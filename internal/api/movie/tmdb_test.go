package movie

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

	result, err := client.SearchMovies("The Matrix")
	if err != nil {
		t.Fatalf("Failed to search movies: %v", err)
	}

	if len(result.Results) == 0 {
		t.Fatal("Expected at least one result for 'The Matrix'")
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

func TestTMDBSearchTV(t *testing.T) {
	skipIfNoTMDBKey(t)
	client := NewTMDBClient()

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

func TestTMDBImageURL(t *testing.T) {
	client := NewTMDBClient()

	posterPath := "/f89U3ADr1oiB1s9GkdPOEpXUk5H.jpg"
	url := client.GetImageURL(posterPath, "w500")
	expected := "https://image.tmdb.org/t/p/w500/f89U3ADr1oiB1s9GkdPOEpXUk5H.jpg"

	if url != expected {
		t.Errorf("Expected URL %s, got %s", expected, url)
	}

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
		result := CleanMediaName(tc.input)
		if result != tc.expected {
			t.Errorf("CleanMediaName(%q) = %q, expected %q", tc.input, result, tc.expected)
		}
	}
}
