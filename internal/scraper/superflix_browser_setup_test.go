package scraper

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isolateCacheDir points os.UserCacheDir at a temp dir for the test. It sets
// both XDG_CACHE_HOME (Linux/BSD) and HOME (macOS) so the marker logic never
// touches the real user cache. t.Setenv forbids t.Parallel, so callers run
// serially.
func isolateCacheDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	t.Setenv("HOME", dir)
}

func TestBrowserSetupMarkerPath(t *testing.T) {
	isolateCacheDir(t)

	p := browserSetupMarkerPath()

	// Must live under <cache>/goanime and be the expected marker file.
	assert.Equal(t, ".browser-ready", filepath.Base(p))
	assert.Equal(t, "goanime", filepath.Base(filepath.Dir(p)))
	cache, err := os.UserCacheDir()
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(p, cache),
		"marker %q must be inside the user cache dir %q", p, cache)
}

func TestBrowserSetupPending(t *testing.T) {
	isolateCacheDir(t)

	// Fresh cache: nothing recorded yet, so setup is pending.
	assert.True(t, BrowserSetupPending(), "should be pending on a fresh cache")

	// After marking ready, it must no longer report pending.
	markBrowserReady()
	assert.False(t, BrowserSetupPending(), "should not be pending once marked ready")
}

func TestMarkBrowserReady(t *testing.T) {
	isolateCacheDir(t)

	markBrowserReady()

	// The marker file must exist and be non-empty (carries a timestamp).
	data, err := os.ReadFile(browserSetupMarkerPath())
	require.NoError(t, err)
	assert.NotEmpty(t, data, "marker file should contain a timestamp")

	// Idempotent: a second call must not error or wipe the marker.
	markBrowserReady()
	assert.False(t, BrowserSetupPending())
}
