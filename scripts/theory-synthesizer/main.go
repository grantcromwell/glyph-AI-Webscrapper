// synthesize-theory — the dog reads two finished expeditions and writes one
// synthesis ACROSS them. Where scribe-theory reports a single hunt, this asks the
// harder question: do two different domains touch? It encodes every engram in each
// bank to its distinctive glyph (baseline removed), builds each domain's centroid,
// measures how near the two domains sit on the dog's field, and surfaces the
// BRIDGES — the pages in one domain that most resonate with the *other* domain's
// whole character. Those bridges are the dog's own answer to "what crosses over".
//
// It also folds in the OLD and NEW per-domain reports: it lists each domain's seam
// topics from both, so the synthesis carries both datasets and shows what the
// improved walk surfaced that the old one missed.
//
//	go run ./scripts/synthesize-theory \
//	  -bank-a sacgeo -title-a "Material Science × Sacred Geometry" \
//	  -bank-b asi    -title-b "Agentic Super Intelligence" \
//	  -old-a theory/material-sacred-geometry_2026-06-21_v1-prefix.md \
//	  -new-a theory/material-sacred-geometry_2026-06-21.md \
//	  -old-b theory/agentic-superintelligence_2026-06-21_v1-prefix.md \
//	  -new-b theory/agentic-superintelligence_2026-06-21.md
package main

import (
	"bufio"
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"glyphai/internal/encoding"
	"glyphai/internal/memory"
)

const memRoot = "memory"

type domain struct {
	bank, title string
	centroid    encoding.Detail
	engrams     []rated
	n           int
}

type rated struct {
	topic, text string
	dd          encoding.Detail // distinctive perception
	cross       float64      // resonance to the OTHER domain's centroid
}

func main() {
	bankA := flag.String("bank-a", "", "first domain's hunt bank (required)")
	bankB := flag.String("bank-b", "", "second domain's hunt bank (required)")
	titleA := flag.String("title-a", "", "first domain title (required)")
	titleB := flag.String("title-b", "", "second domain title (required)")
	oldA := flag.String("old-a", "", "first domain's pre-fix report markdown")
	newA := flag.String("new-a", "", "first domain's new report markdown")
	oldB := flag.String("old-b", "", "second domain's pre-fix report markdown")
	newB := flag.String("new-b", "", "second domain's new report markdown")
	out := flag.String("out", "theory", "directory to write the synthesis into")
	slug := flag.String("slug", "synthesis", "filename prefix")
	top := flag.Int("top", 8, "max bridges to show per direction")
	flag.Parse()
	if *bankA == "" || *bankB == "" || *titleA == "" || *titleB == "" {
		fmt.Fprintln(os.Stderr, "synthesize-theory: -bank-a/-bank-b/-title-a/-title-b are required")
		os.Exit(2)
	}

	a := loadDomain(*bankA, *titleA)
	b := loadDomain(*bankB, *titleB)
	// Bridges: how much each page resembles the OTHER domain's whole character.
	for i := range a.engrams {
		a.engrams[i].cross = encoding.DetailResonance(a.engrams[i].dd, b.centroid)
	}
	for i := range b.engrams {
		b.engrams[i].cross = encoding.DetailResonance(b.engrams[i].dd, a.centroid)
	}
	bind := encoding.DetailResonance(a.centroid, b.centroid)

	date := time.Now().Format("2006-01-02")
	var w strings.Builder
	fmt.Fprintf(&w, "# glyphai — Synthesis across two domains\n\n")
	fmt.Fprintf(&w, "*%s*\n\n", date)
	fmt.Fprintf(&w, "**A · %s** (`%s`, %d engrams)  ×  **B · %s** (`%s`, %d engrams)\n\n---\n\n",
		a.title, a.bank, a.n, b.title, b.bank, b.n)

	fmt.Fprintf(&w, "## How near the two domains sit\n\n")
	fmt.Fprintf(&w, "Distinctive resonance between the two expeditions' centroids: **%.3f**.\n", bind)
	fmt.Fprintf(&w, "- **A reads:** %s\n", encoding.Reading(a.centroid.Coarse()))
	fmt.Fprintf(&w, "- **B reads:** %s\n\n", encoding.Reading(b.centroid.Coarse()))

	fmt.Fprintf(&w, "## Bridges — %s seen through %s\n\n", a.title, b.title)
	fmt.Fprintf(&w, "*Pages the dog caught in A that most resonate with the whole of B — where the first domain reaches toward the second. Cut at the natural break.*\n\n")
	writeBridges(&w, a.engrams, *top)

	fmt.Fprintf(&w, "## Bridges — %s seen through %s\n\n", b.title, a.title)
	fmt.Fprintf(&w, "*And the reverse: pages in B that reach toward A.*\n\n")
	writeBridges(&w, b.engrams, *top)

	// Both datasets: old vs new seam topics, straight from the reports.
	fmt.Fprintf(&w, "## Both datasets — old walk vs improved walk\n\n")
	writeDelta(&w, a.title, *oldA, *newA)
	writeDelta(&w, b.title, *oldB, *newB)

	fmt.Fprintf(&w, "---\n\n*Synthesis is the dog's own perception: distinctive glyphs, cross-domain resonance, and the seam topics its two walks surfaced. The old (pre-driftfix) and improved walks are both represented above.*\n")

	must(os.MkdirAll(*out, 0o755))
	path := filepath.Join(*out, *slug+"_"+date+".md")
	must(os.WriteFile(path, []byte(w.String()), 0o644))
	fmt.Printf("🐾 synthesized %s × %s (bind %.3f) → %s\n", a.bank, b.bank, bind, path)
}

func loadDomain(bank, title string) domain {
	h, err := memory.Open(memRoot, bank)
	must(err)
	all := h.AllEngrams()
	d := domain{bank: bank, title: title, n: len(all)}
	var acc [encoding.Families][encoding.Subfamilies]float64
	for _, e := range all {
		dd := encoding.Distinctive(encoding.Encode(e.Topic + " " + e.Text))
		for f := 0; f < encoding.Families; f++ {
			for s := 0; s < encoding.Subfamilies; s++ {
				acc[f][s] += float64(dd[f][s])
			}
		}
		d.engrams = append(d.engrams, rated{topic: e.Topic, text: e.Text, dd: dd})
	}
	if n := float64(len(all)); n > 0 {
		for f := 0; f < encoding.Families; f++ {
			for s := 0; s < encoding.Subfamilies; s++ {
				d.centroid[f][s] = sat(int(acc[f][s]/n + 0.5))
			}
		}
	}
	return d
}

func writeBridges(w *strings.Builder, items []rated, top int) {
	sort.Slice(items, func(i, j int) bool { return items[i].cross > items[j].cross })
	uniq := items[:0:0]
	seen := map[string]bool{}
	for _, it := range items {
		k := strings.ToLower(strings.TrimSpace(it.topic))
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		uniq = append(uniq, it)
	}
	keep := kneeCut(uniq, top)
	if keep == 0 {
		fmt.Fprintf(w, "*No clear bridge — the domains stayed apart.*\n\n")
		return
	}
	for i := 0; i < keep && i < len(uniq); i++ {
		it := uniq[i]
		fmt.Fprintf(w, "### %s — reach %.3f\n\n", trunc(it.topic, 80), it.cross)
		fmt.Fprintf(w, "> %s\n\n", trunc(collapse(it.text), 240))
		fmt.Fprintf(w, "- **Reads:** %s\n- **Logogram:** `%s`\n\n", encoding.Reading(it.dd.Coarse()), encoding.Logogram(it.dd))
	}
}

// writeDelta lists each report's seam topics (### headings) so both datasets show.
func writeDelta(w *strings.Builder, title, oldPath, newPath string) {
	fmt.Fprintf(w, "### %s\n\n", title)
	oldT := seamTopics(oldPath)
	newT := seamTopics(newPath)
	if len(oldT) == 0 && len(newT) == 0 {
		fmt.Fprintf(w, "*(no reports supplied)*\n\n")
		return
	}
	oldSet := map[string]bool{}
	for _, t := range oldT {
		oldSet[strings.ToLower(t)] = true
	}
	fmt.Fprintf(w, "**Old walk seam:** %s\n\n", joinOr(oldT, "—"))
	fmt.Fprintf(w, "**Improved walk seam:** %s\n\n", joinOr(newT, "—"))
	var fresh []string
	for _, t := range newT {
		if !oldSet[strings.ToLower(t)] {
			fresh = append(fresh, t)
		}
	}
	if len(fresh) > 0 {
		fmt.Fprintf(w, "**New this walk:** %s\n\n", strings.Join(fresh, "; "))
	}
}

// seamTopics pulls "### " headings (the seam catches) from a scribe-theory report.
func seamTopics(path string) []string {
	if path == "" {
		return nil
	}
	f, err := os.Open(path)
	if err != nil {
		return nil
	}
	defer f.Close()
	var out []string
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 1024*1024), 1024*1024)
	for sc.Scan() {
		line := sc.Text()
		if strings.HasPrefix(line, "### ") {
			t := strings.TrimSpace(strings.TrimPrefix(line, "### "))
			// drop any trailing " — reach 0.xxx" the synthesis itself would add
			if i := strings.Index(t, " — "); i > 0 {
				t = t[:i]
			}
			out = append(out, t)
		}
	}
	return out
}

func joinOr(s []string, empty string) string {
	if len(s) == 0 {
		return empty
	}
	return strings.Join(s, "; ")
}

func kneeCut(items []rated, max int) int {
	n := len(items)
	if n > max {
		n = max
	}
	if n <= 1 {
		return n
	}
	cut, gap := n, -1.0
	for i := 0; i < n-1; i++ {
		if d := items[i].cross - items[i+1].cross; d > gap {
			gap, cut = d, i+1
		}
	}
	return cut
}

func sat(v int) uint8 {
	if v > encoding.MaxLevel {
		return uint8(encoding.MaxLevel)
	}
	if v < 0 {
		return 0
	}
	return uint8(v)
}

func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }
func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "synthesize-theory:", err)
		os.Exit(1)
	}
}