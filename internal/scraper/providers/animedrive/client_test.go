package animedrive

import (
	"fmt"
	"net/http"
	"net/http/httptest"
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

		_, _ = fmt.Fprint(w, `
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

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `
        <html>
            <body>
                <div class="animeq-player__servers" role="tablist" aria-label="Fontes de v&#237;deo">
                    <button class="animeq-player__server" data-animeq-switch="0" data-source-name="Legendado">Legendado</button>
                    <button class="animeq-player__server" data-animeq-switch="1" data-source-name="Dublado">Dublado</button>
                </div>
                <video class="animeq-player__video" data-post-id="8861">
                    <source src="https://mangas.cloud/legendado/ep1.mp4" type="video/mp4">
                    <source src="https://mangas.cloud/dublado/ep1.mp4" type="video/mp4">
                </video>
            </body>
        </html>
        `)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	streamURL, metadata, err := client.GetStreamURL("/episodio-1/")
	require.NoError(t, err)
	require.NotEmpty(t, streamURL)
	
	assert.Equal(t, "https://mangas.cloud/dublado/ep1.mp4", streamURL)
	assert.Equal(t, "animedrive", metadata["source"])
	assert.Equal(t, "Dublado", metadata["audio_label"])
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

func TestAnimeDriveGetStreamURL_noracks(t *testing.T) {
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

func TestAnimeDriveHostAssertion(t *testing.T) {

	assert.Equal(t, "https://animesdrive.cloud", AnimeDriveBase)
}