// Spell: phoible (tongue school) — the dog learns the SOUND of any human tongue.
//
// PHOIBLE is a cross-linguistic database of phoneme inventories for ~2000
// languages, each segment tagged with phonological features (consonantal,
// sonorant, nasal, continuant, strident…). This spell looks a language up by
// name or ISO-639-3 code, classifies every phoneme into one of the dog's own
// breath-shapes from those features, and paints the whole inventory into a glyph
// through the SAME family-breath binding the dog hears speech with.
//
// So the dog isn't limited to the four tongues it forages news in: it can FEEL
// the shape of Russian's sibilant-heavy palatalised consonants, Mandarin's tone,
// Spanish's open vowels — the phonology of Earth's languages, as art.
//
//	echo '{"text":"Russian"}'  | go run ./grimoire/tongue/phoible
//	echo '{"text":"cmn"}'      | go run ./grimoire/tongue/phoible
package main

import (
	"encoding/csv"
	"encoding/json"
	"fmt"
	"hash/fnv"
	"os"
	"sort"
	"strconv"
	"strings"

	"glyphai/internal/audio"
	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"glyphai/internal/fetch"
)

const csvPath = "modules/tongue/phoible/phoible.csv"

func main() {
	var in modules.Input
	_ = json.NewDecoder(os.Stdin).Decode(&in)
	query := strings.TrimSpace(in.Text)
	if query == "" {
		query = in.Path
	}

	rows, col, err := load()
	if err != nil {
		fail(err)
	}

	// All distinct languages the dog could know.
	if query == "" {
		langs := map[string]bool{}
		for _, r := range rows {
			langs[r[col["LanguageName"]]] = true
		}
		emit(modules.Output{
			Spell:   "phoible",
			Summary: fmt.Sprintf("the dog can hear %d of Earth's languages — name one (e.g. \"Russian\", \"Mandarin\", or an ISO-639-3 code)", len(langs)),
			Data:    map[string]any{"languages_known": len(langs)},
		})
		return
	}

	name, segs := selectLanguage(rows, col, query)
	if len(segs) == 0 {
		fail(fmt.Errorf("phoible knows no tongue matching %q", query))
	}

	d, profile, tone := paint(segs, col)
	g := d.Coarse()

	// One doc so the hunt/cast can file this as a remembered "sound of a tongue".
	sample := make([]string, 0, 24)
	for _, s := range segs {
		if len(sample) >= 24 {
			break
		}
		sample = append(sample, s[col["Phoneme"]])
	}
	doc := fetch.Doc{
		Title:  name + " — phonology",
		URL:    "phoible:" + name,
		Text:   fmt.Sprintf("%s: %d phonemes. %s. inventory: %s", name, len(segs), profile, strings.Join(sample, " ")),
		Source: "phoible",
	}

	emit(modules.Output{
		Spell:   "phoible",
		Detail:  d,
		Summary: fmt.Sprintf("%s: %d phonemes · %s%s", name, len(segs), profile, tone),
		Data: map[string]any{
			"language":  name,
			"phonemes":  len(segs),
			"glyph":     g.Hex(),
			"reading":   encoding.Reading(g),
			"logogram":  encoding.Logogram(d),
			"breaths":   profile,
			"tone":      tone != "",
			"docs":      []fetch.Doc{doc},
		},
	})
}

// load reads the PHOIBLE csv and returns rows + a column-name index.
func load() ([][]string, map[string]int, error) {
	f, err := os.Open(csvPath)
	if err != nil {
		return nil, nil, err
	}
	defer f.Close()
	r := csv.NewReader(f)
	r.FieldsPerRecord = -1
	recs, err := r.ReadAll()
	if err != nil || len(recs) < 2 {
		return nil, nil, fmt.Errorf("phoible csv unreadable: %v", err)
	}
	col := map[string]int{}
	for i, h := range recs[0] {
		col[h] = i
	}
	return recs[1:], col, nil
}

// selectLanguage matches by ISO-639-3 code (exact) or LanguageName (substring),
// merging every inventory found into one union of phonemes.
func selectLanguage(rows [][]string, col map[string]int, query string) (string, [][]string) {
	q := strings.ToLower(query)
	iso, nameCol := col["ISO6393"], col["LanguageName"]
	var name string
	seen := map[string]bool{}
	var segs [][]string
	for _, r := range rows {
		match := strings.ToLower(r[iso]) == q ||
			strings.Contains(strings.ToLower(r[nameCol]), q)
		if !match {
			continue
		}
		if name == "" {
			name = r[nameCol]
		}
		ph := r[col["Phoneme"]]
		if ph != "" && !seen[ph] {
			seen[ph] = true
			segs = append(segs, r)
		}
	}
	return name, segs
}

// classify turns a PHOIBLE segment's features into one of the dog's breaths,
// using the same shared ear-logic as the cochlea and cmd/attune.
func classify(r []string, col map[string]int) audio.Breath {
	return audio.BreathOf(func(name string) string {
		if i, ok := col[name]; ok && i < len(r) {
			return r[i]
		}
		return ""
	})
}

// paint routes each phoneme to the families that answer its breath,
// at a subfamily fixed by the phoneme itself — so a richer inventory lights more
// of the field, exactly as a richer glyph should.
func paint(segs [][]string, col map[string]int) (encoding.Detail, string, string) {
	var d encoding.Detail
	var counts [audio.NumBreaths]int
	tone := false
	for _, r := range segs {
		b := classify(r, col)
		counts[b]++
		if i, ok := col["tone"]; ok && strings.HasPrefix(r[i], "+") {
			tone = true
		}
		sub := int(hash(r[col["Phoneme"]]) % uint64(encoding.Subfamilies))
		for f := 0; f < encoding.Families; f++ {
			// Check if this family answers to breath b
			if f == 0 || f == 5 || f == 6 || f == 11 || f == 13 || f == 17 {
				// Plosive families
				if b == audio.Plosive {
					cur := int(d[f][sub])
					if cur < encoding.MaxLevel {
						d.Set(f, sub, uint8(cur+2))
					}
				}
			} else if f == 3 || f == 10 || f == 15 || f == 18 || f == 19 {
				// Fricative families
				if b == audio.Fricative {
					cur := int(d[f][sub])
					if cur < encoding.MaxLevel {
						d.Set(f, sub, uint8(cur+2))
					}
				}
			} else if f == 8 || f == 14 {
				// Nasal families
				if b == audio.Nasal {
					cur := int(d[f][sub])
					if cur < encoding.MaxLevel {
						d.Set(f, sub, uint8(cur+2))
					}
				}
			} else if f == 1 || f == 4 || f == 16 {
				// Approximant families
				if b == audio.Approximant {
					cur := int(d[f][sub])
					if cur < encoding.MaxLevel {
						d.Set(f, sub, uint8(cur+2))
					}
				}
			} else if f == 2 || f == 7 || f == 20 {
				// Vowel families
				if b == audio.Vowel {
					cur := int(d[f][sub])
					if cur < encoding.MaxLevel {
						d.Set(f, sub, uint8(cur+2))
					}
				}
			} else if f == 9 || f == 12 {
				// Sibilant families
				if b == audio.Sibilant {
					cur := int(d[f][sub])
					if cur < encoding.MaxLevel {
						d.Set(f, sub, uint8(cur+2))
					}
				}
			}
		}
	}
	// tone is melody → let it ring in the sound families.
	if tone {
		for f := 0; f < encoding.Families; f++ {
			// Tone rings in families 7, 9, 19
			if f == 7 || f == 9 || f == 19 {
				sub := int(hash("tone:"+strconv.Itoa(f)) % uint64(encoding.Subfamilies))
				d.Set(f, sub, encoding.MaxLevel)
			}
		}
	}

	parts := make([]string, 0, audio.NumBreaths)
	order := make([]int, audio.NumBreaths)
	for i := range order {
		order[i] = i
	}
	sort.Slice(order, func(a, b int) bool { return counts[order[a]] > counts[order[b]] })
	for _, b := range order {
		if counts[b] > 0 {
			parts = append(parts, fmt.Sprintf("%d %s", counts[b], audio.BreathName[b]))
		}
	}
	toneStr := ""
	if tone {
		toneStr = " · tonal"
	}
	return d, strings.Join(parts, ", "), toneStr
}

func hash(s string) uint64 {
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}

func emit(o modules.Output) {
	if err := json.NewEncoder(os.Stdout).Encode(o); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "phoible:", err)
	os.Exit(1)
}