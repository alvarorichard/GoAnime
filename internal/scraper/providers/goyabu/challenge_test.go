package goyabu

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/net/publicsuffix"
)

// Goyabu started answering 403 with a Cloudflare managed challenge, and the
// search simply stopped working — "Goyabu blocked the request (captcha/
// challenge)" on every query.
//
// gateTransport hands that interstitial to the browser solver once and replays
// the request with the clearance. What has to hold:
//
//   - unchallenged traffic is untouched, and keeps surf's Chrome fingerprint;
//   - a challenge costs exactly ONE solve, however many requests hit it;
//   - cleared traffic moves to the plain transport, because cf_clearance is
//     bound to the solving browser's User-Agent and surf rewrites it (measured:
//     replaying on surf returns 403, on a plain client 200);
//   - a solve that fails, or a build with no solver at all, hands the caller
//     back the challenge response with its body intact rather than hanging or
//     losing it.

const challengeHTML = `<html><head><title>Just a moment...</title></head>` +
	`<body><script>window.__cf_chl_opt={};</script></body></html>`

// recordingTransport answers a scripted response and records what it received.
type recordingTransport struct {
	mu      sync.Mutex
	calls   atomic.Int32
	lastUA  string
	lastRaw []string
	respond func(*http.Request) *http.Response
}

func (r *recordingTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	r.calls.Add(1)
	r.mu.Lock()
	r.lastUA = req.Header.Get("User-Agent")
	r.lastRaw = nil
	for _, c := range req.Cookies() {
		r.lastRaw = append(r.lastRaw, c.Name)
	}
	r.mu.Unlock()
	return r.respond(req), nil
}

func (r *recordingTransport) ua() string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.lastUA
}

func (r *recordingTransport) cookies() []string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return append([]string(nil), r.lastRaw...)
}

func respondWith(status int, body string, header http.Header) func(*http.Request) *http.Response {
	return func(req *http.Request) *http.Response {
		h := http.Header{}
		for k, v := range header {
			h[k] = append([]string(nil), v...)
		}
		return &http.Response{
			StatusCode: status,
			Header:     h,
			Body:       io.NopCloser(strings.NewReader(body)),
			Request:    req,
		}
	}
}

// fakeSolver stands in for the browser.
type fakeSolver struct {
	solves  atomic.Int32
	ua      string
	err     error
	delay   time.Duration
	visible atomic.Bool
}

func (f *fakeSolver) SolveChallenge(_ context.Context, targetURL string, _ time.Duration, visible bool) (*netx.ChallengeSolveResult, error) {
	f.solves.Add(1)
	f.visible.Store(visible)
	if f.delay > 0 {
		time.Sleep(f.delay)
	}
	if f.err != nil {
		return nil, f.err
	}
	return &netx.ChallengeSolveResult{
		Cookies:   []*http.Cookie{{Name: "cf_clearance", Value: "token", Path: "/"}},
		UserAgent: f.ua,
		FinalURL:  targetURL,
	}, nil
}

// withSolver installs the process-wide browser solver for one test.
//
// The registry is global on purpose — one browser serves the whole process — so
// a test that swaps it CANNOT run in parallel: a sibling's cleanup would pull
// the solver out from under it mid-request. Every test using this helper is
// therefore serial, and only the pure detection table below is parallel.
func withSolver(t *testing.T, s netx.ChallengeSolver) {
	t.Helper()
	netx.RegisterChallengeSolver(s)
	t.Cleanup(func() { netx.RegisterChallengeSolver(nil) })
}

func newTestGate(surf, plain http.RoundTripper) (*gateTransport, http.CookieJar) {
	jar, _ := cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
	return newGateTransport(surf, plain, jar), jar
}

func mustGet(t *testing.T, tr http.RoundTripper, rawURL string) *http.Response {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, rawURL, http.NoBody)
	require.NoError(t, err)
	resp, err := tr.RoundTrip(req)
	require.NoError(t, err)
	return resp
}

func TestGateTransport_LeavesUnchallengedTrafficAlone(t *testing.T) {
	surf := &recordingTransport{respond: respondWith(http.StatusOK, "<html>goyabu</html>", nil)}
	plain := &recordingTransport{respond: respondWith(http.StatusOK, "unused", nil)}
	solver := &fakeSolver{ua: "SolverUA/1.0"}
	withSolver(t, solver)

	gate, _ := newTestGate(surf, plain)
	resp := mustGet(t, gate, "https://goyabu.io/")
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode)
	assert.Equal(t, int32(1), surf.calls.Load(), "the happy path keeps surf's Chrome fingerprint")
	assert.Zero(t, plain.calls.Load())
	assert.Zero(t, solver.solves.Load(), "no browser may open when nothing is challenging us")
}

func TestGateTransport_SolvesOnceThenReplaysWithClearance(t *testing.T) {
	surf := &recordingTransport{respond: respondWith(http.StatusForbidden, challengeHTML,
		http.Header{"Cf-Mitigated": {"challenge"}})}
	plain := &recordingTransport{respond: respondWith(http.StatusOK, "<html>goyabu</html>", nil)}
	solver := &fakeSolver{ua: "SolverUA/1.0"}
	withSolver(t, solver)

	gate, _ := newTestGate(surf, plain)
	resp := mustGet(t, gate, "https://goyabu.io/")
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "the replay must be what the caller gets")
	assert.Equal(t, int32(1), solver.solves.Load())
	assert.True(t, solver.visible.Load(),
		"a minimized window never cleared this gate in measurement; the solve must ask for a visible one")

	assert.Equal(t, "SolverUA/1.0", plain.ua(),
		"cf_clearance is bound to the solving browser's UA; sending any other one is challenged again")
	assert.Contains(t, plain.cookies(), "cf_clearance")
}

// Once cleared, later requests must skip the challenge dance entirely — and
// must not go back to surf, which would rewrite the UA the cookie depends on.
func TestGateTransport_LaterRequestsUseTheClearanceDirectly(t *testing.T) {
	surf := &recordingTransport{respond: respondWith(http.StatusForbidden, challengeHTML,
		http.Header{"Cf-Mitigated": {"challenge"}})}
	plain := &recordingTransport{respond: respondWith(http.StatusOK, "<html>goyabu</html>", nil)}
	solver := &fakeSolver{ua: "SolverUA/1.0"}
	withSolver(t, solver)

	gate, _ := newTestGate(surf, plain)
	for range 3 {
		resp := mustGet(t, gate, "https://goyabu.io/")
		_ = resp.Body.Close()
	}

	assert.Equal(t, int32(1), solver.solves.Load(), "one solve covers the session")
	assert.Equal(t, int32(1), surf.calls.Load(), "only the first request may take the uncleared path")
	assert.Equal(t, int32(3), plain.calls.Load())
}

// Several requests meeting the same gate must open one browser between them,
// not one each.
func TestGateTransport_ConcurrentRequestsShareASingleSolve(t *testing.T) {
	surf := &recordingTransport{respond: respondWith(http.StatusForbidden, challengeHTML,
		http.Header{"Cf-Mitigated": {"challenge"}})}
	plain := &recordingTransport{respond: respondWith(http.StatusOK, "<html>goyabu</html>", nil)}
	solver := &fakeSolver{ua: "SolverUA/1.0", delay: 80 * time.Millisecond}
	withSolver(t, solver)

	gate, _ := newTestGate(surf, plain)

	var wg sync.WaitGroup
	for range 6 {
		wg.Add(1)
		go func() {
			defer wg.Done()
			req, _ := http.NewRequest(http.MethodGet, "https://goyabu.io/", http.NoBody)
			if resp, err := gate.RoundTrip(req); err == nil {
				_ = resp.Body.Close()
			}
		}()
	}
	wg.Wait()

	assert.Equal(t, int32(1), solver.solves.Load(),
		"six challenged requests must not open six browser windows")
}

// A solve that fails must hand back the challenge response — with its body
// still readable, since the caller is about to parse or report it.
func TestGateTransport_FailedSolveReturnsTheChallengeIntact(t *testing.T) {
	surf := &recordingTransport{respond: respondWith(http.StatusForbidden, challengeHTML,
		http.Header{"Cf-Mitigated": {"challenge"}})}
	plain := &recordingTransport{respond: respondWith(http.StatusOK, "unused", nil)}
	withSolver(t, &fakeSolver{err: errors.New("no display")})

	gate, _ := newTestGate(surf, plain)
	resp := mustGet(t, gate, "https://goyabu.io/")
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Contains(t, string(body), "Just a moment",
		"the body was buffered to inspect it; the caller must still be able to read it")
	assert.Zero(t, plain.calls.Load(), "nothing was cleared, so nothing may be replayed")
}

// A build or test that never imported the package owning the browser has no
// solver. That is a plain "cannot help", not a hang and not a panic.
func TestGateTransport_NoSolverRegisteredIsNotAFailure(t *testing.T) {
	surf := &recordingTransport{respond: respondWith(http.StatusForbidden, challengeHTML,
		http.Header{"Cf-Mitigated": {"challenge"}})}
	plain := &recordingTransport{respond: respondWith(http.StatusOK, "unused", nil)}
	withSolver(t, nil)

	gate, _ := newTestGate(surf, plain)
	resp := mustGet(t, gate, "https://goyabu.io/")
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusForbidden, resp.StatusCode)
	assert.Zero(t, plain.calls.Load())
}

// Detection keys on what the live 403 actually carried, and must not fire on
// an ordinary page — a false positive opens a browser for nothing.
func TestChallengeDetection_MatchesTheLiveInterstitialOnly(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		status int
		header http.Header
		body   string
		want   bool
	}{
		"cf-mitigated header":      {http.StatusForbidden, http.Header{"Cf-Mitigated": {"challenge"}}, "anything", true},
		"just a moment":            {http.StatusForbidden, nil, challengeHTML, true},
		"challenge platform":       {http.StatusServiceUnavailable, nil, `<script src="/cdn-cgi/challenge-platform/h/b/x.js">`, true},
		"ordinary 200":             {http.StatusOK, nil, "<html>goyabu</html>", false},
		"real 403 without markers": {http.StatusForbidden, nil, "<html>forbidden</html>", false},
		"real 404":                 {http.StatusNotFound, nil, "<html>gone</html>", false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			resp := respondWith(tt.status, tt.body, tt.header)(&http.Request{URL: &url.URL{}})
			_, got := challengeBody(resp)
			assert.Equal(t, tt.want, got)
		})
	}
}

// deadlineSolver records the budget the solve was actually given.
type deadlineSolver struct {
	ua          string
	gotBudget   time.Duration
	hadDeadline bool
}

func (d *deadlineSolver) SolveChallenge(ctx context.Context, targetURL string, _ time.Duration, _ bool) (*netx.ChallengeSolveResult, error) {
	if dl, ok := ctx.Deadline(); ok {
		d.hadDeadline = true
		d.gotBudget = time.Until(dl)
	}
	return &netx.ChallengeSolveResult{
		Cookies:   []*http.Cookie{{Name: "cf_clearance", Value: "t", Path: "/"}},
		UserAgent: d.ua,
		FinalURL:  targetURL,
	}, nil
}

// A browser solve cannot live inside an HTTP request's deadline.
//
// The shared fast client caps requests at 8s. Running the solve on the
// request's context meant that cap killed the browser mid-challenge: every
// solve died at ~8s, and a search burned four doomed browser launches over 32
// seconds without ever obtaining clearance. A cold managed challenge takes
// about a minute.
//
// So the solve is detached and gets gateSolveTimeout of its own.
func TestGateTransport_SolveGetsItsOwnBudgetNotTheRequestsDeadline(t *testing.T) {
	surf := &recordingTransport{respond: respondWith(http.StatusForbidden, challengeHTML,
		http.Header{"Cf-Mitigated": {"challenge"}})}
	plain := &recordingTransport{respond: respondWith(http.StatusOK, "<html>goyabu</html>", nil)}
	solver := &deadlineSolver{ua: "SolverUA/1.0"}
	withSolver(t, solver)

	gate, _ := newTestGate(surf, plain)

	// A request budget far too short for a real challenge, like the 8s cap that
	// caused the failure.
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, "https://goyabu.io/", http.NoBody)
	require.NoError(t, err)

	resp, err := gate.RoundTrip(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	require.True(t, solver.hadDeadline, "the solve must still be bounded; an unbounded one can hang forever")
	assert.Greater(t, solver.gotBudget, time.Minute,
		"the solve inherited the request's deadline and would be killed before a cold challenge can clear")
	assert.LessOrEqual(t, solver.gotBudget, gateSolveTimeout,
		"and it must not exceed the budget we advertise")
}

// The per-attempt bound has to leave room for a real response while still
// failing a hung host quickly — it replaced the client's global timeout, which
// could not serve both purposes at once.
func TestHTTPAttemptTimeout_IsBoundedOnBothSides(t *testing.T) {
	t.Parallel()
	assert.Greater(t, httpAttemptTimeout, 5*time.Second, "too tight and a slow-but-healthy page fails")
	assert.Less(t, httpAttemptTimeout, gateSolveTimeout, "an HTTP attempt must never outlast a whole solve")
}

// A managed challenge is not always winnable — measured on the same host
// minutes apart, it cleared in 6s once and not at all within 90s later. The
// client retries and several code paths each meet the gate, so without a
// cooldown one stubborn spell became four consecutive 90s solves: a search that
// normally takes two seconds took six minutes.
func TestGateTransport_AFailedSolveIsNotRetriedForEveryRequest(t *testing.T) {
	surf := &recordingTransport{respond: respondWith(http.StatusForbidden, challengeHTML,
		http.Header{"Cf-Mitigated": {"challenge"}})}
	plain := &recordingTransport{respond: respondWith(http.StatusOK, "unused", nil)}
	solver := &fakeSolver{err: errors.New("gate never cleared")}
	withSolver(t, solver)

	gate, _ := newTestGate(surf, plain)
	for range 4 {
		resp := mustGet(t, gate, "https://goyabu.io/")
		_ = resp.Body.Close()
	}

	assert.Equal(t, int32(1), solver.solves.Load(),
		"four challenged requests must cost ONE doomed solve, not four")
}

// The cooldown must not outlive its usefulness: once it lapses, a source that
// recovered has to be reachable again without restarting GoAnime.
func TestGateTransport_CooldownLapsesAndLetsTheSolveRetry(t *testing.T) {
	surf := &recordingTransport{respond: respondWith(http.StatusForbidden, challengeHTML,
		http.Header{"Cf-Mitigated": {"challenge"}})}
	plain := &recordingTransport{respond: respondWith(http.StatusOK, "<html>goyabu</html>", nil)}
	solver := &fakeSolver{err: errors.New("gate never cleared")}
	withSolver(t, solver)

	gate, _ := newTestGate(surf, plain)
	resp := mustGet(t, gate, "https://goyabu.io/")
	_ = resp.Body.Close()
	require.Equal(t, int32(1), solver.solves.Load())

	// Pretend the cooldown has passed, and let the gate clear this time.
	gate.mu.Lock()
	gate.failedAt = time.Now().Add(-2 * solveRetryCooldown)
	gate.mu.Unlock()
	solver.err = nil
	solver.ua = "SolverUA/1.0"

	resp = mustGet(t, gate, "https://goyabu.io/")
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, int32(2), solver.solves.Load(), "the source must get another chance")
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

// The body of a response must stay readable after the transport returns it.
// Bounding each attempt with its own deadline is right; tearing that deadline
// down when the headers arrive is not — the caller is still reading, and
// goquery failed with "context canceled" on a response that had arrived fine.
func TestGateTransport_ResponseBodySurvivesTheAttemptDeadline(t *testing.T) {
	surf := &recordingTransport{respond: respondWith(http.StatusOK, "<html>goyabu</html>", nil)}
	plain := &recordingTransport{respond: respondWith(http.StatusOK, "unused", nil)}
	withSolver(t, nil)

	gate, _ := newTestGate(surf, plain)
	resp := mustGet(t, gate, "https://goyabu.io/")
	defer func() { _ = resp.Body.Close() }()

	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "the per-attempt context must not be cancelled before the body is read")
	assert.Contains(t, string(body), "goyabu")
}

// A clearance that lives only in memory is paid for again on every launch, and
// a ~6s solve does not fit inside the search fan-out's straggler grace — so
// Goyabu was searched, lost the race every time, and looked like it was not
// being searched at all. Persisting it turns the second run onward into a
// plain HTTP request.
func TestGateTransport_ClearanceIsRestoredFromDisk(t *testing.T) {
	path := filepath.Join(t.TempDir(), clearanceStateFile)
	target := &url.URL{Scheme: "https", Host: "goyabu.io", Path: "/"}

	writer, _ := newTestGate(nil, nil)
	writer.saveClearanceTo(path, target, "SolverUA/1.0",
		[]*http.Cookie{{Name: "cf_clearance", Value: "token", Path: "/"}})

	// A new process: same file, nothing shared in memory.
	surf := &recordingTransport{respond: respondWith(http.StatusForbidden, challengeHTML,
		http.Header{"Cf-Mitigated": {"challenge"}})}
	plain := &recordingTransport{respond: respondWith(http.StatusOK, "<html>goyabu</html>", nil)}
	solver := &fakeSolver{ua: "SolverUA/1.0"}
	withSolver(t, solver)

	restarted, _ := newTestGate(surf, plain)
	restarted.loadClearanceFrom(path)

	resp := mustGet(t, restarted, "https://goyabu.io/")
	defer func() { _ = resp.Body.Close() }()

	assert.Zero(t, solver.solves.Load(), "a restart must not pay for a solve it already did")
	assert.Zero(t, surf.calls.Load(), "and must not take the uncleared path either")
	assert.Equal(t, "SolverUA/1.0", plain.ua())
}

// Stale state must not be replayed into a challenge forever.
func TestGateTransport_StaleClearanceIsIgnored(t *testing.T) {
	path := filepath.Join(t.TempDir(), clearanceStateFile)
	blob, err := json.Marshal(storedClearance{
		UserAgent: "SolverUA/1.0",
		Host:      "goyabu.io",
		Cookies:   []storedCookie{{Name: "cf_clearance", Value: "old", Path: "/"}},
		SavedAt:   time.Now().Add(-2 * clearanceMaxAge),
	})
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(path, blob, 0o600))

	gate, _ := newTestGate(nil, nil)
	gate.loadClearanceFrom(path)

	assert.Empty(t, gate.clearance(), "state older than the cap cannot be trusted")
}

func TestGateTransport_CorruptClearanceIsIgnored(t *testing.T) {
	path := filepath.Join(t.TempDir(), clearanceStateFile)
	require.NoError(t, os.WriteFile(path, []byte("{not json"), 0o600))

	gate, _ := newTestGate(nil, nil)
	gate.loadClearanceFrom(path)

	assert.Empty(t, gate.clearance(), "unreadable state must fail open, not wedge the source")
}

// A test binary must never write a clearance into the developer's cache.
func TestClearanceStatePath_TestBinaryKeepsItInMemory(t *testing.T) {
	t.Parallel()
	assert.Empty(t, clearanceStatePath())
}

// A stored clearance can be dead on arrival: cf_clearance expires, and the one
// restored from disk may already have lapsed. Being challenged on a request we
// believed was cleared is the only signal we get.
//
// Without acting on it, ensureCleared saw a non-empty User-Agent, concluded
// there was nothing to do, and never solved again — a restored-but-expired
// clearance left Goyabu answering 403 for the whole life of the process, with
// no solve attempted and nothing in the log to say why.
func TestGateTransport_StaleClearanceIsDiscardedAndResolved(t *testing.T) {
	var surfCalls atomic.Int32
	// The cleared path is challenged (the stored clearance is dead); the solve
	// produces a new UA, and the replay with it succeeds.
	plain := &recordingTransport{respond: func(req *http.Request) *http.Response {
		if req.Header.Get("User-Agent") == "FreshUA/2.0" {
			return respondWith(http.StatusOK, "<html>goyabu</html>", nil)(req)
		}
		return respondWith(http.StatusForbidden, challengeHTML,
			http.Header{"Cf-Mitigated": {"challenge"}})(req)
	}}
	surf := &recordingTransport{respond: func(req *http.Request) *http.Response {
		surfCalls.Add(1)
		return respondWith(http.StatusForbidden, challengeHTML,
			http.Header{"Cf-Mitigated": {"challenge"}})(req)
	}}
	solver := &fakeSolver{ua: "FreshUA/2.0"}
	withSolver(t, solver)

	gate, _ := newTestGate(surf, plain)
	// Simulate a clearance restored from a previous run that has since expired.
	gate.mu.Lock()
	gate.ua = "StaleUA/1.0"
	gate.mu.Unlock()

	resp := mustGet(t, gate, "https://goyabu.io/")
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode, "the source must recover, not stay broken")
	assert.Equal(t, int32(1), solver.solves.Load(),
		"a refused clearance has to trigger a solve; believing it is the bug")
	assert.Equal(t, "FreshUA/2.0", gate.clearance(), "the fresh clearance replaces the dead one")
}

// Only the clearance that was actually refused may be dropped, or a concurrent
// solve's fresh result would be thrown away by a late straggler.
func TestGateTransport_InvalidateOnlyDiscardsTheRefusedClearance(t *testing.T) {
	gate, _ := newTestGate(nil, nil)
	gate.mu.Lock()
	gate.ua = "FreshUA/2.0"
	gate.mu.Unlock()

	gate.invalidateClearance("StaleUA/1.0")

	assert.Equal(t, "FreshUA/2.0", gate.clearance(),
		"a straggler reporting an old failure must not discard a newer clearance")
}
