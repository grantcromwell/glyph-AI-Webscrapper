// Spell: wavelet (sight school) — multi-scale Haar decomposition of what the eye
// saw. Where `spectrum` reads orientation × frequency globally, `wavelet` reads
// structure at every SCALE and WHERE it sits: a 2-D Haar transform peels the
// image into coarse→fine detail bands; each detail coefficient is laid into the
// glyph by geometry — the position of the detail → family (angle from centre),
// the scale it lives at → subfamily (coarse rings inward, fine rings outward),
// its magnitude → strength. Self-normalised (log1p). Exact math on bones; the dog
// casts it to feel an image's structure from broad masses down to fine strokes.
package main

import (
	"encoding/json"
	"fmt"
	gm "math"
	"os"

	"glyphai/internal/math"
	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"glyphai/internal/vision"
)

const grid = 128 // analysis resolution (power of two)

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(err)
	}
	if in.Path == "" {
		fail(fmt.Errorf("wavelet needs an image path"))
	}
	g, err := vision.Load(in.Path, grid)
	if err != nil {
		fail(err)
	}

	d, levels := decompose(g)
	out := modules.Output{
		Spell:   "wavelet",
		Detail:  d,
		Summary: fmt.Sprintf("haar: %d active cells across %d scales", d.Active(), levels),
		Data: map[string]any{
			"complexity":    d.Complexity(),
			"family_spread": d.FamilySpread(),
			"scales":        levels,
		},
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fail(err)
	}
}

// decompose runs a separable 2-D Haar transform level by level. At each level the
// active block splits into LL (recurse) and three detail sub-bands; every detail
// coefficient drops into the glyph by its position (→family) and the level (→
// subfamily ring). Returns the glyph and the number of scales analysed.
func decompose(g [][]float64) (encoding.Detail, int) {
	n := len(g)
	levels := 0
	for s := n; s >= 2; s >>= 1 {
		levels++
	}
	var sum [encoding.Families][encoding.Subfamilies]float64
	li := 0
	for s := n; s >= 2; s, li = s>>1, li+1 {
		// rows: approx → left half, detail → right half.
		for y := 0; y < s; y++ {
			a, dt := math.HaarForward(g[y][:s])
			copy(g[y][:s/2], a)
			copy(g[y][s/2:s], dt)
		}
		// columns: approx → top half, detail → bottom half.
		col := make(math.Vec, s)
		for x := 0; x < s; x++ {
			for y := 0; y < s; y++ {
				col[y] = g[y][x]
			}
			a, dt := math.HaarForward(col)
			for y := 0; y < s/2; y++ {
				g[y][x] = a[y]
				g[y+s/2][x] = dt[y]
			}
		}
		// the three detail sub-bands at this scale: HL, LH, HH.
		half := s / 2
		ring := li * encoding.Subfamilies / levels
		if ring >= encoding.Subfamilies {
			ring = encoding.Subfamilies - 1
		}
		bands := [3][2]int{{0, half}, {half, 0}, {half, half}} // (y0,x0) of HL, LH, HH
		for _, bd := range bands {
			for i := 0; i < half; i++ {
				py := (float64(bd[0]+i) + 0.5) / float64(s)
				for j := 0; j < half; j++ {
					v := gm.Abs(g[bd[0]+i][bd[1]+j])
					if v == 0 {
						continue
					}
					px := (float64(bd[1]+j) + 0.5) / float64(s)
					ang := gm.Atan2(py-0.5, px-0.5)
					if ang < 0 {
						ang += 2 * gm.Pi
					}
					fam := int(ang / (2 * gm.Pi) * encoding.Families)
					if fam >= encoding.Families {
						fam = encoding.Families - 1
					}
					sum[fam][ring] += v
				}
			}
		}
	}
	return normalize(sum), levels
}

func normalize(sum [encoding.Families][encoding.Subfamilies]float64) encoding.Detail {
	max := 0.0
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			if sum[f][s] > max {
				max = sum[f][s]
			}
		}
	}
	var d encoding.Detail
	if max == 0 {
		return d
	}
	lmax := gm.Log1p(max)
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			if sum[f][s] == 0 {
				continue
			}
			if lvl := int(gm.Round(gm.Log1p(sum[f][s]) / lmax * encoding.MaxLevel)); lvl > 0 {
				d.Set(f, s, uint8(lvl))
			}
		}
	}
	return d
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "wavelet:", err)
	os.Exit(1)
}