// fable5site turns a JSONL of model traces into a LOCAL, navigable website the
// dog can hunt through. Each trace becomes a page (the user's ask + the model's
// chain-of-thought). The pages are wired into a semantic web using the dog's OWN
// glyph engine: every page links to its most glyph-resonant neighbours, and the
// link text is the neighbour's gist — so when the dog prowls, its pull
// (ContrastResonance) reads true scent and follows meaning, not chrome.
//
// One-time prep, like walk.sh: generate → serve → hunt → delete the raw file.
//
//	go run ./cmd/fable5site -in gut/source/fable5-raw/fable5_cot_merged.jsonl -out /tmp/fable5-site -k 8
package main

import (
	"bufio"
	"encoding/json"
	"flag"
	"fmt"
	"html"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"glyphai/internal/encoding"
)

type record struct {
	UID        string `json:"uid"`
	Model      string `json:"model"`
	Context    string `json:"context"`
	Cot        string `json:"cot"`
	OutputType string `json:"output_type"`
}

type trace struct {
	idx   int
	uid   string
	title string // short gist (the user's ask) — becomes link text
	body  string // full readable text the dog reads
	g     encoding.Glyph
}

func main() {
	in := flag.String("in", "", "input jsonl")
	out := flag.String("out", "/tmp/fable5-site", "output site dir")
	k := flag.Int("k", 8, "neighbour links per page")
	seeds := flag.Int("seeds", 60, "how many entry links on the index page")
	flag.Parse()
	if *in == "" {
		fmt.Fprintln(os.Stderr, "need -in <jsonl>")
		os.Exit(1)
	}

	traces := load(*in)
	if len(traces) == 0 {
		fmt.Fprintln(os.Stderr, "no usable traces")
		os.Exit(1)
	}
	// Build the title lookup that neighbour links read as scent — must exist
	// before any page is written.
	titles = make([]string, len(traces))
	for i := range traces {
		titles[i] = traces[i].title
	}
	fmt.Printf("loaded %d traces; encoding glyphs…\n", len(traces))

	if err := os.MkdirAll(*out, 0o755); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}

	// Wire the semantic web: each page links to its top-K glyph-resonant peers.
	// Brute-force resonance over coarse glyphs — fast, and it is the dog's own
	// sense of kinship deciding the link structure.
	fmt.Printf("wiring neighbour graph (top-%d by glyph resonance)…\n", *k)
	for i := range traces {
		nb := neighbours(traces, i, *k)
		writePage(*out, traces[i], nb)
	}
	writeIndex(*out, traces, *seeds)

	fmt.Printf("site written to %s (%d pages + index.html)\n", *out, len(traces))
}

func load(path string) []trace {
	f, err := os.Open(path)
	if err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
	defer f.Close()
	sc := bufio.NewScanner(f)
	sc.Buffer(make([]byte, 0, 1<<20), 1<<24)
	var out []trace
	for sc.Scan() {
		line := sc.Bytes()
		if len(line) == 0 {
			continue
		}
		var r record
		if json.Unmarshal(line, &r) != nil {
			continue
		}
		body := strings.TrimSpace(r.Context + "\n\n" + r.Cot)
		if body == "" {
			continue
		}
		idx := len(out)
		out = append(out, trace{
			idx:   idx,
			uid:   r.UID,
			title: gist(r.Cot, r.Context),
			body:  body,
			g:     encoding.EncodeGlyph(clip(body, 6000)),
		})
	}
	return out
}

// gist picks a short, meaningful label for a trace — used as the link text other
// pages point at, so the dog's pull reads what the trace is about. It prefers the
// chain-of-thought (the distinctive reasoning) over the (often repeated/truncated)
// conversation context. Cleaning source artifacts here is ETL, not cognition.
func gist(primary, fallback string) string {
	src := strings.TrimSpace(primary)
	if src == "" {
		src = strings.TrimSpace(fallback)
	}
	src = strings.ReplaceAll(src, "…[earlier truncated]…", " ")
	src = strings.TrimPrefix(src, "USER:")
	src = strings.TrimSpace(src)
	if i := strings.IndexAny(src, ".\n"); i > 24 {
		src = src[:i]
	}
	return clip(strings.Join(strings.Fields(src), " "), 100)
}

func neighbours(ts []trace, i, k int) []int {
	type sc struct {
		j int
		r float64
	}
	scores := make([]sc, 0, len(ts))
	for j := range ts {
		if j == i {
			continue
		}
		scores = append(scores, sc{j, encoding.Resonance(ts[i].g, ts[j].g)})
	}
	sort.Slice(scores, func(a, b int) bool { return scores[a].r > scores[b].r })
	if k > len(scores) {
		k = len(scores)
	}
	out := make([]int, k)
	for n := 0; n < k; n++ {
		out[n] = scores[n].j
	}
	return out
}

func clip(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n])
}

func pageName(idx int) string { return fmt.Sprintf("t%05d.html", idx) }

func writePage(dir string, t trace, nb []int) {
	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\">")
	fmt.Fprintf(&b, "<title>%s</title></head><body>\n", html.EscapeString(t.title))
	fmt.Fprintf(&b, "<article><h1>%s</h1>\n", html.EscapeString(t.title))
	fmt.Fprintf(&b, "<pre>%s</pre>\n", html.EscapeString(t.body))
	b.WriteString("</article>\n<nav><h2>related traces</h2><ul>\n")
	for _, j := range nb {
		// link text = neighbour gist (the scent the dog's pull reads)
		fmt.Fprintf(&b, "<li><a href=\"%s\">%s</a></li>\n",
			pageName(j), html.EscapeString(neighTitle(j)))
	}
	b.WriteString("</ul></nav></body></html>\n")
	_ = os.WriteFile(filepath.Join(dir, pageName(t.idx)), []byte(b.String()), 0o644)
}

// neighTitle is filled in via the package-level index built in writeIndex's caller.
var titles []string

func neighTitle(j int) string {
	if j >= 0 && j < len(titles) {
		return titles[j]
	}
	return pageName(j)
}

func writeIndex(dir string, ts []trace, seeds int) {
	var b strings.Builder
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\">")
	b.WriteString("<title>Fable-5 traces</title></head><body><article><h1>Fable-5 traces</h1>")
	fmt.Fprintf(&b, "<p>%d traces. Entry points into the semantic web:</p></article>\n<nav><ul>\n", len(ts))
	// spread the entry links across the corpus so the dog can reach any region
	step := 1
	if seeds > 0 && len(ts) > seeds {
		step = len(ts) / seeds
	}
	for i := 0; i < len(ts); i += step {
		fmt.Fprintf(&b, "<li><a href=\"%s\">%s</a></li>\n",
			pageName(ts[i].idx), html.EscapeString(ts[i].title))
	}
	b.WriteString("</ul></nav></body></html>\n")
	_ = os.WriteFile(filepath.Join(dir, "index.html"), []byte(b.String()), 0o644)
}