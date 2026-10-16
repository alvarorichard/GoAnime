package tui

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func homeTestEntries() ([]HomeContinueEntry, []HomeRecentSearch) {
	now := time.Now()
	continueAll := []HomeContinueEntry{
		{Key: "k1", AnimeTitle: "Mukashi Banashi", EpisodeTitle: "Urashima Tarou", EpisodeNum: 5, Position: 372, Duration: 580, Source: "Animefire.io", Audio: "Legendado", UpdatedAt: now.Add(-24 * time.Hour)},
		{Key: "k2", AnimeTitle: "Kuro Nyago", EpisodeNum: 1, Position: 160, Duration: 600, Source: "Animefire.io", UpdatedAt: now.Add(-72 * time.Hour)},
		{Key: "k3", AnimeTitle: "Saru to Kani no Gassen", EpisodeNum: 1, Position: 590, Duration: 600, UpdatedAt: now.Add(-7 * 24 * time.Hour)},
	}
	recent := []HomeRecentSearch{
		{Query: "namakura gatana", Count: 5},
		{Query: "momotarou", Count: 3},
	}
	return continueAll, recent
}

func press(code rune, text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: text}
}

func TestNewSearchHomeModel(t *testing.T) {
	t.Parallel()

	continueAll, recent := homeTestEntries()
	model := newSearchHomeModel(SearchHomeOptions{
		Query: "namak", ScopeIndex: 9, SourcesLine: "6 sources on · Goyabu off",
		Continue: continueAll, Recent: recent,
	})

	assert.True(t, model.focusInput)
	assert.Equal(t, "namak", model.input.Value())
	assert.Equal(t, 0, model.scopeIdx, "out-of-range scope clamps to ALL")
	assert.Equal(t, "6 sources on · Goyabu off", model.shell.Context)
	assert.Len(t, model.rows, 5)
	assert.Equal(t, homeRowContinue, model.rows[0].kind)
	assert.Equal(t, homeRowRecent, model.rows[3].kind)
}

func TestSearchHomeEnter(t *testing.T) {
	t.Parallel()

	t.Run("input searches with scope", func(t *testing.T) {
		t.Parallel()
		continueAll, recent := homeTestEntries()
		model := newSearchHomeModel(SearchHomeOptions{Continue: continueAll, Recent: recent, ScopeIndex: 1})
		model.input.SetValue("frieren")

		updated, cmd := model.Update(press(tea.KeyEnter, ""))
		require.NotNil(t, cmd)
		m := updated.(*searchHomeModel)
		assert.Equal(t, HomeSearch, m.result.Action)
		assert.Equal(t, "frieren", m.result.Query)
		assert.Equal(t, "ptbr", m.result.Scope)
	})

	t.Run("short query stays", func(t *testing.T) {
		t.Parallel()
		model := newSearchHomeModel(SearchHomeOptions{})
		model.input.SetValue("x")
		_, cmd := model.Update(press(tea.KeyEnter, ""))
		assert.Nil(t, cmd)
	})

	t.Run("continue resumes", func(t *testing.T) {
		t.Parallel()
		continueAll, recent := homeTestEntries()
		model := newSearchHomeModel(SearchHomeOptions{Continue: continueAll, Recent: recent})
		_, _ = model.Update(press(tea.KeyDown, ""))

		updated, cmd := model.Update(press(tea.KeyEnter, ""))
		require.NotNil(t, cmd)
		m := updated.(*searchHomeModel)
		assert.Equal(t, HomeResume, m.result.Action)
		assert.True(t, m.result.HasEntry)
		assert.Equal(t, "Mukashi Banashi", m.result.Entry.AnimeTitle)
		assert.Equal(t, 5, m.result.Entry.EpisodeNum)
	})

	t.Run("recent searches again", func(t *testing.T) {
		t.Parallel()
		continueAll, recent := homeTestEntries()
		model := newSearchHomeModel(SearchHomeOptions{Continue: continueAll, Recent: recent})
		for range 5 {
			_, _ = model.Update(press(tea.KeyDown, ""))
		}
		updated, cmd := model.Update(press(tea.KeyEnter, ""))
		require.NotNil(t, cmd)
		m := updated.(*searchHomeModel)
		assert.Equal(t, HomeSearch, m.result.Action)
		assert.Equal(t, "momotarou", m.result.Query)
	})
}

func TestSearchHomeRowActions(t *testing.T) {
	t.Parallel()

	t.Run("r from start", func(t *testing.T) {
		t.Parallel()
		continueAll, _ := homeTestEntries()
		model := newSearchHomeModel(SearchHomeOptions{Continue: continueAll})
		_, _ = model.Update(press(tea.KeyDown, ""))
		updated, cmd := model.Update(press('r', "r"))
		require.NotNil(t, cmd)
		assert.Equal(t, HomeFromStart, updated.(*searchHomeModel).result.Action)
	})

	t.Run("e episodes", func(t *testing.T) {
		t.Parallel()
		continueAll, _ := homeTestEntries()
		model := newSearchHomeModel(SearchHomeOptions{Continue: continueAll})
		_, _ = model.Update(press(tea.KeyDown, ""))
		updated, cmd := model.Update(press('e', "e"))
		require.NotNil(t, cmd)
		assert.Equal(t, HomeEpisodes, updated.(*searchHomeModel).result.Action)
	})

	t.Run("x removes the row", func(t *testing.T) {
		t.Parallel()
		continueAll, _ := homeTestEntries()
		var deleted string
		model := newSearchHomeModel(SearchHomeOptions{
			Continue: continueAll,
			DeleteEntry: func(key string) error {
				deleted = key
				return nil
			},
		})
		_, _ = model.Update(press(tea.KeyDown, ""))
		_, cmd := model.Update(press('x', "x"))
		assert.Nil(t, cmd, "remove stays on screen")
		assert.Equal(t, "k1", deleted)
		assert.Len(t, model.continueAll, 2)
		assert.Len(t, model.rows, 2)
	})

	t.Run("x without delete support does nothing", func(t *testing.T) {
		t.Parallel()
		continueAll, _ := homeTestEntries()
		model := newSearchHomeModel(SearchHomeOptions{Continue: continueAll})
		_, _ = model.Update(press(tea.KeyDown, ""))
		_, cmd := model.Update(press('x', "x"))
		assert.Nil(t, cmd)
		assert.Len(t, model.continueAll, 3)
	})

	t.Run("x on recent does nothing", func(t *testing.T) {
		t.Parallel()
		_, recent := homeTestEntries()
		called := false
		model := newSearchHomeModel(SearchHomeOptions{
			Recent:      recent,
			DeleteEntry: func(string) error { called = true; return nil },
		})
		_, _ = model.Update(press(tea.KeyDown, ""))
		_, _ = model.Update(press('x', "x"))
		assert.False(t, called)
	})

	t.Run("delete error keeps the row", func(t *testing.T) {
		t.Parallel()
		continueAll, _ := homeTestEntries()
		model := newSearchHomeModel(SearchHomeOptions{
			Continue:    continueAll,
			DeleteEntry: func(string) error { return errors.New("db locked") },
		})
		_, _ = model.Update(press(tea.KeyDown, ""))
		_, _ = model.Update(press('x', "x"))
		assert.Len(t, model.continueAll, 3)
	})
}

func TestSearchHomeNavigation(t *testing.T) {
	t.Parallel()

	t.Run("tab cycles scopes", func(t *testing.T) {
		t.Parallel()
		model := newSearchHomeModel(SearchHomeOptions{})
		for i, want := range []int{1, 2, 3, 0} {
			_, _ = model.Update(press(tea.KeyTab, ""))
			assert.Equal(t, want, model.scopeIdx, "tab %d", i)
		}
	})

	t.Run("up returns to the box", func(t *testing.T) {
		t.Parallel()
		continueAll, _ := homeTestEntries()
		model := newSearchHomeModel(SearchHomeOptions{Continue: continueAll})
		_, _ = model.Update(press(tea.KeyDown, ""))
		require.False(t, model.focusInput)
		_, _ = model.Update(press(tea.KeyUp, ""))
		assert.True(t, model.focusInput)
	})

	t.Run("esc clears instead of quitting", func(t *testing.T) {
		t.Parallel()
		model := newSearchHomeModel(SearchHomeOptions{Query: "namak"})
		_, cmd := model.Update(press(tea.KeyEscape, ""))
		assert.Nil(t, cmd)
		assert.NoError(t, model.err)
		assert.Equal(t, "", model.input.Value())

		_, cmd = model.Update(press(tea.KeyEscape, ""))
		assert.Nil(t, cmd, "empty box: esc is a noop, never a quit")
		assert.NoError(t, model.err)
	})

	t.Run("ctrl+c cancels", func(t *testing.T) {
		t.Parallel()
		model := newSearchHomeModel(SearchHomeOptions{})
		_, cmd := model.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
		require.NotNil(t, cmd)
		assert.ErrorIs(t, model.err, ErrSearchHomeCancelled)
	})
}

func TestSearchHomeTyping(t *testing.T) {
	t.Parallel()

	t.Run("typing over the list returns to the box", func(t *testing.T) {
		t.Parallel()
		continueAll, _ := homeTestEntries()
		model := newSearchHomeModel(SearchHomeOptions{Continue: continueAll})
		_, _ = model.Update(press(tea.KeyDown, ""))
		require.False(t, model.focusInput)
		_, _ = model.Update(press('f', "f"))
		assert.True(t, model.focusInput)
		assert.Equal(t, "f", model.input.Value())
	})

	t.Run("question in the box types", func(t *testing.T) {
		t.Parallel()
		model := newSearchHomeModel(SearchHomeOptions{})
		_, _ = model.Update(press('?', "?"))
		assert.Equal(t, "?", model.input.Value())
		assert.False(t, model.showHelp)
	})

	t.Run("question over the list opens help", func(t *testing.T) {
		t.Parallel()
		continueAll, _ := homeTestEntries()
		model := newSearchHomeModel(SearchHomeOptions{Continue: continueAll})
		_, _ = model.Update(press(tea.KeyDown, ""))
		_, _ = model.Update(press('?', "?"))
		assert.True(t, model.showHelp)
		_, _ = model.Update(press(tea.KeyEscape, ""))
		assert.False(t, model.showHelp)
	})
}

func TestRelativeDay(t *testing.T) {
	t.Parallel()
	now := time.Date(2026, 10, 14, 12, 0, 0, 0, time.UTC)
	tests := []struct {
		name string
		when time.Time
		want string
	}{
		{"today", now.Add(-2 * time.Hour), "today"},
		{"yesterday", now.Add(-24 * time.Hour), "yesterday"},
		{"days ago", now.Add(-72 * time.Hour), "3 days ago"},
		{"last week", now.Add(-7 * 24 * time.Hour), "last week"},
		{"weeks ago", now.Add(-21 * 24 * time.Hour), "3 weeks ago"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, relativeDay(tt.when, now))
		})
	}
}

func TestHomeContinueEntry(t *testing.T) {
	t.Parallel()
	assert.True(t, HomeContinueEntry{Position: 590, Duration: 600}.Finished())
	assert.False(t, HomeContinueEntry{Position: 372, Duration: 580}.Finished())
	assert.Equal(t, "Show", HomeContinueEntry{AnimeTitle: "Show", EpisodeTitle: "Ep"}.DisplayTitle())
	assert.Equal(t, "Ep", HomeContinueEntry{EpisodeTitle: "Ep"}.DisplayTitle())
	assert.Equal(t, "Unknown title", HomeContinueEntry{}.DisplayTitle())
}

func TestSearchHomeView(t *testing.T) {
	t.Parallel()

	continueAll, recent := homeTestEntries()
	model := newSearchHomeModel(SearchHomeOptions{
		SourcesLine: "6 sources on · Goyabu off", Continue: continueAll, Recent: recent,
	})
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
	view := ansi.Strip(model.View().Content)

	assert.Contains(t, view, "GOANIME")
	assert.Contains(t, view, "SEARCH")
	assert.Contains(t, view, "6 sources on")
	assert.Contains(t, view, "ALL")
	assert.Contains(t, view, "PT-BR")
	assert.Contains(t, view, "CONTINUE")
	assert.Contains(t, view, "Mukashi Banashi")
	assert.Contains(t, view, "Ep 5 · 06:12")
	assert.Contains(t, view, "yesterday")
	assert.Contains(t, view, "Ep 1 · done")
	assert.Contains(t, view, "RECENT")
	assert.Contains(t, view, "namakura gatana")
	assert.Contains(t, view, "5 results")
	assert.Contains(t, view, "Episode 5 · Urashima Tarou")
	assert.Contains(t, view, "Animefire.io · Legendado")
	assert.Contains(t, view, "resume")
	assert.Contains(t, view, "from start")
	assert.Contains(t, view, "enter")
	assert.Contains(t, view, "ctrl+c")
}

func TestSearchHomeViewPlain(t *testing.T) {
	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })

	continueAll, recent := homeTestEntries()
	model := newSearchHomeModel(SearchHomeOptions{Continue: continueAll, Recent: recent})
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
	view := ansi.Strip(model.View().Content)

	assert.Contains(t, view, "CONTINUE")
	assert.NotContains(t, view, "続き", "plain hides Japanese glosses")
	assert.NotContains(t, view, "ゴーアニメ")
	assert.Contains(t, view, ">", "plain cursor is ASCII")
}

func TestSearchHomeEmpty(t *testing.T) {
	t.Parallel()

	model := newSearchHomeModel(SearchHomeOptions{})
	assert.Empty(t, model.rows)
	assert.True(t, model.focusInput)
	_, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})
	view := ansi.Strip(model.View().Content)
	assert.Contains(t, view, "GOANIME")
	assert.NotContains(t, view, "CONTINUE")
	assert.NotContains(t, view, "RECENT")
}

func TestSearchHomeWithRunner(t *testing.T) {
	continueAll, _ := homeTestEntries()

	t.Run("resume passthrough", func(t *testing.T) {
		res, err := searchHomeWithRunner(SearchHomeOptions{Continue: continueAll}, func(model tea.Model) (tea.Model, error) {
			m := model.(*searchHomeModel)
			m.result = SearchHomeResult{Action: HomeResume, Entry: continueAll[0], HasEntry: true}
			return m, nil
		})
		require.NoError(t, err)
		assert.Equal(t, HomeResume, res.Action)
		assert.Equal(t, "Mukashi Banashi", res.Entry.AnimeTitle)
	})

	t.Run("cancelled", func(t *testing.T) {
		_, err := searchHomeWithRunner(SearchHomeOptions{}, func(model tea.Model) (tea.Model, error) {
			m := model.(*searchHomeModel)
			m.err = ErrSearchHomeCancelled
			return m, nil
		})
		assert.ErrorIs(t, err, ErrSearchHomeCancelled)
	})

	t.Run("empty search is a cancel", func(t *testing.T) {
		_, err := searchHomeWithRunner(SearchHomeOptions{}, func(model tea.Model) (tea.Model, error) {
			return model, nil
		})
		assert.ErrorIs(t, err, ErrSearchHomeCancelled)
	})

	t.Run("nil runner", func(t *testing.T) {
		_, err := searchHomeWithRunner(SearchHomeOptions{}, nil)
		assert.ErrorContains(t, err, "runner not configured")
	})

	t.Run("unexpected model", func(t *testing.T) {
		_, err := searchHomeWithRunner(SearchHomeOptions{}, func(tea.Model) (tea.Model, error) {
			return minimalModel{}, nil
		})
		assert.ErrorContains(t, err, "unexpected search home model")
	})

	t.Run("runner error", func(t *testing.T) {
		want := errors.New("no tty")
		_, err := searchHomeWithRunner(SearchHomeOptions{}, func(tea.Model) (tea.Model, error) {
			return nil, want
		})
		assert.ErrorIs(t, err, want)
	})

	t.Run("query with only spaces is a cancel", func(t *testing.T) {
		_, err := searchHomeWithRunner(SearchHomeOptions{}, func(model tea.Model) (tea.Model, error) {
			m := model.(*searchHomeModel)
			m.result = SearchHomeResult{Action: HomeSearch, Query: "   "}
			return m, nil
		})
		assert.ErrorIs(t, err, ErrSearchHomeCancelled)
	})
}
