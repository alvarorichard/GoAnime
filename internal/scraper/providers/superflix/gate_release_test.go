package superflix

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// ReleaseSharedBrowser is deferred on every SuperFlix stream resolve, including
// the cache fast path where no browser window was ever opened. It must be a safe
// no-op then (nil context), never a panic.
func TestReleaseSharedBrowser_SafeWithNoWindow(t *testing.T) {
	// defaultCFSolver has no live context in a unit-test process.
	assert.NotPanics(t, func() {
		ReleaseSharedBrowser()
		ReleaseSharedBrowser() // idempotent
	})
	assert.Nil(t, defaultCFSolver.pctx, "no context must remain after release")
}
