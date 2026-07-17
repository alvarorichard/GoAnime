package tui

import (
	"strings"
	"testing"

	"charm.land/lipgloss/v2"
	"github.com/stretchr/testify/assert"
)

func TestNewShell(t *testing.T) {
	t.Parallel()

	shell := NewShell(NewTheme(true), "Search > Results")

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
			shell := NewShell(NewTheme(true), "")
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

func TestShellRender(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name           string
		width          int
		wantBreadcrumb bool
		wantFooter     string
	}{
		{name: "wide", width: 100, wantBreadcrumb: true, wantFooter: "filter"},
		{name: "compact", width: 40, wantBreadcrumb: true, wantFooter: "enter open"},
		{name: "tiny", width: 20, wantBreadcrumb: false, wantFooter: "enter open"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			shell := NewShell(NewTheme(true), "Search > Results")
			shell.Resize(tt.width, 24)

			got := shell.Render("BODY", "↑↓ navigate  / filter")

			assert.Contains(t, got, "GOANIME")
			assert.Contains(t, got, "BODY")
			assert.Contains(t, got, tt.wantFooter)
			assert.Equal(t, tt.wantBreadcrumb, strings.Contains(got, "Search > Results"))
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
			shell := NewShell(NewTheme(true), strings.Repeat("long breadcrumb ", 20))
			shell.Resize(size.width, size.height)
			body := strings.Repeat("very long body ", 20) + "\n" + strings.Repeat("extra line\n", 20)

			got := shell.Render(body, strings.Repeat("long footer ", 20))

			assert.LessOrEqual(t, lipgloss.Width(got), shell.Width)
			assert.LessOrEqual(t, lipgloss.Height(got), shell.Height)
		}
	})
}
