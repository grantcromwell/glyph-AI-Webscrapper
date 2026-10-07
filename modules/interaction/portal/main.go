// Spell: portal (play school) — the dog's hand into the game.
//
// The world is a separate headless process serving plain HTML over localhost.
// This spell is the modular seam: given a place URL, it GETs the page, reads the
// title + text the dog will perceive and the links it can take onward, and paints
// the place's encoding. Same shape of sight as the scrying eyes (docs + links), so
// the play loop consumes it exactly like roam consumes skim — but pointed at the
// game, not the web. Swap this spell to step into a different world.
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
		fail(fmt.Errorf("portal needs a place url"))
	}
	if !strings.HasPrefix(target, "http") {
		target = "http://" + target // the game lives on localhost http, not https
	}

	page, err := fetch.Get(target)
	if err != nil {
		fail(err)
	}

	title := target
	if m := titleRe.FindStringSubmatch(page); m != nil {
		title = clean(m[1])
	}
	links := extractLinks(page, target, 200)
	body := scriptRe.ReplaceAllString(page, " ")
	text := clean(tagRe.ReplaceAllString(body, " "))
	if len([]rune(text)) > 1200 {
		text = string([]rune(text)[:1200])
	}

	doc := fetch.Doc{Title: title, URL: target, Text: text, Source: "portal"}
	emit(modules.Output{
		Spell:   "portal",
		Detail:  encoding.Encode(title + " " + text),
		Summary: fmt.Sprintf("stepped into %q — %d paths onward", title, len(links)),
		Data: map[string]any{
			"docs":  []fetch.Doc{doc},
			"links": links,
		},
	})
}

func extractLinks(page, base string, max int) []link {
	baseURL, _ := url.Parse(base)
	seen := map[string]bool{}
	var out []link
	for _, m := range linkRe.FindAllStringSubmatch(page, -1) {
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
	fmt.Fprintln(os.Stderr, "portal:", err)
	os.Exit(1)
}