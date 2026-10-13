package tui

import (
	"fmt"
	"math"
	"strings"

	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// minPanelWidth is the narrowest panel worth drawing, shadow included.
const minPanelWidth = 24

// halftone is the site's screen tone in braille dots: dense in a panel's top
// right corner, fading toward the left and down. A row is cut from the left
// when the panel's text needs the room.
var halftone = []string{
	"⠀⠀⠁⠄⠁⠅⠅⠅⠕⢅⠕⢕⢕⢕⢝⢵⢝⢽",
	"⠀⠀⠁⠀⠁⠅⠁⠅⠕⠅⠕⢕⠕⢕⢝⢕⢝⢕",
	"⠀⠀⠀⠀⠁⠄⠁⠅⠅⠅⠕⠅⠕⢕⠕⢕⠕⢕",
	"⠀⠀⠀⠀⠁⠀⠁⠅⠁⠅⠅⠅⠕⠅⠕⠅⠕⢅",
	"⠀⠀⠀⠀⠀⠀⠁⠀⠁⠅⠁⠅⠁⠅⠅⠅⠕⠅",
}

// panelContent is what a paper panel holds, from its heading down.
type panelContent struct {
	label    string // heading, drawn in capitals
	gloss    string // the heading's Japanese
	title    string // bold first line
	progress *panelProgress
	fields   []detailField
	lines    []string // text, wrapped to the panel
	command  string   // a command line, drawn on ink
}

// panelProgress is a position within a known duration, in seconds.
type panelProgress struct{ position, duration float64 }

// detailField is one labelled value in a panel.
type detailField struct{ label, value string }

// fitPanel draws content as a paper panel with its shadow, width cells wide
// (shadow included) and at most maxHeight rows tall. Labels sit above their
// values when that fits; on a shorter screen each field takes one line
// instead; when even that does not fit, the panel is left out ("") rather
// than drawn cut off.
func fitPanel(theme *Theme, content *panelContent, width, maxHeight int) string {
	if width < minPanelWidth {
		return ""
	}
	panelWidth := width - 1 // the shadow column
	inner := panelWidth - 4 // two cells of padding each side
	for _, compact := range []bool{false, true} {
		lines := panelLines(theme, content, inner, compact)
		if len(lines)+2+1 <= maxHeight { // padding rows and the shadow row
			return addShadow(theme, renderPanelBlock(theme, lines, panelWidth))
		}
	}
	return ""
}

// panelLines lays out content within inner cells.
func panelLines(theme *Theme, c *panelContent, inner int, compact bool) []string {
	var out []string
	gap := func() {
		if !compact {
			out = append(out, "")
		}
	}

	head := theme.PanelLabel.Render(ansi.Truncate(strings.ToUpper(singleLine(c.label)), inner, "…"))
	if gloss := singleLine(c.gloss); gloss != "" && lipgloss.Width(head)+2+lipgloss.Width(gloss) <= inner {
		head += theme.PanelMuted.Render("  " + gloss)
	}
	out = append(out, head)

	if title := singleLine(c.title); title != "" {
		gap()
		out = append(out, theme.PanelValue.Render(ansi.Truncate(title, inner, "…")))
	}

	// The line is kept while the duration is still unknown, so the panel does
	// not change height when it arrives.
	if p := c.progress; p != nil {
		gap()
		clock, ratio := "--:-- / --:--", 0.0
		if p.duration > 0 {
			clock, ratio = FormatClock(p.position)+" / "+FormatClock(p.duration), p.position/p.duration
		}
		barWidth := inner - lipgloss.Width(clock) - 2
		line := theme.PanelValue.Render(clock)
		if barWidth >= 6 {
			line = renderBar(barWidth, ratio, &theme.PanelBar, &theme.PanelTrack) + theme.PanelText.Render("  ") + line
		}
		out = append(out, line)
	}

	if compact {
		labelWidth := 0
		for _, f := range c.fields {
			labelWidth = max(labelWidth, lipgloss.Width(f.label))
		}
		for _, f := range c.fields {
			label := ansi.Truncate(f.label+strings.Repeat(" ", labelWidth-lipgloss.Width(f.label)+2), inner, "")
			value := ansi.Truncate(singleLine(f.value), max(inner-lipgloss.Width(label), 0), "…")
			out = append(out, theme.PanelMuted.Render(label)+theme.PanelValue.Render(value))
		}
	} else {
		for _, f := range c.fields {
			out = append(out, "",
				theme.PanelMuted.Render(ansi.Truncate(f.label, inner, "…")),
				theme.PanelValue.Render(ansi.Truncate(singleLine(f.value), inner, "…")))
		}
	}

	if len(c.lines) > 0 {
		gap()
		for _, text := range c.lines {
			for line := range strings.SplitSeq(ansi.Wrap(singleLine(text), inner, ""), "\n") {
				out = append(out, theme.PanelText.Render(line))
			}
		}
	}

	if command := singleLine(c.command); command != "" {
		gap()
		code := ansi.Truncate(" $ "+command, inner, "…")
		out = append(out, theme.PanelCode.Render(code+strings.Repeat(" ", max(inner-lipgloss.Width(code), 0))))
	}
	return out
}

// renderPanelBlock fills a width-cell paper block around lines, with a row of
// padding above and below and the halftone in the top right corner wherever
// the text leaves it room.
func renderPanelBlock(theme *Theme, lines []string, width int) string {
	inner := width - 4
	fill := theme.PanelText
	rows := make([]string, 0, len(lines)+2)
	for r := range len(lines) + 2 {
		content := ""
		if r >= 1 && r-1 < len(lines) {
			content = ansi.Truncate(lines[r-1], inner, "")
		}
		left := fill.Render("  ") + content
		used := lipgloss.Width(left)

		tone := ""
		if r < len(halftone) {
			if room := width - used - 2; room > 0 {
				tone = theme.PanelTone.Render(rightCells(halftone[r], room))
			}
		}
		pad := width - used - lipgloss.Width(tone)
		rows = append(rows, left+fill.Render(strings.Repeat(" ", max(pad, 0)))+tone)
	}
	return strings.Join(rows, "\n")
}

// addShadow puts the site's hard offset shadow under a block: a column of
// full blocks on the right that starts half a row down, and a row of upper
// half blocks underneath that starts one cell in. The result is one cell
// wider and one row taller.
func addShadow(theme *Theme, block string) string {
	lines := strings.Split(block, "\n")
	width := 0
	for _, line := range lines {
		width = max(width, lipgloss.Width(line))
	}
	out := make([]string, 0, len(lines)+1)
	for i, line := range lines {
		line += strings.Repeat(" ", width-lipgloss.Width(line))
		if i == 0 {
			out = append(out, line+theme.Shadow.Render("▄"))
		} else {
			out = append(out, line+theme.Shadow.Render("█"))
		}
	}
	out = append(out, " "+theme.Shadow.Render(strings.Repeat("▀", width)))
	return strings.Join(out, "\n")
}

// renderBar draws ratio of width cells filled.
func renderBar(width int, ratio float64, filled, empty *lipgloss.Style) string {
	if width <= 0 {
		return ""
	}
	if math.IsNaN(ratio) {
		ratio = 0
	}
	n := int(math.Round(math.Max(0, math.Min(1, ratio)) * float64(width)))
	return filled.Render(strings.Repeat("━", n)) + empty.Render(strings.Repeat("━", width-n))
}

// FormatClock renders seconds as mm:ss, or h:mm:ss from an hour up.
func FormatClock(seconds float64) string {
	s := max(int(seconds), 0)
	if s >= 3600 {
		return fmt.Sprintf("%d:%02d:%02d", s/3600, s%3600/60, s%60)
	}
	return fmt.Sprintf("%02d:%02d", s/60, s%60)
}

// rightCells keeps the last n runes of a string of single-cell runes.
func rightCells(s string, n int) string {
	runes := []rune(s)
	if len(runes) <= n {
		return s
	}
	return string(runes[len(runes)-n:])
}
