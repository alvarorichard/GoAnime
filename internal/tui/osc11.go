package tui

import (
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/muesli/cancelreader"
	"golang.org/x/term"
)

// osc11Timeout bounds the single terminal background query. Most terminals
// answer in a few milliseconds; a silent one costs one short wait per
// process, not per screen.
const osc11Timeout = 120 * time.Millisecond

// queryBackgroundIsLight sends OSC 11 (request background color) and parses
// the reply. ok=false means the terminal did not answer in time.
func queryBackgroundIsLight(timeout time.Duration) (light, ok bool) {
	stdout := os.Stdout
	if stdout == nil {
		return false, false
	}
	if _, err := stdout.WriteString("\x1b]11;?\x1b\\"); err != nil {
		return false, false
	}

	if runtime.GOOS == "windows" {
		// A Windows console funnels the terminal's answer and the user's
		// keystrokes through one input queue: the blocking reader below
		// would steal whatever is typed first, and on a silent terminal it
		// outlives its own timeout doing so. Probe via peek/consume/restore
		// instead, which never loses input.
		return queryBackgroundWindows(timeout)
	}

	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		DrainTerminalResponses(20 * time.Millisecond)
		return false, false
	}
	defer func() { _ = term.Restore(fd, state) }()

	in, err := newReplyReader()
	if err != nil {
		// No way to stop a read once started: asking would leave a reader
		// behind that takes the user's keys, so the theme is not asked for.
		return false, false
	}
	defer func() { _ = in.Close() }()
	reply, ok := readOSCReply(in, timeout)
	// Delayed bytes must never leak into the next prompt as literal text.
	DrainTerminalResponses(30 * time.Millisecond)
	if !ok {
		return false, false
	}
	return parseBackgroundIsLight(reply)
}

// replyReader is the terminal's input, read so that a read under way can be
// called off.
type replyReader interface {
	io.Reader
	Cancel() bool
	Close() error
}

// newReplyReader opens stdin for the reply; a seam for the tests.
var newReplyReader = func() (replyReader, error) { return cancelreader.NewReader(os.Stdin) }

// osc11Drain bounds how long a cancelled read may take to return.
const osc11Drain = time.Second

// readOSCReply reads until BEL or ST (ESC \) or the timeout. The OSC 11
// reply looks like "]11;rgb:0000/0000/0000" wrapped in ESC ] ... (BEL|ST).
//
// When the terminal does not answer, the read under way is cancelled and
// waited for before this returns. A read left running would take the next
// keys the user types (up to 64 bytes) from whatever screen comes next: in a
// terminal that never answers (tmux, the Linux console, many SSH sessions)
// the first menu would look frozen.
func readOSCReply(in replyReader, timeout time.Duration) (string, bool) {
	type result struct {
		data string
		ok   bool
	}
	ch := make(chan result, 1)
	go func() {
		var sb strings.Builder
		buf := make([]byte, 1)
		prevESC := false
		for sb.Len() < 64 {
			n, err := in.Read(buf)
			if n <= 0 || err != nil {
				break
			}
			c := buf[0]
			if c == '\a' { // BEL ends the reply
				ch <- result{sb.String(), true}
				return
			}
			if prevESC && c == '\\' { // ST ends the reply
				s := sb.String()
				ch <- result{strings.TrimSuffix(s, "\x1b"), true}
				return
			}
			prevESC = c == '\x1b'
			sb.WriteByte(c)
		}
		ch <- result{sb.String(), sb.Len() > 0}
	}()

	timer := time.NewTimer(timeout)
	defer timer.Stop()
	select {
	case r := <-ch:
		return r.data, r.ok
	case <-timer.C:
		in.Cancel()
		select {
		case <-ch: // the read returned: nothing is left reading the terminal
		case <-time.After(osc11Drain):
		}
		return "", false
	}
}

// parseBackgroundIsLight interprets an OSC 11 body such as
// "]11;rgb:ffff/ffff/ffff" or "11;rgb:ff/ff/ff".
func parseBackgroundIsLight(reply string) (light, ok bool) {
	lower := strings.ToLower(reply)
	idx := strings.Index(lower, "rgb:")
	if idx < 0 {
		return false, false
	}
	parts := strings.Split(lower[idx+4:], "/")
	if len(parts) != 3 {
		return false, false
	}
	var sum float64
	for i, part := range parts {
		part = strings.TrimRight(part, "\x1b\\\a")
		if part == "" || len(part) > 4 {
			return false, false
		}
		// Scale 1–4 hex digits to 0–1.
		v, err := strconv.ParseUint(part, 16, 16)
		if err != nil {
			return false, false
		}
		maxVal := float64(uint64(1)<<(uint(len(part))*4) - 1)
		f := float64(v) / maxVal
		switch i {
		case 0:
			sum += 0.2126 * f
		case 1:
			sum += 0.7152 * f
		case 2:
			sum += 0.0722 * f
		}
	}
	return sum > 0.5, true
}
