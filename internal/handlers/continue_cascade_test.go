package handlers

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/appflow"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/playback"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// continueCascade stubs the whole reopen flow: scripted search, episodes and
// dispatch. Serial: it swaps package seams and one-shot globals.
type continueCascade struct {
	searchTitle string
	searchSrc   string
	seriesCalls int
	movieCalls  int
	preSelect   int
	fallbackTo  string
	resumePos   int
	resumeStart bool
	resumeSet   bool
}

func stubContinueFlow(t *testing.T, anime *models.Anime, episodes []models.Episode, seriesErr, movieErr error) *continueCascade {
	t.Helper()
	c := &continueCascade{}

	prevSearch, prevDetails, prevEps := searchAnimeAutoFn, fetchAnimeDetailsFn, getAnimeEpisodesFn
	prevSeries, prevMovie, prevPlayback := handleSeriesFn, handleMovieFn, handlePlaybackModeFn
	t.Cleanup(func() {
		searchAnimeAutoFn, fetchAnimeDetailsFn, getAnimeEpisodesFn = prevSearch, prevDetails, prevEps
		handleSeriesFn, handleMovieFn, handlePlaybackModeFn = prevSeries, prevMovie, prevPlayback
		clearContinueOverrides()
		player.ClearEpisodeMenu()
	})

	searchAnimeAutoFn = func(title, src string, pick func([]*models.Anime) *models.Anime) (*models.Anime, error) {
		c.searchTitle, c.searchSrc = title, src
		if pick != nil {
			// Exercise the pick wiring with a same-title rival.
			rival := &models.Anime{Name: title, Source: "Rival", URL: "https://rival.test/x"}
			chosen := pick([]*models.Anime{rival, anime})
			if chosen != rival && chosen != anime {
				t.Errorf("pick returned an unknown result")
			}
		}
		return anime, nil
	}
	fetchAnimeDetailsFn = func(*models.Anime) {}
	getAnimeEpisodesFn = func(*models.Anime) ([]models.Episode, error) { return episodes, nil }
	handleSeriesFn = func(_ context.Context, _ *models.Anime, _ []models.Episode, _ int, _ bool) error {
		c.seriesCalls++
		c.preSelect = playback.PreselectEpisodeNum
		c.resumePos, c.resumeStart, c.resumeSet = player.PeekResumeOverride()
		return seriesErr
	}
	handleMovieFn = func(_ context.Context, _ *models.Anime, _ []models.Episode, _ bool) error {
		c.movieCalls++
		c.preSelect = playback.PreselectEpisodeNum
		c.resumePos, c.resumeStart, c.resumeSet = player.PeekResumeOverride()
		return movieErr
	}
	handlePlaybackModeFn = func(name string) { c.fallbackTo = name }
	return c
}

func cascadeEpisodes(n int) []models.Episode {
	eps := make([]models.Episode, 0, n)
	for i := 1; i <= n; i++ {
		eps = append(eps, models.Episode{Number: string(rune('0' + i)), Num: i, URL: "https://cdn.test/e"})
	}
	return eps
}

func TestOpenContinue_ResumeMidEpisode(t *testing.T) {
	anime := &models.Anime{Name: "Mukashi Banashi", Source: "Animefire.io", MediaType: models.MediaTypeAnime, URL: "https://animefire.io/x"}
	c := stubContinueFlow(t, anime, cascadeEpisodes(12), nil, nil)

	entry := &tui.HomeContinueEntry{Key: "k", AnimeTitle: "Mukashi Banashi", EpisodeTitle: "Urashima Tarou",
		EpisodeNum: 5, Position: 372, Duration: 580, Source: "Animefire.io", Audio: "Legendado", MediaType: "anime"}
	require.NoError(t, openContinue(SearchHomeDecision{Entry: entry}))

	assert.Equal(t, "Mukashi Banashi", c.searchTitle)
	assert.Equal(t, "animefire", c.searchSrc)
	assert.Equal(t, 1, c.seriesCalls)
	assert.Equal(t, 0, c.movieCalls)
	assert.Equal(t, 5, c.preSelect, "the tracked episode is preselected")
	assert.True(t, c.resumeSet, "resume override reaches playback")
	assert.False(t, c.resumeStart)
	assert.Equal(t, 372, c.resumePos)
	assert.Empty(t, c.fallbackTo)
}

func TestOpenContinue_FinishedEpisodeOffersNext(t *testing.T) {
	anime := &models.Anime{Name: "Show", Source: "Animefire.io", MediaType: models.MediaTypeAnime}
	c := stubContinueFlow(t, anime, cascadeEpisodes(12), nil, nil)

	entry := &tui.HomeContinueEntry{Key: "k", AnimeTitle: "Show", EpisodeNum: 5, Position: 590, Duration: 600, Source: "Animefire.io", MediaType: "anime"}
	require.NoError(t, openContinue(SearchHomeDecision{Entry: entry}))

	assert.Equal(t, 6, c.preSelect, "finished reopens the next episode, not the credits")
	assert.False(t, c.resumeSet, "no stale position forced")
}

func TestOpenContinue_FinishedMovieReplays(t *testing.T) {
	anime := &models.Anime{Name: "Dark", Source: "StartFlix", MediaType: models.MediaTypeMovie, URL: "https://startflix.test/dark"}
	c := stubContinueFlow(t, anime, cascadeEpisodes(1), nil, nil)

	entry := &tui.HomeContinueEntry{Key: "k", AnimeTitle: "Dark", EpisodeNum: 1, Position: 2900, Duration: 3000, Source: "StartFlix", MediaType: "movie"}
	require.NoError(t, openContinue(SearchHomeDecision{Entry: entry}))

	assert.Equal(t, 0, c.seriesCalls)
	assert.Equal(t, 1, c.movieCalls, "movies never route to the series flow")
	assert.Equal(t, 0, c.preSelect, "movies never preselect (HandleMovie reads none)")
	assert.True(t, c.resumeSet && c.resumeStart, "rewatch starts over without asking")
}

func TestOpenContinue_EpisodesOnlySkipsPreselect(t *testing.T) {
	anime := &models.Anime{Name: "Show", Source: "Animefire.io", MediaType: models.MediaTypeAnime}
	c := stubContinueFlow(t, anime, cascadeEpisodes(12), nil, nil)

	entry := &tui.HomeContinueEntry{Key: "k", AnimeTitle: "Show", EpisodeNum: 5, Position: 372, Duration: 580, MediaType: "anime"}
	require.NoError(t, openContinue(SearchHomeDecision{Entry: entry, EpisodesOnly: true}))

	assert.Equal(t, 1, c.seriesCalls)
	assert.Equal(t, 0, c.preSelect, "the list opens unfiltered")
	assert.False(t, c.resumeSet)
}

func TestOpenContinue_FromStart(t *testing.T) {
	anime := &models.Anime{Name: "Show", Source: "Animefire.io", MediaType: models.MediaTypeAnime}
	c := stubContinueFlow(t, anime, cascadeEpisodes(12), nil, nil)

	entry := &tui.HomeContinueEntry{Key: "k", AnimeTitle: "Show", EpisodeNum: 5, Position: 372, Duration: 580, MediaType: "anime"}
	require.NoError(t, openContinue(SearchHomeDecision{Entry: entry, FromStart: true}))

	assert.Equal(t, 5, c.preSelect)
	assert.True(t, c.resumeSet && c.resumeStart)
}

func TestOpenContinue_SearchErrorFallsBack(t *testing.T) {
	prev := searchAnimeAutoFn
	searchAnimeAutoFn = func(string, string, func([]*models.Anime) *models.Anime) (*models.Anime, error) {
		return nil, errors.New("offline")
	}
	t.Cleanup(func() { searchAnimeAutoFn = prev })

	entry := &tui.HomeContinueEntry{Key: "k", AnimeTitle: "Show", EpisodeNum: 5, MediaType: "anime"}
	err := openContinue(SearchHomeDecision{Entry: entry})
	require.Error(t, err, "openContinue reports; HandleContinue falls back")
}

func TestOpenContinue_EmptyTitleErrors(t *testing.T) {
	err := openContinue(SearchHomeDecision{Entry: &tui.HomeContinueEntry{Key: "k"}})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no title")
}

func TestOpenContinue_MismatchFallsBack(t *testing.T) {
	// A tracked movie resolving to an anime must never play: error out so
	// the caller falls back to the results screen.
	anime := &models.Anime{Name: "Dark", Source: "Animefire.io", MediaType: models.MediaTypeAnime, URL: "https://animefire.io/dark"}
	prev := searchAnimeAutoFn
	searchAnimeAutoFn = func(title, src string, pick func([]*models.Anime) *models.Anime) (*models.Anime, error) {
		assert.Equal(t, "Dark", title)
		return anime, nil
	}
	t.Cleanup(func() { searchAnimeAutoFn = prev; clearContinueOverrides() })

	entry := &tui.HomeContinueEntry{Key: "k", AnimeTitle: "Dark", EpisodeNum: 1, Position: 100, Duration: 3000, Source: "StartFlix", MediaType: "movie"}
	err := openContinue(SearchHomeDecision{Entry: entry})
	require.Error(t, err)
	assert.Contains(t, err.Error(), "anime title")
	assert.Equal(t, 0, playback.PreselectEpisodeNum, "nothing preselected on mismatch")
}

func TestOpenContinue_SourcelessMovieSearchesCatalogs(t *testing.T) {
	anime := &models.Anime{Name: "Dark", Source: "StartFlix", MediaType: models.MediaTypeMovie, URL: "https://startflix.test/dark"}
	c := stubContinueFlow(t, anime, cascadeEpisodes(1), nil, nil)

	var gotSrc string
	prev := searchAnimeAutoFn
	searchAnimeAutoFn = func(title, src string, pick func([]*models.Anime) *models.Anime) (*models.Anime, error) {
		gotSrc = src
		return anime, nil
	}
	t.Cleanup(func() { searchAnimeAutoFn = prev })

	entry := &tui.HomeContinueEntry{Key: "k", AnimeTitle: "Dark", EpisodeNum: 1, Position: 100, Duration: 3000, MediaType: "movie"}
	require.NoError(t, openContinue(SearchHomeDecision{Entry: entry}))
	assert.Equal(t, "movietv", gotSrc, "sourceless movies stay in the movie catalogs")
	_ = c
}

func TestCleanContinueTitle(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Frieren", cleanContinueTitle(&tui.HomeContinueEntry{AnimeTitle: "[PT-BR] Frieren"}))
	assert.Equal(t, "Black Clover", cleanContinueTitle(&tui.HomeContinueEntry{AnimeTitle: "Black Clover (Dublado)"}))
	assert.Equal(t, "Naruto", cleanContinueTitle(&tui.HomeContinueEntry{EpisodeTitle: "Naruto"}))
	assert.Equal(t, "", cleanContinueTitle(&tui.HomeContinueEntry{}))
}

func TestOpenContinue_BackRoutesToSearch(t *testing.T) {
	anime := &models.Anime{Name: "Show", Source: "Animefire.io", MediaType: models.MediaTypeAnime}
	c := stubContinueFlow(t, anime, cascadeEpisodes(12), player.ErrBackToAnimeSelection, nil)

	entry := &tui.HomeContinueEntry{Key: "k", AnimeTitle: "Show", EpisodeNum: 5, Position: 10, Duration: 600, MediaType: "anime"}
	require.NoError(t, openContinue(SearchHomeDecision{Entry: entry}))
	assert.Equal(t, "Show", c.fallbackTo, "back from playback returns to search")
	assert.Equal(t, 0, playback.PreselectEpisodeNum, "one-shot state cleared")
}

func TestHandleContinue_FailureSearchesTitle(t *testing.T) {
	prevSearch, prevPlayback := searchAnimeAutoFn, handlePlaybackModeFn
	t.Cleanup(func() { searchAnimeAutoFn, handlePlaybackModeFn = prevSearch, prevPlayback })
	searchAnimeAutoFn = func(string, string, func([]*models.Anime) *models.Anime) (*models.Anime, error) {
		return nil, errors.New("offline")
	}
	var fallback string
	handlePlaybackModeFn = func(name string) { fallback = name }

	entry := &tui.HomeContinueEntry{Key: "k", AnimeTitle: "Show", EpisodeNum: 5, MediaType: "anime"}
	assert.NotPanics(t, func() { HandleContinue(SearchHomeDecision{Entry: entry, Title: "Show"}) })
	assert.Equal(t, "Show", fallback)
}

func TestDrainPendingHome(t *testing.T) {
	t.Cleanup(func() { appflow.PendingHomeResult = nil })
	appflow.PendingHomeResult = nil

	t.Run("empty drains nothing", func(t *testing.T) {
		assert.False(t, drainPendingHome())
	})

	t.Run("pending reopens and clears", func(t *testing.T) {
		anime := &models.Anime{Name: "Show", Source: "Animefire.io", MediaType: models.MediaTypeAnime}
		c := stubContinueFlow(t, anime, cascadeEpisodes(12), nil, nil)

		entry := tui.HomeContinueEntry{Key: "k", AnimeTitle: "Show", EpisodeNum: 5, Position: 10, Duration: 600, MediaType: "anime"}
		appflow.PendingHomeResult = &tui.SearchHomeResult{Action: tui.HomeResume, Entry: entry, HasEntry: true}

		assert.True(t, drainPendingHome())
		assert.Nil(t, appflow.PendingHomeResult, "consumed once")
		assert.Equal(t, 1, c.seriesCalls)
		assert.Equal(t, 5, c.preSelect)
	})
}
