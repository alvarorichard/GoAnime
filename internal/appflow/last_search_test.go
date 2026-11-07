package appflow

import (
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Issue #229: the user searched "karakai-jouzu", changed the title during
// playback to "wind breaker", and backing out of it landed on the results for
// "karakai-jouzu": the loop that repeats a search kept the first one, so the
// "same title from another source" the screen promised was not there.
// The recorded search is process-wide: none of these run in parallel.

func cleanLastSearch(t *testing.T) {
	t.Helper()
	resetLastSearch()
	t.Cleanup(resetLastSearch)
}

func TestSearchToRepeatBeforeAnySearchIsWhatTheCallerHad(t *testing.T) {
	cleanLastSearch(t)

	query, scope := SearchToRepeat("karakai-jouzu", "hianime")

	assert.Equal(t, "karakai-jouzu", query)
	assert.Equal(t, "hianime", scope)
}

func TestSearchAnimeWithRetryInRemembersWhatFoundTheTitle(t *testing.T) {
	cleanLastSearch(t)
	withOverrides(t, appflowOverrides{searchRetry: func(string, string) (*models.Anime, error) {
		return &models.Anime{Name: "Wind Breaker"}, nil
	}})

	_, err := SearchAnimeWithRetryIn("wind breaker", "hianime")
	require.NoError(t, err)

	query, scope := SearchToRepeat("something else", "")
	assert.Equal(t, "wind breaker", query)
	assert.Equal(t, "hianime", scope)
}

func TestSearchAnimeWithRetryInSearchesTheScopeItIsGiven(t *testing.T) {
	cleanLastSearch(t)
	var gotName, gotScope string
	withOverrides(t, appflowOverrides{searchRetry: func(name, scope string) (*models.Anime, error) {
		gotName, gotScope = name, scope
		return &models.Anime{Name: "x"}, nil
	}})

	_, err := SearchAnimeWithRetryIn("wind breaker", "anroll")

	require.NoError(t, err)
	assert.Equal(t, "wind breaker", gotName)
	assert.Equal(t, "anroll", gotScope)
}

func TestSearchAnimeWithRetryUsesTheGlobalSourceAsItAlwaysDid(t *testing.T) {
	cleanLastSearch(t)
	prev := util.GlobalSource
	util.GlobalSource = "hianime"
	t.Cleanup(func() { util.GlobalSource = prev })
	var gotScope string
	withOverrides(t, appflowOverrides{searchRetry: func(_, scope string) (*models.Anime, error) {
		gotScope = scope
		return &models.Anime{Name: "x"}, nil
	}})

	_, err := SearchAnimeWithRetry("naruto")

	require.NoError(t, err)
	assert.Equal(t, "hianime", gotScope)
	_, scope := SearchToRepeat("", "")
	assert.Equal(t, "hianime", scope)
}

// The search is edited on the way to a title (a failed search, a new term):
// what is remembered is the one that finally found it.
func TestSearchAnimeWithRetryRemembersTheEditedSearchNotTheFirst(t *testing.T) {
	cleanLastSearch(t)
	var attempts int
	withOverrides(t, appflowOverrides{
		searchRetry: func(name, _ string) (*models.Anime, error) {
			attempts++
			if attempts == 1 {
				return nil, errors.New("no result")
			}
			return &models.Anime{Name: name}, nil
		},
		promptForName: func(string) (string, error) { return "wind breaker", nil },
	})

	_, err := SearchAnimeWithRetryIn("wind braker", "")
	require.NoError(t, err)

	query, _ := SearchToRepeat("wind braker", "")
	assert.Equal(t, "wind breaker", query, "the corrected spelling, not the typo that found nothing")
}

func TestAFailedSearchRemembersNothing(t *testing.T) {
	cleanLastSearch(t)
	rememberSearch("karakai-jouzu", "")
	withOverrides(t, appflowOverrides{
		searchRetry:   func(string, string) (*models.Anime, error) { return nil, errors.New("no result") },
		promptForName: func(string) (string, error) { return "", errors.New("search cancelled by user") },
	})

	_, err := SearchAnimeWithRetryIn("zzzz", "")

	require.Error(t, err)
	query, _ := SearchToRepeat("", "")
	assert.Equal(t, "karakai-jouzu", query, "a search that found nothing is not one to come back to")
}

func TestASearchThatReturnsNoTitleRemembersNothing(t *testing.T) {
	cleanLastSearch(t)
	rememberSearch("karakai-jouzu", "")
	withOverrides(t, appflowOverrides{
		searchRetry:   func(string, string) (*models.Anime, error) { return nil, nil },
		promptForName: func(string) (string, error) { return "", errors.New("search cancelled by user") },
	})

	_, _ = SearchAnimeWithRetryIn("home continue", "")

	query, _ := SearchToRepeat("", "")
	assert.Equal(t, "karakai-jouzu", query)
}

func TestTheLatestSearchWins(t *testing.T) {
	cleanLastSearch(t)
	rememberSearch("karakai-jouzu", "")
	rememberSearch("wind breaker", "hianime")

	query, scope := SearchToRepeat("", "")

	assert.Equal(t, "wind breaker", query)
	assert.Equal(t, "hianime", scope)
}

func TestRememberSearchAcceptsAnEmptyScope(t *testing.T) {
	cleanLastSearch(t)
	rememberSearch("wind breaker", "")

	query, scope := SearchToRepeat("fallback", "fallback-scope")

	assert.Equal(t, "wind breaker", query)
	assert.Empty(t, scope, "every source is a scope too, and is not replaced by the caller's")
}

// ---- changing the title during playback ------------------------------------------

func openHomeFor(t *testing.T, res tui.SearchHomeResult) {
	t.Helper()
	prev := OpenSearchHomeFn
	OpenSearchHomeFn = func(string) (tui.SearchHomeResult, error) { return res, nil }
	t.Cleanup(func() { OpenSearchHomeFn = prev })
}

func TestChangeAnimeHomeRemembersTheSearchItMade(t *testing.T) {
	cleanLastSearch(t)
	rememberSearch("karakai-jouzu", "")
	openHomeFor(t, tui.SearchHomeResult{Query: "wind breaker", Scope: "hianime"})
	var gotName, gotScope string
	withOverrides(t, appflowOverrides{
		searchRetry: func(name, scope string) (*models.Anime, error) {
			gotName, gotScope = name, scope
			return &models.Anime{Name: "Wind Breaker"}, nil
		},
		getEpisodes: func(*models.Anime) ([]models.Episode, error) { return []models.Episode{{Num: 1}}, nil },
	})

	anime, episodes, err := ChangeAnimeHome()

	require.NoError(t, err)
	assert.Equal(t, "Wind Breaker", anime.Name)
	assert.Len(t, episodes, 1)
	assert.Equal(t, "wind breaker", gotName)
	assert.Equal(t, "hianime", gotScope)
	query, scope := SearchToRepeat("karakai-jouzu", "")
	assert.Equal(t, "wind breaker", query, "backing out of the new title shows its results")
	assert.Equal(t, "hianime", scope)
}

func TestChangeAnimeHomeKeepsTheOldSearchWhenTheNewTitleHasNoEpisodes(t *testing.T) {
	cleanLastSearch(t)
	rememberSearch("karakai-jouzu", "")
	openHomeFor(t, tui.SearchHomeResult{Query: "wind breaker"})
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) { return &models.Anime{Name: "Wind Breaker"}, nil },
		getEpisodes: func(*models.Anime) ([]models.Episode, error) { return nil, nil },
	})

	_, _, err := ChangeAnimeHome()

	require.Error(t, err)
	query, _ := SearchToRepeat("", "")
	assert.Equal(t, "karakai-jouzu", query, "the user stays with the title already playing, which that search found")
}

func TestChangeAnimeHomeKeepsTheOldSearchWhenEpisodesFail(t *testing.T) {
	cleanLastSearch(t)
	rememberSearch("karakai-jouzu", "")
	openHomeFor(t, tui.SearchHomeResult{Query: "wind breaker"})
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) { return &models.Anime{Name: "Wind Breaker"}, nil },
		getEpisodes: func(*models.Anime) ([]models.Episode, error) { return nil, errors.New("offline") },
	})

	_, _, err := ChangeAnimeHome()

	require.Error(t, err)
	query, _ := SearchToRepeat("", "")
	assert.Equal(t, "karakai-jouzu", query)
}

func TestChangeAnimeHomeKeepsTheOldSearchWhenTheSearchFails(t *testing.T) {
	cleanLastSearch(t)
	rememberSearch("karakai-jouzu", "")
	openHomeFor(t, tui.SearchHomeResult{Query: "wind breaker"})
	withOverrides(t, appflowOverrides{
		searchRetry: func(string, string) (*models.Anime, error) { return nil, api.ErrNoResults },
	})

	_, _, err := ChangeAnimeHome()

	require.Error(t, err)
	query, _ := SearchToRepeat("", "")
	assert.Equal(t, "karakai-jouzu", query)
}

func TestSearchAnimeWithRetryRemembersTheScopeTheRetryScreenChose(t *testing.T) {
	cleanLastSearch(t)
	var attempts int
	withOverrides(t, appflowOverrides{
		searchRetry: func(name, _ string) (*models.Anime, error) {
			attempts++
			if attempts == 1 {
				return nil, errors.New("no result")
			}
			return &models.Anime{Name: name}, nil
		},
		tryScreen: func(tui.TryOptions) (tui.TryDecision, error) {
			return tui.TryDecision{Query: "wind breaker", Scope: "hianime"}, nil
		},
	})

	_, err := SearchAnimeWithRetryIn("wind braker", "")
	require.NoError(t, err)

	query, scope := SearchToRepeat("", "")
	assert.Equal(t, "wind breaker", query)
	assert.Equal(t, "hianime", scope, "the retry that found the title was limited to English sources, and so is the one repeated")
}
