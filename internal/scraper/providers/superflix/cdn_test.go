package superflix

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestCDNPlaybackHeaders_CarriesTheFullContract_2026_08_31 pins the five
// headers the player CDN's hotlink rule requires. Dropping any one of them
// turns a freshly signed master.txt into a 403 — measured by bisecting a live
// URL — which the liveness probe then reports as a dead host, so the whole
// solve is thrown away before mpv runs.
func TestCDNPlaybackHeaders_CarriesTheFullContract_2026_08_31(t *testing.T) {
	t.Parallel()
	const referer = "https://player.example/video/deadbeef"

	h := CDNPlaybackHeaders(referer, "UA/1.0")
	assert.Equal(t, referer, h.Get("Referer"))
	assert.Equal(t, "UA/1.0", h.Get("User-Agent"))
	assert.NotEmpty(t, h.Get("Accept-Language"), "presence is what the rule checks")
	assert.Equal(t, "?0", h.Get("Sec-CH-UA-Mobile"))
	assert.NotEmpty(t, h.Get("Sec-CH-UA-Platform"))
	// The platform hint is a structured-header string: it must be quoted.
	assert.True(t, strings.HasPrefix(h.Get("Sec-CH-UA-Platform"), `"`),
		"Sec-CH-UA-Platform must be a quoted string, got %q", h.Get("Sec-CH-UA-Platform"))
}

func TestCDNPlaybackHeaders_DefaultsTheUserAgent(t *testing.T) {
	t.Parallel()
	assert.Equal(t, SuperFlixUserAgent,
		CDNPlaybackHeaders("https://player.example/video/x", "").Get("User-Agent"))
}

func TestCDNPlaybackHeaders_OmitsAnEmptyReferer(t *testing.T) {
	t.Parallel()
	h := CDNPlaybackHeaders("", "UA/1.0")
	assert.Empty(t, h.Get("Referer"), "an empty Referer must not become a blank header")
	assert.Equal(t, "UA/1.0", h.Get("User-Agent"), "the rest of the contract still applies")
}

// TestCDNPlaybackHeaderFields pins the mpv/ffmpeg wire form: one header per
// element, never comma-joined.
//
// The Accept-Language the CDN demands contains a comma, so the fields CANNOT be
// collapsed into a single mpv --http-header-fields — mpv splits that on commas
// and would send "Accept-Language: en-US" plus a junk field named "en;q=0.9",
// both of which the CDN rejects. Callers must use --http-header-fields-append
// (one per element) or join with CRLF for ffmpeg.
func TestCDNPlaybackHeaderFields(t *testing.T) {
	t.Parallel()
	const referer = "https://player.example/video/deadbeef"

	fields := CDNPlaybackHeaderFields(referer, "UA/1.0")
	require.Len(t, fields, 5)
	assert.Equal(t, "Referer: "+referer, fields[0], "Referer leads so callers can build on it")

	for _, f := range fields {
		assert.Contains(t, f, ": ", "each element is one %q header line", "Name: value")
	}
	for i, name := range []string{"Referer", "User-Agent", "Accept-Language", "Sec-CH-UA-Mobile", "Sec-CH-UA-Platform"} {
		assert.Truef(t, strings.HasPrefix(fields[i], name+": "),
			"field %d should be %s, got %q", i, name, fields[i])
	}

	// This is the property that forces the one-option-per-header form. If the
	// CDN ever stops requiring the comma-bearing value, the player may go back
	// to a single --http-header-fields — until then it must not.
	assert.Contains(t, strings.Join(fields, "\n"), ",",
		"a comma in a value is why these must not be comma-joined")
}

// TestApplyCDNPlaybackHeaders_OverridesDecorateRequest guards the ordering bug
// this replaced: decorateRequest sets its own UA, so the CDN contract has to be
// applied after it and win.
func TestApplyCDNPlaybackHeaders_OverridesDecorateRequest(t *testing.T) {
	t.Parallel()
	req, err := http.NewRequest(http.MethodGet, "https://cdn.example/master.txt", http.NoBody)
	require.NoError(t, err)
	req.Header.Set("User-Agent", "stale/0.1")
	req.Header.Set("Referer", "https://wrong.example/")

	applyCDNPlaybackHeaders(req, "https://player.example/video/abc", "UA/2.0")
	assert.Equal(t, "UA/2.0", req.Header.Get("User-Agent"))
	assert.Equal(t, "https://player.example/video/abc", req.Header.Get("Referer"))
	assert.Equal(t, "?0", req.Header.Get("Sec-CH-UA-Mobile"))
}

// TestStreamURLDead_SendsTheCDNContract is the end-to-end guard: the liveness
// probe must send all five headers, because a CDN that only ever sees Referer
// and User-Agent answers 403 and the probe declares a working stream dead.
func TestStreamURLDead_SendsTheCDNContract(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		// Mimic the real rule: 403 unless every required header is present.
		for _, n := range []string{"Referer", "User-Agent", "Accept-Language", "Sec-CH-UA-Mobile", "Sec-CH-UA-Platform"} {
			if r.Header.Get(n) == "" {
				w.WriteHeader(http.StatusForbidden)
				return
			}
		}
		w.WriteHeader(http.StatusPartialContent)
	}))
	defer srv.Close()

	c := NewClientForTest(srv.URL)
	dead := c.streamURLDead(t.Context(), srv.URL+"/master.txt", "https://player.example/video/abc", "UA/3.0")

	require.NotNil(t, got)
	assert.Equal(t, "https://player.example/video/abc", got.Get("Referer"))
	assert.Equal(t, "UA/3.0", got.Get("User-Agent"),
		"the probe must use the UA the URL was signed for, not the client default")
	assert.Equal(t, "bytes=0-1", got.Get("Range"), "the probe stays a ranged GET")
	assert.False(t, dead, "a 206 from a rule-abiding CDN must not read as dead")
}

// TestFallbackGraceFor pins the wait that dominated every play's latency.
//
// The sniff prefers a getVideo capture and used to wait a flat 8s for one. The
// current player has no getVideo endpoint at all, so that wait could only ever
// expire — 8 seconds added to every single play. Once the raw capture's Referer
// identifies the player page it already carries everything getVideo would have
// added, so there is nothing left to wait for.
func TestFallbackGraceFor(t *testing.T) {
	t.Parallel()

	assert.Zero(t, fallbackGraceFor("https://player.best/video/6f1896bb1bfc3d0bcd2cba28ad968dd4"),
		"a player /video/<hash> Referer already carries host+hash — do not wait")

	for _, ref := range []string{
		"",
		"https://player.best/",
		"https://superflixapi.baby/filme/603",
		"not a url",
	} {
		assert.Positivef(t, fallbackGraceFor(ref),
			"%q does not identify the player, so getVideo still deserves its grace", ref)
	}
}

// TestPlayerIdentityFromReferer covers the parse the grace decision rests on.
func TestPlayerIdentityFromReferer(t *testing.T) {
	t.Parallel()

	host, hash := playerIdentityFromReferer("https://xn--tckasiu6cvova0eb5fua2449g98vg.best/video/6f1896bb1bfc3d0bcd2cba28ad968dd4")
	assert.Equal(t, "https://xn--tckasiu6cvova0eb5fua2449g98vg.best", host)
	assert.Equal(t, "6f1896bb1bfc3d0bcd2cba28ad968dd4", hash)

	for _, bad := range []string{"", "https://player.best/", "https://player.best/video/", "player.best/video/abc"} {
		h, x := playerIdentityFromReferer(bad)
		assert.Emptyf(t, h, "%q must not yield a host", bad)
		assert.Emptyf(t, x, "%q must not yield a hash", bad)
	}
}
