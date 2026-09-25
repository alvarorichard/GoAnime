package superflix

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"runtime"
	"testing"
	"time"

	"github.com/mxschmitt/playwright-go"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The play click is only safe because the traps are disarmed first, so the
// disarming is what has to be proved — in a real browser, against a page that
// springs all three traps, because none of them exist in a mocked DOM.
//
// Local httptest server and a local browser: no upstream, nothing to rate
// limit. Skipped (not failed) where no browser is available, matching the other
// browser-backed tests in this package.

// trapPage serves a play button that tries every route to a pop-under.
func trapServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/opened" {
			_, _ = w.Write([]byte(`<!doctype html><title>POPUNDER</title><body>ad</body>`))
			return
		}
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<!doctype html><html><body style="margin:0">
			<a id="blanklink" href="/opened" target="_blank">ad</a>
			<div class="jw-icon-display" id="play"
			     style="width:200px;height:200px;background:#333"
			     onclick="window.__clicked=true;window.open('/opened','_blank');document.getElementById('blanklink').click();">
			</div>
			<video id="v" muted></video>
			</body></html>`)
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestPopunderGuard_DisarmsThePlayClick(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real-browser test in -short")
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

	srv := trapServer(t)

	t.Run("an armed guard blocks every route to a pop-under", func(t *testing.T) {
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { _ = page.Close() })

		guard := &popunderGuard{}
		disarm := guard.install(page, bctx)
		t.Cleanup(disarm)

		_, err = page.Goto(srv.URL + "/trap")
		require.NoError(t, err)

		before := len(bctx.Pages())
		triggerPlay(page, true)
		time.Sleep(1500 * time.Millisecond) // give a pop-under time to appear

		clicked, evalErr := page.Evaluate("() => !!window.__clicked")
		require.NoError(t, evalErr)
		assert.Equal(t, true, clicked,
			"the overlay was never clicked, so this test proves nothing about the guard")

		// window.open is neutered and target is stripped, so nothing should
		// have opened; anything that slipped through is closed by the handler.
		assert.LessOrEqual(t, len(bctx.Pages()), before,
			"a pop-under survived the guard (blocked=%d)", guard.blocked())

		for _, p := range bctx.Pages() {
			title, tErr := p.Title()
			if tErr == nil {
				assert.NotEqual(t, "POPUNDER", title, "an ad page is on screen")
			}
		}

		// And the player is still the thing on screen.
		where, wErr := page.Evaluate("() => location.pathname")
		require.NoError(t, wErr)
		assert.Equal(t, "/trap", where, "the ad navigated the player away")
	})

	t.Run("the in-page script neuters window.open and target=_blank", func(t *testing.T) {
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { _ = page.Close() })

		guard := &popunderGuard{}
		t.Cleanup(guard.install(page, bctx))

		_, err = page.Goto(srv.URL + "/trap")
		require.NoError(t, err)

		opened, err := page.Evaluate(`() => window.open('/opened','_blank') === null`)
		require.NoError(t, err)
		assert.Equal(t, true, opened, "window.open still returns a window")

		// The anchor click must be cancelled, not merely retargeted: retargeting
		// navigates the CURRENT document, which replaces the player being sniffed.
		navigated, err := page.Evaluate(`() => {
			document.getElementById('blanklink').click();
			return location.pathname;
		}`)
		require.NoError(t, err)
		assert.Equal(t, "/trap", navigated,
			"clicking an ad anchor navigated the player away")
	})

	// Without this, the guard would be a leak: it attaches to a shared context
	// that outlives the sniff, and a handler that never stops firing would close
	// pages other work legitimately opens.
	t.Run("a disarmed guard leaves new pages alone", func(t *testing.T) {
		page, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { _ = page.Close() })

		guard := &popunderGuard{}
		guard.install(page, bctx)()

		other, err := bctx.NewPage()
		require.NoError(t, err)
		t.Cleanup(func() { _ = other.Close() })
		_, err = other.Goto(srv.URL + "/opened")
		require.NoError(t, err)

		time.Sleep(500 * time.Millisecond)
		assert.Zero(t, guard.blocked(), "a disarmed guard closed a page it had no business touching")
		assert.False(t, other.IsClosed())
	})
}

// The overlay click must stay off until a caller asks for it, because the
// caller is what arms the guard. A default-on click would spring the traps on
// every path that calls triggerPlay, including the gate warm-up.
func TestTriggerPlay_OverlayClickIsOptIn(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping real-browser test in -short")
	}
	if runtime.GOOS == "windows" {
		t.Skip("playwright driver path on windows needs extra setup")
	}

	s := &cfBrowserSolver{}
	t.Cleanup(s.Close)
	bctx, err := s.init()
	if err != nil {
		t.Skipf("browser unavailable: %v", err)
	}
	srv := trapServer(t)

	page, err := bctx.NewPage()
	require.NoError(t, err)
	t.Cleanup(func() { _ = page.Close() })
	t.Cleanup((&popunderGuard{}).install(page, bctx))

	_, err = page.Goto(srv.URL + "/trap")
	require.NoError(t, err)

	triggerPlay(page, false)
	time.Sleep(300 * time.Millisecond)
	clicked, err := page.Evaluate("() => !!window.__clicked")
	require.NoError(t, err)
	assert.Equal(t, false, clicked, "the overlay was clicked without the caller asking for it")

	triggerPlay(page, true)
	time.Sleep(300 * time.Millisecond)
	clicked, err = page.Evaluate("() => !!window.__clicked")
	require.NoError(t, err)
	assert.Equal(t, true, clicked, "the caller opted in and the overlay was still not clicked")
}

var _ = playwright.Script{}
