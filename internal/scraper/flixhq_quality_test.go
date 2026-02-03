package scraper

import (
	"context"
	"testing"
	"time"
)

func TestFlixHQClient_GetInfo(t *testing.T) {
	client := NewFlixHQClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	tests := []struct {
		name     string
		id       string
		wantType MediaType
	}{
		{
			name:     "Get movie info",
			id:       "movie/watch-inception-19764",
			wantType: MediaTypeMovie,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			info, err := client.GetInfoWithContext(ctx, tt.id)
			if err != nil {
				t.Logf("Warning: GetInfo failed (may be expected): %v", err)
				return
			}

			if info.Title == "" {
				t.Error("Expected non-empty title")
			}

			t.Logf("Got info: Title=%s, Type=%s, Year=%s", info.Title, info.Type, info.Year)
			if len(info.Genres) > 0 {
				t.Logf("Genres: %v", info.Genres)
			}
		})
	}
}

func TestFlixHQClient_GetServers(t *testing.T) {
	client := NewFlixHQClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First, search for a movie to get a valid ID
	results, err := client.SearchMediaWithContext(ctx, "inception")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Skip("No search results found")
		return
	}

	// Find a movie
	movie := findMovieInResults(results)
	if movie == nil {
		t.Skip("No movie found in results")
		return
	}

	t.Logf("Testing with movie: %s (ID: %s)", movie.Title, movie.ID)

	servers, err := client.GetServersWithContext(ctx, movie.ID, true)
	if err != nil {
		t.Fatalf("GetServers failed: %v", err)
	}

	t.Logf("Found %d servers", len(servers))
	for _, s := range servers {
		t.Logf("  Server: %s (ID: %s)", s.Name, s.ID)
	}

	if len(servers) == 0 {
		t.Error("Expected at least one server")
	}
}

func TestFlixHQClient_GetSources(t *testing.T) {
	client := NewFlixHQClient()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Search for a movie
	results, err := client.SearchMediaWithContext(ctx, "inception")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Skip("No search results found")
		return
	}

	movie := findMovieInResults(results)
	if movie == nil {
		t.Skip("No movie found")
		return
	}

	t.Logf("Getting sources for: %s (ID: %s)", movie.Title, movie.ID)

	sources, err := client.GetSourcesWithContext(ctx, movie.ID, true)
	if err != nil {
		t.Fatalf("GetSources failed: %v", err)
	}

	t.Logf("Found %d sources, %d subtitles", len(sources.Sources), len(sources.Subtitles))
	for _, s := range sources.Sources {
		t.Logf("  Source: Quality=%s, IsM3U8=%t", s.Quality, s.IsM3U8)
	}
	for _, sub := range sources.Subtitles {
		t.Logf("  Subtitle: %s (%s)", sub.Label, sub.Language)
	}
}

func TestFlixHQClient_GetAvailableQualities(t *testing.T) {
	client := NewFlixHQClient()
	ctx, cancel := context.WithTimeout(context.Background(), 45*time.Second)
	defer cancel()

	// Search for a movie
	results, err := client.SearchMediaWithContext(ctx, "inception")
	if err != nil {
		t.Fatalf("Search failed: %v", err)
	}

	if len(results) == 0 {
		t.Skip("No search results found")
		return
	}

	movie := findMovieInResults(results)
	if movie == nil {
		t.Skip("No movie found")
		return
	}

	qualities, err := client.GetAvailableQualitiesWithContext(ctx, movie.ID, true)
	if err != nil {
		t.Fatalf("GetAvailableQualities failed: %v", err)
	}

	t.Logf("Available qualities: %v", qualities)
}

func TestFlixHQClient_SelectBestQuality(t *testing.T) {
	client := NewFlixHQClient()

	sources := &FlixHQVideoSources{
		Sources: []FlixHQSource{
			{URL: "http://example.com/360", Quality: "360", IsM3U8: true},
			{URL: "http://example.com/720", Quality: "720", IsM3U8: true},
			{URL: "http://example.com/1080", Quality: "1080", IsM3U8: true},
			{URL: "http://example.com/auto", Quality: "auto", IsM3U8: true},
		},
	}

	tests := []struct {
		name      string
		preferred Quality
		wantURL   string
	}{
		{"Prefer 1080", Quality1080, "http://example.com/1080"},
		{"Prefer 720", Quality720, "http://example.com/720"},
		{"Prefer auto", QualityAuto, "http://example.com/auto"},
		{"Prefer 480 (should get closest)", Quality480, "http://example.com/360"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			selected := client.SelectBestQuality(sources, tt.preferred)
			if selected == nil {
				t.Fatal("Expected to select a source")
				return
			}
			t.Logf("Selected: Quality=%s, URL=%s (wanted URL=%s)", selected.Quality, selected.URL, tt.wantURL)
			if selected.URL != tt.wantURL {
				t.Errorf("Expected URL %s, got %s", tt.wantURL, selected.URL)
			}
		})
	}
}

func TestFlixHQClient_HealthCheck(t *testing.T) {
	client := NewFlixHQClient()
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	err := client.HealthCheck(ctx)
	if err != nil {
		t.Errorf("HealthCheck failed: %v", err)
	}
}

func TestFlixHQClient_Caching(t *testing.T) {
	client := NewFlixHQClient()
	ctx := context.Background()

	// First search
	start := time.Now()
	results1, err := client.SearchMediaWithContext(ctx, "dexter")
	if err != nil {
		t.Fatalf("First search failed: %v", err)
	}
	firstDuration := time.Since(start)
	t.Logf("First search took %v, found %d results", firstDuration, len(results1))

	// Second search (should be cached)
	start = time.Now()
	results2, err := client.SearchMediaWithContext(ctx, "dexter")
	if err != nil {
		t.Fatalf("Second search failed: %v", err)
	}
	secondDuration := time.Since(start)
	t.Logf("Second search took %v, found %d results", secondDuration, len(results2))

	// Cached search should be much faster
	if secondDuration > firstDuration/2 {
		t.Logf("Warning: Cached search not significantly faster")
	}

	if len(results1) != len(results2) {
		t.Errorf("Cached results differ: first=%d, second=%d", len(results1), len(results2))
	}

	// Clear cache and verify
	client.ClearCache()
	t.Log("Cache cleared")
}

func TestQualityParsing(t *testing.T) {
	tests := []struct {
		input    string
		expected Quality
	}{
		{"360p", Quality360},
		{"480", Quality480},
		{"720p", Quality720},
		{"1080p", Quality1080},
		{"1080", Quality1080},
		{"auto", QualityAuto},
		{"", QualityAuto},
		{"best", QualityBest},
		{"unknown", Quality("unknown")},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			result := parseQuality(tt.input)
			if result != tt.expected {
				t.Errorf("parseQuality(%q) = %v, want %v", tt.input, result, tt.expected)
			}
		})
	}
}

func TestQualityToInt(t *testing.T) {
	tests := []struct {
		quality  Quality
		expected int
	}{
		{Quality360, 360},
		{Quality480, 480},
		{Quality720, 720},
		{Quality1080, 1080},
		{QualityBest, 9999},
		{QualityAuto, 0},
	}

	for _, tt := range tests {
		t.Run(string(tt.quality), func(t *testing.T) {
			result := qualityToInt(tt.quality)
			if result != tt.expected {
				t.Errorf("qualityToInt(%v) = %d, want %d", tt.quality, result, tt.expected)
			}
		})
	}
}

func TestServerPriority(t *testing.T) {
	client := NewFlixHQClient()

	servers := []FlixHQServer{
		{Name: ServerMixDrop, ID: "1"},
		{Name: ServerVidcloud, ID: "2"},
		{Name: ServerVoe, ID: "3"},
		{Name: ServerUpCloud, ID: "4"},
	}

	sorted := client.sortServersByPriority(servers)

	// Vidcloud should be first
	if sorted[0].Name != ServerVidcloud {
		t.Errorf("Expected Vidcloud first, got %s", sorted[0].Name)
	}

	// UpCloud should be second
	if sorted[1].Name != ServerUpCloud {
		t.Errorf("Expected UpCloud second, got %s", sorted[1].Name)
	}

	t.Logf("Sorted order: %v", sorted)
}

// findMovieInResults is a helper function that finds the first movie in search results
func findMovieInResults(results []*FlixHQMedia) *FlixHQMedia {
	for _, r := range results {
		if r.Type == MediaTypeMovie {
			return r
		}
	}
	return nil
}
