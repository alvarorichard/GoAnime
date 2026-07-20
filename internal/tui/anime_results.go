package tui

import (
	"errors"
	"fmt"
	"strings"
	"unicode"

	"charm.land/bubbles/v2/key"
	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/charmbracelet/x/ansi"
)

const wideResultsBreakpoint = 100

var (
	// ErrSelectionBack means the user requested the previous screen.
	ErrSelectionBack = errors.New("back from anime selection")
	// ErrSelectionCancelled means the user quit the result screen.
	ErrSelectionCancelled = errors.New("anime selection cancelled")
	// ErrNoAnimeResults prevents opening an empty result screen.
	ErrNoAnimeResults = errors.New("no anime results to select")
)

type animeResultItem struct {
	anime *models.Anime
}

// FilterValue returns all useful searchable metadata for an anime result.
func (i animeResultItem) FilterValue() string {
	if i.anime == nil {
		return ""
	}
	return strings.Join([]string{
		singleLine(i.anime.Name),
		singleLine(i.anime.Source),
		singleLine(i.anime.Year),
		singleLine(string(i.anime.MediaType)),
		singleLine(i.anime.Quality),
	}, " ")
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

// Description returns compact source, year, type, and quality metadata.
func (i animeResultItem) Description() string {
	if i.anime == nil {
		return "Information unavailable"
	}
	parts := make([]string, 0, 4)
	for _, value := range []string{i.anime.Source, i.anime.Year, string(i.anime.MediaType), i.anime.Quality} {
		if value = singleLine(value); value != "" {
			parts = append(parts, value)
		}
	}
	if len(parts) == 0 {
		return "Information unavailable"
	}
	return strings.Join(parts, "  •  ")
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
	selected      *models.Anime
	err           error
	filterPending bool
}

// newAnimeResultsModel creates the styled, filterable anime result screen.
func newAnimeResultsModel(animes []*models.Anime) *animeResultsModel {
	theme := NewTheme(true)
	shell := NewShell(&theme, "Search > Results")
	items := make([]list.Item, 0, len(animes))
	for _, anime := range animes {
		if anime == nil {
			continue
		}
		items = append(items, animeResultItem{anime: anime})
	}

	delegate := list.NewDefaultDelegate()
	delegate.SetSpacing(1)
	delegate.Styles.NormalTitle = theme.Text.PaddingLeft(2)
	delegate.Styles.NormalDesc = theme.Muted.PaddingLeft(2)
	delegate.Styles.SelectedTitle = theme.SelectedTitle
	delegate.Styles.SelectedDesc = theme.SelectedDescription
	delegate.Styles.DimmedTitle = theme.Muted.PaddingLeft(2)
	delegate.Styles.DimmedDesc = theme.Muted.Faint(true).PaddingLeft(2)
	delegate.Styles.FilterMatch = theme.FilterMatch

	width, height := shell.ContentSize()
	results := list.New(items, delegate, width, height)
	results.SetShowTitle(false)
	results.SetShowHelp(false)
	results.SetStatusBarItemName("result", "results")
	results.FilterInput.Prompt = "Filter: "
	results.Styles.StatusBar = theme.Muted
	results.Styles.StatusEmpty = theme.Muted
	results.Styles.StatusBarActiveFilter = theme.Primary
	results.Styles.StatusBarFilterCount = theme.Muted
	results.Styles.NoItems = theme.Muted
	results.Styles.PaginationStyle = theme.Muted
	results.Styles.ActivePaginationDot = theme.Primary
	results.Styles.InactivePaginationDot = theme.Muted
	results.Styles.Filter.Focused.Prompt = theme.Primary
	results.Styles.Filter.Focused.Text = theme.Text
	results.Styles.Filter.Focused.Placeholder = theme.Muted
	results.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		}
	}

	return &animeResultsModel{theme: theme, shell: shell, results: results}
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
		if width >= wideResultsBreakpoint {
			width = width * 3 / 5
		}
		m.results.SetSize(width, height)
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
	body := m.results.View()
	if width >= wideResultsBreakpoint {
		listWidth := width * 3 / 5
		panelWidth := width - listWidth
		name, source, year, mediaType, quality := "No result selected", "—", "—", "—", "—"
		if item, ok := m.results.SelectedItem().(animeResultItem); ok && item.anime != nil {
			name = item.Title()
			if value := singleLine(item.anime.Source); value != "" {
				source = value
			}
			if value := singleLine(item.anime.Year); value != "" {
				year = value
			}
			if value := singleLine(string(item.anime.MediaType)); value != "" {
				mediaType = value
			}
			if value := singleLine(item.anime.Quality); value != "" {
				quality = value
			}
		}
		details := renderAnimeDetails(&m.theme, name, source, year, mediaType, quality)
		frameWidth, frameHeight := m.theme.Panel.GetFrameSize()
		panel := m.theme.Panel.
			Width(max(panelWidth-frameWidth, 1)).
			Height(max(height-frameHeight, 1)).
			Render(details)
		body = lipgloss.JoinHorizontal(lipgloss.Top, body, panel)
	}

	footer := m.results.Help.ShortHelpView(m.results.ShortHelp())
	view := tea.NewView(m.shell.Render(body, footer))
	view.AltScreen = true
	view.WindowTitle = "GoAnime - Results"
	return view
}

// renderAnimeDetails composes styled lines without JoinVertical's unstyled padding.
func renderAnimeDetails(theme *Theme, name, source, year, mediaType, quality string) string {
	return strings.Join([]string{
		theme.Primary.Render("Details"),
		"",
		theme.Value.Render(name),
		"",
		theme.Label.Render("Source"), theme.Value.Render(source),
		"",
		theme.Label.Render("Year"), theme.Value.Render(year),
		"",
		theme.Label.Render("Type"), theme.Value.Render(mediaType),
		"",
		theme.Label.Render("Quality"), theme.Value.Render(quality),
	}, "\n")
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
