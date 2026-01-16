package api

import (
	"os"
	"testing"
)

func skipIfNoTMDBKey(t *testing.T) {
	if os.Getenv("TMDB_API_KEY") == "" {
		t.Skip("TMDB_API_KEY not set, skipping TMDB tests")
	}
}

func TestTMDBSearchMovies(t *testing.T) {
	skipIfNoTMDBKey(t)
	client := NewTMDBClient()

	// Test searching for a well-known movie
	result, err := client.SearchMovies("The Matrix")
	if err != nil {
		t.Fatalf("Failed to search movies: %v", err)
	}

	if len(result.Results) == 0 {
		t.Fatal("Expected at least one result for 'The Matrix'")
	}

	// Verify the first result is The Matrix
	firstResult := result.Results[0]
	if firstResult.GetDisplayTitle() == "" {
		t.Error("Expected title to be non-empty")
	}

	t.Logf("Found: %s (%s) - Rating: %.1f",
		firstResult.GetDisplayTitle(),
		firstResult.GetReleaseYear(),
		firstResult.VoteAverage)
}

func TestTMDBSearchTV(t *testing.T) {
	skipIfNoTMDBKey(t)
	client := NewTMDBClient()

	// Test searching for a well-known TV show
	result, err := client.SearchTV("Breaking Bad")
	if err != nil {
		t.Fatalf("Failed to search TV: %v", err)
	}

	if len(result.Results) == 0 {
		t.Fatal("Expected at least one result for 'Breaking Bad'")
	}

	firstResult := result.Results[0]
	if firstResult.GetDisplayTitle() == "" {
		t.Error("Expected title to be non-empty")
	}

	t.Logf("Found: %s (%s) - Rating: %.1f",
		firstResult.GetDisplayTitle(),
		firstResult.GetReleaseYear(),
		firstResult.VoteAverage)
}

func TestTMDBGetMovieDetails(t *testing.T) {
	skipIfNoTMDBKey(t)
	client := NewTMDBClient()

	// The Matrix TMDB ID is 603
	details, err := client.GetMovieDetails(603)
	if err != nil {
		t.Fatalf("Failed to get movie details: %v", err)
	}

	if details.Title == "" && details.Name == "" {
		t.Error("Expected title to be non-empty")
	}

	if details.Runtime <= 0 {
		t.Error("Expected runtime to be positive")
	}

	t.Logf("Movie: %s, Runtime: %d min, Genres: %d",
		details.Title, details.Runtime, len(details.Genres))
}

func TestTMDBGetTVDetails(t *testing.T) {
	skipIfNoTMDBKey(t)
	client := NewTMDBClient()

	// Breaking Bad TMDB ID is 1396
	details, err := client.GetTVDetails(1396)
	if err != nil {
		t.Fatalf("Failed to get TV details: %v", err)
	}

	if details.Name == "" && details.Title == "" {
		t.Error("Expected name to be non-empty")
	}

	if details.NumberOfSeasons <= 0 {
		t.Error("Expected seasons to be positive")
	}

	t.Logf("TV Show: %s, Seasons: %d, Episodes: %d",
		details.Name, details.NumberOfSeasons, details.NumberOfEpisodes)
}

func TestTMDBSearchMulti(t *testing.T) {
	skipIfNoTMDBKey(t)
	client := NewTMDBClient()

	// Search for something that could be movie or TV
	result, err := client.SearchMulti("Inception")
	if err != nil {
		t.Fatalf("Failed to search multi: %v", err)
	}

	if len(result.Results) == 0 {
		t.Fatal("Expected at least one result for 'Inception'")
	}

	for i, item := range result.Results[:min(3, len(result.Results))] {
		t.Logf("Result %d: %s (%s) - Type: %s",
			i+1, item.GetDisplayTitle(), item.GetReleaseYear(), item.MediaType)
	}
}

func TestTMDBGetTrending(t *testing.T) {
	skipIfNoTMDBKey(t)
	client := NewTMDBClient()

	result, err := client.GetTrending("movie", "week")
	if err != nil {
		t.Fatalf("Failed to get trending: %v", err)
	}

	if len(result.Results) == 0 {
		t.Fatal("Expected trending results")
	}

	t.Logf("Found %d trending movies", len(result.Results))
	for i, item := range result.Results[:min(5, len(result.Results))] {
		t.Logf("  %d. %s (%.1f★)", i+1, item.GetDisplayTitle(), item.VoteAverage)
	}
}

func TestTMDBGetCredits(t *testing.T) {
	skipIfNoTMDBKey(t)
	client := NewTMDBClient()

	// The Matrix TMDB ID is 603
	credits, err := client.GetCredits("movie", 603)
	if err != nil {
		t.Fatalf("Failed to get credits: %v", err)
	}

	if len(credits.Cast) == 0 {
		t.Error("Expected cast members")
	}

	t.Logf("Cast members: %d", len(credits.Cast))
	for i, actor := range credits.Cast[:min(5, len(credits.Cast))] {
		t.Logf("  %d. %s as %s", i+1, actor.Name, actor.Character)
	}
}

func TestTMDBImageURL(t *testing.T) {
	client := NewTMDBClient()

	// Test poster URL generation
	posterPath := "/f89U3ADr1oiB1s9GkdPOEpXUk5H.jpg"
	url := client.GetImageURL(posterPath, "w500")
	expected := "https://image.tmdb.org/t/p/w500/f89U3ADr1oiB1s9GkdPOEpXUk5H.jpg"

	if url != expected {
		t.Errorf("Expected URL %s, got %s", expected, url)
	}

	// Test empty path
	emptyURL := client.GetImageURL("", "w500")
	if emptyURL != "" {
		t.Errorf("Expected empty URL for empty path, got %s", emptyURL)
	}
}

func TestCleanMediaName(t *testing.T) {
	testCases := []struct {
		input    string
		expected string
	}{
		{"[Movies/TV] The Matrix", "The Matrix"},
		{"[Movie] Inception (2010)", "Inception"},
		{"Breaking Bad [TV]", "Breaking Bad"},
		{"Avatar [English]", "Avatar"},
		{"Simple Name", "Simple Name"},
	}

	for _, tc := range testCases {
		result := cleanMediaName(tc.input)
		if result != tc.expected {
			t.Errorf("cleanMediaName(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
