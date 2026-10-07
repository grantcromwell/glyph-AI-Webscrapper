package spatial

import (
	"path/filepath"
	"testing"
)

func TestIntegrateMapsAndLocates(t *testing.T) {
	p := New()
	p.Integrate("u1", "the number 7", "a prime number, odd, in the field of arithmetic")
	got := p.Integrate("u2", "azure hue", "a blue color at hue 210 degrees, a shade of light")
	if p.Mapped() != 2 {
		t.Fatalf("expected 2 mapped places, got %d", p.Mapped())
	}
	if !p.Seen("u1") || !p.Seen("u2") {
		t.Fatal("places not recorded on the map")
	}
	if p.Here().URL != "u2" {
		t.Fatalf("Here should be the last place, got %q", p.Here().URL)
	}
	if got.Biome == "" {
		t.Fatal("place was not assigned a region")
	}
}

func TestSaveLoadRoundtrip(t *testing.T) {
	p := New()
	p.Integrate("u1", "iron Fe", "a transition metal with a body-centered cubic lattice")
	p.Integrate("u2", "the number 28", "a perfect number, even, composite")
	path := filepath.Join(t.TempDir(), "cartograph.svg")
	if err := p.Save(path); err != nil {
		t.Fatal(err)
	}
	q := New()
	if err := q.Load(path); err != nil {
		t.Fatal(err)
	}
	if q.Mapped() != 2 {
		t.Fatalf("reloaded map has %d places, want 2", q.Mapped())
	}
	if !q.Seen("u1") || !q.Seen("u2") {
		t.Fatal("reloaded map missing places")
	}
}

func TestRegionIsStable(t *testing.T) {
	a := New()
	b := New()
	r1 := a.Integrate("x", "the number 12", "even composite divisible integer arithmetic").Biome
	r2 := b.Integrate("y", "the number 12", "even composite divisible integer arithmetic").Biome
	if r1 != r2 {
		t.Fatalf("region not stable: %q vs %q", r1, r2)
	}
}

func TestLoadMissingFileIsOK(t *testing.T) {
	p := New()
	if err := p.Load(filepath.Join(t.TempDir(), "nope.svg")); err != nil {
		t.Fatalf("missing cartograph should be fine, got %v", err)
	}
}