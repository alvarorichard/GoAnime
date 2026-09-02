package superflix

import (
	"encoding/json"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// writeProfilePrefs lays out a Chromium profile with the given saved window
// rectangle, the way a previous run would have left it.
func writeProfilePrefs(t *testing.T, placement map[string]any) string {
	t.Helper()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "Default"), 0o700))

	prefs := map[string]any{
		// A real Preferences file is full of unrelated state that must survive.
		"profile": map[string]any{"name": "Person 1"},
		"browser": map[string]any{"window_placement": placement},
	}
	raw, err := json.Marshal(prefs)
	require.NoError(t, err)
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Default", "Preferences"), raw, 0o600))
	return dir
}

func readProfilePlacement(t *testing.T, dir string) map[string]any {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join(dir, "Default", "Preferences"))
	require.NoError(t, err)
	var d map[string]any
	require.NoError(t, json.Unmarshal(raw, &d))
	b, _ := d["browser"].(map[string]any)
	require.NotNil(t, b)
	p, _ := b["window_placement"].(map[string]any)
	require.NotNil(t, p)
	return p
}

// TestPinSolverWindowPlacement_Offscreen is the regression for a silent defeat
// of --sf-offscreen: Chromium restores the profile's saved rectangle and that
// OVERRIDES --window-position, so a profile that had ever been on screen kept
// relaunching there no matter what the flag said.
func TestPinSolverWindowPlacement_Offscreen(t *testing.T) {
	t.Parallel()
	dir := writeProfilePrefs(t, map[string]any{
		"left": 0.0, "top": 0.0, "right": 1288.0, "bottom": 851.0, "maximized": true,
	})

	pinSolverWindowPlacement(dir, true)

	got := readProfilePlacement(t, dir)
	assert.EqualValues(t, solverOffscreenX, got["left"])
	assert.EqualValues(t, solverOffscreenX, got["top"])
	assert.EqualValues(t, solverOffscreenX+solverPlacementWidth, got["right"])
	assert.EqualValues(t, solverOffscreenX+solverPlacementHeight, got["bottom"])
	assert.Equal(t, false, got["maximized"],
		"a maximized window ignores the rectangle and would land back on the desktop")
}

// The other direction matters just as much: a profile last used hidden must not
// keep the window invisible after the user asks to see it again.
func TestPinSolverWindowPlacement_BackOnScreen(t *testing.T) {
	t.Parallel()
	dir := writeProfilePrefs(t, map[string]any{
		"left": float64(solverOffscreenX), "top": float64(solverOffscreenX),
		"right": -30900.0, "bottom": -31200.0, "maximized": false,
	})

	pinSolverWindowPlacement(dir, false)

	got := readProfilePlacement(t, dir)
	assert.EqualValues(t, solverOnscreenX, got["left"])
	assert.EqualValues(t, solverOnscreenY, got["top"])
	assert.EqualValues(t, solverOnscreenX+solverPlacementWidth, got["right"])
}

// Unrelated preferences must survive: this file also holds the Cloudflare
// clearance state that makes repeat solves fast.
func TestPinSolverWindowPlacement_PreservesOtherPrefs(t *testing.T) {
	t.Parallel()
	dir := writeProfilePrefs(t, map[string]any{"left": 0.0, "top": 0.0})

	pinSolverWindowPlacement(dir, true)

	raw, err := os.ReadFile(filepath.Join(dir, "Default", "Preferences"))
	require.NoError(t, err)
	var d map[string]any
	require.NoError(t, json.Unmarshal(raw, &d))
	profile, _ := d["profile"].(map[string]any)
	require.NotNil(t, profile, "unrelated preference trees must not be dropped")
	assert.Equal(t, "Person 1", profile["name"])
}

// A fresh profile has no Preferences yet; --window-position applies cleanly and
// there is nothing to rewrite.
func TestPinSolverWindowPlacement_FreshProfileIsNoOp(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	pinSolverWindowPlacement(dir, true) // must not create anything or panic

	_, err := os.Stat(filepath.Join(dir, "Default", "Preferences"))
	assert.True(t, os.IsNotExist(err), "must not fabricate a profile")
}

// A corrupt Preferences file must be left alone rather than overwritten: it
// carries the Cloudflare clearance, and losing it costs a full re-solve.
func TestPinSolverWindowPlacement_CorruptFileIsLeftAlone(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "Default"), 0o700))
	path := filepath.Join(dir, "Default", "Preferences")
	const garbage = "{not json at all"
	require.NoError(t, os.WriteFile(path, []byte(garbage), 0o600))

	pinSolverWindowPlacement(dir, true)

	raw, err := os.ReadFile(path)
	require.NoError(t, err)
	assert.Equal(t, garbage, string(raw), "an unparseable profile must not be clobbered")
}

// A profile with no browser/window_placement tree yet still gets pinned, so the
// very first switch into hidden mode takes effect immediately.
func TestPinSolverWindowPlacement_AddsMissingTree(t *testing.T) {
	t.Parallel()
	dir := t.TempDir()
	require.NoError(t, os.MkdirAll(filepath.Join(dir, "Default"), 0o700))
	require.NoError(t, os.WriteFile(filepath.Join(dir, "Default", "Preferences"),
		[]byte(`{"profile":{"name":"Person 1"}}`), 0o600))

	pinSolverWindowPlacement(dir, true)

	got := readProfilePlacement(t, dir)
	assert.EqualValues(t, solverOffscreenX, got["left"])
}
