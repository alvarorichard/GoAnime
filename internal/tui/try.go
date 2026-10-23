package tui

import (
	"fmt"
	"strings"

	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/charmbracelet/x/ansi"
)

// No-results screen (tela 4): the query found nothing. It keeps the query in
// an editable box, explains the outcome, offers TRY suggestions (a spelling
// fix, fewer words, English sources) and lists every source with its reason
// in the SOURCES panel.

// TrySuggestion is one TRY row: its label, the hint at the right, and the
// search it runs.
type TrySuggestion struct {
	Label string
	Hint  string
	Query string
	Scope string
}

// TrySource is one SOURCES panel row: an answering source (Ok, quiet "none")
// or a failed one (!Ok with its reason).
type TrySource struct {
	Name   string
	Ok     bool
	Detail string
}

// TryOptions configure the no-results screen.
type TryOptions struct {
	// Query is kept from the last search; typing replaces it.
	Query string
	// CurrentScope is the scope the failed search ran with.
	CurrentScope string
	// SourcesLine reads at the right of the header.
	SourcesLine string
	// Summary reads after the "!" block.
	Summary string
	Sources []TrySource
	Off     []models.SearchOffEntry
	// Suggestions are the TRY rows.
	Suggestions []TrySuggestion
	// Note reads under the rows (spelling differences between sources).
	Note string
}

// TryDecision is the retry the screen chose.
type TryDecision struct {
	Query string
	Scope string
}

type tryModel struct {
	theme         Theme
	shell         Shell
	input         textinput.Model
	initial       string
	replaced      bool
	scopeIdx      int
	searchSummary string
	sources       []TrySource
	off           []models.SearchOffEntry
	suggest       []TrySuggestion
	note          string
	cursor        int
	focusBox      bool
	result        TryDecision
	decided       bool
	err           error
	showHelp      bool
}

func newTryModelWithOptions(opts TryOptions) *tryModel {
	theme := CurrentTheme()
	shell := NewShell(&theme, "Search")
	shell.Context = singleLine(opts.SourcesLine)

	input := textinput.New()
	if Plain() {
		input.Prompt = "> "
	} else {
		input.Prompt = "❯ "
	}
	styles := textinput.DefaultStyles(themeIsDark())
	styles.Focused.Text = theme.Text
	styles.Focused.Placeholder = theme.Muted
	styles.Focused.Prompt = theme.Primary
	styles.Blurred = styles.Focused
	styles.Cursor.Color = theme.Primary.GetForeground()
	styles.Cursor.Shape = tea.CursorBlock
	input.SetStyles(styles)
	query := singleLine(opts.Query)
	input.SetValue(query)
	input.CursorEnd()
	input.Focus()

	m := &tryModel{
		theme:         theme,
		shell:         shell,
		input:         input,
		initial:       query,
		scopeIdx:      indexForScope(opts.CurrentScope),
		searchSummary: singleLine(opts.Summary),
		sources:       opts.Sources,
		off:           opts.Off,
		suggest:       opts.Suggestions,
		note:          singleLine(opts.Note),
		focusBox:      true,
	}
	m.sizeInput()
	return m
}

// indexForScope maps a source selector to the scope tab, nearest first.
func indexForScope(scope string) int {
	switch strings.ToLower(strings.TrimSpace(scope)) {
	case "":
		return 0
	case "ptbr", "pt-br", "animefire", "goyabu", "animesdigital":
		return 1
	case "hianime":
		return 2
	case "movietv", "startflix", "topcine", "pobreflix":
		return 3
	default:
		return 0
	}
}

func (m *tryModel) sizeInput() {
	width, _ := m.shell.ContentSize()
	m.input.SetWidth(max(width-lipgloss.Width(m.input.Prompt)-2, 8))
}

// scopeValue is the failed search's scope unless tabbed elsewhere. The tabs
// stay hidden on ALL (tela 4); touching tab reveals them.
func (m *tryModel) scopeValue() string {
	return homeScopes[m.scopeIdx].Value
}

// Init starts the cursor without background work.
func (m *tryModel) Init() tea.Cmd { return nil }

// Update handles resize, box editing, suggestion navigation and retry.
func (m *tryModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				m.err = ErrSelectionCancelled
				return m, tea.Quit
			default:
				return m, nil
			}
		}
		switch msg.String() {
		case "ctrl+c":
			m.err = ErrSelectionCancelled
			return m, tea.Quit
		case "?":
			if !m.focusBox {
				m.showHelp = true
				return m, nil
			}
		case "tab":
			m.scopeIdx = (m.scopeIdx + 1) % len(homeScopes)
			return m, nil
		case "esc":
			// esc edits: it clears the box, never quits and never leaves.
			if m.focusBox {
				if m.input.Value() != "" {
					m.input.SetValue("")
					m.replaced = true
				}
			} else {
				m.focusBox = true
				m.input.Focus()
			}
			return m, nil
		case "enter":
			if res, ok := m.confirm(); ok {
				m.result, m.decided = res, true
				return m, tea.Quit
			}
			return m, nil
		case "up":
			m.move(-1)
			return m, nil
		case "down":
			m.move(1)
			return m, nil
		}
		if m.focusBox {
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
		if isHomeTypeKey(msg) {
			// Typing over the suggestions jumps back to the box. The first
			// keystroke replaces the kept query outright.
			m.focusBox = true
			m.input.Focus()
			if !m.replaced {
				m.input.SetValue("")
				m.replaced = true
			}
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, cmd
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// move steps the highlight; down from the box enters the suggestions, up
// from the first row returns to it.
func (m *tryModel) move(delta int) {
	if len(m.suggest) == 0 {
		m.focusBox = true
		return
	}
	if m.focusBox {
		if delta > 0 {
			m.focusBox = false
			m.input.Blur()
		}
		return
	}
	next := m.cursor + delta
	if next < 0 {
		m.focusBox = true
		m.input.Focus()
		return
	}
	m.cursor = min(next, len(m.suggest)-1)
}

// confirm turns enter into a retry: the box text or a suggestion.
func (m *tryModel) confirm() (TryDecision, bool) {
	if m.focusBox {
		query := strings.TrimSpace(m.input.Value())
		if len([]rune(query)) < minHomeQuery {
			return TryDecision{}, false
		}
		return TryDecision{Query: query, Scope: m.scopeValue()}, true
	}
	if m.cursor < 0 || m.cursor >= len(m.suggest) {
		return TryDecision{}, false
	}
	s := m.suggest[m.cursor]
	if strings.TrimSpace(s.Query) == "" {
		return TryDecision{}, false
	}
	return TryDecision{Query: s.Query, Scope: s.Scope}, true
}

// hints lists the footer keys (tela 4).
func (m *tryModel) hints() []Hint {
	if m.showHelp {
		return []Hint{
			{Key: "?", Desc: "close", Essential: true},
			{Key: "esc", Desc: "close", Essential: true},
		}
	}
	tab := Hint{Key: "tab", Desc: "sources"}
	if name := scopeTabName(m.scopeIdx); name != "ALL" {
		tab.Desc = "sources: " + name
	}
	move := "↓"
	if Plain() {
		move = "dn"
	}
	return []Hint{
		{Key: "enter", Desc: "search", Essential: true},
		{Key: move, Desc: "suggestions"},
		tab,
		{Key: "?", Desc: "help"},
		{Key: "ctrl+c", Desc: "quit", Essential: true},
	}
}

// scopeTabName is the scope tab's name.
func scopeTabName(idx int) string {
	if idx < 0 || idx >= len(homeScopes) {
		return "ALL"
	}
	return homeScopes[idx].Name
}

// View renders the box, the notice, the TRY rows and the SOURCES panel.
func (m *tryModel) View() tea.View {
	width, height := m.shell.ContentSize()
	var body string
	if m.showHelp {
		if help := RenderHelp(&m.theme, "No results", []HelpRow{
			{Key: "enter", Desc: "search"},
			{Key: "esc", Desc: "clear"},
			{Key: "tab", Desc: "sources"},
		}, nil, width, height); help != "" {
			body = help
		}
	} else {
		body = m.tryView(width)
	}
	view := tea.NewView(m.shell.Render(body, m.hints()))
	view.AltScreen = true
	view.WindowTitle = "GoAnime - No results"
	return view
}

// tryView lays out the box, notice, suggestions and panel: the box and the
// notice span the terminal; the panel sits beside the TRY block on a wide
// terminal (as text below it in plain mode).
func (m *tryModel) tryView(width int) string {
	top := m.tryBox(width) + "\n" + renderNotice(&m.theme, m.summary(), width)
	lists := m.tryLists(width)
	panel := m.tryPanel(width - resultsListWidth(width))
	if panel == "" {
		return top + "\n\n" + lists
	}
	if Plain() {
		return top + "\n\n" + lists + "\n\n" + panel
	}
	listW := resultsListWidth(width)
	column := lipgloss.NewStyle().Width(listW).Render(m.tryLists(listW))
	return top + "\n\n" + lipgloss.JoinHorizontal(lipgloss.Top, column, panel)
}

// tryLists renders the TRY header, rows and note at listW cells wide.
func (m *tryModel) tryLists(listW int) string {
	lists := m.sectionHeader("TRY", "試す") + "\n" + m.suggestView(listW)
	if m.note != "" {
		lists += "\n"
		for line := range strings.SplitSeq(ansi.Wrap(m.note, listW, ""), "\n") {
			lists += "\n" + m.theme.Muted.Render(line)
		}
	}
	return lists
}

// tryBox draws the bordered query box with its kept-query hint.
func (m *tryModel) tryBox(width int) string {
	content := m.input.View()
	if !m.replaced && m.input.Value() == m.initial && m.initial != "" {
		hint := "kept from the last search, type to replace it"
		if Plain() {
			hint = "kept from the last search"
		}
		content += m.theme.Muted.Render("  " + hint)
	}
	border := lipgloss.NormalBorder()
	style := lipgloss.NewStyle().Border(border, true).Width(max(width-2, 8))
	style = style.BorderForeground(m.theme.Primary.GetForeground())
	box := style.Render(content)
	if name := scopeTabName(m.scopeIdx); name != "ALL" {
		box += "\n" + renderScopeTabs(&m.theme, m.scopeIdx)
	}
	return box
}

// suggestView renders the TRY rows.
func (m *tryModel) suggestView(width int) string {
	if len(m.suggest) == 0 {
		return ""
	}
	items := make([]homeRowItem, 0, len(m.suggest))
	for _, s := range m.suggest {
		items = append(items, homeRowItem{title: singleLine(s.Label), note: singleLine(s.Hint)})
	}
	selected := -1
	if len(items) > 0 {
		selected = min(max(m.cursor, 0), len(items)-1)
	}
	return renderHomeRows(&m.theme, items, selected, width)
}

// tryPanel draws the SOURCES panel for the failed search.
func (m *tryModel) tryPanel(width int) string {
	if len(m.sources) == 0 && len(m.off) == 0 {
		return ""
	}
	rows := make([]sourcesPanelRow, 0, len(m.sources))
	for _, s := range m.sources {
		rows = append(rows, sourcesPanelRow{
			Name: s.Name, Done: s.Ok, Failed: !s.Ok,
			Count: 0, HideTime: true, Reason: singleLine(s.Detail),
		})
	}
	_, height := m.shell.ContentSize()
	return renderSourcesPanel(&m.theme, rows, m.off, len(rows), len(rows), 0, false, width, height)
}

// sectionHeader draws a TRY heading with its Japanese gloss.
func (m *tryModel) sectionHeader(label, gloss string) string {
	head := m.theme.Primary.Render(label)
	if ShowJapanese() {
		head += m.theme.Gloss.Render("  " + gloss)
	}
	return head
}

// summary reads after the "!" block; renderNotice handles the empty case.
func (m *tryModel) summary() string { return m.searchSummary }

// renderScopeTabs draws the "in ALL PT-BR ENGLISH MOVIES/TV" row.
func renderScopeTabs(theme *Theme, idx int) string {
	var sb strings.Builder
	sb.WriteString(theme.Muted.Render("in "))
	for i, scope := range homeScopes {
		if i == idx {
			sb.WriteString(theme.Tag.Render(scope.Name))
		} else {
			sb.WriteString(theme.Muted.Render(scope.Name))
		}
		if i < len(homeScopes)-1 {
			sb.WriteString("   ")
		}
	}
	return sb.String()
}

type tryRunner func(tea.Model) (tea.Model, error)

// TrySearch opens the no-results screen and returns the retry it chose.
func TrySearch(opts TryOptions) (TryDecision, error) {
	if !HasTerminal() {
		return TryDecision{}, fmt.Errorf("no terminal: stdin is not a TTY")
	}
	return trySearchWithRunner(opts, func(model tea.Model) (tea.Model, error) {
		var final tea.Model
		err := RunClean(func() error {
			var runErr error
			final, runErr = NewProgram(model).Run()
			return runErr
		})
		return final, err
	})
}

// trySearchWithRunner isolates terminal execution for deterministic tests.
func trySearchWithRunner(opts TryOptions, run tryRunner) (TryDecision, error) {
	if run == nil {
		return TryDecision{}, fmt.Errorf("try search runner not configured")
	}
	final, err := run(newTryModelWithOptions(opts))
	if err != nil {
		return TryDecision{}, fmt.Errorf("run no-results screen: %w", err)
	}
	model, ok := final.(*tryModel)
	if !ok || model == nil {
		return TryDecision{}, fmt.Errorf("unexpected no-results model %T", final)
	}
	if model.err != nil {
		return TryDecision{}, model.err
	}
	if !model.decided {
		return TryDecision{}, ErrSelectionCancelled
	}
	return model.result, nil
}
