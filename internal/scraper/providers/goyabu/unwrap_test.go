package goyabu

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Kimetsu no Yaiba Movie (2026-11-22): AnimeDrive fell back to Goyabu, whose
// player data held api.anivideo.net/videohls.php?d=<HLS>&amp;nocache…, still
// HTML-escaped. That wrapper answers 404 "File not found" for every file, even
// one whose d address plays (One Piece 01: d 200, wrapper 404), so mpv was
// handed a dead address. The d address is what plays.

func TestUnwrapPlayerURL(t *testing.T) {
	const hls = "https://cdn-s01.mywallpaper-4k-image.net/stream/sv/o/one-piece/01.mp4/index.m3u8"
	for name, tc := range map[string]struct{ raw, want string }{
		"the wrapper as the page has it": {"https://api.anivideo.net/videohls.php?d=" + hls + "&amp;nocache1753210104", hls},
		"the wrapper unescaped":          {"https://api.anivideo.net/videohls.php?d=" + hls + "&nocache1", hls},
		"the wrapper's host in capitals": {"https://API.anivideo.net/videohls.php?d=" + hls, hls},
		"padded with blanks":             {"  https://api.anivideo.net/videohls.php?d=" + hls + "  ", hls},
		"the wrapper with no d":          {"https://api.anivideo.net/videohls.php?x=1", "https://api.anivideo.net/videohls.php?x=1"},
		"the wrapper with an empty d":    {"https://api.anivideo.net/videohls.php?d=", "https://api.anivideo.net/videohls.php?d="},
		"another anivideo endpoint":      {"https://api.anivideo.net/file/abc?d=" + hls, "https://api.anivideo.net/file/abc?d=" + hls},
		"another host with a d":          {"https://other.example/videohls.php?d=" + hls, "https://other.example/videohls.php?d=" + hls},
		"a Blogger embed, escaped":       {"https://www.blogger.com/video.g?token=A&amp;b=1", "https://www.blogger.com/video.g?token=A&b=1"},
		"what is no URL":                 {"::%zz", "::%zz"},
	} {
		assert.Equal(t, tc.want, unwrapPlayerURL(tc.raw), name)
	}
}

func goyabuServing(t *testing.T, page string) *GoyabuClient {
	t.Helper()
	mux := http.NewServeMux()
	mux.HandleFunc("/episode/1", func(w http.ResponseWriter, _ *http.Request) { _, _ = fmt.Fprint(w, page) })
	mux.HandleFunc("/wp-admin/admin-ajax.php", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = fmt.Fprint(w, `{"success":false,"data":{"message":"Nenhum vídeo encontrado."}}`)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	c := NewGoyabuClient()
	c.baseURL, c.maxRetries, c.retryDelay = srv.URL, 0, 0
	return c
}

func TestGoyabuHandsOverTheHLSBehindTheAnivideoWrapper(t *testing.T) {
	const hls = "https://cdn-s01.mywallpaper-4k-image.net/stream/k/kimetsu/01cambr.mp4/index.m3u8"
	c := goyabuServing(t, `<html><head><script>
		var playersData = [{"name":"Blog","select":"blogger","url":"https://api.anivideo.net/videohls.php?d=`+hls+`&amp;nocache1753210104","blogger_token":"dGVzdA=="}];
	</script></head><body></body></html>`)
	got, err := c.GetEpisodeStreamURL(c.baseURL + "/episode/1")
	require.NoError(t, err)
	assert.Equal(t, hls, got, "the playlist itself, not the wrapper that answers 404")
}

func TestGoyabuUnescapesAndUnwrapsWhatTheScriptsName(t *testing.T) {
	const hls = "https://cdn.example.com/v/ep1/index.m3u8"
	for name, tc := range map[string]struct{ file, want string }{
		"an escaped address":   {"https://cdn.example.com/v/ep1.m3u8?a=1&amp;b=2", "https://cdn.example.com/v/ep1.m3u8?a=1&b=2"},
		"the anivideo wrapper": {"https://api.anivideo.net/videohls.php?d=" + hls + "&amp;nocache1", hls},
	} {
		c := goyabuServing(t, `<html><head><script>var cfg = {"file":"`+tc.file+`"};</script></head><body></body></html>`)
		got, err := c.GetEpisodeStreamURL(c.baseURL + "/episode/1")
		require.NoError(t, err, name)
		assert.Equal(t, tc.want, got, name)
	}
}
