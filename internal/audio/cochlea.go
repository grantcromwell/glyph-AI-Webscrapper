// Package cochlea is Argos's sound lobe: it hears the BREATH-SHAPE of language —
// the physics of the mouth (manner of articulation), not the script. Latin
// letters use the dog's one innate manner table (its native alphabet, anatomy);
// EVERY other script is heard agnostically — each rune's breath emerges from the
// rune itself, never from a per-language table — so a tongue never seen before,
// or invented tomorrow, is still heard. Nothing about specific languages is
// hardcoded; the only place the world's phonologies live is the `phoible`
// knowledge-spell, which the dog may CONSULT but is never wired into the ear.
//
// (Written script is also art: the dog's SIGHT can turn any mark — Latin,
// Heptapod, a glyph that does not yet exist — into a glyph as well. See the
// image-glyph spell. The ear and the eye both always produce a encoding.)
package audio

import (
	"hash/fnv"
	"strings"
	"unicode"
)

// Breath is a manner of articulation — a shape the breath takes.
type Breath uint8

const (
	Plosive     Breath = iota // air held then released: sudden, decisive
	Fricative                 // air through constriction: flowing, constrained
	Nasal                     // bone resonance: resonant, sustained
	Approximant               // narrowed channel: sliding, transitional
	Vowel                     // open channel, pure harmonic: open, vulnerable
	Sibilant                  // high-frequency fricative: sharp, piercing
	NumBreaths
)

// BreathName / BreathSymbol / BreathQuality describe each shape.
var (
	BreathName    = [NumBreaths]string{"plosive", "fricative", "nasal", "approximant", "vowel", "sibilant"}
	BreathSymbol  = [NumBreaths]string{".|", "~", "n~", "~~", "--", "/|\\"}
	BreathQuality = [NumBreaths]string{"sudden", "flowing", "resonant", "sliding", "open", "piercing"}
)

// manner maps each Latin letter to its breath-shape — the dog's one innate
// alphabet (anatomy). Every non-Latin rune is heard emergently (see Manner).
var manner = map[rune]Breath{
	'b': Plosive, 'p': Plosive, 't': Plosive, 'd': Plosive, 'k': Plosive,
	'g': Plosive, 'c': Plosive, 'q': Plosive,
	'f': Fricative, 'v': Fricative, 'h': Fricative,
	'm': Nasal, 'n': Nasal,
	'l': Approximant, 'r': Approximant, 'w': Approximant, 'y': Approximant,
	'a': Vowel, 'e': Vowel, 'i': Vowel, 'o': Vowel, 'u': Vowel,
	's': Sibilant, 'z': Sibilant, 'x': Sibilant, 'j': Sibilant,
}

// Manner returns the breath-shape of a rune. Latin via the innate table; ANY
// other rune (any script, real or invented) by an emergent breath drawn from the
// rune itself — so every glyph is heard, and no language is ever hardcoded.
func Manner(r rune) Breath {
	lr := unicode.ToLower(r)
	if b, ok := manner[lr]; ok {
		return b
	}
	return Breath(hashRune(lr) % uint64(NumBreaths))
}

// BreathOf classifies a PHOIBLE-style segment into a breath from its feature
// values — used only by the `phoible` knowledge-spell (NOT by the ear). feat(name)
// returns the raw feature string ("+","-","0","",…).
func BreathOf(feat func(string) string) Breath {
	pos := func(n string) bool { return strings.HasPrefix(feat(n), "+") }
	switch {
	case pos("click"):
		return Plosive
	case !pos("consonantal") && pos("syllabic"): // peel vowels first: +approximant includes vocoids
		return Vowel
	case pos("nasal"):
		return Nasal
	case pos("trill"), pos("tap"), pos("lateral"), pos("approximant"):
		return Approximant
	case pos("strident"):
		return Sibilant
	case pos("continuant"):
		return Fricative
	default:
		return Plosive
	}
}

func hashRune(r rune) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(string(r)))
	return h.Sum64()
}

// Decompose returns the breath-shape sequence of a text (letters and marks of
// any script).
func Decompose(text string) []Breath {
	var out []Breath
	for _, r := range text {
		if unicode.IsLetter(r) || unicode.IsMark(r) {
			out = append(out, Manner(r))
		}
	}
	return out
}

// Counts tallies how many of each breath-shape a text contains.
func Counts(text string) [NumBreaths]int {
	var c [NumBreaths]int
	for _, b := range Decompose(text) {
		c[b]++
	}
	return c
}

// Profile returns the dominant breath-shape of a text (its mouth-feel).
func Profile(text string) Breath {
	c := Counts(text)
	best, bv := Vowel, -1
	for b := Breath(0); b < NumBreaths; b++ {
		if c[b] > bv {
			bv, best = c[b], b
		}
	}
	return best
}