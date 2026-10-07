#!/usr/bin/env bash
# provision.sh — run on the DEV BOX after the Pi has booted and is reachable.
# Copies the repo to the Pi and runs setup-pi.sh there. Safe to re-run.
set -euo pipefail

HOST="${1:-argos@argos.local}"
REPO_SRC="$(cd "$(dirname "$0")/.." && pwd)"

echo "==> waiting for $HOST ..."
until ssh -o ConnectTimeout=3 -o StrictHostKeyChecking=accept-new "$HOST" true 2>/dev/null; do
  sleep 2
done

echo "==> rsyncing repo -> $HOST:~/argos"
rsync -az --delete \
  --exclude '.git' --exclude 'bin' --exclude 'world' \
  --exclude 'training/corpus' --exclude 'training/data' \
  "$REPO_SRC/" "$HOST:argos/"

echo "==> running setup-pi.sh on the Pi"
ssh "$HOST" 'bash ~/argos/pi/setup-pi.sh'
echo "==> provisioned. Visit:  ssh $HOST  then  ./bin/cerebrum stats"
