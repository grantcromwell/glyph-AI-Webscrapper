// Spell: feed (wild school)
//
// Feeds a directory of text files into Argos's memory as engrams.
// Each .txt file is remembered as a glyph experience the dog can recall.
//
// Input: modules.Input{Path: "training/corpus", Args: {"max_chars": 100000}}
// Output: modules.Output{Summary: "fed 3 files, total N engrams"}
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"glyphai/internal/modules"
	"glyphai/internal/memory"
)

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(fmt.Errorf("reading input: %v", err))
	}

	dir := in.Path
	if dir == "" {
		dir = "training/corpus"
	}

	maxChars := 100000
	if v, ok := in.Args["max_chars"].(float64); ok && v > 0 {
		maxChars = int(v)
	}

	h, err := memory.Open("memory", "records")
	if err != nil {
		fail(fmt.Errorf("open hippocampus: %v", err))
	}

	entries, err := os.ReadDir(dir)
	if err != nil {
		fail(fmt.Errorf("read dir %s: %v", dir, err))
	}

	loreDir := "data/lore"
	os.MkdirAll(loreDir, 0755)

	fed := 0
	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		path := filepath.Join(dir, e.Name())
		base := strings.TrimSuffix(e.Name(), ".txt")
		parts := strings.SplitN(base, "_", 2)
		topic := parts[0]
		source := "corpus"
		if len(parts) > 1 {
			source = "corpus:" + parts[1]
		}

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  feed: skip %s: %v\n", e.Name(), err)
			continue
		}
		text := string(data)
		if len(text) > maxChars {
			text = text[:maxChars]
		}

		if _, err := h.RememberTo(loreDir, topic, text, source); err != nil {
			fmt.Fprintf(os.Stderr, "  feed: %s error: %v\n", e.Name(), err)
			continue
		}
		fed++
	}

	emit(modules.Output{
		Spell:   "feed",
		Summary: fmt.Sprintf("fed %d files from %s. Total engrams: %d", fed, dir, h.Count()),
	})
}

func emit(o modules.Output) {
	if err := json.NewEncoder(os.Stdout).Encode(o); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "feed:", err)
	os.Exit(1)
}