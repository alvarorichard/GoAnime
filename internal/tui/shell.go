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

// brandGloss is GoAnime in katakana, shown beside the header block when the
// terminal has room for it.
const brandGloss = "ゴーアニメ"

// stepGlosses names the screens of the app the way the site labels its
// sections: the word in capitals, its Japanese beside it.
var stepGlosses = map[string]string{
	"search":      "検索",
	"results":     "結果",
	"seasons":     "シーズン",
	"episodes":    "エピソード",
	"quality":     "画質",
	"playback":    "再生",
	"player":      "プレーヤー",
	"next":        "次へ",
	"audio":       "音声",
	"subtitles":   "字幕",
	"download":    "ダウンロード",
	"offline":     "オフライン",
	"watch":       "視聴",
	"watch range": "範囲",
}

// Hint is one key in the footer and what it does.
type Hint struct {
	Key  string
	Desc string
	// Essential hints are the last to go when the footer runs out of room.
	Essential bool
}

// Shell renders common GoAnime navigation around a screen body.
type Shell struct {
	Theme      Theme
	Breadcrumb string
	// Context is shown at the right of the header, such as the show that is
	// playing. It is the first thing dropped on a narrow terminal.
	Context string
	Width   int
	Height  int
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
func (s *Shell) ContentSize() (width, height int) {
	width = s.Width
	if width <= 0 {
		width = defaultShellWidth
	}
	height = s.Height
	if height <= 0 {
		height = defaultShellHeight
	}
	return width, max(height-shellChromeHeight, 1)
}

// Render wraps body content in the header, the rules and the key footer.
func (s *Shell) Render(body string, hints []Hint) string {
	width, height := s.ContentSize()
	separator := s.Theme.Border.Render(strings.Repeat("─", width))
	return lipgloss.JoinVertical(lipgloss.Left,
		fitBlock(s.header(width), width, 1),
		separator,
		fitBlock(body, width, height),
		separator,
		fitBlock(renderHints(&s.Theme, hints, width), width, 1),
	)
}

// header lays out the GOANIME block, the brand gloss, the breadcrumb trail
// and the context, dropping the least useful part until the line fits.
func (s *Shell) header(width int) string {
	block := s.Theme.Header.Render("GOANIME")
	steps := breadcrumbSteps(s.Breadcrumb)
	context := singleLine(s.Context)

	type layout struct {
		brandGloss, stepGloss, context bool
		trail                          int // steps shown, counted from the end; 0 hides the trail
	}
	layouts := []layout{
		{brandGloss: true, stepGloss: true, context: true, trail: len(steps)},
		{brandGloss: true, stepGloss: true, trail: len(steps)},
		{stepGloss: true, trail: len(steps)},
		{trail: len(steps)},
		{trail: min(len(steps), 1)},
	}
	for _, l := range layouts {
		left := block
		if l.brandGloss {
			left += s.Theme.Gloss.Render(" " + brandGloss)
		}
		if l.trail > 0 {
			left += "  " + s.renderTrail(steps, l.trail, l.stepGloss)
		}
		if lipgloss.Width(left) > width {
			continue
		}
		if l.context && context != "" {
			right := s.Theme.Muted.Render(context)
			gap := width - lipgloss.Width(left) - lipgloss.Width(right)
			if gap < 2 {
				continue
			}
			return left + strings.Repeat(" ", gap) + right
		}
		return left
	}
	return block
}

// renderTrail draws the last shown steps of the breadcrumb, the current one
// in the accent color. Steps cut from the front become an ellipsis.
func (s *Shell) renderTrail(steps []string, shown int, gloss bool) string {
	separator := s.Theme.Faint.Render(" › ")
	parts := make([]string, 0, shown+1)
	if shown < len(steps) {
		parts = append(parts, s.Theme.Faint.Render("…"))
	}
	for i, step := range steps[len(steps)-shown:] {
		label, jp := stepLabel(step)
		if i < shown-1 {
			parts = append(parts, s.Theme.Muted.Render(label))
			continue
		}
		current := s.Theme.Primary.Render(label)
		if gloss && jp != "" {
			current += s.Theme.Gloss.Render("  " + jp)
		}
		parts = append(parts, current)
	}
	return strings.Join(parts, separator)
}

// breadcrumbSteps splits a "Search > Results" trail into its steps.
func breadcrumbSteps(breadcrumb string) []string {
	var steps []string
	for _, step := range strings.Split(singleLine(breadcrumb), ">") {
		if step = strings.TrimSpace(step); step != "" {
			steps = append(steps, step)
		}
	}
	return steps
}

// stepLabel returns how a breadcrumb step reads in the header: a known screen
// in capitals with its Japanese name, anything else (a show's title) as is.
func stepLabel(step string) (label, gloss string) {
	if jp, ok := stepGlosses[strings.ToLower(step)]; ok {
		return strings.ToUpper(step), jp
	}
	return step, ""
}

// renderHints draws the footer keys as keycaps. When they do not fit, the
// hints that are not essential go first, from the last one back.
func renderHints(theme *Theme, hints []Hint, width int) string {
	shown := append([]Hint(nil), hints...)
	for {
		line := joinHints(theme, shown)
		if lipgloss.Width(line) <= width || len(shown) <= 1 {
			return line
		}
		drop := len(shown) - 1
		for i := len(shown) - 1; i >= 0; i-- {
			if !shown[i].Essential {
				drop = i
				break
			}
		}
		shown = append(shown[:drop], shown[drop+1:]...)
	}
}

func joinHints(theme *Theme, hints []Hint) string {
	parts := make([]string, 0, len(hints))
	for _, h := range hints {
		parts = append(parts, theme.Keycap.Render(h.Key)+theme.Footer.Render(" "+h.Desc))
	}
	return strings.Join(parts, "   ")
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
