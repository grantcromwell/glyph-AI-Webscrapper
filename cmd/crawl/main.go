// roam — a self-guided hunt with no word-bank. The dog is given NO topics. It
// sets out from a place it already remembers (a URL carried in one of its own
// engrams) and, at every step, follows the link whose glyph the dog is most
// pulled toward given what it is reading right now (contrast-weighted resonance,
// sampled — a Markov walk). Pages that stir it are remembered (and, via
// RememberTo, automatically woven into the association vector space). When a
// trail dies it sets out again from another remembered place. The direction is
// the dog's own; nothing here names a subject. Runs until a deadline.
//
//	go run ./cmd/roam                 # until 12:45 PM America/New_York today
//	go run ./cmd/roam -until 1781455500
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"math/rand"
	"net/url"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"glyphai/internal/memory"
	"glyphai/internal/fetch"
	"glyphai/internal/output"
)

const (
	grimoireDir = "modules"
	memRoot     = "memory"
	memBank     = "records"
	loreDir     = "data/lore"
	hopsPerSeed = 16
)

type link struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

type place struct {
	url string
	g   encoding.Glyph
}

func main() {
	untilFlag := flag.Int64("until", 0, "deadline as unix epoch (0 = today 12:45 PM America/New_York)")
	eyesName := flag.String("eyes", "skim", "scrying spell to see with: skim (fast HTTP) or browse (headless Chromium, full JS)")
	seedFlag := flag.String("seed", "", "comma-separated start URLs to set out from first")
	bankFlag := flag.String("bank", memBank, "memory bank to file into (isolates concurrent hunts)")
	flag.Parse()

	target := deadline(*untilFlag)

	h, err := memory.Open(memRoot, *bankFlag)
	must(err)
	spells, err := modules.Discover(grimoireDir)
	must(err)
	var eyes *modules.Spell
	for i := range spells {
		if spells[i].Name == *eyesName {
			eyes = &spells[i]
		}
	}
	if eyes == nil {
		fmt.Printf("roam: no %q spell to see with\n", *eyesName)
		return
	}

	// Launch points are authoritative: when given, they are where the dog sets out
	// and the only places it RE-seeds from when a trail dies, so a themed hunt
	// stays anchored to its theme instead of falling back into the shared lore pile
	// (which earlier walks have already filled with off-topic pages). They are kept
	// separate from `known` and honoured even if already remembered.
	var launch []string
	for _, u := range strings.Split(*seedFlag, ",") {
		if u = strings.TrimSpace(u); u != "" {
			launch = append(launch, u)
		}
	}

	// Loading every remembered engram at startup can take tens of seconds when the
	// lore pool is large. For anchored (seeded) hunts we don't need the full set of
	// remembered places; for pure self-guided hunts we do, so pay the cost only then.
	var known []place
	var set map[string]bool
	if len(launch) > 0 {
		known, set = []place{}, map[string]bool{}
	} else {
		known, set = loadRememberedPlaces(loreDir)
	}

	if len(launch) == 0 && len(known) == 0 {
		fmt.Println("roam: the dog remembers no place to set out from — give it a -seed or let it prowl once first")
		return
	}
	if len(launch) > 0 {
		fmt.Printf("🐾 roam: %d launch points (anchored) · seeing with %q · self-guided until %s\n",
			len(launch), *eyesName, target.In(target.Location()).Format("3:04 PM MST"))
	} else {
		fmt.Printf("🐾 roam: %d remembered places · seeing with %q · self-guided until %s\n",
			len(known), *eyesName, target.In(target.Location()).Format("3:04 PM MST"))
	}

	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	s := newSenses(rng)
	seedAt, launchAt, filed, hops := 0, 0, 0, 0
	// the dog LEARNS to navigate: pages richer than the running median feed its
	// "substance" sense, poorer ones its "junk" sense; links are then steered by
	// the drives in steer.go toward substance and away from junk and drift. No
	// blocklist — a learned glyph sense plus dynamic host/vitality drives.
	var hist []int

	for time.Now().Before(target) {
		// Choose where to set out. With launch points, cycle through them (and only
		// them) so the hunt keeps returning to its theme; reopen the trail when all
		// have been dived. Without launch points (pure roam), set out from a
		// remembered place not yet walked — scent-guided once the dog has a purpose.
		seed := ""
		if len(launch) > 0 {
			for i := 0; i < len(launch); i++ {
				u := launch[(launchAt+i)%len(launch)]
				if !s.visited[norm(u)] {
					seed = u
					launchAt = (launchAt + i + 1) % len(launch)
					break
				}
			}
			if seed == "" { // all launch points dived — reopen and cycle again
				s.visited = map[string]bool{}
				seed = launch[launchAt%len(launch)]
				launchAt++
			}
		} else {
			for n := 0; n < len(known); n++ {
				p := known[(seedAt+n)%len(known)]
				if !s.visited[norm(p.url)] {
					seed = p.url
					seedAt = (seedAt + n + 1) % len(known)
					break
				}
			}
			if seed == "" { // walked them all — let it revisit
				s.visited = map[string]bool{}
				seed = known[rng.Intn(len(known))].url
			}
		}
		current := seed
		for hop := 0; hop < hopsPerSeed && time.Now().Before(target); hop++ {
			if current == "" || s.visited[norm(current)] {
				break
			}
			s.visited[norm(current)] = true
			s.curHost = hostOf(current)
			s.hostWalks[s.curHost]++ // dwell time per host feeds driveHostFatigue
			out, err := eyes.Cast(context.Background(), modules.Input{Text: current})
			if err != nil {
				break
			}
			docs := decodeDocs(out.Data["docs"])
			links := decodeLinks(out.Data["links"])
			if len(docs) == 0 {
				break
			}
			doc := docs[0]
			dg := encoding.Encode(doc.Title + " " + doc.Text)
			if dg.Active() >= 30 { // only a page that stirs the senses is kept
				if _, e := h.RememberTo(loreDir, doc.Title, doc.Text, "roam:"+doc.URL); e == nil {
					filed++
					s.keep(dg) // this kept page shapes the homing scent
				}
			}
			hops++
			// learn: is this page substance or junk, relative to all it has seen?
			a := dg.Active()
			hist = append(hist, a)
			if a >= medianInt(hist) {
				s.rich.add(dg)
			} else {
				s.junk.add(dg)
			}
			fmt.Printf("   [%s ET · %d filed] %s\n", nowET(target), filed, trunc(doc.Title, 60))
			current = pickNext(dg, links, s, &known, set)
		}
	}

	fmt.Printf("🐾 roam done: %d hops, %d engrams filed; bank now %d.\n", hops, filed, h.Count())
}

// pickNext follows the dog's pull, but the weighting is now the product of every
// drive in steer.go (pull, substance, novelty, cohesion, and the dynamic
// drift-fighters hostFatigue + vitality). It encodes each link, measures its
// contrast-resonance with the current page, multiplies the drives, and samples
// one link weighted by the result. Newly seen links join the remembered places so
// the dog's world keeps widening.
func pickNext(themeD encoding.Detail, links []link, s *senses, known *[]place, set map[string]bool) string {
	if len(links) == 0 {
		return ""
	}
	cands := make([]cand, len(links))
	pool := make([]encoding.Detail, 0, len(links)+1)
	pool = append(pool, themeD)
	for i, l := range links {
		d := encoding.Encode(l.Text)
		cands[i] = cand{link: l, d: d}
		pool = append(pool, d)
		s.anchorAct = append(s.anchorAct, d.Active()) // feed driveVitality's running sense
	}
	contrast := encoding.Contrast(pool)

	type step struct {
		url string
		w   float64
	}
	steps := make([]step, 0, len(cands))
	total := 0.0
	for i := range cands {
		cands[i].pull = encoding.ContrastResonance(themeD, cands[i].d, contrast)
		w := 1.0
		for _, dr := range drives {
			w *= dr.fn(s, &cands[i])
		}
		steps = append(steps, step{cands[i].link.URL, w})
		total += w
		if u := norm(cands[i].link.URL); u != "" && !set[u] && strings.HasPrefix(cands[i].link.URL, "http") {
			set[u] = true
			*known = append(*known, place{url: cands[i].link.URL, g: cands[i].d.Coarse()})
		}
	}
	if total == 0 {
		return ""
	}
	x, acc := s.rng.Float64()*total, 0.0
	pick := steps[len(steps)-1].url
	for _, st := range steps {
		if acc += st.w; x <= acc {
			pick = st.url
			break
		}
	}
	return pick
}

func medianInt(v []int) int {
	c := append([]int(nil), v...)
	sort.Ints(c)
	return c[len(c)/2]
}

// loadRememberedPlaces scans the dog's own engrams for the URLs it has been to —
// its starting points are derived from memory, never a hardcoded list.
func loadRememberedPlaces(dir string) ([]place, map[string]bool) {
	var out []place
	set := map[string]bool{}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return out, set
	}
	for _, de := range entries {
		if de.IsDir() || filepath.Ext(de.Name()) != ".svg" {
			continue
		}
		raw, err := os.ReadFile(filepath.Join(dir, de.Name()))
		if err != nil {
			continue
		}
		e, err := output.Parse(raw)
		if err != nil {
			continue
		}
		i := strings.Index(e.Source, "http")
		if i < 0 {
			continue
		}
		u := e.Source[i:]
		if n := norm(u); n != "" && !set[n] {
			set[n] = true
			g, _ := encoding.ParseHex(e.Glyph)
			out = append(out, place{url: u, g: g})
		}
	}
	return out, set
}

func deadline(epoch int64) time.Time {
	if epoch > 0 {
		return time.Unix(epoch, 0)
	}
	loc, err := time.LoadLocation("America/New_York")
	if err != nil {
		loc = time.Local
	}
	now := time.Now().In(loc)
	t := time.Date(now.Year(), now.Month(), now.Day(), 12, 45, 0, 0, loc)
	if !t.After(time.Now()) {
		t = t.Add(24 * time.Hour)
	}
	return t
}

func nowET(target time.Time) string { return time.Now().In(target.Location()).Format("3:04") }

func decodeDocs(v any) []fetch.Doc {
	raw, _ := json.Marshal(v)
	var d []fetch.Doc
	_ = json.Unmarshal(raw, &d)
	return d
}

func decodeLinks(v any) []link {
	raw, _ := json.Marshal(v)
	var l []link
	_ = json.Unmarshal(raw, &l)
	return l
}

func norm(u string) string {
	p, err := url.Parse(u)
	if err != nil {
		return u
	}
	p.Fragment = ""
	return p.String()
}

func hostOf(u string) string {
	if p, err := url.Parse(u); err == nil {
		return p.Host
	}
	return ""
}

func trunc(s string, n int) string {
	r := []rune(s)
	if len(r) <= n {
		return s
	}
	return string(r[:n]) + "…"
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "roam:", err)
		os.Exit(1)
	}
}