package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

const (
	defaultShellWidth  = 80
	defaultShellHeight = 24
	shellChromeHeight  = 4
)

// Shell renders common GoAnime navigation around a screen body.
type Shell struct {
	Theme      Theme
	Breadcrumb string
	Width      int
	Height     int
}

// NewShell creates a shell with safe dimensions before the first resize event.
func NewShell(theme *Theme, breadcrumb string) Shell {
	return Shell{
		Theme:      *theme,
		Breadcrumb: breadcrumb,
		Width:      defaultShellWidth,
		Height:     defaultShellHeight,
	}
}

// Resize updates the available terminal dimensions.
func (s *Shell) Resize(width, height int) {
	s.Width = max(width, 1)
	s.Height = max(height, shellChromeHeight+1)
}

// ContentSize returns space left after header, separators, and footer.
func (s *Shell) ContentSize() (int, int) {
	width := s.Width
	if width <= 0 {
		width = defaultShellWidth
	}
	height := s.Height
	if height <= 0 {
		height = defaultShellHeight
	}
	return width, max(height-shellChromeHeight, 1)
}

// Render wraps body content in responsive navigation chrome.
func (s *Shell) Render(body, footer string) string {
	width, height := s.ContentSize()
	header := s.Theme.Header.Render("GOANIME")
	if width >= 34 && s.Breadcrumb != "" {
		header += "  " + s.Theme.Breadcrumb.Render(s.Breadcrumb)
	}

	if width < 50 {
		switch {
		case width >= 36:
			footer = "type fzf  ↑↓/jk  enter  esc"
		case width >= 24:
			footer = "type  ↑↓  enter  esc"
		case width >= 12:
			footer = "type  enter  esc"
		case width >= 8:
			footer = "esc back"
		default:
			footer = "esc"
		}
	}
	separator := s.Theme.Border.Render(strings.Repeat("─", width))
	return lipgloss.JoinVertical(lipgloss.Left,
		fitBlock(header, width, 1),
		separator,
		fitBlock(body, width, height),
		separator,
		fitBlock(s.Theme.Footer.Render(footer), width, 1),
	)
}

// fitBlock clips ANSI-styled content to terminal cell dimensions.
func fitBlock(content string, width, height int) string {
	if width <= 0 || height <= 0 {
		return ""
	}
	content = strings.ToValidUTF8(content, "�")
	content = strings.ReplaceAll(content, "\r\n", "\n")
	content = strings.ReplaceAll(content, "\r", "\n")
	lines := strings.Split(content, "\n")
	if len(lines) > height {
		lines = lines[:height]
	}
	for index := range lines {
		lines[index] = ansi.Truncate(lines[index], width, "")
	}
	return strings.Join(lines, "\n")
}
