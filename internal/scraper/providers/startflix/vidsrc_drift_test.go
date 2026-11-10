package startflix

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests pin the three drifts the vidsrc family showed on 2026-11-10,
// which together kept every series (and most movies) from resolving: the
// series vocabulary of vs_src, the api token the host token endpoint wants,
// and a front (vidsrc.io) whose origin never answers.

// TestVIDSRCTypeWord pins the vs_src vocabulary: the endpoint keys a series as
// "tv". GoAnime's own "series" word is answered with the MOVIE page of the
// same TMDB id (live: type=series&id=1396 landed on /embed/movie/1396, not
// Breaking Bad), so it never goes on the wire.
func TestVIDSRCTypeWord(t *testing.T) {
	assert.Equal(t, "tv", vidsrcTypeWord(KindSeries))
	assert.Equal(t, "movie", vidsrcTypeWord(KindMovie))
}

// TestVIDSRCSeriesAsksTVType runs a series through the chain and records the
// type vs_src saw.
func TestVIDSRCSeriesAsksTVType(t *testing.T) {
	var sawType string
	f := &fakeVIDSRC{
		kind: KindSeries, id: "1396",
		playerPages:  []string{fakeTVConfig("SRV/data/api.php?type=tv&tmdb=1396", "tok1")},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/bb/master.m3u8"]}}`,
	}
	srv := f.start(t)
	inner := srv.Config.Handler
	srv.Config.Handler = http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/vs_src.php") {
			sawType = r.URL.Query().Get("type")
		}
		inner.ServeHTTP(w, r)
	})
	vidsrcHTTPClientForTest(t, srv)

	_, err := EpisodeFallback(context.Background(), 1396, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, "tv", sawType)
}

// TestVIDSRCHostTokenCarriesFreshAPIToken pins the host token request: the
// stream hosts' generate.php answers 403 "invalid or mismatched api token"
// unless it carries a player-page apiToken that the stream-data call has not
// spent. The chain re-renders the player page for one and sends it.
func TestVIDSRCHostTokenCarriesFreshAPIToken(t *testing.T) {
	f := &fakeVIDSRC{
		kind:         KindMovie,
		playerPages:  []string{fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/a/master.m3u8"]}}`,
	}
	vidsrcHTTPClientForTest(t, f.start(t))

	stream, err := MovieFallback(context.Background(), "tt0315327", 310)
	require.NoError(t, err)
	assert.Contains(t, stream.URL, "?token=testtoken")
	require.Len(t, f.hostAuths, 1)
	assert.NotEmpty(t, f.hostAuths[0])
	assert.NotEqual(t, "tok1", f.hostAuths[0], "the token spent on stream_urls is dead for generate.php")
}

// TestVIDSRCHostTokenStaleRendersOnce covers a host token that went stale
// anyway: one 403 triggers exactly one more render and one retry, which
// succeeds; the title still resolves.
func TestVIDSRCHostTokenStaleRendersOnce(t *testing.T) {
	f := &fakeVIDSRC{
		kind:               KindMovie,
		playerPages:        []string{fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")},
		plainEnvJSON:       `{"status_code":"200","data":{"stream_urls":["SRV/pl/a/master.m3u8"]}}`,
		generateStaleFirst: true,
	}
	vidsrcHTTPClientForTest(t, f.start(t))

	stream, err := MovieFallback(context.Background(), "tt0315327", 310)
	require.NoError(t, err)
	assert.Contains(t, stream.URL, "?token=testtoken")
	require.Len(t, f.hostAuths, 2, "one refused, one retried")
	assert.NotEqual(t, f.hostAuths[0], f.hostAuths[1], "the retry carries a fresh token")
	assert.Equal(t, 3, f.playerRenders, "stream-data token, host token, retry token")
}

// hangingServer accepts requests and never answers them: an origin that is
// down behind a still-listening front.
func hangingServer(t *testing.T) *httptest.Server {
	t.Helper()
	release := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		select {
		case <-release:
		case <-r.Context().Done():
		}
	}))
	t.Cleanup(func() { close(release); srv.Close() })
	return srv
}

// TestVIDSRCHangingBaseDoesNotBurnTheWalk pins the dead-base fast path: a
// front whose origin never answers (vidsrc.io: Cloudflare 522 after ~20s) is
// cut at the per-base timeout, the walk moves on, and the title resolves from
// the next base.
func TestVIDSRCHangingBaseDoesNotBurnTheWalk(t *testing.T) {
	hang := hangingServer(t)
	second := &fakeVIDSRC{
		kind:         KindMovie,
		playerPages:  []string{fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok2")},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/mirror/master.m3u8"]}}`,
	}
	secondSrv := second.start(t)

	prevBases, prevTimeout := vidsrcBasesFn, vidsrcBaseTimeout
	vidsrcBasesFn = func() []string { return []string{hang.URL, secondSrv.URL} }
	vidsrcBaseTimeout = 150 * time.Millisecond
	t.Cleanup(func() { vidsrcBasesFn, vidsrcBaseTimeout = prevBases, prevTimeout })
	vidsrcHTTPClientForTestPlain(t)
	// A client with no timeout of its own: only the per-base bound can cut
	// the hanging front.
	prevHTTP := vidsrcHTTPFn
	vidsrcHTTPFn = func() *http.Client { return &http.Client{} }
	t.Cleanup(func() { vidsrcHTTPFn = prevHTTP })

	start := time.Now()
	stream, err := MovieFallback(context.Background(), "tt0315327", 310)
	require.NoError(t, err)
	assert.Contains(t, stream.URL, "/pl/mirror/master.m3u8")
	assert.Less(t, time.Since(start), 3*time.Second, "a hanging base is cut at its own timeout")
}

// TestVIDSRCHangingBaseNamedInError: when every base is dead the error names
// the base that did not answer and how long it was given, instead of a bare
// "context deadline exceeded".
func TestVIDSRCHangingBaseNamedInError(t *testing.T) {
	hang := hangingServer(t)

	prevBases, prevTimeout := vidsrcBasesFn, vidsrcBaseTimeout
	vidsrcBasesFn = func() []string { return []string{hang.URL} }
	vidsrcBaseTimeout = 100 * time.Millisecond
	t.Cleanup(func() { vidsrcBasesFn, vidsrcBaseTimeout = prevBases, prevTimeout })
	vidsrcHTTPClientForTestPlain(t)
	prevHTTP := vidsrcHTTPFn
	vidsrcHTTPFn = func() *http.Client { return &http.Client{} }
	t.Cleanup(func() { vidsrcHTTPFn = prevHTTP })

	_, err := MovieFallback(context.Background(), "tt0315327", 310)
	require.Error(t, err)
	assert.Contains(t, err.Error(), hang.URL)
	assert.Contains(t, err.Error(), "did not answer within")
}
