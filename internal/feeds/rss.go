// Package rss is a generic, site-agnostic feed reader — the dog's way of hearing
// the world through its front door (an RSS/Atom feed is the polite, blockless,
// key-free way in). It knows NO sites; WHICH feeds to read is decided by the
// tongue-spells in the grimoire that supply feed URLs. Handles both RSS <item>
// and Atom <entry>, CDATA, and inline HTML in summaries.
package feeds

import (
	"context"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"

	"glyphai/internal/fetch"
)

// browserUA: feeds answer an ordinary browser; the dog knocks as one.
const browserUA = "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0"

// Item is one entry in a feed.
type Item struct {
	Title string
	Link  string
	Text  string
}

var (
	blockRe    = regexp.MustCompile(`(?is)<(?:item|entry)\b[^>]*>(.*?)</(?:item|entry)>`)
	titleRe    = regexp.MustCompile(`(?is)<title\b[^>]*>(.*?)</title>`)
	descRe     = regexp.MustCompile(`(?is)<(?:description|summary|content)\b[^>]*>(.*?)</(?:description|summary|content)>`)
	rssLinkRe  = regexp.MustCompile(`(?is)<link\b[^>]*>([^<]+)</link>`)
	atomLinkRe = regexp.MustCompile(`(?is)<link\b[^>]*href="([^"]+)"`)
	cdataRe    = regexp.MustCompile(`(?is)<!\[CDATA\[(.*?)\]\]>`)
	tagRe      = regexp.MustCompile(`(?s)<[^>]+>`)
)

func clean(s string) string {
	s = cdataRe.ReplaceAllString(s, "$1")
	s = tagRe.ReplaceAllString(s, " ")
	return fetch.Clean(s)
}

// Parse extracts items from RSS or Atom XML.
func Parse(xml string) []Item {
	var items []Item
	for _, m := range blockRe.FindAllStringSubmatch(xml, -1) {
		body := m[1]
		var it Item
		if t := titleRe.FindStringSubmatch(body); t != nil {
			it.Title = clean(t[1])
		}
		if l := rssLinkRe.FindStringSubmatch(body); l != nil && strings.TrimSpace(clean(l[1])) != "" {
			it.Link = strings.TrimSpace(clean(l[1]))
		} else if l := atomLinkRe.FindStringSubmatch(body); l != nil {
			it.Link = strings.TrimSpace(l[1])
		}
		if d := descRe.FindStringSubmatch(body); d != nil {
			it.Text = clean(d[1])
		}
		if it.Title != "" {
			items = append(items, it)
		}
	}
	return items
}

// Fetch GETs a feed (as a browser) and parses it.
func Fetch(feedURL string) ([]Item, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, feedURL, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Accept", "application/rss+xml, application/atom+xml, application/xml, text/xml;q=0.9, */*;q=0.8")
	req.Header.Set("Accept-Language", "*")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 8<<20))
	if err != nil {
		return nil, err
	}
	return Parse(string(raw)), nil
}