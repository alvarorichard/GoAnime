package tui

import (
	"errors"
	"strings"
	"testing"
	"unicode"

	"charm.land/bubbles/v2/list"
	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnimeGroupItemFilterValue(t *testing.T) {
	t.Parallel()

	ptbr := &models.Anime{Name: "[PT-BR] Frieren", Source: "Animefire.io", Year: "2023", MediaType: models.MediaTypeAnime, Audio: "Legendado", Quality: "1080p"}
	en := &models.Anime{Name: "[English] Frieren", Source: "HiAnime", Year: "2023", MediaType: models.MediaTypeAnime, Audio: "Sub", Quality: "1080p"}
	groups := groupAnimeResults([]*models.Anime{ptbr, en})
	require.Len(t, groups, 1)
	item := animeGroupItem{group: &groups[0], pos: 0, number: "01"}

	want := "Frieren 2023 Anime AnimeFire PT-BR PT-BR Legendado 1080p HiAnime EN English Sub 1080p"
	assert.Equal(t, want, item.FilterValue())

	t.Run("nil group", func(t *testing.T) {
		t.Parallel()
		assert.Equal(t, "", animeGroupItem{}.FilterValue())
	})
}

func TestAnimeGroupItemTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		item animeGroupItem
		want string
	}{
		{name: "strips language tag", item: animeGroupItem{group: &animeGroup{title: "Frieren"}}, want: "Frieren"},
		{name: "nil", item: animeGroupItem{}, want: "Unknown title"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.item.Title())
		})
	}
}

func TestAnimeGroupItemDescription(t *testing.T) {
	t.Parallel()

	multi := animeGroup{
		title: "Namakura Gatana", year: "1917", media: "Anime",
		versions: []animeVersion{{anime: &models.Anime{}}, {anime: &models.Anime{}}},
	}
	single := animeGroup{
		title: "Momotarou", year: "1918", media: "Anime",
		versions: []animeVersion{{anime: &models.Anime{Source: "Animefire.io"}}},
	}
	tests := []struct {
		name string
		item animeGroupItem
		want string
	}{
		{name: "versions", item: animeGroupItem{group: &multi}, want: "1917 · Anime · 2 versions"},
		{name: "single shows source", item: animeGroupItem{group: &single}, want: "1918 · Anime · AnimeFire"},
		{name: "nil", item: animeGroupItem{}, want: "Information unavailable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.item.Description())
		})
	}
}

func TestGroupAnimeResults(t *testing.T) {
	t.Parallel()

	t.Run("same work across sources is one group", func(t *testing.T) {
		t.Parallel()
		ptbr := &models.Anime{Name: "[PT-BR] Namakura Gatana", Source: "Animefire.io", Year: "1917", MediaType: models.MediaTypeAnime, Audio: "Legendado"}
		en := &models.Anime{Name: "[English] Namakura Gatana", Source: "HiAnime", Year: "1917", MediaType: models.MediaTypeAnime, Audio: "Sub"}
		other := &models.Anime{Name: "[PT-BR] Saru to Kani no Gassen", Source: "Animefire.io", Year: "1917", MediaType: models.MediaTypeAnime}

		groups := groupAnimeResults([]*models.Anime{ptbr, en, nil, other})

		require.Len(t, groups, 2, "nil entries drop out, the rest group by work")
		assert.Equal(t, "Namakura Gatana", groups[0].title)
		require.Len(t, groups[0].versions, 2)
		assert.Same(t, ptbr, groups[0].versions[0].anime, "first version keeps incoming order")
		assert.Same(t, en, groups[0].versions[1].anime)
		assert.Equal(t, []string{"PT-BR", "EN"}, groups[0].badges())
		assert.Equal(t, "1917", groups[0].year)
		assert.Equal(t, "Saru to Kani no Gassen", groups[1].title)
	})

	t.Run("language falls back to source", func(t *testing.T) {
		t.Parallel()
		groups := groupAnimeResults([]*models.Anime{
			{Name: "Frieren", Source: "HiAnime"},
			{Name: "Naruto", Source: "Goyabu"},
		})
		require.Len(t, groups, 2)
		assert.Equal(t, "EN", groups[0].versions[0].lang)
		assert.Equal(t, "English", groups[0].versions[0].langWord)
		assert.Equal(t, "PT-BR", groups[1].versions[0].lang)
	})

	t.Run("empty", func(t *testing.T) {
		t.Parallel()
		assert.Empty(t, groupAnimeResults(nil))
		assert.Empty(t, groupAnimeResults([]*models.Anime{nil, nil}))
	})
}

func TestCleanResultTitle(t *testing.T) {
	t.Parallel()

	tests := []struct{ name, want string }{
		{"[PT-BR] Namakura Gatana", "Namakura Gatana"},
		{"[Movie] [PT-BR] Inception", "Inception"},
		{"[English] Frieren", "Frieren"},
		{"Attack on Titan (Legendado)", "Attack on Titan"},
		{"Black Clover (Shippuuden)", "Black Clover (Shippuuden)"},
		{"  ", "Unknown title"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, cleanResultTitle(tt.name))
		})
	}
}

func TestSingleLine(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name  string
		value string
		want  string
	}{
		{name: "plain", value: "Frieren", want: "Frieren"},
		{name: "whitespace", value: "  Frieren\tBeyond   Journey  ", want: "Frieren Beyond Journey"},
		{name: "terminal controls", value: "Title\r\nInjected\x1b[2J\x00", want: "Title Injected"},
		{name: "empty", value: "", want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, singleLine(tt.value))
			// Exported wrapper must stay identical to the internal helper.
			assert.Equal(t, tt.want, SingleLine(tt.value))
		})
	}
}

func FuzzSingleLineNeverReturnsControl(f *testing.F) {
	for _, seed := range []string{"Frieren", "Title\r\nInjected\x1b[2J", "日本語 🎬", "\x00\x1b]8;;https://example.com\aevil\x1b]8;;\a"} {
		f.Add(seed)
	}
	f.Fuzz(func(t *testing.T, value string) {
		got := singleLine(value)
		assert.False(t, strings.ContainsFunc(got, unicode.IsControl))
	})
}

func TestNewAnimeResultsModel(t *testing.T) {
	t.Parallel()

	model := newAnimeResultsModel([]*models.Anime{nil, {Name: "Frieren"}, nil, {Name: "Dungeon Meshi"}})

	require.NotNil(t, model)
	assert.Len(t, model.results.Items(), 2)
	assert.False(t, model.results.ShowTitle())
	assert.False(t, model.results.ShowHelp())
	assert.Equal(t, "Search > Results", model.shell.Breadcrumb)
	assert.Equal(t, "❯ ", model.results.FilterInput.Prompt)
}

func TestNewAnimeResultsModelWithQuery(t *testing.T) {
	t.Parallel()

	model := newAnimeResultsModelWithQuery([]*models.Anime{{Name: "[PT-BR] Namakura Gatana"}}, "namakura")

	assert.Equal(t, `"namakura"`, model.shell.Context, "the query reads quoted at the right of the header")
}

func TestAnimeResultsModelGroupsVersions(t *testing.T) {
	t.Parallel()

	ptbr := &models.Anime{Name: "[PT-BR] Namakura Gatana", Source: "Animefire.io", Year: "1917", MediaType: models.MediaTypeAnime, Audio: "Legendado"}
	en := &models.Anime{Name: "[English] Namakura Gatana", Source: "HiAnime", Year: "1917", MediaType: models.MediaTypeAnime, Audio: "Sub"}
	model := newAnimeResultsModel([]*models.Anime{ptbr, en})
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 30})

	t.Run("enter opens the first version", func(t *testing.T) {
		t.Parallel()
		m := newAnimeResultsModel([]*models.Anime{ptbr, en})
		_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		require.NotNil(t, cmd)
		assert.Same(t, ptbr, m.selected)
	})

	t.Run("right picks the next version", func(t *testing.T) {
		m := newAnimeResultsModel([]*models.Anime{ptbr, en})
		_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
		assert.Equal(t, 1, m.versionSel[0])

		_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		require.NotNil(t, cmd)
		assert.Same(t, en, m.selected, "enter opens the marked version")
	})

	t.Run("versions wrap around", func(t *testing.T) {
		m := newAnimeResultsModel([]*models.Anime{ptbr, en})
		_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyLeft})
		assert.Equal(t, 1, m.versionSel[0], "left from the first wraps to the last")
	})

	t.Run("single version has nowhere to move", func(t *testing.T) {
		m := newAnimeResultsModel([]*models.Anime{ptbr})
		updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyRight})
		assert.Nil(t, cmd, "no version to move to: the key reaches the list")
		assert.Same(t, m, updated)
		assert.Empty(t, m.versionSel)
	})

	t.Run("details panel lists versions", func(t *testing.T) {
		view := ansi.Strip(model.View().Content)
		assert.Contains(t, view, "VERSIONS 2 · ←→ to choose")
		assert.Contains(t, view, "▶ AnimeFire PT-BR · Legendado")
		assert.Contains(t, view, "HiAnime English · Sub")
		assert.Contains(t, view, "enter opens it on AnimeFire")
	})

	t.Run("typing filters without slash", func(t *testing.T) {
		m := newAnimeResultsModel([]*models.Anime{
			{Name: "[PT-BR] Namakura Gatana", Source: "Animefire.io"},
			{Name: "[PT-BR] Momotarou", Source: "Animefire.io"},
		})
		_, _ = m.Update(tea.KeyPressMsg{Code: 'm', Text: "m"})
		assert.Equal(t, list.Filtering, m.results.FilterState())
		assert.Equal(t, "m", m.results.FilterInput.Value())
	})
}

func TestAnimeResultsModelInit(t *testing.T) {
	t.Parallel()

	model := newAnimeResultsModel([]*models.Anime{{Name: "Frieren"}})
	assert.Nil(t, model.Init())
}

func TestAnimeResultsModelUpdate(t *testing.T) {
	t.Parallel()

	t.Run("cascade resize navigate select", func(t *testing.T) {
		t.Parallel()
		first := &models.Anime{Name: "Frieren"}
		second := &models.Anime{Name: "Dungeon Meshi"}
		model := newAnimeResultsModel([]*models.Anime{first, second})

		updated, cmd := model.Update(tea.WindowSizeMsg{Width: 120, Height: 40})
		require.Same(t, model, updated)
		assert.Nil(t, cmd)
		assert.Equal(t, 70, model.results.Width(), "3/5 of the width, less the gutter before the details panel")
		assert.Equal(t, 36, model.results.Height())

		_, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyDown})
		assert.Equal(t, 1, model.results.Index())

		_, cmd = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		require.NotNil(t, cmd)
		assert.Same(t, second, model.selected)
		assert.IsType(t, tea.QuitMsg{}, cmd())
	})

	tests := []struct {
		name    string
		msg     tea.KeyPressMsg
		wantErr error
	}{
		{name: "escape goes back", msg: tea.KeyPressMsg{Code: tea.KeyEscape}, wantErr: ErrSelectionBack},
		{name: "q cancels", msg: tea.KeyPressMsg{Code: 'q', Text: "q"}, wantErr: ErrSelectionCancelled},
		{name: "ctrl c cancels", msg: tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl}, wantErr: ErrSelectionCancelled},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			model := newAnimeResultsModel([]*models.Anime{{Name: "Frieren"}})
			_, cmd := model.Update(tt.msg)
			require.NotNil(t, cmd)
			assert.ErrorIs(t, model.err, tt.wantErr)
		})
	}

	t.Run("escape clears applied filter before going back", func(t *testing.T) {
		t.Parallel()
		model := newAnimeResultsModel([]*models.Anime{{Name: "Frieren"}})
		model.results.SetFilterText("fri")
		require.Equal(t, list.FilterApplied, model.results.FilterState())

		_, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})

		assert.NoError(t, model.err)
		assert.Equal(t, list.Unfiltered, model.results.FilterState())

		_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		require.NotNil(t, cmd)
		assert.ErrorIs(t, model.err, ErrSelectionBack)
	})

	t.Run("filter editing does not trigger actions", func(t *testing.T) {
		t.Parallel()
		model := newAnimeResultsModel([]*models.Anime{{Name: "Frieren"}, {Name: "Dungeon Meshi"}})
		model.results.SetFilterState(list.Filtering)

		_, _ = model.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
		assert.NoError(t, model.err)
		assert.Nil(t, model.selected)
		assert.Equal(t, "q", model.results.FilterInput.Value())
		assert.True(t, model.filterPending)

		_, _ = model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
		assert.NoError(t, model.err)
		assert.Nil(t, model.selected)
		assert.Equal(t, list.Filtering, model.results.FilterState())
	})

	t.Run("filtered selection uses visible item", func(t *testing.T) {
		t.Parallel()
		first := &models.Anime{Name: "Frieren"}
		second := &models.Anime{Name: "Dungeon Meshi", Quality: "1080p"}
		model := newAnimeResultsModel([]*models.Anime{first, second})
		model.results.SetFilterText("1080P")
		require.Len(t, model.results.VisibleItems(), 1)

		_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		require.NotNil(t, cmd)
		assert.Same(t, second, model.selected)
	})

	t.Run("empty filtered result cannot quit or select", func(t *testing.T) {
		t.Parallel()
		model := newAnimeResultsModel([]*models.Anime{{Name: "Frieren"}})
		model.results.SetFilterText("does-not-exist")
		require.Empty(t, model.results.VisibleItems())

		_, cmd := model.Update(tea.KeyPressMsg{Code: tea.KeyEnter})

		assert.Nil(t, cmd)
		assert.Nil(t, model.selected)
		assert.NoError(t, model.err)
	})

	t.Run("responsive breakpoint preserves state", func(t *testing.T) {
		t.Parallel()
		model := newAnimeResultsModel([]*models.Anime{{Name: "Frieren"}, {Name: "Dungeon Meshi"}})
		model.results.Select(1)
		for _, size := range []struct {
			width     int
			wantWidth int
		}{{99, 99}, {100, 58}, {101, 58}, {99, 99}} {
			_, _ = model.Update(tea.WindowSizeMsg{Width: size.width, Height: 24})
			assert.Equal(t, size.wantWidth, model.results.Width())
			assert.Equal(t, 1, model.results.Index())
		}
	})
}

func TestAnimeResultsModelView(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name        string
		width       int
		height      int
		wantDetails bool
	}{
		{name: "minimum compact", width: 20, height: 5, wantDetails: false},
		{name: "compact", width: 80, height: 8, wantDetails: false},
		{name: "wide boundary", width: 100, height: 18, wantDetails: true},
		{name: "wide", width: 120, height: 18, wantDetails: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			model := newAnimeResultsModel([]*models.Anime{{Name: "Frieren with a very long title that must not overflow the terminal", Source: "AnimeFire", Year: "2023", MediaType: models.MediaTypeAnime, Quality: "1080p"}})
			_, _ = model.Update(tea.WindowSizeMsg{Width: tt.width, Height: tt.height})

			view := model.View()

			assert.True(t, view.AltScreen)
			assert.Equal(t, "GoAnime - Results", view.WindowTitle)
			if tt.width >= 9 {
				assert.Contains(t, view.Content, "GOANIME")
			}
			assert.Equal(t, tt.wantDetails, strings.Contains(ansi.Strip(view.Content), "DETAILS"))
			assert.LessOrEqual(t, lipgloss.Width(view.Content), tt.width)
			assert.LessOrEqual(t, lipgloss.Height(view.Content), tt.height)
		})
	}
}

// A step that sends the user back to the results explains why on the result
// screen itself, since anything printed before the screen opens is hidden
// behind it. Not parallel: the notice is package-level.
func TestAnimeResultsModel_ShowsNoticeOnce(t *testing.T) {
	NoticeOnNextResults("Sousou no Frieren 3rd Season has no episodes on HiAnime yet.")
	animes := []*models.Anime{{Name: "Sousou no Frieren", Source: "HiAnime"}}

	model := newAnimeResultsModel(animes)
	for _, size := range []tea.WindowSizeMsg{{Width: 80, Height: 12}, {Width: 120, Height: 18}} {
		_, _ = model.Update(size)
		view := model.View().Content
		assert.Contains(t, view, "has no episodes on HiAnime yet", "width %d", size.Width)
		assert.Contains(t, view, "Sousou no Frieren", "the notice must not push the results off screen")
		assert.LessOrEqual(t, lipgloss.Height(view), size.Height)
		assert.LessOrEqual(t, lipgloss.Width(view), size.Width)
	}

	again := newAnimeResultsModel(animes)
	_, _ = again.Update(tea.WindowSizeMsg{Width: 120, Height: 18})
	assert.NotContains(t, again.View().Content, "has no episodes", "a notice is shown once")
}

func TestDetailsPanelLines(t *testing.T) {
	t.Parallel()

	theme := NewTheme(true)
	content := &panelContent{
		label: "Details", gloss: "詳細", title: "The Boys",
		fields: []detailField{{"Source", "StartFlix"}, {"Year", "2019"}, {"Type", "TV"}, {"Audio", "Dubbed and subtitled"}},
	}
	for name, compact := range map[string]bool{"stacked": false, "compact": true} {
		details := ansi.Strip(strings.Join(panelLines(&theme, content, 40, compact), "\n"))
		assert.Contains(t, details, "DETAILS  詳細", name)
		assert.Contains(t, details, "The Boys", name)
		assert.Contains(t, details, "StartFlix", name)
		assert.Contains(t, details, "Audio", name)
		assert.Contains(t, details, "Dubbed and subtitled", name)
		assert.NotContains(t, details, "Quality", "%s: no source fills it, so the panel does not offer it", name)
	}
	assert.Len(t, panelLines(&theme, content, 40, true), 6, "compact: heading, title and one line per field")
}

// On a short terminal the details panel must keep its versions and its
// bottom border instead of running off the screen. A notice above the list
// takes two rows, which is what first pushed it off.
func TestAnimeResultsModel_DetailsPanelFitsAShortScreen(t *testing.T) {
	anime := &models.Anime{Name: "[PT-BR] Frieren e a Jornada para o Além", Source: "Animefire.io",
		Year: "2023", MediaType: models.MediaTypeAnime, Audio: "Legendado"}

	for _, height := range []int{18, 24, 40} {
		NoticeOnNextResults("A title has no episodes yet.") // not parallel: package-level notice
		model := newAnimeResultsModel([]*models.Anime{anime})
		_, _ = model.Update(tea.WindowSizeMsg{Width: 110, Height: height})
		view := model.View().Content

		assert.Contains(t, view, "VERSIONS", "height %d: the versions were cut", height)
		assert.Contains(t, view, "Legendado", "height %d: the version audio was cut", height)
		assert.Contains(t, view, "▀▀▀", "height %d: the panel lost its shadow", height)
		assert.LessOrEqual(t, lipgloss.Height(view), height)
	}

	// Too short for even the compact panel: no panel at all, never half of
	// one. The result rows still carry the same details.
	NoticeOnNextResults("A title has no episodes yet.")
	model := newAnimeResultsModel([]*models.Anime{anime})
	_, _ = model.Update(tea.WindowSizeMsg{Width: 110, Height: 12})
	view := ansi.Strip(model.View().Content)
	assert.NotContains(t, view, "DETAILS")
	assert.Contains(t, view, "AnimeFire")
}

func TestSelectAnimeWithRunner(t *testing.T) {
	animes := []*models.Anime{{Name: "Frieren"}, {Name: "Dungeon Meshi"}}

	t.Run("mocked selection", func(t *testing.T) {
		selected, err := selectAnimeWithRunner(animes, func(model tea.Model) (tea.Model, error) {
			result := model.(*animeResultsModel)
			result.selected = animes[1]
			return result, nil
		})
		require.NoError(t, err)
		assert.Same(t, animes[1], selected)
	})

	t.Run("mocked back", func(t *testing.T) {
		_, err := selectAnimeWithRunner(animes, func(model tea.Model) (tea.Model, error) {
			result := model.(*animeResultsModel)
			result.err = ErrSelectionBack
			return result, nil
		})
		assert.ErrorIs(t, err, ErrSelectionBack)
	})

	t.Run("mocked runner failure", func(t *testing.T) {
		wantErr := errors.New("terminal failed")
		_, err := selectAnimeWithRunner(animes, func(tea.Model) (tea.Model, error) { return nil, wantErr })
		assert.ErrorIs(t, err, wantErr)
	})

	t.Run("mocked unexpected final model", func(t *testing.T) {
		_, err := selectAnimeWithRunner(animes, func(tea.Model) (tea.Model, error) {
			return minimalModel{}, nil
		})
		require.Error(t, err)
		assert.Contains(t, err.Error(), "unexpected anime result model")
	})

	t.Run("mocked quit without selection", func(t *testing.T) {
		_, err := selectAnimeWithRunner(animes, func(model tea.Model) (tea.Model, error) {
			return model, nil
		})
		assert.ErrorIs(t, err, ErrSelectionCancelled)
	})

	t.Run("typed nil final model", func(t *testing.T) {
		assert.NotPanics(t, func() {
			_, err := selectAnimeWithRunner(animes, func(tea.Model) (tea.Model, error) {
				var result *animeResultsModel
				return result, nil
			})
			require.Error(t, err)
		})
	})

	t.Run("all nil results", func(t *testing.T) {
		called := false
		_, err := selectAnimeWithRunner([]*models.Anime{nil, nil}, func(tea.Model) (tea.Model, error) {
			called = true
			return nil, nil
		})
		assert.ErrorIs(t, err, ErrNoAnimeResults)
		assert.False(t, called)
	})

	t.Run("mixed nil results are removed before runner", func(t *testing.T) {
		valid := &models.Anime{Name: "Frieren"}
		selected, err := selectAnimeWithRunner([]*models.Anime{nil, valid, nil}, func(model tea.Model) (tea.Model, error) {
			result := model.(*animeResultsModel)
			require.Len(t, result.results.Items(), 1)
			result.selected = valid
			return result, nil
		})
		require.NoError(t, err)
		assert.Same(t, valid, selected)
	})

	t.Run("nil runner", func(t *testing.T) {
		_, err := selectAnimeWithRunner(animes, nil)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "runner not configured")
	})

	t.Run("empty", func(t *testing.T) {
		called := false
		_, err := selectAnimeWithRunner(nil, func(tea.Model) (tea.Model, error) {
			called = true
			return nil, nil
		})
		assert.ErrorIs(t, err, ErrNoAnimeResults)
		assert.False(t, called)
	})
}

func TestSelectAnime(t *testing.T) {
	t.Parallel()

	for _, animes := range [][]*models.Anime{nil, {}, {nil, nil}} {
		selected, err := SelectAnime(animes)
		assert.Nil(t, selected)
		assert.ErrorIs(t, err, ErrNoAnimeResults)
	}
}
