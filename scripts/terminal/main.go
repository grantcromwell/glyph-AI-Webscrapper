// kennel — a full-screen terminal UI for living with the dog. Pure stdlib (no
// Bubble Tea, no deps): the alternate screen, raw input via stty, a framed
// scrollback pane, a status bar, and an input line. Type a question and the dog
// answers; use a slash command to send it hunting. The dog's output streams into
// the pane live.
//
//	just type…          ask the dog a question
//	/hunt <theme>       prowl the live web toward a theme
//	/roam [min]         loose the dog, self-guided, no topic (fast eyes)
//	/wild [min]         loose it with the full headless browser (Chromium)
//	/cast <spell> [q]   cast a grimoire spell (russian, spectrum, eigen, phoible…)
//	/look <path|url>    let the dog see an image
//	/recall <query>     what it remembers, by resonance
//	/sense <text>       the glyph a phrase lights
//	/help   /quit  (or Esc)
package main

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
	"time"
)

type app struct {
	root, bin    string
	rows, cols   int
	scroll       []string // wrapped output lines
	input        string
	status       string
	ansi         *regexp.Regexp
}

func main() {
	root := repoRoot()
	if root == "" {
		fmt.Println("kennel: run me from inside the glyphai repo.")
		os.Exit(1)
	}
	a := &app{
		root:   root,
		bin:    filepath.Join(root, "bin", "cerebrum"),
		status: "building the dog…",
		ansi:   regexp.MustCompile(`\x1b\[[0-9;?]*[a-zA-Z]`),
	}
	a.rows, a.cols = termSize()

	rawOn()
	defer rawOff()
	fmt.Print("\x1b[?1049h\x1b[?25l") // alt screen, hide cursor
	defer fmt.Print("\x1b[?25h\x1b[?1049l")

	if err := a.runCmd("go", "build", "-o", a.bin, "./cmd/cerebrum"); err != nil {
		a.status = "build failed: " + err.Error()
	}
	a.println(" glyphai · terminal — talk to the dog, or send it hunting.")
	a.helpLines()
	a.refreshStatus()
	a.render()

	in := bufio.NewReader(os.Stdin)
	for {
		b, err := in.ReadByte()
		if err != nil {
			break
		}
		switch b {
		case '\r', '\n':
			line := strings.TrimSpace(a.input)
			a.input = ""
			if line == "/quit" || line == "/q" || line == ":q" {
				return
			}
			if line != "" {
				a.println("\x1b[36myou ❯\x1b[0m " + line)
				a.dispatch(line)
				a.refreshStatus()
			}
			a.rows, a.cols = termSize() // pick up any resize, once per command
			a.render()
		case 127, 8: // backspace
			if r := []rune(a.input); len(r) > 0 {
				a.input = string(r[:len(r)-1])
				a.render()
			}
		case 27: // Esc
			return
		case 3: // Ctrl-C
			return
		default:
			if b >= 32 { // printable
				a.input += string(b)
				a.render()
			}
		}
	}
}

func (a *app) dispatch(line string) {
	if !strings.HasPrefix(line, "/") {
		a.run(a.bin, "ask", line)
		return
	}
	fields := strings.Fields(line)
	verb := fields[0]
	rest := strings.TrimSpace(strings.TrimPrefix(line, verb))
	switch verb {
	case "/help", "/h":
		a.helpLines()
	case "/hunt":
		if rest == "" {
			a.println("   usage: /hunt <theme>")
			return
		}
		a.run(a.bin, "prowl", rest)
	case "/roam", "/wild":
		mins, eyes := 10, "skim"
		if verb == "/wild" {
			eyes = "browse"
		}
		if rest != "" {
			if m, err := strconv.Atoi(strings.Fields(rest)[0]); err == nil && m > 0 {
				mins = m
			}
		}
		until := time.Now().Add(time.Duration(mins) * time.Minute).Unix()
		a.println(fmt.Sprintf("   🐾 loosing the dog for %d min (eyes=%s, self-guided)…", mins, eyes))
		a.run(a.bin, "roam", "-eyes", eyes, "-until", strconv.FormatInt(until, 10))
	case "/cast":
		if rest == "" {
			a.println("   usage: /cast <spell> [query]")
			return
		}
		a.run(a.bin, append([]string{"cast"}, strings.Fields(rest)...)...)
	case "/look":
		if rest == "" {
			a.println("   usage: /look <path|url>")
			return
		}
		a.run(a.bin, "look", rest)
	case "/recall":
		a.run(a.bin, "recall", rest)
	case "/sense":
		a.run(a.bin, "sense", rest)
	default:
		a.println("   unknown command " + verb + " — try /help")
	}
}

// run executes the dog and streams its (ANSI-stripped) output into the pane live.
func (a *app) run(name string, args ...string) {
	c := exec.Command(name, args...)
	c.Dir = a.root
	c.Env = append(os.Environ(), "PATH="+goBin()+":"+os.Getenv("PATH"))
	pr, pw := io.Pipe()
	c.Stdout, c.Stderr = pw, pw
	done := make(chan struct{})
	go func() { _ = c.Run(); pw.Close(); close(done) }()
	sc := bufio.NewScanner(pr)
	sc.Buffer(make([]byte, 0, 64*1024), 1<<20)
	for sc.Scan() {
		a.println(a.ansi.ReplaceAllString(sc.Text(), ""))
	}
	<-done
}

func (a *app) runCmd(name string, args ...string) error {
	c := exec.Command(name, args...)
	c.Dir = a.root
	c.Env = append(os.Environ(), "PATH="+goBin()+":"+os.Getenv("PATH"))
	return c.Run()
}

func (a *app) println(s string) {
	for _, ln := range strings.Split(s, "\n") {
		a.scroll = append(a.scroll, wrap(ln, a.cols-2)...)
	}
	if n := len(a.scroll); n > 4000 {
		a.scroll = a.scroll[n-4000:]
	}
	a.render()
}

func (a *app) helpLines() {
	for _, l := range []string{
		"   just type…        ask the dog a question",
		"   /hunt <theme>     prowl the web toward a theme",
		"   /roam [min]       loose the dog, self-guided (fast eyes)",
		"   /wild [min]       loose it with the full headless browser",
		"   /cast <spell> [q] cast a grimoire spell (russian, spectrum, eigen…)",
		"   /look <path|url>  let the dog see an image",
		"   /recall <query>   what it remembers, by resonance",
		"   /sense <text>     the glyph a phrase lights · /help · /quit",
	} {
		a.scroll = append(a.scroll, l)
	}
}

func (a *app) refreshStatus() {
	out, err := exec.Command(a.bin, "coverage").Output()
	if err == nil {
		a.status = a.ansi.ReplaceAllString(strings.TrimSpace(string(out)), "")
	} else {
		a.status = "the dog waits"
	}
}

// render paints the whole screen: title, scrollback pane, status, input.
func (a *app) render() {
	paneH := a.rows - 3
	if paneH < 1 {
		paneH = 1
	}
	var b strings.Builder
	b.WriteString("\x1b[H")
	// title bar
	title := "  glyphai · terminal — ask, or send the dog hunting "
	b.WriteString("\x1b[7m" + pad(title, a.cols) + "\x1b[0m\r\n")
	// pane: last paneH lines
	start := len(a.scroll) - paneH
	if start < 0 {
		start = 0
	}
	for i := 0; i < paneH; i++ {
		if start+i < len(a.scroll) {
			b.WriteString(clip(a.scroll[start+i], a.cols))
		}
		b.WriteString("\x1b[K\r\n")
	}
	// status
	b.WriteString("\x1b[2m" + clip(pad(" "+a.status, a.cols), a.cols) + "\x1b[0m\r\n")
	// input
	b.WriteString("\x1b[36myou ❯\x1b[0m " + a.input + "\x1b[7m \x1b[0m\x1b[K")
	fmt.Print(b.String())
}

// --- terminal helpers (pure stdlib via stty) ---

func termSize() (rows, cols int) {
	rows, cols = 30, 100
	out, err := exec.Command("stty", "size").Output()
	if err == nil {
		if n, _ := fmt.Sscan(string(out), &rows, &cols); n < 2 {
			rows, cols = 30, 100
		}
	}
	return
}

func rawOn()  { sttyApply("-echo", "-icanon", "min", "1", "time", "0") }
func rawOff() { sttyApply("sane") }

func sttyApply(args ...string) {
	c := exec.Command("stty", args...)
	c.Stdin = os.Stdin
	_ = c.Run()
}

func wrap(s string, w int) []string {
	if w < 1 {
		w = 1
	}
	r := []rune(s)
	if len(r) == 0 {
		return []string{""}
	}
	var out []string
	for len(r) > w {
		out = append(out, string(r[:w]))
		r = r[w:]
	}
	return append(out, string(r))
}

func clip(s string, w int) string {
	r := []rune(s)
	if len(r) <= w {
		return s
	}
	return string(r[:w])
}

func pad(s string, w int) string {
	r := []rune(s)
	if len(r) >= w {
		return string(r[:w])
	}
	return s + strings.Repeat(" ", w-len(r))
}

func goBin() string {
	if home, err := os.UserHomeDir(); err == nil {
		return filepath.Join(home, ".local", "go", "bin")
	}
	return "/usr/local/go/bin"
}

func repoRoot() string {
	dir, _ := os.Getwd()
	for {
		if b, err := os.ReadFile(filepath.Join(dir, "go.mod")); err == nil && strings.Contains(string(b), "module glyphai") {
			return dir
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return ""
		}
		dir = parent
	}
}
