package playback

import (
	"errors"
	"fmt"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Test infrastructure
// ---------------------------------------------------------------------------

// selectorOverrideMu serialises mutations to the package-level injection vars.
// Tests that touch selectEpisodeFunc / extractEpisodeNumberFunc must run
// sequentially against each other; they may still parallel against unrelated
// tests in this package.
var selectorOverrideMu sync.Mutex

// withSelector swaps both package-level injections for the test's duration
// and restores them on cleanup. Caller holds selectorOverrideMu via the
// surrounding sequential test pattern (no t.Parallel inside).
func withSelector(t *testing.T, sel selectEpisodeFuncType, extract extractEpisodeNumberFuncType) {
	t.Helper()
	selectorOverrideMu.Lock()
	oldSel := selectEpisodeFunc
	oldExtract := extractEpisodeNumberFunc
	if sel != nil {
		selectEpisodeFunc = sel
	}
	if extract != nil {
		extractEpisodeNumberFunc = extract
	}
	t.Cleanup(func() {
		selectEpisodeFunc = oldSel
		extractEpisodeNumberFunc = oldExtract
		selectorOverrideMu.Unlock()
	})
}

func sampleEpisodes(n int) []models.Episode {
	out := make([]models.Episode, 0, n)
	for i := 1; i <= n; i++ {
		out = append(out, models.Episode{
			Num:    i,
			Number: fmt.Sprintf("Episode %d", i),
			URL:    fmt.Sprintf("https://example.test/ep%d", i),
		})
	}
	return out
}

// ---------------------------------------------------------------------------
// parseEpisodeSelection — pure post-processing of fuzzy-finder output.
// Table-driven exhaustive.
// ---------------------------------------------------------------------------

func TestParseEpisodeSelection(t *testing.T) {
	t.Parallel()

	cases := []struct {
		name        string
		url         string
		numStr      string
		fuzzyErr    error
		extractor   extractEpisodeNumberFuncType
		wantURL     string
		wantNumStr  string
		wantEpNum   int
		wantErrIs   error
		wantErrSubs string // contains
	}{
		{
			name:       "success_episode_5",
			url:        "https://x/ep5",
			numStr:     "Episode 5",
			extractor:  func(s string) string { return "5" },
			wantURL:    "https://x/ep5",
			wantNumStr: "Episode 5",
			wantEpNum:  5,
		},
		{
			name:       "success_episode_1",
			url:        "https://x/ep1",
			numStr:     "1",
			extractor:  func(string) string { return "1" },
			wantURL:    "https://x/ep1",
			wantNumStr: "1",
			wantEpNum:  1,
		},
		{
			name:       "success_episode_999",
			url:        "https://x/ep999",
			numStr:     "Episodio 999",
			extractor:  func(string) string { return "999" },
			wantURL:    "https://x/ep999",
			wantNumStr: "Episodio 999",
			wantEpNum:  999,
		},
		{
			name:      "back_requested",
			fuzzyErr:  player.ErrBackRequested,
			wantEpNum: -1,
			wantErrIs: player.ErrBackRequested,
		},
		{
			name:        "generic_fuzzy_error",
			fuzzyErr:    errors.New("tcell crash"),
			wantErrSubs: "tcell crash",
		},
		{
			name:        "atoi_failure_non_numeric",
			url:         "https://x/ep",
			numStr:      "OVA Special",
			extractor:   func(string) string { return "abc" },
			wantErrSubs: "invalid syntax",
		},
		{
			name:        "atoi_failure_empty",
			numStr:      "",
			extractor:   func(string) string { return "" },
			wantErrSubs: "invalid syntax",
		},
	}

	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Tests that need extractor override are sequential against other
			// override tests; safe to parallel here only if we don't touch
			// the package var. parseEpisodeSelection accepts the extractor
			// via the package var, so we must serialise.
			withSelector(t, nil, tc.extractor)

			gotURL, gotNumStr, gotEpNum, err := parseEpisodeSelection(tc.url, tc.numStr, tc.fuzzyErr)

			if tc.wantErrIs != nil {
				assert.ErrorIs(t, err, tc.wantErrIs)
				assert.Equal(t, tc.wantEpNum, gotEpNum)
				return
			}
			if tc.wantErrSubs != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tc.wantErrSubs)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tc.wantURL, gotURL)
			assert.Equal(t, tc.wantNumStr, gotNumStr)
			assert.Equal(t, tc.wantEpNum, gotEpNum)
		})
	}
}

// ---------------------------------------------------------------------------
// SelectInitialEpisode — full pipeline through injected mock selector.
// ---------------------------------------------------------------------------

func TestSelectInitialEpisode_Success(t *testing.T) {
	eps := sampleEpisodes(3)
	var callCount int32
	withSelector(t,
		func(got []models.Episode) (string, string, error) {
			atomic.AddInt32(&callCount, 1)
			assert.Equal(t, eps, got, "selector must receive the episode list verbatim")
			return "https://x/ep2", "Episode 2", nil
		},
		func(s string) string {
			assert.Equal(t, "Episode 2", s)
			return "2"
		},
	)

	url, numStr, epNum, err := SelectInitialEpisode(eps)
	require.NoError(t, err)
	assert.Equal(t, "https://x/ep2", url)
	assert.Equal(t, "Episode 2", numStr)
	assert.Equal(t, 2, epNum)
	assert.Equal(t, int32(1), atomic.LoadInt32(&callCount), "selector must be called exactly once")
}

func TestSelectInitialEpisode_BackRequested(t *testing.T) {
	withSelector(t,
		func([]models.Episode) (string, string, error) {
			return "", "", player.ErrBackRequested
		},
		nil,
	)

	url, numStr, epNum, err := SelectInitialEpisode(sampleEpisodes(2))
	assert.ErrorIs(t, err, player.ErrBackRequested)
	assert.Empty(t, url)
	assert.Empty(t, numStr)
	assert.Equal(t, -1, epNum, "back must return sentinel -1 to distinguish from valid 0")
}

func TestSelectInitialEpisode_GenericSelectorError(t *testing.T) {
	want := errors.New("tui crashed")
	withSelector(t,
		func([]models.Episode) (string, string, error) {
			return "", "", want
		},
		nil,
	)

	url, numStr, epNum, err := SelectInitialEpisode(sampleEpisodes(2))
	assert.ErrorIs(t, err, want)
	assert.Empty(t, url)
	assert.Empty(t, numStr)
	assert.Equal(t, 0, epNum, "generic error must return 0 (not -1 which is the back sentinel)")
}

func TestSelectInitialEpisode_AtoiFailureBubblesUp(t *testing.T) {
	withSelector(t,
		func([]models.Episode) (string, string, error) {
			return "https://x/special", "OVA", nil
		},
		func(string) string { return "not-a-number" },
	)

	url, numStr, epNum, err := SelectInitialEpisode(sampleEpisodes(1))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid syntax")
	assert.Empty(t, url)
	assert.Empty(t, numStr)
	assert.Equal(t, 0, epNum)
}

func TestSelectInitialEpisode_EmptyEpisodeList(t *testing.T) {
	// Selector behaviour for empty input is its own responsibility — we just
	// verify SelectInitialEpisode forwards whatever the selector returns.
	want := errors.New("no episodes provided")
	withSelector(t,
		func(got []models.Episode) (string, string, error) {
			assert.Empty(t, got)
			return "", "", want
		},
		nil,
	)

	url, numStr, epNum, err := SelectInitialEpisode(nil)
	assert.ErrorIs(t, err, want)
	assert.Empty(t, url)
	assert.Empty(t, numStr)
	assert.Equal(t, 0, epNum)
}

// TestSelectInitialEpisode_RealExtractor_RouteThrough confirms the default
// production extractor is wired correctly when only the selector is overridden.
func TestSelectInitialEpisode_RealExtractor_RouteThrough(t *testing.T) {
	withSelector(t,
		func([]models.Episode) (string, string, error) {
			// Real player.ExtractEpisodeNumber should pull "7" from this.
			return "https://x/ep7", "Episode 7", nil
		},
		nil, // keep production extractor
	)

	url, numStr, epNum, err := SelectInitialEpisode(sampleEpisodes(10))
	require.NoError(t, err)
	assert.Equal(t, "https://x/ep7", url)
	assert.Equal(t, "Episode 7", numStr)
	assert.Equal(t, 7, epNum)
}

// TestSelectInitialEpisode_RealExtractor_MovieFallback verifies the real
// extractor's fallback to "1" for movie/OVA labels reaches our caller.
func TestSelectInitialEpisode_RealExtractor_MovieFallback(t *testing.T) {
	withSelector(t,
		func([]models.Episode) (string, string, error) {
			return "https://x/movie", "Filme", nil
		},
		nil,
	)

	url, numStr, epNum, err := SelectInitialEpisode(sampleEpisodes(1))
	require.NoError(t, err)
	assert.Equal(t, "https://x/movie", url)
	assert.Equal(t, "Filme", numStr)
	assert.Equal(t, 1, epNum, "movie label must fall back to episode 1")
}

// ---------------------------------------------------------------------------
// Defaults are wired to the production functions.
// ---------------------------------------------------------------------------

func TestPackageDefaults_PointAtProductionFunctions(t *testing.T) {
	// Sequential — reads package-level vars also written by withSelector tests.
	selectorOverrideMu.Lock()
	defer selectorOverrideMu.Unlock()

	require.NotNil(t, selectEpisodeFunc)
	require.NotNil(t, extractEpisodeNumberFunc)
	assert.Equal(t, "5", extractEpisodeNumberFunc("Episode 5"))
	assert.Equal(t, "1", extractEpisodeNumberFunc("OVA Special"))
}
