// Spell: spanish (scrying school) — the dog reads the Hispanic world in Spanish.
// Xataka & El País Tecnología (engineering/tech), El País & El Mundo portada
// (human/general news). Blockless native RSS.
package main

import (
	"encoding/json"
	"os"

	"glyphai/internal/modules"
	"glyphai/internal/news"
)

var feeds = []string{
	"https://www.xataka.com/feedburner.xml",                                      // engineering / tech
	"https://elpais.com/rss/tecnologia/portada.xml",                              // engineering / tech
	"https://feeds.elpais.com/mrss-s/pages/ep/site/elpais.com/portada",           // human / general
	"https://e00-elmundo.uecdn.es/elmundo/rss/portada.xml",                       // human / general
}

func main() {
	var in modules.Input
	_ = json.NewDecoder(os.Stdin).Decode(&in)
	out := news.News(in, "spanish", feeds)
	_ = json.NewEncoder(os.Stdout).Encode(out)
}