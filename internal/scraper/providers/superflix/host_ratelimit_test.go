package superflix

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"path/filepath"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Every retired alias 301s into the SAME live origin, so racing the whole seed
// list lands one homepage GET per alias on that origin inside a second — a
// dozen requests, on the critical path of the first SuperFlix call, to learn
// something the previous run already wrote to disk. A run that remembers a
// working host must walk that host alone; the race stays for the cold case it
// was built for.
//
// Scope note: this is about request volume and latency, NOT about the 429s in
// the "dexter" report. SuperFlix's abuse guard counts per endpoint (measured
// 2026-09-20: `/` answered 200 while `/pesquisar` was still 429ing from the
// same IP, neither cached), so these homepage probes never competed with the
// search request. What starved `/pesquisar` was the transport re-requesting
// THAT path after the search had been abandoned — see retryafter_test.go.

// countingHosts wraps fakeHosts and counts every request per host.
type countingHosts struct {
	inner fakeHosts
	hits  map[string]*atomic.Int32
}

func newCountingHosts(inner fakeHosts) *countingHosts {
	c := &countingHosts{inner: inner, hits: map[string]*atomic.Int32{}}
	for h := range inner {
		c.hits[h] = &atomic.Int32{}
	}
	return c
}

func (c *countingHosts) RoundTrip(r *http.Request) (*http.Response, error) {
	if n, ok := c.hits[r.URL.Host]; ok {
		n.Add(1)
	}
	return c.inner.RoundTrip(r)
}

func (c *countingHosts) count(host string) int32 {
	if n, ok := c.hits[host]; ok {
		return n.Load()
	}
	return 0
}

// proberWithRemembered builds a prober whose state file already names `host`.
func proberWithRemembered(t *testing.T, rt http.RoundTripper, remembered string, seeds ...string) *hostProber {
	t.Helper()
	statePath := filepath.Join(t.TempDir(), hostStateFileName)
	blob, err := json.Marshal(map[string]string{"host": remembered})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(statePath, blob, 0o600))

	return &hostProber{
		client: &http.Client{
			Transport: rt,
			CheckRedirect: func(*http.Request, []*http.Request) error {
				return http.ErrUseLastResponse
			},
		},
		seeds:     seeds,
		statePath: statePath,
	}
}

func TestHostProber_WarmRunWalksOnlyTheRememberedHost(t *testing.T) {
	t.Parallel()
	hosts := newCountingHosts(fakeHosts{
		"superflixapi.monster": respond(200, nil, liveSuperFlixHome),
		"superflixapi.baby":    redirectTo("https://superflixapi.monster/"),
		"superflixapi.beer":    redirectTo("https://superflixapi.monster/"),
		"superflixapi.sbs":     redirectTo("https://superflixapi.monster/"),
		"superflixapi.pro":     redirectTo("https://superflixapi.monster/"),
	})
	p := proberWithRemembered(t, hosts, "superflixapi.monster",
		"superflixapi.monster", "superflixapi.baby", "superflixapi.beer",
		"superflixapi.sbs", "superflixapi.pro")

	host, err := p.resolve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "superflixapi.monster", host)

	assert.Equal(t, int32(1), hosts.count("superflixapi.monster"),
		"a remembered, live host must cost exactly one request against the rate limit")
	for _, alias := range []string{"superflixapi.baby", "superflixapi.beer", "superflixapi.sbs", "superflixapi.pro"} {
		assert.Zerof(t, hosts.count(alias), "%s must not be probed when the remembered host answers", alias)
	}
}

// If the remembered host has died, the full race still runs — the fast path may
// not cost the binary its ability to follow a rotation.
func TestHostProber_DeadRememberedHostFallsBackToTheRace(t *testing.T) {
	t.Parallel()
	hosts := newCountingHosts(fakeHosts{
		// superflixapi.baby (the remembered one) is deliberately absent: dead.
		"superflixapi.pro":     redirectTo("https://superflixapi.monster/"),
		"superflixapi.monster": respond(200, nil, liveSuperFlixHome),
	})
	// Only ONE seed, and it reaches the live host through a redirect. The race
	// cancels its losers as soon as one walker wins, so counting hits on a seed
	// that merely COULD have been walked is timing-dependent; with a single
	// path, "the race ran" is observable without a data race of its own.
	p := proberWithRemembered(t, hosts, "superflixapi.baby", "superflixapi.pro")

	host, err := p.resolve(context.Background())
	require.NoError(t, err, "a dead remembered host must not strand discovery")
	assert.Equal(t, "superflixapi.monster", host)
	assert.Positive(t, hosts.count("superflixapi.pro"), "the alias race must still have run")
}

// A remembered host that answers but is not SuperFlix (a parked page on a
// re-registered domain) must not short-circuit discovery either.
func TestHostProber_ParkedRememberedHostFallsBackToTheRace(t *testing.T) {
	t.Parallel()
	hosts := newCountingHosts(fakeHosts{
		"superflixapi.baby":    respond(200, nil, `<html><title>superflixapi.baby is for sale</title></html>`),
		"superflixapi.pro":     redirectTo("https://superflixapi.monster/"),
		"superflixapi.monster": respond(200, nil, liveSuperFlixHome),
	})
	p := proberWithRemembered(t, hosts, "superflixapi.baby",
		"superflixapi.pro", "superflixapi.monster")

	host, err := p.resolve(context.Background())
	require.NoError(t, err)
	assert.Equal(t, "superflixapi.monster", host,
		"a parked remembered host must not be adopted as live")
}
