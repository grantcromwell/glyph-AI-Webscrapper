// Spell: bridge (reasoning school) — how is A connected to B?
//
// Multi-hop relational reasoning over memory. Starting at A, the dog steps to the
// remembered concept that best carries it TOWARD B (greedy best-first by glyph
// resonance), hop by hop, until it arrives near B — tracing the chain of ideas
// that links the two. "thermodynamics | information" might bridge through entropy.
//
//	echo '{"text":"thermodynamics | information"}' | go run ./grimoire/reasoning/bridge
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"glyphai/internal/memory"
)

func main() {
	var in modules.Input
	_ = json.NewDecoder(os.Stdin).Decode(&in)
	parts := strings.SplitN(in.Text, "|", 2)
	if len(parts) != 2 {
		fmt.Fprintln(os.Stderr, "bridge: needs \"A | B\"")
		os.Exit(1)
	}
	a, b := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1])
	h, err := memory.Open("memory", "records")
	if err != nil {
		fmt.Fprintln(os.Stderr, "bridge:", err)
		os.Exit(1)
	}
	bg := encoding.EncodeGlyph(b)
	cur := encoding.Encode(a)
	chain := []string{a}
	visited := map[string]bool{strings.ToLower(a): true}
	best := 0.0
	for step := 0; step < 6; step++ {
		var pick *memory.Hit
		pickScore := -1.0
		for _, hit := range h.RecallDetail(cur, 10) {
			if hit.Topic == "" || visited[strings.ToLower(hit.Topic)] {
				continue
			}
			hg, _ := encoding.ParseHex(hit.Glyph)
			if toward := encoding.Resonance(hg, bg); toward > pickScore { // progress toward B
				pickScore, pick = toward, &hit
			}
		}
		if pick == nil {
			break
		}
		visited[strings.ToLower(pick.Topic)] = true
		chain = append(chain, pick.Topic)
		cur = encoding.Encode(pick.Topic + " " + pick.Text)
		if pickScore <= best { // no longer getting closer to B — arrived
			break
		}
		best = pickScore
	}
	chain = append(chain, b)
	_ = json.NewEncoder(os.Stdout).Encode(modules.Output{
		Spell:   "bridge",
		Detail:  encoding.Encode(strings.Join(chain, " ")),
		Summary: strings.Join(chain, "  →  "),
		Data:    map[string]any{"chain": chain},
	})
}