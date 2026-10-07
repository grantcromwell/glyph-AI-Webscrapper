package environment

import (
	"fmt"
	"html"
	"strconv"
	"strings"
)

// URL scheme (served by cmd/world):
//
//	/index.html              -> the spawn chunk hub, Coord{0,0}
//	/chunk/<X>.<Y>.html      -> a chunk hub: its nodes + frontier links
//	/node/<X>.<Y>.<i>.html   -> one place the dog stands in and reads
//
// A "place" page is the unit of exploration: rich body text (so it resonates)
// plus links to its glyph-resonant siblings and the four frontier chunks. The
// dog (via the portal spell) reads the text and is pulled down one link.

func chunkPath(c Coord) string { return fmt.Sprintf("/chunk/%d.%d.html", c.X, c.Y) }
func nodePath(c Coord, i int) string {
	return fmt.Sprintf("/node/%d.%d.%d.html", c.X, c.Y, i)
}

// IndexPage is the world's front door — the spawn chunk hub.
func (w *World) IndexPage() string { return w.ChunkPage(Coord{0, 0}) }

// ChunkPage renders a chunk hub: a short scene-set, links to each place in the
// chunk, and frontier links to the four neighbouring (ungenerated) chunks.
func (w *World) ChunkPage(c Coord) string {
	ch := w.Chunk(c)
	var b strings.Builder
	title := fmt.Sprintf("%s at %d, %d", ch.Biome, c.X, c.Y)
	head(&b, title)
	fmt.Fprintf(&b, "<article><h1>%s</h1>\n", html.EscapeString(title))
	fmt.Fprintf(&b, "<p>You stand in the %s. Around you:</p>\n", html.EscapeString(ch.Biome))
	b.WriteString("<ul>\n")
	for i, n := range ch.Nodes {
		fmt.Fprintf(&b, "<li><a href=\"%s\">%s</a></li>\n", nodePath(c, i), html.EscapeString(n.Title))
	}
	b.WriteString("</ul></article>\n")
	w.frontier(&b, c)
	tail(&b)
	return b.String()
}

// NodePage renders one place: its body, its resonant kin (the scent the dog's
// pull reads), and the frontier out of the chunk.
func (w *World) NodePage(c Coord, i int) string {
	ch := w.Chunk(c)
	var b strings.Builder
	if i < 0 || i >= len(ch.Nodes) {
		head(&b, "void")
		b.WriteString("<article><h1>the void</h1><p>Nothing is here.</p></article>")
		w.frontier(&b, c)
		tail(&b)
		return b.String()
	}
	n := ch.Nodes[i]
	head(&b, n.Title)
	fmt.Fprintf(&b, "<article><h1>%s</h1>\n<pre>%s</pre></article>\n", html.EscapeString(n.Title), html.EscapeString(n.Body))
	b.WriteString("<nav><h2>akin to</h2><ul>\n")
	for _, j := range ch.resonantKin(i, w.K) {
		fmt.Fprintf(&b, "<li><a href=\"%s\">%s</a></li>\n", nodePath(c, j), html.EscapeString(ch.Nodes[j].Title))
	}
	b.WriteString("</ul></nav>\n")
	w.frontier(&b, c)
	tail(&b)
	return b.String()
}

// frontier writes the four cardinal links to neighbour chunks, labelled with the
// neighbour's biome (computed without generating it) so the dog's pull reads
// where each path leads.
func (w *World) frontier(b *strings.Builder, c Coord) {
	b.WriteString("<nav><h2>paths onward</h2><ul>\n")
	for k, nb := range c.Neighbours() {
		fmt.Fprintf(b, "<li><a href=\"%s\">%s to the %s</a></li>\n",
			chunkPath(nb), html.EscapeString(w.biomeName(nb)), cardinal[k])
	}
	b.WriteString("</ul></nav>\n")
}

func head(b *strings.Builder, title string) {
	b.WriteString("<!doctype html><html><head><meta charset=\"utf-8\">")
	fmt.Fprintf(b, "<title>%s</title></head><body>\n", html.EscapeString(title))
}
func tail(b *strings.Builder) { b.WriteString("</body></html>\n") }

// ParseChunk parses "<X>.<Y>" (from /chunk/<X>.<Y>.html).
func ParseChunk(s string) (Coord, bool) {
	parts := strings.Split(strings.TrimSuffix(s, ".html"), ".")
	if len(parts) != 2 {
		return Coord{}, false
	}
	x, e1 := strconv.Atoi(parts[0])
	y, e2 := strconv.Atoi(parts[1])
	if e1 != nil || e2 != nil {
		return Coord{}, false
	}
	return Coord{x, y}, true
}

// ParseNode parses "<X>.<Y>.<i>" (from /node/<X>.<Y>.<i>.html).
func ParseNode(s string) (Coord, int, bool) {
	parts := strings.Split(strings.TrimSuffix(s, ".html"), ".")
	if len(parts) != 3 {
		return Coord{}, 0, false
	}
	x, e1 := strconv.Atoi(parts[0])
	y, e2 := strconv.Atoi(parts[1])
	i, e3 := strconv.Atoi(parts[2])
	if e1 != nil || e2 != nil || e3 != nil {
		return Coord{}, 0, false
	}
	return Coord{x, y}, i, true
}