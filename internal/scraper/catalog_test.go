package scraper

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// catalogHandler serves one series (panel 55: season 1 in both audios, season 2
// dubbed only), one movie (panel tt1: a dub and a sub player side by side) and
// a movie whose only player is the dub.
func catalogHandler(w http.ResponseWriter, r *http.Request) {
	player := func(src, label string) string {
		return fmt.Sprintf(`<button data-show-player="true" data-source=%q data-type="jwplayer" data-id="1">%s</button>`, src, label)
	}
	switch {
	case strings.HasSuffix(r.URL.Path, ".mp4"):
		w.Header().Set("Content-Type", "video/mp4")
		w.WriteHeader(http.StatusPartialContent)
	case r.URL.Path == "/series/x/":
		fmt.Fprint(w, `<iframe src="https://painel.test/embed/55"></iframe>`)
	case r.URL.Path == "/filmes/m/":
		fmt.Fprint(w, `<iframe src="https://painel.test/filme/tt1"></iframe>`)
	case r.URL.Path == "/filmes/dubonly/":
		fmt.Fprint(w, `<iframe src="https://painel.test/filme/tt2"></iframe>`)
	case r.URL.Path == "/embed/55":
		fmt.Fprint(w, `<ul class="header-navigation">
			<li data-season-id="s1" data-season-number="1">1</li>
			<li data-season-id="s2" data-season-number="2">2</li></ul>
			<div class="cards">
			<div class="card"><h2 class="card-title">Dublado</h2><ul>
				<li data-season-id="s1" data-episode-id="10"><a>1 - Piloto</a></li>
				<li data-season-id="s1" data-episode-id="11"><a>2 - Segundo</a></li>
				<li data-season-id="s2" data-episode-id="20"><a>1 - Volta</a></li></ul></div>
			<div class="card"><h2 class="card-title">Legendado</h2><ul>
				<li data-season-id="s1" data-episode-id="30"><a>1 - Piloto</a></li></ul></div></div>`)
	case r.URL.Path == "/filme/tt1":
		fmt.Fprint(w, player("https://files.test/dub.mp4", "Player #1 Dublado")+player("https://files.test/sub.mp4", "Player #2 Legendado"))
	case r.URL.Path == "/filme/tt2":
		fmt.Fprint(w, player("https://files.test/dub.mp4", "Player #1 Dublado"))
	case strings.HasPrefix(r.URL.Path, "/episodio/"):
		fmt.Fprint(w, player("https://files.test/ep"+strings.TrimPrefix(r.URL.Path, "/episodio/")+".mp4", "Player #1"))
	default:
		http.NotFound(w, r)
	}
}

func catalogAdapter(t *testing.T) *StartFlixAdapter {
	t.Helper()
	return newStartFlixTestAdapter(t, catalogHandler)
}

const (
	seriesPage = "https://www.startflix.test/series/x/"
	moviePage  = "https://www.startflix.test/filmes/m/"
)

func TestStartFlixCatalogSeasons(t *testing.T) {
	t.Parallel()
	a := catalogAdapter(t)
	got, err := a.CatalogSeasons(context.Background(), seriesPage)
	require.NoError(t, err)
	assert.Equal(t, []CatalogSeason{
		{Number: 1, Audios: []string{"Dublado", "Legendado"}},
		{Number: 2, Audios: []string{"Dublado"}},
	}, got)

	none, err := a.CatalogSeasons(context.Background(), moviePage)
	require.NoError(t, err)
	assert.Empty(t, none, "a movie has no seasons")
}

func TestStartFlixCatalogEpisodes(t *testing.T) {
	t.Parallel()
	a := catalogAdapter(t)
	ctx := context.Background()

	movie, err := a.CatalogEpisodes(ctx, moviePage, 0, "")
	require.NoError(t, err)
	require.Len(t, movie, 1)
	assert.Equal(t, "https://painel.test/filme/tt1", movie[0].URL)
	assert.Equal(t, 1, movie[0].Num)

	_, err = a.CatalogEpisodes(ctx, seriesPage, 0, "")
	require.ErrorIs(t, err, ErrSeasonRequired)

	dub, err := a.CatalogEpisodes(ctx, seriesPage, 1, "")
	require.NoError(t, err)
	require.Len(t, dub, 2, "no audio takes the dub")
	assert.Equal(t, "https://painel.test/episodio/10", dub[0].URL)
	assert.Equal(t, "Piloto", dub[0].Title.English)
	assert.Equal(t, "1", dub[0].SeasonID)
	assert.Equal(t, "Dublado", dub[0].Audio)

	sub, err := a.CatalogEpisodes(ctx, seriesPage, 1, "legendado")
	require.NoError(t, err)
	require.Len(t, sub, 1)
	assert.Equal(t, "https://painel.test/episodio/30", sub[0].URL)
	assert.Equal(t, "Legendado", sub[0].Audio)

	_, err = a.CatalogEpisodes(ctx, seriesPage, 2, "Legendado")
	require.ErrorIs(t, err, ErrAudioUnavailable)
	assert.Contains(t, err.Error(), "Dublado", "the error says what the season has")

	_, err = a.CatalogEpisodes(ctx, seriesPage, 9, "")
	require.ErrorIs(t, err, ErrSeasonNotFound)
	assert.Contains(t, err.Error(), "1, 2")

	_, err = a.CatalogEpisodes(ctx, seriesPage, 1, "klingon")
	require.ErrorIs(t, err, ErrAudioUnavailable)
}

func TestStartFlixCatalogStreamAudio(t *testing.T) {
	t.Parallel()
	a := catalogAdapter(t)
	ctx := context.Background()

	u, meta, err := a.CatalogStream(ctx, "", models.Episode{URL: "https://painel.test/filme/tt1"}, "Legendado")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/sub.mp4", u)
	assert.Equal(t, "Legendado", meta["audio"])
	assert.Equal(t, "startflix", meta["source"])

	u, _, err = a.CatalogStream(ctx, "", models.Episode{URL: "https://painel.test/filme/tt1"}, "dub")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/dub.mp4", u)

	_, _, err = a.CatalogStream(ctx, "", models.Episode{URL: "https://painel.test/filme/tt2"}, "Legendado")
	require.ErrorIs(t, err, ErrAudioUnavailable, "the dub is never played in the sub's place")
	assert.Contains(t, err.Error(), "Dublado")

	u, _, err = a.CatalogStream(ctx, "", models.Episode{URL: "https://painel.test/episodio/10"}, "")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/ep10.mp4", u)

	_, _, err = a.CatalogStream(ctx, "", models.Episode{URL: "https://painel.test/filme/tt1"}, "klingon")
	require.ErrorIs(t, err, ErrAudioUnavailable)
}

func TestParseCatalogAudio(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]startflix.Audio{
		"": "", "Dublado": startflix.AudioDubbed, "DUB": startflix.AudioDubbed, "dubbed": startflix.AudioDubbed,
		"legendado": startflix.AudioSubtitled, "Sub": startflix.AudioSubtitled, "subtitled": startflix.AudioSubtitled,
	} {
		got, err := ParseCatalogAudio(in)
		require.NoError(t, err, in)
		assert.Equal(t, want, got, in)
	}
	got, err := ParseCatalogAudio(" Dublado ")
	require.NoError(t, err)
	assert.Equal(t, startflix.AudioDubbed, got, "surrounding space is ignored")

	_, err = ParseCatalogAudio("dublegendado")
	require.ErrorIs(t, err, ErrAudioUnavailable)
}

// The adapters other than StartFlix are not catalogues yet.
func TestOnlyStartFlixIsACatalogScraper(t *testing.T) {
	t.Parallel()
	for _, st := range []ScraperType{AnimefireType, HiAnimeType, TopCineType, PobreflixType, PomfyType} {
		a, err := NewAdapter(st)
		require.NoError(t, err)
		_, ok := a.(CatalogScraper)
		assert.False(t, ok, "%v", st)
	}
	a, err := NewAdapter(StartFlixType)
	require.NoError(t, err)
	_, ok := a.(CatalogScraper)
	assert.True(t, ok)
}

// stubFallback swaps the vidsrc seam for the test and records what it was asked.
type fallbackCall struct {
	panel           startflix.Panel
	season, episode int
}

func stubFallback(t *testing.T, stream *startflix.Stream, err error) *[]fallbackCall {
	t.Helper()
	var calls []fallbackCall
	old := catalogFallbackFn
	catalogFallbackFn = func(_ context.Context, p startflix.Panel, season, episode int) (*startflix.Stream, error) {
		calls = append(calls, fallbackCall{p, season, episode})
		return stream, err
	}
	t.Cleanup(func() { catalogFallbackFn = old })
	return &calls
}

// A panel with no player at all (the episode's players list is empty) is a
// panel failure: the fallback resolves the title by its ids.
func TestStartFlixCatalogStreamFallsBackWhenThePanelFails(t *testing.T) {
	a := newStartFlixTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/episodio/") {
			fmt.Fprint(w, `<p>nothing</p>`)
			return
		}
		catalogHandler(w, r)
	})
	calls := stubFallback(t, &startflix.Stream{URL: "https://vidsrc/ep.m3u8", Host: "vidsrc.test"}, nil)

	ep := models.Episode{Num: 4, URL: "https://painel.test/episodio/10", SeasonID: "2"}
	u, meta, err := a.CatalogStream(context.Background(), seriesPage, ep, "")
	require.NoError(t, err)
	assert.Equal(t, "https://vidsrc/ep.m3u8", u)
	assert.Equal(t, "Legendado", meta["audio"], "the fallback serves the original audio")
	require.Len(t, *calls, 1)
	assert.Equal(t, fallbackCall{startflix.Panel{URL: "https://painel.test/embed/55", Kind: startflix.KindSeries, TMDBID: 55}, 2, 4}, (*calls)[0])

	// A dub is not the fallback's to give: it is not even asked, and the
	// panel's own failure is what the caller hears, not a made-up audio error.
	n0 := len(*calls)
	_, _, err = a.CatalogStream(context.Background(), seriesPage, ep, "Dublado")
	require.ErrorIs(t, err, startflix.ErrNoPlayers)
	assert.Len(t, *calls, n0, "the fallback serves the original audio only")

	// An episode listed in the dub plays in the dub though the caller says nothing.
	ep.Audio = "Dublado"
	_, _, err = a.CatalogStream(context.Background(), seriesPage, ep, "")
	require.ErrorIs(t, err, startflix.ErrNoPlayers)
	assert.Len(t, *calls, n0)
	ep.Audio = ""

	// Season or episode 0 would mean "the Pilot" to the fallback: a series
	// episode without them is never looked up.
	n := len(*calls)
	_, _, err = a.CatalogStream(context.Background(), seriesPage, models.Episode{URL: ep.URL, Num: 4}, "")
	require.ErrorIs(t, err, startflix.ErrNoPlayers)
	_, _, err = a.CatalogStream(context.Background(), seriesPage, models.Episode{URL: ep.URL, SeasonID: "2"}, "")
	require.ErrorIs(t, err, startflix.ErrNoPlayers)
	assert.Len(t, *calls, n, "no season / no episode number: no lookup")

	// No title page, no fallback: the panel's own error stays.
	_, _, err = a.CatalogStream(context.Background(), "", ep, "")
	require.Error(t, err)
	assert.Len(t, *calls, n, "nothing to look the title up by")
	assert.NotErrorIs(t, err, ErrAudioUnavailable)
}

func TestStartFlixCatalogStreamFallbackMissKeepsThePanelError(t *testing.T) {
	a := newStartFlixTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/episodio/") {
			fmt.Fprint(w, `<p>nothing</p>`)
			return
		}
		catalogHandler(w, r)
	})
	stubFallback(t, nil, errors.New("vidsrc down"))
	_, _, err := a.CatalogStream(context.Background(), seriesPage, models.Episode{URL: "https://painel.test/episodio/10", Num: 1, SeasonID: "1"}, "")
	require.ErrorIs(t, err, startflix.ErrNoPlayers)
}

// A movie whose panel has only the dub: asking for the sub is answered by the
// fallback (the original audio), and by an error when it misses too.
func TestStartFlixCatalogStreamFallbackSuppliesTheMissingAudio(t *testing.T) {
	a := catalogAdapter(t)
	moviePanelEp := models.Episode{Num: 1, URL: "https://painel.test/filme/tt2"}
	dubOnly := "https://www.startflix.test/filmes/dubonly/"

	calls := stubFallback(t, &startflix.Stream{URL: "https://vidsrc/m.m3u8", Audio: startflix.AudioSubtitled}, nil)
	u, meta, err := a.CatalogStream(context.Background(), dubOnly, moviePanelEp, "Legendado")
	require.NoError(t, err)
	assert.Equal(t, "https://vidsrc/m.m3u8", u)
	assert.Equal(t, "Legendado", meta["audio"])
	require.Len(t, *calls, 1)
	assert.Equal(t, startflix.KindMovie, (*calls)[0].panel.Kind)
	assert.Equal(t, "tt2", (*calls)[0].panel.IMDBID)

	// The dub was available on the panel: the fallback is never consulted.
	*calls = nil
	u, _, err = a.CatalogStream(context.Background(), dubOnly, moviePanelEp, "Dublado")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/dub.mp4", u)
	assert.Empty(t, *calls)

	stubFallback(t, nil, errors.New("miss"))
	_, _, err = a.CatalogStream(context.Background(), dubOnly, moviePanelEp, "Legendado")
	require.ErrorIs(t, err, ErrAudioUnavailable, "the dub is still not played in the sub's place")

	// A fallback that also serves the wrong audio does not replace the panel.
	stubFallback(t, &startflix.Stream{URL: "https://vidsrc/x", Audio: startflix.AudioDubbed}, nil)
	_, _, err = a.CatalogStream(context.Background(), dubOnly, moviePanelEp, "Legendado")
	require.ErrorIs(t, err, ErrAudioUnavailable)
}

func emptyEpisodeHandler(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/episodio/") {
		fmt.Fprint(w, `<p>nothing</p>`)
		return
	}
	if r.URL.Path == "/" {
		fmt.Fprint(w, `<iframe src="https://painel.test/embed/55"></iframe>`)
		return
	}
	catalogHandler(w, r)
}

// The fallback is asked only for a title page that resolves to a panel.
func TestStartFlixCatalogStreamFallbackNeedsAResolvablePanel(t *testing.T) {
	a := newStartFlixTestAdapter(t, emptyEpisodeHandler)
	calls := stubFallback(t, &startflix.Stream{URL: "https://vidsrc/ep.m3u8"}, nil)
	ep := models.Episode{Num: 1, URL: "https://painel.test/episodio/10", SeasonID: "1"}

	_, _, err := a.CatalogStream(context.Background(), "https://www.startflix.test/missing/", ep, "")
	require.ErrorIs(t, err, startflix.ErrNoPlayers, "the panel's error stays")
	assert.Empty(t, *calls, "a title page that does not resolve is not looked up by zero ids")

	_, _, err = a.CatalogStream(context.Background(), "", ep, "")
	require.Error(t, err)
	assert.Empty(t, *calls, "an empty title URL never reaches the panel lookup")
}

// A fallback that returns nothing, or a stream with an error, is a miss.
func TestStartFlixCatalogStreamFallbackMisses(t *testing.T) {
	a := newStartFlixTestAdapter(t, emptyEpisodeHandler)
	ep := models.Episode{Num: 1, URL: "https://painel.test/episodio/10", SeasonID: "1"}

	stubFallback(t, nil, nil)
	_, _, err := a.CatalogStream(context.Background(), seriesPage, ep, "")
	require.ErrorIs(t, err, startflix.ErrNoPlayers)

	stubFallback(t, &startflix.Stream{URL: "https://vidsrc/partial"}, errors.New("half done"))
	_, _, err = a.CatalogStream(context.Background(), seriesPage, ep, "")
	require.ErrorIs(t, err, startflix.ErrNoPlayers, "a stream that came with an error is not used")
}

// Without Num the episode is read from its display number.
func TestStartFlixCatalogStreamFallbackReadsTheDisplayNumber(t *testing.T) {
	a := newStartFlixTestAdapter(t, emptyEpisodeHandler)
	calls := stubFallback(t, &startflix.Stream{URL: "https://vidsrc/ep.m3u8"}, nil)
	_, _, err := a.CatalogStream(context.Background(), seriesPage, models.Episode{Number: "6", URL: "https://painel.test/episodio/10", SeasonID: "2"}, "")
	require.NoError(t, err)
	require.Len(t, *calls, 1)
	assert.Equal(t, 6, (*calls)[0].episode)
	assert.Equal(t, 2, (*calls)[0].season)
}

// A cancelled caller gets its cancellation, not a fallback attempt.
func TestStartFlixCatalogStreamCancelledSkipsTheFallback(t *testing.T) {
	ctx, cancel := context.WithCancel(context.Background())
	t.Cleanup(cancel)
	a := newStartFlixTestAdapter(t, func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/episodio/") {
			cancel() // the caller gives up while the panel answers with no player
		}
		emptyEpisodeHandler(w, r)
	})
	calls := stubFallback(t, &startflix.Stream{URL: "https://vidsrc/ep.m3u8"}, nil)
	_, _, err := a.CatalogStream(ctx, seriesPage, models.Episode{Num: 1, URL: "https://painel.test/episodio/10", SeasonID: "1"}, "")
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, context.Canceled, err, "the caller's own error, not the panel's wrapped one")
	assert.Empty(t, *calls)
}

// The panel answers a stream whose audio is the one asked for, or one with no
// audio label at all: neither is a reason to ask the fallback.
func TestStartFlixCatalogStreamFallbackOnlyForAKnownWrongAudio(t *testing.T) {
	a := catalogAdapter(t)
	calls := stubFallback(t, &startflix.Stream{URL: "https://vidsrc/x", Audio: startflix.AudioSubtitled}, nil)

	// No audio asked: the dub the panel plays is kept.
	u, _, err := a.CatalogStream(context.Background(), moviePage, models.Episode{Num: 1, URL: "https://painel.test/filme/tt1"}, "")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/dub.mp4", u)

	// A player with no audio label takes the audio asked for: not "wrong".
	u, _, err = a.CatalogStream(context.Background(), seriesPage, models.Episode{Num: 1, URL: "https://painel.test/episodio/10", SeasonID: "1"}, "Legendado")
	require.NoError(t, err)
	assert.Equal(t, "https://files.test/ep10.mp4", u)
	assert.Empty(t, *calls)
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

// Both the panel lookup and the fallback are bounded by their own budget, not
// by the caller's (longer) deadline.
func TestStartFlixCatalogStreamFallbackPhaseBudget(t *testing.T) {
	var titleLeft, fallbackLeft time.Duration
	srv := httptest.NewServer(http.HandlerFunc(emptyEpisodeHandler))
	t.Cleanup(srv.Close)
	target, _ := url.Parse(srv.URL)
	rec := roundTripFunc(func(req *http.Request) (*http.Response, error) {
		if req.URL.Path == "/series/x/" {
			if dl, ok := req.Context().Deadline(); ok {
				titleLeft = time.Until(dl)
			}
		}
		return sameServer{target}.RoundTrip(req)
	})
	a := NewStartFlixAdapterWithClient(startflix.NewClientForTest(&http.Client{Transport: rec}, "https://www.startflix.test"))

	old := catalogFallbackFn
	catalogFallbackFn = func(ctx context.Context, _ startflix.Panel, _, _ int) (*startflix.Stream, error) {
		if dl, ok := ctx.Deadline(); ok {
			fallbackLeft = time.Until(dl)
		}
		return &startflix.Stream{URL: "https://vidsrc/ep.m3u8"}, nil
	}
	t.Cleanup(func() { catalogFallbackFn = old })

	ctx, cancel := context.WithTimeout(context.Background(), time.Hour)
	defer cancel()
	_, _, err := a.CatalogStream(ctx, seriesPage, models.Episode{Num: 1, URL: "https://painel.test/episodio/10", SeasonID: "1"}, "")
	require.NoError(t, err)
	for name, left := range map[string]time.Duration{"title page": titleLeft, "fallback": fallbackLeft} {
		assert.Greater(t, left, time.Duration(0), name)
		assert.LessOrEqual(t, left, catalogPhaseBudget, name)
	}
}

// The production fallback seam routes a movie to the movie lookup by its IMDb
// id (TMDB when that is all the panel has) and a series to the episode lookup.
func TestCatalogFallbackFnRoutesByKind(t *testing.T) {
	var got []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = append(got, r.URL.Query().Get("type")+":"+r.URL.Query().Get("id"))
		http.NotFound(w, r)
	}))
	t.Cleanup(srv.Close)
	t.Setenv("GOANIME_VIDSRC_FALLBACK_TEST", "1")
	t.Setenv("GOANIME_VIDSRC_URL", srv.URL)

	_, err := catalogFallbackFn(context.Background(), startflix.Panel{Kind: startflix.KindMovie, IMDBID: "tt7"}, 0, 0)
	require.Error(t, err)
	_, err = catalogFallbackFn(context.Background(), startflix.Panel{Kind: startflix.KindMovie, TMDBID: 99}, 0, 0)
	require.Error(t, err)
	_, err = catalogFallbackFn(context.Background(), startflix.Panel{Kind: startflix.KindSeries, TMDBID: 55}, 2, 4)
	require.Error(t, err)
	assert.Equal(t, []string{"movie:tt7", "movie:99", "series:55"}, got)
}
