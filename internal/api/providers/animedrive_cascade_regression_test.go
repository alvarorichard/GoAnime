package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/animedrive"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// This file pins the live failure shapes found on 2026-10-16 through the FULL
// provider → adapter → client cascade. The stub serves search, details and
// player pages; only the network is fake.

type bloggerLiveRoundTripper struct{ fallback http.RoundTripper }

func (t bloggerLiveRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
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

func animeDriveLiveShapeStub(t *testing.T) *httptest.Server {
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
				<div class="result-item"><article><div class="title"><a href="/anime/recoverable/">Recoverable Show</a></div><img src="/r.jpg"/></article></div>
				<div class="result-item"><article><div class="title"><a href="/anime/deadonly/">Dead Only Show</a></div><img src="/d.jpg"/></article></div>
				<div class="result-item"><article><div class="title"><a href="/anime/embedonly/">Embed Only Show</a></div><img src="/e.jpg"/></article></div>
			</body></html>`)
			return
		case r.URL.Path == "/anime/recoverable/":
			_, _ = fmt.Fprint(w, `<html><body>
				<h1 class="entry-title">Recoverable Show</h1>
				<ul class="episodios"><li><a href="/episodio-recoverable-1/">Episode 1</a></li></ul>
			</body></html>`)
			return
		case r.URL.Path == "/anime/deadonly/":
			_, _ = fmt.Fprint(w, `<html><body>
				<h1 class="entry-title">Dead Only Show</h1>
				<ul class="episodios"><li><a href="/episodio-deadonly-1/">Episode 1</a></li></ul>
			</body></html>`)
			return
		case r.URL.Path == "/anime/embedonly/":
			_, _ = fmt.Fprint(w, `<html><body>
				<h1 class="entry-title">Embed Only Show</h1>
				<ul class="episodios"><li><a href="/episodio-embedonly-1/">Movie</a></li></ul>
			</body></html>`)
			return
		case r.URL.Path == "/episodio-recoverable-1/":
			// One Piece 1088 live shape: dead ccdn HLS + live blogger.
			_, _ = fmt.Fprintf(w, `
			<div class="animeq-player__source" data-animeq-source="0"><div><iframe src="https://www.blogger.com/video.g?token=live"></iframe></div></div>
			<div class="animeq-player__source" data-animeq-source="1"><video><source src="%s/dead.m3u8&#038;e=1"></video></div>
			<button class="animeq-player__server" data-animeq-switch="0" data-source-name="Mobile / Celular">Mobile</button>
			<button class="animeq-player__server" data-animeq-switch="1" data-source-name="FullHD / HLS">HLS</button>
			`, server.URL)
			return
		case r.URL.Path == "/episodio-deadonly-1/":
			// Boruto/Naruto live shape: Legendado+Dublado 404s, no fallback.
			_, _ = fmt.Fprintf(w, `
			<div class="animeq-player__source" data-animeq-source="0"><video><source src="%s/dead-a.mp4"></video></div>
			<div class="animeq-player__source" data-animeq-source="1"><video><source src="%s/dead-b.mp4"></video></div>
			<button class="animeq-player__server" data-animeq-switch="0" data-source-name="Legendado">Legendado</button>
			<button class="animeq-player__server" data-animeq-switch="1" data-source-name="Dublado">Dublado</button>
			`, server.URL, server.URL)
			return
		case r.URL.Path == "/episodio-embedonly-1/":
			// Kimetsu movie live shape: third-party embeds only.
			_, _ = fmt.Fprint(w, `
			<div class="animeq-player__source" data-animeq-source="0"><div><iframe src="https://animeshd.cloud/#a"></iframe></div></div>
			<div class="animeq-player__source" data-animeq-source="1"><div><iframe src="https://animes.strp2p.com/#b"></iframe></div></div>
			<div class="animeq-player__source" data-animeq-source="2"><div><iframe src="https://animes.rpmplay.me/#c"></iframe></div></div>
			<button class="animeq-player__server" data-animeq-switch="0" data-source-name="FHD">FHD</button>
			<button class="animeq-player__server" data-animeq-switch="1" data-source-name="FHD">FHD</button>
			<button class="animeq-player__server" data-animeq-switch="2" data-source-name="FHD">FHD</button>
			`)
			return
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

func liveShapeProvider(t *testing.T, server *httptest.Server) *animeDriveProvider {
	t.Helper()
	client := animedrive.NewClientForTest(server.URL)
	client.SetHTTPClientForTest(&http.Client{Transport: bloggerLiveRoundTripper{fallback: http.DefaultTransport}})
	p := &animeDriveProvider{}
	p.adapter.UnifiedScraper = &animeDriveTestAdapter{client: client}
	p.once.Do(func() {})
	return p
}

// animeDriveTestAdapter is a thin local adapter to inject a mocked client
// without importing the scraper package's unexported type.
type animeDriveTestAdapter struct {
	client *animedrive.AnimeDriveClient
}

func (a *animeDriveTestAdapter) SearchAnime(query string, _ ...any) ([]*models.Anime, error) {
	return a.client.SearchAnime(query)
}

func (a *animeDriveTestAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodes(animeURL)
}

func (a *animeDriveTestAdapter) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	quality := "best"
	for _, o := range options {
		if s, ok := o.(string); ok {
			quality = s
			break
		}
	}
	u, meta, err := a.client.GetStreamURLWithQuality(episodeURL, quality)
	if meta == nil {
		meta = map[string]string{}
	}
	meta["source"] = "animedrive"
	return u, meta, err
}

func (a *animeDriveTestAdapter) GetType() scraper.ScraperType { return scraper.AnimeDriveType }

func TestAnimeDriveCascadeRegression_SearchToStreamRecoverable(t *testing.T) {
	// Not parallel: FetchStreamURL stamps process-wide playback globals.
	server := animeDriveLiveShapeStub(t)
	p := liveShapeProvider(t, server)
	ctx := context.Background()

	results, err := p.Search(ctx, "shows")
	require.NoError(t, err)
	require.Len(t, results, 3)

	var recoverable *models.Anime
	for _, a := range results {
		if strings.Contains(a.Name, "Recoverable") {
			recoverable = a
		}
	}
	require.NotNil(t, recoverable)
	_, resolved := source.Resolve(recoverable)
	assert.Equal(t, source.AnimeDrive, resolved.Kind)

	eps, err := p.FetchEpisodes(ctx, recoverable)
	require.NoError(t, err)
	require.Len(t, eps, 1)

	stream, err := p.FetchStreamURL(ctx, &eps[0], recoverable, "best")
	require.NoError(t, err, "dead HLS must fall back to blogger across the full cascade")
	assert.Equal(t, "https://www.blogger.com/video.g?token=live", stream)
	assert.Equal(t, server.URL+"/", util.GetGlobalReferer())
	t.Cleanup(func() { util.SetGlobalReferer("") })
}

func TestAnimeDriveCascadeRegression_DeadOnlyErrorIsActionable(t *testing.T) {
	// Not parallel: asserts process-wide playback globals via FetchStreamURL.
	server := animeDriveLiveShapeStub(t)
	p := liveShapeProvider(t, server)
	ctx := context.Background()

	anime := &models.Anime{Source: "AnimeDrive", URL: server.URL + "/anime/deadonly/"}
	eps, err := p.FetchEpisodes(ctx, anime)
	require.NoError(t, err)
	require.Len(t, eps, 1)

	_, err = p.FetchStreamURL(ctx, &eps[0], anime, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the 2 servers")
	assert.Contains(t, err.Error(), "Legendado")
	assert.Contains(t, err.Error(), "Dublado")
	assert.Equal(t, 1, strings.Count(err.Error(), "AnimeDrive"),
		"provider must not duplicate the leaf prefix: %q", err.Error())
}

func TestAnimeDriveCascadeRegression_EmbedOnlyErrorListsHosts(t *testing.T) {
	// Not parallel: FetchStreamURL stamps process-wide playback globals.
	server := animeDriveLiveShapeStub(t)
	p := liveShapeProvider(t, server)
	ctx := context.Background()

	anime := &models.Anime{Source: "AnimeDrive", URL: server.URL + "/anime/embedonly/"}
	eps, err := p.FetchEpisodes(ctx, anime)
	require.NoError(t, err)
	require.Len(t, eps, 1)

	_, err = p.FetchStreamURL(ctx, &eps[0], anime, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no video tracks")
	assert.Contains(t, err.Error(), "animeshd.cloud")
	assert.Contains(t, err.Error(), "animes.strp2p.com")
	assert.Contains(t, err.Error(), "animes.rpmplay.me")
}
