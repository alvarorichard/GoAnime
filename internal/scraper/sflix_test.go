// Package scraper provides SFlix scraper tests
package scraper

import (
	"context"
	"testing"
	"time"
)

func TestSFlixClient_Search(t *testing.T) {
	client := NewSFlixClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := client.SearchMediaWithContext(ctx, "inception")
	if err != nil {
		t.Skipf("SFlix search failed (may be blocked/unavailable): %v", err)
		return
	}

	t.Logf("Found %d results for 'inception'", len(results))
	for i, r := range results {
		if i >= 5 {
			break
		}
		t.Logf("  [%d] %s (%s) - Type: %s, ID: %s", i, r.Title, r.Year, r.Type, r.ID)
	}
}

func TestSFlixClient_GetInfo(t *testing.T) {
	client := NewSFlixClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First search to get a valid ID
	results, err := client.SearchMediaWithContext(ctx, "inception")
	if err != nil || len(results) == 0 {
		t.Skipf("SFlix search failed or no results: %v", err)
		return
	}

	// Get info for the first result
	info, err := client.GetInfoWithContext(ctx, results[0].ID)
	if err != nil {
		t.Errorf("Failed to get info: %v", err)
		return
	}

	t.Logf("Movie: %s", info.Title)
	t.Logf("  Type: %s", info.Type)
	t.Logf("  Year: %s", info.Year)
	t.Logf("  Rating: %s", info.Rating)
	t.Logf("  Description: %.100s...", info.Description)
	t.Logf("  Genres: %v", info.Genres)
	t.Logf("  Episodes: %d", len(info.Episodes))
}

func TestSFlixClient_HealthCheck(t *testing.T) {
	client := NewSFlixClient()
	ctx, cancel := context.WithTimeout(context.Background(), 15*time.Second)
	defer cancel()

	err := client.HealthCheck(ctx)
	if err != nil {
		t.Skipf("SFlix health check failed (site may be unavailable): %v", err)
		return
	}

	t.Log("SFlix health check passed")
}

func TestSFlixClient_ToAnimeModel(t *testing.T) {
	media := &SFlixMedia{
		ID:          "movie/test-movie-123",
		Title:       "Test Movie",
		Type:        MediaTypeMovie,
		Year:        "2024",
		ImageURL:    "https://example.com/poster.jpg",
		Description: "A test movie description",
		Genres:      []string{"Action", "Drama"},
	}

	anime := media.ToAnimeModel()

	if anime.Name != "Test Movie" {
		t.Errorf("Expected name 'Test Movie', got '%s'", anime.Name)
	}
	if anime.Source != "SFlix" {
		t.Errorf("Expected source 'SFlix', got '%s'", anime.Source)
	}
	if anime.Year != "2024" {
		t.Errorf("Expected year '2024', got '%s'", anime.Year)
	}
	if anime.MediaType != "movie" {
		t.Errorf("Expected media type 'movie', got '%s'", anime.MediaType)
	}

	t.Logf("ToAnimeModel conversion successful: %+v", anime)
}
