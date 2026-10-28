package startflix

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The vidssrc fallback tests run the whole chain — vs_src.php → landing →
// player page → stream-data envelope → wasm decrypt → host token → master
// probe — against one httptest server. Production hops between its own hosts;
// here every payload carries the test server's URL, and the routing transport
// sends even the fixture-decrypted real-world hosts back to the same server,
// so a test never touches the network. The decryptor wasm is the real one,
// stored as testdata; decrypting offline requires no network at all.

// routeVIDSRC sends every request to the test server regardless of the host
// named in the URL, so fixture stream URLs (real hosts) still land here.
type routeVIDSRC struct{ target *url.URL }

func (t routeVIDSRC) RoundTrip(req *http.Request) (*http.Response, error) {
	req.URL.Scheme = t.target.Scheme
	req.URL.Host = t.target.Host
	req.Host = t.target.Host
	return http.DefaultTransport.RoundTrip(req)
}

// vidsrcHTTPClientForTest points the chain at the fake server for the test's
// lifetime and forces the fallback gate on under this test binary.
func vidsrcHTTPClientForTest(t *testing.T, srv *httptest.Server) *url.URL {
	t.Helper()
	t.Setenv("GOANIME_VIDSRC_FALLBACK_TEST", "1")
	target, err := url.Parse(srv.URL)
	require.NoError(t, err)
	prev := vidsrcHTTPFn
	vidsrcHTTPFn = func() *http.Client {
		return &http.Client{Transport: routeVIDSRC{target: target}}
	}
	t.Cleanup(func() { vidsrcHTTPFn = prev })
	return target
}

// vidsrcHTTPClientForTestPointNowhere — supplemental helper kept next to
// vidsrcHTTPClientForTest: a PLAIN client, no host rewrite, for tests that
// stage two REAL httptest bases and let the walk hit them directly.
func vidsrcHTTPClientForTestPlain(t *testing.T) {
	t.Helper()
	t.Setenv("GOANIME_VIDSRC_FALLBACK_TEST", "1")
	prev := vidsrcHTTPFn
	vidsrcHTTPFn = vidsrcHTTPDefault
	t.Cleanup(func() { vidsrcHTTPFn = prev })
}

// vidsrcFixtures loads the recorded decrypt fixtures: the wasm decryptor and
// the {stream_urls: encrypted blob, urls: plaintext lines} pair it resolves to.
func vidsrcFixtures(t *testing.T) (wasm []byte, enc string, want []string) {
	t.Helper()
	wasm, err := os.ReadFile(filepath.Join("testdata", "vidsrc_decrypt_test.wasm"))
	require.NoError(t, err, "the decryptor fixture must ship with the tests")
	raw, err := os.ReadFile(filepath.Join("testdata", "vidsrc_envelope.json"))
	require.NoError(t, err, "the envelope fixture must ship with the tests")
	var fx struct {
		StreamURLs string   `json:"stream_urls"`
		URLs       []string `json:"urls"`
	}
	require.NoError(t, json.Unmarshal(raw, &fx))
	require.NotEmpty(t, fx.URLs)
	return wasm, fx.StreamURLs, fx.URLs
}

// movieMasterPlaylist is what a playable master looks like; the RESOLUTION
// lines feed the height parse.
const movieMasterPlaylist = `#EXTM3U
#EXT-X-INDEPENDENT-SEGMENTS
#EXT-X-STREAM-INF:BANDWIDTH=3905000,CODECS="mp4a.40.2,avc1.640020",RESOLUTION=1920x1080
/master-v1/index.m3u8
#EXT-X-STREAM-INF:BANDWIDTH=1247000,CODECS="mp4a.40.2,avc1.42c01e",RESOLUTION=640x360
/master-v2/index.m3u8
`

// fakeVIDSRC hosts the whole chain in one handler:
//
//   - SRV placeholder in payloads becomes the test server's URL, so pages can
//     point at their own server;
//   - tokenFailFirst makes the first stream-data fetch answer 403 "invalid
//     api token" to exercise the single-use token refresh;
//   - a data call without stream_urls (the series defaults probe) gets the
//     eps-map JSON naming a default season/episode;
//   - mastersBroken makes every master request fail, so the chain reports
//     how many candidates it refused;
//   - generate.php always answers "testtoken", masters always play.
type fakeVIDSRC struct {
	kind           MediaKind
	id             string
	playerPages    []string // CONFIG snippets, one per render; the last repeats
	envJSON        string   // stream_urls envelope (encrypted form)
	plainEnvJSON   string   // same, but with stream_urls as a plain array
	tvMapJSON      string   // series defaults probe
	wasm           []byte
	tokenFailFirst bool
	token429First  bool
	mastersBroken  bool

	playerRenders int
	token429Seen  int32
	dataFetched   []string
	masterHits    int
}

func (f *fakeVIDSRC) start(t *testing.T) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		base := srv.URL
		body := func(s string) string { return strings.ReplaceAll(s, "SRV", base) }
		switch {
		case strings.HasPrefix(r.URL.Path, "/vs_src.php"):
			id := f.id
			if id == "" {
				id = "tt0315327"
			}
			assert.Equal(t, string(f.kind), r.URL.Query().Get("type"))
			assert.Equal(t, id, r.URL.Query().Get("id"))
			_, _ = w.Write([]byte(`{"src":"` + base + `/embed/` + string(f.kind) + `/` + id + `?vs=abc"}`))

		case strings.HasPrefix(r.URL.Path, "/embed/player/"):
			// Regular renders carry the landing referer; the single-use token
			// refresh re-fetches the same page in place, referrer-less as the
			// player's own refreshToken fetch does.
			ref := r.Header.Get("Referer")
			assert.True(t, strings.HasPrefix(ref, base+"/embed/") || ref == "",
				"the player render carries the landing referer or none, got %q", ref)
			f.playerRenders++
			idx := f.playerRenders - 1
			if idx >= len(f.playerPages) {
				idx = len(f.playerPages) - 1
			}
			_, _ = w.Write([]byte(body(f.playerPages[idx])))

		case strings.HasPrefix(r.URL.Path, "/embed/"):
			_, _ = w.Write([]byte(`<!doctype html><script>window.CFG = {"playerUrl":"` +
				`/embed/player/movie/tt0315327?vs=def"}</script>`))

		case strings.HasPrefix(r.URL.Path, "/wasm"):
			_, _ = w.Write(f.wasm)

		case strings.HasPrefix(r.URL.Path, "/data/api.php"):
			if _, ok := r.URL.Query()["stream_urls"]; !ok {
				// The series defaults probe: no token, no stream_urls.
				assert.Empty(t, r.URL.Query().Get("api_token"))
				_, _ = w.Write([]byte(body(f.tvMapJSON)))
				return
			}
			f.dataFetched = append(f.dataFetched, r.URL.RawQuery)
			if f.tokenFailFirst && len(f.dataFetched) == 1 {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"status_code":"403","error":"invalid api token"}`))
				return
			}
			payload := f.envJSON
			if payload == "" {
				payload = f.plainEnvJSON
			}
			_, _ = w.Write([]byte(body(payload)))

		case strings.HasSuffix(r.URL.Path, "/generate.php"):
			if f.token429First && atomic.AddInt32(&f.token429Seen, 1) == 1 {
				w.WriteHeader(http.StatusTooManyRequests)
				_, _ = w.Write([]byte(`{"error":"too many requests"}`))
				return
			}
			_, _ = w.Write([]byte("testtoken"))

		case strings.HasPrefix(r.URL.Path, "/pl/"):
			if f.mastersBroken {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			assert.Equal(t, "testtoken", r.URL.Query().Get("token"), "a master request carries the host token")
			f.masterHits++
			_, _ = w.Write([]byte(movieMasterPlaylist))

		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// jsonQuote encodes a Go string as the JSON string an envelope carries.
func jsonQuote(s string) string {
	b, err := json.Marshal(s)
	if err != nil {
		panic(err)
	}
	return string(b)
}

// fakeMovieConfig renders a player page: CONFIG with streamBase + token.
func fakeMovieConfig(streamBase, token string) string {
	return `<script>window.CONFIG = {"mediaType":"movie","api":"` + streamBase +
		`","apiToken":"` + token + `","apiTokenRefresh":false};</script>`
}

// fakeTVConfig is fakeMovieConfig for a series, keyed by streamBase.
func fakeTVConfig(streamBase, token string) string {
	return `<script>window.CONFIG = {"mediaType":"tv","idType":"tmdb","streamBase":"` + streamBase +
		`","apiToken":"` + token + `","apiTokenRefresh":false};</script>`
}

// TestVIDSRCMovieChainFixture runs the movie chain end to end against the
// recorded wasm fixture: the fake envelope carries the recorded encrypted
// blob, the real decryptor resolves it offline to the recorded masters, and
// the first one is token-stamped and probed.
func TestVIDSRCMovieChainFixture(t *testing.T) {
	wasm, enc, want := vidsrcFixtures(t)
	env := `{"status_code":"200","data":{"stream_urls":` + jsonQuote(enc) +
		`},"vs":{"w":5974048,"wasm_url":"SRV/wasm.bin"}}`
	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")
	f := &fakeVIDSRC{kind: KindMovie, playerPages: []string{playerPage}, envJSON: env, wasm: wasm}

	vidsrcHTTPClientForTest(t, f.start(t))

	stream, err := MovieFallback(context.Background(), "tt0315327", 310)
	require.NoError(t, err)

	// The first decrypted master wins — the recorded fixture's very first URL.
	assert.Equal(t, want[0]+"?token=testtoken", stream.URL)
	assert.Equal(t, 1080, stream.Height, "the tallest RESOLUTION line names the height")
	require.NotEmpty(t, f.dataFetched)
	assert.Contains(t, f.dataFetched[0], "api_token=tok1")
	assert.Contains(t, f.dataFetched[0], "stream_urls=")
	assert.Equal(t, 1, f.playerRenders, "one render, one single-use token")
	assert.Equal(t, 1, f.masterHits)
	assert.Equal(t, AudioSubtitled, stream.Audio, "the fallback is the original audio, never the dub")
}

// TestVIDSRCPlainEnvelopeArray covers the protection-off shape: URLs as a
// plain JSON array, no wasm, straight to token + probe.
func TestVIDSRCPlainEnvelopeArray(t *testing.T) {
	env := `{"status_code":"200","data":{"stream_urls":["SRV/pl/plain/master.m3u8"]}}`
	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")
	f := &fakeVIDSRC{kind: KindMovie, playerPages: []string{playerPage}, plainEnvJSON: env}

	base := vidsrcHTTPClientForTest(t, f.start(t))

	stream, err := MovieFallback(context.Background(), "tt0315327", 310)
	require.NoError(t, err)

	assert.Equal(t, base.String()+"/pl/plain/master.m3u8?token=testtoken", stream.URL)
	assert.Equal(t, 1080, stream.Height)
	require.Len(t, f.dataFetched, 1)
	assert.Contains(t, f.dataFetched[0], "api_token=tok1")
	assert.Equal(t, 1, f.masterHits)
}

// TestVIDSRCTVSeasonAndEpisode pins the series query shape: a caller-named
// season/episode lands in the stream-data call, and the defaults probe does
// not happen at all when both are known.
func TestVIDSRCTVSeasonAndEpisode(t *testing.T) {
	playerPage := fakeTVConfig("SRV/data/api.php?type=tv&tmdb=1429", "tok1")
	env := `{"status_code":"200","data":{"stream_urls":["SRV/pl/s2e3/master.m3u8"]}}`
	f := &fakeVIDSRC{kind: KindSeries, id: "1429", playerPages: []string{playerPage}, plainEnvJSON: env}

	vidsrcHTTPClientForTest(t, f.start(t))

	stream, err := EpisodeFallback(context.Background(), 1429, 2, 3)
	require.NoError(t, err)

	assert.Contains(t, stream.URL, "?token=testtoken")
	require.Len(t, f.dataFetched, 1, "no defaults probe when season and episode are named")
	q := f.dataFetched[0]
	assert.Contains(t, q, "season=2")
	assert.Contains(t, q, "episode=3")
	assert.Contains(t, q, "stream_urls=")
	assert.Contains(t, q, "api_token=tok1")
	assert.Equal(t, AudioSubtitled, stream.Audio, "an episode's fallback is the original audio too")
}

// TestVIDSRCTVDefaultsFromMap covers a caller without season/episode: the
// chain probes the series' own eps map and uses its defaults (its Pilot).
func TestVIDSRCTVDefaultsFromMap(t *testing.T) {
	playerPage := fakeTVConfig("SRV/data/api.php?type=tv&tmdb=1429", "tok1")
	f := &fakeVIDSRC{
		kind: KindSeries, id: "1429",
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/s1e2/master.m3u8"]}}`,
		tvMapJSON:    `{"status_code":"200","data":{"season":1,"episode":"2","eps":{"1":["1","2","3"]}}}`,
	}
	vidsrcHTTPClientForTest(t, f.start(t))

	stream, err := EpisodeFallback(context.Background(), 1429, 0, 0)
	require.NoError(t, err)

	assert.Contains(t, stream.URL, "?token=testtoken")
	require.Len(t, f.dataFetched, 1, "the map probe carries no stream_urls and is not recorded as one")
	q := f.dataFetched[0]
	assert.Contains(t, q, "season=1")
	assert.Contains(t, q, "episode=2")
}

// TestVIDSRCStaleTokenRefreshes exercises the single-use token: the first
// stream-data call answers 403 invalid token, the chain re-renders the player
// page, and the retry with the fresh token succeeds.
func TestVIDSRCStaleTokenRefreshes(t *testing.T) {
	f := &fakeVIDSRC{
		kind: KindMovie,
		playerPages: []string{
			fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1"),
			fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok2"),
		},
		plainEnvJSON:   `{"status_code":"200","data":{"stream_urls":["SRV/pl/md/master.m3u8"]}}`,
		tokenFailFirst: true,
	}
	vidsrcHTTPClientForTest(t, f.start(t))

	stream, err := MovieFallback(context.Background(), "tt0315327", 310)
	require.NoError(t, err)

	assert.Equal(t, 2, f.playerRenders, "a fresh render for the fresh token")
	require.Len(t, f.dataFetched, 2)
	assert.Contains(t, f.dataFetched[0], "api_token=tok1")
	assert.Contains(t, f.dataFetched[1], "api_token=tok2")
	assert.NotEmpty(t, stream.URL)
}

// TestVIDSRCUnlistedTitleIsHonest covers a title the fallback source knows
// nothing about: the chain stops at the envelope and reports it, and the
// caller keeps its original panel error (asserted in the api tests).
func TestVIDSRCUnlistedTitleIsHonest(t *testing.T) {
	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")
	f := &fakeVIDSRC{
		kind:         KindMovie,
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":""}}`,
	}
	vidsrcHTTPClientForTest(t, f.start(t))

	_, err := MovieFallback(context.Background(), "tt0315327", 310)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no stream_urls")
	assert.Zero(t, f.masterHits, "no master probe on an empty listing")
}

// TestVIDSRCUnplayableMastersReported covers every listed master answering
// dead: the chain refuses all of them and fingers the count in the error.
func TestVIDSRCUnplayableMastersReported(t *testing.T) {
	wasm, enc, want := vidsrcFixtures(t)
	require.Greater(t, len(want), 1, "the fixture lists several rotating copies")
	env := `{"status_code":"200","data":{"stream_urls":` + jsonQuote(enc) +
		`},"vs":{"w":5974048,"wasm_url":"SRV/wasm.bin"}}`
	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")
	f := &fakeVIDSRC{kind: KindMovie, playerPages: []string{playerPage}, envJSON: env, wasm: wasm, mastersBroken: true}

	vidsrcHTTPClientForTest(t, f.start(t))

	_, err := MovieFallback(context.Background(), "tt0315327", 310)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none playable")
	assert.Contains(t, err.Error(), strconv.Itoa(len(want)), "the error names how many URLs were listed")
	assert.Zero(t, f.masterHits)
}

// TestVIDSRCDecryptFixture pins the decryptor against the recorded pair so an
func TestVIDSRCDecryptFixture(t *testing.T) {
	wasm, enc, want := vidsrcFixtures(t)
	got, err := decryptVIDSRC(wasm, enc)
	require.NoError(t, err)
	assert.Equal(t, want, got)

	_, err = decryptVIDSRC([]byte("not a wasm module"), enc)
	require.Error(t, err)
	_, err = decryptVIDSRC(wasm, "!!!!not base64!!!!")
	require.Error(t, err)
}

// TestVIDSRCFallbackGate pins the gate: without GOANIME_VIDSRC_FALLBACK_TEST,
// a test binary must not reach the fallback (regression tests exercise panel
// failures against fixtures and would otherwise pull the real source).
func TestVIDSRCFallbackGate(t *testing.T) {
	t.Setenv("GOANIME_VIDSRC_FALLBACK_TEST", "")
	_, err := MovieFallback(context.Background(), "tt1", 0)
	require.ErrorIs(t, err, errVIDSRCDisabled)
	_, err = EpisodeFallback(context.Background(), 310, 1, 1)
	require.ErrorIs(t, err, errVIDSRCDisabled)
}

// TestVIDSRCFallbackSecondBaseRecovers pins the mirror walk: the primary
// front dying (stage-1) does not kill the source — the fallback walks the
// next base and the title still resolves from it.
func TestVIDSRCFallbackSecondBaseRecovers(t *testing.T) {
	var primaryHits int
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		primaryHits++
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer primary.Close()

	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok2")
	second := &fakeVIDSRC{
		kind:         KindMovie,
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/mirror/master.m3u8"]}}`,
	}
	secondSrv := second.start(t)

	prevBases := vidsrcBasesFn
	vidsrcBasesFn = func() []string { return []string{primary.URL, secondSrv.URL} }
	t.Cleanup(func() { vidsrcBasesFn = prevBases })
	vidsrcHTTPClientForTestPlain(t)

	stream, err := MovieFallback(context.Background(), "tt0315327", 310)
	require.NoError(t, err)

	assert.Greater(t, primaryHits, 0, "the walk starts at the primary")
	assert.Equal(t, secondSrv.URL+"/pl/mirror/master.m3u8?token=testtoken", stream.URL)
	assert.Equal(t, 1, second.playerRenders)
}

// TestVIDSRCFallbackSecondBaseRecoversSeries pins the same walk for series.
func TestVIDSRCFallbackSecondBaseRecoversSeries(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer primary.Close()

	playerPage := fakeTVConfig("SRV/data/api.php?type=tv&imdb=1429", "tok3")
	second := &fakeVIDSRC{
		kind:         KindSeries,
		id:           "1429",
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/mirror/s2e3/master.m3u8"]}}`,
	}
	secondSrv := second.start(t)

	prevBases := vidsrcBasesFn
	vidsrcBasesFn = func() []string { return []string{primary.URL, secondSrv.URL} }
	t.Cleanup(func() { vidsrcBasesFn = prevBases })
	vidsrcHTTPClientForTestPlain(t)

	_, err := EpisodeFallback(context.Background(), 1429, 2, 3)
	require.NoError(t, err)
	assert.Equal(t, 1, second.playerRenders)
}

// TestVIDSRCFallbackAllBasesMissKeepsLastError: both primaries down — the
// last base's own diagnosis is what the caller sees.
func TestVIDSRCFallbackAllBasesMissKeepsLastError(t *testing.T) {
	primary := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer primary.Close()

	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tokX")
	second := &fakeVIDSRC{kind: KindMovie, playerPages: []string{playerPage}, mastersBroken: true,
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/dead/master.m3u8"]}}`}
	secondSrv := second.start(t)

	prevBases := vidsrcBasesFn
	vidsrcBasesFn = func() []string { return []string{primary.URL, secondSrv.URL} }
	t.Cleanup(func() { vidsrcBasesFn = prevBases })
	vidsrcHTTPClientForTestPlain(t)

	_, err := MovieFallback(context.Background(), "tt0315327", 310)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "playable", "the walk ends on the last base's diagnosis")
}

// TestVIDSRCRateLimitedTokenRecovers pins the burst case: the host-token
// endpoint answers 429 on the first mint, and after one short pause the
// chain resolves anyway — a replay two seconds later must not read as a
// dead title.
func TestVIDSRCRateLimitedTokenRecovers(t *testing.T) {
	wasm, enc, want := vidsrcFixtures(t)
	env := `{"status_code":"200","data":{"stream_urls":` + jsonQuote(enc) +
		`},"vs":{"w":5974048,"wasm_url":"SRV/wasm.bin"}}`
	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")
	f := &fakeVIDSRC{kind: KindMovie, playerPages: []string{playerPage}, envJSON: env, wasm: wasm, token429First: true}

	vidsrcHTTPClientForTest(t, f.start(t))

	prevPause := vidsrcRatePause
	vidsrcRatePause = 50 * time.Millisecond
	t.Cleanup(func() { vidsrcRatePause = prevPause })

	stream, err := MovieFallback(context.Background(), "tt0315327", 310)
	require.NoError(t, err)
	assert.Equal(t, want[0]+"?token=testtoken", stream.URL)
	assert.Equal(t, int32(2), f.token429Seen, "one refusal, one retry, then the mint")
}

// TestVIDSRCLiveTodoPoderoso is the opt-in integration check for the user's
// report: the fallback must resolve Bruce Almighty (2003) headlessly. Run
// with GOANIME_LIVE=1; it exercises the real chain, so it needs the network.
func TestVIDSRCLiveTodoPoderoso(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") != "1" {
		t.Skip("live test; run with GOANIME_LIVE=1")
	}
	t.Setenv("GOANIME_VIDSRC_FALLBACK_TEST", "1")
	prev := vidsrcHTTPFn
	vidsrcHTTPFn = vidsrcHTTPDefault
	t.Cleanup(func() { vidsrcHTTPFn = prev })

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	stream, err := MovieFallback(ctx, "tt0315327", 310)
	require.NoError(t, err, "the fallback must resolve the title its panel could not")

	assert.NotEmpty(t, stream.Height, "the master names a real resolution")
	assert.NotEmpty(t, stream.Host)
	assert.NotEmpty(t, stream.Referer)

	// The catalog entry is "dublado" and the user asked for PT-BR audio:
	// log what the chain actually serves before promising anything.
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, stream.URL, http.NoBody)
	require.NoError(t, err)
	req.Header.Set("Referer", stream.Referer)
	client := vidsrcHTTPFn()
	if client == nil {
		client = http.DefaultClient
	}
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer resp.Body.Close()
	body, err := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	require.NoError(t, err)
	require.Equal(t, http.StatusOK, resp.StatusCode, "master status: %d", resp.StatusCode)
	variants, audios := countMasterEntries(body)
	t.Logf("master: %d variants, audio groups: %v", variants, audios)
	// Stream.Audio says Legendado because the master carries only the
	// original track (2026-10-21: no audio groups at all). A Portuguese
	// rendition here would make that a lie and the swap question needless.
	for _, a := range audios {
		lower := strings.ToLower(a)
		assert.False(t, strings.Contains(lower, `language="pt`) || strings.Contains(lower, `language="por`),
			"the vidsrc master now carries Portuguese audio; Stream.Audio must stop saying Legendado: %s", a)
	}
}

// countMasterEntries summarizes a master playlist: the variant URL count and
// any audio groups it declares, so a live run says in one line whether the
// title plays with PT-BR audio or the embed's own default.
func countMasterEntries(body []byte) (variants int, audios []string) {
	for _, line := range strings.Split(string(body), "\n") {
		line = strings.TrimSpace(line)
		if strings.HasPrefix(line, "#EXT-X-MEDIA:TYPE=AUDIO") {
			audios = append(audios, line)
			continue
		}
		if line != "" && !strings.HasPrefix(line, "#") {
			variants++
		}
	}
	return variants, audios
}
