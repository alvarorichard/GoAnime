package diag

import (
	"bytes"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"runtime/debug"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/charmbracelet/colorprofile"
	"github.com/charmbracelet/x/term"
)

// Row is one fact; Section groups them under a title.
type Row struct{ Key, Value string }

type Section struct {
	Title string
	Rows  []Row
}

// Report is everything collected, ready to print.
type Report struct{ Sections []Section }

// Tool is an external program GoAnime depends on.
type Tool struct {
	Name        string
	VersionArgs []string // e.g. {"--version"}
}

// Options says what to collect. The zero value collects the machine, the
// terminal and the environment; callers add what only they know.
type Options struct {
	Version string
	Tools   []Tool
	Extra   []Section // appended as given (sources, cfsolve, paths)
	// Timeout bounds each external command. Default 3s.
	Timeout time.Duration
	// Getenv, Environ and Now exist for tests; they default to the os ones.
	Getenv  func(string) string
	Environ func() []string
	Now     func() time.Time
}

func (o *Options) defaults() {
	if o.Timeout <= 0 {
		o.Timeout = 3 * time.Second
	}
	if o.Getenv == nil {
		o.Getenv = os.Getenv
	}
	if o.Environ == nil {
		o.Environ = os.Environ
	}
	if o.Now == nil {
		o.Now = time.Now
	}
}

// Collect gathers the report. External commands run concurrently with a
// timeout each, so a hung tool costs Timeout, not the sum.
func Collect(ctx context.Context, o Options) Report {
	o.defaults()
	var r Report
	r.Sections = append(r.Sections, application(o), system(ctx, o), terminal(o), environment(o))
	if len(o.Tools) > 0 {
		r.Sections = append(r.Sections, tools(ctx, o))
	}
	r.Sections = append(r.Sections, o.Extra...)
	return r
}

func application(o Options) Section {
	s := Section{Title: "Application"}
	add := func(k, v string) {
		if v != "" {
			s.Rows = append(s.Rows, Row{k, v})
		}
	}
	add("Version", o.Version)
	add("Go", runtime.Version())
	if bi, ok := debug.ReadBuildInfo(); ok {
		var rev, vcsTime, modified string
		for _, kv := range bi.Settings {
			switch kv.Key {
			case "vcs.revision":
				rev = kv.Value
			case "vcs.time":
				vcsTime = kv.Value
			case "vcs.modified":
				modified = kv.Value
			}
		}
		if len(rev) > 12 {
			rev = rev[:12]
		}
		if modified == "true" {
			rev += " (modified)"
		}
		add("Commit", rev)
		add("Built", vcsTime)
	}
	if exe, err := os.Executable(); err == nil {
		add("Executable", exe)
	}
	add("PID", fmt.Sprint(os.Getpid()))
	return s
}

func system(ctx context.Context, o Options) Section {
	s := Section{Title: "System"}
	s.Rows = append(s.Rows, Row{"OS / arch", runtime.GOOS + " / " + runtime.GOARCH})
	s.Rows = append(s.Rows, platformInfo(ctx, o)...)
	s.Rows = append(s.Rows, Row{"CPUs", fmt.Sprint(runtime.NumCPU())})
	now := o.Now()
	name, offset := now.Zone()
	s.Rows = append(s.Rows, Row{"Time zone", fmt.Sprintf("%s (UTC%+03d:%02d)", name, offset/3600, abs(offset%3600)/60)})
	locale := firstNonEmpty(o.Getenv("LC_ALL"), o.Getenv("LC_MESSAGES"), o.Getenv("LANG"))
	if locale != "" {
		s.Rows = append(s.Rows, Row{"Locale", locale})
	}
	return s
}

func terminal(o Options) Section {
	s := Section{Title: "Terminal"}
	add := func(k, v string) {
		if v != "" {
			s.Rows = append(s.Rows, Row{k, v})
		}
	}
	add("TERM", o.Getenv("TERM"))
	add("COLORTERM", o.Getenv("COLORTERM"))
	prog := o.Getenv("TERM_PROGRAM")
	if v := o.Getenv("TERM_PROGRAM_VERSION"); prog != "" && v != "" {
		prog += " " + v
	}
	add("Program", prog)
	if o.Getenv("WT_SESSION") != "" {
		add("Windows Terminal", "yes")
	}
	if o.Getenv("TMUX") != "" {
		add("Multiplexer", "tmux")
	} else if o.Getenv("STY") != "" {
		add("Multiplexer", "screen")
	}
	if o.Getenv("SSH_CONNECTION") != "" || o.Getenv("SSH_TTY") != "" {
		add("Session", "ssh")
	}
	add("Shell", firstNonEmpty(o.Getenv("SHELL"), o.Getenv("ComSpec")))
	add("Colors", colorprofile.Detect(os.Stderr, o.Environ()).String())
	if o.Getenv("NO_COLOR") != "" {
		add("NO_COLOR", "set")
	}
	if w, h, err := term.GetSize(os.Stderr.Fd()); err == nil {
		add("Size", fmt.Sprintf("%d×%d", w, h))
	}
	add("TTY", fmt.Sprintf("stdin=%t stdout=%t stderr=%t",
		term.IsTerminal(os.Stdin.Fd()), term.IsTerminal(os.Stdout.Fd()), term.IsTerminal(os.Stderr.Fd())))
	return s
}

// proxyVars are the variables that route traffic elsewhere. Their values are
// shown with credentials removed: a proxy or a VPN is often the whole story.
var proxyVars = []string{"HTTP_PROXY", "HTTPS_PROXY", "ALL_PROXY", "NO_PROXY", "http_proxy", "https_proxy", "all_proxy", "no_proxy"}

func environment(o Options) Section {
	s := Section{Title: "Environment"}
	for _, kv := range o.Environ() {
		k, v, ok := strings.Cut(kv, "=")
		if ok && strings.HasPrefix(k, "GOANIME_") {
			s.Rows = append(s.Rows, Row{k, Redact(v)})
		}
	}
	sort.Slice(s.Rows, func(i, j int) bool { return s.Rows[i].Key < s.Rows[j].Key })
	for _, k := range proxyVars {
		if v := o.Getenv(k); v != "" {
			s.Rows = append(s.Rows, Row{k, Redact(v)})
		}
	}
	if len(s.Rows) == 0 {
		s.Rows = append(s.Rows, Row{"GOANIME_*", "none set"})
	}
	return s
}

func tools(ctx context.Context, o Options) Section {
	s := Section{Title: "Tools"}
	rows := make([]Row, len(o.Tools))
	var wg sync.WaitGroup
	for i, t := range o.Tools {
		wg.Add(1)
		go func() {
			defer wg.Done()
			rows[i] = Row{t.Name, probeTool(ctx, t, o.Timeout)}
		}()
	}
	wg.Wait()
	s.Rows = rows
	return s
}

// probeTool finds t and asks for its version, never waiting past timeout.
func probeTool(ctx context.Context, t Tool, timeout time.Duration) string {
	path := findTool(t.Name)
	if path == "" {
		return "not found"
	}
	if len(t.VersionArgs) == 0 {
		return path
	}
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()
	cmd := exec.CommandContext(ctx, path, t.VersionArgs...) // #nosec G204 -- a fixed tool name and version flag the caller chose, no shell
	var out bytes.Buffer
	cmd.Stdout, cmd.Stderr = &limitedWriter{w: &out, n: 4096}, nil
	// Killing the tool at the timeout is not enough when it left a child
	// holding the output pipe (a shell wrapper around the real binary): Run
	// would wait for that child. WaitDelay closes the pipe shortly after.
	cmd.WaitDelay = 200 * time.Millisecond
	err := cmd.Run()
	if ctx.Err() != nil {
		return path + " (version check timed out)"
	}
	line, _, _ := strings.Cut(strings.TrimSpace(out.String()), "\n")
	line = strings.TrimSpace(line)
	if err != nil && line == "" {
		return path + " (version check failed)"
	}
	return line + "  [" + path + "]"
}

// findTool looks on PATH, then beside the executable (the Windows installer
// puts mpv in bin\ there).
func findTool(name string) string {
	if p, err := exec.LookPath(name); err == nil {
		return p
	}
	exe, err := os.Executable()
	if err != nil {
		return ""
	}
	dir := filepath.Dir(exe)
	file := name
	if runtime.GOOS == "windows" {
		file += ".exe"
	}
	for _, c := range []string{filepath.Join(dir, "bin", file), filepath.Join(dir, file)} {
		if st, err := os.Stat(c); err == nil && !st.IsDir() {
			return c
		}
	}
	return ""
}

// limitedWriter keeps the first n bytes and drops the rest without erroring,
// so a chatty --version cannot fill memory or fail the command.
type limitedWriter struct {
	w *bytes.Buffer
	n int
}

func (l *limitedWriter) Write(p []byte) (int, error) {
	if room := l.n - l.w.Len(); room > 0 {
		l.w.Write(p[:min(len(p), room)])
	}
	return len(p), nil
}

// Text renders the report as aligned "key  value" blocks under their titles,
// with secrets and the home directory removed.
func (r Report) Text() string {
	var b strings.Builder
	for i, s := range r.Sections {
		if i > 0 {
			b.WriteByte('\n')
		}
		fmt.Fprintf(&b, "%s\n", s.Title)
		width := 0
		for _, row := range s.Rows {
			width = max(width, len(row.Key))
		}
		for _, row := range s.Rows {
			fmt.Fprintf(&b, "  %-*s  %s\n", width, row.Key, Redact(row.Value))
		}
	}
	return b.String()
}

func firstNonEmpty(vs ...string) string {
	for _, v := range vs {
		if v != "" {
			return v
		}
	}
	return ""
}

func abs(n int) int {
	if n < 0 {
		return -n
	}
	return n
}
