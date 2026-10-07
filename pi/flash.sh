#!/usr/bin/env bash
# flash.sh — flash Raspberry Pi OS Lite (64-bit) to a microSD and configure the
# boot partition for headless USB-C access. DESTRUCTIVE: it erases the target.
#
#   sudo pi/flash.sh /dev/sdX [path/to/raspios.img.xz]
#
# After it finishes: eject, boot the Pi, plug USB-C into the laptop, then run
# `pi/provision.sh` from the dev box.
set -euo pipefail

DEV="${1:-}"
IMG="${2:-}"
IMG_URL="https://downloads.raspberrypi.com/raspios_lite_arm64_latest"

[ -n "$DEV" ] || { echo "usage: sudo pi/flash.sh /dev/sdX [image.img.xz]"; exit 2; }
[ "$(id -u)" -eq 0 ] || { echo "must run as root (sudo)"; exit 2; }

# --- safety: refuse anything that isn't a real removable SD-sized disk ---
[ -b "$DEV" ] || { echo "$DEV is not a block device"; exit 2; }
case "$DEV" in /dev/sd*|/dev/mmcblk*) : ;; *) echo "refusing non-SD device $DEV"; exit 2;; esac
ROOTDISK="$(lsblk -no PKNAME "$(findmnt -no SOURCE /)" 2>/dev/null || true)"
[ -n "$ROOTDISK" ] && [ "/dev/$ROOTDISK" = "$DEV" ] && { echo "$DEV is the system disk — refusing"; exit 2; }
SIZE_B="$(blockdev --getsize64 "$DEV")"
[ "$SIZE_B" -gt 0 ] || { echo "$DEV reports 0 bytes — reseat the card"; exit 2; }
GB=$(( SIZE_B / 1000000000 ))
echo "Target: $DEV  (~${GB} GB)  $(lsblk -no MODEL "$DEV" | head -1)"
read -r -p "Type ERASE to wipe and flash this device: " ok
[ "$ok" = "ERASE" ] || { echo "aborted"; exit 1; }

# --- fetch image if not supplied ---
if [ -z "$IMG" ]; then
  IMG="/tmp/raspios_lite_arm64.img.xz"
  [ -f "$IMG" ] || { echo "==> downloading Raspberry Pi OS Lite (64-bit)"; curl -fL "$IMG_URL" -o "$IMG"; }
fi

# --- write ---
echo "==> unmounting any mounted partitions on $DEV"
for p in "${DEV}"*; do umount "$p" 2>/dev/null || true; done
echo "==> writing image (this takes a few minutes)"
case "$IMG" in
  *.xz) xzcat "$IMG" | dd of="$DEV" bs=4M conv=fsync status=progress ;;
  *)    dd if="$IMG" of="$DEV" bs=4M conv=fsync status=progress ;;
esac
sync; partprobe "$DEV" 2>/dev/null || true; sleep 2

# --- configure the boot partition (headless + USB-C gadget) ---
BOOT="$(mktemp -d)"
BOOTPART="${DEV}1"; [[ "$DEV" == *mmcblk* ]] && BOOTPART="${DEV}p1"
mount "$BOOTPART" "$BOOT"
echo "==> enabling SSH + USB-C ethernet gadget"
touch "$BOOT/ssh"
grep -q '^dtoverlay=dwc2' "$BOOT/config.txt" || printf '\n# Argos: USB-C as a USB gadget\ndtoverlay=dwc2\n' >> "$BOOT/config.txt"
CMD="$BOOT/cmdline.txt"
grep -q 'modules-load=dwc2,g_ether' "$CMD" || sed -i 's/rootwait/rootwait modules-load=dwc2,g_ether/' "$CMD"
# default user argos / password argos (change after first login)
echo "argos:$(openssl passwd -6 argos)" > "$BOOT/userconf.txt"
sync; umount "$BOOT"; rmdir "$BOOT"

echo "==> flashed. Eject, boot the Pi, plug USB-C into the laptop,"
echo "    then from the dev box:  pi/provision.sh argos@argos.local"
