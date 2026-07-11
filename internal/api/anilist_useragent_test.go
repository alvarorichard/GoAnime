package api

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Issue #184: AniList answers browser User-Agents with a 403 ("The AniList API
// has been temporarily disabled due to severe stability issues") while serving
// plain API clients normally. GoAnime used to lose because the shared surf
// clients impersonate Chrome and REWRITE the User-Agent to a browser one, so the
// bare "GoAnime/1.0" the code set never reached the wire.
//
// The invariant these tests pin: AniList requests must carry a NON-browser UA,
// which means they must not travel on the shared/impersonating clients.

// browserMarkers are the tokens a browser UA is built from. AniList's block keys
// off exactly this shape, so none of them may appear in our AniList UA.
var browserMarkers = []string{"Mozilla", "AppleWebKit", "Chrome", "Firefox", "Safari", "Gecko", "Edg/"}

func TestAPIUserAgent_IsNotBrowserLike(t *testing.T) {
	t.Parallel()
	for _, marker := range browserMarkers {
		assert.NotContains(t, netx.APIUserAgent, marker,
			"AniList rejects browser-shaped User-Agents with a 403 — %q must not appear", marker)
	}
	assert.Contains(t, netx.APIUserAgent, "GoAnime", "the UA should still identify the app")
}

// TestAniListPost_SendsNonBrowserUserAgent is the real regression guard: it
// inspects what actually reaches the wire, which is what the surf client broke.
func TestAniListPost_SendsNonBrowserUserAgent(t *testing.T) {
	t.Parallel()

	var gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotUA = r.Header.Get("User-Agent")
		// Mirror AniList: a browser UA gets the 403 block, everyone else is served.
		for _, marker := range browserMarkers {
			if strings.Contains(gotUA, marker) {
				w.WriteHeader(http.StatusForbidden)
				_, _ = w.Write([]byte(`{"errors":[{"message":"The AniList API has been temporarily disabled"}]}`))
				return
			}
		}
		_, _ = w.Write([]byte(`{"data":{"Media":{"id":20,"idMal":20}}}`))
	}))
	t.Cleanup(srv.Close)

	resp, body, err := aniListPost(srv.URL, []byte(`{"query":"x"}`))
	require.NoError(t, err)

	assert.Equal(t, netx.APIUserAgent, gotUA)
	assert.Equal(t, http.StatusOK, resp.StatusCode,
		"a browser-shaped UA would have been 403'd by AniList")
	assert.Contains(t, string(body), `"id":20`)
}

func TestAniListPost_ReturnsBodyAndStatus(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusTooManyRequests)
		_, _ = w.Write([]byte(`{"errors":[{"message":"rate limited"}]}`))
	}))
	t.Cleanup(srv.Close)

	resp, body, err := aniListPost(srv.URL, []byte(`{"query":"x"}`))
	require.NoError(t, err)
	assert.Equal(t, http.StatusTooManyRequests, resp.StatusCode)
	assert.Contains(t, string(body), "rate limited")
}
