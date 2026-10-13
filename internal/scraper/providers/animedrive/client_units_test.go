package animedrive

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// Constructors, helpers, pure functions
// ---------------------------------------------------------------------------

func TestNewClientForTest_PointsAtServerAndDisablesRetries(t *testing.T) {
	t.Parallel()

	c := NewClientForTest("http://127.0.0.1:1")
	assert.Equal(t, "http://127.0.0.1:1", c.baseURL)
	assert.Equal(t, 0, c.maxRetries)
	assert.Equal(t, AnimeDriveBase, NewAnimeDriveClient().baseURL)
}

func TestDecorateRequest_SetsHeaders(t *testing.T) {
	t.Parallel()

	for _, base := range []string{"https://animesdrive.cloud", "https://animesdrive.cloud/"} {
		c := NewAnimeDriveClient()
		c.baseURL = base
		req, err := http.NewRequest(http.MethodGet, "https://animesdrive.cloud/", http.NoBody)
		require.NoError(t, err)
		c.decorateRequest(req)

		assert.Equal(t, netx.UserAgent, req.Header.Get("User-Agent"))
		assert.Contains(t, req.Header.Get("Accept"), "text/html")
		assert.Equal(t, netx.AcceptLanguage, req.Header.Get("Accept-Language"),
			"the ladder must stay paired with the shared User-Agent")
		assert.Equal(t, "https://animesdrive.cloud/", req.Header.Get("Referer"),
			"CDNs 403 without the trailing slash")
	}
}

func TestResolveURL(t *testing.T) {
	t.Parallel()

	c := NewClientForTest("https://stub.test")
	tests := []struct{ in, want string }{
		{"https://cdn.example/x.mp4", "https://cdn.example/x.mp4"},
		{"/anime/naruto", "https://stub.test/anime/naruto"},
		{"episodio-1", "https://stub.test/episodio-1"},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, c.resolveURL(tt.in))
	}
}

func TestExtractIDFromURL(t *testing.T) {
	t.Parallel()

	c := NewAnimeDriveClient()
	assert.Equal(t, "naruto-shippuden", c.extractIDFromURL("https://animesdrive.cloud/anime/naruto-shippuden"))
	assert.Equal(t, "naruto-shippuden", c.extractIDFromURL("https://animesdrive.cloud/anime/naruto-shippuden/"))
}

func TestAnimeDriveQualityRank(t *testing.T) {
	t.Parallel()

	tests := []struct {
		label string
		want  int
	}{
		{"FHD", 4},
		{"FullHD / HLS", 4},
		{"1080p", 4},
		{"4K", 4},
		{"2160p", 4},
		{"fhd", 4},
		{"HD", 3},
		{"720p", 3},
		{"hd", 3},
		{"SD", 2},
		{"480p", 2},
		{"360p", 2},
		{"Legendado", 1},
		{"Dublado", 1},
		{"Mobile / Celular", 1},
		{"", 1},
	}
	for _, tt := range tests {
		t.Run(tt.label, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, animeDriveQualityRank(tt.label))
		})
	}
}

// ---------------------------------------------------------------------------
// SearchAnime
// ---------------------------------------------------------------------------

func TestAnimeDriveSearch_NormalizesQuery(t *testing.T) {
	t.Parallel()

	var gotQuery string
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("s")
		_, _ = fmt.Fprint(w, `<html><body></body></html>`)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	_, err := client.SearchAnime("  sword-art_online--II  ")
	require.NoError(t, err)
	assert.Equal(t, "sword art online II", gotQuery)
}

func TestAnimeDriveSearch_RetryExhaustionErrors(t *testing.T) {
	t.Parallel()

	attempts := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		attempts++
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	client.maxRetries = 2
	_, err := client.SearchAnime("naruto")
	require.Error(t, err)
	assert.Equal(t, 3, attempts, "maxRetries=2 means 3 attempts total")
}

func TestAnimeDriveSearch_ChallengePageErrors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `<html><head><title>Just a moment...</title></head><body>challenge</body></html>`)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	_, err := client.SearchAnime("naruto")
	require.Error(t, err)
}

func TestAnimeDriveSearch_ResultCardVariants(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		html      string
		wantName  string
		wantURL   string
		wantImage string
	}{
		{
			name:     "dooplay result-item",
			html:     `<div class="result-item"><article><div class="title"><a href="/anime/naruto/">Naruto</a></div><img src="/n.jpg"/></article></div>`,
			wantName: "Naruto", wantURL: "/anime/naruto/", wantImage: "/n.jpg",
		},
		{
			name:     "widget w_item_a with data-src",
			html:     `<article class="w_item_a"><a href="/anime/bleach/" title="Bleach"><img data-src="/b.jpg"/></a><h3>Bleach</h3></article>`,
			wantName: "Bleach", wantURL: "/anime/bleach/", wantImage: "/b.jpg",
		},
		{
			name:     "legacy item with data-lazy-src",
			html:     `<article class="item"><a href="/anime/one-piece/"><img data-lazy-src="/o.jpg"/></a><h3>One Piece</h3></article>`,
			wantName: "One Piece", wantURL: "/anime/one-piece/", wantImage: "/o.jpg",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				_, _ = fmt.Fprintf(w, `<html><body>%s</body></html>`, tt.html)
			}))
			defer server.Close()

			client := NewClientForTest(server.URL)
			results, err := client.SearchAnime("x")
			require.NoError(t, err)
			require.Len(t, results, 1)
			assert.Equal(t, tt.wantName, results[0].Name)
			assert.Equal(t, server.URL+tt.wantURL, results[0].URL)
			assert.Equal(t, tt.wantImage, results[0].ImageURL)
		})
	}
}

func TestAnimeDriveSearch_SkipsUnusableCards(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body>
			<article class="item"><a href="/filme/x/"><img src="/x.jpg"/></a><h3>Not Anime</h3></article>
			<article class="item"><a href="/anime/empty/"><img src="/e.jpg"/></a><h3>   </h3></article>
			<article class="item"><a href="/anime/good/"><img src="/g.jpg"/></a><h3>Good</h3></article>
		</body></html>`)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	results, err := client.SearchAnime("x")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Good", results[0].Name)
}

// ---------------------------------------------------------------------------
// GetAnimeDetails
// ---------------------------------------------------------------------------

func TestAnimeDriveDetails_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	_, err := client.GetAnimeDetails("/anime/x/")
	require.Error(t, err)
}

func TestAnimeDriveDetails_ChallengePageErrors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `<html><head><title>Attention Required! | Cloudflare</title></head></html>`)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	_, err := client.GetAnimeDetails("/anime/x/")
	require.Error(t, err)
}

func TestAnimeDriveDetails_TitlePosterSynopsisFallbacks(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body>
			<div class="sheader"><div class="data"><h1>Solo Leveling</h1></div></div>
			<div class="poster"><img data-src="/poster-lazy.jpg"/></div>
			<div class="description"><p>Weak hunter gets strong.</p></div>
			<ul class="episodios"><li><a href="/episodio-solo-1/">Ep 1</a></li></ul>
		</body></html>`)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	details, err := client.GetAnimeDetails("/anime/solo-leveling/")
	require.NoError(t, err)
	assert.Equal(t, "Solo Leveling", details.Title)
	assert.Equal(t, "/poster-lazy.jpg", details.Thumbnail)
	assert.Equal(t, "Weak hunter gets strong.", details.Synopsis)
	assert.Equal(t, "solo-leveling", details.ID)
}

func TestAnimeDriveDetails_UnknownTitleFallback(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><p>no headings here</p></body></html>`)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	details, err := client.GetAnimeDetails("/anime/x/")
	require.NoError(t, err)
	assert.Equal(t, "Unknown", details.Title)
	assert.Empty(t, details.Episodes)
}

func TestAnimeDriveDetails_EpisodesDedupSortAndFilter(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body>
			<h1 class="entry-title">Test</h1>
			<ul class="episodios">
				<li><a href="/episodio-test-episodio-3/">Episode 3</a></li>
				<li><a href="/episodio-test-episodio-1/">Episode 1</a></li>
				<li><a href="/episodio-test-episodio-3/">Episode 3 dup</a></li>
				<li><a href="/episodio-test-episodio-2/">Episode 2</a></li>
				<li><a href="/filme/test/">Not an episode</a></li>
			</ul>
		</body></html>`)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	details, err := client.GetAnimeDetails("/anime/test/")
	require.NoError(t, err)
	require.Len(t, details.Episodes, 3, "dup collapsed, non-episodio link filtered")
	assert.Equal(t, "1", details.Episodes[0].Number)
	assert.Equal(t, "2", details.Episodes[1].Number)
	assert.Equal(t, "3", details.Episodes[2].Number)
}

func TestAnimeDriveDetails_EpisodeNumberFromAriaLabel(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body>
			<h1 class="entry-title">Test</h1>
			<div class="episode-item">
				<div class="episode-info"><h3 class="episode-title"><a href="https://animesdrive.cloud/episodio/naruto-episodio-07">The Fight</a></h3></div>
				<a href="https://animesdrive.cloud/episodio/naruto-episodio-07" class="episode-mini-number" aria-label="Assistir episódio 7"></a>
			</div>
		</body></html>`)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	details, err := client.GetAnimeDetails("/anime/naruto/")
	require.NoError(t, err)
	require.Len(t, details.Episodes, 1)
	assert.Equal(t, "7", details.Episodes[0].Number)
	assert.Equal(t, "The Fight", details.Episodes[0].Title)
}

// ---------------------------------------------------------------------------
// GetAnimeEpisodes
// ---------------------------------------------------------------------------

func TestAnimeDriveGetAnimeEpisodes_PropagatesDetailsError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	_, err := client.GetAnimeEpisodes("/anime/x/")
	require.Error(t, err)
}

func TestAnimeDriveGetAnimeEpisodes_NonNumericNumberKeepsZero(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body>
			<h1 class="entry-title">Movie</h1>
			<ul class="episodios"><li><a href="/episodio-movie-especial/">Especial</a></li></ul>
		</body></html>`)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	episodes, err := client.GetAnimeEpisodes("/anime/movie/")
	require.NoError(t, err)
	require.Len(t, episodes, 1)
	assert.Equal(t, 0, episodes[0].Num)
	assert.Equal(t, "Especial", episodes[0].Title.Romaji)
}

// ---------------------------------------------------------------------------
// GetStreamURL edge cases
// ---------------------------------------------------------------------------

func TestAnimeDriveGetStreamURL_DefaultsToBest(t *testing.T) {
	t.Parallel()

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".mp4") {
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		_, _ = fmt.Fprintf(w, `
		<div class="animeq-player__source" data-animeq-source="0"><video><source src="%s/sd.mp4"></video></div>
		<div class="animeq-player__source" data-animeq-source="1"><video><source src="%s/hd.mp4"></video></div>
		<button class="animeq-player__server" data-animeq-switch="0" data-source-name="SD">SD</button>
		<button class="animeq-player__server" data-animeq-switch="1" data-source-name="HD">HD</button>
		`, server.URL, server.URL)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	stream, meta, err := client.GetStreamURL("/episodio-1/")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/hd.mp4", stream)
	assert.Equal(t, "HD", meta["audio_lang"])
	assert.Equal(t, "animedrive", meta["source"])
}

func TestAnimeDriveGetStreamURL_WorstPicksLowest(t *testing.T) {
	t.Parallel()

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".mp4") {
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		_, _ = fmt.Fprintf(w, `
		<div class="animeq-player__source" data-animeq-source="0"><video><source src="%s/sd.mp4"></video></div>
		<div class="animeq-player__source" data-animeq-source="1"><video><source src="%s/fhd.mp4"></video></div>
		<button class="animeq-player__server" data-animeq-switch="0" data-source-name="SD">SD</button>
		<button class="animeq-player__server" data-animeq-switch="1" data-source-name="FHD">FHD</button>
		`, server.URL, server.URL)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	stream, meta, err := client.GetStreamURLWithQuality("/episodio-1/", "WORST")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/sd.mp4", stream)
	assert.Equal(t, "SD", meta["audio_lang"])
}

func TestAnimeDriveGetStreamURL_DirectVideoSrc(t *testing.T) {
	t.Parallel()

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".mp4") {
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		_, _ = fmt.Fprintf(w, `
		<div class="animeq-player__source" data-animeq-source="0"><video src="%s/direct.mp4"></video></div>
		<button class="animeq-player__server" data-animeq-switch="0" data-source-name="HD">HD</button>
		`, server.URL)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	stream, _, err := client.GetStreamURL("/episodio-1/")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/direct.mp4", stream)
}

func TestAnimeDriveGetStreamURL_SkipsInvalidTrackURLs(t *testing.T) {
	t.Parallel()

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".mp4") {
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		_, _ = fmt.Fprintf(w, `
		<div class="animeq-player__source" data-animeq-source="0"><video><source src="ftp://cdn.example/bad.mp4"></video></div>
		<div class="animeq-player__source" data-animeq-source="1"><video><source src=":bad-url"></video></div>
		<div class="animeq-player__source" data-animeq-source="2"><video><source src="%s/good.mp4"></video></div>
		<button class="animeq-player__server" data-animeq-switch="0" data-source-name="FHD">FHD</button>
		<button class="animeq-player__server" data-animeq-switch="1" data-source-name="HD">HD</button>
		<button class="animeq-player__server" data-animeq-switch="2" data-source-name="SD">SD</button>
		`, server.URL)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	stream, _, err := client.GetStreamURL("/episodio-1/")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/good.mp4", stream)
}

func TestAnimeDriveGetStreamURL_AllTracksInvalidReportsCount(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `
		<div class="animeq-player__source" data-animeq-source="0"><video><source src="ftp://a.example/a.mp4"></video></div>
		<div class="animeq-player__source" data-animeq-source="1"><video><source src="ftp://b.example/b.mp4"></video></div>
		<button class="animeq-player__server" data-animeq-switch="0" data-source-name="SD">SD</button>
		<button class="animeq-player__server" data-animeq-switch="1" data-source-name="HD">HD</button>
		`)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	_, _, err := client.GetStreamURL("/episodio-1/")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the 2 servers")
}

func TestAnimeDriveGetStreamURL_LegacySkipsEmptySources(t *testing.T) {
	t.Parallel()

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".mp4") {
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		_, _ = fmt.Fprintf(w, `
		<video class="animeq-player__video"><source src=""><source src="%s/only.mp4"></video>
		<button class="animeq-player__server" data-source-name="Legendado">Legendado</button>
		<button class="animeq-player__server" data-source-name="Dublado">Dublado</button>
		`, server.URL)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	stream, meta, err := client.GetStreamURL("/episodio-1/")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/only.mp4", stream)
	// The empty first <source> still occupies button slot 0, so the surviving
	// track keeps its positional label.
	assert.Equal(t, "Dublado", meta["audio_lang"])
}

func TestAnimeDriveGetStreamURL_HTTPError(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	_, _, err := client.GetStreamURL("/episodio-1/")
	require.Error(t, err)
}

func TestAnimeDriveGetStreamURL_ChallengePageErrors(t *testing.T) {
	t.Parallel()

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `<html><head><title>Just a moment...</title></head></html>`)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	_, _, err := client.GetStreamURL("/episodio-1/")
	require.Error(t, err)
}

func TestAnimeDriveGetStreamURL_ProbeSendsRangeAndReferer(t *testing.T) {
	t.Parallel()

	var gotRange, gotReferer string
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".mp4") {
			gotRange = r.Header.Get("Range")
			gotReferer = r.Header.Get("Referer")
			w.WriteHeader(http.StatusPartialContent)
			return
		}
		_, _ = fmt.Fprintf(w, `
		<div class="animeq-player__source" data-animeq-source="0"><video><source src="%s/v.mp4"></video></div>
		<button class="animeq-player__server" data-animeq-switch="0" data-source-name="HD">HD</button>
		`, server.URL)
	}))
	defer server.Close()

	// Base URL with a trailing slash must still produce a single-slash referer.
	client := NewClientForTest(server.URL + "/")
	_, meta, err := client.GetStreamURL("/episodio-1/")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/", meta["referer"])
	assert.Equal(t, "bytes=0-0", gotRange)
	assert.Equal(t, server.URL+"/", gotReferer)
}

// ---------------------------------------------------------------------------
// isAvailableStreamURL
// ---------------------------------------------------------------------------

func TestIsAvailableStreamURL_StatusCodes(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name   string
		status int
		want   bool
	}{
		{"ok", http.StatusOK, true},
		{"partial", http.StatusPartialContent, true},
		{"redirect", http.StatusFound, true},
		{"not found", http.StatusNotFound, false},
		{"forbidden", http.StatusForbidden, false},
		{"server error", http.StatusInternalServerError, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
			}))
			defer server.Close()

			client := NewClientForTest(server.URL)
			assert.Equal(t, tt.want, client.isAvailableStreamURL(server.URL+"/v.mp4"))
		})
	}
}

func TestIsAvailableStreamURL_InvalidURLAndDeadHost(t *testing.T) {
	t.Parallel()

	client := NewClientForTest("https://animesdrive.cloud")
	assert.False(t, client.isAvailableStreamURL("://missing-scheme"),
		"unparseable URL must not probe anything")

	dead := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {}))
	deadURL := dead.URL
	dead.Close()
	assert.False(t, client.isAvailableStreamURL(deadURL+"/v.mp4"),
		"connection refused must read as unavailable")
}
