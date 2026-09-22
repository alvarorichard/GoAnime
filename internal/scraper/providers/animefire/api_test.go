package animefire

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// AnimeFire is read through its JSON API now.
//
// The site became an Angular single-page app and moved to animefire.one; the
// HTML the old scraper parsed is a ~13KB shell with nothing in it, so every
// search returned (nothing, nil) and the source disappeared from the fan-out
// without ever reporting a failure.
//
// These tests replace the HTML-era suite. They keep the same guarantees —
// retries, empty-is-not-an-error, failures are classified, stream selection —
// against the contract that actually exists.
//
// Shapes are taken from live responses captured 2026-09-21.

// apiServer is a fake api.animefire.one.
type apiServer struct {
	*httptest.Server
	hits atomic.Int32
}

func newAPIServer(t *testing.T, routes map[string]func(w http.ResponseWriter)) *apiServer {
	t.Helper()
	s := &apiServer{}
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.hits.Add(1)
		if h, ok := routes[r.URL.Path]; ok {
			h(w)
			return
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(s.Close)
	return s
}

func writeJSON(v any) func(http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(v)
	}
}

func writeStatus(code int, body string) func(http.ResponseWriter) {
	return func(w http.ResponseWriter) {
		w.WriteHeader(code)
		_, _ = w.Write([]byte(body))
	}
}

// spaShell is what animefire.one serves to a plain HTTP client: an app shell
// with no content. Reaching it instead of the API must be reported, not
// mistaken for an empty answer.
const spaShell = `<!doctype html><html lang="pt-BR" data-beasties-container="">` +
	`<head><base href="/"><title></title></head><body><app-root></app-root></body></html>`

// ── Search ──────────────────────────────────────────────────────────────────

func TestSearchAPI_MapsResultsFromTheLiveShape(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/animes/pesquisar": writeJSON(map[string]any{"data": []map[string]any{
			{
				"id":         "eU7t5IvcNKU",
				"titles":     map[string]string{"BR": "Naruto"},
				"audio":      "Dublado & Legendado",
				"poster_src": "https://image.tmdb.org/t/p/original/poster.jpg",
			},
			{
				"id":     "V2Q_qcvaKhb",
				"titles": map[string]string{"BR": "Naruto Shippuden", "US": "Naruto: Shippuden"},
			},
		}}),
	})

	res, err := NewClientForTest(srv.URL).SearchAnime("naruto")

	require.NoError(t, err)
	require.Len(t, res, 2)
	assert.Equal(t, "Naruto", res[0].Name)
	assert.Equal(t, "https://animefire.one/anime/eU7t5IvcNKU", res[0].URL,
		"the URL has to carry the id back to us when episodes are requested")
	assert.Equal(t, "https://image.tmdb.org/t/p/original/poster.jpg", res[0].ImageURL)
	assert.Equal(t, "Naruto Shippuden", res[1].Name, "the Brazilian title wins when several exist")
}

// A title in another language is still a title. Dropping the entry would be
// worse than showing it under the name the API does have.
func TestSearchAPI_FallsBackThroughTheOtherTitleLanguages(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/animes/pesquisar": writeJSON(map[string]any{"data": []map[string]any{
			{"id": "a1", "titles": map[string]string{"JP": "NARUTO"}},
			{"id": "a2", "titles": map[string]string{}},
			{"id": "", "titles": map[string]string{"BR": "sem id"}},
		}}),
	})

	res, err := NewClientForTest(srv.URL).SearchAnime("naruto")

	require.NoError(t, err)
	require.Len(t, res, 1, "entries without a usable title or id are not results")
	assert.Equal(t, "NARUTO", res[0].Name)
}

// An empty result set is an answer, not a failure — the distinction the HTML
// path got wrong for a whole site rewrite.
func TestSearchAPI_EmptyResultIsNotAnError(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/animes/pesquisar": writeJSON(map[string]any{"data": []any{}}),
	})

	res, err := NewClientForTest(srv.URL).SearchAnime("zzzzzzzz")

	require.NoError(t, err)
	assert.Empty(t, res)
}

// Reaching the site's app shell instead of the API must be classified, not
// swallowed. This is the failure that hid the outage.
func TestSearchAPI_HTMLInsteadOfJSONIsReportedAsABrokenParser(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/animes/pesquisar": func(w http.ResponseWriter) {
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte(spaShell))
		},
	})

	res, err := NewClientForTest(srv.URL).SearchAnime("naruto")

	require.Error(t, err, "an HTML answer is a broken contract, not an empty search")
	assert.Empty(t, res)
	var diag *netx.SourceDiagnostic
	require.ErrorAs(t, err, &diag)
	assert.Equal(t, netx.DiagnosticParserBroken, diag.Kind)
}

func TestSearchAPI_HTTPErrorIsSurfaced(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/animes/pesquisar": writeStatus(http.StatusInternalServerError, "boom"),
	})

	_, err := NewClientForTest(srv.URL).SearchAnime("naruto")

	require.Error(t, err)
}

// ── Episodes ────────────────────────────────────────────────────────────────

func TestEpisodesAPI_MapsEpisodesAndCarriesTheIDForward(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/anime/eU7t5IvcNKU": writeJSON(map[string]any{"data": map[string]any{
			"episodes": []map[string]any{
				{"id": "WCrJufyJmQn", "title": "Naruto Uzumaki Chegando!", "season": 1, "number": 1, "synopsis": "s1"},
				{"id": "5y0jBDHTS7V", "title": "Meu Nome é Konohamaru!", "season": 1, "number": 2},
				{"id": "", "title": "sem id", "number": 3},
			},
		}}),
	})

	eps, err := NewClientForTest(srv.URL).GetAnimeEpisodes("https://animefire.one/anime/eU7t5IvcNKU")

	require.NoError(t, err)
	require.Len(t, eps, 2, "an episode with no id cannot be played and is not listed")
	assert.Equal(t, "1", eps[0].Number)
	assert.Equal(t, 1, eps[0].Num)
	assert.Equal(t, "Naruto Uzumaki Chegando!", eps[0].Title.Romaji)
	assert.Equal(t, "s1", eps[0].Synopsis)
	assert.Equal(t, "1", eps[0].SeasonID)
	assert.Equal(t, "https://animefire.one/episode/WCrJufyJmQn", eps[0].URL,
		"the episode URL is what comes back when the stream is requested")
}

func TestEpisodesAPI_RefusesAURLWithNoID(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, nil)

	_, err := NewClientForTest(srv.URL).GetAnimeEpisodes("")

	require.Error(t, err)
	assert.Zero(t, srv.hits.Load(), "a request with nothing to ask for must not be sent")
}

// ── Streams ─────────────────────────────────────────────────────────────────

func TestStreamAPI_ReturnsThePlayableURL(t *testing.T) {
	t.Parallel()
	const hls = "https://akumast.net/i/token/h.jpg"
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/episode/WCrJufyJmQn": writeJSON(map[string]any{"data": map[string]any{
			"streams": []map[string]any{
				{"audio": "legendado", "url": "https://akumast.net/i/sub/h.jpg", "qualities": []string{"480p"}},
				{"audio": "dublado", "url": hls, "qualities": []string{"480p"}},
			},
		}}),
	})

	got, err := NewClientForTest(srv.URL).GetEpisodeStreamURL("https://animefire.one/episode/WCrJufyJmQn")

	require.NoError(t, err)
	assert.Equal(t, hls, got, "a dubbed track wins for a Brazilian audience when both exist")
}

func TestPickStream_PrefersDubbedAndSkipsOffline(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		streams []apiStream
		want    string
		ok      bool
	}{
		"dubbed wins": {
			streams: []apiStream{{Audio: "legendado", URL: "sub"}, {Audio: "dublado", URL: "dub"}},
			want:    "dub", ok: true,
		},
		"subbed when it is all there is": {
			streams: []apiStream{{Audio: "legendado", URL: "sub"}},
			want:    "sub", ok: true,
		},
		"offline entries are listed but have nothing behind them": {
			streams: []apiStream{{Audio: "dublado", URL: "dub", IsOffline: true}, {Audio: "legendado", URL: "sub"}},
			want:    "sub", ok: true,
		},
		"an entry with no url is not a stream": {
			streams: []apiStream{{Audio: "dublado", URL: "   "}},
			ok:      false,
		},
		"nothing at all": {streams: nil, ok: false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			got, ok := pickStream(tt.streams)
			assert.Equal(t, tt.ok, ok)
			if tt.ok {
				assert.Equal(t, tt.want, got.URL)
			}
		})
	}
}

// ── Plumbing ────────────────────────────────────────────────────────────────

func TestIDFromURL_AcceptsBothFormsWeMightBeHanded(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]string{
		"https://animefire.one/anime/eU7t5IvcNKU":    "eU7t5IvcNKU",
		"https://animefire.one/episode/WCrJufyJmQn/": "WCrJufyJmQn",
		"eU7t5IvcNKU":     "eU7t5IvcNKU",
		"  eU7t5IvcNKU  ": "eU7t5IvcNKU",
		"":                "",
	} {
		assert.Equalf(t, want, idFromURL(in), "idFromURL(%q)", in)
	}
}

// The retry policy came over from the HTML client and still has to hold: a
// transient failure is retried, a permanent one is not retried forever.
func TestRetrying_RetriesThenGivesUp(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if calls.Add(1) < 3 {
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"data":[{"id":"a1","titles":{"BR":"Naruto"}}]}`)
	}))
	defer srv.Close()

	c := NewClientForTest(srv.URL)
	c.maxRetries = 2 // three attempts in total

	res, err := c.SearchAnime("naruto")

	require.NoError(t, err, "the third attempt succeeds and must be allowed to happen")
	assert.Len(t, res, 1)
	assert.Equal(t, int32(3), calls.Load())
}

func TestRetrying_StopsAtTheConfiguredLimit(t *testing.T) {
	t.Parallel()
	var calls atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		calls.Add(1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	c := NewClientForTest(srv.URL)
	c.maxRetries = 2

	_, err := c.SearchAnime("naruto")

	require.Error(t, err)
	assert.Equal(t, int32(3), calls.Load(), "maxRetries=2 means three attempts, not an open loop")
}

// The API base has to be injectable, or none of the above could be tested
// without reaching the real host.
func TestNewClientForTest_PointsTheAPIAtTheTestServer(t *testing.T) {
	t.Parallel()
	c := NewClientForTest("http://127.0.0.1:1")

	assert.Equal(t, "http://127.0.0.1:1", c.apiBase)
	assert.NotEqual(t, defaultAPIBase, c.apiBase)
}

// ── Offline content ─────────────────────────────────────────────────────────

// AnimeFire lists episodes it has no file for: `is_offline: true` with a null
// url. Whole titles can be in that state — measured 2026-09-21, Naruto plays
// end to end while Naruto Shippuden is offline on every episode sampled.
//
// That is content state, not a scraping failure, and it has to be reported as
// such. The player used to fall through to the old HTML extractor here and
// answer "no video source found in the page", which blamed a parser for a file
// the source never had.
func TestStreamAPI_OfflineEpisodeIsReportedAsMissingContent(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/episode/dIdRMxXOywZ": writeJSON(map[string]any{"data": map[string]any{
			"streams": []map[string]any{
				{"audio": "dublado", "is_offline": true, "url": nil, "qualities": []string{"480p"}},
				{"audio": "legendado", "is_offline": true, "url": nil, "qualities": []string{"480p"}},
			},
		}}),
	})

	_, err := NewClientForTest(srv.URL).GetEpisodeStreamURL("https://animefire.one/episode/dIdRMxXOywZ")

	require.Error(t, err)
	assert.ErrorIs(t, err, netx.ErrMediaOffline,
		"the player decides what to tell the user from this; a bare string would not do")
	assert.NotErrorIs(t, err, netx.ErrSourceUnavailable,
		"the source answered perfectly — calling it unavailable would send the user to wait for nothing")
}

// A title with a playable track must not be caught by the same check.
func TestStreamAPI_OnlineEpisodeIsNotReportedAsOffline(t *testing.T) {
	t.Parallel()
	srv := newAPIServer(t, map[string]func(http.ResponseWriter){
		"/episode/ok": writeJSON(map[string]any{"data": map[string]any{
			"streams": []map[string]any{
				{"audio": "dublado", "is_offline": true, "url": nil},
				{"audio": "legendado", "url": "https://akumast.net/i/tok/h.jpg"},
			},
		}}),
	})

	got, err := NewClientForTest(srv.URL).GetEpisodeStreamURL("https://animefire.one/episode/ok")

	require.NoError(t, err, "one offline track among several must not condemn the episode")
	assert.Equal(t, "https://akumast.net/i/tok/h.jpg", got)
}

// AnimeFire lists titles it hosts nothing for, and nothing in the anime payload
// says so — status reads "completed", the audio reads "Dublado & Legendado",
// the episode list looks complete. Naruto Shippuden is listed with 500 episodes
// and not one has a file.
//
// The only way to find that out used to be picking an episode, being refused,
// and repeating: four attempts, four identical refusals, no way to tell it was
// the whole title rather than bad luck. So the listing probes.
func TestEpisodesAPI_ProbesForAWhollyOfflineTitle(t *testing.T) {
	t.Parallel()
	var probes atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/anime/offline" {
			_, _ = fmt.Fprint(w, `{"data":{"episodes":[
				{"id":"e1","number":1},{"id":"e2","number":2},{"id":"e3","number":3},{"id":"e4","number":4}]}}`)
			return
		}
		probes.Add(1)
		_, _ = fmt.Fprint(w, `{"data":{"streams":[{"audio":"dublado","is_offline":true,"url":null}]}}`)
	}))
	defer srv.Close()

	eps, err := NewClientForTest(srv.URL).GetAnimeEpisodes("https://animefire.one/anime/offline")

	require.NoError(t, err, "the listing itself is fine; it is the content that is missing")
	assert.Len(t, eps, 4, "the episodes are still listed so the user can look around")
	assert.Equal(t, int32(2), probes.Load(),
		"two probes: one gap is normal and must not condemn a title that mostly works")
}

// A title that plays must cost exactly one probe and no warning — the check
// cannot become a tax on every healthy listing.
func TestEpisodesAPI_StopsProbingAsSoonAsSomethingPlays(t *testing.T) {
	t.Parallel()
	var probes atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/anime/fine" {
			_, _ = fmt.Fprint(w, `{"data":{"episodes":[{"id":"e1","number":1},{"id":"e2","number":2}]}}`)
			return
		}
		probes.Add(1)
		_, _ = fmt.Fprint(w, `{"data":{"streams":[{"audio":"dublado","url":"https://akumast.net/i/t/h.jpg"}]}}`)
	}))
	defer srv.Close()

	eps, err := NewClientForTest(srv.URL).GetAnimeEpisodes("https://animefire.one/anime/fine")

	require.NoError(t, err)
	assert.Len(t, eps, 2)
	assert.Equal(t, int32(1), probes.Load(), "the first episode played; there is nothing left to establish")
}

// A one-episode listing has nothing to cross-check, so it must not probe at all
// rather than warn on a single sample.
func TestEpisodesAPI_DoesNotProbeAOneEpisodeListing(t *testing.T) {
	t.Parallel()
	var probes atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		if r.URL.Path == "/anime/single" {
			_, _ = fmt.Fprint(w, `{"data":{"episodes":[{"id":"e1","number":1}]}}`)
			return
		}
		probes.Add(1)
		_, _ = fmt.Fprint(w, `{"data":{"streams":[]}}`)
	}))
	defer srv.Close()

	eps, err := NewClientForTest(srv.URL).GetAnimeEpisodes("https://animefire.one/anime/single")

	require.NoError(t, err)
	assert.Len(t, eps, 1)
	assert.Zero(t, probes.Load(), "one sample cannot tell a gap from an empty title")
}
