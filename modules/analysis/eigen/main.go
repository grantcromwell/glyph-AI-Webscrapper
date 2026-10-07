// Spell: eigen (reasoning school)
// A deterministic, exact analysis — like wavelet or fft, but over meaning.
// Given a set of topics, it builds their glyph-resonance graph and finds the
// most central by eigenvector centrality (power iteration). The dog casts this
// to discover where the centre of a field lies — e.g. which AGI topic to start
// a walk-forward training prowl from. The seed is emergent, never hand-picked.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"glyphai/internal/math"
	"glyphai/internal/encoding"
	"glyphai/internal/modules"
)

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(err)
	}
	topics := topicsFrom(in)
	if len(topics) < 2 {
		fail(fmt.Errorf("eigen needs >=2 topics (text lines/commas or a path)"))
	}

	// resonance graph
	n := len(topics)
	gs := make([]encoding.Glyph, n)
	for i, t := range topics {
		gs[i] = encoding.EncodeGlyph(t)
	}
	A := math.NewMat(n, n)
	for i := 0; i < n; i++ {
		for j := 0; j < n; j++ {
			if i != j {
				A.Set(i, j, encoding.Resonance(gs[i], gs[j]))
			}
		}
	}
	// power iteration -> dominant eigenvector (centrality)
	v := make(math.Vec, n)
	for i := range v {
		v[i] = 1
	}
	for iter := 0; iter < 60; iter++ {
		v = A.MatVec(v)
		if nrm := math.Norm(v); nrm > 0 {
			v = math.Scale(v, 1/nrm)
		}
	}
	order := make([]int, n)
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool { return v[order[a]] > v[order[b]] })

	type ranked struct {
		Topic string  `json:"topic"`
		Score float64 `json:"score"`
	}
	out := make([]ranked, n)
	combined := encoding.Detail{}
	for i, idx := range order {
		out[i] = ranked{topics[idx], v[idx]}
	}
	for _, t := range topics {
		d := encoding.Encode(t)
		for f := 0; f < encoding.Families; f++ {
			for s := 0; s < encoding.Subfamilies; s++ {
				combined.Set(f, s, d[f][s])
			}
		}
	}

	emit(modules.Output{
		Spell:   "eigen",
		Detail:  combined,
		Summary: fmt.Sprintf("centre of %d topics: %s", n, out[0].Topic),
		Data:    map[string]any{"ranked": out},
	})
}

func topicsFrom(in modules.Input) []string {
	text := in.Text
	if in.Path != "" {
		if raw, err := os.ReadFile(in.Path); err == nil {
			text = string(raw)
		}
	}
	text = strings.ReplaceAll(text, ",", "\n")
	var out []string
	for _, l := range strings.Split(text, "\n") {
		if l = strings.TrimSpace(l); l != "" {
			out = append(out, l)
		}
	}
	return out
}

func emit(o modules.Output) {
	if err := json.NewEncoder(os.Stdout).Encode(o); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "eigen:", err)
	os.Exit(1)
}