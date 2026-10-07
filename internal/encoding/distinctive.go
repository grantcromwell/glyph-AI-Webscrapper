package encoding

import (
	"encoding/base64"
	"os"
	"regexp"
	"sync"
)

// Distinctive denotation. Every word the dog hears shares a common hum (the Voice
// families fire for all language), so bare logograms blur together. The dog keeps
// a BASELINE — the mean of everything it perceives — and denotes a concept by
// what rises ABOVE that baseline: the part of its glyph that is special to it.
// Subtracting the shared hum leaves a crisp, concept-specific logogram, so two
// concepts (and the point between them) are clearly told apart.

var (
	baselineOnce sync.Once
	baseline     Detail
	baselineRe   = regexp.MustCompile(`(?s)<!--ARGOS-DATA:baseline\n(.*?)\n:END-->`)
)

// loadBaseline reads the dog's mean-perception glyph from myelin/baseline.svg
// once. Missing → a zero baseline (Distinctive then equals the raw glyph).
func loadBaseline() {
	baselineOnce.Do(func() {
		path := os.Getenv("ARGOS_BASELINE")
		if path == "" {
			path = "weights/baseline.svg"
		}
		data, err := os.ReadFile(path)
		if err != nil {
			return
		}
		m := baselineRe.FindSubmatch(data)
		if m == nil {
			return
		}
		raw, err := base64.StdEncoding.DecodeString(string(m[1]))
		if err != nil || len(raw) < Families*Subfamilies {
			return
		}
		for f := 0; f < Families; f++ {
			for s := 0; s < Subfamilies; s++ {
				baseline[f][s] = raw[f*Subfamilies+s]
			}
		}
	})
}

// Distinctive returns a glyph's signature above the dog's baseline — what makes
// the concept itself, with the shared hum subtracted away.
func Distinctive(d Detail) Detail {
	loadBaseline()
	var out Detail
	for f := 0; f < Families; f++ {
		for s := 0; s < Subfamilies; s++ {
			if v := int(d[f][s]) - int(baseline[f][s]); v > 0 {
				out[f][s] = uint8(v)
			}
		}
	}
	return out
}

// ConceptLogogram denotes a concept by its distinctive signature — the dog's
// crisp logogram for the idea, free of the Voice every word carries.
func ConceptLogogram(d Detail) string { return Logogram(Distinctive(d)) }

// BaselineBytes flattens a baseline glyph to the 1512-byte payload the loader
// reads (used by the trainer to persist it).
func BaselineBytes(d Detail) []byte {
	b := make([]byte, Families*Subfamilies)
	for f := 0; f < Families; f++ {
		for s := 0; s < Subfamilies; s++ {
			b[f*Subfamilies+s] = d[f][s]
		}
	}
	return b
}