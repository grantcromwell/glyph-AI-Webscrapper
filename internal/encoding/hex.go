package encoding

import "fmt"

// ParseHex reads a raven-style 21-pair glyph string ("070005...") back into a
// Glyph. Each pair is a level 0..7; out-of-range pairs are clamped.
func ParseHex(s string) (Glyph, error) {
	var g Glyph
	if len(s) != Families*2 {
		return g, fmt.Errorf("glyph: hex must be %d chars, got %d", Families*2, len(s))
	}
	for i := 0; i < Families; i++ {
		var v int
		if _, err := fmt.Sscanf(s[i*2:i*2+2], "%02d", &v); err != nil {
			return g, fmt.Errorf("glyph: bad pair %d: %w", i, err)
		}
		g[i] = clampLevel(v)
	}
	return g, nil
}