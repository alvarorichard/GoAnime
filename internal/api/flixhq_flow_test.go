package api

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
)

func TestFlixHQFullFlow(t *testing.T) {
	util.IsDebug = true

	// 1. Search for Dexter
	t.Log("=== Step 1: Searching for Dexter ===")
	sm := scraper.NewScraperManager()
	flixhqType := scraper.FlixHQType
	results, err := sm.SearchAnime("dexter", &flixhqType)
	if err != nil {
		t.Fatalf("Search error: %v", err)
	}

	if len(results) == 0 {
		t.Fatal("No results found")
	}

	// Find TV show
	anime := findTVShowInResults(results)
	if anime == nil {
		t.Fatal("No TV show found")
		return
	}

	for _, r := range results {
		t.Logf("Found: %s (Source: %s, MediaType: %s, URL: %s)", r.Name, r.Source, r.MediaType, r.URL)
	}

	t.Logf("=== Step 2: Selected anime - Source: %s, MediaType: %s ===", anime.Source, anime.MediaType)

	// 2. Get seasons directly without user interaction
	t.Log("=== Step 3: Getting seasons ===")
	flixhqClient := scraper.NewFlixHQClient()

	// Extract media ID from URL (get last number from URL like "watch-dexter-39448")
	mediaID := extractMediaIDFromURL(anime.URL)
	t.Logf("Extracted mediaID: %s from URL: %s", mediaID, anime.URL)

	seasons, err := flixhqClient.GetSeasons(mediaID)
	if err != nil {
		t.Fatalf("Get seasons error: %v", err)
	}

	if len(seasons) == 0 {
		t.Fatal("No seasons found")
	}

	t.Logf("Found %d seasons", len(seasons))
	t.Logf("Season 1: %s (ID: %s)", seasons[0].Title, seasons[0].ID)

	// 3. Get episodes for first season
	t.Log("=== Step 4: Getting episodes for Season 1 ===")
	episodes, err := flixhqClient.GetEpisodes(seasons[0].ID)
	if err != nil {
		t.Fatalf("Get episodes error: %v", err)
	}

	if len(episodes) == 0 {
		t.Fatal("No episodes found")
	}

	t.Logf("Found %d episodes", len(episodes))
	t.Logf("First episode: Title=%s, Number=%d, DataID=%s", episodes[0].Title, episodes[0].Number, episodes[0].DataID)

	// 4. Convert to models.Episode
	modelEpisode := models.Episode{
		Number:   "1",
		Num:      1,
		URL:      episodes[0].DataID,
		DataID:   episodes[0].DataID,
		SeasonID: seasons[0].ID,
	}

	// 5. Get stream URL
	t.Log("=== Step 5: Getting stream URL ===")
	streamURL, err := GetEpisodeStreamURL(&modelEpisode, anime, "1080")
	if err != nil {
		t.Fatalf("Get stream URL error: %v", err)
	}

	t.Logf("=== SUCCESS! Stream URL: %s ===", streamURL)
}

// findTVShowInResults finds the first TV show in the search results
func findTVShowInResults(results []*models.Anime) *models.Anime {
	for _, r := range results {
		if r.MediaType == models.MediaTypeTV {
			return r
		}
	}
	return nil
}
