// Package logview shows GoAnime's debug logs in the terminal: the newest one,
// or a followed one (goanime logs -f), in the colours of the interface.
//
// The log file itself holds plain text so that editors, less and bug reports
// read it as it is. The colour is added here, at display, the way docker logs
// and journalctl do. Colouring never changes a character of the line: take the
// escape sequences away and the original is what is left.
package logview

import (
	"regexp"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
)

// problemResults are the outcomes of a provider check that mean something is
// wrong; "ok" is left quiet so the problems are what the eye finds.
var problemResults = map[string]bool{
	"challenge": true, "blocked": true, "down": true, "slow": true,
	"dns": true, "refused": true, "tls": true, "error": true,
}

var (
	// 19:13:38 INFO  message — the label is only on INFO, WARN, ERROR and FATAL.
	logLineRe = regexp.MustCompile(`^(\d{2}:\d{2}:\d{2}(?:\.\d+)?)( (?:INFO |WARN |ERROR|FATAL))? (.*)$`)
	// key=value and key="quoted value", only where a key starts a word.
	fieldRe = regexp.MustCompile(`([A-Za-z_][A-Za-z0-9_.-]*)=("(?:[^"\\]|\\.)*"|\S*)`)
	// [PERF] and [TRACE] markers.
	tagRe = regexp.MustCompile(`\[(?:PERF|TRACE)\]`)
	// "  Key   value" in the header and in the environment report.
	reportRowRe = regexp.MustCompile(`^(\s*)([A-Za-z][A-Za-z0-9_*/ ().-]*?)(\s{2,})(\S.*)$`)
)

// Colorizer styles log lines one at a time. It remembers whether it is inside
// the environment report, whose lines read differently from log lines.
type Colorizer struct {
	theme *tui.Theme
	inEnv bool
}

// NewColorizer styles with theme.
func NewColorizer(theme *tui.Theme) *Colorizer { return &Colorizer{theme: theme} }

// Reset forgets the report state, for the start of another file.
func (c *Colorizer) Reset() { c.inEnv = false }

// Line returns line styled. Without its escape sequences it is line, exactly.
func (c *Colorizer) Line(line string) string {
	t := c.theme
	switch {
	case line == "":
		return ""
	case strings.HasPrefix(line, "====="):
		return render(&t.Primary, line)
	case line == "--- Environment ---":
		c.inEnv = true
		return render(&t.Faint, line)
	case line == "--- Log ---":
		c.inEnv = false
		return render(&t.Faint, line)
	}
	if c.inEnv {
		return c.reportLine(line)
	}
	if m := logLineRe.FindStringSubmatch(line); m != nil {
		return c.logLine(m[1], m[2], m[3])
	}
	// The header block before the log ("Session  abc", "Started  ...").
	if row := c.reportRow(line); row != "" {
		return row
	}
	return render(&t.Soft, line)
}

// reportLine styles a line of the environment report: a title, or a row.
func (c *Colorizer) reportLine(line string) string {
	if !strings.HasPrefix(line, " ") {
		return render(&c.theme.Primary, line) // "System", "Terminal", ...
	}
	if row := c.reportRow(line); row != "" {
		return row
	}
	return render(&c.theme.Text, line)
}

func (c *Colorizer) reportRow(line string) string {
	m := reportRowRe.FindStringSubmatch(line)
	if m == nil {
		return ""
	}
	return m[1] + render(&c.theme.Muted, m[2]) + m[3] + render(&c.theme.Text, m[4])
}

// logLine styles "time [LEVEL] message".
func (c *Colorizer) logLine(clock, label, msg string) string {
	t := c.theme
	var b strings.Builder
	b.WriteString(render(&t.Muted, clock))
	if label != "" {
		b.WriteString(" ")
		b.WriteString(c.level(strings.TrimPrefix(label, " ")))
	}
	b.WriteString(" ")
	b.WriteString(c.message(msg))
	return b.String()
}

// level styles INFO, WARN and ERROR the way the console does: the accent, an
// inverted badge, the crimson block. The padding of those styles is removed
// because it would add characters to the line.
func (c *Colorizer) level(label string) string {
	t := c.theme
	switch strings.TrimSpace(label) {
	case "INFO":
		return render(&t.Primary, label)
	case "WARN":
		badge := t.Tag.Padding(0, 0)
		return render(&badge, label)
	default: // ERROR, FATAL
		block := t.Notice.Padding(0, 0)
		return render(&block, label)
	}
}

// message styles the text of a log line: plain words, [PERF] markers, and
// key=value fields with the key and the value told apart.
func (c *Colorizer) message(msg string) string {
	t := c.theme
	var b strings.Builder
	last := 0
	for _, loc := range fieldLocations(msg) {
		start, end, kEnd := loc[0], loc[1], loc[3]
		b.WriteString(c.words(msg[last:start]))
		key := msg[start:kEnd]
		value := msg[loc[4]:end]
		b.WriteString(render(&t.Muted, key))
		b.WriteString(render(&t.Faint, "="))
		b.WriteString(c.value(key, value))
		last = end
	}
	b.WriteString(c.words(msg[last:]))
	return b.String()
}

// fieldLocations are the key=value matches that start a word.
func fieldLocations(msg string) [][]int {
	var out [][]int
	for _, loc := range fieldRe.FindAllStringSubmatchIndex(msg, -1) {
		if loc[0] == 0 || msg[loc[0]-1] == ' ' || msg[loc[0]-1] == '(' {
			out = append(out, loc)
		}
	}
	return out
}

// value styles a field's value; the result of a provider check that went wrong
// is the accent, so it is what the eye finds.
func (c *Colorizer) value(key, value string) string {
	if key == "result" && problemResults[strings.Trim(value, `"`)] {
		return render(&c.theme.Primary, value)
	}
	return render(&c.theme.Soft, value)
}

// words styles the text between fields, setting the [PERF] markers apart.
func (c *Colorizer) words(s string) string {
	if s == "" {
		return ""
	}
	var b strings.Builder
	last := 0
	for _, loc := range tagRe.FindAllStringIndex(s, -1) {
		b.WriteString(render(&c.theme.Text, s[last:loc[0]]))
		b.WriteString(render(&c.theme.Muted, s[loc[0]:loc[1]]))
		last = loc[1]
	}
	b.WriteString(render(&c.theme.Text, s[last:]))
	return b.String()
}

// render styles s without converting its tabs: lipgloss turns a tab into four
// spaces by default, which would change the line. A log line is shown as it was
// written.
func render(st *lipgloss.Style, s string) string {
	return st.TabWidth(lipgloss.NoTabConversion).Render(s)
}
