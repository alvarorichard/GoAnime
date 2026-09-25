package superflix

import (
	"context"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
)

// A window the user can SEE has to go the moment it is done.
//
// The idle watchdog closes an unused browser, and measured it did exactly that
// — sixteen seconds after the solve. Correct, and far too long to sit looking
// at a browser you did not ask for; it was reported as the window not closing
// at all. A hidden window can wait out the timer, a revealed one cannot, so the
// solve closes its own context when it was the one that put it on screen.
//
// Measured after the change: 1.2s.
//
// Live-gated: this is about a real OS window, and a fake cannot be seen.
// Opt in with GOANIME_LIVE=1; never runs in CI.
func TestRevealedWindowClosesWithoutWaitingForTheIdleTimer(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against a real browser")
	}

	// Only real browser binaries: a pgrep over full command lines also matches
	// the shell and test commands that mention the same paths, which is how an
	// earlier version of this counter reported two browsers that were its own
	// tooling.
	browsers := func() int {
		n := 0
		for _, name := range []string{"chrome", "chromium"} {
			out, _ := exec.Command("pgrep", "-a", "-x", name).Output()
			for _, l := range strings.Split(strings.TrimSpace(string(out)), "\n") {
				if strings.TrimSpace(l) != "" {
					n++
				}
			}
		}
		return n
	}

	before := browsers()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	// RevealAlways so the window is definitely on screen, whatever the gate does.
	// Whether it CLEARS is upstream's business; what is being tested is that the
	// window it opened goes away.
	_, _ = defaultCFSolver.solveGate(ctx, "https://goyabu.io/", 20*time.Second, netx.RevealAlways)

	// Well inside browserIdleTimeout, so a pass cannot be the watchdog.
	deadline := time.Now().Add(5 * time.Second)
	for time.Now().Before(deadline) {
		if browsers() <= before {
			return
		}
		time.Sleep(250 * time.Millisecond)
	}
	assert.LessOrEqual(t, browsers(), before,
		"a window the solve put on screen was still there %s later; the idle watchdog is a backstop, not the answer for a visible window",
		browserIdleTimeout)
}
