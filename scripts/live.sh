#!/usr/bin/env bash
# live.sh — keep Argos playing its world forever.
#
# The always-on loop the argos-play service runs: the dog steps through the portal
# into the headless game and plays in rolling legs. cmd/play re-enters at the
# spawn door whenever a trail dies, and the service restarts this script if it
# ever exits, so the dog never stops exploring.
#
# Requires Go on PATH (spells, incl. portal, run via `go run`) and the world
# server (argos-world.service) up on $ARGOS_WORLD.
set -u

export PATH="$HOME/.local/go/bin:$PATH"
cd "$(dirname "$0")/.." || exit 1   # repo root (so grimoire/ is discoverable)

WORLD_URL="${ARGOS_WORLD:-http://127.0.0.1:8088/index.html}"
BANK="${ARGOS_BANK:-world}"
MAP="${ARGOS_MAP:-hippocampus/cartograph.svg}"
LEG="${ARGOS_LEG:-1800}"           # seconds per leg (30 min)

PLAY="./bin/play"
[ -x "$PLAY" ] || PLAY="go run ./cmd/play"

while true; do
  $PLAY -until "$(( $(date +%s) + LEG ))" -seed "$WORLD_URL" -bank "$BANK" -map "$MAP"
  sleep 1
done
