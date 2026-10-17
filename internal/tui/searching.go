package tui

import (
	"fmt"
	"sort"
	"strings"
	"time"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/charmbracelet/x/ansi"
)

// Searching screen (tela 2): results arrive live while every source answers.
// The SOURCES panel tracks each source; enter opens the highlighted title
// now instead of waiting; when every source is done the screen closes and
// the caller continues with the collected results.

// searchTickInterval refreshes the asking timers and drains source events.
const searchTickInterval = 100 * time.Millisecond

// SearchLiveOptions configure the SEARCHING screen.
type SearchLiveOptions struct {
	Query string
}

// SearchLiveOutcome is how the SEARCHING screen ended: Chosen when enter
// opened a title early, Completed when every source answered.
type SearchLiveOutcome struct {
	Chosen    *models.Anime
	Completed bool
}

// searchSource tracks one searched source's outcome.
type searchSource struct {
	name     string
	done     bool
	failed   bool
	count    int
	elapsed  time.Duration
	reason   string
	finished bool
}

type searchingModel struct {
	theme    Theme
	shell    Shell
	results  list.Model
	groups   []animeGroup
	all      []*models.Anime
	query    string
	events   <-chan models.SearchEvent
	searched []string
	off      []models.SearchOffEntry
	states   map[string]*searchSource
	started  time.Time
	giveUp   time.Duration
	planned  bool

	outcome  SearchLiveOutcome
	err      error
	showHelp bool
}

type searchTickMsg struct{}

// newSearchingModel creates the live search screen.
func newSearchingModel(query string, events <-chan models.SearchEvent) *searchingModel {
	theme := CurrentTheme()
	shell := NewShell(&theme, "Search > Searching")
	query = singleLine(query)
	if query != "" {
		shell.Context = `"` + query + `"`
	}
	delegate := rowDelegate{theme: &theme, twoLine: true, spacing: 1}
	m := &searchingModel{theme: theme, shell: shell, query: query, events: events, states: make(map[string]*searchSource)}
	width, height := shell.ContentSize()
	results := list.New(nil, delegate, width, height-1)
	keyMap := fancyListKeyMap()
	keyMap.Filter.SetEnabled(false) // no filter on this screen (tela 2)
	results.KeyMap = keyMap
	results.DisableQuitKeybindings()
	results.SetShowStatusBar(false) // the status line above the list counts instead (tela 2)
	styleList(&results, &m.theme, "")
	results.SetStatusBarItemName("title", "titles")
	m.results = results
	return m
}

// Init starts the event polling.
func (m *searchingModel) Init() tea.Cmd {
	return tea.Tick(searchTickInterval, func(time.Time) tea.Msg { return searchTickMsg{} })
}

// Update handles resize, event polling, navigation, early open and cancel.
func (m *searchingModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.shell.Resize(msg.Width, msg.Height)
		m.sizeList()
		return m, nil

	case searchTickMsg:
		m.drainEvents()
		if m.allTerminal() {
			m.outcome.Completed = true
			return m, tea.Quit
		}
		return m, tea.Tick(searchTickInterval, func(time.Time) tea.Msg { return searchTickMsg{} })

	case tea.KeyPressMsg:
		if m.showHelp {
			switch msg.String() {
			case "?", "esc":
				m.showHelp = false
				return m, nil
			case "ctrl+c":
				m.err = ErrSelectionCancelled
				return m, tea.Quit
			default:
				return m, nil
			}
		}
		switch msg.String() {
		case "?":
			m.showHelp = true
			return m, nil
		case "ctrl+c":
			m.err = ErrSelectionCancelled
			return m, tea.Quit
		case "q":
			return m, nil // ctrl+c quits; q never does here
		case "esc":
			m.err = ErrSelectionBack
			return m, tea.Quit
		case "enter":
			if item, ok := m.results.SelectedItem().(animeGroupItem); ok && item.group != nil && len(item.group.versions) > 0 {
				if anime := item.group.versions[0].anime; anime != nil {
					m.outcome.Chosen = anime
					return m, tea.Quit
				}
			}
			return m, nil
		}
	}

	var cmd tea.Cmd
	m.results, cmd = m.results.Update(msg)
	return m, cmd
}

// drainEvents applies every pending fan-out event.
func (m *searchingModel) drainEvents() {
	for {
		select {
		case ev, ok := <-m.events:
			if !ok {
				return
			}
			m.applyEvent(ev)
		default:
			return
		}
	}
}

// applyEvent folds one fan-out message into the screen.
func (m *searchingModel) applyEvent(ev models.SearchEvent) {
	if ev.Plan {
		m.searched = ev.Searched
		m.off = ev.Off
		m.giveUp = ev.GiveUp
		m.started = time.Now()
		m.planned = true
		for _, name := range ev.Searched {
			m.states[name] = &searchSource{name: name}
		}
		m.sizeList()
		return
	}
	st, ok := m.states[ev.Name]
	if !ok {
		st = &searchSource{name: ev.Name}
		m.states[ev.Name] = st
	}
	st.finished = true
	st.elapsed = ev.Elapsed
	switch ev.State {
	case models.SearchDone:
		st.done = true
		st.count = ev.Count
		for _, anime := range ev.Results {
			if anime != nil {
				m.all = append(m.all, anime)
			}
		}
		m.rebuildGroups()
	case models.SearchFailed:
		st.failed = true
		st.reason = singleLine(ev.Reason)
	}
	m.sizeList()
}

// rebuildGroups regroups the arrived results, closest match first, keeping
// the highlight on its work.
func (m *searchingModel) rebuildGroups() {
	selected := ""
	if item, ok := m.results.SelectedItem().(animeGroupItem); ok && item.group != nil {
		selected = item.group.title
	}
	m.groups = groupAnimeResults(m.all)
	groupScore := func(g *animeGroup) int {
		best := 0
		for _, v := range g.versions {
			if v.anime == nil {
				continue
			}
			if s := models.RelevanceScore(m.query, v.anime.Name, v.anime.Year); s > best {
				best = s
			}
		}
		return best
	}
	sort.SliceStable(m.groups, func(i, j int) bool {
		// Recomputed per comparison (n is small): a precomputed index
		// table would go stale as sort moves elements mid-sort.
		si, sj := groupScore(&m.groups[i]), groupScore(&m.groups[j])
		if si != sj {
			return si > sj
		}
		return liveLanguagePriority(m.groups[i].versions[0].anime.Name) < liveLanguagePriority(m.groups[j].versions[0].anime.Name)
	})
	digits := max(len(fmt.Sprintf("%d", len(m.groups))), 2)
	items := make([]list.Item, 0, len(m.groups))
	at := 0
	for i := range m.groups {
		items = append(items, animeGroupItem{group: &m.groups[i], pos: i, number: fmt.Sprintf("%0*d", digits, i+1)})
		if selected != "" && m.groups[i].title == selected {
			at = i
		}
	}
	m.results.SetItems(items)
	if len(items) > 0 {
		m.results.Select(at)
	}
	m.sizeList()
}

// liveLanguagePriority sorts live results like the result screen: Portuguese
// first. It mirrors api.languagePriority, which tui cannot import.
func liveLanguagePriority(name string) int {
	lower := strings.ToLower(name)
	if strings.Contains(lower, "[pt-br]") || strings.Contains(lower, "[portuguese]") || strings.Contains(lower, "[português]") {
		return 0
	}
	switch {
	case strings.HasPrefix(lower, "[multilanguage]"):
		return 1
	case strings.HasPrefix(lower, "[english]"):
		return 2
	case strings.HasPrefix(lower, "[movie]") || strings.HasPrefix(lower, "[tv]") || strings.HasPrefix(lower, "[movies/tv]"):
		return 3
	default:
		return 4
	}
}

// terminalCount is how many searched sources answered or failed.
func (m *searchingModel) terminalCount() int {
	n := 0
	for _, name := range m.searched {
		if st, ok := m.states[name]; ok && st.finished {
			n++
		}
	}
	return n
}

// pendingCount is how many searched sources are still asking.
func (m *searchingModel) pendingCount() int {
	return len(m.searched) - m.terminalCount()
}

// allTerminal reports whether every searched source answered. With no plan
// yet it is false: the screen waits for the fan-out, not for nothing.
func (m *searchingModel) allTerminal() bool {
	return m.planned && m.pendingCount() == 0
}

// sizeList fits the list beside the sources panel on a wide terminal,
// leaving room for the status line above and the skeleton rows of the
// sources still asking.
func (m *searchingModel) sizeList() {
	width, height := m.shell.ContentSize()
	listWidth := width
	if width >= wideResultsBreakpoint {
		listWidth = resultsListWidth(width) - resultsGutter
	}
	room := height - 1 // the status line above
	if m.pendingCount() > 0 {
		room -= 2 // the skeleton rows below
	}
	m.results.SetSize(listWidth, max(room, 1))
}

// statusLine reads "3 results so far, more arriving · enter opens one now".
func (m *searchingModel) statusLine() string {
	mark := "⁖ "
	if Plain() {
		mark = "... "
	}
	n := len(m.groups)
	if n == 0 {
		if Plain() {
			return "... asking the sources..."
		}
		return "⁖ asking the sources…"
	}
	sep := " · "
	if Plain() {
		sep = " - "
	}
	results := "result"
	if n != 1 {
		results = "results"
	}
	line := fmt.Sprintf("%d %s so far, more arriving%senter opens one now", n, results, sep)
	return m.theme.Muted.Render(mark) + m.theme.Text.Render(line)
}

// hints lists the footer keys (tela 2).
func (m *searchingModel) hints() []Hint {
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
		{Key: "enter", Desc: "open now", Essential: true},
		{Key: move, Desc: "move"},
		{Key: "esc", Desc: "edit search", Essential: true},
		{Key: "ctrl+c", Desc: "quit"},
	}
}

// View renders the live results with the sources panel.
func (m *searchingModel) View() tea.View {
	width, height := m.shell.ContentSize()
	var body string
	if m.showHelp {
		if help := RenderHelp(&m.theme, "Searching", []HelpRow{
			{Key: "enter", Desc: "open now"},
			{Key: "esc", Desc: "edit the search"},
			{Key: "ctrl+c", Desc: "quit"},
		}, nil, width, height); help != "" {
			body = help
		}
	} else {
		body = m.searchView(width, height)
	}
	view := tea.NewView(m.shell.Render(body, m.hints()))
	view.AltScreen = true
	view.WindowTitle = "GoAnime - Searching"
	return view
}

// searchView lays out the status line, the live list with skeleton rows for
// the sources still asking, and the sources panel beside them.
func (m *searchingModel) searchView(width, _ int) string {
	left := m.statusLine() + "\n" + m.results.View()
	if m.pendingCount() > 0 {
		left += "\n" + m.skeletons(width)
	}
	panel := m.panelRows(width - resultsListWidth(width))
	if panel == "" {
		return left
	}
	if Plain() {
		// Tela 9: sources as text below the results.
		return left + "\n\n" + panel
	}
	column := lipgloss.NewStyle().Width(resultsListWidth(width)).Render(left)
	return lipgloss.JoinHorizontal(lipgloss.Top, column, panel)
}

// skeletons draws placeholder bars for the sources still asking.
func (m *searchingModel) skeletons(width int) string {
	listWidth := width
	if width >= wideResultsBreakpoint {
		listWidth = resultsListWidth(width) - resultsGutter
	}
	char := "░"
	if Plain() {
		char = "-"
	}
	bar := func(frac float64) string {
		return m.theme.Faint.Render(strings.Repeat(char, max(int(float64(listWidth)*frac), 8)))
	}
	return bar(0.45) + "\n" + bar(0.3)
}

// panelRows draws the SOURCES panel for the current fan-out state.
func (m *searchingModel) panelRows(width int) string {
	rows := make([]sourcesPanelRow, 0, len(m.searched))
	now := time.Now()
	for _, name := range m.searched {
		st, ok := m.states[name]
		if !ok {
			st = &searchSource{name: name}
		}
		elapsed := st.elapsed
		if !st.finished {
			elapsed = now.Sub(m.started)
		}
		rows = append(rows, sourcesPanelRow{
			Name: name, Done: st.finished && !st.failed, Failed: st.failed,
			Count: st.count, Elapsed: elapsed, Reason: st.reason,
		})
	}
	_, height := m.shell.ContentSize()
	return renderSourcesPanel(&m.theme, rows, m.off, m.terminalCount(), len(m.searched), m.giveUp, true, width, height)
}

// sourcesPanelRow is one searched source on the SOURCES panel.
type sourcesPanelRow struct {
	Name     string
	Done     bool
	Failed   bool
	Count    int
	Elapsed  time.Duration // live timer for asking rows
	HideTime bool          // the TRY screen lists no durations
	Reason   string        // failure reason underneath, or "" for none
}

// sourcesPanel draws the SOURCES panel: every searched source with its
// outcome, the sitters-out, and, unless showProgress is false, a progress
// bar with the give-up budget. It returns "" when too short to be useful.
func renderSourcesPanel(theme *Theme, rows []sourcesPanelRow, off []models.SearchOffEntry, done, total int, giveUp time.Duration, showProgress bool, width, maxHeight int) string {
	if width < minPanelWidth || (len(rows) == 0 && len(off) == 0) {
		return ""
	}
	panelWidth := width - 1 // the shadow column
	inner := panelWidth - 4 // two cells of padding each side

	var lines []string
	head := theme.PanelLabel.Render("SOURCES")
	if !Plain() {
		head += theme.PanelMuted.Render("  ソース")
	}
	lines = append(lines, head)

	for _, row := range rows {
		lines = append(lines, sourcePanelLines(theme, row, inner)...)
	}
	for _, entry := range off {
		lines = append(lines, offPanelLine(theme, entry, inner))
	}

	if showProgress {
		lines = append(lines, renderBar(inner, float64(done)/float64(max(total, 1)), &theme.PanelBar, &theme.PanelTrack))
		sep := " · "
		if Plain() {
			sep = " - "
		}
		footer := fmt.Sprintf("%d of %d done%sgives up at %s", done, total, sep, shortDuration(giveUp))
		lines = append(lines, theme.PanelText.Render(ansi.Truncate(footer, inner, "…")))
	}

	if Plain() {
		// Tela 9: sources as text, no paper panel, shadow or halftone.
		return strings.Join(lines, "\n")
	}
	if len(lines)+2+1 > maxHeight {
		return ""
	}
	return addShadow(theme, renderPanelBlock(theme, lines, panelWidth))
}

// sourcePanelLines draws one searched source: its mark, name, count and
// time, plus the failure reason underneath when it failed.
func sourcePanelLines(theme *Theme, row sourcesPanelRow, inner int) []string {
	mark, nameStyle, rightStyle := sourcePanelMark(theme, row)
	right := ""
	switch {
	case row.Failed:
		right = ""
	case row.Done && row.Count == 0:
		right = "none"
	default:
		if row.Done {
			if row.Count == 1 {
				right = "1 found"
			} else {
				right = fmt.Sprintf("%d found", row.Count)
			}
		} else {
			right = "asking..."
		}
	}
	if !row.HideTime {
		stamp := formatElapsed(row.Elapsed)
		if right != "" {
			stamp = "  " + stamp
		}
		if row.Done || row.Elapsed > 0 {
			right += stamp
		}
	}
	line := mark + panelGap(theme, 1) + nameStyle.Render(ansi.Truncate(row.Name, inner, "…"))
	if right != "" {
		gap := inner - lipgloss.Width(line) - lipgloss.Width(rightStyle.Render(right))
		if gap < 1 {
			line = mark + panelGap(theme, 1) + nameStyle.Render(ansi.Truncate(row.Name, max(inner-lipgloss.Width(rightStyle.Render(right))-2, 1), "…"))
			gap = inner - lipgloss.Width(line) - lipgloss.Width(rightStyle.Render(right))
		}
		line += panelGap(theme, max(gap, 1)) + rightStyle.Render(right)
	}
	out := []string{line}
	if row.Failed && row.Reason != "" {
		for line := range strings.SplitSeq(ansi.Wrap("  "+row.Reason, inner, ""), "\n") {
			out = append(out, theme.PanelMuted.Render(line))
		}
	}
	return out
}

// sourcePanelMark is the panel mark per outcome: ✓ done, ⁖ asking, ! failed.
func sourcePanelMark(theme *Theme, row sourcesPanelRow) (mark string, name, right lipgloss.Style) {
	if Plain() {
		switch {
		case row.Failed:
			return theme.PanelLabel.Render("!"), theme.PanelText, theme.PanelMuted
		case row.Done:
			return theme.PanelText.Render("+"), theme.PanelText, theme.PanelMuted
		default:
			return theme.PanelMuted.Render("~"), theme.PanelMuted, theme.PanelMuted
		}
	}
	switch {
	case row.Failed:
		return theme.Notice.Render("!"), theme.PanelText, theme.PanelMuted
	case row.Done:
		return theme.PanelLabel.Render("✓"), theme.PanelText, theme.PanelMuted
	default:
		return theme.PanelMuted.Render("⁖"), theme.PanelMuted, theme.PanelMuted
	}
}

// panelGap pads a paper-panel line with the panel background. Plain spaces
// would show the terminal's own background through the paper — a black bar
// on dark terminals, a white one on light ones.
func panelGap(theme *Theme, n int) string {
	if n <= 0 {
		return ""
	}
	return theme.PanelText.Render(strings.Repeat(" ", n))
}

// offPanelLine draws a sitting-out source: "– Goyabu  off by default".
func offPanelLine(theme *Theme, off models.SearchOffEntry, inner int) string {
	mark := "–"
	if Plain() {
		mark = "-"
	}
	line := theme.PanelMuted.Render(mark+" ") + theme.PanelText.Render(ansi.Truncate(off.Name, inner, "…"))
	right := theme.PanelMuted.Render(off.Reason)
	gap := inner - lipgloss.Width(line) - lipgloss.Width(right)
	if gap < 1 {
		return line
	}
	return line + panelGap(theme, gap) + right
}

// formatElapsed reads "0.8s".
func formatElapsed(d time.Duration) string {
	if d < 0 {
		d = 0
	}
	return fmt.Sprintf("%.1fs", d.Seconds())
}

// shortDuration reads "15s".
func shortDuration(d time.Duration) string {
	if d <= 0 {
		return "—"
	}
	return fmt.Sprintf("%ds", int(d.Seconds()))
}

type searchLiveRunner func(tea.Model) (tea.Model, error)

// SearchLiveResults opens the SEARCHING screen over events and returns how
// it ended: the early pick, the completed wait, or the cancel/back error.
func SearchLiveResults(opts SearchLiveOptions, events <-chan models.SearchEvent) (SearchLiveOutcome, error) {
	if !HasTerminal() {
		return SearchLiveOutcome{}, fmt.Errorf("no terminal: stdin is not a TTY")
	}
	return searchLiveWithRunner(opts, events, func(model tea.Model) (tea.Model, error) {
		var final tea.Model
		err := RunClean(func() error {
			var runErr error
			final, runErr = NewProgram(model).Run()
			return runErr
		})
		return final, err
	})
}

// searchLiveWithRunner isolates terminal execution for deterministic tests.
func searchLiveWithRunner(opts SearchLiveOptions, events <-chan models.SearchEvent, run searchLiveRunner) (SearchLiveOutcome, error) {
	if run == nil {
		return SearchLiveOutcome{}, fmt.Errorf("search live runner not configured")
	}
	final, err := run(newSearchingModel(opts.Query, events))
	if err != nil {
		return SearchLiveOutcome{}, fmt.Errorf("run searching screen: %w", err)
	}
	model, ok := final.(*searchingModel)
	if !ok || model == nil {
		return SearchLiveOutcome{}, fmt.Errorf("unexpected searching model %T", final)
	}
	if model.err != nil {
		return SearchLiveOutcome{}, model.err
	}
	return model.outcome, nil
}
