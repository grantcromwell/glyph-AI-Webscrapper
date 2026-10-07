// Spell: analogy (reasoning school) — A is to B as C is to ?
//
// On the logogram plane a relation is a vector. The dog builds the target glyph
// B − A + C (the relation B-from-A carried onto C), then asks memory which known
// concept lands nearest that point. Relational inference straight from the glyph
// geometry — e.g. "puppy | dog | kitten" → cat.
//
//	echo '{"text":"puppy | dog | kitten"}' | go run ./grimoire/reasoning/analogy
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
	parts := strings.Split(in.Text, "|")
	if len(parts) != 3 {
		fmt.Fprintln(os.Stderr, "analogy: needs \"A | B | C\"  (A is to B as C is to ?)")
		os.Exit(1)
	}
	a, b, c := strings.TrimSpace(parts[0]), strings.TrimSpace(parts[1]), strings.TrimSpace(parts[2])
	da, db, dc := encoding.Encode(a), encoding.Encode(b), encoding.Encode(c)

	// target = C + (B − A), clamped into glyph levels: carry the relation onto C.
	var t encoding.Detail
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			v := int(dc[f][s]) + int(db[f][s]) - int(da[f][s])
			if v < 0 {
				v = 0
			} else if v > encoding.MaxLevel {
				v = encoding.MaxLevel
			}
			if v > 0 {
				t.Set(f, s, uint8(v))
			}
		}
	}

	h, err := memory.Open("memory", "records")
	if err != nil {
		fmt.Fprintln(os.Stderr, "analogy:", err)
		os.Exit(1)
	}
	answer := "(nothing lands there)"
	skip := strings.ToLower(a + " " + b + " " + c)
	for _, hit := range h.RecallDetail(t, 8) {
		if lt := strings.ToLower(hit.Topic); lt != "" && !strings.Contains(skip, lt) && !strings.Contains(lt, strings.ToLower(c)) {
			answer = hit.Topic
			break
		}
	}
	x, y := encoding.Locate(t)
	denote := encoding.ConceptLogogram(t) // the dog's own logogram for the in-between concept
	_ = json.NewEncoder(os.Stdout).Encode(modules.Output{
		Spell:   "analogy",
		Detail:  t,
		Summary: fmt.Sprintf("%s : %s :: %s : %s   (the dog denotes it %s)", a, b, c, answer, denote),
		Data:    map[string]any{"answer": answer, "x": x, "y": y, "denotes": denote},
	})
}