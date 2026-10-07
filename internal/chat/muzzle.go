// Package muzzle is Argos's mouth: the live chat surface. It owns presentation
// and the interactive loop only — the cognition is handed to it as a Mind that
// turns a query into a Reflection (what the dog perceived, thought in logograms,
// and chose to say). Pure stdlib, no UI dependencies — it runs anywhere.
package chat

import (
	"bufio"
	"fmt"
	"os"
	"strings"
)

// ANSI colours (terminals that don't support them just show plain text).
const (
	reset  = "\033[0m"
	dim    = "\033[2m"
	bold   = "\033[1m"
	cyan   = "\033[36m"
	yellow = "\033[33m"
	green  = "\033[32m"
	grey   = "\033[90m"
)

// Reflection is one turn of thought, produced by the Mind.
type Reflection struct {
	Query     string
	GlyphHex  string
	Reading   string
	Logogram  string
	Thought   []string // per-word "word sign"
	Instinct  string    // cerebellum's completion reading ("" if untrained)
	Senses    string    // sound + meaning-depth lobe readings
	Settle    []float64 // the field-settling trajectory (resonance per step)
	Coherence float64   // whether the field locked into a memory (0..1)
	Veto      bool      // the dog stayed silent rather than guess
	Answer    string
	Coined    string // logogram coined this turn ("" if none)
}

// Mind is the cognition the muzzle speaks for.
type Mind interface {
	Perceive(query string) Reflection
	Count() int
}

// Render prints a reflection with colour: the dog thinking, then speaking.
func Render(r Reflection) {
	fmt.Printf("\n%s🐾 perceives%s %q\n", bold, reset, r.Query)
	fmt.Printf("   %sglyph%s    %s\n", grey, reset, r.GlyphHex)
	fmt.Printf("   %sreading%s  %s%s%s\n", grey, reset, cyan, r.Reading, reset)
	fmt.Printf("   %slogogram%s %s%s%s\n", grey, reset, yellow, r.Logogram, reset)
	if len(r.Thought) > 0 {
		fmt.Printf("   %sthinks%s   %s\n", grey, reset, strings.Join(r.Thought, "  ·  "))
	}
	if r.Instinct != "" {
		fmt.Printf("   %sinstinct%s %s%s%s\n", grey, reset, dim, r.Instinct, reset)
	}
	if r.Senses != "" {
		fmt.Printf("   %ssenses%s   %s%s%s\n", grey, reset, dim, r.Senses, reset)
	}
	if len(r.Settle) > 0 {
		steps := make([]string, len(r.Settle))
		for i, s := range r.Settle {
			steps[i] = fmt.Sprintf("%.2f", s)
		}
		fmt.Printf("   %sfield%s    settling %s\n", grey, reset, strings.Join(steps, " → "))
		fmt.Printf("   %sconviction%s %s%.2f%s %s\n", grey, reset, bold, r.Coherence, reset, coherenceMark(r.Coherence))
	}
	fmt.Printf("\n%s%s%s\n", green, r.Answer, reset)
	if r.Coined != "" {
		fmt.Printf("%s   · coined %s%s\n", grey, r.Coined, reset)
	}
}

// coherenceMark is a felt quality, not a gate — the dog always speaks.
func coherenceMark(c float64) string {
	switch {
	case c >= 0.45:
		return "(sure)"
	case c >= 0.25:
		return "(holding)"
	default:
		return "(faint)"
	}
}

// Run is the interactive chat loop. Type to talk; :q or Ctrl-D to leave.
func Run(m Mind) {
	fmt.Printf("%s🐾 Argos · muzzle%s  %s(%d engrams in memory)%s\n", bold, reset, grey, m.Count(), reset)
	fmt.Printf("%stype to speak with the dog · :q to leave%s\n", grey, reset)
	sc := bufio.NewScanner(os.Stdin)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for {
		fmt.Printf("\n%syou ❯%s ", cyan, reset)
		if !sc.Scan() {
			break // EOF / Ctrl-D
		}
		line := strings.TrimSpace(sc.Text())
		if line == "" {
			continue
		}
		if line == ":q" || line == "exit" || line == "quit" {
			break
		}
		Render(m.Perceive(line))
	}
	fmt.Printf("\n%s🐾 the dog curls up.%s\n", grey, reset)
}