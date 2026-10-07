// Spell: arxiv (temporal lobe, a "site-spell")
// The warlock's familiar for arXiv. Casts the query at arXiv's export API and
// returns paper titles + abstracts (dense, clean knowledge). The site lives
// here, not in the fetch.
package main

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"net/url"
	"os"
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
		fail(fmt.Errorf("arxiv spell needs a query (text)"))
	}
	limit := 3
	if v, ok := in.Args["limit"].(float64); ok {
		limit = int(v)
	}

	body, err := fetch.Get("http://export.arxiv.org/api/query?max_results=" + fmt.Sprint(limit) +
		"&search_query=all:" + url.QueryEscape(in.Text))
	if err != nil {
		fail(err)
	}
	var feed struct {
		Entries []struct {
			Title   string `xml:"title"`
			Summary string `xml:"summary"`
			ID      string `xml:"id"`
		} `xml:"entry"`
	}
	if err := xml.Unmarshal([]byte(body), &feed); err != nil {
		fail(err)
	}

	var docs []fetch.Doc
	combined := encoding.Detail{}
	for _, e := range feed.Entries {
		abstract := fetch.Clean(strings.TrimSpace(e.Summary))
		if len([]rune(abstract)) < 80 {
			continue
		}
		docs = append(docs, fetch.Doc{
			Title:  fetch.Clean(e.Title),
			URL:    strings.TrimSpace(e.ID),
			Text:   abstract,
			Source: "arxiv",
		})
		merge(&combined, encoding.Encode(abstract))
	}

	emit(modules.Output{
		Spell:   "arxiv",
		Detail:  combined,
		Summary: fmt.Sprintf("arxiv brought back %d abstracts for %q", len(docs), in.Text),
		Data:    map[string]any{"docs": docs},
	})
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
	fmt.Fprintln(os.Stderr, "arxiv:", err)
	os.Exit(1)
}