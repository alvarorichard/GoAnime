package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	debugLogPath = "/home/krone/.local/share/goanime/logs/goanime_2026-11-02_15-49-12.log"
	debugFollow  = `tail -f "` + debugLogPath + `"`
)

func TestDebugPanelIsAPaperPanelWithTheWholeCommand(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)

	panel := debugPanel(&theme, debugLogPath, debugFollow, 100)

	require.NotEmpty(t, panel)
	plain := ansi.Strip(panel)
	assert.Contains(t, plain, "DEBUG")
	assert.Contains(t, plain, debugLogPath, "the path is shown whole")
	assert.Contains(t, plain, debugFollow, "the command is shown whole, ready to copy")
	assert.Contains(t, plain, "▀▀▀", "the hard shadow of the site's panels")
	assert.LessOrEqual(t, lipgloss.Width(panel), 100)
	assert.NotContains(t, plain, "…", "nothing is cut")
}

// The command is copied from the screen, so a panel that would have to cut it
// steps aside for the plain line, which prints it whole.
func TestDebugPanelStepsAsideRatherThanCutTheCommand(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)

	for _, width := range []int{0, 20, 60, len(debugFollow) + 3} {
		assert.Empty(t, debugPanel(&theme, debugLogPath, debugFollow, width), "width %d", width)
	}
	assert.NotEmpty(t, debugPanel(&theme, debugLogPath, debugFollow, len(debugFollow)+3+5),
		"the first width that holds the command")
}

func TestDebugPanelCommandOnWindows(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)
	cmd := `Get-Content -Wait -Tail 50 "C:\Users\k\AppData\Local\GoAnime\logs\goanime_2026-11-02_15-49-12.log"`

	panel := debugPanel(&theme, `C:\Users\k\AppData\Local\GoAnime\logs\goanime_2026-11-02_15-49-12.log`, cmd, 120)

	require.NotEmpty(t, panel)
	assert.Contains(t, ansi.Strip(panel), cmd)
	assert.False(t, strings.Contains(ansi.Strip(panel), "…"))
}
