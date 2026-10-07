// Spell: russian (scrying school) — the dog reads Russia in Russian.
// Native front door, no English filter: Habr + N+1 (engineering/science),
// Lenta + RIA (human/general news). Blockless RSS.
package main

import (
	"encoding/json"
	"os"

	"glyphai/internal/modules"
	"glyphai/internal/news"
)

var feeds = []string{
	"https://habr.com/ru/rss/articles/?fl=ru",      // engineering / tech community
	"https://nplus1.ru/rss",                         // science / engineering
	"https://lenta.ru/rss/news",                     // human / general
	"https://ria.ru/export/rss2/archive/index.xml",  // human / general
}

func main() {
	var in modules.Input
	_ = json.NewDecoder(os.Stdin).Decode(&in)
	out := news.News(in, "russian", feeds)
	_ = json.NewEncoder(os.Stdout).Encode(out)
}