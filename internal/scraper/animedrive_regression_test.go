package scraper

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/providers/animedrive"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// bloggerLiveTransport answers blogger embeds the way the live site does
// (200 HTML), keeping the adapter cascade hermetic.
type animeDriveBloggerLiveTransport struct{ fallback http.RoundTripper }

func (t animeDriveBloggerLiveTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Host, "blogger.com") {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("<html>blogger</html>")),
			Request:    req,
		}, nil
	}
	return t.fallback.RoundTrip(req)
}

// adapterCascadeServer serves search → details → player for two titles:
// recoverable (dead HLS + live blogger) and dead-only (Legendado/Dublado
// 404s, the Boruto live shape). It proves the adapter glue forwards quality
// and surfaces actionable errors end to end.
func adapterCascadeServer(t *testing.T) *httptest.Server {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, ".mp4"), strings.HasSuffix(r.URL.Path, ".m3u8"):
			if strings.Contains(r.URL.Path, "dead") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusPartialContent)
			return
		case r.URL.Path == "/" && r.URL.Query().Has("s"):
			_, _ = fmt.Fprint(w, `<html><body>
				<div class="result-item"><article><div class="title"><a href="/anime/recoverable/">Recoverable</a></div><img src="/r.jpg"/></article></div>
				<div class="result-item"><article><div class="title"><a href="/anime/deadonly/">Dead Only</a></div><img src="/d.jpg"/></article></div>
			</body></html>`)
			return
		case r.URL.Path == "/anime/recoverable/":
			_, _ = fmt.Fprint(w, `<html><body>
				<h1 class="entry-title">Recoverable</h1>
				<ul class="episodios"><li><a href="/episodio-recoverable-1/">Episode 1</a></li></ul>
			</body></html>`)
			return
		case r.URL.Path == "/anime/deadonly/":
			_, _ = fmt.Fprint(w, `<html><body>
				<h1 class="entry-title">Dead Only</h1>
				<ul class="episodios"><li><a href="/episodio-deadonly-1/">Episode 1</a></li></ul>
			</body></html>`)
			return
		case r.URL.Path == "/episodio-recoverable-1/":
			_, _ = fmt.Fprintf(w, `
			<div class="animeq-player__source" data-animeq-source="0"><div><iframe src="https://www.blogger.com/video.g?token=live"></iframe></div></div>
			<div class="animeq-player__source" data-animeq-source="1"><video><source src="%s/dead.m3u8&#038;e=1"></video></div>
			<button class="animeq-player__server" data-animeq-switch="0" data-source-name="Mobile / Celular">Mobile</button>
			<button class="animeq-player__server" data-animeq-switch="1" data-source-name="FullHD / HLS">HLS</button>
			`, server.URL)
			return
		case r.URL.Path == "/episodio-deadonly-1/":
			_, _ = fmt.Fprintf(w, `
			<div class="animeq-player__source" data-animeq-source="0"><video><source src="%s/dead-a.mp4"></video></div>
			<div class="animeq-player__source" data-animeq-source="1"><video><source src="%s/dead-b.mp4"></video></div>
			<button class="animeq-player__server" data-animeq-switch="0" data-source-name="Legendado">Legendado</button>
			<button class="animeq-player__server" data-animeq-switch="1" data-source-name="Dublado">Dublado</button>
			`, server.URL, server.URL)
			return
		case r.URL.Path == "/episodio-embedonly-1/":
			_, _ = fmt.Fprint(w, `
			<div class="animeq-player__source" data-animeq-source="0"><div><iframe src="https://animeshd.cloud/#a"></iframe></div></div>
			<div class="animeq-player__source" data-animeq-source="1"><div><iframe src="https://animes.strp2p.com/#b"></iframe></div></div>
			<button class="animeq-player__server" data-animeq-switch="0" data-source-name="FHD">FHD</button>
			<button class="animeq-player__server" data-animeq-switch="1" data-source-name="FHD">FHD</button>
			`)
			return
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func adapterForCascade(t *testing.T, serverURL string) *AnimeDriveAdapter {
	t.Helper()
	client := animedrive.NewClientForTest(serverURL)
	client.SetHTTPClientForTest(&http.Client{Transport: animeDriveBloggerLiveTransport{fallback: http.DefaultTransport}})
	return &AnimeDriveAdapter{client: client}
}

func TestAnimeDriveAdapterCascade_SearchToStreamRecoverable(t *testing.T) {
	t.Parallel()

	server := adapterCascadeServer(t)
	adapter := adapterForCascade(t, server.URL)

	results, err := adapter.SearchAnime("recoverable")
	require.NoError(t, err)
	require.Len(t, results, 2)

	var recoverableURL string
	for _, a := range results {
		if a.Name == "Recoverable" {
			recoverableURL = a.URL
		}
	}
	require.NotEmpty(t, recoverableURL)

	episodes, err := adapter.GetAnimeEpisodes(recoverableURL)
	require.NoError(t, err)
	require.Len(t, episodes, 1)

	stream, meta, err := adapter.GetStreamURL(episodes[0].URL, "best")
	require.NoError(t, err, "dead HLS must fall back to live blogger through the adapter")
	assert.Equal(t, "https://www.blogger.com/video.g?token=live", stream)
	assert.Equal(t, "animedrive", meta["source"])
	assert.Equal(t, "Mobile / Celular", meta["audio_lang"])
}

func TestAnimeDriveAdapterCascade_DeadOnlyErrorCarriesLabels(t *testing.T) {
	t.Parallel()

	server := adapterCascadeServer(t)
	adapter := adapterForCascade(t, server.URL)

	episodes, err := adapter.GetAnimeEpisodes(server.URL + "/anime/deadonly/")
	require.NoError(t, err)
	require.Len(t, episodes, 1)

	_, meta, err := adapter.GetStreamURL(episodes[0].URL, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the 2 servers")
	assert.Contains(t, err.Error(), "Legendado")
	assert.Contains(t, err.Error(), "Dublado")
	require.NotNil(t, meta, "adapter must keep metadata on failure")
	assert.Equal(t, "animedrive", meta["source"])
}

func TestAnimeDriveAdapterCascade_EmbedOnlyErrorCarriesHosts(t *testing.T) {
	t.Parallel()

	server := adapterCascadeServer(t)
	adapter := adapterForCascade(t, server.URL)

	_, _, err := adapter.GetStreamURL(server.URL+"/episodio-embedonly-1/", "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no video tracks")
	assert.Contains(t, err.Error(), "animeshd.cloud")
	assert.Contains(t, err.Error(), "animes.strp2p.com")
}
