// Spell: autonomy-loop (wild school) — Argos's autonomous EE loop.
// Wakes, assesses hardware, plans work, executes, and reports.
// Runs in cycles: sense -> plan -> act -> report -> sleep.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"
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

	cycles := in.Args["cycles"]
	if cycles == "" { cycles = "1" }
	mode := in.Args["mode"]
	if mode == "" { mode = "assess" }

	switch mode {
	case "assess":
		// Sense phase: check what hardware is available
		var report strings.Builder
		report.WriteString("=== Argos EE Autonomy Loop ===\n")
		report.WriteString(fmt.Sprintf("Cycle started: %s\n", time.Now().Format(time.RFC3339)))

		// Check GPIO
		if _, err := os.Stat("/sys/class/gpio"); err == nil {
			report.WriteString("GPIO: available\n")
		} else {
			report.WriteString("GPIO: not available\n")
		}

		// Check I2C
		if _, err := os.Stat("/dev/i2c-1"); err == nil {
			report.WriteString("I2C: available\n")
		} else {
			report.WriteString("I2C: not available\n")
		}

		// Check SPI
		if _, err := os.Stat("/dev/spidev0.0"); err == nil {
			report.WriteString("SPI: available\n")
		} else {
			report.WriteString("SPI: not available\n")
		}

		// Check CAN
		canOut, _ := exec.Command("ip", "link", "show", "can0").CombinedOutput()
		if strings.Contains(string(canOut), "can0") {
			report.WriteString("CAN: available\n")
		} else {
			report.WriteString("CAN: not available\n")
		}

		// Check GPU
		gpuOut, _ := exec.Command("nvidia-smi", "--query-gpu=name,memory.total", "--format=csv,noheader").CombinedOutput()
		if len(gpuOut) > 0 {
			report.WriteString(fmt.Sprintf("GPU: %s", strings.TrimSpace(string(gpuOut))))
		}

		// Check serial ports
		serialOut, _ := exec.Command("python3", "-c", "import serial.tools.list_ports; ports = serial.tools.list_ports.comports(); [print(p.device) for p in ports]").CombinedOutput()
		if len(serialOut) > 0 {
			report.WriteString(fmt.Sprintf("Serial ports:\n%s", string(serialOut)))
		}

		// Check toolchain
		tools := []string{"arm-none-eabi-gcc", "aarch64-linux-gnu-gcc", "openocd", "sigrok-cli", "ngspice", "kicad-cli", "verilator", "yosys"}
		for _, t := range tools {
			if path, err := exec.LookPath(t); err == nil {
				report.WriteString(fmt.Sprintf("Tool %s: %s\n", t, path))
			}
		}

		report.WriteString(fmt.Sprintf("Cycle complete: %s\n", time.Now().Format(time.RFC3339)))
		json.NewEncoder(os.Stdout).Encode(Output{Summary: report.String()})

	case "plan":
		// Plan phase: read memory and decide what to do
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("Argos planning %s cycles of EE work", cycles)})

	case "act":
		// Execute phase: run the planned work
		task := in.Args["task"]
		if task == "" { task = "assess hardware" }
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("Argos executing: %s", task)})

	case "report":
		// Report phase: summarize what was done
		json.NewEncoder(os.Stdout).Encode(Output{Summary: "Argos EE cycle report"})

	default:
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("unknown mode: %s", mode)})
	}
}
