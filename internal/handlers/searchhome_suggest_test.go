package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/suggest"
	"github.com/alvarorichard/Goanime/v2/internal/tracking"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isolateHome points the home directory at a temp dir so the watch-history
// database never touches the real one, and closes the tracker it opened
// before the dir is removed (Windows cannot delete an open database file).
func isolateHome(t *testing.T) {
	t.Helper()
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("USERPROFILE", home)
	t.Setenv("LOCALAPPDATA", home)
	t.Cleanup(func() { _ = tracking.CloseGlobalTracker() })
}

func TestAdaptSuggestions(t *testing.T) {
	t.Parallel()
	rows, err := adaptSuggestions([]suggest.Suggestion{
		{Title: "Naruto", Year: 2002, Kind: suggest.KindAnime},
		{Title: "Dune", Kind: suggest.KindMovie},
	}, nil)
	require.NoError(t, err)
	assert.Equal(t, []tui.HomeSuggestion{
		{Title: "Naruto", Detail: "Anime · 2002"},
		{Title: "Dune", Detail: "Movie"},
	}, rows)

	rows, err = adaptSuggestions(nil, nil)
	require.NoError(t, err)
	assert.Empty(t, rows)

	boom := errors.New("boom")
	rows, err = adaptSuggestions([]suggest.Suggestion{{Title: "x"}}, boom)
	require.ErrorIs(t, err, boom)
	assert.Nil(t, rows, "an error never leaks half an answer")
}

func TestHomeSuggestShortQueryStaysOffline(t *testing.T) {
	t.Parallel()
	rows, err := homeSuggest(t.Context(), "a", "")
	require.NoError(t, err)
	assert.Empty(t, rows, "one letter never reaches the network")
}

// TestSearchHomeWiresAutocomplete guards the integration seam: the home
// screen must be opened with the suggestion engine attached, or the box
// silently loses its autocomplete.
func TestSearchHomeWiresAutocomplete(t *testing.T) {
	isolateHome(t)

	var got tui.SearchHomeOptions
	want := tui.SearchHomeResult{Action: tui.HomeSearch, Query: "kept"}
	res, err := openSearchHome("naruto", func(opts tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
		got = opts
		return want, nil
	})
	require.NoError(t, err)
	assert.Equal(t, want, res, "the screen's decision passes straight through")
	assert.Equal(t, "naruto", got.Query)
	require.NotNil(t, got.Suggest, "OpenSearchHome must attach the autocomplete")

	rows, err := got.Suggest(t.Context(), "a", "")
	require.NoError(t, err)
	assert.Empty(t, rows, "the attached func is the engine adapter (a one-letter query stays offline)")
}

// TestOpenSearchHomeEndToEnd runs the public entry point with the terminal
// and the screen stubbed: no terminal is a clear error, a terminal reaches
// the screen with the autocomplete attached.
func TestOpenSearchHomeEndToEnd(t *testing.T) {
	isolateHome(t)
	origTerm, origRun := stdinIsTerminal, runSearchHome
	t.Cleanup(func() { stdinIsTerminal, runSearchHome = origTerm, origRun })

	stdinIsTerminal = func() bool { return false }
	runSearchHome = func(tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
		t.Fatal("the screen must not open without a terminal")
		return tui.SearchHomeResult{}, nil
	}
	_, err := OpenSearchHome("")
	require.ErrorContains(t, err, "no interactive terminal")

	stdinIsTerminal = func() bool { return true }
	var opened tui.SearchHomeOptions
	runSearchHome = func(opts tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
		opened = opts
		return tui.SearchHomeResult{Action: tui.HomeSearch, Query: "dune"}, nil
	}
	res, err := OpenSearchHome("du")
	require.NoError(t, err)
	assert.Equal(t, "dune", res.Query)
	assert.Equal(t, "du", opened.Query)
	assert.NotNil(t, opened.Suggest)
}

// TestHomeSuggestPassesTheScope guards that the scope tab narrows what the
// engine is asked for.
func TestHomeSuggestPassesTheScope(t *testing.T) {
	t.Parallel()
	var asked []suggest.Kind
	rows, err := homeSuggestWith(t.Context(), "naruto", "ptbr", func(_ context.Context, _ string, kinds []suggest.Kind) ([]suggest.Suggestion, error) {
		asked = kinds
		return []suggest.Suggestion{{Title: "Naruto", Kind: suggest.KindAnime}}, nil
	})
	require.NoError(t, err)
	assert.Equal(t, []suggest.Kind{suggest.KindAnime}, asked)
	assert.Equal(t, []tui.HomeSuggestion{{Title: "Naruto", Detail: "Anime"}}, rows)

	_, err = homeSuggestWith(t.Context(), "naruto", "movietv", func(_ context.Context, _ string, kinds []suggest.Kind) ([]suggest.Suggestion, error) {
		asked = kinds
		return nil, nil
	})
	require.NoError(t, err)
	assert.Equal(t, []suggest.Kind{suggest.KindMovie, suggest.KindSeries}, asked)
}
