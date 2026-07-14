package superflix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ReleaseSharedBrowser is deferred on every SuperFlix stream resolve, including
// the cache fast path where no browser window was ever opened. With no live
// context it must be a safe, idempotent no-op — never a panic. It closes the
// context (window) but must NOT tear the driver down: keeping the Playwright
// driver warm is what makes the next episode's relaunch fast (the full teardown
// happens only at app exit via Close).
func TestReleaseSharedBrowser_SafeWithNoWindow(t *testing.T) {
	// defaultCFSolver has no live context/driver in a unit-test process.
	assert.NotPanics(t, func() {
		ReleaseSharedBrowser()
		ReleaseSharedBrowser() // idempotent
	})
	assert.Nil(t, defaultCFSolver.pctx, "no context should remain after release")
	assert.Nil(t, defaultCFSolver.pw, "release must keep (not stop) the Playwright driver")
}
