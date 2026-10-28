package tui

import "charm.land/lipgloss/v2"

// Theme contains the shared visual language for GoAnime's interactive screens.
//
// It follows goanime.online: an ink ground, paper panels, crimson as the only
// accent, square corners and hard shadows. On a dark terminal the values are
// the site's own; on a light one the panel turns to ink so it still stands
// out, and crimson replaces the light accent, which is too faint on white.
//
// Contrast on the ink ground: text 16:1, soft 12:1, muted 6.4:1, primary
// 6.6:1, white on crimson 8.1:1. On paper: ink-soft 8.8:1, crimson 6.8:1.
type Theme struct {
	Text    lipgloss.Style // body text
	Soft    lipgloss.Style // secondary text that still reads as content
	Muted   lipgloss.Style // descriptions, counts and help text
	Faint   lipgloss.Style // separators; never carries meaning alone
	Primary lipgloss.Style // the accent for text: current step, marks, prompts

	Header lipgloss.Style // the GOANIME block
	Gloss  lipgloss.Style // the katakana beside a label
	Border lipgloss.Style // horizontal rules
	Keycap lipgloss.Style // a key in the footer
	Footer lipgloss.Style // what a key does

	Selected      lipgloss.Style // the cursor row, filled in crimson
	SelectedMuted lipgloss.Style // secondary text on the cursor row
	SelectedFaint lipgloss.Style // the empty part of a bar on the cursor row
	FilterMatch   lipgloss.Style // matched characters while filtering
	Tag           lipgloss.Style // a badge such as FILLER, inverted against the ground
	Notice        lipgloss.Style // the "!" block before a notice

	Panel      lipgloss.Style // the paper fill behind a panel
	PanelLabel lipgloss.Style // a panel's heading
	PanelText  lipgloss.Style
	PanelMuted lipgloss.Style // labels inside a panel
	PanelValue lipgloss.Style // values inside a panel
	PanelTone  lipgloss.Style // the halftone dots in a panel's corner
	PanelCode  lipgloss.Style // a command line set inside a panel
	PanelBar   lipgloss.Style // the filled part of a bar inside a panel
	PanelTrack lipgloss.Style // the rest of a bar inside a panel
	Shadow     lipgloss.Style // the hard shadow under a panel

	Meter      lipgloss.Style // the filled part of a bar
	MeterEmpty lipgloss.Style // the rest of a bar
}

// NewTheme builds a deterministic light or dark GoAnime theme.
func NewTheme(isDark bool) Theme {
	lightDark := lipgloss.LightDark(isDark)

	ink := lipgloss.Color("#0B0B0E")
	paper := lipgloss.Color("#F5F7FA")
	crimson := lipgloss.Color("#A3001F")
	white := lipgloss.Color("#FFFFFF")

	text := lightDark(ink, lipgloss.Color("#E6EBF0"))
	soft := lightDark(lipgloss.Color("#3A3F47"), lipgloss.Color("#C9CED6"))
	muted := lightDark(lipgloss.Color("#5B606B"), lipgloss.Color("#8E93A0"))
	faint := lightDark(lipgloss.Color("#A9AEB8"), lipgloss.Color("#4A4A56"))
	primary := lightDark(crimson, lipgloss.Color("#FF5C70"))
	border := lightDark(lipgloss.Color("#C9CED6"), lipgloss.Color("#34343F"))
	raised := lightDark(lipgloss.Color("#E6EBF0"), lipgloss.Color("#17171D"))
	tagFg := lightDark(lipgloss.Color("#E6EBF0"), ink)
	tagBg := lightDark(ink, lipgloss.Color("#E6EBF0"))

	panel := lightDark(lipgloss.Color("#111116"), paper)
	panelText := lightDark(lipgloss.Color("#E6EBF0"), ink)
	panelMuted := lightDark(lipgloss.Color("#A9AEB8"), lipgloss.Color("#3A3F47"))
	panelLabel := lightDark(lipgloss.Color("#FF5C70"), crimson)
	panelTone := lightDark(lipgloss.Color("#3A3F47"), lipgloss.Color("#9AA3AD"))

	onPanel := lipgloss.NewStyle().Background(panel)
	onCrimson := lipgloss.NewStyle().Background(crimson)

	return Theme{
		Text:    lipgloss.NewStyle().Foreground(text),
		Soft:    lipgloss.NewStyle().Foreground(soft),
		Muted:   lipgloss.NewStyle().Foreground(muted),
		Faint:   lipgloss.NewStyle().Foreground(faint),
		Primary: lipgloss.NewStyle().Foreground(primary).Bold(true),

		Header: onCrimson.Foreground(white).Bold(true).Padding(0, 1),
		Gloss:  lipgloss.NewStyle().Foreground(muted),
		Border: lipgloss.NewStyle().Foreground(border),
		Keycap: lipgloss.NewStyle().Foreground(soft).Background(raised).Padding(0, 1),
		Footer: lipgloss.NewStyle().Foreground(muted),

		Selected:      onCrimson.Foreground(white).Bold(true),
		SelectedMuted: onCrimson.Foreground(lipgloss.Color("#FFD9DF")),
		SelectedFaint: onCrimson.Foreground(lipgloss.Color("#D9707F")),
		FilterMatch:   lipgloss.NewStyle().Foreground(primary).Underline(true),
		Tag:           lipgloss.NewStyle().Foreground(tagFg).Background(tagBg).Bold(true).Padding(0, 1),
		Notice:        onCrimson.Foreground(white).Bold(true).Padding(0, 1),

		Panel:      onPanel.Foreground(panelText).Padding(1, 2),
		PanelLabel: onPanel.Foreground(panelLabel).Bold(true),
		PanelText:  onPanel.Foreground(panelText),
		PanelMuted: onPanel.Foreground(panelMuted),
		PanelValue: onPanel.Foreground(panelText).Bold(true),
		PanelTone:  onPanel.Foreground(panelTone),
		PanelCode:  lipgloss.NewStyle().Foreground(lipgloss.Color("#E6EBF0")).Background(lightDark(lipgloss.Color("#24242C"), ink)),
		PanelBar:   onPanel.Foreground(panelLabel),
		PanelTrack: onPanel.Foreground(lightDark(lipgloss.Color("#34343F"), lipgloss.Color("#C9CED6"))),
		Shadow:     lipgloss.NewStyle().Foreground(crimson),

		Meter:      lipgloss.NewStyle().Foreground(primary),
		MeterEmpty: lipgloss.NewStyle().Foreground(border),
	}
}
