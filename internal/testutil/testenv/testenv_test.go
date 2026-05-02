package testenv

import "testing"

func TestRequireLiveNetwork(t *testing.T) {
	t.Run("enabled", func(t *testing.T) {
		t.Setenv("GOANIME_LIVE_TESTS", "1")
		RequireLiveNetwork(t)
	})

	t.Run("disabled skips", func(t *testing.T) {
		t.Setenv("GOANIME_LIVE_TESTS", "0")
		RequireLiveNetwork(t)
		t.Fatal("expected RequireLiveNetwork to skip when live tests are disabled")
	})
}

func TestRequireInteractiveTerminal(t *testing.T) {
	t.Run("env disabled skips", func(t *testing.T) {
		t.Setenv("GOANIME_INTERACTIVE_TESTS", "0")
		RequireInteractiveTerminal(t)
		t.Fatal("expected RequireInteractiveTerminal to skip when interactive tests are disabled")
	})

	t.Run("env enabled without tty skips", func(t *testing.T) {
		t.Setenv("GOANIME_INTERACTIVE_TESTS", "1")
		RequireInteractiveTerminal(t)
		t.Fatal("expected RequireInteractiveTerminal to skip when no terminal is available")
	})
}
