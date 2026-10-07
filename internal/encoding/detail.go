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
// Richness principle
// ------------------
// A glyph carries as much information as its source. A bare triangle should
// fire only a handful of subfamilies; a dense stained-glass painting should
// fire hundreds across many families. The system must perceive that gap, so a
// Detail exposes how much it actually holds: how many cells are active and how
// evenly the signal is spread (entropy). Modules paint detail with Set, then the
// orchestrator reads Active/Complexity to know how rich the perception was.
package encoding

import "math"

// Set raises cell (family, subfamily) to level (0..7), keeping the max if the
// cell was already brighter. Spells use this to paint perception into a Detail.
func (d *Detail) Set(family, subfamily int, level uint8) {
	if family < 0 || family >= Families || subfamily < 0 || subfamily >= Subfamilies {
		return
	}
	if level > MaxLevel {
		level = MaxLevel
	}
	if level > d[family][subfamily] {
		d[family][subfamily] = level
	}
}

// Add accumulates intensity into a cell, saturating at MaxLevel. Useful when a
// spell tallies many observations into the same subfamily.
func (d *Detail) Add(family, subfamily int, delta int) {
	if family < 0 || family >= Families || subfamily < 0 || subfamily >= Subfamilies {
		return
	}
	v := int(d[family][subfamily]) + delta
	d[family][subfamily] = clampLevel(v)
}

// Active counts non-zero cells out of all 1512 — the breadth of perception.
func (d Detail) Active() int {
	n := 0
	for f := 0; f < Families; f++ {
		for s := 0; s < Subfamilies; s++ {
			if d[f][s] > 0 {
				n++
			}
		}
	}
	return n
}

// Energy is the total summed intensity across every cell.
func (d Detail) Energy() int {
	sum := 0
	for f := 0; f < Families; f++ {
		for s := 0; s < Subfamilies; s++ {
			sum += int(d[f][s])
		}
	}
	return sum
}

// Complexity is the richness of the glyph in [0,1]: the Shannon entropy of the
// cell-intensity distribution, normalised by the maximum possible entropy.
// Energy concentrated in a few cells (a triangle) -> near 0; energy spread
// across many cells (a painting) -> near 1.
func (d Detail) Complexity() float64 {
	total := float64(d.Energy())
	if total == 0 {
		return 0
	}
	var h float64
	for f := 0; f < Families; f++ {
		for s := 0; s < Subfamilies; s++ {
			if d[f][s] == 0 {
				continue
			}
			p := float64(d[f][s]) / total
			h -= p * math.Log(p)
		}
	}
	return h / math.Log(float64(Cells)) // normalise by max entropy
}

// FamilySpread returns how many of the 21 families have at least one active
// subfamily — coarse breadth, complementing Glyph.Active.
func (d Detail) FamilySpread() int {
	n := 0
	for f := 0; f < Families; f++ {
		for s := 0; s < Subfamilies; s++ {
			if d[f][s] > 0 {
				n++
				break
			}
		}
	}
	return n
}