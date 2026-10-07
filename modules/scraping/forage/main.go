// Spell: forage (scrying school) — the dog's home browser onto the open web.
//
// Forked from `browse` (same headless sight), but where browse VISITS a known
// URL, forage SEARCHES: DuckDuckGo is the dog's front door to everything. It
// walks in through DDG's no-JavaScript HTML endpoint (html.duckduckgo.com) with
// an ordinary browser User-Agent — the blockless path, the same door a text
// browser uses, so it is never rate-limited or captcha-walled. If that ever
// fails it falls back to driving a real headless Chromium (full JS) exactly like
// browse. Returns clean docs + result links the hunt can then follow.
//
//	echo '{"text":"optical illusions that hypnotize"}' | go run ./grimoire/scrying/forage
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
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

// Browser-shaped User-Agent: forage knocks as a normal browser, never as a bot.
const browserUA = "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0"

var (
	resultRe  = regexp.MustCompile(`(?is)<a[^>]+class="result__a"[^>]*href="([^"]+)"[^>]*>(.*?)</a>`)
	snippetRe = regexp.MustCompile(`(?is)class="result__snippet"[^>]*>(.*?)</a>`)
	tagRe     = regexp.MustCompile(`(?s)<[^>]+>`)
	uddgRe    = regexp.MustCompile(`[?&]uddg=([^&]+)`)
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
	q := strings.TrimSpace(in.Text)
	if q == "" {
		q = in.Path
	}
	if q == "" {
		fail(fmt.Errorf("forage needs a query (text)"))
	}
	limit := 10
	if v, ok := in.Args["limit"].(float64); ok && v > 0 {
		limit = int(v)
	}

	html, mode, err := search(q)
	if err != nil {
		fail(err)
	}

	docs, links := parse(html, limit)
	if len(links) == 0 {
		fail(fmt.Errorf("forage found nothing for %q (mode %s)", q, mode))
	}

	combined := encoding.Detail{}
	for _, d := range docs {
		merge(&combined, encoding.Encode(d.Title+" "+d.Text))
	}
	emit(modules.Output{
		Spell:   "forage",
		Detail:  combined,
		Summary: fmt.Sprintf("foraged (%s) %q — %d results", mode, q, len(links)),
		Data: map[string]any{
			"docs":  docs,
			"links": links,
			"mode":  mode,
		},
	})
}

// search returns DuckDuckGo's results HTML. The blockless HTML endpoint first;
// a real headless Chromium only if that door is shut.
func search(q string) (html, mode string, err error) {
	endpoint := "https://html.duckduckgo.com/html/?q=" + url.QueryEscape(q)
	if body, e := fetchUA(endpoint); e == nil && strings.Contains(body, "result__a") {
		return body, "ddg-html", nil
	}
	// Fallback: drive a headless browser to render DDG fully (same as browse).
	if bin := findBrowser(); bin != "" {
		ctx, cancel := context.WithTimeout(context.Background(), 35*time.Second)
		defer cancel()
		cmd := exec.CommandContext(ctx, bin,
			"--headless=new", "--disable-gpu", "--no-sandbox",
			"--user-agent="+browserUA, "--virtual-time-budget=6000", "--dump-dom",
			"https://html.duckduckgo.com/html/?q="+url.QueryEscape(q))
		if out, e := cmd.Output(); e == nil && len(out) > 200 {
			return string(out), "chromium", nil
		}
	}
	// Last resort: the polite nose GET.
	body, e := fetch.Get(endpoint)
	return body, "http", e
}

// fetchUA does a GET as an ordinary browser would.
func fetchUA(u string) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return "", err
	}
	req.Header.Set("User-Agent", browserUA)
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.5")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("forage: %s -> %s", u, resp.Status)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, 4<<20))
	return string(raw), err
}

func parse(html string, limit int) ([]fetch.Doc, []link) {
	snips := snippetRe.FindAllStringSubmatch(html, -1)
	var docs []fetch.Doc
	var links []link
	for i, m := range resultRe.FindAllStringSubmatch(html, -1) {
		if len(links) >= limit {
			break
		}
		href := realURL(m[1])
		title := clean(m[2])
		if href == "" || title == "" {
			continue
		}
		snip := ""
		if i < len(snips) {
			snip = clean(snips[i][1])
		}
		docs = append(docs, fetch.Doc{Title: title, URL: href, Text: title + ". " + snip, Source: "duckduckgo"})
		links = append(links, link{Text: title, URL: href})
	}
	return docs, links
}

// realURL unwraps DDG's redirect (//duckduckgo.com/l/?uddg=ENCODED) to the bare
// destination; otherwise normalises a scheme-relative href.
func realURL(href string) string {
	if m := uddgRe.FindStringSubmatch(href); m != nil {
		if dec, err := url.QueryUnescape(m[1]); err == nil {
			return dec
		}
	}
	href = strings.TrimSpace(html2(href))
	if strings.HasPrefix(href, "//") {
		return "https:" + href
	}
	if strings.HasPrefix(href, "http") {
		return href
	}
	return ""
}

func clean(s string) string { return fetch.Clean(tagRe.ReplaceAllString(s, "")) }
func html2(s string) string { return strings.ReplaceAll(s, "&amp;", "&") }

func merge(dst *encoding.Detail, d encoding.Detail) {
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			if d[f][s] > dst[f][s] {
				dst[f][s] = d[f][s]
			}
		}
	}
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

func emit(o modules.Output) {
	if err := json.NewEncoder(os.Stdout).Encode(o); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "forage:", err)
	os.Exit(1)
}