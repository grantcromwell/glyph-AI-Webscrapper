package main

import (
	"math"
	"math/rand"

	"glyphai/internal/encoding"
)

// steer.go is the dog's NAVIGATIONAL INSTINCT, as a modular array of drives.
//
// Each step of a roam, the dog must choose which link to follow. That choice is
// the product of many competing instincts — pull toward what resonates, a learned
// nose for substance over junk, a taste for novelty, a pull to stay coherent, and
// (the dynamic part) a restlessness that breaks it out of any rut it falls into.
// Rather than bury these as inline factors, each instinct is a `drive`: a small
// function that scores one candidate link and returns a weight multiplier
// (>1 attracts, <1 repels, ~0 forbids). pickNext multiplies every drive together,
// so the `drives` slice below IS the dog's whole steering character — add a drive
// to give it a new behaviour, remove one to silence it, reorder freely.
//
// The drift-fighting drives (hostFatigue, vitality) are DYNAMIC: they read the
// dog's own running experience (how long it has dwelt on each host, the median
// richness of the link text it keeps seeing) and adjust themselves. Nothing here
// names a site or keyword — drift is recognised statistically and corrected.

// These three weights are the inherited navigation tuning (previously inline in
// pickNext). Named here so the instinct is transparent in one place.
const (
	pullGain      = 6.0  // how sharply contrast-resonance bends the walk
	visitedDamp   = 0.03 // a place already seen this roam is all but closed
	sameHostBonus = 1.15 // a mild pull to finish reading a site before leaving it
)

// senses is the dog's running navigational state — what it has learned so far on
// this walk. Drives read it; the main loop updates it.
type senses struct {
	rich, junk centroid        // learned "substance" vs "junk" page senses
	scent      centroid        // mean field of pages it has KEPT — its sense of purpose
	scentN     int             // how many kept pages have shaped the scent
	hostWalks  map[string]int  // how many pages it has dwelt on, per host (fatigue)
	anchorAct  []int           // running Active() of every link text seen (vitality)
	visited    map[string]bool // places already walked this roam
	curHost    string          // host of the page being read right now
	rng        *rand.Rand
}

// keep folds a page the dog chose to remember into its scent — the running mean
// of everything it has kept. Because it is a true mean, the pages it sets out
// among dominate it, so the scent anchors to the walk's purpose and resists drift.
func (s *senses) keep(d encoding.Detail) {
	s.scent.add(d)
	s.scentN++
}

func newSenses(rng *rand.Rand) *senses {
	return &senses{hostWalks: map[string]int{}, visited: map[string]bool{}, rng: rng}
}

// cand is one candidate link, with the per-step values drives need precomputed.
type cand struct {
	link link
	d    encoding.Detail // glyph of the link text
	pull float64      // contrast-resonance with what the dog is reading now
}

// drive is one instinct: name it, and give it a function that weights a candidate.
type drive struct {
	name string
	fn   func(s *senses, c *cand) float64
}

// drives is the ordered instinct stack. Their outputs are multiplied, so order is
// cosmetic — but reading top to bottom is the dog's priority of mind.
var drives = []drive{
	{"pull", drivePull},               // toward what resonates with the current page
	{"homing", driveHoming},           // DYNAMIC: toward the scent of what it set out to find
	{"substance", driveSubstance},     // toward the learned "rich" sense, away from "junk"
	{"novelty", driveNovelty},         // away from places already walked
	{"cohesion", driveCohesion},       // mildly toward the host being read
	{"hostFatigue", driveHostFatigue}, // DYNAMIC: away from a host it has over-dwelt
	{"vitality", driveVitality},       // DYNAMIC: away from thin/chrome link text
}

func drivePull(s *senses, c *cand) float64 { return math.Exp(pullGain * c.pull) }

// driveHoming is the anti-drift conscience. A topic-free walk has no defence
// against dense-but-off-topic pages (foreign news, encyclopaedic tangents): they
// are rich, so pull/substance/vitality all wave them through. Homing gives the dog
// a sense of PURPOSE — the mean glyph of the pages it has kept so far — and damps
// candidates by how far they fall from it. News will not resonate with a scent
// built of crystals and geometry, so it sinks. Neutral until the first page is
// kept; the scent is the dog's own accumulation, never a supplied topic.
func driveHoming(s *senses, c *cand) float64 {
	if s.scentN == 0 {
		return 1
	}
	return cosCentroid(c.d, &s.scent)
}

func driveSubstance(s *senses, c *cand) float64 { return substance(c.d, &s.rich, &s.junk) }

func driveNovelty(s *senses, c *cand) float64 {
	if s.visited[norm(c.link.URL)] {
		return visitedDamp
	}
	return 1
}

func driveCohesion(s *senses, c *cand) float64 {
	if hostOf(c.link.URL) == s.curHost {
		return sameHostBonus
	}
	return 1
}

// driveHostFatigue is the heart of the dynamic drift fix. A host the dog has
// already walked far more than its AVERAGE host is almost certainly a chrome trap
// (a wiki's infrastructure, a CMS's nav). The attraction to it decays as
// mean/(mean+n): neutral for a fresh host, 0.5 once a host hits the mean dwell,
// and ever smaller as it dominates — so the walk is forced to diversify. The bar
// is the dog's own mean dwell, not a fixed number, so it adapts to every walk.
func driveHostFatigue(s *senses, c *cand) float64 {
	n := s.hostWalks[hostOf(c.link.URL)]
	if n == 0 || len(s.hostWalks) == 0 {
		return 1
	}
	total := 0
	for _, v := range s.hostWalks {
		total += v
	}
	mean := float64(total) / float64(len(s.hostWalks))
	return mean / (mean + float64(n))
}

// driveVitality damps thin link text — "Jump to content", "Create account",
// "Privacy Policy" — without any keyword list. Such anchors encode to sparse
// glyphs (low Active); content anchors are richer. A candidate below the running
// median anchor richness is scaled down by its ratio to that median; at or above
// it is left neutral. The median is the dog's own evolving sense of "a normal
// link", so the bar moves with what it is actually seeing.
func driveVitality(s *senses, c *cand) float64 {
	med := medianInt(s.anchorAct)
	if med == 0 {
		return 1
	}
	if r := float64(c.d.Active()) / float64(med); r < 1 {
		return r
	}
	return 1
}

// --- learned page senses (substance vs junk), shared by the substance drive ---

// centroid is a running sum of glyph fields — the dog's learned sense of a kind
// of page (substance, or junk).
type centroid [encoding.Families][encoding.Subfamilies]float64

func (c *centroid) add(d encoding.Detail) {
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			c[f][s] += float64(d[f][s])
		}
	}
}

// substance scores a candidate by how much more it resembles the dog's learned
// "substance" sense than its "junk" sense, in [0,1]. Neutral 0.5 until learned.
func substance(d encoding.Detail, rich, junk *centroid) float64 {
	cr, cj := cosCentroid(d, rich), cosCentroid(d, junk)
	if cr+cj == 0 {
		return 0.5
	}
	return cr / (cr + cj)
}

func cosCentroid(d encoding.Detail, c *centroid) float64 {
	var dot, nd, nc float64
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			dv, cv := float64(d[f][s]), c[f][s]
			dot += dv * cv
			nd += dv * dv
			nc += cv * cv
		}
	}
	if nd == 0 || nc == 0 {
		return 0
	}
	return dot / (math.Sqrt(nd) * math.Sqrt(nc))
}