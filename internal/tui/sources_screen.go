package tui

import (
	"fmt"
	"strings"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// SourceRow is one line of the SOURCES screen: a catalogue the search can ask.
type SourceRow struct {
	ID     string // what SetSource receives
	Name   string
	Note   string // why it is in this state, or a caveat
	On     bool
	Locked bool // the environment decides; the screen cannot change it
}

// LoadSourcesFunc lists the sources as they stand now.
type LoadSourcesFunc func() []SourceRow

// SetSourceFunc saves the user's choice for one source.
type SetSourceFunc func(id string, on bool) error

// SourcesSummary reads "6 sources on · Goyabu off" for the header: how many
// sources search, and the ones that sit out.
func SourcesSummary(rows []SourceRow) string {
	sep := " · "
	if Plain() {
		sep = " - "
	}
	var off []string
	on := 0
	for _, r := range rows {
		if r.On {
			on++
		} else {
			off = append(off, r.Name)
		}
	}
	noun := "sources"
	if on == 1 {
		noun = "source"
	}
	line := fmt.Sprintf("%d %s on", on, noun)
	if len(off) > 0 {
		line += sep + strings.Join(off, ", ") + " off"
	}
	return line
}

// sourcesPanel is the SOURCES screen's state: a checklist the user walks with
// the arrows and flips with space or enter. Every flip is saved at once, so
// there is no "apply" step to forget.
type sourcesPanel struct {
	theme  *Theme
	load   LoadSourcesFunc
	set    SetSourceFunc
	rows   []SourceRow
	cursor int

	status    string
	statusBad bool
}

func newSourcesPanel(theme *Theme, load LoadSourcesFunc, set SetSourceFunc) *sourcesPanel {
	p := &sourcesPanel{theme: theme, load: load, set: set}
	p.refresh()
	return p
}

// refresh reloads the rows and keeps the cursor on a valid row.
func (p *sourcesPanel) refresh() {
	p.rows = p.load()
	p.cursor = max(min(p.cursor, len(p.rows)-1), 0)
}

// move steps the highlight, stopping at both ends.
func (p *sourcesPanel) move(delta int) {
	p.cursor = max(min(p.cursor+delta, len(p.rows)-1), 0)
}

// onCount is how many sources search right now.
func (p *sourcesPanel) onCount() int {
	n := 0
	for _, r := range p.rows {
		if r.On {
			n++
		}
	}
	return n
}

// toggle flips the highlighted source and saves it. A locked source says who
// holds the lock; a failed save leaves the row as it was and says so.
func (p *sourcesPanel) toggle() {
	if len(p.rows) == 0 {
		return
	}
	row := p.rows[p.cursor]
	if row.Locked {
		p.status, p.statusBad = row.Name+" is "+row.Note+"; unset it to change it here.", true
		return
	}
	if p.set == nil {
		return
	}
	if err := p.set(row.ID, !row.On); err != nil {
		p.status, p.statusBad = "Could not save: "+singleLine(err.Error()), true
		return
	}
	p.refresh()
	state := "off"
	if !row.On {
		state = "on"
	}
	p.status, p.statusBad = row.Name+" "+state+" - saved", false
}

// update handles one key and reports whether the screen should close.
func (p *sourcesPanel) update(msg tea.KeyPressMsg) (closed bool) {
	switch msg.String() {
	case "esc", "ctrl+s", "q":
		return true
	case "up", "k":
		p.move(-1)
	case "down", "j":
		p.move(1)
	case "home", "g":
		p.cursor = 0
	case "end", "G":
		p.cursor = max(len(p.rows)-1, 0)
	case "space", " ", "enter", "x":
		p.toggle()
	}
	return false
}

// hints is the footer while the screen is open.
func (p *sourcesPanel) hints() []Hint {
	move := "↑↓"
	if Plain() {
		move = "up/dn"
	}
	return []Hint{
		{Key: "space", Desc: "on/off", Essential: true},
		{Key: move, Desc: "move"},
		{Key: "esc", Desc: "done", Essential: true},
	}
}

// view draws the heading, the checklist and the status line.
func (p *sourcesPanel) view(width int) string {
	theme := p.theme
	lines := []string{
		theme.Primary.Render("SOURCES"),
		theme.Muted.Render(ansi.Truncate("Choose where GoAnime searches. Changes are saved as you make them.", width, "…")),
		"",
	}
	nameWidth := 0
	for _, r := range p.rows {
		nameWidth = max(nameWidth, lipgloss.Width(singleLine(r.Name)))
	}
	for i, r := range p.rows {
		lines = append(lines, p.row(i, r, nameWidth, width))
	}
	lines = append(lines, "", theme.Muted.Render(fmt.Sprintf("%d of %d on", p.onCount(), len(p.rows))))
	if len(p.rows) > 0 && p.onCount() == 0 {
		lines = append(lines, theme.Primary.Render(ansi.Truncate("No source is on: searches will find nothing.", width, "…")))
	}
	if p.status != "" {
		style := theme.Muted
		if p.statusBad {
			style = theme.Primary
		}
		lines = append(lines, style.Render(ansi.Truncate(p.status, width, "…")))
	}
	return strings.Join(lines, "\n")
}

// row draws one checklist line: pointer, [x]/[ ] box, name, and the note at
// the right. The box is text, so state never depends on colour.
func (p *sourcesPanel) row(i int, r SourceRow, nameWidth, width int) string {
	theme := p.theme
	selected := i == p.cursor
	pointer := "   "
	if selected {
		pointer = drawGlyph(" ▶ ", "> ")
	}
	box := "[ ]"
	if r.On {
		box = "[x]"
	}
	name := singleLine(r.Name)
	left := pointer + box + " " + name + strings.Repeat(" ", nameWidth-lipgloss.Width(name))
	note := singleLine(r.Note)
	if room := width - lipgloss.Width(left) - 3; note != "" && room > 0 {
		note = ansi.Truncate(note, room, "…")
	} else {
		note = ""
	}
	line := left
	if note != "" {
		line += "  " + note
	}
	line = ansi.Truncate(line, width, "")
	switch {
	case selected:
		return theme.Selected.Render(line + strings.Repeat(" ", max(width-lipgloss.Width(line), 0)))
	case r.Locked || !r.On:
		return theme.Muted.Render(line)
	default:
		return theme.Text.Render(line)
	}
}
