package tui

import (
	"bytes"
	"os"
	"runtime"
	"strings"
	"testing"

	"github.com/charmbracelet/colorprofile"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestResolveColorProfile_WindowsWithoutVT_ForcesASCII is the regression guard
// for classic cmd.exe garbage (←[38;2;...m). Detect alone returns TrueColor on
// Win10+ by OS build; without VT we MUST force ASCII so nothing emits ANSI.
func TestResolveColorProfile_WindowsWithoutVT_ForcesASCII(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only regression: VT-disabled console")
	}

	// Simulate a TTY-ish env that Detect upgrades to TrueColor on Win10+.
	env := []string{
		"TERM=",
		"COLORTERM=truecolor",
	}
	// Use os.Stderr so Detect sees a real file handle; VT flag forced false.
	got := ResolveColorProfile(os.Stderr, env, false)
	assert.LessOrEqual(t, got, colorprofile.ASCII,
		"Windows console without VT must not emit color ANSI (got %v)", got)
}

// TestResolveColorProfile_WindowsWithVT_AllowsColor ensures we do not
// over-downgrade modern hosts (Windows Terminal, cmd with VT on).
// TTY_FORCE makes Detect treat the writer as a TTY even under go test pipes.
func TestResolveColorProfile_WindowsWithVT_AllowsColor(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only")
	}

	env := []string{
		"TTY_FORCE=1",
		"WT_SESSION=test-session",
		"COLORTERM=truecolor",
	}
	got := ResolveColorProfile(os.Stderr, env, true)
	assert.Greater(t, got, colorprofile.ASCII,
		"VT-enabled Windows console should keep color (got %v)", got)

	// Same env without VT must still force ASCII (the cmd.exe regression).
	downgraded := ResolveColorProfile(os.Stderr, env, false)
	assert.LessOrEqual(t, downgraded, colorprofile.ASCII)
}

// TestResolveColorProfile_NoColorEnv_StaysPlain covers NO_COLOR on any OS.
func TestResolveColorProfile_NoColorEnv_StaysPlain(t *testing.T) {
	t.Parallel()
	env := []string{"NO_COLOR=1", "TERM=xterm-256color", "COLORTERM=truecolor"}
	got := ResolveColorProfile(os.Stderr, env, true)
	assert.LessOrEqual(t, got, colorprofile.ASCII)
}

// TestResolveColorProfile_NonWindows_IgnoresVTFlag documents that Unix
// hosts always trust Detect (vtEnabled is irrelevant).
func TestResolveColorProfile_NonWindows_IgnoresVTFlag(t *testing.T) {
	t.Parallel()
	if runtime.GOOS == "windows" {
		t.Skip("Unix-only branch")
	}
	env := []string{"TERM=xterm-256color", "COLORTERM=truecolor"}
	// Even with vtEnabled=false, non-Windows must not force ASCII solely for that.
	with := ResolveColorProfile(os.Stderr, env, true)
	without := ResolveColorProfile(os.Stderr, env, false)
	assert.Equal(t, with, without)
}

// TestResolveColorProfile_PipeWriter_NoTTY ensures redirected writers stay plain.
func TestResolveColorProfile_PipeWriter_NoTTY(t *testing.T) {
	t.Parallel()
	var buf bytes.Buffer
	got := ResolveColorProfile(&buf, []string{"TERM=xterm-256color"}, false)
	assert.LessOrEqual(t, got, colorprofile.ASCII,
		"non-TTY writer must not use color profiles (got %v)", got)
}

func TestEnableVirtualTerminal_DoesNotPanic(t *testing.T) {
	t.Parallel()
	_ = EnableVirtualTerminal()
	_ = EnableVirtualTerminal()
}

func TestHasVirtualTerminal_NilFile(t *testing.T) {
	t.Parallel()
	assert.False(t, HasVirtualTerminal(nil))
}

func TestSupportsANSI_NilFile(t *testing.T) {
	t.Parallel()
	assert.False(t, SupportsANSI(nil))
}

// TestEnableVirtualTerminal_EnablesFlagWhenConsole is the live console check.
// Skips when stdout is not a console (CI pipes, go test capture).
func TestEnableVirtualTerminal_EnablesFlagWhenConsole(t *testing.T) {
	if runtime.GOOS != "windows" {
		t.Skip("Windows console mode flag only")
	}
	ok := EnableVirtualTerminal()
	if !HasVirtualTerminal(os.Stdout) && !HasVirtualTerminal(os.Stderr) {
		t.Skip("no console attached (piped CI) — cannot assert VT flag")
	}
	require.True(t, ok, "EnableVirtualTerminal should succeed on a real console")
	assert.True(t, HasVirtualTerminal(os.Stdout) || HasVirtualTerminal(os.Stderr))
}

// TestConsoleColorProfile_NeverTrueColorWithoutVT hard-guards the user-facing
// bug: profile used by logger/TUI must never be TrueColor/ANSI256 when VT is off.
func TestConsoleColorProfile_NeverTrueColorWithoutVT(t *testing.T) {
	t.Parallel()
	if runtime.GOOS != "windows" {
		t.Skip("Windows-only contract")
	}
	// Force the pure path with vtEnabled=false regardless of host state.
	p := ResolveColorProfile(os.Stderr, os.Environ(), false)
	assert.LessOrEqual(t, p, colorprofile.ASCII)
}

// TestRestoreTerminalState_NoANSIWhenUnsupported ensures exit cleanup does not
// dump TerminalResetSequence into classic cmd.exe.
func TestRestoreTerminalState_NoANSIWhenUnsupported(t *testing.T) {
	t.Parallel()
	// bytes.Buffer is not *os.File → SupportsANSI path not taken; sequence writes.
	// Test the *os.File branch with a temp file (not a console → no VT).
	f, err := os.CreateTemp(t.TempDir(), "restore-*.txt")
	require.NoError(t, err)
	t.Cleanup(func() { _ = f.Close() })

	if runtime.GOOS == "windows" {
		// Temp file is not a console → SupportsANSI false → no write.
		RestoreTerminalState(f)
		_, _ = f.Seek(0, 0)
		data, err := os.ReadFile(f.Name())
		require.NoError(t, err)
		assert.Empty(t, data, "must not write ANSI reset to non-console Windows handle")
		assert.NotContains(t, string(data), "\x1b")
		return
	}
	// Non-Windows: sequence is written (SupportsANSI true for any non-nil file).
	RestoreTerminalState(f)
	_, _ = f.Seek(0, 0)
	data, err := os.ReadFile(f.Name())
	require.NoError(t, err)
	assert.Contains(t, string(data), "\x1b[?25h")
}

// TestTerminalResetSequence_ContainsNoRIS documents we never hard-reset the
// screen (would wipe scrollback) — keeps restore safe.
func TestTerminalResetSequence_ContainsNoRIS(t *testing.T) {
	t.Parallel()
	assert.NotContains(t, TerminalResetSequence, "\x1bc")
	assert.NotContains(t, TerminalResetSequence, "\x1b[2J")
	assert.True(t, strings.Contains(TerminalResetSequence, "\x1b[?25h"))
}
