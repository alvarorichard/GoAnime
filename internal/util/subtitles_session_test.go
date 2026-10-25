package util

import (
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func twoTracks() []SubtitleInfo {
	return []SubtitleInfo{
		{URL: "https://cdn.example.com/en.vtt", Language: "eng", Label: "English"},
		{URL: "https://cdn.example.com/pt.vtt", Language: "por", Label: "Portuguese"},
	}
}

func TestIsPortugueseSubtitle(t *testing.T) {
	t.Parallel()
	tests := []struct {
		sub  SubtitleInfo
		want bool
	}{
		{SubtitleInfo{Language: "por"}, true},
		{SubtitleInfo{Language: "pt"}, true},
		{SubtitleInfo{Language: "pt-BR", Label: "Portuguese"}, true},
		{SubtitleInfo{Label: "Portuguese (Brazil)"}, true},
		{SubtitleInfo{Label: "Portuguese - Brazilian Portuguese"}, true},
		{SubtitleInfo{Label: "Legendado"}, true},
		{SubtitleInfo{Label: "Dublado"}, true},
		{SubtitleInfo{Language: "eng", Label: "English"}, false},
		{SubtitleInfo{Language: "sub"}, false},
		{SubtitleInfo{Language: "dub"}, false},
		{SubtitleInfo{Language: "spa", Label: "Spanish"}, false},
		{SubtitleInfo{}, false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isPortugueseSubtitle(tt.sub), "%+v", tt.sub)
	}
}

func TestMatchSubtitleLanguage(t *testing.T) {
	t.Parallel()
	tracks := twoTracks()
	assert.Equal(t, 1, matchSubtitleLanguage(tracks, "por"))
	assert.Equal(t, 1, matchSubtitleLanguage(tracks, "POR"))
	assert.Equal(t, 0, matchSubtitleLanguage(tracks, "eng"))
	assert.Equal(t, -1, matchSubtitleLanguage(tracks, "fre"))
	assert.Equal(t, -1, matchSubtitleLanguage(tracks, ""))
	assert.Equal(t, 0, matchSubtitleLanguage([]SubtitleInfo{{Label: "English"}}, "English"), "label fallback")
}

func TestApplySubtitleSession(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	t.Run("unset asks", func(t *testing.T) {
		resetSubtitleState()
		SetGlobalSubtitles(twoTracks())
		assert.False(t, applySubtitleSession())
		assert.Len(t, GlobalSubtitles, 2)
	})

	t.Run("recorded language applies", func(t *testing.T) {
		resetSubtitleState()
		SetGlobalSubtitles(twoTracks())
		subtitleSessionChoice = subtitleSession{set: true, language: "por"}
		assert.True(t, applySubtitleSession())
		require.Len(t, GlobalSubtitles, 1)
		assert.Equal(t, "Portuguese", GlobalSubtitles[0].Label)
	})

	t.Run("recorded all keeps", func(t *testing.T) {
		resetSubtitleState()
		SetGlobalSubtitles(twoTracks())
		subtitleSessionChoice = subtitleSession{set: true, all: true}
		assert.True(t, applySubtitleSession())
		assert.Len(t, GlobalSubtitles, 2)
	})

	t.Run("recorded none clears", func(t *testing.T) {
		resetSubtitleState()
		SetGlobalSubtitles(twoTracks())
		subtitleSessionChoice = subtitleSession{set: true, none: true}
		assert.True(t, applySubtitleSession())
		assert.Empty(t, GlobalSubtitles)
	})

	t.Run("gone language asks again", func(t *testing.T) {
		resetSubtitleState()
		SetGlobalSubtitles([]SubtitleInfo{{URL: "u", Language: "fre", Label: "French"}})
		subtitleSessionChoice = subtitleSession{set: true, language: "por"}
		assert.False(t, applySubtitleSession())
	})

	t.Run("portuguese rule wins", func(t *testing.T) {
		resetSubtitleState()
		SetGlobalSubtitles(twoTracks())
		subtitleAlwaysPT = true
		assert.True(t, applySubtitleSession())
		require.Len(t, GlobalSubtitles, 1)
		assert.Equal(t, "por", GlobalSubtitles[0].Language)
	})

	t.Run("portuguese rule without portuguese falls through", func(t *testing.T) {
		resetSubtitleState()
		SetGlobalSubtitles([]SubtitleInfo{{URL: "u", Language: "eng", Label: "English"}})
		subtitleAlwaysPT = true
		assert.False(t, applySubtitleSession())
	})
}

func TestSelectSubtitlesSession(t *testing.T) {
	t.Cleanup(resetSubtitleState)
	prev := subtitleScreenFn
	t.Cleanup(func() { subtitleScreenFn = prev })

	t.Run("screen choice filters and remembers", func(t *testing.T) {
		resetSubtitleState()
		SetGlobalSubtitles(twoTracks())
		var seenOpts tui.SubtitleOptions
		subtitleScreenFn = func(opts tui.SubtitleOptions) (tui.SubtitleChoice, error) {
			seenOpts = opts
			return tui.SubtitleChoice{Kind: "track", Language: "por", Session: true}, nil
		}

		SelectSubtitles()

		require.Len(t, GlobalSubtitles, 1)
		assert.Equal(t, "por", GlobalSubtitles[0].Language)
		assert.Equal(t, 2, len(seenOpts.Tracks))
		assert.Equal(t, "Portuguese", subtitleLastLabel)

		// Next episode with fresh tracks: no screen, same language.
		SetGlobalSubtitles([]SubtitleInfo{
			{URL: "e2-en", Language: "eng", Label: "English"},
			{URL: "e2-pt", Language: "por", Label: "Portuguese"},
		})
		SelectSubtitles()
		require.Len(t, GlobalSubtitles, 1)
		assert.Equal(t, "e2-pt", GlobalSubtitles[0].URL, "asked once")
	})

	t.Run("unchecked session asks every time", func(t *testing.T) {
		resetSubtitleState()
		SetGlobalSubtitles(twoTracks())
		calls := 0
		subtitleScreenFn = func(tui.SubtitleOptions) (tui.SubtitleChoice, error) {
			calls++
			return tui.SubtitleChoice{Kind: "all", Session: false}, nil
		}

		SelectSubtitles()
		SelectSubtitles()
		assert.Equal(t, 2, calls)
		assert.Len(t, GlobalSubtitles, 2)
	})

	t.Run("unchecking clears a recorded session", func(t *testing.T) {
		resetSubtitleState()
		SetSubtitleContext("Show A", "", "", "")
		SetGlobalSubtitles(twoTracks())
		subtitleScreenFn = func(tui.SubtitleOptions) (tui.SubtitleChoice, error) {
			return tui.SubtitleChoice{Kind: "track", Language: "por", Session: true}, nil
		}
		SelectSubtitles()
		require.Len(t, GlobalSubtitles, 1)

		// Another show asks again; answering unchecked clears everything.
		SetSubtitleContext("Show B", "", "", "")
		SetGlobalSubtitles(twoTracks())
		subtitleScreenFn = func(tui.SubtitleOptions) (tui.SubtitleChoice, error) {
			return tui.SubtitleChoice{Kind: "track", Language: "eng", Session: false}, nil
		}
		SelectSubtitles()
		require.Len(t, GlobalSubtitles, 1)
		assert.Equal(t, "eng", GlobalSubtitles[0].Language)

		// Back on show A: nothing remembered, asks again.
		asked := false
		SetSubtitleContext("Show A", "", "", "")
		SetGlobalSubtitles(twoTracks())
		subtitleScreenFn = func(tui.SubtitleOptions) (tui.SubtitleChoice, error) {
			asked = true
			return tui.SubtitleChoice{Kind: "all", Session: false}, nil
		}
		SelectSubtitles()
		assert.True(t, asked, "unchecking must not remember")
	})

	t.Run("cancel keeps tracks", func(t *testing.T) {
		resetSubtitleState()
		SetGlobalSubtitles(twoTracks())
		subtitleScreenFn = func(tui.SubtitleOptions) (tui.SubtitleChoice, error) {
			return tui.SubtitleChoice{}, tui.ErrPickBack
		}

		SelectSubtitles()
		assert.Len(t, GlobalSubtitles, 2)
	})
}

func TestSubtitleSessionBoundToShow(t *testing.T) {
	t.Cleanup(resetSubtitleState)
	prev := subtitleScreenFn
	t.Cleanup(func() { subtitleScreenFn = prev })

	resetSubtitleState()
	SetSubtitleContext("Show A", "HiAnime", "", "")
	SetGlobalSubtitles(twoTracks())
	calls := 0
	subtitleScreenFn = func(tui.SubtitleOptions) (tui.SubtitleChoice, error) {
		calls++
		return tui.SubtitleChoice{Kind: "track", Language: "por", Session: true}, nil
	}
	SelectSubtitles()
	require.Len(t, GlobalSubtitles, 1)
	require.Equal(t, 1, calls)

	// Same show, fresh tracks: remembered, no screen.
	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "a2-en", Language: "eng", Label: "English"},
		{URL: "a2-pt", Language: "por", Label: "Portuguese"},
	})
	SelectSubtitles()
	assert.Equal(t, 1, calls, "same show reuses the remembered language")
	require.Len(t, GlobalSubtitles, 1)
	assert.Equal(t, "a2-pt", GlobalSubtitles[0].URL)

	// Another show: the remembered language must NOT leak across titles.
	SetSubtitleContext("Show B", "HiAnime", "", "")
	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "b-en", Language: "eng", Label: "English"},
		{URL: "b-por", Language: "por", Label: "Portuguese"},
	})
	SelectSubtitles()
	assert.Equal(t, 2, calls, "a new show asks again")
}

func TestPromptSubtitleLanguageSession(t *testing.T) {
	t.Cleanup(resetSubtitleState)
	prev := subtitleScreenFn
	t.Cleanup(func() { subtitleScreenFn = prev })

	t.Run("prefer portuguese skips the screen", func(t *testing.T) {
		resetSubtitleState()
		SetGlobalSubtitles(twoTracks())
		called := false
		subtitleScreenFn = func(tui.SubtitleOptions) (tui.SubtitleChoice, error) {
			called = true
			return tui.SubtitleChoice{Kind: "track", Language: "eng"}, nil
		}
		subtitleAlwaysPT = true

		PromptSubtitleLanguage()
		assert.False(t, called)
		require.Len(t, GlobalSubtitles, 1)
		assert.Equal(t, "por", GlobalSubtitles[0].Language)
	})

	t.Run("single track still asks", func(t *testing.T) {
		resetSubtitleState()
		SetGlobalSubtitles([]SubtitleInfo{{URL: "u", Language: "eng", Label: "English"}})
		var gotTracks int
		subtitleScreenFn = func(opts tui.SubtitleOptions) (tui.SubtitleChoice, error) {
			gotTracks = len(opts.Tracks)
			return tui.SubtitleChoice{Kind: "track", Language: "eng", Session: true}, nil
		}

		PromptSubtitleLanguage()
		assert.Equal(t, 1, gotTracks)
		require.Len(t, GlobalSubtitles, 1)
	})

	t.Run("none disables", func(t *testing.T) {
		resetSubtitleState()
		SetGlobalSubtitles(twoTracks())
		subtitleScreenFn = func(tui.SubtitleOptions) (tui.SubtitleChoice, error) {
			return tui.SubtitleChoice{Kind: "none", Session: true}, nil
		}

		PromptSubtitleLanguage()
		assert.Empty(t, GlobalSubtitles)

		// Following episodes stay silent too.
		SetGlobalSubtitles(twoTracks())
		PromptSubtitleLanguage()
		assert.Empty(t, GlobalSubtitles)
	})
}

func TestSubtitleContext(t *testing.T) {
	t.Cleanup(resetSubtitleState)

	SetSubtitleContext("Namakura Gatana", "HiAnime", "Sub", "Episode 2")
	opts := buildSubtitleOptions()
	assert.Equal(t, "Namakura Gatana", opts.Title)
	assert.Equal(t, "HiAnime · Sub", opts.Context)
	assert.Equal(t, "Namakura Gatana", opts.NextTitle)
	assert.Equal(t, "Episode 2", opts.NextLine)

	SetSubtitleContext("", "", "", "")
	opts = buildSubtitleOptions()
	assert.Equal(t, "", opts.Title)
	assert.Equal(t, "", opts.Context)
}

// TestSubtitleEpisodeLifecycle is the cascade: episodes of one show ask
// once, a new show asks again, and the marker follows the choice.
func TestSubtitleEpisodeLifecycle(t *testing.T) {
	t.Cleanup(resetSubtitleState)
	prev := subtitleScreenFn
	t.Cleanup(func() { subtitleScreenFn = prev })

	resetSubtitleState()
	SetSubtitleContext("Show A", "HiAnime", "", "")
	var seenLabels [][]string
	var seenLast []string
	subtitleScreenFn = func(opts tui.SubtitleOptions) (tui.SubtitleChoice, error) {
		var labels []string
		for _, tr := range opts.Tracks {
			labels = append(labels, tr.Label)
		}
		seenLabels = append(seenLabels, labels)
		seenLast = append(seenLast, opts.LastLabel)
		return tui.SubtitleChoice{Kind: "track", Language: "por", Session: true}, nil
	}

	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "a1-en", Language: "eng", Label: "English"},
		{URL: "a1-pt", Language: "por", Label: "Portuguese"},
	})
	PromptSubtitleLanguage()
	require.Len(t, GlobalSubtitles, 1)
	assert.Equal(t, "a1-pt", GlobalSubtitles[0].URL)
	require.Len(t, seenLabels, 1, "episode 1 asks")

	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "a2-en", Language: "eng", Label: "English"},
		{URL: "a2-pt", Language: "por", Label: "Portuguese"},
	})
	PromptSubtitleLanguage()
	assert.Len(t, seenLabels, 1, "episode 2 same show: no screen")
	require.Len(t, GlobalSubtitles, 1)
	assert.Equal(t, "a2-pt", GlobalSubtitles[0].URL)

	SetSubtitleContext("Show B", "HiAnime", "", "")
	SetGlobalSubtitles([]SubtitleInfo{
		{URL: "b-en", Language: "eng", Label: "English"},
		{URL: "b-pt", Language: "por", Label: "Portuguese"},
	})
	PromptSubtitleLanguage()
	assert.Len(t, seenLabels, 2, "another show asks again")
	assert.Equal(t, "", seenLast[1], "no marker for an unmatched last choice")
	require.Len(t, GlobalSubtitles, 1)
	assert.Equal(t, "b-pt", GlobalSubtitles[0].URL)
}

func TestSubtitleScreenError(t *testing.T) {
	t.Cleanup(resetSubtitleState)
	prev := subtitleScreenFn
	t.Cleanup(func() { subtitleScreenFn = prev })

	subtitleScreenFn = func(tui.SubtitleOptions) (tui.SubtitleChoice, error) {
		return tui.SubtitleChoice{}, errors.New("no tty")
	}
	assert.NotPanics(t, func() {
		SetGlobalSubtitles(twoTracks())
		SelectSubtitles()
		PromptSubtitleLanguage()
	})
}
