// Package output is how glyphai leaves marks: it renders meaning as an SVG that is
// at once (a) artistic, (b) a transposition of language through the Artistic
// Key, and (c) machine-recoverable knowledge. The dog draws an SVG when it
// wants to remember; the hippocampus reads it back later.
//
// The knowledge travels in an XML comment marker (ARGOS-ENGRAM <json>) so the
// art and the data live in one file — "knowledge as art."
package output

import (
	"encoding/json"
	"fmt"
	"math"
	"regexp"
	"strings"

	"glyphai/internal/encoding"
)

// Engram is one remembered piece of knowledge.
type Engram struct {
	ID      string `json:"id"`
	Topic   string `json:"topic"`
	Text    string `json:"text"`
	Glyph   string `json:"glyph"`   // coarse glyph hex
	Created string `json:"created"` // RFC3339
	Source  string `json:"source,omitempty"`
	detail  encoding.Detail
}

// SetDetail attaches the fine field so the art can render it (not serialised).
func (e *Engram) SetDetail(d encoding.Detail) { e.detail = d }

const (
	svgW  = 640
	svgH  = 700
	gridL = 40
	gridR = 600
	acx   = 320.0 // art centre x
	acy   = 300.0 // art centre y
	aR    = 235.0 // art radius
)

// renderCtx is the shared data every layer draws from — computed once. Layers
// read it; none of them owns the encoding.
type renderCtx struct {
	e          Engram
	g          encoding.Glyph
	d          encoding.Detail
	rot        float64 // per-glyph orientation
	hueShift   float64 // per-glyph palette shift
	complexity float64
	seal       []encoding.Syllable
	h1, h2     float64 // the two dominant family hues
}

// Layer is one independent module of the picture. Add a layer, or grow one
// toward per-pixel resolution, without touching the others.
type Layer func(b *strings.Builder, c *renderCtx)

// Layers is the ordered composition. It is a plain slice so the dog's art is
// modular: append a module, reorder, or swap one out freely.
var Layers = []Layer{
	layerGround,      // the canvas colour = dominant meanings
	layerField,       // adaptive raster of the whole field — scales toward every pixel
	layerStarfield,   // active cells as bright stars
	layerWeb,         // threads between the figures
	layerFigures,     // the dominant families as characters
	layerCentre,      // the dog's mark
	layerFrame,       // per-subfamily rhythm along the foot
	layerInscription, // topic, glyph, reading, logogram
	layerLocus,       // where the topic sits on the logogram plane + mood + curiosity
	layerKnowledge,   // the readable page
}

// Render composes the engram into SVG art by running every layer in order.
func Render(e Engram) string {
	g, _ := encoding.ParseHex(e.Glyph)
	d := e.detail
	if d == (encoding.Detail{}) {
		d = encoding.Encode(e.Topic + " " + e.Text)
	}
	seed := hash64(g.Hex())
	c := &renderCtx{
		e: e, g: g, d: d,
		rot:        float64(seed%3600) / 3600 * 2 * math.Pi,
		hueShift:   float64((seed/7)%40) - 20,
		complexity: d.Complexity(),
		seal:       encoding.Seal(d, 8),
	}
	c.h1, c.h2 = groundHues(g)

	var b strings.Builder
	meta, _ := json.Marshal(e)
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fmt.Fprintf(&b, "<!--ARGOS-ENGRAM %s -->\n", string(meta))
	fmt.Fprintf(&b, `<svg xmlns="http://www.w3.org/2000/svg" width="%d" height="%d" viewBox="0 0 %d %d" font-family="monospace">`+"\n", svgW, svgH, svgW, svgH)
	for _, l := range Layers {
		l(&b, c)
	}
	b.WriteString(`</svg>` + "\n")
	return b.String()
}

// fieldAt samples the glyph field at a canvas point: angle -> family,
// distance-from-centre -> subfamily. This mapping lets the picture be read at
// any resolution, down to a single pixel. Returns the level and the family.
func fieldAt(c *renderCtx, x, y float64) (uint8, int) {
	ang := math.Mod(math.Atan2(y-acy, x-acx)-c.rot, 2*math.Pi)
	if ang < 0 {
		ang += 2 * math.Pi
	}
	f := int(ang / (2 * math.Pi) * encoding.Families)
	if f < 0 {
		f = 0
	} else if f >= encoding.Families {
		f = encoding.Families - 1
	}
	rn := math.Hypot(x-acx, y-acy) / aR
	if rn > 1 {
		rn = 1
	}
	s := int(rn * (encoding.Subfamilies - 1))
	return c.d[f][s], f
}

func layerGround(b *strings.Builder, c *renderCtx) {
	fmt.Fprintf(b, `<defs><radialGradient id="ground" cx="50%%" cy="40%%" r="80%%">`+
		`<stop offset="0%%" stop-color="hsl(%.0f,55%%,%d%%)"/>`+
		`<stop offset="100%%" stop-color="hsl(%.0f,65%%,5%%)"/></radialGradient></defs>`+"\n",
		c.h1+c.hueShift, 7+int(c.complexity*16), c.h2+c.hueShift)
	fmt.Fprintf(b, `<rect width="%d" height="%d" fill="url(#ground)"/>`+"\n", svgW, svgH)
}

// layerField is the modular "use every pixel" layer. Its resolution GROWS with
// how much the glyph knows: sparse coarse tiles when the field is thin, finer
// and finer as it fills — raise the cap and it tiles down to the pixel.
func layerField(b *strings.Builder, c *renderCtx) {
	res := 14 + c.d.Active()/10
	if res > 64 { // safety cap today; lift it and the dog paints every pixel
		res = 64
	}
	tw := float64(svgW) / float64(res)
	th := float64(svgH) / float64(res)
	for ty := 0; ty < res; ty++ {
		for tx := 0; tx < res; tx++ {
			x := (float64(tx) + 0.5) * tw
			y := (float64(ty) + 0.5) * th
			lvl, f := fieldAt(c, x, y)
			if lvl == 0 {
				continue
			}
			l := float64(lvl) / encoding.MaxLevel
			fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="hsl(%.0f,60%%,%d%%)" opacity="%.2f"/>`+"\n",
				float64(tx)*tw, float64(ty)*th, tw+0.6, th+0.6, encoding.FamilyHue(f)+c.hueShift, 10+int(l*22), 0.10+l*0.20)
		}
	}
}

func layerStarfield(b *strings.Builder, c *renderCtx) {
	for f := 0; f < encoding.Families; f++ {
		base := float64(f)/encoding.Families*2*math.Pi + c.rot
		hue := encoding.FamilyHue(f) + c.hueShift
		for s := 0; s < encoding.Subfamilies; s++ {
			lvl := c.d[f][s]
			if lvl == 0 {
				continue
			}
			l := float64(lvl) / encoding.MaxLevel
			ang := base + (float64(s)/encoding.Subfamilies-0.5)*(2*math.Pi/encoding.Families)
			r := 46 + float64(s)/encoding.Subfamilies*aR
			fmt.Fprintf(b, `<circle cx="%.1f" cy="%.1f" r="%.2f" fill="hsl(%.0f,80%%,%d%%)" opacity="%.2f"/>`+"\n",
				acx+r*math.Cos(ang), acy+r*math.Sin(ang), 0.6+l*1.8, hue, 50+int(l*35), 0.22+l*0.5)
		}
	}
}

func figurePos(c *renderCtx, sy encoding.Syllable) (x, y, l float64) {
	ang := float64(sy.Family)/encoding.Families*2*math.Pi + c.rot
	l = float64(sy.Level) / encoding.MaxLevel
	r := 60 + l*(aR-70)
	return acx + r*math.Cos(ang), acy + r*math.Sin(ang), l
}

func layerWeb(b *strings.Builder, c *renderCtx) {
	for _, sy := range c.seal {
		x, y, l := figurePos(c, sy)
		fmt.Fprintf(b, `<line x1="%.0f" y1="%.0f" x2="%.1f" y2="%.1f" stroke="hsl(%.0f,60%%,55%%)" stroke-width="%.1f" opacity="0.33"/>`+"\n",
			acx, acy, x, y, encoding.FamilyHue(sy.Family)+c.hueShift, 0.5+l*2.5)
	}
}

func layerFigures(b *strings.Builder, c *renderCtx) {
	for _, sy := range c.seal {
		x, y, l := figurePos(c, sy)
		hue := encoding.FamilyHue(sy.Family) + c.hueShift
		size := 16 + int(l*34)
		fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="middle" font-size="%d" fill="hsl(%.0f,80%%,68%%)">%s</text>`+"\n",
			x, y+float64(size)/3, size, hue, encoding.FamilySymbol(sy.Family))
		fmt.Fprintf(b, `<text x="%.1f" y="%.1f" text-anchor="middle" font-size="%d" fill="hsl(%.0f,70%%,82%%)" opacity="0.85">%s</text>`+"\n",
			x+float64(size)/2, y-float64(size)/3, size/2+4, hue, encoding.SubfamilySymbols[sy.Subfamily])
	}
}

func layerCentre(b *strings.Builder, c *renderCtx) {
	heart := 16 + int(c.g.Active())
	fmt.Fprintf(b, `<text x="%.0f" y="%.1f" text-anchor="middle" font-size="%d" fill="#eef2f8">🐾</text>`+"\n", acx, acy+float64(heart)/3, heart)
}

func layerFrame(b *strings.Builder, c *renderCtx) {
	col := columnActivity(c.d)
	for s := 0; s < encoding.Subfamilies; s++ {
		if col[s] <= 0 {
			continue
		}
		x := float64(gridL) + float64(s)/encoding.Subfamilies*float64(gridR-gridL)
		hgt := 2 + col[s]*34
		fmt.Fprintf(b, `<rect x="%.1f" y="%.1f" width="6.5" height="%.1f" fill="hsl(%.0f,70%%,55%%)" opacity="0.7"/>`+"\n",
			x, 578-hgt, hgt, c.h1+c.hueShift)
	}
}

func layerInscription(b *strings.Builder, c *renderCtx) {
	fmt.Fprintf(b, `<text x="%d" y="30" fill="#eef2f8" font-size="18" font-weight="bold">🐾 %s</text>`+"\n", gridL, xmlEsc(trunc(c.e.Topic, 46)))
	fmt.Fprintf(b, `<text x="%d" y="50" fill="#aeb6c8" font-size="11">glyph %s · complexity %.2f · %d/1512 cells</text>`+"\n", gridL, c.g.Hex(), c.complexity, c.d.Active())
	fmt.Fprintf(b, `<text x="%d" y="70" fill="#cfd6e6" font-size="13" font-style="italic">“%s”</text>`+"\n", gridL, xmlEsc(trunc(encoding.Reading(c.g), 60)))
	fmt.Fprintf(b, `<text x="%d" y="30" text-anchor="end" fill="#e6c97a" font-size="17">%s</text>`+"\n", gridR, xmlEsc(encoding.Logogram(c.d)))
}

// layerLocus plots WHERE this topic sits on the logogram plane (its level-weighted
// centroid), draws a crosshair there, and inscribes the coordinate, the two
// logograms it lies between, the dog's mood toward it, and its curiosity. The
// picture thus holds position, correlation, curiosity, and emotion.
func layerLocus(b *strings.Builder, c *renderCtx) {
	x, y := encoding.Locate(c.d)
	px, py := acx+x*aR, acy+y*aR
	fmt.Fprintf(b, `<circle cx="%.1f" cy="%.1f" r="6.5" fill="none" stroke="#eef" stroke-width="1.4"/>`+"\n", px, py)
	fmt.Fprintf(b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#eef" stroke-width="0.6"/>`+"\n", px-11, py, px+11, py)
	fmt.Fprintf(b, `<line x1="%.1f" y1="%.1f" x2="%.1f" y2="%.1f" stroke="#eef" stroke-width="0.6"/>`+"\n", px, py-11, px, py+11)
	betw := ""
	if len(c.seal) >= 2 {
		betw = encoding.FamilySymbol(c.seal[0].Family) + encoding.SubfamilySymbols[c.seal[0].Subfamily] +
			" ↔ " + encoding.FamilySymbol(c.seal[1].Family) + encoding.SubfamilySymbols[c.seal[1].Subfamily]
	}
	line := fmt.Sprintf("⌖ (%.2f, %.2f)  between %s  ·  mood %s  ·  curiosity %.2f",
		x, y, betw, encoding.Emotion(c.d), encoding.Curiosity(c.d))
	fmt.Fprintf(b, `<text x="28" y="%d" fill="#8fa0c0" font-size="10">%s</text>`+"\n", svgH-14, xmlEsc(line))
}

func layerKnowledge(b *strings.Builder, c *renderCtx) {
	fmt.Fprintf(b, `<text x="%d" y="600" fill="#9aa6bd" font-size="11" font-weight="bold">memory</text>`+"\n", gridL)
	for i, line := range wrap(c.e.Text, 80, 7) {
		fmt.Fprintf(b, `<text x="%d" y="%d" fill="#8893a8" font-size="10">%s</text>`+"\n", gridL, 618+i*13, xmlEsc(line))
	}
}

var engramRe = regexp.MustCompile(`(?s)<!--ARGOS-ENGRAM (\{.*?\}) -->`)

// Parse recovers the Engram knowledge embedded in an SVG the dog drew.
func Parse(svg []byte) (Engram, error) {
	m := engramRe.FindSubmatch(svg)
	if m == nil {
		return Engram{}, fmt.Errorf("paw: no ARGOS-ENGRAM marker found")
	}
	var e Engram
	if err := json.Unmarshal(m[1], &e); err != nil {
		return Engram{}, fmt.Errorf("paw: bad engram json: %w", err)
	}
	return e, nil
}

func trunc(s string, n int) string {
	s = strings.TrimSpace(strings.ReplaceAll(s, "\n", " "))
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}

func wrap(s string, width, maxLines int) []string {
	s = strings.Join(strings.Fields(s), " ")
	var lines []string
	for len(s) > 0 && len(lines) < maxLines {
		if len(s) <= width {
			lines = append(lines, s)
			break
		}
		cut := width
		if i := strings.LastIndex(s[:width], " "); i > 0 {
			cut = i
		}
		lines = append(lines, s[:cut])
		s = strings.TrimSpace(s[cut:])
	}
	if len(s) > 0 && len(lines) == maxLines {
		lines[maxLines-1] = trunc(lines[maxLines-1]+" "+s, width)
	}
	return lines
}

func xmlEsc(s string) string {
	r := strings.NewReplacer("&", "&amp;", "<", "&lt;", ">", "&gt;", `"`, "&quot;")
	return r.Replace(s)
}

// hash64 (FNV-1a) seeds per-glyph rotation and palette so distinct memories
// look distinct even when their meaning rhymes.
func hash64(s string) uint64 {
	h := uint64(1469598103934665603)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= 1099511628211
	}
	return h
}

// groundHues returns the hues of the two strongest families — the colours that
// paint the canvas ground.
func groundHues(g encoding.Glyph) (h1, h2 float64) {
	a, b := -1, -1
	for f := 0; f < encoding.Families; f++ {
		if a < 0 || g[f] > g[a] {
			b, a = a, f
		} else if b < 0 || g[f] > g[b] {
			b = f
		}
	}
	h1, h2 = 220, 260
	if a >= 0 {
		h1 = encoding.FamilyHue(a)
	}
	if b >= 0 {
		h2 = encoding.FamilyHue(b)
	}
	return
}

// columnActivity returns, per subfamily, how active it is across all families,
// normalised to 0..1 — the frame's rhythm.
func columnActivity(d encoding.Detail) [encoding.Subfamilies]float64 {
	var col [encoding.Subfamilies]float64
	max := 0.0
	for s := 0; s < encoding.Subfamilies; s++ {
		var sum float64
		for f := 0; f < encoding.Families; f++ {
			sum += float64(d[f][s])
		}
		col[s] = sum
		if sum > max {
			max = sum
		}
	}
	if max > 0 {
		for s := range col {
			col[s] /= max
		}
	}
	return col
}