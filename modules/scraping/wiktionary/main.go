// Spell: wiktionary (temporal lobe, a "site-spell")
// The warlock's familiar for etymology. Given a query, it picks the most
// meaning-bearing words and pulls their real lineage (PIE -> Latin/Greek ->
// English) from Wiktionary's official API — the depth the temporal lobe can't
// derive on its own. Returns each etymology as a doc for the nose to file.
package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"sort"
	"strings"

	"glyphai/internal/modules"
	"glyphai/internal/encoding"
	"glyphai/internal/fetch"
)

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(err)
	}
	if in.Text == "" {
		fail(fmt.Errorf("wiktionary needs a query"))
	}

	var docs []fetch.Doc
	combined := encoding.Detail{}
	for _, w := range salientWords(in.Text, 2) {
		etym, err := etymology(w)
		if err != nil || len([]rune(etym)) < 30 {
			continue
		}
		docs = append(docs, fetch.Doc{
			Title:  "etymology of " + w,
			URL:    "https://en.wiktionary.org/wiki/" + url.PathEscape(w),
			Text:   etym,
			Source: "wiktionary",
		})
		merge(&combined, encoding.Encode(etym))
	}

	emit(modules.Output{
		Spell:   "wiktionary",
		Detail:  combined,
		Summary: fmt.Sprintf("traced %d etymologies for %q", len(docs), in.Text),
		Data:    map[string]any{"docs": docs},
	})
}

var sectionRe = regexp.MustCompile(`(?m)^(Pronunciation|Noun|Verb|Adjective|Adverb|Proper noun|Interjection|Determiner|References|Anagrams|Derived terms|Related terms|Translations|Usage notes|Synonyms|Antonyms|Etymology \d)`)

// etymology fetches a word's Etymology section as plain text from Wiktionary.
func etymology(word string) (string, error) {
	body, err := fetch.Get("https://en.wiktionary.org/w/api.php?action=query&prop=extracts" +
		"&explaintext=1&redirects=1&format=json&titles=" + url.QueryEscape(word))
	if err != nil {
		return "", err
	}
	var r struct {
		Query struct {
			Pages map[string]struct {
				Extract string `json:"extract"`
			} `json:"pages"`
		} `json:"query"`
	}
	if err := json.Unmarshal([]byte(body), &r); err != nil {
		return "", err
	}
	for _, p := range r.Query.Pages {
		ex := p.Extract
		i := strings.Index(ex, "Etymology")
		if i < 0 {
			return "", fmt.Errorf("no etymology")
		}
		seg := ex[i+len("Etymology"):]
		seg = strings.TrimLeft(seg, " 0123456789\n")
		if loc := sectionRe.FindStringIndex(seg); loc != nil {
			seg = seg[:loc[0]]
		}
		seg = fetch.Clean(seg)
		if len([]rune(seg)) > 600 {
			seg = string([]rune(seg)[:600])
		}
		return seg, nil
	}
	return "", fmt.Errorf("no page")
}

// salientWords picks the n longest meaning-bearing words of a query.
func salientWords(text string, n int) []string {
	seen := map[string]bool{}
	var ws []string
	for _, w := range strings.Fields(strings.ToLower(text)) {
		w = strings.Trim(w, ".,!?;:\"'()-")
		if len(w) > 4 && !seen[w] {
			seen[w] = true
			ws = append(ws, w)
		}
	}
	sort.SliceStable(ws, func(i, j int) bool { return len(ws[i]) > len(ws[j]) })
	if len(ws) > n {
		ws = ws[:n]
	}
	return ws
}

func merge(dst *encoding.Detail, src encoding.Detail) {
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			dst.Set(f, s, src[f][s])
		}
	}
}

func emit(o modules.Output) {
	if err := json.NewEncoder(os.Stdout).Encode(o); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "wiktionary:", err)
	os.Exit(1)
}