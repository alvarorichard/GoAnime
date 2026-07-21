package superflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newExtrasTestClient returns a client whose transport is a plain one. The
// production client refuses loopback addresses (SSRF guard), so an httptest
// server is unreachable through it.
func newExtrasTestClient(baseURL string) *SuperFlixClient {
	c := NewSuperFlixClient()
	c.SetTestConfig(baseURL, &http.Client{})
	return c
}

// realPlayerPage mirrors what the warezcdn player actually serves (captured live
// from a Dexter S1E1 stream): the HLS carries several audio tracks and the
// Portuguese subtitles are a separate file. This is exactly the data the browser
// path used to throw away.
const realPlayerPage = `<html><head><title>Player</title></head><body>
<script>
var defaultAudio = ["por","und","eng","spa","kor","jpn","chi","und"];
var defaultCaptions = {"und":"por","por":"por","eng":"por"};
var playerjsSubtitle = "[Portuguese]https://salvavidas.buzz/q/abc123.html";
var playerjsDefaultSubtitle = "Portuguese";
</script>
</body></html>`

// TestFetchPlayerExtras pins the recovery of the tracks the production (browser)
// path used to drop: it sniffs the media URL out of network traffic and never
// reads the player page, so streams arrived with zero subtitles and no audio
// info. The player host is not Cloudflare-gated, so one plain GET restores both.
func TestFetchPlayerExtras(t *testing.T) {
	t.Parallel()

	var gotPath, gotReferer string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotPath = r.URL.Path
		gotReferer = r.Header.Get("Referer")
		_, _ = fmt.Fprint(w, realPlayerPage)
	}))
	t.Cleanup(srv.Close)

	c := newExtrasTestClient(srv.URL)
	audio, subs := c.fetchPlayerExtras(context.Background(), srv.URL, "deadbeefhash")

	assert.Equal(t, "/video/deadbeefhash", gotPath, "must read the player page for this (host, hash)")
	assert.Equal(t, srv.URL+"/", gotReferer, "the player host rejects requests without its own referer")

	assert.Equal(t, []string{"por", "und", "eng", "spa", "kor", "jpn", "chi", "und"}, audio)
	require.Len(t, subs, 1)
	assert.Equal(t, "Portuguese", subs[0].Lang)
	assert.Equal(t, "https://salvavidas.buzz/q/abc123.html", subs[0].URL)
}

// Extras only enrich playback, so a player page we cannot reach must never take
// the stream down with it.
func TestFetchPlayerExtras_FailuresAreNonFatal(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	c := newExtrasTestClient(srv.URL)

	tests := []struct {
		name       string
		host, hash string
	}{
		{"server error", srv.URL, "hash"},
		{"unreachable host", "http://127.0.0.1:1", "hash"},
		{"no host", "", "hash"},
		{"no hash", srv.URL, ""},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.NotPanics(t, func() {
				audio, subs := c.fetchPlayerExtras(context.Background(), tt.host, tt.hash)
				assert.Empty(t, audio)
				assert.Empty(t, subs)
			})
		})
	}
}

// The raw-media fallback capture yields no (host, hash), so there is nothing to
// read extras from — it must not fire a bogus request at an empty URL.
func TestFetchPlayerExtras_SkipsWhenNoPlayerHost(t *testing.T) {
	t.Parallel()

	var called bool
	srv := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		called = true
	}))
	t.Cleanup(srv.Close)

	c := newExtrasTestClient(srv.URL)
	audio, subs := c.fetchPlayerExtras(context.Background(), "", "")

	assert.False(t, called)
	assert.Empty(t, audio)
	assert.Empty(t, subs)
	assert.True(t, strings.HasPrefix(srv.URL, "http"), "server was started but must go untouched")
}

type extrasBrowserTrap struct {
	t      *testing.T
	called bool
}

func (s *extrasBrowserTrap) Solve(context.Context, string, time.Duration) (*CFSolveResult, error) {
	s.called = true
	s.t.Error("optional player extras must never escalate to the headed browser")
	return nil, fmt.Errorf("unexpected browser solve")
}

// A retired player route currently returns a 200 HTML shell that resembles the
// provider's verification page.  The shared transport normally opens Chromium
// for that response; extras are optional, so this request must opt out and let
// playback proceed with the stream that was already captured.
func TestFetchPlayerExtras_DoesNotOpenBrowserForDeadPlayerPage(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, `<html><title>Verificação</title><div class="cf-turnstile"></div></html>`)
	}))
	t.Cleanup(srv.Close)

	trap := &extrasBrowserTrap{t: t}
	transport := &cfFallbackTransport{
		base:    http.DefaultTransport,
		solver:  trap,
		timeout: time.Second,
	}
	c := newExtrasTestClient(srv.URL)
	c.client = &http.Client{Transport: transport}

	audio, subs := c.fetchPlayerExtras(context.Background(), srv.URL, "retired-hash")

	assert.False(t, trap.called)
	assert.Empty(t, audio)
	assert.Empty(t, subs)
}
