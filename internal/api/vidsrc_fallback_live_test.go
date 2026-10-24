package api

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/pobreflix"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/topcine"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/require"
)

// Todo Poderoso (2003) is the user's report: the panel record points at five
// file-host domains that no longer answer (checked 2026-10-17), and playback
// ended as "servers GoAnime can't play yet". These live tests drive the two
// ways the user reaches the movie, and are the standing regression for the
// case. Both need the vidsrc fallback gate, so they flip it on explicitly.
// Opt in with GOANIME_LIVE=1; never runs in CI.

func liveFallbackSetup(t *testing.T) context.Context {
	t.Helper()
	if os.Getenv("GOANIME_LIVE") != "1" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	t.Setenv("TMDB_API_KEY", "")
	t.Setenv("OMDB_API_KEY", "")
	t.Setenv("GOANIME_VIDSRC_FALLBACK_TEST", "1")
	util.InitLogger()
	if os.Getenv("GOANIME_LIVE_DEBUG") != "" {
		util.IsDebug = true
	}
	prevPick, prevSwap := sfxPickFn, sfxAudioSwapFn
	sfxPickFn = func(string, []string) (int, error) { return 0, nil } // Dublado
	// The fallback is the original audio: take it, and say what was asked.
	sfxAudioSwapFn = func(swap audioSwap) (bool, error) {
		t.Logf("asked: %s instead of %s — %s", swap.Have, swap.Want, strings.Join(swap.Why, " / "))
		return true, nil
	}
	t.Cleanup(func() { sfxPickFn, sfxAudioSwapFn = prevPick, prevSwap })
	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	t.Cleanup(cancel)
	return ctx
}

// TestLiveTopCineTodoPoderosoPlaysViaFallback is the user's exact path: the
// TopCine search lists the movie, playback goes to the StartFlix panel, the
// panel's hosts are dead, and the next thing failure must be the fallback's
// master — the movie plays.
func TestLiveTopCineTodoPoderosoPlaysViaFallback(t *testing.T) {
	ctx := liveFallbackSetup(t)

	results, err := topcine.Shared().Search(ctx, "todo poderoso")
	require.NoError(t, err, "search")

	var anime *models.Anime
	for _, r := range results {
		t.Logf("result: %s (%s) %s", r.Title, r.Year, r.URL)
		// The original is the 2003 one; "A Volta ..." (2007) and "Evan ..."
		// (2007) must not win the match.
		if r.Kind == topcine.KindMovie && strings.EqualFold(r.Title, "Todo Poderoso") && r.Year == "2003" {
			anime = r.ToAnimeModel()
			break
		}
	}
	require.NotNil(t, anime, "the 2003 movie is not in the TopCine catalog")

	eps, err := GetTopCineEpisodes(anime)
	require.NoError(t, err, "listing")
	require.Len(t, eps, 1)
	t.Logf("anime: tmdb %d imdb %q, first %q", anime.TMDBID, anime.IMDBID, eps[0].URL)

	streamURL, err := GetTopCineStreamURL(anime, &eps[0], "best")
	require.NoError(t, err, "the fallback must save the title its panel could not")

	require.True(t, strings.HasPrefix(streamURL, "http"), "URL shape: %s", head(streamURL, 100))
	t.Logf("stream: %s", head(streamURL, 100))
	// The fallback has no Portuguese track: mpv must not be told otherwise.
	require.Equal(t, originalALang, util.GetGlobalAudioLanguage())
}

// TestLivePobreflixTodoPoderosoPlays is the user's original report, re-anchored
// on the live catalog. The old .sale listing page is gone; current pages embed
// a direct MP4, so playback no longer depends on the dead StartFlix panel
// hosts. The panel/vidsrc fallback keeps unit coverage for old-theme pages.
func TestLivePobreflixTodoPoderosoPlays(t *testing.T) {
	ctx := liveFallbackSetup(t)

	results, err := pobreflix.Shared().Search(ctx, "todo poderoso")
	require.NoError(t, err, "search")

	var anime *models.Anime
	for _, r := range results {
		t.Logf("result: %s (%s) %s", r.Title, r.Kind, r.URL)
		if r.Kind == pobreflix.KindMovie && r.Title == "Todo Poderoso" {
			anime = r.ToAnimeModel()
			break
		}
	}
	require.NotNil(t, anime, "Todo Poderoso is not in the live catalog")

	eps, err := GetPobreflixEpisodes(anime)
	require.NoError(t, err, "listing")
	require.Len(t, eps, 1)
	t.Logf("anime: tmdb %d imdb %q, first %q", anime.TMDBID, anime.IMDBID, eps[0].URL)

	streamURL, err := GetPobreflixStreamURL(anime, &eps[0], "best")
	require.NoError(t, err, "the fallback must save the title its panel could not")
	require.True(t, strings.HasPrefix(streamURL, "http"), "URL shape: %s", head(streamURL, 100))
	t.Logf("stream: %s", head(streamURL, 100))
}

func head(s string, limit int) string {
	if len(s) > limit {
		return s[:limit]
	}
	return s
}
