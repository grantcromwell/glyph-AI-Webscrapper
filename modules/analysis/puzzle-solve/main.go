// puzzle-solve — ARC-AGI deterministic solver spell for Argos.
// Takes train examples + test input, returns predicted output.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
)

type Grid [][]int
type Example struct {
	Input  Grid `json:"input"`
	Output Grid `json:"output"`
}

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail("bad input: %v", err)
	}
	if in.Text == "" {
		fail("needs puzzle JSON in text field")
	}

	var pd struct {
		Train []Example `json:"train"`
		Test  Grid      `json:"test"`
	}
	if err := json.Unmarshal([]byte(in.Text), &pd); err != nil {
		fail("bad puzzle JSON: %v", err)
	}

	pred, rule, families := solve(pd.Train, pd.Test)

	var d encoding.Detail
	for i, v := range families {
		level := int(v)
		if level > 7 {
			level = 7
		}
		for s := 0; s < level && s < encoding.Subfamilies; s++ {
			d[i][s] = uint8(level)
		}
	}

	emit(modules.Output{
		Spell:   "puzzle-solve",
		Detail:  d,
		Summary: fmt.Sprintf("solved: %s", rule),
		Data: map[string]any{
			"predicted": pred,
			"rule":      rule,
		},
	})
}

func solve(train []Example, test Grid) (Grid, string, []float64) {
	g := make([]float64, 21)

	if pred, rule := tryCorrespondence(train, test); pred != nil {
		g[0], g[1] = 0, 7
		return pred, rule, g
	}
	if pred, rule := tryTiling(train, test); pred != nil {
		g[0], g[1] = 1, 7
		return pred, rule, g
	}
	if pred, rule := tryFloodFill(train, test); pred != nil {
		g[0], g[1] = 2, 7
		return pred, rule, g
	}
	if pred, rule := tryShapeDuplicator(train, test); pred != nil {
		g[0], g[1] = 3, 7
		return pred, rule, g
	}
	if pred, rule := tryColorShift(train, test); pred != nil {
		g[0], g[1] = 4, 7
		return pred, rule, g
	}
	if pred, rule := tryVerticalStacking(train, test); pred != nil {
		g[0], g[1] = 5, 4
		return pred, rule, g
	}

	g[0], g[1] = 7, 0
	return test, "unknown pattern", g
}

func tryCorrespondence(train []Example, test Grid) (Grid, string) {
	if len(train) == 0 {
		return nil, ""
	}
	h, w := len(train[0].Input), len(train[0].Input[0])
	for col := 0; col < w; col++ {
		vals := make([]int, h)
		for r := 0; r < h; r++ {
			vals[r] = train[0].Input[r][col]
		}
		if vals[0] == 0 {
			continue
		}
		allSame := true
		for _, v := range vals {
			if v != vals[0] {
				allSame = false
				break
			}
		}
		if !allSame {
			continue
		}
		for _, op := range []string{"AND", "OR", "XOR"} {
			ok := true
			for _, ex := range train {
				l, r := splitCol(ex.Input, col)
				ol, _ := splitCol(ex.Output, col)
				if !gridsEqual(applyOp(l, r, op), ol) {
					ok = false
					break
				}
			}
			if ok {
				l, r := splitCol(test, col)
				return applyOp(l, r, op), fmt.Sprintf("correspondence: %s at col %d", op, col)
			}
		}
	}
	return nil, ""
}

func tryTiling(train []Example, test Grid) (Grid, string) {
	if len(train) == 0 {
		return nil, ""
	}
	ex := train[0]
	hi, wi := len(ex.Input), len(ex.Input[0])
	ho, wo := len(ex.Output), len(ex.Output[0])
	if ho%hi != 0 || wo%wi != 0 {
		return nil, ""
	}
	ty, tx := ho/hi, wo/wi
	h, w := len(test), len(test[0])
	out := make(Grid, h*ty)
	for y := 0; y < h*ty; y++ {
		row := make([]int, w*tx)
		for x := 0; x < w*tx; x++ {
			row[x] = test[y%h][x%w]
		}
		out[y] = row
	}
	return out, fmt.Sprintf("tile %dx%d", ty, tx)
}

func tryFloodFill(train []Example, test Grid) (Grid, string) {
	if len(train) == 0 {
		return nil, ""
	}
	ex := train[0]
	hi, wi := len(ex.Input), len(ex.Input[0])
	ho, wo := len(ex.Output), len(ex.Output[0])
	if hi != ho || wi != wo {
		return nil, ""
	}
	inCols := colorSet(ex.Input)
	outCols := colorSet(ex.Output)
	newCols := diffSet(outCols, inCols)
	if len(newCols) == 0 {
		return nil, ""
	}
	fillColor := -1
	for c := range newCols {
		fillColor = c
		break
	}
	return applyFloodFill(test, fillColor), fmt.Sprintf("flood fill with %d", fillColor)
}

func tryShapeDuplicator(train []Example, test Grid) (Grid, string) {
	if len(train) == 0 {
		return nil, ""
	}
	ex := train[0]
	hi, wi := len(ex.Input), len(ex.Input[0])
	ho, wo := len(ex.Output), len(ex.Output[0])
	if ho == hi*3 && wo == wi*3 {
		return applyShapeDuplicator(test, 3), "shape duplicator 3x3"
	}
	return nil, ""
}

func tryColorShift(train []Example, test Grid) (Grid, string) {
	m := map[int]int{}
	for _, ex := range train {
		hi, wi := len(ex.Input), len(ex.Input[0])
		ho, wo := len(ex.Output), len(ex.Output[0])
		if hi != ho || wi != wo {
			return nil, ""
		}
		for y := 0; y < hi; y++ {
			for x := 0; x < wi; x++ {
				iv := ex.Input[y][x]
				ov := ex.Output[y][x]
				if iv != 0 {
					if p, ok := m[iv]; ok && p != ov {
						return nil, ""
					}
					m[iv] = ov
				}
			}
		}
	}
	if len(m) == 0 {
		return nil, ""
	}
	out := make(Grid, len(test))
	for y := 0; y < len(test); y++ {
		row := make([]int, len(test[0]))
		for x := 0; x < len(test[0]); x++ {
			if v, ok := m[test[y][x]]; ok {
				row[x] = v
			} else {
				row[x] = test[y][x]
			}
		}
		out[y] = row
	}
	return out, fmt.Sprintf("color shift: %v", m)
}

func tryVerticalStacking(train []Example, test Grid) (Grid, string) {
	if len(train) == 0 {
		return nil, ""
	}
	ex := train[0]
	hi, wi := len(ex.Input), len(ex.Input[0])
	ho, wo := len(ex.Output), len(ex.Output[0])
	if wo != wi || ho < hi {
		return nil, ""
	}
	h := len(test)
	out := make(Grid, ho)
	for y := 0; y < ho; y++ {
		out[y] = make([]int, wi)
		for x := 0; x < wi; x++ {
			out[y][x] = test[y%h][x]
		}
	}
	return out, fmt.Sprintf("vertical stack to %d", ho)
}

func splitCol(g Grid, col int) (Grid, Grid) {
	h := len(g)
	l, r := make(Grid, h), make(Grid, h)
	for y := 0; y < h; y++ {
		l[y] = append([]int{}, g[y][:col]...)
		r[y] = append([]int{}, g[y][col+1:]...)
	}
	return l, r
}

func applyOp(l, r Grid, op string) Grid {
	h, w := len(l), len(l[0])
	out := make(Grid, h)
	for y := 0; y < h; y++ {
		out[y] = make([]int, w)
		for x := 0; x < w; x++ {
			switch op {
			case "AND":
				if l[y][x] > 0 && r[y][x] > 0 {
					out[y][x] = 2
				}
			case "OR":
				if l[y][x] > 0 || r[y][x] > 0 {
					out[y][x] = 2
				}
			case "XOR":
				if (l[y][x] > 0 || r[y][x] > 0) && l[y][x] != r[y][x] {
					out[y][x] = 2
				}
			}
		}
	}
	return out
}

func applyFloodFill(input Grid, fillColor int) Grid {
	h, w := len(input), len(input[0])
	out := make(Grid, h)
	for y := 0; y < h; y++ {
		out[y] = append([]int{}, input[y]...)
	}
	bg := make([][]bool, h)
	for y := 0; y < h; y++ {
		bg[y] = make([]bool, w)
	}
	type pt struct{ y, x int }
	q := []pt{}
	for y := 0; y < h; y++ {
		if input[y][0] == 0 {
			q = append(q, pt{y, 0})
			bg[y][0] = true
		}
		if input[y][w-1] == 0 {
			q = append(q, pt{y, w - 1})
			bg[y][w-1] = true
		}
	}
	for x := 0; x < w; x++ {
		if input[0][x] == 0 {
			q = append(q, pt{0, x})
			bg[0][x] = true
		}
		if input[h-1][x] == 0 {
			q = append(q, pt{h - 1, x})
			bg[h-1][x] = true
		}
	}
	for len(q) > 0 {
		p := q[0]
		q = q[1:]
		for _, d := range [][2]int{{0, 1}, {0, -1}, {1, 0}, {-1, 0}} {
			ny, nx := p.y+d[0], p.x+d[1]
			if ny >= 0 && ny < h && nx >= 0 && nx < w && !bg[ny][nx] && input[ny][nx] == 0 {
				bg[ny][nx] = true
				q = append(q, pt{ny, nx})
			}
		}
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if input[y][x] == 0 && !bg[y][x] {
				out[y][x] = fillColor
			}
		}
	}
	return out
}

func applyShapeDuplicator(input Grid, factor int) Grid {
	h, w := len(input), len(input[0])
	out := make(Grid, h*factor)
	for y := 0; y < h*factor; y++ {
		out[y] = make([]int, w*factor)
	}
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			if input[y][x] > 0 {
				for dy := 0; dy < h; dy++ {
					for dx := 0; dx < w; dx++ {
						if input[dy][dx] > 0 {
							out[y*factor+dy][x*factor+dx] = input[dy][dx]
						}
					}
				}
			}
		}
	}
	return out
}

func colorSet(g Grid) map[int]bool {
	s := map[int]bool{}
	for _, row := range g {
		for _, v := range row {
			s[v] = true
		}
	}
	return s
}

func diffSet(a, b map[int]bool) map[int]bool {
	r := map[int]bool{}
	for k := range a {
		if !b[k] {
			r[k] = true
		}
	}
	return r
}

func gridsEqual(a, b Grid) bool {
	if len(a) != len(b) {
		return false
	}
	for y := range a {
		if len(a[y]) != len(b[y]) {
			return false
		}
		for x := range a[y] {
			if a[y][x] != b[y][x] {
				return false
			}
		}
	}
	return true
}

func emit(o modules.Output) { _ = json.NewEncoder(os.Stdout).Encode(o) }
func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "puzzle-solve: "+format+"\n", args...)
	os.Exit(1)
}