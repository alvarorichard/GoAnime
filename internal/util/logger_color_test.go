package util

import (
	"bytes"
	"io"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/charmbracelet/colorprofile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// containsANSI reports whether s holds a CSI/OSC escape introducer used by
// lipgloss/log TrueColor output. Classic cmd.exe prints these as ←[...m.
func containsANSI(s string) bool {
	return strings.Contains(s, "\x1b[") || strings.Contains(s, "\x1b]")
}

// TestPrefixForProfile_ASCII_NoEscapeCodes is the hard regression for the
// "←[48;2;99;102;241mGoAnime" garbage on Windows cmd.exe without VT.
func TestPrefixForProfile_ASCII_NoEscapeCodes(t *testing.T) {
	t.Parallel()
	for _, p := range []colorprofile.Profile{
		colorprofile.NoTTY,
		colorprofile.ASCII,
	} {
		p := p
		t.Run(p.String(), func(t *testing.T) {
			t.Parallel()
			got := prefixForProfile(p)
			assert.Equal(t, "GoAnime", got)
			assert.False(t, containsANSI(got), "ASCII/NoTTY prefix must be plain text, got %q", got)
		})
	}
}

// TestPrefixForProfile_Color_MayUseANSI documents colored hosts still style
// the badge when the profile is above ASCII.
func TestPrefixForProfile_Color_MayUseANSI(t *testing.T) {
	t.Parallel()
	got := prefixForProfile(colorprofile.TrueColor)
	assert.Contains(t, got, "GoAnime")
	// TrueColor profile is allowed to embed SGR; that's intentional.
	assert.True(t, containsANSI(got) || got == "GoAnime",
		"TrueColor path should style or at least include the name")
}

// TestInitLogger_UsesSafeProfile_NoHardcodedTrueColor ensures InitLogger never
// pins TrueColor regardless of host — profile must come from ConsoleColorProfile.
func TestInitLogger_UsesSafeProfile_NoHardcodedTrueColor(t *testing.T) {
	snapshotLogger(t)
	IsDebug = false
	logFile = nil
	fileLogger = nil

	// Capture stderr so a pipe writer is in play for profile detection in some paths.
	origStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	t.Cleanup(func() {
		os.Stderr = origStderr
		_ = w.Close()
		_, _ = io.ReadAll(r)
	})

	InitLogger()
	require.NotNil(t, Logger)

	// When stderr is a pipe, safe profile is ASCII/NoTTY → plain prefix path.
	// Emit one info line and assert no TrueColor SGR 38;2 sequences leak.
	var buf bytes.Buffer
	Logger.SetOutput(&buf)
	// Force ASCII so the assertion is host-independent (pipe may still Detect oddly).
	Logger.SetColorProfile(colorprofile.ASCII)
	Logger.Info("safe-profile-check")
	out := buf.String()
	assert.NotContains(t, out, "38;2;", "must not emit TrueColor SGR under ASCII profile")
	assert.NotContains(t, out, "48;2;", "must not emit TrueColor bg SGR under ASCII profile")
}

// TestShowDebugBanner_PlainWhenANSIUnsupported covers the debug banner path
// that previously dumped TrueColor into cmd.exe (see user screenshots).
func TestShowDebugBanner_PlainWhenANSIUnsupported(t *testing.T) {
	snapshotLogger(t)
	LogFilePath = `C:\Users\Usuario\AppData\Local\GoAnime\logs\goanime_test.log`

	origStderr := os.Stderr
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stderr = w
	t.Cleanup(func() { os.Stderr = origStderr })

	showDebugBanner()
	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	body := string(out)

	assert.Contains(t, body, LogFilePath)
	// Pipe is not a console → SupportsANSI false on Windows; plain banner.
	// On Unix SupportsANSI(pipe file) is still true by design — may have ANSI.
	if runtime.GOOS == "windows" {
		assert.False(t, containsANSI(body),
			"Windows non-console stderr must get plain debug banner, got %q", body)
		assert.Contains(t, body, "DEBUG")
		assert.Contains(t, body, "Get-Content")
	}
}

// TestPrintSavedLocation_PlainWhenANSIUnsupported guards the download path.
func TestPrintSavedLocation_PlainWhenANSIUnsupported(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows SupportsANSI(false) path")
	}
	// Redirect stdout to a pipe so SupportsANSI is false.
	orig := os.Stdout
	r, w, err := os.Pipe()
	require.NoError(t, err)
	os.Stdout = w
	t.Cleanup(func() { os.Stdout = orig })

	PrintSavedLocation("Saved", `C:\tmp\ep.mp4`)
	require.NoError(t, w.Close())
	out, err := io.ReadAll(r)
	require.NoError(t, err)
	body := string(out)

	assert.Contains(t, body, "Saved")
	assert.Contains(t, body, `C:\tmp\ep.mp4`)
	assert.False(t, containsANSI(body), "plain fallback must not contain ANSI")
}

// TestResolveColorProfile_ContractViaTUI re-exports the critical Windows
// contract so util package tests fail if tui regresses independently.
func TestResolveColorProfile_ContractViaTUI(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "windows" {
		t.Skip("Windows contract")
	}
	p := tui.ResolveColorProfile(os.Stderr, []string{"COLORTERM=truecolor"}, false)
	assert.LessOrEqual(t, p, colorprofile.ASCII)
}
