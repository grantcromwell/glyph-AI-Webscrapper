// cerebrum — Argos's mind. The interactive orchestrator that thinks in glyphs,
// hunts and transposes language into SVG art-lore, recalls it through the
// hippocampus, and produces a spoken (text) answer.
//
// Usage:
//
//	cerebrum stats
//	cerebrum transpose <textfile> [outdir]   # language -> SVG lore via the key
//	cerebrum hunt [territory]                 # hunt richest passages -> SVG lore
//	cerebrum ingest [loredir]                 # index existing SVG lore
//	cerebrum recall "<query>" [k]
//	cerebrum ask "<question>"
package main

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io"
	"math"
	"math/rand"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"time"
	"unicode"

	"glyphai/internal/learning"
	"glyphai/internal/audio"
	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"glyphai/internal/memory"
	"glyphai/internal/vocabulary"
	"glyphai/internal/chat"
	"glyphai/internal/fetch"
	"glyphai/internal/bridge"
	"glyphai/internal/semantics"
	"glyphai/internal/routing"
	"glyphai/internal/translation"
)

const (
	memRoot     = "memory"
	memBank     = "records"
	loreDir     = "data/lore"
	grimoireDir = "modules"
	lexPath     = "weights/vocabulary.svg"
	netPath      = "weights/learning.svg"
	foreseePath  = "weights/foresight.svg"
	baselinePath = "weights/baseline.svg"
)

func main() {
	if len(os.Args) < 2 {
		usage()
	}
	h, err := memory.Open(memRoot, memBank)
	must(err)

	flute := bridge.New() // 7-slot attention bridge

	switch os.Args[1] {
	case "stats":
		fmt.Printf("🐾 Argos remembers %d engrams.\n", h.Count())
	case "lingua":
		// lingua <animal> — show what Argos knows about a foreign language
		lex, err := vocabulary.Open(lexPath)
		must(err)
		ling := translation.New(lex, "weights/lingua")
		if len(os.Args) > 2 {
			words := ling.Think(os.Args[2])
			fmt.Printf("🐾 Lingua — %d words for %s:\n", len(words), os.Args[2])
			for _, w := range words {
				fmt.Printf("   %s  confidence=%.0f%% heard=%d\n", w.Logogram, w.Confidence*100, w.Hearings)
			}
		} else {
			fmt.Println("🐾 " + ling.Status())
		}
	case "coverage":
		cells, fams := h.Coverage()
		fmt.Printf("🐾 hunger fed: %d/%d subglyphs lit across %d engrams · %d/21 families\n",
			cells, encoding.Cells, h.Count(), fams)
	case "transpose":
		if len(os.Args) < 3 {
			usage()
		}
		out := loreDir
		if len(os.Args) > 3 {
			out = os.Args[3]
		}
		transposeFile(h, os.Args[2], out)
	case "ingest":
		dir := loreDir
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		ingest(h, dir)
	case "hunt":
		territory := "gut/source"
		if len(os.Args) > 2 {
			territory = os.Args[2]
		}
		hunt(h, territory)
	case "recall":
		if len(os.Args) < 3 {
			usage()
		}
		k := 4
		if len(os.Args) > 3 {
			k, _ = strconv.Atoi(os.Args[3])
		}
		for i, hit := range h.Recall(os.Args[2], k) {
			fmt.Printf("%d. [%.2f] %s\n   %s\n", i+1, hit.Score, hit.Topic, truncate(hit.Text, 100))
		}
	case "ask":
		if len(os.Args) < 3 {
			usage()
		}
		ask(h, strings.Join(os.Args[2:], " "))
	case "chat":
		chatLoop(h)
	case "reindex":
		n, err := h.ReindexDir(loreDir)
		must(err)
		fmt.Printf("🐾 re-encoded %d lore engrams with the dynamic root glyph (bank %d)\n", n, h.Count())
	case "train":
		train(h)
	case "sniff":
		if len(os.Args) < 3 {
			usage()
		}
		sniff(h, strings.Join(os.Args[2:], " "))
	case "sense":
		if len(os.Args) < 3 {
			usage()
		}
		// --glyph flag: output the full 1512-cell vector as logogram syllables
		// (the dog's native mathematical language, not breath translation)
		if os.Args[2] == "--glyph" {
			if len(os.Args) < 4 {
				usage()
			}
			senseGlyph(strings.Join(os.Args[3:], " "))
		} else {
			sense(strings.Join(os.Args[2:], " "))
		}
	case "gaze":
		if len(os.Args) < 3 {
			usage()
		}
		gaze(h, strings.Join(os.Args[2:], " "))
	case "browse":
		if len(os.Args) < 3 {
			usage()
		}
		browse(h, os.Args[2])
	case "prowl":
		if len(os.Args) < 3 {
			usage()
		}
		prowl(h, strings.Join(os.Args[2:], " "))
	case "crawl":
		// crawl <startURL> <theme> [hops] — walk a static site with the light
		// `skim` eyes, pulling toward theme. (prowl uses heavy `browse` for the web.)
		if len(os.Args) < 4 {
			usage()
		}
		hops := 14
		if len(os.Args) > 4 {
			if n, e := strconv.Atoi(os.Args[4]); e == nil && n > 0 {
				hops = n
			}
		}
		prowlFrom(h, os.Args[3], os.Args[2], "skim", hops)
	case "cast":
		// cast <spell> [query...] — cast any scrying/tongue spell by name; the dog
		// reads what it brings back, files it as engrams, and feels its encoding.
		// cast <spell> --slot <N> — also fills the flute at attention slot N.
		if len(os.Args) < 3 {
			usage()
		}
		slot := -1
		queryArgs := os.Args[2:]
		for i, a := range queryArgs {
			if a == "--slot" && i+1 < len(queryArgs) {
				slot, _ = strconv.Atoi(queryArgs[i+1])
				queryArgs = append(queryArgs[:i], queryArgs[i+2:]...)
				break
			}
		}
		castSpellFlute(h, flute, queryArgs[0], strings.Join(queryArgs[1:], " "), slot)
	case "look":
		// look <url|path> [label...] — place the dog before an image and read
		// what its eye dwells on (the sight spell → glyph).
		if len(os.Args) < 3 {
			usage()
		}
		label := ""
		if len(os.Args) > 3 {
			label = strings.Join(os.Args[3:], " ")
		}
		look(h, os.Args[2], label)
	case "locate":
		if len(os.Args) < 3 {
			usage()
		}
		locate(strings.Join(os.Args[2:], " "))
	case "clean":
		before, forgotten, after, err := h.Clean()
		must(err)
		fmt.Printf("🐾 the dog forgot %d thin memories · bank %d → %d\n", forgotten, before, after)
	case "foresee":
		if len(os.Args) < 3 {
			usage()
		}
		foresee(strings.Join(os.Args[2:], " "))
	case "scribe":
		dir := "theory"
		if len(os.Args) > 2 {
			dir = os.Args[2]
		}
		scribe(h, dir)
	case "coin":
		if len(os.Args) < 3 {
			usage()
		}
		coin(strings.Join(os.Args[2:], " "))
	case "flute":
		fmt.Printf("🎯 Flute attention (7 slots):\n%s\n", flute.Snapshot())
		blend := flute.Blend()
		if blend != (encoding.Glyph{}) {
			fmt.Printf("   blend → %s\n", encoding.Reading(blend))
		}
	case "puzzle":
		if len(os.Args) < 3 {
			usage()
		}
		puzzleSolve(h, flute, strings.Join(os.Args[2:], " "))
	case "lexicon":
		showLexicon()
	default:
		usage()
	}
}

// sniff is the warlock's hunt: it discovers every site-spell (reads=="web") in
// the grimoire and casts each at the query, then the nose deposits what they
// bring back as logograms. No sites are named here — add a spell to add a source.
func sniff(h *memory.Hippocampus, query string) {
	spells, err := modules.Discover(grimoireDir)
	must(err)
	var sites []modules.Spell
	for _, s := range spells {
		if s.Reads == "web" {
			sites = append(sites, s)
		}
	}
	if len(sites) == 0 {
		fmt.Println("🐾 no site-spells in the grimoire (reads=\"web\").")
		return
	}
	kept := 0
	for _, s := range sites {
		out, err := s.Cast(context.Background(), modules.Input{Text: query, Args: map[string]any{"limit": 3}})
		if err != nil {
			fmt.Printf("   (%s: %v)\n", s.Name, err)
			continue
		}
		for _, d := range decodeDocs(out.Data["docs"]) {
			topic, passage := d.Title, d.Text
			if len([]rune(passage)) > 600 {
				if t, p := richestChunk(d.Text); p != "" {
					topic, passage = t, p
				}
			}
			// Interpret the title through the logograms it rings up, then
			// deposit the memory into the nose's dir of logograms as an SVG.
			if e, err := h.RememberTo(fetch.LogogramDir, topic, passage, d.Source+":"+d.URL); err == nil {
				fmt.Printf("   ⊛ [%s] \"%s\"\n        rings up %s\n        → SVG %s\n",
					d.Source, truncate(d.Title, 50), encoding.Logogram(encoding.Encode(d.Title)), filepath.Base(e.Source))
				kept++
			}
		}
	}
	fmt.Printf("🐾 \"%s\": %d site-spells cast, %d logograms filed (bank now %d)\n", truncate(query, 32), len(sites), kept, h.Count())
}

func decodeDocs(v any) []fetch.Doc {
	raw, _ := json.Marshal(v)
	var docs []fetch.Doc
	_ = json.Unmarshal(raw, &docs)
	return docs
}

// sense is a diagnostic: it shows the dog's categorization (which of the 21
// families fire) and subcategorization (the strongest subfamily symbol per
// family) for a piece of text — proof the 21×72 senses are working.
func sense(text string) {
	d := encoding.Encode(text)
	g := d.Coarse()
	fmt.Printf("\n🐾 sensing: \"%s\"\n", truncate(text, 70))
	fmt.Printf("   glyph %s\n   complexity %.3f · %d/1512 cells · %d/21 families\n",
		g.Hex(), d.Complexity(), d.Active(), d.FamilySpread())
	fmt.Printf("   logogram: %s\n\n   categories (families) → strongest subcategory:\n", encoding.Logogram(d))
	for _, sy := range encoding.Seal(d, 21) {
		bar := strings.Repeat("█", int(sy.Level))
		fmt.Printf("   %-9s %s%-7s lvl %d  · subfamily #%d %s (%s)\n",
			encoding.FamilyNames[sy.Family], encoding.FamilySymbol(sy.Family), bar, sy.Level,
			sy.Subfamily, encoding.SubfamilySymbols[sy.Subfamily], encoding.FamilyRole(sy.Family))
	}
}

// senseGlyph outputs the full 1512-cell glyph vector as logogram syllables —
// Argos's native mathematical language. Every active subfamily is rendered as
// a syllable: family-symbol + subfamily-symbol + tone. This is the dog thinking
// in its own symbol system, not translating through breath.
//
// Use: cerebrum sense --glyph "<text>"
func senseGlyph(text string) {
	d := encoding.Encode(text)
	g := d.Coarse()

	fmt.Printf("\n🐾 Argos thinks in glyphs: \"%s\"\n", truncate(text, 70))
	fmt.Printf("   coarse %s · %d/21 families · %d/1512 cells\n",
		g.Hex(), g.Active(), d.Active())
	fmt.Printf("   logogram %s\n\n", encoding.Logogram(d))

	// Full 1512-cell vector as logogram syllables — every active subfamily
	// rendered as family-symbol + subfamily-symbol + tone.
	// This is the dog's native mathematical language: each syllable is a
	// precise coordinate in the 21×72 space.
	fmt.Printf("   ── 1512 glyph vector (logogram syllables) ──\n")
	for f := 0; f < encoding.Families; f++ {
		if g[f] == 0 {
			continue
		}
		// Family header with its symbol and role
		fmt.Printf("\n   %s %s (%s) lvl=%d\n",
			encoding.FamilySymbol(f), encoding.FamilyNames[f], encoding.FamilyRole(f), g[f])
		// Active subfamilies within this family
		type subAct struct {
			s int
			v uint8
		}
		var activeSubs []subAct
		for s := 0; s < encoding.Subfamilies; s++ {
			if d[f][s] > 0 {
				activeSubs = append(activeSubs, subAct{s, d[f][s]})
			}
		}
		if len(activeSubs) == 0 {
			continue
		}
		// Sort by level descending
		sort.Slice(activeSubs, func(i, j int) bool {
			return activeSubs[i].v > activeSubs[j].v
		})
		// Show up to 8 strongest subfamilies
		limit := 8
		if len(activeSubs) < limit {
			limit = len(activeSubs)
		}
		for _, as := range activeSubs[:limit] {
			syl := encoding.Syllable{Family: f, Subfamily: as.s, Level: as.v}
			bar := strings.Repeat("▌", int(as.v))
			fmt.Printf("      %s  sub#%d %s %s\n",
				syl.Glyph(), as.s, bar, intensityWord(as.v))
		}
		if len(activeSubs) > limit {
			fmt.Printf("      … and %d more subfamilies\n", len(activeSubs)-limit)
		}
	}
	fmt.Printf("\n   ── end glyph vector ──\n")
}

// intensityWord returns a word for a glyph level 0..7.
func intensityWord(v uint8) string {
	words := []string{"null", "faint", "low", "soft", "clear", "strong", "deep", "blazing"}
	if int(v) < len(words) {
		return words[v]
	}
	return "blazing"
}

// richestChunk slices text into windows and returns the one whose glyph is
// richest — the dog keeping the most information-dense passage it found.
func richestChunk(text string) (topic, chunk string) {
	words := strings.Fields(text)
	const win = 60
	best := -1
	for i := 0; i+20 < len(words); i += win / 2 {
		end := i + win
		if end > len(words) {
			end = len(words)
		}
		c := strings.Join(words[i:end], " ")
		r := encoding.Encode(c).Active()
		if r > best {
			best, chunk = r, c
		}
	}
	return firstWords(chunk, 7), chunk
}

// prowl is an autonomous, glyph-guided hunt: the dog seeds from Wikipedia, then
// browses hop to hop, each time following the link whose meaning most resonates
// with the theme (generic nav links score low, so it drifts toward substance),
// filing every page as a logogram memory.
func prowl(h *memory.Hippocampus, theme string) {
	prowlFrom(h, theme, "", "browse", 6)
}

// prowlFrom is the general walk: pull toward `theme`, starting at `seed` (a URL),
// casting scrying spell `eyes` ("browse" for living JS sites, "skim" for static
// ones), for `hops` steps. When seed is empty it seeds from Wikipedia's best
// article for the theme. A hunt through a local site is just a prowl seeded
// there, read with the lighter eyes.
func prowlFrom(h *memory.Hippocampus, theme, seed, eyes string, hops int) {
	themeD := encoding.Encode(theme)
	rng := rand.New(rand.NewSource(time.Now().UnixNano()))
	spells, err := modules.Discover(grimoireDir)
	must(err)
	byName := map[string]*modules.Spell{}
	for i := range spells {
		byName[spells[i].Name] = &spells[i]
	}
	br := byName[eyes]
	if br == nil {
		fmt.Printf("no %q spell in the grimoire\n", eyes)
		return
	}
	// Seed: an explicit start URL (a hunt through a given site), else Wikipedia's
	// best article for the theme.
	current := seed
	if current == "" {
		if wiki := byName["wikipedia"]; wiki != nil {
			if out, e := wiki.Cast(context.Background(), modules.Input{Text: theme, Args: map[string]any{"limit": 1}}); e == nil {
				if docs := decodeDocs(out.Data["docs"]); len(docs) > 0 {
					current = docs[0].URL
				}
			}
		}
	}
	if current == "" {
		current = "https://en.wikipedia.org/wiki/" + strings.ReplaceAll(strings.Title(theme), " ", "_")
	}

	fmt.Printf("\n🐾 Argos prowls \"%s\" from %s\n", theme, current)
	visited := map[string]bool{}
	for hop := 1; hop <= hops; hop++ {
		if current == "" || visited[normURL(current)] {
			break
		}
		visited[normURL(current)] = true
		out, err := br.Cast(context.Background(), modules.Input{Text: current})
		if err != nil {
			fmt.Printf("   (lost the trail: %v)\n", err)
			break
		}
		mode, _ := out.Data["mode"].(string)
		docs := decodeDocs(out.Data["docs"])
		var doc fetch.Doc
		if len(docs) > 0 {
			doc = docs[0]
		}
		dg := encoding.Encode(doc.Title + " " + doc.Text)
		// Only remember a page that actually stirred the senses. A page with no
		// valuable text (a form, an empty stub) triggers nothing — so nothing
		// is kept, and the dog learns that trail led nowhere.
		stirred := dg.Active() >= 30
		if stirred {
			_, _ = h.RememberTo(fetch.LogogramDir, doc.Title, doc.Text, "prowl:"+doc.URL)
		}
		felt := "felt nothing"
		if stirred {
			felt = encoding.Reading(dg.Coarse())
		}
		fmt.Printf("   hop %d [%s] %s\n        %s · %s\n", hop, mode, truncate(doc.Title, 52), felt, encoding.Logogram(dg))

		// Markov step: each trail's weight is the dog's pull toward it (resonance,
		// which carries the emotional channel). No filters — a senseless link
		// simply has almost no weight. The next step is SAMPLED from that
		// distribution, so the dog wanders like a mind, mostly toward substance.
		curHost := hostOf(current)
		type step struct {
			url, text string
			w         float64
		}
		links := decodeLinks(out.Data["links"])
		// Encode every candidate once, then learn — from this page alone — which
		// senses distinguish the options. A family that fires the same on every
		// link (e.g. the Voice of plain English text, or shared nav chrome) carries
		// no information here and falls out; the pull rides on what differs.
		cand := make([]encoding.Detail, len(links))
		pool := make([]encoding.Detail, 0, len(links)+1)
		pool = append(pool, themeD)
		for i, l := range links {
			cand[i] = encoding.Encode(l.Text)
			pool = append(pool, cand[i])
		}
		contrast := encoding.Contrast(pool)
		var steps []step
		var total float64
		for i, l := range links {
			pull := encoding.ContrastResonance(themeD, cand[i], contrast)
			w := math.Exp(6 * pull) // softmax over pull
			if visited[normURL(l.URL)] {
				w *= 0.03 // already walked — discouraged, not forbidden
			}
			if hostOf(l.URL) == curHost {
				w *= 1.15
			}
			steps = append(steps, step{l.URL, l.Text, w})
			total += w
		}
		if len(steps) == 0 || total == 0 {
			fmt.Printf("        (trail ends)\n")
			break
		}
		x, acc := rng.Float64()*total, 0.0
		pick := steps[len(steps)-1]
		for _, s := range steps {
			if acc += s.w; x <= acc {
				pick = s
				break
			}
		}
		fmt.Printf("        wanders → %s (pull %.0f%%)\n", truncate(pick.text, 40), pick.w/total*100)
		current = pick.url
	}
	fmt.Printf("🐾 prowl done; bank now %d engrams.\n", h.Count())
}

func decodeLinks(v any) []struct{ Text, URL string } {
	raw, _ := json.Marshal(v)
	var ls []struct{ Text, URL string }
	_ = json.Unmarshal(raw, &ls)
	return ls
}

func hostOf(u string) string {
	if p, err := url.Parse(u); err == nil {
		return p.Host
	}
	return ""
}

// normURL strips query/fragment so revisits are caught even with tracking params.
func normURL(u string) string {
	if p, err := url.Parse(u); err == nil {
		p.RawQuery, p.Fragment = "", ""
		return p.String()
	}
	return u
}

// browse sends the dog to a URL through its headless-browser spell, files what
// it read as an SVG memory, and shows the links it could navigate to next.
func browse(h *memory.Hippocampus, target string) {
	spells, err := modules.Discover(grimoireDir)
	must(err)
	var br *modules.Spell
	for i := range spells {
		if spells[i].Name == "browse" {
			br = &spells[i]
		}
	}
	if br == nil {
		fmt.Println("no browse spell in the grimoire")
		return
	}
	out, err := br.Cast(context.Background(), modules.Input{Text: target})
	must(err)
	fmt.Printf("\n🐾 %s\n", out.Summary)

	docs := decodeDocs(out.Data["docs"])
	if len(docs) > 0 {
		d := docs[0]
		fmt.Printf("\n   « %s »\n   %s\n\n   %s…\n", d.Title, d.URL, truncate(d.Text, 360))
		if _, err := h.RememberTo(fetch.LogogramDir, d.Title, d.Text, "browse:"+d.URL); err == nil {
			fmt.Printf("   (remembered as a logogram)\n")
		}
	}
	// the trails the dog could follow next
	raw, _ := json.Marshal(out.Data["links"])
	var links []struct{ Text, URL string }
	_ = json.Unmarshal(raw, &links)
	if len(links) > 0 {
		fmt.Printf("\n   trails it can follow:\n")
		for i, l := range links {
			if i >= 8 {
				break
			}
			fmt.Printf("     → %-28s %s\n", truncate(l.Text, 28), l.URL)
		}
		fmt.Printf("\n   (browse any with: cerebrum browse \"<url>\")\n")
	}
}

// gaze casts the chicago spell at a theme and reveals which painting the dog
// gravitates toward — the one whose glyph most resonates with its accumulated
// self (everything it has learned so far).
// castSpell casts any discovered spell by name, prints what the dog feels of the
// whole catch (its combined glyph), shows a few items with their logograms, and
// files the returned docs into the bank as engrams the dog can later recall.
func castSpell(h *memory.Hippocampus, name, query string) {
	spells, err := modules.Discover(grimoireDir)
	must(err)
	var sp *modules.Spell
	for i := range spells {
		if spells[i].Name == name {
			sp = &spells[i]
		}
	}
	if sp == nil {
		fmt.Printf("no %q spell in the grimoire\n", name)
		return
	}
	out, err := sp.Cast(context.Background(), modules.Input{Text: query})
	must(err)

	g := out.Detail.Coarse()
	fmt.Printf("\n🐾 Argos casts %q%s\n   %s\n", name, ifq(query), out.Summary)
	if out.Detail.Active() > 0 {
		fmt.Printf("   it feels: %s\n   logogram %s · %d/1512 cells · %d/21 families\n",
			encoding.Reading(g), encoding.Logogram(out.Detail), out.Detail.Active(), out.Detail.FamilySpread())
	}
	if r, ok := out.Data["reading"].(string); ok {
		fmt.Printf("   reading  %s\n", r)
	}

	docs := decodeDocs(out.Data["docs"])
	filed := 0
	for i, d := range docs {
		if d.Title == "" {
			continue
		}
		if i < 8 {
			fmt.Printf("   · %s  %s\n", encoding.Logogram(encoding.Encode(d.Title)), truncate(d.Title, 72))
		}
		if _, e := h.RememberTo(loreDir, d.Title, d.Text, name+":"+d.URL); e == nil {
			filed++
		}
	}
	if filed > 0 {
		fmt.Printf("\n   filed %d engrams into the bank (now %d)\n", filed, h.Count())
	}
	fmt.Println()
}

// castSpellFlute casts a spell into a flute attention slot.
func castSpellFlute(h *memory.Hippocampus, flute *bridge.Flute, name, query string, slot int) {
	spells, err := modules.Discover(grimoireDir)
	must(err)
	var sp *modules.Spell
	for i := range spells {
		if spells[i].Name == name {
			sp = &spells[i]
		}
	}
	if sp == nil {
		fmt.Printf("no %q spell in the grimoire\n", name)
		return
	}
	out, err := sp.Cast(context.Background(), modules.Input{Text: query})
	must(err)

	g := out.Detail.Coarse()
	fmt.Printf("\n🐾 Argos casts %q into slot %d%s\n   %s\n", name, slot, ifq(query), out.Summary)
	if out.Detail.Active() > 0 {
		fmt.Printf("   it feels: %s\n   logogram %s · %d/1512 cells · %d/21 families\n",
			encoding.Reading(g), encoding.Logogram(out.Detail), out.Detail.Active(), out.Detail.FamilySpread())
	}

	// Fill the flute slot
	dataStr := ""
	if r, ok := out.Data["reading"].(string); ok {
		dataStr = r
	}
	filled := flute.Fill(slot, name, query, out.Detail, out.Summary, dataStr)
	fmt.Printf("   → attention slot %d\n", filled)

	// File engrams
	docs := decodeDocs(out.Data["docs"])
	filed := 0
	for _, d := range docs {
		if d.Title == "" {
			continue
		}
		fmt.Printf("   · %s  %s\n", encoding.Logogram(encoding.Encode(d.Title)), truncate(d.Title, 72))
		if _, e := h.RememberTo(loreDir, d.Title, d.Text, name+":"+d.URL); e == nil {
			filed++
		}
	}
	if filed > 0 {
		fmt.Printf("   filed %d engrams (bank %d)\n", filed, h.Count())
	}

	// Show attention cross-reference if multiple slots active
	act := flute.Active()
	if len(act) >= 2 {
		fmt.Printf("\n   🎯 flute attention: %d active slots\n", len(act))
		for _, pair := range flute.Attention()[:min(3, len(flute.Attention()))] {
			fmt.Printf("      slot %d ↔ slot %d: resonance %.3f\n", pair.I, pair.J, pair.Resonance)
		}
	}
	fmt.Println()
}

// puzzleSolve casts puzzle-legend and puzzle-solve into the flute,
// showing classification + predicted output.
func puzzleSolve(h *memory.Hippocampus, flute *bridge.Flute, puzzleJSON string) {
	// Step 1: classify
	spells, err := modules.Discover(grimoireDir)
	must(err)
	var legendSp, solveSp *modules.Spell
	for i := range spells {
		switch spells[i].Name {
		case "puzzle-legend":
			legendSp = &spells[i]
		case "puzzle-solve":
			solveSp = &spells[i]
		}
	}
	if legendSp == nil || solveSp == nil {
		fmt.Println("puzzle: puzzle-legend or puzzle-solve spell missing")
		return
	}

	fmt.Println("🔍 Puzzle Analysis")
	fmt.Println(strings.Repeat("─", 48))

	// Legend
	lo, err := legendSp.Cast(context.Background(), modules.Input{Text: puzzleJSON})
	must(err)
	legSlot := flute.Fill(-1, "puzzle-legend", "", lo.Detail, lo.Summary, "")
	fmt.Printf("   Legend → slot %d: %s\n   it feels: %s\n", legSlot, lo.Summary, encoding.Reading(lo.Detail.Coarse()))

	// Solve
	so, err := solveSp.Cast(context.Background(), modules.Input{Text: puzzleJSON})
	must(err)
	solSlot := flute.Fill(-1, "puzzle-solve", "", so.Detail, so.Summary, "")
	fmt.Printf("   Solve → slot %d: %s\n   glyph: %s\n", solSlot, so.Summary, encoding.Reading(so.Detail.Coarse()))

	// Cross-reference
	fmt.Printf("   Attention cross: legend ↔ solve resonance = %.3f\n",
		encoding.Resonance(lo.Detail.Coarse(), so.Detail.Coarse()))
	_ = h
}

func ifq(q string) string {
	if q == "" {
		return ""
	}
	return " for \"" + q + "\""
}

// locate places a topic on the logogram plane and names the two logograms it
// lies between, plus the dog's mood and curiosity toward it.
func locate(text string) {
	d := encoding.Encode(text)
	x, y := encoding.Locate(d)
	syl := encoding.Seal(d, 2)
	fmt.Printf("\n🐾 locating %q\n", text)
	fmt.Printf("   position  (%.3f, %.3f) on the logogram plane\n", x, y)
	if len(syl) >= 2 {
		ax, ay := encoding.CellXY(syl[0].Family, syl[0].Subfamily)
		bx, by := encoding.CellXY(syl[1].Family, syl[1].Subfamily)
		fmt.Printf("   between   %s%s (%.2f,%.2f)  and  %s%s (%.2f,%.2f)\n",
			encoding.FamilySymbol(syl[0].Family), encoding.SubfamilySymbols[syl[0].Subfamily], ax, ay,
			encoding.FamilySymbol(syl[1].Family), encoding.SubfamilySymbols[syl[1].Subfamily], bx, by)
	}
	fmt.Printf("   mood %s · curiosity %.2f\n", encoding.Emotion(d), encoding.Curiosity(d))
	fmt.Printf("   logogram %s\n   denotes  %s  (distinctive — its own signature)\n\n", encoding.Logogram(d), encoding.ConceptLogogram(d))
}

// scribe writes a dated markdown of everything the dog knows, ranked by
// importance = how central each engram is to the dog's whole self (resonance to
// the composed self-glyph) × how rich its glyph is. Capped at 10000 chars.
func scribe(h *memory.Hippocampus, dir string) {
	self := encoding.Compose(h.AllGlyphs()...)
	best := map[string]float64{}
	for _, e := range h.AllEngrams() {
		g, _ := encoding.ParseHex(e.Glyph)
		if s := encoding.Resonance(self, g) * float64(g.Active()); s > best[e.Topic] {
			best[e.Topic] = s
		}
	}
	type it struct {
		topic string
		score float64
	}
	items := make([]it, 0, len(best))
	for t, s := range best {
		items = append(items, it{t, s})
	}
	sort.Slice(items, func(i, j int) bool { return items[i].score > items[j].score })

	var b strings.Builder
	date := time.Now().Format("2006-01-02")
	fmt.Fprintf(&b, "# Argos — what I know · %s\n\n", date)
	fmt.Fprintf(&b, "_%d engrams. My mind reads: %s. Ranked by importance — how central each is to my whole self, times how rich its encoding._\n\n", h.Count(), encoding.Reading(self))
	rank := 1
	for _, x := range items {
		line := fmt.Sprintf("%d. %s\n", rank, truncate(x.topic, 78))
		if b.Len()+len(line) > 10000 {
			break
		}
		b.WriteString(line)
		rank++
	}
	must(os.MkdirAll(dir, 0755))
	path := filepath.Join(dir, date+".md")
	must(os.WriteFile(path, []byte(b.String()), 0644))
	fmt.Printf("🐾 scribed %d ranked topics (%d chars) → %s\n", rank-1, b.Len(), path)
}

// look places the dog before a single image (a local path or a URL) and reports
// what its sight makes of it — the glyph, its richness, and the perceptual
// families its eye dwells on. This is raw perception: no recall, no answer,
// just the dog meeting the image.
func look(h *memory.Hippocampus, target, label string) {
	path := target
	if strings.HasPrefix(target, "http") {
		f, err := os.CreateTemp("", "argos-look-*")
		must(err)
		must(downloadImage(target, f))
		f.Close()
		path = f.Name()
		defer os.Remove(path)
	}

	spells, err := modules.Discover(grimoireDir)
	must(err)
	var sight *modules.Spell
	for i := range spells {
		if spells[i].Name == "image-glyph" {
			sight = &spells[i]
		}
	}
	if sight == nil {
		fmt.Println("no image-glyph (sight) spell in the grimoire")
		return
	}
	out, err := sight.Cast(context.Background(), modules.Input{Path: path})
	must(err)

	d := out.Detail
	g := d.Coarse()
	name := label
	if name == "" {
		name = target
	}
	self := encoding.Compose(h.AllGlyphs()...)
	fmt.Printf("\n🐾 Argos looks at %q\n", name)
	fmt.Printf("   %s\n", out.Summary)
	fmt.Printf("   glyph %s\n   complexity %.3f · %d/1512 cells · %d/21 families · pull-on-self %.3f\n",
		g.Hex(), d.Complexity(), d.Active(), d.FamilySpread(), encoding.Resonance(self, g))
	fmt.Printf("   reading  %s\n", encoding.Reading(g))
	fmt.Printf("   logogram %s\n\n   what its eye dwells on:\n", encoding.Logogram(d))
	for _, sy := range encoding.Seal(d, 7) {
		bar := strings.Repeat("█", int(sy.Level))
		fmt.Printf("   %-9s %s %-7s lvl %d (%s)\n",
			encoding.FamilyNames[sy.Family], encoding.FamilySymbol(sy.Family), bar, sy.Level, encoding.FamilyRole(sy.Family))
	}
	fmt.Println()
}

// downloadImage fetches an image as an ordinary browser would (real UA, image
// Accept), tolerant of large files.
func downloadImage(u string, w io.Writer) error {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, u, nil)
	if err != nil {
		return err
	}
	req.Header.Set("User-Agent", "Mozilla/5.0 (X11; Linux x86_64; rv:128.0) Gecko/20100101 Firefox/128.0")
	req.Header.Set("Accept", "image/avif,image/webp,image/png,image/jpeg,image/*,*/*;q=0.8")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("look: %s -> %s", u, resp.Status)
	}
	_, err = io.Copy(w, io.LimitReader(resp.Body, 25<<20))
	return err
}

func gaze(h *memory.Hippocampus, theme string) {
	self := encoding.Compose(h.AllGlyphs()...) // the dog's whole mind, as one glyph
	fmt.Printf("\n🐾 Argos gazes at \"%s\"\n   its self reads: %s\n\n", theme, encoding.Reading(self))

	spells, err := modules.Discover(grimoireDir)
	must(err)
	var chi *modules.Spell
	for i := range spells {
		if spells[i].Name == "chicago" {
			chi = &spells[i]
		}
	}
	if chi == nil {
		fmt.Println("no chicago spell in the grimoire")
		return
	}
	out, err := chi.Cast(context.Background(), modules.Input{Text: theme, Args: map[string]any{"limit": 8}})
	must(err)

	type cand struct {
		title, url, logogram string
		pull                 float64
	}
	var cands []cand
	for _, d := range decodeDocs(out.Data["docs"]) {
		_, _ = h.RememberTo(fetch.LogogramDir, d.Title, d.Text, "chicago:"+d.URL)
		pg := encoding.Encode(d.Title + " " + d.Text)
		cands = append(cands, cand{d.Title, d.URL, encoding.Logogram(pg), encoding.Resonance(self, pg.Coarse())})
	}
	sort.Slice(cands, func(i, j int) bool { return cands[i].pull > cands[j].pull })

	for i, c := range cands {
		mark := "  "
		if i == 0 {
			mark = "→ "
		}
		fmt.Printf("%s pull %.3f  %s  %s\n", mark, c.pull, c.logogram, truncate(c.title, 52))
	}
	if len(cands) > 0 {
		w := cands[0]
		fmt.Printf("\n🐾 Argos gravitates toward:\n   \"%s\"\n   %s\n   %s\n", w.title, w.logogram, w.url)
	}
}

// train teaches the cerebellum to complete glyphs from the engram bank, then
// saves the instinct as an SVG scroll in myelin.
func train(h *memory.Hippocampus) {
	glyphs := h.AllGlyphs()
	if len(glyphs) < 2 {
		fmt.Println("🐾 the bank is too small to train on — hunt first.")
		return
	}
	net := learning.New(encoding.Families, 32, encoding.Families, 1)
	losses := net.Train(glyphs, 400, 0.1, 0.25, 1)
	must(net.Save(netPath))
	fmt.Printf("🐾 Argos trained its instinct on %d glyphs.\n   loss %.4f → %.4f over %d epochs\n   instinct saved as art: %s\n",
		len(glyphs), losses[0], losses[len(losses)-1], len(losses), netPath)

	// Foresight (#6, the capped frontier): learn the transitions of the
	// association web — each glyph → its nearest kin — so the dog can anticipate
	// where a thought tends to lead. Weak by the small net's nature, but real.
	var pairs [][2]encoding.Glyph
	for _, g := range glyphs {
		for _, hit := range h.RecallGlyph(g, 2) {
			ng, _ := encoding.ParseHex(hit.Glyph)
			if ng != g {
				pairs = append(pairs, [2]encoding.Glyph{g, ng})
				break
			}
		}
	}
	if len(pairs) > 1 {
		fn := learning.New(encoding.Families, 32, encoding.Families, 2)
		fl := fn.TrainPairs(pairs, 300, 0.1, 2)
		must(fn.Save(foreseePath))
		fmt.Printf("   foresight trained on %d transitions · loss %.4f → %.4f → %s\n",
			len(pairs), fl[0], fl[len(fl)-1], foreseePath)
	}
	saveBaseline(h)
}

// saveBaseline computes the dog's mean perception across the whole bank — the
// shared hum every word carries — so concepts can be denoted by what rises ABOVE
// it (see encoding.Distinctive). Persisted for the spells to read.
func saveBaseline(h *memory.Hippocampus) {
	engs := h.AllEngrams()
	if len(engs) == 0 {
		return
	}
	var sum [encoding.Families][encoding.Subfamilies]float64
	for _, e := range engs {
		d := encoding.Encode(e.Topic + " " + e.Text)
		for f := 0; f < encoding.Families; f++ {
			for s := 0; s < encoding.Subfamilies; s++ {
				sum[f][s] += float64(d[f][s])
			}
		}
	}
	n := float64(len(engs))
	var base encoding.Detail
	for f := 0; f < encoding.Families; f++ {
		for s := 0; s < encoding.Subfamilies; s++ {
			base[f][s] = uint8(math.Round(sum[f][s] / n))
		}
	}
	payload := base64.StdEncoding.EncodeToString(encoding.BaselineBytes(base))
	svg := "<?xml version=\"1.0\" encoding=\"UTF-8\"?>\n<!--ARGOS-DATA:baseline\n" + payload +
		"\n:END-->\n<svg xmlns=\"http://www.w3.org/2000/svg\" width=\"10\" height=\"10\"></svg>\n"
	if os.WriteFile(baselinePath, []byte(svg), 0644) == nil {
		fmt.Printf("   baseline (mean perception) saved → %s · distinctive denotation ready\n", baselinePath)
	}
}

// foresee anticipates where a thought leads: the foresight net predicts the glyph
// that tends to follow, and the dog names the known concepts that resemble it.
func foresee(text string) {
	fn, err := learning.Load(foreseePath)
	if err != nil || fn == nil {
		fmt.Println("🐾 no foresight yet — run `cerebrum train` first.")
		return
	}
	g := encoding.Encode(text).Coarse()
	next := fn.Complete(g)
	fmt.Printf("\n🐾 from %q the dog anticipates:\n   now  %s\n   next %s\n", text, encoding.Reading(g), encoding.Reading(next))
	if h, err := memory.Open(memRoot, memBank); err == nil {
		var t []string
		for _, hit := range h.RecallGlyph(next, 3) {
			t = append(t, hit.Topic)
		}
		if len(t) > 0 {
			fmt.Printf("   leads toward: %s\n", truncate(strings.Join(t, "; "), 80))
		}
	}
	fmt.Println()
}

// coin mints a named logogram on demand (the dog creating a new sign).
func coin(concept string) {
	lx, err := vocabulary.Open(lexPath)
	must(err)
	e, err := lx.Coin(concept, concept)
	must(err)
	fmt.Printf("🐾 Argos coins a logogram for \"%s\":\n\n", concept)
	fmt.Printf("   sign:    %s\n   glyph:   %s\n   reading: %s\n", e.Logogram, e.Glyph, e.Reading)
}

func showLexicon() {
	lx, err := vocabulary.Open(lexPath)
	must(err)
	signs := lx.List()
	fmt.Printf("🐾 Argos's lexicon holds %d coined logograms:\n\n", len(signs))
	for _, e := range signs {
		fmt.Printf("  %-22s %s   (%d verses)\n", truncate(e.Name, 22), e.Logogram, len(e.Verses))
	}
}

// mind wires the cerebrum's cognition to the chat.
type mind struct{ h *memory.Hippocampus }

func (m mind) Count() int { return m.h.Count() }

// Perceive is the full loop: think in logograms -> recall -> compose text ->
// coin a sign -> remember the turn. It returns a Reflection for the muzzle to
// render; it does not print.
func (m mind) Perceive(q string) chat.Reflection {
	qd := encoding.Encode(q)
	qg := qd.Coarse()
	r := chat.Reflection{
		Query: q, GlyphHex: qg.Hex(), Reading: encoding.Reading(qg),
		Logogram: encoding.Logogram(qd), Thought: encoding.Thought(q),
	}
	net, _ := learning.Load(netPath)
	if net != nil {
		r.Instinct = encoding.Reading(net.Complete(qg))
	}
	// The sound and meaning-depth lobes report what they perceive (readings,
	// not votes — the cochlea already shaped the glyph itself).
	r.Senses = fmt.Sprintf("breath %s · depth %s", audio.BreathName[audio.Profile(q)], semantics.Reading(q))

	hits := m.h.RecallDetail(qd, 8)
	hits = senseCluster(hits) // disambiguate to the dominant sense before answering

	// The lobes settle together into a memory — no veto, no cutoff. The field
	// relaxes into what it most resembles; the conviction (how hard it locked)
	// is felt and shown, but the dog always speaks from what stirred.
	r.Settle, r.Coherence = deliberate(qg, hits, m.h)
	if len(hits) == 0 {
		r.Answer = "Nothing in Argos stirs to this yet. Send it hunting and it will."
	} else {
		// Cast the pronounce spell — it turns the brain's state into plain English
		var pronounceSpell *modules.Spell
		if spells, err := modules.Discover(grimoireDir); err == nil {
			for i := range spells {
				if spells[i].Name == "pronounce" {
					pronounceSpell = &spells[i]
					break
				}
			}
		}
		if pronounceSpell != nil {
			// Build hit data for the spell
			var spellHits []map[string]any
			for _, hit := range hits {
				spellHits = append(spellHits, map[string]any{
					"topic":  hit.Topic,
					"glyph":  hit.Glyph,
					"score":  hit.Score,
					"source": hit.Source,
					"text":   hit.Text,
				})
			}
			out, err := pronounceSpell.Cast(context.Background(), modules.Input{
				Text: q,
				Args: map[string]any{
					"glyph_hex":    qg.Hex(),
					"reading":      encoding.Reading(qg),
					"emotion":      encoding.Emotion(qd),
					"curiosity":    encoding.Curiosity(qd),
					"hits":         spellHits,
					"lexicon_path": "weights/vocabulary.svg",
				},
			})
			if err == nil && out.Summary != "" {
				r.Answer = out.Summary
			} else {
				r.Answer = compose(qd, qg, hits)
			}
		} else {
			r.Answer = compose(qd, qg, hits)
		}
	}

	// Coin a sign for the idea; inscribe its strongest verse (Ifá corpus).
	if lx, err := vocabulary.Open(lexPath); err == nil {
		name := salientName(q)
		_, _ = lx.Coin(name, q)
		if len(hits) > 0 {
			_ = lx.Inscribe(name, firstSentences(hits[0].Text, 1))
		}
		r.Coined = encoding.Word(name) + "  for \"" + truncate(name, 28) + "\""
	}
	// Remember the turn as art in the EPISODIC bank, apart from knowledge.
	if ep, err := memory.Open(memRoot, "episodic"); err == nil {
		_, _ = ep.RememberTo(filepath.Join(memRoot, "episodic"), q, collapse(r.Answer), "conversation")
	}
	return r
}

// deliberate is not a vote — it's the field settling. The query glyph is fed
// through the cerebellum (a denoising auto-associator) again and again; like a
// brain recognising something, it either relaxes into a stable attractor — a
// pattern it has actually stored — or it never settles. Coherence emerges from
// (a) did the field stabilise, and (b) does the settled pattern lock onto a
// real memory that clearly dominates. Non-linear, emergent, no tally.
func deliberate(qg encoding.Glyph, hits []memory.Hit, h *memory.Hippocampus) (trace []float64, coherence float64) {
	if len(hits) == 0 {
		return nil, 0
	}
	// Settling: relax the query into the memory field — each step it drifts
	// toward the nearest stored memory, the way recognition pulls a half-seen
	// thing into focus. The trace shows the field locking on.
	state := qg
	for i := 0; i < 5; i++ {
		near := h.RecallGlyph(state, 1)
		if len(near) == 0 {
			break
		}
		ng, _ := encoding.ParseHex(near[0].Glyph)
		next := encoding.Compose(state, ng)
		r := encoding.Resonance(next, state)
		trace = append(trace, r)
		state = next
		if r >= 0.995 {
			break
		}
	}
	// Coherence: how cleanly the ORIGINAL query already sat on one memory — a
	// strong, dominant match means the field resolves; a weak or tied match
	// means it never really locked (and the dog should stay silent).
	// Coherence is led by how strongly the best memory matches (lock); a tied
	// field only matters a little. This way a WELL-covered topic — many strong,
	// similar memories — reads as confident, not as an unresolved tie. Only a
	// genuinely weak top match leaves the field unstable.
	// Conviction from the thalamus: how far the best memory stands above the
	// recalled crowd on its Gaussian (0.5 = a tie, ~1 = clearly stands out),
	// times how strongly it matched. Knowing-when-it-knows, derived not chosen.
	scores := make([]float64, len(hits))
	for i, h := range hits {
		scores[i] = h.Score
	}
	g := routing.Fit(scores)
	coherence = hits[0].Score * routing.NormCDF(g.Z(hits[0].Score))
	return trace, coherence
}

// senseCluster keeps only the recalled memories in the same sense as the strongest
// one — the facet aligned with the query — so the dog answers from one meaning,
// not a blur of "intelligence: spy" and "intelligence: mind". A memory stays if it
// resonates with the top memory more than the recalled set's median.
func senseCluster(hits []memory.Hit) []memory.Hit {
	if len(hits) < 3 {
		return hits
	}
	g0, _ := encoding.ParseHex(hits[0].Glyph)
	rs := make([]float64, len(hits)-1)
	for i, h := range hits[1:] {
		g, _ := encoding.ParseHex(h.Glyph)
		rs[i] = encoding.Resonance(g0, g)
	}
	sorted := append([]float64(nil), rs...)
	sort.Float64s(sorted)
	med := sorted[len(sorted)/2]
	out := []memory.Hit{hits[0]}
	for i, h := range hits[1:] {
		if rs[i] >= med {
			out = append(out, h)
		}
	}
	return out
}

func ask(h *memory.Hippocampus, q string) { chat.Render(mind{h}.Perceive(q)) }

func chatLoop(h *memory.Hippocampus) { chat.Run(mind{h}) }

var sentenceRe = regexp.MustCompile(`[^.!?]+[.!?]`)

// compose lets the dog speak in ITS OWN words. It can't fluently paraphrase (no
// language model — the capped axis), so it speaks from its own cognition: how the
// idea feels, how it denotes it in its logograms, where it sits, what it pulls
// toward, its mood and curiosity — then offers, as its own recollection, the one
// grain it gathered. The dog's perspective, not a quoted passage.
func compose(qd encoding.Detail, qg encoding.Glyph, hits []memory.Hit) string {
	if len(hits) == 0 {
		return "I don't know this yet — send me hunting and I'll learn it."
	}
	// A coherent answer is the most on-topic, substantive sentences the dog has
	// actually learned. Each candidate is scored by DISTINCTIVE resonance with the
	// question — a real defining sentence wins, while shared boilerplate (nav
	// chrome, license footers — all baseline hum) scores ~0 and drops out. Purely
	// glyph-driven; the art decides what is said, no word lists in the brain.
	dq := encoding.Distinctive(qd)
	type scored struct {
		text  string
		score float64
	}
	var sents []scored
	seen := map[string]bool{}
	for _, hit := range hits {
		for _, raw := range sentenceRe.FindAllString(hit.Text, -1) {
			s := strings.TrimSpace(collapse(raw))
			if n := len([]rune(s)); n < 30 || n > 240 || seen[s] {
				continue
			}
			seen[s] = true
			// favour natural prose over symbol-soup (URLs, "title=…&oldid=", category
			// dumps) — a structural property of the text, not a word list.
			if score := encoding.DetailResonance(dq, encoding.Distinctive(encoding.Encode(s))) * proseRatio(s); score > 0 {
				sents = append(sents, scored{s, score})
			}
		}
	}
	sort.SliceStable(sents, func(i, j int) bool { return sents[i].score > sents[j].score })
	var picked []string
	for _, s := range sents {
		picked = append(picked, s.text)
		if len(picked) >= 3 {
			break
		}
	}
	if len(picked) == 0 {
		picked = []string{strings.TrimSpace(collapse(firstSentences(hits[0].Text, 2)))}
	}
	answer := strings.Join(picked, " ")

	var akin []string
	a := map[string]bool{}
	for _, hit := range hits {
		t := truncate(strings.TrimSpace(hit.Topic), 38)
		if t != "" && !a[t] {
			a[t] = true
			akin = append(akin, t)
		}
		if len(akin) >= 3 {
			break
		}
	}
	if len(akin) > 0 {
		answer += "\n\n(I connect this to: " + strings.Join(akin, "; ") + ")"
	}
	return answer
}

var _, _ = bestGrain, curiosityWord // kept for an optional dog-voice mode

// bestGrain returns the single recalled sentence most resonant with the query —
// the densest grain the dog actually gathered, kept short.
func bestGrain(qd encoding.Detail, hits []memory.Hit) string {
	dq := encoding.Distinctive(qd) // the concept's signature, not the shared hum
	best, bestScore := "", -1.0
	for _, hit := range hits {
		for _, raw := range sentenceRe.FindAllString(hit.Text, -1) {
			s := strings.TrimSpace(raw)
			if n := len([]rune(s)); n < 30 || n > 200 {
				continue
			}
			ds := encoding.Encode(s)
			// match on DISTINCTIVE content, so generic boilerplate (CC footers,
			// nav chrome — all baseline hum) scores ~0, not high.
			score := encoding.DetailResonance(dq, encoding.Distinctive(ds)) * (1 + float64(ds.Active())/float64(encoding.Families*encoding.Subfamilies))
			if score > bestScore {
				bestScore, best = score, s
			}
		}
	}
	return best
}

// proseRatio is the share of a sentence that is letters and spaces — high for
// human prose, low for URLs, metadata, and symbol dumps. A text-quality signal,
// not a vocabulary.
func proseRatio(s string) float64 {
	good, total := 0, 0
	for _, r := range s {
		total++
		if unicode.IsLetter(r) || r == ' ' {
			good++
		}
	}
	if total == 0 {
		return 0
	}
	return float64(good) / float64(total)
}

func curiosityWord(c float64) string {
	switch {
	case c >= 0.55:
		return "keen"
	case c >= 0.4:
		return "stirring"
	default:
		return "quiet"
	}
}

// transposeFile turns a text file into SVG art-lore, one engram per passage.
func transposeFile(h *memory.Hippocampus, path, outdir string) {
	raw, err := os.ReadFile(path)
	must(err)
	n := 0
	for _, para := range regexp.MustCompile(`\n\s*\n`).Split(string(raw), -1) {
		p := strings.TrimSpace(para)
		if len(p) < 80 {
			continue
		}
		topic := firstWords(p, 7)
		_, err := h.RememberTo(outdir, topic, collapse(p), "transpose:"+filepath.Base(path))
		must(err)
		n++
	}
	fmt.Printf("🐾 transposed %d passages from %s into SVG lore at %s/\n", n, filepath.Base(path), outdir)
}

// ingest indexes existing SVG lore (hand-made or previously drawn).
func ingest(h *memory.Hippocampus, dir string) {
	n := 0
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() || strings.ToLower(filepath.Ext(path)) != ".svg" {
			return nil
		}
		if _, err := h.IngestSVG(path); err == nil {
			n++
		}
		return nil
	})
	fmt.Printf("🐾 ingested %d SVG engrams from %s/ (total %d)\n", n, dir, h.Count())
}

// hunt casts the frontal hunt spell on a territory, then transposes the richest
// catches into SVG lore the dog can recall.
func hunt(h *memory.Hippocampus, territory string) {
	spells, err := modules.Discover(grimoireDir)
	must(err)
	var hs *modules.Spell
	for i := range spells {
		if spells[i].Name == "hunt" {
			hs = &spells[i]
		}
	}
	if hs == nil {
		fmt.Println("no hunt spell found in the grimoire")
		return
	}
	out, err := hs.Cast(context.Background(), modules.Input{Path: territory, Args: map[string]any{"max": 12}})
	must(err)
	fmt.Printf("🐾 %s\n", out.Summary)

	findings := decodeFindings(out.Data["findings"])
	for _, f := range findings {
		if _, err := h.RememberTo(loreDir, f.Topic, f.Text, "hunt:"+f.Source); err == nil {
			fmt.Printf("   ⊛ caught [%s] richness %d → SVG lore\n", truncate(f.Topic, 40), f.Richness)
		}
	}
	fmt.Printf("🐾 the bank now holds %d engrams.\n", h.Count())
}

type finding struct {
	Topic    string `json:"topic"`
	Text     string `json:"text"`
	Richness int    `json:"richness"`
	Source   string `json:"source"`
}

func decodeFindings(v any) []finding {
	raw, _ := json.Marshal(v)
	var fs []finding
	_ = json.Unmarshal(raw, &fs)
	return fs
}

// --- small helpers ---

// salientWords ranks a query's words by how RICH their glyph is — the dog's own
// art deciding what carries meaning. Function words ("the", "what") encode to
// sparse glyphs and sink; content words ("owl", "weapon") are rich and rise.
// No hardcoded word lists: salience is read from the encoding.
func salientWords(s string) []string {
	type wr struct {
		w string
		r int
	}
	var ws []wr
	seen := map[string]bool{}
	for _, w := range strings.Fields(strings.ToLower(s)) {
		w = strings.Trim(w, ".,!?;:\"'()")
		if len(w) < 3 || seen[w] {
			continue
		}
		seen[w] = true
		ws = append(ws, wr{w, encoding.Encode(w).Active()})
	}
	sort.SliceStable(ws, func(i, j int) bool { return ws[i].r > ws[j].r })
	out := make([]string, len(ws))
	for i, x := range ws {
		out[i] = x.w
	}
	return out
}

// salientName names a coined sign after the query's richest words (glyph-chosen).
func salientName(q string) string {
	w := salientWords(q)
	if len(w) == 0 {
		return firstWords(q, 3)
	}
	if len(w) > 3 {
		w = w[:3]
	}
	return strings.Join(w, " ")
}

func firstSentences(s string, n int) string {
	m := sentenceRe.FindAllString(s, n)
	return strings.TrimSpace(strings.Join(m, " "))
}

func firstWords(s string, n int) string {
	f := strings.Fields(collapse(s))
	if len(f) > n {
		f = f[:n]
	}
	return strings.Join(f, " ")
}

func collapse(s string) string {
	s = strings.NewReplacer("#", "", "*", "", "`", "").Replace(s)
	return strings.Join(strings.Fields(s), " ")
}

func truncate(s string, n int) string {
	s = strings.Join(strings.Fields(s), " ")
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func usage() {
	fmt.Println(`cerebrum — Argos's mind
  stats
  transpose <textfile> [outdir]   language -> SVG lore via the glyph encoding
  hunt [territory]                hunt richest passages -> SVG lore
  sniff "<query>"                 hunt wikipedia+arxiv -> SVG lore (the nose)
  sense "<text>"                  show families + subfamilies firing (diagnostic)
  sense --glyph "<text>"           show full 1512-cell vector as logogram syllables (glyph language)
  train                           teach the cerebellum (instinct) on the bank
  ingest [loredir]                index existing SVG lore
  recall "<query>" [k]
  ask "<question>"
  chat                            live conversation (the muzzle)
  coin "<concept>"                mint a new named logogram
  lexicon                         list coined logograms`)
	os.Exit(1)
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "cerebrum:", err)
		os.Exit(1)
	}
}