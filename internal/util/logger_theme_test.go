package util

import (
	"bytes"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The console log used the library's own palette (cyan, yellow, pink, violet),
// the "violet on gray" the screens left behind. Each level now speaks the
// site's identity.
func TestConsoleLogLevelsUseTheSiteColors(t *testing.T) {
	var buf bytes.Buffer
	l := newConsoleLogger(&buf, colorprofile.TrueColor, false)

	l.Info("info line")
	l.Warn("warn line")
	l.Error("error line")
	out := buf.String()

	for _, old := range []string{"38;5;86", "38;5;192", "38;5;204", "38;5;63", "38;5;134"} {
		assert.NotContains(t, out, old, "the library's default level colour %s", old)
	}
	assert.Contains(t, out, "48;2;163;0;31", "ERROR is the crimson notice block")
	assert.Contains(t, out, "ERROR", "the level is still spelled out")
	for _, line := range strings.Split(strings.TrimSpace(out), "\n") {
		assert.Contains(t, line, "GoAnime", "every line keeps the GoAnime block")
	}
}

// Without colour the levels still line up, so a log read in a pipe or a screen
// reader is as tidy as the coloured one.
func TestConsoleLogMessagesStartInTheSameColumn(t *testing.T) {
	var buf bytes.Buffer
	l := newConsoleLogger(&buf, colorprofile.NoTTY, false)

	l.Info("one")
	l.Warn("two")
	l.Error("three")

	assert.NotContains(t, buf.String(), "\x1b", "NoTTY carries no escape codes at all")
	lines := strings.Split(strings.TrimSuffix(buf.String(), "\n"), "\n") // keep each line's leading space
	require.Len(t, lines, 3)
	col := strings.Index(lines[0], "one")
	assert.Positive(t, col)
	assert.Equal(t, col, strings.Index(lines[1], "two"))
	assert.Equal(t, col, strings.Index(lines[2], "three"))
}
