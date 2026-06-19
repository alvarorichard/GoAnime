package scraper

import (
	"os"
	"strings"
	"testing"
)

// skipInCI skips tests that hit real upstream services (sflix.to, 9animetv.to,
// superflix.to, etc.) when running in CI environments. The surf/utls HTTP/2
// client does not always honor context cancellation when a CDN accepts the
// TLS handshake but stalls on the response body — local timeouts work, but CI
// runners experience indefinite hangs that exhaust the workflow budget.
//
// We detect CI via the standard GITHUB_ACTIONS and CI environment variables.
// Local developers run the tests normally; the contract is "best-effort
// against real CDNs locally, deterministic skip in CI".
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
