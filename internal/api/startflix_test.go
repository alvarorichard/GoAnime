package api

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/movie"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The StartFlix tests swap package-level seams (client, pickers) and read the
// playback globals, so none of them run in parallel.

// fakeStartFlix serves a minimal catalog: one series with two seasons (S1 has
// both audios, S2 is dubbed only), one movie, and a direct-file player.
type fakeStartFlix struct{ srv *httptest.Server }

const fakeSeriesPage = `<html><body><div id="info">
	<iframe src="https://www.youtube.com/embed/trailer01"></iframe>
	<iframe src="https://painel.test/embed/100"></iframe></div></body></html>`

const fakeMoviePage = `<html><body><iframe src="https://painel.test/filme/tt100"></iframe></body></html>`

const fakeSeriesPanel = `<main class="main">
<ul class="header-navigation">
	<li data-season-id="s1" data-season-number="1">1 TEMP</li>
	<li data-season-id="s2" data-season-number="2">2 TEMP</li>
</ul>
<div class="cards">
	<div class="card"><div class="card-header"><h2 class="card-title">Dublado</h2></div><div class="card-body"><ul>
		<li data-season-id="s1" data-episode-id="1001"><a href="#">1 - Episódio</a></li>
		<li data-season-id="s2" data-episode-id="2001"><a href="#">1 - Episódio</a></li>
		<li data-season-id="s1" data-episode-id="1002"><a href="#">2 - Episódio</a></li>
		<li data-season-id="s2" data-episode-id="2002"><a href="#">2 - Episódio</a></li>
		<li data-season-id="s2" data-episode-id="2003"><a href="#">3 - Episódio</a></li>
	</ul></div></div>
	<div class="card"><div class="card-header"><h2 class="card-title">Legendado</h2></div><div class="card-body"><ul>
		<li data-season-id="s1" data-episode-id="1101"><a href="#">1 - O Começo</a></li>
	</ul></div></div>
</div></main>`

func fakePlayers(file string) string {
	return fmt.Sprintf(`<div id="players">
	<button class="btn" data-show-player="true" data-source="https://playembedapi.site/?v=x" data-subtitles="" data-type="iframe" data-id="1">Player #1</button>
	<button class="btn" data-show-player="true" data-source="https://files.test/%s" data-subtitles="https://files.test/pt.vtt" data-type="jwplayer" data-id="2">Player #2</button>
	</div>`, file)
}

func newFakeStartFlix(t *testing.T) *fakeStartFlix {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		host := r.Header.Get("X-Original-Host")
		switch {
		case host == "www.startflix.test" && r.URL.Path == "/series/show/":
			_, _ = w.Write([]byte(fakeSeriesPage))
		case host == "www.startflix.test" && r.URL.Path == "/filmes/film/":
			_, _ = w.Write([]byte(fakeMoviePage))
		case host == "painel.test" && r.URL.Path == "/embed/999":
			// The panel's answer for a title it does not carry.
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("Conteúdo não encontrado."))
		case host == "painel.test" && r.URL.Path == "/embed/100":
			_, _ = w.Write([]byte(fakeSeriesPanel))
		case host == "painel.test" && r.URL.Path == "/filme/tt100":
			_, _ = w.Write([]byte(fakePlayers("movie.mp4")))
		case host == "painel.test" && strings.HasPrefix(r.URL.Path, "/episodio/"):
			_, _ = w.Write([]byte(fakePlayers("ep-" + strings.TrimPrefix(r.URL.Path, "/episodio/") + ".mp4")))
		case host == "files.test":
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusPartialContent)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return &fakeStartFlix{srv: srv}
}

type routeAll struct{ target *url.URL }

func (r routeAll) RoundTrip(req *http.Request) (*http.Response, error) {
	out := req.Clone(req.Context())
	out.Header.Set("X-Original-Host", req.URL.Host)
	out.URL.Scheme, out.URL.Host, out.Host = r.target.Scheme, r.target.Host, r.target.Host
	return http.DefaultTransport.RoundTrip(out)
}

// useFakeStartFlix points the api seams at the fake site and fixed picker
// answers, restoring everything afterwards.
func useFakeStartFlix(t *testing.T, seasonIdx, audioIdx int) {
	t.Helper()
	fake := newFakeStartFlix(t)
	target, _ := url.Parse(fake.srv.URL)
	client := startflix.NewClientForTest(&http.Client{Transport: routeAll{target}}, "https://www.startflix.test")

	prevClient, prevSeason, prevPick, prevEnrich := sfxClientFn, sfxSeasonPickFn, sfxPickFn, sfxEnrichByIDFn
	prevAudio, prevExplicit := util.GetGlobalAudioLanguage(), util.GlobalAudioLanguageExplicit
	sfxClientFn = func() *startflix.Client { return client }
	// No metadata lookups over the network: by default nothing is found and
	// the StartFlix name stands. Tests that need a record swap this again.
	sfxEnrichByIDFn = func(*models.Media, bool) error { return movie.ErrNoOfficialRecord }
	sfxSeasonPickFn = func([]tui.PickItem, tui.PickOptions) (int, error) {
		if seasonIdx < 0 {
			return 0, tui.ErrPickBack
		}
		return seasonIdx, nil
	}
	sfxPickFn = func(string, []string) (int, error) { return audioIdx, nil }
	resetStartFlixState()
	t.Cleanup(func() {
		sfxClientFn, sfxSeasonPickFn, sfxPickFn, sfxEnrichByIDFn = prevClient, prevSeason, prevPick, prevEnrich
		resetStartFlixState()
		// Put back what other tests in this package expect to find.
		util.SetGlobalAudioLanguage(prevAudio)
		util.GlobalAudioLanguageExplicit = prevExplicit
	})
}

func resetStartFlixState() {
	sfxAudioMu.Lock()
	sfxAudioChoices = map[string]startflix.Audio{}
	sfxAudioMu.Unlock()
	// What the real binary starts with when --audio is not passed: the flag's
	// default, not an empty string.
	util.SetGlobalAudioLanguage("pt-BR,pt,english")
	util.GlobalAudioLanguageExplicit = false
	util.ClearGlobalSubtitles()
}

func TestGetStartFlixEpisodes_SeriesWithBothAudios(t *testing.T) {
	useFakeStartFlix(t, 0, 1) // season 1, Legendado
	media := &models.Anime{Name: "Show", URL: "https://www.startflix.test/series/show/", Source: "StartFlix", MediaType: models.MediaTypeTV}

	eps, err := GetStartFlixEpisodes(media)
	require.NoError(t, err)
	require.Len(t, eps, 1)
	assert.Equal(t, "https://painel.test/episodio/1101", eps[0].URL)
	assert.Equal(t, "1101", eps[0].DataID)
	assert.Equal(t, "1", eps[0].SeasonID)
	assert.Equal(t, "O Começo", eps[0].Title.English)
	assert.Equal(t, 1, media.CurrentSeason)
	assert.Equal(t, 100, media.TMDBID)

	audio, ok := recallStartFlixAudio(media.URL)
	assert.True(t, ok)
	assert.Equal(t, startflix.AudioSubtitled, audio, "the audio choice is remembered for the title")
}

func TestGetStartFlixEpisodes_SingleAudioSeasonDoesNotAsk(t *testing.T) {
	useFakeStartFlix(t, 1, 1)
	asked := false
	sfxPickFn = func(string, []string) (int, error) { asked = true; return 1, nil }
	media := &models.Anime{Name: "Show", URL: "https://www.startflix.test/series/show/", Source: "StartFlix", MediaType: models.MediaTypeTV}

	eps, err := GetStartFlixEpisodes(media)
	require.NoError(t, err)
	assert.False(t, asked, "season 2 is dubbed only; there is nothing to ask")
	require.Len(t, eps, 3)
	for i, ep := range eps {
		assert.Equal(t, i+1, ep.Num)
		assert.Equal(t, "2", ep.SeasonID)
	}
}

func TestGetStartFlixEpisodes_Movie(t *testing.T) {
	useFakeStartFlix(t, 0, 0)
	media := &models.Anime{Name: "Film", URL: "https://www.startflix.test/filmes/film/", Source: "StartFlix", MediaType: models.MediaTypeTV}

	eps, err := GetStartFlixEpisodes(media)
	require.NoError(t, err)
	require.Len(t, eps, 1)
	assert.Equal(t, "https://painel.test/filme/tt100", eps[0].URL)
	assert.Equal(t, models.MediaTypeMovie, media.MediaType, "the panel corrects the search page's guess")
	assert.Equal(t, "tt100", media.IMDBID)
}

// officialRecordFake answers EnrichByID with a fixed record and remembers
// what it was asked.
type officialRecordFake struct {
	calls     int
	movie     bool
	tmdbID    int
	imdbID    string
	hadRecord bool
	title     string
	year      string
}

func (f *officialRecordFake) enrich(m *models.Media, isMovie bool) error {
	f.calls++
	f.movie, f.tmdbID, f.imdbID, f.hadRecord = isMovie, m.TMDBID, m.IMDBID, m.TMDBDetails != nil
	if f.title == "" {
		return movie.ErrNoOfficialRecord
	}
	m.TMDBDetails = &models.TMDBDetails{Title: f.title}
	m.Year = f.year
	return nil
}

// TestGetStartFlixEpisodes_MovieNamedAfterOfficialRecord pins the point of
// the lookup: a movie listed under its Brazilian name is downloaded under the
// official English title and year, with its IMDb id on the folder, the way
// SuperFlix titles were.
func TestGetStartFlixEpisodes_MovieNamedAfterOfficialRecord(t *testing.T) {
	useFakeStartFlix(t, 0, 0)
	fake := &officialRecordFake{title: "Heart of the Beast", year: "2026"}
	sfxEnrichByIDFn = fake.enrich
	media := &models.Anime{Name: "Coração Selvagem", URL: "https://www.startflix.test/filmes/film/", Source: "StartFlix", MediaType: models.MediaTypeMovie}

	_, err := GetStartFlixEpisodes(media)
	require.NoError(t, err)
	require.Equal(t, 1, fake.calls)
	assert.True(t, fake.movie, "a /filme/ panel is looked up as a movie")
	assert.Equal(t, "tt100", fake.imdbID, "looked up by the panel's IMDb id")
	assert.Equal(t, "Heart of the Beast", media.OfficialTitle())

	meta := &util.MediaMeta{OfficialTitle: media.OfficialTitle(), Year: media.Year, TMDBID: media.TMDBID, IMDBID: media.IMDBID}
	assert.Equal(t, "/dl/Heart of the Beast (2026) {imdb-tt100}/Heart of the Beast (2026).mp4",
		util.FormatPlexMoviePath("/dl", media.Name, "", meta))
}

func TestGetStartFlixEpisodes_SeriesLookedUpByTMDBID(t *testing.T) {
	useFakeStartFlix(t, 0, 0)
	fake := &officialRecordFake{}
	sfxEnrichByIDFn = fake.enrich
	media := &models.Anime{Name: "Show", URL: "https://www.startflix.test/series/show/", Source: "StartFlix", MediaType: models.MediaTypeTV}

	_, err := GetStartFlixEpisodes(media)
	require.NoError(t, err)
	require.Equal(t, 1, fake.calls)
	assert.False(t, fake.movie)
	assert.Equal(t, 100, fake.tmdbID)
	assert.Equal(t, "Show", media.OfficialTitle(), "with no record found the StartFlix name stands")
}

// TestGetStartFlixEpisodes_PanelIDsReplaceNameGuess pins that the panel's ids
// win over a match the selection-time search made by the Portuguese name:
// the guessed ids and the record they brought in are dropped before the
// lookup, so the wrong film's title never names the download.
func TestGetStartFlixEpisodes_PanelIDsReplaceNameGuess(t *testing.T) {
	useFakeStartFlix(t, 0, 0)
	fake := &officialRecordFake{}
	sfxEnrichByIDFn = fake.enrich
	media := &models.Anime{
		Name: "Coração Selvagem", URL: "https://www.startflix.test/filmes/film/", Source: "StartFlix",
		MediaType: models.MediaTypeMovie, TMDBID: 483, IMDBID: "tt0100935", Rating: 7.2,
		TMDBDetails: &models.TMDBDetails{Title: "Wild at Heart"},
	}

	_, err := GetStartFlixEpisodes(media)
	require.NoError(t, err)
	assert.Equal(t, "tt100", fake.imdbID)
	assert.Zero(t, fake.tmdbID, "the guessed TMDB id belonged to the other film")
	assert.False(t, fake.hadRecord, "the other film's record is dropped before the lookup")
	assert.Equal(t, "Coração Selvagem", media.OfficialTitle())
	assert.Zero(t, media.Rating)
}

func TestGetStartFlixEpisodes_MatchingGuessIsKept(t *testing.T) {
	useFakeStartFlix(t, 0, 0)
	fake := &officialRecordFake{}
	sfxEnrichByIDFn = fake.enrich
	media := &models.Anime{
		Name: "Film", URL: "https://www.startflix.test/filmes/film/", Source: "StartFlix",
		MediaType: models.MediaTypeMovie, TMDBID: 7, IMDBID: "tt100",
		TMDBDetails: &models.TMDBDetails{Title: "The Film"},
	}

	_, err := GetStartFlixEpisodes(media)
	require.NoError(t, err)
	assert.Equal(t, 7, fake.tmdbID, "ids that agree with the panel stay")
	assert.Equal(t, "The Film", media.OfficialTitle())
}

func TestGetStartFlixEpisodes_BackFromSeasonPicker(t *testing.T) {
	useFakeStartFlix(t, -1, 0)
	media := &models.Anime{Name: "Show", URL: "https://www.startflix.test/series/show/", Source: "StartFlix"}
	_, err := GetStartFlixEpisodes(media)
	assert.ErrorIs(t, err, ErrBackToSearch)
}

func TestGetStartFlixEpisodes_TitleWithoutPanel(t *testing.T) {
	useFakeStartFlix(t, 0, 0)
	media := &models.Anime{Name: "Gone", URL: "https://www.startflix.test/series/gone/", Source: "StartFlix"}
	_, err := GetStartFlixEpisodes(media)
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "startflix:", "the user sees a plain message, not the raw cause")
}

func TestGetStartFlixStreamURL_SetsPlaybackGlobals(t *testing.T) {
	useFakeStartFlix(t, 0, 1)
	media := &models.Anime{Name: "Show", URL: "https://www.startflix.test/series/show/", Source: "StartFlix", MediaType: models.MediaTypeTV}
	eps, err := GetStartFlixEpisodes(media)
	require.NoError(t, err)

	got, err := GetStartFlixStreamURL(media, &eps[0], "best")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/ep-1101.mp4", got, "the unsupported Abyss player is skipped")
	assert.Equal(t, "https://painel.test/", util.GetGlobalReferer())
	assert.Equal(t, audioLangForStartFlix(startflix.AudioSubtitled), util.GetGlobalAudioLanguage())
	subs := util.GetGlobalSubtitles()
	require.Len(t, subs, 1)
	assert.Equal(t, "https://files.test/pt.vtt", subs[0].URL)
}

func TestGetStartFlixStreamURL_MovieResetsAudioToDub(t *testing.T) {
	useFakeStartFlix(t, 0, 1)
	rememberStartFlixAudio("https://www.startflix.test/filmes/film/", startflix.AudioSubtitled)
	util.SetGlobalAudioLanguage("jpn") // left behind by a previous Legendado title

	media := &models.Anime{Name: "Film", URL: "https://www.startflix.test/filmes/film/", Source: "StartFlix", MediaType: models.MediaTypeMovie}
	got, err := GetStartFlixStreamURL(media, &models.Episode{Number: "1", URL: "https://painel.test/filme/tt100"}, "")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/movie.mp4", got)
	assert.Equal(t, audioLangForStartFlix(startflix.AudioDubbed), util.GetGlobalAudioLanguage())
}

func TestGetStartFlixStreamURL_RebuildsEpisodeFromHistory(t *testing.T) {
	useFakeStartFlix(t, 0, 0)
	media := &models.Anime{Name: "Show", URL: "https://www.startflix.test/series/show/", Source: "StartFlix", MediaType: models.MediaTypeTV}
	// What a restored history entry knows: season and number, no panel URL.
	got, err := GetStartFlixStreamURL(media, &models.Episode{Number: "3", Num: 3, SeasonID: "2"}, "")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/ep-2003.mp4", got)
}

func TestSelectStartFlixAudio(t *testing.T) {
	both := startflix.Season{Number: 1,
		Dubbed:    []startflix.Episode{{Number: 1}},
		Subtitled: []startflix.Episode{{Number: 1}}}
	subOnly := startflix.Season{Number: 2, Subtitled: []startflix.Episode{{Number: 1}}}

	t.Run("single audio needs no prompt", func(t *testing.T) {
		useFakeStartFlix(t, 0, 0)
		sfxPickFn = func(string, []string) (int, error) { t.Fatal("asked"); return 0, nil }
		assert.Equal(t, startflix.AudioSubtitled, selectStartFlixAudio("t", subOnly))
	})
	t.Run("asks once per title", func(t *testing.T) {
		useFakeStartFlix(t, 0, 0)
		calls := 0
		sfxPickFn = func(string, []string) (int, error) { calls++; return 1, nil }
		assert.Equal(t, startflix.AudioSubtitled, selectStartFlixAudio("t", both))
		assert.Equal(t, startflix.AudioSubtitled, selectStartFlixAudio("t", both))
		assert.Equal(t, 1, calls)
	})
	t.Run("remembered audio absent from the season is not forced", func(t *testing.T) {
		useFakeStartFlix(t, 0, 0)
		rememberStartFlixAudio("t", startflix.AudioDubbed)
		assert.Equal(t, startflix.AudioSubtitled, selectStartFlixAudio("t", subOnly))
	})
	t.Run("picker failure falls back to the dub", func(t *testing.T) {
		useFakeStartFlix(t, 0, 0)
		sfxPickFn = func(string, []string) (int, error) { return 0, errors.New("no tty") }
		assert.Equal(t, startflix.AudioDubbed, selectStartFlixAudio("t", both))
	})
	t.Run("the --audio default is not a choice", func(t *testing.T) {
		useFakeStartFlix(t, 0, 0) // global holds "pt-BR,pt,english", not explicit
		asked := false
		sfxPickFn = func(string, []string) (int, error) { asked = true; return 1, nil }
		assert.Equal(t, startflix.AudioSubtitled, selectStartFlixAudio("t", both))
		assert.True(t, asked, "Legendado must still be offered when --audio was not passed")
	})
	t.Run("an explicit --audio decides", func(t *testing.T) {
		useFakeStartFlix(t, 0, 0)
		util.SetGlobalAudioLanguage("jpn,eng")
		util.GlobalAudioLanguageExplicit = true
		sfxPickFn = func(string, []string) (int, error) { t.Fatal("asked despite --audio"); return 0, nil }
		assert.Equal(t, startflix.AudioSubtitled, selectStartFlixAudio("t", both))
	})
}

func TestDescribeStartFlixErr(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		err  error
		want string
	}{
		{"unsupported hosts are named", &startflix.NoStreamError{Unsupported: []string{"playembedapi.site", "vidsrcme.su"}}, "playembedapi.site, vidsrcme.su"},
		{"supported but failing", &startflix.NoStreamError{Failures: []error{errors.New("x")}}, "None of StartFlix's servers worked"},
		{"no players", startflix.ErrNoPlayers, "No video sources"},
		{"not on panel", fmt.Errorf("wrap: %w", startflix.ErrNotOnPanel), "has no video for it"},
		{"no panel", startflix.ErrNoPanel, "has no video for it"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := describeStartFlixErr(tt.err)
			assert.Contains(t, got.Error(), tt.want)
			assert.ErrorIs(t, got, tt.err, "the cause stays reachable")
		})
	}
	assert.NoError(t, describeStartFlixErr(nil))
	plain := errors.New("plain")
	assert.Same(t, plain, describeStartFlixErr(plain))
}
