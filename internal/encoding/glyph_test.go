package encoding

import "testing"

func TestEncodeDeterministic(t *testing.T) {
	a := EncodeGlyph("the owl counts in the dark")
	b := EncodeGlyph("the owl counts in the dark")
	if a != b {
		t.Fatalf("encode not deterministic: %v != %v", a, b)
	}
}

func TestSelfResonanceIsOne(t *testing.T) {
	g := EncodeGlyph("wavelet decomposition reveals hidden structure")
	if r := Resonance(g, g); r < 0.999 {
		t.Fatalf("self resonance = %v, want ~1", r)
	}
}

func TestSimilarMoreResonantThanDifferent(t *testing.T) {
	base := EncodeGlyph("the dog runs across the green field chasing the ball")
	near := EncodeGlyph("the dog runs across the field chasing a ball")
	far := EncodeGlyph("quarterly fiscal taxation revenue compliance audit")
	rNear := Resonance(base, near)
	rFar := Resonance(base, far)
	if rNear <= rFar {
		t.Fatalf("expected near (%v) > far (%v)", rNear, rFar)
	}
}

func TestCoarseLevelsInRange(t *testing.T) {
	d := Encode("frequency analysis splits light into families")
	g := d.Coarse()
	for f := 0; f < Families; f++ {
		if g[f] > MaxLevel {
			t.Fatalf("family %d level %d out of range", f, g[f])
		}
	}
}

func TestHexRoundTripsLength(t *testing.T) {
	g := EncodeGlyph("sumerian wrote the first song")
	if len(g.Hex()) != Families*2 {
		t.Fatalf("hex len = %d, want %d", len(g.Hex()), Families*2)
	}
}

func TestEmptyGlyphScoresZero(t *testing.T) {
	var g Glyph
	if g.Score() != 0 || g.Active() != 0 {
		t.Fatalf("empty glyph: score=%d active=%d", g.Score(), g.Active())
	}
}