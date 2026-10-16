package tui

import (
	"errors"
	"testing"

	tea "charm.land/bubbletea/v2"
	"charm.land/lipgloss/v2"
	"github.com/charmbracelet/x/ansi"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func subtitleTestOptions() SubtitleOptions {
	return SubtitleOptions{
		Title:   "Namakura Gatana",
		Context: "HiAnime · Sub",
		Tracks: []SubtitleTrack{
			{Label: "Portuguese (Brazil)", Language: "por"},
			{Label: "English", Language: "eng"},
		},
		LastLabel:      "Portuguese (Brazil)",
		SessionChecked: true,
		NextTitle:      "Namakura Gatana",
		NextLine:       "Episode 1 · 1080p",
	}
}

func subPress(code rune, text string) tea.KeyPressMsg {
	return tea.KeyPressMsg{Code: code, Text: text}
}

func TestNewSubtitlesModel(t *testing.T) {
	t.Parallel()

	model := newSubtitlesModel(subtitleTestOptions())

	// 2 tracks + Both + None + 2 checkboxes.
	require.Len(t, model.rows, 6)
	assert.Equal(t, subtitleRowTrack, model.rows[0].kind)
	assert.Equal(t, "1", model.rows[0].key)
	assert.Equal(t, "2", model.rows[1].key)
	assert.Equal(t, subtitleRowBoth, model.rows[2].kind)
	assert.Equal(t, "3", model.rows[2].key)
	assert.Equal(t, subtitleRowNone, model.rows[3].kind)
	assert.Equal(t, "0", model.rows[3].key)
	assert.True(t, model.rows[0].current, "last choice marked")
	assert.Equal(t, 0, model.cursor, "last choice preselected")
	assert.True(t, model.session, "session checked by default")
	assert.False(t, model.preferPT)
	assert.Equal(t, "Namakura Gatana > Subtitles", model.shell.Breadcrumb)
	assert.Equal(t, "HiAnime · Sub", model.shell.Context)
}

func TestSubtitlesSingleTrack(t *testing.T) {
	t.Parallel()
	opts := subtitleTestOptions()
	opts.Tracks = opts.Tracks[:1]
	model := newSubtitlesModel(opts)

	// 1 track + None + 2 checkboxes: no Both row.
	require.Len(t, model.rows, 4)
	assert.Equal(t, subtitleRowNone, model.rows[1].kind)
	assert.Equal(t, "1 0", model.digitRange())
}

func TestSubtitlesNoBothNote(t *testing.T) {
	t.Parallel()
	model := newSubtitlesModel(subtitleTestOptions())
	assert.Equal(t, "1-3 0", model.digitRange())
	assert.Equal(t, "This episode has 2 subtitle tracks", model.heading(model.trackCount()))
	assert.Equal(t, "This episode has 1 subtitle track", model.heading(1))
}

func TestSubtitlesEnter(t *testing.T) {
	t.Parallel()

	t.Run("highlighted track confirms", func(t *testing.T) {
		t.Parallel()
		model := newSubtitlesModel(subtitleTestOptions())
		updated, cmd := model.Update(subPress(tea.KeyEnter, ""))
		require.NotNil(t, cmd)
		m := updated.(*subtitlesModel)
		require.True(t, m.decided)
		assert.Equal(t, "track", m.result.Kind)
		assert.Equal(t, "por", m.result.Language)
		assert.True(t, m.result.Session, "checkbox state travels with the choice")
		assert.False(t, m.result.PreferPT)
	})

	t.Run("checkbox enter toggles instead", func(t *testing.T) {
		t.Parallel()
		model := newSubtitlesModel(subtitleTestOptions())
		for range 4 {
			_, _ = model.Update(subPress(tea.KeyDown, ""))
		}
		updated, cmd := model.Update(subPress(tea.KeyEnter, ""))
		assert.Nil(t, cmd, "stays open")
		m := updated.(*subtitlesModel)
		assert.False(t, m.decided)
		assert.False(t, m.session)
	})
}

func TestSubtitlesDigits(t *testing.T) {
	t.Parallel()

	t.Run("digit picks directly", func(t *testing.T) {
		t.Parallel()
		model := newSubtitlesModel(subtitleTestOptions())
		updated, cmd := model.Update(subPress('2', "2"))
		require.NotNil(t, cmd)
		m := updated.(*subtitlesModel)
		assert.Equal(t, "track", m.result.Kind)
		assert.Equal(t, "eng", m.result.Language)
	})

	t.Run("zero picks none", func(t *testing.T) {
		t.Parallel()
		model := newSubtitlesModel(subtitleTestOptions())
		updated, cmd := model.Update(subPress('0', "0"))
		require.NotNil(t, cmd)
		assert.Equal(t, "none", updated.(*subtitlesModel).result.Kind)
	})

	t.Run("both keeps all", func(t *testing.T) {
		t.Parallel()
		model := newSubtitlesModel(subtitleTestOptions())
		updated, cmd := model.Update(subPress('3', "3"))
		require.NotNil(t, cmd)
		m := updated.(*subtitlesModel)
		assert.Equal(t, "all", m.result.Kind)
		assert.True(t, m.result.Session)
	})
}

func TestSubtitlesSpaceToggles(t *testing.T) {
	t.Parallel()

	model := newSubtitlesModel(subtitleTestOptions())
	for range 4 {
		_, _ = model.Update(subPress(tea.KeyDown, ""))
	}
	_, _ = model.Update(subPress(' ', " "))
	assert.False(t, model.session, "space flips the focused checkbox")

	_, _ = model.Update(subPress(tea.KeyDown, ""))
	_, _ = model.Update(subPress(' ', " "))
	assert.True(t, model.preferPT)
}

func TestSubtitlesSpaceOnTrackNoops(t *testing.T) {
	t.Parallel()
	model := newSubtitlesModel(subtitleTestOptions())
	updated, cmd := model.Update(subPress(' ', " "))
	assert.Nil(t, cmd)
	assert.False(t, updated.(*subtitlesModel).decided)
	assert.True(t, model.session, "session untouched")
}

func TestSubtitlesNavigation(t *testing.T) {
	t.Parallel()

	t.Run("cursor clamps", func(t *testing.T) {
		t.Parallel()
		model := newSubtitlesModel(subtitleTestOptions())
		_, _ = model.Update(subPress(tea.KeyUp, ""))
		assert.Equal(t, 0, model.cursor)
		for range 20 {
			_, _ = model.Update(subPress(tea.KeyDown, ""))
		}
		assert.Equal(t, 5, model.cursor)
	})

	t.Run("esc backs out", func(t *testing.T) {
		t.Parallel()
		model := newSubtitlesModel(subtitleTestOptions())
		_, cmd := model.Update(subPress(tea.KeyEscape, ""))
		require.NotNil(t, cmd)
		assert.ErrorIs(t, model.err, ErrPickBack)
	})

	t.Run("ctrl+c cancels", func(t *testing.T) {
		t.Parallel()
		model := newSubtitlesModel(subtitleTestOptions())
		_, cmd := model.Update(tea.KeyPressMsg{Code: 'c', Mod: tea.ModCtrl})
		require.NotNil(t, cmd)
		assert.ErrorIs(t, model.err, ErrPickCancelled)
	})

	t.Run("help toggles", func(t *testing.T) {
		t.Parallel()
		model := newSubtitlesModel(subtitleTestOptions())
		_, _ = model.Update(subPress('?', "?"))
		assert.True(t, model.showHelp)
		_, _ = model.Update(subPress(tea.KeyEscape, ""))
		assert.False(t, model.showHelp)
	})
}

func TestSubtitlesView(t *testing.T) {
	t.Parallel()

	model := newSubtitlesModel(subtitleTestOptions())
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
	view := model.View()
	assert.True(t, view.AltScreen)
	assert.Equal(t, "GoAnime - Subtitles", view.WindowTitle)
	stripped := ansi.Strip(view.Content)

	assert.Contains(t, stripped, "GOANIME")
	assert.Contains(t, stripped, "SUBTITLES")
	assert.Contains(t, stripped, "Namakura Gatana")
	assert.Contains(t, stripped, "HiAnime")
	assert.Contains(t, stripped, "This episode has 2 subtitle tracks")
	assert.Contains(t, stripped, "Portuguese (Brazil)")
	assert.Contains(t, stripped, "your choice last time")
	assert.Contains(t, stripped, "Both, switch in mpv with j")
	assert.Contains(t, stripped, "None")
	assert.Contains(t, stripped, "[x] Use this choice for the rest of the session")
	assert.Contains(t, stripped, "[ ] Always pick Portuguese when a source has it")
	assert.Contains(t, stripped, "UP NEXT")
	assert.Contains(t, stripped, "Episode 1 · 1080p")
	assert.Contains(t, stripped, "Asked once")
	assert.Contains(t, stripped, "in the player menu changes it")
	assert.LessOrEqual(t, lipgloss.Width(view.Content), 120)
	assert.LessOrEqual(t, lipgloss.Height(view.Content), 31)
}

func TestSubtitlesViewPlain(t *testing.T) {
	SetPlain(true)
	t.Cleanup(func() { SetPlain(false) })

	model := newSubtitlesModel(subtitleTestOptions())
	_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
	stripped := ansi.Strip(model.View().Content)

	assert.Contains(t, stripped, "SUBTITLES")
	assert.NotContains(t, stripped, "字幕", "plain hides Japanese glosses")
	assert.NotContains(t, stripped, "次へ")
	assert.Contains(t, stripped, "[x] Use this choice")
}

func TestSelectSubtitleWithRunner(t *testing.T) {
	t.Parallel()

	t.Run("choice passthrough", func(t *testing.T) {
		t.Parallel()
		res, err := selectSubtitleWithRunner(subtitleTestOptions(), func(model tea.Model) (tea.Model, error) {
			m := model.(*subtitlesModel)
			m.result = SubtitleChoice{Kind: "track", Language: "por", Session: true}
			m.decided = true
			return m, nil
		})
		require.NoError(t, err)
		assert.Equal(t, "track", res.Kind)
		assert.Equal(t, "por", res.Language)
		assert.True(t, res.Session)
	})

	t.Run("undecided cancels", func(t *testing.T) {
		t.Parallel()
		_, err := selectSubtitleWithRunner(subtitleTestOptions(), func(model tea.Model) (tea.Model, error) {
			return model, nil
		})
		assert.ErrorIs(t, err, ErrPickCancelled)
	})

	t.Run("model error", func(t *testing.T) {
		t.Parallel()
		_, err := selectSubtitleWithRunner(subtitleTestOptions(), func(model tea.Model) (tea.Model, error) {
			m := model.(*subtitlesModel)
			m.err = ErrPickBack
			return m, nil
		})
		assert.ErrorIs(t, err, ErrPickBack)
	})

	t.Run("nil runner", func(t *testing.T) {
		t.Parallel()
		_, err := selectSubtitleWithRunner(subtitleTestOptions(), nil)
		assert.ErrorContains(t, err, "runner not configured")
	})

	t.Run("unexpected model", func(t *testing.T) {
		t.Parallel()
		_, err := selectSubtitleWithRunner(subtitleTestOptions(), func(tea.Model) (tea.Model, error) {
			return minimalModel{}, nil
		})
		assert.ErrorContains(t, err, "unexpected subtitles model")
	})

	t.Run("runner error", func(t *testing.T) {
		t.Parallel()
		want := errors.New("no tty")
		_, err := selectSubtitleWithRunner(subtitleTestOptions(), func(tea.Model) (tea.Model, error) {
			return nil, want
		})
		assert.ErrorIs(t, err, want)
	})
}

func TestSubtitlesExtraRows(t *testing.T) {
	t.Parallel()

	t.Run("empty labels fall back", func(t *testing.T) {
		t.Parallel()
		opts := subtitleTestOptions()
		opts.Tracks = []SubtitleTrack{{Label: "", Language: "por"}, {Label: "", Language: ""}}
		opts.LastLabel = "por"
		model := newSubtitlesModel(opts)

		assert.Equal(t, "por", model.rows[0].label, "language fills a blank label")
		assert.Equal(t, "Subtitle 2", model.rows[1].label)
		assert.True(t, model.rows[0].current, "last choice matches the fallback label")
		assert.Equal(t, 0, model.cursor)
	})

	t.Run("unmatched last label preselects nothing", func(t *testing.T) {
		t.Parallel()
		opts := subtitleTestOptions()
		opts.LastLabel = "Klingon"
		model := newSubtitlesModel(opts)

		for _, row := range model.rows {
			assert.False(t, row.current)
		}
		assert.Equal(t, 0, model.cursor)
	})

	t.Run("many tracks cap digit keys", func(t *testing.T) {
		t.Parallel()
		var tracks []SubtitleTrack
		for range 12 {
			tracks = append(tracks, SubtitleTrack{Label: "T", Language: "l"})
		}
		model := newSubtitlesModel(SubtitleOptions{Tracks: tracks})

		assert.Equal(t, "9", model.rows[8].key)
		assert.Equal(t, "", model.rows[9].key, "past 9 there is no digit")
		assert.Equal(t, "", model.rows[10].key)
		_, cmd := model.Update(subPress('0', "0"))
		assert.NotNil(t, cmd, "None keeps 0")
	})

	t.Run("letters are ignored", func(t *testing.T) {
		t.Parallel()
		model := newSubtitlesModel(subtitleTestOptions())
		updated, cmd := model.Update(subPress('a', "a"))
		assert.Nil(t, cmd)
		assert.False(t, updated.(*subtitlesModel).decided)
	})

	t.Run("help view shows keys", func(t *testing.T) {
		t.Parallel()
		model := newSubtitlesModel(subtitleTestOptions())
		_, _ = model.Update(subPress('?', "?"))
		_, _ = model.Update(tea.WindowSizeMsg{Width: 120, Height: 31})
		view := ansi.Strip(model.View().Content)
		assert.Contains(t, view, "KEYS")
		assert.Contains(t, view, "SUBTITLES")
	})

	t.Run("narrow panel drops below", func(t *testing.T) {
		t.Parallel()
		model := newSubtitlesModel(subtitleTestOptions())
		_, _ = model.Update(tea.WindowSizeMsg{Width: 80, Height: 31})
		view := ansi.Strip(model.View().Content)
		assert.Contains(t, view, "UP NEXT")
		assert.Contains(t, view, "Portuguese (Brazil)")
	})

	t.Run("out-of-range confirm never quits", func(t *testing.T) {
		t.Parallel()
		model := newSubtitlesModel(subtitleTestOptions())
		model.cursor = 99
		_, cmd := model.Update(subPress(tea.KeyEnter, ""))
		assert.Nil(t, cmd)
		assert.False(t, model.decided)
	})
}
