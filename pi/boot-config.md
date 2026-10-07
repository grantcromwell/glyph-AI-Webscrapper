# Boot-partition edits for the USB-C ethernet gadget (Raspberry Pi 4)

These two edits turn the Pi 4's USB-C port into a USB ethernet device, so plugging
it into a laptop makes the dog reachable at `10.55.0.1` / `argos.local` with no
monitor and no network setup. Apply them on the **boot** partition (`bootfs`) after
flashing — `pi/flash.sh` does this automatically.

## 1. `config.txt` — append
```
# Argos: USB-C as a USB gadget
dtoverlay=dwc2
```

## 2. `cmdline.txt` — insert right after `rootwait` (stay on ONE line, space-separated)
```
modules-load=dwc2,g_ether
```

## 3. Headless access (also on the boot partition)
- Create an empty file named `ssh` to enable SSH on first boot.
- Create `userconf.txt` with `argos:<openssl-passwd-6-hash>` to set the `argos` user
  (or use Raspberry Pi Imager's "set username" — `pi/flash.sh` writes this for you).

## On the root partition (`rootfs`), `pi/setup-pi.sh` finishes the job:
- installs `pi/usb0.network` to `/etc/systemd/network/` and enables `systemd-networkd`,
- installs Avahi so `argos.local` resolves,
- installs the Go arm64 toolchain to `~/.local/go`,
- drops `argos-world.service` + `argos-play.service` and enables lingering.

> Pi 4 only. The Pi 5's USB-C is power-delivery oriented and its gadget path differs.
