package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Unit Tests: NormalizeSuperFlixImageURL
// These test the core fix — extracting direct TMDB URLs from CloudFront wrappers
// =============================================================================

func TestNormalizeSuperFlixImageURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "CloudFront w342 URL is normalized to direct TMDB w500",
			input:    "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
			expected: "https://image.tmdb.org/t/p/w500/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
		},
		{
			name:     "CloudFront w185 URL is normalized to direct TMDB w500",
			input:    "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w185/poster123.jpg",
			expected: "https://image.tmdb.org/t/p/w500/poster123.jpg",
		},
		{
			name:     "CloudFront w154 URL is normalized to direct TMDB w500",
			input:    "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w154/tiny_poster.jpg",
			expected: "https://image.tmdb.org/t/p/w500/tiny_poster.jpg",
		},
		{
			name:     "CloudFront w500 URL is normalized to direct TMDB w500 (no size change needed)",
			input:    "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w500/already_w500.jpg",
			expected: "https://image.tmdb.org/t/p/w500/already_w500.jpg",
		},
		{
			name:     "CloudFront original size URL is normalized to direct TMDB",
			input:    "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/original/highres.jpg",
			expected: "https://image.tmdb.org/t/p/original/highres.jpg",
		},
		{
			name:     "Direct TMDB URL is returned unchanged",
			input:    "https://image.tmdb.org/t/p/w500/poster.jpg",
			expected: "https://image.tmdb.org/t/p/w500/poster.jpg",
		},
		{
			name:     "Empty string returns empty",
			input:    "",
			expected: "",
		},
		{
			name:     "Non-TMDB URL is returned unchanged",
			input:    "https://example.com/some_poster.jpg",
			expected: "https://example.com/some_poster.jpg",
		},
		{
			name:     "AniList cover URL is returned unchanged",
			input:    "https://s4.anilist.co/file/anilistcdn/media/anime/cover/large/bx1.png",
			expected: "https://s4.anilist.co/file/anilistcdn/media/anime/cover/large/bx1.png",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := NormalizeSuperFlixImageURL(tc.input)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestNormalizeSuperFlixImageURL_OriginalBug demonstrates the original bug:
// CloudFront URLs caused Discord RPC to show "?" instead of cover images
func TestNormalizeSuperFlixImageURL_OriginalBug(t *testing.T) {
	t.Parallel()

	// These are REAL URLs extracted from the SuperFlix search page for "Dexter"
	// Before the fix, these were passed directly to Discord and showed "?"
	realCloudFrontURLs := map[string]string{
		"Dexter":                  "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
		"Dexter: New Blood":       "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/v95YfP2MvoGOC7FrBD1v5nlWBHv.jpg",
		"Dexter: Pecado Original": "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/w1iM0vkMz9XVCH1ZBiU3m1288G5.jpg",
	}

	for title, cloudFrontURL := range realCloudFrontURLs {
		t.Run(title, func(t *testing.T) {
			normalized := NormalizeSuperFlixImageURL(cloudFrontURL)

			// Must NOT contain cloudfront.net — Discord can't proxy it
			assert.NotContains(t, normalized, "cloudfront.net",
				"normalized URL must not contain CloudFront domain")

			// Must start with direct TMDB URL
			assert.True(t, len(normalized) > 0 && normalized[:len("https://image.tmdb.org")] == "https://image.tmdb.org",
				"normalized URL must be a direct TMDB URL, got: %s", normalized)

			// Must use w500 for Discord display quality
			assert.Contains(t, normalized, "/w500/",
				"normalized URL must use w500 size for Discord")

			// Must preserve the poster path
			assert.True(t, len(normalized) > 40,
				"normalized URL must preserve the poster path")
		})
	}
}

// =============================================================================
// Unit Tests: parseCards with image extraction
// These verify that search result cards correctly extract and normalize cover URLs
// =============================================================================

// superflixSearchHTML returns realistic SuperFlix search HTML for testing.
// This matches the actual HTML structure observed on the live site.
func superflixSearchHTML(serverURL string) string {
	return fmt.Sprintf(`
<html>
<body>
<div class="group/card">
	<img alt="Dexter" src="https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg" />
	<h3>Dexter</h3>
	<button data-msg="Copiar TMDB" data-copy="1405">TMDB</button>
	<button data-msg="Copiar IMDB" data-copy="tt0773262">IMDB</button>
	<button data-msg="Copiar Link" data-copy="%s/serie/1405">Link</button>
	<div class="mt-3">2006 | SÉRIE</div>
</div>
<div class="group/card">
	<img alt="Inception" src="https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/qmDpIHrmpJINaRKAfWQfftjCdyi.jpg" />
	<h3>Inception</h3>
	<button data-msg="Copiar TMDB" data-copy="27205">TMDB</button>
	<button data-msg="Copiar IMDB" data-copy="tt1375666">IMDB</button>
	<button data-msg="Copiar Link" data-copy="%s/filme/27205">Link</button>
	<div class="mt-3">2010 | FILME</div>
</div>
<div class="group/card">
	<img alt="O Laboratório de Dexter" src="https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/12rxsv2if1i0TudBRFfP7WznJw0.jpg" />
	<h3>O Laboratório de Dexter</h3>
	<button data-msg="Copiar TMDB" data-copy="4229">TMDB</button>
	<button data-msg="Copiar IMDB" data-copy="tt0115157">IMDB</button>
	<button data-msg="Copiar Link" data-copy="%s/serie/4229">Link</button>
	<div class="mt-3">1996 | ANIME</div>
</div>
</body>
</html>`, serverURL, serverURL, serverURL)
}

func TestParseCards_ExtractsAndNormalizesImageURL(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/pesquisar", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, superflixSearchHTML(srv.URL))
	})

	client := NewSuperFlixClient()
	client.baseURL = srv.URL
	client.client = &http.Client{} // bypass SSRF-safe transport for localhost tests
	client.maxRetries = 0

	results, err := client.SearchMedia("Dexter")
	require.NoError(t, err)
	require.Len(t, results, 3, "should find 3 cards")

	// Card 0: Dexter (TV series)
	t.Run("Dexter TV series has normalized TMDB image", func(t *testing.T) {
		dexter := results[0]
		assert.Equal(t, "Dexter", dexter.Title)
		assert.Equal(t, "1405", dexter.TMDBID)
		assert.Equal(t, "tt0773262", dexter.IMDBID)
		assert.Equal(t, "serie", dexter.SFType)

		// THE FIX: image must be a direct TMDB URL, not CloudFront wrapper
		assert.NotContains(t, dexter.ImageURL, "cloudfront.net",
			"BUG: CloudFront URL should have been normalized")
		assert.Equal(t, "https://image.tmdb.org/t/p/w500/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
			dexter.ImageURL)
	})

	// Card 1: Inception (Movie)
	t.Run("Inception movie has normalized TMDB image", func(t *testing.T) {
		inception := results[1]
		assert.Equal(t, "Inception", inception.Title)
		assert.Equal(t, "27205", inception.TMDBID)
		assert.Equal(t, "filme", inception.SFType)

		assert.NotContains(t, inception.ImageURL, "cloudfront.net")
		assert.Equal(t, "https://image.tmdb.org/t/p/w500/qmDpIHrmpJINaRKAfWQfftjCdyi.jpg",
			inception.ImageURL)
	})

	// Card 2: Dexter's Lab (Anime)
	t.Run("Anime card has normalized TMDB image", func(t *testing.T) {
		lab := results[2]
		assert.Equal(t, "O Laboratório de Dexter", lab.Title)
		assert.Equal(t, "4229", lab.TMDBID)

		assert.NotContains(t, lab.ImageURL, "cloudfront.net")
		assert.Equal(t, "https://image.tmdb.org/t/p/w500/12rxsv2if1i0TudBRFfP7WznJw0.jpg",
			lab.ImageURL)
	})
}

// TestParseCards_NoImageFallback tests cards without images
func TestParseCards_NoImageFallback(t *testing.T) {
	t.Parallel()

	htmlNoImage := `
<html><body>
<div class="group/card">
	<h3>No Image Show</h3>
	<button data-msg="Copiar TMDB" data-copy="99999">TMDB</button>
	<button data-msg="Copiar Link" data-copy="http://example.com/serie/99999">Link</button>
	<div class="mt-3">2024 | SÉRIE</div>
</div>
</body></html>`

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/pesquisar", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, htmlNoImage)
	})

	client := NewSuperFlixClient()
	client.baseURL = srv.URL
	client.client = &http.Client{}
	client.maxRetries = 0

	results, err := client.SearchMedia("nothing")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "", results[0].ImageURL, "card without img should have empty ImageURL")
}

// TestParseCards_DataSrcFallback tests lazy-loaded images using data-src
func TestParseCards_DataSrcFallback(t *testing.T) {
	t.Parallel()

	htmlDataSrc := `
<html><body>
<div class="group/card">
	<img alt="Lazy Show" data-src="https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/lazy_poster.jpg" src="data:image/gif;base64,R0lGODlhAQABAIAAAAAAAP///yH5BAEAAAAALAAAAAABAAEAAAIBRAA7" />
	<h3>Lazy Show</h3>
	<button data-msg="Copiar TMDB" data-copy="55555">TMDB</button>
	<button data-msg="Copiar Link" data-copy="http://example.com/serie/55555">Link</button>
	<div class="mt-3">2025 | SÉRIE</div>
</div>
</body></html>`

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/pesquisar", func(w http.ResponseWriter, r *http.Request) {
		_, _ = fmt.Fprint(w, htmlDataSrc)
	})

	client := NewSuperFlixClient()
	client.baseURL = srv.URL
	client.client = &http.Client{}
	client.maxRetries = 0

	results, err := client.SearchMedia("lazy")
	require.NoError(t, err)
	require.Len(t, results, 1)

	// data: URI placeholder should be skipped, data-src should be used and normalized
	assert.NotContains(t, results[0].ImageURL, "data:image",
		"data: URI placeholder must be skipped")
	assert.NotContains(t, results[0].ImageURL, "cloudfront.net",
		"CloudFront wrapper must be normalized")
	assert.Equal(t, "https://image.tmdb.org/t/p/w500/lazy_poster.jpg", results[0].ImageURL)
}

// =============================================================================
// Unit Tests: ToAnimeModel propagates ImageURL and TMDBID
// =============================================================================

func TestToAnimeModel_ImageAndTMDBID(t *testing.T) {
	t.Parallel()

	t.Run("propagates normalized image URL to anime model", func(t *testing.T) {
		t.Parallel()
		media := &SuperFlixMedia{
			Title:    "Dexter",
			Year:     "2006",
			Type:     "SÉRIE",
			SFType:   "serie",
			TMDBID:   "1405",
			IMDBID:   "tt0773262",
			ImageURL: "https://image.tmdb.org/t/p/w500/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
		}

		anime := media.ToAnimeModel()

		assert.Equal(t, "Dexter", anime.Name)
		assert.Equal(t, "SuperFlix", anime.Source)
		assert.Equal(t, models.MediaTypeTV, anime.MediaType)
		assert.Equal(t, "https://image.tmdb.org/t/p/w500/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
			anime.ImageURL, "ImageURL must be propagated to anime model")
		assert.Equal(t, 1405, anime.TMDBID,
			"TMDBID must be parsed as int for TMDB API enrichment")
		assert.Equal(t, "tt0773262", anime.IMDBID)
	})

	t.Run("movie type is set correctly", func(t *testing.T) {
		t.Parallel()
		media := &SuperFlixMedia{
			Title:    "Inception",
			SFType:   "filme",
			TMDBID:   "27205",
			ImageURL: "https://image.tmdb.org/t/p/w500/poster.jpg",
		}

		anime := media.ToAnimeModel()
		assert.Equal(t, models.MediaTypeMovie, anime.MediaType)
		assert.Equal(t, 27205, anime.TMDBID)
	})

	t.Run("empty image URL is preserved as empty", func(t *testing.T) {
		t.Parallel()
		media := &SuperFlixMedia{
			Title:  "No Image",
			SFType: "serie",
			TMDBID: "99999",
		}

		anime := media.ToAnimeModel()
		assert.Equal(t, "", anime.ImageURL)
	})

	t.Run("invalid TMDB ID does not crash", func(t *testing.T) {
		t.Parallel()
		media := &SuperFlixMedia{
			Title:  "Bad ID",
			SFType: "serie",
			TMDBID: "not-a-number",
		}

		anime := media.ToAnimeModel()
		assert.Equal(t, 0, anime.TMDBID, "non-numeric TMDB ID should result in 0")
	})
}

// =============================================================================
// Mock Tests: SuperFlix GetVideoAPI with thumbnail normalization
// =============================================================================

// TestGetVideoAPI_ReturnsCloudFrontThumb tests that GetVideoAPI returns the raw thumb
// (normalization is applied in GetStreamURL which calls NormalizeSuperFlixImageURL on it)
func TestGetVideoAPI_ReturnsCloudFrontThumb(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/player/index.php", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]string{
			"securedLink": "https://stream.example.com/video.m3u8",
			"videoImage":  "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/movie_thumb.jpg",
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	client := NewSuperFlixClient()
	client.client = &http.Client{}

	ctx := t.Context()
	streamURL, thumbURL, err := client.GetVideoAPI(ctx, srv.URL, "abc123hash", srv.URL+"/video/abc123hash")
	require.NoError(t, err)

	assert.Equal(t, "https://stream.example.com/video.m3u8", streamURL)
	assert.Equal(t, "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/movie_thumb.jpg",
		thumbURL, "GetVideoAPI returns raw thumb; normalization happens in GetStreamURL")

	// Simulate what GetStreamURL does after GetVideoAPI
	normalizedThumb := NormalizeSuperFlixImageURL(thumbURL)
	assert.Equal(t, "https://image.tmdb.org/t/p/w500/movie_thumb.jpg", normalizedThumb,
		"NormalizeSuperFlixImageURL must strip CloudFront and upgrade to w500")
	assert.NotContains(t, normalizedThumb, "cloudfront.net",
		"BUG REGRESSION: normalized thumb must not contain CloudFront")
}

// TestGetVideoAPI_EmptyThumb verifies behavior when videoImage is empty
func TestGetVideoAPI_EmptyThumb(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/player/index.php", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := map[string]string{
			"securedLink": "https://stream.example.com/video.m3u8",
			"videoImage":  "",
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	client := NewSuperFlixClient()
	client.client = &http.Client{}

	ctx := t.Context()
	_, thumbURL, err := client.GetVideoAPI(ctx, srv.URL, "hash", srv.URL+"/")
	require.NoError(t, err)

	assert.Empty(t, thumbURL)
	assert.Empty(t, NormalizeSuperFlixImageURL(thumbURL),
		"empty thumb should remain empty after normalization")
}

// TestGetStreamURL_ThumbNormalizationFlow verifies the full thumb normalization
// flow as implemented in GetStreamURL, without actually calling the full pipeline
func TestGetStreamURL_ThumbNormalizationFlow(t *testing.T) {
	t.Parallel()

	// These represent real thumbnail URLs returned by the video API
	thumbCases := []struct {
		name     string
		rawThumb string
		expected string
	}{
		{
			name:     "CloudFront w342 thumb from real API",
			rawThumb: "https://d1muf25xaso8hp.cloudfront.net/https://image.tmdb.org/t/p/w342/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
			expected: "https://image.tmdb.org/t/p/w500/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
		},
		{
			name:     "Direct TMDB thumb (no wrapping)",
			rawThumb: "https://image.tmdb.org/t/p/w500/poster.jpg",
			expected: "https://image.tmdb.org/t/p/w500/poster.jpg",
		},
		{
			name:     "Empty thumb",
			rawThumb: "",
			expected: "",
		},
	}

	for _, tc := range thumbCases {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			// This is exactly what GetStreamURL line does: Thumb: NormalizeSuperFlixImageURL(thumbURL)
			result := NormalizeSuperFlixImageURL(tc.rawThumb)
			assert.Equal(t, tc.expected, result)
			if tc.rawThumb != "" {
				assert.NotContains(t, result, "cloudfront.net")
			}
		})
	}
}

// =============================================================================
// Integration Test: Real SuperFlix search with image extraction
// =============================================================================

func TestSuperFlix_RealSearch_ImageExtraction(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test that hits real SuperFlix API")
	}

	client := NewSuperFlixClient()
	results, err := client.SearchMedia("Dexter")
	if err != nil {
		t.Skipf("SuperFlix API unavailable: %v", err)
	}

	require.NotEmpty(t, results, "should find at least one result for 'Dexter'")

	for _, media := range results {
		t.Run(media.Title, func(t *testing.T) {
			// Every result should have an image URL
			if media.ImageURL == "" {
				t.Logf("WARNING: %s has no image URL", media.Title)
				return
			}

			// No CloudFront URLs should survive normalization
			assert.NotContains(t, media.ImageURL, "cloudfront.net",
				"CloudFront URL leaked for %s: %s", media.Title, media.ImageURL)

			// Should be a valid TMDB URL
			assert.Contains(t, media.ImageURL, "image.tmdb.org/t/p/",
				"image URL should be direct TMDB for %s", media.Title)

			// Should use w500 quality
			assert.Contains(t, media.ImageURL, "/w500/",
				"should use w500 quality for %s", media.Title)
		})
	}

	// Verify ToAnimeModel propagation for first result
	anime := results[0].ToAnimeModel()
	assert.NotEmpty(t, anime.ImageURL, "anime model should have ImageURL from search")
	assert.NotContains(t, anime.ImageURL, "cloudfront.net", "anime ImageURL must be normalized")
	assert.Equal(t, "SuperFlix", anime.Source)
	if results[0].TMDBID != "" {
		assert.Greater(t, anime.TMDBID, 0, "TMDBID should be parsed as int")
	}
}

// TestSuperFlix_RealSearch_ImageURLIsAccessible verifies the normalized URLs actually load
func TestSuperFlix_RealSearch_ImageURLIsAccessible(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test that hits real APIs")
	}

	client := NewSuperFlixClient()
	results, err := client.SearchMedia("Breaking Bad")
	if err != nil {
		t.Skipf("SuperFlix API unavailable: %v", err)
	}

	require.NotEmpty(t, results)

	// Find a result with an image
	var imageURL string
	for _, m := range results {
		if m.ImageURL != "" {
			imageURL = m.ImageURL
			break
		}
	}

	if imageURL == "" {
		t.Skip("No image URLs found in search results")
	}

	// The normalized URL should be directly accessible (HTTP 200)
	resp, err := http.Head(imageURL) //nolint:gosec // test URL from trusted TMDB CDN
	if err != nil {
		t.Skipf("Network error accessing image: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()

	assert.Equal(t, http.StatusOK, resp.StatusCode,
		"normalized image URL should be accessible: %s", imageURL)
	assert.Contains(t, resp.Header.Get("Content-Type"), "image/",
		"response should be an image")
}
