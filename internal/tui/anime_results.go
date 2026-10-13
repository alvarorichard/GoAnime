package tui

import (
	"errors"
	"fmt"
	"strconv"
	"strings"
	"sync"
	"unicode"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/charmbracelet/x/ansi"
)

const (
	wideResultsBreakpoint = 100
	// resultsGutter separates the result list from the details panel.
	resultsGutter = 2
	// maxNoticeLines is how many lines a notice wraps to before it is cut.
	maxNoticeLines = 2
)

// resultsListWidth is the column the result list takes on a wide screen; the
// details panel gets the rest.
func resultsListWidth(width int) int {
	return width * 3 / 5
}

var (
	// ErrSelectionBack means the user requested the previous screen.
	ErrSelectionBack = errors.New("back from anime selection")
	// ErrSelectionCancelled means the user quit the result screen.
	ErrSelectionCancelled = errors.New("anime selection cancelled")
	// ErrNoAnimeResults prevents opening an empty result screen.
	ErrNoAnimeResults = errors.New("no anime results to select")
)

// resultsNotice holds a message for the next result screen to show above the
// list.
var resultsNotice struct {
	sync.Mutex
	text string
}

// NoticeOnNextResults makes the next result screen show text above the list.
//
// It is how a step that sends the user back to the results, such as opening a
// title that has no episodes yet, says why: anything printed to the terminal
// before the screen opens is hidden behind it.
func NoticeOnNextResults(text string) {
	resultsNotice.Lock()
	resultsNotice.text = singleLine(text)
	resultsNotice.Unlock()
}

// takeResultsNotice returns the pending notice and clears it, so it is shown
// once.
func takeResultsNotice() string {
	resultsNotice.Lock()
	defer resultsNotice.Unlock()
	text := resultsNotice.text
	resultsNotice.text = ""
	return text
}

// renderNotice draws a notice after a crimson "!" block, wrapped to at most
// maxNoticeLines lines and followed by a blank line. It is "" for no notice.
func renderNotice(theme *Theme, notice string, width int) string {
	if notice == "" {
		return ""
	}
	block := theme.Notice.Render("!")
	indent := lipgloss.Width(block) + 1
	room := width - indent
	if room < minTitleWidth {
		return theme.Primary.Render(ansi.Truncate(notice, width, "…")) + "\n"
	}
	wrapped := strings.Split(ansi.Wrap(notice, room, ""), "\n")
	if len(wrapped) > maxNoticeLines {
		wrapped = append(wrapped[:maxNoticeLines-1],
			ansi.Truncate(strings.Join(wrapped[maxNoticeLines-1:], " "), room, "…"))
	}
	lines := make([]string, len(wrapped))
	for i, text := range wrapped {
		lead := strings.Repeat(" ", indent)
		if i == 0 {
			lead = block + " "
		}
		lines[i] = lead + theme.Text.Render(text)
	}
	return strings.Join(lines, "\n") + "\n"
}

type animeResultItem struct {
	anime  *models.Anime
	number string // the result's position, zero padded to the widest one
}

// FilterValue returns all useful searchable metadata for an anime result.
func (i animeResultItem) FilterValue() string {
	if i.anime == nil {
		return ""
	}
	fields := make([]string, 0, 6)
	for _, value := range []string{
		i.anime.Name, i.anime.Source, i.anime.Year, string(i.anime.MediaType), i.anime.Audio, i.anime.Quality,
	} {
		if value = singleLine(value); value != "" {
			fields = append(fields, value)
		}
	}
	return strings.Join(fields, " ")
}

// mediaTypeLabel is how a media type reads on screen.
func mediaTypeLabel(mediaType models.MediaType) string {
	switch mediaType {
	case models.MediaTypeAnime:
		return "Anime"
	case models.MediaTypeMovie:
		return "Movie"
	case models.MediaTypeTV:
		return "TV"
	default:
		return singleLine(string(mediaType))
	}
}

// Title returns the primary result label.
func (i animeResultItem) Title() string {
	if i.anime == nil {
		return "Unknown title"
	}
	title := singleLine(i.anime.Name)
	if title == "" {
		return "Unknown title"
	}
	return title
}

// Description returns compact source, year, type, audio, and quality metadata.
func (i animeResultItem) Description() string {
	if i.anime == nil {
		return "Information unavailable"
	}
	parts := make([]string, 0, 5)
	for _, value := range []string{i.anime.Source, i.anime.Year, mediaTypeLabel(i.anime.MediaType), i.anime.Audio, i.anime.Quality} {
		if value = singleLine(value); value != "" {
			parts = append(parts, value)
		}
	}
	if len(parts) == 0 {
		return "Information unavailable"
	}
	return strings.Join(parts, " · ")
}

func (i animeResultItem) row() row {
	return row{number: i.number, title: i.Title(), desc: i.Description()}
}

// SingleLine strips terminal controls and collapses metadata to one line.
// Exported so callers can sanitize titles before building breadcrumb trails.
func SingleLine(value string) string {
	return singleLine(value)
}

// singleLine strips terminal controls and collapses metadata to one line.
func singleLine(value string) string {
	value = ansi.Strip(value)
	value = strings.Map(func(r rune) rune {
		if unicode.IsControl(r) {
			return ' '
		}
		return r
	}, value)
	return strings.Join(strings.Fields(value), " ")
}

type animeResultsModel struct {
	theme         Theme
	shell         Shell
	results       list.Model
	notice        string
	selected      *models.Anime
	err           error
	filterPending bool
}

// newAnimeResultsModel creates the styled, filterable anime result screen.
func newAnimeResultsModel(animes []*models.Anime) *animeResultsModel {
	theme := NewTheme(true)
	shell := NewShell(&theme, "Search > Results")
	valid := make([]*models.Anime, 0, len(animes))
	for _, anime := range animes {
		if anime != nil {
			valid = append(valid, anime)
		}
	}
	digits := max(len(strconv.Itoa(len(valid))), 2)
	items := make([]list.Item, 0, len(valid))
	for i, anime := range valid {
		items = append(items, animeResultItem{anime: anime, number: fmt.Sprintf("%0*d", digits, i+1)})
	}

	delegate := rowDelegate{theme: &theme, twoLine: true, spacing: 1}

	m := &animeResultsModel{theme: theme, shell: shell, notice: takeResultsNotice()}
	width, height := shell.ContentSize()
	results := list.New(items, delegate, width, max(height-m.noticeHeight(width), 1))
	styleList(&results, &m.theme, "")
	results.SetStatusBarItemName("result", "results")
	results.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		}
	}
	m.results = results
	return m
}

// noticeHeight is the rows the notice takes above the list at width.
func (m *animeResultsModel) noticeHeight(width int) int {
	if m.notice == "" {
		return 0
	}
	return lipgloss.Height(renderNotice(&m.theme, m.notice, width))
}

// Init starts the result screen without background work.
func (m *animeResultsModel) Init() tea.Cmd {
	return nil
}

// Update handles resize, filtering, navigation, selection, and cancellation.
func (m *animeResultsModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	if _, ok := msg.(list.FilterMatchesMsg); ok {
		m.filterPending = false
	}
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.shell.Resize(msg.Width, msg.Height)
		width, height := m.shell.ContentSize()
		listWidth := width
		if width >= wideResultsBreakpoint {
			listWidth = resultsListWidth(width) - resultsGutter
		}
		m.results.SetSize(listWidth, max(height-m.noticeHeight(width), 1))
		return m, nil

	case tea.KeyPressMsg:
		filtering := m.results.FilterState() == list.Filtering
		switch msg.String() {
		case "ctrl+c":
			m.err = ErrSelectionCancelled
			return m, tea.Quit
		case "q":
			if !filtering {
				m.err = ErrSelectionCancelled
				return m, tea.Quit
			}
		case "esc":
			if m.results.FilterState() == list.Unfiltered {
				m.err = ErrSelectionBack
				return m, tea.Quit
			}
		case "enter":
			if !filtering {
				if item, ok := m.results.SelectedItem().(animeResultItem); ok && item.anime != nil {
					m.selected = item.anime
					return m, tea.Quit
				}
			} else if m.filterPending {
				return m, nil
			}
		}
	}

	filterBefore := m.results.FilterValue()
	var cmd tea.Cmd
	m.results, cmd = m.results.Update(msg)
	if m.results.FilterState() == list.Filtering && filterBefore != m.results.FilterValue() {
		m.filterPending = true
	}
	return m, cmd
}

// View renders compact results or a wide list-and-details layout.
func (m *animeResultsModel) View() tea.View {
	width, height := m.shell.ContentSize()
	height = max(height-m.noticeHeight(width), 1)
	body := m.results.View()
	if width >= wideResultsBreakpoint {
		listWidth := resultsListWidth(width)
		if panel := fitPanel(&m.theme, m.detailsContent(), width-listWidth, height); panel != "" {
			// A fixed-width column keeps the panel in place whatever the
			// length of the visible rows.
			column := lipgloss.NewStyle().Width(listWidth).Render(body)
			body = lipgloss.JoinHorizontal(lipgloss.Top, column, panel)
		}
	}
	if notice := renderNotice(&m.theme, m.notice, width); notice != "" {
		body = notice + "\n" + body
	}

	view := tea.NewView(m.shell.Render(body, m.hints()))
	view.AltScreen = true
	view.WindowTitle = "GoAnime - Results"
	return view
}

// detailsContent is the details panel for the selected result.
func (m *animeResultsModel) detailsContent() *panelContent {
	name, source, year, mediaType, audio := "No result selected", "—", "—", "—", "—"
	if item, ok := m.results.SelectedItem().(animeResultItem); ok && item.anime != nil {
		name = item.Title()
		if value := singleLine(item.anime.Source); value != "" {
			source = value
		}
		if value := singleLine(item.anime.Year); value != "" {
			year = value
		}
		if value := mediaTypeLabel(item.anime.MediaType); value != "" {
			mediaType = value
		}
		if value := singleLine(item.anime.Audio); value != "" {
			audio = value
		}
	}
	return &panelContent{
		label:  "Details",
		gloss:  "詳細",
		title:  name,
		fields: []detailField{{"Source", source}, {"Year", year}, {"Type", mediaType}, {"Audio", audio}},
	}
}

func (m *animeResultsModel) hints() []Hint {
	if m.results.FilterState() != list.Unfiltered {
		return []Hint{
			{Key: "↑↓", Desc: "move"},
			{Key: "enter", Desc: "open", Essential: true},
			{Key: "esc", Desc: "clear filter", Essential: true},
		}
	}
	return []Hint{
		{Key: "↑↓", Desc: "move"},
		{Key: "/", Desc: "filter"},
		{Key: "enter", Desc: "open", Essential: true},
		{Key: "esc", Desc: "back", Essential: true},
		{Key: "q", Desc: "quit"},
	}
}

type animeResultsRunner func(tea.Model) (tea.Model, error)

// SelectAnime opens the result screen and returns the chosen anime.
func SelectAnime(animes []*models.Anime) (*models.Anime, error) {
	return selectAnimeWithRunner(animes, func(model tea.Model) (tea.Model, error) {
		var final tea.Model
		err := RunClean(func() error {
			var runErr error
			final, runErr = NewProgram(model).Run()
			return runErr
		})
		return final, err
	})
}

// selectAnimeWithRunner isolates terminal execution for deterministic tests.
func selectAnimeWithRunner(animes []*models.Anime, run animeResultsRunner) (*models.Anime, error) {
	valid := make([]*models.Anime, 0, len(animes))
	for _, anime := range animes {
		if anime != nil {
			valid = append(valid, anime)
		}
	}
	if len(valid) == 0 {
		return nil, ErrNoAnimeResults
	}
	if run == nil {
		return nil, fmt.Errorf("anime result runner not configured")
	}

	final, err := run(newAnimeResultsModel(valid))
	if err != nil {
		return nil, fmt.Errorf("run anime result screen: %w", err)
	}
	model, ok := final.(*animeResultsModel)
	if !ok || model == nil {
		return nil, fmt.Errorf("unexpected anime result model %T", final)
	}
	if model.err != nil {
		return nil, model.err
	}
	if model.selected == nil {
		return nil, ErrSelectionCancelled
	}
	return model.selected, nil
}
