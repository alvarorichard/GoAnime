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

// PickAction is a screen-level single-key shortcut, returned instead of a
// selection. The episodes screen uses r/d/D/A/m (tela 5): r resumes from the
// start, d/D/A download, m opens every option. Other screens leave empty.
type PickAction struct {
	Key   string // single character, checked before type-to-filter
	Label string // section and help row: "this episode"
	Hint  string // footer hint: "download"; "" hides it from the footer
	// Section lists the action under ActionsLabel ("Download"). r resumes
	// instead of downloading, so it stays out of the section.
	Section bool
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
	// Actions are screen-level single-key shortcuts (tela 5).
	Actions []PickAction
	// ActionsLabel heads the actions section ("Download"); "" hides it.
	ActionsLabel string
	// ActionsGloss is the heading's Japanese ("ダウンロード").
	ActionsGloss string
	// FilterPlaceholder is the filter line's hint.
	FilterPlaceholder string
	// BackDesc is what esc does, for the footer ("back", "results", ...).
	BackDesc string
	// HelpHint adds "? help" to the footer.
	HelpHint bool
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
	actionKey     string // screen action picked instead of a row
	err           error
	filterPending bool

	// shortcuts maps a row's key to its index in the items slice; keys
	// keeps them in row order for the footer.
	shortcuts map[string]int
	keys      []string
	// actionKeys maps a screen action's key to the actions slice.
	actionKeys map[string]int

	panel      *PickPanel
	progress   panelProgress
	panelSide  bool // beside the list rather than above it
	panelWidth int  // 0 when the panel does not fit
	// plainPanel is the panel as plain text below the list (tela 9).
	plainPanel string
	// showHelp replaces the list with the KEYS overlay (tela 10).
	showHelp bool
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
	theme := CurrentTheme()
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
	actionKeys := make(map[string]int)
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
	for i, action := range opts.Actions {
		if k := singleLine(action.Key); utf8.RuneCountInString(k) == 1 {
			if _, taken := actionKeys[k]; !taken {
				actionKeys[k] = i
			}
		}
	}

	// Dense rows: long catalogs (One Piece, etc.) show more episodes per screen.
	delegate := rowDelegate{theme: &theme, marks: marks}

	width, height := shell.ContentSize()
	entries := list.New(listItems, delegate, width, height)
	entries.KeyMap = fancyListKeyMap()
	entries.DisableQuitKeybindings() // quit/back owned by picker (ctrl+c / esc)
	placeholder := singleLine(opts.FilterPlaceholder)
	if placeholder == "" {
		placeholder = "type number or title…"
	}
	styleList(&entries, &theme, placeholder)
	if singleLine(opts.FilterPlaceholder) != "" {
		// Tela 5: the jump hint reads persistently above the list, not
		// only while filtering.
		prompt := drawGlyph("❯ ", "> ")
		entries.Title = theme.Primary.Render(prompt) + theme.Muted.Render(placeholder)
		entries.SetShowTitle(true)
	}
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
		actionKeys:    actionKeys,
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
		if m.showHelp {
			switch msg.String() {
			case "?", "esc":
				m.showHelp = false
				return m, nil
			case "ctrl+c":
				m.err = ErrPickCancelled
				return m, tea.Quit
			default:
				return m, nil
			}
		}
		filtering := m.entries.FilterState() == list.Filtering
		switch msg.String() {
		case "?":
			if !filtering {
				m.showHelp = true
				return m, nil
			}
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
				// Screen actions win over row shortcuts and filtering:
				// on the episodes screen r/d/D/A/m must fire anywhere.
				if _, ok := m.actionKeys[msg.String()]; ok {
					m.actionKey = msg.String()
					if entry, ok := m.entries.SelectedItem().(pickEntry); ok {
						m.selectedIndex = entry.index
					}
					return m, tea.Quit
				}
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
// In plain mode the panel is always plain text below the list (tela 9).
func (m *pickerModel) layout() {
	width, height := m.shell.ContentSize()
	m.panelSide, m.panelWidth, m.plainPanel = false, 0, ""
	listWidth, listHeight := width, height
	if m.panel != nil {
		if Plain() {
			if panel := fitPanel(&m.theme, m.panelContent(), width, height); panel != "" {
				m.panelWidth = width
				m.plainPanel = panel
				listWidth, listHeight = width, height-lipgloss.Height(panel)-1
			}
		} else if width >= wideResultsBreakpoint {
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
// a menu that has them. Screens with actions (tela 5) lead with enter and
// the action hints, like the design's episode footer.
func (m *pickerModel) hints() []Hint {
	if m.showHelp {
		return []Hint{
			{Key: "?", Desc: "close", Essential: true},
			{Key: "esc", Desc: "close", Essential: true},
		}
	}
	action := singleLine(m.options.Action)
	if action == "" {
		action = "select"
	}
	back := singleLine(m.options.BackDesc)
	if back == "" {
		back = "back"
	}
	if m.entries.FilterState() != list.Unfiltered {
		return []Hint{
			{Key: "↑↓", Desc: "move"},
			{Key: "enter", Desc: action, Essential: true},
			{Key: "esc", Desc: "clear filter", Essential: true},
		}
	}
	move := "↑↓"
	if Plain() {
		move = "up/dn"
	}
	if len(m.options.Actions) > 0 {
		hints := []Hint{{Key: "enter", Desc: action, Essential: true}}
		for _, a := range m.options.Actions {
			if a.Hint != "" {
				hints = append(hints, Hint{Key: singleLine(a.Key), Desc: singleLine(a.Hint)})
			}
		}
		hints = append(hints, Hint{Key: move, Desc: "move"})
		if m.options.HelpHint {
			hints = append(hints, Hint{Key: "?", Desc: "help"})
		}
		return append(hints, Hint{Key: "esc", Desc: back, Essential: true})
	}
	hints := []Hint{{Key: move, Desc: "move"}}
	if len(m.keys) > 0 {
		hints = append(hints, Hint{Key: strings.Join(m.keys, " "), Desc: "shortcuts"})
	} else {
		hints = append(hints, Hint{Key: "type", Desc: "filter"})
	}
	if len(m.entries.Items()) > m.entries.Paginator.PerPage {
		hints = append(hints, Hint{Key: "g/G", Desc: "first/last"})
	}
	hints = append(hints, Hint{Key: "enter", Desc: action, Essential: true})
	if m.options.HelpHint {
		hints = append(hints, Hint{Key: "?", Desc: "help"})
	}
	return append(hints, Hint{Key: "esc", Desc: back, Essential: true})
}

// actionsSection draws the actions block under the panel (tela 5's DOWNLOAD
// section): the heading and one line per section action.
func (m *pickerModel) actionsSection() string {
	var labels []PickAction
	for _, a := range m.options.Actions {
		if a.Section {
			labels = append(labels, a)
		}
	}
	if singleLine(m.options.ActionsLabel) == "" || len(labels) == 0 {
		return ""
	}
	head := m.theme.Primary.Render(strings.ToUpper(singleLine(m.options.ActionsLabel)))
	if gloss := singleLine(m.options.ActionsGloss); gloss != "" && ShowJapanese() {
		head += m.theme.Gloss.Render("  " + gloss)
	}
	lines := []string{head}
	for _, a := range labels {
		lines = append(lines, m.theme.Keycap.Render(singleLine(a.Key))+m.theme.Footer.Render("  "+singleLine(a.Label)))
	}
	return strings.Join(lines, "\n")
}

// View renders the list, with the panel beside or above it when there is one.
func (m *pickerModel) View() tea.View {
	width, height := m.shell.ContentSize()
	body := m.entries.View()
	if m.showHelp {
		if help := RenderHelp(&m.theme, helpScreenName(m.options.Breadcrumb), nil, m.helpExtra(), width, height); help != "" {
			body = help
		}
	} else if m.panelWidth > 0 {
		if Plain() {
			// The clock text follows playback; re-render, same height.
			if panel := fitPanel(&m.theme, m.panelContent(), width, height); panel != "" {
				body += "\n\n" + panel
				if section := m.actionsSection(); section != "" {
					body += "\n\n" + section
				}
			}
		} else if m.panelSide {
			side := ""
			if panel := fitPanel(&m.theme, m.panelContent(), m.panelWidth, height); panel != "" {
				side = panel
			}
			if section := m.actionsSection(); section != "" {
				if side != "" {
					side += "\n\n"
				}
				side += section
			}
			if side != "" {
				column := lipgloss.NewStyle().Width(resultsListWidth(width)).Render(body)
				body = lipgloss.JoinHorizontal(lipgloss.Top, column, side)
			}
		} else if panel := fitPanel(&m.theme, m.panelContent(), m.panelWidth, m.panelRoomAbove(height)); panel != "" {
			body = panel + "\n\n" + body
			if section := m.actionsSection(); section != "" {
				body += "\n\n" + section
			}
		} else if section := m.actionsSection(); section != "" {
			body += "\n\n" + section
		}
	} else if section := m.actionsSection(); section != "" {
		body += "\n\n" + section
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

// helpExtra documents the screen's keys in the KEYS overlay: row shortcuts
// and, when present, the screen actions under their section heading.
func (m *pickerModel) helpExtra() map[string][]HelpRow {
	extra := helpShortcuts(m.keys)
	for _, a := range m.options.Actions {
		if a.Label == "" {
			continue
		}
		section := singleLine(m.options.ActionsLabel)
		if section == "" {
			section = "Shortcuts"
		}
		if extra == nil {
			extra = make(map[string][]HelpRow)
		}
		extra[strings.ToUpper(section)] = append(extra[strings.ToUpper(section)], HelpRow{Key: singleLine(a.Key), Desc: singleLine(a.Label)})
	}
	return extra
}

// helpScreenName is the screen for the KEYS overlay: the last breadcrumb
// step, uppercased ("Episodes" from "Search > Episodes").
func helpScreenName(breadcrumb string) string {
	steps := breadcrumbSteps(breadcrumb)
	if len(steps) == 0 {
		return ""
	}
	return strings.ToUpper(steps[len(steps)-1])
}

// helpShortcuts documents a menu's single-key shortcuts in the overlay.
func helpShortcuts(keys []string) map[string][]HelpRow {
	if len(keys) == 0 {
		return nil
	}
	return map[string][]HelpRow{
		"SHORTCUTS": {{Key: strings.Join(keys, " "), Desc: "pick at once"}},
	}
}

// startFilterWithKey opens the list filter (as if "/" was pressed) then applies
// the typed character so "1000" works without a separate filter chord.
func (m *pickerModel) startFilterWithKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	filterBefore := m.entries.FilterValue()
	cmd := feedKeyToFilter(&m.entries, msg)
	if m.entries.FilterState() == list.Filtering && filterBefore != m.entries.FilterValue() {
		m.filterPending = true
	}
	return m, cmd
}

// feedKeyToFilter opens a list's filter as if "/" was pressed, then applies
// the typed key. Shared with the results screen so typing filters every list
// the same way (tela 3: "type to filter these results").
func feedKeyToFilter(entries *list.Model, msg tea.KeyPressMsg) tea.Cmd {
	var cmds []tea.Cmd
	var cmd tea.Cmd
	*entries, cmd = entries.Update(tea.KeyPressMsg{Code: '/', Text: "/"})
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	*entries, cmd = entries.Update(msg)
	if cmd != nil {
		cmds = append(cmds, cmd)
	}
	return tea.Batch(cmds...)
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
	index, _, err := PickActioned(items, opts)
	return index, err
}

// PickActioned opens a picker that also offers screen actions (tela 5). It
// returns the chosen item index, or -1 with the picked action's key.
func PickActioned(items []PickItem, opts PickOptions) (index int, action string, err error) {
	if len(items) == 0 {
		return -1, "", ErrNoPickItems
	}
	// Fail fast without a TTY: Bubble Tea blocks on Windows (getConsoleInput)
	// while it errors on Linux. Without this, headless CI on windows-latest
	// hangs the full go-test timeout instead of erroring like Linux.
	if !HasTerminal() {
		return -1, "", fmt.Errorf("no terminal: stdin is not a TTY")
	}
	return pickActionedWithRunner(items, opts, func(model tea.Model) (tea.Model, error) {
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
	index, _, err := pickActionedWithRunner(items, opts, run)
	return index, err
}

// pickActionedWithRunner is pickWithRunner with the picked screen action.
func pickActionedWithRunner(items []PickItem, opts PickOptions, run pickRunner) (index int, action string, err error) {
	if len(items) == 0 {
		return -1, "", ErrNoPickItems
	}
	if run == nil {
		return -1, "", fmt.Errorf("picker runner not configured")
	}

	final, err := run(newPickerModel(items, opts))
	if err != nil {
		return -1, "", fmt.Errorf("run picker screen: %w", err)
	}
	model, ok := final.(*pickerModel)
	if !ok || model == nil {
		return -1, "", fmt.Errorf("unexpected picker model %T", final)
	}
	if model.err != nil {
		return -1, "", model.err
	}
	if model.actionKey != "" {
		return model.selectedIndex, model.actionKey, nil
	}
	if model.selectedIndex < 0 || model.selectedIndex >= len(items) {
		return -1, "", ErrPickCancelled
	}
	return model.selectedIndex, "", nil
}
