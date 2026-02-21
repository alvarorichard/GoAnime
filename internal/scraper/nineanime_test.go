// Package scraper provides 9anime scraper tests
package scraper

import (
	"context"
	"testing"
	"time"
)

func TestNineAnimeClient_Search(t *testing.T) {
	client := NewNineAnimeClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	results, err := client.SearchAnimeWithContext(ctx, "naruto")
	if err != nil {
		t.Skipf("9anime search failed (may be blocked/unavailable): %v", err)
		return
	}

	if len(results) == 0 {
		t.Skip("No search results returned (site may be unavailable)")
		return
	}

	t.Logf("Found %d results for 'naruto'", len(results))
	for i, r := range results {
		if i >= 5 {
			break
		}
		t.Logf("  [%d] %s - Source: %s, URL: %s", i, r.Name, r.Source, r.URL)
	}
}

func TestNineAnimeClient_GetEpisodes(t *testing.T) {
	client := NewNineAnimeClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// First search to get a valid anime ID
	results, err := client.SearchAnimeWithContext(ctx, "one piece")
	if err != nil || len(results) == 0 {
		t.Skipf("9anime search failed or no results: %v", err)
		return
	}

	// The URL field stores the anime ID
	animeID := results[0].URL
	t.Logf("Getting episodes for anime ID: %s (%s)", animeID, results[0].Name)

	episodes, err := client.GetEpisodesWithContext(ctx, animeID)
	if err != nil {
		t.Skipf("Failed to get episodes (site may be unavailable): %v", err)
		return
	}

	t.Logf("Found %d episodes", len(episodes))
	for i, ep := range episodes {
		if i >= 5 {
			break
		}
		t.Logf("  Ep %d: %s (ID: %s)", ep.Number, ep.Title, ep.EpisodeID)
	}
}

func TestNineAnimeClient_GetServers(t *testing.T) {
	client := NewNineAnimeClient()
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Search → get episodes → get servers
	results, err := client.SearchAnimeWithContext(ctx, "naruto")
	if err != nil || len(results) == 0 {
		t.Skipf("9anime search failed or no results: %v", err)
		return
	}

	animeID := results[0].URL
	episodes, err := client.GetEpisodesWithContext(ctx, animeID)
	if err != nil || len(episodes) == 0 {
		t.Skipf("Failed to get episodes: %v", err)
		return
	}

	// Get servers for the first episode
	servers, err := client.GetServersWithContext(ctx, episodes[0].EpisodeID)
	if err != nil {
		t.Skipf("Failed to get servers: %v", err)
		return
	}

	t.Logf("Found %d servers for episode %d", len(servers), episodes[0].Number)
	for i, s := range servers {
		t.Logf("  [%d] %s (ID: %s, Type: %s)", i, s.Name, s.DataID, s.AudioType)
	}
}

func TestNineAnimeClient_GetAnimeEpisodes(t *testing.T) {
	client := NewNineAnimeClient()

	// Test the models.Episode conversion
	results, err := client.SearchAnime("naruto")
	if err != nil || len(results) == 0 {
		t.Skipf("9anime search failed or no results: %v", err)
		return
	}

	animeID := results[0].URL
	modelEpisodes, err := client.GetAnimeEpisodes(animeID)
	if err != nil {
		t.Skipf("Failed to get anime episodes: %v", err)
		return
	}

	if len(modelEpisodes) == 0 {
		t.Skip("No episodes returned")
		return
	}

	t.Logf("Converted %d episodes to models.Episode", len(modelEpisodes))
	ep := modelEpisodes[0]
	t.Logf("  First: Number=%s, Num=%d, Title=%s, DataID=%s",
		ep.Number, ep.Num, ep.Title.English, ep.DataID)
}

func TestNineAnimeClient_ToAnimeModel(t *testing.T) {
	result := &NineAnimeResult{
		Title:   "Naruto",
		URL:     "/watch/naruto-677",
		AnimeID: "677",
		Extra:   "SUB DUB Ep 220/220",
	}

	anime := result.ToAnimeModel()

	if anime.Name != "Naruto" {
		t.Errorf("Expected name 'Naruto', got '%s'", anime.Name)
	}
	if anime.Source != "9Anime" {
		t.Errorf("Expected source '9Anime', got '%s'", anime.Source)
	}
	if anime.URL != "677" {
		t.Errorf("Expected URL '677', got '%s'", anime.URL)
	}

	t.Logf("ToAnimeModel conversion successful: %+v", anime)
}

func TestNineAnimeAdapter_Interface(t *testing.T) {
	// Verify NineAnimeAdapter implements UnifiedScraper
	adapter := &NineAnimeAdapter{client: NewNineAnimeClient()}

	var _ UnifiedScraper = adapter // compile-time check

	if adapter.GetType() != NineAnimeType {
		t.Errorf("Expected type NineAnimeType, got %v", adapter.GetType())
	}

	t.Log("NineAnimeAdapter correctly implements UnifiedScraper interface")
}
