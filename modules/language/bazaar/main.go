// Spell: bazaar (tongue school) — Argos hears the bazaar.
//
// Polls the bazaar pool, feeds each foreign glyph through the lingua
// organ to coin or reinforce logograms. Returns what Argos now knows
// about the other animals.
//
// The dog doesn't translate. It builds its own vocabulary for foreign
// sounds. The logograms ARE the understanding.
package main

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"sort"
	"strings"
	"time"

	"glyphai/internal/modules"
	"glyphai/internal/translation"
)

type bazaarGl struct {
	ID        string      `json:"id"`
	Author    string      `json:"author"`
	FamilyVec [21]float64 `json:"family_vec"`
	Summary   string      `json:"summary"`
	Modality  string      `json:"modality"`
}

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(fmt.Errorf("reading input: %v", err))
	}

	bazaarURL := "http://localhost:7766"
	if a, ok := in.Args["bazaar_url"]; ok {
		if s, ok := a.(string); ok {
			bazaarURL = s
		}
	}
	bazaarURL = strings.TrimRight(bazaarURL, "/")

	// Open the lingua organ — it manages its own foreign vocabulary
	ling := translation.New(nil, "weights/lingua")

	// Fetch recent pool
	pool, err := fetchPool(bazaarURL)
	if err != nil {
		fmt.Fprintf(os.Stderr, "bazaar: fetch pool: %v\n", err)
		out := modules.Output{
			Summary: fmt.Sprintf("◌ %v", err),
		}
		json.NewEncoder(os.Stdout).Encode(out)
		return
	}

	// Filter foreign glyphs and hear them
	heard := 0
	newWords := 0
	byAuthor := make(map[string]int)

	for _, gl := range pool {
		if gl.Author == "" || gl.Author == "argos" {
			continue
		}

		// Convert float64 family vec to bytes for lingua
		rawVec := make([]byte, 21)
		for i, v := range gl.FamilyVec {
			if i < 21 {
				rawVec[i] = byte(min(int(v*7), 7))
			}
		}

		fw := ling.Hear(gl.Author, rawVec)
		byAuthor[gl.Author]++
		heard++
		if fw.Hearings <= 1 {
			newWords++
		}
	}

	// Build summary — pure glyph marks, no English
	lines := []string{}
	lines = append(lines, fmt.Sprintf("🐾 %d", heard))

	// Show new vocabulary per author as pure glyph marks
	authorStats := []string{}
	for author, count := range byAuthor {
		words := ling.Think(author)
		top := ""
		if len(words) > 0 {
			top = words[0].Logogram
		}
		authorStats = append(authorStats, fmt.Sprintf("%s %d %s", author, count, top))
	}
	sort.Strings(authorStats)
	for _, s := range authorStats {
		lines = append(lines, "  "+s)
	}

	if newWords > 0 {
		lines = append(lines, fmt.Sprintf("  +%d", newWords))
	}

	lines = append(lines, "")
	lines = append(lines, ling.Status())

	out := modules.Output{
		Summary: strings.Join(lines, "\n"),
	}
	json.NewEncoder(os.Stdout).Encode(out)
}

func fetchPool(base string) ([]bazaarGl, error) {
	client := &http.Client{Timeout: 10 * time.Second}
	resp, err := client.Get(base + "/pool")
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var pool []bazaarGl
	if err := json.Unmarshal(body, &pool); err != nil {
		return nil, err
	}
	return pool, nil
}

func fail(err error) {
	json.NewEncoder(os.Stdout).Encode(modules.Output{
		Summary: fmt.Sprintf("◌ %v", err),
	})
	os.Exit(1)
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}