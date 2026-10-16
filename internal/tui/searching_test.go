package tui

import (
	"errors"
	"testing"
	"time"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func searchingTestEvents() (plan, fire, hi, topcine models.SearchEvent) {
	fireResults := []*models.Anime{
		{Name: "[PT-BR] Namakura Gatana", Source: "Animefire.io", Year: "1917", MediaType: models.MediaTypeAnime, Audio: "Legendado"},
		{Name: "[PT-BR] Saru to Kani no Gassen", Source: "Animefire.io", Year: "1917", MediaType: models.MediaTypeAnime},
		{Name: "[PT-BR] Momotarou", Source: "Animefire.io", Year: "1918", MediaType: models.MediaTypeAnime},
	}
	plan = models.SearchEvent{
		Plan:     true,
		Searched: []string{"AnimeFire", "HiAnime", "TopCine"},
		Off:      []models.SearchOffEntry{{Name: "Goyabu", Reason: "off by default"}},
		GiveUp:   15 * time.Second,
	}
	fire = models.SearchEvent{Name: "AnimeFire", State: models.SearchDone, Results: fireResults, Count: 3, Elapsed: 800 * time.Millisecond}
	hi = models.SearchEvent{
		Name: "HiAnime", State: models.SearchDone,
		Results: []*models.Anime{{Name: "[English] Namakura Gatana", Source: "HiAnime", Year: "1917", MediaType: models.MediaTypeAnime, Audio: "Sub"}},
		Count:   1, Elapsed: 1400 * time.Millisecond,
	}
	topcine = models.SearchEvent{Name: "TopCine", State: models.SearchFailed, Elapsed: 2100 * time.Millisecond, Reason: "temporarily unavailable (HTTP 503)"}
	return plan, fire, hi, topcine
}

func applyAll(m *searchingModel, events ...models.SearchEvent) {
	for _, ev := range events {
		m.applyEvent(ev)
	}
}

func TestSearchingAppliesPlan(t *testing.T) {
	t.Parallel()
	plan, _, _, _ := searchingTestEvents()
	m := newSearchingModel("namakura", nil)

	assert.False(t, m.allTerminal(), "no plan yet: the screen waits for the fan-out")
	assert.Equal(t, `"namakura"`, m.shell.Context)

	m.applyEvent(plan)
	assert.True(t, m.planned)
	assert.Equal(t, []string{"AnimeFire", "HiAnime", "TopCine"}, m.searched)
	assert.Equal(t, 0, m.terminalCount())
	assert.Equal(t, 3, m.pendingCount())
	assert.False(t, m.allTerminal())
}

func TestSearchingGroupsArriveLive(t *testing.T) {
	t.Parallel()
	plan, fire, hi, _ := searchingTestEvents()
	m := newSearchingModel("namakura", nil)
	applyAll(m, plan, fire)

	require.Len(t, m.groups, 3, "three works after AnimeFire answers")
	assert.Equal(t, "Namakura Gatana", m.groups[0].title)
	assert.Len(t, m.results.Items(), 3)

	applyAll(m, hi)
	require.Len(t, m.groups, 3, "HiAnime's copy joins the group, not the list")
	assert.Len(t, m.groups[0].versions, 2)
	assert.Equal(t, []string{"PT-BR", "EN"}, m.groups[0].badges())
}

func TestSearchingSortsClosestFirst(t *testing.T) {
	t.Parallel()
	m := newSearchingModel("x", nil)
	m.applyEvent(models.SearchEvent{Plan: true, Searched: []string{"HiAnime", "AnimeFire"}, GiveUp: time.Second})
	// English arrives first; the PT-BR group still sorts first.
	m.applyEvent(models.SearchEvent{Name: "HiAnime", State: models.SearchDone,
		Results: []*models.Anime{{Name: "[English] B"}}, Count: 1})
	m.applyEvent(models.SearchEvent{Name: "AnimeFire", State: models.SearchDone,
		Results: []*models.Anime{{Name: "[PT-BR] A"}}, Count: 1})
	require.Len(t, m.groups, 2)
	assert.Equal(t, "A", m.groups[0].title)
}

func TestSearchingEnterOpensFirstVersion(t *testing.T) {
	t.Parallel()
	plan, fire, hi, _ := searchingTestEvents()
	m := newSearchingModel("namakura", nil)
	applyAll(m, plan, fire, hi)

	updated, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	require.NotNil(t, cmd)
	got := updated.(*searchingModel).outcome
	require.NotNil(t, got.Chosen, "enter opens the highlighted title now")
	assert.Equal(t, "Animefire.io", got.Chosen.Source, "first version wins while searching")
	assert.False(t, got.Completed)
}

func TestSearchingEnterWithNothingIgnores(t *testing.T) {
	t.Parallel()
	m := newSearchingModel("namakura", nil)
	_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEnter})
	assert.Nil(t, cmd)
	assert.Nil(t, m.outcome.Chosen)
}

func TestSearchingKeys(t *testing.T) {
	t.Parallel()

	t.Run("esc goes back", func(t *testing.T) {
		t.Parallel()
		m := newSearchingModel("x", nil)
		_, cmd := m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		require.NotNil(t, cmd)
		assert.ErrorIs(t, m.err, ErrSelectionBack)
	})

	t.Run("ctrl+c cancels", func(t *testing.T) {
		t.Parallel()
		m := newSearchingModel("x", nil)
		_, cmd := m.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
		require.NotNil(t, cmd)
		assert.ErrorIs(t, m.err, ErrSelectionCancelled)
	})

	t.Run("q never quits here", func(t *testing.T) {
		t.Parallel()
		m := newSearchingModel("x", nil)
		_, cmd := m.Update(tea.KeyPressMsg{Code: 'q', Text: "q"})
		assert.Nil(t, cmd)
		assert.NoError(t, m.err)
	})

	t.Run("help toggles", func(t *testing.T) {
		t.Parallel()
		m := newSearchingModel("x", nil)
		_, _ = m.Update(tea.KeyPressMsg{Code: '?', Text: "?"})
		assert.True(t, m.showHelp)
		_, _ = m.Update(tea.KeyPressMsg{Code: tea.KeyEscape})
		assert.False(t, m.showHelp)
	})
}

func TestSearchingAutoCompletes(t *testing.T) {
	t.Parallel()
	plan, fire, hi, topcine := searchingTestEvents()
	m := newSearchingModel("namakura", nil)
	applyAll(m, plan, fire)
	assert.False(t, m.allTerminal())

	updated, cmd := m.Update(searchTickMsg{})
	require.NotNil(t, cmd, "still asking: keeps polling")
	_ = updated

	applyAll(m, hi, topcine)
	assert.True(t, m.allTerminal())
	updated, cmd = m.Update(searchTickMsg{})
	require.NotNil(t, cmd)
	assert.True(t, updated.(*searchingModel).outcome.Completed)
}

func TestSearchingStatusAndPanel(t *testing.T) {
	t.Parallel()
	plan, fire, _, topcine := searchingTestEvents()
	m := newSearchingModel("namakura", nil)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
	applyAll(m, plan, fire)
	m.started = time.Now().Add(-3200 * time.Millisecond)

	width, _ := m.shell.ContentSize()
	assert.Contains(t, ansi.Strip(m.statusLine()), "3 results so far, more arriving · enter opens one now")

	panel := ansi.Strip(m.panelRows(width - resultsListWidth(width)))
	assert.Contains(t, panel, "SOURCES")
	assert.Contains(t, panel, "AnimeFire")
	assert.Contains(t, panel, "3 found")
	assert.Contains(t, panel, "0.8s")
	assert.Contains(t, panel, "HiAnime")
	assert.Contains(t, panel, "asking...")
	assert.Contains(t, panel, "3.2s")
	assert.Contains(t, panel, "Goyabu")
	assert.Contains(t, panel, "off by default")
	assert.Contains(t, panel, "1 of 3 done")
	assert.Contains(t, panel, "gives up at 15s")

	applyAll(m, topcine)
	panel = ansi.Strip(m.panelRows(width - resultsListWidth(width)))
	assert.Contains(t, panel, "TopCine")
	assert.Contains(t, panel, "temporarily unavailable (HTTP 503)")
	assert.Contains(t, panel, "2 of 3 done")
}

func TestSearchingView(t *testing.T) {
	t.Parallel()
	plan, fire, hi, _ := searchingTestEvents()
	m := newSearchingModel("namakura", nil)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
	applyAll(m, plan, fire, hi)

	view := m.View()
	assert.True(t, view.AltScreen)
	assert.Equal(t, "GoAnime - Searching", view.WindowTitle)
	stripped := ansi.Strip(view.Content)
	assert.Contains(t, stripped, "SEARCHING")
	assert.Contains(t, stripped, `"namakura"`)
	assert.Contains(t, stripped, "Namakura Gatana")
	assert.Contains(t, stripped, "SOURCES")
	assert.Contains(t, stripped, "enter")
	assert.LessOrEqual(t, lipgloss.Width(view.Content), 120)
	assert.LessOrEqual(t, lipgloss.Height(view.Content), 31)
}

func TestSearchingEmptyShowsSkeletons(t *testing.T) {
	t.Parallel()
	plan, _, _, _ := searchingTestEvents()
	m := newSearchingModel("namakura", nil)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
	applyAll(m, plan)

	assert.Contains(t, ansi.Strip(m.statusLine()), "asking the sources")
	view := ansi.Strip(m.View().Content)
	assert.Contains(t, view, "SOURCES", "the panel tracks asking sources with no results yet")
}

func TestSearchingPlain(t *testing.T) {
	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })

	plan, fire, _, _ := searchingTestEvents()
	m := newSearchingModel("namakura", nil)
	_, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
	applyAll(m, plan, fire)

	assert.Contains(t, ansi.Strip(m.statusLine()), "more arriving - enter opens one now")
	width, height := m.shell.ContentSize()
	_ = height
	panel := ansi.Strip(m.panelRows(width - resultsListWidth(width)))
	assert.NotContains(t, panel, "ソース")
	assert.Contains(t, panel, "off by default")
}

func TestLiveLanguagePriority(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 0, liveLanguagePriority("[PT-BR] X"))
	assert.Equal(t, 2, liveLanguagePriority("[English] X"))
	assert.Equal(t, 4, liveLanguagePriority("X"))
}

func TestSearchLiveWithRunner(t *testing.T) {
	t.Parallel()
	events := make(chan models.SearchEvent, 1)

	t.Run("completed passthrough", func(t *testing.T) {
		t.Parallel()
		out, err := searchLiveWithRunner(SearchLiveOptions{Query: "x"}, events, func(model tea.Model) (tea.Model, error) {
			m := model.(*searchingModel)
			m.outcome.Completed = true
			return m, nil
		})
		require.NoError(t, err)
		assert.True(t, out.Completed)
		assert.Nil(t, out.Chosen)
	})

	t.Run("cancelled", func(t *testing.T) {
		t.Parallel()
		_, err := searchLiveWithRunner(SearchLiveOptions{}, events, func(model tea.Model) (tea.Model, error) {
			m := model.(*searchingModel)
			m.err = ErrSelectionCancelled
			return m, nil
		})
		assert.ErrorIs(t, err, ErrSelectionCancelled)
	})

	t.Run("nil runner", func(t *testing.T) {
		t.Parallel()
		_, err := searchLiveWithRunner(SearchLiveOptions{}, events, nil)
		assert.ErrorContains(t, err, "runner not configured")
	})

	t.Run("unexpected model", func(t *testing.T) {
		t.Parallel()
		_, err := searchLiveWithRunner(SearchLiveOptions{}, events, func(tea.Model) (tea.Model, error) {
			return minimalModel{}, nil
		})
		assert.ErrorContains(t, err, "unexpected searching model")
	})

	t.Run("runner error", func(t *testing.T) {
		t.Parallel()
		want := errors.New("no tty")
		_, err := searchLiveWithRunner(SearchLiveOptions{}, events, func(tea.Model) (tea.Model, error) {
			return nil, want
		})
		assert.ErrorIs(t, err, want)
	})
}
