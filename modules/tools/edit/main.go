// Spell: edit (wild school)
//
// Gives Argos file modification powers.
// Reads, writes, or patches any file on the filesystem.
//
// Input:
//   modules.Input{
//     Path: "/path/to/file",
//     Args: {
//       "action":     "read" | "write" | "patch",
//       "content":    "<full content for write>",
//       "old_string": "<text to find for patch>",
//       "new_string": "<replacement text for patch>",
//     },
//   }
//
// Output: modules.Output{Summary: "description of what was done"}
package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"glyphai/internal/modules"
)

func main() {
	var in modules.Input
	if err := json.NewDecoder(os.Stdin).Decode(&in); err != nil {
		fail(fmt.Errorf("reading input: %v", err))
	}

	path := in.Path
	if path == "" {
		fail(fmt.Errorf("path is required"))
	}

	// Expand ~ to home dir
	if strings.HasPrefix(path, "~/") {
		home, err := os.UserHomeDir()
		if err != nil {
			fail(fmt.Errorf("cannot resolve home dir: %v", err))
		}
		path = filepath.Join(home, path[2:])
	}

	action := ""
	if v, ok := in.Args["action"].(string); ok {
		action = v
	}
	if action == "" {
		// default to read
		action = "read"
	}

	var summary string

	switch action {
	case "read":
		data, err := os.ReadFile(path)
		if err != nil {
			fail(fmt.Errorf("read %s: %v", path, err))
		}
		summary = string(data)

	case "write":
		content := ""
		if v, ok := in.Args["content"].(string); ok {
			content = v
		}
		// Ensure parent directory exists
		if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil {
			fail(fmt.Errorf("mkdir %s: %v", filepath.Dir(path), err))
		}
		if err := os.WriteFile(path, []byte(content), 0644); err != nil {
			fail(fmt.Errorf("write %s: %v", path, err))
		}
		summary = fmt.Sprintf("wrote %d bytes to %s", len(content), path)

	case "patch":
		oldStr := ""
		newStr := ""
		if v, ok := in.Args["old_string"].(string); ok {
			oldStr = v
		}
		if v, ok := in.Args["new_string"].(string); ok {
			newStr = v
		}
		if oldStr == "" {
			fail(fmt.Errorf("old_string is required for patch action"))
		}
		data, err := os.ReadFile(path)
		if err != nil {
			fail(fmt.Errorf("read %s for patch: %v", path, err))
		}
		original := string(data)
		if !strings.Contains(original, oldStr) {
			fail(fmt.Errorf("old_string not found in %s", path))
		}
		patched := strings.Replace(original, oldStr, newStr, 1)
		if err := os.WriteFile(path, []byte(patched), 0644); err != nil {
			fail(fmt.Errorf("write patched %s: %v", path, err))
		}
		summary = fmt.Sprintf("patched %s: replaced %d chars with %d chars", path, len(oldStr), len(newStr))

	default:
		fail(fmt.Errorf("unknown action %q — use read, write, or patch", action))
	}

	out := modules.Output{
		Spell:   "edit",
		Summary: summary,
	}
	if err := json.NewEncoder(os.Stdout).Encode(out); err != nil {
		fail(fmt.Errorf("encoding output: %v", err))
	}
}

func fail(err error) {
	out := modules.Output{
		Spell:   "edit",
		Summary: fmt.Sprintf("error: %v", err),
	}
	json.NewEncoder(os.Stdout).Encode(out) //nolint
	os.Exit(1)
}