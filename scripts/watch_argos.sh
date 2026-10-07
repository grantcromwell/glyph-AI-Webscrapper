#!/bin/bash
# Watch what Argos is doing - check for new files, modified files, and his latest thoughts
WATCH_DIRS=("$HOME/Desktop/native/glyphify" "$HOME/Desktop/native/obsidian-mind/brain" "$HOME/Desktop/native/gut/lore")
REPORT="$HOME/Desktop/native/argos_activity_log.md"

echo "=== Argos Activity Check: $(date) ===" >> "$REPORT"

# Check for new/modified files in glyphify
if [ -d "$HOME/Desktop/native/glyphify" ]; then
    echo "" >> "$REPORT"
    echo "## glyphify/ — new files" >> "$REPORT"
    find "$HOME/Desktop/native/glyphify" -type f -newer /tmp/argos_last_check 2>/dev/null | head -20 >> "$REPORT"
    echo "" >> "$REPORT"
    echo "## glyphify/ — all files" >> "$REPORT"
    find "$HOME/Desktop/native/glyphify" -type f 2>/dev/null | head -30 >> "$REPORT"
fi

# Check for new engrams (what he's been thinking about)
if [ -d "$HOME/Desktop/native/gut/lore" ]; then
    echo "" >> "$REPORT"
    echo "## New engrams (last 10)" >> "$REPORT"
    ls -lt "$HOME/Desktop/native/gut/lore/" 2>/dev/null | head -10 >> "$REPORT"
fi

# Check his memory vault for new entries
if [ -f "$HOME/Desktop/native/obsidian-mind/brain/Memories.md" ]; then
    echo "" >> "$REPORT"
    echo "## Memory vault size" >> "$REPORT"
    wc -l "$HOME/Desktop/native/obsidian-mind/brain/"*.md 2>/dev/null >> "$REPORT"
fi

# Check what he's coined recently
if [ -f "$HOME/Desktop/native/myelin/lexicon.svg" ]; then
    echo "" >> "$REPORT"
    echo "## Lexicon entries" >> "$REPORT"
    grep -c "entry" "$HOME/Desktop/native/myelin/lexicon.svg" 2>/dev/null >> "$REPORT"
fi

touch /tmp/argos_last_check
