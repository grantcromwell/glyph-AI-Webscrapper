// Spell: skim (scrying school) — the dog's quick read.
//
// Where `browse` drives a full headless browser for living, JS-rendered sites,
// `skim` is the lighter incantation: a single polite HTTP GET, then read the
// title, the text, and the links. No browser, no JavaScript — instant. The dog
// casts this over static pages (a plain HTML site, a local library it has laid
// out for itself) where a full render would buy nothing but seconds of waiting.
//
// Same shape of sight as browse (docs + links), so a hunt can follow either.
package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"regexp"
	"strings"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"glyphai/internal/fetch"
)

var (
	titleRe  = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	linkRe   = regexp.MustCompile(`(?is)<a\s[^>]*href="([^"#][^"]*)"[^>]*>(.*?)</a>`)
	scriptRe = regexp.MustCompile(`(?is)<(script|style|noscript|head)[^>]*>.*?</(script|style|noscript|head)>`)
	tagRe    = regexp.MustCompile(`(?s)<[^>]+>`)
)

type link struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(err)
	}
	target := strings.TrimSpace(in.Text)
	if target == "" {
		target = in.Path
	}
	if target == "" {
		fail(fmt.Errorf("skim needs a url"))
	}
	if !strings.HasPrefix(target, "http") {
		target = "https://" + target
	}

	html, err := fetch.Get(target)
	if err != nil {
		fail(err)
	}

	title := target
	if m := titleRe.FindStringSubmatch(html); m != nil {
		title = clean(m[1])
	}
	links := extractLinks(html, target, 200)
	body := scriptRe.ReplaceAllString(html, " ")
	text := clean(tagRe.ReplaceAllString(body, " "))
	if len([]rune(text)) > 1200 {
		text = string([]rune(text)[:1200])
	}

	doc := fetch.Doc{Title: title, URL: target, Text: text, Source: "skim"}
	emit(modules.Output{
		Spell:   "skim",
		Detail:  encoding.Encode(title + " " + text),
		Summary: fmt.Sprintf("skimmed (http) %q — %d links", title, len(links)),
		Data: map[string]any{
			"docs":  []fetch.Doc{doc},
			"links": links,
			"mode":  "http",
		},
	})
}

func extractLinks(html, base string, max int) []link {
	baseURL, _ := url.Parse(base)
	seen := map[string]bool{}
	var out []link
	for _, m := range linkRe.FindAllStringSubmatch(html, -1) {
		href := m[1]
		if strings.HasPrefix(href, "javascript:") || strings.HasPrefix(href, "mailto:") {
			continue
		}
		abs := href
		if baseURL != nil {
			if ref, err := url.Parse(href); err == nil {
				abs = baseURL.ResolveReference(ref).String()
			}
		}
		text := clean(m[2])
		if abs == "" || seen[abs] || len(text) < 2 {
			continue
		}
		seen[abs] = true
		out = append(out, link{Text: text, URL: abs})
		if len(out) >= max {
			break
		}
	}
	return out
}

func clean(s string) string {
	s = tagRe.ReplaceAllString(s, "")
	return fetch.Clean(s)
}

func emit(o modules.Output) {
	if err := json.NewEncoder(os.Stdout).Encode(o); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "skim:", err)
	os.Exit(1)
}