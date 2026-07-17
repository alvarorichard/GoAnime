package superflix

import (
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"time"

	"github.com/mxschmitt/playwright-go"
)

// ErrPlaywrightUnavailable is returned when the Playwright driver or its bundled
// Chromium can't be initialized (first run needs network to download them).

// installPlaywright installs the Playwright driver (node) and, unless skipped, the
// bundled Chromium. The installer's stdout/stderr are discarded and Verbose is off
// so its progress and compatibility notices — notably the "BEWARE: your OS is not
// officially supported by Playwright; downloading fallback build …" line on
// unsupported distros — can't bleed into and corrupt the loading spinner. Real
// failures still propagate through the returned error.
func installPlaywright(skipBrowsers bool) error {
	opts := &playwright.RunOptions{
		SkipInstallBrowsers: skipBrowsers,
		Verbose:             false,
		Stdout:              io.Discard,
		Stderr:              io.Discard,
	}
	if !skipBrowsers {
		opts.Browsers = []string{"chromium"}
	}
	return playwright.Install(opts)
}

// solverProfileDir returns the persistent profile directory for the given channel.
// System Chrome and bundled Chromium get separate dirs so a profile created by one
// is never opened by the other (which can refuse or corrupt it). Empty channel ==
// bundled Chromium, keeping the historical "cf-playwright-profile" path.
func solverProfileDir(cache, channel string) string {
	name := "cf-playwright-profile"
	if channel != "" {
		name = "cf-" + sanitizeProfileSegment(channel) + "-profile"
	}
	dir := filepath.Join(cache, "goanime", name)
	_ = os.MkdirAll(dir, 0o700)
	return dir
}

// sanitizeProfileSegment reduces an arbitrary string (the browser channel, which
// can come from the GOANIME_SF_CHROME_CHANNEL env var) to a single safe path
// segment. Channel feeds a directory name, so without this a value like
// "../../etc" would let filepath.Join escape the cache dir. We keep only
// [A-Za-z0-9-_] and collapse everything else; an empty result falls back to a
// fixed token so the path is always a valid single component.
func sanitizeProfileSegment(s string) string {
	var b strings.Builder
	for _, r := range s {
		switch {
		case r >= 'a' && r <= 'z', r >= 'A' && r <= 'Z', r >= '0' && r <= '9', r == '-', r == '_':
			b.WriteRune(r)
		}
	}
	if b.Len() == 0 {
		return "unknown"
	}
	return b.String()
}

// browserSetupMarkerPath is the on-disk marker proving the Cloudflare-bypass
// browser engine has been initialized successfully at least once on this
// machine. Its absence means the next init may download Playwright's driver and
// (when system Chrome is missing) a bundled Chromium — up to ~150MB — which can
// look frozen behind a spinner. Callers check BrowserSetupPending to warn first.
func browserSetupMarkerPath() string {
	cache, _ := os.UserCacheDir()
	return filepath.Join(cache, "goanime", ".browser-ready")
}

// BrowserSetupPending reports whether the next Cloudflare-bypass browser init
// is likely to be slow: either first-ever use (no marker), or the marker exists
// but no usable browser binary does anymore — e.g. system Chrome was removed
// since the marker was written, forcing a silent multi-minute bundled-Chromium
// download behind the spinner. Callers surface a "preparing browser" notice
// before that otherwise invisible wait.
func BrowserSetupPending() bool {
	if _, err := os.Stat(browserSetupMarkerPath()); err != nil {
		return true
	}
	cfg := loadSuperflixConfig()
	switch cfg.resolveChannel() {
	case "chrome":
		// Chrome preferred, bundled Chromium is the fallback: setup is only
		// pending when NEITHER is present.
		return !systemChromeAvailableFn() && !bundledChromiumInstalledFn()
	case "":
		return !bundledChromiumInstalledFn()
	default:
		// Custom channel via GOANIME_SF_CHROME_CHANNEL: the user manages that
		// install themselves; don't second-guess it.
		return false
	}
}

// Indirection points for the browser-availability probes, overridable in tests
// so BrowserSetupPending can be exercised without depending on what happens to
// be installed on the machine running the tests.
var (
	systemChromeAvailableFn    = systemChromeAvailable
	bundledChromiumInstalledFn = bundledChromiumInstalled
)

// systemChromeAvailable reports whether a system Google Chrome install exists
// where Playwright's "chrome" channel will look for it.
func systemChromeAvailable() bool {
	switch runtime.GOOS {
	case "windows":
		for _, base := range []string{
			os.Getenv("ProgramFiles"),
			os.Getenv("ProgramFiles(x86)"),
			os.Getenv("LOCALAPPDATA"),
		} {
			if base == "" {
				continue
			}
			// #nosec G703 -- base is a trusted Windows env var (ProgramFiles/LOCALAPPDATA); the path suffix is a constant, and os.Stat only probes existence (no open/read/write/exec).
			if _, err := os.Stat(filepath.Join(base, "Google", "Chrome", "Application", "chrome.exe")); err == nil {
				return true
			}
		}
		return false
	case "darwin":
		_, err := os.Stat("/Applications/Google Chrome.app/Contents/MacOS/Google Chrome")
		return err == nil
	default:
		for _, name := range []string{"google-chrome", "google-chrome-stable", "chrome"} {
			if _, err := exec.LookPath(name); err == nil {
				return true
			}
		}
		return false
	}
}

// bundledChromiumInstalled reports whether Playwright's bundled Chromium has
// already been downloaded into its browsers cache (PLAYWRIGHT_BROWSERS_PATH or
// the ms-playwright dir under the user cache), meaning the fallback launch
// won't need a fresh ~150MB download.
func bundledChromiumInstalled() bool {
	dir := os.Getenv("PLAYWRIGHT_BROWSERS_PATH")
	if dir == "" {
		cache, err := os.UserCacheDir()
		if err != nil {
			return false
		}
		dir = filepath.Join(cache, "ms-playwright")
	}
	entries, err := os.ReadDir(dir)
	if err != nil {
		return false
	}
	for _, e := range entries {
		if e.IsDir() && strings.HasPrefix(e.Name(), "chromium-") {
			return true
		}
	}
	return false
}

// markBrowserReady records that the bypass browser engine initialized
// successfully, so BrowserSetupPending stops returning true on later runs.
// Best-effort: a write failure only means the first-run notice may show again.
func markBrowserReady() {
	p := browserSetupMarkerPath()
	// 0o700 dir / 0o600 file to match solverProfileDir's owner-only perms in
	// this same cache tree (and satisfy gosec G301/G306).
	_ = os.MkdirAll(filepath.Dir(p), 0o700)
	_ = os.WriteFile(p, []byte(time.Now().UTC().Format(time.RFC3339)+"\n"), 0o600)
}
