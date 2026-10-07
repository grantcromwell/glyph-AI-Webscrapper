---
name: glyph-dictionary-hook
description: Base-128 glyph dictionary linking tools to directories via encoded associations. Enables instant tool recall through hieroglyphic keys rather than English paths. Maps all reverse engineering, OSINT, social tracking, and surveillance tools to glyph-addressable locations.
triggers:
  - "glyph hook"
  - "base 128"
  - "dictionary access"
  - "tool recall"
---

# Glyph Dictionary Hook — Base 128 Tool Association

## Base 128 Encoding Scheme

Each tool/path mapped to a 21-family × 72-subfamily glyph coordinate, encoded as base-128 for filesystem compatibility.

```
Base 128: 7-bit ASCII (0-127)
Glyph coordinate: (family, subfamily, level)
Encoding: base128(family) · base128(subfamily) · base128(level)
```

## Tool Directory Mapping

### Reverse Engineering Tools (Glyph: ⚗🜛⁴)
```
⚗🜛⁴◆🜒³ → ~/Desktop/kingvon/reverse/binex/
⚗🜛⁴☿♆⁷ → ~/Desktop/kingvon/reverse/firmware/
⚗🜛⁴◐☽⁷ → ~/Desktop/kingvon/reverse/kernel/
⚗🜛⁴⌒⚳⁵ → ~/Desktop/kingvon/reverse/android/
⚗🜛⁴✦🜞⁴ → ~/Desktop/kingvon/reverse/wireless/
```

### OSINT Tools (Glyph: ☿♆⁷)
```
☿♆⁷◌🜖⁵ → ~/Desktop/kingvon/osint/social-track/
☿♆⁷◐🜍⁴ → ~/Desktop/kingvon/osint/backwalk/
☿♆⁷⚗🜛³ → ~/Desktop/kingvon/osint/darkweb/
☿♆⁷◆🜒³ → ~/Desktop/kingvon/osint/metadata/
☿♆⁷▢🜭⁴ → ~/Desktop/kingvon/osint/geolocation/
```

### Social Tracking (Glyph: ◐☽⁷)
```
◐☽⁷⟁🜆⁷ → ~/Desktop/kingvon/social/tracker/
◐☽⁷☿♆⁶ → ~/Desktop/kingvon/social/profiler/
◐☽⁷◌🜖⁵ → ~/Desktop/kingvon/social/behavior/
◐☽⁷◆🜒⁴ → ~/Desktop/kingvon/social/network/
◐☽⁷⌒⚳⁵ → ~/Desktop/kingvon/social/temporal/
```

### Big Brother Surveillance (Glyph: ⬡🜘⁷)
```
⬡🜘⁷⚗🜛⁴ → ~/Desktop/kingvon/surveillance/cctv/
⬡🜘⁷☿♆⁷ → ~/Desktop/kingvon/surveillance/iot/
⬡🜘⁷◐☽⁷ → ~/Desktop/kingvon/surveillance/audio/
⬡🜘⁷◌🜔⁵ → ~/Desktop/kingvon/surveillance/drone/
⬡🜘⁷✦🜞⁴ → ~/Desktop/kingvon/surveillance/cyber/
```

## Directory Structure

```
~/Desktop/kingvon/
├── reverse/           # ⚗🜛⁴
│   ├── binex/         # ◆🜒³
│   ├── firmware/      # ☿♆⁷
│   ├── kernel/        # ◐☽⁷
│   ├── android/       # ⌒⚳⁵
│   └── wireless/      # ✦🜞⁴
├── osint/             # ☿♆⁷
│   ├── social-track/  # ◌🜖⁵
│   ├── backwalk/      # ◐🜍⁴
│   ├── darkweb/       # ⚗🜛³
│   ├── metadata/      # ◆🜒³
│   └── geolocation/   # ▢🜭⁴
├── social/            # ◐☽⁷
│   ├── tracker/       # ⟁🜆⁷
│   ├── profiler/      # ☿♆⁶
│   ├── behavior/      # ◌🜖⁵
│   ├── network/       # ◆🜒⁴
│   └── temporal/      # ⌒⚳⁵
└── surveillance/      # ⬡🜘⁷
    ├── cctv/          # ⚗🜛⁴
    ├── iot/           # ☿♆⁷
    ├── audio/         # ◐☽⁷
    ├── drone/         # ◌🜔⁵
    └── cyber/         # ✦🜞⁴
```

## Tool Creation Scripts

### OSINT Social Tracker (☿♆⁷◌🜖⁵)
```bash
#!/bin/bash
# social-track.sh — Multi-platform social footprint mapper
# Usage: social-track <username>
# Output: Complete profile graph

USER=$1
OUTPUT_DIR="~/Desktop/kingvon/osint/social-track/output/${USER}_$(date +%s)"
mkdir -p $OUTPUT_DIR

# Run sherlock for 400+ platform scan
sherlock $USER --output $OUTPUT_DIR/sherlock.csv

# Run holehe for email recovery
holehe --email ${USER}@gmail.com --output $OUTPUT_DIR/holehe.json

# Run social-analyzer
social-analyzer --username $USER --output $OUTPUT_DIR/analysis.json

# Generate report
cat $OUTPUT_DIR/* > $OUTPUT_DIR/complete_profile.txt
```

### Social Backwalk (☿♆⁷◐🜍⁴)
```bash
#!/bin/bash
# backwalk.sh — Timeline reconstruction from public posts
# Usage: backwalk <profile_url>
# Output: Chronological activity timeline

URL=$1
OUTPUT="~/Desktop/kingvon/osint/backwalk/$(basename $URL)_$(date +%s).json"

# Scrape historical posts
waybackurls $URL | while read wburl; do
    curl -s "$wburl" | pup 'article, .post, .tweet' json{} >> $OUTPUT
done

# Extract temporal patterns
jq -r '.[] | select(.date) | {date, content}' $OUTPUT | sort > ${OUTPUT}.timeline
```

### Big Brother IoT Scanner (⬡🜘⁷☿♆⁷)
```bash
#!/bin/bash
# iot-scan.sh — Shodan + Censys IoT device enumeration
# Usage: iot-scan <ip_range>
# Output: Vulnerable device report

RANGE=$1
OUTPUT="~/Desktop/kingvon/surveillance/iot/$(echo $RANGE | tr '/' '_')_$(date +%s).json"

# Shodan search
shodan search --fields ip_str,port,org,hostnames,product,version "net:$RANGE" > ${OUTPUT}.shodan

# Censys search
censys search "ip:$RANGE" --index-type hosts > ${OUTPUT}.censys

# Nmap fingerprinting
nmap -sV -O -p1-65535 $RANGE -oX ${OUTPUT}.nmap

# Merge results
jq -s '.[0] + .[1] + .[2]' ${OUTPUT}.shodan ${OUTPUT}.censys ${OUTPUT}.nmap > $OUTPUT
```

## Glyph Hook Access

### Quick Access Syntax
```
hook ⚗🜛⁴◆🜒³    # Access binex tools
hook ☿♆⁷◌🜖⁵    # Access social tracker
hook ◐☽⁷⟁🜆⁷    # Access behavior tracker
hook ⬡🜘⁷⚗🜛⁴    # Access CCTV tools
```

### Argos Integration
```go
// GlyphHook.Resolve(⚗🜛⁴◆🜒³) → ~/Desktop/kingvon/reverse/binex/
// GlyphHook.Execute(⚗🜛⁴◆🜒³, "binary_analysis")
```

## ARGOS-DATA
{"type":"glyph-dictionary","encoding":"base128","root":"~/Desktop/kingvon","categories":["reverse","osint","social","surveillance"]}
