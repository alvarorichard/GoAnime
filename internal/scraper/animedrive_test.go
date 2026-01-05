package scraper

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

func TestAnimeDriveGetVideoOptions(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `
        <html>
            <body>
                <li class="dooplay_player_option" data-post="12345" data-type="tv" data-nume="1">
                    <span class="title">HD</span>
                </li>
                <li class="dooplay_player_option" data-post="12345" data-type="tv" data-nume="2">
                    <span class="title">FullHD</span>
                </li>
                <li class="dooplay_player_option" data-post="12345" data-type="tv" data-nume="3">
                    <span class="title">Mobile</span>
                </li>
            </body>
        </html>
        `)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	options, err := client.GetVideoOptions("/episodio-1/")
	require.NoError(t, err)
	require.Len(t, options, 3)

	assert.Equal(t, "HD", options[0].Label)
	assert.Equal(t, QualityHD, options[0].Quality)
	assert.Equal(t, "12345", options[0].PostID)
	assert.Equal(t, "1", options[0].Nume)

	assert.Equal(t, "FullHD", options[1].Label)
	assert.Equal(t, QualityFullHD, options[1].Quality)

	assert.Equal(t, "Mobile", options[2].Label)
	assert.Equal(t, QualityMobile, options[2].Quality)
}

func TestAnimeDriveResolveVideoURL(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Check if this is the API call
		if r.URL.Path == "/wp-json/dooplayer/v2/12345/tv/1" {
			w.Header().Set("Content-Type", "application/json")
			_, _ = fmt.Fprint(w, `{
                "embed_url": "https://player.example.com/jwplayer?source=https%3A%2F%2Ftityos.feralhosting.com%2Fvideo.mp4",
                "type": "mp4"
            }`)
			return
		}

		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	option := VideoOption{
		Label:       "HD",
		Quality:     QualityHD,
		ServerName:  "Server 1",
		ServerIndex: 0,
		PostID:      "12345",
		Type:        "tv",
		Nume:        "1",
	}

	videoURL, videoType, err := client.ResolveVideoURLWithType(option)
	require.NoError(t, err)
	assert.Equal(t, "mp4", videoType)
	assert.Equal(t, "https://tityos.feralhosting.com/video.mp4", videoURL)
}

func TestParseVideoQuality(t *testing.T) {
	t.Parallel()

	tests := []struct {
		input    string
		expected VideoQuality
	}{
		{"Mobile", QualityMobile},
		{"Celular", QualityMobile},
		{"mobile / celular", QualityMobile},
		{"SD", QualitySD},
		{"HD", QualityHD},
		{"FullHD", QualityFullHD},
		{"HLS", QualityFullHD},
		{"FHD", QualityFHD},
		{"Unknown Quality", QualityUnknown},
		{"", QualityUnknown},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := ParseVideoQuality(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnimeDriveGetAnimesByPage(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `
        <html>
            <body>
                <article class="item">
                    <a href="/anime/naruto/"><h3>Naruto</h3></a>
                    <img src="/naruto.jpg" />
                    <span class="year">2002</span>
                </article>
                <article class="item">
                    <a href="/anime/one-piece/"><h3>One Piece</h3></a>
                    <img src="/onepiece.jpg" />
                    <span class="year">1999</span>
                </article>
                <div class="pagination">
                    <a href="/page/1/">1</a>
                    <a href="/page/2/">2</a>
                    <a href="/page/371/">371</a>
                </div>
            </body>
        </html>
        `)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	results, err := client.GetAnimesByPage(1)
	require.NoError(t, err)
	require.Len(t, results, 2)

	assert.Equal(t, "Naruto", results[0].Title)
	assert.Equal(t, "2002", results[0].Year)
	assert.Equal(t, "One Piece", results[1].Title)
	assert.Equal(t, "1999", results[1].Year)
}

func TestAnimeDriveGetLatestReleases(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `
        <html>
            <body>
                <div class="items">
                    <article>
                        <a href="/anime/demon-slayer/" title="Demon Slayer">
                            <h3>Demon Slayer</h3>
                        </a>
                        <img src="/demon-slayer.jpg" />
                    </article>
                    <article>
                        <a href="/anime/jujutsu-kaisen/" title="Jujutsu Kaisen">
                            <h3>Jujutsu Kaisen</h3>
                        </a>
                        <img src="/jujutsu.jpg" />
                    </article>
                </div>
            </body>
        </html>
        `)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	results, err := client.GetLatestReleases()
	require.NoError(t, err)
	require.Len(t, results, 2)

	assert.Equal(t, "Demon Slayer", results[0].Title)
	assert.Equal(t, "Jujutsu Kaisen", results[1].Title)
}

func TestAnimeDriveGetFilms(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `
        <html>
            <body>
                <article class="item">
                    <a href="/filme/your-name/"><h3>Your Name</h3></a>
                    <img src="/yourname.jpg" />
                    <span class="rating">8.9</span>
                </article>
                <article class="item">
                    <a href="/filme/spirited-away/"><h3>Spirited Away</h3></a>
                    <img src="/spirited.jpg" />
                    <span class="rating">9.0</span>
                </article>
            </body>
        </html>
        `)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	results, err := client.GetFilms(1)
	require.NoError(t, err)
	require.Len(t, results, 2)

	assert.Equal(t, "Your Name", results[0].Title)
	assert.Equal(t, "8.9", results[0].Rating)
	assert.Equal(t, "Spirited Away", results[1].Title)
	assert.Equal(t, "9.0", results[1].Rating)
}

func TestAnimeDriveGetGenres(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, `
        <html>
            <body>
                <nav>
                    <a href="/genre/action/">Action</a>
                    <a href="/genre/comedy/">Comedy</a>
                    <a href="/genre/drama/">Drama</a>
                </nav>
            </body>
        </html>
        `)
	}))
	defer server.Close()

	client := NewAnimeDriveClient()
	client.baseURL = server.URL
	client.maxRetries = 1
	client.retryDelay = 0

	genres, err := client.GetGenres()
	require.NoError(t, err)
	require.Len(t, genres, 3)

	assert.Equal(t, "Action", genres[0].Name)
	assert.Equal(t, "action", genres[0].ID)
	assert.Equal(t, "Comedy", genres[1].Name)
	assert.Equal(t, "Drama", genres[2].Name)
}

func TestAnimeDrivePreferredDomains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		url      string
		expected bool
	}{
		{"https://tityos.feralhosting.com/video.mp4", true},
		{"https://archive.org/video.mp4", true},
		{"https://other.feralhosting.com/video.mp4", true},
		{"https://aniplay.online/video.mp4", false},
		{"https://random.com/video.mp4", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isPreferredDomain(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestAnimeDriveProblematicDomains(t *testing.T) {
	t.Parallel()

	tests := []struct {
		url      string
		expected bool
	}{
		{"https://aniplay.online/video.mp4", true},
		{"https://animeshd.cloud/video.mp4", true},
		{"https://animes.strp2p.com/video.mp4", true},
		{"https://tityos.feralhosting.com/video.mp4", false},
		{"https://random.com/video.mp4", false},
	}

	for _, tt := range tests {
		t.Run(tt.url, func(t *testing.T) {
			result := isProblematicDomain(tt.url)
			assert.Equal(t, tt.expected, result)
		})
	}
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

func TestAnimeDriveAlphabetLetters(t *testing.T) {
	t.Parallel()

	client := NewAnimeDriveClient()
	letters := client.AlphabetLetters()

	assert.Len(t, letters, 27)
	assert.Equal(t, "#", letters[0])
	assert.Equal(t, "A", letters[1])
	assert.Equal(t, "Z", letters[26])
}
