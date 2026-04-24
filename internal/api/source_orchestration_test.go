package api

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
)

func TestGetAnimeEpisodesEnhancedSFlixMovie(t *testing.T) {
	anime := &models.Anime{
		Name:      "Inception",
		Source:    "SFlix",
		MediaType: models.MediaTypeMovie,
		URL:       "https://sflix.to/movie/free-inception-hd-27205",
	}

	episodes, err := GetAnimeEpisodesEnhanced(anime)
	if err != nil {
		t.Fatalf("GetAnimeEpisodesEnhanced returned error: %v", err)
	}
	if len(episodes) != 1 {
		t.Fatalf("GetAnimeEpisodesEnhanced returned %d episodes, want 1", len(episodes))
	}
	if episodes[0].URL != "27205" {
		t.Fatalf("episode URL = %q, want %q", episodes[0].URL, "27205")
	}
}
