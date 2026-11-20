package tui

import (
	"fmt"
	"slices"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

// The FAVORITES tab with Options.Lists set has two levels. The first is the
// user's lists (Favorites, then the ones they made) with "+ Create a new
// list" on top; enter on a list opens the second, which is the works of that
// list and behaves like the single Favorites list always did. esc goes back.
//
// Everything here is reached from search_home.go through a few one-line
// hooks, and every hook does nothing while Options.Lists is nil, so the tab
// is exactly what it was without lists.

const (
	// homeListNameLimit is how many characters a list's name can have.
	homeListNameLimit = 40
	// homeListPanelTitles is how many titles the details panel names before it
	// says "+N more".
	homeListPanelTitles = 8
	// homeListDateLayout is how a list's creation date is written.
	homeListDateLayout = "2006-01-02"

	homeListsPlaceholder = "type to filter your lists, enter opens the first one"
	homeWorksPlaceholder = "type to filter this list, enter opens the first one"
	homeListCreateTitle  = "+ Create a new list"
	homeListPanelTip     = "enter opens it"
	// homeListMax is how many lists of their own a user can have.
	homeListMax = 50
	// homeListNoteWidth is the narrowest terminal with room to say on the
	// Favorites row what that list is.
	homeListNoteWidth = 70
	// homeListFavoritesNote says what the Favorites list is.
	homeListFavoritesNote = "what you save with ctrl+f"
)

type homeListsLevel int

const (
	homeLevelLists homeListsLevel = iota // the user's lists
	homeLevelWorks                       // the works of one list
)

type homeListEntry int

const (
	homeEntryNone homeListEntry = iota
	homeEntryCreate
	homeEntryRename
)

// homeLists is the state of the FAVORITES tab's lists.
type homeLists struct {
	api   FavoriteLists
	level homeListsLevel
	all   []HomeList // the lists, as Lists() last answered
	// loadErr is the sentence for a failed Lists(), drawn in the tab while it
	// lasts; a failed Works() is a notice instead.
	loadErr string

	open  HomeList       // the list whose works show, at homeLevelWorks
	works []HomeFavorite // its works

	// The inline name entry for a new or renamed list.
	name     textinput.Model
	entry    homeListEntry
	entryID  string // the list being renamed
	entryErr string // why the last enter was refused
	wasFocus bool   // whether the box had the focus before the entry opened

	doomed HomeList // the list the delete question is about

	// removed is the title last taken off the open list, which u puts back.
	removed *HomeFavorite
}

// newHomeLists is the state for the host's lists, or nil when it gave none.
func newHomeLists(api FavoriteLists, styles *textinput.Styles) *homeLists {
	if api == nil {
		return nil
	}
	name := textinput.New()
	name.Prompt = "Name your list: "
	name.Placeholder = "e.g. Weekend marathon"
	name.CharLimit = homeListNameLimit
	name.SetStyles(*styles)
	return &homeLists{api: api, name: name}
}

// reload asks for the lists again. A failure keeps what was known and puts
// the reason in loadErr.
func (h *homeLists) reload() {
	all, err := h.api.Lists()
	if err != nil {
		h.loadErr = homeListsReadFailed(err)
		return
	}
	h.all = slices.Clone(all)
	h.loadErr = ""
}

// homeListsReadFailed is the sentence for a list that could not be read.
func homeListsReadFailed(err error) string {
	return "Could not read your lists. " + friendly(err)
}

// listsActive reports whether the FAVORITES tab is showing and has lists.
func (m *searchHomeModel) listsActive() bool {
	return m.lists != nil && m.onFavorites()
}

// listsLevelOne reports whether the tab shows the user's lists.
func (m *searchHomeModel) listsLevelOne() bool {
	return m.listsActive() && m.lists.level == homeLevelLists
}

// listsLevelTwo reports whether the tab shows the works of one list.
func (m *searchHomeModel) listsLevelTwo() bool {
	return m.listsActive() && m.lists.level == homeLevelWorks
}

// listsOnScope puts the tab right when the FAVORITES tab opens: always at the
// first level, with the lists read afresh.
func (m *searchHomeModel) listsOnScope() {
	if !m.listsActive() {
		return
	}
	m.lists.level = homeLevelLists
	m.lists.entry = homeEntryNone
	m.lists.reload()
	m.input.Placeholder = homeListsPlaceholder
}

// tabFavorites is the works the tab's rows index: the open list's, or the
// Favorites list the screen was given.
func (m *searchHomeModel) tabFavorites() []HomeFavorite {
	if m.listsLevelTwo() {
		return m.lists.works
	}
	return m.favoriteAll
}

// tabRows is the rows the box leaves on the tab.
func (m *searchHomeModel) tabRows() []homeRow {
	if m.listsLevelOne() {
		return m.listRows()
	}
	matches := m.matchingFavorites()
	rows := make([]homeRow, 0, len(matches))
	for _, i := range matches {
		rows = append(rows, homeRow{kind: homeRowFavorite, index: i})
	}
	return rows
}

// homeListsHasAllTerms reports whether hay holds every one of terms.
func homeListsHasAllTerms(hay string, terms []string) bool {
	for _, term := range terms {
		if !strings.Contains(hay, term) {
			return false
		}
	}
	return true
}

// listRows is "+ Create a new list" while the box is empty, then the lists
// whose name holds every word typed.
func (m *searchHomeModel) listRows() []homeRow {
	h := m.lists
	terms := strings.Fields(foldForMatch(m.input.Value()))
	var rows []homeRow
	if len(terms) == 0 && h.loadErr == "" {
		rows = append(rows, homeRow{kind: homeRowNewList})
	}
	for i := range h.all {
		if homeListsHasAllTerms(foldForMatch(singleLine(h.all[i].Name)), terms) {
			rows = append(rows, homeRow{kind: homeRowList, index: i})
		}
	}
	return rows
}

// ---- words on the tab ---------------------------------------------------------

// tabLabel is the tab's heading: FAVORITES, or the name of the open list cut
// to the width of the screen (a name of 40 wide characters is 80 cells).
func (m *searchHomeModel) tabLabel(width int) string {
	if m.listsLevelTwo() {
		return ansi.Truncate(singleLine(m.lists.open.Name), max(width-12, 12), "…")
	}
	return "FAVORITES"
}

// tabCount is the muted count after the heading: "12", or "3 of 12" while the
// box hides some.
func (m *searchHomeModel) tabCount(shown, total int) string {
	if shown == total {
		return m.theme.Muted.Render(fmt.Sprintf("  %d", total))
	}
	return m.theme.Muted.Render(fmt.Sprintf("  %d of %d", shown, total))
}

// tabEmptyText is what an empty tab says.
func (m *searchHomeModel) tabEmptyText() string {
	if !m.listsLevelTwo() {
		return "You have no favorites yet. Search for a title, then press ctrl+f on it to add it here."
	}
	key := "ctrl+l"
	if m.lists.open.Auto {
		key = "ctrl+f"
	}
	return "This list is empty. Search for a title, then press " + key + " on it to add it here."
}

// tabNoMatchText is what the tab says when the box matches nothing.
func (m *searchHomeModel) tabNoMatchText() string {
	what := "favorite"
	if m.listsLevelTwo() {
		what = "title"
	}
	return fmt.Sprintf("No %s matches %q. Delete a letter to see them all.", what, strings.TrimSpace(m.input.Value()))
}

// homeListTitleCount is "12 titles", "1 title" or "empty".
func homeListTitleCount(n int) string {
	if n <= 0 {
		return "empty"
	}
	return countOf(n, "title", "titles")
}

// listItem is the row of a list, or of "+ Create a new list".
func (m *searchHomeModel) listItem(row homeRow, width int) homeRowItem {
	if row.kind == homeRowNewList {
		if m.listsFull() {
			return homeRowItem{title: homeListCreateTitle, note: fmt.Sprintf("you have %d, the most there can be", homeListMax)}
		}
		return homeRowItem{title: homeListCreateTitle}
	}
	l := m.lists.all[row.index]
	item := homeRowItem{title: singleLine(l.Name), note: homeListTitleCount(l.Count)}
	if l.Auto {
		item.title = drawGlyph("★", "*") + " " + item.title
		if width >= homeListNoteWidth {
			item.key = homeListFavoritesNote
		}
	} else if !l.Created.IsZero() {
		item.key = "created " + l.Created.Format(homeListDateLayout)
	}
	return item
}

// ---- the lists level ----------------------------------------------------------

// listsTabView is the first level: the lists in a window that follows the
// cursor, the name entry above them while it is open, and the details of the
// highlighted list beside them (below them in plain mode).
func (m *searchHomeModel) listsTabView(width, height int) string {
	top := m.listsTop(width)
	extra := 0
	if top != "" {
		extra = lipgloss.Height(top)
	}
	head := m.sectionHeader("YOUR LISTS", "") + m.listsCount()
	if len(m.rows) == 0 {
		return homeListsJoin(top, head, m.theme.Muted.Render("  "+m.noListsText()))
	}
	teach := ""
	if m.teachesLists() {
		lines := strings.Split(ansi.Wrap(homeListTeaching, max(width-2, 10), ""), "\n")
		for i := range lines {
			lines[i] = m.theme.Muted.Render("  " + lines[i])
		}
		teach = strings.Join(lines, "\n")
		extra += len(lines)
	}

	side, below := m.listPanels(width, height, extra)
	listWidth := width
	if side != "" {
		listWidth = resultsListWidth(width)
	}
	extraBelow := 0
	if below != "" {
		extraBelow = lipgloss.Height(below) + 1
	}
	capacity, arrows := m.tabCapacity(width, height, extra+extraBelow)
	win := windowAround(len(m.rows), m.cursor, capacity, arrows)
	items := make([]homeRowItem, 0, win.end-win.start)
	for _, row := range m.rows[win.start:win.end] {
		items = append(items, m.listItem(row, listWidth))
	}
	body := m.windowLines(head, items, win, listWidth)
	if side != "" {
		body = lipgloss.JoinHorizontal(lipgloss.Top, lipgloss.NewStyle().Width(listWidth).Render(body), side)
	}
	if below != "" {
		body += "\n\n" + below
	}
	return homeListsJoin(top, body, teach)
}

// homeListTeaching tells a user with no lists of their own what to do next.
const homeListTeaching = `A list keeps the titles you want to find again. You have none of your own yet: press n to make one, for example "Weekend marathon".`

// teachesLists reports whether the tab should explain lists: the user has none
// of their own and the box is not hiding anything.
func (m *searchHomeModel) teachesLists() bool {
	h := m.lists
	if h.loadErr != "" || strings.TrimSpace(m.input.Value()) != "" {
		return false
	}
	return !slices.ContainsFunc(h.all, func(l HomeList) bool { return !l.Auto })
}

// listsFull reports whether the user has as many lists of their own as there can be.
func (m *searchHomeModel) listsFull() bool {
	own := 0
	for i := range m.lists.all {
		if !m.lists.all[i].Auto {
			own++
		}
	}
	return own >= homeListMax
}

// homeListsJoin joins the parts that hold something, one per line.
func homeListsJoin(parts ...string) string {
	kept := parts[:0:0]
	for _, p := range parts {
		if p != "" {
			kept = append(kept, p)
		}
	}
	return strings.Join(kept, "\n")
}

// listsTop is what sits above the heading: the name entry and a failed read.
func (m *searchHomeModel) listsTop(width int) string {
	var parts []string
	if entry := m.entryView(width); entry != "" {
		parts = append(parts, entry)
	}
	if m.lists.loadErr != "" {
		parts = append(parts, strings.TrimRight(renderNotice(&m.theme, m.lists.loadErr, width), "\n"))
	}
	return strings.Join(parts, "\n")
}

// listsCount is the number of lists after the heading.
func (m *searchHomeModel) listsCount() string {
	shown := 0
	for _, row := range m.rows {
		if row.kind == homeRowList {
			shown++
		}
	}
	return m.tabCount(shown, len(m.lists.all))
}

// noListsText is what the first level says when it has no rows.
func (m *searchHomeModel) noListsText() string {
	if m.lists.loadErr != "" {
		return "Press tab and then tab back here to try again."
	}
	return fmt.Sprintf("No list matches %q. Delete a letter to see them all.", strings.TrimSpace(m.input.Value()))
}

// panelList is the list the details panel is about: the highlighted one,
// whether the box or the list has the focus.
func (m *searchHomeModel) panelList() (HomeList, bool) {
	if m.cursor < 0 || m.cursor >= len(m.rows) || m.rows[m.cursor].kind != homeRowList {
		return HomeList{}, false
	}
	return m.lists.all[m.rows[m.cursor].index], true
}

// listPanels draws the details of the highlighted list: side is the panel
// beside the rows on a wide terminal, below the same details as text in plain
// mode. Both are "" when there is nothing to show or no room for it.
func (m *searchHomeModel) listPanels(width, height, extra int) (side, below string) {
	l, ok := m.panelList()
	if !ok {
		return "", ""
	}
	avail := height - m.tabAbove(width, extra)
	if Plain() {
		// The rows keep at least their heading and a few lines.
		return "", m.fitListPanel(&l, width, min(avail/2, avail-3))
	}
	if width < wideResultsBreakpoint {
		return "", ""
	}
	return m.fitListPanel(&l, width-resultsListWidth(width), avail), ""
}

// fitListPanel draws the panel of l in at most maxHeight lines, naming fewer
// titles until it fits. It is "" when even the bare panel does not.
func (m *searchHomeModel) fitListPanel(l *HomeList, width, maxHeight int) string {
	// A panel that keeps its tip with fewer titles beats one with the titles
	// and no tip: the tip is what says what enter does.
	for _, wantTip := range []bool{true, false} {
		for shown := min(homeListPanelTitles, len(l.Titles)); shown >= 0; shown-- {
			panel := fitPanel(&m.theme, homeListPanelContent(l, shown), width, maxHeight)
			if panel != "" && (!wantTip || strings.Contains(ansi.Strip(panel), homeListPanelTip)) {
				return panel
			}
		}
	}
	return ""
}

// homeListPanelContent is what the panel holds: the list's count and date, then
// shown of its titles and how many more there are.
func homeListPanelContent(l *HomeList, shown int) *panelContent {
	created := "Always there"
	if !l.Auto {
		created = "Unknown"
		if !l.Created.IsZero() {
			created = l.Created.Format(homeListDateLayout)
		}
	}
	var lines []string
	for _, title := range l.Titles[:min(shown, len(l.Titles))] {
		lines = append(lines, singleLine(title))
	}
	if more := max(l.Count, len(l.Titles)) - len(lines); more > 0 && len(lines) > 0 {
		lines = append(lines, fmt.Sprintf("+%d more", more))
	}
	if l.Count == 0 && len(l.Titles) == 0 {
		lines = []string{"Nothing in it yet."}
	}
	return &panelContent{
		label:  "List",
		title:  singleLine(l.Name),
		fields: []detailField{{label: "Titles", value: fmt.Sprint(l.Count)}, {label: "Created", value: created}},
		lines:  lines,
		tip:    homeListPanelTip,
	}
}

// ---- the name entry -------------------------------------------------------------

// sizeListsEntry fits the name entry to the terminal.
func (m *searchHomeModel) sizeListsEntry() {
	if m.lists == nil {
		return
	}
	width, _ := m.shell.ContentSize()
	m.lists.name.SetWidth(max(width-lipgloss.Width(m.lists.name.Prompt)-2, 8))
}

// entryView is the name entry, its instructions and why the last enter was
// refused. "" while the entry is closed.
func (m *searchHomeModel) entryView(width int) string {
	h := m.lists
	if h.entry == homeEntryNone {
		return ""
	}
	verb := "create"
	if h.entry == homeEntryRename {
		verb = "rename"
	}
	count := len([]rune(h.name.Value()))
	counter := fmt.Sprintf("%d/%d", count, homeListNameLimit)
	if count >= homeListNameLimit {
		counter += " · that is the longest a name can be"
	}
	lines := []string{h.name.View(), m.theme.Muted.Render("  enter " + verb + " · esc cancel   " + counter)}
	if h.entryErr != "" {
		lines = append(lines, strings.TrimRight(renderNotice(&m.theme, h.entryErr, width), "\n"))
	}
	return strings.Join(lines, "\n")
}

// openEntry opens the name entry, empty to create a list or holding the name
// of l to rename it. The box gives up its caret while it is open.
func (m *searchHomeModel) openEntry(kind homeListEntry, l HomeList) {
	h := m.lists
	h.entry, h.entryID, h.entryErr, h.wasFocus = kind, l.ID, "", m.focusInput
	h.name.SetValue(l.Name)
	h.name.CursorEnd()
	h.name.Focus()
	m.input.Blur()
	m.sizeListsEntry()
}

// closeEntry closes the name entry and gives the box its caret back if it had it.
func (m *searchHomeModel) closeEntry() {
	h := m.lists
	h.entry = homeEntryNone
	h.name.Blur()
	if h.wasFocus {
		m.focusInput = true
		m.input.Focus()
	}
}

// submitEntry creates or renames the list with the name typed. A name the
// lists refuse stays in the entry, with the reason beside it.
func (m *searchHomeModel) submitEntry() {
	h := m.lists
	name := h.name.Value()
	if h.entry == homeEntryRename {
		if err := h.api.Rename(h.entryID, name); err != nil {
			h.entryErr = friendly(err)
			return
		}
		m.closeEntry()
		h.reload()
		m.showList(h.entryID)
		return
	}
	created, err := h.api.Create(name)
	if err != nil {
		h.entryErr = friendly(err)
		return
	}
	m.closeEntry()
	h.reload()
	m.openList(&created)
	if h.level == homeLevelLists {
		m.showList(created.ID)
	}
}

// entryKey answers a key while the name entry is open: enter creates or
// renames, esc cancels, ctrl+c quits and the rest types the name.
func (m *searchHomeModel) entryKey(msg tea.KeyPressMsg) tea.Cmd {
	h := m.lists
	switch msg.String() {
	case "ctrl+c":
		m.err = ErrSearchHomeCancelled
		return tea.Quit
	case "esc":
		m.closeEntry()
	case "enter":
		m.submitEntry()
	default:
		h.entryErr = ""
		var cmd tea.Cmd
		h.name, cmd = h.name.Update(msg)
		return cmd
	}
	return nil
}

// ---- keys -----------------------------------------------------------------------

// listsKey is the lists' part of a key press: it reports whether it took the
// key. The name entry takes every key while it is open; otherwise enter, esc
// and the row keys mean something different here than on the other tabs.
func (m *searchHomeModel) listsKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd, bool) {
	if !m.listsActive() {
		return m, nil, false
	}
	if m.lists.entry != homeEntryNone {
		cmd := m.entryKey(msg)
		return m, cmd, true
	}
	switch key := msg.String(); key {
	case "esc":
		if m.lists.level == homeLevelWorks && m.input.Value() == "" {
			m.backToLists()
			return m, nil, true
		}
	case "enter":
		if m.lists.level == homeLevelLists {
			m.openHighlighted()
			return m, nil, true
		}
	case "n":
		if m.lists.level == homeLevelLists && !m.focusInput {
			m.newList()
			return m, nil, true
		}
	case "u":
		if m.lists.level == homeLevelWorks && !m.focusInput && m.lists.removed != nil {
			m.undoRemove()
			return m, nil, true
		}
	case "r", "e", "x", "X":
		if !m.focusInput {
			m.listsRowKey(key)
			return m, nil, true
		}
	}
	return m, nil, false
}

// listsRowKey runs r, e, x or X on the highlighted row of either level.
func (m *searchHomeModel) listsRowKey(key string) {
	if m.lists.level == homeLevelLists {
		m.listKey(key)
		return
	}
	switch key {
	case "x":
		m.removeWork()
	case "X":
		if len(m.lists.works) > 0 {
			m.clearing = clearListWorks
		}
	}
}

// listKey runs r (rename) and x (delete) on the highlighted list.
func (m *searchHomeModel) listKey(key string) {
	l, ok := m.highlightedList()
	if !ok {
		if key == "r" || key == "x" {
			m.notice = "Pick a list first."
		}
		return
	}
	switch {
	case key == "r" && l.Auto:
		m.notice = m.cannotRename(&l)
	case key == "r":
		m.openEntry(homeEntryRename, l)
	case key == "x" && l.Auto:
		m.notice = "The " + singleLine(l.Name) + " list cannot be deleted."
	case key == "x":
		m.lists.doomed = l
		m.clearing = clearDeleteList
	}
}

// cannotRename asks the lists why the automatic list cannot be renamed, in
// the words they use.
func (m *searchHomeModel) cannotRename(l *HomeList) string {
	if err := m.lists.api.Rename(l.ID, l.Name); err != nil {
		return friendly(err)
	}
	return "The " + singleLine(l.Name) + " list cannot be renamed."
}

// highlightedList is the list the highlight is on, when it is on one.
func (m *searchHomeModel) highlightedList() (HomeList, bool) {
	row, ok := m.currentRow()
	if !ok || row.kind != homeRowList {
		return HomeList{}, false
	}
	return m.lists.all[row.index], true
}

// openHighlighted runs enter on the first level: the highlighted row. While
// the box has the focus it opens the first list that matches, never "+ Create
// a new list": making a list takes a highlight on that row or the n key.
func (m *searchHomeModel) openHighlighted() {
	pos := m.cursor
	if m.focusInput {
		pos = slices.IndexFunc(m.rows, func(r homeRow) bool { return r.kind == homeRowList })
	}
	if pos < 0 || pos >= len(m.rows) {
		return
	}
	switch row := m.rows[pos]; row.kind {
	case homeRowNewList:
		m.newList()
	case homeRowList:
		l := m.lists.all[row.index]
		m.openList(&l)
	}
}

// newList opens the name entry for a new list, or says up front that there
// is no room for one.
func (m *searchHomeModel) newList() {
	if m.listsFull() {
		m.notice = fmt.Sprintf("You have %d lists, the most there can be. Delete one to make a new one.", homeListMax)
		return
	}
	m.openEntry(homeEntryCreate, HomeList{})
}

// listsLanded puts the highlight on the first list when the tab opens, so
// that enter, r and x act on a row the user can see.
func (m *searchHomeModel) listsLanded() {
	if !m.listsLevelOne() {
		return
	}
	if i := slices.IndexFunc(m.rows, func(r homeRow) bool { return r.kind == homeRowList }); i >= 0 {
		m.cursor, m.focusInput = i, false
		m.input.Blur()
	}
}

// openList shows the works of l. A list that cannot be read leaves the lists
// showing, with the reason as a notice.
func (m *searchHomeModel) openList(l *HomeList) {
	h := m.lists
	works, err := h.api.Works(l.ID)
	if err != nil {
		m.notice = homeListsReadFailed(err)
		return
	}
	h.level, h.open, h.works = homeLevelWorks, *l, slices.Clone(works)
	m.input.Placeholder = homeWorksPlaceholder
	m.resetBox()
}

// backToLists goes up from the works of a list, with the highlight on the
// list it came from.
func (m *searchHomeModel) backToLists() {
	h := m.lists
	h.level = homeLevelLists
	h.removed = nil
	h.reload()
	m.input.Placeholder = homeListsPlaceholder
	m.showList(h.open.ID)
}

// resetBox empties the box, gives it the focus and redraws the rows.
func (m *searchHomeModel) resetBox() {
	m.input.SetValue("")
	m.focusInput = true
	m.input.Focus()
	m.cursor = 0
	m.queryChanged()
}

// showList empties the box and puts the highlight on the list with id.
func (m *searchHomeModel) showList(id string) {
	m.resetBox()
	for i, row := range m.rows {
		if row.kind == homeRowList && m.lists.all[row.index].ID == id {
			m.cursor = i
			m.focusInput = false
			m.input.Blur()
			return
		}
	}
}

// ---- works of one list ------------------------------------------------------------

// removeWork takes the highlighted work off the open list.
func (m *searchHomeModel) removeWork() {
	row, ok := m.currentRow()
	if !ok {
		return
	}
	h := m.lists
	gone := h.works[row.index]
	if err := h.api.RemoveWork(h.open.ID, gone.Key); err != nil {
		m.notice = "Could not remove the title. " + friendly(err)
		return
	}
	h.removed = &gone
	m.worksChanged()
	if h.level == homeLevelWorks && m.notice == "" {
		m.notice = "Removed " + singleLine(gone.Title) + ". Press u to put it back."
	}
}

// undoRemove puts the title removed last back on the open list.
func (m *searchHomeModel) undoRemove() {
	h := m.lists
	back := *h.removed
	h.removed = nil
	if _, err := h.api.ToggleWork(h.open.ID, back); err != nil {
		m.notice = "Could not put the title back. " + friendly(err)
		return
	}
	m.worksChanged()
	if h.level == homeLevelWorks && m.notice == "" {
		m.notice = singleLine(back.Title) + " is back in the list."
	}
}

// worksChanged reads the open list again after a removal or a clear, and
// keeps the ALL tab's FAVORITES section in step when it is the automatic list.
func (m *searchHomeModel) worksChanged() {
	h := m.lists
	works, err := h.api.Works(h.open.ID)
	if err != nil {
		m.notice = homeListsReadFailed(err)
		m.backToLists()
		return
	}
	h.works = slices.Clone(works)
	if h.open.Auto {
		m.refreshFavorites()
	}
	m.rowsChanged()
}

// refreshFavorites reads the Favorites list again for the ALL tab's section.
func (m *searchHomeModel) refreshFavorites() {
	works, err := m.lists.api.Works(FavoritesListID)
	if err != nil {
		m.notice = homeListsReadFailed(err)
		return
	}
	m.favoriteAll = slices.Clone(works)
}

// listsConfirmed runs the y of a lists question: clear the open list, or
// delete the list asked about.
func (m *searchHomeModel) listsConfirmed(target clearTarget) {
	h := m.lists
	if target == clearListWorks {
		if err := h.api.Clear(h.open.ID); err != nil {
			m.notice = "Could not clear the list. " + friendly(err)
			return
		}
		m.worksChanged()
		return
	}
	if err := h.api.Delete(h.doomed.ID); err != nil {
		m.notice = "Could not delete the list. " + friendly(err)
		return
	}
	h.reload()
	m.rowsChanged()
}

// listsQuestion is the question under the tabs for the lists' y/n questions.
func (m *searchHomeModel) listsQuestion() string {
	h := m.lists
	if m.clearing == clearListWorks {
		return fmt.Sprintf("Clear all %s from %s? y to clear, n to keep.", countOf(len(h.works), "title", "titles"), homeListQuoted(h.open.Name))
	}
	name := homeListQuoted(h.doomed.Name)
	switch n := h.doomed.Count; {
	case n <= 0:
		return fmt.Sprintf("Delete the empty list %s? y to delete, n to keep.", name)
	case n == 1:
		return fmt.Sprintf("Delete the list %s? The title stays available, only the list goes. y to delete, n to keep.", name)
	default:
		return fmt.Sprintf("Delete the list %s? The %d titles stay available, only the list goes. y to delete, n to keep.", name, n)
	}
}

// homeListQuoted is a list's name between plain quotes, on one line. Unlike
// %q it does not turn a quote or an accent in the name into an escape.
func homeListQuoted(name string) string {
	return `"` + singleLine(name) + `"`
}

// ---- footer and help ----------------------------------------------------------------

// listsModalHints are the footer keys while the name entry or the delete
// question is open. nil otherwise.
func (m *searchHomeModel) listsModalHints() []Hint {
	if !m.listsActive() {
		return nil
	}
	switch {
	case m.lists.entry == homeEntryCreate:
		return []Hint{{Key: "enter", Desc: "create", Essential: true}, {Key: "esc", Desc: "cancel", Essential: true}, {Key: "ctrl+c", Desc: "quit"}}
	case m.lists.entry == homeEntryRename:
		return []Hint{{Key: "enter", Desc: "rename", Essential: true}, {Key: "esc", Desc: "cancel", Essential: true}, {Key: "ctrl+c", Desc: "quit"}}
	case m.clearing == clearDeleteList:
		return []Hint{{Key: "y", Desc: "delete", Essential: true}, {Key: "n", Desc: "keep", Essential: true}, {Key: "ctrl+c", Desc: "quit"}}
	}
	return nil
}

// listsFooter words the footer for the lists: enter names what it will do, a
// user's list offers delete and rename, and the works of a list say how to go
// back.
func (m *searchHomeModel) listsFooter(hints []Hint) []Hint {
	if !m.listsActive() {
		return hints
	}
	if m.lists.level == homeLevelWorks {
		back := "back to your lists"
		if m.input.Value() != "" {
			back = "clear filter"
		}
		return slices.Insert(hints, 1, Hint{Key: "esc", Desc: back, Essential: true})
	}
	hints = slices.DeleteFunc(hints, func(h Hint) bool { return h.Key == "x" || h.Key == "X" })
	if i := slices.IndexFunc(hints, func(h Hint) bool { return h.Key == "enter" }); i >= 0 {
		hints[i].Desc = m.enterWord()
	}
	var more []Hint
	if l, ok := m.highlightedList(); ok && !l.Auto {
		more = append(more, Hint{Key: "x", Desc: "delete", Essential: true}, Hint{Key: "r", Desc: "rename", Essential: true})
	}
	if !m.focusInput {
		// n types into the box while the box has the focus.
		more = append(more, Hint{Key: "n", Desc: "new list"})
	}
	at := slices.IndexFunc(hints, func(h Hint) bool { return h.Key == "?" })
	if at < 0 {
		at = len(hints)
	}
	return slices.Insert(hints, at, more...)
}

// enterWord is what enter does on the first level: create on the "+ Create"
// row, open otherwise (with the box focused it opens the first match).
func (m *searchHomeModel) enterWord() string {
	if !m.focusInput && m.cursor >= 0 && m.cursor < len(m.rows) && m.rows[m.cursor].kind == homeRowNewList {
		return "create"
	}
	return "open"
}

// listsHelpRows are the lists' lines in the help overlay.
func (m *searchHomeModel) listsHelpRows() []HelpRow {
	if m.lists == nil {
		return nil
	}
	return []HelpRow{
		{Key: "enter", Desc: "FAVORITES: open a list or make a new one"},
		{Key: "esc", Desc: "FAVORITES: back to your lists"},
		{Key: "n", Desc: "FAVORITES: make a new list"},
		{Key: "r", Desc: "FAVORITES: rename a list"},
		{Key: "x", Desc: "FAVORITES: delete a list or remove a title (u puts it back)"},
	}
}
