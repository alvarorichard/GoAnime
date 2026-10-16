package tui

import (
	"os"
	"strconv"
	"strings"
	"time"

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

	fd := int(os.Stdin.Fd())
	state, err := term.MakeRaw(fd)
	if err != nil {
		DrainTerminalResponses(20 * time.Millisecond)
		return false, false
	}
	defer func() { _ = term.Restore(fd, state) }()

	reply, ok := readOSCReply(fd, timeout)
	// Delayed bytes must never leak into the next prompt as literal text.
	DrainTerminalResponses(30 * time.Millisecond)
	if !ok {
		return false, false
	}
	return parseBackgroundIsLight(reply)
}

// readOSCReply reads until BEL or ST (ESC \) or the timeout. The OSC 11
// reply looks like "]11;rgb:0000/0000/0000" wrapped in ESC ] ... (BEL|ST).
func readOSCReply(fd int, timeout time.Duration) (string, bool) {
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
			n, err := os.Stdin.Read(buf)
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

	select {
	case r := <-ch:
		return r.data, r.ok
	case <-time.After(timeout):
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
