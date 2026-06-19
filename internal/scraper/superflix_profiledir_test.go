package scraper

import (
	"path/filepath"
	"strings"
	"testing"
)

func TestSanitizeProfileSegment(t *testing.T) {
	t.Parallel()
	cases := map[string]string{
		"chrome":         "chrome",
		"chrome-beta":    "chrome-beta",
		"msedge_dev":     "msedge_dev",
		"../../etc":      "etc",     // path separators + dots stripped
		"..":             "unknown", // nothing safe left
		"/":              "unknown",
		"a/b/c":          "abc",
		"foo bar":        "foobar", // spaces stripped
		"x..y":           "xy",
		"":               "unknown",
		"日本語":            "unknown", // non-ASCII dropped
		"Chrome-Canary9": "Chrome-Canary9",
	}
	for in, want := range cases {
		if got := sanitizeProfileSegment(in); got != want {
			t.Errorf("sanitizeProfileSegment(%q) = %q, want %q", in, got, want)
		}
	}
}

// TestSolverProfileDirNoTraversal is the security regression: a hostile
// GOANIME_SF_CHROME_CHANNEL must never let the profile dir escape the cache root.
func TestSolverProfileDirNoTraversal(t *testing.T) {
	t.Parallel()
	cache := t.TempDir()
	root := filepath.Join(cache, "goanime")

	for _, channel := range []string{
		"../../../../etc/cron.d/x",
		"../../..",
		"/etc/passwd",
		"a/../../b",
	} {
		dir := solverProfileDir(cache, channel)
		rel, err := filepath.Rel(root, dir)
		if err != nil {
			t.Fatalf("channel %q: Rel(%q,%q): %v", channel, root, dir, err)
		}
		if strings.HasPrefix(rel, "..") || filepath.IsAbs(rel) {
			t.Errorf("channel %q produced dir %q which escapes %q (rel=%q)", channel, dir, root, rel)
		}
	}
}

func TestSecureIntn(t *testing.T) {
	t.Parallel()
	if got := secureIntn(0); got != 0 {
		t.Errorf("secureIntn(0) = %d, want 0", got)
	}
	if got := secureIntn(-5); got != 0 {
		t.Errorf("secureIntn(-5) = %d, want 0", got)
	}
	for i := 0; i < 1000; i++ {
		if got := secureIntn(500); got < 0 || got >= 500 {
			t.Fatalf("secureIntn(500) = %d, out of [0,500)", got)
		}
	}
}
