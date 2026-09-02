package superflix

import (
	"os"
	"runtime"
	"strings"
)

// superflixConfig captures the SuperFlix Cloudflare-bypass browser knobs,
// resolved once from the environment. Centralizing them (instead of scattered
// os.Getenv calls across the solver) keeps every switch documented in one place
// and makes the launch decisions unit-testable.
//
// The CLI flags --sf-headless / --sf-bundled / --sf-browser / --sf-mask set the
// corresponding env vars (see util.FlagParser), so reading the env here covers
// both flag-supplied and manually-exported input.
type superflixConfig struct {
	// Headless runs the bypass browser without a visible window
	// (GOANIME_SF_HEADLESS). Advanced/escape-hatch only: Cloudflare Turnstile
	// typically rejects headless, so the gate will not auto-pass.
	Headless bool
	// ForceBundled forces Playwright's bundled Chromium instead of the user's
	// system Chrome (GOANIME_SF_BUNDLED).
	ForceBundled bool
	// Channel pins an explicit browser distribution channel such as "chrome",
	// "chrome-beta" or "msedge" (GOANIME_SF_CHROME_CHANNEL). Empty = auto.
	Channel string
	// Mask injects the navigator fingerprint mask script (GOANIME_SF_MASK).
	// Off by default because it usually BREAKS the managed challenge.
	Mask bool
	// Offscreen keeps the bypass browser out of the user's way, minimized
	// instead of opening in their face.
	//
	// ON BY DEFAULT. Set GOANIME_SF_OFFSCREEN=0 (or pass --sf-window) to get the
	// old behaviour of a window that always shows. Hiding is only safe as a
	// default because it is not absolute: the window surfaces on its own the
	// moment the challenge needs a human, so nobody is left waiting on a captcha
	// they cannot see.
	//
	// It is "smart" rather than absolute, in two senses. The window is pulled
	// back on screen (see revealSolverWindow) the moment the challenge shows it
	// needs a human — either it reports a load failure or it simply refuses to
	// clear — because a window that could never surface would turn a solvable
	// challenge into a silent failure; Cloudflare does not always auto-pass.
	// And the hiding itself is partial: the window stays minimized through the
	// launch and warm-up but the SuperFlix embed raises it when it loads. See
	// reveal.go for the measurements behind both.
	//
	// This is NOT headless. Cloudflare Turnstile rejects headless outright —
	// measured 2026-09-01, a headless run sits on the "Verificação" page
	// indefinitely (>150s) with both the bundled Chromium and the new headless
	// mode, while a headed run clears it in under 5s. What Turnstile checks is
	// that a real browser is rendering, not that a human can see it: the same
	// run with the window parked at -32000,-32000 clears the gate just as fast.
	Offscreen bool
}

// envFlagDefaultOn reads a boolean env var that defaults to ON, so only an
// explicit "0"/"false"/"no"/"off" turns it off. An unset or unrecognized value
// keeps the default rather than silently disabling a feature over a typo.
func envFlagDefaultOn(name string) bool {
	switch strings.ToLower(strings.TrimSpace(os.Getenv(name))) {
	case "0", "false", "no", "off":
		return false
	default:
		return true
	}
}

// loadSuperflixConfig reads the bypass-browser knobs from the environment.
func loadSuperflixConfig() superflixConfig {
	return superflixConfig{
		Headless:     os.Getenv("GOANIME_SF_HEADLESS") != "",
		ForceBundled: os.Getenv("GOANIME_SF_BUNDLED") != "",
		Channel:      os.Getenv("GOANIME_SF_CHROME_CHANNEL"),
		Mask:         os.Getenv("GOANIME_SF_MASK") != "",
		Offscreen:    envFlagDefaultOn("GOANIME_SF_OFFSCREEN"),
	}
}

// resolveChannel returns the browser channel to launch FIRST: an explicit
// override wins; otherwise system Chrome ("chrome") unless bundled Chromium is
// forced (empty string == bundled Chromium).
func (c superflixConfig) resolveChannel() string {
	if c.Channel != "" {
		return c.Channel
	}
	if c.ForceBundled {
		return ""
	}
	return "chrome"
}

// displayAvailable reports whether a graphical display is available for the
// headed bypass browser. Cloudflare Turnstile only auto-passes in a real,
// on-screen window, so on a pure headless host (an SSH session, a container, a
// box with no compositor) the solve stalls until timeout.
//
// Only X11/Wayland env vars gate the GUI on Linux/BSD; macOS and Windows always
// have a window server, so a display is assumed present there.
func displayAvailable() bool {
	switch runtime.GOOS {
	case "windows", "darwin":
		return true
	default:
		return os.Getenv("DISPLAY") != "" || os.Getenv("WAYLAND_DISPLAY") != ""
	}
}

// HeadlessEnvironment reports whether GoAnime is running without a usable
// graphical display while the user has NOT opted into headless mode. In that
// state the bypass browser cannot clear Turnstile, so callers should warn the
// user up front (e.g. suggest running on a desktop session or passing
// --sf-headless) instead of letting them wait out a doomed solve.
func HeadlessEnvironment() bool {
	cfg := loadSuperflixConfig()
	if cfg.Headless {
		// The user explicitly chose headless; their accepted trade-off, not an
		// unexpected state worth warning about.
		return false
	}
	return !displayAvailable()
}
