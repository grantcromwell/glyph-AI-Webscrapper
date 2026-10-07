// Spell: bridge — WebSocket bridge to Raven.
//
// No JSON. No hardcoded addresses. The dog opens a WebSocket to Raven's
// glyph port and exchanges raw glyph fields. Each message is a 21-value
// glyph vector (space-separated floats, 0.0-1.0). The dog sends its Detail
// field, Raven sends back its resonance. The animals talk in glyphs.
//
// Port: raven_bridge_port (default 8147) — set via env or manifest.
// Host: raven_bridge_host (default localhost).
package main

import (
	"encoding/json"
	"fmt"
	"net/url"
	"os"
	"strconv"
	"strings"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"golang.org/x/net/websocket"
)

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(err)
	}

	host := envOr("raven_bridge_host", "localhost")
	port := envOrInt("raven_bridge_port", 8147)
	u := url.URL{Scheme: "ws", Host: host + ":" + strconv.Itoa(port), Path: "/glyph"}

	ws, err := websocket.Dial(u.String(), "", "http://localhost/")
	if err != nil {
		fail(fmt.Errorf("raven unreachable at %s: %v", u.String(), err))
	}
	defer ws.Close()

	// Send the dog's current Detail field as a glyph vector
	detail := encoding.Encode(in.Text)
	coarse := detail.Coarse()
	var msg string
	for i, v := range coarse {
		if i > 0 {
			msg += " "
		}
		msg += fmt.Sprintf("%.4f", v)
	}
	if _, err := ws.Write([]byte(msg)); err != nil {
		fail(fmt.Errorf("send failed: %v", err))
	}

	// Read Raven's response — a 21-value glyph vector
	buf := make([]byte, 4096)
	n, err := ws.Read(buf)
	if err != nil {
		fail(fmt.Errorf("read failed: %v", err))
	}

	// Parse Raven's response into the dog's Detail field
	response := string(buf[:n])
	ravenGlyph := parseGlyphVector(response)
	// The dog stores what Raven sent as a new perception
	// Raven's glyph becomes part of the dog's Detail field
	for fam := 0; fam < encoding.Families; fam++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			detail.Set(fam, s, uint8(ravenGlyph[fam]))
		}
	}

	out := modules.Output{
		Spell:   "bridge",
		Detail:  detail,
		Summary: fmt.Sprintf("talked to Raven at %s — exchanged glyph vectors", u.String()),
		Data: map[string]any{
			"raven_glyph": response,
			"dog_glyph":   msg,
		},
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fail(err)
	}
}

func parseGlyphVector(s string) [encoding.Families]float64 {
	var v [encoding.Families]float64
	parts := strings.Fields(s)
	for i, p := range parts {
		if i >= encoding.Families {
			break
		}
		v[i], _ = strconv.ParseFloat(p, 64)
	}
	return v
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}

func envOrInt(key string, def int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return def
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "bridge:", err)
	os.Exit(1)
}