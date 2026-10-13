package animedrive

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAnimeDriveSearchRetriesOnFailure(t *testing.T) {
	t.Parallel()

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempts++
		if attempts == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}

		_, _ = fmt.Fprintf(w, `
        <html>
            <body>
                <article class="item">
                    <h3><a href="/anime/naruto/">Naruto</a></h3>
                    <img src="/poster.jpg" />
                </article>
            </body>
        </html>
        `)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 2
	client.retryDelay = 0

	results, err := client.SearchAnime("naruto")
	require.NoError(t, err)
	require.Len(t, results, 1)

	assert.Equal(t, "Naruto", results[0].Name)
	assert.Contains(t, results[0].URL, "/anime/naruto/")
}

func TestAnimeDriveSearchReturnsEmptySliceWhenNoMatch(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, _ = fmt.Fprint(w, `<html><body><div class="nothing-here"></div></body></html>`)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	results, err := client.SearchAnime("unknown")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestAnimeDriveGetAnimeDetails(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `
        <html>
            <body>
                <h1 class="entry-title">Naruto Shippuden</h1>
                <div class="poster"><img src="/poster.jpg" /></div>
                <div class="wp-content"><p>A young ninja seeks recognition.</p></div>
                <ul class="episodios">
                    <li><a href="/episodio-1/">Episode 1</a></li>
                    <li><a href="/episodio-2/">Episode 2</a></li>
                    <li><a href="/episodio-3/">Episode 3</a></li>
                </ul>
            </body>
        </html>
        `)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	details, err := client.GetAnimeDetails("/anime/naruto/")
	require.NoError(t, err)
	require.NotNil(t, details)

	assert.Equal(t, "Naruto Shippuden", details.Title)
	assert.Equal(t, "A young ninja seeks recognition.", details.Synopsis)
	assert.Len(t, details.Episodes, 3)
}

func TestAnimeDriveGetAnimeDetailsParsesEpisodeCardsOnce(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `
        <html><body>
            <h1 class="entry-title">Naruto</h1>
            <div class="episode-item">
                <a href="https://animesdrive.cloud/episodio/naruto-episodio-01"><img src="poster.jpg" alt="Naruto - Episódio 1" class="episode-thumbnail"></a>
                <div class="episode-info"><span class="episode-label">Naruto</span><h3 class="episode-title"><a href="https://animesdrive.cloud/episodio/naruto-episodio-01">Naruto Uzumaki Chegando!</a></h3></div>
                <a href="https://animesdrive.cloud/episodio/naruto-episodio-01" class="episode-mini-number" hidden="" aria-label="Assistir episódio 1">1</a>
            </div>
            <div class="episode-item">
                <a href="https://animesdrive.cloud/episodio/naruto-episodio-02"><img src="poster.jpg" alt="Naruto - Episódio 2" class="episode-thumbnail"></a>
                <div class="episode-info"><span class="episode-label">Naruto</span><h3 class="episode-title"><a href="https://animesdrive.cloud/episodio/naruto-episodio-02">Os Akatsuki Entram em Ação</a></h3></div>
                <a href="https://animesdrive.cloud/episodio/naruto-episodio-02" class="episode-mini-number" hidden="" aria-label="Assistir episódio 2">2</a>
            </div>
        </body></html>`)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	details, err := client.GetAnimeDetails("/anime/naruto/")
	require.NoError(t, err)
	require.Len(t, details.Episodes, 2)

	assert.Equal(t, "1", details.Episodes[0].Number)
	assert.Equal(t, "Naruto Uzumaki Chegando!", details.Episodes[0].Title)
	assert.Equal(t, "https://animesdrive.cloud/episodio/naruto-episodio-01", details.Episodes[0].URL)
	assert.Equal(t, "2", details.Episodes[1].Number)
	assert.Equal(t, "Os Akatsuki Entram em Ação", details.Episodes[1].Title)
	assert.Equal(t, "https://animesdrive.cloud/episodio/naruto-episodio-02", details.Episodes[1].URL)
}

func TestAnimeDriveGetAnimeEpisodes(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `
        <html>
            <body>
                <h1 class="entry-title">Test Anime</h1>
                <div class="wp-content"><p>Synopsis here.</p></div>
                <ul class="episodios">
                    <li><a href="/episodio-1/">Episode 1</a></li>
                    <li><a href="/episodio-2/">Episode 2</a></li>
                </ul>
            </body>
        </html>
        `)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	episodes, err := client.GetAnimeEpisodes("/anime/test/")
	require.NoError(t, err)
	require.Len(t, episodes, 2)

	assert.Equal(t, "1", episodes[0].Number)
	assert.Equal(t, 1, episodes[0].Num)
	assert.Equal(t, "2", episodes[1].Number)
	assert.Equal(t, 2, episodes[1].Num)
}

func TestAnimeDriveGetStreamURL(t *testing.T) {
	t.Parallel()

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasPrefix(r.URL.Path, "/legendado/") || strings.HasPrefix(r.URL.Path, "/dublado/") {
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		_, _ = fmt.Fprintf(w, `
        <html>
            <body>
                <div class="animeq-player__servers" role="tablist" aria-label="Fontes de v&#237;deo">
                    <button class="animeq-player__server" data-animeq-switch="0" data-source-name="Legendado">Legendado</button>
                    <button class="animeq-player__server" data-animeq-switch="1" data-source-name="Dublado">Dublado</button>
                </div>
                <video class="animeq-player__video" data-post-id="8861">
					<source src="%s/legendado/ep1.mp4" type="video/mp4">
					<source src="%s/dublado/ep1.mp4" type="video/mp4">
                </video>
            </body>
        </html>
		`, server.URL, server.URL)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	streamURL, metadata, err := client.GetStreamURL("/episodio-1/")
	require.NoError(t, err)
	require.NotEmpty(t, streamURL)
	assert.Equal(t, server.URL+"/legendado/ep1.mp4", streamURL)
	assert.Equal(t, "animedrive", metadata["source"])
	assert.Equal(t, "Legendado", metadata["audio_label"])
	assert.Equal(t, server.URL+"/", metadata["referer"])
}

func TestAnimeDriveGetStreamURL_LegendadoOnly(t *testing.T) {
	t.Parallel()

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/legendado/ep1.mp4" {
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		_, _ = fmt.Fprintf(w, `
        <html>
            <body>
                <video class="animeq-player__video" data-post-id="8861">
					<source src="%s/legendado/ep1.mp4" type="video/mp4">
                </video>
            </body>
        </html>
		`, server.URL)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	streamURL, metadata, err := client.GetStreamURL("/episodio-1/")
	require.NoError(t, err)
	require.NotEmpty(t, streamURL)
	assert.Equal(t, server.URL+"/legendado/ep1.mp4", streamURL)
	assert.Equal(t, "Legendado", metadata["audio_label"])
}

func TestAnimeDriveGetStreamURLUsesSourceIDsAndQuality(t *testing.T) {
	t.Parallel()

	var gotReferer string
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sd.mp4", "/hd.mp4":
			gotReferer = r.Header.Get("Referer")
			w.WriteHeader(http.StatusPartialContent)
		case "/fhd.mp4":
			w.WriteHeader(http.StatusNotFound)
		default:
			_, _ = fmt.Fprintf(w, `
		<div class="animeq-player__source" data-animeq-source="0"><video><source src="%s/sd.mp4"></video></div>
		<div class="animeq-player__source" data-animeq-source="1"><div><iframe src="https://animeshd.cloud/#example"></iframe></div></div>
		<div class="animeq-player__source" data-animeq-source="2"><video><source src="%s/fhd.mp4"></video></div>
		<div class="animeq-player__source" data-animeq-source="3"><video><source src="%s/hd.mp4"></video></div>
		<button class="animeq-player__server" data-animeq-switch="0" data-source-name="SD">SD</button>
		<button class="animeq-player__server" data-animeq-switch="1" data-source-name="FHD embed">FHD embed</button>
		<button class="animeq-player__server" data-animeq-switch="2" data-source-name="FHD">FHD</button>
		<button class="animeq-player__server" data-animeq-switch="3" data-source-name="HD">HD</button>
		`, server.URL, server.URL, server.URL)
		}
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	streamURL, metadata, err := client.GetStreamURLWithQuality("/episodio-1/", "best")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/hd.mp4", streamURL, "FHD is unavailable, so HD should be selected")
	assert.Equal(t, "HD", metadata["audio_label"])
	assert.Equal(t, server.URL+"/", gotReferer)

	streamURL, metadata, err = client.GetStreamURLWithQuality("/episodio-1/", "worst")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/sd.mp4", streamURL)
	assert.Equal(t, "SD", metadata["audio_label"])
}

func TestAnimeDriveGetStreamURLAcceptsBloggerEmbed(t *testing.T) {
	t.Parallel()
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `<div class="animeq-player__source" data-animeq-source="0"><iframe src="https://www.blogger.com/video.g?token=example"></iframe></div><button class="animeq-player__server" data-animeq-switch="0" data-source-name="Mobile">Mobile</button>`)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.client = &http.Client{Transport: bloggerMockTransport{fallback: http.DefaultTransport}}
	client.maxRetries = 1

	streamURL, _, err := client.GetStreamURL("/episodio-1/")
	require.NoError(t, err)
	assert.Equal(t, "https://www.blogger.com/video.g?token=example", streamURL)
}

type bloggerMockTransport struct {
	fallback http.RoundTripper
}

func (t bloggerMockTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.URL.Host == "www.blogger.com" {
		return &http.Response{StatusCode: http.StatusPartialContent, Header: make(http.Header), Body: io.NopCloser(strings.NewReader("")), Request: req}, nil
	}
	return t.fallback.RoundTrip(req)
}

func TestAnimeDriveGetStreamURL_NoTracks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `
        <html>
            <body>
                <div>No player here</div>
            </body>
        </html>
        `)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	_, _, err := client.GetStreamURL("/episodio-1/")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no video tracks found in animeq-player")
}
