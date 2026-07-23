// Package allanime — regression suite for the dynamic per-epoch key derivation
// introduced with the mkissa migration (ani-cli PR #1779, ported 2026-07-22).
//
// What broke and what these pin:
//   - AllAnime dropped the static AES key. The key is now derived per epoch:
//     scrape the referer page for `epoch`/`partB`, follow the entry bundle to a
//     code-split chunk holding a 64-hex "mask", then key = mask XOR partB.
//     TestFetchAAKeys_* pin that whole flow end-to-end against a mock bundle so a
//     future upstream reshuffle fails loudly here instead of silently returning
//     an unusable key (which surfaces downstream as "no source URLs").
//   - The session bug that motivated this file: source extraction now needs a
//     key from getAAKeys, and a client WITHOUT a cached key would reach out to
//     the live bundle over the network — even inside tests. TestGetAAKeys_*
//     pin that a cached/injected key is reused with ZERO network, and that a
//     fresh derive is cached (never re-scraped within its TTL).
package allanime

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockKeyBundleServer stands in for the whole mkissa key-bundle surface: the
// referer page (epoch/partB + entry bundle URL), the entry bundle (imports a
// chunk), and the chunk (embeds the 64-hex mask). Returns the server plus a
// live request counter so tests can assert the network is (not) touched.
func mockKeyBundleServer(t *testing.T, epoch string, partB, mask []byte) (*httptest.Server, *int32) {
	t.Helper()
	var hits int32
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&hits, 1)
		switch r.URL.Path {
		case "/entry/app.abc123.js":
			_, _ = fmt.Fprint(w, `import "../chunks/deadbeef.js";`)
		case "/chunks/deadbeef.js":
			_, _ = fmt.Fprintf(w, `var m="%s";`, hex.EncodeToString(mask))
		default: // referer page
			_, _ = fmt.Fprintf(w, `x={"epoch":%s,"partB":"%s"};var s="%s/entry/app.abc123.js";`,
				epoch, base64.StdEncoding.EncodeToString(partB), srv.URL)
		}
	}))
	t.Cleanup(srv.Close)
	return srv, &hits
}

// randBytes returns n cryptographically-random bytes for fixture material.
func randBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	_, err := io.ReadFull(rand.Reader, b)
	require.NoError(t, err)
	return b
}

// xor returns a[i]^b[i] — the exact key-derivation step fetchAAKeys performs.
func xor(a, b []byte) []byte {
	out := make([]byte, len(a))
	for i := range out {
		out[i] = a[i] ^ b[i]
	}
	return out
}

// failTransport fails every request and counts attempts. Injected to PROVE a
// code path performs no network I/O (the exact regression this file guards).
type failTransport struct{ hits *int32 }

func (f failTransport) RoundTrip(*http.Request) (*http.Response, error) {
	atomic.AddInt32(f.hits, 1)
	return nil, errors.New("network disabled in test")
}

// ---------------------------------------------------------------------------
// getText
// ---------------------------------------------------------------------------

func TestGetText_ReturnsBodyOnSuccess(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, "hello-bundle")
	}))
	t.Cleanup(srv.Close)

	c := &AllAnimeClient{client: util.GetFastClient(), referer: srv.URL, userAgent: netx.UserAgent}
	body, err := c.getText(srv.URL)
	require.NoError(t, err)
	assert.Equal(t, "hello-bundle", body)
}

func TestGetText_ErrorsOnNon200(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	c := &AllAnimeClient{client: util.GetFastClient(), referer: srv.URL, userAgent: netx.UserAgent}
	_, err := c.getText(srv.URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "status 500")
}

// ---------------------------------------------------------------------------
// fetchAAKeys — the protocol pin
// ---------------------------------------------------------------------------

// TestFetchAAKeys_DerivesKeyFromMaskXorPartB pins the entire ani-cli PR #1779
// derivation: page → entry bundle → chunk mask, key = mask XOR partB, epoch
// carried through. Drift in ANY step (regex, chunk resolution, XOR) fails here.
func TestFetchAAKeys_DerivesKeyFromMaskXorPartB(t *testing.T) {
	t.Parallel()
	partB := randBytes(t, 32)
	mask := randBytes(t, 32)
	srv, _ := mockKeyBundleServer(t, "4128", partB, mask)

	c := &AllAnimeClient{client: util.GetFastClient(), referer: srv.URL, userAgent: netx.UserAgent}
	keys, err := c.fetchAAKeys()
	require.NoError(t, err)
	assert.Equal(t, xor(mask, partB), keys.key, "key must be mask XOR partB")
	assert.Len(t, keys.key, 32, "derived key must be a 32-byte AES-256 key")
	assert.Equal(t, "4128", keys.epoch)
}

// TestFetchAAKeys_ErrorsWhenPagePiecesMissing pins the guard: a page lacking
// epoch/partB/app-bundle must error, not return a half-derived key.
func TestFetchAAKeys_ErrorsWhenPagePiecesMissing(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `x={"epoch":1};`) // no partB, no app bundle
	}))
	t.Cleanup(srv.Close)

	c := &AllAnimeClient{client: util.GetFastClient(), referer: srv.URL, userAgent: netx.UserAgent}
	_, err := c.fetchAAKeys()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "epoch/partB/app")
}

// TestFetchAAKeys_ErrorsWhenMaskAbsent pins the guard for a chunk that carries
// no 64-hex mask — derivation must fail rather than XOR against garbage.
func TestFetchAAKeys_ErrorsWhenMaskAbsent(t *testing.T) {
	t.Parallel()
	partB := randBytes(t, 32)
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/entry/app.abc123.js":
			_, _ = fmt.Fprint(w, `import "../chunks/deadbeef.js";`)
		case "/chunks/deadbeef.js":
			_, _ = fmt.Fprint(w, `var m="tooshort";`) // no 64-hex mask
		default:
			_, _ = fmt.Fprintf(w, `x={"epoch":7,"partB":"%s"};var s="%s/entry/app.abc123.js";`,
				base64.StdEncoding.EncodeToString(partB), srv.URL)
		}
	}))
	t.Cleanup(srv.Close)

	c := &AllAnimeClient{client: util.GetFastClient(), referer: srv.URL, userAgent: netx.UserAgent}
	_, err := c.fetchAAKeys()
	require.Error(t, err)
	assert.Contains(t, err.Error(), "mask not found")
}

// ---------------------------------------------------------------------------
// getAAKeys — caching / no-network guarantees
// ---------------------------------------------------------------------------

// TestGetAAKeys_ReturnsCachedKeyWithoutNetwork pins the exact session bug: a
// client with a live (unexpired) key must return it and perform NO network I/O.
// A failing transport counts any request — the count must stay zero.
func TestGetAAKeys_ReturnsCachedKeyWithoutNetwork(t *testing.T) {
	t.Parallel()
	var hits int32
	c := &AllAnimeClient{
		client:  &http.Client{Transport: failTransport{&hits}},
		referer: "http://unused.invalid",
		keys:    &aaKeys{key: allAnimeKey, epoch: testAAEpoch},
		keysExp: time.Now().Add(time.Hour),
	}
	got, err := c.getAAKeys()
	require.NoError(t, err)
	assert.Equal(t, allAnimeKey, got.key)
	assert.Equal(t, testAAEpoch, got.epoch)
	assert.Zero(t, atomic.LoadInt32(&hits), "cached key must not trigger any network request")
}

// TestGetAAKeys_DerivesAndCachesOnFirstCall pins that a first call scrapes and
// derives, and a second call inside the TTL is served from cache (no re-scrape).
func TestGetAAKeys_DerivesAndCachesOnFirstCall(t *testing.T) {
	t.Parallel()
	partB := randBytes(t, 32)
	mask := randBytes(t, 32)
	srv, hits := mockKeyBundleServer(t, "4128", partB, mask)

	c := &AllAnimeClient{client: util.GetFastClient(), referer: srv.URL, userAgent: netx.UserAgent}

	k1, err := c.getAAKeys()
	require.NoError(t, err)
	assert.Equal(t, xor(mask, partB), k1.key)
	afterFirst := atomic.LoadInt32(hits)
	require.Positive(t, afterFirst, "first call must scrape the bundle")

	k2, err := c.getAAKeys()
	require.NoError(t, err)
	assert.Same(t, k1, k2, "second call must return the cached key pointer")
	assert.Equal(t, afterFirst, atomic.LoadInt32(hits), "second call must not re-scrape within TTL")
}

// TestGetAAKeys_RefetchesAfterExpiry pins TTL behaviour: an expired cache entry
// is discarded and a fresh key derived (the epoch rotates server-side).
func TestGetAAKeys_RefetchesAfterExpiry(t *testing.T) {
	t.Parallel()
	partB := randBytes(t, 32)
	mask := randBytes(t, 32)
	srv, hits := mockKeyBundleServer(t, "4128", partB, mask)

	c := &AllAnimeClient{
		client:    util.GetFastClient(),
		referer:   srv.URL,
		userAgent: netx.UserAgent,
		keys:      &aaKeys{key: allAnimeKey, epoch: "stale"},
		keysExp:   time.Now().Add(-time.Minute), // already expired
	}

	keys, err := c.getAAKeys()
	require.NoError(t, err)
	assert.Equal(t, "4128", keys.epoch, "expired entry must be replaced with the freshly-scraped epoch")
	assert.Equal(t, xor(mask, partB), keys.key)
	assert.Positive(t, atomic.LoadInt32(hits), "expiry must force a re-scrape")
}
