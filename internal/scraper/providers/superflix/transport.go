package superflix

import (
	"bytes"
	"context"
	"io"
	"net/http"
	"net/http/cookiejar"
	"net/url"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"golang.org/x/net/publicsuffix"
)

// Retry-After backoff bounds for 429 rate-limits. Cloudflare sends a clear
// Retry-After value on a 429; honoring it (instead of retrying immediately or
// pushing the 429 into the browser solver) is what clears the block. The caps
// stop a hostile/misconfigured edge advertising an enormous delay from hanging
// the request forever.
const (
	maxRetryAfterWait  = 30 * time.Second // longest single wait we'll honor
	retryAfterBudget   = 90 * time.Second // total time spent waiting across retries
	maxRetryAfterTries = 4
	defaultRetryAfter  = 5 * time.Second // used when the header is missing/unparseable
)

// parseRetryAfter extracts the wait from a Retry-After header, which is either
// delta-seconds ("120") or an HTTP-date. Returns 0 when absent or unparseable.
func parseRetryAfter(resp *http.Response) time.Duration {
	v := strings.TrimSpace(resp.Header.Get("Retry-After"))
	if v == "" {
		return 0
	}
	if secs, err := strconv.Atoi(v); err == nil {
		if secs < 0 {
			return 0
		}
		return time.Duration(secs) * time.Second
	}
	if when, err := http.ParseTime(v); err == nil {
		if d := time.Until(when); d > 0 {
			return d
		}
	}
	return 0
}

// cfSolver is the minimal interface cfFallbackTransport needs from a
// challenge solver. Defined here (not as `*cfBrowserSolver`) so tests can
// substitute a fake without launching Chromium.
type cfSolver interface {
	Solve(ctx context.Context, targetURL string, timeout time.Duration) (*CFSolveResult, error)
}

// cfFallbackTransport wraps an existing RoundTripper. When the underlying
// response signals a Cloudflare interstitial (status 403/503/429 + CF
// markers), it delegates to a browser solver to obtain cf_clearance, stores
// the resulting cookies in the shared jar, and retries the original request
// once with the cookies attached.
//
// Why a transport (not a per-call helper): SuperFlix has six distinct call
// sites that hit superflixapi.best. A transport intercepts all of them
// without forcing every caller to know about the fallback.
//
// Body buffering cost: only paid on 403/503/429 (the suspect statuses).
// 2xx responses stream through untouched, so the steady-state perf cost is
// zero.
type cfFallbackTransport struct {
	base    http.RoundTripper
	solver  cfSolver
	jar     http.CookieJar
	timeout time.Duration

	// solvedUA holds the User-Agent of the real browser that solved the gate.
	// Cloudflare binds cf_clearance to the UA, so once we have it every
	// subsequent request must send the SAME UA or the clearance is rejected.
	// Guarded by uaMu; empty until the first successful solve.
	uaMu     sync.Mutex
	solvedUA string
}

func (t *cfFallbackTransport) getSolvedUA() string {
	t.uaMu.Lock()
	defer t.uaMu.Unlock()
	return t.solvedUA
}

func (t *cfFallbackTransport) setSolvedUA(ua string) {
	t.uaMu.Lock()
	t.solvedUA = ua
	t.uaMu.Unlock()
}

// suspectStatus returns true for HTTP statuses that Cloudflare classically
// uses for challenges.
func suspectStatus(code int) bool {
	return code == http.StatusForbidden ||
		code == http.StatusServiceUnavailable ||
		code == http.StatusTooManyRequests
}

// honorRetryAfter429 waits out Cloudflare's Retry-After on a 429 and retries the
// request, up to maxRetryAfterTries times and within retryAfterBudget total.
//
// Only idempotent, body-less requests (GET/HEAD) are retried transparently — a
// POST could have side effects, so it's returned untouched. When the header is
// missing/unparseable it falls back to defaultRetryAfter so we still ease off
// instead of spinning. Respects request-context cancellation while waiting.
// Returns the latest response (possibly still 429 if the budget runs out, which
// then falls through to the normal CF-marker handling).
func (t *cfFallbackTransport) honorRetryAfter429(req *http.Request, resp *http.Response) (*http.Response, error) {
	if resp.StatusCode != http.StatusTooManyRequests {
		return resp, nil
	}
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		return resp, nil
	}

	var spent time.Duration
	for tries := 0; tries < maxRetryAfterTries && resp.StatusCode == http.StatusTooManyRequests; tries++ {
		wait := parseRetryAfter(resp)
		if wait <= 0 {
			wait = defaultRetryAfter
		}
		if wait > maxRetryAfterWait {
			wait = maxRetryAfterWait
		}
		if spent+wait > retryAfterBudget {
			break // out of budget — hand the 429 back to the caller
		}
		spent += wait

		util.Debug("SuperFlix 429 rate-limited; honoring Retry-After", "wait", wait, "try", tries+1, "url", req.URL.String())

		// Discard the 429 body before dropping the response.
		_, _ = io.Copy(io.Discard, io.LimitReader(resp.Body, 1<<16))
		_ = resp.Body.Close()

		select {
		case <-req.Context().Done():
			return nil, req.Context().Err()
		case <-time.After(wait):
		}

		newResp, err := t.base.RoundTrip(req)
		if err != nil {
			return nil, err
		}
		resp = newResp
	}
	return resp, nil
}

// shouldInspect reports whether a response is worth buffering for a CF-marker
// check. Besides the classic challenge statuses, SuperFlix's current host
// serves a Turnstile gate page on a plain 200 with Content-Type text/html
// (title "Verificação"). We therefore also inspect HTML 200s — but never
// JSON/video 200s, so the steady-state cost stays zero for real API and media
// responses.
func shouldInspect(resp *http.Response) bool {
	if suspectStatus(resp.StatusCode) {
		return true
	}
	if resp.StatusCode == http.StatusOK {
		ct := resp.Header.Get("Content-Type")
		return strings.Contains(ct, "text/html")
	}
	return false
}

func (t *cfFallbackTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Once a solve has happened, force every request to carry the browser's UA
	// so the UA-bound cf_clearance cookie stays valid.
	if ua := t.getSolvedUA(); ua != "" {
		req.Header.Set("User-Agent", ua)
	}

	resp, err := t.base.RoundTrip(req)
	if err != nil {
		return resp, err
	}

	// Honor Cloudflare's Retry-After on a 429 BEFORE anything else. Waiting out
	// the advertised delay and retrying is the documented way to clear a
	// rate-limit block; retrying immediately or handing the 429 to the browser
	// solver only deepens it.
	resp, err = t.honorRetryAfter429(req, resp)
	if err != nil {
		return resp, err
	}

	if !shouldInspect(resp) {
		return resp, nil
	}

	// Buffer body for marker check. 1MB cap protects against a hostile
	// edge that streams a multi-GB body on 403 — won't happen in
	// practice, but the cap is cheap insurance.
	body, readErr := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	_ = resp.Body.Close()
	if readErr != nil {
		return nil, readErr
	}

	// Route through the solver on a CF challenge OR the post-gate restricted
	// embed page (which needs the browser to follow its embed iframe). With a
	// warm CF profile the gate auto-clears and the server returns the restricted
	// page directly — no CF markers — so checking only IsCloudflareChallenge
	// would hand that useless page back to the caller.
	if !IsCloudflareChallenge(resp, body) && !isRestrictedEmbedPage(body) {
		resp.Body = io.NopCloser(bytes.NewReader(body))
		return resp, nil
	}

	if t.solver == nil {
		// CF detected but no solver wired (test/headless-disabled path).
		// Restore body so caller sees the real challenge HTML.
		resp.Body = io.NopCloser(bytes.NewReader(body))
		return resp, nil
	}

	util.Debug("SuperFlix CF challenge detected, delegating to browser solver", "url", req.URL.String(), "status", resp.StatusCode)

	timeout := t.timeout
	if timeout <= 0 {
		timeout = 30 * time.Second
	}
	target := req.URL.String()
	res, solveErr := t.solver.Solve(req.Context(), target, timeout)
	if solveErr != nil {
		util.Debug("SuperFlix CF solve failed", "err", solveErr)
		resp.Body = io.NopCloser(bytes.NewReader(body))
		return resp, nil
	}

	if res.UserAgent != "" {
		// Remember the browser UA so all follow-up requests match the
		// cf_clearance binding (and fix up THIS request's retry below).
		t.setSolvedUA(res.UserAgent)
		req.Header.Set("User-Agent", res.UserAgent)
	}

	if t.jar != nil && len(res.Cookies) > 0 {
		// Populate jar so subsequent http.Client.Do() calls auto-attach
		// the cookies (the site's verification cookie) on follow-up requests.
		t.jar.SetCookies(req.URL, res.Cookies)
	}

	// Fast path: the solver may have already captured the REAL player HTML
	// (read from the embed iframe in the browser). If so, hand it straight back
	// for a GET — no HTTP re-fetch can do better and the bare/embed URL fetched
	// over HTTP may not reproduce the iframe Sec-Fetch context.
	if req.Method == http.MethodGet &&
		(strings.Contains(res.HTML, "CSRF_TOKEN") || strings.Contains(res.HTML, "ALL_EPISODES")) {
		hdr := http.Header{}
		hdr.Set("Content-Type", "text/html; charset=utf-8")
		return &http.Response{
			StatusCode:    http.StatusOK,
			Status:        "200 OK",
			Proto:         resp.Proto,
			ProtoMajor:    resp.ProtoMajor,
			ProtoMinor:    resp.ProtoMinor,
			Header:        hdr,
			Body:          io.NopCloser(strings.NewReader(res.HTML)),
			ContentLength: int64(len(res.HTML)),
			Request:       req,
		}, nil
	}

	// Redirect the retry to the POST-GATE url. After Turnstile clears, the gate
	// bounces to serie/<id>?cfv=<JWT> (scope "embed"). Loaded top-level that URL
	// shows an "Acesso Restrito / Visualização Externa" page — the real player
	// is only served when the SAME ?cfv url is fetched in IFRAME context. Our
	// original request already carries the iframe Sec-Fetch-* headers (set by
	// GetPlayerPage), so re-issuing it against the ?cfv url returns the real
	// player HTML with ALL_EPISODES/CSRF_TOKEN. The bare url would just re-gate.
	if res.FinalURL != "" {
		if u, perr := url.Parse(res.FinalURL); perr == nil && u.Host != "" {
			req.URL = u
			req.Host = u.Host
		}
	}

	// Attach verification cookies. We are inside RoundTrip — the http.Client's
	// jar→Cookie-header step already ran for THIS request and won't run again.
	for _, c := range res.Cookies {
		req.AddCookie(c)
	}

	// Reseed the body for POSTs: the first attempt drained it, but net/http sets
	// GetBody for *bytes.Reader/*bytes.Buffer/*strings.Reader, so the form-
	// encoded SuperFlix POSTs (strings.NewReader) reseed correctly.
	if req.GetBody != nil {
		nb, gErr := req.GetBody()
		if gErr == nil {
			req.Body = nb
		}
	}

	return t.base.RoundTrip(req)
}

// newCookieJar returns a public-suffix-aware cookie jar, matching the
// behavior of a real browser. We need PSL handling so cf_clearance set
// for `.superflixapi.best` correctly attaches to subdomains too.
func newCookieJar() (http.CookieJar, error) {
	return cookiejar.New(&cookiejar.Options{PublicSuffixList: publicsuffix.List})
}
