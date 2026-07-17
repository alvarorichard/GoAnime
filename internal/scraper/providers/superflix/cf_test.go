package superflix

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"net/url"
	"runtime"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// helper: make a *http.Response with a body and given status for marker tests.
func makeResp(t *testing.T, status int, body string, headers map[string]string) (*http.Response, []byte) {
	t.Helper()
	rec := httptest.NewRecorder()
	for k, v := range headers {
		rec.Header().Set(k, v)
	}
	rec.WriteHeader(status)
	_, _ = rec.WriteString(body)
	resp := rec.Result()
	buf, _ := io.ReadAll(resp.Body)
	_ = resp.Body.Close()
	return resp, buf
}

func TestIsCloudflareChallenge(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		status  int
		body    string
		headers map[string]string
		want    bool
	}{
		{"200 JSON not a challenge", 200, `{"ok":true}`, nil, false},
		{"403 JSON rate limit not a challenge", 403, `{"error":"rate"}`, nil, false},
		{"403 HTML treated as challenge", 403, `<html><body>blocked</body></html>`, nil, true},
		{"503 HTML treated as challenge", 503, `<html>maintenance</html>`, nil, true},
		{"cf-mitigated header forces challenge", 200, `{}`, map[string]string{"cf-mitigated": "challenge"}, true},
		{"body contains challenge-platform path", 200, `<script src="/cdn-cgi/challenge-platform/foo.js"></script>`, nil, true},
		{"body contains turnstile script", 200, `<script src="https://challenges.cloudflare.com/turnstile/v0/api.js"></script>`, nil, true},
		{"body contains cf_chl_opt", 200, `var cf_chl_opt = {};`, nil, true},
		{"body contains Just a moment", 200, `<title>Just a moment...</title>`, nil, true},
		{"body contains __cf_chl_", 200, `window.__cf_chl_tk = "abc";`, nil, true},
		{"nil response is safe", 0, "", nil, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			if tt.name == "nil response is safe" {
				assert.False(t, IsCloudflareChallenge(nil, nil))
				return
			}
			resp, body := makeResp(t, tt.status, tt.body, tt.headers)
			assert.Equal(t, tt.want, IsCloudflareChallenge(resp, body))
		})
	}
}

func TestSuspectStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code int
		want bool
	}{
		{200, false},
		{301, false},
		{400, false},
		{403, true},
		{404, false},
		{429, true},
		{500, false},
		{502, false},
		{503, true},
	}
	for _, tt := range tests {
		t.Run(fmt.Sprintf("status_%d", tt.code), func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, suspectStatus(tt.code))
		})
	}
}

func TestShouldInspect(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		status      int
		contentType string
		want        bool
	}{
		{"403 inspected", 403, "", true},
		{"503 inspected", 503, "", true},
		{"429 inspected", 429, "", true},
		// 2026-06-05: SuperFlix's .fit host serves a Turnstile gate on 200
		// text/html — must be inspected even though the status is OK.
		{"200 html inspected", 200, "text/html; charset=utf-8", true},
		{"200 json skipped", 200, "application/json", false},
		{"200 no content-type skipped", 200, "", false},
		{"200 video skipped", 200, "video/mp4", false},
		{"301 skipped", 301, "text/html", false},
		{"404 skipped", 404, "text/html", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			resp := &http.Response{StatusCode: tt.status, Header: http.Header{}}
			if tt.contentType != "" {
				resp.Header.Set("Content-Type", tt.contentType)
			}
			assert.Equal(t, tt.want, shouldInspect(resp))
		})
	}
}

// TestCFFallbackTransport_Solves200HTMLChallenge guards the SuperFlix .fit
// migration: the Turnstile gate arrives on a 200 text/html response, so the
// transport must detect it, solve, and retry — not pass the gate HTML through
// as if it were the real page (which produced "no seasons found").
func TestCFFallbackTransport_Solves200HTMLChallenge(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			// Turnstile gate on HTTP 200 + text/html.
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`<html><head><title>Verificação</title></head><body><div class="cf-turnstile" data-sitekey="x"></div><script src="https://challenges.cloudflare.com/turnstile/v0/api.js"></script></body></html>`))
			return
		}
		c, err := r.Cookie("cf_clearance")
		if err != nil || c.Value != "good-token" {
			w.WriteHeader(401)
			_, _ = w.Write([]byte("no clearance"))
			return
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`<html>var ALL_EPISODES = {"1":[]};</html>`))
	}))
	t.Cleanup(srv.Close)

	jar, err := newCookieJar()
	require.NoError(t, err)
	solver := &fakeSolver{cookies: []*http.Cookie{{Name: "cf_clearance", Value: "good-token", Path: "/"}}}
	tr := &cfFallbackTransport{base: http.DefaultTransport, solver: solver, jar: jar, timeout: time.Second}
	client := &http.Client{Transport: tr, Jar: jar}

	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(body), "ALL_EPISODES")
	assert.Equal(t, int32(1), solver.calls.Load(), "solver called once on the 200 gate")
	assert.Equal(t, int32(2), hits.Load(), "server hit twice (gate + retry)")
}

// TestCFFallbackTransport_GETRefetchesFinalURL guards the SuperFlix gate fix:
// after Turnstile clears, the gate bounces to serie/<id>?cfv=<JWT> (scope
// "embed"). The real player is only served when that ?cfv URL is fetched in
// iframe context, so the transport must re-issue the request against the
// solver's FinalURL (with cookies) — NOT the bare gated URL, and NOT the
// browser's top-level HTML (which is the "Acesso Restrito" embed page).
func TestCFFallbackTransport_GETRefetchesFinalURL(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		hits.Add(1)
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		// The real player is served only when the ?cfv verification token is
		// present (mimics the embed flow). The bare URL re-gates.
		if r.URL.Query().Get("cfv") != "" {
			w.WriteHeader(200)
			_, _ = w.Write([]byte(`<html><head><title>Player | The Boys</title></head><body>var ALL_EPISODES = {"1":[]};</body></html>`))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`<html><head><title>Verificação</title></head><body><form id="cf-turnstile-form"></form></body></html>`))
	}))
	t.Cleanup(srv.Close)

	jar, err := newCookieJar()
	require.NoError(t, err)
	solver := &fakeSolver{
		// The browser's own HTML is the restricted embed page — must be ignored
		// in favour of an HTTP re-fetch of FinalURL.
		html:     `<html><body>Acesso Restrito - Visualização Externa</body></html>`,
		finalURL: srv.URL + "/serie/76479?cfv=token123",
		cookies:  []*http.Cookie{{Name: "sf_verified", Value: "1", Path: "/"}},
	}
	tr := &cfFallbackTransport{base: http.DefaultTransport, solver: solver, jar: jar, timeout: time.Second}
	client := &http.Client{Transport: tr, Jar: jar}

	resp, err := client.Get(srv.URL + "/serie/76479")
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, 200, resp.StatusCode)
	assert.Contains(t, string(body), "ALL_EPISODES", "must return the real player fetched from FinalURL")
	assert.NotContains(t, string(body), "Acesso Restrito", "must NOT return the restricted embed page")
	assert.Equal(t, int32(1), solver.calls.Load())
}

func TestExtractSuperFlixEmbedURL(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		html string
		want string
	}{
		{
			"iframe with cfv",
			`<iframe src="https://superflixapi.fit/serie/76479?cfv=eyJ0.abc" allowfullscreen></iframe>`,
			"https://superflixapi.fit/serie/76479?cfv=eyJ0.abc",
		},
		{
			"html-escaped ampersand",
			`<iframe src="https://x/serie/1?cfv=tok&amp;foo=1"></iframe>`,
			"https://x/serie/1?cfv=tok&foo=1",
		},
		{"no embed", `<div>Acesso Restrito</div>`, ""},
		{"non-cfv src ignored", `<img src="https://x/poster.jpg">`, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, extractSuperFlixEmbedURL(tt.html))
		})
	}
}

func TestHasVerificationParam(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want bool
	}{
		// 2026-06-05: SuperFlix gate bounces through serie/<id>?cfv=<JWT> on
		// success — that transient page has no real content, so it must not be
		// mistaken for the final page.
		{"cfv token", "https://superflixapi.fit/serie/76479?cfv=eyJ0", true},
		{"cf_chl challenge", "https://x/serie/1?__cf_chl_tk=abc", true},
		{"clean serie url", "https://superflixapi.fit/serie/76479", false},
		{"clean with other query", "https://superflixapi.fit/serie/76479?s=1", false},
		{"empty", "", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, hasVerificationParam(tt.url))
		})
	}
}

func TestConvertPlaywrightCookies(t *testing.T) {
	t.Parallel()
	in := []playwright.Cookie{
		{Name: "cf_clearance", Value: "abc", Domain: ".superflixapi.best", Path: "/", HttpOnly: true, Secure: true},
		{Name: "__cf_bm", Value: "xyz", Domain: "superflixapi.best", Path: "/"},
	}
	out := convertPlaywrightCookies(in)
	require.Len(t, out, 2)
	assert.Equal(t, "cf_clearance", out[0].Name)
	assert.Equal(t, "abc", out[0].Value)
	assert.Equal(t, ".superflixapi.best", out[0].Domain)
	assert.True(t, out[0].HttpOnly)
	assert.True(t, out[0].Secure)
	assert.Equal(t, "__cf_bm", out[1].Name)
	assert.False(t, out[1].HttpOnly)
}

func TestNewCookieJar(t *testing.T) {
	t.Parallel()
	jar, err := newCookieJar()
	require.NoError(t, err)
	require.NotNil(t, jar)
	u, _ := url.Parse("https://superflixapi.best/")
	jar.SetCookies(u, []*http.Cookie{{Name: "k", Value: "v"}})
	got := jar.Cookies(u)
	require.Len(t, got, 1)
	assert.Equal(t, "k", got[0].Name)
	assert.Equal(t, "v", got[0].Value)
}

// fakeSolver implements cfSolver in-memory for transport tests.
type fakeSolver struct {
	calls    atomic.Int32
	cookies  []*http.Cookie
	html     string
	finalURL string // when set, overrides the echoed target as FinalURL
	err      error
}

func (f *fakeSolver) Solve(_ context.Context, target string, _ time.Duration) (*CFSolveResult, error) {
	f.calls.Add(1)
	if f.err != nil {
		return nil, f.err
	}
	final := target
	if f.finalURL != "" {
		final = f.finalURL
	}
	return &CFSolveResult{Cookies: f.cookies, HTML: f.html, FinalURL: final}, nil
}

func TestCFFallbackTransport_PassesThrough200(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	solver := &fakeSolver{}
	tr := &cfFallbackTransport{base: http.DefaultTransport, solver: solver, timeout: time.Second}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, 200, resp.StatusCode)
	assert.JSONEq(t, `{"ok":true}`, string(body))
	assert.Equal(t, int32(0), solver.calls.Load(), "solver must not be called on 200")
}

func TestCFFallbackTransport_PassesThrough403JSON(t *testing.T) {
	t.Parallel()
	// 403 with JSON body (e.g. rate-limit) must NOT trigger the solver —
	// only true CF interstitials should.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"error":"forbidden"}`))
	}))
	t.Cleanup(srv.Close)

	solver := &fakeSolver{}
	tr := &cfFallbackTransport{base: http.DefaultTransport, solver: solver, timeout: time.Second}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, 403, resp.StatusCode)
	assert.Equal(t, int32(0), solver.calls.Load())
	assert.Contains(t, string(body), "forbidden")
}

func TestCFFallbackTransport_SolvesAndRetries(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		if n == 1 {
			// First hit: CF challenge.
			w.WriteHeader(403)
			_, _ = w.Write([]byte(`<html><script src="/cdn-cgi/challenge-platform/h/g/orchestrate/jsch/v1"></script></html>`))
			return
		}
		// Retry: must carry cf_clearance from the solver.
		c, err := r.Cookie("cf_clearance")
		if err != nil || c.Value != "good-token" {
			w.WriteHeader(401)
			_, _ = w.Write([]byte("no clearance"))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	jar, err := newCookieJar()
	require.NoError(t, err)
	solver := &fakeSolver{cookies: []*http.Cookie{{Name: "cf_clearance", Value: "good-token", Path: "/"}}}
	tr := &cfFallbackTransport{base: http.DefaultTransport, solver: solver, jar: jar, timeout: time.Second}
	client := &http.Client{Transport: tr, Jar: jar}

	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, 200, resp.StatusCode)
	assert.JSONEq(t, `{"ok":true}`, string(body))
	assert.Equal(t, int32(1), solver.calls.Load(), "solver called once")
	assert.Equal(t, int32(2), hits.Load(), "server hit twice (challenge + retry)")
}

func TestCFFallbackTransport_SolverErrorReturnsChallenge(t *testing.T) {
	t.Parallel()
	// If the solver fails (no browser installed, timeout, etc.) the
	// transport must surface the original challenge response to the caller
	// instead of dropping it. That lets higher layers report a clear
	// "CF blocked, install Chromium" message.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`<html>Just a moment...</html>`))
	}))
	t.Cleanup(srv.Close)

	solver := &fakeSolver{err: errors.New("no browser")}
	tr := &cfFallbackTransport{base: http.DefaultTransport, solver: solver, timeout: time.Second}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, 403, resp.StatusCode)
	assert.Contains(t, string(body), "Just a moment")
	assert.Equal(t, int32(1), solver.calls.Load())
}

func TestCFFallbackTransport_NilSolverPassThrough(t *testing.T) {
	t.Parallel()
	// A SuperFlixClient constructed without a solver (test path) must
	// still return the challenge body so tests can assert on it.
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(503)
		_, _ = w.Write([]byte(`<html>Just a moment...</html>`))
	}))
	t.Cleanup(srv.Close)

	tr := &cfFallbackTransport{base: http.DefaultTransport, solver: nil, timeout: time.Second}
	client := &http.Client{Transport: tr}
	resp, err := client.Get(srv.URL)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, 503, resp.StatusCode)
	assert.Contains(t, string(body), "Just a moment")
}

func TestCFFallbackTransport_RetryPreservesPOSTBody(t *testing.T) {
	t.Parallel()
	// SuperFlix's Bootstrap/Source endpoints are POST application/x-www-form-urlencoded.
	// After the first attempt drains the body, the retry must reseed it so
	// the server actually receives the form payload. http.NewRequest sets
	// GetBody for strings.NewReader, which makes this work — guard it.
	var hits atomic.Int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		n := hits.Add(1)
		_ = r.ParseForm()
		if n == 1 {
			w.WriteHeader(403)
			_, _ = w.Write([]byte(`<html>/cdn-cgi/challenge-platform</html>`))
			return
		}
		if r.PostForm.Get("contentid") != "42" {
			w.WriteHeader(400)
			_, _ = w.Write([]byte("missing form on retry"))
			return
		}
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{"ok":true}`))
	}))
	t.Cleanup(srv.Close)

	jar, _ := newCookieJar()
	solver := &fakeSolver{cookies: []*http.Cookie{{Name: "cf_clearance", Value: "tok", Path: "/"}}}
	tr := &cfFallbackTransport{base: http.DefaultTransport, solver: solver, jar: jar, timeout: time.Second}
	client := &http.Client{Transport: tr, Jar: jar}

	req, err := http.NewRequest("POST", srv.URL, strings.NewReader("contentid=42"))
	require.NoError(t, err)
	req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	resp, err := client.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(resp.Body)
	assert.Equal(t, 200, resp.StatusCode)
	assert.JSONEq(t, `{"ok":true}`, string(body))
}

// TestCFBrowserSolver_Init asserts that init() either succeeds (system Chrome +
// driver present) or fails cleanly without panicking. Launching the real
// browser and fetching the node driver need a desktop + network on first run,
// so a clean error is the expected outcome in offline/headless CI.
func TestCFBrowserSolver_Init(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real-browser launch in -short")
	}
	if runtime.GOOS == "windows" {
		t.Skip("playwright driver path on windows requires extra setup; covered by integration tests")
	}
	s := &cfBrowserSolver{}
	t.Cleanup(s.Close)
	_, err := s.init()
	if err != nil {
		// Acceptable when no system Chrome or no network for the driver.
		// Either way the contract under test is "no panic".
		t.Logf("init returned error (acceptable offline/headless): %v", err)
	}
}

// TestCFBrowserSolver_Close is safe to call multiple times.
func TestCFBrowserSolver_Close(t *testing.T) {
	t.Parallel()
	s := &cfBrowserSolver{}
	s.Close() // before init: no-op
	s.Close() // idempotent
}

// TestCFBrowserSolver_CloseDoesNotBlockOnSolve guards the SIGINT-shutdown fix:
// Close() must use lifeMu, not the solve mutex mu, so it can tear the browser
// down while a solve still holds mu. If Close ever takes mu again, killing the
// app mid-solve would hang for the whole solve budget (~90s) and the Playwright
// driver would EPIPE-crash.
func TestCFBrowserSolver_CloseDoesNotBlockOnSolve(t *testing.T) {
	t.Parallel()
	s := &cfBrowserSolver{}

	// Simulate an in-flight solve holding the solve mutex.
	s.mu.Lock()
	defer s.mu.Unlock()

	done := make(chan struct{})
	go func() {
		s.Close() // no browser launched → just exercises the locking path
		close(done)
	}()

	select {
	case <-done:
		// Close returned without waiting on mu — correct.
	case <-time.After(2 * time.Second):
		t.Fatal("Close blocked while a solve held mu — shutdown would hang on SIGINT")
	}
}

// TestCFBrowserSolver_Solve exercises the real Chrome over CDP. Requires a
// system Chrome + the node driver + a desktop session. Skipped (not failed)
// when init can't bring up the browser so headless/offline CI doesn't hang.
func TestCFBrowserSolver_Solve(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping browser integration in -short")
	}
	s := &cfBrowserSolver{}
	t.Cleanup(s.Close)
	if _, err := s.init(); err != nil {
		t.Skipf("chrome unavailable: %v", err)
	}
	// example.com has no CF gate, so Solve must succeed immediately (the
	// gate-cleared poll returns at once) and return real HTML. This exercises
	// navigation, content capture, cookie + UA capture.
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	res, err := s.Solve(ctx, "https://example.com/", 20*time.Second)
	require.NoError(t, err, "example.com is not gated; solve must succeed")
	require.NotNil(t, res)
	assert.Contains(t, res.HTML, "Example Domain")
	assert.NotEmpty(t, res.UserAgent)
}
