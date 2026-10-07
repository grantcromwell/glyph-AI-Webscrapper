// Package encoding is the core of the glyph AI system.
//
// DESIGN PATTERN: All perception, memory, and analysis converges on a single
// 21-family x 72-subfamily glyph field (1512 cells). Every input type (text,
// image, audio, web page) encodes to the same field. This is not a parser that
// decomposes inputs into pieces -- it is a combinator that projects all signals
// into a shared geometric space where resonance (cosine similarity) is the
// universal comparison operator.
//
// The 21 families are inspired by alchemical and planetary archetypes, not
// random labels. Each family has a breath-shape affinity (plosive, fricative,
// etc.) that determines which phonemes light it up. This means the encoding is
// language-agnostic: the system hears the breath, not the script.
//
// Key invariants:
//   - Families=21, Subfamilies=72, MaxLevel=7 (3 bits per cell)
//   - Detail [21][72]uint8 is the fine perception field
//   - Glyph [21]uint8 is the coarse fingerprint (used for indexing)
//   - All thresholds derive from data geometry, never hardcoded
//   - Ghost detection at 32% baseline: the most absent family is often the
//     most informative signal
//
// Every input, memory, spell and concept is encoded into a glyph: an
// interpretable fingerprint in a 21-family space. Each family has 72
// subfamilies, so the dog can perceive in fine detail (21*72 = 1512 cells)
// yet still think and index in the coarse 21-family space.
//
//	Detail  [21][72]uint8   fine perception   (what spells produce/consume)
//	Glyph   [21]uint8       coarse fingerprint (what the O(K) index keys on)
//
// A Detail coarsens to a Glyph; a Glyph is what gets spoken, scored, indexed
// and recalled. Levels run 0..7 (3 bits) so a coarse Glyph packs to 21 bytes.
package encoding

import (
	"fmt"
	"math"
	"strings"
)

const (
	Families    = 21 // the dog's 21 axes of meaning
	Subfamilies = 72 // detail per family
	MaxLevel    = 7  // every cell is 0..7
	Cells       = Families * Subfamilies
)

// FamilyNames are opaque labels for the 21 encoding axes.
// These labels carry no semantic meaning about the families.
// The original text cannot be recovered from a glyph.
var FamilyNames = [Families]string{
	"F00", "F01", "F02", "F03", "F04",
	"F05", "F06", "F07", "F08", "F09",
	"F10", "F11", "F12", "F13", "F14",
	"F15", "F16", "F17", "F18", "F19",
	"F20",
}

// familyWeights are uniform — no family is weighted higher than another.
// The original weighting scheme is not published.
var familyWeights = [Families]float64{
	1.0, 1.0, 1.0, 1.0, 1.0,
	1.0, 1.0, 1.0, 1.0, 1.0,
	1.0, 1.0, 1.0, 1.0, 1.0,
	1.0, 1.0, 1.0, 1.0, 1.0,
	1.0,
}

// Glyph is the coarse 21-family fingerprint: the dog's spoken word and the
// key used by the O(K) additive index.
type Glyph [Families]uint8

// Detail is the fine 21x72 perception field produced and consumed by spells.
type Detail [Families][Subfamilies]uint8

// FamilyIndex returns the axis index for a family name, or -1.
func FamilyIndex(name string) int {
	for i, n := range FamilyNames {
		if strings.EqualFold(n, name) {
			return i
		}
	}
	return -1
}

// Coarse aggregates each family's 72 subfamilies down to a single 0..7 level.
// Coarse aggregates each family to a 0..7 level — PEAKED and self-normalised so
// glyphs are distinctive, not uniformly lit. Each family scores by its
// concentration (peak-led with a breadth bonus); scores are then normalised by
// the strongest family, so the dominant axes stand out at 7 and weak ones fall
// to 0. This is what lets resonance actually tell two meanings apart.
func (d Detail) Coarse() Glyph {
	var g Glyph
	var score [Families]float64
	max := 0.0
	for f := 0; f < Families; f++ {
		var sum, peak float64
		for s := 0; s < Subfamilies; s++ {
			v := float64(d[f][s])
			sum += v
			if v > peak {
				peak = v
			}
		}
		mean := sum / float64(Subfamilies)
		score[f] = 0.7*peak + 0.3*mean // concentration, in 0..7 units
		if score[f] > max {
			max = score[f]
		}
	}
	if max == 0 {
		return g
	}
	for f := 0; f < Families; f++ {
		// Relative to the strongest family, with a gentle gamma so mid families
		// don't all flatten to the same level.
		rel := score[f] / max
		g[f] = clampLevel(int(math.Round(math.Pow(rel, 1.35) * MaxLevel)))
	}
	return g
}

// Active reports how many families are non-zero — raven's "active families"
// signal of glyph complexity.
func (g Glyph) Active() int {
	n := 0
	for _, v := range g {
		if v > 0 {
			n++
		}
	}
	return n
}

// Score is the superiority score: a weighted, complexity-bonused sum of family
// levels. Used to rank concepts/memories the way raven ranked objects.
func (g Glyph) Score() int {
	var s float64
	for f := 0; f < Families; f++ {
		s += float64(g[f]) * familyWeights[f]
	}
	// Complexity bonus: glyphs that light up many families score higher.
	s *= 1.0 + 0.25*float64(g.Active())/Families
	return int(math.Round(s * 10))
}

// Resonance is the dog's similarity sense over coarse glyphs: a family-weighted
// cosine in [0,1]. 1.0 == identical direction, 0 == orthogonal/empty.
func Resonance(a, b Glyph) float64 {
	var dot, na, nb float64
	for f := 0; f < Families; f++ {
		w := familyWeights[f]
		av := float64(a[f]) * w
		bv := float64(b[f]) * w
		dot += av * bv
		na += av * av
		nb += bv * bv
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// DetailResonance compares full 1512-cell fields for fine-grained matching.
func DetailResonance(a, b Detail) float64 {
	var dot, na, nb float64
	for f := 0; f < Families; f++ {
		for s := 0; s < Subfamilies; s++ {
			av := float64(a[f][s])
			bv := float64(b[f][s])
			dot += av * bv
			na += av * av
			nb += bv * bv
		}
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// ContrastResonance is DetailResonance with a per-cell weight: each glyph-cell
// counts in proportion to how informative it is. A sense the dog feels equally
// toward every option (weight→0) drops out of the comparison; the senses that
// actually distinguish the options decide the pull. No cell is filtered — an
// uninformative one simply weighs nothing.
func ContrastResonance(a, b Detail, w *[Families][Subfamilies]float64) float64 {
	var dot, na, nb float64
	for f := 0; f < Families; f++ {
		for s := 0; s < Subfamilies; s++ {
			cw := w[f][s]
			av := float64(a[f][s]) * cw
			bv := float64(b[f][s]) * cw
			dot += av * bv
			na += av * av
			nb += bv * bv
		}
	}
	if na == 0 || nb == 0 {
		return 0
	}
	return dot / (math.Sqrt(na) * math.Sqrt(nb))
}

// Contrast returns a per-cell weight = how much each glyph-cell varies (std dev)
// across the given fields. A sense that fires the same on every field carries
// no information for telling them apart, so its weight is ~0; a sense that
// differs across them gets full weight. Emergent salience: the dog listens to
// what differs, never to a hand-written list of what to ignore.
func Contrast(ds []Detail) *[Families][Subfamilies]float64 {
	var w [Families][Subfamilies]float64
	n := float64(len(ds))
	if n == 0 {
		return &w
	}
	var mean [Families][Subfamilies]float64
	for _, d := range ds {
		for f := 0; f < Families; f++ {
			for s := 0; s < Subfamilies; s++ {
				mean[f][s] += float64(d[f][s])
			}
		}
	}
	for f := 0; f < Families; f++ {
		for s := 0; s < Subfamilies; s++ {
			mean[f][s] /= n
		}
	}
	for _, d := range ds {
		for f := 0; f < Families; f++ {
			for s := 0; s < Subfamilies; s++ {
				diff := float64(d[f][s]) - mean[f][s]
				w[f][s] += diff * diff
			}
		}
	}
	for f := 0; f < Families; f++ {
		for s := 0; s < Subfamilies; s++ {
			w[f][s] = math.Sqrt(w[f][s] / n)
		}
	}
	return &w
}

// Compose blends glyphs into one (saturating mean), e.g. to merge a query with
// recalled memories before responding.
func Compose(gs ...Glyph) Glyph {
	var out Glyph
	if len(gs) == 0 {
		return out
	}
	for f := 0; f < Families; f++ {
		var sum int
		for _, g := range gs {
			sum += int(g[f])
		}
		out[f] = clampLevel(int(math.Round(float64(sum) / float64(len(gs)))))
	}
	return out
}

// Pack serialises the coarse glyph to its 21-byte index key.
func (g Glyph) Pack() [Families]byte {
	var b [Families]byte
	for i, v := range g {
		b[i] = v
	}
	return b
}

// Hex renders the raven-style 21-pair glyph string, e.g. "070005050206...".
func (g Glyph) Hex() string {
	var sb strings.Builder
	for _, v := range g {
		fmt.Fprintf(&sb, "%02d", v)
	}
	return sb.String()
}

// String shows the glyph as named families that are firing, strongest first.
func (g Glyph) String() string {
	type fv struct {
		name string
		v    uint8
	}
	var on []fv
	for f := 0; f < Families; f++ {
		if g[f] > 0 {
			on = append(on, fv{FamilyNames[f], g[f]})
		}
	}
	for i := 1; i < len(on); i++ { // tiny insertion sort, strongest first
		for j := i; j > 0 && on[j].v > on[j-1].v; j-- {
			on[j], on[j-1] = on[j-1], on[j]
		}
	}
	parts := make([]string, len(on))
	for i, p := range on {
		parts[i] = fmt.Sprintf("%s=%d", p.name, p.v)
	}
	return "{" + strings.Join(parts, " ") + "}"
}

func clampLevel(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > MaxLevel {
		return MaxLevel
	}
	return uint8(v)
}