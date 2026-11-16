package appflow

import (
	"errors"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/api/providers"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
)

// Back on the TRY screen (esc on an empty box) goes to the SEARCH home screen,
// as back from the results does: the program is not left, and from the home
// anything can be done (search, CONTINUE, favorites). It used to leave no way
// out but ctrl+c, which quits.

func withHome(t *testing.T, fn func(string) (tui.SearchHomeResult, error)) {
	t.Helper()
	prev := OpenSearchHomeFn
	OpenSearchHomeFn = fn
	t.Cleanup(func() {
		OpenSearchHomeFn = prev
		PendingHomeResult = nil
	})
}

func TestTryBackOpensTheSearchHomeAndSearchesWhatIsTypedThere(t *testing.T) {
	want := &models.Anime{Name: "iCarly"}
	var searched []string
	withOverrides(t, appflowOverrides{
		searchRetry: func(name, scope string) (*models.Anime, error) {
			searched = append(searched, name+"@"+scope)
			if len(searched) == 1 {
				return nil, api.ErrNoResults
			}
			return want, nil
		},
		tryScreen: func(opts tui.TryOptions) (tui.TryDecision, error) {
			assert.Equal(t, "icarly", opts.Query)
			return tui.TryDecision{}, tui.ErrSelectionBack
		},
		promptForName: func(string) (string, error) {
			t.Error("back goes to the home screen, not the bare prompt")
			return "", errors.New("unexpected")
		},
	})
	var homeQueries []string
	withHome(t, func(q string) (tui.SearchHomeResult, error) {
		homeQueries = append(homeQueries, q)
		return tui.SearchHomeResult{Action: tui.HomeSearch, Query: "iCarly", Scope: "movietv"}, nil
	})

	got, err := SearchAnimeWithRetryIn("icarly", "ptbr")
	require.NoError(t, err, "back is not a reason to leave")
	assert.Same(t, want, got)
	assert.Equal(t, []string{"icarly"}, homeQueries, "the home opens with the query kept")
	assert.Equal(t, []string{"icarly@ptbr", "iCarly@movietv"}, searched, "and its search runs with its own scope")
}

func TestTryBackToAContinueRowPendsIt(t *testing.T) {
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) { return nil, api.ErrNoResults },
		tryScreen:   func(tui.TryOptions) (tui.TryDecision, error) { return tui.TryDecision{}, tui.ErrSelectionBack },
	})
	entry := tui.HomeContinueEntry{AnimeTitle: "Show", EpisodeNum: 4}
	withHome(t, func(string) (tui.SearchHomeResult, error) {
		return tui.SearchHomeResult{Action: tui.HomeResume, Entry: entry, HasEntry: true}, nil
	})

	anime, err := SearchAnimeWithRetryIn("x", "")
	require.NoError(t, err)
	assert.Nil(t, anime, "the caller reopens the row")
	require.NotNil(t, PendingHomeResult)
	assert.Equal(t, 4, PendingHomeResult.Entry.EpisodeNum)
}

func TestTryBackThenQuittingTheHomeQuits(t *testing.T) {
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) { return nil, api.ErrNoResults },
		tryScreen:   func(tui.TryOptions) (tui.TryDecision, error) { return tui.TryDecision{}, tui.ErrSelectionBack },
	})
	withHome(t, func(string) (tui.SearchHomeResult, error) { return tui.SearchHomeResult{}, tui.ErrSearchHomeCancelled })

	_, err := SearchAnimeWithRetryIn("x", "")
	require.ErrorIs(t, err, api.ErrSearchAborted, "ctrl+c on the home is what quits")
}

func TestTryBackWithoutAHomeScreenFallsBackToThePrompt(t *testing.T) {
	want := &models.Anime{Name: "Typed"}
	calls := 0
	withOverrides(t, appflowOverrides{
		searchRetry: func(name, _ string) (*models.Anime, error) {
			calls++
			if calls == 1 {
				return nil, api.ErrNoResults
			}
			assert.Equal(t, "typed", name)
			return want, nil
		},
		tryScreen:     func(tui.TryOptions) (tui.TryDecision, error) { return tui.TryDecision{}, tui.ErrSelectionBack },
		promptForName: func(string) (string, error) { return "typed", nil },
	})
	withHome(t, nil) // no home screen wired

	got, err := SearchAnimeWithRetryIn("x", "")
	require.NoError(t, err)
	assert.Same(t, want, got)
}

func TestTryBackFromAnEmptyHomeFallsBackToThePrompt(t *testing.T) {
	calls := 0
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) {
			calls++
			if calls == 1 {
				return nil, api.ErrNoResults
			}
			return &models.Anime{Name: "ok"}, nil
		},
		tryScreen:     func(tui.TryOptions) (tui.TryDecision, error) { return tui.TryDecision{}, tui.ErrSelectionBack },
		promptForName: func(string) (string, error) { return "typed", nil },
	})
	withHome(t, func(string) (tui.SearchHomeResult, error) { return tui.SearchHomeResult{Action: tui.HomeSearch}, nil })

	got, err := SearchAnimeWithRetryIn("x", "")
	require.NoError(t, err)
	assert.Equal(t, "ok", got.Name)
}

func TestTryBackReportsAHomeScreenThatFails(t *testing.T) {
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) { return nil, api.ErrNoResults },
		tryScreen:   func(tui.TryOptions) (tui.TryDecision, error) { return tui.TryDecision{}, tui.ErrSelectionBack },
	})
	boom := errors.New("terminal gone")
	withHome(t, func(string) (tui.SearchHomeResult, error) { return tui.SearchHomeResult{}, boom })

	_, err := SearchAnimeWithRetryIn("x", "")
	require.ErrorIs(t, err, boom)
}

func TestTryScreenErrorsOtherThanBackStillEndTheSearch(t *testing.T) {
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) { return nil, api.ErrNoResults },
		tryScreen: func(tui.TryOptions) (tui.TryDecision, error) {
			return tui.TryDecision{}, errors.Join(api.ErrSearchAborted, tui.ErrSelectionCancelled)
		},
	})
	withHome(t, func(string) (tui.SearchHomeResult, error) {
		t.Error("ctrl+c on the TRY screen quits; it does not go to the home")
		return tui.SearchHomeResult{}, nil
	})
	_, err := SearchAnimeWithRetryIn("x", "")
	require.ErrorIs(t, err, api.ErrSearchAborted)
}

func TestReportSearchOutcomeNamesTheReason(t *testing.T) {
	var lines []string
	var failures []*providers.SearchFailure
	prevErrorf, prevFailure := searchErrorfFn, reportSearchFailureFn
	t.Cleanup(func() { searchErrorfFn, reportSearchFailureFn = prevErrorf, prevFailure })
	searchErrorfFn = func(format string, a ...any) { lines = append(lines, fmt.Sprintf(format, a...)) }
	reportSearchFailureFn = func(query string, f *providers.SearchFailure) {
		assert.Equal(t, "q", query)
		failures = append(failures, f)
	}

	reportSearchOutcome("q", nil)
	reportSearchOutcome("q", fmt.Errorf("wrapped: %w", api.ErrNoResults))
	reportSearchOutcome("q", errors.New("connection refused"))
	failure := &providers.SearchFailure{}
	reportSearchOutcome("q", fmt.Errorf("search: %w", failure))

	assert.Equal(t, []string{
		"No anime found with the name: q",
		"No anime found with the name: q",
		`Search failed for "q": connection refused`,
	}, lines, "nothing found is not the same as a source that failed")
	require.Len(t, failures, 1, "a per-source failure gets its own report")
	assert.Same(t, failure, failures[0])
}
