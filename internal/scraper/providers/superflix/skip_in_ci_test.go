package superflix

import (
	"os"
	"strings"
	"testing"
)

// skipInCI skips tests that hit real upstream hosts when running in CI, where
// those hosts are unreachable or rate-limited and would flake the suite.
func skipInCI(t *testing.T) {
	t.Helper()
	if isCI() {
		t.Skipf("skipped in CI: real upstream call (CI=%q, GITHUB_ACTIONS=%q)",
			os.Getenv("CI"), os.Getenv("GITHUB_ACTIONS"))
	}
}

func isCI() bool {
	if v := strings.ToLower(os.Getenv("GITHUB_ACTIONS")); v == "true" || v == "1" {
		return true
	}
	if v := strings.ToLower(os.Getenv("CI")); v == "true" || v == "1" {
		return true
	}
	return false
}
