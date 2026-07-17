package tui

import "charm.land/lipgloss/v2"

// Theme contains the shared visual language for GoAnime's interactive screens.
type Theme struct {
	Primary             lipgloss.Style
	Text                lipgloss.Style
	Muted               lipgloss.Style
	Header              lipgloss.Style
	Breadcrumb          lipgloss.Style
	Border              lipgloss.Style
	Footer              lipgloss.Style
	Panel               lipgloss.Style
	Label               lipgloss.Style
	Value               lipgloss.Style
	SelectedTitle       lipgloss.Style
	SelectedDescription lipgloss.Style
	FilterMatch         lipgloss.Style
}

// NewTheme builds a deterministic light or dark GoAnime theme.
func NewTheme(isDark bool) Theme {
	lightDark := lipgloss.LightDark(isDark)
	primary := lightDark(lipgloss.Color("#6D28D9"), lipgloss.Color("#A78BFA"))
	text := lightDark(lipgloss.Color("#1F2937"), lipgloss.Color("#F9FAFB"))
	muted := lightDark(lipgloss.Color("#6B7280"), lipgloss.Color("#9CA3AF"))
	border := lightDark(lipgloss.Color("#D1D5DB"), lipgloss.Color("#374151"))
	surface := lightDark(lipgloss.Color("#F3F4F6"), lipgloss.Color("#1F2937"))

	return Theme{
		Primary:    lipgloss.NewStyle().Foreground(primary).Bold(true),
		Text:       lipgloss.NewStyle().Foreground(text),
		Muted:      lipgloss.NewStyle().Foreground(muted),
		Header:     lipgloss.NewStyle().Foreground(lipgloss.Color("#FFFFFF")).Background(primary).Bold(true).Padding(0, 1),
		Breadcrumb: lipgloss.NewStyle().Foreground(muted),
		Border:     lipgloss.NewStyle().Foreground(border),
		Footer:     lipgloss.NewStyle().Foreground(muted),
		Panel: lipgloss.NewStyle().
			Background(surface).
			Border(lipgloss.RoundedBorder()).
			BorderForeground(border).
			Padding(1, 2),
		Label:               lipgloss.NewStyle().Foreground(muted),
		Value:               lipgloss.NewStyle().Foreground(text).Bold(true),
		SelectedTitle:       lipgloss.NewStyle().Foreground(primary).Bold(true).Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(primary).PaddingLeft(1),
		SelectedDescription: lipgloss.NewStyle().Foreground(primary).Border(lipgloss.NormalBorder(), false, false, false, true).BorderForeground(primary).PaddingLeft(1),
		FilterMatch:         lipgloss.NewStyle().Foreground(primary).Underline(true),
	}
}
