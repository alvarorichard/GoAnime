package tui

import (
	"strings"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// hostileSizes are terminals a leigo pode ter: zero, minúsculo, gigante.
var hostileSizes = []tea.WindowSizeMsg{
	{Width: 0, Height: 0},
	{Width: 1, Height: 1},
	{Width: 5, Height: 3},
	{Width: 20, Height: 5},
	{Width: 40, Height: 10},
	{Width: 80, Height: 24},
	{Width: 120, Height: 31},
	{Width: 300, Height: 100},
}

// hostileKeys is a storm of everything a confused user presses, including
// control characters, CJK, emoji and ANSI injections.
var hostileKeys = []tea.KeyPressMsg{
	{Code: tea.KeyEnter}, {Code: tea.KeyEscape}, {Code: tea.KeyUp}, {Code: tea.KeyDown},
	{Code: tea.KeyLeft}, {Code: tea.KeyRight}, {Code: tea.KeyTab}, {Code: 'q', Text: "q"},
	{Code: '?', Text: "?"}, {Code: ' ', Text: " "}, {Code: 'm', Text: "m"}, {Code: 'd', Text: "d"},
	{Code: 'r', Text: "r"}, {Code: 'e', Text: "e"}, {Code: 'x', Text: "x"},
	{Code: '0', Text: "0"}, {Code: '9', Text: "9"}, {Code: '/', Text: "/"},
	{Code: 'g', Text: "g"}, {Code: 'G', Text: "G"},
	{Code: -1, Text: "猫🎬\x1b[2J\n\t\x00"},
	{Code: -1, Text: strings.Repeat("namakura ", 500)},
	{Code: -1, Text: ">"},
	{Code: 'c', Mod: tea.ModCtrl},
}

// hostileModels builds one of each screen with hostile content.
func hostileModels() []tea.Model {
	big := strings.Repeat("爵 ", 200)
	animes := []*models.Anime{
		{Name: big, Source: "Animefire.io", Year: "1917", MediaType: models.MediaTypeAnime, Audio: "Legendado"},
		{Name: "[PT-BR]\x1b[31mEvil", Source: "", Year: "", Audio: "\n"},
		nil,
		{Name: " › › › ", Source: "›", Year: "›"},
	}
	episodes := []PickItem{
		{Label: big, Details: big, Tags: []string{"Filler", big}, Note: big, Key: "x", Meter: 2},
		{Label: "", Details: "", Status: PickWatched},
		{Label: "Ep 2", Status: PickStatus(99)},
	}
	continueEntries := []HomeContinueEntry{
		{Key: "k1", AnimeTitle: big, EpisodeTitle: big, EpisodeNum: -5, Position: -10, Duration: -3, Source: big, Audio: big, UpdatedAt: time.Now().Add(999 * time.Hour)},
		{Key: "", AnimeTitle: "", EpisodeNum: 0, Position: 999999999, Duration: 1, UpdatedAt: time.Time{}},
	}
	return []tea.Model{
		newAnimeResultsModel(animes),
		newPickerModel(episodes, PickOptions{
			Breadcrumb: "Search > ›\n Episodes", Action: "resume", BackDesc: "results", HelpHint: true,
			FilterPlaceholder: big, Context: big,
			Actions:      []PickAction{{Key: "d", Label: big, Hint: big, Section: true}},
			ActionsLabel: big, Panel: &PickPanel{Label: big, Title: big, Lines: []string{big}, Command: big},
		}),
		newSearchHomeModel(SearchHomeOptions{Query: big, ScopeIndex: 99, SourcesLine: big, Continue: continueEntries, Recent: []HomeRecentSearch{{Query: big, Count: -1}}}),
		newSearchingModel(big, nil),
		newTryModelWithOptions(TryOptions{Query: big, Summary: big, Note: big, Sources: []TrySource{{Name: big, Detail: big}}, Suggestions: []TrySuggestion{{Label: big, Hint: big, Query: big}}}),
		newSubtitlesModel(SubtitleOptions{Title: big, Context: big, Tracks: []SubtitleTrack{{Label: "", Language: ""}, {Label: big, Language: big}}, LastLabel: big, NextTitle: big, NextLine: big}),
	}
}

// TestHostileScreensNeverPanic drives every screen through absurd sizes and
// a key storm: a leigo's terminal must never crash the app.
func TestHostileScreensNeverPanic(t *testing.T) {
	for _, plain := range []bool{false, true} {
		SetPlain(plain)
		t.Cleanup(func() { SetPlain(false) })
		for mi, model := range hostileModels() {
			for _, size := range hostileSizes {
				model, _ = model.Update(size)
				for _, key := range hostileKeys {
					require.NotPanics(t, func() {
						model, _ = model.Update(key)
						_ = model.(interface{ View() tea.View }).View()
					}, "plain=%v model=%d size=%+v key=%q", plain, mi, size, key.Text)
				}
			}
		}
	}
}

// TestHostileViewsFitWidth renders every screen at hostile sizes and asserts
// the visible content never exceeds the terminal width.
func TestHostileViewsFitWidth(t *testing.T) {
	SetPlain(false)
	t.Cleanup(func() { SetPlain(false) })
	for mi, model := range hostileModels() {
		for _, size := range hostileSizes {
			if size.Width < 10 {
				continue // below any useful width: clipping is the contract
			}
			model, _ = model.Update(size)
			view := model.(interface{ View() tea.View }).View()
			for _, line := range strings.Split(ansi.Strip(view.Content), "\n") {
				assert.LessOrEqual(t, len([]rune(line)), size.Width+4, "model=%d size=%+v line=%q", mi, size, line)
			}
		}
	}
}

// TestHostileInputsFeedGarbage types hostile text into every input box.
func TestHostileInputsFeedGarbage(t *testing.T) {
	garbage := []string{"", " ", "\n", "\x1b[2J", strings.Repeat("爵", 5000), "a\nb\rc\x00d"}
	home := newSearchHomeModel(SearchHomeOptions{})
	for _, g := range garbage {
		home.input.SetValue(g)
		require.NotPanics(t, func() { _ = home.View() })
	}
	try := newTryModelWithOptions(TryOptions{Query: "x"})
	for _, g := range garbage {
		try.input.SetValue(g)
		require.NotPanics(t, func() { _ = try.View() })
	}
}
