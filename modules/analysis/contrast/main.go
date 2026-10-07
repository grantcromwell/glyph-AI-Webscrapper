// Spell: contrast (reasoning school) — compare two things the dog couldn't before.
//
// The dog could recall but not RELATE. Given "A | B", it hears both, measures how
// much they resonate, and names the families where they most AGREE (both strong)
// and where they most DIFFER. Pure glyph geometry; the dog's first comparative
// thought.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"sort"
	"strings"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
)

func main() {
	var in modules.Input
	_ = json.NewDecoder(os.Stdin).Decode(&in)
	a, b := split(in.Text)
	if a == "" || b == "" {
		fmt.Fprintln(os.Stderr, "contrast: needs \"A | B\""); os.Exit(1)
	}
	da, db := encoding.Encode(a), encoding.Encode(b)
	ga, gb := da.Coarse(), db.Coarse()
	res := encoding.DetailResonance(da, db)

	type fam struct {
		name        string
		agree, diff int
	}
	fams := make([]fam, encoding.Families)
	for f := 0; f < encoding.Families; f++ {
		la, lb := int(ga[f]), int(gb[f])
		fams[f] = fam{encoding.FamilyNames[f] + " " + encoding.FamilyRole(f), min(la, lb), abs(la - lb)}
	}
	shared := append([]fam(nil), fams...)
	sort.Slice(shared, func(i, j int) bool { return shared[i].agree > shared[j].agree })
	differ := append([]fam(nil), fams...)
	sort.Slice(differ, func(i, j int) bool { return differ[i].diff > differ[j].diff })

	emit(modules.Output{
		Spell:   "contrast",
		Detail:  merge(da, db),
		Summary: fmt.Sprintf("%.0f%% kin · agree on %s · differ on %s", res*100, shared[0].name, differ[0].name),
		Data: map[string]any{
			"resonance":   res,
			"shared":      []string{shared[0].name, shared[1].name, shared[2].name},
			"divergent":   []string{differ[0].name, differ[1].name, differ[2].name},
		},
	})
}

func split(s string) (string, string) {
	for _, sep := range []string{"|", " vs ", " versus "} {
		if i := strings.Index(s, sep); i >= 0 {
			return strings.TrimSpace(s[:i]), strings.TrimSpace(s[i+len(sep):])
		}
	}
	return "", ""
}
func merge(a, b encoding.Detail) encoding.Detail {
	var d encoding.Detail
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			if a[f][s] > b[f][s] {
				d[f][s] = a[f][s]
			} else {
				d[f][s] = b[f][s]
			}
		}
	}
	return d
}
func min(a, b int) int { if a < b { return a }; return b }
func abs(a int) int    { if a < 0 { return -a }; return a }
func emit(o modules.Output) { _ = json.NewEncoder(os.Stdout).Encode(o) }