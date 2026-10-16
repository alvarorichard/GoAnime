package tui

import (
	"fmt"
	"io"
	"strings"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// minTitleWidth is the room a row keeps for its title before it gives up
// the text at its right, then its badges.
const minTitleWidth = 12

// row is what one list row shows. Result and picker entries both provide it.
type row struct {
	number string     // before the title, such as "01"
	status PickStatus // the mark before the title
	title  string
	desc   string // under the title on two-line rows, at the right otherwise
	tags   []string
	note   string  // at the right, in place of desc
	key    string  // at the far right
	meter  float64 // 0–1, a bar at the right in place of desc; 0 draws none
}

// rowItem is a list item that can be drawn by rowDelegate.
type rowItem interface {
	list.Item
	row() row
}

// rowDelegate draws list rows the way the site draws its selection: the
// cursor row filled in crimson from edge to edge, numbers and marks in a
// muted gutter, badges inverted against the ground.
type rowDelegate struct {
	theme   *Theme
	twoLine bool // description under the title instead of at the right
	spacing int
	marks   bool // keep a column for status marks
}

func (d rowDelegate) Height() int {
	if d.twoLine {
		return 2
	}
	return 1
}

func (d rowDelegate) Spacing() int { return d.spacing }

func (d rowDelegate) Update(tea.Msg, *list.Model) tea.Cmd { return nil }

// rowStyles are the styles of one row in one state.
type rowStyles struct {
	cursor                      string
	fill, title, desc, gutter   lipgloss.Style
	mark, match, note, key, tag lipgloss.Style
	meter, meterEmpty           lipgloss.Style
}

func (d rowDelegate) styles(selected, dimmed bool, status PickStatus) rowStyles {
	t := d.theme
	if selected {
		cursor := " ▶ "
		if Plain() {
			// Tela 9: without color the cursor is ">" with reverse
			// video, never a color fill.
			cursor = "> "
		}
		return rowStyles{
			cursor: cursor,
			fill:   t.Selected.UnsetBold(), title: t.Selected, desc: t.SelectedMuted, gutter: t.SelectedMuted,
			mark: t.Selected, match: t.Selected.Underline(true), note: t.SelectedMuted, key: t.Selected,
			tag: t.Tag, meter: t.Selected, meterEmpty: t.SelectedFaint,
		}
	}
	s := rowStyles{
		cursor: "   ",
		fill:   lipgloss.NewStyle(), title: t.Text, desc: t.Muted, gutter: t.Muted,
		mark: t.Muted, match: t.FilterMatch, note: t.Muted, key: t.Muted,
		tag: t.Tag, meter: t.Meter, meterEmpty: t.MeterEmpty,
	}
	switch status {
	case PickWatched:
		s.title = t.Muted
	case PickInProgress, PickCurrent:
		s.mark, s.note = t.Primary, t.Primary.UnsetBold()
	}
	if dimmed {
		s.title, s.desc, s.gutter = t.Muted, t.Faint, t.Faint
	}
	return s
}

// Render draws the item at index, at the list's full width.
//
//nolint:gocritic // hugeParam: list.ItemDelegate fixes the signature.
func (d rowDelegate) Render(w io.Writer, m list.Model, index int, item list.Item) {
	it, ok := item.(rowItem)
	width := m.Width()
	if !ok || width <= 0 {
		return
	}
	r := it.row()
	st := d.styles(index == m.Index(), m.FilterState() == list.Filtering && m.FilterValue() == "", r.status)
	var matches []int
	if m.FilterState() != list.Unfiltered {
		matches = m.MatchesForItem(index)
	}
	if d.twoLine {
		_, _ = fmt.Fprint(w, d.renderTwoLine(r, &st, matches, width))
		return
	}
	_, _ = fmt.Fprint(w, d.renderOneLine(r, &st, matches, width))
}

// gutter draws the cursor, the status mark and the number before a title.
func (d rowDelegate) gutter(r row, st *rowStyles) string {
	out := st.fill.Render(st.cursor)
	if d.marks {
		out += st.mark.Render(statusMark(r.status)) + st.fill.Render(" ")
	}
	if r.number != "" {
		out += st.gutter.Render(r.number) + st.fill.Render("  ")
	}
	return out
}

func (d rowDelegate) renderOneLine(r row, st *rowStyles, matches []int, width int) string {
	left := d.gutter(r, st)
	tags := renderTags(r.tags, st)

	side := ""
	switch {
	case r.meter > 0:
		side = renderBar(max(min(width/4, 30), 8), r.meter, &st.meter, &st.meterEmpty)
	case r.note != "":
		side = st.note.Render(singleLine(r.note))
	case r.desc != "":
		side = st.desc.Render(r.desc)
	}
	right := func() string {
		out := side
		if r.key != "" {
			if out != "" {
				out += st.fill.Render("   ")
			}
			out += st.key.Render(r.key)
		}
		if out != "" {
			out += st.fill.Render(" ")
		}
		return out
	}
	room := func() int {
		return width - lipgloss.Width(left) - lipgloss.Width(tags) - lipgloss.Width(right()) - 2
	}
	if room() < minTitleWidth {
		side = ""
	}
	if room() < minTitleWidth {
		tags = ""
	}

	line := left + renderTitle(r.title, max(room(), 1), st, matches) + tags
	end := right()
	pad := width - lipgloss.Width(line) - lipgloss.Width(end)
	return ansi.Truncate(line+st.fill.Render(strings.Repeat(" ", max(pad, 0)))+end, width, "")
}

func (d rowDelegate) renderTwoLine(r row, st *rowStyles, matches []int, width int) string {
	left := d.gutter(r, st)
	tags := renderTags(r.tags, st)
	room := width - lipgloss.Width(left) - lipgloss.Width(tags) - 1
	if room < minTitleWidth {
		tags = ""
		room = width - lipgloss.Width(left) - 1
	}
	first := left + renderTitle(r.title, max(room, 1), st, matches) + tags

	indent := lipgloss.Width(left)
	desc := ansi.Truncate(r.desc, max(width-indent-1, 1), "…")
	second := st.fill.Render(strings.Repeat(" ", indent)) + st.desc.Render(desc)

	pad := func(line string) string {
		return ansi.Truncate(line+st.fill.Render(strings.Repeat(" ", max(width-lipgloss.Width(line), 0))), width, "")
	}
	return pad(first) + "\n" + pad(second)
}

// renderTitle truncates a title to room cells and underlines the characters
// the filter matched.
func renderTitle(title string, room int, st *rowStyles, matches []int) string {
	title = ansi.Truncate(title, room, "…")
	if len(matches) == 0 {
		return st.title.Render(title)
	}
	return lipgloss.StyleRunes(title, matches, st.match, st.title)
}

func renderTags(tags []string, st *rowStyles) string {
	out := ""
	for _, tag := range tags {
		if tag = singleLine(tag); tag != "" {
			out += st.fill.Render(" ") + st.tag.Render(strings.ToUpper(tag))
		}
	}
	return out
}

// styleList applies the theme to everything a list draws around its rows:
// the filter line, the status line and the page dots.
func styleList(l *list.Model, theme *Theme, placeholder string) {
	l.SetShowTitle(false)
	l.SetShowHelp(false)

	prompt := "❯ "
	if Plain() {
		prompt = "> "
	}
	l.FilterInput.Prompt = prompt
	l.FilterInput.Placeholder = placeholder
	input := l.FilterInput.Styles()
	input.Focused.Prompt = theme.Primary
	input.Blurred.Prompt = theme.Primary
	input.Focused.Text = theme.Text
	input.Blurred.Text = theme.Text
	input.Focused.Placeholder = theme.Muted
	input.Blurred.Placeholder = theme.Muted
	input.Cursor.Color = theme.Primary.GetForeground()
	l.FilterInput.SetStyles(input)

	l.Styles.TitleBar = lipgloss.NewStyle().Padding(0, 0, 1, 1)
	l.Styles.StatusBar = theme.Muted.Padding(0, 0, 1, 1)
	l.Styles.StatusEmpty = theme.Muted
	l.Styles.StatusBarActiveFilter = theme.Primary
	l.Styles.StatusBarFilterCount = theme.Muted
	l.Styles.DividerDot = theme.Faint.SetString(" · ")
	l.Styles.NoItems = theme.Muted.PaddingLeft(1)
	l.Styles.PaginationStyle = lipgloss.NewStyle().PaddingLeft(3)
	l.Styles.ArabicPagination = theme.Muted
	l.Styles.ActivePaginationDot = theme.Primary.SetString("●")
	l.Styles.InactivePaginationDot = theme.Faint.SetString("○")
	if Plain() {
		l.Styles.ActivePaginationDot = theme.Primary.SetString("*")
		l.Styles.InactivePaginationDot = theme.Faint.SetString("-")
	}
	l.Paginator.ActiveDot = l.Styles.ActivePaginationDot.String()
	l.Paginator.InactiveDot = l.Styles.InactivePaginationDot.String()
}

// statusMark is the one-cell mark drawn before a title. In plain mode the
// marks become ASCII (tela 9): x for watched, ~ for in progress, * for
// current, so screen readers and dumb terminals read them as text.
func statusMark(status PickStatus) string {
	if Plain() {
		switch status {
		case PickWatched:
			return "x"
		case PickInProgress:
			return "~"
		case PickCurrent:
			return "*"
		default:
			return " "
		}
	}
	switch status {
	case PickWatched:
		return "✓"
	case PickInProgress:
		return "◐"
	case PickCurrent:
		return "●"
	default:
		return " "
	}
}
