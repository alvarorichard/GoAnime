package api

import (
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

func TestSmokeGoyabu(t *testing.T) {
	if testing.Short() {
		t.Skip("skipping Goyabu smoke test in short mode")
	}

	anime := &models.Anime{
		Name:   "[PT-BR] Naruto Shippuden Dublado",
		URL:    "https://goyabu.io/anime/naruto-shippuden-dublado-online-hd",
		Source: "Goyabu",
	}

	episodes, err := GetAnimeEpisodesEnhanced(anime)
	if err != nil {
		if errors.Is(err, scraper.ErrSourceUnavailable) {
			t.Skipf("skipping Goyabu smoke test while upstream is unavailable: %v", err)
		}
		t.Fatalf("GetAnimeEpisodesEnhanced failed: %v", err)
	}

	if len(episodes) == 0 {
		t.Fatal("GetAnimeEpisodesEnhanced returned no episodes")
	}

	streamURL, err := GetEpisodeStreamURL(&episodes[0], anime, "best")
	if err != nil {
		if errors.Is(err, scraper.ErrSourceUnavailable) {
			t.Skipf("skipping Goyabu stream smoke step while upstream is unavailable: %v", err)
		}
		t.Fatalf("GetEpisodeStreamURL failed: %v", err)
	}

	if streamURL == "" {
		t.Fatal("GetEpisodeStreamURL returned an empty stream URL")
	}
}
