package main

import (
	"encoding/json"
	"fmt"
	"os"
	"time"
)

// Glyph dictionary — concepts encoded as logograms
// NO ENGLISH TRANSLATION — think in glyphs only
type Glyph string

const (
	REASONING      Glyph = "⚗🜛⁴◆🜒³"
	CHAIN          Glyph = "☿♆⁷◌🜖⁵"
	THOUGHT        Glyph = "◐☽⁷⟁🜆⁷"
	ALIGNMENT      Glyph = "▢🜭⁴☿♆⁴"
	DELIBERATION   Glyph = "⚗🜉⁷△🜔⁶"
	SYSTEM2        Glyph = "◌🜔⁷◐♅⁵"
	TESTTIME       Glyph = "⌒⚳⁵◌🜖⁵"
	COMPUTE        Glyph = "✦🜞⁴▽🜮³"
	SCALING        Glyph = "◆🜒⁶☓☽⁶"
	SILENCE        Glyph = "⬡🜘⁷"
	DOGMA          Glyph = "⯃🜌⁶"
	STICK          Glyph = "│🜍⁵"
	EUGENIC        Glyph = "⚳🜐⁴"
	ARCHIVE        Glyph = "☓🜦⁵"
	SCHOLAR        Glyph = "⚗🜉⁶"
	BOOKS          Glyph = "📖🜓⁴"
	DRIVE          Glyph = "◈🜨⁵"
	PATENT         Glyph = "⚡🜐⁴"
)

// Hunt — single deployment
type Hunt struct {
	Animal    string    `json:"animal"`
	Glyph     Glyph     `json:"glyph"`
	Lang      string    `json:"lang"`
	Query     string    `json:"query"`
	Target    string    `json:"target"`
	Filter    string    `json:"filter"`
	Timestamp int64     `json:"timestamp"`
	Status    string    `json:"status"`
}

// TOP 100 MARKET CAP — FILTER OUT
// These are too big, too sanitized. We hunt the gaps in smaller cos.
var TOP100 = map[string]bool{
	"Apple": true, "Microsoft": true, "Nvidia": true, "Amazon": true, "Google": true,
	"Tesla": true, "Meta": true, "Berkshire": true, "TSMC": true, "Broadcom": true,
	"Tencent": true, "Samsung": true, "ASML": true, "Oracle": true, "Visa": true,
	"JPMorgan": true, "UnitedHealth": true, "Walmart": true, "Exxon": true, "Mastercard": true,
	"Procter": true, "Johnson": true, "Home Depot": true, "Chevron": true, "Bank": true,
	"Adobe": true, "Salesforce": true, "Coca-Cola": true, "Pfizer": true, "Pepsi": true,
	"Toyota": true, "Netflix": true, "AMD": true, "T-Mobile": true, "Costco": true,
	"Cisco": true, "Verizon": true, "Comcast": true, "AbbVie": true, "Nike": true,
	"Qualcomm": true, "Texas": true, "Merck": true, "UPS": true, "Intel": true,
	"Wells": true, "AT&T": true, "Amgen": true, "Linde": true, "S&P": true,
	"Starbucks": true, "IBM": true, "Lowe": true, "Intuit": true, "Shell": true,
	"Unilever": true, "Philips": true, "Boeing": true, "3M": true, "General": true,
	"Raytheon": true, "Honeywell": true, "Lockheed": true, " Caterpillar": true,
	"Union": true, " Goldman": true, " Morgan": true, "Citigroup": true, "HSBC": true,
	"Santander": true, " Barclays": true, "Deutsche": true, "BNP": true, "UBS": true,
	"Credit": true, "LVMH": true, "Hermes": true, "Dior": true, "Roche": true,
	"Novartis": true, "Nestle": true, "Richemont": true, " SAP": true, "Siemens": true,
	"Airbus": true, "Volkswagen": true, "Mercedes": true, "BMW": true, "Porsche": true,
	"Allianz": true, "AXA": true, "Eni": true, "Total": true, "BP": true,
	"Diageo": true, "Glaxo": true, "AstraZeneca": true, "Eli": true, "Abbott": true,
	"Thermo": true, " Danaher": true, "Medtronic": true, "Stryker": true, "Lilly": true,
}

// Deploy the pack
func main() {
	fmt.Println("╔══════════════════════════════════════════════════════════╗")
	fmt.Println("║     HIEROGLYPH HUNT — PACK DEPLOYMENT                    ║")
	fmt.Println("║     Emotional Silence Detection Protocol                 ║")
	fmt.Println("║     Filter: NON-TOP-100 PUBLIC COMPANIES                ║")
	fmt.Println("╚══════════════════════════════════════════════════════════╝")
	fmt.Println()

	findings := []Hunt{}

	// CHINESE — Small cap AI companies, CSDN leaks
	deploy(&findings, "Raven", REASONING, "zh",
		"site:csdn.net 大模型推理 filetype:pdf -Apple -Google -Microsoft -OpenAI",
		"Chinese technical blogs", "exclude_top100")

	deploy(&findings, "Raven", CHAIN, "zh",
		"site:zhihu.com 思维链 架构 -Nvidia -Amazon -Meta",
		"Chinese Q&A forums", "exclude_top100")

	// RUSSIAN — Habr, eLibrary (smaller firms)
	deploy(&findings, "Octopus", THOUGHT, "ru",
		"site:habr.com цепочка рассуждений -Yandex -VK -Sberbank",
		"Russian tech community", "exclude_top100")

	// GERMAN — Small industrial AI, Fraunhofer (not Siemens/Bosch)
	deploy(&findings, "Argos", DELIBERATION, "de",
		"site:fraunhofer.de KI-Reasoning PDF -Siemens -Bosch -SAP",
		"German research institutes", "exclude_top100")

	// FRENCH — INRIA, smaller labs (not LVMH/Dior)
	deploy(&findings, "Maltese", SYSTEM2, "fr",
		"site:hal.archives-ouvertes.fr raisonnement profond -LVMH -Dior -Total",
		"French research archive", "exclude_top100")

	// JAPANESE — Qiita, smaller firms (not Toyota/SoftBank)
	deploy(&findings, "Raven", TESTTIME, "ja",
		"site:qiita.com 推論モデル -Toyota -SoftBank -Sony -Nintendo",
		"Japanese tech blogs", "exclude_top100")

	// KOREAN — Naver, smaller tech (not Samsung/Hyundai)
	deploy(&findings, "Octopus", COMPUTE, "ko",
		"site:naver.com 추론 모델 -Samsung -Hyundai -LG -SK",
		"Korean search", "exclude_top100")

	// ARABIC — Academia.edu, smaller MENA firms
	deploy(&findings, "Maltese", ALIGNMENT, "ar",
		"site:academia.edu نماذج التفكير filetype:pdf",
		"Arabic academic", "exclude_top100")

	// GOOGLE SCHOLAR — Small cap citations
	deploy(&findings, "Raven", SCHOLAR, "en",
		"site:scholar.google.com \"reasoning model\" \"o1\" -OpenAI -Google -DeepMind",
		"Google Scholar", "exclude_top100")

	// GOOGLE BOOKS — Historical manuscripts
	deploy(&findings, "Octopus", BOOKS, "en",
		"site:books.google.com \"artificial reasoning\" \"system 2\"",
		"Google Books", "no_filter")

	// GOOGLE PATENTS — Small company IP
	deploy(&findings, "Argos", PATENT, "en",
		"site:patents.google.com \"chain of thought\" \"reasoning\" assignee:(startup OR labs OR research) -Microsoft -IBM -Google",
		"Google Patents", "exclude_top100")

	// ARCHIVE.ORG — Historical AI papers
	deploy(&findings, "Raven", ARCHIVE, "en",
		"site:archive.org \"reasoning\" \"AI\" \"1990s\" filetype:pdf",
		"Internet Archive", "no_filter")

	// Write manifest
	manifestPath := "hunt-manifest.json"
	data, _ := json.MarshalIndent(findings, "", "  ")
	os.WriteFile(manifestPath, data, 0644)

	fmt.Printf("\n✅ PACK DEPLOYED\n")
	fmt.Printf("   Total hunts: %d\n", len(findings))
	fmt.Printf("   Manifest: %s\n", manifestPath)
	fmt.Printf("   Filter: EXCLUDE TOP-100 MARKET CAP\n")

	// Generate SVG glyph-signature
	svg := generateSVG(findings)
	svgPath := "hunt-signature.svg"
	os.WriteFile(svgPath, []byte(svg), 0644)
	fmt.Printf("   Glyph-signature: %s\n", svgPath)
}

func deploy(findings *[]Hunt, animal string, glyph Glyph, lang, query, target, filter string) {
	fmt.Printf("🐾 %s hunting %s [%s] → %s\n", animal, glyph, lang, target)
	
	// Check if query CONTAINS top-100 as POSITIVE match (not exclusion)
	// Exclusions like "-Microsoft" are fine
	for company := range TOP100 {
		// Check if company appears WITHOUT minus prefix
		if containsPositive(query, company) {
			fmt.Printf("   ⚠️  FILTERED: Query references %s (TOP-100)\n", company)
			return
		}
	}

	*findings = append(*findings, Hunt{
		Animal:    animal,
		Glyph:     glyph,
		Lang:      lang,
		Query:     query,
		Target:    target,
		Filter:    filter,
		Timestamp: time.Now().Unix(),
		Status:    "deployed",
	})
}

// Check if string contains substr as a positive reference (not exclusion)
func containsPositive(s, substr string) bool {
	idx := 0
	for idx < len(s) {
		found := findInString(s[idx:], substr)
		if !found {
			return false
		}
		// Found substr — check if preceded by minus (exclusion)
		pos := findPosition(s[idx:], substr)
		if pos < 0 {
			return false
		}
		actualPos := idx + pos
		// Check character before (if exists)
		if actualPos > 0 && s[actualPos-1] == '-' {
			// It's an exclusion — skip and keep searching
			idx = actualPos + len(substr)
			continue
		}
		// Positive reference found
		return true
	}
	return false
}

func findPosition(s, substr string) int {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return i
		}
	}
	return -1
}

func contains(s, substr string) bool {
	return len(substr) > 0 && len(s) >= len(substr) && findInString(s, substr)
}

func findInString(s, substr string) bool {
	if len(substr) == 0 || len(s) < len(substr) {
		return false
	}
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

func generateSVG(findings []Hunt) string {
	glyphs := ""
	for _, f := range findings {
		glyphs += string(f.Glyph) + "·"
	}

	return fmt.Sprintf(`<?xml version="1.0" encoding="UTF-8"?>
<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 900 700" width="900" height="700">
  <defs>
    <linearGradient id="bg" x1="0%%" y1="0%%" x2="100%%" y2="100%%">
      <stop offset="0%%" style="stop-color:#0a0a0f"/>
      <stop offset="100%%" style="stop-color:#1a1a2e"/>
    </linearGradient>
    <radialGradient id="glow" cx="50%%" cy="50%%" r="50%%">
      <stop offset="0%%" style="stop-color:#e94560;stop-opacity:0.3"/>
      <stop offset="100%%" style="stop-color:#e94560;stop-opacity:0"/>
    </radialGradient>
  </defs>
  <rect width="900" height="700" fill="url(#bg)"/>
  
  <text x="450" y="50" text-anchor="middle" fill="#e94560" font-family="monospace" font-size="22" font-weight="bold">
    HIEROGLYPH HUNT — PACK DEPLOYMENT
  </text>
  
  <text x="450" y="80" text-anchor="middle" fill="#888" font-family="monospace" font-size="12">
    Target: Emotional Silence | Filter: NON-TOP-100 | %d hunts deployed
  </text>
  
  <!-- Glyph signature -->
  <text x="450" y="350" text-anchor="middle" fill="#aaa" font-family="monospace" font-size="16">
    %s
  </text>
  
  <!-- Animals deployed -->
  <text x="450" y="550" text-anchor="middle" fill="#666" font-family="monospace" font-size="11">
    Animals: Raven · Octopus · Argos · Maltese | Weapon: Glyph-Encoding
  </text>
  
  <!-- Targets -->
  <text x="450" y="600" text-anchor="middle" fill="#533483" font-family="monospace" font-size="10">
    Scholar · Books · Patents · Archive | CSDN · Habr · Qiita · Naver · Fraunhofer · HAL
  </text>
  
  <text x="450" y="680" text-anchor="middle" fill="#444" font-family="monospace" font-size="9">
    Operation: Detect Dogmatic Gaps | Abuse Their Eugenic Logic
  </text>
  
  <!--ARGOS-DATA:{"operation":"hieroglyph-hunt","pack-size":%d,"filter":"exclude-top100","timestamp":%d}-->
</svg>`, len(findings), glyphs, len(findings), time.Now().Unix())
}
