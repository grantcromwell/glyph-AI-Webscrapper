package output

import (
	"encoding/base64"
	"fmt"
	"regexp"
	"strings"
)

// A Scroll is how glyphai persists state as art instead of JSON. Every saved
// artifact — the lexicon, the recall index, the cerebellum's weights — is an
// SVG that (a) shows its content as a small artwork and (b) carries the exact
// data losslessly in an ARGOS-DATA marker, so it reloads perfectly. Nothing the
// dog keeps is "non-art": open any file and it is a picture.

// Inscribe wraps a plain-text payload of a kind into an SVG scroll with the
// given title and inner art body. The payload is NOT JSON — callers use the
// record helpers (or raw text) below.
func Inscribe(kind, title, payload, body string) string {
	var b strings.Builder
	b.WriteString(`<?xml version="1.0" encoding="UTF-8"?>` + "\n")
	fmt.Fprintf(&b, "<!--ARGOS-DATA:%s\n%s\n:END-->\n", kind, payload)
	b.WriteString(`<svg xmlns="http://www.w3.org/2000/svg" width="660" height="440" viewBox="0 0 660 440" font-family="monospace">` + "\n")
	b.WriteString(`<rect width="660" height="440" fill="#06070d"/>` + "\n")
	fmt.Fprintf(&b, `<text x="28" y="34" fill="#cfd6e6" font-size="18" font-weight="bold">🐾 %s</text>`+"\n", xmlEsc(title))
	fmt.Fprintf(&b, `<text x="28" y="54" fill="#6b7488" font-size="11">glyphai scroll · kind=%s</text>`+"\n", xmlEsc(kind))
	b.WriteString(body)
	b.WriteString("</svg>\n")
	return b.String()
}

// XMLEsc escapes text for safe inclusion in SVG (exported for other organs).
func XMLEsc(s string) string { return xmlEsc(s) }

var dataRe = regexp.MustCompile(`(?s)<!--ARGOS-DATA:(\S+)\n(.*?)\n:END-->`)

// ReadScroll extracts the payload of the given kind from an SVG scroll.
func ReadScroll(svg []byte, kind string) (string, error) {
	m := dataRe.FindSubmatch(svg)
	if m == nil {
		return "", fmt.Errorf("paw: no ARGOS-DATA marker")
	}
	if string(m[1]) != kind {
		return "", fmt.Errorf("paw: scroll kind %q, want %q", m[1], kind)
	}
	return string(m[2]), nil
}

// EncodeRecords serialises rows of string fields into a payload (base64 per
// field so any bytes survive; tab between fields, newline between rows). This
// is the dog's own line script — deliberately not JSON.
func EncodeRecords(rows [][]string) string {
	var b strings.Builder
	for _, row := range rows {
		enc := make([]string, len(row))
		for i, f := range row {
			enc[i] = base64.StdEncoding.EncodeToString([]byte(f))
		}
		b.WriteString(strings.Join(enc, "\t"))
		b.WriteByte('\n')
	}
	return strings.TrimRight(b.String(), "\n")
}

// DecodeRecords reverses EncodeRecords.
func DecodeRecords(payload string) [][]string {
	var rows [][]string
	for _, line := range strings.Split(payload, "\n") {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}
		fields := strings.Split(line, "\t")
		row := make([]string, len(fields))
		for i, f := range fields {
			if dec, err := base64.StdEncoding.DecodeString(f); err == nil {
				row[i] = string(dec)
			}
		}
		rows = append(rows, row)
	}
	return rows
}