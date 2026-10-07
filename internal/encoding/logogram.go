package encoding

import (
	"sort"
	"strings"
)

// Logograms — Argos's derived script
// ----------------------------------
// Inspired by (not copied from) Ifá Odù and held to a Fable-5 pretense of
// open-ended expression. Two borrowed principles, one new mechanism:
//
//   • From Ifá: a sign is COMPOSED of elemental marks, and a named sign is a
//     VESSEL that accumulates wisdom (its "verses"). Casting reads the vessel.
//   • New mechanism: a logogram is SYLLABIC. Each syllable is
//        family-mark (one of 21 alchemical "consonants" — the glyph encoding)
//      + subfamily-symbol (one of 72 alchemical "vowels")
//      + tone (the level 0..7).
//     Syllables compose without limit, so the script GROWS — the Fable-5
//     pretense — rather than capping at 256.
//
// The dog thinks by sealing meaning into a logogram, then reading the verses
// the matching vessel holds (see the lexicon).

// SubfamilySymbols give each of the 72 subfamilies an alchemical "vowel".
// Filled from a curated alchemical/planetary set, the user's favourites first.
var SubfamilySymbols [Subfamilies]string

func init() {
	pre := []rune{
		'♀', '⊛', '⋈', '☉', '☽', '☿', '♁', '♂', '♃', '♄', '♅', '♆',
		'⚳', '⚴', '⚵', '⚶', '⚷', '⚸', '⚹', '⚺', '⚻', '⚼', '🜔', '🜍',
	}
	i := 0
	for ; i < len(pre) && i < Subfamilies; i++ {
		SubfamilySymbols[i] = string(pre[i])
	}
	// Fill the remainder from the Alchemical Symbols block (U+1F700…).
	r := rune(0x1F700)
	for ; i < Subfamilies; i++ {
		SubfamilySymbols[i] = string(r)
		r++
	}
}

var tones = [8]rune{'⁰', '¹', '²', '³', '⁴', '⁵', '⁶', '⁷'}

// Syllable is one stroke of a logogram: a family, its strongest subfamily, tone.
type Syllable struct {
	Family    int
	Subfamily int
	Level     uint8
}

// Glyph renders a single syllable: consonant(family) · vowel(subfamily) · tone.
func (s Syllable) Glyph() string {
	if s.Family < 0 || s.Family >= Families {
		return ""
	}
	return FamilySymbol(s.Family) + SubfamilySymbols[s.Subfamily] + string(tones[s.Level])
}

// Seal derives the logogram syllables of a detail field: for each active family
// it takes that family's strongest subfamily as the syllable, then keeps the
// maxSyll loudest syllables (the dominant strokes), tone = coarse level.
func Seal(d Detail, maxSyll int) []Syllable {
	g := d.Coarse()
	var sy []Syllable
	for f := 0; f < Families; f++ {
		if g[f] == 0 {
			continue
		}
		bestSub, bestVal := 0, uint8(0)
		for s := 0; s < Subfamilies; s++ {
			if d[f][s] > bestVal {
				bestVal, bestSub = d[f][s], s
			}
		}
		sy = append(sy, Syllable{Family: f, Subfamily: bestSub, Level: g[f]})
	}
	sort.Slice(sy, func(a, b int) bool { return sy[a].Level > sy[b].Level })
	if maxSyll > 0 && len(sy) > maxSyll {
		sy = sy[:maxSyll]
	}
	return sy
}

// Logogram is the written sign for a meaning: its dominant syllables strung
// together. This is what the dog "writes" when it thinks.
func Logogram(d Detail) string {
	sy := Seal(d, 4)
	parts := make([]string, len(sy))
	for i, s := range sy {
		parts[i] = s.Glyph()
	}
	return strings.Join(parts, "")
}

// Word is the logogram for a single concept of text.
func Word(text string) string { return Logogram(Encode(text)) }

// Thought renders the dog thinking in logograms: each salient word mapped to
// its sign. Returns "word→sign" pairs, deduplicated, in order of appearance.
func Thought(text string) []string {
	seen := map[string]bool{}
	var out []string
	for _, w := range strings.Fields(strings.ToLower(text)) {
		w = strings.Trim(w, ".,!?;:\"'()")
		if len(w) <= 3 || seen[w] {
			continue
		}
		seen[w] = true
		out = append(out, w+" "+Word(w))
	}
	return out
}