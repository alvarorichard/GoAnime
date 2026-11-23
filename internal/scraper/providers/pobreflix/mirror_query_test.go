package pobreflix

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"net/url"
	"strings"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// On 2026-11-23 the old default mirror (pobreflixtvfree.cv) answered
// /api/busca?q=icarly with a 301 to https://pobreflixtvfree.cc/api — the query
// string gone — and the new host answered the bare endpoint with "[]". Every
// first search of a run came back empty, Pobreflix showed nothing, and only the
// next request (already on .cc) worked. These pin the three parts of that: the
// default host, the re-ask when a mirror redirect drops the query, and the
// redirect loop the old /api?action=catalog endpoint turned into.

const icarlyBuscaJSON = `[{"title":"iCarly","year":2007,"type":"series","url":"/series/online/icarly-dublado-5371","poster":"x"}]`

// buscaMirrors returns an old mirror that redirects without the query and a
// new one that answers with results only when the query arrives.
func buscaMirrors(t *testing.T, newHits, oldHits *atomic.Int32) (oldURL, newURL string) {
	t.Helper()
	fresh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		newHits.Add(1)
		if r.URL.Query().Get("q") == "" {
			_, _ = w.Write([]byte("[]"))
			return
		}
		_, _ = w.Write([]byte(icarlyBuscaJSON))
	}))
	t.Cleanup(fresh.Close)
	old := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		oldHits.Add(1)
		http.Redirect(w, r, fresh.URL+r.URL.Path, http.StatusMovedPermanently) // the query is dropped
	}))
	t.Cleanup(old.Close)
	return old.URL, fresh.URL
}

func TestClientAsksTheNewMirrorWhenARedirectDropsTheQuery(t *testing.T) {
	var newHits, oldHits atomic.Int32
	oldURL, newURL := buscaMirrors(t, &newHits, &oldHits)
	c := NewClientForTest(http.DefaultClient, oldURL, t.TempDir()+"/corpus.json")

	body, err := c.get(context.Background(), oldURL+"/api/busca?q=icarly", getOptions{accept: "application/json"})
	require.NoError(t, err)
	assert.Contains(t, string(body), "iCarly", "the answer is the search's, not the bare endpoint's")
	assert.Equal(t, newURL, c.BaseURL(), "the client moved to the live mirror")
	assert.EqualValues(t, 1, oldHits.Load())
	assert.EqualValues(t, 2, newHits.Load(), "the bare answer, then the real request")
}

func TestBuscaSearchOnTheOldMirrorFindsTheTitle(t *testing.T) {
	var newHits, oldHits atomic.Int32
	oldURL, _ := buscaMirrors(t, &newHits, &oldHits)
	c := NewClientForTest(http.DefaultClient, oldURL, t.TempDir()+"/corpus.json")
	got, err := c.buscaSearch(context.Background(), "icarly")
	require.NoError(t, err)
	require.Len(t, got, 1, "the first search of a run must not come back empty")
	assert.Equal(t, "iCarly", got[0].Title)
	assert.Equal(t, 5371, got[0].VideoID)
}

func TestClientDoesNotRehostWhenTheRedirectKeepsTheQuery(t *testing.T) {
	var hits atomic.Int32
	fresh := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		_, _ = w.Write([]byte(icarlyBuscaJSON))
	}))
	t.Cleanup(fresh.Close)
	old := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, fresh.URL+r.URL.RequestURI(), http.StatusMovedPermanently)
	}))
	t.Cleanup(old.Close)
	c := NewClientForTest(http.DefaultClient, old.URL, t.TempDir()+"/corpus.json")
	_, err := c.get(context.Background(), old.URL+"/api/busca?q=icarly", getOptions{})
	require.NoError(t, err)
	assert.EqualValues(t, 1, hits.Load(), "nothing was lost, so nothing is asked twice")
}

func TestClientDoesNotRehostARequestWithoutAQuery(t *testing.T) {
	var newHits, oldHits atomic.Int32
	oldURL, _ := buscaMirrors(t, &newHits, &oldHits)
	c := NewClientForTest(http.DefaultClient, oldURL, t.TempDir()+"/corpus.json")
	_, err := c.get(context.Background(), oldURL+"/api/busca", getOptions{})
	require.NoError(t, err)
	assert.EqualValues(t, 1, newHits.Load())
}

// Two mirrors that each send a request carrying the query to the other, with
// the query dropped, must end, not ping-pong. Each answers the bare endpoint
// itself. Past a few hits both stop, so a client with no guard ends the test
// instead of hanging it.
func TestClientRehostsAtMostOnce(t *testing.T) {
	var total atomic.Int32
	var aURL, bURL string
	bounce := func(to *string) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			if total.Add(1) > 30 || r.URL.Query().Get("q") == "" {
				_, _ = w.Write([]byte("[]"))
				return
			}
			http.Redirect(w, r, *to+r.URL.Path, http.StatusMovedPermanently)
		}
	}
	a := httptest.NewServer(bounce(&bURL))
	t.Cleanup(a.Close)
	b := httptest.NewServer(bounce(&aURL))
	t.Cleanup(b.Close)
	aURL, bURL = a.URL, b.URL
	c := NewClientForTest(http.DefaultClient, aURL, t.TempDir()+"/corpus.json")
	_, _ = c.get(context.Background(), aURL+"/api/busca?q=x", getOptions{})
	assert.LessOrEqual(t, total.Load(), int32(6), "one ask and one re-ask, never a chain")
}

func TestRehostedURL(t *testing.T) {
	t.Parallel()
	parse := func(s string) *url.URL {
		u, err := url.Parse(s)
		require.NoError(t, err)
		return u
	}
	for _, tc := range []struct {
		name, requested, final, want string
	}{
		{"a mirror that dropped the query", "https://pobreflixtvfree.cv/api/busca?q=icarly", "https://pobreflixtvfree.cc/api/busca", "https://pobreflixtvfree.cc/api/busca?q=icarly"},
		{"the path and every parameter come along", "https://pobreflixtvfree.cv/api?action=catalog&type=series&search=a%20b", "https://pobreflixtvfree.cc/api", "https://pobreflixtvfree.cc/api?action=catalog&type=series&search=a%20b"},
		{"the query survived", "https://pobreflixtvfree.cv/api/busca?q=icarly", "https://pobreflixtvfree.cc/api/busca?q=icarly", ""},
		{"nothing was asked", "https://pobreflixtvfree.cv/api/busca", "https://pobreflixtvfree.cc/api/busca", ""},
		{"another site", "https://pobreflixtvfree.cv/api/busca?q=icarly", "https://evil.example/api/busca", ""},
		{"a downgrade to http", "https://pobreflixtvfree.cv/api/busca?q=icarly", "http://pobreflixtvfree.cc/api/busca", ""},
		{"the same host", "https://pobreflixtvfree.cc/api/busca?q=icarly", "https://pobreflixtvfree.cc/api/busca", ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			got, ok := rehostedURL(parse(tc.requested), parse(tc.final))
			assert.Equal(t, tc.want != "", ok)
			assert.Equal(t, tc.want, got)
		})
	}
	_, ok := rehostedURL(nil, parse("https://pobreflixtvfree.cc/"))
	assert.False(t, ok)
	_, ok = rehostedURL(parse("https://pobreflixtvfree.cv/?q=1"), nil)
	assert.False(t, ok)
}

// TestDefaultBaseIsNotTheRedirectingMirror pins the host the client starts on:
// the old .cv redirected without the query string.
func TestDefaultBaseIsNotTheRedirectingMirror(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "https://pobreflixtvfree.cc", DefaultBase)
}

func TestIsRedirectLoop(t *testing.T) {
	t.Parallel()
	assert.True(t, isRedirectLoop(http.ErrUseLastResponse))
	assert.True(t, isRedirectLoop(errors.New(`Get "https://x/api?a=b": net/http: use last response`)))
	assert.True(t, isRedirectLoop(errors.New(`Get "https://x": stopped after 10 redirects`)))
	assert.False(t, isRedirectLoop(errors.New("connection refused")))
	assert.False(t, isRedirectLoop(context.DeadlineExceeded))
	assert.False(t, isRedirectLoop(nil))
}

// The old /api?action=catalog endpoint now redirects to /api/ and back. A loop
// answers the same way every time, so it must be tried once, not retried.
func TestClientDoesNotRetryARedirectLoop(t *testing.T) {
	var hits atomic.Int32
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		next := "/api/"
		if strings.HasPrefix(r.URL.Path, "/api/") {
			next = "/api"
		}
		http.Redirect(w, r, srv.URL+next+"?"+r.URL.RawQuery, http.StatusMovedPermanently)
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(http.DefaultClient, srv.URL, t.TempDir()+"/corpus.json")
	c.maxRetries = 2
	c.retryDelay = 1

	_, err := c.get(context.Background(), srv.URL+"/api?action=catalog&search=icarly", getOptions{})
	require.Error(t, err)
	assert.LessOrEqual(t, hits.Load(), int32(11), "one pass through the client's redirect limit, not three")
}
