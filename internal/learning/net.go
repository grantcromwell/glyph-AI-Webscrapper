// Package learning — denoising auto-associator for glyph completion.
//
// DESIGN PATTERN: This is the only learned component in the system. It is a
// 2-layer MLP (21 -> hidden -> 21) trained as a denoising auto-associator over
// the system's own engram glyphs. It learns the manifold of "what real meaning
// looks like" so a partial or noisy query glyph is nudged toward a real memory
// before recall.
//
// This is not a language model. It does not generate prose. It sharpens recall.
// The net is genuinely small (21 -> 8 -> 21 = 365 weights). It is trained on
// the system's own memories, not on external data. The weights are persisted
// as SVG art (the "myelin" heatmap), not as JSON or binary.
//
// Pattern analysis (P28 Hormesis, P25 Constraint Compression):
// The small net size is a hormetic constraint: too few parameters to memorize,
// just enough to learn the manifold. The denoising training (dropping random
// families) forces the net to learn the structure of the glyph space, not the
// specific values. This is constraint compression of the memory manifold.
//
// Key invariants:
//   - 2-layer MLP: input -> ReLU(hidden) -> Sigmoid(output)
//   - Trained as denoising auto-associator (corrupt input, reconstruct clean)
//   - TrainPairs teaches foresight: given a glyph, anticipate the next
//   - Weights persisted as SVG heatmap (not JSON, not binary)
//   - All thresholds from data geometry, never hardcoded
package learning

import (
	"fmt"
	gm "math"
	"math/rand"
	"os"
	"strconv"
	"strings"

	"glyphai/internal/math"
	"glyphai/internal/encoding"
	"glyphai/internal/output"
)

// Net is a 2-layer MLP: input -> ReLU(hidden) -> Sigmoid(output).
type Net struct {
	In, Hidden, Out int
	W1, W2          *math.Mat // W1: hidden×in, W2: out×hidden
	B1, B2          math.Vec
}

// New builds a net with random small weights.
func New(in, hidden, out int, seed int64) *Net {
	r := rand.New(rand.NewSource(seed))
	n := &Net{In: in, Hidden: hidden, Out: out,
		W1: math.NewMat(hidden, in), W2: math.NewMat(out, hidden),
		B1: make(math.Vec, hidden), B2: make(math.Vec, out)}
	scale1 := gm.Sqrt(2.0 / float64(in))
	scale2 := gm.Sqrt(2.0 / float64(hidden))
	for i := range n.W1.Data {
		n.W1.Data[i] = r.NormFloat64() * scale1
	}
	for i := range n.W2.Data {
		n.W2.Data[i] = r.NormFloat64() * scale2
	}
	return n
}

// forward returns the hidden pre-activation, hidden activation, and output.
func (n *Net) forward(x math.Vec) (z1, h, y math.Vec) {
	z1 = math.Add(n.W1.MatVec(x), n.B1)
	h = append(math.Vec(nil), z1...)
	math.ReLU(h)
	z2 := math.Add(n.W2.MatVec(h), n.B2)
	y = make(math.Vec, len(z2))
	for i, v := range z2 {
		y[i] = math.Sigmoid(v)
	}
	return
}

// trainStep does one SGD step on (x -> target) with MSE loss; returns the loss.
func (n *Net) trainStep(x, target math.Vec, lr float64) float64 {
	z1, h, y := n.forward(x)

	// Output error and gradient through sigmoid.
	dz2 := make(math.Vec, n.Out)
	var loss float64
	for i := range y {
		e := y[i] - target[i]
		loss += e * e
		dz2[i] = e * y[i] * (1 - y[i]) // MSE * sigmoid'
	}
	loss /= float64(n.Out)

	// Grad to hidden, through ReLU.
	dh := n.W2.Transpose().MatVec(dz2)
	dz1 := make(math.Vec, n.Hidden)
	for i := range dh {
		if z1[i] > 0 {
			dz1[i] = dh[i]
		}
	}

	// Update W2,b2 (out×hidden) and W1,b1 (hidden×in).
	for o := 0; o < n.Out; o++ {
		row := n.W2.Row(o)
		g := dz2[o]
		for k := 0; k < n.Hidden; k++ {
			row[k] -= lr * g * h[k]
		}
		n.B2[o] -= lr * g
	}
	for hh := 0; hh < n.Hidden; hh++ {
		row := n.W1.Row(hh)
		g := dz1[hh]
		for k := 0; k < n.In; k++ {
			row[k] -= lr * g * x[k]
		}
		n.B1[hh] -= lr * g
	}
	return loss
}

// Train runs denoising auto-association over the given clean glyphs. Each epoch
// corrupts inputs (drops families) and learns to reconstruct the clean encoding.
// Returns the mean loss per epoch so callers can watch it fall.
func (n *Net) Train(glyphs []encoding.Glyph, epochs int, lr, dropRate float64, seed int64) []float64 {
	r := rand.New(rand.NewSource(seed))
	losses := make([]float64, 0, epochs)
	for e := 0; e < epochs; e++ {
		r.Shuffle(len(glyphs), func(i, j int) { glyphs[i], glyphs[j] = glyphs[j], glyphs[i] })
		var sum float64
		for _, g := range glyphs {
			target := toVec(g)
			x := append(math.Vec(nil), target...)
			for i := range x { // corrupt: drop some families to zero
				if r.Float64() < dropRate {
					x[i] = 0
				}
			}
			sum += n.trainStep(x, target, lr)
		}
		if len(glyphs) > 0 {
			losses = append(losses, sum/float64(len(glyphs)))
		}
	}
	return losses
}

// TrainPairs teaches FORESIGHT: given a glyph, anticipate the one that tends to
// follow or associate with it. It trains forward(prev) → next over observed
// transitions. This is the dog's capped frontier — a small net's sense of "if
// this, then that" — weak, but a genuine step from recognition toward inference.
func (n *Net) TrainPairs(pairs [][2]encoding.Glyph, epochs int, lr float64, seed int64) []float64 {
	r := rand.New(rand.NewSource(seed))
	losses := make([]float64, 0, epochs)
	for e := 0; e < epochs; e++ {
		r.Shuffle(len(pairs), func(i, j int) { pairs[i], pairs[j] = pairs[j], pairs[i] })
		var sum float64
		for _, p := range pairs {
			sum += n.trainStep(toVec(p[0]), toVec(p[1]), lr)
		}
		if len(pairs) > 0 {
			losses = append(losses, sum/float64(len(pairs)))
		}
	}
	return losses
}

// Complete runs the learned instinct: a glyph in, a completed glyph out.
func (n *Net) Complete(g encoding.Glyph) encoding.Glyph {
	_, _, y := n.forward(toVec(g))
	var out encoding.Glyph
	for i := 0; i < encoding.Families && i < len(y); i++ {
		out[i] = clamp(int(gm.Round(y[i] * encoding.MaxLevel)))
	}
	return out
}

func toVec(g encoding.Glyph) math.Vec {
	v := make(math.Vec, encoding.Families)
	for i := 0; i < encoding.Families; i++ {
		v[i] = float64(g[i]) / encoding.MaxLevel
	}
	return v
}

func clamp(v int) uint8 {
	if v < 0 {
		return 0
	}
	if v > encoding.MaxLevel {
		return encoding.MaxLevel
	}
	return uint8(v)
}

// Save / Load persist the trained instinct as an SVG scroll — the weights drawn
// as a heatmap (the dog's "myelin") with the exact numbers embedded. No JSON.

func (n *Net) Save(path string) error {
	body := n.heatmap()
	svg := output.Inscribe("cerebellum", "Argos · learned instinct (myelin)", n.encode(), body)
	return os.WriteFile(path, []byte(svg), 0644)
}

func Load(path string) (*Net, error) {
	raw, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}
	payload, err := output.ReadScroll(raw, "cerebellum")
	if err != nil {
		return nil, err
	}
	lines := strings.Split(payload, "\n")
	if len(lines) < 5 {
		return nil, fmt.Errorf("cerebellum: short scroll")
	}
	var in, hid, out int
	fmt.Sscanf(lines[0], "%d %d %d", &in, &hid, &out)
	n := &Net{In: in, Hidden: hid, Out: out,
		W1: &math.Mat{Rows: hid, Cols: in, Data: parseFloats(lines[1])},
		B1: parseFloats(lines[2]),
		W2: &math.Mat{Rows: out, Cols: hid, Data: parseFloats(lines[3])},
		B2: parseFloats(lines[4])}
	return n, nil
}

func (n *Net) encode() string {
	var b strings.Builder
	fmt.Fprintf(&b, "%d %d %d\n", n.In, n.Hidden, n.Out)
	b.WriteString(floats(n.W1.Data) + "\n")
	b.WriteString(floats(n.B1) + "\n")
	b.WriteString(floats(n.W2.Data) + "\n")
	b.WriteString(floats(n.B2))
	return b.String()
}

// heatmap draws W1 (hidden×in) as art: warm cells = positive weights, cool =
// negative, brightness = magnitude.
func (n *Net) heatmap() string {
	var b strings.Builder
	const x0, y0, cw, ch = 28.0, 78.0, 27.0, 10.0
	maxAbs := 1e-9
	for _, w := range n.W1.Data {
		if a := gm.Abs(w); a > maxAbs {
			maxAbs = a
		}
	}
	for r := 0; r < n.Hidden; r++ {
		for c := 0; c < n.In; c++ {
			w := n.W1.At(r, c)
			hue := 205.0
			if w > 0 {
				hue = 25.0
			}
			light := 12 + int(gm.Abs(w)/maxAbs*55)
			fmt.Fprintf(&b, `<rect x="%.1f" y="%.1f" width="%.1f" height="%.1f" fill="hsl(%.0f,70%%,%d%%)"/>`+"\n",
				x0+float64(c)*cw, y0+float64(r)*ch, cw-0.5, ch-0.5, hue, light)
		}
	}
	fmt.Fprintf(&b, `<text x="28" y="420" fill="#7a849a" font-size="10">W1 %d×%d · warm=+ cool=− · trained instinct</text>`+"\n", n.Hidden, n.In)
	return b.String()
}

func floats(v math.Vec) string {
	parts := make([]string, len(v))
	for i, x := range v {
		parts[i] = strconv.FormatFloat(x, 'g', -1, 64)
	}
	return strings.Join(parts, " ")
}

func parseFloats(line string) math.Vec {
	fields := strings.Fields(line)
	v := make(math.Vec, len(fields))
	for i, f := range fields {
		v[i], _ = strconv.ParseFloat(f, 64)
	}
	return v
}