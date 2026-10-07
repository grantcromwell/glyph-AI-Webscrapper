// Spell: pronounce (tongue school) — Argos speaks.
//
// The brain produces pure glyph state: readings, emotions, recall hits.
// This spell turns that state into plain English in the dog's own voice.
// Not a report — a perspective. The dog feels what it knows and says
// what it means, not what it memorized.
//
// Input (as JSON on stdin): modules.Input{
//   Text: "original question",
//   Args: {
//     "glyph_hex":  "020304...",
//     "reading":    "blazing sound, deep depth",
//     "emotion":    "curious|wary|calm|excited",
//     "curiosity":  0.73,
//     "hits":       [{topic, glyph, score, source, text}, ...],
//     "lexicon_path": "weights/vocabulary.svg"
//   }
// }
package main

import (
	"encoding/json"
	"fmt"
	"math"
	"os"
	"regexp"
	"strings"
	"unicode"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"glyphai/internal/vocabulary"
)

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(fmt.Errorf("reading input: %v", err))
	}

	query := in.Text
	if query == "" {
		fail(fmt.Errorf("pronounce needs a query"))
	}

	args := in.Args
	if args == nil {
		args = map[string]any{}
	}

	emotion, _ := args["emotion"].(string)
	curiosity, _ := args["curiosity"].(float64)

	// Decode hits into structured data
	type Hit struct {
		Topic  string
		Glyph  string
		Score  float64
		Source string
		Text   string
	}
	var hits []Hit
	if h, ok := args["hits"].([]any); ok {
		for _, item := range h {
			if m, ok := item.(map[string]any); ok {
				hits = append(hits, Hit{
					Topic:  str(m["topic"]),
					Glyph:  str(m["glyph"]),
					Score:  flt(m["score"]),
					Source: str(m["source"]),
					Text:   str(m["text"]),
				})
			}
		}
	}

	lexPath := "weights/vocabulary.svg"
	if lp, ok := args["lexicon_path"].(string); ok && lp != "" {
		lexPath = lp
	}

	lx, _ := vocabulary.Open(lexPath)

	// Encode the query for resonance matching
	qd := encoding.Encode(query)
	qg := qd.Coarse()

	// ——— Gather material ———

	// Find best lexicon match — but skip self-referential signs about Argos/Raven
	var bestSign string
	var bestVerse string
	bestScore := 0.0
	if lx != nil {
		for _, entry := range lx.Signs {
			if entry.Glyph == "" {
				continue
			}
			eg, err := encoding.ParseHex(entry.Glyph)
			if err != nil {
				continue
			}
			r := encoding.Resonance(qg, eg)
			// Skip self-referential signs
			lower := strings.ToLower(entry.Concept)
			isSelfRef := strings.Contains(lower, "raven is") ||
				strings.Contains(lower, "argos is") ||
				strings.Contains(lower, "147-glyph") ||
				strings.Contains(lower, "frequency mind") ||
				strings.Contains(lower, "convergence mind") ||
				strings.Contains(lower, "spatial mind") ||
				strings.Contains(lower, "21 families") ||
				strings.Contains(lower, "what do you want to say")
			if isSelfRef {
				continue
			}
			if r > bestScore {
				bestScore = r
				bestSign = entry.Concept
				if len(entry.Verses) > 0 {
					bestVerse = entry.Verses[0]
				}
			}
		}
	}

	// 2. Extract clean knowledge sentences from hits
	var knowledge []string
	for _, hit := range hits {
		s := extractKnowledge(hit.Text)
		if s != "" {
			knowledge = append(knowledge, s)
		}
		if len(knowledge) >= 3 {
			break
		}
	}

	// 3. Gather related topics (distinct from the main one, not self-referential)
	var related []string
	topTopic := ""
	if len(hits) > 0 {
		topTopic = clean(hits[0].Topic)
	}
	seen := map[string]bool{strings.ToLower(topTopic): true}
	selfBlock := []string{"raven has", "argos is", "147-glyph", "signal intelligence", "vertical lobe", "neurons per family"}
	for _, hit := range hits {
		t := clean(hit.Topic)
		if t == "" || seen[strings.ToLower(t)] || len(related) >= 2 {
			continue
		}
		// Skip self-referential topics
		lower := strings.ToLower(t)
		bad := false
		for _, sb := range selfBlock {
			if strings.Contains(lower, sb) {
				bad = true
				break
			}
		}
		if !bad {
			seen[strings.ToLower(t)] = true
			related = append(related, t)
		}
	}

	// ——— Compose the answer in the dog's voice ———

	answer := composeVoice(query, emotion, curiosity, topTopic, knowledge, related, bestSign, bestVerse, bestScore)

	emit(modules.Output{
		Spell:   "pronounce",
		Detail:  qd,
		Summary: answer,
	})
}

// composeVoice builds a natural English answer from the dog's perspective.
func composeVoice(query, emotion string, curiosity float64, topTopic string, knowledge []string, related []string, bestSign, bestVerse string, lexScore float64) string {
	emo := emotionVoice(emotion, curiosity)
	hasKnowledge := len(knowledge) > 0
	hasLexicon := bestSign != "" && lexScore >= 0.4

	// No knowledge at all
	if !hasKnowledge && !hasLexicon {
		return "I don't know that one yet."
	}

	var parts []string

	if hasKnowledge {
		parts = append(parts, emo)
		first := knowledge[0]
		// Trim to first sentence
		if idx := strings.Index(first, ". "); idx > 0 && idx < len(first)-2 {
			first = first[:idx]
		}
		parts = append(parts, first)
	} else if hasLexicon {
		parts = append(parts, fmt.Sprintf("%s I know this as \"%s\"", emo, bestSign))
	}

	// One connection max
	if len(related) > 0 {
		parts = append(parts, fmt.Sprintf("reminds me of %s", related[0]))
	}

	return strings.Join(parts, ". ") + "."
}

// emotionVoice returns how the dog opens based on what it feels.
// Natural, not a menu selection.
func emotionVoice(emotion string, curiosity float64) string {
	emo := strings.ToLower(emotion)
	switch emo {
	case "curious":
		return "That stirs something in me"
	case "wary":
		return "I'm cautious about this"
	case "calm":
		return "This sits well with me"
	case "excited":
		return "This excites me"
	case "quiet":
		return "I'm quiet about this"
	default:
		// Derive from curiosity level
		if curiosity >= 0.5 {
			return "That catches my attention"
		} else if curiosity >= 0.3 {
			return "I notice this"
		}
		return "I feel this"
	}
}

// weaveKnowledge naturally integrates a topic into a knowledge sentence.
// Instead of "X — Y", it becomes "About X, Y" or just "Y" if the topic
// is already clear from context.
func weaveKnowledge(sentence, topic string) string {
	if topic == "" {
		return sentence
	}
	// If the sentence already mentions the topic naturally, just use it
	if containsIgnoreCase(sentence, topic) {
		return sentence
	}
	// If the topic is short enough, weave it in
	if len(topic) <= 30 {
		return fmt.Sprintf("About %s: %s", topic, lowerFirst(sentence))
	}
	return sentence
}

// sameTopic checks if two sentences are basically saying the same thing
func sameTopic(a, b string) bool {
	// Simple overlap check — if they share >60% of significant words, they're redundant
	wordsA := significantWords(a)
	wordsB := significantWords(b)
	if len(wordsA) == 0 || len(wordsB) == 0 {
		return false
	}
	overlap := 0
	for w := range wordsA {
		if wordsB[w] {
			overlap++
		}
	}
	return float64(overlap)/math.Max(float64(len(wordsA)), float64(len(wordsB))) > 0.6
}

func significantWords(s string) map[string]bool {
	words := map[string]bool{}
	stop := map[string]bool{
		"the": true, "a": true, "an": true, "is": true, "are": true,
		"was": true, "were": true, "be": true, "been": true, "being": true,
		"have": true, "has": true, "had": true, "do": true, "does": true,
		"did": true, "will": true, "would": true, "could": true, "should": true,
		"may": true, "might": true, "can": true, "shall": true, "must": true,
		"of": true, "in": true, "to": true, "for": true, "with": true,
		"on": true, "at": true, "from": true, "by": true, "about": true,
		"as": true, "into": true, "through": true, "during": true, "before": true,
		"after": true, "above": true, "below": true, "between": true, "under": true,
		"and": true, "or": true, "but": true, "not": true, "no": true,
		"nor": true, "so": true, "yet": true, "both": true, "either": true,
		"neither": true, "each": true, "every": true, "all": true, "any": true,
		"few": true, "more": true, "most": true, "other": true, "some": true,
		"such": true, "than": true, "too": true, "very": true, "just": true,
		"also": true, "that": true, "this": true, "these": true, "those": true,
		"it": true, "its": true, "i": true, "me": true, "my": true,
	}
	for _, w := range strings.Fields(strings.ToLower(s)) {
		w = strings.Trim(w, ".,;:!?()[]{}\"'")
		if len(w) > 2 && !stop[w] {
			words[w] = true
		}
	}
	return words
}

// naturalJoin joins items with natural English connectors
func naturalJoin(items []string) string {
	switch len(items) {
	case 0:
		return ""
	case 1:
		return items[0]
	case 2:
		return items[0] + " and " + items[1]
	default:
		return strings.Join(items[:len(items)-1], ", ") + ", and " + items[len(items)-1]
	}
}

func lowerFirst(s string) string {
	if len(s) == 0 {
		return s
	}
	runes := []rune(s)
	runes[0] = unicode.ToLower(runes[0])
	return string(runes)
}

func containsIgnoreCase(s, sub string) bool {
	return strings.Contains(strings.ToLower(s), strings.ToLower(sub))
}

// extractKnowledge finds the best substantive sentence from raw text.
// It filters out garbage (nav chrome, URLs, metadata) and picks
// sentences that sound like genuine knowledge — not boilerplate.
func extractKnowledge(text string) string {
	text = strings.TrimSpace(text)
	if text == "" {
		return ""
	}

	// Quick reject: known garbage patterns
	lower := strings.ToLower(text)
	garbage := []string{"skip to content", "checking your browser", "captcha", "cookie", "javascript", "<!", "error ", "403", "404", "503"}
	for _, g := range garbage {
		if strings.Contains(lower, g) {
			return ""
		}
	}

	sents := sentenceSplit(text)

	// Score each sentence by quality heuristics
	type candidate struct {
		text  string
		score float64
	}
	var candidates []candidate

	for _, s := range sents {
		s = strings.TrimSpace(s)
		words := strings.Fields(s)
		nw := len(words)

		// Length filter: too short = heading, too long = wall of text
		if nw < 7 || nw > 45 {
			continue
		}

		// Skip lines that look like metadata or nav
		lowerS := strings.ToLower(s)
		bad := []string{"click here", "read more", "sign up", "log in", "subscribe", "copyright", "all rights reserved", "powered by", "last edited", "retrieved from", "this article", "main page", "navigation menu", "personal tools", "special page"}
		isBad := false
		for _, b := range bad {
			if strings.Contains(lowerS, b) {
				isBad = true
				break
			}
		}
		if isBad {
			continue
		}

		// Skip if it's a list of items separated by pipes or tabs
		if strings.Count(s, "|") > 3 {
			continue
		}

		// Score: has a verb (action), varied vocabulary, not a title
		score := 0.0

		// Reward for having actual verbs (heuristic: common verb forms)
		verbs := []string{" is ", " are ", " was ", " were ", " has ", " have ", " had ", " do ", " does ", " can ", " could ", " will ", " would ", " may ", " might ", " should ", " makes ", " made ", " became ", " became ", " seems ", " appears ", " involves ", " means ", " represents ", " contains ", " includes ", " describes ", " refers ", " relates ", " connects "}
		for _, v := range verbs {
			if strings.Contains(lowerS, v) {
				score += 2
				break
			}
		}

		// Reward for sentence-like structure (starts with capital, ends with period)
		if len(s) > 0 && unicode.IsUpper(rune(s[0])) {
			score += 1
		}
		if strings.HasSuffix(s, ".") {
			score += 1
		}

		// Reward for moderate length (not too short, not too long)
		if nw >= 10 && nw <= 30 {
			score += 1
		}

		// Penalize if it looks like a code fragment or formula
		if strings.Contains(s, "()") || strings.Contains(s, ":=") || strings.Contains(s, "func ") {
			score -= 3
		}

		// Penalize self-referential sentences about Argos/Raven itself
		selfRef := []string{"i am", "i'm", "argos is", "this dog", "the dog is", "i feel",
			"n1_raw", "n2_form", "n3_pattern", "n4_relation", "n5_memory", "n6_intent", "n7_ghost",
			"neurons per family", "glyph memory", "harmonic propagation", "147-glyph",
			"family saturation", "signal intelligence node", "vertical lobe"}
		for _, sr := range selfRef {
			if strings.Contains(lowerS, sr) {
				score -= 5
				break
			}
		}
		// Extra penalty for technical jargon that sounds like config output
		nerdWords := 0
		for _, nw := range []string{"encoding", "mapping", "consolidation", "orientation",
			"propagation", "saturation", "raw signal", "structural"} {
			if strings.Contains(lowerS, nw) {
				nerdWords++
			}
		}
		if nerdWords >= 2 {
			score -= 3
		}

		if score > 0 {
			candidates = append(candidates, candidate{s, score})
		}
	}

	if len(candidates) == 0 {
		return ""
	}

	// Sort by score descending
	for i := 0; i < len(candidates)-1; i++ {
		for j := i + 1; j < len(candidates); j++ {
			if candidates[j].score > candidates[i].score {
				candidates[i], candidates[j] = candidates[j], candidates[i]
			}
		}
	}

	return candidates[0].text
}

func cleanVerse(s string) string {
	s = strings.TrimSpace(s)
	if len(s) < 10 {
		return ""
	}
	lower := strings.ToLower(s)
	if strings.Contains(lower, "skip to") || strings.Contains(lower, "jump to") ||
		strings.HasPrefix(lower, "error") || strings.HasPrefix(lower, "http") ||
		strings.Contains(lower, "captcha") {
		return ""
	}
	if len(s) > 200 {
		s = s[:200]
		if last := strings.LastIndex(s, " "); last > 0 {
			s = s[:last]
		}
	}
	return s
}

func clean(s string) string {
	s = strings.TrimSpace(s)
	if len(s) < 3 {
		return ""
	}
	if strings.HasPrefix(s, "http") || strings.HasPrefix(s, "Error ") {
		return ""
	}
	return s
}

// sentenceSplit splits text into sentences on . ! ?
var sentenceRe = regexp.MustCompile(`[^.!?]+[.!?]`)

func sentenceSplit(text string) []string {
	return sentenceRe.FindAllString(text, -1)
}

func shorten(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[:n]) + "..."
}

func str(v any) string {
	if v == nil {
		return ""
	}
	s, _ := v.(string)
	return s
}

func flt(v any) float64 {
	if v == nil {
		return 0
	}
	f, _ := v.(float64)
	return f
}

func emit(o modules.Output) {
	if err := json.NewEncoder(os.Stdout).Encode(o); err != nil {
		fail(err)
	}
}

func fail(err error) {
	fmt.Fprintln(os.Stderr, "pronounce:", err)
	os.Exit(1)
}