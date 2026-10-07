// Spell: entropy (reasoning school) — how much INFORMATION a text really carries.
//
// The dog kept getting hijacked by keyword-spam pages (a disambiguation page that
// says "Intelligence" 40 times) because raw resonance rewards repetition. This
// spell measures the Shannon entropy of the text's glyph field: rich, varied
// meaning spreads energy across many cells (high entropy); thin, repetitive spam
// concentrates it on a few (low entropy). The ratio to the maximum possible is
// the dog's read on "is this substance or noise" — derived, no threshold.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
)

func main() {
	var in modules.Input
	_ = json.NewDecoder(os.Stdin).Decode(&in)
	if in.Text == "" {
		fail()
	}
	d := encoding.Encode(in.Text)

	var weights []float64
	var sum float64
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			if v := float64(d[f][s]); v > 0 {
				weights = append(weights, v)
				sum += v
			}
		}
	}
	H, ratio := 0.0, 0.0
	if sum > 0 && len(weights) > 1 {
		for _, v := range weights {
			p := v / sum
			H -= p * math.Log2(p)
		}
		ratio = H / math.Log2(float64(len(weights))) // 1 = energy spread evenly, 0 = concentrated
	}

	emit(modules.Output{
		Spell:   "entropy",
		Detail:  d,
		Summary: fmt.Sprintf("information %.2f bits across %d cells · spread %.2f (1=rich, 0=repetitive)", H, len(weights), ratio),
		Data:    map[string]any{"bits": H, "active_cells": len(weights), "spread": ratio},
	})
}

func emit(o modules.Output) { _ = json.NewEncoder(os.Stdout).Encode(o) }
func fail()                  { fmt.Fprintln(os.Stderr, "entropy: needs text"); os.Exit(1) }