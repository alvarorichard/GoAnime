package superflix

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const testLayerKey = "a0b567c03b96daacbf60368ce59272ab6a965bd14XrbympqU6l5iDg3WFVbjA"

// TestFetchLayerKey_2026_09_01 pins the extraction of the rotating key from the
// player bundle. The key is not a constant: three consecutive live sessions
// produced three different values, so a hardcoded one would work once and then
// answer 410 Gone forever.
func TestFetchLayerKey_2026_09_01(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Contains(t, r.URL.Path, "/player/assets/scripts.php")
		fmt.Fprintf(w, `$.ajax({type:"POST",url:"/layer/%s/"+ID+"/",data:{hash:ID,r:document.referrer}});`, testLayerKey)
	}))
	defer srv.Close()

	c := NewClientForTest(srv.URL)
	key, err := c.fetchLayerKey(t.Context(), srv.URL, srv.URL+"/video/abc")
	require.NoError(t, err)
	assert.Equal(t, testLayerKey, key)
}

func TestFetchLayerKey_MissingKeyIsAnError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		fmt.Fprint(w, `console.log("no layer endpoint here");`)
	}))
	defer srv.Close()

	c := NewClientForTest(srv.URL)
	_, err := c.fetchLayerKey(t.Context(), srv.URL, "")
	require.Error(t, err, "a bundle without the key means the contract moved again")
	assert.Contains(t, err.Error(), "no layer key")
}

// TestGetVideoViaLayer covers the two-request signing flow end to end.
func TestGetVideoViaLayer(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/player/assets/scripts.php"):
			fmt.Fprintf(w, `url:"/layer/%s/"+ID+"/"`, testLayerKey)
		case strings.HasPrefix(r.URL.Path, "/layer/"):
			assert.Equal(t, "/layer/"+testLayerKey+"/hash42/", r.URL.Path)
			require.NoError(t, r.ParseForm())
			assert.Equal(t, "hash42", r.PostForm.Get("hash"))
			assert.NotEmpty(t, r.PostForm.Get("r"), "the player sends its referrer")
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"hls":true,"videoSource":"https://cdn.example/a/b/1/master.txt","securedLink":"https://cdn.example/stale.txt","videoImage":"https://img.example/t.jpg"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := NewClientForTest(srv.URL)
	streamURL, thumb, err := c.getVideoViaLayer(t.Context(), srv.URL, "hash42", srv.URL+"/video/hash42")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/a/b/1/master.txt", streamURL,
		"videoSource wins over securedLink, as it does for getVideo")
	assert.Equal(t, "https://img.example/t.jpg", thumb)
}

// An HTML error page must surface as an actionable error, not a JSON decode
// failure — the layer endpoint answers 410 Gone for a stale key.
func TestGetVideoViaLayer_HTMLErrorIsActionable(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/player/assets/scripts.php") {
			fmt.Fprintf(w, `url:"/layer/%s/"+ID+"/"`, testLayerKey)
			return
		}
		w.WriteHeader(http.StatusGone)
		fmt.Fprint(w, "<html><head><title>410 Gone</title></head><body>410</body></html>")
	}))
	defer srv.Close()

	c := NewClientForTest(srv.URL)
	_, _, err := c.getVideoViaLayer(t.Context(), srv.URL, "hash42", "")
	require.Error(t, err)
	assert.NotContains(t, err.Error(), "looking for beginning of value",
		"an HTML body must not surface as a raw JSON decode error")
}

// TestSignStreamURL_FallsBackToLegacyGetVideo covers a player host that still
// serves the old contract: the layer attempt fails, getVideo answers, and the
// caller never sees the difference.
func TestSignStreamURL_FallsBackToLegacyGetVideo(t *testing.T) {
	t.Parallel()
	var legacyHits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/player/index.php") {
			legacyHits++
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"videoSource":"https://cdn.example/legacy.m3u8"}`)
			return
		}
		// No layer key in the bundle, and no /layer/ endpoint.
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewClientForTest(srv.URL)
	streamURL, _, err := c.signStreamURL(t.Context(), srv.URL, "hash42", "")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/legacy.m3u8", streamURL)
	assert.Equal(t, 1, legacyHits)
}

// When both contracts are gone the error must name the layer failure, since
// that is the one worth acting on.
func TestSignStreamURL_BothGone(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer srv.Close()

	c := NewClientForTest(srv.URL)
	_, _, err := c.signStreamURL(t.Context(), srv.URL, "hash42", "")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "layer signing failed")
	assert.Contains(t, err.Error(), "legacy getVideo also failed")
}

// The key regex must not be fooled by unrelated short paths in the bundle.
func TestSFLayerKeyRe(t *testing.T) {
	t.Parallel()
	assert.Nil(t, sfLayerKeyRe.FindStringSubmatch(`url:"/layer/short/"+ID+"/"`),
		"a short segment is not a key")

	m := sfLayerKeyRe.FindStringSubmatch(`url: "/layer/` + testLayerKey + `/"+ID+"/"`)
	require.Len(t, m, 2)
	assert.Equal(t, testLayerKey, m[1])
}

// TestEffectiveUserAgent guards the UA a result reports against the UA its
// requests actually carry.
//
// cfFallbackTransport rewrites every request to the solving browser's UA so the
// UA-bound cf_clearance stays valid. A result that reported c.userAgent instead
// would hand mpv a UA the CDN never signed for, and every fetch would 403 — the
// same failure mode as sending no client hints at all.
func TestEffectiveUserAgent(t *testing.T) {
	c := NewSuperFlixClient()
	assert.Equal(t, c.userAgent, c.effectiveUserAgent(),
		"before any solve the client's own UA is what goes out")

	tr, ok := c.client.Transport.(*cfFallbackTransport)
	require.True(t, ok, "the production client wraps cfFallbackTransport")

	const solved = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/151.0.0.0 Safari/537.36"
	tr.setSolvedUA(solved)
	assert.Equal(t, solved, c.effectiveUserAgent(),
		"after a solve the transport rewrites the UA, so that is the one to report")

	// A test client has a plain transport and must fall back cleanly.
	tc := NewClientForTest("https://sf.test")
	assert.Equal(t, tc.userAgent, tc.effectiveUserAgent())
}

// TestStreamFromServer_SignsViaLayer is the regression for the warning a user
// saw on a server that was working fine:
//
//	SuperFlix: the chosen server failed; falling back server="Servidor 324278"
//	error="... video API endpoint returned HTML (status 403,
//	url=".../player/index.php?data=...&do=getVideo")"
//
// The server had resolved correctly; only the retired getVideo endpoint was
// gone. Signing through /layer/ first keeps a good server from being reported
// as failed and from falling back to a full browser solve.
func TestStreamFromServer_SignsViaLayer(t *testing.T) {
	var legacyHits int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.Contains(r.URL.Path, "/player/index.php"):
			// The retired endpoint, answering exactly as the live one does.
			legacyHits++
			w.WriteHeader(http.StatusForbidden)
			fmt.Fprint(w, "<html><body>403</body></html>")
		case strings.HasPrefix(r.URL.Path, "/player/assets/scripts.php"):
			fmt.Fprintf(w, `url:"/layer/%s/"+ID+"/"`, testLayerKey)
		case strings.HasPrefix(r.URL.Path, "/layer/"):
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, `{"videoSource":"https://cdn.example/signed/master.txt"}`)
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer srv.Close()

	c := NewClientForTest(srv.URL)
	streamURL, _, err := c.signStreamURL(t.Context(), srv.URL, "hash42", srv.URL+"/video/hash42")

	require.NoError(t, err, "a live /layer/ must not be reported as a failed server")
	assert.Equal(t, "https://cdn.example/signed/master.txt", streamURL)
	assert.Zero(t, legacyHits, "the retired getVideo must not be tried first")
}
