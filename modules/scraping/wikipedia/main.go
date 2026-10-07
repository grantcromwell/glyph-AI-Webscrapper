// Spell: wikipedia (temporal lobe, a "site-spell")
// A warlock's familiar for Wikipedia. The hunt casts it with a query; it walks
// in through Wikipedia's official front door (opensearch + plain-text extracts,
// no scraping, no blocking) and returns clean docs for the nose to file as
// logograms. The site lives HERE, never in the fetch.
package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"

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
		fail(fmt.Errorf("wikipedia spell needs a query (text)"))
	}
	limit := 3
	if v, ok := in.Args["limit"].(float64); ok {
		limit = int(v)
	}

	body, err := fetch.Get("https://en.wikipedia.org/w/api.php?action=opensearch&namespace=0&format=json&limit=" +
		fmt.Sprint(limit) + "&search=" + url.QueryEscape(in.Text))
	if err != nil {
		fail(err)
	}
	var os4 []json.RawMessage
	if err := json.Unmarshal([]byte(body), &os4); err != nil || len(os4) < 4 {
		fail(fmt.Errorf("bad opensearch response"))
	}
	var titles, urls []string
	_ = json.Unmarshal(os4[1], &titles)
	_ = json.Unmarshal(os4[3], &urls)

	var docs []fetch.Doc
	combined := encoding.Detail{}
	for i, t := range titles {
		text, err := extract(t)
		if err != nil || len([]rune(text)) < 120 {
			continue
		}
		u := ""
		if i < len(urls) {
			u = urls[i]
		}
		docs = append(docs, fetch.Doc{Title: t, URL: u, Text: text, Source: "wikipedia"})
		merge(&combined, encoding.Encode(text))
	}

	emit(modules.Output{
		Spell:   "wikipedia",
		Detail:  combined,
		Summary: fmt.Sprintf("wikipedia brought back %d articles for %q", len(docs), in.Text),
		Data:    map[string]any{"docs": docs},
	})
}

func extract(title string) (string, error) {
	body, err := fetch.Get("https://en.wikipedia.org/w/api.php?action=query&prop=extracts&explaintext=1" +
		"&redirects=1&format=json&titles=" + url.QueryEscape(title))
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
		return fetch.Clean(p.Extract), nil
	}
	return "", fmt.Errorf("no extract")
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
	fmt.Fprintln(os.Stderr, "wikipedia:", err)
	os.Exit(1)
}