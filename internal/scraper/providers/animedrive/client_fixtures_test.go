package animedrive

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fixtureServer serves a real captured player page (testdata/*.html) with the
// CDN media hosts rewritten to the test server, so availability probing stays
// hermetic. videoStatus controls the probe outcome per media path; nil means
// 206 for every media file.
func fixtureServer(t *testing.T, fixture string, videoStatus func(path string) int, gotReferer *string) *httptest.Server {
	t.Helper()
	raw, err := os.ReadFile("testdata/" + fixture)
	require.NoError(t, err)

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".mp4") {
			if gotReferer != nil {
				*gotReferer = r.Header.Get("Referer")
			}
			status := http.StatusPartialContent
			if videoStatus != nil {
				status = videoStatus(r.URL.Path)
			}
			w.WriteHeader(status)
			return
		}
		page := string(raw)
		for _, host := range []string{"animes.click", "aniplay.online", "animeflix.blog", "mangas.cloud"} {
			page = strings.ReplaceAll(page, "https://"+host, server.URL)
		}
		_, _ = fmt.Fprint(w, page)
	}))
	t.Cleanup(server.Close)
	return server
}

// bloggerOKClient answers Blogger embeds with 206 without network, like the
// unit test transport, while routing everything else normally.
func bloggerOKClient() *http.Client {
	return &http.Client{Transport: bloggerMockTransport{fallback: http.DefaultTransport}}
}

func TestAnimeDriveFixture_MagicalExplorer_QualityOrder(t *testing.T) {
	t.Parallel()

	var gotReferer string
	server := fixtureServer(t, "mag.html", nil, &gotReferer)

	client := NewClientForTest(server.URL)
	client.client = bloggerOKClient()

	best, meta, err := client.GetStreamURLWithQuality("/episodio-01/", "best")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/Midias/Animes/Letra-M/Magical-Explorer/01.mp4", best)
	assert.Equal(t, "FHD", meta["audio_lang"])
	assert.Equal(t, server.URL+"/", meta["referer"])
	assert.Equal(t, server.URL+"/", gotReferer)

	worst, meta, err := client.GetStreamURLWithQuality("/episodio-01/", "worst")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/CDN/Midia/Letra-M/Magical-Explorer/01.mp4", worst)
	assert.Equal(t, "SD", meta["audio_lang"])
}

func TestAnimeDriveFixture_Frieren_BloggerFallback(t *testing.T) {
	t.Parallel()

	server := fixtureServer(t, "fri.html", func(path string) int {
		if strings.Contains(path, "Frieren") {
			return http.StatusNotFound
		}
		return http.StatusPartialContent
	}, nil)

	client := NewClientForTest(server.URL)
	client.client = bloggerOKClient()

	stream, meta, err := client.GetStreamURL("/episodio-01/")
	require.NoError(t, err)
	assert.Contains(t, stream, "blogger.com/video.g")
	assert.Equal(t, "Mobile / Celular", meta["audio_lang"])
	assert.Equal(t, server.URL+"/", meta["referer"])
}

func TestAnimeDriveFixture_Boruto_AllDead(t *testing.T) {
	t.Parallel()

	server := fixtureServer(t, "bor.html", func(string) int {
		return http.StatusNotFound
	}, nil)

	client := NewClientForTest(server.URL)

	_, _, err := client.GetStreamURL("/episodio-01/")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the 2 servers")
}

func TestAnimeDriveFixture_Kimetsu_NoPlayableServer(t *testing.T) {
	t.Parallel()

	server := fixtureServer(t, "kim.html", nil, nil)

	client := NewClientForTest(server.URL)

	_, _, err := client.GetStreamURL("/episodio-01/")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no video tracks found in animeq-player")
}

func TestAnimeDriveStream_ChallengePage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, `<html><head><title>Just a moment...</title></head><body>Verifying you are human</body></html>`)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)

	_, _, err := client.GetStreamURL("/episodio-01/")
	require.Error(t, err)
}
