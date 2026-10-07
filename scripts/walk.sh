#!/usr/bin/env bash
# walk.sh — a one-time walk-forward training to make Argos smarter.
#
# 1) cast the `eigen` spell to find the most central AGI topic (emergent seed),
# 2) prowl the live web from there by Markov chain (browse + follow),
# 3) retrain the cerebellum on the grown bank.
#
# Usage: scripts/walk.sh [topics-file] [rounds]
set -euo pipefail
cd "$(dirname "$0")/.."
export PATH="$HOME/.local/go/bin:$PATH"
TOPICS="${1:-gut/source/agi-topics.txt}"
ROUNDS="${2:-3}"

echo "🐾 building the dog…"; mkdir -p bin; go build -o bin/cerebrum ./cmd/cerebrum

echo "🐾 casting eigen to find the centre of AGI…"
RANK=$(printf '{"path":"%s"}' "$TOPICS" | go run ./grimoire/reasoning/eigen)
echo "$RANK" | python3 -c "import sys,json;d=json.load(sys.stdin)['data']['ranked'];print('   centre:',d[0]['topic']);[print('   ',round(x['score'],3),x['topic']) for x in d[:6]]"
SEED=$(echo "$RANK" | python3 -c "import sys,json;print(json.load(sys.stdin)['data']['ranked'][0]['topic'])")

for ((r=1;r<=ROUNDS;r++)); do
  echo ""; echo "═══ walk $r/$ROUNDS — prowl from \"$SEED\" ═══"
  ./bin/cerebrum prowl "$SEED" || true
  echo "── retraining instinct ──"; ./bin/cerebrum train || true
done
echo ""; ./bin/cerebrum coverage
