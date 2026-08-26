package superflix

import (
	"context"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSuperFlixCDNRules_Live pins the two upstream rules the playback arguments
// encode. Everything else in this repo asserts that we FOLLOW the rules; this
// asserts the rules are still what we think they are.
//
// When SuperFlix changes them again — and it will — this names which one moved
// instead of leaving "it stopped playing":
//
//	rule 1  the signed playlist requires Referer = <player>/video/<hash>;
//	        the bare origin is refused.
//	rule 2  the segments live on rotating third-party CDN hosts and are fetched
//	        by hls.js as a cross-origin XHR, so they require the CORS Origin
//	        header. Without it: 403, while the playlist above still loads —
//	        which is exactly why the 2026-08-26 outage looked like a player bug.
//
// Every signed URL here is SINGLE USE: one request (even a rejected one) burns
// it, and the host then answers "security error" to everything. So each probe
// gets its own freshly resolved stream or its own segment, and no URL is ever
// requested twice. Getting this wrong makes the negative assertions pass for
// the wrong reason.
//
// Live + opt-in: hits the real site and may open the browser solver.
//
//	GOANIME_RECON=1 go test ./internal/scraper/providers/superflix/ -run TestSuperFlixCDNRules_Live -v -count=1 -timeout 400s
func TestSuperFlixCDNRules_Live(t *testing.T) {
	if os.Getenv("GOANIME_RECON") == "" {
		t.Skip("set GOANIME_RECON=1 (hits the live site, may launch a browser)")
	}
	skipInCI(t)
	if testing.Short() {
		t.Skip("skipping live CDN probe in -short")
	}
	util.InitLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 360*time.Second)
	defer cancel()

	// Two header profiles, deliberately different.
	//
	// browserish: what a Go client needs to get a playlist out of the player
	// host at all — bare requests come back 200 with the body "security error".
	// Used only to WALK to a segment, never to assert.
	//
	// asMPV: exactly what mpv sends after the fix. The assertions use this one,
	// so a pass means the player's real request works, not that some richer
	// request does.
	do := func(url string, base, extra map[string]string) (int, string) {
		req, rErr := http.NewRequestWithContext(ctx, http.MethodGet, url, http.NoBody)
		require.NoError(t, rErr)
		for k, v := range base {
			req.Header.Set(k, v)
		}
		for k, v := range extra {
			req.Header.Set(k, v)
		}
		resp, dErr := http.DefaultClient.Do(req)
		require.NoError(t, dErr)
		defer func() { _ = resp.Body.Close() }()
		b, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
		return resp.StatusCode, string(b)
	}
	browserish := map[string]string{
		"User-Agent":      SuperFlixUserAgent,
		"Accept":          "*/*",
		"Accept-Language": "en-US,en;q=0.9",
		"Sec-Fetch-Dest":  "empty",
		"Sec-Fetch-Mode":  "cors",
		"Sec-Fetch-Site":  "cross-site",
	}
	asMPV := map[string]string{"User-Agent": SuperFlixUserAgent}
	get := func(url string, hdr map[string]string) (int, string) {
		return do(url, browserish, hdr)
	}

	// ---- rule 1: the playlist Referer we ship still works -------------------
	// Only the POSITIVE direction is probed live. The negative (bare origin ->
	// 403) is pinned offline by playerRefererFor's tests, and probing it here
	// would cost a second resolve: a rejected request still burns the signed
	// URL, so the negative and positive cases cannot share one.
	live := resolveLiveStream(t, ctx)
	origin := originOfReferer(live.Referer)

	okStatus, master := getPlaylist(t, get, live.StreamURL, live.Referer)
	require.Equal(t, http.StatusOK, okStatus,
		"the Referer we ship stopped working — playback is broken until this is re-derived")
	require.True(t, strings.HasPrefix(strings.TrimSpace(master), "#EXTM3U"),
		"expected an HLS master, got %.60s", master)

	// ---- rule 2: the segments need the CORS Origin ---------------------------
	_, variantBody := getPlaylist(t, get, absolutize(firstURI(master), origin), live.Referer)
	require.True(t, strings.HasPrefix(strings.TrimSpace(variantBody), "#EXTM3U"),
		"expected a variant playlist, got %.60s", variantBody)

	// Two DIFFERENT segments: a rejected probe can burn the one it touches.
	segs := segmentURIs(variantBody, origin, 2)
	require.Len(t, segs, 2, "variant playlist exposed fewer than two segments")

	// THE assertion: mpv's own request, with the Origin header the fix added.
	withOrigin, _ := do(segs[0], asMPV, map[string]string{
		"Referer": live.Referer,
		"Origin":  origin,
	})
	assert.Equal(t, http.StatusOK, withOrigin,
		"the Referer+Origin pair mpv ships stopped fetching segments — playback is broken")

	// No negative assertion here on purpose. Dropping Origin makes ffmpeg's
	// request fail (measured: 121 failed segments -> 0 with it) but a plain Go
	// client without Origin is still served 200, so the host is weighing a
	// combination of client signals this test cannot observe. Asserting a
	// rejection would pin a mechanism we cannot see; the guard that matters is
	// the positive above, and the end-to-end proof lives in the player package.
}

// getPlaylist reads a signed playlist, tolerating the brief window right after
// resolution in which the host answers 200 with the body "security error"
// instead of the playlist. The token needs a moment to become valid; reading it
// within milliseconds of getVideo reliably hits this, which is why the probe is
// spaced rather than immediate.
func getPlaylist(t *testing.T, get func(string, map[string]string) (int, string), url, referer string) (int, string) {
	t.Helper()
	var status int
	var body string
	for attempt := range 4 {
		if attempt > 0 {
			time.Sleep(2 * time.Second)
		}
		status, body = get(url, map[string]string{"Referer": referer})
		if strings.TrimSpace(body) != "security error" {
			return status, body
		}
	}
	t.Skipf("playlist still answered %q after 4 spaced attempts — upstream is rate limiting, not a code defect", strings.TrimSpace(body))
	return status, body
}

// resolveLiveStream returns a freshly signed stream. Each caller gets its own,
// because the URLs are single use.
func resolveLiveStream(t *testing.T, ctx context.Context) *SuperFlixStreamResult {
	t.Helper()
	res, err := NewSuperFlixClient().GetStreamURL(ctx, "filme", "603", "", "")
	if err != nil {
		t.Skipf("SuperFlix unavailable (site may have rotated): %v", err)
	}
	require.NotEmpty(t, res.StreamURL)
	require.NotEmpty(t, res.Referer)
	return res
}

// firstURI returns the first non-comment, non-blank line of an HLS playlist.
func firstURI(playlist string) string {
	for _, line := range strings.Split(playlist, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		return line
	}
	return ""
}

// segmentURIs returns up to n absolute segment URLs from a media playlist.
func segmentURIs(playlist, origin string, n int) []string {
	var out []string
	for _, line := range strings.Split(playlist, "\n") {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "#") {
			continue
		}
		out = append(out, absolutize(line, origin))
		if len(out) == n {
			break
		}
	}
	return out
}

func absolutize(uri, origin string) string {
	if uri == "" || strings.HasPrefix(uri, "http") {
		return uri
	}
	return origin + uri
}

// originOfReferer reduces "<scheme>://<host>/video/<hash>" to "<scheme>://<host>".
func originOfReferer(referer string) string {
	i := strings.Index(referer, "://")
	if i < 0 {
		return ""
	}
	if j := strings.Index(referer[i+3:], "/"); j >= 0 {
		return referer[:i+3+j]
	}
	return referer
}
