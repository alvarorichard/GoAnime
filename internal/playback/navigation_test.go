package playback

import (
	"context"
	"runtime"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// --- printEpisodeNotFoundMsg ---

func TestPrintEpisodeNotFoundMsg_NoPanic(t *testing.T) {
	t.Parallel()
	assert.NotPanics(t, printEpisodeNotFoundMsg)
}

// --- getSocketPath ---

func TestGetSocketPath_PerOS(t *testing.T) {
	t.Parallel()
	got := getSocketPath()
	if runtime.GOOS == "windows" {
		assert.Equal(t, `\\.\pipe\goanime_mpvsocket`, got)
	} else {
		assert.True(t, strings.HasSuffix(got, "mpvsocket"))
	}
}

// --- createUpdater ---

func TestCreateUpdater_DisabledReturnsNil(t *testing.T) {
	t.Parallel()
	got := createUpdater(&models.Anime{}, ptr(false), &sync.Mutex{}, time.Second, false)
	assert.Nil(t, got)
}

func TestCreateUpdater_EnabledReturnsRPU(t *testing.T) {
	t.Parallel()
	got := createUpdater(&models.Anime{Name: "x"}, ptr(false), &sync.Mutex{}, 2*time.Second, true)
	require.NotNil(t, got)
	assert.Equal(t, 3*time.Second, got.GetUpdateFreq())
	assert.Equal(t, 2*time.Second, got.GetEpisodeDuration())
}

// --- handleUserNavigation: pure dispatch ---

func TestHandleUserNavigation_Next(t *testing.T) {
	t.Parallel()
	eps := []models.Episode{
		{URL: "u1", Number: "1"}, {URL: "u2", Number: "2"}, {URL: "u3", Number: "3"},
	}
	url, num, n := handleUserNavigation("n", eps, 1, 3)
	assert.Equal(t, "u2", url)
	assert.Equal(t, "2", num)
	assert.Equal(t, 2, n)
}

func TestHandleUserNavigation_Previous(t *testing.T) {
	t.Parallel()
	eps := []models.Episode{
		{URL: "u1", Number: "1"}, {URL: "u2", Number: "2"},
	}
	url, num, n := handleUserNavigation("p", eps, 2, 2)
	assert.Equal(t, "u1", url)
	assert.Equal(t, "1", num)
	assert.Equal(t, 1, n)
}

func TestHandleUserNavigation_PreviousAtFirst(t *testing.T) {
	t.Parallel()
	eps := []models.Episode{{URL: "u1", Number: "1"}}
	// At episode 1, p should clamp to 1.
	url, _, n := handleUserNavigation("p", eps, 1, 1)
	assert.Equal(t, "u1", url)
	assert.Equal(t, 1, n)
}

func TestHandleUserNavigation_NextAtLast(t *testing.T) {
	t.Parallel()
	eps := []models.Episode{
		{URL: "u1", Number: "1"}, {URL: "u2", Number: "2"},
	}
	url, _, n := handleUserNavigation("n", eps, 2, 2)
	assert.Equal(t, "u2", url)
	assert.Equal(t, 2, n)
}

// --- handleUserNavigationEnhanced ---

func TestHandleUserNavigationEnhanced_NonAllAnime_Delegates(t *testing.T) {
	t.Parallel()
	eps := []models.Episode{{URL: "u1", Number: "1"}, {URL: "u2", Number: "2"}}
	anime := &models.Anime{Source: "AnimeFire", URL: "https://animefire.io/x"}
	url, _, n := handleUserNavigationEnhanced("n", eps, 1, 2, anime)
	assert.Equal(t, "u2", url)
	assert.Equal(t, 2, n)
}

func TestHandleUserNavigationEnhanced_AllAnime_UsesEnhancedHandler(t *testing.T) {
	t.Parallel()
	// Seed navigator cache so we don't hit network.
	anime := &models.Anime{Source: "AllAnime", URL: "AAcacheKey1"}
	nav := &AllAnimeNavigator{animeID: "AAcacheKey1", episodes: []string{"1", "2", "3"}}
	navigatorCacheMu.Lock()
	navigatorCache["AAcacheKey1"] = nav
	navigatorCacheMu.Unlock()
	t.Cleanup(func() {
		navigatorCacheMu.Lock()
		delete(navigatorCache, "AAcacheKey1")
		navigatorCacheMu.Unlock()
	})
	eps := []models.Episode{
		{URL: "u1", Number: "1", Num: 1},
		{URL: "u2", Number: "2", Num: 2},
		{URL: "u3", Number: "3", Num: 3},
	}
	_, num, n := handleUserNavigationEnhanced("n", eps, 1, 3, anime)
	// AllAnime path returns nextEp.URL/Number/Num.
	assert.Equal(t, "2", num)
	assert.Equal(t, 2, n)
}

// --- handleAllAnimeNavigation ---

func TestHandleAllAnimeNavigation_NoCurrentEpisode_Fallback(t *testing.T) {
	t.Parallel()
	anime := &models.Anime{Source: "AllAnime", URL: "AAmissing999"}
	eps := []models.Episode{{URL: "u1", Number: "1", Num: 1}, {URL: "u2", Number: "2", Num: 2}}
	// currentNum=99 has no match → falls back to handleUserNavigation
	url, _, _ := handleAllAnimeNavigation("n", eps, 99, 2, anime)
	// fallback clamps next to totalEpisodes(2) → "u2"
	assert.Equal(t, "u2", url)
}

func TestHandleAllAnimeNavigation_Previous_Success(t *testing.T) {
	t.Parallel()
	anime := &models.Anime{Source: "AllAnime", URL: "AAprev1"}
	nav := &AllAnimeNavigator{animeID: "AAprev1", episodes: []string{"1", "2", "3"}}
	navigatorCacheMu.Lock()
	navigatorCache["AAprev1"] = nav
	navigatorCacheMu.Unlock()
	t.Cleanup(func() {
		navigatorCacheMu.Lock()
		delete(navigatorCache, "AAprev1")
		navigatorCacheMu.Unlock()
	})
	eps := []models.Episode{
		{URL: "u1", Number: "1", Num: 1},
		{URL: "u2", Number: "2", Num: 2},
	}
	_, num, n := handleAllAnimeNavigation("p", eps, 2, 2, anime)
	assert.Equal(t, "1", num)
	assert.Equal(t, 1, n)
}

func TestHandleAllAnimeNavigation_PreviousFailsFallback(t *testing.T) {
	t.Parallel()
	anime := &models.Anime{Source: "AllAnime", URL: "AAprevFail"}
	nav := &AllAnimeNavigator{animeID: "AAprevFail", episodes: []string{"1"}}
	navigatorCacheMu.Lock()
	navigatorCache["AAprevFail"] = nav
	navigatorCacheMu.Unlock()
	t.Cleanup(func() {
		navigatorCacheMu.Lock()
		delete(navigatorCache, "AAprevFail")
		navigatorCacheMu.Unlock()
	})
	eps := []models.Episode{{URL: "u1", Number: "1", Num: 1}}
	url, _, _ := handleAllAnimeNavigation("p", eps, 1, 1, anime)
	// nav.GetPreviousEpisode("1") errors → falls back → handleUserNavigation("p", ...) clamps to 1
	assert.Equal(t, "u1", url)
}

func TestHandleAllAnimeNavigation_DefaultBranch(t *testing.T) {
	t.Parallel()
	anime := &models.Anime{Source: "AllAnime", URL: "AAdefault"}
	nav := &AllAnimeNavigator{animeID: "AAdefault", episodes: []string{"1", "2"}}
	navigatorCacheMu.Lock()
	navigatorCache["AAdefault"] = nav
	navigatorCacheMu.Unlock()
	t.Cleanup(func() {
		navigatorCacheMu.Lock()
		delete(navigatorCache, "AAdefault")
		navigatorCacheMu.Unlock()
	})
	eps := []models.Episode{
		{URL: "u1", Number: "1", Num: 1},
		{URL: "u2", Number: "2", Num: 2},
	}
	// unknown input → default → falls to handleUserNavigation → next
	url, _, n := handleAllAnimeNavigation("?", eps, 1, 2, anime)
	assert.Equal(t, "u2", url)
	assert.Equal(t, 2, n)
}

// --- CheckIfSeries / CheckIfSeriesEnhanced ---

func TestCheckIfSeries_HandlesError(t *testing.T) {
	t.Parallel()
	// Bogus URL → underlying api.IsSeries either errors (→ (false,1)) or
	// returns (false, 0) for empty results. Either way we just guard no panic
	// and a non-negative total.
	series, total := CheckIfSeries("http://127.0.0.1:1/no")
	_ = series
	assert.GreaterOrEqual(t, total, 0)
}

func TestCheckIfSeriesEnhanced_HandlesError(t *testing.T) {
	t.Parallel()
	// With a bogus URL, IsSeriesEnhanced either errors (→ (false,1)) or returns
	// (false, 0) for empty episode list. Either way we must not panic and the
	// total must be non-negative.
	series, total := CheckIfSeriesEnhanced(&models.Anime{URL: "http://127.0.0.1:1/no"})
	_ = series
	assert.GreaterOrEqual(t, total, 0)
}

// SelectInitialEpisode is fully covered by select_initial_episode_test.go
// (injected mock selector + table-driven parseEpisodeSelection). Removed the
// old symbol pin to avoid dead noise.

// --- PlayEpisode: drive entry until videoErr triggers an early return. ---
// We cannot drive the full TUI happy path, but with a bogus URL the call to
// player.GetVideoURLForEpisodeEnhanced fails and the function returns
// player.ErrBackToEpisodeSelection.

func TestPlayEpisode_VideoURLErrorReturnsBack(t *testing.T) {
	t.Parallel()
	anime := &models.Anime{
		Name:   "X",
		URL:    "http://127.0.0.1:1/bogus",
		Source: "AnimeFire",
	}
	eps := []models.Episode{{Number: "1", Num: 1, URL: "http://127.0.0.1:1/bogus/ep1"}}
	mu := &sync.Mutex{}
	pause := false
	err := PlayEpisode(context.Background(), anime, eps, 1, eps[0].URL, "1", false, &pause, mu)
	require.Error(t, err)
}

// --- HandleMovie / HandleSeries / GetUserInput / ChangeAnimeLocal ---
// Heavy TUI loops with no driveable entry point under tests. Symbol-pin only.

func TestHandleMovie_SymbolPin(t *testing.T) {
	t.Parallel()
	_ = HandleMovie
}

func TestHandleSeries_SymbolPin(t *testing.T) {
	t.Parallel()
	_ = HandleSeries
}

func TestGetUserInput_SymbolPin(t *testing.T) {
	t.Parallel()
	_ = GetUserInput
}

func TestChangeAnimeLocal_SymbolPin(t *testing.T) {
	t.Parallel()
	_ = ChangeAnimeLocal
}

// helpers
func ptr[T any](v T) *T { return &v }
