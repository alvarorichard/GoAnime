package tui

import (
	"errors"
	"os"
	"strings"
	"testing"
)

func TestSafeBubbleTeaEnvironmentSuppressesCapabilityQueries(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	t.Setenv("TERM_PROGRAM", "Ghostty")
	t.Setenv("WT_SESSION", "1")

	env := safeBubbleTeaEnvironment()

	assertEnvValue(t, env, "TERM", "xterm-256color")
	assertEnvValue(t, env, "TERM_PROGRAM", "Apple_Terminal")
	assertEnvMissing(t, env, "WT_SESSION")
}

func TestRunCleanRestoresEnvironmentAndPropagatesError(t *testing.T) {
	t.Setenv("TERM", "xterm-ghostty")
	t.Setenv("TERM_PROGRAM", "Ghostty")
	t.Setenv("WT_SESSION", "1")

	wantErr := errors.New("boom")
	err := RunClean(func() error {
		if got := os.Getenv("TERM"); got != "xterm-256color" {
			t.Fatalf("TERM during RunClean = %q, want xterm-256color", got)
		}
		if got := os.Getenv("TERM_PROGRAM"); got != "Apple_Terminal" {
			t.Fatalf("TERM_PROGRAM during RunClean = %q, want Apple_Terminal", got)
		}
		if _, ok := os.LookupEnv("WT_SESSION"); ok {
			t.Fatalf("WT_SESSION should be unset during RunClean")
		}
		return wantErr
	})

	if !errors.Is(err, wantErr) {
		t.Fatalf("RunClean error = %v, want %v", err, wantErr)
	}
	if got := os.Getenv("TERM"); got != "xterm-ghostty" {
		t.Fatalf("TERM after RunClean = %q, want xterm-ghostty", got)
	}
	if got := os.Getenv("TERM_PROGRAM"); got != "Ghostty" {
		t.Fatalf("TERM_PROGRAM after RunClean = %q, want Ghostty", got)
	}
	if got := os.Getenv("WT_SESSION"); got != "1" {
		t.Fatalf("WT_SESSION after RunClean = %q, want 1", got)
	}
}

func assertEnvValue(t *testing.T, env []string, key, want string) {
	t.Helper()
	prefix := key + "="
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			if got := strings.TrimPrefix(item, prefix); got != want {
				t.Fatalf("%s = %q, want %q", key, got, want)
			}
			return
		}
	}
	t.Fatalf("%s missing from environment", key)
}

func assertEnvMissing(t *testing.T, env []string, key string) {
	t.Helper()
	prefix := key + "="
	for _, item := range env {
		if strings.HasPrefix(item, prefix) {
			t.Fatalf("%s should be absent, got %q", key, item)
		}
	}
}
