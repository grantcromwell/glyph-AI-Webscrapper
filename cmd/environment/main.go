// world — the headless GAME the dog plays. A standalone process (not the brain):
// it generates a Minecraft-blueprint world lazily, one chunk at a time, and
// serves it over localhost HTTP. The dog reaches it only through the portal spell.
//
//	world serve   [-addr 127.0.0.1:8088] [-seed WORD] [-dir DIR] [-k 8]
//	world prewarm [-radius 3] [-seed WORD] [-dir DIR] [-k 8]
//
// Generation is deterministic: same seed + coord => identical terrain, so the
// on-disk cache under -dir is only a speed-up, never authoritative.
package main

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"glyphai/models"
)

func main() {
	if len(os.Args) < 2 {
		fmt.Fprintln(os.Stderr, "usage: world <serve|prewarm> [flags]")
		os.Exit(2)
	}
	switch os.Args[1] {
	case "serve":
		serve(os.Args[2:])
	case "prewarm":
		prewarm(os.Args[2:])
	default:
		fmt.Fprintln(os.Stderr, "unknown subcommand:", os.Args[1])
		os.Exit(2)
	}
}

func serve(args []string) {
	addr, seed, dir, k, _ := flags("serve", args)
	w := environment.New(seed, k)
	if dir != "" {
		_ = os.MkdirAll(dir, 0o755)
	}
	h := func(rw http.ResponseWriter, req *http.Request) {
		page, ok := render(w, req.URL.Path)
		if !ok {
			http.NotFound(rw, req)
			return
		}
		if dir != "" {
			cache(dir, req.URL.Path, page)
		}
		rw.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = rw.Write([]byte(page))
	}
	http.HandleFunc("/", h)
	fmt.Printf("world %q serving on http://%s/index.html (cache %q)\n", seed, addr, dir)
	if err := http.ListenAndServe(addr, nil); err != nil {
		fmt.Fprintln(os.Stderr, "world:", err)
		os.Exit(1)
	}
}

// render maps a URL path to a generated page. Returns false for unknown routes.
func render(w *environment.World, path string) (string, bool) {
	switch {
	case path == "/" || path == "/index.html":
		return w.IndexPage(), true
	case strings.HasPrefix(path, "/chunk/"):
		if c, ok := environment.ParseChunk(strings.TrimPrefix(path, "/chunk/")); ok {
			return w.ChunkPage(c), true
		}
	case strings.HasPrefix(path, "/node/"):
		if c, i, ok := environment.ParseNode(strings.TrimPrefix(path, "/node/")); ok {
			return w.NodePage(c, i), true
		}
	}
	return "", false
}

func prewarm(args []string) {
	_, seed, dir, k, _ := flags("prewarm", args)
	radius := intFlag(args, "-radius", 3)
	if dir == "" {
		fmt.Fprintln(os.Stderr, "prewarm needs -dir")
		os.Exit(2)
	}
	w := environment.New(seed, k)
	_ = os.MkdirAll(dir, 0o755)
	n := 0
	for x := -radius; x <= radius; x++ {
		for y := -radius; y <= radius; y++ {
			c := environment.Coord{X: x, Y: y}
			cache(dir, fmt.Sprintf("/chunk/%d.%d.html", x, y), w.ChunkPage(c))
			n++
		}
	}
	cache(dir, "/index.html", w.IndexPage())
	fmt.Printf("prewarmed %d chunks of world %q into %s\n", n, seed, dir)
}

// cache writes a generated page under dir, mirroring the URL path, so the world
// materialises on disk as the dog explores (useful for inspection).
func cache(dir, urlPath, page string) {
	rel := strings.TrimPrefix(urlPath, "/")
	if rel == "" {
		rel = "index.html"
	}
	p := filepath.Join(dir, filepath.FromSlash(rel))
	_ = os.MkdirAll(filepath.Dir(p), 0o755)
	_ = os.WriteFile(p, []byte(page), 0o644)
}

// --- tiny flag parsing (keeps the entrypoint dependency-free) ---

func flags(sub string, args []string) (addr, seed, dir string, k, maxGB int) {
	addr = strFlag(args, "-addr", "127.0.0.1:8088")
	seed = strFlag(args, "-seed", "argos")
	dir = strFlag(args, "-dir", "")
	k = intFlag(args, "-k", 8)
	maxGB = intFlag(args, "-max-gb", 2)
	return
}

func strFlag(args []string, name, def string) string {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == name {
			return args[i+1]
		}
	}
	return def
}

func intFlag(args []string, name string, def int) int {
	for i := 0; i < len(args)-1; i++ {
		if args[i] == name {
			var v int
			if _, err := fmt.Sscanf(args[i+1], "%d", &v); err == nil {
				return v
			}
		}
	}
	return def
}