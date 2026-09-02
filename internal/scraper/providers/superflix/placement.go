package superflix

import (
	"encoding/json"
	"os"
	"path/filepath"

	"github.com/alvarorichard/Goanime/internal/util"
)

// Chromium remembers where its window was last placed and restores it on the
// next launch from the persistent profile, which OVERRIDES --window-position.
//
// That silently defeated --sf-offscreen: a profile that had ever been on screen
// kept relaunching at its old coordinates no matter what the flag said
// (observed: window_placement {left:0, top:0} winning over
// --window-position=-32000,-32000). It also breaks the other direction — a
// profile last used offscreen would keep the window invisible after the user
// turns the flag OFF, with no way to reach it.
//
// So the saved placement is pinned to match the mode before every launch,
// rather than trusting the command-line flag alone.

// solverPlacementSize is the window size that goes with solverWindowBounds; the
// stored rectangle is left/top/right/bottom, not left/top/width/height.
const (
	solverPlacementWidth  = 1100
	solverPlacementHeight = 800
	solverOnscreenX       = 60
	solverOnscreenY       = 60
)

// pinSolverWindowPlacement rewrites the profile's saved window rectangle so the
// next launch lands where the current mode wants it.
//
// Best-effort and non-fatal: on a fresh profile there is nothing to rewrite
// (--window-position then applies cleanly), and a malformed or unreadable
// Preferences file is left untouched rather than risking the profile that holds
// the Cloudflare clearance.
func pinSolverWindowPlacement(profileDir string, offscreen bool) {
	path := filepath.Join(profileDir, "Default", "Preferences")
	raw, err := os.ReadFile(path) // #nosec G304 -- path is built from our own cache dir
	if err != nil {
		return // fresh profile: the launch flag is enough
	}

	var prefs map[string]any
	if err := json.Unmarshal(raw, &prefs); err != nil {
		util.Debug("SuperFlix: solver profile preferences unreadable; leaving placement alone", "err", err)
		return
	}

	browser, ok := prefs["browser"].(map[string]any)
	if !ok {
		browser = map[string]any{}
		prefs["browser"] = browser
	}
	placement, ok := browser["window_placement"].(map[string]any)
	if !ok {
		placement = map[string]any{}
		browser["window_placement"] = placement
	}

	x, y := solverOnscreenX, solverOnscreenY
	if offscreen {
		x, y = solverOffscreenX, solverOffscreenX
	}
	placement["left"] = x
	placement["top"] = y
	placement["right"] = x + solverPlacementWidth
	placement["bottom"] = y + solverPlacementHeight
	// A maximized window ignores the rectangle entirely, which would drag an
	// offscreen solve back onto the desktop.
	placement["maximized"] = false

	out, err := json.Marshal(prefs)
	if err != nil {
		return
	}
	if err := os.WriteFile(path, out, 0o600); err != nil {
		util.Debug("SuperFlix: could not pin the solver window placement", "err", err)
		return
	}
	util.Debug("SuperFlix: pinned solver window placement", "offscreen", offscreen, "left", x, "top", y)
}
