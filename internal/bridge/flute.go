// Package pons is Argos's attention bridge — 7 focus slots that can hold
// simultaneous spell results and cross-reference them. Named after the
// pontine nuclei (pons), the brainstem structure that relays attention
// signals between cerebellum and cerebrum.
//
// Each slot is a short-term memory holding a spell's output: the glyph,
// its detail field, summary text, and conviction. Spells can be cast into
// any slot; slots can be blended, compared, or attended to.
package bridge

import (
	"fmt"
	"math"
	"sort"
	"strings"
	"time"

	"glyphai/internal/encoding"
)

// NumSlots is the dog's attention breadth — 7, like a human's working memory.
const NumSlots = 7

// A Slot holds one spell's output and the dog's attention on it.
type Slot struct {
	ID        int          `json:"id"`
	Spell     string       `json:"spell"`              // name of the spell that filled this slot
	Query     string       `json:"query"`              // what was asked
	Glyph     encoding.Glyph  `json:"glyph"`              // the coarse 21-byte reading
	Detail    encoding.Detail `json:"detail"`             // the full 1512-cell field
	Summary   string       `json:"summary,omitempty"`  // what the spell returned
	Data      string       `json:"data,omitempty"`     // additional structured text
	Conviction float64     `json:"conviction"`         // how strongly this slot fires (0-1)
	TakenAt   time.Time    `json:"taken_at"`
}

// Flute is the 7-slot attention bridge.
type Flute struct {
	Slots [NumSlots]*Slot
}

// New returns an empty flute.
func New() *Flute { return &Flute{} }

// Fill places a spell result into slot n (0-6). If slot is -1, finds the
// emptiest slot. Returns the slot index used.
func (f *Flute) Fill(slot int, spell, query string, d encoding.Detail, summary, data string) int {
	if slot < 0 || slot >= NumSlots {
		// Auto-assign to the emptiest slot (lowest conviction, or oldest)
		slot = emptiest(f)
	}
	f.Slots[slot] = &Slot{
		ID:         slot,
		Spell:      spell,
		Query:      query,
		Glyph:      d.Coarse(),
		Detail:     d,
		Summary:    summary,
		Data:       data,
		Conviction: float64(d.Active()) / float64(encoding.Cells),
		TakenAt:    time.Now(),
	}
	return slot
}

// Read returns the slot at index n, or nil if empty.
func (f *Flute) Read(n int) *Slot {
	if n < 0 || n >= NumSlots {
		return nil
	}
	return f.Slots[n]
}

// Clear empties slot n. If n is -1, clears all slots.
func (f *Flute) Clear(n int) {
	if n < 0 {
		for i := range f.Slots {
			f.Slots[i] = nil
		}
		return
	}
	if n < NumSlots {
		f.Slots[n] = nil
	}
}

// Active returns the indices of slots that have been filled.
func (f *Flute) Active() []int {
	var ids []int
	for i, s := range f.Slots {
		if s != nil {
			ids = append(ids, i)
		}
	}
	return ids
}

// Blend composes a single glyph from all active slots, weighted by conviction.
// This is the dog's "gestalt" — what it attends to collectively.
func (f *Flute) Blend() encoding.Glyph {
	var glyphs []encoding.Glyph
	for _, s := range f.Slots {
		if s != nil {
			glyphs = append(glyphs, s.Glyph)
		}
	}
	if len(glyphs) == 0 {
		return encoding.Glyph{}
	}
	return encoding.Compose(glyphs...)
}

// Attention computes the pairwise resonance matrix between all active slots.
// Returns a list of {slot_i, slot_j, resonance} sorted by strongest connection.
type AttnPair struct {
	I, J      int
	Resonance float64
	SpellI, SpellJ string
	SummaryI, SummaryJ string
}

func (f *Flute) Attention() []AttnPair {
	act := f.Active()
	if len(act) < 2 {
		return nil
	}
	var pairs []AttnPair
	for i := 0; i < len(act); i++ {
		for j := i + 1; j < len(act); j++ {
			si := f.Slots[act[i]]
			sj := f.Slots[act[j]]
			if si == nil || sj == nil {
				continue
			}
			r := encoding.Resonance(si.Glyph, sj.Glyph)
			pairs = append(pairs, AttnPair{
				I: act[i], J: act[j],
				Resonance:  r,
				SpellI: si.Spell, SpellJ: sj.Spell,
				SummaryI: truncate(si.Summary, 40), SummaryJ: truncate(sj.Summary, 40),
			})
		}
	}
	sort.Slice(pairs, func(a, b int) bool { return pairs[a].Resonance > pairs[b].Resonance })
	return pairs
}

// Decay reduces conviction of all slots over time, simulating attention drift.
// Slots below threshold are cleared.
func (f *Flute) Decay(threshold float64) {
	for i, s := range f.Slots {
		if s == nil {
			continue
		}
		age := time.Since(s.TakenAt).Minutes()
		s.Conviction *= math.Exp(-age * 0.3) // halve every ~2.3 minutes
		if s.Conviction < threshold {
			f.Slots[i] = nil
		}
	}
}

// Snapshot returns a human-readable summary of all active slots.
func (f *Flute) Snapshot() string {
	var b strings.Builder
	act := f.Active()
	if len(act) == 0 {
		b.WriteString("   (all slots empty)\n")
		return b.String()
	}
	for _, i := range act {
		s := f.Slots[i]
		if s == nil {
			continue
		}
		logogram := encoding.Logogram(s.Detail)
		reading := encoding.Reading(s.Glyph)
		age := time.Since(s.TakenAt).Truncate(time.Second).String()
		b.WriteString(fmtFluteSlot(i, s.Spell, s.Query, logogram, reading, s.Conviction, s.Summary, age))
	}
	return b.String()
}

func emptiest(f *Flute) int {
	best, bi := 999.0, 0
	for i, s := range f.Slots {
		if s == nil {
			return i
		}
		if s.Conviction < best {
			best, bi = s.Conviction, i
		}
	}
	_ = best
	return bi
}

func truncate(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "…"
}

func fmtFluteSlot(i int, spell, query, logogram, reading string, conviction float64, summary, age string) string {
	var b strings.Builder
	fmt.Fprintf(&b, "   slot %d [%s] ", i, spell)
	if query != "" {
		fmt.Fprintf(&b, "«%s» ", truncate(query, 24))
	}
	fmt.Fprintf(&b, "· %s %s\n", logogram, reading)
	fmt.Fprintf(&b, "           conviction %.2f · %s\n", conviction, age)
	if summary != "" {
		fmt.Fprintf(&b, "           %s\n", truncate(summary, 80))
	}
	return b.String()
}