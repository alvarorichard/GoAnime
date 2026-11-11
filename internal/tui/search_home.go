package tui

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	bubblekey "charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/textinput"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"golang.org/x/text/unicode/norm"
)

var (
	// ErrSearchHomeBack means the user went back from the search home screen.
	ErrSearchHomeBack = errors.New("back from search home")
	// ErrSearchHomeCancelled means the user quit the search home screen.
	ErrSearchHomeCancelled = errors.New("search home cancelled")
)

// searchScopes are the search tabs: the row under the search box. Value is
// the source selector the search runs with ("", "ptbr", "hianime" or
// "movietv").
var searchScopes = []struct{ Name, Value string }{
	{"ALL", ""},
	{"PT-BR", "ptbr"},
	{"ENGLISH", "hianime"},
	{"MOVIES/TV", "movietv"},
}

// homeScopes are the home screen's tabs: the search tabs, then FAVORITES.
var homeScopes = append(slices.Clone(searchScopes), struct{ Name, Value string }{"FAVORITES", favoritesScope})

// favoritesScope is the FAVORITES tab's value. It is not a search scope: on
// that tab the box filters the user's favorites and nothing is searched until
// one is opened, and a result never carries it.
const favoritesScope = "favorites"

// favoritesPlaceholder is the box's hint on the FAVORITES tab.
const favoritesPlaceholder = "type to filter your favorites, enter opens the first one"

// homePageRows is how far pgup and pgdown move on a list.
const homePageRows = 10

// homeFavoritesShown is how many FAVORITES rows the home screen draws; the
// rest wait behind a "+N more" line and slide in as rows are removed.
const homeFavoritesShown = 8

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

// HomeFavorite is one row of the FAVORITES list.
type HomeFavorite struct {
	Key    string // favorites key, for removal
	Title  string
	Year   string
	Media  string // "Anime", "Movie" or "TV"
	Source string
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
	// Favorites is the FAVORITES list: the works the user marked. Enter
	// searches the title. DeleteFavorite removes a row by key and
	// ClearFavorites empties the list (X, after a y/n question); nil disables
	// x or X there.
	Favorites []HomeFavorite
	// Lists turns the FAVORITES tab into two levels: first the user's lists
	// (Favorites, then the ones they made), then the works of the list they
	// open. nil keeps the single list of Favorites above.
	Lists FavoriteLists
	// Notice is a line to show when the screen opens, such as a warning that
	// a saved file was damaged. It goes at the first key.
	Notice         string
	DeleteFavorite func(key string) error
	ClearFavorites func() error
	// DeleteEntry removes a CONTINUE row by key; nil disables x there.
	DeleteEntry func(key string) error
	// DeleteRecent removes a RECENT row by its query; nil disables x there.
	DeleteRecent func(query string) error
	// ClearContinue and ClearRecent empty a whole list (X, after a y/n
	// question); nil disables X on that list.
	ClearContinue func() error
	ClearRecent   func() error
	// Suggest feeds the autocomplete under the box: it returns titles for the
	// typed query in the given scope. nil turns the autocomplete off. It runs
	// off the UI loop and must honour ctx, which is cancelled when the user
	// types again.
	Suggest SuggestFunc
	// Sources lists the sources for the ctrl+s screen and SetSource saves a
	// choice. Either nil leaves the screen out.
	Sources   LoadSourcesFunc
	SetSource SetSourceFunc

	// Restore reopens the screen where the user left it when they come back
	// from a search it started: its tab, row and box win over ScopeIndex and
	// Query. The zero value changes nothing.
	Restore HomeRestore
}

// HomeSuggestion is one autocomplete row: the title that fills the box and a
// short detail at the right ("Anime · 2002").
type HomeSuggestion struct {
	Title  string
	Detail string
}

// SuggestFunc produces autocomplete rows for query in scope.
type SuggestFunc func(ctx context.Context, query, scope string) ([]HomeSuggestion, error)

// Autocomplete timing and size: the debounce waits for a pause in typing, the
// timeout bounds one lookup, and at most maxHomeSuggestions rows show.
const (
	suggestDebounce    = 180 * time.Millisecond
	suggestTimeout     = 4 * time.Second
	maxHomeSuggestions = 6
)

// suggestDueMsg fires after the debounce; it is stale unless seq is current.
type suggestDueMsg struct{ seq int }

// suggestResultMsg carries a lookup's answer back to the UI loop.
type suggestResultMsg struct {
	seq   int
	items []HomeSuggestion
	err   error
}

// SearchHomeResult is the home screen's decision.
type SearchHomeResult struct {
	Action    HomeAction
	Query     string
	Scope     string
	Entry     HomeContinueEntry
	HasEntry  bool
	Cancelled bool

	// Restore is where the screen was, to reopen it there after this decision.
	Restore HomeRestore
}

type homeRowKind int

const (
	homeRowContinue homeRowKind = iota
	homeRowRecent
	homeRowFavorite
	homeRowList    // a list on the first level of the FAVORITES tab
	homeRowNewList // "+ Create a new list" on that level
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

// clearTarget is the list a "clear all?" question is about.
type clearTarget int

const (
	clearNothing clearTarget = iota
	clearContinueRows
	clearRecentRows
	clearFavoriteRows
	clearListWorks // X on the works of a list (see home_lists.go)
	clearDeleteList
)

type searchHomeModel struct {
	theme Theme
	shell Shell
	input textinput.Model
	// searchPlaceholder is the box's hint on the search tabs.
	searchPlaceholder string
	scopes            []struct{ Name, Value string }
	scopeIdx          int
	continueAll       []HomeContinueEntry
	recentAll         []HomeRecentSearch
	favoriteAll       []HomeFavorite
	lists             *homeLists // nil unless Options.Lists is set
	rows              []homeRow
	cursor            int
	focusInput        bool
	deleteEntry       func(key string) error
	now               func() time.Time

	// Forgetting history: the RECENT row remover, the two list clearers, the
	// list the user is being asked to clear (clearNothing while no question is
	// open) and what to say when a removal failed.
	deleteRecent   func(query string) error
	deleteFavorite func(key string) error
	clearContinue  func() error
	clearRecent    func() error
	clearFavorites func() error
	clearing       clearTarget
	notice         string

	// Autocomplete state. seq numbers the typing bursts: a debounce tick or a
	// lookup answer with an older seq is stale and dropped. asked is the
	// query+scope the current suggestions (or pending lookup) belong to.
	suggest     SuggestFunc
	debounce    time.Duration
	suggestions []HomeSuggestion
	sugCursor   int // -1: the box itself is highlighted
	seq         int
	asked       string
	cancel      context.CancelFunc

	// The ctrl+s sources screen: sources is non-nil while it is open.
	loadSources LoadSourcesFunc
	setSource   SetSourceFunc
	sources     *sourcesPanel

	result   SearchHomeResult
	restore  HomeRestore // where the screen was when result was decided
	err      error
	showHelp bool
}

func newSearchHomeModel(opts SearchHomeOptions) *searchHomeModel {
	opts = opts.withRestore()
	theme := CurrentTheme()
	shell := NewShell(&theme, "Search")
	shell.Context = singleLine(opts.SourcesLine)

	input := textinput.New()
	input.Prompt = drawGlyph("❯ ", "> ")
	input.Placeholder = "title, or " + drawGlyph("↓", "down") + " to pick from the lists below"
	styles := textinput.DefaultStyles(themeIsDark())
	styles.Focused.Text = theme.Text
	styles.Focused.Placeholder = theme.Muted
	styles.Focused.Prompt = theme.Primary
	styles.Blurred = styles.Focused
	styles.Focused.Suggestion = theme.Muted
	styles.Blurred.Suggestion = theme.Muted
	styles.Cursor.Color = theme.Primary.GetForeground()
	styles.Cursor.Shape = tea.CursorBlock
	input.SetStyles(styles)
	// The inline ghost text (the rest of the best suggestion, drawn after the
	// cursor) is driven by syncGhost; the widget's own accept/cycle keys would
	// fight the scope tab and the dropdown, so they are off.
	input.ShowSuggestions = true
	input.KeyMap.AcceptSuggestion = bubblekey.NewBinding(bubblekey.WithDisabled())
	input.KeyMap.NextSuggestion = bubblekey.NewBinding(bubblekey.WithDisabled())
	input.KeyMap.PrevSuggestion = bubblekey.NewBinding(bubblekey.WithDisabled())
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
		favoriteAll: slices.Clone(opts.Favorites),
		cursor:      0,
		focusInput:  true,
		deleteEntry: opts.DeleteEntry,
		now:         time.Now,

		deleteRecent:   opts.DeleteRecent,
		deleteFavorite: opts.DeleteFavorite,
		clearContinue:  opts.ClearContinue,
		clearRecent:    opts.ClearRecent,
		clearFavorites: opts.ClearFavorites,
		suggest:        opts.Suggest,
		loadSources:    opts.Sources,
		setSource:      opts.SetSource,
		debounce:       suggestDebounce,
		sugCursor:      -1,
	}
	inputStyles := input.Styles()
	m.lists = newHomeLists(opts.Lists, &inputStyles)
	m.asked = m.suggestKey()
	m.notice = singleLine(opts.Notice)
	m.searchPlaceholder = input.Placeholder
	if m.onFavorites() {
		m.input.Placeholder = favoritesPlaceholder
	}
	m.listsOnScope()
	m.rebuildRows()
	m.listsLanded()
	m.placeRestore(opts.Restore)
	return m
}

// themeIsDark reports the process theme for the input's base styles.
func themeIsDark() bool {
	return IsDark()
}

func (m *searchHomeModel) rebuildRows() {
	m.rows = m.rows[:0]
	if m.onFavorites() {
		// The tab shows every favorite that matches the box, and nothing else.
		m.rows = append(m.rows, m.tabRows()...)
		if m.cursor >= len(m.rows) {
			m.cursor = max(len(m.rows)-1, 0)
		}
		if len(m.rows) == 0 {
			m.focusInput = true
		}
		return
	}
	for i := range m.continueAll {
		m.rows = append(m.rows, homeRow{kind: homeRowContinue, index: i})
	}
	for i := range min(len(m.favoriteAll), homeFavoritesShown) {
		m.rows = append(m.rows, homeRow{kind: homeRowFavorite, index: i})
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

// jumpRows moves the highlight a page or to an end of the list. It reports
// whether it did: it does not when the box has the focus or there are no rows.
func (m *searchHomeModel) jumpRows(key string) bool {
	if m.focusInput || len(m.rows) == 0 {
		return false
	}
	switch key {
	case "pgup":
		m.cursor = max(m.cursor-homePageRows, 0)
	case "pgdown":
		m.cursor = min(m.cursor+homePageRows, len(m.rows)-1)
	case "home":
		m.cursor = 0
	default:
		m.cursor = len(m.rows) - 1
	}
	return true
}

// onFavorites reports whether the FAVORITES tab is the one showing.
func (m *searchHomeModel) onFavorites() bool {
	return m.scopes[m.scopeIdx].Value == favoritesScope
}

// matchingFavorites is the positions in favoriteAll of the favorites the box
// matches: every one when it is empty, else those whose title, year, type or
// source contain what was typed, ignoring case.
func (m *searchHomeModel) matchingFavorites() []int {
	terms := strings.Fields(foldForMatch(m.input.Value()))
	works := m.tabFavorites()
	out := make([]int, 0, len(works))
	for i, f := range works {
		hay := foldForMatch(strings.Join([]string{
			singleLine(f.Title), singleLine(f.Year), singleLine(f.Media), displaySource(f.Source),
		}, " "))
		if homeListsHasAllTerms(hay, terms) {
			out = append(out, i)
		}
	}
	return out
}

// searchableTitle is a favorite's title as a search query: its words, with
// the punctuation sources put in titles ("Pokémon: Aventuras") turned into
// spaces, which the searches turn into one clean query.
func searchableTitle(title string) string {
	words := strings.FieldsFunc(singleLine(title), func(r rune) bool {
		return !unicode.IsLetter(r) && !unicode.IsDigit(r) && !unicode.Is(unicode.Mn, r)
	})
	if len(words) == 0 {
		return singleLine(title)
	}
	return strings.Join(words, " ")
}

// foldForMatch lowercases s and drops its accents, so "pokemon" finds
// "Pokémon" on a keyboard with no accent keys. Every typed word must match.
func foldForMatch(s string) string {
	decomposed := norm.NFD.String(strings.ToLower(s))
	return strings.Map(func(r rune) rune {
		if unicode.Is(unicode.Mn, r) {
			return -1
		}
		return r
	}, decomposed)
}

// scopeChanged puts the screen right after the tab moved: the box takes the
// focus, the rows are the new tab's, and the box says what typing does there.
func (m *searchHomeModel) scopeChanged() tea.Cmd {
	m.cursor = 0
	m.focusInput = true
	m.input.Focus()
	if m.onFavorites() {
		m.input.Placeholder = favoritesPlaceholder
	} else {
		m.input.Placeholder = m.searchPlaceholder
	}
	m.listsOnScope()
	m.rebuildRows()
	cmd := m.queryChanged()
	m.listsLanded()
	return cmd
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

// Update handles resize, typing, scope tabs, navigation and actions, then
// refreshes the inline ghost text to match the resulting state.
func (m *searchHomeModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	model, cmd := m.update(msg)
	m.syncGhost()
	return model, cmd
}

func (m *searchHomeModel) update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.shell.Resize(msg.Width, msg.Height)
		m.sizeInput()
		return m, nil

	case suggestDueMsg:
		cmd := m.startLookup(msg.seq)
		return m, cmd

	case suggestResultMsg:
		m.applySuggestions(msg)
		return m, nil

	case tea.KeyPressMsg:
		m.notice = ""
		if m.sources != nil {
			return m.updateSources(msg)
		}
		if m.clearing != clearNothing {
			return m.updateClearQuestion(msg)
		}
		if m.showHelp {
			return m.updateHelp(msg)
		}
		if model, cmd, ok := m.listsKey(msg); ok {
			return model, cmd
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
		case "ctrl+s":
			m.openSources()
			return m, nil
		case "tab":
			m.scopeIdx = (m.scopeIdx + 1) % len(m.scopes)
			cmd := m.scopeChanged()
			return m, cmd
		case "esc":
			// esc closes the suggestions first, then clears the box; it
			// never quits (tela 10).
			if m.suggestionsOpen() {
				m.dropSuggestions()
				return m, nil
			}
			if m.input.Value() != "" {
				m.input.SetValue("")
				m.focusInput = true
				m.input.Focus()
			}
			cmd := m.queryChanged()
			return m, cmd
		case "enter":
			if res, ok := m.confirm(); ok {
				m.dropSuggestions()
				m.result = res
				m.markRestore()
				return m, tea.Quit
			}
			return m, nil
		case "pgup", "pgdown", "home", "end":
			// On a list these jump a page or to either end; in the box they
			// keep moving the caret.
			if m.jumpRows(msg.String()) {
				return m, nil
			}
		case "up":
			if m.suggestionsOpen() {
				m.moveSuggestion(-1)
				return m, nil
			}
			m.move(-1)
			return m, nil
		case "down":
			if m.suggestionsOpen() && m.moveSuggestion(1) {
				return m, nil
			}
			m.move(1)
			return m, nil
		case "right":
			// right at the end of the text completes it with the highlighted
			// (or first) suggestion.
			if _, ok := m.completionPick(); ok && m.input.Position() >= len([]rune(m.input.Value())) {
				m.completeSuggestion()
				return m, nil
			}
		case "r", "e", "x", "X":
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
			return m, tea.Batch(cmd, m.queryChanged())
		}
		// Typing over the lists jumps back to the box with the character.
		if isHomeTypeKey(msg) {
			m.focusInput = true
			m.input.Focus()
			var cmd tea.Cmd
			m.input, cmd = m.input.Update(msg)
			return m, tea.Batch(cmd, m.queryChanged())
		}
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// updateHelp answers a key while the help overlay is open: ? or esc closes
// it, ctrl+c quits and every other key does nothing.
func (m *searchHomeModel) updateHelp(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "?", "esc":
		m.showHelp = false
	case "ctrl+c":
		m.err = ErrSearchHomeCancelled
		return m, tea.Quit
	}
	return m, nil
}

// openSources opens the sources screen when the host provided one.
func (m *searchHomeModel) openSources() {
	if m.loadSources == nil {
		return
	}
	m.dropSuggestions()
	m.sources = newSourcesPanel(&m.theme, m.loadSources, m.setSource)
}

// updateSources feeds a key to the open sources screen, closing it on done and
// keeping the header's "N sources on" in step with what was just switched.
func (m *searchHomeModel) updateSources(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	if msg.String() == "ctrl+c" {
		m.err = ErrSearchHomeCancelled
		return m, tea.Quit
	}
	closed := m.sources.update(msg)
	m.shell.Context = SourcesSummary(m.sources.rows)
	if closed {
		m.sources = nil
	}
	return m, nil
}

// suggestKey identifies what the autocomplete would ask: the trimmed query
// and the scope it runs in.
func (m *searchHomeModel) suggestKey() string {
	return m.scopes[m.scopeIdx].Value + "\x00" + strings.TrimSpace(m.input.Value())
}

// suggestionsOpen reports whether the dropdown shows under the box.
func (m *searchHomeModel) suggestionsOpen() bool {
	return m.focusInput && len(m.suggestions) > 0
}

// dropSuggestions closes the dropdown and cancels any lookup in flight. It
// bumps seq so an answer already on its way is dropped on arrival.
func (m *searchHomeModel) dropSuggestions() {
	m.seq++
	if m.cancel != nil {
		m.cancel()
		m.cancel = nil
	}
	m.suggestions = nil
	m.sugCursor = -1
}

// queryChanged reacts to typing, a scope change or a clear: when what the
// autocomplete would ask has changed, the old suggestions go and, for a long
// enough query, a debounce tick is scheduled.
func (m *searchHomeModel) queryChanged() tea.Cmd {
	if m.onFavorites() {
		// Nothing is looked up on the FAVORITES tab: typing filters the list.
		// The highlight only goes back to the top when the filter changed.
		key := m.suggestKey()
		m.dropSuggestions()
		if key != m.asked {
			m.asked = key
			m.cursor = 0
		}
		m.rebuildRows()
		return nil
	}
	key := m.suggestKey()
	if key == m.asked {
		return nil
	}
	m.asked = key
	m.dropSuggestions()
	query := strings.TrimSpace(m.input.Value())
	if m.suggest == nil || len([]rune(query)) < minHomeQuery {
		return nil
	}
	seq := m.seq
	return tea.Tick(m.debounce, func(time.Time) tea.Msg { return suggestDueMsg{seq: seq} })
}

// startLookup runs the lookup for the burst seq, off the UI loop. A stale
// tick (the user typed on) does nothing.
func (m *searchHomeModel) startLookup(seq int) tea.Cmd {
	if seq != m.seq || m.suggest == nil {
		return nil
	}
	ctx, cancel := context.WithTimeout(context.Background(), suggestTimeout)
	m.cancel = cancel
	query := strings.TrimSpace(m.input.Value())
	scope := m.scopes[m.scopeIdx].Value
	fn := m.suggest
	return func() tea.Msg {
		defer cancel()
		items, err := fn(ctx, query, scope)
		return suggestResultMsg{seq: seq, items: items, err: err}
	}
}

// applySuggestions shows a lookup's answer unless the user moved on. A failed
// lookup shows nothing: the box works without autocomplete.
func (m *searchHomeModel) applySuggestions(msg suggestResultMsg) {
	if msg.seq != m.seq {
		return
	}
	m.cancel = nil
	m.suggestions, m.sugCursor = nil, -1
	if msg.err != nil {
		return
	}
	query := strings.TrimSpace(m.input.Value())
	for _, item := range msg.items {
		title := singleLine(item.Title)
		// A row repeating exactly what is typed adds nothing.
		if title == "" || strings.EqualFold(title, query) {
			continue
		}
		m.suggestions = append(m.suggestions, HomeSuggestion{Title: title, Detail: singleLine(item.Detail)})
		if len(m.suggestions) == maxHomeSuggestions {
			break
		}
	}
}

// moveSuggestion steps the dropdown highlight. Up from the first row returns
// to the box. It reports false when down ran off the last row, so the caller
// carries on into the lists below.
func (m *searchHomeModel) moveSuggestion(delta int) bool {
	next := m.sugCursor + delta
	switch {
	case next < -1:
		return true
	case next >= len(m.suggestions):
		m.sugCursor = -1
		return false
	}
	m.sugCursor = next
	return true
}

// completionPick is the suggestion → would complete the box with: the
// highlighted row, else the first row that continues what is typed (the one
// the inline ghost text shows). ok is false when there is none.
func (m *searchHomeModel) completionPick() (int, bool) {
	if !m.suggestionsOpen() {
		return 0, false
	}
	if m.sugCursor >= 0 {
		return m.sugCursor, true
	}
	typed := strings.ToLower(m.input.Value())
	for i, s := range m.suggestions {
		if strings.HasPrefix(strings.ToLower(s.Title), typed) {
			return i, true
		}
	}
	return 0, false
}

// syncGhost shows the completion pick's remainder after the cursor, like an
// editor's inline suggestion, and clears it whenever no completion applies.
func (m *searchHomeModel) syncGhost() {
	if pick, ok := m.completionPick(); ok {
		m.input.SetSuggestions([]string{m.suggestions[pick].Title})
		return
	}
	m.input.SetSuggestions(nil)
}

// completeSuggestion fills the box with the completion pick without
// searching, and closes the dropdown.
func (m *searchHomeModel) completeSuggestion() {
	pick, _ := m.completionPick()
	m.input.SetValue(m.suggestions[pick].Title)
	m.input.CursorEnd()
	m.asked = m.suggestKey()
	m.dropSuggestions()
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
	if m.onFavorites() {
		// enter opens the highlighted favorite, or the first match from the box.
		pos := m.cursor
		if m.focusInput {
			pos = 0
		}
		if pos < 0 || pos >= len(m.rows) {
			return SearchHomeResult{}, false
		}
		return SearchHomeResult{Action: HomeSearch, Query: searchableTitle(m.tabFavorites()[m.rows[pos].index].Title)}, true
	}
	if m.focusInput {
		query := strings.TrimSpace(m.input.Value())
		if m.sugCursor >= 0 {
			query = m.suggestions[m.sugCursor].Title
		}
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
	case homeRowFavorite:
		// A favorite can be an anime or a movie: search every scope for it.
		return SearchHomeResult{Action: HomeSearch, Query: searchableTitle(m.favoriteAll[row.index].Title)}, true
	default:
		entry := m.continueAll[row.index]
		return SearchHomeResult{Action: HomeResume, Entry: entry, HasEntry: true, Scope: scope}, true
	}
}

// rowAction runs r (from start), e (episodes), x (remove) and X (clear the
// list) on the highlighted row. It returns a quit command when the action
// leaves the screen. r and e belong to continue rows; x and X to both lists.
func (m *searchHomeModel) rowAction(key string) tea.Cmd {
	row, ok := m.currentRow()
	if !ok {
		return nil
	}
	if key == "X" {
		m.askToClear(row.kind)
		return nil
	}
	if row.kind == homeRowRecent {
		if key == "x" {
			m.forgetRecent(row.index)
		}
		return nil
	}
	if row.kind == homeRowFavorite {
		if key == "x" {
			m.forgetFavorite(row.index)
		}
		return nil
	}
	entry := m.continueAll[row.index]
	switch key {
	case "r":
		m.result = SearchHomeResult{Action: HomeFromStart, Entry: entry, HasEntry: true, Scope: m.scopes[m.scopeIdx].Value}
		m.markRestore()
		return tea.Quit
	case "e":
		m.result = SearchHomeResult{Action: HomeEpisodes, Entry: entry, HasEntry: true, Scope: m.scopes[m.scopeIdx].Value}
		m.markRestore()
		return tea.Quit
	case "x":
		m.forgetContinue(row.index)
	}
	return nil
}

// forgetContinue removes the CONTINUE row at index from the history and the
// screen. A removal that fails keeps the row and says so.
func (m *searchHomeModel) forgetContinue(index int) {
	if m.deleteEntry == nil {
		return
	}
	if err := m.deleteEntry(m.continueAll[index].Key); err != nil {
		m.failed("remove the episode", err)
		return
	}
	m.continueAll = append(m.continueAll[:index], m.continueAll[index+1:]...)
	m.rowsChanged()
}

// forgetRecent removes the RECENT row at index from the history and the
// screen. A removal that fails keeps the row and says so.
func (m *searchHomeModel) forgetRecent(index int) {
	if m.deleteRecent == nil {
		return
	}
	if err := m.deleteRecent(m.recentAll[index].Query); err != nil {
		m.failed("remove the search", err)
		return
	}
	m.recentAll = append(m.recentAll[:index], m.recentAll[index+1:]...)
	m.rowsChanged()
}

// forgetFavorite takes the FAVORITES row at index off the list and the
// screen. A removal that fails keeps the row and says so.
func (m *searchHomeModel) forgetFavorite(index int) {
	if m.deleteFavorite == nil {
		return
	}
	if err := m.deleteFavorite(m.favoriteAll[index].Key); err != nil {
		m.failed("remove the favorite", err)
		return
	}
	m.favoriteAll = append(m.favoriteAll[:index], m.favoriteAll[index+1:]...)
	m.rowsChanged()
}

// askToClear opens the "clear all?" question for the list kind belongs to. It
// opens nothing when the host cannot clear that list or the list is empty.
func (m *searchHomeModel) askToClear(kind homeRowKind) {
	switch {
	case kind == homeRowContinue && m.clearContinue != nil && len(m.continueAll) > 0:
		m.clearing = clearContinueRows
	case kind == homeRowRecent && m.clearRecent != nil && len(m.recentAll) > 0:
		m.clearing = clearRecentRows
	case kind == homeRowFavorite && m.clearFavorites != nil && len(m.favoriteAll) > 0:
		m.clearing = clearFavoriteRows
	}
}

// updateClearQuestion answers the open "clear all?" question: y clears the
// list, n or esc keeps it, ctrl+c quits, and nothing else gets through (a
// stray key must not type into the box or act on a row behind the question).
func (m *searchHomeModel) updateClearQuestion(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "y", "Y":
		m.clearConfirmed()
	case "n", "N", "esc":
		m.clearing = clearNothing
	case "ctrl+c":
		m.err = ErrSearchHomeCancelled
		return m, tea.Quit
	}
	return m, nil
}

// clearConfirmed empties the list the question was about. A clear that fails
// keeps the list and says so.
func (m *searchHomeModel) clearConfirmed() {
	target := m.clearing
	m.clearing = clearNothing
	switch target {
	case clearContinueRows:
		if err := m.clearContinue(); err != nil {
			m.failed("clear CONTINUE", err)
			return
		}
		m.continueAll = nil
	case clearRecentRows:
		if err := m.clearRecent(); err != nil {
			m.failed("clear RECENT", err)
			return
		}
		m.recentAll = nil
	case clearFavoriteRows:
		if err := m.clearFavorites(); err != nil {
			m.failed("clear FAVORITES", err)
			return
		}
		m.favoriteAll = nil
	case clearListWorks, clearDeleteList:
		m.listsConfirmed(target)
		return
	}
	m.rowsChanged()
}

// rowsChanged redraws the rows after a list shrank. With nothing left to
// pick, the box takes the focus back.
func (m *searchHomeModel) rowsChanged() {
	m.rebuildRows()
	if len(m.rows) == 0 {
		m.focusInput = true
		m.input.Focus()
	}
}

// failed puts what went wrong on the screen until the next key.
func (m *searchHomeModel) failed(what string, err error) {
	m.notice = fmt.Sprintf("Could not %s: %v", what, err)
}

// questionView is the line under the scope tabs while the screen asks
// whether to clear a list, or says that a removal failed. "" otherwise.
func (m *searchHomeModel) questionView(width int) string {
	switch m.clearing {
	case clearContinueRows:
		return renderNotice(&m.theme, fmt.Sprintf("Clear all %s from CONTINUE? y to clear, n to keep.", countOf(len(m.continueAll), "episode", "episodes")), width)
	case clearRecentRows:
		return renderNotice(&m.theme, fmt.Sprintf("Clear all %s from RECENT? y to clear, n to keep.", countOf(len(m.recentAll), "search", "searches")), width)
	case clearFavoriteRows:
		return renderNotice(&m.theme, fmt.Sprintf("Clear all %s from FAVORITES? y to clear, n to keep.", countOf(len(m.favoriteAll), "favorite", "favorites")), width)
	case clearListWorks, clearDeleteList:
		return renderNotice(&m.theme, m.listsQuestion(), width)
	}
	return renderNotice(&m.theme, singleLine(m.notice), width)
}

// countOf is "1 episode", "3 episodes".
func countOf(n int, one, many string) string {
	if n == 1 {
		return "1 " + one
	}
	return fmt.Sprintf("%d %s", n, many)
}

func (m *searchHomeModel) sizeInput() {
	width, _ := m.shell.ContentSize()
	m.input.SetWidth(max(width-lipgloss.Width(m.input.Prompt)-2, 8))
	m.sizeListsEntry()
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
	if m.sources != nil {
		return m.sources.hints()
	}
	if m.showHelp {
		return []Hint{
			{Key: "?", Desc: "close", Essential: true},
			{Key: "esc", Desc: "close", Essential: true},
		}
	}
	if modal := m.listsModalHints(); modal != nil {
		return modal
	}
	if m.clearing != clearNothing {
		return []Hint{
			{Key: "y", Desc: "clear all", Essential: true},
			{Key: "n", Desc: "keep", Essential: true},
			{Key: "ctrl+c", Desc: "quit"},
		}
	}
	move := "↑↓"
	if Plain() {
		move = "up/dn"
	}
	enter := "search"
	if m.onFavorites() {
		enter = "open"
	}
	hints := []Hint{
		{Key: "enter", Desc: enter, Essential: true},
		{Key: move, Desc: "move"},
		{Key: "tab", Desc: "scope"},
		{Key: "?", Desc: "help"},
		{Key: "ctrl+c", Desc: "quit", Essential: true},
	}
	if !m.focusInput {
		hints = append(hints[:3], append([]Hint{{Key: "x", Desc: "remove", Essential: true}, {Key: "X", Desc: "clear list"}}, hints[3:]...)...)
	}
	if m.loadSources != nil {
		hints = append(hints[:3], append([]Hint{{Key: "ctrl+s", Desc: "sources", Essential: true}}, hints[3:]...)...)
	}
	if _, ok := m.completionPick(); ok {
		right := "→"
		if Plain() {
			right = "right"
		}
		hints = append(hints[:2], append([]Hint{{Key: right, Desc: "complete"}}, hints[2:]...)...)
	}
	return m.listsFooter(hints)
}

// View renders the search box, the scope tabs, the continue and recent
// lists, and the details panel of the highlighted continue row.
func (m *searchHomeModel) View() tea.View {
	width, height := m.shell.ContentSize()
	var body string
	switch {
	case m.sources != nil:
		body = m.sources.view(width)
	case m.showHelp:
		body = m.helpView(width, height)
	default:
		body = m.homeView(width, height)
	}
	view := tea.NewView(m.shell.Render(body, m.hints()))
	view.AltScreen = true
	view.WindowTitle = "GoAnime"
	return view
}

func (m *searchHomeModel) helpView(width, height int) string {
	rows := []HelpRow{
		{Key: "enter", Desc: "search, resume"},
		{Key: "tab", Desc: "change scope; FAVORITES lists your favorites"},
		{Key: "ctrl+s", Desc: "choose sources"},
		{Key: "r", Desc: "from the start"},
		{Key: "e", Desc: "episodes"},
		{Key: "x", Desc: "remove"},
		{Key: "X", Desc: "clear the list"},
	}
	if help := RenderHelp(&m.theme, "Search", append(rows, m.listsHelpRows()...), nil, width, height); help != "" {
		return help
	}
	return ""
}

// homeView lays out the box, tabs and lists, with the panel beside the
// continue list on a wide terminal (as text below it in plain mode).
func (m *searchHomeModel) homeView(width, height int) string {
	sections := []string{m.searchBox(width)}
	if drop := m.suggestionsView(width); drop != "" {
		sections = append(sections, drop)
	}
	sections = append(sections, m.scopeTabs())
	if question := m.questionView(width); question != "" {
		sections = append(sections, question)
	}
	if m.onFavorites() {
		sections = append(sections, "", m.favoritesTabView(width, height))
		return strings.Join(sections, "\n")
	}
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
	sections = append(sections, "", m.sectionHeader("FAVORITES", ""), m.favoritesView(width))
	if recent := m.recentView(width); recent != "" {
		sections = append(sections, "", m.sectionHeader("RECENT", "最近"), recent)
	}
	return strings.Join(sections, "\n")
}

// suggestionsView draws the autocomplete rows under the box: the title and,
// at the right, its kind and year.
func (m *searchHomeModel) suggestionsView(width int) string {
	if !m.suggestionsOpen() {
		return ""
	}
	items := make([]homeRowItem, 0, len(m.suggestions))
	for _, s := range m.suggestions {
		items = append(items, homeRowItem{title: s.Title, note: s.Detail})
	}
	return renderHomeRows(&m.theme, items, m.sugCursor, width)
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
			if Plain() {
				// No colour to tell the tab by: brackets do it in text.
				name = "[" + name + "]"
			}
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
	if gloss != "" && ShowJapanese() {
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
		cursor = drawGlyph(" ▶ ", "> ")
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

// favoritesTabView is the FAVORITES tab: every favorite the box matches, in a
// window that follows the cursor, with how many are above and below it. With
// Options.Lists it is the works of one list, or the lists themselves.
func (m *searchHomeModel) favoritesTabView(width, height int) string {
	if m.listsLevelOne() {
		return m.listsTabView(width, height)
	}
	works := m.tabFavorites()
	total := len(works)
	head := m.sectionHeader(m.tabLabel(width), "")
	switch {
	case total == 0:
		return head + "\n" + m.theme.Muted.Render("  "+m.tabEmptyText())
	case len(m.rows) == 0:
		return head + "\n" + m.theme.Muted.Render("  "+m.tabNoMatchText())
	}
	head += m.tabCount(len(m.rows), total)

	capacity, arrows := m.tabCapacity(width, height, 0)
	win := windowAround(len(m.rows), m.cursor, capacity, arrows)
	items := make([]homeRowItem, 0, win.end-win.start)
	for _, row := range m.rows[win.start:win.end] {
		f := works[row.index]
		var note []string
		if y := singleLine(f.Year); y != "" {
			note = append(note, y)
		}
		if media := singleLine(f.Media); media != "" {
			note = append(note, media)
		}
		items = append(items, homeRowItem{title: singleLine(f.Title), note: strings.Join(note, " · "), key: displaySource(f.Source)})
	}
	return m.windowLines(head, items, win, width)
}

// rowWindow is the slice of m.rows a tab draws and whether it says how many
// rows are above and below it.
type rowWindow struct {
	start, end int
	arrows     bool
}

// tabAbove is how many lines sit above the tab's heading: the box, the scope
// tabs, the blank line, the question and anything extra.
func (m *searchHomeModel) tabAbove(width, extra int) int {
	above := lipgloss.Height(m.searchBox(width)) + 1 + 1 + extra
	if question := m.questionView(width); question != "" {
		above += lipgloss.Height(question)
	}
	return above
}

// tabCapacity is how many rows fit under the tab's heading, leaving room for
// the "more above/below" lines. On a tiny terminal only the highlighted row
// shows, so it is never cut off.
func (m *searchHomeModel) tabCapacity(width, height, extra int) (capacity int, arrows bool) {
	capacity = height - m.tabAbove(width, extra) - 1 - 2
	if capacity < 1 {
		return 1, false
	}
	return capacity, true
}

// windowAround is the window of capacity rows (of total) that keeps cursor in
// view, as near the middle as the ends allow.
func windowAround(total, cursor, capacity int, arrows bool) rowWindow {
	start := 0
	if total > capacity {
		start = min(max(cursor-capacity/2, 0), total-capacity)
	}
	return rowWindow{start: start, end: min(start+capacity, total), arrows: arrows}
}

// windowLines draws a heading, the rows of win and the lines that say how many
// rows are above and below it.
func (m *searchHomeModel) windowLines(head string, items []homeRowItem, win rowWindow, width int) string {
	cursor := -1
	if !m.focusInput {
		cursor = m.cursor - win.start
	}
	lines := []string{head}
	if win.start > 0 && win.arrows {
		lines = append(lines, m.theme.Muted.Render(fmt.Sprintf("  ↑ %d more above", win.start)))
	}
	lines = append(lines, renderHomeRows(&m.theme, items, cursor, width))
	if win.end < len(m.rows) && win.arrows {
		lines = append(lines, m.theme.Muted.Render(fmt.Sprintf("  ↓ %d more below", len(m.rows)-win.end)))
	}
	return strings.Join(lines, "\n")
}

// favoritesView renders the FAVORITES rows: the title and, at the right, its
// year and type.
func (m *searchHomeModel) favoritesView(width int) string {
	if len(m.favoriteAll) == 0 {
		// Always on the menu, so it can be found before the first favorite.
		return m.theme.Muted.Render("  Nothing here yet · press ctrl+f on a result to add it")
	}
	shown := m.favoriteAll[:min(len(m.favoriteAll), homeFavoritesShown)]
	items := make([]homeRowItem, 0, len(shown))
	for _, f := range shown {
		var note []string
		if y := singleLine(f.Year); y != "" {
			note = append(note, y)
		}
		if media := singleLine(f.Media); media != "" {
			note = append(note, media)
		}
		items = append(items, homeRowItem{title: singleLine(f.Title), note: strings.Join(note, " · "), key: displaySource(f.Source)})
	}
	rows := renderHomeRows(&m.theme, items, m.sectionCursor(homeRowFavorite), width)
	if more := len(m.favoriteAll) - len(shown); more > 0 {
		rows += "\n" + m.theme.Muted.Render(fmt.Sprintf("  +%d more · press tab until FAVORITES to see them all", more))
	}
	return rows
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
	return model.resultWithRestore(), nil
}
