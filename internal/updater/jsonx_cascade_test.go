package updater

import (
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// serveJSON is a plain loopback server: checkForUpdatesFromURL builds its own
// http.Client, so the in-memory test server cannot be used here.
func serveJSON(t *testing.T, handler http.HandlerFunc) string {
	t.Helper()
	srv := httptest.NewServer(handler)
	t.Cleanup(srv.Close)
	return srv.URL
}

// TestUpdateCheckCascade_DecodesRelease is the happy path through the release
// check after json.NewDecoder was replaced with the bounded jsonx.Decode.
func TestUpdateCheckCascade_DecodesRelease(t *testing.T) {
	url := serveJSON(t, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{
			"tag_name":"v99.0.0",
			"name":"GoAnime 99",
			"body":"notes",
			"assets":[{"name":"goanime_linux_amd64.tar.gz","browser_download_url":"https://example.com/a.tar.gz"}]
		}`))
	})

	release, isNewer, err := checkForUpdatesFromURL(url, "1.0.0")
	if err != nil {
		t.Fatalf("checkForUpdatesFromURL: %v", err)
	}
	if !isNewer {
		t.Error("v99.0.0 should be newer than 1.0.0")
	}
	if release.TagName != "v99.0.0" || len(release.Assets) != 1 {
		t.Fatalf("unexpected release: %+v", release)
	}
	if release.Assets[0].BrowserDownloadURL != "https://example.com/a.tar.gz" {
		t.Errorf("asset URL not decoded: %+v", release.Assets[0])
	}
}

// TestUpdateCheckCascade_CaseInsensitiveFields proves the v1 field-matching
// semantics reach this call site too.
func TestUpdateCheckCascade_CaseInsensitiveFields(t *testing.T) {
	url := serveJSON(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"TAG_NAME":"v99.0.0","Name":"GoAnime 99","ASSETS":[]}`))
	})

	release, _, err := checkForUpdatesFromURL(url, "1.0.0")
	if err != nil {
		t.Fatalf("checkForUpdatesFromURL: %v", err)
	}
	if release.TagName != "v99.0.0" {
		t.Fatalf("case-insensitive matching lost: %+v", release)
	}
}

// TestUpdateCheckCascade_OversizedResponseIsRejected covers the security half of
// the change: json.NewDecoder(resp.Body) read without any bound, so a hostile
// update endpoint could stream until the process died. The decode now stops at
// maxJSONResponseBytes.
func TestUpdateCheckCascade_OversizedResponseIsRejected(t *testing.T) {
	url := serveJSON(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v99.0.0","body":"`))
		chunk := strings.Repeat("x", 256*1024)
		for range (maxJSONResponseBytes / len(chunk)) + 8 {
			if _, err := w.Write([]byte(chunk)); err != nil {
				return
			}
		}
		_, _ = w.Write([]byte(`"}`))
	})

	_, _, err := checkForUpdatesFromURL(url, "1.0.0")
	if err == nil {
		t.Fatal("expected an oversized response to be rejected")
	}
	if !errors.Is(err, jsonx.ErrTooLarge) {
		t.Fatalf("expected jsonx.ErrTooLarge, got %v", err)
	}
}

// TestUpdateCheckCascade_TrailingDataIsRejected documents a deliberate
// behaviour change: json.NewDecoder decoded the first JSON value and ignored
// whatever followed it, so a response that was only *prefixed* with valid JSON
// was accepted. jsonx.Decode requires the body to be exactly one JSON value.
func TestUpdateCheckCascade_TrailingDataIsRejected(t *testing.T) {
	url := serveJSON(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"tag_name":"v99.0.0"}<html>injected</html>`))
	})

	_, _, err := checkForUpdatesFromURL(url, "1.0.0")
	if err == nil {
		t.Fatal("expected trailing data after the JSON value to be rejected")
	}
	if !strings.Contains(fmt.Sprint(err), "decode release data") {
		t.Errorf("error should come from the decode step, got %v", err)
	}
}

// TestUpdateCheckCascade_TrailingWhitespaceIsAccepted keeps the strictness above
// from becoming a false positive on well-formed responses.
func TestUpdateCheckCascade_TrailingWhitespaceIsAccepted(t *testing.T) {
	url := serveJSON(t, func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte("{\"tag_name\":\"v99.0.0\"}\n\n  \t\n"))
	})

	release, _, err := checkForUpdatesFromURL(url, "1.0.0")
	if err != nil {
		t.Fatalf("trailing whitespace must be tolerated: %v", err)
	}
	if release.TagName != "v99.0.0" {
		t.Fatalf("unexpected release: %+v", release)
	}
}
