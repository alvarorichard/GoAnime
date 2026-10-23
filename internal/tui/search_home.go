package tui

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
)

var (
	// ErrSearchHomeBack means the user went back from the search home screen.
	ErrSearchHomeBack = errors.New("back from search home")
	// ErrSearchHomeCancelled means the user quit the search home screen.
	ErrSearchHomeCancelled = errors.New("search home cancelled")
)

// Home scopes: the tab row under the search box. Value is the source
// selector the search runs with ("", "ptbr", "hianime" or "movietv").
var homeScopes = []struct{ Name, Value string }{
	{"ALL", ""},
	{"PT-BR", "ptbr"},
	{"ENGLISH", "hianime"},
	{"MOVIES/TV", "movietv"},
}

// minHomeQuery is the shortest query enter searches for.
const minHomeQuery = 2

// HomeAction is what the home screen decided.
type HomeAction int

const (
	// HomeSearch searches Query with Scope.
	HomeSearch HomeAction = iota
	// HomeResume reopens Entry's episode at its saved position.
	HomeResume
	// HomeFromStart reopens Entry's episode from zero.
	HomeFromStart
	// HomeEpisodes opens Entry's episode list.
	HomeEpisodes
)

// HomeContinueEntry is one row of the CONTINUE list.
type HomeContinueEntry struct {
	Key          string // tracking key, for deletion
	AnimeTitle   string
	EpisodeTitle string
	EpisodeNum   int
	Position     int // seconds
	Duration     int // seconds
	Source       string
	Audio        string
	MediaType    string // "anime" or "movie", for continue routing
	UpdatedAt    time.Time
}

// Finished reports whether the episode was watched into its last tenth,
// where the end credits run (same rule as the episode picker).
func (e HomeContinueEntry) Finished() bool {
	return e.Duration > 0 && e.Position >= e.Duration*9/10
}

// DisplayTitle is the row's title: the show's name, the episode's when the
// row predates show names.
func (e HomeContinueEntry) DisplayTitle() string {
	if t := singleLine(e.AnimeTitle); t != "" {
		return t
	}
	if t := singleLine(e.EpisodeTitle); t != "" {
		return t
	}
	return "Unknown title"
}

// HomeRecentSearch is one row of the RECENT list.
type HomeRecentSearch struct {
	Query string
	Count int
}

// SearchHomeOptions configure the SEARCH home screen.
type SearchHomeOptions struct {
	// Query prefills the search box (kept from the last search).
	Query string
	// ScopeIndex preselects the scope tab.
	ScopeIndex int
	// SourcesLine reads at the right of the header ("6 sources on · Goyabu off").
	SourcesLine string
	Continue    []HomeContinueEntry
	Recent      []HomeRecentSearch
	// DeleteEntry removes a CONTINUE row by key; nil disables x.
	DeleteEntry func(key string) error
}

// SearchHomeResult is the home screen's decision.
type SearchHomeResult struct {
	Action    HomeAction
	Query     string
	Scope     string
	Entry     HomeContinueEntry
	HasEntry  bool
	Cancelled bool
}

type homeRowKind int

const (
	homeRowContinue homeRowKind = iota
	homeRowRecent
)

type homeRow struct {
	kind  homeRowKind
	index int // into Continue or Recent
}

// homeRowItem is one home list row: status, title, right note and far-right key.
type homeRowItem struct {
	status PickStatus
	title  string
	note   string // at the right: "Ep 5 · 06:12" or "5 results"
	key    string // at the far right: "yesterday"
}

// relativeDay words how long ago t was: yesterday, N days ago, last week.
func relativeDay(t, now time.Time) string {
	days := int(now.Sub(t).Hours() / 24)
	switch {
	case days <= 0:
		return "today"
	case days == 1:
		return "yesterday"
	case days <= 6:
		return fmt.Sprintf("%d days ago", days)
	case days <= 13:
		return "last week"
	case days <= 59:
		return fmt.Sprintf("%d weeks ago", days/7)
	default:
		return t.Format("2006-01-02")
	}
}

type searchHomeModel struct {
	theme       Theme
	shell       Shell
	input       textinput.Model
	scopes      []struct{ Name, Value string }
	scopeIdx    int
	continueAll []HomeContinueEntry
	recentAll   []HomeRecentSearch
	rows        []homeRow
	cursor      int
	focusInput  bool
	deleteEntry func(key string) error
	now         func() time.Time

	result   SearchHomeResult
	err      error
	showHelp bool
}

func newSearchHomeModel(opts SearchHomeOptions) *searchHomeModel {
	theme := CurrentTheme()
	shell := NewShell(&theme, "Search")
	shell.Context = singleLine(opts.SourcesLine)

	input := textinput.New()
	if Plain() {
		input.Prompt = "> "
		input.Placeholder = "title, or down to pick from the lists below"
	} else {
		input.Prompt = "❯ "
		input.Placeholder = "title, or ↓ to pick from the lists below"
	}
	styles := textinput.DefaultStyles(themeIsDark())
	styles.Focused.Text = theme.Text
	styles.Focused.Placeholder = theme.Muted
	styles.Focused.Prompt = theme.Primary
	styles.Blurred = styles.Focused
	styles.Cursor.Color = theme.Primary.GetForeground()
	styles.Cursor.Shape = tea.CursorBlock
	input.SetStyles(styles)
	if query := singleLine(opts.Query); query != "" {
		input.SetValue(query)
		input.CursorEnd()
	}
	input.Focus()

	scopeIdx := opts.ScopeIndex
	if scopeIdx < 0 || scopeIdx >= len(homeScopes) {
		scopeIdx = 0
	}

	m := &searchHomeModel{
		theme:       theme,
		shell:       shell,
		input:       input,
		scopes:      homeScopes,
		scopeIdx:    scopeIdx,
		continueAll: opts.Continue,
		recentAll:   opts.Recent,
		cursor:      0,
		focusInput:  true,
		deleteEntry: opts.DeleteEntry,
		now:         time.Now,
	}
	m.rebuildRows()
	return m
}

// themeIsDark reports the process theme for the input's base styles.
func themeIsDark() bool {
	return IsDark()
}

func (m *searchHomeModel) rebuildRows() {
	m.rows = m.rows[:0]
	for i := range m.continueAll {
		m.rows = append(m.rows, homeRow{kind: homeRowContinue, index: i})
	}
	for i := range m.recentAll {
		m.rows = append(m.rows, homeRow{kind: homeRowRecent, index: i})
	}
	if m.cursor >= len(m.rows) {
		m.cursor = max(len(m.rows)-1, 0)
	}
	if len(m.rows) == 0 {
		m.focusInput = true
	}
}

func (m *searchHomeModel) currentRow() (homeRow, bool) {
	if m.focusInput || m.cursor < 0 || m.cursor >= len(m.rows) {
		return homeRow{}, false
	}
	return m.rows[m.cursor], true
}

// Init focuses the search box without background work.
func (m *searchHomeModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update handles resize, typing, scope tabs, navigation and actions.
func (m *searchHomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.shell.Resize(msg.Width, msg.Height)
		m.sizeInput()
		return m, nil

	case tea.KeyPressMsg:
		if m.showHelp {
			switch msg.String() {
			case "?", "esc":
				m.showHelp = false
				return m, nil
			case "ctrl+c":
				m.err = ErrSearchHomeCancelled
				return m, tea.Quit
			default:
				return m, nil
			}
		}
		switch msg.String() {
		case "ctrl+c":
			m.err = ErrSearchHomeCancelled
			return m, tea.Quit
		case "?":
			if !m.focusInput {
				m.showHelp = true
				return m, nil
			}
		case "tab":
			m.scopeIdx = (m.scopeIdx + 1) % len(m.scopes)
			return m, nil
		case "esc":
			// esc clears the box; it never quits (tela 10).
			if m.input.Value() != "" {
				m.input.SetValue("")
				m.focusInput = true
				m.input.Focus()
			}
			return m, nil
		case "enter":
			if res, ok := m.confirm(); ok {
				m.result = res
				return m, tea.Quit
			}
			return m, nil
		case "up":
			m.move(-1)
			return m, nil
		case "down":
			m.move(1)
			return m, nil
		case "r", "e", "x":
			if !m.focusInput {
				if cmd := m.rowAction(msg.String()); cmd != nil {
					return m, cmd
				}
				return m, nil
			}
		}
		if m.focusInput {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
		// Typing over the lists jumps back to the box with the character.
		if isHomeTypeKey(msg) {
			m.focusInput = true
			m.input.Focus()
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// move steps the highlight; up from the first row returns to the box.
func (m *searchHomeModel) move(delta int) {
	if len(m.rows) == 0 {
		m.focusInput = true
		return
	}
	if m.focusInput {
		if delta > 0 {
			m.focusInput = false
			m.input.Blur()
		}
		return
	}
	next := m.cursor + delta
	if next < 0 {
		m.focusInput = true
		m.input.Focus()
		return
	}
	m.cursor = min(next, len(m.rows)-1)
}

// confirm turns enter into a result: a search from the box or a recent row,
// a resume from a continue row.
func (m *searchHomeModel) confirm() (SearchHomeResult, bool) {
	scope := m.scopes[m.scopeIdx].Value
	if m.focusInput {
		query := strings.TrimSpace(m.input.Value())
		if len([]rune(query)) < minHomeQuery {
			return SearchHomeResult{}, false
		}
		return SearchHomeResult{Action: HomeSearch, Query: query, Scope: scope}, true
	}
	row, ok := m.currentRow()
	if !ok {
		return SearchHomeResult{}, false
	}
	switch row.kind {
	case homeRowRecent:
		return SearchHomeResult{Action: HomeSearch, Query: m.recentAll[row.index].Query, Scope: scope}, true
	default:
		entry := m.continueAll[row.index]
		return SearchHomeResult{Action: HomeResume, Entry: entry, HasEntry: true, Scope: scope}, true
	}
}

// rowAction runs r (from start), e (episodes) and x (remove) on a continue
// row. It returns a quit command when the action leaves the screen.
func (m *searchHomeModel) rowAction(key string) tea.Cmd {
	row, ok := m.currentRow()
	if !ok || row.kind != homeRowContinue {
		return nil
	}
	entry := m.continueAll[row.index]
	switch key {
	case "r":
		m.result = SearchHomeResult{Action: HomeFromStart, Entry: entry, HasEntry: true, Scope: m.scopes[m.scopeIdx].Value}
		return tea.Quit
	case "e":
		m.result = SearchHomeResult{Action: HomeEpisodes, Entry: entry, HasEntry: true, Scope: m.scopes[m.scopeIdx].Value}
		return tea.Quit
	case "x":
		if m.deleteEntry == nil {
			return nil
		}
		if err := m.deleteEntry(entry.Key); err != nil {
			return nil
		}
		m.continueAll = append(m.continueAll[:row.index], m.continueAll[row.index+1:]...)
		m.rebuildRows()
		if len(m.rows) == 0 {
			m.focusInput = true
			m.input.Focus()
		}
		return nil
	}
	return nil
}

func (m *searchHomeModel) sizeInput() {
	width, _ := m.shell.ContentSize()
	m.input.SetWidth(max(width-lipgloss.Width(m.input.Prompt)-2, 8))
}

// isHomeTypeKey reports keys that type into the search box: printable runes,
// including space for multi-word titles.
func isHomeTypeKey(msg tea.KeyPressMsg) bool {
	if msg.Text == "" {
		return false
	}
	r, _ := utf8.DecodeRuneInString(msg.Text)
	return r >= 32 && r != 127 && !unicode.IsControl(r)
}

// hints lists the footer keys (tela 1).
func (m *searchHomeModel) hints() []Hint {
	if m.showHelp {
		return []Hint{
			{Key: "?", Desc: "close", Essential: true},
			{Key: "esc", Desc: "close", Essential: true},
		}
	}
	move := "↑↓"
	if Plain() {
		move = "up/dn"
	}
	return []Hint{
		{Key: "enter", Desc: "search", Essential: true},
		{Key: move, Desc: "move"},
		{Key: "tab", Desc: "sources"},
		{Key: "?", Desc: "help"},
		{Key: "ctrl+c", Desc: "quit", Essential: true},
	}
}

// View renders the search box, the scope tabs, the continue and recent
// lists, and the details panel of the highlighted continue row.
func (m *searchHomeModel) View() tea.View {
	width, height := m.shell.ContentSize()
	var body string
	if m.showHelp {
		body = m.helpView(width, height)
	} else {
		body = m.homeView(width, height)
	}
	view := tea.NewView(m.shell.Render(body, m.hints()))
	view.AltScreen = true
	view.WindowTitle = "GoAnime"
	return view
}

func (m *searchHomeModel) helpView(width, height int) string {
	if help := RenderHelp(&m.theme, "Search", []HelpRow{
		{Key: "enter", Desc: "search, resume"},
		{Key: "tab", Desc: "sources"},
		{Key: "r", Desc: "from the start"},
		{Key: "e", Desc: "episodes"},
		{Key: "x", Desc: "remove"},
	}, nil, width, height); help != "" {
		return help
	}
	return ""
}

// homeView lays out the box, tabs and lists, with the panel beside the
// continue list on a wide terminal (as text below it in plain mode).
func (m *searchHomeModel) homeView(width, height int) string {
	sections := []string{m.searchBox(width), m.scopeTabs()}
	lists, panel := m.listsView(width, height)
	if lists != "" {
		lists = m.sectionHeader("CONTINUE", "続き") + "\n" + lists
		if panel != "" {
			if Plain() {
				// Tela 9: details as text below the lists.
				lists += "\n\n" + panel
			} else {
				column := lipgloss.NewStyle().Width(resultsListWidth(width)).Render(lists)
				lists = lipgloss.JoinHorizontal(lipgloss.Top, column, panel)
			}
		}
		sections = append(sections, "", lists)
	}
	if recent := m.recentView(width); recent != "" {
		sections = append(sections, "", m.sectionHeader("RECENT", "最近"), recent)
	}
	return strings.Join(sections, "\n")
}

// searchBox draws the bordered query box.
func (m *searchHomeModel) searchBox(width int) string {
	content := m.input.View()
	border := lipgloss.NormalBorder()
	style := lipgloss.NewStyle().Border(border, true).Width(max(width-2, 8))
	style = style.BorderForeground(m.theme.Primary.GetForeground())
	return style.Render(content)
}

// scopeTabs draws the "in ALL PT-BR ENGLISH MOVIES/TV tab change" row.
func (m *searchHomeModel) scopeTabs() string {
	var sb strings.Builder
	sb.WriteString(m.theme.Muted.Render("in "))
	for i, scope := range m.scopes {
		name := scope.Name
		if i == m.scopeIdx {
			sb.WriteString(m.theme.Tag.Render(name))
		} else {
			sb.WriteString(m.theme.Muted.Render(name))
		}
		if i < len(m.scopes)-1 {
			sb.WriteString("   ")
		}
	}
	sb.WriteString("     ")
	sb.WriteString(m.theme.Keycap.Render("tab"))
	sb.WriteString(m.theme.Footer.Render(" change"))
	return sb.String()
}

// sectionHeader draws a CONTINUE/RECENT heading with its Japanese gloss.
func (m *searchHomeModel) sectionHeader(label, gloss string) string {
	head := m.theme.Primary.Render(label)
	if ShowJapanese() {
		head += m.theme.Gloss.Render("  " + gloss)
	}
	return head
}

// listsView renders the continue rows and the details panel of the
// highlighted continue row: beside the lists on a wide terminal, as text
// below them in plain mode (tela 9).
func (m *searchHomeModel) listsView(width, height int) (lists, panel string) {
	if len(m.continueAll) == 0 {
		return "", ""
	}
	items := make([]homeRowItem, 0, len(m.continueAll))
	now := m.now()
	for _, e := range m.continueAll {
		status := PickInProgress
		note := fmt.Sprintf("Ep %d · %s", e.EpisodeNum, FormatClock(float64(e.Position)))
		if e.Finished() {
			status = PickWatched
			note = fmt.Sprintf("Ep %d · done", e.EpisodeNum)
		}
		items = append(items, homeRowItem{status: status, title: e.DisplayTitle(), note: note, key: relativeDay(e.UpdatedAt, now)})
	}
	lists = renderHomeRows(&m.theme, items, m.sectionCursor(homeRowContinue), m.listWidth(width))

	row, ok := m.currentRowForPanel()
	if !ok {
		return lists, ""
	}
	content := m.continuePanel(&m.continueAll[row.index])
	if Plain() {
		return lists, renderPlainPanel(&m.theme, content, width, height)
	}
	if width < wideResultsBreakpoint {
		return lists, ""
	}
	if panel := fitPanel(&m.theme, content, width-resultsListWidth(width), height); panel != "" {
		return lists, panel
	}
	return lists, ""
}

// currentRowForPanel is the highlighted continue row for the panel, whether
// or not the box has focus.
func (m *searchHomeModel) currentRowForPanel() (homeRow, bool) {
	if m.cursor < 0 || m.cursor >= len(m.rows) || m.rows[m.cursor].kind != homeRowContinue {
		return homeRow{}, false
	}
	return m.rows[m.cursor], true
}

// listWidth is the width rows render at: the list column when the panel
// sits beside them, full width otherwise.
func (m *searchHomeModel) listWidth(width int) int {
	if Plain() || width < wideResultsBreakpoint {
		return width
	}
	if _, ok := m.currentRowForPanel(); ok {
		return resultsListWidth(width)
	}
	return width
}

// renderHomeRows draws items one line each, highlighting selected (-1 for
// none: the highlight belongs to the other section then).
func renderHomeRows(theme *Theme, items []homeRowItem, selected, width int) string {
	lines := make([]string, 0, len(items))
	for i := range items {
		lines = append(lines, renderHomeRow(theme, i == selected, items[i], width))
	}
	return strings.Join(lines, "\n")
}

// renderHomeRow draws one home row: the mark, the title, and the note and
// relative time at the right. It mirrors rowDelegate's one-line rows.
func renderHomeRow(theme *Theme, selected bool, item homeRowItem, width int) string {
	title := singleLine(item.title)
	if title == "" {
		title = "Untitled"
	}
	note := singleLine(item.note)
	key := singleLine(item.key)

	styleTitle := func(text string) string {
		if selected {
			return theme.Selected.Render(text)
		}
		if item.status == PickWatched {
			return theme.Muted.Render(text)
		}
		return theme.Text.Render(text)
	}
	styleSide := func(text string) string {
		if selected {
			return theme.SelectedMuted.Render(text)
		}
		if item.status == PickInProgress || item.status == PickCurrent {
			return theme.Primary.UnsetBold().Render(text)
		}
		return theme.Muted.Render(text)
	}

	cursor := "   "
	if selected {
		cursor = " ▶ "
		if Plain() {
			cursor = "> "
		}
	}
	rawMark := statusMark(item.status)
	var mark string
	switch {
	case selected:
		mark = theme.Selected.Render(rawMark)
	case item.status == PickInProgress || item.status == PickCurrent:
		mark = theme.Primary.Render(rawMark)
	default:
		mark = theme.Muted.Render(rawMark)
	}

	var right string
	if note != "" {
		right += styleSide(note)
	}
	if key != "" {
		if right != "" {
			if selected {
				right += theme.Selected.Render("   ")
			} else {
				right += "   "
			}
		}
		if selected {
			right += theme.Selected.Render(key)
		} else {
			right += theme.Muted.Render(key)
		}
	}

	gutter := cursor + mark + " "
	left := gutter + styleTitle(title)
	if room := width - lipgloss.Width(left) - lipgloss.Width(right) - 1; room < minTitleWidth {
		left = gutter + styleTitle(ansi.Truncate(title, max(width-lipgloss.Width(gutter)-lipgloss.Width(right)-1, 1), "…"))
	}
	pad := max(width-lipgloss.Width(left)-lipgloss.Width(right), 0)
	line := left + strings.Repeat(" ", pad) + right
	if selected {
		// Fill the row to the edge like the other screens' cursor row.
		line += theme.Selected.UnsetBold().Render(strings.Repeat(" ", max(width-lipgloss.Width(line), 0)))
	}
	return ansi.Truncate(line, width, "")
}

// sectionCursor maps the flat cursor to a section index, or -1 when the
// highlight is in the other section.
func (m *searchHomeModel) sectionCursor(kind homeRowKind) int {
	ord := 0
	for i := range m.rows {
		if m.rows[i].kind != kind {
			continue
		}
		if i == m.cursor {
			return ord
		}
		ord++
	}
	return -1
}

// continuePanel is the DETAILS-style panel of a continue row: the episode,
// its progress, source and audio, and the row's keys.
func (m *searchHomeModel) continuePanel(e *HomeContinueEntry) *panelContent {
	title := e.DisplayTitle()
	sub := fmt.Sprintf("Episode %d", e.EpisodeNum)
	if ep := singleLine(e.EpisodeTitle); ep != "" {
		sub += " · " + ep
	}
	meta := ""
	if source := singleLine(e.Source); source != "" {
		meta = source
	}
	if audio := singleLine(e.Audio); audio != "" {
		if meta != "" {
			meta += " · "
		}
		meta += audio
	}
	lines := []string{sub}
	if meta != "" {
		lines = append(lines, meta)
	}
	return &panelContent{
		label: "Continue",
		gloss: "続き",
		title: title,
		lines: lines,
		progress: &panelProgress{
			position: float64(e.Position),
			duration: float64(e.Duration),
		},
		keys: []Hint{
			{Key: "enter", Desc: "resume"},
			{Key: "r", Desc: "from start"},
			{Key: "e", Desc: "episodes"},
			{Key: "x", Desc: "remove"},
		},
	}
}

// recentView renders the recent searches: the query and its result count.
func (m *searchHomeModel) recentView(width int) string {
	items := make([]homeRowItem, 0, len(m.recentAll))
	for _, r := range m.recentAll {
		count := fmt.Sprintf("%d results", r.Count)
		if r.Count == 1 {
			count = "1 result"
		}
		items = append(items, homeRowItem{title: singleLine(r.Query), note: count})
	}
	return renderHomeRows(&m.theme, items, m.sectionCursor(homeRowRecent), width)
}

type searchHomeRunner func(tea.Model) (tea.Model, error)

// SearchHome opens the SEARCH home screen and returns its decision.
func SearchHome(opts SearchHomeOptions) (SearchHomeResult, error) {
	if !HasTerminal() {
		return SearchHomeResult{}, fmt.Errorf("no terminal: stdin is not a TTY")
	}
	return searchHomeWithRunner(opts, func(model tea.Model) (tea.Model, error) {
		var final tea.Model
		err := RunClean(func() error {
			var runErr error
			final, runErr = NewProgram(model).Run()
			return runErr
		})
		return final, err
	})
}

// searchHomeWithRunner isolates terminal execution for deterministic tests.
func searchHomeWithRunner(opts SearchHomeOptions, run searchHomeRunner) (SearchHomeResult, error) {
	if run == nil {
		return SearchHomeResult{}, fmt.Errorf("search home runner not configured")
	}
	final, err := run(newSearchHomeModel(opts))
	if err != nil {
		return SearchHomeResult{}, fmt.Errorf("run search home screen: %w", err)
	}
	model, ok := final.(*searchHomeModel)
	if !ok || model == nil {
		return SearchHomeResult{}, fmt.Errorf("unexpected search home model %T", final)
	}
	if model.err != nil {
		return SearchHomeResult{}, model.err
	}
	if model.result.Action == HomeSearch && !model.result.HasEntry && strings.TrimSpace(model.result.Query) == "" {
		return SearchHomeResult{}, ErrSearchHomeCancelled
	}
	return model.result, nil
}
