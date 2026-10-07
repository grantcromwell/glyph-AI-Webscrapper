// olfaction.go — the BIOLOGY of the nose: how raw air becomes perceived scent.
//
// fetch.go fetches air (HTTP) and strips it to readable text. This file models what
// a real canine nose does with that air, as anatomy rather than I/O:
//
//   - Sniff       — active, intermittent sampling: a dog does not smell in one
//     continuous draw but in a burst of discrete breaths. A page is
//     therefore a TRAIL of scents, not one averaged smell.
//   - Turbinates  — the nasal conchae split airflow: most inhaled air is bulk
//     "respiratory" air that passes through, while the scent-laden
//     fraction is routed over the olfactory recess. Here the bulk
//     (boilerplate, chrome) is breathed out and the scent kept.
//   - Epithelium  — the receptor sheet ADAPTS: a scent met over and over
//     desensitises its receptors (habituation), so only what exceeds
//     the accumulated familiar exposure is newly perceived. This is
//     why a dog stops noticing a constant background smell — and why
//     the web's shared "hum" fades for Argos without any blocklist.
//
// Everything here emerges from the glyph: a breath is a natural unit of text, the
// turbinate cut is the dog's own median scent strength, adaptation is the running
// mean of what the receptors have been bathed in. No word lists, no tuned numbers.
// (Perceive is ORTHONASAL — smelling the outer world; smelling inward at recalled
// memory would be the retronasal counterpart, a future organ.)
package fetch

import (
	"regexp"
	"sort"
	"strings"

	"glyphai/internal/encoding"
)

// A Whiff is one discrete sample from a sniff: a breath of text, the glyph the
// receptors fire for it, and its intensity (how strongly it stirs them).
type Whiff struct {
	Text      string
	Scent     encoding.Detail
	Intensity int
}

// breathRe carves text into breaths at sentence terminators (Latin and CJK), so a
// breath is a natural parcel of air, not a fixed-size chunk. Script-agnostic: text
// with no terminators is taken as a single breath.
var breathRe = regexp.MustCompile(`[^.!?。！？…]+[.!?。！？…]*`)

func breaths(text string) []string {
	var out []string
	for _, s := range breathRe.FindAllString(text, -1) {
		if s = strings.TrimSpace(s); s != "" {
			out = append(out, s)
		}
	}
	if len(out) == 0 {
		if s := strings.TrimSpace(text); s != "" {
			out = append(out, s)
		}
	}
	return out
}

// Sniff samples text the way a dog samples air — in a burst of discrete breaths.
// Each breath fires a glyph; the whiffs are returned in order, so locality is kept
// (the start of a page may smell different from its end).
func Sniff(text string) []Whiff {
	bs := breaths(text)
	whiffs := make([]Whiff, 0, len(bs))
	for _, b := range bs {
		d := encoding.Encode(b)
		whiffs = append(whiffs, Whiff{Text: b, Scent: d, Intensity: d.Active()})
	}
	return whiffs
}

// Turbinates split a sniff as the nasal conchae split airflow: breaths carrying
// above-typical scent (intensity at or over the median of this sniff) are routed
// to the olfactory recess; the fainter bulk is shed as respiratory air. The bar is
// the dog's own median for THIS sniff, so a page of uniform boilerplate keeps
// little and a page with a few rich passages concentrates onto them — no list of
// words to ignore, just scent strength.
func Turbinates(whiffs []Whiff) (olfactory, respiratory []Whiff) {
	if len(whiffs) == 0 {
		return nil, nil
	}
	med := medianIntensity(whiffs)
	for _, w := range whiffs {
		if w.Intensity >= med {
			olfactory = append(olfactory, w)
		} else {
			respiratory = append(respiratory, w)
		}
	}
	return olfactory, respiratory
}

func medianIntensity(whiffs []Whiff) int {
	v := make([]int, len(whiffs))
	for i, w := range whiffs {
		v[i] = w.Intensity
	}
	sort.Ints(v)
	return v[len(v)/2]
}

// Epithelium is the receptor sheet's running state. It accumulates EXPOSURE — the
// sum of every scent it has met — and adapts to it: smelling a whiff returns that
// scent minus the sheet's mean exposure (clamped at zero), so a familiar smell
// barely registers while a novel one fires fully. This is olfactory habituation.
type Epithelium struct {
	exposure scentField
	n        int
}

// Smell returns the ADAPTED perception of a whiff and then breathes it in, raising
// the sheet's habituation for next time. Before any exposure the scent is perceived
// whole (the first breath is always vivid).
func (e *Epithelium) Smell(w Whiff) encoding.Detail {
	var out encoding.Detail
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			v := float64(w.Scent[f][s])
			if e.n > 0 {
				v -= e.exposure[f][s] / float64(e.n)
			}
			out[f][s] = clampLevel(v)
		}
	}
	e.exposure.add(w.Scent)
	e.n++
	return out
}

// Habituated reports how saturated the sheet is (how many breaths it has adapted
// to). Reset clears it — a fresh nose for a new outing.
func (e *Epithelium) Habituated() int { return e.n }
func (e *Epithelium) Reset()          { *e = Epithelium{} }

// Nose is the perceiving organ: a receptor sheet that remembers what it has
// smelled across a session, so adaptation carries between pages.
type Nose struct{ epi Epithelium }

func New() *Nose { return &Nose{} }

// Scent is the dog's biological reading of a page: the concentrated, adapted glyph,
// the passages that carried it, and how strongly it perceives the result.
type Scent struct {
	Glyph     encoding.Detail
	Passages  []string
	Intensity int
}

// Perceive runs the whole olfactory pipeline on fetched text: sniff it into
// breaths, route the scent-bearing fraction through the turbinates, and smell each
// through the adapting epithelium. The returned glyph is the mean of the adapted
// olfactory breaths — chrome shed by the turbinates, the familiar hum faded by
// habituation — i.e. what the dog newly, actually smells on this page.
func (n *Nose) Perceive(text string) Scent {
	whiffs := Sniff(text)
	olf, _ := Turbinates(whiffs)
	if len(olf) == 0 { // a faint page: smell all of it rather than nothing
		olf = whiffs
	}
	var acc scentField
	passages := make([]string, 0, len(olf))
	for _, w := range olf {
		acc.add(n.epi.Smell(w))
		passages = append(passages, w.Text)
	}
	g := acc.mean(len(olf))
	return Scent{Glyph: g, Passages: passages, Intensity: g.Active()}
}

// Bulb models the olfactory bulb: the receptors' signals converge on glomeruli
// where LATERAL INHIBITION sharpens the pattern — a strongly firing region
// suppresses the diffuse background, so the dominant scent stands out crisply.
// Implemented as center-surround: each cell keeps only what it carries above the
// mean activity of the lit field (the surround it competes against). The surround
// is the scent's own mean, so nothing is tuned. A flat, featureless scent is left
// as-is (no peak to sharpen).
func Bulb(d encoding.Detail) encoding.Detail {
	var sum float64
	var lit int
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			if v := float64(d[f][s]); v > 0 {
				sum += v
				lit++
			}
		}
	}
	if lit == 0 {
		return d
	}
	mean := sum / float64(lit)
	var out encoding.Detail
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			out[f][s] = clampLevel(float64(d[f][s]) - mean)
		}
	}
	return out
}

// scentField is a running sum of glyph fields, used for exposure and for averaging
// a set of breaths into one perceived scent.
type scentField [encoding.Families][encoding.Subfamilies]float64

func (c *scentField) add(d encoding.Detail) {
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			c[f][s] += float64(d[f][s])
		}
	}
}

func (c *scentField) mean(n int) encoding.Detail {
	var out encoding.Detail
	if n <= 0 {
		return out
	}
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			out[f][s] = clampLevel(c[f][s] / float64(n))
		}
	}
	return out
}

func clampLevel(v float64) uint8 {
	r := int(v + 0.5)
	if r < 0 {
		return 0
	}
	if r > encoding.MaxLevel {
		return uint8(encoding.MaxLevel)
	}
	return uint8(r)
}