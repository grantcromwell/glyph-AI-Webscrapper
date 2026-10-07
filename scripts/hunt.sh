#!/usr/bin/env bash
# hunt.sh — send Argos hunting until its lore is rich.
#
# Each round, for each topic, the hunt casts the grimoire's site-spells like a
# warlock (currently wikipedia + arxiv — add a spell to add a source; the nose
# names no sites). What they bring back the nose files as logograms, then the
# dog retrains its instinct on the grown bank. The lore grows toward the Fable-5
# pretense of broad, dense knowledge — more rounds, richer bank, better answers.
#
# Usage:  scripts/hunt.sh [rounds]      (default 8)
#         ARGOS_TOPICS="a;b;c" scripts/hunt.sh 3   # custom topics
#
# Honest note: "until it reminds you of Fable 5" is an aspiration, not a hard
# stop — richer lore = better recall, but reasoning is still bounded by the
# small instinct net. Ctrl-C any time; progress is saved as art after each sniff.

set -euo pipefail
cd "$(dirname "$0")/.."
export PATH="$HOME/.local/go/bin:$PATH"

ROUNDS="${1:-8}"

# Build the mind once so the loop doesn't recompile on every call.
echo "🐾 building the dog…"
mkdir -p bin
go build -o bin/cerebrum ./cmd/cerebrum
DOG="./bin/cerebrum"

# Default topics: subjects Wikipedia + arXiv cover well (pattern analysis, the
# dog's own cosmology, cognition).
DEFAULT_TOPICS="wavelet transform;fourier analysis;information theory;pattern recognition;neural network;owl;corvidae intelligence;border collie;cuneiform;alchemy"
IFS=';' read -r -a TOPICS <<< "${ARGOS_TOPICS:-$DEFAULT_TOPICS}"

echo "🐾 Argos begins the hunt — $ROUNDS rounds over ${#TOPICS[@]} scents."
for ((r=1; r<=ROUNDS; r++)); do
  echo ""
  echo "═══ round $r/$ROUNDS ═══"
  for t in "${TOPICS[@]}"; do
    [ -z "$t" ] && continue
    "$DOG" sniff "$t" || echo "   (skipped: $t)"
  done
  echo "── retraining instinct on the grown bank ──"
  "$DOG" train || true
done

echo ""
echo "🐾 hunt complete."
"$DOG" stats
echo "Ask the dog:  "$DOG" ask \"...\""
