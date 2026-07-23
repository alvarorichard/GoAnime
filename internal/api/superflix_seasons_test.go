package api

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// sfSeasonsHarness stubs the episode-listing seams and the browser-warning seams
// so a test can observe exactly which path fetchSuperFlixSeasons took. It records
// whether the browser was reached and whether the "a browser will open" warnings
// fired. Tests using it must NOT be parallel — they swap package-level vars.
type sfSeasonsHarness struct {
	tvmazeCalls  int
	browserCalls int
	warned       bool
}

func (h *sfSeasonsHarness) install(
	t *testing.T,
	tvmaze func(context.Context, string) (map[string][]superflix.SuperFlixEpisode, error),
	browser func(context.Context, *superflix.SuperFlixClient, string) (map[string][]superflix.SuperFlixEpisode, error),
) {
	t.Helper()
	prevTV, prevBrowser := sfTVmazeEpisodesFn, sfBrowserEpisodesFn
	prevHeadless, prevPending := sfHeadlessEnvFn, sfSetupPendingFn
	prevWarn, prevInfo := sfWarnFn, sfInfoFn
	t.Cleanup(func() {
		sfTVmazeEpisodesFn, sfBrowserEpisodesFn = prevTV, prevBrowser
		sfHeadlessEnvFn, sfSetupPendingFn = prevHeadless, prevPending
		sfWarnFn, sfInfoFn = prevWarn, prevInfo
	})

	sfTVmazeEpisodesFn = func(ctx context.Context, imdbID string) (map[string][]superflix.SuperFlixEpisode, error) {
		h.tvmazeCalls++
		return tvmaze(ctx, imdbID)
	}
	sfBrowserEpisodesFn = func(ctx context.Context, c *superflix.SuperFlixClient, tmdbID string) (map[string][]superflix.SuperFlixEpisode, error) {
		h.browserCalls++
		return browser(ctx, c, tmdbID)
	}
	// preflightSuperFlixBrowser only warns when it thinks a browser is needed;
	// force both notices on so any call to it is unmistakable.
	sfHeadlessEnvFn = func() bool { return true }
	sfSetupPendingFn = func() bool { return true }
	sfWarnFn = func(any, ...any) { h.warned = true }
	sfInfoFn = func(any, ...any) { h.warned = true }
}

func sfEpisodes(n int) map[string][]superflix.SuperFlixEpisode {
	eps := make([]superflix.SuperFlixEpisode, n)
	return map[string][]superflix.SuperFlixEpisode{"1": eps}
}

func noBrowser(t *testing.T) func(context.Context, *superflix.SuperFlixClient, string) (map[string][]superflix.SuperFlixEpisode, error) {
	t.Helper()
	return func(context.Context, *superflix.SuperFlixClient, string) (map[string][]superflix.SuperFlixEpisode, error) {
		t.Error("the browser must not be used when TVmaze can answer")
		return nil, nil
	}
}

// TVmaze is the browser-free path and must be tried FIRST. If it answers, the
// headed browser must never be started — and the user must not be told a browser
// window is about to open. Emitting those warnings up front (the old behavior)
// scared users on the common path, and scraping SuperFlix is unreliable anyway:
// it now often serves an embed-only shell with no episode list (issue #184).
func TestFetchSuperFlixSeasons_TVmazeWinsAndSkipsBrowser(t *testing.T) {
	h := &sfSeasonsHarness{}
	h.install(t,
		func(_ context.Context, imdbID string) (map[string][]superflix.SuperFlixEpisode, error) {
			assert.Equal(t, "tt0857297", imdbID)
			return sfEpisodes(24), nil
		},
		noBrowser(t),
	)

	media := &models.Anime{Name: "NHK ni Youkoso", IMDBID: "tt0857297", URL: "42821"}
	got, err := fetchSuperFlixSeasons(nil, media, "42821")

	require.NoError(t, err)
	assert.Len(t, got["1"], 24)
	assert.Equal(t, 1, h.tvmazeCalls)
	assert.Equal(t, 0, h.browserCalls, "browser must not be reached")
	assert.False(t, h.warned, "no browser warning may be shown when TVmaze answered")
}

func TestFetchSuperFlixSeasons_FallsBackToBrowser(t *testing.T) {
	tests := []struct {
		name   string
		imdb   string
		tvmaze func(context.Context, string) (map[string][]superflix.SuperFlixEpisode, error)
		wantTV int
	}{
		{
			name: "TVmaze errors",
			imdb: "tt0000001",
			tvmaze: func(context.Context, string) (map[string][]superflix.SuperFlixEpisode, error) {
				return nil, errors.New("tvmaze down")
			},
			wantTV: 1,
		},
		{
			name:   "TVmaze has no listing",
			imdb:   "tt0000002",
			tvmaze: func(context.Context, string) (map[string][]superflix.SuperFlixEpisode, error) { return nil, nil },
			wantTV: 1,
		},
		{
			// Without an IMDB id there is nothing to look TVmaze up by.
			name: "no IMDB id skips TVmaze entirely",
			imdb: "",
			tvmaze: func(context.Context, string) (map[string][]superflix.SuperFlixEpisode, error) {
				return sfEpisodes(9), nil
			},
			wantTV: 0,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &sfSeasonsHarness{}
			h.install(t, tt.tvmaze,
				func(context.Context, *superflix.SuperFlixClient, string) (map[string][]superflix.SuperFlixEpisode, error) {
					return sfEpisodes(12), nil
				},
			)

			media := &models.Anime{Name: "X", IMDBID: tt.imdb, URL: "1396"}
			got, err := fetchSuperFlixSeasons(nil, media, "1396")

			require.NoError(t, err)
			assert.Len(t, got["1"], 12, "episodes must come from the browser fallback")
			assert.Equal(t, tt.wantTV, h.tvmazeCalls)
			assert.Equal(t, 1, h.browserCalls)
			assert.True(t, h.warned, "reaching the browser must warn the user first")
		})
	}
}

// When neither path yields a listing the user must get an actionable message —
// not the bare "no seasons found" that issue #184 dead-ended on — and the cause
// must still carry the ids needed to debug it.
func TestFetchSuperFlixSeasons_BothPathsEmptyGivesActionableError(t *testing.T) {
	h := &sfSeasonsHarness{}
	h.install(t,
		func(context.Context, string) (map[string][]superflix.SuperFlixEpisode, error) { return nil, nil },
		func(context.Context, *superflix.SuperFlixClient, string) (map[string][]superflix.SuperFlixEpisode, error) {
			return nil, nil
		},
	)

	media := &models.Anime{Name: "X", IMDBID: "tt0857297", URL: "42821"}
	_, err := fetchSuperFlixSeasons(nil, media, "42821")

	require.Error(t, err)
	assert.NotEqual(t, "no seasons found", err.Error(), "the old dead-end message must be gone")
	assert.Contains(t, err.Error(), "another source", "must point the user somewhere useful")

	// The technical cause stays reachable for debugging, with both ids.
	cause := errors.Unwrap(err)
	require.NotNil(t, cause)
	assert.Contains(t, cause.Error(), "tmdb=42821")
	assert.Contains(t, cause.Error(), "tt0857297")
}

// A browser scrape that finds no episode list is a scrape failure. It must reach
// the user as plain language via describeSuperFlixErr, never as raw jargon.
func TestFetchSuperFlixSeasons_NoEpisodeListErrorIsFriendly(t *testing.T) {
	h := &sfSeasonsHarness{}
	h.install(t,
		func(context.Context, string) (map[string][]superflix.SuperFlixEpisode, error) { return nil, nil },
		func(context.Context, *superflix.SuperFlixClient, string) (map[string][]superflix.SuperFlixEpisode, error) {
			return nil, superflix.ErrSuperFlixNoEpisodeList
		},
	)

	media := &models.Anime{Name: "X", IMDBID: "tt1", URL: "46260"}
	_, err := fetchSuperFlixSeasons(nil, media, "46260")

	require.Error(t, err)
	assert.ErrorIs(t, err, superflix.ErrSuperFlixNoEpisodeList, "the sentinel must survive for callers")
	assert.Contains(t, err.Error(), "another source")
	assert.NotContains(t, err.Error(), "exposed no episode list", "internal wording must not leak to the user")
}

func TestDescribeSuperFlixErr_MapsNoEpisodeList(t *testing.T) {
	t.Parallel()
	got := describeSuperFlixErr(superflix.ErrSuperFlixNoEpisodeList)
	require.Error(t, got)
	assert.ErrorIs(t, got, superflix.ErrSuperFlixNoEpisodeList)
	assert.Contains(t, got.Error(), "another source")
}

// The season picker used sort.Strings, which orders "10" before "2" — a show with
// ten or more seasons listed them scrambled.
func TestSortedSeasonNumbers(t *testing.T) {
	t.Parallel()

	seasons := func(keys ...string) map[string][]superflix.SuperFlixEpisode {
		m := make(map[string][]superflix.SuperFlixEpisode, len(keys))
		for _, k := range keys {
			m[k] = []superflix.SuperFlixEpisode{{}}
		}
		return m
	}

	tests := []struct {
		name string
		in   map[string][]superflix.SuperFlixEpisode
		want []string
	}{
		{
			name: "double digits sort numerically, not lexicographically",
			in:   seasons("1", "2", "10", "11", "3"),
			want: []string{"1", "2", "3", "10", "11"},
		},
		{"single season", seasons("1"), []string{"1"}},
		{"empty", seasons(), []string{}},
		{
			// TVmaze exposes year-based "seasons" for some long-running anime.
			name: "year-like seasons stay ordered",
			in:   seasons("2003", "2002", "2007"),
			want: []string{"2002", "2003", "2007"},
		},
		{
			name: "non-numeric keys sort after numeric ones, deterministically",
			in:   seasons("2", "especial", "1"),
			want: []string{"1", "2", "especial"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, sortedSeasonNumbers(tt.in))
		})
	}
}
