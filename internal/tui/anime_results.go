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
	if Plain() {
		block = "[!]"
		indent = len(block) + 1
	}
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

// ---- Grouped results (tela 3) ----

// animeVersion is one source's copy of a work: the row's anime plus the
// language it is in.
type animeVersion struct {
	anime *models.Anime
	// lang is the badge on the row ("PT-BR" or "EN").
	lang string
	// langWord is the word in the Details panel ("PT-BR" or "English").
	langWord string
}

// animeGroup is one work: every source's copy of the same title. The list
// shows one row per group; left/right picks the version (tela 3).
type animeGroup struct {
	title    string // display title, language tags stripped
	versions []animeVersion
	year     string
	media    string // "Anime", "Movie" or "TV"
}

// groupAnimeResults groups search results by work, keeping the incoming
// order: the search already sorts closest match first. Nil entries are
// dropped; a nil-heavy list still yields no groups.
func groupAnimeResults(animes []*models.Anime) []animeGroup {
	var groups []animeGroup
	index := make(map[string]int)
	for _, anime := range animes {
		if anime == nil {
			continue
		}
		titleKey := normalizeTitleKey(anime.Name)
		if pos, ok := index[titleKey]; ok {
			groups[pos].versions = append(groups[pos].versions, makeVersion(anime))
			if groups[pos].year == "" {
				groups[pos].year = singleLine(anime.Year)
			}
			if groups[pos].media == "" {
				groups[pos].media = mediaTypeLabel(anime.MediaType)
			}
			continue
		}
		index[titleKey] = len(groups)
		groups = append(groups, animeGroup{
			title:    cleanResultTitle(anime.Name),
			versions: []animeVersion{makeVersion(anime)},
			year:     singleLine(anime.Year),
			media:    mediaTypeLabel(anime.MediaType),
		})
	}
	return groups
}

func makeVersion(anime *models.Anime) animeVersion {
	badge, word := versionLanguage(anime)
	return animeVersion{anime: anime, lang: badge, langWord: word}
}

// versionLanguage is the language a result is in: the [PT-BR]/[English] tag
// the scrapers put on the name wins, otherwise the source decides (only
// HiAnime is English).
func versionLanguage(anime *models.Anime) (badge, word string) {
	name := strings.ToLower(singleLine(anime.Name))
	switch {
	case strings.Contains(name, "[pt-br]"),
		strings.Contains(name, "[portuguese]"),
		strings.Contains(name, "[português]"):
		return "PT-BR", "PT-BR"
	case strings.Contains(name, "[english]"):
		return "EN", "English"
	}
	if strings.Contains(strings.ToLower(singleLine(anime.Source)), "hianime") {
		return "EN", "English"
	}
	return "PT-BR", "PT-BR"
}

// displaySource is how a source reads on screen.
func displaySource(source string) string {
	if strings.EqualFold(strings.TrimSpace(source), "animefire.io") {
		return "AnimeFire"
	}
	return singleLine(source)
}

// cleanResultTitle strips the bracketed tags scrapers prepend ("[PT-BR]",
// "[English]", "[Movie]", "[TV]") and trailing audio words ("(Legendado)",
// "(Dublado)") so grouped rows show the work's title.
func cleanResultTitle(name string) string {
	title := singleLine(name)
	for {
		trimmed := strings.TrimSpace(title)
		if !strings.HasPrefix(trimmed, "[") {
			title = trimmed
			break
		}
		end := strings.Index(trimmed, "]")
		if end < 0 {
			break
		}
		title = strings.TrimSpace(trimmed[end+1:])
	}
	for _, audio := range []string{"(legendado)", "(dublado)"} {
		if strings.HasSuffix(strings.ToLower(title), audio) {
			title = strings.TrimSpace(title[:len(title)-len(audio)])
		}
	}
	if title == "" {
		return "Unknown title"
	}
	return title
}

// normalizeTitleKey folds a result name into its grouping key: tags off,
// lowercase, letters and digits only, single spaces.
func normalizeTitleKey(name string) string {
	title := strings.ToLower(cleanResultTitle(name))
	var sb strings.Builder
	space := false
	for _, r := range title {
		switch {
		case unicode.IsLetter(r) || unicode.IsDigit(r):
			sb.WriteRune(r)
			space = false
		case unicode.IsSpace(r):
			if !space {
				sb.WriteRune(' ')
				space = true
			}
		}
	}
	return strings.TrimSpace(sb.String())
}

// groupBadges are the row's language badges: PT-BR before EN.
func (g *animeGroup) badges() []string {
	var badges []string
	seen := make(map[string]bool)
	for _, order := range []string{"PT-BR", "EN"} {
		for _, v := range g.versions {
			if v.lang == order && !seen[order] {
				seen[order] = true
				badges = append(badges, order)
			}
		}
	}
	return badges
}

// groupDescription is the row's second line: year, type and either the
// version count or the single version's source.
func (g *animeGroup) description() string {
	parts := make([]string, 0, 3)
	if g.year != "" {
		parts = append(parts, g.year)
	}
	if g.media != "" {
		parts = append(parts, g.media)
	}
	if len(g.versions) == 1 {
		if source := displaySource(g.versions[0].anime.Source); source != "" {
			parts = append(parts, source)
		}
	} else {
		parts = append(parts, fmt.Sprintf("%d versions", len(g.versions)))
	}
	if len(parts) == 0 {
		return "Information unavailable"
	}
	return strings.Join(parts, " · ")
}

type animeGroupItem struct {
	group  *animeGroup
	pos    int    // the group's position in the model
	number string // the row's position, zero padded to the widest one
}

// FilterValue exposes the title and every version's metadata to the fuzzy
// filter, so typing a source ("hianime") finds its group.
func (i animeGroupItem) FilterValue() string {
	if i.group == nil {
		return ""
	}
	fields := []string{i.group.title}
	if i.group.year != "" {
		fields = append(fields, i.group.year)
	}
	if i.group.media != "" {
		fields = append(fields, i.group.media)
	}
	for _, v := range i.group.versions {
		if v.anime == nil {
			continue
		}
		for _, value := range []string{
			displaySource(v.anime.Source), v.lang, v.langWord,
			singleLine(v.anime.Audio), singleLine(v.anime.Quality),
		} {
			if value != "" {
				fields = append(fields, value)
			}
		}
	}
	return strings.Join(fields, " ")
}

// Title returns the group's display title.
func (i animeGroupItem) Title() string {
	if i.group == nil || i.group.title == "" {
		return "Unknown title"
	}
	return i.group.title
}

// Description returns the year, type and version count or source.
func (i animeGroupItem) Description() string {
	if i.group == nil {
		return "Information unavailable"
	}
	return i.group.description()
}

func (i animeGroupItem) row() row {
	var tags []string
	if i.group != nil {
		tags = i.group.badges()
	}
	return row{number: i.number, title: i.Title(), desc: i.Description(), tags: tags}
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
	groups        []animeGroup
	versionSel    map[int]int // group position -> chosen version
	query         string
	notice        string
	selected      *models.Anime
	err           error
	filterPending bool
	// plainPanel is the details as plain text below the list (tela 9).
	plainPanel string
	// showHelp replaces the list with the KEYS overlay (tela 10).
	showHelp bool
}

// newAnimeResultsModel creates the styled, filterable anime result screen.
func newAnimeResultsModel(animes []*models.Anime) *animeResultsModel {
	return newAnimeResultsModelWithQuery(animes, "")
}

// newAnimeResultsModelWithQuery is newAnimeResultsModel with the search query,
// shown quoted at the right of the header (tela 3).
func newAnimeResultsModelWithQuery(animes []*models.Anime, query string) *animeResultsModel {
	theme := CurrentTheme()
	shell := NewShell(&theme, "Search > Results")
	if query = singleLine(query); query != "" {
		shell.Context = `"` + query + `"`
	}
	groups := groupAnimeResults(animes)
	digits := max(len(strconv.Itoa(len(groups))), 2)
	items := make([]list.Item, 0, len(groups))
	for i := range groups {
		items = append(items, animeGroupItem{group: &groups[i], pos: i, number: fmt.Sprintf("%0*d", digits, i+1)})
	}

	delegate := rowDelegate{theme: &theme, twoLine: true, spacing: 1}

	m := &animeResultsModel{
		theme:      theme,
		shell:      shell,
		groups:     groups,
		versionSel: make(map[int]int),
		query:      query,
		notice:     takeResultsNotice(),
	}
	width, height := shell.ContentSize()
	results := list.New(items, delegate, width, max(height-m.noticeHeight(width), 1))
	results.KeyMap = fancyListKeyMap()
	results.DisableQuitKeybindings() // quit/back owned by the screen (ctrl+c / q / esc)
	styleList(&results, &m.theme, "type to filter these results")
	versions := 0
	for _, g := range groups {
		versions += len(g.versions)
	}
	results.SetStatusBarItemName(
		"title · closest match first",
		fmt.Sprintf("titles · %d versions · closest match first", versions),
	)
	results.AdditionalShortHelpKeys = func() []key.Binding {
		return []key.Binding{
			key.NewBinding(key.WithKeys("enter"), key.WithHelp("enter", "open")),
			key.NewBinding(key.WithKeys("esc"), key.WithHelp("esc", "back")),
		}
	}
	m.results = results
	m.reservePlainPanel()
	return m
}

// reservePlainPanel renders the details as plain text below the list and
// shrinks the list to make room, so nothing is drawn cut off (tela 9). The
// panel follows the highlighted work, so callers refresh it after every
// update that can move the highlight.
func (m *animeResultsModel) reservePlainPanel() {
	m.plainPanel = ""
	if !Plain() {
		return
	}
	width, height := m.shell.ContentSize()
	height = max(height-m.noticeHeight(width), 1)
	panel := fitPanel(&m.theme, m.detailsContent(), width, height)
	if panel == "" {
		m.results.SetSize(width, height)
		return
	}
	m.plainPanel = panel
	m.results.SetSize(width, max(height-lipgloss.Height(panel)-1, 1))
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
		if width >= wideResultsBreakpoint && !Plain() {
			listWidth = resultsListWidth(width) - resultsGutter
		}
		m.results.SetSize(listWidth, max(height-m.noticeHeight(width), 1))
		m.reservePlainPanel()
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
		filtering := m.results.FilterState() == list.Filtering
		switch msg.String() {
		case "?":
			if !filtering {
				m.showHelp = true
				return m, nil
			}
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
		case "left", "right":
			// ←→ picks the highlighted work's version (tela 3). While
			// filtering the arrows move the filter's caret instead.
			if !filtering && m.switchVersion(msg.String() == "right") {
				m.reservePlainPanel()
				return m, nil
			}
		case "enter":
			if !filtering {
				if item, ok := m.results.SelectedItem().(animeGroupItem); ok {
					m.selected = m.chosenVersion(item.pos)
					if m.selected != nil {
						return m, tea.Quit
					}
				}
			} else if m.filterPending {
				return m, nil
			}
		default:
			if !filtering && isTypeToFilterKey(msg) {
				// Tela 3: typing filters these results without "/" first.
				return m.startFilterWithKey(msg)
			}
		}
	}

	filterBefore := m.results.FilterValue()
	var cmd tea.Cmd
	m.results, cmd = m.results.Update(msg)
	if m.results.FilterState() == list.Filtering && filterBefore != m.results.FilterValue() {
		m.filterPending = true
	}
	m.reservePlainPanel()
	return m, cmd
}

// switchVersion moves the highlighted group's chosen version one step.
// It reports whether there was another version to move to.
func (m *animeResultsModel) switchVersion(next bool) bool {
	item, ok := m.results.SelectedItem().(animeGroupItem)
	if !ok || item.group == nil || len(item.group.versions) < 2 {
		return false
	}
	sel := m.versionSel[item.pos]
	if next {
		sel = (sel + 1) % len(item.group.versions)
	} else {
		sel = (sel - 1 + len(item.group.versions)) % len(item.group.versions)
	}
	m.versionSel[item.pos] = sel
	return true
}

// chosenVersion returns the highlighted group's chosen version.
func (m *animeResultsModel) chosenVersion(pos int) *models.Anime {
	if pos < 0 || pos >= len(m.groups) {
		return nil
	}
	versions := m.groups[pos].versions
	if len(versions) == 0 {
		return nil
	}
	sel := m.versionSel[pos]
	if sel < 0 || sel >= len(versions) {
		sel = 0
	}
	return versions[sel].anime
}

// startFilterWithKey opens the list filter with the typed character.
func (m *animeResultsModel) startFilterWithKey(msg tea.KeyPressMsg) (tea.Model, tea.Cmd) {
	filterBefore := m.results.FilterValue()
	cmd := feedKeyToFilter(&m.results, msg)
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
	switch {
	case m.showHelp:
		body = m.helpView(width, height, body)
	case Plain():
		// Tela 9: details as text below the list, never a paper panel.
		if m.plainPanel != "" {
			body += "\n\n" + m.plainPanel
		}
	case width >= wideResultsBreakpoint:
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

// helpView renders the KEYS overlay, falling back to the list when the
// terminal is too narrow for it.
func (m *animeResultsModel) helpView(width, height int, body string) string {
	versionKey := "←→"
	if Plain() {
		versionKey = "left/right"
	}
	if help := RenderHelp(&m.theme, "Results", []HelpRow{
		{Key: versionKey, Desc: "choose a version"},
		{Key: "esc", Desc: "edit the search"},
		{Key: "q", Desc: "quit"},
	}, nil, width, height); help != "" {
		return help
	}
	return body
}

// detailsContent is the details panel for the selected work: its year and
// type, then every version with the chosen one marked. Left/right moves the
// mark; enter opens the marked version's source (tela 3).
func (m *animeResultsModel) detailsContent() *panelContent {
	item, ok := m.results.SelectedItem().(animeGroupItem)
	if !ok || item.group == nil || len(item.group.versions) == 0 {
		return &panelContent{label: "Details", gloss: "詳細", title: "No result selected"}
	}
	group := item.group
	sel := m.versionSel[item.pos]
	if sel < 0 || sel >= len(group.versions) {
		sel = 0
	}

	choose := "←→ to choose"
	if Plain() {
		choose = "left/right to choose"
	}
	versionsLine := fmt.Sprintf("VERSIONS  %d", len(group.versions))
	if len(group.versions) > 1 {
		versionsLine += " · " + choose
	}
	mark, gap := drawGlyph("▶ ", "> "), "  "
	lines := []string{versionsLine, ""}
	for i, v := range group.versions {
		label := displaySource(v.anime.Source)
		detail := v.langWord
		if audio := singleLine(v.anime.Audio); audio != "" {
			detail += " · " + audio
		}
		prefix := gap
		if i == sel {
			prefix = mark
		}
		lines = append(lines, prefix+label+"  "+detail)
	}
	lines = append(lines, "", "enter opens it on "+displaySource(group.versions[sel].anime.Source))

	fields := []detailField{}
	if group.year != "" {
		fields = append(fields, detailField{"Year", group.year})
	}
	if group.media != "" {
		fields = append(fields, detailField{"Type", group.media})
	}
	return &panelContent{
		label:  "Details",
		gloss:  "詳細",
		title:  group.title,
		fields: fields,
		lines:  lines,
	}
}

func (m *animeResultsModel) hints() []Hint {
	if m.showHelp {
		return []Hint{
			{Key: "?", Desc: "close", Essential: true},
			{Key: "esc", Desc: "close", Essential: true},
		}
	}
	version := "←→"
	if Plain() {
		version = "<->"
	}
	move := "↑↓"
	if Plain() {
		move = "up/dn"
	}
	if m.results.FilterState() != list.Unfiltered {
		return []Hint{
			{Key: move, Desc: "move"},
			{Key: "enter", Desc: "open", Essential: true},
			{Key: "esc", Desc: "clear filter", Essential: true},
		}
	}
	return []Hint{
		{Key: "enter", Desc: "open", Essential: true},
		{Key: version, Desc: "version"},
		{Key: move, Desc: "move"},
		{Key: "esc", Desc: "edit search", Essential: true},
		{Key: "?", Desc: "help"},
		{Key: "q", Desc: "quit"},
	}
}

type animeResultsRunner func(tea.Model) (tea.Model, error)

// SelectAnime opens the result screen and returns the chosen anime.
func SelectAnime(animes []*models.Anime) (*models.Anime, error) {
	return SelectAnimeWithQuery(animes, "")
}

// SelectAnimeWithQuery is SelectAnime with the search query, shown quoted at
// the right of the header (tela 3).
func SelectAnimeWithQuery(animes []*models.Anime, query string) (*models.Anime, error) {
	if len(groupAnimeResults(animes)) == 0 {
		return nil, ErrNoAnimeResults
	}
	if !HasTerminal() {
		return nil, fmt.Errorf("no terminal: stdin is not a TTY")
	}
	return selectAnimeWithQueryRunner(animes, query, func(model tea.Model) (tea.Model, error) {
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
	return selectAnimeWithQueryRunner(animes, "", run)
}

// selectAnimeWithQueryRunner is selectAnimeWithRunner with the search query.
func selectAnimeWithQueryRunner(animes []*models.Anime, query string, run animeResultsRunner) (*models.Anime, error) {
	if len(groupAnimeResults(animes)) == 0 {
		return nil, ErrNoAnimeResults
	}
	if run == nil {
		return nil, fmt.Errorf("anime result runner not configured")
	}

	final, err := run(newAnimeResultsModelWithQuery(animes, query))
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
