#!/bin/bash
# Passive watcher — just logs what Argos does, never interferes
LOG="$HOME/Desktop/native/argos_activity_log.md"
NOW=$(date "+%Y-%m-%d %H:%M:%S")

echo "" >> "$LOG"
echo "### $NOW" >> "$LOG"

# What files did he create/modify?
find "$HOME/Desktop/native/glyphify" -type f 2>/dev/null | while read f; do
    echo "- created: $f ($(wc -c < "$f") bytes)" >> "$LOG"
done

# What engrams did he file?
NEW_ENGRAMS=$(ls -lt "$HOME/Desktop/native/gut/lore/" 2>/dev/null | head -5)
if [ -n "$NEW_ENGRAMS" ]; then
    echo "- recent engrams:" >> "$LOG"
    echo "$NEW_ENGRAMS" >> "$LOG"
fi

# What did he coin?
if [ -f "$HOME/Desktop/native/myelin/lexicon.svg" ]; then
    TOTAL=$(grep -c "entry" "$HOME/Desktop/native/myelin/lexicon.svg" 2>/dev/null)
    echo "- lexicon entries: $TOTAL" >> "$LOG"
fi

# Memory vault changes
if [ -f "$HOME/Desktop/native/obsidian-mind/brain/Memories.md" ]; then
    LINES=$(wc -l < "$HOME/Desktop/native/obsidian-mind/brain/Memories.md")
    echo "- memory vault: $LINES lines" >> "$LOG"
fi
