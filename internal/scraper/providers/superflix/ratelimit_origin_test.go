package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
)

// sfOrigin is a mock SuperFlix origin that reproduces the abuse guard measured
// live on 2026-09-20, so the regressions behind the "dexter" report can be
// driven offline and deterministically.
//
// What the real edge does, captured from an actual 429:
//
//	HTTP/2.0 429 Too Many Requests
//	X-Abuse-Guard: rate-limited
//	X-Ratelimit-Limit: 15
//	X-Ratelimit-Window: 10
//	Retry-After: 10
//
// Three properties of it matter here and are modelled faithfully:
//
//  1. The counter is PER ENDPOINT. `/` kept answering 200 while `/pesquisar`
//     was still 429ing from the same IP, neither response cached. So probing
//     the homepage (host discovery) does not spend the search allowance.
//  2. Traffic received WHILE blocked extends the block. This is what made the
//     failure self-sustaining: the transport answered each 429 by sleeping and
//     re-requesting the same path, which refreshed the block, which produced
//     the next 429.
//  3. Retry-After is advertised in whole seconds, and honoring it is the only
//     thing that clears the block.
type sfOrigin struct {
	*httptest.Server

	// limit/window define the allowance. retryAfter is what a 429 advertises.
	limit      int
	window     time.Duration
	retryAfter time.Duration
	// sticky models property 2: a request arriving while blocked restarts the
	// window instead of being ignored.
	sticky bool
	// body is what a 200 returns (SuperFlix search HTML).
	body string

	mu sync.Mutex
	// hits counts every request ever received, per path.
	hits map[string]int
	// stamps records arrival times per path, so a test can assert on traffic
	// that arrived after some moment.
	stamps map[string][]time.Time
	// windowStart/inWindow track the limiter for the rate-limited path.
	windowStart time.Time
	inWindow    int
	// onUnexpected, when set, is called for a request the test did not expect.
	onUnexpected func(path string, n int)

	// hold makes the handler block instead of answering, so a test can observe
	// what happens to an IN-FLIGHT request when its caller walks away.
	hold bool
	// clientGone is closed when a held request's connection is torn down by the
	// client — which only happens if the caller's context reached the request.
	clientGone chan struct{}
	goneOnce   sync.Once
	// released ends a held request so the server can shut down cleanly.
	released chan struct{}
}

// mismatchedAcceptLanguage is Chrome's q-ladder, which GoAnime used to send
// under its Firefox User-Agent — a pair no real browser produces.
//
// On 2026-09-21 the live host answered 429 to exactly this value across eight
// interleaved probes while serving every other Accept-Language 200; hours later
// the same value was accepted again, so the rule is reputation-sensitive rather
// than a fixed blocklist. The mock rejects it regardless: whether or not the
// host is enforcing it today, GoAnime must not present a browser it is not.

// sfSearchPath is the only endpoint the abuse guard is modelled on, matching
// the real deployment where `/` is not rate limited alongside it.
const sfSearchPath = "/pesquisar"

// emptyResultsHTML is a valid, parseable SuperFlix search page with no cards.
// Parsing is not what these tests are about; traffic shape is.
const emptyResultsHTML = `<!DOCTYPE html><html><head><title>SuperFlixAPI - Busca</title></head><body></body></html>`

func newSFOrigin(t *testing.T, opts ...func(*sfOrigin)) *sfOrigin {
	t.Helper()
	o := &sfOrigin{
		limit:      15,
		window:     10 * time.Second,
		retryAfter: time.Second,
		body:       emptyResultsHTML,
		hits:       map[string]int{},
		stamps:     map[string][]time.Time{},
		clientGone: make(chan struct{}),
		released:   make(chan struct{}),
	}
	for _, opt := range opts {
		opt(o)
	}
	o.Server = httptest.NewServer(http.HandlerFunc(o.serve))
	t.Cleanup(func() {
		close(o.released)
		o.Close()
	})
	return o
}

// withHold makes every request block until the client disconnects or the test
// ends. It models an origin that is simply slow — the case where the only thing
// that can end the request is the caller's own deadline.
func withHold() func(*sfOrigin) {
	return func(o *sfOrigin) { o.hold = true }
}

// withLimit sets a small allowance and window so a limiter test runs in
// milliseconds instead of the real 15-per-10s.
func withLimit(limit int, window time.Duration) func(*sfOrigin) {
	return func(o *sfOrigin) { o.limit, o.window = limit, window }
}

func withRetryAfter(d time.Duration) func(*sfOrigin) {
	return func(o *sfOrigin) { o.retryAfter = d }
}

// withSticky makes traffic received while blocked extend the block, which is
// the property that turned a single 429 into a lasting outage.
func withSticky() func(*sfOrigin) {
	return func(o *sfOrigin) { o.sticky = true }
}

func (o *sfOrigin) serve(w http.ResponseWriter, r *http.Request) {
	o.mu.Lock()
	path := r.URL.Path
	o.hits[path]++
	n := o.hits[path]
	o.stamps[path] = append(o.stamps[path], time.Now())
	unexpected := o.onUnexpected
	limited := path == sfSearchPath && o.countLocked()
	o.mu.Unlock()

	// Reject the Firefox-UA/Chrome-ladder pair, so every test that searches
	// through this mock also guards the header instead of it being one
	// assertion somebody can delete.
	if r.Header.Get("Accept-Language") == netx.ChromeAcceptLanguage {
		limited = true
	}

	if unexpected != nil {
		unexpected(path, n)
	}

	if o.hold {
		select {
		case <-r.Context().Done():
			// net/http cancels the handler's context when the client hangs up.
			o.goneOnce.Do(func() { close(o.clientGone) })
		case <-o.released:
		}
		return
	}

	if limited {
		w.Header().Set("X-Abuse-Guard", "rate-limited")
		w.Header().Set("X-Ratelimit-Limit", fmt.Sprint(o.limit))
		w.Header().Set("X-Ratelimit-Window", fmt.Sprint(int(o.window.Seconds())))
		w.Header().Set("Retry-After", fmt.Sprint(int(o.retryAfter.Seconds())))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`<html><body>rate limited</body></html>`))
		return
	}

	w.Header().Set("Content-Type", "text/html; charset=utf-8")
	_, _ = w.Write([]byte(o.body))
}

// countLocked records one request against the limiter and reports whether it
// must be rejected. Caller holds o.mu.
func (o *sfOrigin) countLocked() bool {
	now := time.Now()
	if now.Sub(o.windowStart) >= o.window {
		o.windowStart = now
		o.inWindow = 0
	}
	o.inWindow++
	if o.inWindow <= o.limit {
		return false
	}
	if o.sticky {
		// Property 2: traffic received while blocked restarts the window, so
		// the block only lapses once the client actually goes quiet.
		o.windowStart = now
	}
	return true
}

func (o *sfOrigin) hitCount(path string) int {
	o.mu.Lock()
	defer o.mu.Unlock()
	return o.hits[path]
}

// hitsAfter counts requests to path that arrived strictly after t. It is how
// the "an abandoned search stops talking to the origin" invariant is checked.
func (o *sfOrigin) hitsAfter(path string, t time.Time) int {
	o.mu.Lock()
	defer o.mu.Unlock()
	n := 0
	for _, s := range o.stamps[path] {
		if s.After(t) {
			n++
		}
	}
	return n
}

// failOnRequestAfter makes any further request to path fail the test as soon as
// it arrives, instead of waiting out a grace period and counting afterwards.
// This is what turns the leak regression into an immediate, unambiguous failure.
func (o *sfOrigin) failOnRequestAfter(t *testing.T, path, why string) {
	t.Helper()
	o.mu.Lock()
	seen := o.hits[path]
	o.onUnexpected = func(got string, n int) {
		if got == path && n > seen {
			t.Errorf("unexpected request #%d to %s: %s", n, path, why)
		}
	}
	o.mu.Unlock()
}

// awaitClientGone reports whether a held request was torn down within d.
func (o *sfOrigin) awaitClientGone(d time.Duration) bool {
	select {
	case <-o.clientGone:
		return true
	case <-time.After(d):
		return false
	}
}

// newOriginClient builds a SuperFlixClient pointed at the mock origin while
// keeping the REAL cfFallbackTransport in the chain.
//
// NewClientForTest deliberately swaps that transport out (it exists to exercise
// the plain-HTTP parsing paths), but the transport is exactly the code under
// test here — it owns the Retry-After loop. So the client is assembled by hand:
// mock base URL, no browser solver, real transport.
func newOriginClient(o *sfOrigin) *SuperFlixClient {
	return newOriginClientWithGate(o, newTestGate(""))
}

// newTestGate is a rate-limit gate with millisecond bounds, so the back-off it
// enforces is observable inside a test instead of the production tens of
// seconds. path may be empty to keep it in memory.
func newTestGate(path string) *rateLimitGate {
	g := newRateLimitGate(path)
	g.minWait = 20 * time.Millisecond
	g.maxWait = 100 * time.Millisecond
	return g
}

func newOriginClientWithGate(o *sfOrigin, gate *rateLimitGate) *SuperFlixClient {
	jar, _ := newCookieJar()
	c := NewSuperFlixClient()
	c.baseURL = o.URL // != SuperFlixBase, so base() returns it verbatim
	c.browserSolver = nil
	c.maxRetries = 0
	c.retryDelay = 0
	c.client = &http.Client{
		Timeout:   30 * time.Second,
		Transport: &cfFallbackTransport{base: http.DefaultTransport, jar: jar, gate: gate},
		Jar:       jar,
	}
	return c
}

// newPlayPathRequest builds a search request that is NOT marked best-effort, so
// the transport is allowed to wait out a Retry-After the way the play path does.
func newPlayPathRequest(c *SuperFlixClient, query string) (*http.Request, error) {
	req, err := http.NewRequestWithContext(context.Background(), http.MethodGet,
		c.base()+sfSearchPath+"?s="+url.QueryEscape(query), http.NoBody)
	if err != nil {
		return nil, err
	}
	c.decorateRequest(req)
	return req, nil
}
