package movie

import (
	"testing"
)

func TestOMDbSearchByTitle(t *testing.T) {
	client := NewOMDbClient()

	result, err := client.SearchByTitle("The Matrix", "movie")
	if err != nil {
		t.Fatalf("Failed to search: %v", err)
	}

	if len(result.Search) == 0 {
		t.Fatal("Expected at least one result for 'The Matrix'")
	}

	firstResult := result.Search[0]
	if firstResult.Title == "" {
		t.Error("Expected title to be non-empty")
	}

	t.Logf("Found: %s (%s) - IMDB: %s", firstResult.Title, firstResult.Year, firstResult.IMDBID)
}

func TestOMDbGetByIMDBID(t *testing.T) {
	client := NewOMDbClient()

	media, err := client.GetByIMDBID("tt0133093")
	if err != nil {
		t.Fatalf("Failed to get by IMDB ID: %v", err)
	}

	if media.Title == "" {
		t.Error("Expected title to be non-empty")
	}

	if media.GetRuntimeMinutes() <= 0 {
		t.Error("Expected runtime to be positive")
	}

	t.Logf("Movie: %s, Runtime: %d min, Rating: %.1f, Genres: %v",
		media.Title, media.GetRuntimeMinutes(), media.GetRating(), media.GetGenres())
}

func TestOMDbGetByTitle(t *testing.T) {
	client := NewOMDbClient()

	media, err := client.GetByTitle("The Matrix", "1999")
	if err != nil {
		t.Fatalf("Failed to get by title: %v", err)
	}

	if media.Title == "" {
		t.Error("Expected title to be non-empty")
	}

	if media.IMDBID != "tt0133093" {
		t.Errorf("Expected IMDB ID tt0133093, got %s", media.IMDBID)
	}

	t.Logf("Movie: %s (%s) - Director: %s", media.Title, media.Year, media.Director)
}

func TestOMDbRuntimeParsing(t *testing.T) {
	media := &OMDbMedia{Runtime: "136 min"}
	if media.GetRuntimeMinutes() != 136 {
		t.Errorf("Expected 136 minutes, got %d", media.GetRuntimeMinutes())
	}

	media2 := &OMDbMedia{Runtime: "N/A"}
	if media2.GetRuntimeMinutes() != 0 {
		t.Errorf("Expected 0 minutes for N/A, got %d", media2.GetRuntimeMinutes())
	}
}

func TestOMDbRatingParsing(t *testing.T) {
	media := &OMDbMedia{IMDBRating: "8.7"}
	if media.GetRating() != 8.7 {
		t.Errorf("Expected 8.7, got %f", media.GetRating())
	}

	media2 := &OMDbMedia{IMDBRating: "N/A"}
	if media2.GetRating() != 0 {
		t.Errorf("Expected 0 for N/A, got %f", media2.GetRating())
	}
}

func TestOMDbGenreParsing(t *testing.T) {
	media := &OMDbMedia{Genre: "Action, Sci-Fi, Drama"}
	genres := media.GetGenres()

	if len(genres) != 3 {
		t.Errorf("Expected 3 genres, got %d", len(genres))
	}

	if genres[0] != "Action" || genres[1] != "Sci-Fi" || genres[2] != "Drama" {
		t.Errorf("Unexpected genres: %v", genres)
	}

	media2 := &OMDbMedia{Genre: "N/A"}
	if len(media2.GetGenres()) != 0 {
		t.Error("Expected empty genres for N/A")
	}
}
