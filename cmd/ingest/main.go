// feed-corpus — feeds training/corpus/*.txt into Argos's memory.
// Each file becomes one engram with topic from filename.
package main

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"glyphai/internal/memory"
)

func main() {
	corpusDir := "training/corpus"
	if len(os.Args) > 1 {
		corpusDir = os.Args[1]
	}

	h, err := memory.Open("memory", "records")
	if err != nil {
		fmt.Fprintf(os.Stderr, "open hippocampus: %v\n", err)
		os.Exit(1)
	}

	entries, err := os.ReadDir(corpusDir)
	if err != nil {
		fmt.Fprintf(os.Stderr, "read corpus dir: %v\n", err)
		os.Exit(1)
	}

	loreDir := "data/lore"
	os.MkdirAll(loreDir, 0755)

	for _, e := range entries {
		if e.IsDir() || !strings.HasSuffix(e.Name(), ".txt") {
			continue
		}
		path := filepath.Join(corpusDir, e.Name())
		base := strings.TrimSuffix(e.Name(), ".txt")
		parts := strings.SplitN(base, "_", 2)
		topic := parts[0]
		lang := ""
		if len(parts) > 1 {
			lang = parts[1]
		}

		data, err := os.ReadFile(path)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  skip %s: %v\n", e.Name(), err)
			continue
		}
		text := string(data)
		if len(text) > 100000 {
			text = text[:100000]
			fmt.Fprintf(os.Stderr, "  %s truncated to 100K chars\n", e.Name())
		}

		source := "corpus"
		if lang != "" {
			source = "corpus:" + lang
		}

		_, err = h.RememberTo(loreDir, topic, text, source)
		if err != nil {
			fmt.Fprintf(os.Stderr, "  %s remember error: %v\n", e.Name(), err)
			continue
		}
		fmt.Printf("  ✓ %s (%s, %d chars)\n", e.Name(), source, len(text))
	}

	fmt.Printf("\n🐾 corpus fed. Total engrams: %d\n", h.Count())
}