//go:build windows

package tui

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"golang.org/x/sys/windows/registry"
)

// These run on a real Windows machine, and that is the point: the guarantee
// they make is that installing GoAnime Glyphs lands both files in the fonts
// folder, gdi answers for the glyphs of the very font just loaded, and a
// second install changes nothing. The folder is the test's own scratch dir
// (LOCALAPPDATA pointed there), and register stays false so the machine's
// HKCU font registry is never written by a test.

// TestFontBindingsAnswer is the guard the old hand-rolled proc table needed:
// the typed Windows bindings behind the probes must resolve and answer on
// this machine, not merely compile. It makes no claim that a console font
// exists here: under ConPTY consoleFontFace answers "", which is the
// documented lost answer, while a classic conhost names the face it draws
// with and gdi maps the probe runes through it.
func TestFontBindingsAnswer(t *testing.T) {
	faces, ok := enumFontFaces()
	require.True(t, ok, "gdi enumerates the faces installed on this machine")
	require.NotEmpty(t, faces, "a Windows machine has fonts")

	face := consoleFontFace()
	if face == "" {
		// ConPTY, a pipe, or a refused API: the console probe must fail
		// closed rather than let a missing answer count as coverage.
		assert.False(t, probeConsoleFontCovers(glossProbeRunes),
			"an unknown console face never covers the gloss runes")
		t.Log("consoleFontFace has no face here (ConPTY or no console); the lost answer is the documented one")
		return
	}

	dc, done := screenDC()
	require.NotZero(t, dc, "the screen context backs the console font probe")
	defer done()
	drawn, ok := gdiFaceGlyphs(dc, face, glossProbeRunes)
	require.True(t, ok, "gdi maps glyphs for the console's face %q", face)
	assert.Len(t, drawn, len(glossProbeRunes), "every console probe rune gets an answer")
}

// TestGdiFaceGlyphsMapsLongRequests pins the mapping buffers to the request:
// the old fixed 32-rune window silently stopped at 32, so a long probe needs
// a whole answer, not a prefix.
func TestGdiFaceGlyphsMapsLongRequests(t *testing.T) {
	faces, ok := enumFontFaces()
	require.True(t, ok, "gdi enumerates the faces installed on this machine")
	require.NotEmpty(t, faces, "a Windows machine has fonts")
	dc, done := screenDC()
	require.NotZero(t, dc, "gdi's screen context is available")
	defer done()

	// 52 BMP runes, well past the old window. Some faces (symbol and device
	// fonts) cannot be selected or mapped; the first one that answers must
	// answer for every rune.
	runes := []rune("ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz")
	for _, face := range faces {
		drawn, mapped := gdiFaceGlyphs(dc, face, runes)
		if !mapped {
			continue
		}
		assert.Len(t, drawn, len(runes), "gdi answers for every rune of %q", face)
		return
	}
	t.Fatal("no installed face answered a 52-rune glyph-mapping request")
}

func TestInstallGlyphFontLandsAndAnswers(t *testing.T) {
	scratch := t.TempDir()
	t.Setenv("LOCALAPPDATA", scratch)

	dir, err := userFontDir()
	require.NoError(t, err, "a scratch LOCALAPPDATA still yields a fonts folder")
	note := installGlyphFontInto(dir, false)

	require.Contains(t, note, "installed GoAnime Glyphs in", "the note owns the install: %q", note)
	require.Contains(t, note, "maps every Japanese glyph", "the install is machine-verified, not promised: %q", note)
	require.Contains(t, note, dir, "the note names where the font landed: %q", note)

	info, err := os.Stat(filepath.Join(dir, glyphFontFile))
	require.NoError(t, err, "the font file lands in the fonts folder")
	assert.Equal(t, int64(len(glyphFont)), info.Size(), "the font is all of its bytes")

	license, err := os.ReadFile(filepath.Join(dir, "OFL.txt"))
	require.NoError(t, err, "the OFL license travels with the font")
	assert.Contains(t, string(license), "SIL Open Font License", "the license is the real one")

	// The machine-verified part, asked again by the test itself: the face
	// this very session loaded draws every Japanese probe glyph.
	drawn, ok := gdiFaceGlyphsOnScreen(glyphFontFamily, glossProbeRunes)
	require.True(t, ok, "gdi answers for the glyphs font")
	assert.True(t, every(drawn), "the freshly loaded font draws every probe glyph: %v", drawn)
}

func TestInstallGlyphFontTwiceChangesNothing(t *testing.T) {
	scratch := t.TempDir()
	t.Setenv("LOCALAPPDATA", scratch)

	dir, err := userFontDir()
	require.NoError(t, err)
	first := installGlyphFontInto(dir, false)
	require.Contains(t, first, "installed GoAnime Glyphs in")

	second := installGlyphFontInto(dir, false)
	assert.NotEmpty(t, second, "the second install still speaks for itself")
	assert.Contains(t, second, "maps every Japanese glyph", "the second load is verified the same way: %q", second)

	info, err := os.Stat(filepath.Join(dir, glyphFontFile))
	require.NoError(t, err, "the font survives the second install")
	assert.Equal(t, int64(len(glyphFont)), info.Size(), "no rewrite half-did the font")
}

func TestInstallGlyphFontUnregisteredLeavesTheRegistryAlone(t *testing.T) {
	scratch := t.TempDir()
	t.Setenv("LOCALAPPDATA", scratch)

	dir, err := userFontDir()
	require.NoError(t, err)
	_ = installGlyphFontInto(dir, false)

	key, err := registry.OpenKey(registry.CURRENT_USER,
		`Software\Microsoft\Windows NT\CurrentVersion\Fonts`, registry.QUERY_VALUE)
	require.NoError(t, err, "the per-user Fonts key is readable")
	defer func() { _ = key.Close() }()
	value, _, err := key.GetStringValue(glyphFontFamily + " (TrueType)")
	if err != nil {
		t.Log("no GoAnime Glyphs value in the font registry — the machine never had a real install; nothing to check")
		return
	}
	assert.NotContains(t, value, scratch, "a test's scratch path never reaches the font registry")
}

func TestUserFontDirShape(t *testing.T) {
	local := t.TempDir()
	t.Setenv("LOCALAPPDATA", local)
	dir, err := userFontDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(local, "Microsoft", "Windows", "Fonts"), dir,
		"the per-user fonts folder is Windows' own path")

	home, err := os.UserHomeDir()
	require.NoError(t, err)

	// No LOCALAPPDATA, or a relative one: neither is trusted — the user's
	// own AppData\Local carries the folder instead.
	t.Setenv("LOCALAPPDATA", "")
	dir, err = userFontDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "AppData", "Local", "Microsoft", "Windows", "Fonts"), dir)

	t.Setenv("LOCALAPPDATA", filepath.Join("relative", "path"))
	dir, err = userFontDir()
	require.NoError(t, err)
	assert.Equal(t, filepath.Join(home, "AppData", "Local", "Microsoft", "Windows", "Fonts"), dir)
	assert.False(t, strings.Contains(dir, "relative"), "a relative LOCALAPPDATA never names the folder")
}
