// Package lexicon is Argos's growing book of coined logograms. Each entry is a
// named sign (a "vessel" in the Ifá sense) that holds its logogram, its glyph,
// and the verses — fragments of knowledge — that have resonated with it. The
// dog coins new signs when it meets ideas worth a name; the book persists, so
// the script grows toward the Fable-5 pretense of open expression.
package vocabulary

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"glyphai/internal/encoding"
	"glyphai/internal/output"
)

// Entry is one coined logogram and the verses it has gathered.
type Entry struct {
	Name     string   `json:"name"`     // what the dog calls this sign
	Concept  string   `json:"concept"`  // the seed meaning it was coined from
	Logogram string   `json:"logogram"` // the written sign
	Glyph    string   `json:"glyph"`    // coarse glyph hex
	Reading  string   `json:"reading"`  // artistic reading of the glyph
	Verses   []string `json:"verses"`   // accumulated wisdom (the Ifá corpus)
	Coined   string   `json:"coined"`   // RFC3339
}

// Lexicon is the on-disk book of signs — persisted as an SVG scroll, not JSON.
type Lexicon struct {
	path  string
	Signs map[string]Entry // keyed by lowercased name
}

const verseSep = "\x1f"

// Open loads (or starts) a lexicon at an SVG path.
func Open(path string) (*Lexicon, error) {
	lx := &Lexicon{path: path, Signs: map[string]Entry{}}
	raw, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return lx, os.MkdirAll(filepath.Dir(path), 0755)
		}
		return nil, err
	}
	payload, err := output.ReadScroll(raw, "lexicon")
	if err != nil {
		return lx, nil // unreadable/empty scroll: start fresh
	}
	for _, r := range output.DecodeRecords(payload) {
		if len(r) < 7 {
			continue
		}
		var verses []string
		if r[6] != "" {
			verses = strings.Split(r[6], verseSep)
		}
		lx.Signs[key(r[0])] = Entry{
			Name: r[0], Concept: r[1], Logogram: r[2], Glyph: r[3],
			Reading: r[4], Coined: r[5], Verses: verses,
		}
	}
	return lx, nil
}

// Has reports whether a sign already exists for the given name/concept.
func (lx *Lexicon) Has(name string) bool {
	_, ok := lx.Signs[key(name)]
	return ok
}

// Coin mints (or returns) a named logogram for a concept. The dog derives the
// sign from the concept's glyph through the glyph encoding + syllabic script.
func (lx *Lexicon) Coin(name, concept string) (Entry, error) {
	if e, ok := lx.Signs[key(name)]; ok {
		return e, nil
	}
	d := encoding.Encode(concept)
	g := d.Coarse()
	e := Entry{
		Name:     name,
		Concept:  concept,
		Logogram: encoding.Logogram(d),
		Glyph:    g.Hex(),
		Reading:  encoding.Reading(g),
		Coined:   time.Now().Format(time.RFC3339),
	}
	lx.Signs[key(name)] = e
	return e, lx.save()
}

// Inscribe adds a verse (a fragment of resonant knowledge) to a sign's corpus.
func (lx *Lexicon) Inscribe(name, verse string) error {
	e, ok := lx.Signs[key(name)]
	if !ok {
		return os.ErrNotExist
	}
	for _, v := range e.Verses {
		if v == verse {
			return nil
		}
	}
	e.Verses = append(e.Verses, verse)
	lx.Signs[key(name)] = e
	return lx.save()
}

// List returns all coined signs.
func (lx *Lexicon) List() []Entry {
	out := make([]Entry, 0, len(lx.Signs))
	for _, e := range lx.Signs {
		out = append(out, e)
	}
	return out
}

func (lx *Lexicon) Count() int { return len(lx.Signs) }

func (lx *Lexicon) save() error {
	var rows [][]string
	var body strings.Builder
	y := 90
	for _, e := range lx.Signs {
		rows = append(rows, []string{e.Name, e.Concept, e.Logogram, e.Glyph,
			e.Reading, e.Coined, strings.Join(e.Verses, verseSep)})
		if y < 420 {
			fmt.Fprintf(&body, `<text x="28" y="%d" fill="#e6c97a" font-size="16">%s</text>`+"\n", y, output.XMLEsc(e.Logogram))
			fmt.Fprintf(&body, `<text x="240" y="%d" fill="#7a849a" font-size="12">%s</text>`+"\n", y, output.XMLEsc(trunc(e.Name, 32)))
			y += 22
		}
	}
	svg := output.Inscribe("lexicon", "Argos · book of coined logograms", output.EncodeRecords(rows), body.String())
	return os.WriteFile(lx.path, []byte(svg), 0644)
}

func trunc(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func key(s string) string { return strings.ToLower(strings.TrimSpace(s)) }