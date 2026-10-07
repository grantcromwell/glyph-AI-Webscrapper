package fetch

import (
	"testing"

	"glyphai/internal/encoding"
)

// Sniffing samples a page as discrete breaths in order, one per sentence.
func TestSniffBreathsInOrder(t *testing.T) {
	whiffs := Sniff("Alpha beta gamma. Delta epsilon zeta! Eta theta iota?")
	if len(whiffs) != 3 {
		t.Fatalf("want 3 breaths, got %d", len(whiffs))
	}
	for i, w := range whiffs {
		if w.Text == "" || w.Intensity != w.Scent.Active() {
			t.Fatalf("whiff %d malformed: %+v", i, w)
		}
	}
	// text with no terminators is a single breath, never zero
	if got := Sniff("just one long unterminated breath of air"); len(got) != 1 {
		t.Fatalf("want 1 breath for unterminated text, got %d", len(got))
	}
	if got := Sniff("   "); len(got) != 0 {
		t.Fatalf("want 0 breaths for blank air, got %d", len(got))
	}
}

// The turbinates route above-median scent to perception and shed the faint bulk.
func TestTurbinatesConcentrate(t *testing.T) {
	whiffs := []Whiff{
		{Text: "a", Intensity: 1},
		{Text: "b", Intensity: 2},
		{Text: "c", Intensity: 5},
		{Text: "d", Intensity: 8},
		{Text: "e", Intensity: 9},
	}
	olf, resp := Turbinates(whiffs)
	if len(olf)+len(resp) != len(whiffs) {
		t.Fatalf("turbinates dropped air: %d+%d != %d", len(olf), len(resp), len(whiffs))
	}
	// median is 5; the olfactory fraction keeps everything >= 5
	for _, w := range olf {
		if w.Intensity < 5 {
			t.Fatalf("faint whiff routed to perception: %+v", w)
		}
	}
	if len(olf) == 0 || len(olf) == len(whiffs) {
		t.Fatalf("expected a real split, got %d olfactory of %d", len(olf), len(whiffs))
	}
}

// The epithelium habituates: smelling the same scent over and over fades it.
func TestEpitheliumHabituates(t *testing.T) {
	e := &Epithelium{}
	w := Sniff("crystalline lattice symmetry of a quasicrystal.")[0]
	first := e.Smell(w) // n==0, perceived whole
	if first.Active() != w.Intensity {
		t.Fatalf("first smell should be vivid: got %d want %d", first.Active(), w.Intensity)
	}
	for i := 0; i < 12; i++ {
		e.Smell(w) // bathe the receptors in the same scent
	}
	last := e.Smell(w)
	if last.Active() >= first.Active() {
		t.Fatalf("scent did not habituate: first %d, last %d", first.Active(), last.Active())
	}
}

// A novel scent still fires through a habituated sheet.
func TestNoveltyBreaksHabituation(t *testing.T) {
	e := &Epithelium{}
	familiar := Sniff("the same dull background hum over and over again.")[0]
	for i := 0; i < 12; i++ {
		e.Smell(familiar)
	}
	novel := Sniff("a sudden sharp unfamiliar pungent foreign trace.")[0]
	got := e.Smell(novel)
	if got.Active() == 0 {
		t.Fatalf("novel scent was wrongly suppressed by habituation")
	}
}

// Perceive sheds chrome and returns the passages that carried the scent.
func TestPerceiveKeepsScentBearingPassages(t *testing.T) {
	n := New()
	page := "Skip to content. Crystallography studies the arrangement of atoms in crystalline solids and their symmetry groups. Privacy policy."
	sc := n.Perceive(page)
	if len(sc.Passages) == 0 {
		t.Fatal("perceive kept nothing")
	}
	if sc.Glyph.Active() == 0 {
		t.Fatal("perceive returned an empty scent")
	}
}

// The bulb sharpens: a peaked scent loses its diffuse background to lateral
// inhibition while keeping the peak; an empty scent is untouched.
func TestBulbSharpens(t *testing.T) {
	d := Sniff("crystallography symmetry lattice diffraction.")[0].Scent
	sharp := Bulb(d)
	if sharp.Active() == 0 {
		t.Fatal("bulb suppressed everything")
	}
	if sharp.Active() > d.Active() {
		t.Fatalf("lateral inhibition should not light new cells: %d > %d", sharp.Active(), d.Active())
	}
	var empty encoding.Detail
	if Bulb(empty).Active() != 0 {
		t.Fatal("bulb should leave an empty scent empty")
	}
}