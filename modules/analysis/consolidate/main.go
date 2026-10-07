// Spell: consolidate (reasoning school) — the dog reorganises its own memory.
//
// Like an animal restructuring its recall during sleep, the dog lets memories
// MERGE instead of deleting them. Each engram reaches for its nearest kin and
// fuses with it with probability equal to their resonance — so alike memories
// (near-duplicate chrome, the same idea seen in many tongues) collapse into one
// schema, while distinct memories stand. The number that fuse into one is never
// fixed; it emerges from the probability. Only the mechanism is hardcoded.
//
//	echo '{}'                 | go run ./grimoire/reasoning/consolidate   # engrams bank
//	echo '{"text":"babel"}'   | go run ./grimoire/reasoning/consolidate   # a named bank
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"glyphai/internal/modules"
	"glyphai/internal/memory"
)

func main() {
	var in modules.Input
	_ = json.NewDecoder(os.Stdin).Decode(&in)
	bank := strings.TrimSpace(in.Text)
	if bank == "" {
		bank = "records"
	}
	h, err := memory.Open("memory", bank)
	if err != nil {
		fmt.Fprintln(os.Stderr, "consolidate:", err)
		os.Exit(1)
	}
	before, schemas, after, err := h.Consolidate("data/lore")
	if err != nil {
		fmt.Fprintln(os.Stderr, "consolidate:", err)
		os.Exit(1)
	}
	_ = json.NewEncoder(os.Stdout).Encode(modules.Output{
		Spell:   "consolidate",
		Summary: fmt.Sprintf("slept on %d memories → fused into %d schemas · bank %d → %d", before, schemas, before, after),
		Data:    map[string]any{"before": before, "schemas": schemas, "after": after},
	})
}