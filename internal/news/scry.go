// Package scry is the shared body of the dog's tongue-spells: read a set of
// native-language news feeds, bring back the headlines as docs the hunt can file
// and follow, and paint their combined glyph so the dog can feel — wordlessly —
// what a whole foreign news-front is humming about. Each tongue-spell is then a
// thin shell that just hands over its own feeds.
package news

import (
	"fmt"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"glyphai/internal/fetch"
	"glyphai/internal/feeds"
)

// News reads the given feeds, balanced across them, and returns an Output with
// docs (title+summary), links (to follow), and the combined glyph of the lot.
// `source` names the tongue (e.g. "russian"); feeds are its native sources.
func News(in modules.Input, source string, feedList []string) modules.Output {
	limit := 16
	if v, ok := in.Args["limit"].(float64); ok && v > 0 {
		limit = int(v)
	}
	perFeed := limit/len(feedList) + 1

	var docs []fetch.Doc
	combined := encoding.Detail{}
	for _, f := range feedList {
		items, err := feeds.Fetch(f)
		if err != nil {
			continue
		}
		n := 0
		for _, it := range items {
			text := it.Title
			if it.Text != "" {
				text += ". " + it.Text
			}
			docs = append(docs, fetch.Doc{Title: it.Title, URL: it.Link, Text: text, Source: source})
			merge(&combined, encoding.Encode(text))
			if n++; n >= perFeed {
				break
			}
		}
	}

	links := make([]map[string]any, 0, len(docs))
	for _, d := range docs {
		links = append(links, map[string]any{"text": d.Title, "url": d.URL})
	}
	return modules.Output{
		Spell:   source,
		Detail:  combined,
		Summary: fmt.Sprintf("%s: %d native headlines across %d feeds", source, len(docs), len(feedList)),
		Data:    map[string]any{"docs": docs, "links": links},
	}
}

func merge(dst *encoding.Detail, d encoding.Detail) {
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			if d[f][s] > dst[f][s] {
				dst[f][s] = d[f][s]
			}
		}
	}
}