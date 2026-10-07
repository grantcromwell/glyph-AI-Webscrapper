// babel makes an SVG art-engram for every language the dog knows, so a recall
// vector space of tongues forms. Each language's glyph is heard from its real
// PHOIBLE phoneme inventory (the engram's Text IS those phonemes, so the glyph
// re-derives identically on recall). Each engram also carries, drawn into the
// art and embedded as cipher data, the tongues it is most AKIN to — the top
// glyph-resonant neighbours — so association is delineated and the connected
// vector space emerges in the index. One dense cryptographic art piece per tongue.
//
//	go run ./cmd/babel            # all languages
//	go run ./cmd/babel -limit 50  # a sample
package main

import (
	"encoding/csv"
	"flag"
	"fmt"
	"os"
	"sort"
	"strings"

	"glyphai/internal/encoding"
	"glyphai/internal/memory"
)

const (
	csvPath  = "modules/tongue/phoible/phoible.csv"
	babelDir = "gut/lore/babel"
	memRoot  = "memory"
	memBank  = "babel"
)

type lang struct {
	name     string
	phonemes []string
	seen     map[string]bool
	d        encoding.Detail // the tongue's full phonological field (1512 cells)
}

func main() {
	limit := flag.Int("limit", 0, "max languages (0 = all)")
	k := flag.Int("k", 6, "akin-to neighbours per tongue")
	flag.Parse()

	langs := load()
	if len(langs) == 0 {
		fmt.Fprintln(os.Stderr, "babel: no languages")
		os.Exit(1)
	}
	// hear each tongue from its phonemes (full field, so association is real even
	// when dense inventories saturate the coarse glyph)
	for _, l := range langs {
		l.d = encoding.Encode(l.name + " " + strings.Join(l.phonemes, " "))
	}
	order := make([]string, 0, len(langs))
	for name := range langs {
		order = append(order, name)
	}
	sort.Strings(order)
	if *limit > 0 && *limit < len(order) {
		order = order[:*limit]
	}
	fmt.Printf("🐾 babel: %d tongues heard; weaving the akin-to web (top-%d)…\n", len(order), *k)

	h, err := memory.Open(memRoot, memBank)
	if err != nil {
		fmt.Fprintln(os.Stderr, "babel:", err)
		os.Exit(1)
	}

	for i, name := range order {
		akin := neighbours(langs, name, order, *k)
		text := fmt.Sprintf("phonemes (%d): %s · akin to: %s",
			len(langs[name].phonemes), strings.Join(langs[name].phonemes, " "), strings.Join(akin, ", "))
		if _, err := h.RememberTo(babelDir, name, text, "phoible"); err != nil {
			fmt.Fprintf(os.Stderr, "  %s: %v\n", name, err)
			continue
		}
		if (i+1)%200 == 0 {
			fmt.Printf("   %d/%d tongues inscribed\n", i+1, len(order))
		}
	}
	cells, fams := h.Coverage()
	fmt.Printf("🐾 babel done: %d tongues → %s/ · vector space lights %d/1512 cells, %d/21 families\n",
		h.Count(), babelDir, cells, fams)
}

// load reads PHOIBLE and unions each language's phoneme inventory.
func load() map[string]*lang {
	f, err := os.Open(csvPath)
	if err != nil {
		fmt.Fprintln(os.Stderr, "babel:", err)
		os.Exit(1)
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	recs, err := r.ReadAll()
	if err != nil || len(recs) < 2 {
		fmt.Fprintln(os.Stderr, "babel: bad csv")
		os.Exit(1)
	}
	col := map[string]int{}
	for i, hdr := range recs[0] {
		col[hdr] = i
	}
	nameC, phC := col["LanguageName"], col["Phoneme"]
	out := map[string]*lang{}
	for _, row := range recs[1:] {
		name := strings.TrimSpace(row[nameC])
		ph := strings.TrimSpace(row[phC])
		if name == "" || ph == "" {
			continue
		}
		l := out[name]
		if l == nil {
			l = &lang{name: name, seen: map[string]bool{}}
			out[name] = l
		}
		if !l.seen[ph] {
			l.seen[ph] = true
			l.phonemes = append(l.phonemes, ph)
		}
	}
	return out
}

// neighbours returns the names of the top-k tongues most glyph-resonant with name.
func neighbours(langs map[string]*lang, name string, order []string, k int) []string {
	type sc struct {
		n string
		r float64
	}
	me := langs[name].d
	scores := make([]sc, 0, len(order))
	for _, other := range order {
		if other == name {
			continue
		}
		scores = append(scores, sc{other, encoding.DetailResonance(me, langs[other].d)})
	}
	sort.Slice(scores, func(a, b int) bool { return scores[a].r > scores[b].r })
	if k > len(scores) {
		k = len(scores)
	}
	out := make([]string, k)
	for i := 0; i < k; i++ {
		out[i] = scores[i].n
	}
	return out
}