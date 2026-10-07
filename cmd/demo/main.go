// play — the dog playing the game. This is the "kid with a Minecraft modpack":
// an always-on loop in which the dog steps through the portal into the headless
// world, reads the place it lands in, lets its parietal lobe place it on the map,
// remembers whatever stirs it, and lets its will (telekinesis) pull it down the
// next path. Worldgen lives entirely in the separate game process; this loop only
// perceives, maps, remembers, and wills.
//
//	go run ./cmd/play                       # play forever (Ctrl-C / service stop)
//	go run ./cmd/play -until 1781455500     # play until a unix deadline
//	go run ./cmd/play -seed http://127.0.0.1:8088/index.html -bank world
package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"os"
	"time"

	"glyphai/internal/encoding"
	"glyphai/internal/modules"
	"glyphai/internal/memory"
	"glyphai/internal/fetch"
	"glyphai/internal/spatial"
	"glyphai/internal/action"
)

const (
	grimoireDir = "modules"
	memRoot     = "memory"
	loreDir     = "data/lore"
	activeGate  = 30 // a place is worth remembering only if it stirs the senses (as roam)
	saveEvery   = 25
)

type link struct {
	Text string `json:"text"`
	URL  string `json:"url"`
}

func main() {
	until := flag.Int64("until", 0, "unix deadline (0 = play until stopped)")
	start := flag.String("seed", "http://127.0.0.1:8088/index.html", "the place to step in at")
	bank := flag.String("bank", "world", "memory bank to file places into")
	mapPath := flag.String("map", "hippocampus/cartograph.svg", "where the parietal cartograph lives")
	flag.Parse()

	deadline := time.Now().Add(100 * 365 * 24 * time.Hour)
	if *until > 0 {
		deadline = time.Unix(*until, 0)
	}

	spells, err := modules.Discover(grimoireDir)
	must(err)
	var portal *modules.Spell
	for i := range spells {
		if spells[i].Name == "portal" {
			portal = &spells[i]
		}
	}
	if portal == nil {
		fmt.Fprintln(os.Stderr, "play: the portal spell is missing (grimoire/play/portal)")
		os.Exit(1)
	}

	h, err := memory.Open(memRoot, *bank)
	must(err)
	par := spatial.New()
	_ = par.Load(*mapPath)
	will := action.New(1)

	visited := map[string]bool{}
	current := *start
	steps, filed := 0, 0

	for time.Now().Before(deadline) {
		if current == "" {
			current = *start
		}
		out, err := will.Reach(portal, current)
		if err != nil {
			// the game may not be up yet — wait and try the door again
			time.Sleep(2 * time.Second)
			current = *start
			continue
		}
		docs := decodeDocs(out.Data["docs"])
		links := decodeLinks(out.Data["links"])
		if len(docs) == 0 {
			current = *start
			continue
		}
		doc := docs[0]
		visited[doc.URL] = true

		par.Integrate(doc.URL, doc.Title, doc.Text)
		if encoding.Encode(doc.Title+" "+doc.Text).Active() >= activeGate {
			if _, e := h.RememberTo(loreDir, doc.Title, doc.Text, "play:"+doc.URL); e == nil {
				filed++
			}
		}

		// prefer the unmapped (novelty); fall back to all paths if exhausted
		fresh := make([]spatial.Link, 0, len(links))
		for _, l := range links {
			if !visited[l.URL] {
				fresh = append(fresh, spatial.Link{Text: l.Text, URL: l.URL})
			}
		}
		if len(fresh) == 0 {
			for _, l := range links {
				fresh = append(fresh, spatial.Link{Text: l.Text, URL: l.URL})
			}
		}
		current = will.Sample(par.Options(fresh))

		steps++
		if steps%saveEvery == 0 {
			_ = par.Save(*mapPath)
			fmt.Printf("play: %d steps, %d remembered, %d places mapped\n", steps, filed, par.Mapped())
		}
	}

	_ = par.Save(*mapPath)
	fmt.Printf("play: done — %d steps, %d remembered, %d places mapped\n", steps, filed, par.Mapped())
}

func decodeDocs(v any) []fetch.Doc {
	raw, _ := json.Marshal(v)
	var d []fetch.Doc
	_ = json.Unmarshal(raw, &d)
	return d
}

func decodeLinks(v any) []link {
	raw, _ := json.Marshal(v)
	var l []link
	_ = json.Unmarshal(raw, &l)
	return l
}

func must(err error) {
	if err != nil {
		fmt.Fprintln(os.Stderr, "play:", err)
		os.Exit(1)
	}
}