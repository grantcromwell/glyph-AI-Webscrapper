// Spell: distill (reasoning school) — find the densest grain of a long text.
//
// The dog drowned in long pages and dumped raw boilerplate. distill slides a
// window across the text and keeps the passage whose glyph is RICHEST (most
// active cells) — the most information-dense grain — so the dog answers from
// substance, not chrome. Salience derived from glyph richness, never a wordlist.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
)

const win = 50 // words per window (structural, like a fovea)

func main() {
	var in modules.Input
	_ = json.NewDecoder(os.Stdin).Decode(&in)
	words := strings.Fields(in.Text)
	if len(words) == 0 {
		fmt.Fprintln(os.Stderr, "distill: needs text"); os.Exit(1)
	}
	best, bestD, bestScore := in.Text, encoding.Encode(in.Text), -1
	for i := 0; i < len(words); i += win / 2 {
		j := i + win
		if j > len(words) {
			j = len(words)
		}
		passage := strings.Join(words[i:j], " ")
		d := encoding.Encode(passage)
		if a := d.Active(); a > bestScore {
			bestScore, best, bestD = a, passage, d
		}
		if j == len(words) {
			break
		}
	}
	emit(modules.Output{
		Spell:   "distill",
		Detail:  bestD,
		Summary: fmt.Sprintf("densest grain (%d/1512 cells) of %d words", bestScore, len(words)),
		Data:    map[string]any{"passage": best, "richness": bestScore},
	})
}

func emit(o modules.Output) { _ = json.NewEncoder(os.Stdout).Encode(o) }