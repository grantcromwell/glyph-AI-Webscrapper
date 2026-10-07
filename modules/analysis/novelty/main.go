// Spell: novelty (reasoning school) — how NEW is this to the dog?
//
// The dog couldn't tell fresh knowledge from what it already held, so it re-filed
// redundant stubs and wandered in circles. novelty hears a text and asks its
// memory how strongly anything already resonates with it: novelty = 1 − the
// strongest familiarity. Near 1 = unknown ground worth hunting; near 0 = it has
// this already. The dog's sense of its own edge of knowing.
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
		fmt.Fprintln(os.Stderr, "novelty: needs text"); os.Exit(1)
	}
	d := encoding.Encode(in.Text)
	h, err := memory.Open("memory", "records")
	if err != nil {
		fmt.Fprintln(os.Stderr, "novelty:", err); os.Exit(1)
	}
	hits := h.RecallDetail(d, 3)
	familiarity, nearest := 0.0, "(nothing)"
	if len(hits) > 0 {
		familiarity, nearest = hits[0].Score, hits[0].Topic
	}
	novelty := 1 - familiarity
	emit(modules.Output{
		Spell:   "novelty",
		Detail:  d,
		Summary: fmt.Sprintf("novelty %.2f (1=new) · nearest known: %s", novelty, nearest),
		Data:    map[string]any{"novelty": novelty, "familiarity": familiarity, "nearest": nearest},
	})
}

func emit(o modules.Output) { _ = json.NewEncoder(os.Stdout).Encode(o) }