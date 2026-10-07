package environment

import (
	"fmt"
	"math"
	"math/rand"
	"strings"
)

// ---------------------------------------------------------------------------
// Colors biome — native to the substrate (hue->angle, sat->radius mirrors the
// logogram plane). The color wheel re-forms from glyph resonance alone, because
// near-hue colors share hue words. Anchor: Light (family 20).
// ---------------------------------------------------------------------------

type Colors struct{}

func (Colors) Name() string { return "colorlands" }
func (Colors) Anchor() int  { return 20 } // Light

var hueNames = []string{
	"red", "orange", "amber", "yellow", "chartreuse", "green",
	"spring green", "cyan", "azure", "blue", "violet", "magenta",
}

func (Colors) Fill(rng *rand.Rand, c Coord) []Node {
	// chunk's base hue is a smooth function of position so neighbouring chunks
	// hold neighbouring hues; jitter within the chunk spreads a small band.
	base := math.Mod(float64(c.X)*30+float64(c.Y)*11+360000, 360)
	n := 5 + rng.Intn(3)
	out := make([]Node, 0, n)
	for i := 0; i < n; i++ {
		hue := math.Mod(base+float64(rng.Intn(40))-20+360, 360)
		sat := 0.45 + rng.Float64()*0.5
		light := 0.30 + rng.Float64()*0.45
		name := hueNames[int(hue/30)%len(hueNames)]
		r, g, b := hslToRGB(hue, sat, light)
		title := fmt.Sprintf("%s — hsl(%.0f, %.0f%%, %.0f%%)", name, hue, sat*100, light*100)
		var body strings.Builder
		fmt.Fprintf(&body, "A %s color at hue %.0f degrees, saturation %.0f percent, lightness %.0f percent. ",
			name, hue, sat*100, light*100)
		fmt.Fprintf(&body, "Its red green blue value is approximately (%d, %d, %d). ", r, g, b)
		fmt.Fprintf(&body, "On the color wheel %s sits among %s and %s. ",
			name, hueNames[(int(hue/30)+len(hueNames)-1)%len(hueNames)], hueNames[(int(hue/30)+1)%len(hueNames)])
		if w := hueWavelength(hue); w > 0 {
			fmt.Fprintf(&body, "As spectral light it is near %d nanometres. ", w)
		}
		body.WriteString("It is a hue, a color, a shade of light the eye can see.")
		out = append(out, Node{ID: nodeID(c, i), Title: title, Body: body.String()})
	}
	return out
}

func hslToRGB(h, s, l float64) (int, int, int) {
	c := (1 - math.Abs(2*l-1)) * s
	x := c * (1 - math.Abs(math.Mod(h/60, 2)-1))
	m := l - c/2
	var r, g, b float64
	switch {
	case h < 60:
		r, g, b = c, x, 0
	case h < 120:
		r, g, b = x, c, 0
	case h < 180:
		r, g, b = 0, c, x
	case h < 240:
		r, g, b = 0, x, c
	case h < 300:
		r, g, b = x, 0, c
	default:
		r, g, b = c, 0, x
	}
	return int((r + m) * 255), int((g + m) * 255), int((b + m) * 255)
}

// hueWavelength is a rough hue->wavelength map across the visible band (red hues
// long, violet hues short). 0 for the magentas, which have no single wavelength.
func hueWavelength(h float64) int {
	if h > 300 {
		return 0
	}
	return int(700 - (h/300)*(700-400))
}

// ---------------------------------------------------------------------------
// Material science biome — embedded factual data as RAG world content (allowed:
// knowledge lives outside the model). Anchor: Iron (family 0).
// ---------------------------------------------------------------------------

type Materials struct{}

func (Materials) Name() string { return "forge" }
func (Materials) Anchor() int  { return 0 } // Iron

type element struct {
	sym, name, cat, crystal string
	z, group, period        int
}

var elements = []element{
	{"H", "hydrogen", "reactive nonmetal", "hexagonal", 1, 1, 1},
	{"He", "helium", "noble gas", "hexagonal close packed", 2, 18, 1},
	{"Li", "lithium", "alkali metal", "body-centered cubic", 3, 1, 2},
	{"C", "carbon", "reactive nonmetal", "hexagonal (graphite) or diamond cubic", 6, 14, 2},
	{"N", "nitrogen", "reactive nonmetal", "hexagonal", 7, 15, 2},
	{"O", "oxygen", "reactive nonmetal", "cubic", 8, 16, 2},
	{"Na", "sodium", "alkali metal", "body-centered cubic", 11, 1, 3},
	{"Mg", "magnesium", "alkaline earth metal", "hexagonal close packed", 12, 2, 3},
	{"Al", "aluminium", "post-transition metal", "face-centered cubic", 13, 13, 3},
	{"Si", "silicon", "metalloid", "diamond cubic", 14, 14, 3},
	{"S", "sulfur", "reactive nonmetal", "orthorhombic", 16, 16, 3},
	{"K", "potassium", "alkali metal", "body-centered cubic", 19, 1, 4},
	{"Ca", "calcium", "alkaline earth metal", "face-centered cubic", 20, 2, 4},
	{"Ti", "titanium", "transition metal", "hexagonal close packed", 22, 4, 4},
	{"Cr", "chromium", "transition metal", "body-centered cubic", 24, 6, 4},
	{"Fe", "iron", "transition metal", "body-centered cubic", 26, 8, 4},
	{"Ni", "nickel", "transition metal", "face-centered cubic", 28, 10, 4},
	{"Cu", "copper", "transition metal", "face-centered cubic", 29, 11, 4},
	{"Zn", "zinc", "transition metal", "hexagonal close packed", 30, 12, 4},
	{"Ag", "silver", "transition metal", "face-centered cubic", 47, 11, 5},
	{"Sn", "tin", "post-transition metal", "tetragonal", 50, 14, 5},
	{"W", "tungsten", "transition metal", "body-centered cubic", 74, 6, 6},
	{"Au", "gold", "transition metal", "face-centered cubic", 79, 11, 6},
	{"Pb", "lead", "post-transition metal", "face-centered cubic", 82, 14, 6},
}

var crystalSystems = []string{
	"cubic", "tetragonal", "orthorhombic", "hexagonal", "trigonal", "monoclinic", "triclinic",
}

func (Materials) Fill(rng *rand.Rand, c Coord) []Node {
	n := 4 + rng.Intn(3)
	out := make([]Node, 0, n)
	start := abs(c.X*7+c.Y*13) % len(elements)
	for i := 0; i < n; i++ {
		e := elements[(start+i)%len(elements)]
		var body strings.Builder
		fmt.Fprintf(&body, "%s (%s), atomic number %d, a %s. ", strings.Title(e.name), e.sym, e.z, e.cat)
		fmt.Fprintf(&body, "It sits in group %d, period %d of the periodic table. ", e.group, e.period)
		fmt.Fprintf(&body, "Its solid forms a %s crystal lattice. ", e.crystal)
		body.WriteString("A chemical element, a material, an atom of matter with mass and structure.")
		out = append(out, Node{ID: nodeID(c, i), Title: fmt.Sprintf("%s (%s)", strings.Title(e.name), e.sym), Body: body.String()})
	}
	// a crystal-system landmark sometimes seeds the chunk's geometry vocabulary
	if rng.Intn(3) == 0 {
		cs := crystalSystems[rng.Intn(len(crystalSystems))]
		body := fmt.Sprintf("The %s crystal system, one of the seven lattice systems. It describes the symmetry of how atoms pack into a repeating three-dimensional lattice of matter.", cs)
		out = append(out, Node{ID: nodeID(c, len(out)), Title: fmt.Sprintf("%s crystal system", strings.Title(cs)), Body: body})
	}
	return out
}

// ---------------------------------------------------------------------------
// Numerology biome — integers described by derived number theory. Pure math;
// numbers with shared structure cluster by resonance. Anchor: Number (family 5).
// ---------------------------------------------------------------------------

type Numerology struct{}

func (Numerology) Name() string { return "numberfields" }
func (Numerology) Anchor() int  { return 5 } // Number

func (Numerology) Fill(rng *rand.Rand, c Coord) []Node {
	base := abs(c.X*60+c.Y*7) + 2 + rng.Intn(20)
	n := 5 + rng.Intn(3)
	out := make([]Node, 0, n)
	for i := 0; i < n; i++ {
		v := base + i
		out = append(out, Node{ID: nodeID(c, i), Title: fmt.Sprintf("the number %d", v), Body: describeNumber(v)})
	}
	return out
}

func describeNumber(v int) string {
	var b strings.Builder
	fmt.Fprintf(&b, "The number %d. ", v)
	if v%2 == 0 {
		b.WriteString("It is even. ")
	} else {
		b.WriteString("It is odd. ")
	}
	div := divisors(v)
	if len(div) == 2 {
		b.WriteString("It is a prime number, divisible only by one and itself. ")
	} else {
		fmt.Fprintf(&b, "It is composite, with divisors %s. ", joinInts(div))
	}
	sum := 0
	for _, d := range div {
		if d != v {
			sum += d
		}
	}
	if sum == v && v > 1 {
		b.WriteString("Its proper divisors sum to itself, so it is a perfect number. ")
	}
	fmt.Fprintf(&b, "Its digit sum is %d. ", digitSum(v))
	if isSquare(v) {
		b.WriteString("It is a perfect square. ")
	}
	if isTriangular(v) {
		b.WriteString("It is a triangular number. ")
	}
	if isFibonacci(v) {
		b.WriteString("It belongs to the Fibonacci sequence. ")
	}
	b.WriteString("An integer, a count, a number in the field of arithmetic.")
	return b.String()
}

// ---------------------------------------------------------------------------
// helpers
// ---------------------------------------------------------------------------

func nodeID(c Coord, i int) string { return fmt.Sprintf("%d.%d.%d", c.X, c.Y, i) }

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}

func divisors(n int) []int {
	if n < 1 {
		return nil
	}
	var d []int
	for i := 1; i <= n; i++ {
		if n%i == 0 {
			d = append(d, i)
		}
	}
	return d
}

func digitSum(n int) int {
	n = abs(n)
	s := 0
	for n > 0 {
		s += n % 10
		n /= 10
	}
	return s
}

func isSquare(n int) bool {
	if n < 0 {
		return false
	}
	r := int(math.Sqrt(float64(n)))
	return r*r == n || (r+1)*(r+1) == n
}

func isTriangular(n int) bool {
	// n triangular iff 8n+1 is a perfect square
	return isSquare(8*n + 1)
}

func isFibonacci(n int) bool {
	return isSquare(5*n*n+4) || isSquare(5*n*n-4)
}

func joinInts(xs []int) string {
	parts := make([]string, len(xs))
	for i, x := range xs {
		parts[i] = fmt.Sprintf("%d", x)
	}
	return strings.Join(parts, ", ")
}