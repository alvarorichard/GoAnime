package testenv

import (
	"os"
	"testing"

	"golang.org/x/term"
)

// RequireLiveNetwork skips the test unless live network access was explicitly enabled.
func RequireLiveNetwork(t *testing.T) {
	t.Helper()

	if os.Getenv("GOANIME_LIVE_TESTS") != "1" {
		t.Skip("Skipping live network test; set GOANIME_LIVE_TESTS=1 to enable it")
	}
}

// RequireInteractiveTerminal skips the test unless interactive terminal access was explicitly enabled.
func RequireInteractiveTerminal(t *testing.T) {
	t.Helper()

	if os.Getenv("GOANIME_INTERACTIVE_TESTS") != "1" {
		t.Skip("Skipping interactive test; set GOANIME_INTERACTIVE_TESTS=1 to enable it")
	}

	if !term.IsTerminal(int(os.Stdin.Fd())) || !term.IsTerminal(int(os.Stdout.Fd())) {
		t.Skip("Skipping interactive test; no interactive terminal is available")
	}
}
