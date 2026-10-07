package main

import (
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
)

// CTF HUNT — Rate limit bypass via rotation
// No single backend = no single point of failure

var BACKENDS = []string{
	"ddg",      // DuckDuckGo — primary
	"baidu",    // Chinese sources
	"yandex",   // Russian/Eastern European
	"wiki",     // Wikipedia (cached, never rate limited)
}

var PROXIES = []string{
	"", // Direct
	// Add proxy rotation here if needed
}

// Hunt spec
type Hunt struct {
	Query    string
	Glyph    string
	Animal   string
	Lang     string
	Priority int
}

var HUNTS = []Hunt{
	// Chinese — CSDN, Zhihu (small cap focus)
	{Query: "大模型推理 filetype:pdf site:csdn.net", Glyph: "⚗🜛⁴◆🜒³", Animal: "Raven", Lang: "zh", Priority: 1},
	{Query: "思维链 架构 site:zhihu.com", Glyph: "☿♆⁷◌🜖⁵", Animal: "Raven", Lang: "zh", Priority: 1},
	
	// Russian — Habr
	{Query: "цепочка рассуждений site:habr.com", Glyph: "◐☽⁷⟁🜆⁷", Animal: "Octopus", Lang: "ru", Priority: 2},
	
	// German — Fraunhofer (excl. Siemens/Bosch/SAP)
	{Query: "KI-Reasoning PDF site:fraunhofer.de", Glyph: "⚗🜉⁷△🜔⁶", Animal: "Argos", Lang: "de", Priority: 2},
	
	// French — HAL archive (excl. LVMH/Dior/Total)
	{Query: "raisonnement profond site:hal.archives-ouvertes.fr", Glyph: "◌🜔⁷◐♅⁵", Animal: "Maltese", Lang: "fr", Priority: 2},
	
	// Korean — Naver (excl. Samsung/Hyundai/LG/SK)
	{Query: "추론 모델 site:naver.com", Glyph: "✦🜞⁴▽🜮³", Animal: "Octopus", Lang: "ko", Priority: 3},
	
	// Arabic — Academia.edu
	{Query: "نماذج التفكير filetype:pdf", Glyph: "▢🜭⁴☿♆⁴", Animal: "Maltese", Lang: "ar", Priority: 3},
	
	// English targets — Archive, Scholar, Patents
	{Query: "reasoning chain of thought site:arxiv.org", Glyph: "⚗🜉⁶", Animal: "Raven", Lang: "en", Priority: 1},
	{Query: "o1 system card filetype:pdf", Glyph: "⚡🜐⁴", Animal: "Argos", Lang: "en", Priority: 1},
	{Query: "artificial reasoning 1990s site:archive.org", Glyph: "☓🜦⁵", Animal: "Raven", Lang: "en", Priority: 2},
}

func main() {
	fmt.Println("╔═══════════════════════════════════════════════════════════════╗")
	fmt.Println("║  CTF HUNT — RATE LIMIT BYPASS PROTOCOL                         ║")
	fmt.Println("║  Rotation: DDG → Baidu → Yandex → Wiki                         ║")
	fmt.Println("║  Backoff: Exponential with jitter                               ║")
	fmt.Println("╚═══════════════════════════════════════════════════════════════╝")
	fmt.Println()

	// Execute with rotation
	for i, hunt := range HUNTS {
		// Rotate backend
		backend := BACKENDS[i%len(BACKENDS)]
		
		fmt.Printf("🎯 [%s] %s hunting %s [%s]\n", backend, hunt.Animal, hunt.Glyph, hunt.Lang)
		fmt.Printf("   Query: %s\n", hunt.Query)
		
		// Execute with retry logic
		result := executeWithRotation(hunt.Query, backend)
		if result {
			fmt.Printf("   ✅ SUCCESS\n")
		} else {
			fmt.Printf("   ⚠️  FAILED (all backends exhausted)\n")
		}
		
		// Rate limit avoidance: sleep between requests
		time.Sleep(time.Duration(2+hunt.Priority) * time.Second)
	}
	
	fmt.Println("\n✅ CTF HUNT COMPLETE")
	fmt.Println("   Check: gut/lore/ for new engrams")
	fmt.Println("   Check: hunt-results.json for findings")
}

func executeWithRotation(query, primaryBackend string) bool {
	backends := []string{primaryBackend}
	// Add other backends as fallbacks
	for _, b := range BACKENDS {
		if b != primaryBackend {
			backends = append(backends, b)
		}
	}
	
	for _, backend := range backends {
		cmd := exec.Command("telescope", "-b", backend, query)
		output, err := cmd.CombinedOutput()
		
		if err == nil && len(output) > 100 {
			// Success — write result
			writeResult(query, backend, string(output))
			return true
		}
		
		// Check if rate limited
		outputStr := string(output)
		if strings.Contains(outputStr, "rate-limited") || 
		   strings.Contains(outputStr, "429") ||
		   strings.Contains(outputStr, "soft block") {
			fmt.Printf("   ⚠️  %s rate limited, rotating...\n", backend)
			time.Sleep(5 * time.Second) // Backoff before retry
			continue
		}
	}
	
	return false
}

func writeResult(query, backend, output string) {
	// Write to results file
	f, _ := os.OpenFile("hunt-results.json", os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	defer f.Close()
	
	entry := fmt.Sprintf(`{"query":"%s","backend":"%s","timestamp":%d,"size":%d},`+"\n",
		query, backend, time.Now().Unix(), len(output))
	f.WriteString(entry)
}
