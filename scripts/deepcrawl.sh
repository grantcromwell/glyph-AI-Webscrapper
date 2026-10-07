#!/usr/bin/env bash
# Deep crawl of a local static site with the dog's light `skim` eyes.
# Spreads hunt seeds across the whole corpus and rotates the pull-theme so the
# Markov walk reaches every region, retraining the cerebellum periodically so the
# gathered engrams sharpen instinct as it goes.
#
#   scripts/deepcrawl.sh <baseURL> <pageCount> [hopsPerHunt] [seedStride]
# e.g. scripts/deepcrawl.sh http://127.0.0.1:8077 4665 32 30
set -euo pipefail
cd "$(dirname "$0")/.."
export PATH="$HOME/.local/go/bin:$PATH"

BASE="${1:-http://127.0.0.1:8077}"
PAGES="${2:-4665}"
HOPS="${3:-32}"
STRIDE="${4:-30}"

THEMES=(reasoning debugging testing attention memory rendering refactor \
        performance architecture scaffolding verification concurrency error \
        optimization network graphics planning recap)

go build -o bin/cerebrum ./cmd/cerebrum/

echo "🐾 deep crawl: base=$BASE pages=$PAGES hops=$HOPS stride=$STRIDE"
echo "   start bank: $(./bin/cerebrum coverage 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g')"

hunt=0
ti=0
for ((p=0; p<PAGES; p+=STRIDE)); do
  url=$(printf "%s/t%05d.html" "$BASE" "$p")
  theme="${THEMES[$((ti % ${#THEMES[@]}))]}"
  ti=$((ti+1)); hunt=$((hunt+1))
  ./bin/cerebrum crawl "$url" "$theme" "$HOPS" >/dev/null 2>&1 || true
  if (( hunt % 25 == 0 )); then
    ./bin/cerebrum train >/dev/null 2>&1 || true
    cov=$(./bin/cerebrum coverage 2>/dev/null | sed 's/\x1b\[[0-9;]*m//g')
    echo "   [$hunt hunts] $cov"
  fi
done

echo "🐾 final retrain…"
./bin/cerebrum train 2>&1 | sed 's/\x1b\[[0-9;]*m//g' | tail -3
echo "🐾 deep crawl done after $hunt hunts."
./bin/cerebrum coverage 2>&1 | sed 's/\x1b\[[0-9;]*m//g'
