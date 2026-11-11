package handlers

import (
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// freshHome isolates the history files and forgets the last home decision.
func freshHome(t *testing.T) {
	t.Helper()
	isolateHome(t)
	resetLastHome()
	t.Cleanup(resetLastHome)
}

// answer is a screen runner that returns res and records the options it got.
func answer(res tui.SearchHomeResult, err error, got *tui.SearchHomeOptions) func(tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
	return func(o tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
		*got = o
		return res, err
	}
}

// reopen is the way back: the home opened with the query kept, and left
// again with esc, which decides nothing and so leaves the memory as it was.
func reopen(t *testing.T, query string) tui.SearchHomeOptions {
	t.Helper()
	var got tui.SearchHomeOptions
	_, err := openSearchHome(query, answer(tui.SearchHomeResult{}, tui.ErrSearchHomeCancelled, &got))
	require.ErrorIs(t, err, tui.ErrSearchHomeCancelled)
	return got
}

func decide(t *testing.T, res tui.SearchHomeResult) {
	t.Helper()
	var got tui.SearchHomeOptions
	_, err := openSearchHome("", answer(res, nil, &got))
	require.NoError(t, err)
}

func TestFirstOpeningRestoresNothing(t *testing.T) {
	freshHome(t)
	assert.True(t, reopen(t, "").Restore.IsZero())
	assert.True(t, reopen(t, "naruto").Restore.IsZero(), "nothing was decided yet")
}

func TestBackFromAFavoriteRestoresTheFavoritesTabAndRow(t *testing.T) {
	freshHome(t)
	want := tui.HomeRestore{Tab: tui.TabFavorites, RowKey: "favorite:abc"}
	decide(t, tui.SearchHomeResult{Action: tui.HomeSearch, Query: "Sousou no Frieren", Restore: want})

	// main hyphenates the query before the search, and that is what comes back
	got := reopen(t, util.TreatingAnimeName("Sousou no Frieren"))
	assert.Equal(t, want, got.Restore)
	assert.Equal(t, "sousou-no-frieren", got.Query, "the query is still passed as today; the screen decides what to show")
}

func TestBackRestoresEachRoute(t *testing.T) {
	routes := map[string]struct {
		res   tui.SearchHomeResult
		query string
	}{
		"box ALL":         {tui.SearchHomeResult{Action: tui.HomeSearch, Query: "one piece", Restore: tui.HomeRestore{Tab: tui.TabAll, Query: "one piece"}}, "one-piece"},
		"box PT-BR":       {tui.SearchHomeResult{Action: tui.HomeSearch, Query: "one piece", Scope: "ptbr", Restore: tui.HomeRestore{Tab: tui.TabPTBR, Query: "one piece"}}, "one-piece"},
		"box ENGLISH":     {tui.SearchHomeResult{Action: tui.HomeSearch, Query: "one piece", Scope: "hianime", Restore: tui.HomeRestore{Tab: tui.TabEnglish, Query: "one piece"}}, "one-piece"},
		"box MOVIES/TV":   {tui.SearchHomeResult{Action: tui.HomeSearch, Query: "one piece", Scope: "movietv", Restore: tui.HomeRestore{Tab: tui.TabMoviesTV, Query: "one piece"}}, "one-piece"},
		"recent row":      {tui.SearchHomeResult{Action: tui.HomeSearch, Query: "momotarou", Restore: tui.HomeRestore{Tab: tui.TabAll, RowKey: "recent:momotarou"}}, "momotarou"},
		"favorite on ALL": {tui.SearchHomeResult{Action: tui.HomeSearch, Query: "A Origem", Restore: tui.HomeRestore{Tab: tui.TabAll, RowKey: "favorite:k"}}, "a-origem"},
		"favorites tab":   {tui.SearchHomeResult{Action: tui.HomeSearch, Query: "A Origem", Restore: tui.HomeRestore{Tab: tui.TabFavorites, RowKey: "favorite:k"}}, "a-origem"},
	}
	for name, rt := range routes {
		t.Run(name, func(t *testing.T) {
			freshHome(t)
			decide(t, rt.res)
			assert.Equal(t, rt.res.Restore, reopen(t, rt.query).Restore)
		})
	}
}

func TestBackFromAContinueRowFindsTheTitleTheFallbackSearches(t *testing.T) {
	freshHome(t)
	want := tui.HomeRestore{Tab: tui.TabPTBR, RowKey: "continue:t:naruto\x00animefire.io"}
	for _, action := range []tui.HomeAction{tui.HomeResume, tui.HomeFromStart, tui.HomeEpisodes} {
		resetLastHome()
		decide(t, tui.SearchHomeResult{
			Action: action, HasEntry: true, Scope: "ptbr", Restore: want,
			Entry: tui.HomeContinueEntry{Key: "u:ep3", AnimeTitle: "[PT-BR] Naruto (Legendado)", Source: "Animefire.io"},
		})
		// openContinue backs out to HandlePlaybackMode(cleanContinueTitle(entry))
		assert.Equal(t, want, reopen(t, "Naruto").Restore, "action %d", action)
		assert.Equal(t, want, reopen(t, "naruto").Restore, "action %d", action)
	}
}

func TestBackAfterADifferentSearchRestoresNothing(t *testing.T) {
	freshHome(t)
	decide(t, tui.SearchHomeResult{Action: tui.HomeSearch, Query: "one piece", Restore: tui.HomeRestore{Tab: tui.TabFavorites, RowKey: "favorite:k"}})
	// the search was edited on the way (a TRY suggestion, say)
	assert.True(t, reopen(t, "piece").Restore.IsZero())
	assert.True(t, reopen(t, "one-piece-2").Restore.IsZero())
	assert.True(t, reopen(t, "").Restore.IsZero(), "a screen opened with no query is a fresh one (change anime)")
	assert.True(t, reopen(t, "   ").Restore.IsZero())
}

func TestTheLastDecisionReplacesTheOneBefore(t *testing.T) {
	freshHome(t)
	decide(t, tui.SearchHomeResult{Action: tui.HomeSearch, Query: "one piece", Restore: tui.HomeRestore{Tab: tui.TabFavorites, RowKey: "favorite:k"}})
	decide(t, tui.SearchHomeResult{Action: tui.HomeSearch, Query: "naruto", Restore: tui.HomeRestore{Tab: tui.TabPTBR, Query: "naruto"}})
	assert.True(t, reopen(t, "one-piece").Restore.IsZero(), "the first search is no longer the one being left")
	assert.Equal(t, tui.HomeRestore{Tab: tui.TabPTBR, Query: "naruto"}, reopen(t, "naruto").Restore)

	// a decision with nothing to restore clears the memory instead of keeping a stale one
	decide(t, tui.SearchHomeResult{Action: tui.HomeSearch, Query: "naruto"})
	assert.True(t, reopen(t, "naruto").Restore.IsZero())
}

func TestACancelledOrFailedHomeKeepsTheLastDecision(t *testing.T) {
	freshHome(t)
	want := tui.HomeRestore{Tab: tui.TabEnglish, Query: "bleach"}
	decide(t, tui.SearchHomeResult{Action: tui.HomeSearch, Query: "bleach", Restore: want})
	for _, e := range []error{tui.ErrSearchHomeCancelled, errors.New("boom")} {
		var got tui.SearchHomeOptions
		_, err := openSearchHome("bleach", answer(tui.SearchHomeResult{Restore: tui.HomeRestore{Tab: tui.TabAll}, Query: "x"}, e, &got))
		require.ErrorIs(t, err, e)
		assert.Equal(t, want, got.Restore)
	}
	assert.Equal(t, want, reopen(t, "bleach").Restore)
}

func TestTheScreensDecisionStillPassesStraightThrough(t *testing.T) {
	freshHome(t)
	want := tui.SearchHomeResult{Action: tui.HomeSearch, Query: "kept", Restore: tui.HomeRestore{Tab: tui.TabAll, Query: "kept"}}
	var got tui.SearchHomeOptions
	res, err := openSearchHome("", answer(want, nil, &got))
	require.NoError(t, err)
	assert.Equal(t, want, res)
}

func TestSameSearch(t *testing.T) {
	assert.True(t, sameSearch("One Piece", "one-piece"))
	assert.True(t, sameSearch("  one   piece ", "one-piece"))
	assert.True(t, sameSearch("one-piece", "one-piece"))
	assert.False(t, sameSearch("one piece", "one-piece-film"))
	assert.False(t, sameSearch("one piece", "onepiece"))
}

func TestSearchedByHome(t *testing.T) {
	assert.Equal(t, "q", searchedByHome(tui.SearchHomeResult{Query: "q"}))
	assert.Equal(t, "Bleach", searchedByHome(tui.SearchHomeResult{
		HasEntry: true, Query: "ignored", Entry: tui.HomeContinueEntry{AnimeTitle: "[HiAnime] Bleach (Dublado)"},
	}))
}

// The favorites tab is a tab, never a search scope: nothing about reopening
// the home touches the source the searches run with.
func TestRestoreNeverReachesTheSearchScope(t *testing.T) {
	freshHome(t)
	prev := util.GlobalSource
	t.Cleanup(func() { util.GlobalSource = prev })
	util.GlobalSource = "hianime"

	decide(t, tui.SearchHomeResult{Action: tui.HomeSearch, Query: "Frieren", Restore: tui.HomeRestore{Tab: tui.TabFavorites, RowKey: "favorite:k"}})
	got := reopen(t, "frieren")

	assert.Equal(t, "hianime", util.GlobalSource, "reopening the home leaves the source alone")
	assert.Equal(t, 2, got.ScopeIndex, "ScopeIndex is still the --source flag's; the screen lets the restore win")
	assert.Equal(t, tui.TabFavorites, got.Restore.Tab)
}

// Through the public entry points, with the terminal and the screen stubbed,
// as main and the retry loop use them: home, search, back.
func TestOpenSearchHomeBackPathEndToEnd(t *testing.T) {
	freshHome(t)
	prevTerm, prevRun := stdinIsTerminal, runSearchHome
	t.Cleanup(func() { stdinIsTerminal, runSearchHome = prevTerm, prevRun })
	stdinIsTerminal = func() bool { return true }

	want := tui.HomeRestore{Tab: tui.TabFavorites, RowKey: "favorite:k"}
	var opened []tui.SearchHomeOptions
	runSearchHome = func(o tui.SearchHomeOptions) (tui.SearchHomeResult, error) {
		opened = append(opened, o)
		if len(opened) == 1 {
			return tui.SearchHomeResult{Action: tui.HomeSearch, Query: "Sousou no Frieren", Restore: want}, nil
		}
		return tui.SearchHomeResult{}, tui.ErrSearchHomeCancelled
	}

	decision, err := HandleSearchHome() // main
	require.NoError(t, err)
	assert.Equal(t, SearchHomeDecision{Query: "Sousou no Frieren"}, decision, "the decision main acts on is unchanged: empty scope, so GlobalSource is not touched")

	_, err = OpenSearchHome(util.TreatingAnimeName(decision.Query)) // appflow, on esc from the results
	require.ErrorIs(t, err, tui.ErrSearchHomeCancelled)

	require.Len(t, opened, 2)
	assert.True(t, opened[0].Restore.IsZero())
	assert.Equal(t, want, opened[1].Restore)
}
