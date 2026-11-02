package diag

import (
	"net"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func probe(t *testing.T, url string, timeout time.Duration) Result {
	t.Helper()
	var got []Result
	var mu sync.Mutex
	ProbeAll(t.Context(), http.DefaultClient, []Probe{{Name: "Site", URL: url}}, timeout, func(r Result) {
		mu.Lock()
		got = append(got, r)
		mu.Unlock()
	})
	require.Len(t, got, 1)
	return got[0]
}

func TestProbeClassifiesWhatAProviderAnswered(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name    string
		handler http.HandlerFunc
		outcome string
		status  int
	}{
		{"healthy", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }, "ok", 200},
		{"cloudflare challenge header", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Cf-Mitigated", "challenge")
			w.WriteHeader(403)
		}, "challenge", 403},
		{"cloudflare 403", func(w http.ResponseWriter, _ *http.Request) {
			w.Header().Set("Server", "cloudflare")
			w.WriteHeader(403)
		}, "challenge", 403},
		{"plain 403 is a block", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(403) }, "blocked", 403},
		{"rate limited", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(429) }, "blocked", 429},
		{"site down", func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(502) }, "down", 502},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			srv := httptest.NewServer(tc.handler)
			t.Cleanup(srv.Close)
			r := probe(t, srv.URL, 5*time.Second)
			assert.Equal(t, tc.outcome, r.Outcome)
			assert.Equal(t, tc.status, r.Status)
			assert.Empty(t, r.FinalURL, "no redirect happened")
		})
	}
}

// The site moved: the probe says where it ended, so a mirror change is visible.
func TestProbeReportsAMovedDomain(t *testing.T) {
	t.Parallel()
	moved := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) { w.WriteHeader(200) }))
	t.Cleanup(moved.Close)
	old := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		http.Redirect(w, r, moved.URL+"/home", http.StatusMovedPermanently)
	}))
	t.Cleanup(old.Close)

	r := probe(t, old.URL, 5*time.Second)

	assert.Equal(t, "ok", r.Outcome)
	assert.Equal(t, moved.URL+"/home", r.FinalURL)
}

func TestProbeNamesTheFailureNotJustThatThereWasOne(t *testing.T) {
	t.Parallel()
	t.Run("refused", func(t *testing.T) {
		t.Parallel()
		l, err := net.Listen("tcp", "127.0.0.1:0")
		require.NoError(t, err)
		addr := l.Addr().String()
		require.NoError(t, l.Close()) // nothing listens there now
		r := probe(t, "http://"+addr, 3*time.Second)
		assert.Equal(t, "refused", r.Outcome)
		assert.Zero(t, r.Status)
	})
	t.Run("slow", func(t *testing.T) {
		t.Parallel()
		release := make(chan struct{})
		srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) { <-release }))
		t.Cleanup(func() { close(release); srv.Close() })
		start := time.Now()
		r := probe(t, srv.URL, 300*time.Millisecond)
		assert.Equal(t, "slow", r.Outcome)
		assert.Less(t, time.Since(start), 3*time.Second)
	})
	t.Run("dns", func(t *testing.T) {
		t.Parallel()
		r := probe(t, "http://no-such-host.invalid/", 5*time.Second)
		assert.Equal(t, "dns", r.Outcome)
	})
}

func TestProbeAllReportsEveryProviderAndNeverMoreThanFourAtOnce(t *testing.T) {
	t.Parallel()
	var mu sync.Mutex
	inFlight, peak := 0, 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		mu.Lock()
		inFlight++
		peak = max(peak, inFlight)
		mu.Unlock()
		time.Sleep(60 * time.Millisecond)
		mu.Lock()
		inFlight--
		mu.Unlock()
	}))
	t.Cleanup(srv.Close)
	var probes []Probe
	for i := range 10 {
		probes = append(probes, Probe{Name: string(rune('A' + i)), URL: srv.URL})
	}

	seen := map[string]bool{}
	ProbeAll(t.Context(), http.DefaultClient, probes, 5*time.Second, func(r Result) { seen[r.Name] = true })

	assert.Len(t, seen, 10)
	assert.LessOrEqual(t, peak, 4)
}
