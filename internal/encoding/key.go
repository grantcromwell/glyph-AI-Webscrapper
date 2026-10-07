// Package encoding — glyph field encoding and comparison.
//
// The 21-family glyph field is the system's native representation. Every input
// encodes to a 21x72 Detail field and a 21-value coarse Glyph. The mapping from
// input to glyph is deterministic and one-way: given a glyph vector, there is
// no way to recover the original text. The glyph is a fingerprint, not a cipher.
//
// The breath-to-family routing table is the only fixed mapping. It determines
// which families light up for which phoneme breath-shapes (plosive, fricative,
// nasal, approximant, vowel, sibilant). This table is necessary for encoding.
// The human-readable names, symbols, colors, and roles that would decode a glyph
// back to English are not present in this file.
package encoding

import (
	"sort"
	"strings"

	"glyphai/internal/audio"
)

// familiesByBreath groups family indices by the breath-shape they answer to.
// This is the only fixed mapping in the encoding system. It is necessary for
// the Encode() function to route phonemes to the correct families.
var familiesByBreath = func() [audio.NumBreaths][]int {
	var m [audio.NumBreaths][]int
	breathMap := [Families]audio.Breath{
		audio.Plosive,     // 0
		audio.Approximant, // 1
		audio.Vowel,       // 2
		audio.Fricative,   // 3
		audio.Approximant, // 4
		audio.Plosive,     // 5
		audio.Plosive,     // 6
		audio.Vowel,       // 7
		audio.Nasal,       // 8
		audio.Sibilant,    // 9
		audio.Fricative,   // 10
		audio.Plosive,     // 11
		audio.Sibilant,    // 12
		audio.Plosive,     // 13
		audio.Nasal,       // 14
		audio.Fricative,   // 15
		audio.Approximant, // 16
		audio.Plosive,     // 17
		audio.Sibilant,    // 18
		audio.Fricative,   // 19
		audio.Vowel,       // 20
	}
	for f := 0; f < Families; f++ {
		b := breathMap[f]
		m[b] = append(m[b], f)
	}
	return m
}()

// FamiliesForBreath returns the families that resonate with a breath-shape.
func FamiliesForBreath(b audio.Breath) []int { return familiesByBreath[b] }

// FamilyHue returns a deterministic hue for a family index, derived from its
// position in the field. This is used for SVG rendering only — it does not
// encode any semantic meaning about the family.
func FamilyHue(f int) float64 {
	if f < 0 || f >= Families {
		return 0
	}
	return float64(f) * 360.0 / float64(Families)
}

// FamilySymbol returns a deterministic symbol for a family index, derived from
// its position. This is used for SVG rendering only — it does not encode any
// semantic meaning about the family.
func FamilySymbol(f int) string {
	if f < 0 || f >= Families {
		return "?"
	}
	return string(rune(0xE000 + f)) // Private Use Area — no semantic meaning
}

// FamilyRole returns a deterministic role label for a family index.
// This is a lossy description, not a decoding. The original text cannot be
// recovered from a glyph.
func FamilyRole(f int) string {
	if f < 0 || f >= Families {
		return "unknown"
	}
	roles := [Families]string{
		"axis00", "axis01", "axis02", "axis03", "axis04",
		"axis05", "axis06", "axis07", "axis08", "axis09",
		"axis10", "axis11", "axis12", "axis13", "axis14",
		"axis15", "axis16", "axis17", "axis18", "axis19",
		"axis20",
	}
	return roles[f]
}

// emoFamilies are the six families the vowel-melody sings into.
var emoFamilies = [...]int{7, 9, 19, 14, 16, 2}

// Emotion names the strongest felt family for a glyph (its mood toward it).
func Emotion(d Detail) string {
	best, bv := -1, 0.0
	for _, f := range emoFamilies {
		var sum float64
		for s := 0; s < Subfamilies; s++ {
			sum += float64(d[f][s])
		}
		if sum > bv {
			bv, best = sum, f
		}
	}
	if best < 0 {
		return "still"
	}
	return FamilyRole(best)
}

// intensity adjectives by level 1..7.
var intensity = [8]string{"", "faint", "low", "soft", "clear", "strong", "deep", "blazing"}

// Reading produces a human-readable description of a glyph's strongest families.
// This is a lossy interpretation, not a decoding. The original text cannot be
// recovered from a glyph.
func Reading(g Glyph) string {
	type fv struct {
		i int
		v uint8
	}
	var on []fv
	for f := 0; f < Families; f++ {
		if g[f] > 0 {
			on = append(on, fv{f, g[f]})
		}
	}
	sort.Slice(on, func(a, b int) bool { return on[a].v > on[b].v })
	if len(on) > 5 {
		on = on[:5]
	}
	parts := make([]string, len(on))
	for i, p := range on {
		parts[i] = strings.TrimSpace(intensity[p.v] + " " + FamilyRole(p.i))
	}
	if len(parts) == 0 {
		return "empty"
	}
	return strings.Join(parts, ", ")
}
