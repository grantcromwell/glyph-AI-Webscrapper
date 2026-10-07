package encoding

import "math"

// The logogram plane. Every subglyph (family f, subfamily s) has a fixed place:
// the family sets the ANGLE (its mark's bearing), the subfamily sets the RADIUS
// (its alchemical symbol's ring) — the very polar geometry the eye reads and the
// paw draws. So each glyph cell is a point, each logogram a constellation of
// points, and a TOPIC is where its lit cells average. A topic woven from two
// logograms lands between them: betweenness becomes a coordinate, not a guess.

// CellXY is the fixed position of subglyph (f, s) on the logogram plane, in the
// unit disc [-1,1]².
func CellXY(f, s int) (x, y float64) {
	ang := float64(f) / float64(Families) * 2 * math.Pi
	r := float64(s) / float64(Subfamilies-1)
	return r * math.Cos(ang), r * math.Sin(ang)
}

// Locate places a whole glyph (a topic) at one point: the level-weighted centroid
// of its lit cells. Two logograms are two points; a topic from both lands between.
func Locate(d Detail) (x, y float64) {
	var sx, sy, sw float64
	for f := 0; f < Families; f++ {
		for s := 0; s < Subfamilies; s++ {
			w := float64(d[f][s])
			if w == 0 {
				continue
			}
			cx, cy := CellXY(f, s)
			sx += cx * w
			sy += cy * w
			sw += w
		}
	}
	if sw == 0 {
		return 0, 0
	}
	return sx / sw, sy / sw
}

// Curiosity is how much the system is drawn to explore a glyph — its information
// richness (complexity). Sparse = settled; rich = it wants to know more.
func Curiosity(d Detail) float64 { return d.Complexity() }