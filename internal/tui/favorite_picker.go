package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// newFavoriteLists opens the saved lists the result screen offers on ctrl+l.
// When it returns nil the whole feature is absent: no key, no hint, no tip. A
// variable so tests install fakes and the app installs the real store.
var newFavoriteLists = func() FavoriteLists { return nil }

const (
	// maxListNameRunes is how long a list's name can be typed.
	maxListNameRunes = 40
	// pickerStatusLines is the room the status line has under the rows: a
	// sentence wraps to this many lines before it is cut.
	pickerStatusLines = 2
	// pickerWrapSlack is the cells a wrapped line may lose to a word break.
	pickerWrapSlack = 8
	// pickerNamePlaceholder shows what a name could look like.
	pickerNamePlaceholder = "e.g. Weekend marathon"
)

// favoritePicker is the "Add to a list" overlay of the result screen: the
// lists to put the highlighted work in, and the entry that names a new one.
//
// It reads the lists once when it opens and keeps its copy in step with every
// change it makes, so drawing it never touches the saved file.
type favoritePicker struct {
	theme *Theme
	lists FavoriteLists
	work  HomeFavorite
	// title is the work as the screen names it, for the heading and the status.
	title string

	items   []HomeList
	member  map[string]bool // list ID -> the work is in it
	loadErr error           // Lists() failed: there is nothing to pick from
	// favKeys is the result screen's set of FAVORITES keys, kept in step when
	// the Favorites list changes so the row's star follows.
	favKeys map[string]bool

	cursor int // 0 is "+ Create a new list", then one row per list
	offset int // the first row drawn
	room   int // how many rows fit; 0 until the screen has been measured

	status pickerStatus

	naming  bool
	input   textinput.Model
	nameErr string
	// created remembers a list that was made but could not take the work, so
	// pressing enter again adds the work to it instead of making a second one.
	createdID, createdName string
}

// pickerStatus is the sentence under the rows. The work's title sits between
// prefix and suffix so a long title is what gets shortened.
type pickerStatus struct {
	prefix, work, suffix string
	failed               bool
}

func (s pickerStatus) text(titleRoom int) string {
	if s.work == "" || titleRoom <= 0 {
		return s.prefix + s.work + s.suffix
	}
	return s.prefix + ansi.Truncate(s.work, titleRoom, "…") + s.suffix
}

// newFavoritePicker opens the picker for work, named title on screen. It reads
// the lists and the work's membership once.
func newFavoritePicker(theme *Theme, lists FavoriteLists, work HomeFavorite, title string, favKeys map[string]bool) *favoritePicker {
	p := &favoritePicker{
		theme:   theme,
		lists:   lists,
		work:    work,
		title:   title,
		member:  map[string]bool{},
		favKeys: favKeys,
		input:   newNameInput(theme),
	}
	p.items, p.loadErr = lists.Lists()
	if p.loadErr != nil {
		p.items = nil
		return p
	}
	for id, in := range lists.Membership(work.Key) {
		if in {
			p.member[id] = true
		}
	}
	return p
}

// newNameInput is the text entry that names a new list.
func newNameInput(theme *Theme) textinput.Model {
	input := textinput.New()
	input.Prompt = drawGlyph("❯ ", "> ")
	input.Placeholder = pickerNamePlaceholder
	input.CharLimit = maxListNameRunes
	styles := textinput.DefaultStyles(themeIsDark())
	styles.Focused.Text = theme.Text
	styles.Focused.Placeholder = theme.Muted
	styles.Focused.Prompt = theme.Primary
	styles.Blurred = styles.Focused
	styles.Cursor.Color = theme.Primary.GetForeground()
	styles.Cursor.Shape = tea.CursorBlock
	input.SetStyles(styles)
	return input
}

// rowCount is how many rows the picker offers: the create row and the lists.
func (p *favoritePicker) rowCount() int {
	if p.loadErr != nil {
		return 0
	}
	return 1 + len(p.items)
}

// pickerLayout is how the picker's height is shared out.
type pickerLayout struct {
	title       bool
	spacers     bool // a line above and below the rows, for the "more" notes
	statusLines int
	rows        int
}

// layoutFor shares height lines between the heading, the rows and the status:
// the rows shrink last, the spacers go first.
func layoutFor(height int) pickerLayout {
	switch {
	case height >= 9:
		return pickerLayout{title: true, spacers: true, statusLines: pickerStatusLines, rows: height - 3 - pickerStatusLines}
	case height >= 5:
		return pickerLayout{title: true, statusLines: pickerStatusLines, rows: height - 1 - pickerStatusLines}
	case height >= 3:
		return pickerLayout{title: true, statusLines: 1, rows: height - 2}
	case height == 2:
		return pickerLayout{title: true, rows: 1}
	}
	return pickerLayout{rows: 1}
}

// windowStart is the first row drawn so the cursor row is on screen, moving
// the window as little as possible.
func windowStart(offset, cursor, total, rows int) int {
	rows = max(rows, 1)
	if cursor < offset {
		offset = cursor
	}
	if cursor >= offset+rows {
		offset = cursor - rows + 1
	}
	return max(min(offset, total-rows), 0)
}

// fit measures the picker for a body width by height and keeps the cursor row
// on screen. It is called when the picker opens and on every resize.
func (p *favoritePicker) fit(width, height int) {
	p.room = layoutFor(height).rows
	p.offset = windowStart(p.offset, p.cursor, p.rowCount(), p.room)
	p.input.SetWidth(max(min(width-lipgloss.Width(p.input.Prompt)-2, maxListNameRunes+1), 8))
}

// handleKey acts on a key press and reports whether the picker should close.
func (p *favoritePicker) handleKey(msg tea.KeyPressMsg) (tea.Cmd, bool) {
	if p.naming {
		return p.nameKey(msg), false
	}
	switch msg.String() {
	case "esc", "q":
		return nil, true
	case "enter":
		return p.choose(), false
	case "up":
		p.move(-1)
	case "down":
		p.move(1)
	case "pgup":
		p.move(-max(p.room-1, 1))
	case "pgdown":
		p.move(max(p.room-1, 1))
	case "home":
		p.move(-p.rowCount())
	case "end":
		p.move(p.rowCount())
	}
	return nil, false
}

// move shifts the cursor by delta rows, stopping at the first and last.
func (p *favoritePicker) move(delta int) {
	if p.rowCount() == 0 {
		return
	}
	p.cursor = max(min(p.cursor+delta, p.rowCount()-1), 0)
	p.offset = windowStart(p.offset, p.cursor, p.rowCount(), p.room)
}

// choose is enter on the highlighted row.
func (p *favoritePicker) choose() tea.Cmd {
	if p.rowCount() == 0 {
		return nil
	}
	if p.cursor == 0 {
		return p.startNaming()
	}
	p.toggle(p.cursor - 1)
	return nil
}

// toggle puts the work in the list at index, or takes it off, and says so.
func (p *favoritePicker) toggle(index int) {
	list := p.items[index]
	on, err := p.lists.ToggleWork(list.ID, p.work)
	if err != nil {
		p.status = pickerStatus{prefix: fmt.Sprintf("Could not change %s. %s", quoted(list.Name), friendly(err)), failed: true}
		return
	}
	p.applyToggle(index, on)
	if on {
		p.status = pickerStatus{prefix: `Added "`, work: p.title, suffix: `" to ` + singleLine(list.Name) + "."}
		return
	}
	p.status = pickerStatus{prefix: `Removed "`, work: p.title, suffix: `" from ` + singleLine(list.Name) + "."}
}

// applyToggle records that the work is now in the list at index (or not): the
// membership, the list's count and, for Favorites, the result row's star.
func (p *favoritePicker) applyToggle(index int, on bool) {
	list := &p.items[index]
	was := p.member[list.ID]
	switch {
	case on && !was:
		list.Count++
	case !on && was:
		list.Count = max(list.Count-1, 0)
	}
	if on {
		p.member[list.ID] = true
	} else {
		delete(p.member, list.ID)
	}
	if list.ID != FavoritesListID || p.favKeys == nil {
		return
	}
	if on {
		p.favKeys[p.work.Key] = true
	} else {
		delete(p.favKeys, p.work.Key)
	}
}

// startNaming opens the name entry.
func (p *favoritePicker) startNaming() tea.Cmd {
	p.naming = true
	p.status = pickerStatus{}
	return p.input.Focus()
}

// nameKey handles a key press in the name entry.
func (p *favoritePicker) nameKey(msg tea.KeyPressMsg) tea.Cmd {
	switch msg.String() {
	case "esc":
		p.stopNaming()
		return nil
	case "enter":
		p.submitName()
		return nil
	}
	return p.forward(msg)
}

// forward hands a message that is not a key press (a pasted text, the
// cursor's blink) to the name entry.
func (p *favoritePicker) forward(msg tea.Msg) tea.Cmd {
	before := p.input.Value()
	var cmd tea.Cmd
	p.input, cmd = p.input.Update(msg)
	if p.input.Value() != before {
		p.nameErr = ""
	}
	return cmd
}

// stopNaming goes back to the lists, creating nothing more. The entry is left
// empty for the next time.
func (p *favoritePicker) stopNaming() {
	p.naming = false
	p.nameErr = ""
	p.createdID, p.createdName = "", ""
	p.input.Reset()
	p.input.Blur()
}

// submitName makes the list and puts the work in it. A failure stays in the
// name entry, with the name as typed.
func (p *favoritePicker) submitName() {
	name := p.input.Value()
	index := p.indexOf(p.createdID)
	if p.createdID == "" || name != p.createdName || index < 0 {
		list, err := p.lists.Create(name)
		if err != nil {
			p.nameErr = friendly(err)
			return
		}
		p.items = append(p.items, list)
		index = len(p.items) - 1
		p.createdID, p.createdName = list.ID, name
	}
	list := p.items[index]
	on, err := p.lists.ToggleWork(list.ID, p.work)
	if err != nil {
		p.nameErr = fmt.Sprintf("Made the list %s, but could not add %s. %s", quoted(list.Name), quoted(p.title), friendly(err))
		return
	}
	p.applyToggle(index, on)
	p.stopNaming()
	p.cursor = 1 + index
	p.offset = windowStart(p.offset, p.cursor, p.rowCount(), p.room)
	if !on {
		p.status = pickerStatus{prefix: "Created " + quoted(list.Name) + ".", suffix: ""}
		return
	}
	p.status = pickerStatus{prefix: "Created " + quoted(list.Name) + ` and added "`, work: p.title, suffix: `".`}
}

// indexOf is the position of the list with id in items, or -1.
func (p *favoritePicker) indexOf(id string) int {
	for i := range p.items {
		if id != "" && p.items[i].ID == id {
			return i
		}
	}
	return -1
}

// quoted is a name between quotes, on one line.
func quoted(name string) string {
	return `"` + singleLine(name) + `"`
}

// hints are the footer's keys while the picker is open.
func (p *favoritePicker) hints() []Hint {
	if p.naming {
		return []Hint{
			{Key: "enter", Desc: "create", Essential: true},
			{Key: "esc", Desc: "back", Essential: true},
		}
	}
	if p.rowCount() == 0 {
		return []Hint{{Key: "esc", Desc: "close", Essential: true}}
	}
	move := drawGlyph("↑↓", "up/dn")
	return []Hint{
		{Key: "enter", Desc: p.enterWord(), Essential: true},
		{Key: move, Desc: "move"},
		{Key: "esc", Desc: "close", Essential: true},
	}
}

// enterWord is what enter will do on the highlighted row.
func (p *favoritePicker) enterWord() string {
	switch {
	case p.cursor == 0:
		return "new list"
	case p.member[p.items[p.cursor-1].ID]:
		return "take off"
	}
	return "add"
}

// view draws the picker in a width by height body.
func (p *favoritePicker) view(width, height int) string {
	if p.naming {
		return p.nameView(width, height)
	}
	layout := layoutFor(height)
	var lines []string
	if layout.title {
		lines = append(lines, p.heading(width))
	}
	if p.loadErr != nil {
		text := "Could not read your lists. " + friendly(p.loadErr)
		notice := strings.TrimRight(renderNotice(p.theme, text, width-pickerIndent), "\n")
		lines = append(lines, "", indented(strings.ReplaceAll(notice, "\n", "\n"+indented(""))), "", indented(p.theme.Muted.Render("esc closes this.")))
		return strings.Join(lines, "\n")
	}
	start := windowStart(p.offset, p.cursor, p.rowCount(), layout.rows)
	end := min(start+layout.rows, p.rowCount())
	if layout.spacers {
		lines = append(lines, p.more(start > 0, aboveNote(start), width))
	}
	for i := start; i < end; i++ {
		lines = append(lines, p.renderRow(i, width))
	}
	if layout.spacers {
		lines = append(lines, p.more(end < p.rowCount(), belowNote(p.rowCount()-end), width))
	}
	lines = append(lines, p.statusLines(width, layout.statusLines)...)
	return strings.Join(lines, "\n")
}

func aboveNote(n int) string { return drawGlyph("↑ ", "^ ") + fmt.Sprintf("%d more above", n) }
func belowNote(n int) string { return drawGlyph("↓ ", "v ") + fmt.Sprintf("%d more below", n) }

// more is the line above or below the rows: a note when rows are out of sight
// that way, empty otherwise.
func (p *favoritePicker) more(show bool, note string, width int) string {
	if !show {
		return ""
	}
	return p.theme.Muted.Render(ansi.Truncate("   "+note, width, "…"))
}

// heading is the first line: what is being put in a list.
func (p *favoritePicker) heading(width int) string {
	const lead, tail = `Add "`, `" to a list`
	room := width - pickerIndent - lipgloss.Width(lead) - lipgloss.Width(tail)
	title := singleLine(p.title)
	text := lead + title + tail
	if room >= 1 {
		text = lead + ansi.Truncate(title, room, "…") + tail
	}
	return indented(p.theme.Text.Bold(true).Render(ansi.Truncate(text, width-pickerIndent, "…")))
}

// pickerIndent is the margin the picker's sentences keep, like the result
// list's.
const pickerIndent = 1

// indented puts the margin before a line.
func indented(line string) string {
	return strings.Repeat(" ", pickerIndent) + line
}

// statusLines is the sentence about what just happened, wrapped to at most
// lines lines.
func (p *favoritePicker) statusLines(width, lines int) []string {
	if lines == 0 {
		return nil
	}
	if p.status.prefix == "" && p.status.work == "" {
		return make([]string, lines)
	}
	// Room for the work's title: what the sentence leaves of the lines, less
	// some slack for the words that wrap.
	room := width*lines - lipgloss.Width(p.status.prefix) - lipgloss.Width(p.status.suffix) - lines*pickerWrapSlack
	text := p.status.text(max(room, minTitleWidth))
	var out []string
	width = max(width-pickerIndent, 1)
	if p.status.failed {
		out = strings.Split(strings.TrimRight(renderNotice(p.theme, text, width), "\n"), "\n")
	} else {
		wrapped := strings.Split(ansi.Wrap(text, width, ""), "\n")
		if len(wrapped) > lines {
			wrapped = append(wrapped[:lines-1], ansi.Truncate(strings.Join(wrapped[lines-1:], " "), width, "…"))
		}
		for _, line := range wrapped {
			out = append(out, p.theme.Primary.Render(line))
		}
	}
	for i := range out {
		out[i] = indented(out[i])
	}
	if len(out) > lines {
		out = out[:lines]
	}
	return out
}

// pickerRow is what one row says.
type pickerRow struct {
	create bool
	star   bool
	name   string
	count  string
	in     bool
}

// rowFor describes row i of the picker.
func (p *favoritePicker) rowFor(i int) pickerRow {
	if i == 0 {
		return pickerRow{create: true, name: "Create a new list"}
	}
	list := p.items[i-1]
	return pickerRow{
		star:  list.ID == FavoritesListID,
		name:  list.Name,
		count: workCount(list.Count),
		in:    p.member[list.ID],
	}
}

// workCount is "12 works", "1 work" or "empty".
func workCount(n int) string {
	switch {
	case n <= 0:
		return "empty"
	case n == 1:
		return "1 work"
	}
	return fmt.Sprintf("%d works", n)
}

func (p *favoritePicker) renderRow(i, width int) string {
	return renderPickerRow(p.theme, p.rowFor(i), i == p.cursor, width)
}

// renderPickerRow draws one row like the home screen's rows: the cursor, the
// name, then the count and whether the work is in the list at the right. That
// it is in the list is said in words (and "[x]" in plain mode), never by
// colour alone.
func renderPickerRow(theme *Theme, r pickerRow, selected bool, width int) string {
	body, side, mark := theme.Text, theme.Muted, theme.Primary
	if selected {
		body, side, mark = theme.Selected, theme.SelectedMuted, theme.Selected
	}
	cursor := drawGlyph(" ▶ ", " > ")
	if !selected {
		cursor = "   "
	}
	left := cursor
	switch {
	case r.create:
		left += mark.Render("+") + body.Render(" "+r.name)
	default:
		if Plain() {
			box := "[ ] "
			if r.in {
				box = "[x] "
			}
			left += body.Render(box)
		}
		star := "  "
		if r.star {
			star = drawGlyph("★ ", "* ")
		}
		left += mark.Render(star) + body.Render(r.name)
	}
	fill := func(n int) string {
		if selected {
			return theme.Selected.UnsetBold().Render(strings.Repeat(" ", max(n, 0)))
		}
		return strings.Repeat(" ", max(n, 0))
	}
	right := pickerRowSide(&side, &mark, r, width-lipgloss.Width(cursor)-minTitleWidth-1)
	if room := width - lipgloss.Width(left) - lipgloss.Width(right) - 1; room < 1 {
		left = ansi.Truncate(left, max(width-lipgloss.Width(right)-1, 1), "…")
	}
	line := left + fill(width-lipgloss.Width(left)-lipgloss.Width(right)) + right
	if selected {
		line += fill(width - lipgloss.Width(line))
	}
	return ansi.Truncate(line, width, "")
}

// pickerRowSide is the right part of a row: the count and the membership
// note, the note cut to a bare tick (or nothing in plain mode, where "[x]"
// already says it) when the row is narrow.
func pickerRowSide(side, mark *lipgloss.Style, r pickerRow, room int) string {
	if r.create {
		return ""
	}
	full := side.Render(r.count)
	if r.in {
		note := drawGlyph("✓ in this list", "in this list")
		if lipgloss.Width(full)+3+lipgloss.Width(note) > room {
			note = drawGlyph("✓", "")
		}
		if note != "" {
			full += side.Render("   ") + mark.Render(note)
		}
	}
	return full
}

// nameView draws the name entry.
func (p *favoritePicker) nameView(width, height int) string {
	lines := []string{
		p.theme.Text.Bold(true).Render("Name your list"),
		p.theme.Muted.Render(ansi.Truncate(`It will start with `+quoted(p.title)+` in it.`, width-pickerIndent, "…")),
	}
	roomy := height >= 7
	if roomy {
		lines = append(lines, "")
	}
	lines = append(lines, p.input.View())
	if p.nameErr != "" {
		lines = append(lines, strings.TrimRight(renderNotice(p.theme, p.nameErr, width-pickerIndent), "\n"))
	} else {
		lines = append(lines, "")
	}
	if roomy {
		lines = append(lines, "")
	}
	lines = append(lines, p.theme.Muted.Render("enter create "+drawGlyph("·", "-")+" esc back"))
	for i := range lines {
		lines[i] = ansi.Truncate(indented(lines[i]), width, "")
	}
	return strings.Join(lines, "\n")
}

// openPicker opens the "Add to a list" overlay for the highlighted work. It
// reports false when there is nothing to open it for.
func (m *animeResultsModel) openPicker() bool {
	item, ok := m.results.SelectedItem().(animeGroupItem)
	if !ok || item.group == nil || m.lists == nil {
		return false
	}
	entry := m.favoriteEntry(item)
	work := HomeFavorite{Key: entry.Key, Title: entry.Title, Year: entry.Year, Media: entry.Media, Source: entry.Source}
	m.picker = newFavoritePicker(&m.theme, m.lists, work, item.Title(), m.favKeys)
	m.fitPicker()
	return true
}

// fitPicker re-measures the open picker for the screen's size.
func (m *animeResultsModel) fitPicker() {
	if m.picker == nil {
		return
	}
	width, height := m.shell.ContentSize()
	m.picker.fit(width, max(height-m.noticeHeight(width), 1))
}

// updatePicker routes a key press to the open picker. ctrl+c quits like it
// does everywhere on the screen.
func (m *animeResultsModel) updatePicker(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		m.err = ErrSelectionCancelled
		return m, tea.Quit
	}
	cmd, closed := m.picker.handleKey(msg)
	if closed {
		m.picker = nil
	}
	// The picker may have starred or unstarred the work in FAVORITES.
	m.reservePlainPanel()
	return m, cmd
}
