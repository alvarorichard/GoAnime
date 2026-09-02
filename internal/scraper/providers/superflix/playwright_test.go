package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// localEmbedServer serves the controlled, Cloudflare-free pages the Playwright
// helper tests drive against. Deterministic: no live host, no gate.
func localEmbedServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/player", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><body>
<script>var ALL_EPISODES = {"1":[{"epi_num":"1"}]};</script>
player ready</body></html>`))
	})
	mux.HandleFunc("/child", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><body>child frame</body></html>`))
	})
	mux.HandleFunc("/parent", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><body>
<iframe src="/child" style="width:300px;height:200px"></iframe></body></html>`))
	})
	// Reproduces SuperFlix's challenge shell: the retry control is display:none
	// until the Turnstile widget reports a load failure, then gets the "show"
	// class. ?failed=1 renders it already shown.
	mux.HandleFunc("/challenge", func(w http.ResponseWriter, r *http.Request) {
		show := ""
		if r.URL.Query().Get("failed") != "" {
			show = " show"
		}
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><head><style>
.captcha-retry{display:none;width:300px;height:44px}
.captcha-retry.show{display:block}
</style></head><body>
<div id="cfw-deadbeef" class="cf-turnstile-placeholder"></div>
<button type="button" id="cfr-deadbeef" class="captcha-retry` + show + `"
        onclick="window.__retryClicks=(window.__retryClicks||0)+1">Tentar novamente</button>
</body></html>`))
	})
	mux.HandleFunc("/normal", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(`<!doctype html><html><body><h1>regular page</h1>
<video src="/x.mp4"></video></body></html>`))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv
}

// TestPlaywrightHelpers exercises the real page-driving helpers against a local
// server, so the browser code paths (frame walking, iframe injection, content
// capture) get real Chromium coverage WITHOUT the flakiness of a live gated
// host. One browser is launched for all subtests (efficiency); each subtest gets
// a fresh page.
//
// Skipped (not failed) under -short or when no browser/desktop is available, so
// headless/offline CI never hangs — matching TestCFBrowserSolver_Solve.
func TestPlaywrightHelpers(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real-browser helper tests in -short")
	}
	if runtime.GOOS == "windows" {
		t.Skip("playwright driver path on windows needs extra setup")
	}

	s := &cfBrowserSolver{}
	t.Cleanup(s.Close)
	bctx, err := s.init()
	if err != nil {
		t.Skipf("browser unavailable (acceptable offline/headless): %v", err)
	}

	srv := localEmbedServer(t)

	t.Run("child frame content is readable across the frame boundary", func(t *testing.T) {
		// readEmbeddedPlayer walks page.Frames() and reads each frame's Content().
		// This proves that capability end-to-end in real Chromium: a child iframe's
		// HTML is reachable from the parent page object.
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { _ = page.Close() })

		_, err = page.Goto(srv.URL + "/parent")
		require.NoError(t, err)

		var childHTML string
		require.Eventually(t, func() bool {
			for _, fr := range page.Frames() {
				if c, cErr := fr.Content(); cErr == nil && strings.Contains(c, "child frame") {
					childHTML = c
					return true
				}
			}
			return false
		}, 10*time.Second, 200*time.Millisecond, "child frame content must become readable")
		assert.Contains(t, childHTML, "child frame")
	})

	t.Run("readEmbeddedPlayer times out cleanly without player markers", func(t *testing.T) {
		// /child has no ALL_EPISODES / data-episode-id, so readEmbeddedPlayer must
		// drive the iframe wrapper, find nothing real, and return its deadline error
		// (rather than hang or panic). Exercises about:blank reset + SetContent +
		// the frame-poll loop on the negative path.
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { _ = page.Close() })

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		_, err = readEmbeddedPlayer(ctx, page, srv.URL+"/child", time.Now().Add(3*time.Second))
		require.Error(t, err)
		assert.Contains(t, err.Error(), "did not yield content")
	})

	t.Run("embedFrameLive and pageBlankedOut track frame state", func(t *testing.T) {
		// Live page with a child iframe.
		live, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { _ = live.Close() })
		_, err = live.Goto(srv.URL + "/parent")
		require.NoError(t, err)
		assert.False(t, pageBlankedOut(live), "navigated page is not blank")
		assert.True(t, embedFrameLive(live), "the embedded child frame is live")

		// A fresh page sits on about:blank with only its main frame.
		blank, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { _ = blank.Close() })
		assert.True(t, pageBlankedOut(blank), "fresh page is blanked out")
		assert.False(t, embedFrameLive(blank), "no live child frame on a blank page")
	})

	t.Run("injectEmbedCrossOrigin yields a live child frame", func(t *testing.T) {
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { _ = page.Close() })

		require.NoError(t, injectEmbedCrossOrigin(page, srv.URL+"/child"))
		assert.True(t, embedFrameLive(page), "injected cross-origin embed must be a live frame")
	})

	t.Run("clickChallengeRetry presses a shown retry button", func(t *testing.T) {
		// The recovery for a Turnstile that fails to LOAD: there is no checkbox
		// to tick, only the page's own retry control.
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { forgetRevealedPage(page); _ = page.Close() })

		_, err = page.Goto(srv.URL + "/challenge?failed=1")
		require.NoError(t, err)

		assert.True(t, clickChallengeRetry(page), "a visible retry button must be pressed")

		clicks, err := page.Evaluate(`() => window.__retryClicks || 0`)
		require.NoError(t, err)
		assert.EqualValues(t, 1, clicks, "the press must reach the button's handler")
	})

	t.Run("clickChallengeRetry ignores a hidden retry button", func(t *testing.T) {
		// While the widget is still verifying the button is display:none.
		// Pressing then would interrupt an attempt that may yet succeed.
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { forgetRevealedPage(page); _ = page.Close() })

		_, err = page.Goto(srv.URL + "/challenge")
		require.NoError(t, err)

		assert.False(t, clickChallengeRetry(page), "a hidden button means no failure to recover from")

		clicks, err := page.Evaluate(`() => window.__retryClicks || 0`)
		require.NoError(t, err)
		assert.EqualValues(t, 0, clicks)
	})

	t.Run("clickChallengeRetry honours its cooldown", func(t *testing.T) {
		// The widget needs seconds to retry and fail again; clicking every tick
		// would interrupt an attempt still in flight.
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { forgetRevealedPage(page); _ = page.Close() })

		_, err = page.Goto(srv.URL + "/challenge?failed=1")
		require.NoError(t, err)

		require.True(t, clickChallengeRetry(page))
		assert.False(t, clickChallengeRetry(page), "an immediate second press must be held back")

		clicks, err := page.Evaluate(`() => window.__retryClicks || 0`)
		require.NoError(t, err)
		assert.EqualValues(t, 1, clicks, "exactly one press inside the cooldown")
	})

	t.Run("clickTurnstile does not click the empty widget mount", func(t *testing.T) {
		// /challenge renders div.cf-turnstile-placeholder with no Cloudflare
		// iframe inside it. Clicking that container is useless AND harmful: it
		// consumes the one allowed click and keeps a verifying widget spinning.
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { forgetRevealedPage(page); _ = page.Close() })

		_, err = page.Goto(srv.URL + "/challenge?failed=1")
		require.NoError(t, err)

		clickTurnstile(page)

		// The retry button sits right below the mount; if clickTurnstile had
		// matched the placeholder it could have landed on it.
		clicks, err := page.Evaluate(`() => window.__retryClicks || 0`)
		require.NoError(t, err)
		assert.EqualValues(t, 0, clicks, "no stray click from the turnstile path")
	})

	t.Run("hide then reveal moves the solver window", func(t *testing.T) {
		// The smart-offscreen mechanic end to end, in real Chromium: minimize,
		// then restore on demand.
		t.Setenv("GOANIME_SF_OFFSCREEN", "1")
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { forgetRevealedPage(page); _ = page.Close() })

		_, err = page.Goto(srv.URL + "/normal")
		require.NoError(t, err)

		windowState := func() string {
			cdp, cErr := bctx.NewCDPSession(page)
			if cErr != nil {
				t.Skipf("CDP unavailable: %v", cErr)
			}
			win, wErr := cdp.Send("Browser.getWindowForTarget", nil)
			if wErr != nil {
				t.Skipf("window not addressable: %v", wErr)
			}
			m, _ := win.(map[string]any)
			b, _ := m["bounds"].(map[string]any)
			require.NotNil(t, b)
			return fmt.Sprint(b["windowState"])
		}

		hideSolverWindow(page, bctx)
		assert.Eventually(t, func() bool { return windowState() == "minimized" },
			5*time.Second, 250*time.Millisecond, "hiding must minimize the window")

		revealSolverWindow(page, bctx, "test")
		assert.Eventually(t, func() bool { return windowState() == "normal" },
			5*time.Second, 250*time.Millisecond, "revealing must restore it for the user")

		// And hiding must not undo a deliberate reveal.
		hideSolverWindow(page, bctx)
		assert.Equal(t, "normal", windowState(),
			"re-hiding after a reveal would take the captcha away mid-solve")
	})

	t.Run("warmGateTopLevel returns fast on a non-gated page", func(t *testing.T) {
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { _ = page.Close() })

		const budget = 10 * time.Second
		start := time.Now()
		warmGateTopLevel(page, bctx, srv.URL+"/normal", budget)
		elapsed := time.Since(start)
		assert.Less(t, elapsed, budget-2*time.Second,
			"a page with no challenge markup must clear well before the budget")
	})

	t.Run("clickTurnstile is a safe no-op when no widget is present", func(t *testing.T) {
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { _ = page.Close() })

		_, err = page.Goto(srv.URL + "/normal")
		require.NoError(t, err)
		assert.NotPanics(t, func() { clickTurnstile(page) })
	})

	t.Run("triggerPlay does not error on a page with a video element", func(t *testing.T) {
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { _ = page.Close() })

		_, err = page.Goto(srv.URL + "/normal")
		require.NoError(t, err)
		assert.NotPanics(t, func() { triggerPlay(page) })
	})
}
