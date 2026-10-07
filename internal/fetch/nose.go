// Package nose is Argos's generic sense of the wider world. It knows NO sites —
// the dog never hardcodes where to look. The nose only does two things:
//
//   - it is bound to the dir of logograms (LogogramDir), where everything it
//     brings home is deposited as logogram art;
//   - it offers a polite, generic web GET (Get) and a readability pass (Fetch).
//
// WHERE to sniff is decided by site-SPELLS in the grimoire (cortex/.../wikipedia,
// .../arxiv, …) that the hunt casts like a warlock. Adding a source means adding
// a spell, never editing the fetch.
package fetch

import (
	"context"
	"fmt"
	"html"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// LogogramDir is the one fixed thing the nose is wired to: the directory of
// logograms where sniffed knowledge is filed as art.
const LogogramDir = "data/lore"

const userAgent = "Argos/0.1 (canine knowledge agent; +https://localhost) polite-fetch"

// Doc is a unit of knowledge a site-spell brings back.
type Doc struct {
	Title  string `json:"title"`
	URL    string `json:"url"`
	Text   string `json:"text"`
	Source string `json:"source"`
}

var (
	tagRe    = regexp.MustCompile(`(?s)<[^>]+>`)
	scriptRe = regexp.MustCompile(`(?s)<(script|style|noscript)[^>]*>.*?</(script|style|noscript)>`)
	wsRe     = regexp.MustCompile(`\s+`)
)

// Get performs a polite, generic HTTP GET (the raw sense). Site-spells use this;
// the nose itself targets nothing in particular. Blockless principle: a real
// User-Agent and the front door, no scraping tricks.
func Get(u string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", userAgent)
	req.Header.Set("Accept", "application/json, application/atom+xml, text/html;q=0.8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("nose: %s -> %s", u, resp.Status)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	return string(raw), err
}

// Fetch GETs a page and strips it to readable text (generic, site-agnostic).
func Fetch(u string) (string, error) {
	body, err := Get(u)
	if err != nil {
		return "", err
	}
	body = scriptRe.ReplaceAllString(body, " ")
	body = tagRe.ReplaceAllString(body, " ")
	return strings.TrimSpace(wsRe.ReplaceAllString(html.UnescapeString(body), " ")), nil
}

// Clean collapses whitespace and unescapes entities (helper for site-spells).
func Clean(s string) string {
	return strings.TrimSpace(wsRe.ReplaceAllString(html.UnescapeString(s), " "))
}