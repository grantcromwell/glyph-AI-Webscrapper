// Package world is the GAME — a separate, headless Minecraft-style world the dog
// plays. It is NOT a brain organ: it lives outside internal/ and imports only the
// shared glyph substrate (internal/glyph). It owns 100% of worldgen. The dog never
// imports this package; it reaches the world only through the portal spell.
//
// Blueprint = Minecraft (1.18+):
//
//	world seed (one number)        -> a glyph seed (a word), deterministic
//	multi-noise climate space      -> the logogram plane (glyph/coord.go)
//	a biome owns a climate region  -> the biome with the strongest anchor family
//	chunk (16x16), setSeed(seed^p) -> a coordinate cell, rng seeded by hash(seed,coord)
//	terrain->features->structures  -> climate -> fill -> link -> frontier
//	chunks load as you walk        -> frontier links generate the next chunk on visit
//
// Biome borders are not hardcoded adjacency: smooth value-noise over the chunk
// grid decides which anchor family is strongest where, so regions emerge.
package environment

import (
	"hash/fnv"
	"math"
	"math/rand"
	"sort"

	"glyphai/internal/encoding"
)

// Coord is an integer chunk position on the infinite grid.
type Coord struct{ X, Y int }

// Node is one place the dog can stand in and read — the unit of exploration.
type Node struct {
	ID    string
	Title string
	Body  string
}

// Biome is the modpack hook: add a domain to the world by adding a Biome.
// Anchor is the biome's home family on the 21-axis plane; Fill emits a chunk's
// content deterministically from a seeded rng.
type Biome interface {
	Name() string
	Anchor() int // family index 0..encoding.Families-1
	Fill(rng *rand.Rand, c Coord) []Node
}

// World is the game state: a seed, the link fan-out k, and the registered biomes.
type World struct {
	Seed   uint64
	K      int
	biomes []Biome
}

// New builds a world from a seed word. The biomes are the v1 modpack.
func New(seedWord string, k int) *World {
	if k <= 0 {
		k = 8
	}
	return &World{
		Seed:   fnvWord(seedWord),
		K:      k,
		biomes: []Biome{Colors{}, Materials{}, Numerology{}},
	}
}

// Chunk is a generated cell: which biome owns it and the nodes it holds, with
// each node's coarse glyph cached so links can be wired by resonance.
type Chunk struct {
	Coord  Coord
	Biome  string
	Nodes  []Node
	glyphs []encoding.Glyph
}

// Chunk generates (deterministically) the cell at c. Same seed+coord => identical.
func (w *World) Chunk(c Coord) Chunk {
	b := w.biomeFor(c)
	rng := rand.New(rand.NewSource(int64(chunkSeed(w.Seed, c))))
	nodes := b.Fill(rng, c)
	gl := make([]encoding.Glyph, len(nodes))
	for i := range nodes {
		gl[i] = encoding.EncodeGlyph(nodes[i].Title + " " + nodes[i].Body)
	}
	return Chunk{Coord: c, Biome: b.Name(), Nodes: nodes, glyphs: gl}
}

// biomeName cheaply reports which biome owns a neighbour cell, WITHOUT filling it
// (so frontier links can be labelled without cascading generation).
func (w *World) biomeName(c Coord) string { return w.biomeFor(c).Name() }

// biomeFor picks the biome whose anchor family is strongest in the chunk's
// climate — argmax, so every cell has exactly one biome (no gaps, no overlap).
func (w *World) biomeFor(c Coord) Biome {
	lv := w.climate(c)
	best, bestV := w.biomes[0], lv[w.biomes[0].Anchor()]
	for _, b := range w.biomes[1:] {
		if v := lv[b.Anchor()]; v > bestV {
			best, bestV = b, v
		}
	}
	return best
}

// climate is the multi-noise field: per family, a smooth value over the chunk
// grid. Low frequencies => biomes span several chunks (contiguous regions).
func (w *World) climate(c Coord) [encoding.Families]float64 {
	var lv [encoding.Families]float64
	x, y := float64(c.X), float64(c.Y)
	for f := 0; f < encoding.Families; f++ {
		// per-family phases drawn from the seed so each family's region differs
		p1 := float64((w.Seed>>uint(f%64))&0xff) / 255 * 2 * math.Pi
		p2 := float64((w.Seed>>uint((f*7)%64))&0xff) / 255 * 2 * math.Pi
		v := math.Sin(x*0.31+p1) + math.Sin(y*0.27+p2) + math.Sin((x+y)*0.19+p1+p2)
		lv[f] = (v + 3) / 6 // -> 0..1
	}
	return lv
}

// Neighbours returns the four cardinal frontier cells (the world's edges).
func (c Coord) Neighbours() []Coord {
	return []Coord{{c.X, c.Y - 1}, {c.X + 1, c.Y}, {c.X, c.Y + 1}, {c.X - 1, c.Y}}
}

var cardinal = []string{"north", "east", "south", "west"}

// resonantKin returns the indices of node i's top-k glyph-resonant siblings —
// the dog's own sense of kinship deciding the link graph (lifted from fable5site).
func (ch Chunk) resonantKin(i, k int) []int {
	type sc struct {
		j int
		r float64
	}
	scores := make([]sc, 0, len(ch.glyphs))
	for j := range ch.glyphs {
		if j == i {
			continue
		}
		scores = append(scores, sc{j, encoding.Resonance(ch.glyphs[i], ch.glyphs[j])})
	}
	sort.Slice(scores, func(a, b int) bool { return scores[a].r > scores[b].r })
	if k > len(scores) {
		k = len(scores)
	}
	out := make([]int, k)
	for n := 0; n < k; n++ {
		out[n] = scores[n].j
	}
	return out
}

func chunkSeed(world uint64, c Coord) uint64 {
	const prime = 1099511628211
	h := world
	h = (h * prime) ^ uint64(int64(c.X))
	h = (h * prime) ^ uint64(int64(c.Y))
	h ^= h >> 33
	return h
}

func fnvWord(s string) uint64 {
	if s == "" {
		s = "glyphai"
	}
	h := fnv.New64a()
	_, _ = h.Write([]byte(s))
	return h.Sum64()
}