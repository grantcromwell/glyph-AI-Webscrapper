// Spell: browse (scrying school) — the dog's headless browser.
//
// Inspired by browser-use (github.com/browser-use/browser-use): give the agent
// a real browser action-space. Here, in pure Go: the spell visits a URL,
// renders it to readable text, and extracts the links so the dog can NAVIGATE —
// follow a link, then another, browsing like we do.
//
// It auto-upgrades: if a headless Chromium/Chrome is installed it drives it
// (`--headless --dump-dom`) for full JavaScript rendering; otherwise it does a
// polite HTTP fetch and parses the HTML. No Go dependencies either way.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"glyphai/internal/fetch"
)

var (
	titleRe  = regexp.MustCompile(`(?is)<title[^>]*>(.*?)</title>`)
	linkRe   = regexp.MustCompile(`(?is)<a\s[^>]*href="([^"#][^"]*)"[^>]*>(.*?)</a>`)
	scriptRe = regexp.MustCompile(`(?is)<(script|style|noscript|head)[^>]*>.*?</(script|style|noscript|head)>`)
	tagRe    = regexp.MustCompile(`(?s)<[^>]+>`)
	wsRe     = regexp.MustCompile(`\s+`)
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
		fail(fmt.Errorf("browse needs a url"))
	}
	if !strings.HasPrefix(target, "http") {
		target = "https://" + target
	}

	html, mode, err := render(target)
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

	doc := fetch.Doc{Title: title, URL: target, Text: text, Source: "browse"}
	emit(modules.Output{
		Spell:   "browse",
		Detail:  encoding.Encode(title + " " + text),
		Summary: fmt.Sprintf("browsed (%s) %q — %d links", mode, title, len(links)),
		Data: map[string]any{
			"docs":  []fetch.Doc{doc},
			"links": links,
			"mode":  mode,
		},
	})
}

// render returns the page HTML/DOM. It drives a headless browser if one exists
// (full JS), else falls back to a polite HTTP GET.
func render(u string) (html, mode string, err error) {
	if bin := findBrowser(); bin != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, bin,
			"--headless=new", "--disable-gpu", "--no-sandbox",
			"--virtual-time-budget=5000", "--dump-dom", u)
		if out, e := cmd.Output(); e == nil && len(out) > 200 {
			return string(out), "chromium", nil
		}
	}
	body, e := fetch.Get(u)
	return body, "http", e
}

func findBrowser() string {
	for _, b := range []string{"chromium", "chromium-browser", "google-chrome",
		"google-chrome-stable", "chrome", "brave", "brave-browser", "microsoft-edge"} {
		if p, err := exec.LookPath(b); err == nil {
			return p
		}
	}
	return ""
}

// extractLinks pulls the navigable links so the dog can follow them.
func extractLinks(html, base string, max int) []link {
	src := html
	baseURL, _ := url.Parse(base)
	seen := map[string]bool{}
	var out []link
	for _, m := range linkRe.FindAllStringSubmatch(src, -1) {
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
	fmt.Fprintln(os.Stderr, "browse:", err)
	os.Exit(1)
}