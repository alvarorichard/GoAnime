package logview

import (
	"bytes"
	"context"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/charmbracelet/colorprofile"
)

const usage = `Usage: goanime logs [-f] [-n N] [file]

Shows the newest debug log (written by goanime --debug), in colour when the
terminal has it.

  -f        keep following the log as it grows, and move to the new session
            when goanime --debug is started again
  -n N      show only the last N lines first (with -f the default is 20)
  file      a log file name from the logs folder, or a path; default: the newest
`

const (
	// followHistory is how many lines -f shows before it starts following.
	followHistory = 20
	// maxHistory bounds how much of a file is read to show its tail.
	maxHistory = 8 << 20
	// maxPerPoll bounds what one poll reads while following.
	maxPerPoll = 4 << 20
)

// Variables so that tests can speed the loop up and point it at a folder.
var (
	pollEvery = 200 * time.Millisecond
	logDirFn  = util.GetLogDir
)

// UseColor reports whether f can show the interface's colours: not in plain
// mode, not with NO_COLOR, not into a pipe.
func UseColor(f *os.File) bool {
	return !tui.Plain() && tui.ConsoleColorProfile(f) > colorprofile.ASCII
}

// Main is Run on the process's standard streams.
func Main(ctx context.Context, args []string) int {
	color := UseColor(os.Stdout)
	var out io.Writer = os.Stdout
	if color {
		out = colorprofile.NewWriter(os.Stdout, os.Environ()) // downsample to the terminal
	}
	return Run(ctx, args, out, os.Stderr, color)
}

type options struct {
	follow bool
	lines  int // -1: not given
	file   string
}

func parseArgs(args []string, errw io.Writer) (options, bool) {
	fs := flag.NewFlagSet("goanime logs", flag.ContinueOnError)
	fs.SetOutput(io.Discard)
	opt := options{lines: -1}
	fs.BoolVar(&opt.follow, "f", false, "")
	fs.BoolVar(&opt.follow, "follow", false, "")
	fs.IntVar(&opt.lines, "n", -1, "")
	// The flag package stops at the first argument that is not a flag; parse
	// again after the file name so "logs file -f" works as well as "logs -f file".
	rest := args
	for {
		if err := fs.Parse(rest); err != nil {
			if errors.Is(err, flag.ErrHelp) {
				_, _ = fmt.Fprint(errw, usage)
				return opt, false
			}
			_, _ = fmt.Fprintf(errw, "goanime logs: %v\n\n%s", err, usage)
			return opt, false
		}
		if fs.NArg() == 0 {
			break
		}
		if opt.file != "" {
			_, _ = fmt.Fprintf(errw, "goanime logs: more than one file given\n\n%s", usage)
			return opt, false
		}
		opt.file = fs.Arg(0)
		rest = fs.Args()[1:]
	}
	if opt.lines < -1 {
		_, _ = fmt.Fprintf(errw, "goanime logs: -n must not be negative\n\n%s", usage)
		return opt, false
	}
	return opt, true
}

// Run shows a log. It returns the process's exit code: 0, 1 when there is
// nothing to show or the file is missing, 2 for a usage error.
func Run(ctx context.Context, args []string, out, errw io.Writer, color bool) int {
	opt, ok := parseArgs(args, errw)
	if !ok {
		return 2
	}
	theme := tui.CurrentTheme()
	v := &viewer{out: out, errw: errw, color: color, col: NewColorizer(&theme), theme: &theme}

	path, err := resolve(opt.file)
	switch {
	case err == nil:
	case errors.Is(err, errNoLogs) && opt.follow && opt.file == "":
		// Following before the first debug session: wait for it.
		_, _ = fmt.Fprintln(errw, "No debug log yet. Waiting for one: run goanime --debug in another terminal.")
	default:
		_, _ = fmt.Fprintf(errw, "goanime logs: %v\n", err)
		return 1
	}

	lines := opt.lines
	if lines < 0 && opt.follow {
		lines = followHistory
	}
	var offset int64
	if path != "" {
		offset, err = v.showTail(path, lines)
		if err != nil {
			_, _ = fmt.Fprintf(errw, "goanime logs: %v\n", err)
			return 1
		}
	}
	if !opt.follow {
		return 0
	}
	v.follow(ctx, path, offset, opt.file == "")
	return 0
}

type viewer struct {
	out, errw io.Writer
	color     bool
	col       *Colorizer
	theme     *tui.Theme
}

func (v *viewer) emit(line string) {
	if v.color {
		line = v.col.Line(line)
	}
	_, _ = fmt.Fprintln(v.out, line)
}

// note prints a separator between sessions, set apart from the log lines.
func (v *viewer) note(text string) {
	if v.color {
		text = render(&v.theme.Faint, text)
	}
	_, _ = fmt.Fprintln(v.out, text)
}

// showTail prints the last n lines of path (all of them when n < 0) and returns
// the offset the next byte to read starts at. Every line goes through the
// colorizer so it knows whether it is inside the environment report; only the
// last n are shown. A file larger than maxHistory is read from the end.
func (v *viewer) showTail(path string, n int) (int64, error) {
	f, err := os.Open(path) // #nosec G304 -- a log the user named or the newest in the log folder
	if err != nil {
		return 0, err
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil {
		return 0, err
	}
	start := int64(0)
	if st.Size() > maxHistory {
		start = st.Size() - maxHistory
		if _, err := f.Seek(start, io.SeekStart); err != nil {
			return 0, err
		}
	}
	data, err := io.ReadAll(io.LimitReader(f, maxHistory))
	if err != nil {
		return 0, err
	}
	base := start
	if start > 0 { // began mid-line: drop the fragment, and count it in the offset
		if i := bytes.IndexByte(data, '\n'); i >= 0 {
			data = data[i+1:]
			base += int64(i + 1)
		}
	}
	v.col.Reset()
	lines := splitLines(data)
	first := 0
	if n >= 0 && len(lines) > n {
		first = len(lines) - n
	}
	for i, line := range lines {
		if i < first {
			if v.color {
				v.col.Line(line) // for its state only
			}
			continue
		}
		v.emit(line)
	}
	// Only whole lines were shown: a last line without its newline is still
	// being written and is left for the follower.
	return base + int64(consumed(data)), nil
}

// splitLines returns the complete lines of data, without "\r\n" or "\n".
func splitLines(data []byte) []string {
	var out []string
	for len(data) > 0 {
		i := bytes.IndexByte(data, '\n')
		if i < 0 {
			break
		}
		out = append(out, strings.TrimSuffix(string(data[:i]), "\r"))
		data = data[i+1:]
	}
	return out
}

// consumed is the number of bytes of data that splitLines turns into lines.
func consumed(data []byte) int {
	if i := bytes.LastIndexByte(data, '\n'); i >= 0 {
		return i + 1
	}
	return 0
}

// follow prints what is appended to path. With auto it also moves to a newer
// session log when one appears, which is the usual way to use it: leave
// "goanime logs -f" open in one terminal and run "goanime --debug" in another.
func (v *viewer) follow(ctx context.Context, path string, offset int64, auto bool) {
	var partial []byte
	for {
		if auto {
			if newest, err := resolve(""); err == nil && newest != path {
				if path != "" {
					v.note("── new session: " + filepath.Base(newest) + " ──")
				}
				path, offset, partial = newest, 0, nil
				v.col.Reset()
			}
		}
		if path != "" {
			partial = v.drain(path, &offset, partial)
		}
		select {
		case <-ctx.Done():
			return
		case <-time.After(pollEvery):
		}
	}
}

// drain prints the complete lines appended to path since offset and returns the
// unfinished last line, to be completed by the next poll.
func (v *viewer) drain(path string, offset *int64, partial []byte) []byte {
	f, err := os.Open(path) // #nosec G304 -- see showTail
	if err != nil {
		return partial // gone (pruned) or not there yet: the next poll looks again
	}
	defer func() { _ = f.Close() }()
	st, err := f.Stat()
	if err != nil {
		return partial
	}
	if st.Size() < *offset { // truncated or replaced
		v.note("── log was reset ──")
		*offset, partial = 0, nil
		v.col.Reset()
	}
	if st.Size() == *offset {
		return partial
	}
	if _, err := f.Seek(*offset, io.SeekStart); err != nil {
		return partial
	}
	data, err := io.ReadAll(io.LimitReader(f, maxPerPoll))
	if err != nil || len(data) == 0 {
		return partial
	}
	*offset += int64(len(data))
	data = append(partial, data...)
	for _, line := range splitLines(data) {
		v.emit(line)
	}
	return append([]byte(nil), data[consumed(data):]...)
}

var errNoLogs = errors.New("no debug log yet: run goanime --debug first")

// resolve finds the log to show: the named one, or the newest session log.
func resolve(name string) (string, error) {
	dir := logDirFn()
	if name == "" {
		return newest(dir)
	}
	for _, candidate := range []string{name, filepath.Join(dir, name)} {
		if st, err := os.Stat(candidate); err == nil && !st.IsDir() {
			return candidate, nil
		}
	}
	return "", fmt.Errorf("no such log: %s (looked in %s)", name, dir)
}

// newest is the latest session log in dir. The names carry their timestamp, so
// the last name is the latest run. Crash reports are not session logs.
func newest(dir string) (string, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		return "", errNoLogs
	}
	var names []string
	for _, e := range entries {
		if !e.IsDir() && strings.HasPrefix(e.Name(), "goanime_") && strings.HasSuffix(e.Name(), ".log") {
			names = append(names, e.Name())
		}
	}
	if len(names) == 0 {
		return "", errNoLogs
	}
	sort.Strings(names)
	return filepath.Join(dir, names[len(names)-1]), nil
}
