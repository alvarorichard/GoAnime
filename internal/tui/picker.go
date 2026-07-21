package tui

import (
	"errors"
	"fmt"
	"strings"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
)

var (
	// ErrPickBack means the user requested the previous screen.
	ErrPickBack = errors.New("back from picker")
	// ErrPickCancelled means the user quit the picker screen.
	ErrPickCancelled = errors.New("picker cancelled")
	// ErrNoPickItems prevents opening an empty picker screen.
	ErrNoPickItems = errors.New("no items to pick")
)

// PickItem is one selectable row on a generic picker screen.
type PickItem struct {
	Label   string   // primary row text
	Details string   // secondary row text (optional)
	Preview []string // side panel lines on wide terminals (optional)
}

// PickOptions configure a generic picker screen.
type PickOptions struct {
	Breadcrumb   string // shell navigation trail, e.g. "Search > Seasons"
	WindowTitle  string // terminal window title
	ItemSingular string // status bar noun, e.g. "season"
	ItemPlural   string // status bar plural noun, e.g. "seasons"
	InitialIndex int    // preselected row in the items slice
}

type pickEntry struct {
	index int
	item  PickItem
}

// FilterValue exposes label and details to the fuzzy filter.
func (e pickEntry) FilterValue() string {
	return strings.TrimSpace(singleLine(e.item.Label) + " " + singleLine(e.item.Details))
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

type pickerModel struct {
	theme         Theme
	shell         Shell
	options       PickOptions
	entries       list.Model
	selectedIndex int
	err           error
	filterPending bool
}

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

	listItems := make([]list.Item, 0, len(items))
	for index, item := range items {
		listItems = append(listItems, pickEntry{index: index, item: item})
	}

	delegate := list.NewDefaultDelegate()
	// Dense rows: long catalogs (One Piece, etc.) show more episodes per screen.
	delegate.SetSpacing(0)
	delegate.Styles.NormalTitle = theme.Text.PaddingLeft(2)
	delegate.Styles.NormalDesc = theme.Muted.PaddingLeft(4)
	delegate.Styles.SelectedTitle = theme.SelectedTitle
	delegate.Styles.SelectedDesc = theme.SelectedDescription.PaddingLeft(1)
	delegate.Styles.DimmedTitle = theme.Muted.PaddingLeft(2)
	delegate.Styles.DimmedDesc = theme.Muted.Faint(true).PaddingLeft(4)
	delegate.Styles.FilterMatch = theme.FilterMatch

	width, height := shell.ContentSize()
	entries := list.New(listItems, delegate, width, height)
	entries.KeyMap = fancyListKeyMap()
	entries.DisableQuitKeybindings() // quit/back owned by picker (ctrl+c / esc)
	entries.SetShowTitle(false)
	entries.SetShowHelp(false)
	entries.SetStatusBarItemName(opts.ItemSingular, opts.ItemPlural)
	entries.FilterInput.Prompt = "❯ "
	entries.FilterInput.Placeholder = "type number or title…"
	entries.Styles.StatusBar = theme.Muted
	entries.Styles.StatusEmpty = theme.Muted
	entries.Styles.StatusBarActiveFilter = theme.Primary
	entries.Styles.StatusBarFilterCount = theme.Muted
	entries.Styles.NoItems = theme.Muted
	entries.Styles.PaginationStyle = theme.Muted
	entries.Styles.ActivePaginationDot = theme.Primary
	entries.Styles.InactivePaginationDot = theme.Muted
	entries.Styles.Filter.Focused.Prompt = theme.Primary
	entries.Styles.Filter.Focused.Text = theme.Text
	entries.Styles.Filter.Focused.Placeholder = theme.Muted
	entries.AdditionalShortHelpKeys = fancyListHelpKeys
	if opts.InitialIndex > 0 && opts.InitialIndex < len(listItems) {
		entries.Select(opts.InitialIndex)
	}

	return &pickerModel{
		theme:         theme,
		shell:         shell,
		options:       opts,
		entries:       entries,
		selectedIndex: -1,
	}
}

// Init starts the picker screen without background work.
func (m *pickerModel) Init() tea.Cmd {
	return nil
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
		width, height := m.shell.ContentSize()
		// Full-width list — no side preview panel (it only duplicated the row).
		m.entries.SetSize(width, height)
		return m, nil

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
			// fzf-style: typing while browsing opens the filter with that char.
			if !filtering && isTypeToFilterKey(msg) {
				return m.startFilterWithKey(msg)
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
		r := []rune(msg.Text)[0]
		return r >= 32 && r != 127
	}
	if msg.Code >= 32 && msg.Code < 127 {
		return true
	}
	return false
}

// View renders the full-width picker list (no side panel).
func (m *pickerModel) View() tea.View {
	footer := m.entries.Help.ShortHelpView(m.entries.ShortHelp())
	view := tea.NewView(m.shell.Render(m.entries.View(), footer))
	view.AltScreen = true
	title := m.options.WindowTitle
	if title == "" {
		title = "GoAnime"
	}
	view.WindowTitle = title
	return view
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
