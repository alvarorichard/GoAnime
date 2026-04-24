package scraper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// newFlixHQClientForTest creates a FlixHQClient pointing at a test server,
// bypassing the real FlixHQ endpoints.
func newFlixHQClientForTest(baseURL string) *FlixHQClient {
	return &FlixHQClient{
		client:         &http.Client{},
		baseURL:        baseURL,
		apiURL:         baseURL,
		fallbackAPIURL: baseURL,
		userAgent:      FlixHQUserAgent,
		maxRetries:     0,
	}
}

func TestFlixHQClient_SearchMedia(t *testing.T) {
	client := NewFlixHQClient()

	tests := []struct {
		name    string
		query   string
		wantErr bool
	}{
		{
			name:    "Search for popular movie",
			query:   "avengers",
			wantErr: false,
		},
		{
			name:    "Search for TV show",
			query:   "breaking bad",
			wantErr: false,
		},
		{
			name:    "Search with special characters",
			query:   "spider-man",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			results, err := client.SearchMedia(tt.query)
			if err != nil && isFlixHQUnavailable(err) {
				t.Skipf("Skipping - external service unavailable: %v", err)
			}
			if (err != nil) != tt.wantErr {
				t.Errorf("SearchMedia() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && len(results) == 0 {
				t.Logf("Warning: No results found for query: %s (this may be expected)", tt.query)
			}

			for _, media := range results {
				if media.ID == "" {
					t.Errorf("Media ID is empty for: %s", media.Title)
				}
				if media.Title == "" {
					t.Errorf("Media title is empty for ID: %s", media.ID)
				}
				if media.Type != MediaTypeMovie && media.Type != MediaTypeTV {
					t.Errorf("Invalid media type: %s", media.Type)
				}
			}
		})
	}
}

func TestFlixHQClient_MediaTypeDetection(t *testing.T) {
	client := NewFlixHQClient()

	// Search for something that should return both movies and TV shows
	results, err := client.SearchMedia("batman")
	if err != nil {
		t.Skipf("Skipping test due to error: %v", err)
	}

	hasMovies := false
	hasTV := false

	for _, media := range results {
		if media.Type == MediaTypeMovie {
			hasMovies = true
		}
		if media.Type == MediaTypeTV {
			hasTV = true
		}
	}

	t.Logf("Search results: %d total, hasMovies=%v, hasTV=%v", len(results), hasMovies, hasTV)
}

func TestFlixHQClient_ToAnimeModel(t *testing.T) {
	media := &FlixHQMedia{
		ID:       "12345",
		Title:    "Test Movie",
		Type:     MediaTypeMovie,
		Year:     "2024",
		ImageURL: "https://example.com/image.jpg",
		URL:      "https://flixhq.to/movie/test-movie-12345",
	}

	anime := media.ToAnimeModel()

	if anime.Name != media.Title {
		t.Errorf("Name mismatch: got %s, want %s", anime.Name, media.Title)
	}
	if anime.URL != media.URL {
		t.Errorf("URL mismatch: got %s, want %s", anime.URL, media.URL)
	}
	if anime.ImageURL != media.ImageURL {
		t.Errorf("ImageURL mismatch: got %s, want %s", anime.ImageURL, media.ImageURL)
	}
	if anime.Source != "FlixHQ" {
		t.Errorf("Source mismatch: got %s, want FlixHQ", anime.Source)
	}
}

func TestFlixHQClient_ExtractLanguageFromLabel(t *testing.T) {
	client := NewFlixHQClient()

	tests := []struct {
		label    string
		expected string
	}{
		{"English", "en"},
		{"Spanish", "es"},
		{"Portuguese (Brazil)", "pt"},
		{"French", "fr"},
		{"Unknown Language", "unknown language"},
	}

	for _, tt := range tests {
		result := client.extractLanguageFromLabel(tt.label)
		if result != tt.expected {
			t.Errorf("extractLanguageFromLabel(%s) = %s, want %s", tt.label, result, tt.expected)
		}
	}
}

func TestMediaManager_Creation(t *testing.T) {
	mm := NewMediaManager()

	if mm.scraperManager == nil {
		t.Error("ScraperManager should not be nil")
	}
	if mm.flixhqClient == nil {
		t.Error("FlixHQClient should not be nil")
	}

	// Check that FlixHQ scraper is registered
	scraper, err := mm.scraperManager.GetScraper(FlixHQType)
	if err != nil {
		t.Errorf("FlixHQ scraper should be registered: %v", err)
	}
	if scraper == nil {
		t.Error("FlixHQ scraper should not be nil")
	}
}

func TestFlixHQEpisode_ToEpisodeModel(t *testing.T) {
	episode := FlixHQEpisode{
		ID:       "ep123",
		DataID:   "data456",
		Title:    "Pilot",
		Number:   1,
		SeasonID: "season1",
	}

	model := episode.ToEpisodeModel()

	if model.Num != episode.Number {
		t.Errorf("Number mismatch: got %d, want %d", model.Num, episode.Number)
	}
	if model.Title.English != episode.Title {
		t.Errorf("Title mismatch: got %s, want %s", model.Title.English, episode.Title)
	}
}

func TestConvertFlixHQToAnime(t *testing.T) {
	media := []*FlixHQMedia{
		{
			ID:    "1",
			Title: "Movie 1",
			Type:  MediaTypeMovie,
		},
		{
			ID:    "2",
			Title: "TV Show 1",
			Type:  MediaTypeTV,
		},
	}

	animes := ConvertFlixHQToAnime(media)

	if len(animes) != len(media) {
		t.Errorf("Length mismatch: got %d, want %d", len(animes), len(media))
	}

	for i, anime := range animes {
		if anime.Name != media[i].Title {
			t.Errorf("Title mismatch at %d: got %s, want %s", i, anime.Name, media[i].Title)
		}
	}
}

func TestFlixHQAdapter_GetType(t *testing.T) {
	adapter := &FlixHQAdapter{client: NewFlixHQClient()}

	if adapter.GetType() != FlixHQType {
		t.Errorf("GetType() = %v, want %v", adapter.GetType(), FlixHQType)
	}
}

func TestFlixHQClient_ResolveURL(t *testing.T) {
	client := NewFlixHQClient()

	tests := []struct {
		ref      string
		expected string
	}{
		{"/movie/test-123", FlixHQBase + "/movie/test-123"},
		{"movie/test-123", FlixHQBase + "/movie/test-123"},
		{"https://example.com/test", "https://example.com/test"},
	}

	for _, tt := range tests {
		result := client.resolveURL(tt.ref)
		if result != tt.expected {
			t.Errorf("resolveURL(%s) = %s, want %s", tt.ref, result, tt.expected)
		}
	}
}

func TestUnifiedScraperManager_FlixHQIntegration(t *testing.T) {
	sm := NewScraperManager()

	// Verify FlixHQ is registered
	displayName := sm.getScraperDisplayName(FlixHQType)
	if displayName != "FlixHQ" {
		t.Errorf("Display name = %s, want FlixHQ", displayName)
	}

	languageTag := sm.getLanguageTag(FlixHQType)
	if languageTag != "[English]" {
		t.Errorf("Language tag should be [English]: got %s", languageTag)
	}
}

// TestFlixHQClient_GetTVShowStream verifies the full TV show scraping flow
// (search → seasons → episodes → server ID → embed link) using a local
// httptest server instead of the real FlixHQ API.
func TestFlixHQClient_GetTVShowStream(t *testing.T) {
	mux := http.NewServeMux()

	// Search results: one TV show item with the expected HTML structure.
	mux.HandleFunc("/search/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body>
			<div class="flw-item">
				<div class="film-name">
					<a href="/tv/watch-dexter-39448" title="Dexter">Dexter</a>
				</div>
				<span class="fdi-item">2006</span>
			</div>
		</body></html>`)
	})

	// Seasons for media ID 39448.
	mux.HandleFunc("/ajax/v2/tv/seasons/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body>
			<a data-id="s1id" href="javascript:;">Season 1</a>
		</body></html>`)
	})

	// Episodes for season s1id.
	mux.HandleFunc("/ajax/v2/season/episodes/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body>
			<div class="nav-item"><a data-id="ep1data" title="Pilot">Ep 1</a></div>
		</body></html>`)
	})

	// Available streaming servers for episode ep1data.
	mux.HandleFunc("/ajax/v2/episode/servers/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		fmt.Fprint(w, `<html><body>
			<div class="nav-item"><a data-id="srv1id" title="Vidcloud">Vidcloud</a></div>
		</body></html>`)
	})

	// Embed link JSON for server srv1id.
	mux.HandleFunc("/ajax/episode/sources/", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		fmt.Fprint(w, `{"link": "https://example.com/embed/mock123"}`)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := newFlixHQClientForTest(srv.URL)

	// Step 1: search.
	results, err := client.SearchMedia("dexter")
	require.NoError(t, err)
	require.NotEmpty(t, results, "search should return at least one result")

	tvShow := findTVShowInFlixHQResults(results)
	require.NotNil(t, tvShow, "search should return a TV show")
	assert.Equal(t, "Dexter", tvShow.Title)
	assert.Equal(t, MediaTypeTV, tvShow.Type)
	assert.Equal(t, "39448", tvShow.ID)

	// Step 2: seasons.
	seasons, err := client.GetSeasons(tvShow.ID)
	require.NoError(t, err)
	require.NotEmpty(t, seasons, "GetSeasons should return at least one season")
	assert.Equal(t, "s1id", seasons[0].ID)
	assert.Equal(t, "Season 1", seasons[0].Title)

	// Step 3: episodes.
	episodes, err := client.GetEpisodes(seasons[0].ID)
	require.NoError(t, err)
	require.NotEmpty(t, episodes, "GetEpisodes should return at least one episode")
	assert.Equal(t, "ep1data", episodes[0].DataID)
	assert.Equal(t, "Pilot", episodes[0].Title)

	// Step 4: server ID.
	serverID, err := client.GetEpisodeServerID(episodes[0].DataID, "Vidcloud")
	require.NoError(t, err)
	assert.Equal(t, "srv1id", serverID)

	// Step 5: embed link.
	embedLink, err := client.GetEmbedLink(serverID)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/embed/mock123", embedLink)
}

// findTVShowInFlixHQResults finds the first TV show in FlixHQ search results
func findTVShowInFlixHQResults(results []*FlixHQMedia) *FlixHQMedia {
	for _, media := range results {
		if media.Type == MediaTypeTV {
			return media
		}
	}
	return nil
}
