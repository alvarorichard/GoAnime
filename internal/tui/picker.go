package tui

import (
	"errors"
	"fmt"
	"strings"
	"time"
	"unicode/utf8"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

var (
	// ErrPickBack means the user requested the previous screen.
	ErrPickBack = errors.New("back from picker")
	// ErrPickCancelled means the user quit the picker screen.
	ErrPickCancelled = errors.New("picker cancelled")
	// ErrNoPickItems prevents opening an empty picker screen.
	ErrNoPickItems = errors.New("no items to pick")
)

// PickStatus is the mark drawn before a row's label.
type PickStatus int

const (
	PickNone       PickStatus = iota
	PickWatched               // ✓ seen to the end
	PickInProgress            // ◐ stopped partway
	PickCurrent               // ● the one in use, such as the playing audio track
)

// PickItem is one selectable row on a generic picker screen.
type PickItem struct {
	Label   string // primary row text
	Details string // secondary text at the right of the row (optional)
	// Tags are short badges drawn after the label, such as "Filler".
	Tags   []string
	Status PickStatus
	// Note replaces Details at the right of the row, such as where playback
	// stopped.
	Note string
	// Key is a single character that picks the row at once. Rows with a key
	// keep type-to-filter for every other character.
	Key string
	// Meter, from 0 to 1, draws a bar at the right of the row in place of
	// Details, such as a resolution against the best one offered.
	Meter float64
}

// PickPanel is a paper panel shown beside the list on a wide terminal and
// above it on a narrow one.
type PickPanel struct {
	Label   string   // heading, drawn in capitals, such as "Now playing"
	Gloss   string   // the heading's Japanese, such as "再生中"
	Title   string   // bold first line
	Lines   []string // text under the title
	Command string   // a command line drawn on ink, such as a flag to try
	// Progress reports a position and a duration in seconds; ok false keeps
	// the last answer. It is asked once when the screen opens, and then every
	// second when Live is set, so the panel follows what mpv is playing.
	Progress func() (position, duration float64, ok bool)
	Live     bool
}

// PickOptions configure a generic picker screen.
type PickOptions struct {
	Breadcrumb   string // shell navigation trail, e.g. "Search > Seasons"
	WindowTitle  string // terminal window title
	ItemSingular string // status bar noun, e.g. "season"
	ItemPlural   string // status bar plural noun, e.g. "seasons"
	InitialIndex int    // preselected row in the items slice
	Action       string // what enter does, for the footer: "play", "open"; "select" when empty
	Context      string // right side of the header, such as the show being played
	Panel        *PickPanel
}

const (
	// listChromeRows is what a list draws around its rows: the filter line,
	// the status line and its gap, the page dots and their gap.
	listChromeRows = 5
	// minPanelListRows is how many rows a panel above the list leaves visible.
	minPanelListRows = 6
)

// narrowPanelWidth caps a panel drawn above the list.
const narrowPanelWidth = 64

type pickEntry struct {
	index int
	item  PickItem
}

// FilterValue exposes label, details and tags to the fuzzy filter, so typing
// "filler" finds the filler episodes.
func (e pickEntry) FilterValue() string {
	parts := []string{singleLine(e.item.Label), singleLine(e.item.Details)}
	for _, tag := range e.item.Tags {
		parts = append(parts, singleLine(tag))
	}
	return strings.Join(strings.Fields(strings.Join(parts, " ")), " ")
}

// Title returns the sanitized primary row text.
func (e pickEntry) Title() string {
	if label := singleLine(e.item.Label); label != "" {
		return label
	}
	return "Untitled"
}

// Description returns the sanitized secondary row text.
func (e pickEntry) Description() string {
	return singleLine(e.item.Details)
}

func (e pickEntry) row() row {
	return row{
		status: e.item.Status,
		title:  e.Title(),
		desc:   e.Description(),
		tags:   e.item.Tags,
		note:   e.item.Note,
		key:    singleLine(e.item.Key),
		meter:  e.item.Meter,
	}
}

type pickerModel struct {
	theme         Theme
	shell         Shell
	options       PickOptions
	entries       list.Model
	selectedIndex int
	err           error
	filterPending bool

	// shortcuts maps a row's key to its index in the items slice; keys
	// keeps them in row order for the footer.
	shortcuts map[string]int
	keys      []string

	panel      *PickPanel
	progress   panelProgress
	panelSide  bool // beside the list rather than above it
	panelWidth int  // 0 when the panel does not fit
}

type (
	progressTickMsg struct{}
	progressMsg     struct {
		position, duration float64
		ok                 bool
	}
)

// newPickerModel creates a styled, filterable picker screen.
func newPickerModel(items []PickItem, opts PickOptions) *pickerModel {
	theme := NewTheme(true)
	opts.Breadcrumb = singleLine(opts.Breadcrumb)
	opts.WindowTitle = singleLine(opts.WindowTitle)
	if opts.Breadcrumb == "" {
		opts.Breadcrumb = "Select"
	}
	if opts.ItemSingular == "" || opts.ItemPlural == "" {
		opts.ItemSingular, opts.ItemPlural = "item", "items"
	}
	shell := NewShell(&theme, opts.Breadcrumb)
	shell.Context = opts.Context

	listItems := make([]list.Item, 0, len(items))
	shortcuts := make(map[string]int)
	var keys []string
	marks := false
	for index, item := range items {
		listItems = append(listItems, pickEntry{index: index, item: item})
		if k := singleLine(item.Key); utf8.RuneCountInString(k) == 1 {
			if _, taken := shortcuts[k]; !taken {
				shortcuts[k] = index
				keys = append(keys, k)
			}
		}
		marks = marks || item.Status != PickNone
	}

	// Dense rows: long catalogs (One Piece, etc.) show more episodes per screen.
	delegate := rowDelegate{theme: &theme, marks: marks}

	width, height := shell.ContentSize()
	entries := list.New(listItems, delegate, width, height)
	entries.KeyMap = fancyListKeyMap()
	entries.DisableQuitKeybindings() // quit/back owned by picker (ctrl+c / esc)
	styleList(&entries, &theme, "type number or title…")
	entries.SetStatusBarItemName(opts.ItemSingular, opts.ItemPlural)
	entries.AdditionalShortHelpKeys = fancyListHelpKeys
	if opts.InitialIndex > 0 && opts.InitialIndex < len(listItems) {
		entries.Select(opts.InitialIndex)
	}

	m := &pickerModel{
		theme:         theme,
		shell:         shell,
		options:       opts,
		entries:       entries,
		selectedIndex: -1,
		shortcuts:     shortcuts,
		keys:          keys,
		panel:         opts.Panel,
	}
	if m.panel != nil && m.panel.Progress != nil && !m.panel.Live {
		m.setProgress(m.panel.Progress())
	}
	m.layout()
	return m
}

// Init starts polling the panel's progress when it follows playback.
func (m *pickerModel) Init() tea.Cmd {
	if m.panel != nil && m.panel.Progress != nil && m.panel.Live {
		return m.pollProgress()
	}
	return nil
}

func (m *pickerModel) pollProgress() tea.Cmd {
	poll := m.panel.Progress
	return func() tea.Msg {
		position, duration, ok := poll()
		return progressMsg{position: position, duration: duration, ok: ok}
	}
}

func (m *pickerModel) setProgress(position, duration float64, ok bool) {
	if ok {
		m.progress = panelProgress{position: position, duration: duration}
	}
}

// Update handles resize, filtering, navigation, selection, and cancellation.
// Printable keys type-to-filter like fzf (no need to press "/" first), so
// typing "1000" jumps straight to episode 1000 in long lists.
func (m *pickerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(list.FilterMatchesMsg); ok {
		m.filterPending = false
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.shell.Resize(msg.Width, msg.Height)
		m.layout()
		return m, nil

	case progressMsg:
		m.setProgress(msg.position, msg.duration, msg.ok)
		return m, tea.Tick(time.Second, func(time.Time) tea.Msg { return progressTickMsg{} })

	case progressTickMsg:
		poll := m.pollProgress()
		return m, poll

	case tea.KeyPressMsg:
		filtering := m.entries.FilterState() == list.Filtering
		switch msg.String() {
		case "ctrl+c":
			m.err = ErrPickCancelled
			return m, tea.Quit
		case "esc":
			if m.entries.FilterState() == list.Unfiltered {
				m.err = ErrPickBack
				return m, tea.Quit
			}
		case "enter":
			// fzf-style: Enter always confirms the highlighted row, even while
			// the filter is open — type "150" + Enter, done (no second Enter).
			if filtering && m.filterPending {
				return m, nil
			}
			if entry, ok := m.entries.SelectedItem().(pickEntry); ok {
				m.selectedIndex = entry.index
				return m, tea.Quit
			}
			if filtering {
				return m, nil
			}
		default:
			if !filtering {
				if index, ok := m.shortcuts[msg.String()]; ok {
					m.selectedIndex = index
					return m, tea.Quit
				}
				// fzf-style: typing while browsing opens the filter with that char.
				if isTypeToFilterKey(msg) {
					return m.startFilterWithKey(msg)
				}
			}
		}
	}

	filterBefore := m.entries.FilterValue()
	var cmd tea.Cmd
	m.entries, cmd = m.entries.Update(msg)
	if m.entries.FilterState() == list.Filtering && filterBefore != m.entries.FilterValue() {
		m.filterPending = true
	}
	return m, cmd
}

// layout sizes the list around the panel: beside it on a wide terminal,
// above it on a narrow one when both fit, and the list alone otherwise.
func (m *pickerModel) layout() {
	width, height := m.shell.ContentSize()
	m.panelSide, m.panelWidth = false, 0
	listWidth, listHeight := width, height
	if m.panel != nil {
		if width >= wideResultsBreakpoint {
			m.panelSide = true
			m.panelWidth = width - resultsListWidth(width)
			listWidth = resultsListWidth(width) - resultsGutter
		} else if panel := fitPanel(&m.theme, m.panelContent(), min(width, narrowPanelWidth), m.panelRoomAbove(height)); panel != "" {
			m.panelWidth = min(width, narrowPanelWidth)
			listHeight = height - lipgloss.Height(panel) - 1
		}
	}
	m.entries.SetSize(listWidth, max(listHeight, 1))
}

// panelRoomAbove is the height a panel above the list may take out of
// height: what is left once the list keeps its chrome, a blank line, and
// enough rows to still be a list.
func (m *pickerModel) panelRoomAbove(height int) int {
	return height - 1 - listChromeRows - min(len(m.entries.Items()), minPanelListRows)
}

func (m *pickerModel) panelContent() *panelContent {
	p := m.panel
	content := &panelContent{label: p.Label, gloss: p.Gloss, title: p.Title, lines: p.Lines, command: p.Command}
	if p.Progress != nil {
		content.progress = &m.progress
	}
	return content
}

// hints lists the footer keys, the shortcuts in place of type-to-filter on
// a menu that has them.
func (m *pickerModel) hints() []Hint {
	action := singleLine(m.options.Action)
	if action == "" {
		action = "select"
	}
	back := "back"
	if m.entries.FilterState() != list.Unfiltered {
		back = "clear filter"
	}
	hints := []Hint{{Key: "↑↓", Desc: "move"}}
	if len(m.keys) > 0 {
		hints = append(hints, Hint{Key: strings.Join(m.keys, " "), Desc: "shortcuts"})
	} else {
		hints = append(hints, Hint{Key: "type", Desc: "filter"})
	}
	if len(m.entries.Items()) > m.entries.Paginator.PerPage {
		hints = append(hints, Hint{Key: "g/G", Desc: "first/last"})
	}
	return append(hints,
		Hint{Key: "enter", Desc: action, Essential: true},
		Hint{Key: "esc", Desc: back, Essential: true})
}

// View renders the list, with the panel beside or above it when there is one.
func (m *pickerModel) View() tea.View {
	width, height := m.shell.ContentSize()
	body := m.entries.View()
	if m.panelWidth > 0 {
		if m.panelSide {
			if panel := fitPanel(&m.theme, m.panelContent(), m.panelWidth, height); panel != "" {
				column := lipgloss.NewStyle().Width(resultsListWidth(width)).Render(body)
				body = lipgloss.JoinHorizontal(lipgloss.Top, column, panel)
			}
		} else if panel := fitPanel(&m.theme, m.panelContent(), m.panelWidth, m.panelRoomAbove(height)); panel != "" {
			body = panel + "\n\n" + body
		}
	}

	view := tea.NewView(m.shell.Render(body, m.hints()))
	view.AltScreen = true
	title := m.options.WindowTitle
	if title == "" {
		title = "GoAnime"
	}
	view.WindowTitle = title
	return view
}

// startFilterWithKey opens the list filter (as if "/" was pressed) then applies
// the typed character so "1000" works without a separate filter chord.
func (m *pickerModel) startFilterWithKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	m.entries, cmd = m.entries.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	filterBefore := m.entries.FilterValue()
	m.entries, cmd = m.entries.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	if m.entries.FilterState() == list.Filtering && filterBefore != m.entries.FilterValue() {
		m.filterPending = true
	}
	return m, tea.Batch(cmds...)
}

// fancyListKeyMap is arrows + vim + emacs-style nav, shared with type-to-filter.
// Users can scroll with ↑↓ or hjkl, page with ^u/^d, jump with g/G, and still
// type digits/letters for fzf-style filtering.
func fancyListKeyMap() list.KeyMap {
	km := list.DefaultKeyMap()
	km.CursorUp = key.NewBinding(
		key.WithKeys("up", "k", "ctrl+p"),
		key.WithHelp("↑/k", "up"),
	)
	km.CursorDown = key.NewBinding(
		key.WithKeys("down", "j", "ctrl+n"),
		key.WithHelp("↓/j", "down"),
	)
	km.PrevPage = key.NewBinding(
		key.WithKeys("left", "h", "pgup", "ctrl+u", "b"),
		key.WithHelp("←/h/^u", "page up"),
	)
	km.NextPage = key.NewBinding(
		key.WithKeys("right", "l", "pgdown", "ctrl+d", "f"),
		key.WithHelp("→/l/^d", "page down"),
	)
	km.GoToStart = key.NewBinding(
		key.WithKeys("home", "g"),
		key.WithHelp("g/home", "first"),
	)
	km.GoToEnd = key.NewBinding(
		key.WithKeys("end", "G"),
		key.WithHelp("G/end", "last"),
	)
	km.Filter = key.NewBinding(
		key.WithKeys("/"),
		key.WithHelp("type//", "filter"),
	)
	km.Quit.SetEnabled(false)
	km.ForceQuit.SetEnabled(false)
	return km
}

// fancyListHelpKeys documents the dual input model: type-to-filter (fzf) and
// arrows/vim navigation.
func fancyListHelpKeys() []key.Binding {
	return []key.Binding{
		key.NewBinding(key.WithKeys("up", "down", "j", "k"), key.WithHelp("↑↓/jk", "move")),
		key.NewBinding(key.WithKeys("0", "1", "a", "z"), key.WithHelp("type", "fzf filter")),
		key.NewBinding(key.WithKeys("g", "G"), key.WithHelp("g/G", "first/last")),
		key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "select")),
		key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
	}
}

// isTypeToFilterKey reports printable keys that should open the filter instead
// of navigating. Navigation keys (arrows, vim hjkl, emacs ^n/^p, page, g/G)
// stay as movement so the list adapts to each user's habit.
func isTypeToFilterKey(msg tea.KeyPressMsg) bool {
	switch msg.String() {
	case "up", "down", "left", "right",
		"k", "j", "h", "l",
		"ctrl+p", "ctrl+n", "ctrl+u", "ctrl+d",
		"pgup", "pgdown", "home", "end",
		"g", "G", "b", "u", "f", "d",
		"enter", "esc", "tab", "shift+tab",
		"ctrl+c", "/",
		"space": // keep space for browsing; use letters/digits to filter
		return false
	}
	// Printable runes: digits, letters, punctuation (e.g. "1000", "one piece").
	if msg.Text != "" {
		r, _ := utf8.DecodeRuneInString(msg.Text)
		return r >= 32 && r != 127
	}
	if msg.Code >= 32 && msg.Code < 127 {
		return true
	}
	return false
}

type pickRunner func(tea.Model) (tea.Model, error)

// Pick opens a styled fuzzy picker and returns the chosen item index.
func Pick(items []PickItem, opts PickOptions) (int, error) {
	return pickWithRunner(items, opts, func(model tea.Model) (tea.Model, error) {
		var final tea.Model
		err := RunClean(func() error {
			var runErr error
			final, runErr = NewProgram(model).Run()
			return runErr
		})
		return final, err
	})
}

// PickLabels is a convenience for simple one-line menus (download options,
// yes/no, quality labels, post-playback actions). Same fancy list + fzf typing
// as Pick, without requiring callers to build PickItem values.
func PickLabels(labels []string, opts PickOptions) (int, error) {
	items := make([]PickItem, len(labels))
	for i, label := range labels {
		items[i] = PickItem{Label: label}
	}
	if opts.ItemSingular == "" {
		opts.ItemSingular = "option"
	}
	if opts.ItemPlural == "" {
		opts.ItemPlural = "options"
	}
	return Pick(items, opts)
}

// pickWithRunner isolates terminal execution for deterministic tests.
func pickWithRunner(items []PickItem, opts PickOptions, run pickRunner) (int, error) {
	if len(items) == 0 {
		return -1, ErrNoPickItems
	}
	if run == nil {
		return -1, fmt.Errorf("picker runner not configured")
	}

	final, err := run(newPickerModel(items, opts))
	if err != nil {
		return -1, fmt.Errorf("run picker screen: %w", err)
	}
	model, ok := final.(*pickerModel)
	if !ok || model == nil {
		return -1, fmt.Errorf("unexpected picker model %T", final)
	}
	if model.err != nil {
		return -1, model.err
	}
	if model.selectedIndex < 0 || model.selectedIndex >= len(items) {
		return -1, ErrPickCancelled
	}
	return model.selectedIndex, nil
}
