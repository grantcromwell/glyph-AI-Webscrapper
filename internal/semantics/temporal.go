// Package temporal is Argos's meaning-depth lobe. Where the octopus carries
// each glyph's etymology (PIE -> historical -> current), this lobe traces a
// word's depth through the layers the dog can derive on its own:
//
//	surface  — the word as spoken
//	root     — its consonant skeleton (the bones of meaning)
//	breath   — the mouth-physics shape that makes it
//
// The deeper and richer those layers, the more "meaning weight" the word
// carries. (Real PIE/Sanskrit etymology is a future Wiktionary site-spell; this
// lobe is the structural depth the dog reads unaided.)
package semantics

import (
	"strings"

	"glyphai/internal/audio"
)

// Depth is the layered reading of a word.
type Depth struct {
	Surface string
	Root    string  // consonant skeleton
	Breath  string  // dominant breath-shape name
	Score   float64 // 0..1 meaning weight
}

// Trace layers a single word.
func Trace(word string) Depth {
	w := strings.ToLower(strings.TrimSpace(word))
	root := consonantSkeleton(w)
	prof := audio.Profile(w)
	// Depth weight: longer roots and consonant-rich words carry more meaning;
	// pure-vowel function words carry little.
	score := 0.0
	if len(w) > 0 {
		score = float64(len(root)) / float64(len(w))
	}
	return Depth{Surface: w, Root: root, Breath: audio.BreathName[prof], Score: score}
}

// Weight is the mean meaning-depth across a text's words — how much the phrase
// carries beyond grammatical scaffolding.
func Weight(text string) float64 {
	words := strings.Fields(text)
	if len(words) == 0 {
		return 0
	}
	var sum float64
	for _, w := range words {
		sum += Trace(w).Score
	}
	return sum / float64(len(words))
}

// Reading describes the deepest (most meaning-bearing) word of a text.
func Reading(text string) string {
	var best Depth
	for _, w := range strings.Fields(text) {
		if d := Trace(w); d.Score > best.Score {
			best = d
		}
	}
	if best.Surface == "" {
		return "shallow"
	}
	return best.Surface + " ⟵ root " + best.Root + " · " + best.Breath
}

func consonantSkeleton(w string) string {
	var b strings.Builder
	for _, r := range w {
		if r >= 'a' && r <= 'z' && !isVowel(r) {
			b.WriteRune(r)
		}
	}
	return b.String()
}

func isVowel(r rune) bool {
	switch r {
	case 'a', 'e', 'i', 'o', 'u', 'y':
		return true
	}
	return false
}