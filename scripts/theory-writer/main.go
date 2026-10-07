// scribe-theory — the dog writes a daily theory report from one hunt's bank.
//
// It reads a memory bank grown by a themed hunt (cmd/roam -bank NAME), reads what
// it caught along each of two domain "axes", and — the point of the report —
// finds the SEAM: the engrams that resonate with BOTH axes at once. That seam is
// the dog's own answer to "axis-a × axis-b": where, in its glyph field, the two
// domains actually touch.
//
// Ranking is by DISTINCTIVE detail resonance (encoding.Distinctive): the learned
// baseline — the shared "Voice hum" that boilerplate and web chrome are made of —
// is subtracted first, so a page is judged by what is SPECIAL about it, not by the
// generic navigation text every site shares. Everything here is the dog's own
// perception (Encode, Distinctive, DetailResonance, Reading, Logogram); no prose
// is invented, only retrieved and read.
//
//	go run ./scripts/scribe-theory -bank sacgeo \
//	   -title "Material Science × Sacred Geometry" -slug material-sacred-geometry \
//	   -axis-a "materials science crystal lattice" -axis-b "sacred geometry"
package main

import (
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

type scored struct {
	topic, text string
	det         encoding.Detail // raw perception of topic+text
	a, b, seam  float64      // distinctive resonance toward each axis, and the seam
}

func main() {
	bank := flag.String("bank", "", "the hunt's memory bank to read (required)")
	out := flag.String("out", "theory", "directory to write the daily report into")
	title := flag.String("title", "", "report title / theme (required)")
	slug := flag.String("slug", "", "filename prefix (defaults to the bank name)")
	axisA := flag.String("axis-a", "", "first domain probe (required)")
	axisB := flag.String("axis-b", "", "second domain probe (required)")
	top := flag.Int("top", 10, "how many catches to show per section")
	flag.Parse()
	if *bank == "" || *title == "" || *axisA == "" || *axisB == "" {
		fmt.Fprintln(os.Stderr, "scribe-theory: -bank, -title, -axis-a and -axis-b are required")
		os.Exit(2)
	}
	if *slug == "" {
		*slug = *bank
	}

	h, err := memory.Open(memRoot, *bank)
	must(err)

	// Axis fingerprints, baseline removed — what is SPECIAL about each domain.
	dA := encoding.Distinctive(encoding.Encode(*axisA))
	dB := encoding.Distinctive(encoding.Encode(*axisB))
	bind := encoding.DetailResonance(dA, dB)

	all := h.AllEngrams()
	items := make([]scored, 0, len(all))
	var acc [encoding.Families][encoding.Subfamilies]float64 // running sum for the mean field
	for _, e := range all {
		det := encoding.Encode(e.Topic + " " + e.Text)
		dd := encoding.Distinctive(det)
		for f := 0; f < encoding.Families; f++ {
			for s := 0; s < encoding.Subfamilies; s++ {
				acc[f][s] += float64(dd[f][s])
			}
		}
		rA := encoding.DetailResonance(dA, dd)
		rB := encoding.DetailResonance(dB, dd)
		seam := rA
		if rB < seam {
			seam = rB
		}
		items = append(items, scored{e.Topic, e.Text, det, rA, rB, seam})
	}
	// The expedition's centroid is the MEAN distinctive field (averaging keeps it
	// in 0..7; summing would saturate every family and read the same for any bank).
	var centroid encoding.Detail
	if n := float64(len(all)); n > 0 {
		for f := 0; f < encoding.Families; f++ {
			for s := 0; s < encoding.Subfamilies; s++ {
				centroid[f][s] = sat(int(acc[f][s]/n + 0.5))
			}
		}
	}

	date := time.Now().Format("2006-01-02")
	var b strings.Builder
	fmt.Fprintf(&b, "# glyphai — %s\n\n", *title)
	fmt.Fprintf(&b, "*Daily theory report · %s*\n", date)
	fmt.Fprintf(&b, "*Hunt bank: `%s` · %d engrams caught*\n\n---\n\n", *bank, len(all))
	fmt.Fprintf(&b, "## What this expedition's mind reads\n\n")
	fmt.Fprintf(&b, "Setting aside the shared hum of the web, the *distinctive* centroid of everything caught reads: **%s**.\n", encoding.Reading(centroid.Coarse()))
	fmt.Fprintf(&b, "Its logogram: `%s`\n\n", encoding.Logogram(centroid))
	fmt.Fprintf(&b, "The two domains themselves resonate at **%.3f** (distinctive) — that is how near *%s* and *%s* sit on the dog's field before it read a single page.\n\n",
		bind, *axisA, *axisB)

	// Each section is cut at the natural BREAK in its sorted scores (the largest
	// gap among the top candidates), not a fixed count: a drift-heavy walk shows
	// only the few real catches before the noise floor; a clean walk shows more.
	writeAxis(&b, "I. "+title2(*axisA), items, func(s scored) float64 { return s.a }, *top)
	writeAxis(&b, "II. "+title2(*axisB), items, func(s scored) float64 { return s.b }, *top)

	// The seam — the heart of the report.
	fmt.Fprintf(&b, "## III. The seam — where the two domains touch\n\n")
	fmt.Fprintf(&b, "*Engrams that pull on **both** axes at once, scored by the weaker of the two distinctive resonances so only true bridges rank. The list is cut where the scores break — below that gap is the walk's noise floor. This is the dog's own sense of the \"×\".*\n\n")
	sort.Slice(items, func(i, j int) bool { return items[i].seam > items[j].seam })
	uniq := dedupe(items, func(s scored) string { return norm(s.topic) })
	keep := kneeCut(uniq, func(s scored) float64 { return s.seam }, *top)
	seen := map[string]bool{}
	shown := 0
	for _, it := range uniq {
		if shown >= keep {
			break
		}
		key := norm(it.topic)
		if key == "" || seen[key] {
			continue
		}
		seen[key] = true
		fmt.Fprintf(&b, "### %s\n\n", trunc(it.topic, 80))
		fmt.Fprintf(&b, "> %s\n\n", trunc(collapse(it.text), 300))
		fmt.Fprintf(&b, "- **Reads:** %s\n", encoding.Reading(it.det.Coarse()))
		fmt.Fprintf(&b, "- **Logogram:** `%s`\n", encoding.Logogram(it.det))
		fmt.Fprintf(&b, "- **Pull → %s:** %.3f · **→ %s:** %.3f · **seam:** %.3f\n\n",
			firstWord(*axisA), it.a, firstWord(*axisB), it.b, it.seam)
		shown++
	}
	if shown == 0 {
		fmt.Fprintf(&b, "*Nothing this hunt caught resonated with both axes — the walk drifted off the seam.*\n\n")
	}
	fmt.Fprintf(&b, "---\n\n*Written by the dog from its `%s` bank. Readings, logograms and resonances are its own perception, ranked after subtracting the learned baseline; the quoted prose is retrieved, not generated.*\n", *bank)

	must(os.MkdirAll(*out, 0o755))
	path := filepath.Join(*out, *slug+"_"+date+".md")
	must(os.WriteFile(path, []byte(b.String()), 0o644))
	fmt.Printf("🐾 scribed theory report (%d seam catches, %d chars) → %s\n", shown, b.Len(), path)
}

// writeAxis lists the top engrams along one axis, ranked by distinctive
// resonance and cut at the natural break in the scores.
func writeAxis(b *strings.Builder, heading string, items []scored, key func(scored) float64, k int) {
	fmt.Fprintf(b, "## %s\n\n", heading)
	cp := make([]scored, len(items))
	copy(cp, items)
	sort.Slice(cp, func(i, j int) bool { return key(cp[i]) > key(cp[j]) })
	uniq := dedupe(cp, func(s scored) string { return norm(s.topic) })
	keep := kneeCut(uniq, key, k)
	if keep == 0 {
		fmt.Fprintf(b, "*No catches resonated with this axis.*\n\n")
		return
	}
	for i := 0; i < keep && i < len(uniq); i++ {
		it := uniq[i]
		fmt.Fprintf(b, "%d. **%s** — pull %.3f\n", i+1, trunc(it.topic, 78), key(it))
		fmt.Fprintf(b, "   > %s\n", trunc(collapse(it.text), 180))
	}
	fmt.Fprintln(b)
}

// dedupe keeps the first occurrence of each key (input must already be sorted
// best-first), so the strongest reading of a repeated topic is the one shown.
func dedupe(items []scored, key func(scored) string) []scored {
	seen := map[string]bool{}
	out := items[:0:0]
	for _, it := range items {
		k := key(it)
		if k == "" || seen[k] {
			continue
		}
		seen[k] = true
		out = append(out, it)
	}
	return out
}

// kneeCut finds where a descending score series breaks: the index of the largest
// drop between consecutive values within the first max entries. It returns how
// many to keep (up to and including the value before the break) — a data-derived
// "this is where signal becomes noise", with no tuned threshold.
func kneeCut(items []scored, key func(scored) float64, max int) int {
	n := len(items)
	if n > max {
		n = max
	}
	if n <= 1 {
		return n
	}
	cut, gap := n, -1.0
	for i := 0; i < n-1; i++ {
		if d := key(items[i]) - key(items[i+1]); d > gap {
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

func norm(s string) string     { return strings.ToLower(strings.TrimSpace(s)) }
func collapse(s string) string { return strings.Join(strings.Fields(s), " ") }
func title2(s string) string   { return strings.Title(s) }
func firstWord(s string) string {
	if f := strings.Fields(s); len(f) > 0 {
		return f[0]
	}
	return s
}
func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}
func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "scribe-theory:", err)
		os.Exit(1)
	}
}