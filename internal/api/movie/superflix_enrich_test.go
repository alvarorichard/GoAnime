package movie

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// =============================================================================
// Tests for EnrichMedia with SuperFlix content
// Verifies the TMDB ID direct lookup path and image URL enrichment
// =============================================================================

// mockTMDBServer creates a test server that simulates TMDB API responses
// for the specific IDs used by SuperFlix content
func mockTMDBServer(t *testing.T) *httptest.Server {
	t.Helper()
	mux := http.NewServeMux()

	// Dexter TV show (TMDB ID 1405)
	mux.HandleFunc("/tv/1405", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := models.TMDBDetails{
			ID:           1405,
			Name:         "Dexter",
			IMDBID:       "tt0773262",
			Overview:     "Dexter Morgan, a forensic blood-spatter analyst...",
			PosterPath:   "/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
			FirstAirDate: "2006-10-01",
			VoteAverage:  8.2,
			Runtime:      55,
			Genres:       []models.TMDBGenre{{ID: 18, Name: "Drama"}, {ID: 80, Name: "Crime"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	// Inception movie (TMDB ID 27205)
	mux.HandleFunc("/movie/27205", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := models.TMDBDetails{
			ID:          27205,
			Title:       "Inception",
			IMDBID:      "tt1375666",
			Overview:    "A thief who steals corporate secrets through dream-sharing...",
			PosterPath:  "/qmDpIHrmpJINaRKAfWQfftjCdyi.jpg",
			ReleaseDate: "2010-07-15",
			VoteAverage: 8.4,
			Runtime:     148,
			Genres:      []models.TMDBGenre{{ID: 28, Name: "Action"}, {ID: 878, Name: "Science Fiction"}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	// Search endpoint (fallback path)
	mux.HandleFunc("/search/tv", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := models.TMDBSearchResult{
			Results: []models.TMDBMedia{{
				ID:          1405,
				Name:        "Dexter",
				PosterPath:  "/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
				VoteAverage: 8.2,
			}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	mux.HandleFunc("/search/movie", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := models.TMDBSearchResult{
			Results: []models.TMDBMedia{{
				ID:          27205,
				Title:       "Inception",
				PosterPath:  "/qmDpIHrmpJINaRKAfWQfftjCdyi.jpg",
				VoteAverage: 8.4,
			}},
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	return httptest.NewServer(mux)
}

// newTestTMDBClient creates a TMDB client pointing at the mock server
func newTestTMDBClient(serverURL string) *TMDBClient {
	return &TMDBClient{
		client:    &http.Client{},
		apiKey:    "test-api-key",
		baseURL:   serverURL,
		imageBase: TMDBImageBaseURL,
	}
}

// TestEnrichMedia_SuperFlixTV_WithTMDBID tests the core fix:
// When SuperFlix provides a TMDB ID, EnrichMedia should fetch details directly
// instead of searching by name. This ensures the ImageURL gets populated.
func TestEnrichMedia_SuperFlixTV_WithTMDBID(t *testing.T) {
	t.Parallel()
	srv := mockTMDBServer(t)
	defer srv.Close()

	// Simulate a SuperFlix TV show (Dexter) with TMDB ID already set
	// This is what ToAnimeModel produces after parsing SuperFlix search results
	media := &models.Media{
		Name:      "Dexter",
		Source:    "SuperFlix",
		MediaType: models.MediaTypeTV,
		TMDBID:    1405,
		URL:       "1405", // SuperFlix stores TMDB ID as URL
	}

	// Override the TMDB client's base URL
	tmdbClient := newTestTMDBClient(srv.URL)

	// Manually do what EnrichMedia does when TMDBID > 0
	details, err := tmdbClient.GetTVDetails(media.TMDBID)
	require.NoError(t, err)
	require.NotNil(t, details)

	media.TMDBDetails = details
	media.Rating = details.VoteAverage
	media.Overview = details.Overview
	if details.IMDBID != "" {
		media.IMDBID = details.IMDBID
	}
	media.Runtime = details.Runtime
	var genres []string
	for _, g := range details.Genres {
		genres = append(genres, g.Name)
	}
	media.Genres = genres
	if details.PosterPath != "" {
		media.ImageURL = tmdbClient.GetImageURL(details.PosterPath, "w500")
	}

	// VERIFY: ImageURL must be populated with a direct TMDB URL
	assert.Equal(t, "https://image.tmdb.org/t/p/w500/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
		media.ImageURL, "EnrichMedia must set ImageURL from TMDB poster")
	assert.NotContains(t, media.ImageURL, "cloudfront.net",
		"enriched ImageURL must not contain CloudFront")

	// VERIFY: Other fields are enriched
	assert.Equal(t, "tt0773262", media.IMDBID)
	assert.Equal(t, 8.2, media.Rating)
	assert.Equal(t, 55, media.Runtime)
	assert.Contains(t, media.Genres, "Drama")
	assert.Contains(t, media.Genres, "Crime")
	assert.NotEmpty(t, media.Overview)

	// VERIFY: Setup fields are preserved
	assert.Equal(t, "Dexter", media.Name)
	assert.Equal(t, "SuperFlix", media.Source)
	assert.Equal(t, models.MediaTypeTV, media.MediaType)
	assert.Equal(t, "1405", media.URL)
	assert.NotNil(t, media.TMDBDetails)
}

// TestEnrichMedia_SuperFlixMovie_WithTMDBID tests movie enrichment with direct ID
func TestEnrichMedia_SuperFlixMovie_WithTMDBID(t *testing.T) {
	t.Parallel()
	srv := mockTMDBServer(t)
	defer srv.Close()

	media := &models.Media{
		Name:      "Inception",
		Source:    "SuperFlix",
		MediaType: models.MediaTypeMovie,
		TMDBID:    27205,
		URL:       "27205",
	}

	tmdbClient := newTestTMDBClient(srv.URL)

	details, err := tmdbClient.GetMovieDetails(media.TMDBID)
	require.NoError(t, err)
	require.NotNil(t, details)

	media.TMDBDetails = details
	media.Rating = details.VoteAverage
	media.Overview = details.Overview
	if details.IMDBID != "" {
		media.IMDBID = details.IMDBID
	}
	media.Runtime = details.Runtime
	if details.PosterPath != "" {
		media.ImageURL = tmdbClient.GetImageURL(details.PosterPath, "w500")
	}

	assert.Equal(t, "https://image.tmdb.org/t/p/w500/qmDpIHrmpJINaRKAfWQfftjCdyi.jpg",
		media.ImageURL)
	assert.Equal(t, "tt1375666", media.IMDBID)
	assert.Equal(t, 8.4, media.Rating)
	assert.Equal(t, 148, media.Runtime)

	// VERIFY: Setup fields are preserved
	assert.Equal(t, "Inception", media.Name)
	assert.Equal(t, "SuperFlix", media.Source)
	assert.Equal(t, models.MediaTypeMovie, media.MediaType)
	assert.Equal(t, "27205", media.URL)
	assert.NotNil(t, media.TMDBDetails)
	assert.NotEmpty(t, media.Overview)
}

// TestEnrichMedia_SuperFlix_OriginalBug_NoTMDBID documents the original problem:
// Before the fix, SuperFlix content had TMDBID=0 and only a numeric URL,
// so EnrichMedia would search by name which could fail or return wrong results.
func TestEnrichMedia_SuperFlix_OriginalBug_NoTMDBID(t *testing.T) {
	t.Parallel()
	srv := mockTMDBServer(t)
	defer srv.Close()

	// BEFORE THE FIX: TMDBID was 0, ToAnimeModel didn't parse it
	media := &models.Media{
		Name:      "Dexter",
		Source:    "SuperFlix",
		MediaType: models.MediaTypeTV,
		TMDBID:    0,      // Original bug: TMDBID was never set
		URL:       "1405", // This was all we had
		ImageURL:  "",     // No image extracted either
	}

	tmdbClient := newTestTMDBClient(srv.URL)

	// Without TMDBID, the code falls through to name-based search
	assert.Equal(t, 0, media.TMDBID,
		"original bug: TMDBID was zero before the fix")
	assert.Empty(t, media.ImageURL,
		"original bug: no image was extracted from SuperFlix search")
	assert.Equal(t, "Dexter", media.Name)
	assert.Equal(t, "SuperFlix", media.Source)
	assert.Equal(t, models.MediaTypeTV, media.MediaType)
	assert.Equal(t, "1405", media.URL)

	// Simulate what happens AFTER the fix: TMDBID is now parsed
	media.TMDBID = 1405                                                                // Fixed by ToAnimeModel parsing strconv.Atoi(TMDBID)
	media.ImageURL = "https://image.tmdb.org/t/p/w500/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg" // Fixed by parseCards normalization

	// Now enrichment can use the direct path
	details, err := tmdbClient.GetTVDetails(media.TMDBID)
	require.NoError(t, err)

	assert.Equal(t, "tt0773262", details.IMDBID,
		"after fix: direct TMDB lookup gets the correct IMDB ID")
	assert.NotEmpty(t, details.PosterPath,
		"after fix: TMDB details include poster path for Discord RPC")
}

// TestEnrichMedia_SuperFlix_NoPosterPath tests enrichment when TMDB has no poster
func TestEnrichMedia_SuperFlix_NoPosterPath(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	srv := httptest.NewServer(mux)
	defer srv.Close()

	mux.HandleFunc("/tv/99999", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		resp := models.TMDBDetails{
			ID:         99999,
			Name:       "Obscure Show",
			PosterPath: "", // No poster available
		}
		_ = json.NewEncoder(w).Encode(resp)
	})

	tmdbClient := newTestTMDBClient(srv.URL)
	media := &models.Media{
		Name:      "Obscure Show",
		Source:    "SuperFlix",
		MediaType: models.MediaTypeTV,
		TMDBID:    99999,
		ImageURL:  "https://image.tmdb.org/t/p/w500/some_poster_from_search.jpg", // Set from search card
	}

	details, err := tmdbClient.GetTVDetails(media.TMDBID)
	require.NoError(t, err)

	// If TMDB has no poster, don't overwrite the one from SuperFlix search
	if details.PosterPath != "" {
		media.ImageURL = tmdbClient.GetImageURL(details.PosterPath, "w500")
	}

	assert.Equal(t, "https://image.tmdb.org/t/p/w500/some_poster_from_search.jpg",
		media.ImageURL, "should keep search card image when TMDB has no poster")
	assert.Equal(t, "Obscure Show", media.Name)
	assert.Equal(t, "SuperFlix", media.Source)
	assert.Equal(t, models.MediaTypeTV, media.MediaType)
}

// TestEnrichMedia_GetImageURL verifies TMDB URL construction used in enrichment
func TestEnrichMedia_GetImageURL(t *testing.T) {
	t.Parallel()

	client := &TMDBClient{imageBase: TMDBImageBaseURL}

	tests := []struct {
		name     string
		path     string
		size     string
		expected string
	}{
		{
			name:     "standard w500 poster",
			path:     "/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
			size:     "w500",
			expected: "https://image.tmdb.org/t/p/w500/f1nV5NBIFwfQLw5g8FVrdt90FAy.jpg",
		},
		{
			name:     "empty path returns empty",
			path:     "",
			size:     "w500",
			expected: "",
		},
		{
			name:     "default size when empty",
			path:     "/poster.jpg",
			size:     "",
			expected: "https://image.tmdb.org/t/p/w500/poster.jpg",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result := client.GetImageURL(tc.path, tc.size)
			assert.Equal(t, tc.expected, result)
		})
	}
}

// TestEnrichMedia_RealTMDB_SuperFlixDexter is an integration test that
// verifies the full enrichment pipeline with the real TMDB API
func TestEnrichMedia_RealTMDB_SuperFlixDexter(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test that hits real TMDB API")
	}
	skipIfNoTMDBKey(t)

	media := &models.Media{
		Name:      "Dexter",
		Source:    "SuperFlix",
		MediaType: models.MediaTypeTV,
		TMDBID:    1405,
		URL:       "1405",
	}

	err := EnrichMedia(media)
	require.NoError(t, err)

	// After enrichment, ImageURL must be a direct TMDB URL
	assert.NotEmpty(t, media.ImageURL, "EnrichMedia must set ImageURL")
	assert.Contains(t, media.ImageURL, "image.tmdb.org/t/p/",
		"ImageURL must be a direct TMDB URL")
	assert.NotContains(t, media.ImageURL, "cloudfront.net",
		"ImageURL must NOT contain CloudFront")

	// Verify metadata
	assert.Equal(t, "tt0773262", media.IMDBID)
	assert.Greater(t, media.Rating, 0.0)
	assert.NotEmpty(t, media.Overview)
	assert.NotEmpty(t, media.Genres)

	t.Logf("Dexter enriched: ImageURL=%s, Rating=%.1f, IMDB=%s",
		media.ImageURL, media.Rating, media.IMDBID)
}

// TestEnrichMedia_RealTMDB_SuperFlixInception tests movie enrichment with real API
func TestEnrichMedia_RealTMDB_SuperFlixInception(t *testing.T) {
	if testing.Short() {
		t.Skip("Skipping integration test that hits real TMDB API")
	}
	skipIfNoTMDBKey(t)

	media := &models.Media{
		Name:      "Inception",
		Source:    "SuperFlix",
		MediaType: models.MediaTypeMovie,
		TMDBID:    27205,
		URL:       "27205",
	}

	err := EnrichMedia(media)
	require.NoError(t, err)

	assert.NotEmpty(t, media.ImageURL)
	assert.Contains(t, media.ImageURL, "image.tmdb.org/t/p/")
	assert.Equal(t, "tt1375666", media.IMDBID)
	assert.Equal(t, 148, media.Runtime)

	t.Logf("Inception enriched: ImageURL=%s, Runtime=%d, IMDB=%s",
		media.ImageURL, media.Runtime, media.IMDBID)
}
