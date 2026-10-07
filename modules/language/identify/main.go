// Spell: identify (tongue school) — name the tongue a text is written in.
//
// The dog roamed dozens of the world's Wikipedias but could not say WHICH tongue
// it was reading. It already knows ~2716 tongues as glyphs (the babel bank). This
// spell hears the given text, then asks the babel vector space which tongue it
// most resonates with — the dog naming a language by its sound-shape, not a
// lookup table.
package main

import (
	"encoding/json"
	"fmt"
	"os"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"glyphai/internal/memory"
)

func main() {
	var in modules.Input
	_ = json.NewDecoder(os.Stdin).Decode(&in)
	if in.Text == "" {
		fmt.Fprintln(os.Stderr, "identify: needs text"); os.Exit(1)
	}
	d := encoding.Encode(in.Text)
	h, err := memory.Open("memory", "babel")
	if err != nil {
		fmt.Fprintln(os.Stderr, "identify:", err); os.Exit(1)
	}
	hits := h.RecallDetail(d, 5)
	var tongues []string
	for _, hit := range hits {
		tongues = append(tongues, fmt.Sprintf("%s (%.0f%%)", hit.Topic, hit.Score*100))
	}
	best := "unknown"
	if len(hits) > 0 {
		best = hits[0].Topic
	}
	emit(modules.Output{
		Spell:   "identify",
		Detail:  d,
		Summary: fmt.Sprintf("sounds most like %s", best),
		Data:    map[string]any{"tongues": tongues},
	})
}

func emit(o modules.Output) { _ = json.NewEncoder(os.Stdout).Encode(o) }