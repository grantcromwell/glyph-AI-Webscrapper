// Package modules — deterministic analysis modules (spells) discovered on disk.
//
// DESIGN PATTERN: The module system is the opposite of an LLM plugin framework.
// Each module is a standalone executable that reads JSON on stdin and writes
// JSON on stdout. The orchestrator discovers them by walking a directory tree
// for manifest.json files. There is no shared library, no plugin ABI, no
// runtime code loading. Modules can be written in any language (Go, Python,
// Rust, C) as long as they speak the JSON contract.
//
// This is a deliberate architectural choice: modules are decoupled from the
// core by a process boundary. A crash in a module does not crash the system.
// A module can be updated without recompiling the core. A module can be
// written in the language best suited to its task.
//
// Pattern analysis (P9 SPI MITM, P25 Constraint Compression):
// The JSON contract between core and modules is a constraint compression of
// the interface surface. The core never needs to know what language a module
// is written in, what libraries it uses, or how it works internally. The
// module never needs to know about the core's memory model, concurrency, or
// persistence. The interface is the compression boundary.
//
// Key invariants:
//   - Modules are discovered by walking a directory tree for manifest.json
//   - Each module is a subprocess with stdin/stdout JSON contract
//   - Modules have a 30-second timeout
//   - The core augments PATH so modules can invoke Go/Python toolchains
//   - No shared memory, no plugin ABI, no runtime code loading
package modules

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"sort"
	"strings"
	"time"

	"glyphai/internal/encoding"
)

// Input is what the cortex hands a spell on stdin.
type Input struct {
	Text string         `json:"text,omitempty"` // text to analyse
	Path string         `json:"path,omitempty"` // file/image path to analyse
	Args map[string]any `json:"args,omitempty"` // spell-specific knobs
}

// Output is what a spell returns on stdout. A spell fills Detail (the fine
// 21x72 field); the cortex derives the coarse Glyph and richness from it.
type Output struct {
	Spell      string         `json:"spell"`
	Detail     encoding.Detail   `json:"detail"`
	Summary    string         `json:"summary,omitempty"`
	Data       map[string]any `json:"data,omitempty"`
	Complexity float64        `json:"-"` // filled by cortex
	Active     int            `json:"-"` // filled by cortex
}

// Glyph coarsens the spell's detail field into the dog's spoken encoding.
func (o Output) Glyph() encoding.Glyph { return o.Detail.Coarse() }

// Manifest describes one spell on disk.
type Manifest struct {
	Name        string   `json:"name"`
	School      string   `json:"school"`      // occipital | temporal | frontal
	Lang        string   `json:"lang"`        // go | python | c | ...
	Cmd         []string `json:"cmd"`         // argv, run from the project root
	Reads       string   `json:"reads"`       // text | image | signal
	Description string   `json:"description"` // one line
	dir         string   // resolved spell directory
}

// Spell is a discovered, castable spell.
type Spell struct {
	Manifest
	root string // project root, for resolving relative cmd paths
}

// Discover walks cortexRoot for manifest.json files and returns the spells it
// finds, sorted by name.
func Discover(cortexRoot string) ([]Spell, error) {
	var spells []Spell
	err := filepath.WalkDir(cortexRoot, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		if d.IsDir() || d.Name() != "manifest.json" {
			return nil
		}
		raw, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		var m Manifest
		if err := json.Unmarshal(raw, &m); err != nil {
			return fmt.Errorf("%s: %w", path, err)
		}
		m.dir = filepath.Dir(path)
		spells = append(spells, Spell{Manifest: m, root: filepath.Dir(cortexRoot)})
		return nil
	})
	if err != nil {
		return nil, err
	}
	sort.Slice(spells, func(i, j int) bool { return spells[i].Name < spells[j].Name })
	return spells, nil
}

// Cast runs the spell on the given input and returns its output. The spell is
// executed as a subprocess from the project root; stdin/stdout carry JSON.
func (s Spell) Cast(ctx context.Context, in Input) (Output, error) {
	if len(s.Cmd) == 0 {
		return Output{}, fmt.Errorf("spell %s: empty cmd", s.Name)
	}
	if ctx == nil {
		ctx = context.Background()
	}
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	payload, err := json.Marshal(in)
	if err != nil {
		return Output{}, err
	}
	// Resolve the spell executable against the augmented PATH before creating the
	// command. exec.Command resolves the name at construction time using the
	// parent process PATH; patching cmd.Env or cmd.Path later is not enough when
	// the parent PATH lacks the toolchain.
	env := augmentPath(os.Environ())
	cmd0 := s.Cmd[0]
	if !filepath.IsAbs(cmd0) && !strings.Contains(cmd0, string(filepath.Separator)) {
		for _, e := range env {
			if strings.HasPrefix(e, "PATH=") {
				_, pathVal, _ := strings.Cut(e, "=")
				if abs, err := lookInPath(cmd0, pathVal); err == nil {
					cmd0 = abs
				}
				break
			}
		}
	}
	cmd := exec.CommandContext(ctx, cmd0, s.Cmd[1:]...)
	cmd.Dir = s.root
	cmd.Stdin = bytes.NewReader(payload)
	var stdout, stderr bytes.Buffer
	cmd.Stdout = &stdout
	cmd.Stderr = &stderr
	// Spells may invoke sibling binaries (e.g. `go run`). Ensure the Go toolchain
	// is discoverable from standard locations even when the parent process was
	// launched with a minimal environment.
	cmd.Env = env
	if err := cmd.Run(); err != nil {
		return Output{}, fmt.Errorf("spell %s failed: %w: %s", s.Name, err, stderr.String())
	}
	var out Output
	if err := json.Unmarshal(stdout.Bytes(), &out); err != nil {
		return Output{}, fmt.Errorf("spell %s: bad output: %w", s.Name, err)
	}
	if out.Spell == "" {
		out.Spell = s.Name
	}
	out.Complexity = out.Detail.Complexity()
	out.Active = out.Detail.Active()
	return out, nil
}

// augmentPath prepends common Go toolchain locations to PATH when the Go
// binary is not already reachable, so spells can invoke `go run` from inside
// a subprocess launched by a compiled binary.
func augmentPath(env []string) []string {
	home := os.Getenv("HOME")
	for _, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			if _, rest, ok := strings.Cut(e, "="); ok {
				paths := strings.Split(rest, string(os.PathListSeparator))
				for _, p := range paths {
					if p == "" {
						continue
					}
					goBin := filepath.Join(p, "go")
					if runtime.GOOS == "windows" {
						goBin += ".exe"
					}
					if _, err := os.Stat(goBin); err == nil {
						return env
					}
				}
			}
			break
		}
	}
	candidates := []string{
		"/usr/local/go/bin",
		"/usr/lib/go/bin",
		filepath.Join(home, "go", "bin"),
		filepath.Join(home, ".local", "go", "bin"),
	}
	extra := ""
	for _, p := range candidates {
		if p == "" {
			continue
		}
		goBin := filepath.Join(p, "go")
		if runtime.GOOS == "windows" {
			goBin += ".exe"
		}
		if _, err := os.Stat(goBin); err == nil {
			if extra != "" {
				extra += string(os.PathListSeparator)
			}
			extra += p
		}
	}
	if extra == "" {
		return env
	}
	for i, e := range env {
		if strings.HasPrefix(e, "PATH=") {
			env[i] = "PATH=" + extra + string(os.PathListSeparator) + strings.TrimPrefix(e, "PATH=")
			return env
		}
	}
	return append(env, "PATH="+extra)
}

// lookInPath resolves name inside the provided PATH string (using the OS path
// separator). It is used because exec.LookPath reads the *parent* process PATH,
// not a custom cmd.Env.
func lookInPath(name, pathVal string) (string, error) {
	for _, p := range strings.Split(pathVal, string(os.PathListSeparator)) {
		if p == "" {
			continue
		}
		abs := filepath.Join(p, name)
		if runtime.GOOS == "windows" {
			abs += ".exe"
		}
		if info, err := os.Stat(abs); err == nil && !info.IsDir() {
			if info.Mode()&0o111 != 0 {
				return abs, nil
			}
		}
	}
	return "", fmt.Errorf("%s not found in provided PATH", name)
}