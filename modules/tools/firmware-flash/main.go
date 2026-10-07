// Spell: firmware-flash (wild school) — Argos flashes firmware to microcontrollers.
// Supports OpenOCD, ST-Link, and direct SWD access.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
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
	target := in.Args["target"]
	firmwareFile := in.Path
	if firmwareFile == "" { firmwareFile = in.Args["file"] }
	iface := in.Args["interface"]
	if iface == "" { iface = "stlink" }

	switch action {
	case "flash":
		if firmwareFile == "" {
			json.NewEncoder(os.Stdout).Encode(Output{Error: "firmware file required"})
			return
		}
		// Build OpenOCD command
		args := []string{"-f", fmt.Sprintf("interface/%s.cfg", iface), "-f", "target/" + target + ".cfg",
			"-c", fmt.Sprintf("program %s verify reset exit", firmwareFile)}
		cmd := exec.Command("openocd", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", err, string(out))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("flashed %s to %s via %s", firmwareFile, target, iface)})

	case "read":
		outFile := in.Args["output"]
		if outFile == "" { outFile = target + "_dump.bin" }
		args := []string{"-f", fmt.Sprintf("interface/%s.cfg", iface), "-f", "target/" + target + ".cfg",
			"-c", fmt.Sprintf("flash read_bank 0 %s 0 0x100000; reset; exit", outFile)}
		cmd := exec.Command("openocd", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", err, string(out))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("read flash from %s to %s", target, outFile)})

	case "erase":
		args := []string{"-f", fmt.Sprintf("interface/%s.cfg", iface), "-f", "target/" + target + ".cfg",
			"-c", "flash erase_sector 0 0 last; reset; exit"}
		cmd := exec.Command("openocd", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", err, string(out))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("erased %s", target)})

	case "verify":
		if firmwareFile == "" {
			json.NewEncoder(os.Stdout).Encode(Output{Error: "firmware file required"})
			return
		}
		args := []string{"-f", fmt.Sprintf("interface/%s.cfg", iface), "-f", "target/" + target + ".cfg",
			"-c", fmt.Sprintf("program %s verify reset exit", firmwareFile)}
		cmd := exec.Command("openocd", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("verify failed: %s: %s", err, string(out))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("verified %s on %s", firmwareFile, target)})

	default:
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("unknown action: %s", action)})
	}
}
