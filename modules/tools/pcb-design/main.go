// Spell: pcb-design (wild school) — Argos designs PCBs.
// Shells out to KiCad CLI for project creation, schematic, layout, and Gerber export.
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
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
	name := in.Args["name"]
	dir := in.Path
	if dir == "" { dir = "." }

	switch action {
	case "create":
		proj := filepath.Join(dir, name+".pro")
		if err := os.WriteFile(proj, []byte(fmt.Sprintf("[%s]\nversion=1\n", name)), 0644); err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: err.Error()})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("created KiCad project %s in %s", name, dir)})

	case "schematic":
		sch := filepath.Join(dir, name+".kicad_sch")
		template := `(kicad_sch (version 2024) (generator "argos")
  (paper "A4")
  (title_block (title "` + name + `") (date "` + timeNow() + `") (rev "1.0"))
)`
		if err := os.WriteFile(sch, []byte(template), 0644); err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: err.Error()})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("created schematic %s", sch)})

	case "gerber":
		outDir := filepath.Join(dir, "gerber")
		os.MkdirAll(outDir, 0755)
		cmd := exec.Command("kicad-cli", "pcb", "export", "gerbers", "-o", outDir, filepath.Join(dir, name+".kicad_pcb"))
		if out, err := cmd.CombinedOutput(); err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", err, string(out))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("exported Gerbers to %s", outDir)})

	case "bom":
		cmd := exec.Command("kicad-cli", "sch", "export", "bom", "-o", filepath.Join(dir, name+"_bom.csv"), filepath.Join(dir, name+".kicad_sch"))
		if out, err := cmd.CombinedOutput(); err != nil {
			json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("%s: %s", err, string(out))})
			return
		}
		json.NewEncoder(os.Stdout).Encode(Output{Summary: fmt.Sprintf("exported BOM for %s", name)})

	default:
		json.NewEncoder(os.Stdout).Encode(Output{Error: fmt.Sprintf("unknown action: %s", action)})
	}
}

func timeNow() string {
	return "2026-07-24"
}
