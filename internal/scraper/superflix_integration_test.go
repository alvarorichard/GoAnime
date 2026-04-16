package scraper_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Integration Tests: SuperFlix end-to-end flows
// These tests use httptest servers to simulate the complete SuperFlix pipeline
// without hitting real external services.
// =============================================================================

// buildSuperFlixTestServer creates a fully functional mock SuperFlix server
// that simulates the entire streaming pipeline.
func buildSuperFlixTestServer(t *testing.T, opts testServerOpts) *httptest.Server {
	t.Helper()

	mux := http.NewServeMux()

	// Search endpoint
	mux.HandleFunc("/pesquisar", func(w http.ResponseWriter, r *http.Request) {
		query := r.URL.Query().Get("s")
		if opts.searchHTML != "" {
			w.Header().Set("Content-Type", "text/html; charset=utf-8")
			fmt.Fprint(w, opts.searchHTML)
			return
		}
		if opts.searchError != 0 {
			w.WriteHeader(opts.searchError)
			return
		}
		// Default search response
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		fmt.Fprint(w, defaultSearchHTML(query))
	})

	srv := httptest.NewServer(mux)

	// Player page endpoint (dynamic path, so register after we know srv.URL)
	mux.HandleFunc("/filme/", func(w http.ResponseWriter, r *http.Request) {
		if opts.playerPageHTML != "" {
			fmt.Fprint(w, opts.playerPageHTML)
			return
		}
		fmt.Fprint(w, defaultPlayerPageHTML("filme", srv.URL))
	})
	mux.HandleFunc("/serie/", func(w http.ResponseWriter, r *http.Request) {
		if opts.playerPageHTML != "" {
			fmt.Fprint(w, opts.playerPageHTML)
			return
		}
		fmt.Fprint(w, defaultPlayerPageHTML("serie", srv.URL))
	})

	// Bootstrap endpoint
	mux.HandleFunc("/player/bootstrap", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if opts.bootstrapJSON != "" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, opts.bootstrapJSON)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"data":{"options":[{"ID":"sv1","name":"Primary Server"}]}}`)
	})

	// Source endpoint
	mux.HandleFunc("/player/source", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if opts.sourceJSON != "" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, opts.sourceJSON)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprintf(w, `{"data":{"video_url":"%s/video/testhash"}}`, srv.URL)
	})

	// Video player page (redirect destination)
	mux.HandleFunc("/video/", func(w http.ResponseWriter, _ *http.Request) {
		if opts.videoPlayerHTML != "" {
			fmt.Fprint(w, opts.videoPlayerHTML)
			return
		}
		fmt.Fprint(w, `<html>
			<script>
				var defaultAudio = ["Portuguese","English"];
				var playerjsSubtitle = "[Portuguese]https://subs.example.com/pt.vtt,[English]https://subs.example.com/en.vtt";
			</script>
		</html>`)
	})

	// Video API endpoint
	mux.HandleFunc("/player/index.php", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			w.WriteHeader(http.StatusMethodNotAllowed)
			return
		}
		if opts.videoAPIJSON != "" {
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprint(w, opts.videoAPIJSON)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"securedLink":"https://cdn.example.com/stream.m3u8","videoImage":"https://image.tmdb.org/t/p/w500/thumb.jpg"}`)
	})

	return srv
}

type testServerOpts struct {
	searchHTML      string
	searchError     int
	playerPageHTML  string
	bootstrapJSON   string
	sourceJSON      string
	videoPlayerHTML string
	videoAPIJSON    string
}

func defaultSearchHTML(query string) string {
	return fmt.Sprintf(`<html><body>
		<div class="group/card">
			<img alt="%s Result" src="https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/test_poster.jpg" />
			<h3>%s Result</h3>
			<button data-msg="Copiar TMDB" data-copy="12345">TMDB</button>
			<button data-msg="Copiar IMDB" data-copy="tt9876543">IMDB</button>
			<button data-msg="Copiar Link" data-copy="http://example.com/filme/12345">Link</button>
			<div class="mt-3">2024 | FILME</div>
		</div>
	</body></html>`, query, query)
}

func defaultPlayerPageHTML(contentType, _ string) string {
	return fmt.Sprintf(`<html>
		<script>
			var CSRF_TOKEN = "integration_csrf";
			var PAGE_TOKEN = "integration_page_token";
			var INITIAL_CONTENT_ID = 12345;
			var CONTENT_TYPE = "%s";
		</script>
		<title>Player | Integration Test</title>
	</html>`, contentType)
}

// =============================================================================
// Integration Test: Complete search-to-stream pipeline
// =============================================================================

func TestIntegration_SuperFlix_SearchToStream(t *testing.T) {
	t.Parallel()

	srv := buildSuperFlixTestServer(t, testServerOpts{})
	defer srv.Close()

	client := scraper.NewSuperFlixClient()
	setTestClientURL(client, srv.URL)

	// Step 1: Search for media
	results, err := client.SearchMedia("integration test")
	require.NoError(t, err)
	require.NotEmpty(t, results)

	media := results[0]
	assert.NotEmpty(t, media.Title)
	assert.NotEmpty(t, media.TMDBID)
	assert.NotContains(t, media.ImageURL, "cloudfront.net", "image URL must be normalized")

	// Step 2: Get stream URL for the found media
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	result, err := client.GetStreamURL(ctx, media.SFType, media.TMDBID, "", "")
	require.NoError(t, err)

	assert.NotEmpty(t, result.StreamURL, "stream URL must not be empty")
	assert.Equal(t, "Integration Test", result.Title)
	assert.NotEmpty(t, result.Referer)
	assert.Len(t, result.Subtitles, 2)
	assert.Len(t, result.DefaultAudio, 2)
}

// =============================================================================
// Integration Test: Search → ToAnimeModel → tag flow
// =============================================================================

func TestIntegration_SuperFlix_SearchToAnimeModelFlow(t *testing.T) {
	t.Parallel()

	srv := buildSuperFlixTestServer(t, testServerOpts{})
	defer srv.Close()

	client := scraper.NewSuperFlixClient()
	setTestClientURL(client, srv.URL)

	results, err := client.SearchMedia("model flow")
	require.NoError(t, err)
	require.NotEmpty(t, results)

	// Convert to anime model
	anime := results[0].ToAnimeModel()

	assert.Equal(t, "SuperFlix", anime.Source)
	assert.NotEmpty(t, anime.Name)
	assert.NotEmpty(t, anime.URL, "URL should contain TMDB ID")
	assert.NotContains(t, anime.ImageURL, "cloudfront.net")

	// Verify TMDBID is parsed as int
	if results[0].TMDBID != "" {
		assert.Greater(t, anime.TMDBID, 0)
	}
}

// =============================================================================
// Integration Test: Series with seasons and episodes
// =============================================================================

func TestIntegration_SuperFlix_SeriesEpisodes(t *testing.T) {
	t.Parallel()

	episodesHTML := `<html><script>
		var ALL_EPISODES = {"1":[{"epi_num":"1","title":"Pilot","air_date":"2020-01-15"},{"epi_num":"2","title":"Episode 2","air_date":"2020-01-22"},{"epi_num":"3","title":"Episode 3","air_date":"2020-01-29"}],"2":[{"epi_num":"1","title":"Season 2 Premiere","air_date":"2021-03-01"},{"epi_num":"2","title":"S2E2","air_date":"2021-03-08"}]};
		var CSRF_TOKEN = "csrf";
		var PAGE_TOKEN = "page";
		var INITIAL_CONTENT_ID = 5555;
		var CONTENT_TYPE = "serie";
	</script></html>`

	srv := buildSuperFlixTestServer(t, testServerOpts{
		playerPageHTML: episodesHTML,
	})
	defer srv.Close()

	client := scraper.NewSuperFlixClient()
	setTestClientURL(client, srv.URL)

	episodes, err := client.GetEpisodes(context.Background(), "5555")
	require.NoError(t, err)
	require.NotNil(t, episodes)

	// Verify season structure
	require.Len(t, episodes["1"], 3, "Season 1 should have 3 episodes")
	require.Len(t, episodes["2"], 2, "Season 2 should have 2 episodes")

	// Verify episode data integrity
	assert.Equal(t, "Pilot", episodes["1"][0].Title)
	assert.Equal(t, json.Number("1"), episodes["1"][0].EpiNum)
	assert.Equal(t, "2020-01-15", episodes["1"][0].AirDate)

	assert.Equal(t, "Season 2 Premiere", episodes["2"][0].Title)
}

// =============================================================================
// Integration Test: Search result image pipeline
// =============================================================================

func TestIntegration_SuperFlix_ImageNormalizationPipeline(t *testing.T) {
	t.Parallel()

	htmlWithImages := `<html><body>
		<div class="group/card">
			<img alt="CloudFront Movie" src="https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/poster1.jpg" />
			<h3>CloudFront Movie</h3>
			<button data-msg="Copiar TMDB" data-copy="1">TMDB</button>
			<button data-msg="Copiar Link" data-copy="http://x.com/filme/1">Link</button>
			<div class="mt-3">2024 | FILME</div>
		</div>
		<div class="group/card">
			<img alt="Direct TMDB Movie" src="https://image.tmdb.org/t/p/w500/poster2.jpg" />
			<h3>Direct TMDB Movie</h3>
			<button data-msg="Copiar TMDB" data-copy="2">TMDB</button>
			<button data-msg="Copiar Link" data-copy="http://x.com/filme/2">Link</button>
			<div class="mt-3">2024 | FILME</div>
		</div>
		<div class="group/card">
			<img alt="No Image Movie" />
			<h3>No Image Movie</h3>
			<button data-msg="Copiar TMDB" data-copy="3">TMDB</button>
			<button data-msg="Copiar Link" data-copy="http://x.com/filme/3">Link</button>
			<div class="mt-3">2024 | FILME</div>
		</div>
	</body></html>`

	srv := buildSuperFlixTestServer(t, testServerOpts{searchHTML: htmlWithImages})
	defer srv.Close()

	client := scraper.NewSuperFlixClient()
	setTestClientURL(client, srv.URL)

	results, err := client.SearchMedia("images")
	require.NoError(t, err)
	require.Len(t, results, 3)

	// CloudFront URL normalized to direct TMDB w500
	assert.Equal(t, "https://image.tmdb.org/t/p/w500/poster1.jpg", results[0].ImageURL)
	assert.NotContains(t, results[0].ImageURL, "cloudfront.net")

	// Already direct TMDB URL preserved
	assert.Equal(t, "https://image.tmdb.org/t/p/w500/poster2.jpg", results[1].ImageURL)

	// Empty image URL preserved
	assert.Empty(t, results[2].ImageURL)

	// Verify propagation to anime models
	for _, r := range results {
		anime := r.ToAnimeModel()
		assert.Equal(t, r.ImageURL, anime.ImageURL, "ImageURL must propagate to anime model for %s", r.Title)
	}
}

// =============================================================================
// Integration Test: Concurrent searches don't interfere
// =============================================================================

func TestIntegration_SuperFlix_ConcurrentSearches(t *testing.T) {
	t.Parallel()

	srv := buildSuperFlixTestServer(t, testServerOpts{})
	defer srv.Close()

	client := scraper.NewSuperFlixClient()
	setTestClientURL(client, srv.URL)

	var wg sync.WaitGroup
	errors := make(chan error, 10)

	queries := []string{"naruto", "dexter", "breaking bad", "inception", "one piece",
		"attack on titan", "death note", "fullmetal", "dragon ball", "demon slayer"}

	for _, q := range queries {
		wg.Add(1)
		go func(query string) {
			defer wg.Done()
			results, err := client.SearchMedia(query)
			if err != nil {
				errors <- fmt.Errorf("search %q failed: %w", query, err)
				return
			}
			if len(results) == 0 {
				errors <- fmt.Errorf("search %q returned no results", query)
			}
		}(q)
	}

	wg.Wait()
	close(errors)

	for err := range errors {
		t.Error(err)
	}
}

// =============================================================================
// Integration Test: Error recovery - server returns errors mid-pipeline
// =============================================================================

func TestIntegration_SuperFlix_BootstrapError(t *testing.T) {
	t.Parallel()

	srv := buildSuperFlixTestServer(t, testServerOpts{
		bootstrapJSON: `{"error":"internal server error"}`,
	})
	defer srv.Close()

	client := scraper.NewSuperFlixClient()
	setTestClientURL(client, srv.URL)

	ctx := context.Background()
	_, err := client.GetStreamURL(ctx, "filme", "1", "", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no servers available")
}

func TestIntegration_SuperFlix_SourceURLError(t *testing.T) {
	t.Parallel()

	srv := buildSuperFlixTestServer(t, testServerOpts{
		sourceJSON: `{"data":{"video_url":""}}`,
	})
	defer srv.Close()

	client := scraper.NewSuperFlixClient()
	setTestClientURL(client, srv.URL)

	ctx := context.Background()
	_, err := client.GetStreamURL(ctx, "filme", "1", "", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no video URL")
}

func TestIntegration_SuperFlix_VideoAPIError(t *testing.T) {
	t.Parallel()

	srv := buildSuperFlixTestServer(t, testServerOpts{
		videoAPIJSON: `{"securedLink":"","videoSource":""}`,
	})
	defer srv.Close()

	client := scraper.NewSuperFlixClient()
	setTestClientURL(client, srv.URL)

	ctx := context.Background()
	_, err := client.GetStreamURL(ctx, "filme", "1", "", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no stream URL")
}

// =============================================================================
// Integration Test: Context cancellation propagates through pipeline
// =============================================================================

func TestIntegration_SuperFlix_ContextCancellation(t *testing.T) {
	t.Parallel()

	// Server that hangs on bootstrap
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/pesquisar" {
			fmt.Fprint(w, defaultSearchHTML("test"))
			return
		}
		if strings.HasPrefix(r.URL.Path, "/filme/") {
			fmt.Fprint(w, defaultPlayerPageHTML("filme", ""))
			return
		}
		if r.URL.Path == "/player/bootstrap" {
			// Hang to trigger context timeout
			time.Sleep(10 * time.Second)
			return
		}
	}))
	defer srv.Close()

	client := scraper.NewSuperFlixClient()
	setTestClientURL(client, srv.URL)

	ctx, cancel := context.WithTimeout(context.Background(), 100*time.Millisecond)
	defer cancel()

	_, err := client.GetStreamURL(ctx, "filme", "1", "", "")
	require.Error(t, err)
}

// =============================================================================
// Integration Test: Subtitle and audio metadata extraction
// =============================================================================

func TestIntegration_SuperFlix_SubtitleAndAudioMetadata(t *testing.T) {
	t.Parallel()

	videoPlayerHTML := `<html><script>
		var defaultAudio = ["Portuguese","Japanese","English"];
		var playerjsSubtitle = "[Portuguese]https://subs.example.com/pt.vtt,[English]https://subs.example.com/en.vtt,[Spanish]https://subs.example.com/es.vtt";
	</script></html>`

	srv := buildSuperFlixTestServer(t, testServerOpts{
		videoPlayerHTML: videoPlayerHTML,
	})
	defer srv.Close()

	client := scraper.NewSuperFlixClient()
	setTestClientURL(client, srv.URL)

	ctx := context.Background()
	result, err := client.GetStreamURL(ctx, "filme", "12345", "", "")
	require.NoError(t, err)

	// Verify audio languages
	require.Len(t, result.DefaultAudio, 3)
	assert.Equal(t, "Portuguese", result.DefaultAudio[0])
	assert.Equal(t, "Japanese", result.DefaultAudio[1])
	assert.Equal(t, "English", result.DefaultAudio[2])

	// Verify subtitles
	require.Len(t, result.Subtitles, 3)
	assert.Equal(t, "Portuguese", result.Subtitles[0].Lang)
	assert.Equal(t, "https://subs.example.com/pt.vtt", result.Subtitles[0].URL)
	assert.Equal(t, "English", result.Subtitles[1].Lang)
	assert.Equal(t, "Spanish", result.Subtitles[2].Lang)
}

// =============================================================================
// Integration Test: Mixed media types in search results
// =============================================================================

func TestIntegration_SuperFlix_MixedMediaTypes(t *testing.T) {
	t.Parallel()

	mixedHTML := `<html><body>
		<div class="group/card">
			<img alt="Action Movie" src="https://image.tmdb.org/t/p/w500/movie.jpg" />
			<h3>Action Movie</h3>
			<button data-msg="Copiar TMDB" data-copy="100">TMDB</button>
			<button data-msg="Copiar Link" data-copy="http://x.com/filme/100">Link</button>
			<div class="mt-3">PG-13 | 2024 | FILME</div>
		</div>
		<div class="group/card">
			<img alt="Drama Series" src="https://image.tmdb.org/t/p/w500/series.jpg" />
			<h3>Drama Series</h3>
			<button data-msg="Copiar TMDB" data-copy="200">TMDB</button>
			<button data-msg="Copiar Link" data-copy="http://x.com/serie/200">Link</button>
			<div class="mt-3">PG | 2023 | SÉRIE</div>
		</div>
		<div class="group/card">
			<img alt="Cool Anime" src="https://image.tmdb.org/t/p/w500/anime.jpg" />
			<h3>Cool Anime</h3>
			<button data-msg="Copiar TMDB" data-copy="300">TMDB</button>
			<button data-msg="Copiar Link" data-copy="http://x.com/serie/300">Link</button>
			<div class="mt-3">PG | 2022 | ANIME</div>
		</div>
		<div class="group/card">
			<img alt="K-Drama" src="https://image.tmdb.org/t/p/w500/dorama.jpg" />
			<h3>K-Drama</h3>
			<button data-msg="Copiar TMDB" data-copy="400">TMDB</button>
			<button data-msg="Copiar Link" data-copy="http://x.com/serie/400">Link</button>
			<div class="mt-3">PG | 2021 | DORAMA</div>
		</div>
	</body></html>`

	srv := buildSuperFlixTestServer(t, testServerOpts{searchHTML: mixedHTML})
	defer srv.Close()

	client := scraper.NewSuperFlixClient()
	setTestClientURL(client, srv.URL)

	results, err := client.SearchMedia("mixed")
	require.NoError(t, err)
	require.Len(t, results, 4)

	// Verify type detection
	animeModels := make([]*models.Anime, len(results))
	for i, r := range results {
		animeModels[i] = r.ToAnimeModel()
	}

	assert.Equal(t, models.MediaTypeMovie, animeModels[0].MediaType)
	assert.Equal(t, models.MediaTypeTV, animeModels[1].MediaType)
	assert.Equal(t, models.MediaTypeAnime, animeModels[2].MediaType)
	assert.Equal(t, models.MediaTypeAnime, animeModels[3].MediaType) // Dorama maps to anime

	// Verify years
	assert.Equal(t, "2024", results[0].Year)
	assert.Equal(t, "2023", results[1].Year)
	assert.Equal(t, "2022", results[2].Year)
	assert.Equal(t, "2021", results[3].Year)
}

// =============================================================================
// Integration Test: Token extraction edge cases in full pipeline
// =============================================================================

func TestIntegration_SuperFlix_EmptyPlayerPageTokens(t *testing.T) {
	t.Parallel()

	srv := buildSuperFlixTestServer(t, testServerOpts{
		playerPageHTML: `<html><body>Empty page with no tokens</body></html>`,
	})
	defer srv.Close()

	client := scraper.NewSuperFlixClient()
	setTestClientURL(client, srv.URL)

	ctx := context.Background()
	_, err := client.GetStreamURL(ctx, "filme", "1", "", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to extract tokens")
}

func TestIntegration_SuperFlix_PartialTokens(t *testing.T) {
	t.Parallel()

	// Only CSRF, missing PAGE_TOKEN
	srv := buildSuperFlixTestServer(t, testServerOpts{
		playerPageHTML: `<html><script>var CSRF_TOKEN = "only_csrf";</script></html>`,
	})
	defer srv.Close()

	client := scraper.NewSuperFlixClient()
	setTestClientURL(client, srv.URL)

	ctx := context.Background()
	_, err := client.GetStreamURL(ctx, "filme", "1", "", "")

	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to extract tokens")
}

// =============================================================================
// Integration Test: Search → Adapter → ScraperManager flow
// =============================================================================

func TestIntegration_SuperFlix_ThroughScraperManager(t *testing.T) {
	t.Parallel()

	mixedHTML := `<html><body>
		<div class="group/card">
			<img alt="Test Movie" src="https://image.tmdb.org/t/p/w500/poster.jpg" />
			<h3>Test Movie</h3>
			<button data-msg="Copiar TMDB" data-copy="999">TMDB</button>
			<button data-msg="Copiar IMDB" data-copy="tt0000999">IMDB</button>
			<button data-msg="Copiar Link" data-copy="http://x.com/filme/999">Link</button>
			<div class="mt-3">2024 | FILME</div>
		</div>
		<div class="group/card">
			<img alt="Test Anime" src="https://image.tmdb.org/t/p/w500/anime.jpg" />
			<h3>Test Anime</h3>
			<button data-msg="Copiar TMDB" data-copy="888">TMDB</button>
			<button data-msg="Copiar Link" data-copy="http://x.com/serie/888">Link</button>
			<div class="mt-3">2023 | ANIME</div>
		</div>
	</body></html>`

	srv := buildSuperFlixTestServer(t, testServerOpts{searchHTML: mixedHTML})
	defer srv.Close()

	sfClient := scraper.NewSuperFlixClient()
	setTestClientURL(sfClient, srv.URL)

	// Create the adapter manually
	adapter := scraper.NewSuperFlixAdapterWithClient(sfClient)

	// Search through the adapter interface
	results, err := adapter.SearchAnime("test")
	require.NoError(t, err)
	require.Len(t, results, 2)

	// Verify anime models
	assert.Equal(t, "Test Movie", results[0].Name)
	assert.Equal(t, models.MediaTypeMovie, results[0].MediaType)
	assert.Equal(t, "SuperFlix", results[0].Source)
	assert.Equal(t, 999, results[0].TMDBID)

	assert.Equal(t, "Test Anime", results[1].Name)
	assert.Equal(t, models.MediaTypeAnime, results[1].MediaType)
}

// =============================================================================
// Integration Test: Real SuperFlix API (skipped in CI/short mode)
// =============================================================================

func TestIntegration_RealSuperFlix_SearchAndVerify(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test that hits real SuperFlix API")
	}

	client := scraper.NewSuperFlixClient()

	queries := []string{"Breaking Bad", "Naruto", "Inception"}

	for _, query := range queries {
		t.Run(query, func(t *testing.T) {
			results, err := client.SearchMedia(query)
			if err != nil {
				t.Skipf("SuperFlix API unavailable: %v", err)
			}

			if len(results) == 0 {
				t.Skipf("No results for %q (API may have changed)", query)
			}

			for _, media := range results {
				// Title must not be empty
				assert.NotEmpty(t, media.Title, "every result must have a title")

				// TMDB ID should be present for most results
				if media.TMDBID != "" {
					// Verify it's numeric
					var tmdbID int
					_, err := fmt.Sscanf(media.TMDBID, "%d", &tmdbID)
					assert.NoError(t, err, "TMDB ID should be numeric: %s", media.TMDBID)
					assert.Greater(t, tmdbID, 0)
				}

				// Image URL must be normalized
				if media.ImageURL != "" {
					assert.NotContains(t, media.ImageURL, "cloudfront.net",
						"leaked CloudFront URL for %s", media.Title)
				}

				// SFType must be valid
				assert.True(t, media.SFType == "filme" || media.SFType == "serie",
					"invalid SFType %q for %s", media.SFType, media.Title)

				// Type must be present
				assert.NotEmpty(t, media.Type, "Type must not be empty for %s", media.Title)

				// ToAnimeModel must not panic
				anime := media.ToAnimeModel()
				assert.Equal(t, "SuperFlix", anime.Source)
			}
		})
	}
}

func TestIntegration_RealSuperFlix_GetEpisodes(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test that hits real SuperFlix API")
	}

	client := scraper.NewSuperFlixClient()

	// Dexter is a well-known series with multiple seasons
	episodes, err := client.GetEpisodes(context.Background(), "1405")
	if err != nil {
		t.Skipf("SuperFlix API unavailable: %v", err)
	}

	if episodes == nil {
		t.Skip("No episodes returned (API may have changed)")
	}

	// Should have at least season 1
	s1, exists := episodes["1"]
	require.True(t, exists, "should have season 1")
	assert.NotEmpty(t, s1, "season 1 should have episodes")

	// Verify episode structure
	for seasonNum, eps := range episodes {
		for _, ep := range eps {
			assert.NotEmpty(t, ep.EpiNum, "episode number must not be empty in season %s", seasonNum)
		}
	}
}

// =============================================================================
// Helper to set test URL on exported client
// We use a constructor helper to access unexported fields
// =============================================================================

// setTestClientURL is a helper that sets the base URL and HTTP client for testing.
// Since SuperFlixClient has unexported fields, we need a test-friendly way to configure it.
func setTestClientURL(client *scraper.SuperFlixClient, url string) {
	// Use the test helper method we'll expose
	client.SetTestConfig(url, &http.Client{Timeout: 5 * time.Second, Transport: http.DefaultTransport})
}
