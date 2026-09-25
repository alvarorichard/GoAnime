package superflix

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
)

// Respecting a rate limit BETWEEN requests, not just inside one.
//
// SuperFlix's edge answers 429 with a Retry-After (10s, measured 2026-09-20)
// and its abuse guard is sticky: traffic that arrives while the block is up
// refreshes it. So the only thing that ends a block is going quiet for the
// window the server asked for.
//
// GoAnime used to throw that information away the moment a request finished.
// The transport honored Retry-After inside one request, but nothing remembered
// it afterwards, so the next search — typed five seconds later, or started by a
// brand new process — sent another request straight into the block and renewed
// it. Three searches in nineteen seconds was enough to keep an IP locked out
// indefinitely, which is exactly what the user hit while retyping "tehran".
//
// The circuit breaker in netx does not cover this. It needs three failures
// before it opens, it opens for a flat ten minutes, it ignores the delay the
// server actually asked for, and — decisive here — it lives in memory, so it is
// empty again on the next `goanime` launch.
//
// rateLimitGate closes those gaps: one 429 arms it, the wait is the server's
// own Retry-After, consecutive 429s back off further, and the state is on disk
// so a restart cannot walk straight back into the block.

// ErrRateLimited is returned INSTEAD OF sending a request, while a host's
// advertised back-off is still running. It is not a failure of the request —
// there was no request — so callers should report it as "the source asked us to
// wait", never retry it in a loop.
var ErrRateLimited = errors.New("superflix: rate limited, backing off")

const (
	// backoffStateFileName holds the gate across process restarts, next to the
	// host state and the stream cache.
	backoffStateFileName = "superflix-backoff.json"
	// minBackoff is the floor for every block, not just one with no usable
	// Retry-After.
	//
	// It is a floor because this host's advertised Retry-After is not true. It
	// sends "Retry-After: 10" and then stays blocked for minutes: measured
	// 2026-09-25, /pesquisar was still refusing after three and six minutes of
	// complete silence, and by then had escalated from answering 429 to
	// dropping the connection. Obeying the advertised 10 seconds meant walking
	// back into an active block on every retry, and the guard is sticky — a
	// probe sent during a block renews it. So the floor is set from what the
	// host does rather than from what it says.
	minBackoff = 60 * time.Second
	// maxBackoff caps the escalation. A blocked host is probed at worst every
	// few minutes, which is a trickle the abuse guard cannot mistake for abuse,
	// while still recovering on its own without the user restarting anything.
	maxBackoff = 5 * time.Minute
	// staleBackoff discards state left by a long-finished run, so a crash
	// during a block cannot gate a host forever.
	staleBackoff = 30 * time.Minute
)

// backoffKey identifies what a back-off applies to: one endpoint on one host.
//
// Keying by host alone over-blocks. SuperFlix's abuse guard counts per endpoint
// — measured 2026-09-21, `/` answered 200 while `/pesquisar` was still 429ing
// from the same IP, neither cached — so a blocked search would otherwise
// silence the episode listing, the metadata lookup and playback, all of which
// were still working. The query string is dropped: the limit is on the
// endpoint, not on what is being searched for.
func backoffKey(host, path string) string {
	if host == "" {
		return ""
	}
	if path == "" {
		path = "/"
	}
	return host + path
}

// hostBackoff is the gate for one endpoint.
type hostBackoff struct {
	// Until is when requests may resume.
	Until time.Time `json:"until"`
	// Strikes counts consecutive 429s, and drives the escalation. Reset by any
	// successful response.
	Strikes int `json:"strikes"`
}

// rateLimitGate remembers, per host, when we are allowed to talk again.
//
// Reads and writes go through a mutex and are mirrored to disk on every change.
// Persistence is best-effort: losing it costs one extra request, never
// correctness.
type rateLimitGate struct {
	mu sync.Mutex
	// hosts is the in-memory view (keyed by backoffKey), lazily loaded once.
	hosts  map[string]hostBackoff
	loaded bool
	// path is the state file; empty disables persistence (tests, or a platform
	// with no cache dir).
	path string
	// now is swappable so tests can drive the clock.
	now func() time.Time
	// minWait/maxWait bound the back-off. Fields rather than constants so a
	// test can exercise the escalation in milliseconds instead of minutes.
	minWait, maxWait time.Duration
}

// newRateLimitGate builds a gate with the production bounds. path may be empty
// to keep the gate in memory only.
func newRateLimitGate(path string) *rateLimitGate {
	return &rateLimitGate{
		path:    path,
		now:     time.Now,
		minWait: minBackoff,
		maxWait: maxBackoff,
	}
}

// defaultRateLimitGate is the process-wide gate, shared by every SuperFlix
// client so two clients cannot each spend a request on the same blocked host.
var defaultRateLimitGate = newRateLimitGate(defaultBackoffStatePath())

// defaultBackoffStatePath is the on-disk gate, or "" when there is nowhere to
// put it.
//
// Test binaries get "" as well: the gate is process-wide, and a test that
// happens to provoke a 429 must not write a back-off into the developer's real
// cache — where it would then suppress their next actual search. Tests that
// need persistence build their own gate with an explicit temp path.
func defaultBackoffStatePath() string {
	if testing.Testing() {
		return ""
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "goanime", backoffStateFileName)
}

// knows reports whether this endpoint has already rate limited us at least
// once, whether or not the block is currently active.
//
// It is what lets a refused connection be read as the same block escalating
// rather than as an unrelated network failure.
func (g *rateLimitGate) knows(host, path string) bool {
	key := backoffKey(host, path)
	if g == nil || key == "" {
		return false
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.loadLocked()
	b, ok := g.hosts[key]
	return ok && b.Strikes > 0
}

// retryIn reports how long an endpoint must stay untouched, or 0 when it is free.
func (g *rateLimitGate) retryIn(host, path string) time.Duration {
	key := backoffKey(host, path)
	if g == nil || key == "" {
		return 0
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.loadLocked()

	b, ok := g.hosts[key]
	if !ok {
		return 0
	}
	d := b.Until.Sub(g.clock())
	if d <= 0 {
		return 0
	}
	return d
}

// arm records a 429 for host and returns how long we will now stay quiet.
//
// retryAfter is what the server advertised; it is the floor, not the whole
// story. Consecutive strikes double it, because a host that keeps refusing is
// telling us its advertised delay was not enough — and every probe we send
// during a sticky block is what keeps that block alive.
func (g *rateLimitGate) arm(host, path string, retryAfter time.Duration) time.Duration {
	key := backoffKey(host, path)
	if g == nil || key == "" {
		return 0
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.loadLocked()

	wait := max(retryAfter, g.minBound())

	b := g.hosts[key]
	b.Strikes++
	// First strike waits exactly what was asked; each further one doubles.
	for range b.Strikes - 1 {
		if wait >= g.maxBound() {
			break
		}
		wait *= 2
	}
	wait = min(wait, g.maxBound())

	b.Until = g.clock().Add(wait)
	g.hosts[key] = b
	g.persistLocked()
	return wait
}

// clear forgets a host's back-off. Called on any response that is not a 429:
// the host is talking to us again, so the strike count must not carry over into
// the next unrelated hiccup.
func (g *rateLimitGate) clear(host, path string) {
	key := backoffKey(host, path)
	if g == nil || key == "" {
		return
	}
	g.mu.Lock()
	defer g.mu.Unlock()
	g.loadLocked()

	if _, ok := g.hosts[key]; !ok {
		return
	}
	delete(g.hosts, key)
	g.persistLocked()
}

// minBound/maxBound fall back to the production bounds so a zero-value gate is
// still safe to use.
func (g *rateLimitGate) minBound() time.Duration {
	if g.minWait > 0 {
		return g.minWait
	}
	return minBackoff
}

func (g *rateLimitGate) maxBound() time.Duration {
	if g.maxWait > 0 {
		return g.maxWait
	}
	return maxBackoff
}

func (g *rateLimitGate) clock() time.Time {
	if g.now != nil {
		return g.now()
	}
	return time.Now()
}

// loadLocked reads the state file once per process. Caller holds g.mu.
func (g *rateLimitGate) loadLocked() {
	if g.loaded {
		return
	}
	g.loaded = true
	g.hosts = map[string]hostBackoff{}
	if g.path == "" {
		return
	}
	raw, err := os.ReadFile(g.path)
	if err != nil {
		return
	}
	var stored map[string]hostBackoff
	if json.Unmarshal(raw, &stored) != nil {
		return
	}
	now := g.clock()
	for host, b := range stored {
		// Drop entries that already lapsed, and ones so old they most likely
		// outlived the run that wrote them.
		if b.Until.Before(now) || b.Until.Sub(now) > staleBackoff {
			continue
		}
		g.hosts[host] = b
	}
}

// persistLocked writes the gate to disk. Caller holds g.mu. Best-effort.
func (g *rateLimitGate) persistLocked() {
	if g.path == "" {
		return
	}
	blob, err := json.Marshal(g.hosts)
	if err != nil {
		return
	}
	// 0o700/0o600 to match the other owner-only state in this cache tree.
	if mkErr := os.MkdirAll(filepath.Dir(g.path), 0o700); mkErr != nil {
		return
	}
	_ = os.WriteFile(g.path, blob, 0o600)
}

// rateLimitedError reports a request we deliberately did not send, and when the
// host becomes usable again. Wrapping ErrRateLimited lets callers tell it apart
// from a 429 we actually received.
func rateLimitedError(endpoint string, in time.Duration) error {
	return fmt.Errorf("%w: %s asked us to wait, retrying in %s (no request sent)",
		ErrRateLimited, endpoint, in.Round(time.Second))
}

// noteRateLimit arms the gate from a 429 response and logs what it cost.
func noteRateLimit(g *rateLimitGate, host, path string, retryAfter time.Duration) {
	wait := g.arm(host, path, retryAfter)
	util.Debug("SuperFlix rate limited; staying off this endpoint until the block can lapse",
		"endpoint", backoffKey(host, path), "retryAfter", retryAfter, "quietFor", wait)
}
