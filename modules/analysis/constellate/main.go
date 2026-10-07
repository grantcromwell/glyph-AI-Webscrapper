// Spell: constellate (reasoning school) — find the structure in a SET of things.
//
// eigen finds the centre of a field; constellate finds its constellations. Given
// many items (lines or comma-separated), it hears each as a glyph and groups them
// by mutual resonance — two items join the same constellation when they resonate
// above the set's OWN median (the threshold is derived from the data, never
// hardcoded). The dog's way of seeing families emerge from a heap of fragments.
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
)

func main() {
	var in modules.Input
	_ = json.NewDecoder(os.Stdin).Decode(&in)
	items := itemsOf(in.Text)
	if len(items) < 2 {
		fmt.Fprintln(os.Stderr, "constellate: needs >=2 items (lines or commas)"); os.Exit(1)
	}
	gs := make([]encoding.Detail, len(items))
	for i, it := range items {
		gs[i] = encoding.Encode(it) // fine field: coarse saturates and blurs short items
	}
	// median pairwise resonance = the set's own kinship threshold
	var rs []float64
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			rs = append(rs, encoding.DetailResonance(gs[i], gs[j]))
		}
	}
	// link only UNUSUALLY strong pairs (above mean + 1σ of the set's own kinship),
	// so weak chains don't merge distinct constellations. Both derived from data.
	var mean float64
	for _, r := range rs {
		mean += r
	}
	mean /= float64(len(rs))
	var varc float64
	for _, r := range rs {
		varc += (r - mean) * (r - mean)
	}
	std := math.Sqrt(varc / float64(len(rs)))
	thr := mean + std

	// union-find: join items whose resonance exceeds the median
	parent := make([]int, len(items))
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
	for i := 0; i < len(items); i++ {
		for j := i + 1; j < len(items); j++ {
			if encoding.DetailResonance(gs[i], gs[j]) > thr {
				parent[find(i)] = find(j)
			}
		}
	}
	groups := map[int][]string{}
	combined := encoding.Detail{}
	for i, it := range items {
		groups[find(i)] = append(groups[find(i)], it)
		merge(&combined, encoding.Encode(it))
	}
	var constellations [][]string
	for _, g := range groups {
		constellations = append(constellations, g)
	}
	sort.Slice(constellations, func(a, b int) bool { return len(constellations[a]) > len(constellations[b]) })

	emit(modules.Output{
		Spell:   "constellate",
		Detail:  combined,
		Summary: fmt.Sprintf("%d items → %d constellations (kinship > %.2f)", len(items), len(constellations), thr),
		Data:    map[string]any{"constellations": constellations},
	})
}

func itemsOf(text string) []string {
	text = strings.ReplaceAll(text, ",", "\n")
	var out []string
	for _, l := range strings.Split(text, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}
func merge(dst *encoding.Detail, d encoding.Detail) {
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			if d[f][s] > dst[f][s] {
				dst[f][s] = d[f][s]
			}
		}
	}
}
func emit(o modules.Output) { _ = json.NewEncoder(os.Stdout).Encode(o) }