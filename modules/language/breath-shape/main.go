// Spell: breath-shape (temporal lobe)
// Decomposes text into the physics of the mouth — the breath-shape sequence the
// cochlea hears — and returns it as a glyph plus a readable breakdown. This is
// the deterministic, exact form of the dog's sound sense.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"glyphai/internal/audio"
	"glyphai/internal/modules"
	"glyphai/internal/encoding"
)

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(err)
	}
	text := in.Text
	if text == "" {
		text = in.Path
	}
	if text == "" {
		fail(fmt.Errorf("breath-shape needs text"))
	}

	counts := audio.Counts(text)
	var parts []string
	for b := audio.Breath(0); b < audio.NumBreaths; b++ {
		if counts[b] > 0 {
			parts = append(parts, fmt.Sprintf("%s %s×%d", audio.BreathSymbol[b], audio.BreathName[b], counts[b]))
		}
	}
	prof := audio.Profile(text)

	out := modules.Output{
		Spell:   "breath-shape",
		Detail:  encoding.Encode(text),
		Summary: fmt.Sprintf("breath: %s · dominant=%s", strings.Join(parts, " | "), audio.BreathName[prof]),
		Data: map[string]any{
			"dominant": audio.BreathName[prof],
			"quality":  audio.BreathQuality[prof],
		},
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "breath-shape:", err)
	os.Exit(1)
}