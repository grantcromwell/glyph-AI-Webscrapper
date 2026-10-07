// Package telekinesis is Argos's outward WILL — its agency to act on a world it
// is not part of. It does no I/O and no worldgen itself; it is the will, not the
// hand. To act it casts the portal spell (the hand, grimoire/play/portal), which
// does the actual contact with the headless game. Telekinesis chooses where to
// reach (sampling the parietal lobe's pull-weighted paths) and carries back what
// the hand brings.
//
// Split of labour: parietal = WHERE am I (the map); telekinesis = the WILL to
// reach; the portal spell = the HAND that touches the world.
package action

import (
	"context"
	"math/rand"

	"glyphai/internal/modules"
	"glyphai/internal/spatial"
)

type Telekinesis struct{ rng *rand.Rand }

func New(seed int64) *Telekinesis { return &Telekinesis{rng: rand.New(rand.NewSource(seed))} }

// Reach casts the portal spell to step to a place and returns what comes back.
func (t *Telekinesis) Reach(portal *modules.Spell, target string) (modules.Output, error) {
	return portal.Cast(context.Background(), modules.Input{Text: target})
}

// Sample picks the next path by the dog's pull — a weighted (Markov) draw, never
// greedy. Zero/negative weights fall back to a uniform draw so a senseless field
// still moves rather than stalling.
func (t *Telekinesis) Sample(opts []spatial.Weighted) string {
	if len(opts) == 0 {
		return ""
	}
	sum := 0.0
	for _, o := range opts {
		if o.Weight > 0 {
			sum += o.Weight
		}
	}
	if sum <= 0 {
		return opts[t.rng.Intn(len(opts))].URL
	}
	r := t.rng.Float64() * sum
	for _, o := range opts {
		if o.Weight <= 0 {
			continue
		}
		r -= o.Weight
		if r <= 0 {
			return o.URL
		}
	}
	return opts[len(opts)-1].URL
}