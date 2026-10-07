// Spell: can-bus (wild school) — Argos interfaces with CAN bus.
// Uses SocketCAN for sniffing, sending, filtering, and analyzing CAN frames.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
)

type Input struct {
	Path string            `json:"path"`
	Args map[string]string `json:"args"`
}

type Output struct {
	Summary string `json:"summary"`
	Error   string `json:"error,omitempty"`
}

func main() {
	var in Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("reading input: %v", err)})
		return
	}
	action := in.Args["action"]
	iface := in.Args["interface"]
	if iface == "" { iface = "can0" }
	bitrate := in.Args["bitrate"]
	if bitrate == "" { bitrate = "500000" }

	switch action {
	case "setup":
		// Bring up CAN interface
		cmds := [][]string{
			{"ip", "link", "set", iface, "down"},
			{"ip", "link", "set", iface, "type", "can", "bitrate", bitrate},
			{"ip", "link", "set", iface, "up"},
		}
		for _, c := range cmds {
			cmd := exec.Command(c[0], c[1:]...)
			if out, err := cmd.CombinedOutput(); err != nil {
				json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", err, string(out))})
				return
			}
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("CAN %s up at %s bps", iface, bitrate)})

	case "sniff":
		duration := in.Args["duration"]
		if duration == "" { duration = "5" }
		cmd := exec.Command("candump", iface, "-L", "-n", "100", "-T", duration)
		sniffOut, sniffErr := cmd.CombinedOutput()
		if sniffErr != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", sniffErr, string(sniffOut))})
			return
		}
		lines := strings.Split(strings.TrimSpace(string(sniffOut)), "\n")
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("captured %d CAN frames on %s", len(lines), iface)})

	case "send":
		id := in.Args["id"]
		data := in.Args["data"]
		if id == "" || data == "" {
			json.NewEncoder(os.Stdout).Encode(Output{Error: "id and data required"})
			return
		}
		cmd := exec.Command("cansend", iface, fmt.Sprintf("%s#%s", id, data))
		sendOut, sendErr := cmd.CombinedOutput()
		if sendErr != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", sendErr, string(sendOut))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("sent CAN %s#%s on %s", id, data, iface)})

	case "filter":
		id := in.Args["id"]
		mask := in.Args["mask"]
		if mask == "" { mask = "7FF" }
		cmd := exec.Command("cangw", "-A", "-s", iface, "-d", iface, "-f", fmt.Sprintf("%s:%s", id, mask))
		filtOut, filtErr := cmd.CombinedOutput()
		if filtErr != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", filtErr, string(filtOut))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("CAN filter set: id=%s mask=%s", id, mask)})

	case "analyze":
		// Dump CAN stats
		cmd := exec.Command("ip", "-details", "-statistics", "link", "show", iface)
		statOut, statErr := cmd.CombinedOutput()
		if statErr != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", statErr, string(statOut))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: strings.TrimSpace(string(statOut))})

	case "bitrate":
		cmd := exec.Command("ip", "-details", "link", "show", iface)
		rateOut, rateErr := cmd.CombinedOutput()
		if rateErr != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", rateErr, string(rateOut))})
			return
		}
		for _, line := range strings.Split(string(rateOut), "\n") {
			if strings.Contains(line, "bitrate") {
				json.NewEncoder(os.Stdout).Encode(Output{Summary: strings.TrimSpace(line)})
				return
			}
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("CAN %s status", iface)})

	default:
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("unknown action: %s", action)})
	}
}
