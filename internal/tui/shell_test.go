package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
)

// newTestShell builds a shell over an addressable test theme.
func newTestShell(breadcrumb string) Shell {
	theme := NewTheme(true)
	return NewShell(&theme, breadcrumb)
}

func TestNewShell(t *testing.T) {
	t.Parallel()

	shell := newTestShell("Search > Results")

	assert.Equal(t, defaultShellWidth, shell.Width)
	assert.Equal(t, defaultShellHeight, shell.Height)
	assert.Equal(t, "Search > Results", shell.Breadcrumb)
}

func TestShellResize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		width      int
		height     int
		wantWidth  int
		wantHeight int
	}{
		{name: "normal", width: 120, height: 40, wantWidth: 120, wantHeight: 40},
		{name: "minimum", width: 0, height: 0, wantWidth: 1, wantHeight: shellChromeHeight + 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			shell := newTestShell("")
			shell.Resize(tt.width, tt.height)
			assert.Equal(t, tt.wantWidth, shell.Width)
			assert.Equal(t, tt.wantHeight, shell.Height)
		})
	}

}

func TestShellContentSize(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		shell      Shell
		wantWidth  int
		wantHeight int
	}{
		{name: "defaults for zero value", shell: Shell{}, wantWidth: defaultShellWidth, wantHeight: defaultShellHeight - shellChromeHeight},
		{name: "resized", shell: Shell{Width: 100, Height: 30}, wantWidth: 100, wantHeight: 26},
		{name: "minimum body", shell: Shell{Width: 10, Height: 2}, wantWidth: 10, wantHeight: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			width, height := tt.shell.ContentSize()
			assert.Equal(t, tt.wantWidth, width)
			assert.Equal(t, tt.wantHeight, height)
		})
	}

}

func TestFitBlock(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		content string
		width   int
		height  int
		want    string
	}{
		{name: "width", content: "123456", width: 4, height: 1, want: "1234"},
		{name: "height", content: "one\ntwo\nthree", width: 10, height: 2, want: "one\ntwo"},
		{name: "carriage returns", content: "one\r\ntwo\rthree", width: 10, height: 3, want: "one\ntwo\nthree"},
		{name: "ansi aware", content: "\x1b[31mabcdef\x1b[0m", width: 3, height: 1, want: "\x1b[31mabc\x1b[0m"},
		{name: "invalid dimensions", content: "value", width: 0, height: 1, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, fitBlock(tt.content, tt.width, tt.height))
		})
	}

}

func FuzzFitBlockRespectsBounds(f *testing.F) {
	for _, seed := range []string{"plain", "one\ntwo\nthree", "\x1b[31mcolored\x1b[0m", "日本語 🎬"} {
		f.Add(seed, uint8(20), uint8(5))
	}
	f.Fuzz(func(t *testing.T, content string, rawWidth, rawHeight uint8) {
		width := int(rawWidth%80) + 1
		height := int(rawHeight%30) + 1
		got := fitBlock(content, width, height)
		assert.LessOrEqual(t, lipgloss.Width(got), width)
		assert.LessOrEqual(t, lipgloss.Height(got), height)
	})
}

var testHints = []Hint{
	{Key: "↑↓", Desc: "move"},
	{Key: "type", Desc: "filter"},
	{Key: "enter", Desc: "select", Essential: true},
	{Key: "esc", Desc: "back", Essential: true},
}

func TestShellRender(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		width       int
		wantTrail   bool
		wantFooter  string
		droppedHint string
	}{
		{name: "wide", width: 100, wantTrail: true, wantFooter: "filter"},
		{name: "compact", width: 40, wantTrail: true, wantFooter: "select", droppedHint: "filter"},
		{name: "tiny", width: 20, wantTrail: false, wantFooter: "enter", droppedHint: "move"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			theme := NewTheme(true)
			shell := NewShell(&theme, "Search > Results")
			shell.Resize(tt.width, 24)

			got := shell.Render("BODY", testHints)
			plain := ansi.Strip(got)

			assert.Contains(t, plain, "GOANIME")
			assert.Contains(t, plain, "BODY")
			assert.Contains(t, plain, tt.wantFooter)
			if tt.droppedHint != "" {
				assert.NotContains(t, plain, tt.droppedHint)
			}
			assert.Equal(t, tt.wantTrail, strings.Contains(plain, "SEARCH › RESULTS"))
			assert.Contains(t, got, strings.Repeat("─", tt.width))
			assert.GreaterOrEqual(t, lipgloss.Height(got), 5)
		})
	}

	t.Run("never exceeds terminal bounds", func(t *testing.T) {
		t.Parallel()
		for _, size := range []struct {
			width  int
			height int
		}{{1, 5}, {8, 5}, {20, 6}, {33, 8}, {80, 12}} {
			shell := newTestShell(strings.Repeat("long breadcrumb ", 20))
			shell.Context = strings.Repeat("context ", 20)
			shell.Resize(size.width, size.height)
			body := strings.Repeat("very long body ", 20) + "\n" + strings.Repeat("extra line\n", 20)

			got := shell.Render(body, []Hint{{Key: "x", Desc: strings.Repeat("long footer ", 20)}})

			assert.LessOrEqual(t, lipgloss.Width(got), shell.Width)
			assert.LessOrEqual(t, lipgloss.Height(got), shell.Height)
		}
	})
}

func TestShellHeader(t *testing.T) {
	t.Parallel()

	header := func(breadcrumb, context string, width int) string {
		shell := newTestShell(breadcrumb)
		shell.Context = context
		return ansi.Strip(shell.header(width))
	}

	t.Run("everything fits", func(t *testing.T) {
		t.Parallel()
		got := header("Search > Results", "Momotarou", 100)
		assert.True(t, strings.HasPrefix(got, " GOANIME  ゴーアニメ  SEARCH › RESULTS  結果"), got)
		assert.True(t, strings.HasSuffix(got, "Momotarou"), got)
		assert.Equal(t, 100, lipgloss.Width(got), "the context is pushed to the right edge")
	})

	t.Run("context goes first, then the glosses", func(t *testing.T) {
		t.Parallel()
		got := header("Search > Results", "Momotarou", 50)
		assert.NotContains(t, got, "Momotarou")
		assert.Contains(t, got, "ゴーアニメ")

		got = header("Search > Results", "", 40)
		assert.NotContains(t, got, "ゴーアニメ")
		assert.Contains(t, got, "結果")
	})

	t.Run("the trail keeps its last step", func(t *testing.T) {
		t.Parallel()
		got := header("Search > Results > Episodes", "", 30)
		assert.Equal(t, " GOANIME   … › EPISODES", got)
	})

	t.Run("a title is not a step name", func(t *testing.T) {
		t.Parallel()
		got := header("Search > Saru to Kani > Seasons", "", 120)
		assert.Contains(t, got, "SEARCH › Saru to Kani › SEASONS  シーズン")
	})
}

func TestBreadcrumbSteps(t *testing.T) {
	t.Parallel()
	assert.Equal(t, []string{"Search", "Results"}, breadcrumbSteps("Search > Results"))
	assert.Equal(t, []string{"Player"}, breadcrumbSteps("  Player  "))
	assert.Equal(t, []string{"Search", "Evil"}, breadcrumbSteps("Search >> \x1b[31mEvil\x1b[0m >"))
	assert.Empty(t, breadcrumbSteps(""))
}

func TestRenderHints(t *testing.T) {
	t.Parallel()
	theme := NewTheme(true)

	full := ansi.Strip(renderHints(&theme, testHints, 200))
	assert.Equal(t, " ↑↓  move    type  filter    enter  select    esc  back", full)

	// Short of room, the hints that are not essential go first, from the
	// last one back; enter and esc stay.
	for width := 60; width >= 1; width-- {
		got := ansi.Strip(renderHints(&theme, testHints, width))
		if lipgloss.Width(got) > width {
			assert.Equal(t, "enter", strings.Fields(got)[0], "width %d: only one hint left", width)
			continue
		}
		if strings.Contains(got, "move") || strings.Contains(got, "filter") {
			assert.Contains(t, got, "select", "width %d", width)
			assert.Contains(t, got, "back", "width %d", width)
		}
	}
}
