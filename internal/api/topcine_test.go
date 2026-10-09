package api

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/movie"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/topcine"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Like the StartFlix tests, these swap package seams and read playback
// globals, so none run in parallel.

// fakeTopCineSeriesPage names the series by TMDB id 100 — the fake StartFlix
// panel's series — and gives episodes the names the panel lacks.
const fakeTopCineSeriesPage = `<html><body>
<select data-lista-temporada="1">
	<option value="" selected disabled>Escolha</option>
	<option data-temporada="1" data-episodio="2" data-titulo="Segundo" data-player="https://azullog.test/serie/100/1/2">EP 02</option>
	<option data-temporada="1" data-episodio="1" data-titulo="Piloto" data-player="https://azullog.test/serie/100/1/1">EP 01</option>
</select>
<select data-lista-temporada="2">
	<option data-temporada="2" data-episodio="1" data-titulo="Episódio 1" data-player="https://azullog.test/serie/100/2/1">EP 01</option>
	<option data-temporada="2" data-episodio="2" data-titulo="Volta" data-player="https://azullog.test/serie/100/2/2">EP 02</option>
	<option data-temporada="2" data-episodio="3" data-titulo="" data-player="https://azullog.test/serie/100/2/3">EP 03</option>
</select></body></html>`

// fakeTopCineMissingPage is a series the panel does not carry.
const fakeTopCineMissingPage = `<html><body><select>
	<option data-temporada="1" data-episodio="1" data-titulo="Piloto" data-player="https://azullog.test/serie/999/1/1">EP 01</option>
</select></body></html>`

const fakeTopCineMoviePage = `<html><body><iframe src="https://azullog.test/filme/555"></iframe></body></html>`

// useFakeTopCine is useFakeStartFlix plus a fake TopCine site, with the panel
// built from TMDB ids pointed at the fake panel.
func useFakeTopCine(t *testing.T, seasonIdx, audioIdx int) {
	t.Helper()
	useFakeStartFlix(t, seasonIdx, audioIdx)
	t.Setenv("GOANIME_STARTFLIX_PANEL_URL", "https://painel.test")

	tc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/serie/show":
			_, _ = w.Write([]byte(fakeTopCineSeriesPage))
		case "/serie/missing":
			_, _ = w.Write([]byte(fakeTopCineMissingPage))
		case "/filme/film":
			_, _ = w.Write([]byte(fakeTopCineMoviePage))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(tc.Close)
	tcURL, _ := url.Parse(tc.URL)
	client := topcine.NewClientForTest(&http.Client{Transport: routeAll{tcURL}}, "https://topcine.test")

	prevClient, prevLangs := tcClientFn, tcLanguagesFn
	tcClientFn = func() *topcine.Client { return client }
	tcLanguagesFn = func(context.Context, string) (topcine.Languages, error) {
		t.Fatal("the player is only asked to explain a failure")
		return topcine.Languages{}, nil
	}
	t.Cleanup(func() { tcClientFn, tcLanguagesFn = prevClient, prevLangs })
}

func topCineSeries(slug string) *models.Anime {
	return &models.Anime{Name: "Show", URL: "https://topcine.test/serie/" + slug, Source: topcine.SourceName, MediaType: models.MediaTypeTV}
}

func TestGetTopCineEpisodes_SeriesThroughThePanel(t *testing.T) {
	useFakeTopCine(t, 0, 0) // season 1, Dublado
	var prompt string
	sfxPickFn = func(p string, _ []string) (int, error) { prompt = p; return 0, nil }
	media := topCineSeries("show")

	eps, err := GetTopCineEpisodes(media)
	require.NoError(t, err)
	require.Len(t, eps, 2)
	assert.Equal(t, "https://painel.test/episodio/1001", eps[0].URL, "episodes play from the panel the TMDB id names")
	assert.Equal(t, "https://painel.test/episodio/1002", eps[1].URL)
	assert.Equal(t, "Piloto", eps[0].Title.English, "the panel names no episode; TopCine's name fills in")
	assert.Equal(t, "Segundo", eps[1].Title.English)
	assert.Equal(t, 100, media.TMDBID)
	assert.Equal(t, 1, media.CurrentSeason)
	assert.True(t, strings.HasPrefix(prompt, "TopCine > "), "the audio picker is titled after TopCine, got %q", prompt)

	streamURL, err := GetTopCineStreamURL(media, &eps[0], "best")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/ep-1001.mp4", streamURL)
	assert.Equal(t, portugueseALang, util.GetGlobalAudioLanguage())
}

func TestGetTopCineEpisodes_PanelNameIsKept(t *testing.T) {
	useFakeTopCine(t, 0, 1) // season 1, Legendado: the panel names it "O Começo"
	eps, err := GetTopCineEpisodes(topCineSeries("show"))
	require.NoError(t, err)
	require.Len(t, eps, 1)
	assert.Equal(t, "O Começo", eps[0].Title.English)
}

func TestGetTopCineSeasonEpisodes_ForDownloads(t *testing.T) {
	useFakeTopCine(t, 0, 0)
	media := topCineSeries("show")

	seasons, err := GetTopCineSeasonNumbers(media)
	require.NoError(t, err)
	assert.Equal(t, []int{1, 2}, seasons)

	eps, err := GetTopCineSeasonEpisodes(media, 2)
	require.NoError(t, err)
	require.Len(t, eps, 3)
	assert.Equal(t, 2, media.CurrentSeason)
	assert.Empty(t, eps[0].Title.English, "a generic TopCine name is not used")
	assert.Equal(t, "Volta", eps[1].Title.English)
	assert.Empty(t, eps[2].Title.English)

	_, err = GetTopCineSeasonEpisodes(media, 7)
	require.ErrorContains(t, err, "season 7 is not on TopCine")
}

func TestGetTopCineEpisodes_MovieByIMDbID(t *testing.T) {
	useFakeTopCine(t, 0, 0)
	fake := &officialRecordFake{title: "Film", year: "2026"}
	sfxEnrichByIDFn = func(m *models.Media, isMovie bool) error {
		err := fake.enrich(m, isMovie)
		m.IMDBID = "tt100" // what the official record links TMDB 555 to
		return err
	}
	media := &models.Anime{Name: "Filme", URL: "https://topcine.test/filme/film", Source: topcine.SourceName, MediaType: models.MediaTypeMovie}

	eps, err := GetTopCineEpisodes(media)
	require.NoError(t, err)
	require.Len(t, eps, 1)
	assert.Equal(t, "https://painel.test/filme/tt100", eps[0].URL)
	assert.Equal(t, 555, fake.tmdbID, "the movie is looked up by the TMDB id TopCine names")
	assert.True(t, fake.movie)
	assert.Equal(t, "tt100", media.IMDBID)

	streamURL, err := GetTopCineStreamURL(media, &eps[0], "best")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/movie.mp4", streamURL)
}

func TestGetTopCineEpisodes_MovieWithoutIMDbID(t *testing.T) {
	useFakeTopCine(t, 0, 0)
	media := &models.Anime{Name: "Filme", URL: "https://topcine.test/filme/film", Source: topcine.SourceName, MediaType: models.MediaTypeMovie}

	_, err := GetTopCineEpisodes(media)
	require.ErrorIs(t, err, ErrTopCineNoIMDb)
	require.ErrorIs(t, err, movie.ErrNoOfficialRecord, "the lookup's own error stays reachable")
	assert.Contains(t, err.Error(), "couldn't link this TopCine movie")
}

func TestTopCinePanel_PageIDReplacesANameGuess(t *testing.T) {
	useFakeTopCine(t, 0, 0)
	media := topCineSeries("show")
	media.TMDBID, media.IMDBID, media.Overview, media.Rating = 7, "tt7", "another show", 9

	panel, err := topCinePanel(context.Background(), media)
	require.NoError(t, err)
	assert.Equal(t, startflix.SeriesPanel(100), panel)
	assert.Equal(t, 100, media.TMDBID)
	assert.Empty(t, media.IMDBID)
	assert.Empty(t, media.Overview, "what the wrong guess brought in is dropped")
	assert.Zero(t, media.Rating)
}

func TestGetTopCineEpisodes_NotOnThePanel(t *testing.T) {
	tests := []struct {
		name  string
		langs topcine.Languages
		err   error
		want  string
	}{
		{"only on the gated host", topcine.Languages{Dubbed: true, Hosts: []string{"1take.top"}}, nil, "only on 1take.top, which asks for a browser check"},
		{"no video at all", topcine.Languages{}, nil, "no video for this title yet"},
		{"player check fails", topcine.Languages{}, errors.New("timeout"), "a server that asks for a browser check"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			useFakeTopCine(t, 0, 0)
			var asked string
			tcLanguagesFn = func(_ context.Context, player string) (topcine.Languages, error) {
				asked = player
				return tt.langs, tt.err
			}
			_, err := GetTopCineEpisodes(topCineSeries("missing"))
			require.ErrorIs(t, err, startflix.ErrNotOnPanel)
			assert.Contains(t, err.Error(), tt.want)
			assert.Equal(t, "https://azullog.test/serie/999/1/1", asked)
		})
	}
}

func TestDescribeTopCineErr(t *testing.T) {
	media := topCineSeries("show")
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"no player", topcine.ErrNoPlayer, "no player for this title"},
		{"panel servers named after TopCine", &startflix.NoStreamError{Failures: []error{errors.New("x")}}, "None of TopCine's servers worked"},
		{"timeout", context.DeadlineExceeded, "TopCine took too long"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := describeTopCineErr(media, tt.err)
			assert.Contains(t, got.Error(), tt.want)
			assert.ErrorIs(t, got, tt.err)
		})
	}
	assert.NoError(t, describeTopCineErr(media, nil))
	plain := errors.New("plain")
	assert.Same(t, plain, describeTopCineErr(media, plain))
}

func TestTopCineAnyPlayer(t *testing.T) {
	title := topcine.Title{Seasons: []topcine.Season{
		{Number: 1, Episodes: []topcine.Episode{{PlayerURL: "s1"}}},
		{Number: 2, Episodes: []topcine.Episode{{PlayerURL: "s2"}}},
	}}
	assert.Equal(t, "s2", topCineAnyPlayer(title, 2))
	assert.Equal(t, "s1", topCineAnyPlayer(title, 0))
	assert.Equal(t, "s1", topCineAnyPlayer(title, 9), "an unknown season falls back to the first")
	assert.Equal(t, "m", topCineAnyPlayer(topcine.Title{PlayerURL: "m"}, 0))
	assert.Empty(t, topCineAnyPlayer(topcine.Title{}, 0))
}

func TestPanelSourceName(t *testing.T) {
	assert.Equal(t, "TopCine", panelSourceName(topCineSeries("x")))
	assert.Equal(t, "StartFlix", panelSourceName(&models.Anime{Source: "StartFlix"}))
	assert.Equal(t, "StartFlix", panelSourceName(nil))
}

func TestGetTopCineEpisodes_NoTitle(t *testing.T) {
	for _, media := range []*models.Anime{nil, {Name: "x", Source: topcine.SourceName}} {
		_, err := GetTopCineEpisodes(media)
		require.ErrorContains(t, err, "no TopCine page")
	}
}

func TestGetTopCineEpisodes_PageWithoutPlayer(t *testing.T) {
	useFakeTopCine(t, 0, 0)
	_, err := GetTopCineEpisodes(topCineSeries("gone")) // the fake site answers 404
	require.Error(t, err)

	media := topCineSeries("show")
	media.URL = "https://topcine.test/serie/empty"
	prev := tcClientFn
	tcClientFn = func() *topcine.Client {
		return topcine.NewClientForTest(&http.Client{Transport: staticBody(`<p>Em breve</p>`)}, "https://topcine.test")
	}
	t.Cleanup(func() { tcClientFn = prev })
	_, err = GetTopCineEpisodes(media)
	require.ErrorIs(t, err, topcine.ErrNoPlayer)
	assert.Contains(t, err.Error(), "no player for this title yet")
}

// staticBody answers every request with the same page.
type staticBody string

func (s staticBody) RoundTrip(r *http.Request) (*http.Response, error) {
	rec := httptest.NewRecorder()
	_, _ = rec.WriteString(string(s))
	resp := rec.Result()
	resp.Request = r
	return resp, nil
}

func TestTopCinePanel_MovieLookupFindsNoIMDbID(t *testing.T) {
	useFakeTopCine(t, 0, 0)
	sfxEnrichByIDFn = func(*models.Media, bool) error { return nil } // a record, but no IMDb id on it
	media := &models.Anime{Name: "Filme", URL: "https://topcine.test/filme/film", Source: topcine.SourceName}

	_, err := topCinePanel(context.Background(), media)
	require.ErrorIs(t, err, ErrTopCineNoIMDb)
	assert.Contains(t, err.Error(), "tmdb 555")
}

func TestTopCinePanel_KnownIMDbIDSkipsTheLookup(t *testing.T) {
	useFakeTopCine(t, 0, 0)
	sfxEnrichByIDFn = func(*models.Media, bool) error { t.Fatal("looked up an id already known"); return nil }
	media := &models.Anime{URL: "https://topcine.test/filme/film", Source: topcine.SourceName, TMDBID: 555, IMDBID: "tt100"}

	panel, err := topCinePanel(context.Background(), media)
	require.NoError(t, err)
	assert.Equal(t, startflix.MoviePanel("tt100"), panel)
}

func TestGetTopCineEpisodes_SeriesTaggedAsMovieBecomesTV(t *testing.T) {
	useFakeTopCine(t, 0, 0)
	media := topCineSeries("show")
	media.MediaType = models.MediaTypeMovie
	_, err := GetTopCineEpisodes(media)
	require.NoError(t, err)
	assert.Equal(t, models.MediaTypeTV, media.MediaType, "the panel is the authority on the kind")
}

func TestGetTopCineSeasonNumbers_Movie(t *testing.T) {
	useFakeTopCine(t, 0, 0)
	sfxEnrichByIDFn = func(m *models.Media, _ bool) error { m.IMDBID = "tt100"; return nil }
	_, err := GetTopCineSeasonNumbers(&models.Anime{URL: "https://topcine.test/filme/film", Source: topcine.SourceName})
	require.ErrorIs(t, err, ErrStartFlixNotSeries)
}

func TestTopCineUnplayableMessage_Fallbacks(t *testing.T) {
	const fallback = "a server that asks for a browser check"
	t.Run("title page fails", func(t *testing.T) {
		useFakeTopCine(t, 0, 0)
		assert.Contains(t, topCineUnplayableMessage(topCineSeries("gone")), fallback)
	})
	t.Run("page names no player", func(t *testing.T) {
		useFakeTopCine(t, 0, 0)
		prev := tcClientFn
		tcClientFn = func() *topcine.Client {
			return topcine.NewClientForTest(&http.Client{Transport: staticBody(`<p>x</p>`)}, "https://topcine.test")
		}
		t.Cleanup(func() { tcClientFn = prev })
		assert.Contains(t, topCineUnplayableMessage(topCineSeries("show")), fallback)
	})
	t.Run("tracks without a host", func(t *testing.T) {
		useFakeTopCine(t, 0, 0)
		tcLanguagesFn = func(context.Context, string) (topcine.Languages, error) {
			return topcine.Languages{Subtitled: true}, nil
		}
		assert.Contains(t, topCineUnplayableMessage(topCineSeries("show")), fallback)
	})
}

func TestTcLanguagesFnAsksTheTopCineClient(t *testing.T) {
	var asked string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		asked = r.URL.Path + "?" + r.URL.Query().Get("acao")
		_, _ = w.Write([]byte(`{"sucesso":true,"tem_dublado":true,"url_dublado":"https://1take.top/e/x"}`))
	}))
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	prevClient, prevLangs := tcClientFn, tcLanguagesFn
	t.Cleanup(func() { tcClientFn, tcLanguagesFn = prevClient, prevLangs })
	client := topcine.NewClientForTest(&http.Client{Transport: routeAll{target}}, "https://topcine.test")
	tcClientFn = func() *topcine.Client { return client }

	langs, err := prevLangs(context.Background(), "https://azullog.test/filme/1")
	require.NoError(t, err)
	assert.Equal(t, "/filme/1?carregar_idiomas", asked)
	assert.Equal(t, []string{"1take.top"}, langs.Hosts)
}

func TestFillTopCineEpisodeTitles(t *testing.T) {
	useFakeTopCine(t, 0, 0)
	media := topCineSeries("show")
	media.CurrentSeason = 1

	fillTopCineEpisodeTitles(media, nil) // nothing to fill: no request either

	eps := []models.Episode{
		{Num: 1},
		{Num: 2, Title: models.TitleDetails{English: "Kept"}},
		{Num: 9},
	}
	fillTopCineEpisodeTitles(media, eps)
	assert.Equal(t, "Piloto", eps[0].Title.English)
	assert.Equal(t, "Piloto", eps[0].Title.Romaji)
	assert.Equal(t, "Kept", eps[1].Title.English, "a name the panel gave is kept")
	assert.Empty(t, eps[2].Title.English, "an episode TopCine does not list stays unnamed")

	gone := topCineSeries("gone")
	unnamed := []models.Episode{{Num: 1}}
	fillTopCineEpisodeTitles(gone, unnamed)
	assert.Empty(t, unnamed[0].Title.English, "an unreadable page leaves the list as it is")
}

func TestDescribeMediaErr(t *testing.T) {
	err := startflix.ErrNoPlayers
	assert.Contains(t, describeMediaErr(&models.Anime{Source: "StartFlix"}, err).Error(), "on StartFlix")
	assert.Contains(t, describeMediaErr(topCineSeries("x"), err).Error(), "on TopCine")
	assert.Contains(t, describeMediaErr(nil, err).Error(), "on StartFlix")
}

func TestIsTopCine(t *testing.T) {
	assert.True(t, isTopCine(&models.Anime{Source: "TopCine"}))
	assert.False(t, isTopCine(&models.Anime{Source: "StartFlix"}))
	assert.False(t, isTopCine(&models.Anime{Source: "topcine"}), "Source is matched exactly, as the registry stamps it")
	assert.False(t, isTopCine(nil))
}

func TestGetTopCineStreamURL_SubtitledAudioPreference(t *testing.T) {
	useFakeTopCine(t, 0, 1) // season 1, Legendado
	media := topCineSeries("show")
	eps, err := GetTopCineEpisodes(media)
	require.NoError(t, err)

	streamURL, err := GetTopCineStreamURL(media, &eps[0], "best")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/ep-1101.mp4", streamURL)
	assert.Equal(t, originalALang, util.GetGlobalAudioLanguage(), "Legendado prefers the original audio")
	assert.Equal(t, "https://painel.test/", util.GetGlobalReferer())
	assert.NotEmpty(t, util.GetGlobalSubtitles(), "the panel's subtitle track reaches mpv")
}

func TestGetTopCineStreamURL_EpisodeFromHistory(t *testing.T) {
	useFakeTopCine(t, 0, 0)
	media := topCineSeries("show")
	// History keeps the season and number, not the panel endpoint.
	ep := &models.Episode{Number: "2", SeasonID: "2"}
	streamURL, err := GetTopCineStreamURL(media, ep, "best")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/ep-2002.mp4", streamURL)
}
