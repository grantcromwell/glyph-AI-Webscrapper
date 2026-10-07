// Spell: facet (reasoning school) — separate the SENSES of a word.
//
// The dog drifted because "intelligence" pulled one heap of memories mixing the
// spy sense and the mind sense, and the heaviest won. facet recalls what a query
// stirs, then splits those memories into facets: clusters that resonate with each
// other above the recalled set's OWN median, but not across. So the dog can SEE
// that "intelligence" has an espionage facet and a cognition facet, instead of
// blurring them. Disambiguation by resonance, nothing hardcoded.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"glyphai/internal/memory"
)

func main() {
	var in modules.Input
	_ = json.NewDecoder(os.Stdin).Decode(&in)
	if in.Text == "" {
		fmt.Fprintln(os.Stderr, "facet: needs a query"); os.Exit(1)
	}
	h, err := memory.Open("memory", "records")
	if err != nil {
		fmt.Fprintln(os.Stderr, "facet:", err); os.Exit(1)
	}
	q := encoding.Encode(in.Text)
	hits := h.RecallDetail(q, 14)
	if len(hits) < 2 {
		emit(modules.Output{Spell: "facet", Detail: q, Summary: "too few memories to facet"})
		return
	}
	gs := make([]encoding.Glyph, len(hits))
	for i, hit := range hits {
		gs[i], _ = encoding.ParseHex(hit.Glyph)
	}
	// the recalled set's own median kinship is the split line
	var rs []float64
	for i := 0; i < len(hits); i++ {
		for j := i + 1; j < len(hits); j++ {
			rs = append(rs, encoding.Resonance(gs[i], gs[j]))
		}
	}
	sort.Float64s(rs)
	thr := rs[len(rs)/2]

	parent := make([]int, len(hits))
	for i := range parent {
		parent[i] = i
	}
	var find func(int) int
	find = func(x int) int {
		for parent[x] != x {
			parent[x] = parent[parent[x]]
			x = parent[x]
		}
		return x
	}
	for i := 0; i < len(hits); i++ {
		for j := i + 1; j < len(hits); j++ {
			if encoding.Resonance(gs[i], gs[j]) > thr {
				parent[find(i)] = find(j)
			}
		}
	}
	facets := map[int][]string{}
	for i, hit := range hits {
		facets[find(i)] = append(facets[find(i)], hit.Topic)
	}
	var out [][]string
	for _, f := range facets {
		out = append(out, f)
	}
	sort.Slice(out, func(a, b int) bool { return len(out[a]) > len(out[b]) })

	emit(modules.Output{
		Spell:   "facet",
		Detail:  q,
		Summary: fmt.Sprintf("%q has %d facets in memory", in.Text, len(out)),
		Data:    map[string]any{"facets": out},
	})
}

func emit(o modules.Output) { _ = json.NewEncoder(os.Stdout).Encode(o) }