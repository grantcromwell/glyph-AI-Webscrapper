// Spell: chinese (scrying school) — the dog reads China in Chinese.
// Solidot (engineering/tech, the Chinese Slashdot), People's Daily (human/state
// news), and Google News-zh (broad native aggregation). Blockless RSS.
package main

import (
	"encoding/json"
	"os"

	"glyphai/internal/modules"
	"glyphai/internal/news"
)

var feeds = []string{
	"https://www.solidot.org/index.rss",                          // engineering / tech
	"http://www.people.com.cn/rss/politics.xml",                  // human / state
	"https://news.google.com/rss?hl=zh-CN&gl=CN&ceid=CN:zh-Hans", // broad native
}

func main() {
	var in modules.Input
	_ = json.NewDecoder(os.Stdin).Decode(&in)
	out := news.News(in, "chinese", feeds)
	_ = json.NewEncoder(os.Stdout).Encode(out)
}