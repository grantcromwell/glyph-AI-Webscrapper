// Spell: serial-debug (wild school) — Argos communicates over serial/UART.
// Lists, opens, reads, writes, and monitors serial ports.
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
	port := in.Args["port"]
	baud := in.Args["baud"]
	if baud == "" { baud = "115200" }

	switch action {
	case "list":
		cmd := exec.Command("python3", "-c", `
import serial.tools.list_ports
ports = serial.tools.list_ports.comports()
for p in sorted(ports):
    print(f"{p.device} - {p.description} ({p.hwid})")
`)
		listOut, listErr := cmd.CombinedOutput()
		if listErr != nil {
			entries, _ := os.ReadDir("/dev")
			var ports []string
			for _, e := range entries {
				n := e.Name()
				if strings.HasPrefix(n, "ttyUSB") || strings.HasPrefix(n, "ttyACM") || strings.HasPrefix(n, "ttyS") {
					ports = append(ports, "/dev/"+n)
				}
			}
			json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("serial ports: %s", strings.Join(ports, ", "))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: strings.TrimSpace(string(listOut))})

	case "open":
		script := fmt.Sprintf(`
import serial, sys, time
try:
    s = serial.Serial('%s', %s, timeout=2)
    print(f"opened {s.port} @ {s.baudrate} baud")
    s.close()
except Exception as e:
    print(f"error: {e}")
`, port, baud)
		cmd := exec.Command("python3", "-c", script)
		openOut, openErr := cmd.CombinedOutput()
		if openErr != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", openErr, string(openOut))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: strings.TrimSpace(string(openOut))})

	case "read":
		bytes := in.Args["bytes"]
		if bytes == "" { bytes = "64" }
		timeout := in.Args["timeout"]
		if timeout == "" { timeout = "2" }
		script := fmt.Sprintf(`
import serial, sys, time
try:
    s = serial.Serial('%s', %s, timeout=%s)
    time.sleep(0.1)
    data = s.read(%s)
    s.close()
    hexdata = ' '.join(f'{b:02x}' for b in data)
    print(f"read {len(data)} bytes: {hexdata}")
except Exception as e:
    print(f"error: {e}")
`, port, baud, timeout, bytes)
		cmd := exec.Command("python3", "-c", script)
		readOut, readErr := cmd.CombinedOutput()
		if readErr != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", readErr, string(readOut))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: strings.TrimSpace(string(readOut))})

	case "write":
		data := in.Args["data"]
		if data == "" {
			json.NewEncoder(os.Stdout).Encode(Output{Error: "data required"})
			return
		}
		script := fmt.Sprintf(`
import serial, sys, time
try:
    s = serial.Serial('%s', %s, timeout=2)
    written = s.write(bytes.fromhex('%s'))
    s.close()
    print(f"wrote {written} bytes")
except Exception as e:
    print(f"error: {e}")
`, port, baud, data)
		cmd := exec.Command("python3", "-c", script)
		writeOut, writeErr := cmd.CombinedOutput()
		if writeErr != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", writeErr, string(writeOut))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: strings.TrimSpace(string(writeOut))})

	case "monitor":
		duration := in.Args["duration"]
		if duration == "" { duration = "10" }
		script := fmt.Sprintf(`
import serial, sys, time
try:
    s = serial.Serial('%s', %s, timeout=1)
    deadline = time.time() + %s
    count = 0
    while time.time() < deadline:
        data = s.read(1)
        if data:
            sys.stdout.write(data.decode('utf-8', errors='replace'))
            sys.stdout.flush()
            count += 1
    s.close()
    print(f"\n--- {count} chars received ---")
except Exception as e:
    print(f"error: {e}")
`, port, baud, duration)
		cmd := exec.Command("python3", "-c", script)
		monOut, monErr := cmd.CombinedOutput()
		if monErr != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", monErr, string(monOut))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: strings.TrimSpace(string(monOut))})

	default:
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("unknown action: %s", action)})
	}
}
