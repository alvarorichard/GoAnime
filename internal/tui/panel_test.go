package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAddShadow(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)

	got := strings.Split(ansi.Strip(addShadow(&theme, "abcd\nefgh\nijkl")), "\n")

	// One cell to the right and half a row down, like the site's offset
	// shadow: a lower half block on the first row, full blocks below it,
	// and upper half blocks under the panel starting one cell in.
	assert.Equal(t, []string{"abcd▄", "efgh█", "ijkl█", " ▀▀▀▀"}, got)
}

func TestFitPanel(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)
	content := &panelContent{
		label: "Details", gloss: "詳細", title: "Momotarou",
		fields: []detailField{{"Source", "AnimeFire"}, {"Year", "1918"}},
	}

	t.Run("stacked when there is room", func(t *testing.T) {
		t.Parallel()
		panel := fitPanel(&theme, content, 40, 30)
		require.NotEmpty(t, panel)
		plain := ansi.Strip(panel)
		assert.Contains(t, plain, "DETAILS  詳細")
		assert.Contains(t, plain, "Momotarou")
		assert.Equal(t, 40, lipgloss.Width(panel), "the width given includes the shadow")
		// Heading, gap, title, then a gap, label and value per field, inside
		// a row of padding above and below, and the shadow row.
		assert.Equal(t, 3+2*3+2+1, lipgloss.Height(panel))
	})

	t.Run("one line per field when short", func(t *testing.T) {
		t.Parallel()
		panel := fitPanel(&theme, content, 40, 8)
		require.NotEmpty(t, panel)
		assert.Contains(t, ansi.Strip(panel), "Source  AnimeFire")
		assert.LessOrEqual(t, lipgloss.Height(panel), 8)
	})

	t.Run("left out rather than cut", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, fitPanel(&theme, content, 40, 6))
		assert.Empty(t, fitPanel(&theme, content, minPanelWidth-1, 30))
	})

	t.Run("every row is filled to the panel's edge", func(t *testing.T) {
		t.Parallel()
		panel := fitPanel(&theme, content, 40, 30)
		lines := strings.Split(panel, "\n")
		for _, line := range lines[:len(lines)-1] {
			assert.Equal(t, 40, lipgloss.Width(line))
		}
	})
}

func TestPanelLines_Progress(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)

	content := &panelContent{label: "Now playing", title: "Momotarou", progress: &panelProgress{position: 492, duration: 1420}}
	plain := ansi.Strip(strings.Join(panelLines(&theme, content, 40, false), "\n"))
	assert.Contains(t, plain, "08:12 / 23:40")
	assert.Contains(t, plain, "━")

	// Before mpv reports a duration the line is already there, so the panel
	// keeps its height when the time arrives.
	content.progress = &panelProgress{}
	plain = ansi.Strip(strings.Join(panelLines(&theme, content, 40, false), "\n"))
	assert.Contains(t, plain, "--:-- / --:--")
}

func TestPanelLines_TextAndCommand(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)
	content := &panelContent{
		label:   "Skip this menu",
		lines:   []string{"Start with a quality and this menu opens only when an episode does not have it."},
		command: "goanime --quality 720p",
	}
	lines := panelLines(&theme, content, 30, false)
	for _, line := range lines {
		assert.LessOrEqual(t, lipgloss.Width(line), 30)
	}
	plain := ansi.Strip(strings.Join(lines, "\n"))
	assert.Contains(t, plain, "SKIP THIS MENU")
	assert.Contains(t, plain, " $ goanime --quality 720p")
	assert.Greater(t, strings.Count(plain, "\n"), 4, "the text wraps to the panel")
}

func TestRenderPanelBlock_HalftoneGivesWay(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)

	short := strings.Split(ansi.Strip(renderPanelBlock(&theme, []string{"A"}, 40)), "\n")
	assert.True(t, strings.HasSuffix(short[1], rightCells(halftone[1], 18)), "a short line leaves the whole tone")

	long := strings.Split(ansi.Strip(renderPanelBlock(&theme, []string{strings.Repeat("x", 36)}, 40)), "\n")
	assert.Equal(t, 40, lipgloss.Width(long[1]))
	assert.Contains(t, long[1], strings.Repeat("x", 36), "text always wins over the tone")
}

func TestRenderBar(t *testing.T) {
	t.Parallel()
	plain := lipgloss.NewStyle()
	assert.Equal(t, "", renderBar(0, 0.5, &plain, &plain))
	assert.Equal(t, 10, lipgloss.Width(renderBar(10, 0.5, &plain, &plain)))
	assert.Equal(t, 10, lipgloss.Width(renderBar(10, 7, &plain, &plain)), "a ratio over 1 is clamped")
	assert.Equal(t, 10, lipgloss.Width(renderBar(10, -1, &plain, &plain)))
}

func TestFormatClock(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "00:00", FormatClock(0))
	assert.Equal(t, "00:00", FormatClock(-5))
	assert.Equal(t, "14:32", FormatClock(872.9))
	assert.Equal(t, "1:02:03", FormatClock(3723))
}
