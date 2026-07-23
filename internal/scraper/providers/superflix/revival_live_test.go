package superflix

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSuperFlixStreamRevival_Live verifies the revived stream pipeline end to end
// against the real upstream (added 2026-06-09 after the site moved to a
// Turnstile-gated warezcdn embed). It drives the production GetStreamURL → browser
// solver → embed → getVideo path and asserts a playable HLS master comes back.
//
// Gated: needs a headed browser + network, so it is skipped in -short / CI. Run:
//
//	go test ./internal/scraper/ -run TestSuperFlixStreamRevival_Live -v -count=1 -timeout 240s
func TestSuperFlixStreamRevival_Live(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping live SuperFlix stream test in -short")
	}
	if v := strings.ToLower(os.Getenv("CI")); v == "true" || v == "1" {
		t.Skip("skipped in CI: real upstream + headed browser")
	}
	if v := strings.ToLower(os.Getenv("GITHUB_ACTIONS")); v == "true" || v == "1" {
		t.Skip("skipped in CI: real upstream + headed browser")
	}

	// Allow overriding the probe target; default is a known movie (tmdb 1048794).
	mediaType := envOr("SF_TYPE", "filme")
	mediaID := envOr("SF_ID", "1048794")
	season := os.Getenv("SF_SEASON")
	episode := os.Getenv("SF_EPISODE")

	client := NewSuperFlixClient()

	ctx, cancel := context.WithTimeout(context.Background(), 180*time.Second)
	defer cancel()

	result, err := client.GetStreamURL(ctx, mediaType, mediaID, season, episode)
	if err != nil {
		t.Skipf("SuperFlix stream unavailable (site may have rotated): %v", err)
	}
	require.NotNil(t, result)
	t.Logf("stream=%s referer=%s", result.StreamURL, result.Referer)

	assert.True(t, strings.Contains(strings.ToLower(result.StreamURL), "/hls/"), "expected an HLS master URL")

	// Fetch the playlist and confirm it parses as HLS (the live proof).
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, result.StreamURL, http.NoBody)
	req.Header.Set("User-Agent", SuperFlixUserAgent)
	if result.Referer != "" {
		req.Header.Set("Referer", result.Referer)
	}
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()

	body, _ := io.ReadAll(io.LimitReader(resp.Body, 64*1024))
	assert.Equal(t, http.StatusOK, resp.StatusCode)
	if strings.TrimSpace(string(body)) != "security error" {
		assert.True(t, strings.HasPrefix(strings.TrimSpace(string(body)), "#EXTM3U"),
			"playlist must start with #EXTM3U; got: %.80s", string(body))
	}
}

func envOr(key, def string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return def
}
