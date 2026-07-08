package superflix

import (
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// isolateCacheDir points os.UserCacheDir at a temp dir for the test. It sets
// XDG_CACHE_HOME (Linux/BSD), HOME (macOS) and LocalAppData (Windows) so the
// marker logic never touches the real user cache. t.Setenv forbids t.Parallel,
// so callers run serially.
func isolateCacheDir(t *testing.T) {
	t.Helper()
	dir := t.TempDir()
	t.Setenv("XDG_CACHE_HOME", dir)
	t.Setenv("HOME", dir)
	t.Setenv("LocalAppData", dir)
}

// stubBrowserProbes replaces the browser-availability indirection points for
// the duration of the test so BrowserSetupPending is deterministic regardless
// of what is installed on the machine.
func stubBrowserProbes(t *testing.T, chrome, chromium bool) {
	t.Helper()
	origChrome, origChromium := systemChromeAvailableFn, bundledChromiumInstalledFn
	systemChromeAvailableFn = func() bool { return chrome }
	bundledChromiumInstalledFn = func() bool { return chromium }
	t.Cleanup(func() {
		systemChromeAvailableFn = origChrome
		bundledChromiumInstalledFn = origChromium
	})
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
	tests := []struct {
		name     string
		marked   bool
		chrome   bool
		chromium bool
		want     bool
	}{
		{"fresh cache is pending", false, true, true, true},
		{"marked with chrome present", true, true, false, false},
		{"marked with chromium present", true, false, true, false},
		{"marked but no browser left", true, false, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			isolateCacheDir(t)
			stubBrowserProbes(t, tt.chrome, tt.chromium)
			// Pin the channel to the default ("chrome" with bundled-Chromium
			// fallback) regardless of the host's GOANIME_SF_* env.
			t.Setenv("GOANIME_SF_CHROME_CHANNEL", "")
			t.Setenv("GOANIME_SF_BUNDLED", "")
			if tt.marked {
				markBrowserReady()
			}
			assert.Equal(t, tt.want, BrowserSetupPending())
		})
	}
}

func TestMarkBrowserReady(t *testing.T) {
	isolateCacheDir(t)
	stubBrowserProbes(t, true, true)

	markBrowserReady()

	// The marker file must exist and be non-empty (carries a timestamp).
	data, err := os.ReadFile(browserSetupMarkerPath())
	require.NoError(t, err)
	assert.NotEmpty(t, data, "marker file should contain a timestamp")

	// Idempotent: a second call must not error or wipe the marker.
	markBrowserReady()
	assert.False(t, BrowserSetupPending())
}

func TestSystemChromeAvailable(t *testing.T) {
	switch runtime.GOOS {
	case "windows":
		// Point every probed base at empty temp dirs → not available.
		empty := t.TempDir()
		t.Setenv("ProgramFiles", empty)
		t.Setenv("ProgramFiles(x86)", empty)
		t.Setenv("LOCALAPPDATA", empty)
		assert.False(t, systemChromeAvailable())

		// Create the expected chrome.exe layout under one base → available.
		appDir := filepath.Join(empty, "Google", "Chrome", "Application")
		require.NoError(t, os.MkdirAll(appDir, 0o750))
		require.NoError(t, os.WriteFile(filepath.Join(appDir, "chrome.exe"), []byte("x"), 0o600))
		assert.True(t, systemChromeAvailable())
	case "darwin":
		// /Applications can't be isolated; just ensure the probe runs.
		_ = systemChromeAvailable()
	default:
		// Empty PATH → LookPath finds nothing.
		t.Setenv("PATH", t.TempDir())
		assert.False(t, systemChromeAvailable())
	}
}

func TestBundledChromiumInstalled(t *testing.T) {
	dir := t.TempDir()
	t.Setenv("PLAYWRIGHT_BROWSERS_PATH", dir)

	// Empty browsers dir → nothing downloaded yet.
	assert.False(t, bundledChromiumInstalled())

	// A stray file must not count, only a chromium-* directory.
	require.NoError(t, os.WriteFile(filepath.Join(dir, "chromium-notadir"), []byte("x"), 0o600))
	assert.False(t, bundledChromiumInstalled())

	require.NoError(t, os.MkdirAll(filepath.Join(dir, "chromium-1187"), 0o750))
	assert.True(t, bundledChromiumInstalled())
}
