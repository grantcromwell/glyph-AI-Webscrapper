// mythoshunt — a Markov hunt that sets out from the original Fable 5 and Mythos
// articles. It finds those two pages through the dog's own DuckDuckGo spell, then
// looses the self-guided Markov walk (cmd/roam) from them, with the dog's LEARNED
// junk-avoidance steering each step toward substance and away from chrome.
//
//	go run ./scripts/mythoshunt            # 15 min, fast eyes
//	go run ./scripts/mythoshunt -min 30 -eyes browse
package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"glyphai/internal/modules"
)

type link struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

func main() {
	mins := flag.Int("min", 15, "hunt length in minutes")
	eyes := flag.String("eyes", "skim", "skim (fast) or browse (headless Chromium)")
	flag.Parse()

	spells, err := modules.Discover("modules")
	if err != nil {
		fail(err)
	}
	var forage *modules.Spell
	for i := range spells {
		if spells[i].Name == "forage" {
			forage = &spells[i]
		}
	}
	if forage == nil {
		fail(fmt.Errorf("no forage spell"))
	}

	var seeds []string
	for _, q := range []string{"Fable 5 Anthropic model", "Mythos 5 model"} {
		out, e := forage.Cast(context.Background(), modules.Input{Text: q, Args: map[string]any{"limit": 4}})
		if e != nil {
			continue
		}
		for _, l := range decodeLinks(out.Data["links"]) {
			if strings.HasPrefix(l.URL, "http") {
				seeds = append(seeds, l.URL)
				fmt.Printf("🐾 seed for %q → %s\n", q, l.URL)
				break
			}
		}
	}
	if len(seeds) == 0 {
		fail(fmt.Errorf("could not find the Fable 5 / Mythos articles"))
	}

	until := time.Now().Add(time.Duration(*mins) * time.Minute).Unix()
	roam := filepath.Join("bin", "roam")
	cmd := exec.Command(roam, "-eyes", *eyes, "-seed", strings.Join(seeds, ","),
		"-until", strconv.FormatInt(until, 10))
	cmd.Env = append(os.Environ(), "PATH="+goBin()+":"+os.Getenv("PATH"))
	cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
	fmt.Printf("🐾 Markov hunt from %d seeds for %d min (eyes=%s, learning junk-avoidance)…\n", len(seeds), *mins, *eyes)
	_ = cmd.Run()
}

func decodeLinks(v any) []link {
	raw, _ := json.Marshal(v)
	var l []link
	_ = json.Unmarshal(raw, &l)
	return l
}

func goBin() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "go", "bin")
	}
	return "/usr/local/go/bin"
}

func fail(err error) { fmt.Fprintln(os.Stderr, "mythoshunt:", err); os.Exit(1) }