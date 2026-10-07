#!/usr/bin/env bash
# setup-pi.sh — run ON the Raspberry Pi (once) to finish provisioning Argos.
# Installs the Go arm64 toolchain, wires the USB-gadget network + mDNS, builds the
# binaries, and enables the always-on world + play services.
set -euo pipefail

GO_VER="1.26.4"
REPO="$HOME/argos"

echo "==> Argos Pi setup"
[ -d "$REPO" ] || { echo "expected the repo at $REPO (run pi/provision.sh first)"; exit 1; }

# --- Go toolchain (MANDATORY: spells, incl. portal, run via `go run`) ---
if ! "$HOME/.local/go/bin/go" version >/dev/null 2>&1; then
  echo "==> installing Go $GO_VER (arm64)"
  mkdir -p "$HOME/.local"
  curl -fsSL "https://go.dev/dl/go${GO_VER}.linux-arm64.tar.gz" | tar -C "$HOME/.local" -xz
fi
export PATH="$HOME/.local/go/bin:$PATH"
grep -q '.local/go/bin' "$HOME/.bashrc" 2>/dev/null || \
  echo 'export PATH="$HOME/.local/go/bin:$PATH"' >> "$HOME/.bashrc"
go version

# --- network: USB-C gadget address + mDNS ---
echo "==> network (usb0 + avahi)"
sudo install -m 0644 "$REPO/pi/usb0.network" /etc/systemd/network/usb0.network
sudo systemctl enable --now systemd-networkd
sudo apt-get update -y && sudo apt-get install -y avahi-daemon
sudo systemctl enable --now avahi-daemon

# --- build the dog + the game ---
echo "==> building binaries"
cd "$REPO"
mkdir -p bin
go build -o bin/world ./cmd/world
go build -o bin/play ./cmd/play
go build -o bin/cerebrum ./cmd/cerebrum || echo "  (cerebrum optional; skipped)"

# --- services (always-on) ---
echo "==> services"
mkdir -p "$HOME/.config/systemd/user"
install -m 0644 "$REPO/pi/argos-world.service" "$HOME/.config/systemd/user/argos-world.service"
install -m 0644 "$REPO/pi/argos-play.service"  "$HOME/.config/systemd/user/argos-play.service"
chmod +x "$REPO/scripts/live.sh"
sudo loginctl enable-linger "$USER"
systemctl --user daemon-reload
systemctl --user enable --now argos-world.service argos-play.service

echo "==> done. The dog is playing."
echo "    systemctl --user status argos-play"
echo "    journalctl --user -u argos-play -f"
