// Spell: image-glyph — the dog's EYE (native geometric sight).
//
// The eye is anatomy: it interprets GEOMETRY as information. It reads an image
// and lays its visual structure into the glyph by the dog's OWN field geometry —
// the same polar convention the paw renders glyphs with (angle around the centre
// → family, distance from the centre → subfamily). Seeing is the exact inverse
// of drawing. A cell's strength is the visual structure (local contrast/edges)
// found at that angle and radius; levels are self-normalised the same way the ear
// normalises sound (log1p, relative to the strongest cell). There is no tuned
// threshold, no per-feature family table — the only constants are the glyph's own
// anatomy (Families / Subfamilies / MaxLevel). Any image yields a glyph: a triangle,
// a painting, a Cyrillic word, an alien logogram.
//
// DECOMPOSITION and pattern analysis (wavelet, spectrum, …) are SPELLS the dog
// casts on what the eye has seen, to come to conclusions — never part of the eye.
package main

import (
	"encoding/json"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"math"
	"os"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
)

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(err)
	}
	if in.Path == "" {
		fail(fmt.Errorf("image-glyph needs a path"))
	}
	f, err := os.Open(in.Path)
	if err != nil {
		fail(err)
	}
	defer f.Close()
	img, format, err := image.Decode(f)
	if err != nil {
		fail(err)
	}

	d, active := perceive(img)
	out := modules.Output{
		Spell:   "image-glyph",
		Detail:  d,
		Summary: fmt.Sprintf("saw %s image — geometry read into %d glyph cells", format, active),
		Data: map[string]any{
			"format":        format,
			"complexity":    d.Complexity(),
			"family_spread": d.FamilySpread(),
		},
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fail(err)
	}
}

// perceive lays an image's visual structure into the glyph by pure geometry:
// angle-from-centre → family, radius-from-centre → subfamily, local contrast →
// strength. Self-normalised (log1p, peaked). The eye's inverse is the paw's
// fieldAt, which turns a glyph back into a picture.
func perceive(img image.Image) (encoding.Detail, int) {
	b := img.Bounds()
	w, h := b.Dx(), b.Dy()
	if w == 0 || h == 0 {
		return encoding.Detail{}, 0
	}
	lum := func(x, y int) float64 {
		r, g, bl, _ := img.At(b.Min.X+x, b.Min.Y+y).RGBA()
		return 0.299*float64(r>>8) + 0.587*float64(g>>8) + 0.114*float64(bl>>8)
	}
	cx, cy := float64(w)/2, float64(h)/2
	maxR := math.Hypot(cx, cy)
	if maxR == 0 {
		return encoding.Detail{}, 0
	}

	var sum [encoding.Families][encoding.Subfamilies]float64
	for y := 0; y < h; y++ {
		for x := 0; x < w; x++ {
			l := lum(x, y)
			e := 0.0 // structural energy = local contrast (where the image changes)
			if x+1 < w {
				e += math.Abs(l - lum(x+1, y))
			}
			if y+1 < h {
				e += math.Abs(l - lum(x, y+1))
			}
			if e == 0 {
				continue
			}
			dx, dy := float64(x)-cx, float64(y)-cy
			ang := math.Atan2(dy, dx)
			if ang < 0 {
				ang += 2 * math.Pi
			}
			fam := int(ang / (2 * math.Pi) * encoding.Families)
			if fam >= encoding.Families {
				fam = encoding.Families - 1
			}
			sub := int(math.Hypot(dx, dy) / maxR * encoding.Subfamilies)
			if sub >= encoding.Subfamilies {
				sub = encoding.Subfamilies - 1
			}
			sum[fam][sub] += e
		}
	}

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
		return d, 0
	}
	lmax := math.Log1p(max)
	active := 0
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			if sum[f][s] == 0 {
				continue
			}
			lvl := int(math.Round(math.Log1p(sum[f][s]) / lmax * encoding.MaxLevel))
			if lvl > 0 {
				d.Set(f, s, uint8(lvl))
				active++
			}
		}
	}
	return d, active
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "image-glyph:", err)
	os.Exit(1)
}