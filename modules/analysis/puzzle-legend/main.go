// puzzle-legend — ARC-AGI pattern classifier spell for Argos.
// Input: modules.Input.Text = JSON with train+test examples.
// Output: detected puzzle type as glyph + summary.
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

type PuzzleClass struct {
	Type       string      `json:"type"`
	Rule       string      `json:"rule"`
	Confidence float64     `json:"confidence"`
	Separator  int         `json:"separator_col,omitempty"`
	TileY      int         `json:"tile_y,omitempty"`
	TileX      int         `json:"tile_x,omitempty"`
	Factor     int         `json:"factor,omitempty"`
	FillColor  int         `json:"fill_color,omitempty"`
	ColorMap   map[int]int `json:"color_map,omitempty"`
}

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail("bad input: %v", err)
	}
	if in.Text == "" {
		fail("needs puzzle JSON in text field")
	}

	var puzzleData struct {
		Train []Example `json:"train"`
	}
	if err := json.Unmarshal([]byte(in.Text), &puzzleData); err != nil {
		fail("bad puzzle JSON: %v", err)
	}

	legend := classify(puzzleData.Train)
	g := legendToGlyph(legend, puzzleData.Train)

	// Convert to encoding.Detail (21x72 field) for the grimoire output
	var d encoding.Detail
	for i, v := range g {
		level := int(v)
		if level > 7 {
			level = 7
		}
		for s := 0; s < level && s < encoding.Subfamilies; s++ {
			d[i][s] = uint8(level)
		}
	}

	emit(modules.Output{
		Spell:   "puzzle-legend",
		Detail:  d,
		Summary: fmt.Sprintf("puzzle class: %s (%.0f%%)", legend.Type, legend.Confidence*100),
		Data: map[string]any{
			"legend": legend,
			"type":   legend.Type,
			"rule":   legend.Rule,
		},
	})
}

func classify(train []Example) PuzzleClass {
	if len(train) == 0 {
		return PuzzleClass{Type: "unknown", Confidence: 0, Rule: "no training data"}
	}
	var candidates []PuzzleClass
	if c := checkCorrespondence(train); c != nil {
		candidates = append(candidates, *c)
	}
	if c := checkTiling(train); c != nil {
		candidates = append(candidates, *c)
	}
	if c := checkFloodFill(train); c != nil {
		candidates = append(candidates, *c)
	}
	if c := checkShapeDuplicator(train); c != nil {
		candidates = append(candidates, *c)
	}
	if c := checkColorShift(train); c != nil {
		candidates = append(candidates, *c)
	}
	if c := checkVerticalStacking(train); c != nil {
		candidates = append(candidates, *c)
	}
	if len(candidates) > 0 {
		return candidates[0]
	}
	return PuzzleClass{Type: "unknown", Confidence: 0, Rule: "no known pattern"}
}

func checkCorrespondence(train []Example) *PuzzleClass {
	if len(train) == 0 {
		return nil
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
				// Compare against the corresponding expected output, also split by separator
				ol, _ := splitCol(ex.Output, col)
				if !gridsEqual(applyOp(l, r, op), ol) {
					ok = false
					break
				}
			}
			if ok {
				return &PuzzleClass{
					Type: "correspondence", Rule: fmt.Sprintf("split col %d %s", col, op),
					Confidence: 1.0, Separator: col,
				}
			}
		}
	}
	return nil
}

func checkTiling(train []Example) *PuzzleClass {
	ex := train[0]
	hi, wi := len(ex.Input), len(ex.Input[0])
	ho, wo := len(ex.Output), len(ex.Output[0])
	if ho%hi != 0 || wo%wi != 0 {
		return nil
	}
	return &PuzzleClass{
		Type: "tiling", Rule: fmt.Sprintf("tile %dx%d", ho/hi, wo/wi),
		Confidence: 1.0, TileY: ho / hi, TileX: wo / wi,
	}
}

func checkFloodFill(train []Example) *PuzzleClass {
	ex := train[0]
	hi, wi := len(ex.Input), len(ex.Input[0])
	ho, wo := len(ex.Output), len(ex.Output[0])
	if hi != ho || wi != wo {
		return nil
	}
	inCols := colorSet(ex.Input)
	outCols := colorSet(ex.Output)
	newCols := diffSet(outCols, inCols)
	if len(newCols) == 0 {
		return nil
	}
	fillColor := -1
	for c := range newCols {
		fillColor = c
		break
	}
	for _, ex := range train {
		if !verifyFloodFill(ex.Input, ex.Output, fillColor) {
			return nil
		}
	}
	return &PuzzleClass{
		Type: "flood-fill", Rule: fmt.Sprintf("fill with %d", fillColor),
		Confidence: 1.0, FillColor: fillColor,
	}
}

func checkShapeDuplicator(train []Example) *PuzzleClass {
	ex := train[0]
	hi, wi := len(ex.Input), len(ex.Input[0])
	ho, wo := len(ex.Output), len(ex.Output[0])
	if ho == hi*3 && wo == wi*3 {
		return &PuzzleClass{
			Type: "shape-duplicator", Rule: "3x3 dup",
			Confidence: 1.0, Factor: 3,
		}
	}
	return nil
}

func checkColorShift(train []Example) *PuzzleClass {
	m := map[int]int{}
	for _, ex := range train {
		hi, wi := len(ex.Input), len(ex.Input[0])
		ho, wo := len(ex.Output), len(ex.Output[0])
		if hi != ho || wi != wo {
			return nil
		}
		for y := 0; y < hi; y++ {
			for x := 0; x < wi; x++ {
				iv := ex.Input[y][x]
				ov := ex.Output[y][x]
				if iv != 0 {
					if p, ok := m[iv]; ok && p != ov {
						return nil
					}
					m[iv] = ov
				}
			}
		}
	}
	if len(m) > 0 {
		return &PuzzleClass{
			Type: "color-shift", Rule: fmt.Sprintf("map %v", m),
			Confidence: 1.0, ColorMap: m,
		}
	}
	return nil
}

func checkVerticalStacking(train []Example) *PuzzleClass {
	ex := train[0]
	hi, wi := len(ex.Input), len(ex.Input[0])
	ho, wo := len(ex.Output), len(ex.Output[0])
	if wo == wi && ho >= hi {
		return &PuzzleClass{
			Type: "vertical-stacking", Rule: fmt.Sprintf("stack to %d", ho),
			Confidence: 1.0, TileY: ho / hi,
		}
	}
	return nil
}

// ─── Helpers ─────────────────────────────────────────

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

func verifyFloodFill(input, output Grid, fillColor int) bool {
	h, w := len(input), len(input[0])
	if h != len(output) || w != len(output[0]) {
		return false
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
				if output[y][x] != fillColor && output[y][x] != 0 {
					return false
				}
			}
			if input[y][x] != 0 && output[y][x] != input[y][x] {
				return false
			}
		}
	}
	return true
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

func legendToGlyph(l PuzzleClass, train []Example) []float64 {
	g := make([]float64, 21)
	typeIdx := map[string]int{
		"correspondence":    0,
		"tiling":            1,
		"flood-fill":        2,
		"shape-duplicator":  3,
		"color-shift":       4,
		"vertical-stacking": 5,
		"unknown":           7,
	}
	g[0] = float64(typeIdx[l.Type])
	g[1] = l.Confidence * 7
	g[2] = float64(l.Separator)
	if len(train) > 0 {
		g[3] = float64(len(train[0].Input))
	}
	if len(train) > 0 && len(train[0].Input) > 0 {
		g[4] = float64(len(train[0].Input[0]))
	}
	g[5] = float64(len(train))
	g[6] = float64(l.FillColor)
	g[7] = float64(l.TileY * l.TileX)
	if l.Separator > 0 {
		g[8] = 7
	}
	if len(train) > 0 {
		allColors := map[int]bool{}
		for _, ex := range train {
			for _, row := range ex.Input {
				for _, v := range row {
					allColors[v] = true
				}
			}
		}
		g[9] = float64(len(allColors))
	}
	return g
}

func emit(o modules.Output) { _ = json.NewEncoder(os.Stdout).Encode(o) }
func fail(format string, args ...any) {
	fmt.Fprintf(os.Stderr, "puzzle-legend: "+format+"\n", args...)
	os.Exit(1)
}