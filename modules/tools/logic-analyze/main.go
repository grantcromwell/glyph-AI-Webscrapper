// Spell: logic-analyze (wild school) — Argos captures and decodes logic signals.
// Uses sigrok-cli for protocol decoding.
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
	protocol := in.Args["protocol"]
	duration := in.Args["duration"]
	if duration == "" { duration = "1" }
	channels := in.Args["channels"]
	if channels == "" { channels = "0=CS,1=MISO,2=MOSI,3=CLK" }

	switch action {
	case "capture":
		driver := in.Args["driver"]
		if driver == "" { driver = "fx2lafw" }
		outFile := in.Path
		if outFile == "" { outFile = "/tmp/argos_capture.sr" }
		args := []string{"--driver", driver, "--config", fmt.Sprintf("samplerate=%s", in.Args["samplerate"]),
			"--channels", channels, "--time", duration, "-o", outFile}
		cmd := exec.Command("sigrok-cli", args...)
		if out, err := cmd.CombinedOutput(); err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", err, string(out))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("captured %ss to %s", duration, outFile)})

	case "decode":
		srFile := in.Path
		if srFile == "" { srFile = "/tmp/argos_capture.sr" }
		args := []string{"--input-file", srFile, "--protocol-decoders", protocol, "--channels", channels}
		cmd := exec.Command("sigrok-cli", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", err, string(out))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("decoded %s: %s", protocol, string(out))})

	case "analyze":
		// Run decode then analyze the output
		srFile := in.Path
		if srFile == "" { srFile = "/tmp/argos_capture.sr" }
		args := []string{"--input-file", srFile, "--protocol-decoders", protocol, "--channels", channels}
		cmd := exec.Command("sigrok-cli", args...)
		out, err := cmd.CombinedOutput()
		if err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", err, string(out))})
			return
		}
		lines := strings.Split(string(out), "\n")
		summary := fmt.Sprintf("decoded %s: %d lines, first: %s", protocol, len(lines), lines[0])
		if len(lines) > 1 { summary += fmt.Sprintf(", last: %s", lines[len(lines)-1]) }
		json.NewEncoder(os.Stdout).Encode(Output{Summary: summary})

	default:
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("unknown action: %s", action)})
	}
}
