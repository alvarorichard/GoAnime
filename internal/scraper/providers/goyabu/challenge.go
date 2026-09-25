package goyabu

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
)

// Goyabu now sits behind a Cloudflare managed challenge.
//
// Measured 2026-09-21 from a challenged IP: goyabu.io answers 403 with
// cf-mitigated: challenge, __cf_chl markers and "Just a moment…". No header
// tuning gets past it — it wants a real browser.
//
// GoAnime already drives one for SuperFlix, offered through
// netx.ChallengeSolver. This transport is the piece that uses it: it notices a
// challenge, has the browser clear it once, and replays the request with the
// resulting clearance.
//
// Two measurements shaped the design. The first, taken with the profile wiped
// between runs — the COLD case only:
//
//	window minimized (--sf-offscreen)   never cleared in 70s
//	window visible                      cleared every time (~1min cold, ~6s warm)
//	headless                            never cleared
//
// Re-measured 2026-09-24 without wiping, which is how a real second run looks:
//
//	cold profile, hidden    never cleared in 90s
//	warm profile, hidden    cleared in 1.8s
//	warm profile, visible   cleared in 1.6s
//
// Reading only the first table is what made this ask for a forced window and
// charge every search the cold case's price. The window is now earned, not
// assumed: see netx.RevealWhenStuck.
//
//	clearance replayed on a plain client   200, real page
//	clearance replayed on the surf client  403, still challenged
//
// The second is the subtle one. cf_clearance is bound to the User-Agent that
// solved the challenge, and the shared surf client impersonates Chrome by
// REWRITING the User-Agent — so it sends a UA the cookie was not issued for and
// is challenged again. Cleared traffic therefore moves to a plain client whose
// headers we control; unchallenged traffic keeps surf's fingerprint, which is
// what gets it through in the first place.

const (
	// gateSolveTimeout bounds one browser solve. A cold managed challenge took
	// ~1 minute in measurement, so this leaves room without letting a hopeless
	// solve hold the search forever.
	gateSolveTimeout = 90 * time.Second

	// httpAttemptTimeout bounds one HTTP attempt, and is what keeps ordinary
	// failures fast now that the client carries no global timeout of its own.
	//
	// It has to be separate from the solve budget. The shared fast client caps
	// everything at 8s, and running the solve inside the request's lifetime
	// meant that cap killed the browser mid-challenge: every solve died at ~8s
	// and every search failed after four doomed attempts (measured: 32s per
	// search, eight browser launches, no clearance). A minute-long solve simply
	// cannot live inside an eight-second request.
	httpAttemptTimeout = 20 * time.Second

	// solveRetryCooldown is how long a FAILED solve suppresses further attempts.
	//
	// A managed challenge is not always winnable: measured on the same host
	// minutes apart, it cleared in 6s once and refused to clear at all within
	// 90s later. Without this, one stubborn spell turned a search into four
	// consecutive 90s solves — the client retries, and several code paths each
	// meet the gate — and a two-second search took six minutes. One attempt,
	// then fail fast until the mood upstream has had a chance to change.
	solveRetryCooldown = 5 * time.Minute

	// clearanceStateFile remembers a solved challenge across runs.
	//
	// Without it every `goanime` pays for a solve, because the clearance lived
	// only in memory. That cost more than the seconds it took: the search
	// fan-out returns once the first source answers plus a short straggler
	// grace, and a ~6s solve never fit inside it — so Goyabu was searched,
	// timed out of the grace, and the user saw only the fast sources and
	// concluded it was not being searched at all.
	//
	// Cloudflare issues cf_clearance with its own lifetime; this file just
	// stops us throwing it away at exit.
	clearanceStateFile = "goyabu-clearance.json"
)

// gateTransport clears a Cloudflare challenge once and replays with clearance.
type gateTransport struct {
	// surf carries the default, unchallenged traffic: a Chrome TLS fingerprint,
	// at the cost of rewriting the User-Agent.
	surf http.RoundTripper
	// plain preserves the headers we set, which is what a UA-bound clearance
	// cookie requires.
	plain http.RoundTripper
	// jar holds the clearance cookies.
	jar http.CookieJar

	mu sync.Mutex
	// ua is the browser's User-Agent once a solve has succeeded; empty before.
	// Its presence is what marks us as cleared.
	ua string
	// solving serialises solves, so several concurrent requests meeting the
	// same gate open one browser between them rather than one each.
	solving bool
	// failedAt is when the last solve gave up, and starts the cooldown that
	// keeps a hopeless gate from being retried into a six-minute search.
	failedAt time.Time
	// cond wakes the requests that waited for an in-flight solve.
	cond *sync.Cond
}

func newGateTransport(surf, plain http.RoundTripper, jar http.CookieJar) *gateTransport {
	t := &gateTransport{surf: surf, plain: plain, jar: jar}
	t.cond = sync.NewCond(&t.mu)
	t.loadClearance()
	return t
}

// storedClearance is the on-disk form of a solved challenge.
type storedClearance struct {
	UserAgent string         `json:"user_agent"`
	Host      string         `json:"host"`
	Cookies   []storedCookie `json:"cookies"`
	SavedAt   time.Time      `json:"saved_at"`
}

type storedCookie struct {
	Name   string    `json:"name"`
	Value  string    `json:"value"`
	Path   string    `json:"path"`
	Domain string    `json:"domain"`
	Expiry time.Time `json:"expires,omitzero"`
	// Secure and HttpOnly are carried through rather than dropped. They are how
	// the challenge issued cf_clearance, and rebuilding the cookie without them
	// would quietly widen where it can be sent.
	Secure   bool `json:"secure"`
	HTTPOnly bool `json:"http_only"`
}

// clearanceStatePath is where the clearance is kept, or "" when there is
// nowhere to put it. Test binaries get "" so a test run cannot leave a
// clearance in the developer's cache.
func clearanceStatePath() string {
	if testing.Testing() {
		return ""
	}
	dir, err := os.UserCacheDir()
	if err != nil {
		return ""
	}
	return filepath.Join(dir, "goanime", clearanceStateFile)
}

// clearanceMaxAge discards state we kept too long. cf_clearance has its own
// expiry, which we do not get to see reliably; this is the outer bound that
// stops a stale cookie being replayed into a challenge forever.
const clearanceMaxAge = 12 * time.Hour

// loadClearance restores a previous run's clearance, if there is one.
func (t *gateTransport) loadClearance() { t.loadClearanceFrom(clearanceStatePath()) }

// loadClearanceFrom is loadClearance against an explicit file, so a test can
// exercise the restore without touching the real cache.
func (t *gateTransport) loadClearanceFrom(path string) {
	if path == "" {
		return
	}
	raw, err := os.ReadFile(path) // #nosec G304 -- our own file in the user cache dir
	if err != nil {
		return
	}
	var st storedClearance
	if json.Unmarshal(raw, &st) != nil || st.UserAgent == "" || st.Host == "" {
		return
	}
	if time.Since(st.SavedAt) > clearanceMaxAge {
		return
	}

	u := &url.URL{Scheme: "https", Host: st.Host, Path: "/"}
	cookies := make([]*http.Cookie, 0, len(st.Cookies))
	for _, c := range st.Cookies {
		// #nosec G124 -- this cookie is rebuilt to be SENT on an outgoing
		// request, not issued in a Set-Cookie. Secure/HttpOnly only constrain a
		// server handing a cookie to a browser; on the wire to goyabu.io they
		// are inert. The values the challenge issued are carried through as
		// stored rather than invented, which is the honest thing to do here —
		// gosec cannot see that because they come from variables.
		cookies = append(cookies, &http.Cookie{
			Name: c.Name, Value: c.Value, Path: c.Path, Domain: c.Domain, Expires: c.Expiry,
			Secure: c.Secure, HttpOnly: c.HTTPOnly, SameSite: http.SameSiteLaxMode,
		})
	}
	if len(cookies) == 0 {
		return
	}
	t.jar.SetCookies(u, cookies)
	t.mu.Lock()
	t.ua = st.UserAgent
	t.mu.Unlock()
	util.Debug("Goyabu: reusing the clearance from a previous run", "cookies", len(cookies))
}

// saveClearance persists a solve so the next run does not repeat it.
// Best-effort: losing it costs one more solve, never correctness.
func (t *gateTransport) saveClearance(target *url.URL, ua string, cookies []*http.Cookie) {
	t.saveClearanceTo(clearanceStatePath(), target, ua, cookies)
}

// saveClearanceTo is saveClearance against an explicit file.
func (t *gateTransport) saveClearanceTo(path string, target *url.URL, ua string, cookies []*http.Cookie) {
	if path == "" || ua == "" || len(cookies) == 0 {
		return
	}
	st := storedClearance{UserAgent: ua, Host: target.Host, SavedAt: time.Now()}
	for _, c := range cookies {
		st.Cookies = append(st.Cookies, storedCookie{
			Name: c.Name, Value: c.Value, Path: c.Path, Domain: c.Domain, Expiry: c.Expires,
			Secure: c.Secure, HTTPOnly: c.HttpOnly,
		})
	}
	blob, err := json.Marshal(st)
	if err != nil {
		return
	}
	if mkErr := os.MkdirAll(filepath.Dir(path), 0o700); mkErr != nil {
		return
	}
	_ = os.WriteFile(path, blob, 0o600)
}

// invalidateClearance forgets a clearance that turned out not to work.
//
// staleUA guards against discarding a clearance a concurrent solve has just
// replaced: only the exact one that was refused is dropped. The on-disk copy
// goes too, so a restart does not begin by replaying the same dead cookie.
func (t *gateTransport) invalidateClearance(staleUA string) {
	t.mu.Lock()
	if t.ua != staleUA {
		t.mu.Unlock()
		return
	}
	t.ua = ""
	t.mu.Unlock()

	util.Debug("Goyabu: stored clearance was refused; discarding it and solving again")
	if path := clearanceStatePath(); path != "" {
		_ = os.Remove(path)
	}
}

// clearance returns the solved User-Agent, or "" while we are not cleared.
func (t *gateTransport) clearance() string {
	t.mu.Lock()
	defer t.mu.Unlock()
	return t.ua
}

func (t *gateTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	used := t.clearance()
	resp, err := t.send(req, used)
	if err != nil {
		return resp, err
	}

	body, isChallenge := challengeBody(resp)
	if !isChallenge {
		return resp, nil
	}
	// Put the buffered body back for the caller in case we cannot clear it.
	restoreBody(resp, body)

	if !netx.ChallengeSolverAvailable() {
		util.Debug("Goyabu: challenged, but no browser solver is registered", "url", req.URL.String())
		return resp, nil
	}

	// Being challenged on a request we believed was CLEARED means the clearance
	// is stale — cf_clearance expires, and a restored one can already be dead on
	// arrival. Without dropping it here, ensureCleared sees a non-empty
	// User-Agent, concludes there is nothing to do, and the source stays broken
	// for the life of the process: a restored-but-expired clearance turned
	// Goyabu into a permanent 403 with no solve ever attempted.
	if used != "" {
		t.invalidateClearance(used)
	}

	ua, ok := t.ensureCleared(req)
	if !ok {
		return resp, nil // solve failed; hand back the challenge we have
	}

	_ = resp.Body.Close()
	return t.send(req, ua)
}

// send issues req, using the cleared path when ua is set.
//
// The per-attempt deadline must outlive this function. A `defer cancel()` here
// tears the context down the moment the headers are back, while the caller is
// still reading the body — goquery then fails with "context canceled" on a
// response that arrived perfectly. So the cancel is attached to the BODY and
// fires when the caller closes it, which is the only moment the request is
// genuinely over.
func (t *gateTransport) send(req *http.Request, ua string) (*http.Response, error) {
	ctx := req.Context()
	cancel := context.CancelFunc(func() {})
	if _, ok := ctx.Deadline(); !ok {
		ctx, cancel = context.WithTimeout(ctx, httpAttemptTimeout)
	}

	clone := req.Clone(ctx)
	rt := t.surf
	if ua != "" {
		// Cleared: our headers must survive, so the plain transport carries it.
		clone.Header.Set("User-Agent", ua)
		for _, c := range t.jar.Cookies(clone.URL) {
			clone.AddCookie(c)
		}
		rt = t.plain
	}

	resp, err := rt.RoundTrip(clone)
	if err != nil || resp == nil {
		cancel()
		return resp, err
	}
	resp.Body = &cancelOnClose{ReadCloser: resp.Body, cancel: cancel}
	return resp, nil
}

// cancelOnClose releases a request's context when its body is closed.
type cancelOnClose struct {
	io.ReadCloser
	once   sync.Once
	cancel context.CancelFunc
}

func (c *cancelOnClose) Close() error {
	err := c.ReadCloser.Close()
	c.once.Do(c.cancel)
	return err
}

// ensureCleared runs a solve, or waits for the one already running, and reports
// the User-Agent to replay with.
func (t *gateTransport) ensureCleared(req *http.Request) (string, bool) {
	t.mu.Lock()
	for t.solving {
		t.cond.Wait()
	}
	if t.ua != "" {
		ua := t.ua
		t.mu.Unlock()
		return ua, true
	}
	if !t.failedAt.IsZero() && time.Since(t.failedAt) < solveRetryCooldown {
		t.mu.Unlock()
		util.Debug("Goyabu: skipping the solve; the last one failed recently",
			"retryIn", solveRetryCooldown-time.Since(t.failedAt))
		return "", false
	}
	t.solving = true
	t.mu.Unlock()

	ua, ok := t.solve(req)

	t.mu.Lock()
	t.solving = false
	if ok {
		t.ua = ua
		t.failedAt = time.Time{}
	} else {
		t.failedAt = time.Now()
	}
	t.cond.Broadcast()
	t.mu.Unlock()
	return ua, ok
}

func (t *gateTransport) solve(req *http.Request) (string, bool) {
	target := &url.URL{Scheme: req.URL.Scheme, Host: req.URL.Host, Path: "/"}
	// Says what is happening, not what will be on screen. The window usually is
	// not: with a warm profile the gate falls in under two seconds with nothing
	// shown, and solveGate only surfaces it once it is clear the challenge will
	// not pass on its own — at which point it prints its own line.
	util.Info("Goyabu pediu verificação — resolvendo automaticamente (leva alguns segundos).")

	// The solve gets its OWN budget, detached from the request that triggered
	// it. A caller's deadline is sized for an HTTP round trip, not for a browser
	// working through a managed challenge, and inheriting it aborts the solve
	// before it can finish. Values are kept so logging and tracing still follow.
	ctx, cancel := context.WithTimeout(context.WithoutCancel(req.Context()), gateSolveTimeout)
	defer cancel()

	// RevealWhenStuck, not a forced window. The earlier measurement that put
	// visible=true here was taken with the profile wiped between runs, which is
	// the cold case only; re-measured 2026-09-24 on three consecutive solves, a
	// warm profile clears hidden in 1.8s against 1.6s visible, and only a cold
	// one fails to clear hidden at all. Forcing the window made every search
	// pay the cold case's price.
	res, _, err := netx.SolveChallengeFor(ctx, target.String(), gateSolveTimeout, netx.RevealWhenStuck)
	if err != nil || res == nil {
		util.Debug("Goyabu: challenge solve failed", "url", target.String(), "err", err)
		return "", false
	}
	if len(res.Cookies) > 0 {
		t.jar.SetCookies(target, res.Cookies)
		t.saveClearance(target, res.UserAgent, res.Cookies)
	}
	util.Debug("Goyabu: challenge cleared", "cookies", len(res.Cookies), "ua", res.UserAgent)
	return res.UserAgent, res.UserAgent != ""
}

// challengeBody reads resp far enough to tell whether it is a Cloudflare
// interstitial, returning the bytes it consumed so they can be put back.
//
// Only suspect responses are buffered: a 200 that is not HTML streams through
// untouched, so the steady-state cost is nil.
func challengeBody(resp *http.Response) (body []byte, isChallenge bool) {
	if resp == nil {
		return nil, false
	}
	if resp.Header.Get("cf-mitigated") == "challenge" {
		body, _ = io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return body, true
	}
	switch resp.StatusCode {
	case http.StatusForbidden, http.StatusServiceUnavailable, http.StatusTooManyRequests:
	default:
		return nil, false
	}
	body, _ = io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return body, hasChallengeMarker(body)
}

// restoreBody makes an already-read body readable again by the caller.
func restoreBody(resp *http.Response, body []byte) {
	_ = resp.Body.Close()
	resp.Body = io.NopCloser(bytes.NewReader(body))
}

// cfChallengeMarkers are the substrings Cloudflare's interstitial carries.
// Captured from the live 403: "Just a moment…", the challenge-platform script
// and the __cf_chl / cf_chl globals it defines.
var cfChallengeMarkers = [][]byte{
	[]byte("__cf_chl"),
	[]byte("cf_chl_opt"),
	[]byte("/cdn-cgi/challenge-platform/h/"),
	[]byte("Just a moment..."),
	[]byte("Just a moment…"),
	[]byte("challenges.cloudflare.com/turnstile"),
	[]byte("Checking your browser before accessing"),
}

func hasChallengeMarker(body []byte) bool {
	for _, m := range cfChallengeMarkers {
		if bytes.Contains(body, m) {
			return true
		}
	}
	return false
}
