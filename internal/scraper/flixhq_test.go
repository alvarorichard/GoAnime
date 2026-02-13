package scraper

import (
	"strings"
	"testing"
)

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
	if !strings.Contains(languageTag, "Movies") && !strings.Contains(languageTag, "TV") {
		t.Errorf("Language tag should mention Movies/TV: got %s", languageTag)
	}
}

func TestFlixHQClient_GetTVShowStream(t *testing.T) {
	// Skip if network is unavailable
	client := NewFlixHQClient()

	// Search for a TV show
	results, err := client.SearchMedia("dexter")
	if err != nil {
		t.Skipf("Skipping test due to network error: %v", err)
	}

	if len(results) == 0 {
		t.Skip("No search results found for 'dexter'")
		return
	}

	// Find a TV show in the results
	tvShow := findTVShowInFlixHQResults(results)
	if tvShow == nil {
		t.Skip("No TV show found in search results")
		return
	}

	t.Logf("Found TV show: %s (ID: %s)", tvShow.Title, tvShow.ID)

	// Get seasons
	seasons, err := client.GetSeasons(tvShow.ID)
	if err != nil {
		t.Fatalf("Failed to get seasons: %v", err)
	}

	if len(seasons) == 0 {
		t.Fatal("No seasons found")
	}

	t.Logf("Found %d seasons", len(seasons))

	// Get episodes for first season
	episodes, err := client.GetEpisodes(seasons[0].ID)
	if err != nil {
		t.Fatalf("Failed to get episodes: %v", err)
	}

	if len(episodes) == 0 {
		t.Fatal("No episodes found")
	}

	t.Logf("Found %d episodes in season 1", len(episodes))

	// Get server ID for first episode
	serverID, err := client.GetEpisodeServerID(episodes[0].DataID, "Vidcloud")
	if err != nil {
		t.Fatalf("Failed to get server ID: %v", err)
	}

	t.Logf("Got server ID: %s", serverID)

	// Get embed link
	embedLink, err := client.GetEmbedLink(serverID)
	if err != nil {
		t.Fatalf("Failed to get embed link: %v", err)
	}

	t.Logf("Got embed link: %s", embedLink)

	// Extract stream info (depends on third-party embed services which may change)
	streamInfo, err := client.ExtractStreamInfo(embedLink, "", "english")
	if err != nil {
		t.Skipf("Skipping stream extraction - external embed service unavailable: %v", err)
	}

	if streamInfo.VideoURL == "" {
		t.Skip("Skipping - video URL empty (external embed service may have changed)")
	}

	t.Logf("Got video URL: %s", streamInfo.VideoURL)
	t.Logf("Found %d subtitle tracks", len(streamInfo.Subtitles))
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
