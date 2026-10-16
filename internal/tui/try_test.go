package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func tryTestOptions() TryOptions {
	return TryOptions{
		Query:        "namakra gatana",
		CurrentScope: "",
		SourcesLine:  "6 sources on · Goyabu off",
		Summary:      `Nothing for "namakra gatana". 5 sources answered without a match, 1 could not be reached.`,
		Sources: []TrySource{
			{Name: "AnimeFire", Ok: true},
			{Name: "HiAnime", Ok: true},
			{Name: "TopCine", Detail: "temporarily unavailable (HTTP 503)"},
		},
		Off: []models.SearchOffEntry{{Name: "Goyabu", Reason: "off by default"}},
		Suggestions: []TrySuggestion{
			{Label: "Namakura Gatana", Hint: "did you mean · AniList", Query: "Namakura Gatana", Scope: ""},
			{Label: "gatana", Hint: "fewer words", Query: "gatana", Scope: ""},
			{Label: "namakra gatana", Hint: "English sources only", Query: "namakra gatana", Scope: "hianime"},
		},
		Note: "Titles often differ between sources: romaji, English, and the Portuguese name on the PT-BR sites.",
	}
}

func tryPress(code rune, text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: text}
}

func TestNewTryModel(t *testing.T) {
	t.Parallel()

	model := newTryModelWithOptions(tryTestOptions())

	assert.True(t, model.focusBox)
	assert.Equal(t, "namakra gatana", model.input.Value())
	assert.Equal(t, 0, model.scopeIdx)
	assert.Equal(t, "6 sources on · Goyabu off", model.shell.Context)
	assert.False(t, model.replaced)
}

func TestTryEnter(t *testing.T) {
	t.Parallel()

	t.Run("box searches its text", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		updated, cmd := model.Update(tryPress(tea.KeyEnter, ""))
		require.NotNil(t, cmd)
		m := updated.(*tryModel)
		require.True(t, m.decided)
		assert.Equal(t, "namakra gatana", m.result.Query)
		assert.Equal(t, "", m.result.Scope)
	})

	t.Run("short box stays", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		model.input.SetValue("x")
		model.replaced = true
		_, cmd := model.Update(tryPress(tea.KeyEnter, ""))
		assert.Nil(t, cmd)
		assert.False(t, model.decided)
	})

	t.Run("suggestion searches it", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		_, _ = model.Update(tryPress(tea.KeyDown, ""))
		_, _ = model.Update(tryPress(tea.KeyDown, ""))
		updated, cmd := model.Update(tryPress(tea.KeyEnter, ""))
		require.NotNil(t, cmd)
		m := updated.(*tryModel)
		assert.Equal(t, "gatana", m.result.Query)
		assert.Equal(t, "", m.result.Scope)
	})

	t.Run("english suggestion scopes", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		for range 4 {
			_, _ = model.Update(tryPress(tea.KeyDown, ""))
		}
		updated, cmd := model.Update(tryPress(tea.KeyEnter, ""))
		require.NotNil(t, cmd)
		m := updated.(*tryModel)
		assert.Equal(t, "namakra gatana", m.result.Query)
		assert.Equal(t, "hianime", m.result.Scope)
	})
}

func TestTryNavigation(t *testing.T) {
	t.Parallel()

	t.Run("down enters suggestions, up returns", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		_, _ = model.Update(tryPress(tea.KeyDown, ""))
		assert.False(t, model.focusBox)
		assert.Equal(t, 0, model.cursor)
		_, _ = model.Update(tryPress(tea.KeyUp, ""))
		assert.True(t, model.focusBox)
	})

	t.Run("cursor clamps at the ends", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		for range 10 {
			_, _ = model.Update(tryPress(tea.KeyDown, ""))
		}
		assert.Equal(t, 2, model.cursor)
	})

	t.Run("tab cycles scopes", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		for i, want := range []int{1, 2, 3, 0} {
			_, _ = model.Update(tryPress(tea.KeyTab, ""))
			assert.Equal(t, want, model.scopeIdx, "tab %d", i)
		}
	})

	t.Run("box enter uses the tabbed scope", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		_, _ = model.Update(tryPress(tea.KeyTab, ""))
		_, _ = model.Update(tryPress(tea.KeyTab, ""))
		updated, cmd := model.Update(tryPress(tea.KeyEnter, ""))
		require.NotNil(t, cmd)
		assert.Equal(t, "hianime", updated.(*tryModel).result.Scope)
	})

	t.Run("esc clears the box, never quits", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		_, cmd := model.Update(tryPress(tea.KeyEscape, ""))
		assert.Nil(t, cmd)
		assert.NoError(t, model.err)
		assert.Equal(t, "", model.input.Value())

		_, _ = model.Update(tryPress(tea.KeyDown, ""))
		_, cmd = model.Update(tryPress(tea.KeyEscape, ""))
		assert.Nil(t, cmd, "over suggestions esc returns to the box")
		assert.True(t, model.focusBox)
		assert.NoError(t, model.err)
	})
}

func TestTryTypingReplacesKeptQuery(t *testing.T) {
	t.Parallel()

	model := newTryModelWithOptions(tryTestOptions())
	_, _ = model.Update(tryPress(tea.KeyDown, ""))
	require.False(t, model.focusBox)

	_, _ = model.Update(tryPress('g', "g"))
	assert.True(t, model.focusBox)
	assert.Equal(t, "g", model.input.Value(), "first keystroke replaces the kept query")

	_, _ = model.Update(tryPress('a', "a"))
	assert.Equal(t, "ga", model.input.Value(), "then it edits normally")
}

func TestTryHelp(t *testing.T) {
	t.Parallel()

	model := newTryModelWithOptions(tryTestOptions())
	_, _ = model.Update(tryPress(tea.KeyDown, ""))
	_, _ = model.Update(tryPress('?', "?"))
	assert.True(t, model.showHelp)
	_, _ = model.Update(tryPress(tea.KeyEscape, ""))
	assert.False(t, model.showHelp)
}

func TestTryView(t *testing.T) {
	t.Parallel()

	model := newTryModelWithOptions(tryTestOptions())
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
	view := ansi.Strip(model.View().Content)

	assert.Contains(t, view, "GOANIME")
	assert.Contains(t, view, "SEARCH")
	assert.Contains(t, view, "namakra gatana")
	assert.Contains(t, view, "kept from the last search")
	assert.Contains(t, view, "Nothing for")
	assert.Contains(t, view, "TRY")
	assert.Contains(t, view, "Namakura Gatana")
	assert.Contains(t, view, "did you mean")
	assert.Contains(t, view, "fewer words")
	assert.Contains(t, view, "English sources only")
	assert.Contains(t, view, "Titles often differ")
	assert.Contains(t, view, "SOURCES")
	assert.Contains(t, view, "AnimeFire")
	assert.Contains(t, view, "TopCine")
	assert.Contains(t, view, "temporarily unavailable (HTTP 503)")
	assert.Contains(t, view, "Goyabu")
	assert.Contains(t, view, "enter")
	assert.Contains(t, view, "suggestions")
	assert.True(t, model.View().AltScreen)
	assert.LessOrEqual(t, lipgloss.Width(model.View().Content), 120)
	assert.LessOrEqual(t, lipgloss.Height(model.View().Content), 31)
}

func TestTryViewPlain(t *testing.T) {
	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })

	model := newTryModelWithOptions(tryTestOptions())
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
	view := ansi.Strip(model.View().Content)

	assert.Contains(t, view, "TRY")
	assert.NotContains(t, view, "試す", "plain hides Japanese glosses")
	assert.NotContains(t, view, "ソース")
	assert.Contains(t, view, "SOURCES")
}

func TestTryEmptySuggestions(t *testing.T) {
	t.Parallel()

	opts := tryTestOptions()
	opts.Suggestions = nil
	model := newTryModelWithOptions(opts)
	_, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 24})

	_, _ = model.Update(tryPress(tea.KeyDown, ""))
	assert.True(t, model.focusBox, "nowhere to go without suggestions")

	_, cmd := model.Update(tryPress(tea.KeyEnter, ""))
	assert.NotNil(t, cmd, "the box still searches")
}

func TestTryCtrlCCancels(t *testing.T) {
	t.Parallel()
	model := newTryModelWithOptions(tryTestOptions())
	_, cmd := model.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
	require.NotNil(t, cmd)
	assert.ErrorIs(t, model.err, ErrSelectionCancelled)
}

func TestTryWithRunner(t *testing.T) {
	t.Parallel()

	t.Run("decision passthrough", func(t *testing.T) {
		t.Parallel()
		res, err := trySearchWithRunner(tryTestOptions(), func(model tea.Model) (tea.Model, error) {
			m := model.(*tryModel)
			m.result = TryDecision{Query: "gatana", Scope: ""}
			m.decided = true
			return m, nil
		})
		require.NoError(t, err)
		assert.Equal(t, "gatana", res.Query)
	})

	t.Run("undecided is a cancel", func(t *testing.T) {
		t.Parallel()
		_, err := trySearchWithRunner(tryTestOptions(), func(model tea.Model) (tea.Model, error) {
			return model, nil
		})
		assert.ErrorIs(t, err, ErrSelectionCancelled)
	})

	t.Run("cancelled", func(t *testing.T) {
		t.Parallel()
		_, err := trySearchWithRunner(tryTestOptions(), func(model tea.Model) (tea.Model, error) {
			m := model.(*tryModel)
			m.err = ErrSelectionCancelled
			return m, nil
		})
		assert.ErrorIs(t, err, ErrSelectionCancelled)
	})

	t.Run("nil runner", func(t *testing.T) {
		t.Parallel()
		_, err := trySearchWithRunner(tryTestOptions(), nil)
		assert.ErrorContains(t, err, "runner not configured")
	})

	t.Run("unexpected model", func(t *testing.T) {
		t.Parallel()
		_, err := trySearchWithRunner(tryTestOptions(), func(tea.Model) (tea.Model, error) {
			return minimalModel{}, nil
		})
		assert.ErrorContains(t, err, "unexpected no-results model")
	})

	t.Run("runner error", func(t *testing.T) {
		t.Parallel()
		want := errors.New("no tty")
		_, err := trySearchWithRunner(tryTestOptions(), func(tea.Model) (tea.Model, error) {
			return nil, want
		})
		assert.ErrorIs(t, err, want)
	})
}

func TestIndexForScope(t *testing.T) {
	t.Parallel()
	assert.Equal(t, 0, indexForScope(""))
	assert.Equal(t, 1, indexForScope("ptbr"))
	assert.Equal(t, 1, indexForScope("animefire"))
	assert.Equal(t, 2, indexForScope("hianime"))
	assert.Equal(t, 3, indexForScope("movietv"))
	assert.Equal(t, 3, indexForScope("topcine"))
	assert.Equal(t, 0, indexForScope("weird"))
}

func TestTryBoxHint(t *testing.T) {
	t.Parallel()

	t.Run("hint shows until replaced", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
		assert.Contains(t, ansi.Strip(model.tryBox(120)), "kept from the last search")

		_, _ = model.Update(tryPress('x', "x"))
		assert.NotContains(t, ansi.Strip(model.tryBox(120)), "kept from the last search")
	})

	t.Run("box typing edits normally when focused", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		_, _ = model.Update(tryPress('!', "!"))
		assert.Equal(t, "namakra gatana!", model.input.Value(), "focused box appends, never replaces")
	})
}

func TestTryTabsVisibility(t *testing.T) {
	t.Parallel()

	t.Run("hidden on ALL", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
		assert.NotContains(t, ansi.Strip(model.tryBox(120)), "MOVIES/TV")
	})

	t.Run("shown once tabbed", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
		_, _ = model.Update(tryPress(tea.KeyTab, ""))
		box := ansi.Strip(model.tryBox(120))
		assert.Contains(t, box, "MOVIES/TV", "tabbed scope reveals the tabs row")
		assert.Contains(t, box, "PT-BR")
	})

	t.Run("footer names the tabbed scope", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		_, _ = model.Update(tryPress(tea.KeyTab, ""))
		_, _ = model.Update(tryPress(tea.KeyTab, ""))
		hints := model.hints()
		found := false
		for _, h := range hints {
			if h.Key == "tab" {
				found = true
				assert.Equal(t, "sources: ENGLISH", h.Desc)
			}
		}
		assert.True(t, found)
	})
}

func TestTryResizeSizesInput(t *testing.T) {
	t.Parallel()
	model := newTryModelWithOptions(tryTestOptions())
	_, _ = model.Update(tea.WindowSizeMsg{Width: 100, Height: 30})
	assert.Equal(t, 100, model.shell.Width)
	assert.Greater(t, model.input.Width(), 50)
}

func TestTryHelpView(t *testing.T) {
	t.Parallel()

	m := newTryModelWithOptions(tryTestOptions())
	_, _ = m.Update(tryPress(tea.KeyDown, ""))
	_, _ = m.Update(tryPress('?', "?"))
	_, _ = m.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
	view := ansi.Strip(m.View().Content)
	assert.Contains(t, view, "KEYS")
	assert.Contains(t, view, "NO RESULTS")
	assert.Contains(t, view, "esc never quits")
}

func TestTryConfirmGuards(t *testing.T) {
	t.Parallel()

	t.Run("blank suggestion never confirms", func(t *testing.T) {
		t.Parallel()
		opts := tryTestOptions()
		opts.Suggestions = []TrySuggestion{{Label: "x", Hint: "h", Query: "  "}}
		model := newTryModelWithOptions(opts)
		_, _ = model.Update(tryPress(tea.KeyDown, ""))
		_, cmd := model.Update(tryPress(tea.KeyEnter, ""))
		assert.Nil(t, cmd)
		assert.False(t, model.decided)
	})

	t.Run("cursor beyond rows never confirms", func(t *testing.T) {
		t.Parallel()
		model := newTryModelWithOptions(tryTestOptions())
		model.focusBox = false
		model.cursor = 99
		_, cmd := model.Update(tryPress(tea.KeyEnter, ""))
		assert.Nil(t, cmd)
		assert.False(t, model.decided)
	})
}

func TestTryPanelEmpty(t *testing.T) {
	t.Parallel()
	model := newTryModelWithOptions(TryOptions{Query: "x", Summary: "Nothing."})
	assert.Equal(t, "", model.tryPanel(60))
}

func TestTryPlainFooter(t *testing.T) {
	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })

	model := newTryModelWithOptions(tryTestOptions())
	for _, h := range model.hints() {
		assert.NotContains(t, h.Key, "↓")
		assert.NotContains(t, h.Desc, "↓")
	}
}
