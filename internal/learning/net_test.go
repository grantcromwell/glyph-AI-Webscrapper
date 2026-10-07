package learning

import (
	"testing"

	"glyphai/internal/encoding"
)

// The instinct net must actually learn: reconstruction loss should fall as it
// trains on a set of glyphs.
func TestTrainingLowersLoss(t *testing.T) {
	var set []encoding.Glyph
	for _, s := range []string{
		"the owl is the most superior animal that knows in silence",
		"the most superior weapon is the knife born in the forge",
		"language is alive and outranks writing it transforms reality",
		"the dog is loyalty as flow water and rhythm anticipation",
		"pattern recognition is the highest bridge the cortex sees itself",
	} {
		set = append(set, encoding.EncodeGlyph(s))
	}
	n := New(encoding.Families, 32, encoding.Families, 1)
	losses := n.Train(set, 200, 0.1, 0.2, 1)
	if len(losses) < 2 {
		t.Fatalf("expected loss history, got %d", len(losses))
	}
	first, last := losses[0], losses[len(losses)-1]
	if last >= first {
		t.Fatalf("loss did not fall: first=%.4f last=%.4f", first, last)
	}
	if last > 0.05 {
		t.Logf("note: final loss %.4f (small set, ok)", last)
	}
}

func TestCompleteReturnsGlyph(t *testing.T) {
	g := encoding.EncodeGlyph("the owl knows in silence")
	n := New(encoding.Families, 24, encoding.Families, 7)
	n.Train([]encoding.Glyph{g}, 50, 0.1, 0.0, 7)
	got := n.Complete(g)
	if got.Active() == 0 {
		t.Fatalf("completion produced an empty glyph")
	}
}