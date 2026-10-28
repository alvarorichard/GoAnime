package tui

import (
	"fmt"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
)

// Subtitles screen (tela 7): the episode's subtitle tracks as numbered rows,
// session checkboxes below, and the UP NEXT panel explaining that the choice
// is asked once.

// SubtitleTrack is one subtitle track choice.
type SubtitleTrack struct {
	Label    string
	Language string
}

// SubtitleOptions configure the subtitles screen.
type SubtitleOptions struct {
	// Title is the show for the "Title > Subtitles" breadcrumb ("" hides it).
	Title string
	// Context reads at the right of the header ("HiAnime · Sub", "" hides it).
	Context string
	// Tracks are the numbered rows. Both/None rows are added automatically.
	Tracks []SubtitleTrack
	// LastLabel marks the "your choice last time" row and preselects it.
	LastLabel string
	// SessionChecked pre-checks "rest of the session"; PreferChecked the
	// always-Portuguese rule.
	SessionChecked bool
	PreferChecked  bool
	// NextTitle and NextLine feed the UP NEXT panel ("" omits each line).
	NextTitle string
	NextLine  string
}

// SubtitleChoice is what the screen chose.
type SubtitleChoice struct {
	// Kind is "track", "all" (both) or "none".
	Kind     string
	Language string // the track's language, for Kind "track"
	Session  bool   // use the choice for the rest of the session
	PreferPT bool   // always pick Portuguese when a source has it
}

type subtitleRowKind int

const (
	subtitleRowTrack subtitleRowKind = iota
	subtitleRowBoth
	subtitleRowNone
	subtitleRowSession
	subtitleRowPrefer
)

type subtitleRow struct {
	kind     subtitleRowKind
	track    int // track index for track rows
	label    string
	language string
	key      string
	current  bool // "your choice last time"
	checked  bool // checkbox rows
}

type subtitlesModel struct {
	theme    Theme
	shell    Shell
	rows     []subtitleRow
	cursor   int
	session  bool
	preferPT bool
	next     panelContent
	hasNext  bool

	result   SubtitleChoice
	decided  bool
	err      error
	showHelp bool
}

// newSubtitlesModel creates the subtitles screen.
func newSubtitlesModel(opts SubtitleOptions) *subtitlesModel {
	theme := CurrentTheme()
	breadcrumb := "Subtitles"
	if title := singleLine(opts.Title); title != "" {
		breadcrumb = title + " > Subtitles"
	}
	shell := NewShell(&theme, breadcrumb)
	shell.Context = singleLine(opts.Context)

	m := &subtitlesModel{
		theme:    theme,
		shell:    shell,
		cursor:   0,
		session:  opts.SessionChecked,
		preferPT: opts.PreferChecked,
	}
	key := 0
	nextKey := func() string {
		key++
		if key > 9 {
			return ""
		}
		return fmt.Sprintf("%d", key)
	}
	for i, track := range opts.Tracks {
		label := singleLine(track.Label)
		if label == "" {
			label = singleLine(track.Language)
		}
		if label == "" {
			label = fmt.Sprintf("Subtitle %d", i+1)
		}
		m.rows = append(m.rows, subtitleRow{
			kind: subtitleRowTrack, track: i, label: label,
			language: singleLine(track.Language), key: nextKey(),
			current: singleLine(opts.LastLabel) != "" && singleLine(opts.LastLabel) == label,
		})
		if m.rows[len(m.rows)-1].current {
			m.cursor = len(m.rows) - 1
		}
	}
	if len(opts.Tracks) >= 2 {
		m.rows = append(m.rows, subtitleRow{kind: subtitleRowBoth, label: "Both, switch in mpv with j", key: nextKey()})
	}
	m.rows = append(m.rows,
		subtitleRow{kind: subtitleRowNone, label: "None", key: "0"},
		subtitleRow{kind: subtitleRowSession, label: "Use this choice for the rest of the session", checked: m.session},
		subtitleRow{kind: subtitleRowPrefer, label: "Always pick Portuguese when a source has it", checked: m.preferPT},
	)

	next := panelContent{label: "Up next", gloss: "次へ", title: singleLine(opts.NextTitle)}
	var lines []string
	if line := singleLine(opts.NextLine); line != "" {
		lines = append(lines, line)
	}
	lines = append(lines, "Asked once. The next episodes start with the same track.")
	next.lines = lines
	next.keys = []Hint{{Key: "t", Desc: "in the player menu changes it"}}
	m.next = next
	m.hasNext = true
	return m
}

// Init starts the screen without background work.
func (m *subtitlesModel) Init() tea.Cmd { return nil }

// Update handles navigation, digit picks, checkbox toggles and confirm.
func (m *subtitlesModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.shell.Resize(msg.Width, msg.Height)
		return m, nil

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
		switch msg.String() {
		case "ctrl+c":
			m.err = ErrPickCancelled
			return m, tea.Quit
		case "?":
			m.showHelp = true
			return m, nil
		case "esc":
			m.err = ErrPickBack
			return m, tea.Quit
		case "enter":
			if res, ok := m.confirm(); ok {
				m.result, m.decided = res, true
				return m, tea.Quit
			}
			return m, nil
		case "up", "k":
			m.cursor = max(m.cursor-1, 0)
			return m, nil
		case "down", "j":
			m.cursor = min(m.cursor+1, len(m.rows)-1)
			return m, nil
		case "space":
			m.toggle()
			return m, nil
		default:
			if row, ok := m.rowByKey(msg.String()); ok {
				if res, done := m.pick(row); done {
					m.result, m.decided = res, true
					return m, tea.Quit
				}
				return m, nil
			}
		}
	}
	return m, nil
}

// rowByKey finds the row a digit picks.
func (m *subtitlesModel) rowByKey(key string) (subtitleRow, bool) {
	if key == "" {
		return subtitleRow{}, false
	}
	for _, row := range m.rows {
		if row.key == key {
			return row, true
		}
	}
	return subtitleRow{}, false
}

// toggle flips the focused checkbox.
func (m *subtitlesModel) toggle() {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return
	}
	switch m.rows[m.cursor].kind {
	case subtitleRowSession:
		m.session = !m.session
		m.rows[m.cursor].checked = m.session
	case subtitleRowPrefer:
		m.preferPT = !m.preferPT
		m.rows[m.cursor].checked = m.preferPT
	}
}

// confirm turns enter into a choice: options confirm, checkboxes toggle.
func (m *subtitlesModel) confirm() (SubtitleChoice, bool) {
	if m.cursor < 0 || m.cursor >= len(m.rows) {
		return SubtitleChoice{}, false
	}
	row := m.rows[m.cursor]
	if res, done := m.pick(row); done {
		return res, true
	}
	m.toggle()
	return SubtitleChoice{}, false
}

// pick maps a row to its choice; checkboxes never confirm.
func (m *subtitlesModel) pick(row subtitleRow) (SubtitleChoice, bool) {
	base := SubtitleChoice{Session: m.session, PreferPT: m.preferPT}
	switch row.kind {
	case subtitleRowTrack:
		base.Kind = "track"
		base.Language = row.language
		return base, true
	case subtitleRowBoth:
		base.Kind = "all"
		return base, true
	case subtitleRowNone:
		base.Kind = "none"
		return base, true
	default:
		return SubtitleChoice{}, false
	}
}

// heading reads "This episode has 2 subtitle tracks".
func (m *subtitlesModel) heading(trackCount int) string {
	if trackCount == 1 {
		return "This episode has 1 subtitle track"
	}
	return fmt.Sprintf("This episode has %d subtitle tracks", trackCount)
}

// trackCount counts the track rows.
func (m *subtitlesModel) trackCount() int {
	n := 0
	for _, row := range m.rows {
		if row.kind == subtitleRowTrack {
			n++
		}
	}
	return n
}

// hints lists the footer keys (tela 7).
func (m *subtitlesModel) hints() []Hint {
	if m.showHelp {
		return []Hint{
			{Key: "?", Desc: "close", Essential: true},
			{Key: "esc", Desc: "close", Essential: true},
		}
	}
	digits := "1"
	if n := m.digitRange(); n != "" {
		digits = n
	}
	move := "↑↓"
	if Plain() {
		move = "up/dn"
	}
	return []Hint{
		{Key: "enter", Desc: "play", Essential: true},
		{Key: digits, Desc: "pick"},
		{Key: "space", Desc: "toggle"},
		{Key: move, Desc: "move"},
		{Key: "esc", Desc: "back", Essential: true},
	}
}

// digitRange reads the digit hints: "1-3 0" for two tracks plus Both.
func (m *subtitlesModel) digitRange() string {
	top := 0
	for _, row := range m.rows {
		if row.kind == subtitleRowTrack || row.kind == subtitleRowBoth {
			if row.key != "" && row.key != "0" {
				top++
			}
		}
	}
	if top == 0 {
		return "0"
	}
	if top == 1 {
		return "1 0"
	}
	return fmt.Sprintf("1-%d 0", top)
}

// View renders the tracks, the checkboxes and the UP NEXT panel.
func (m *subtitlesModel) View() tea.View {
	width, height := m.shell.ContentSize()
	var body string
	if m.showHelp {
		if help := RenderHelp(&m.theme, "Subtitles", []HelpRow{
			{Key: "enter", Desc: "play"},
			{Key: "space", Desc: "toggle"},
			{Key: "esc", Desc: "back"},
		}, nil, width, height); help != "" {
			body = help
		}
	} else {
		body = m.subtitlesView(width, height)
	}
	view := tea.NewView(m.shell.Render(body, m.hints()))
	view.AltScreen = true
	view.WindowTitle = "GoAnime - Subtitles"
	return view
}

// subtitlesView lays out the heading, rows and panel.
func (m *subtitlesModel) subtitlesView(width, height int) string {
	side := !Plain() && width >= wideResultsBreakpoint
	listW := width
	if side {
		// A panel beside the list narrows its rows, or the right side
		// wraps onto the next line.
		listW = resultsListWidth(width)
	}
	items := make([]homeRowItem, 0, len(m.rows))
	for _, row := range m.rows {
		title := row.label
		if row.kind == subtitleRowSession || row.kind == subtitleRowPrefer {
			box := "[ ]"
			if row.checked {
				box = "[x]"
			}
			title = box + " " + row.label
		}
		var status PickStatus
		if row.current {
			status = PickCurrent
		}
		note := ""
		if row.current {
			note = "your choice last time"
		}
		items = append(items, homeRowItem{status: status, title: title, note: note, key: row.key})
	}
	selected := -1
	if len(items) > 0 {
		selected = min(max(m.cursor, 0), len(items)-1)
	}
	lists := m.heading(m.trackCount()) + "\n\n" + renderHomeRows(&m.theme, items, selected, listW)
	if !m.hasNext {
		return lists
	}
	if side {
		if panel := fitPanel(&m.theme, &m.next, width-resultsListWidth(width), height); panel != "" {
			column := lipgloss.NewStyle().Width(resultsListWidth(width)).Render(lists)
			return lipgloss.JoinHorizontal(lipgloss.Top, column, panel)
		}
	}
	var panel string
	if Plain() {
		panel = renderPlainPanel(&m.theme, &m.next, width, height)
	} else {
		panel = fitPanel(&m.theme, &m.next, width, height)
	}
	if panel == "" {
		return lists
	}
	return lists + "\n\n" + panel
}

type subtitleRunner func(tea.Model) (tea.Model, error)

// SelectSubtitleTrack opens the subtitles screen and returns its choice.
func SelectSubtitleTrack(opts SubtitleOptions) (SubtitleChoice, error) {
	// Fail fast without a TTY (see PickActioned): Windows blocks on console
	// input while Linux errors. Headless callers (e.g. batch subtitle choice)
	// treat any error as "keep current tracks".
	if !HasTerminal() {
		return SubtitleChoice{}, fmt.Errorf("no terminal: stdin is not a TTY")
	}
	return selectSubtitleWithRunner(opts, func(model tea.Model) (tea.Model, error) {
		var final tea.Model
		err := RunClean(func() error {
			var runErr error
			final, runErr = NewProgram(model).Run()
			return runErr
		})
		return final, err
	})
}

// selectSubtitleWithRunner isolates terminal execution for deterministic tests.
func selectSubtitleWithRunner(opts SubtitleOptions, run subtitleRunner) (SubtitleChoice, error) {
	if run == nil {
		return SubtitleChoice{}, fmt.Errorf("subtitle runner not configured")
	}
	final, err := run(newSubtitlesModel(opts))
	if err != nil {
		return SubtitleChoice{}, fmt.Errorf("run subtitles screen: %w", err)
	}
	model, ok := final.(*subtitlesModel)
	if !ok || model == nil {
		return SubtitleChoice{}, fmt.Errorf("unexpected subtitles model %T", final)
	}
	if model.err != nil {
		return SubtitleChoice{}, model.err
	}
	if !model.decided {
		return SubtitleChoice{}, ErrPickCancelled
	}
	return model.result, nil
}
