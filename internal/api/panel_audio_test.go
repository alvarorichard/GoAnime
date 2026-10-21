package api

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The audio a panel title plays in, beyond Todo Poderoso: a movie panel mixes
// "Dublado" and "Legendado" players on one list (24 of 24 live movies sampled
// on 2026-10-21), and a series episode's players can fail over to the
// original-audio fallback. Whatever plays, it is what the user asked for, or
// they were asked first.

// panelPlayer renders one movie-panel button the way the live panel does.
func panelPlayer(src, label string) string {
	return fmt.Sprintf(`<button class="btn" data-show-player="true" data-source=%q data-subtitles="" data-type="jwplayer" data-id="1">&nbsp; %s </button>`, src, label)
}

// useMoviePanel serves a StartFlix movie whose panel lists players, with
// direct files on files.test that play unless their name says "dead".
func useMoviePanel(t *testing.T, players ...string) *models.Anime {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host, p := r.Header.Get("X-Original-Host"), r.URL.Path
		switch {
		case host == "www.startflix.test" && p == "/filmes/mixed/":
			_, _ = w.Write([]byte(`<html><body><iframe src="https://painel.test/filme/tt200"></iframe></body></html>`))
		case host == "painel.test" && p == "/filme/tt200":
			_, _ = w.Write([]byte(`<div id="players">` + strings.Join(players, "\n") + `</div>`))
		case host == "files.test" && !strings.Contains(p, "dead"):
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusPartialContent)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	useFakeStartFlix(t, 0, 0)
	usePanelTransport(t, srv)
	return &models.Anime{Name: "Mixed", URL: "https://www.startflix.test/filmes/mixed/", Source: "StartFlix", MediaType: models.MediaTypeMovie}
}

func playMovie(t *testing.T, media *models.Anime) (string, error) {
	t.Helper()
	eps, err := GetStartFlixEpisodes(media)
	require.NoError(t, err)
	require.Len(t, eps, 1)
	return GetStartFlixStreamURL(media, &eps[0], "best")
}

// TestPanelMovieMixedAudioPlaysTheDub: the Legendado player is listed first
// and plays; the dub plays too, and it is the one the user gets.
func TestPanelMovieMixedAudioPlaysTheDub(t *testing.T) {
	media := useMoviePanel(t,
		panelPlayer("https://files.test/leg.mp4", "Legendado"),
		panelPlayer("https://files.test/dub.mp4", "1 Dublado"),
	)
	asked := recordSwaps(false, nil)

	got, err := playMovie(t, media)
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/dub.mp4", got)
	assert.Empty(t, *asked)
	assert.Equal(t, portugueseALang, util.GetGlobalAudioLanguage())
}

// TestPanelMovieDeadDubAsksBeforeLegendado: the panel offers the dub but its
// server is dead; the Legendado that plays is offered, never slipped in.
func TestPanelMovieDeadDubAsksBeforeLegendado(t *testing.T) {
	media := useMoviePanel(t,
		panelPlayer("https://files.test/dead-dub.mp4", "1 Dublado"),
		panelPlayer("https://files.test/leg.mp4", "Legendado"),
	)
	asked := recordSwaps(false, nil)

	_, err := playMovie(t, media)
	require.ErrorIs(t, err, ErrAudioUnavailable)
	require.Len(t, *asked, 1)
	assert.Equal(t, startflix.AudioDubbed, (*asked)[0].Want)
	assert.Equal(t, startflix.AudioSubtitled, (*asked)[0].Have)
	assert.Equal(t, "StartFlix", (*asked)[0].Source)

	*asked = nil
	sfxAudioSwapFn = func(swap audioSwap) (bool, error) { *asked = append(*asked, swap); return true, nil }
	got, err := playMovie(t, media)
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/leg.mp4", got)
	assert.Equal(t, originalALang, util.GetGlobalAudioLanguage())
}

// TestPanelMovieLegendadoOnlyPlaysUnasked: a movie the panel has only
// Legendado never had a dub to lose — like a single-audio season, it plays
// without a question, and mpv is told it is the original audio.
func TestPanelMovieLegendadoOnlyPlaysUnasked(t *testing.T) {
	media := useMoviePanel(t,
		panelPlayer("https://files.test/leg1.mp4", "1 Legendado"),
		panelPlayer("https://files.test/leg2.mp4", "2 Legendado"),
	)
	asked := recordSwaps(false, nil)

	got, err := playMovie(t, media)
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/leg1.mp4", got)
	assert.Empty(t, *asked)
	assert.Equal(t, originalALang, util.GetGlobalAudioLanguage(), "not Portuguese: there is no dub in it")
}

// TestPanelSeriesFallbackAsksBeforeOriginalAudio: the user watches the
// Dublado list; the episode's servers fail and the fallback answers with the
// original audio. The seam's stream does not say its audio — the api takes
// the fallback for the original audio regardless.
func TestPanelSeriesFallbackAsksBeforeOriginalAudio(t *testing.T) {
	useFakeStartFlix(t, 0, 0) // season 1, Dublado
	media := &models.Anime{Name: "Show", URL: "https://www.startflix.test/series/show/", Source: "StartFlix", MediaType: models.MediaTypeTV}
	eps, err := GetStartFlixEpisodes(media)
	require.NoError(t, err)
	require.NotEmpty(t, eps)

	prev := sfxVIDSRCFallbackFn
	sfxVIDSRCFallbackFn = func(context.Context, *models.Anime, *models.Episode) (*startflix.Stream, error) {
		return &startflix.Stream{URL: "https://vidsrc.test/master.m3u8", Host: "vidsrc.test"}, nil
	}
	t.Cleanup(func() { sfxVIDSRCFallbackFn = prev })
	asked := recordSwaps(false, nil)

	// An episode whose panel entry is gone: the panel fails, the fallback
	// answers.
	dead := eps[0]
	dead.URL = "https://painel.test/episodio/gone"
	_, err = GetStartFlixStreamURL(media, &dead, "best")
	require.ErrorIs(t, err, ErrAudioUnavailable)
	require.Len(t, *asked, 1)
	assert.Equal(t, startflix.AudioDubbed, (*asked)[0].Want, "the list the episode came from")
	assert.Equal(t, startflix.AudioSubtitled, (*asked)[0].Have)

	// The live episode itself still plays its dub, unasked.
	got, err := GetStartFlixStreamURL(media, &eps[0], "best")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/ep-1001.mp4", got)
	assert.Len(t, *asked, 1)
	assert.Equal(t, portugueseALang, util.GetGlobalAudioLanguage())
}

// TestWantedStreamAudioFollowsTheListedSeason: a Legendado-only season is
// taken without a picker, and its episodes are known to be Legendado anyway.
func TestWantedStreamAudioFollowsTheListedSeason(t *testing.T) {
	useFakeStartFlix(t, 0, 0)
	media := &models.Anime{Name: "Show", URL: "https://www.startflix.test/series/sub/", MediaType: models.MediaTypeTV}
	subOnly := startflix.Season{Number: 3, Subtitled: []startflix.Episode{{Number: 1, ID: "9"}}}

	eps := startFlixSeasonEpisodes(media, startflix.SeriesPanel(1), subOnly)
	require.Len(t, eps, 1)
	assert.Equal(t, startflix.AudioSubtitled, wantedStreamAudio(media, &eps[0]))

	media.MediaType = models.MediaTypeMovie
	assert.Equal(t, startflix.AudioDubbed, wantedStreamAudio(media, &eps[0]), "a movie is played dubbed unless pinned")
}
