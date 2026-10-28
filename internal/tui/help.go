package tui

import (
	"strings"

	"charm.land/lipgloss/v2"
)

// HelpRow is one key and what it does in the KEYS overlay.
type HelpRow struct {
	Key  string
	Desc string
}

// Help for every screen, from the design's KEYS overlay (tela 10): esc takes
// one step back and never quits; ctrl+c always quits GoAnime.
func everyScreenHelp() []HelpRow {
	move := "↑↓ j k"
	if Plain() {
		move = "up/dn j k"
	}
	return []HelpRow{
		{Key: move, Desc: "move"},
		{Key: "enter", Desc: "open, play, resume"},
		{Key: "esc", Desc: "one step back"},
		{Key: "ctrl+c", Desc: "quit GoAnime"},
		{Key: "?", Desc: "this help"},
	}
}

// ListsHelp is the LISTS section of the KEYS overlay.
func listsHelp() []HelpRow {
	return []HelpRow{
		{Key: "type", Desc: "filter the list"},
		{Key: "g G", Desc: "first, last"},
		{Key: "pgup pgdn", Desc: "page"},
	}
}

// RenderHelp draws the KEYS overlay for a screen: the shared keys on the
// left, this screen's keys and the extra sections on the right, and the
// "esc never quits" promise at the bottom. screen names the current screen
// ("RESULTS", "EPISODES"), this lists its keys, extra adds sections such as
// EPISODES download keys. It returns "" when too narrow to be useful.
func RenderHelp(theme *Theme, screen string, this []HelpRow, extra map[string][]HelpRow, width, maxHeight int) string {
	if width < minPanelWidth {
		return ""
	}
	left := []HelpRow{{Key: "EVERY SCREEN", Desc: ""}}
	left = append(left, everyScreenHelp()...)
	left = append(left, HelpRow{}, HelpRow{Key: "LISTS", Desc: ""})
	left = append(left, listsHelp()...)

	rightTitle := "THIS SCREEN"
	if screen = singleLine(screen); screen != "" {
		sep := " · "
		if Plain() {
			sep = " - "
		}
		rightTitle += sep + strings.ToUpper(screen)
	}
	right := append([]HelpRow{{Key: rightTitle, Desc: ""}}, this...)
	names := sortedKeys(extra)
	for _, name := range names {
		right = append(right, HelpRow{}, HelpRow{Key: name, Desc: ""})
		right = append(right, extra[name]...)
	}

	keyWidth := func(rows []HelpRow) int {
		w := 0
		for _, r := range rows {
			if r.Desc == "" {
				continue
			}
			w = max(w, lipgloss.Width(r.Key))
		}
		return w
	}
	lw, rw := keyWidth(left), keyWidth(right)

	renderRow := func(r HelpRow, kw int) string {
		if r.Key == "" {
			return ""
		}
		if r.Desc == "" {
			return theme.PanelMuted.Render(r.Key)
		}
		key := theme.PanelCode.Render(" " + r.Key + strings.Repeat(" ", max(kw-lipgloss.Width(r.Key), 0)) + " ")
		return key + "  " + theme.PanelText.Render(r.Desc)
	}

	height := max(len(left), len(right))
	rows := make([]string, 0, height+2)
	for i := 0; i < height; i++ {
		l, r := "", ""
		if i < len(left) {
			l = renderRow(left[i], lw)
		}
		if i < len(right) {
			r = renderRow(right[i], rw)
		}
		rows = append(rows, rstrip(l+"    "+r))
	}
	rows = append(rows, "", theme.PanelMuted.Render("esc never quits. ctrl+c always does."))

	content := &panelContent{label: "Keys", gloss: "キー", lines: rows}
	if Plain() {
		// Tela 9: help is plain text, no paper panel.
		return "KEYS\n\n" + strings.Join(rows, "\n")
	}
	if panel := fitPanel(theme, content, width, maxHeight); panel != "" {
		return panel
	}
	return strings.Join(rows, "\n")
}

func rstrip(s string) string {
	return strings.TrimRight(s, " ")
}

func sortedKeys(m map[string][]HelpRow) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	for i := 1; i < len(keys); i++ {
		for j := i; j > 0 && keys[j] < keys[j-1]; j-- {
			keys[j], keys[j-1] = keys[j-1], keys[j]
		}
	}
	return keys
}
