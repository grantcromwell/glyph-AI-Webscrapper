// Spell: french (scrying school) — the dog reads the Francophone world in French.
// Les Numériques (engineering/tech), Le Monde, France24 & Le Figaro (human/
// general news). Blockless native RSS.
package main

import (
	"encoding/json"
	"os"

	"glyphai/internal/modules"
	"glyphai/internal/news"
)

var feeds = []string{
	"https://www.lesnumeriques.com/feeds.xml",              // engineering / tech
	"https://www.lemonde.fr/rss/une.xml",                 // human / general
	"https://www.france24.com/fr/rss",                    // human / general
	"https://www.lefigaro.fr/rss/figaro_actualites.xml",  // human / general
}

func main() {
	var in modules.Input
	_ = json.NewDecoder(os.Stdin).Decode(&in)
	out := news.News(in, "french", feeds)
	_ = json.NewEncoder(os.Stdout).Encode(out)
}