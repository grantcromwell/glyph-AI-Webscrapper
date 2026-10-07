// Package lingua — Argos's foreign-language ear.
//
// When Argos hears a glyph from another animal (21 raw bytes, coarse family vector),
// it doesn't translate. It encodes that foreign sound into its own 21×72 Detail space,
// then finds the nearest logogram in its lexicon — or coins a new one.
//
// The logogram IS the understanding. No dictionaries, no translation tables,
// no hardcoded mappings between animals. Just the dog building its own vocabulary
// for the sounds it hears.
//
// A foreign glyph matures like a real word: first exposure = tentative coinage,
// repeated resonance = solidified meaning, disuse = decay.

package translation

import (
	"encoding/hex"
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"glyphai/internal/encoding"
	"glyphai/internal/vocabulary"
)

// ForeignWord is a logogram Argos coined for something another animal said.
// The glyph is Argos's internal encoding; the tag remembers which animal it heard it from.
type ForeignWord struct {
	Logogram   string    // the coined logogram glyph (e.g. "▽⚷⁷✦🜞⁷")
	Internal   encoding.Detail // Argos's native encoding of the foreign thought
	Coarse     encoding.Glyph  // the 21-value fingerprint
	Tag        string    // which animal (raven, octopus)
	Born       time.Time
	LastHeard  time.Time
	Hearings   int       // how many times this logogram was activated
	Confidence float64   // 0..1 — how well Argos feels it understands this word
}

// Lingua is the organ that learns foreign glyphs as logograms.
type Lingua struct {
	lex       *vocabulary.Lexicon
	foreign   []ForeignWord // growing vocabulary of foreign concepts
	decay     float64       // how fast unused words fade (per day)
	dir       string        // persistence directory
	maxWords  int           // maximum foreign words to keep
}

func New(lex *vocabulary.Lexicon, dir string) *Lingua {
	l := &Lingua{
		lex:      lex,
		foreign:  []ForeignWord{},
		decay:    0.5,       // 50% confidence loss per day of silence
		dir:      dir,
		maxWords: 256,       // don't let foreign vocabulary explode
	}
	os.MkdirAll(dir, 0755)
	l.load()
	return l
}

// Hear processes a foreign glyph vector (21 raw bytes from the bazaar).
// Returns the logogram Argos coined or matched — the "word" it now knows.
func (l *Lingua) Hear(author string, foreignVec []byte) ForeignWord {
	if len(foreignVec) < 21 {
		return ForeignWord{Logogram: "?", Confidence: 0}
	}

	// Encode the foreign vector into Argos's native 21×72 Detail space.
	// Each byte (0..7) maps to the dominant subfamily within that family.
	var detail encoding.Detail
	for fam := 0; fam < 21 && fam < len(foreignVec); fam++ {
		level := int(foreignVec[fam])
		if level > 7 {
			level = 7
		}
		// Spread across subfamilies proportional to level
		for sub := 0; sub < 72; sub++ {
			if level == 0 {
				detail[fam][sub] = 0
			} else {
				// Higher levels activate more subfamilies in a gradient
				dist := float64(sub) / 71.0
				val := float64(level) * math.Exp(-dist*float64(level)*0.5)
				if val > 7 {
					val = 7
				}
				detail[fam][sub] = uint8(val)
			}
		}
	}

	coarse := detail.Coarse()

	// Find the closest existing foreign word by glyph resonance
	bestIdx := -1
	bestScore := 0.0
	for i := range l.foreign {
		score := encoding.DetailResonance(detail, l.foreign[i].Internal)
		if score > bestScore {
			bestScore = score
			bestIdx = i
		}
	}

	// If the resonance is high enough and from the same animal, reinforce it
	if bestScore > 0.65 && bestIdx >= 0 && l.foreign[bestIdx].Tag == author {
		l.foreign[bestIdx].Hearings++
		l.foreign[bestIdx].LastHeard = time.Now()
		// Confidence grows with repetition but asymptotes
		l.foreign[bestIdx].Confidence = math.Min(1.0, l.foreign[bestIdx].Confidence+0.1)
		// Refine the internal encoding — blend the new hearing with memory
		for fam := 0; fam < 21; fam++ {
			for sub := 0; sub < 72; sub++ {
				old := float64(l.foreign[bestIdx].Internal[fam][sub])
				new := float64(detail[fam][sub])
				blended := old*0.7 + new*0.3
				if blended > 7 {
					blended = 7
				}
				l.foreign[bestIdx].Internal[fam][sub] = uint8(blended)
			}
		}
		l.save()
		return l.foreign[bestIdx]
	}

	// It's a new sound — coin a logogram
	fw := ForeignWord{
		Internal:   detail,
		Coarse:     coarse,
		Tag:        author,
		Born:       time.Now(),
		LastHeard:  time.Now(),
		Hearings:   1,
		Confidence: 0.3, // tentative — barely heard
	}

	// Build a logogram from the coarse glyph
	fw.Logogram = coinFromGlyph(&coarse, author)

	l.foreign = append(l.foreign, fw)
	if len(l.foreign) > l.maxWords {
		// Forget the oldest, least-confident word
		sort.Slice(l.foreign, func(i, j int) bool {
			return l.foreign[i].LastHeard.Before(l.foreign[j].LastHeard)
		})
		l.foreign = l.foreign[len(l.foreign)-l.maxWords:]
	}
	l.save()
	return fw
}

// Think returns Argos's current foreign vocabulary — the words it knows for each animal.
func (l *Lingua) Think(author string) []ForeignWord {
	var result []ForeignWord
	for _, fw := range l.foreign {
		if fw.Tag == author {
			result = append(result, fw)
		}
	}
	// Sort by confidence, highest first
	sort.Slice(result, func(i, j int) bool {
		return result[i].Confidence > result[j].Confidence
	})
	return result
}

// Decay runs once per day — foreign words fade if not heard.
func (l *Lingua) Decay() {
	now := time.Now()
	for i := range l.foreign {
		daysSince := now.Sub(l.foreign[i].LastHeard).Hours() / 24.0
		if daysSince > 0 {
			l.foreign[i].Confidence *= math.Pow(1.0-l.decay, daysSince)
			if l.foreign[i].Confidence < 0.05 {
				l.foreign[i].Confidence = 0
			}
		}
	}
	l.save()
}

// Status returns what Argos knows about foreign languages.
func (l *Lingua) Status() string {
	byAuthor := make(map[string]int)
	for _, fw := range l.foreign {
		byAuthor[fw.Tag]++
	}
	parts := []string{}
	for author, count := range byAuthor {
		parts = append(parts, fmt.Sprintf("%s %d", author, count))
	}
	return fmt.Sprintf("%d %s", len(l.foreign), strings.Join(parts, " "))
}

// ─── Persistence — words live on disk as SVGs ─────────────────────

func (l *Lingua) save() {
	// Wipe old and rewrite — the lexicon grows, it doesn't shrink on save
	for _, fw := range l.foreign {
		l.writeWord(fw)
	}
}

func (l *Lingua) writeWord(fw ForeignWord) {
	hexGlyph := hex.EncodeToString(fw.Coarse[:])
	name := fmt.Sprintf("%s-%s-%d.svg", fw.Tag, hexGlyph[:8], fw.Born.Unix())
	path := filepath.Join(l.dir, name)

	// Build an SVG art-engram — the logogram as visual memory
	svg := fmt.Sprintf(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 300 150">
<rect width="300" height="150" fill="#0a0a12"/>
<text x="150" y="30" text-anchor="middle" fill="#%s" font-size="12" font-family="monospace">%s</text>
<text x="150" y="50" text-anchor="middle" fill="#888" font-size="10">from %s · heard %d times</text>
<text x="150" y="70" text-anchor="middle" fill="#666" font-size="9">confidence: %.0f%%</text>
`,
		colorFor(fw.Tag), fw.Logogram, fw.Tag, fw.Hearings, fw.Confidence*100)

	// Draw the 21-family glyph as circles
	for fam := 0; fam < 21; fam++ {
		level := float64(fw.Coarse[fam]) / 7.0
		if level < 0.1 {
			continue
		}
		angle := float64(fam) * 17.14 * math.Pi / 180.0
		x := 150 + int(40*level*math.Cos(angle))
		y := 100 + int(40*level*math.Sin(angle))
		r := int(2 + level*6)
		svg += fmt.Sprintf(`<circle cx="%d" cy="%d" r="%d" fill="#%s" opacity="0.5"/>
`, x, y, r, colorFor(fw.Tag))
	}
	svg += `</svg>`
	os.WriteFile(path, []byte(svg), 0644)
}

func (l *Lingua) load() {
	entries, _ := os.ReadDir(l.dir)
	for _, e := range entries {
		if !strings.HasSuffix(e.Name(), ".svg") {
			continue
		}
		path := filepath.Join(l.dir, e.Name())
		l.loadWord(path)
	}
}

func (l *Lingua) loadWord(path string) {
	data, err := os.ReadFile(path)
	if err != nil {
		return
	}
	content := string(data)

	// Minimal parsing — extract what we can from the SVG metadata
	var tag string
	var hearings int
	var confidence float64
	var logogram string
	var glyphVec [21]uint8

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if strings.Contains(line, "from") && strings.Contains(line, "heard") {
			fmt.Sscanf(line, "from %s · heard %d times", &tag, &hearings)
			tag = strings.TrimSuffix(tag, "·")
			tag = strings.TrimSpace(tag)
		}
		if strings.Contains(line, "confidence:") {
			fmt.Sscanf(line, "confidence: %f%%", &confidence)
			confidence /= 100.0
		}
		if strings.Contains(line, "font-size=\"12\"") && strings.Contains(line, "monospace") {
			// Extract logogram text from the SVG text element
			start := strings.Index(line, ">")
			end := strings.LastIndex(line, "<")
			if start >= 0 && end > start {
				logogram = line[start+1 : end]
			}
		}
		// Extract glyph circles
		var cx, cy, r int
		var col string
		if n, _ := fmt.Sscanf(line, `<circle cx="%d" cy="%d" r="%d" fill="#%s"`, &cx, &cy, &r, &col); n == 4 {
			// Map circle position to family
			dx := float64(cx - 150)
			dy := float64(cy - 100)
			dist := math.Sqrt(dx*dx + dy*dy)
			level := uint8(dist / 40.0 * 7.0)
			if level > 7 {
				level = 7
			}
			angle := math.Atan2(dy, dx)
			if angle < 0 {
				angle += 2 * math.Pi
			}
			fam := int(angle / (2 * math.Pi) * 21)
			if fam >= 0 && fam < 21 {
				glyphVec[fam] = level
			}
		}
	}

	if tag == "" || logogram == "" {
		return
	}

	fw := ForeignWord{
		Logogram:   logogram,
		Coarse:     glyphVec,
		Tag:        tag,
		Hearings:   hearings,
		Confidence: confidence,
	}
	if fw.Confidence <= 0 {
		fw.Confidence = 0.3
	}
	if fw.Hearings <= 0 {
		fw.Hearings = 1
	}
	l.foreign = append(l.foreign, fw)
}

// ─── Helpers ───────────────────────────────────────────────────────

func coinFromGlyph(g *encoding.Glyph, author string) string {
	// Pick the top 3 families by activation
	type famAct struct {
		idx int
		val uint8
	}
	active := make([]famAct, 0, 21)
	for i, v := range g {
		if v > 0 {
			active = append(active, famAct{i, v})
		}
	}
	sort.Slice(active, func(i, j int) bool {
		return active[i].val > active[j].val
	})

	if len(active) == 0 {
		return "◌"
	}

	top := active[:min(3, len(active))]

	// Build a logogram from the dominant families
	logos := []string{}
	for _, f := range top {
		// Each family has a glyph symbol based on its index
		sym := string(rune(0x25CB + f.idx)) // ◌ ◍ ◎ ● ◐ ◑ ◒ ◓ ◔ ◕ ◖ ◗ ...
		// Subfamily derived from the value
		sub := f.val
		if sub > 0 {
			logos = append(logos, fmt.Sprintf("%s%c%d", sym, rune('₀'+rune(sub%10)), sub))
		} else {
			logos = append(logos, fmt.Sprintf("%s", sym))
		}
	}
	return strings.Join(logos, "")
}

func colorFor(author string) string {
	switch author {
	case "raven":
		return "ff6633"
	case "argos":
		return "44aaff"
	case "octopus":
		return "cc44ff"
	default:
		return "88aacc"
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}