package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// scriptedSolver is a deterministic, browser-free stand-in for the headed CF
// solver. It satisfies BOTH cfSolver (Solve) and embedStreamSolver
// (SniffEmbedStream) so it exercises every fallback fork in GetEpisodes /
// GetStreamURL without launching Chromium. All calls are recorded so a test can
// assert the exact cascade order.
type scriptedSolver struct {
	mu sync.Mutex

	solveByURL map[string]*CFSolveResult
	solveErr   map[string]error
	solveCalls []string

	sniff      *CFStreamResult
	sniffErr   error
	sniffCalls []string
}

func (s *scriptedSolver) Solve(_ context.Context, u string, _ time.Duration) (*CFSolveResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.solveCalls = append(s.solveCalls, u)
	if err, ok := s.solveErr[u]; ok {
		return nil, err
	}
	if r, ok := s.solveByURL[u]; ok {
		return r, nil
	}
	return &CFSolveResult{}, nil
}

func (s *scriptedSolver) SniffEmbedStream(_ context.Context, u string, _ time.Duration) (*CFStreamResult, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sniffCalls = append(s.sniffCalls, u)
	if s.sniffErr != nil {
		return nil, s.sniffErr
	}
	return s.sniff, nil
}

func (s *scriptedSolver) calls() ([]string, []string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	return append([]string(nil), s.solveCalls...), append([]string(nil), s.sniffCalls...)
}

// --- getEpisodesViaBrowser cascade -----------------------------------------

// frontend serie page that carries ONLY anchors (no window.allEpisodes / no
// ALL_EPISODES), forcing the per-season fetch-and-merge cascade.
const sfSeason1Anchors = `<!doctype html><html><body>
<nav>
 <a href="/serie/demo-show/1">T1</a>
 <a href="/serie/demo-show/2">T2</a>
 <a href="/serie/demo-show/3">T3</a>
</nav>
<a data-episode-id="11" data-season="1" data-episode="1">E1</a>
<a data-episode-id="12" data-season="1" data-episode="2">E2</a>
</body></html>`

const sfSeason2Anchors = `<!doctype html><html><body>
<a data-episode-id="21" data-season="2" data-episode="1">E1</a>
</body></html>`

func TestGetEpisodesViaBrowser_FrontendPerSeasonMerge(t *testing.T) {
	t.Parallel()
	solver := &scriptedSolver{
		solveByURL: map[string]*CFSolveResult{
			"https://sf.test/serie/123": {
				HTML:     sfSeason1Anchors,
				FinalURL: "https://frontend.test/serie/demo-show/1",
			},
			"https://frontend.test/serie/demo-show/2": {HTML: sfSeason2Anchors},
		},
		solveErr: map[string]error{
			// Season 3 fails: the cascade must swallow it and keep what it has.
			"https://frontend.test/serie/demo-show/3": fmt.Errorf("season 3 boom"),
		},
	}
	c := NewSuperFlixClient()
	c.baseURL = "https://sf.test"
	c.browserSolver = solver

	got, err := c.GetEpisodes(context.Background(), "123")
	require.NoError(t, err)

	// Season 1 from the top solve, season 2 from the merged per-season fetch,
	// season 3 dropped (its solve errored).
	require.Contains(t, got, "1")
	require.Contains(t, got, "2")
	assert.NotContains(t, got, "3")
	assert.Len(t, got["1"], 2)
	assert.Len(t, got["2"], 1)

	solveCalls, _ := solver.calls()
	assert.Contains(t, solveCalls, "https://sf.test/serie/123")
	assert.Contains(t, solveCalls, "https://frontend.test/serie/demo-show/2")
	assert.Contains(t, solveCalls, "https://frontend.test/serie/demo-show/3")
	// Season 1 already present from the top page → must NOT be re-fetched.
	assert.NotContains(t, solveCalls, "https://frontend.test/serie/demo-show/1")
}

func TestGetEpisodesViaBrowser_WindowAllEpisodesSingleSolve(t *testing.T) {
	t.Parallel()
	// The real frontend fixture carries window.allEpisodes for every season, so a
	// single solve covers them all — no per-season fetch.
	solver := &scriptedSolver{
		solveByURL: map[string]*CFSolveResult{
			"https://sf.test/serie/777": {HTML: loadFrontendFixture(t)},
		},
	}
	c := NewSuperFlixClient()
	c.baseURL = "https://sf.test"
	c.browserSolver = solver

	got, err := c.GetEpisodes(context.Background(), "777")
	require.NoError(t, err)
	require.Contains(t, got, "1")
	require.Contains(t, got, "5")

	solveCalls, _ := solver.calls()
	assert.Equal(t, []string{"https://sf.test/serie/777"}, solveCalls,
		"window.allEpisodes path must solve exactly once (no per-season cascade)")
}

func TestGetEpisodesViaBrowser_TopLevelSolveError(t *testing.T) {
	t.Parallel()
	solver := &scriptedSolver{
		solveErr: map[string]error{"https://sf.test/serie/9": fmt.Errorf("gate down")},
	}
	c := NewSuperFlixClient()
	c.baseURL = "https://sf.test"
	c.browserSolver = solver

	_, err := c.GetEpisodes(context.Background(), "9")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to load serie page")
}

// A solved page with nothing to parse is a scrape failure, not a title with zero
// episodes. Returning (nil, nil) here — the old behavior — made the two
// indistinguishable and surfaced to the user as a bare "no seasons found"
// (issue #184), so it must report an error carrying ErrSuperFlixNoEpisodeList.
func TestGetEpisodesViaBrowser_NoEpisodesReturnsError(t *testing.T) {
	t.Parallel()
	solver := &scriptedSolver{
		solveByURL: map[string]*CFSolveResult{
			"https://sf.test/serie/0": {HTML: "<html><body>nothing here</body></html>"},
		},
	}
	c := NewSuperFlixClient()
	c.baseURL = "https://sf.test"
	c.browserSolver = solver

	got, err := c.GetEpisodes(context.Background(), "0")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSuperFlixNoEpisodeList)
	assert.Nil(t, got)
}

// --- getStreamViaBrowser cascade -------------------------------------------

func TestGetStreamViaBrowser_EmbedURLShape(t *testing.T) {
	// NOT parallel: the subtests swap the package-global defaultStreamCache, which
	// races with any other parallel test touching it.
	cases := []struct {
		name                     string
		mtype, mid, season, epis string
		wantURL                  string
	}{
		{"movie", "filme", "1048794", "", "", "https://" + SuperFlixEmbedHost + "/filme/1048794"},
		{"serie defaults s/e", "serie", "76479", "", "", "https://" + SuperFlixEmbedHost + "/serie/76479/1/1"},
		{"serie explicit s/e", "serie", "76479", "2", "5", "https://" + SuperFlixEmbedHost + "/serie/76479/2/5"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			// Not parallel: swaps the package-global stream cache.
			saved := defaultStreamCache
			t.Cleanup(func() { defaultStreamCache = saved })
			defaultStreamCache = &streamCache{path: filepath.Join(t.TempDir(), "c.json")}

			solver := &scriptedSolver{
				sniff: &CFStreamResult{
					StreamURL:  "https://cdn.test/master.m3u8",
					Referer:    "https://" + SuperFlixEmbedHost + "/",
					PlayerHost: "https://host.test",
					VideoHash:  "deadbeef",
				},
			}
			c := NewSuperFlixClient()
			c.browserSolver = solver

			res, err := c.GetStreamURL(context.Background(), tc.mtype, tc.mid, tc.season, tc.epis)
			require.NoError(t, err)
			assert.Equal(t, "https://cdn.test/master.m3u8", res.StreamURL)

			_, sniffCalls := solver.calls()
			require.Len(t, sniffCalls, 1)
			assert.Equal(t, tc.wantURL, sniffCalls[0])
		})
	}
}

func TestGetStreamViaBrowser_SniffErrorPropagates(t *testing.T) {
	// Not parallel: swaps the package-global stream cache.
	saved := defaultStreamCache
	t.Cleanup(func() { defaultStreamCache = saved })
	defaultStreamCache = &streamCache{path: filepath.Join(t.TempDir(), "c.json")}

	solver := &scriptedSolver{sniffErr: fmt.Errorf("turnstile timeout")}
	c := NewSuperFlixClient()
	c.browserSolver = solver

	_, err := c.GetStreamURL(context.Background(), "filme", "1", "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "embed stream sniff failed")
}

// TestGetStreamViaBrowser_StaleCacheReSolves is the core stale-cache cascade:
// a cached (host, hash) whose getVideo no longer resolves must fall back to a
// browser re-solve, and the freshly sniffed (host, hash) must replace the stale
// cache entry.
func TestGetStreamViaBrowser_StaleCacheReSolves(t *testing.T) {
	// Stale host: returns 500 so the pure-HTTP getVideo replay fails.
	stale := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(stale.Close)

	saved := defaultStreamCache
	t.Cleanup(func() { defaultStreamCache = saved })
	defaultStreamCache = &streamCache{path: filepath.Join(t.TempDir(), "c.json")}

	key := streamCacheKey("filme", "555", "", "")
	defaultStreamCache.put(key, streamCacheEntry{Host: stale.URL, Hash: "stalehash"})

	solver := &scriptedSolver{
		sniff: &CFStreamResult{
			StreamURL:  "https://cdn.test/fresh/master.m3u8",
			PlayerHost: "https://fresh.host",
			VideoHash:  "freshhash",
		},
	}
	c := NewSuperFlixClient()
	c.client = stale.Client() // reach the loopback stale host (bypass anti-SSRF)
	c.browserSolver = solver

	res, err := c.GetStreamURL(context.Background(), "filme", "555", "", "")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.test/fresh/master.m3u8", res.StreamURL)

	// Browser re-solve happened, and the cache now holds the fresh pair.
	_, sniffCalls := solver.calls()
	assert.Len(t, sniffCalls, 1, "stale cache must trigger exactly one re-solve")
	got, ok := defaultStreamCache.get(key)
	require.True(t, ok)
	assert.Equal(t, "https://fresh.host", got.Host)
	assert.Equal(t, "freshhash", got.Hash)
}

// TestGetStreamViaBrowser_DeadSniffedStream guards the fresh-path liveness probe:
// the browser can capture a signed URL for a player host that has already rotated
// out of the CDN. That link 404s a few seconds into playback, so the sniff must be
// rejected — with an error the caller can fail over on — and the dead (host, hash)
// must never enter the cache.
func TestGetStreamViaBrowser_DeadSniffedStream(t *testing.T) {
	// Rotated-out CDN: the freshly signed URL answers 404.
	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(dead.Close)

	saved := defaultStreamCache
	t.Cleanup(func() { defaultStreamCache = saved })
	defaultStreamCache = &streamCache{path: filepath.Join(t.TempDir(), "c.json")}

	solver := &scriptedSolver{
		sniff: &CFStreamResult{
			StreamURL:  dead.URL + "/signed/master.m3u8",
			PlayerHost: "https://rotated.host",
			VideoHash:  "deadhash",
		},
	}
	c := NewSuperFlixClient()
	c.client = dead.Client() // reach the loopback dead host (bypass anti-SSRF)
	c.browserSolver = solver

	_, err := c.GetStreamURL(context.Background(), "filme", "777", "", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "dead stream host")

	// The doomed (host, hash) must not have been cached.
	_, ok := defaultStreamCache.get(streamCacheKey("filme", "777", "", ""))
	assert.False(t, ok, "a dead sniff must not poison the cache")
}
