// Spell: spectrum (sight school) — 2-D Fourier decomposition of what the eye saw.
//
// The dog casts this on an image to read its STRUCTURE in frequency space: a 2-D
// FFT turns the picture into energy spread over orientation × spatial-frequency.
// That energy is laid into the glyph by the dog's own field geometry — the angle
// of a frequency component → family, its radial frequency → subfamily — the same
// polar convention the eye and the paw use. Self-normalised (log1p). So a script
// of mostly horizontal strokes concentrates energy at one orientation; a circular
// Heptapod logogram spreads it evenly around. Exact, deterministic math on math.
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

const grid = 128 // analysis resolution (power of two for the FFT)

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(err)
	}
	if in.Path == "" {
		fail(fmt.Errorf("spectrum needs an image path"))
	}
	g, err := vision.Load(in.Path, grid)
	if err != nil {
		fail(err)
	}

	mag := fft2(g)        // 2-D magnitude spectrum, DC at [0][0]
	d := toGlyph(mag)     // lay frequency energy into the glyph by geometry
	dom := dominantAngle(mag)

	out := modules.Output{
		Spell:   "spectrum",
		Detail:  d,
		Summary: fmt.Sprintf("fourier: %d active cells · dominant orientation %.0f°", d.Active(), dom),
		Data: map[string]any{
			"complexity":    d.Complexity(),
			"family_spread": d.FamilySpread(),
			"dominant_deg":  dom,
		},
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fail(err)
	}
}

// fft2 runs a separable 2-D FFT (rows then columns) and returns the magnitude.
func fft2(g [][]float64) [][]float64 {
	n := len(g)
	buf := make([][]math.Complex, n)
	for y := 0; y < n; y++ {
		buf[y] = make([]math.Complex, n)
		for x := 0; x < n; x++ {
			buf[y][x] = math.Complex{Re: g[y][x], Im: 0}
		}
		math.FFT(buf[y], false)
	}
	col := make([]math.Complex, n)
	for x := 0; x < n; x++ {
		for y := 0; y < n; y++ {
			col[y] = buf[y][x]
		}
		math.FFT(col, false)
		for y := 0; y < n; y++ {
			buf[y][x] = col[y]
		}
	}
	mag := make([][]float64, n)
	for y := 0; y < n; y++ {
		mag[y] = make([]float64, n)
		for x := 0; x < n; x++ {
			mag[y][x] = buf[y][x].Abs()
		}
	}
	return mag
}

// toGlyph maps each frequency component into the glyph by geometry: its angle
// (orientation) → family, its radial frequency → subfamily, magnitude → strength.
// DC (overall brightness) is skipped — it carries no structure.
func toGlyph(mag [][]float64) encoding.Detail {
	n := len(mag)
	var sum [encoding.Families][encoding.Subfamilies]float64
	for y := 0; y < n; y++ {
		ky := wrap(y, n) // signed frequency (−n/2..n/2)
		for x := 0; x < n; x++ {
			kx := wrap(x, n)
			if kx == 0 && ky == 0 {
				continue // DC
			}
			ang := gm.Atan2(float64(ky), float64(kx))
			if ang < 0 {
				ang += 2 * gm.Pi
			}
			fam := int(ang / (2 * gm.Pi) * encoding.Families)
			if fam >= encoding.Families {
				fam = encoding.Families - 1
			}
			r := gm.Hypot(float64(kx), float64(ky)) / (float64(n) / 2)
			if r > 1 {
				r = 1
			}
			sub := int(r * (encoding.Subfamilies - 1))
			sum[fam][sub] += mag[y][x]
		}
	}
	return normalize(sum)
}

func wrap(i, n int) int {
	if i > n/2 {
		return i - n
	}
	return i
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

// dominantAngle reports the orientation carrying the most spectral energy.
func dominantAngle(mag [][]float64) float64 {
	n := len(mag)
	var bins [180]float64
	for y := 0; y < n; y++ {
		ky := wrap(y, n)
		for x := 0; x < n; x++ {
			kx := wrap(x, n)
			if kx == 0 && ky == 0 {
				continue
			}
			deg := gm.Mod(gm.Atan2(float64(ky), float64(kx))*180/gm.Pi+180, 180)
			bins[int(deg)%180] += mag[y][x]
		}
	}
	best, bv := 0, -1.0
	for i, v := range bins {
		if v > bv {
			bv, best = v, i
		}
	}
	return float64(best)
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "spectrum:", err)
	os.Exit(1)
}