// Package parietal is Argos's parietal lobe — its sense of WHERE it is. Like the
// real lobe, it does spatial awareness and sensory integration: it turns each
// perceived place into a point on the logogram plane (the same plane the game
// uses to place biomes), names the region it is standing in, and grows a map (a
// "cartograph") of everywhere it has been. It is pure cognition — no I/O, no
// worldgen. It hands the will (internal/telekinesis) a sense of pull over the
// paths leading onward.
//
// Why this lobe and the game's biomes are the same idea: encoding.Locate places any
// perception at an (x,y); the game lays biomes over that very plane. So the dog's
// FELT position is the world's coordinate — egocentric perception and the
// allocentric map are one transform, not a guess.
package spatial

import (
	"fmt"
	"os"

	"glyphai/internal/encoding"
	"glyphai/internal/output"
)

// Link is one path onward the dog could take (text it reads + where it leads).
type Link struct{ Text, URL string }

// Weighted is a candidate path with the dog's pull toward it.
type Weighted struct {
	URL    string
	Weight float64
}

// Place is the dog's reading of one location: where it sits on the plane, the
// region (biome) it belongs to, and its coarse encoding.
type Place struct {
	URL, Title, Biome string
	X, Y              float64
	Glyph             encoding.Glyph
}

// Parietal holds the live position and the growing cartograph.
type Parietal struct {
	here Place
	seen map[string]Place
}

func New() *Parietal { return &Parietal{seen: map[string]Place{}} }

// Integrate folds a freshly perceived place into the spatial model: it locates
// the place on the plane, names its region, records it on the map, and makes it
// "here". This is the egocentric->allocentric step.
func (p *Parietal) Integrate(url, title, text string) Place {
	d := encoding.Encode(title + " " + text)
	x, y := encoding.Locate(d)
	pl := Place{URL: url, Title: title, Biome: regionOf(d), X: x, Y: y, Glyph: d.Coarse()}
	p.here = pl
	p.seen[url] = pl
	return pl
}

// Here is the dog's current place. Seen reports whether a url is already mapped.
func (p *Parietal) Here() Place { return p.here }
func (p *Parietal) Seen(url string) bool {
	_, ok := p.seen[url]
	return ok
}

// Mapped is how many distinct places the cartograph holds.
func (p *Parietal) Mapped() int { return len(p.seen) }

// Options weighs the paths onward by the dog's pull: the glyph resonance between
// what it is reading now and the scent of each path. (Novelty — preferring the
// unmapped — is applied by the caller, exactly as roam does, so no tuned constant
// lives in the lobe.)
func (p *Parietal) Options(links []Link) []Weighted {
	cur := p.here.Glyph
	out := make([]Weighted, 0, len(links))
	for _, l := range links {
		g := encoding.EncodeGlyph(l.Text)
		out = append(out, Weighted{URL: l.URL, Weight: encoding.Resonance(cur, g)})
	}
	return out
}

// regionOf names the biome by the strongest DISTINCTIVE family (baseline removed,
// so the name reflects what is special here, not the shared Voice hum).
func regionOf(d encoding.Detail) string {
	dd := encoding.Distinctive(d)
	best, bestV := 0, -1.0
	for f := 0; f < encoding.Families; f++ {
		s := 0.0
		for sub := 0; sub < encoding.Subfamilies; sub++ {
			s += float64(dd[f][sub])
		}
		if s > bestV {
			bestV, best = s, f
		}
	}
	return encoding.FamilyNames[best]
}

// Save writes the cartograph as an SVG art-engram (persistence is SVG art), so
// the map survives reboots and grows across sessions.
func (p *Parietal) Save(path string) error {
	rows := make([][]string, 0, len(p.seen))
	for _, pl := range p.seen {
		rows = append(rows, []string{pl.URL, pl.Title, pl.Biome,
			fmt.Sprintf("%.4f", pl.X), fmt.Sprintf("%.4f", pl.Y)})
	}
	payload := output.EncodeRecords(rows)
	svg := output.Inscribe("cartograph", "Argos cartograph", payload,
		fmt.Sprintf("%d places mapped", len(rows)))
	return os.WriteFile(path, []byte(svg), 0o644)
}

// Load restores a cartograph saved earlier. Missing file is not an error.
func (p *Parietal) Load(path string) error {
	b, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	payload, err := output.ReadScroll(b, "cartograph")
	if err != nil {
		return err
	}
	for _, r := range output.DecodeRecords(payload) {
		if len(r) < 5 {
			continue
		}
		var x, y float64
		fmt.Sscanf(r[3], "%f", &x)
		fmt.Sscanf(r[4], "%f", &y)
		p.seen[r[0]] = Place{URL: r[0], Title: r[1], Biome: r[2], X: x, Y: y}
	}
	return nil
}