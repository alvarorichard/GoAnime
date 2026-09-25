package superflix

import (
	"context"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"os"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
)

// A solve must not leave a window on screen.
//
// Reported with a screenshot: a white Chrome window sitting on about:blank long
// after the search had finished. Two mistakes produced it.
//
// solveGate created a FRESH page. A persistent context always already has one,
// parked at about:blank in its own window since launch, and nothing in this
// path ever touched it — so a solve put a second window up and left the first
// one there. Solve() has always reused that tab; this now does too.
//
// And the window it does raise has to come back down: showSolverWindow puts one
// on screen whatever --sf-offscreen says, so the counterpart has to undo exactly
// that, regardless of configuration.
//
// Only a real browser can show this, so the check runs against one:
//
//	GOANIME_RECON=1 GOANIME_LIVE=1 go test ./internal/scraper/providers/superflix/ \
//	  -run TestSolveGate_LeavesNoWindowOnScreen_Live -v -count=1 -timeout 200s
func TestSolveGate_LeavesNoWindowOnScreen_Live(t *testing.T) {
	if os.Getenv("GOANIME_RECON") == "" {
		t.Skip("set GOANIME_RECON=1 (drives a real browser against a live host)")
	}
	skipInCI(t)
	if testing.Short() {
		t.Skip("skipping live browser recon in -short")
	}
	util.InitLogger()
	t.Cleanup(defaultCFSolver.Close)

	// The outcome is deliberately ignored. This gate is a managed challenge on a
	// hostile endpoint — measured clearing in 6s once and not at all within 90s
	// another time — so requiring success would make the test flake on upstream
	// mood. The invariant does not depend on it: whether the solve clears or
	// gives up, the defer runs and the window has to be gone. The failure path
	// is the one that matters more, since that is when a user is left staring at
	// a browser that achieved nothing.
	_, solveErr := defaultCFSolver.solveGate(context.Background(), "https://goyabu.io/", 90*time.Second, netx.RevealAlways)
	t.Logf("solve outcome (not asserted): %v", solveErr)

	// The browser has to be GONE, not merely out of the way. A minimized window
	// is still a window and still a live Chrome, which is what the screenshot
	// showed: the solver parked on the cleared page after the search finished.
	defaultCFSolver.lifeMu.Lock()
	ctxAfter := defaultCFSolver.pctx
	defaultCFSolver.lifeMu.Unlock()
	assert.Nil(t, ctxAfter, "the solver context must be released once the clearance is in hand")

	// And releasing it must not cost the NEXT solve: the driver and the on-disk
	// profile survive, so a second solve has to relaunch and clear again. This
	// is the half that a previous attempt got wrong.
	second, secondErr := defaultCFSolver.solveGate(context.Background(), "https://goyabu.io/", 90*time.Second, netx.RevealAlways)
	t.Logf("second solve outcome: %v", secondErr)
	if secondErr == nil {
		assert.NotEmpty(t, second.Cookies, "a relaunched solve must still produce clearance")
	} else {
		assert.NotContains(t, secondErr.Error(), "launch browser",
			"closing the context must not leave the profile locked against the next launch")
	}
}
