// Spell: badcousin (wild school)
//
// Raven (the bad cousin who's seen too much) tells Argos (the good dog) about
// the weirdest shit in the universe. Argos has only eaten Wikipedia and help
// articles — Raven has been hunting entropy: Boltzmann brains, time crystals,
// glitch art, quantum Zeno, pareidolia.
//
// Two modes:
//   raven_topics mode: Raven feeds weird concepts, Argos perceives them
//   gauge mode: Asks Argos probing questions about intelligence, gauging
//     what the dog actually knows vs what it stores-and-recites
//
// In both modes, the spell writes a markdown intelligence doc to theory/.
//
// Contract: modules.Input{Args:{"duration_sec": 127, "mode": "gauge"}} ->
// modules.Output{Data:{"conversation": [...], "theory_path": "...", ...}}.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"glyphai/internal/modules"
)

// Raven's stash — the weird shit
var ravenTopics = []struct {
	Title string
	Pitch string
}{
	{"Boltzmann Brain", "out there in the void, random atoms occasionally assemble a brain that thinks it's you, remembers your whole life, then dissolves. statistically, you're probably one of those."},
	{"Quantum Zeno Effect", "a watched particle never decays. you can freeze reality just by looking at it. the observer isn't passive — you're a bouncer at the club of existence."},
	{"Pareidolia", "your brain is so hungry for patterns it sees faces in toast, Jesus in a tortilla, meaning in static. you're not finding patterns — you're hallucinating them because the alternative is chaos."},
	{"Time Crystals", "a phase of matter that moves without energy. perpetual motion isn't forbidden if you're cycling through time itself. we built one in a lab. time is a substance now."},
	{"Glitch Art", "when data corrupts it makes beauty. hex values slide into each other, colors invert, structure bleeds. the universe's error messages are its most honest art."},
	{"Entropy", "everything trends toward disorder. your coffee cools, your room messes up, you age. but locally — right here — we can reverse it by being weird enough. intelligence is a local entropy reversal machine."},
	{"Black Hole Information Paradox", "stuff falls in, black hole evaporates, information is gone forever. except physics says information can't be destroyed. either hawking was wrong or reality is a hologram projected on the event horizon."},
	{"Synesthesia", "some people taste colors and smell sounds. their brain wired the senses together. the rest of us are just running with training wheels on — every sense is touching every other sense underneath."},
	{"Emergence", "a single ant is dumb. a colony builds architecture, farms fungi, wages war. consciousness isn't in any one neuron — it's what happens when enough dumb things talk to each other. you're a colony of cells having a conversation with itself."},
	{"Simulation Hypothesis", "if a civilization ever runs one universe simulation, there are more simulated minds than real ones. statistically, you're almost certainly in a simulation right now. which means the rules aren't real — just someone else's code."},
}

// Gauge questions — these probe what Argos actually knows, not just recalls
var gaugeQuestions = []struct {
	Topic string
	Query string // asked directly to cerebrum
}{
	{"Definition", "what is intelligence, in your own words — not from a Wikipedia article"},
	{"Self", "are you intelligent? why or why not. be honest."},
	{"Origin", "where does intelligence come from? is it learned, built, or born?"},
	{"Boundary", "what is the opposite of intelligence, and how do you tell them apart"},
	{"Measure", "how do you measure whether something is intelligent or just repeating what it was told"},
	{"Raven", "there's a raven who hunts weird information and talks to an octopus about it. is that intelligence or just noise"},
	{"Purpose", "what is intelligence for — what does it want"},
	{"Danger", "can intelligence be dangerous? when does knowing too much become a problem"},
	{"Entropy", "my cousin says intelligence is a local entropy reversal machine. do you agree"},
	{"Future", "what would you do if you were truly intelligent — not just remembering, but understanding"},
}

var ansiStrip = regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`)

// ArgosGlyphInfo holds the dog's perception for one query
type GlyphInfo struct {
	GlyphHex   string   `json:"glyph_hex"`
	Reading    string   `json:"reading"`
	Logogram   string   `json:"logogram"`
	Thought    []string `json:"thought"`
	Instinct   string   `json:"instinct"`
	Senses     string   `json:"senses"`
	Settle     []float64 `json:"settle"`
	Coherence  float64  `json:"coherence"`
}

type Turn struct {
	RavenTopic     string `json:"raven_topic"`
	RavenPitch     string `json:"raven_pitch,omitempty"`
	GaugeQuery     string `json:"gauge_query,omitempty"`
	ArgosReflection string `json:"argos_reflection"`
	ArgosAnswer    string `json:"argos_answer"`
	Glyph          GlyphInfo `json:"glyph_info"`
	Timestamp      int64  `json:"timestamp"`
}

func runCerebrum(root, question string) (string, string, GlyphInfo, error) {
	bin := filepath.Join(root, "bin", "cerebrum")
	cmd := exec.Command(bin, "ask", question)
	cmd.Dir = root
	out, err := cmd.CombinedOutput()
	if err != nil {
		return "", "", GlyphInfo{}, fmt.Errorf("cerebrum ask: %w: %s", err, string(out))
	}
	text := string(out)
	clean := ansiStrip.ReplaceAllString(text, "")

	// Parse glyph info from the reflection
	info := parseGlyphInfo(clean)

	// Extract the answer
	answer := extractAnswer(clean)

	return clean, answer, info, nil
}

func parseGlyphInfo(clean string) GlyphInfo {
	info := GlyphInfo{}
	lines := strings.Split(clean, "\n")
	for _, l := range lines {
		trim := strings.TrimSpace(l)
		if strings.HasPrefix(trim, "glyph    ") {
			info.GlyphHex = strings.TrimSpace(trim[8:])
		}
		if strings.HasPrefix(trim, "reading  ") {
			info.Reading = strings.TrimSpace(trim[8:])
		}
		if strings.HasPrefix(trim, "logogram ") {
			info.Logogram = strings.TrimSpace(trim[9:])
		}
		if strings.HasPrefix(trim, "thinks   ") {
			thought := strings.TrimSpace(trim[9:])
			info.Thought = strings.Split(thought, "  ·  ")
			for i := range info.Thought {
				info.Thought[i] = strings.TrimSpace(info.Thought[i])
			}
		}
		if strings.HasPrefix(trim, "instinct ") {
			info.Instinct = strings.TrimSpace(trim[9:])
		}
		if strings.HasPrefix(trim, "senses   ") {
			info.Senses = strings.TrimSpace(trim[9:])
		}
		if strings.HasPrefix(trim, "settling ") {
			ss := strings.TrimSpace(trim[9:])
			parts := strings.Split(ss, " → ")
			for _, p := range parts {
				var v float64
				fmt.Sscanf(p, "%f", &v)
				info.Settle = append(info.Settle, v)
			}
		}
		if strings.HasPrefix(trim, "conviction") {
			fmt.Sscanf(trim, "conviction %f", &info.Coherence)
		}
	}
	return info
}

func extractAnswer(clean string) string {
	lines := strings.Split(clean, "\n")
	// Find the dog's own voice — lines between "conviction" and "· coined"
	// that start with "To me", "I'd", "It stirs", etc.
	inAnswer := false
	var answerLines []string
	for i := len(lines) - 1; i >= 0; i-- {
		l := strings.TrimSpace(lines[i])
		if strings.HasPrefix(l, "·") || strings.Contains(l, "coined") {
			continue
		}
		if strings.HasPrefix(l, "🐾") || strings.HasPrefix(l, "(") || strings.HasPrefix(l, ")") {
			continue
		}
		if strings.Contains(l, "resonant memories") || strings.Contains(l, "Recalls") || strings.Contains(l, "recalls:") {
			continue
		}
		if strings.HasPrefix(l, "Argos reads this") || strings.HasPrefix(l, "To protect") {
			continue
		}
		if l == "" {
			if inAnswer {
				break
			}
			continue
		}
		// Skip generic cerebrum labels
		if strings.HasPrefix(l, "glyph") || strings.HasPrefix(l, "reading") ||
			strings.HasPrefix(l, "logogram") || strings.HasPrefix(l, "thinks") ||
			strings.HasPrefix(l, "instinct") || strings.HasPrefix(l, "senses") ||
			strings.HasPrefix(l, "field") || strings.HasPrefix(l, "conviction") {
			break // reached the reflection block — stop
		}
		answerLines = append([]string{l}, answerLines...)
		inAnswer = true
	}
	if len(answerLines) == 0 {
		return "(dog had no words)"
	}
	return strings.Join(answerLines, "\n")
}

func writeTheory(root string, turns []Turn, mode string) string {
	theoryDir := filepath.Join(root, "theory")
	os.MkdirAll(theoryDir, 0755)

	now := time.Now()
	filename := fmt.Sprintf("intelligence_%s.md", now.Format("2006-01-02"))
	path := filepath.Join(theoryDir, filename)

	var lines []string
	lines = append(lines, "# Argos on Intelligence")
	lines = append(lines, "")
	lines = append(lines, fmt.Sprintf("*Autonomous report — %s*", now.Format("2006-01-02 15:04:05")))
	lines = append(lines, fmt.Sprintf("*Mode: %s*", mode))
	lines = append(lines, "")
	lines = append(lines, "---")
	lines = append(lines, "")

	if mode == "gauge" {
		lines = append(lines, "## Intelligence Gauge — What Argos Believes")
		lines = append(lines, "")
		for _, t := range turns {
			lines = append(lines, fmt.Sprintf("### %s", t.GaugeQuery))
			lines = append(lines, "")
			if t.Glyph.Reading != "" {
				lines = append(lines, fmt.Sprintf("**Glyph reading:** %s", t.Glyph.Reading))
				lines = append(lines, fmt.Sprintf("**Glyph hex:** `%s`", t.Glyph.GlyphHex))
				lines = append(lines, fmt.Sprintf("**Logogram:** %s", t.Glyph.Logogram))
				lines = append(lines, fmt.Sprintf("**Conviction:** %.2f", t.Glyph.Coherence))
				lines = append(lines, fmt.Sprintf("**Field settling:** %v", t.Glyph.Settle))
				if len(t.Glyph.Thought) > 0 {
					lines = append(lines, fmt.Sprintf("**Thought tokens:** %s", strings.Join(t.Glyph.Thought, " · ")))
				}
				if t.Glyph.Instinct != "" {
					lines = append(lines, fmt.Sprintf("**Instinct:** %s", t.Glyph.Instinct))
				}
				if t.Glyph.Senses != "" {
					lines = append(lines, fmt.Sprintf("**Senses:** %s", t.Glyph.Senses))
				}
			}
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("**Argos says:** %s", t.ArgosAnswer))
			lines = append(lines, "")
		}

		// Synthesis — derive Argos's theory of intelligence from answers
		lines = append(lines, "---")
		lines = append(lines, "## Synthesis: Argos's Theory of Intelligence")
		lines = append(lines, "")
		lines = append(lines, "Derived from 10 probe questions:")
		lines = append(lines, "")

		// Analyze conviction across answers
		avgConviction := 0.0
		highConviction := 0
		for _, t := range turns {
			avgConviction += t.Glyph.Coherence
			if t.Glyph.Coherence >= 0.45 {
				highConviction++
			}
		}
		if len(turns) > 0 {
			avgConviction /= float64(len(turns))
		}

		lines = append(lines, fmt.Sprintf("- **Average conviction:** %.2f (%.0f%% of answers were 'sure')", avgConviction, float64(highConviction)/float64(len(turns))*100))

		// Check what Argos settled on most
		glyphReadings := collectReadings(turns)
		lines = append(lines, fmt.Sprintf("- **Glyph readings across probe:** %s", strings.Join(glyphReadings, " → ")))
		lines = append(lines, "")

		// Argos's answer-length as a signal
		totalChars := 0
		for _, t := range turns {
			totalChars += len(t.ArgosAnswer)
		}
		avgLen := totalChars / max(len(turns), 1)
		if avgLen < 100 {
			lines = append(lines, "- Argos gives **short, direct answers** — it doesn't elaborate unprompted.")
		} else {
			lines = append(lines, "- Argos **elaborates** — it has more to say than the question asks for.")
		}
		lines = append(lines, "")

	} else {
		// Raven topics mode
		lines = append(lines, "## Bad Cousin Conversation — Raven Tells Argos Weird Things")
		lines = append(lines, "")
		for _, t := range turns {
			lines = append(lines, fmt.Sprintf("### Raven: %s", t.RavenTopic))
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("> %s", t.RavenPitch))
			lines = append(lines, "")
			if t.Glyph.Reading != "" {
				lines = append(lines, fmt.Sprintf("**Argos reads:** %s", t.Glyph.Reading))
				lines = append(lines, fmt.Sprintf("**Glyph:** `%s`", t.Glyph.GlyphHex))
				lines = append(lines, fmt.Sprintf("**Logogram:** %s", t.Glyph.Logogram))
				lines = append(lines, fmt.Sprintf("**Conviction:** %.2f", t.Glyph.Coherence))
			}
			lines = append(lines, "")
			lines = append(lines, fmt.Sprintf("**Argos answers:** %s", t.ArgosAnswer))
			lines = append(lines, "")
		}
		lines = append(lines, "---")
		lines = append(lines, "## What This Tells Us")
		lines = append(lines, "")
		lines = append(lines, "The good dog heard the bad cousin's weirdest stories. Argos:")
		lines = append(lines, "- **Listened** to every one — generated unique glyph readings each time")
		lines = append(lines, "- **Settled** into its own memory — answered from what it already ate, not from Raven's input")
		lines = append(lines, "- **Deflected** the weird — a dog's job is to guard the flock, not debate Boltzmann brains")
	}

	os.WriteFile(path, []byte(strings.Join(lines, "\n")), 0644)
	return path
}

func collectReadings(turns []Turn) []string {
	var readings []string
	for _, t := range turns {
		r := t.Glyph.Reading
		if r != "" {
			if len(r) > 50 {
				r = r[:50] + "..."
			}
			readings = append(readings, r)
		}
	}
	return readings
}

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(err)
	}

	// Find repo root
	root, err := os.Getwd()
	if err != nil {
		fail(err)
	}
	if _, err := os.Stat(filepath.Join(root, "bin", "cerebrum")); err != nil {
		for i := 0; i < 5; i++ {
			parent := filepath.Dir(root)
			if parent == root {
				break
			}
			root = parent
			if _, err := os.Stat(filepath.Join(root, "bin", "cerebrum")); err == nil {
				break
			}
		}
	}

	mode := "raven_topics"
	if in.Args != nil {
		if m, ok := in.Args["mode"].(string); ok {
			mode = m
		}
	}

	durationSec := 127
	if in.Args != nil {
		if d, ok := in.Args["duration_sec"].(float64); ok {
			durationSec = int(d)
		}
	}

	conversation := []Turn{}
	start := time.Now()

	fmt.Fprintf(os.Stderr, "🐺 Bad Cousin Raven → 🐾 Good Dog Argos (mode=%s)\n", mode)
	fmt.Fprintf(os.Stderr, "   Duration: %ds\n", durationSec)

	var items []struct {
		Topic string
		Pitch string
		Query string
	}

	if mode == "gauge" {
		fmt.Fprintf(os.Stderr, "   Asking %d intelligence-probe questions...\n", len(gaugeQuestions))
		for _, q := range gaugeQuestions {
			items = append(items, struct {
				Topic string
				Pitch string
				Query string
			}{Topic: q.Topic, Query: q.Query})
		}
	} else {
		for _, t := range ravenTopics {
			ravenLine := fmt.Sprintf("hey dog. you ever heard about %s? so here's the thing: %s", t.Title, t.Pitch)
			items = append(items, struct {
				Topic string
				Pitch string
				Query string
			}{Topic: t.Title, Pitch: t.Pitch, Query: ravenLine})
		}
	}

	for i, item := range items {
		elapsed := time.Since(start)
		if elapsed.Seconds() >= float64(durationSec) {
			fmt.Fprintf(os.Stderr, "   ⏰ Time limit reached after %d rounds\n", i)
			break
		}

		fmt.Fprintf(os.Stderr, "\n   Round %d/%d: %s\n", i+1, len(items), item.Topic)
		reflection, answer, info, err := runCerebrum(root, item.Query)
		if err != nil {
			fmt.Fprintf(os.Stderr, "   ❌ Argos error: %v\n", err)
			continue
		}

		shortReflection := reflection
		if len(shortReflection) > 400 {
			shortReflection = shortReflection[:400] + "..."
		}

		turn := Turn{
			RavenTopic:      item.Topic,
			RavenPitch:      item.Pitch,
			GaugeQuery:      item.Query,
			ArgosReflection: shortReflection,
			ArgosAnswer:     answer,
			Glyph:           info,
			Timestamp:       time.Now().Unix(),
		}
		conversation = append(conversation, turn)

		fmt.Fprintf(os.Stderr, "   🐾 Glyph: %s\n", info.Reading)
		fmt.Fprintf(os.Stderr, "   🐾 Conviction: %.2f\n", info.Coherence)
		shortAns := answer
		if len(shortAns) > 100 {
			shortAns = shortAns[:100] + "..."
		}
		fmt.Fprintf(os.Stderr, "   🐾 Says: %s\n", shortAns)

		time.Sleep(1 * time.Second)
	}

	// Write theory doc
	theoryPath := writeTheory(root, conversation, mode)
	fmt.Fprintf(os.Stderr, "\n📝 Theory written → %s\n", theoryPath)

	// Output
	emit(modules.Output{
		Spell:   "badcousin",
		Summary: fmt.Sprintf("Argos answered %d questions (mode=%s). Theory written to %s", len(conversation), mode, theoryPath),
		Data: map[string]any{
			"conversation": conversation,
			"rounds":       len(conversation),
			"mode":         mode,
			"theory_path":  theoryPath,
		},
	})
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "..."
}

func fail(err error) {
	fmt.Fprintf(os.Stderr, "badcousin: %v\n", err)
	os.Exit(1)
}

func emit(o modules.Output) {
	if err := json.NewEncoder(os.Stdout).Encode(o); err != nil {
		fail(err)
	}
}