// Spell: hunt (frontal lobe)
// Argos's autonomous knowledge hunt — descended from raven's hunt, but glyph-
// driven and SVG-bound. Given a territory (a directory of .md/.txt/.go files),
// the dog reads every passage, encodes each into a glyph, and brings back the
// RICHEST ones (highest glyph complexity × family spread). The cerebrum then
// files each catch as an SVG art-engram in the bank.
//
// Contract: modules.Input{Path: territory, Args:{"max": N, "minlen": N}} ->
// modules.Output{Detail: combined glyph, Data:{"findings": [...]}}.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"glyphai/internal/modules"
	"glyphai/internal/encoding"
)

type finding struct {
	Topic      string  `json:"topic"`
	Text       string  `json:"text"`
	Glyph      string  `json:"glyph"`
	Complexity float64 `json:"complexity"`
	Richness   int     `json:"richness"`
	Source     string  `json:"source"`
}

var splitRe = regexp.MustCompile(`\n\s*\n`)

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(err)
	}
	territory := in.Path
	if territory == "" {
		fail(fmt.Errorf("hunt needs a territory path"))
	}
	maxFind := argInt(in.Args, "max", 8)
	minLen := argInt(in.Args, "minlen", 120)

	var founds []finding
	combined := encoding.Detail{}

	_ = filepath.WalkDir(territory, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if ext != ".md" && ext != ".txt" && ext != ".go" && ext != ".rs" && ext != ".c" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return nil
		}
		for _, para := range splitRe.Split(string(raw), -1) {
			p := strings.TrimSpace(para)
			if len(p) < minLen {
				continue
			}
			det := encoding.Encode(p)
			richness := det.Active() + 10*det.FamilySpread()
			founds = append(founds, finding{
				Topic:      firstWords(p, 7),
				Text:       collapse(p),
				Glyph:      det.Coarse().Hex(),
				Complexity: det.Complexity(),
				Richness:   richness,
				Source:     filepath.Base(path),
			})
		}
		return nil
	})

	// The dog keeps only the richest catches.
	sort.Slice(founds, func(i, j int) bool { return founds[i].Richness > founds[j].Richness })
	if len(founds) > maxFind {
		founds = founds[:maxFind]
	}
	// Combine the catches into one glyph (what the whole hunt "smelled like").
	for _, f := range founds {
		fd := encoding.Encode(f.Text)
		for fam := 0; fam < encoding.Families; fam++ {
			for s := 0; s < encoding.Subfamilies; s++ {
				combined.Set(fam, s, fd[fam][s])
			}
		}
	}

	out := modules.Output{
		Spell:   "hunt",
		Detail:  combined,
		Summary: fmt.Sprintf("hunted %s, brought back %d rich catches", territory, len(founds)),
		Data:    map[string]any{"findings": founds},
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fail(err)
	}
}

func firstWords(s string, n int) string {
	f := strings.Fields(collapse(s))
	if len(f) > n {
		f = f[:n]
	}
	return strings.Join(f, " ")
}

func collapse(s string) string {
	s = strings.ReplaceAll(s, "#", "")
	s = strings.ReplaceAll(s, "*", "")
	return strings.Join(strings.Fields(s), " ")
}

func argInt(m map[string]any, key string, def int) int {
	if m == nil {
		return def
	}
	if v, ok := m[key]; ok {
		if f, ok := v.(float64); ok {
			return int(f)
		}
	}
	return def
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "hunt:", err)
	os.Exit(1)
}