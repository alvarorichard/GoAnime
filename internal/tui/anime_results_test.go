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
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnimeResultItemFilterValue(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		item animeResultItem
		want string
	}{
		{name: "metadata", item: animeResultItem{anime: &models.Anime{Name: "Frieren", Source: "AnimeFire", Year: "2023", MediaType: models.MediaTypeAnime, Quality: "1080p"}}, want: "Frieren AnimeFire 2023 anime 1080p"},
		{name: "control sequences", item: animeResultItem{anime: &models.Anime{Name: "Title\nfake\x1b[2J", Source: "Source\rInjected"}}, want: "Title fake Source Injected   "},
		{name: "nil", item: animeResultItem{}, want: ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.item.FilterValue())
		})
	}
}

func TestAnimeResultItemTitle(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		item animeResultItem
		want string
	}{
		{name: "title", item: animeResultItem{anime: &models.Anime{Name: "Frieren"}}, want: "Frieren"},
		{name: "single line", item: animeResultItem{anime: &models.Anime{Name: "Title\nfake\x1b[2J"}}, want: "Title fake"},
		{name: "blank", item: animeResultItem{anime: &models.Anime{Name: "  "}}, want: "Unknown title"},
		{name: "nil", item: animeResultItem{}, want: "Unknown title"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.item.Title())
		})
	}
}

func TestAnimeResultItemDescription(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name string
		item animeResultItem
		want string
	}{
		{name: "complete", item: animeResultItem{anime: &models.Anime{Source: "AnimeFire", Year: "2023", MediaType: models.MediaTypeAnime, Quality: "1080p"}}, want: "AnimeFire  •  2023  •  anime  •  1080p"},
		{name: "partial", item: animeResultItem{anime: &models.Anime{Source: "AllAnime"}}, want: "AllAnime"},
		{name: "single line", item: animeResultItem{anime: &models.Anime{Source: "Source\r\nInjected\x1b[2J"}}, want: "Source Injected"},
		{name: "empty", item: animeResultItem{anime: &models.Anime{}}, want: "Information unavailable"},
		{name: "nil", item: animeResultItem{}, want: "Information unavailable"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.item.Description())
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
	assert.Equal(t, "Filter: ", model.results.FilterInput.Prompt)
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
		assert.Equal(t, 72, model.results.Width())
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
		}{{99, 99}, {100, 60}, {101, 60}, {99, 99}} {
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
		{name: "wide boundary", width: 100, height: 10, wantDetails: true},
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
			assert.Equal(t, tt.wantDetails, strings.Contains(view.Content, "Details"))
			assert.LessOrEqual(t, lipgloss.Width(view.Content), tt.width)
			assert.LessOrEqual(t, lipgloss.Height(view.Content), tt.height)
		})
	}
}

func TestRenderAnimeDetails(t *testing.T) {
	t.Parallel()

	theme := NewTheme(true)
	details := renderAnimeDetails(theme, "The Boys", "SuperFlix", "2019", "tv", "—")

	assert.Contains(t, details, "Details")
	assert.Contains(t, details, "The Boys")
	assert.Contains(t, details, "SuperFlix")
	assert.NotContains(t, details, theme.Primary.Render("Details")+" ")
	for _, line := range strings.Split(details, "\n") {
		assert.False(t, strings.HasSuffix(line, " "), "line has unstyled trailing padding: %q", line)
	}
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
