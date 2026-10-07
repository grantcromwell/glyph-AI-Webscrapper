// Package encoding — text-to-glyph encoding via breath-shape physics.
//
// DESIGN PATTERN: This is the perceptual front door. Every text input passes
// through Encode(), which transposes language into the 21x72 glyph field using
// mouth-physics (breath-shapes), not surface script. This is the key insight:
// the system hears the breath, not the alphabet. A plosive in Arabic lights
// the same families as a plosive in English. This makes the encoding
// language-agnostic by construction.
//
// The encoding has six layers, each capturing a different aspect of language:
//   1. Phoneme breath-shape -> family routing (the primary signal)
//   2. Phoneme cluster -> subfamily (fine detail within a breath)
//   3. Vowel melody -> emotional families (prosody, the song of speech)
//   4. Word-to-word transitions -> dynamic flow (language in motion)
//   5. Absence encoding -> ghost field (what breaths are NOT present)
//   6. Rhythm encoding -> pattern of presence/absence (the missed beats)
//
// Layers 5 and 6 (absence and rhythm) are the maximum-entropy extensions that
// make this encoding fundamentally different from bag-of-words or TF-IDF.
// They capture what is NOT there, which is often more informative than what is.
//
// Pattern analysis (P25 Constraint Compression, P28 Hormesis):
// The 3-bit level (0-7) per cell is a constraint compression: 1512 cells at
// 3 bits each = 567 bytes per perception. This is the Kolmogorov complexity
// of the input signal. Two inputs that compress to the same glyph are
// structurally similar regardless of surface form.
package encoding

import (
	"math"
	"strings"
	"unicode"

	"glyphai/internal/audio"
)

// Encode transposes text into a fine Detail glyph through BREATH-SHAPE physics,
// not English surface form. The dog hears the breath, not the script:
//
//   - each phoneme's breath-shape (plosive/fricative/nasal/approximant/vowel/
//     sibilant) routes its signal to the families that answer that breath — so
//     a family lights up because its FEATURE (its breath) was spoken;
//   - the phoneme cluster picks the subfamily — the fine subcategory a real
//     sound maps onto;
//   - vowel melody sings into the emotional families (the song English drops);
//   - breath-shifts between words carry the dynamic flow (language in motion).
//   - TRANSITIONS between phonemes encode the "between" — the void between
//     breaths that triconsonantal languages use for meaning;
//   - ABSENCE encodes what breaths are NOT present — the ghost field;
//   - RHYTHM encodes the pattern of presence/absence — the missed beats.
//
// Deterministic; words that feel alike in the mouth resonate.
func Encode(text string) Detail {
	text = strings.ToLower(strings.TrimSpace(text))
	var counts [Families][Subfamilies]int
	max := 0
	bump := func(f, s, w int) {
		counts[f][s] += w
		if counts[f][s] > max {
			max = counts[f][s]
		}
	}
	emo := [...]int{7, 9, 19, 14, 16, 2} // Voice Rhythm Heat Water Air Void
	addEmo := func(h uint64, w int) {
		f := emo[int(h%uint64(len(emo)))]
		bump(f, int((h/uint64(len(emo)))%Subfamilies), w)
	}

	var shapes []string
	var allPhonemes []int // for transition encoding
	for _, w := range strings.Fields(text) {
		rs := letters(w)
		if len(rs) == 0 {
			continue
		}
		for i := 0; i < len(rs); i++ {
			b := audio.Manner(rs[i])
			fams := familiesByBreath[b]
			if len(fams) == 0 {
				continue
			}
			sig := string(rs[i])
			if i+1 < len(rs) {
				sig += string(rs[i+1]) // the phoneme in its local cluster
			}
			h := hash64("p:" + sig)
			f := fams[int(hash64("f:"+sig)%uint64(len(fams)))] // which family of this breath
			bump(f, int(h%Subfamilies), 2)
			allPhonemes = append(allPhonemes, int(b))
			if i+2 < len(rs) { // trigram -> more subglyphs
				h3 := hash64("p3:" + string(rs[i:i+3]))
				bump(f, int(h3%Subfamilies), 1)
			}
		}
		// vowel melody -> emotional subglyphs (the word's song)
		vs := []rune(vowelMelody(w))
		addEmo(hash64("vw:"+string(vs)), 2)
		for i := 0; i < len(vs); i++ {
			addEmo(hash64("v:"+string(vs[i])), 1)
			if i+1 < len(vs) {
				addEmo(hash64("vv:"+string(vs[i:i+2])), 2)
			}
		}
		shapes = append(shapes, breathString(rs))
	}
	// Dynamic flow: shifts in mouth-shape from one word to the next.
	for i := 0; i+1 < len(shapes); i++ {
		addEmo(hash64("t:"+shapes[i]+">"+shapes[i+1]), 3)
	}

	// === MAXIMUM-ENTROPY EXTENSIONS ===

	// 1. TRANSITION ENCODING: encode the between-breath transitions
	// This captures what triconsonantal languages do — meaning in the gaps
	if len(allPhonemes) > 1 {
		for i := 0; i < len(allPhonemes)-1; i++ {
			from := allPhonemes[i]
			to := allPhonemes[i+1]
			// The transition itself maps to a family
			tf := (from + to) % Families
			ts := (from*7 + to*3) % Subfamilies
			bump(tf, ts, 1)
		}
	}

	// 2. ABSENCE ENCODING: what breaths are NOT present
	// The ghost field — absence as primary signal
	present := make(map[int]bool)
	for _, b := range allPhonemes {
		present[b] = true
	}
	for b := 0; b < 256; b++ {
		if !present[b] {
			af := b % Families
			as := (b * 7) % Subfamilies
			bump(af, as, 1)
		}
	}

	// 3. RHYTHM ENCODING: pattern of presence/absence
	// The missed beats — rhythm holes
	if len(allPhonemes) > 0 {
		// Encode runs of same breath type
		runStart := 0
		for i := 1; i <= len(allPhonemes); i++ {
			if i == len(allPhonemes) || allPhonemes[i] != allPhonemes[runStart] {
				runLen := i - runStart
				rf := runLen % Families
				rs := (runLen * 11) % Subfamilies
				bump(rf, rs, runLen)
				runStart = i
			}
		}
		// Encode gaps between different breath types
		for i := 0; i < len(allPhonemes)-1; i++ {
			if allPhonemes[i] != allPhonemes[i+1] {
				gap := (allPhonemes[i+1] - allPhonemes[i] + 256) % 256
				gf := gap % Families
				gs := (gap * 5) % Subfamilies
				bump(gf, gs, 1)
			}
		}
	}

	// === END MAXIMUM-ENTROPY EXTENSIONS ===

	var d Detail
	if max == 0 {
		return d
	}
	lmax := math.Log1p(float64(max))
	for f := 0; f < Families; f++ {
		for s := 0; s < Subfamilies; s++ {
			if counts[f][s] == 0 {
				continue
			}
			v := math.Log1p(float64(counts[f][s])) / lmax * MaxLevel
			d[f][s] = clampLevel(int(math.Round(v)))
		}
	}
	return d
}

// EncodeGlyph is the coarse convenience: Encode then Coarse.
func EncodeGlyph(text string) Glyph { return Encode(text).Coarse() }

// letters returns the lowercase letters (and combining marks) of a word, in ANY
// script — so the dog hears Cyrillic, Han, Arabic, Devanagari, etc., not just
// Latin. The cochlea decides the breath of each rune.
func letters(w string) []rune {
	var out []rune
	for _, r := range w {
		if unicode.IsLetter(r) || unicode.IsMark(r) {
			out = append(out, unicode.ToLower(r))
		}
	}
	return out
}

// breathString renders a word's breath-shape sequence as a signature.
func breathString(rs []rune) string {
	var b strings.Builder
	for _, r := range rs {
		b.WriteByte(byte('0' + audio.Manner(r)))
	}
	return b.String()
}

// vowelMelody keeps only the vowels — the emotional/prosodic song of a word —
// where "vowel" is any rune the ear hears as the open breath, in any script.
func vowelMelody(w string) string {
	var b strings.Builder
	for _, r := range w {
		if (unicode.IsLetter(r) || unicode.IsMark(r)) && audio.Manner(r) == audio.Vowel {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// hash64 is FNV-1a, vendored so the package has no dependencies.
func hash64(s string) uint64 {
	const (
		offset = 1469598103934665603
		prime  = 1099511628211
	)
	h := uint64(offset)
	for i := 0; i < len(s); i++ {
		h ^= uint64(s[i])
		h *= prime
	}
	return h
}