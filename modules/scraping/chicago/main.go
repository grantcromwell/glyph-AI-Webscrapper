// Spell: chicago (scrying school, a site-spell)
// The witch's familiar for the Art Institute of Chicago. Casts a theme at the
// AIC public API and brings back artworks — title, artist, date, medium,
// classification, subject terms — for the nose to file as logograms. Each
// painting becomes a glyph the dog can be drawn toward.
package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strings"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"glyphai/internal/fetch"
)

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(err)
	}
	if in.Text == "" {
		fail(fmt.Errorf("chicago needs a theme (text)"))
	}
	limit := 6
	if v, ok := in.Args["limit"].(float64); ok {
		limit = int(v)
	}

	fields := "id,title,artist_display,date_display,medium_display,classification_title,term_titles,thumbnail"
	body, err := fetch.Get("https://api.artic.edu/api/v1/artworks/search?limit=" + fmt.Sprint(limit) +
		"&fields=" + url.QueryEscape(fields) + "&q=" + url.QueryEscape(in.Text))
	if err != nil {
		fail(err)
	}
	var resp struct {
		Data []struct {
			ID            int      `json:"id"`
			Title         string   `json:"title"`
			Artist        string   `json:"artist_display"`
			Date          string   `json:"date_display"`
			Medium        string   `json:"medium_display"`
			Class         string   `json:"classification_title"`
			Terms         []string `json:"term_titles"`
			Thumbnail     struct {
				Alt string `json:"alt_text"`
			} `json:"thumbnail"`
		} `json:"data"`
	}
	if err := json.Unmarshal([]byte(body), &resp); err != nil {
		fail(err)
	}

	var docs []fetch.Doc
	combined := encoding.Detail{}
	for _, a := range resp.Data {
		if a.Title == "" {
			continue
		}
		parts := []string{a.Artist, a.Date, a.Medium, a.Class}
		parts = append(parts, a.Terms...)
		if a.Thumbnail.Alt != "" {
			parts = append(parts, a.Thumbnail.Alt)
		}
		text := fetch.Clean(strings.Join(filterEmpty(parts), ". "))
		if len([]rune(text)) < 20 {
			text = a.Title
		}
		docs = append(docs, fetch.Doc{
			Title:  a.Title,
			URL:    fmt.Sprintf("https://www.artic.edu/artworks/%d", a.ID),
			Text:   text,
			Source: "chicago",
		})
		merge(&combined, encoding.Encode(a.Title+" "+text))
	}

	emit(modules.Output{
		Spell:   "chicago",
		Detail:  combined,
		Summary: fmt.Sprintf("the Art Institute offered %d works for %q", len(docs), in.Text),
		Data:    map[string]any{"docs": docs},
	})
}

func filterEmpty(s []string) []string {
	var out []string
	for _, x := range s {
		if strings.TrimSpace(x) != "" {
			out = append(out, x)
		}
	}
	return out
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
	fmt.Fprintln(os.Stderr, "chicago:", err)
	os.Exit(1)
}