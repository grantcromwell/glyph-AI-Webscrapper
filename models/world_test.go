package environment

import (
	"reflect"
	"testing"

	"glyphai/internal/encoding"
)

// Same seed + coord must regenerate byte-identical content (Minecraft determinism).
func TestChunkDeterministic(t *testing.T) {
	w1 := New("glyphai", 8)
	w2 := New("glyphai", 8)
	for _, c := range []Coord{{0, 0}, {3, -2}, {-7, 5}} {
		a := w1.Chunk(c)
		b := w2.Chunk(c)
		if !reflect.DeepEqual(a.Nodes, b.Nodes) || a.Biome != b.Biome {
			t.Fatalf("chunk %v not deterministic", c)
		}
	}
	// a different seed should (somewhere) diverge
	w3 := New("other-seed", 8)
	same := true
	for x := -10; x <= 10 && same; x++ {
		for y := -10; y <= 10; y++ {
			if w1.biomeName(Coord{x, y}) != w3.biomeName(Coord{x, y}) {
				same = false
				break
			}
		}
	}
	if same {
		t.Fatal("seed had no effect on the world")
	}
}

// Every chunk in a wide region must resolve to a known biome with content.
func TestBiomeCoverage(t *testing.T) {
	w := New("glyphai", 8)
	known := map[string]bool{"colorlands": true, "forge": true, "numberfields": true}
	seen := map[string]bool{}
	for x := -10; x <= 10; x++ {
		for y := -10; y <= 10; y++ {
			ch := w.Chunk(Coord{x, y})
			if !known[ch.Biome] {
				t.Fatalf("unknown biome %q at %d,%d", ch.Biome, x, y)
			}
			if len(ch.Nodes) == 0 {
				t.Fatalf("empty chunk at %d,%d", x, y)
			}
			seen[ch.Biome] = true
		}
	}
	if len(seen) < 3 {
		t.Fatalf("not all biomes appear in the region: %v", seen)
	}
}

// The substrate must cluster meaning: two places from the same domain resonate
// more than places from different domains. This is what makes the color wheel /
// number clusters emerge from glyph resonance rather than hardcoded links.
func TestDomainsClusterByResonance(t *testing.T) {
	w := New("glyphai", 8)
	pick := func(biome string) encoding.Glyph {
		for x := -20; x <= 20; x++ {
			for y := -20; y <= 20; y++ {
				ch := w.Chunk(Coord{x, y})
				if ch.Biome == biome {
					return ch.glyphs[0]
				}
			}
		}
		t.Fatalf("biome %q never appeared", biome)
		return encoding.Glyph{}
	}
	color1 := pick("colorlands")
	number1 := pick("numberfields")
	// a second colorlands sample, from a far chunk
	var color2 encoding.Glyph
	found := false
	for x := 20; x >= -20 && !found; x-- {
		for y := 20; y >= -20; y-- {
			ch := w.Chunk(Coord{x, y})
			if ch.Biome == "colorlands" {
				color2 = ch.glyphs[0]
				found = true
				break
			}
		}
	}
	within := encoding.Resonance(color1, color2)
	across := encoding.Resonance(color1, number1)
	if within <= across {
		t.Fatalf("domains did not cluster: within(color,color)=%.3f across(color,number)=%.3f", within, across)
	}
}