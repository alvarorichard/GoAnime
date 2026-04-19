package appflow

import (
	"os"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestGetAnimeEpisodes_EmptyResult verifies that GetAnimeEpisodes returns an
// error instead of calling log.Fatal when no episodes are found.
// Before the fix this scenario would kill the process with os.Exit(1).
func TestGetAnimeEpisodes_EmptyResult(t *testing.T) {
	anime := &models.Anime{
		Name:   "NonExistentAnime12345",
		URL:    "https://invalid.example.com/anime/does-not-exist",
		Source: "test",
	}

	episodes, err := GetAnimeEpisodes(anime)

	assert.Error(t, err, "expected an error for anime with no episodes")
	assert.Nil(t, episodes, "expected nil episodes on error")
	t.Logf("Got expected error: %v", err)
}

// TestSearchAnime_InvalidName verifies that SearchAnime returns an error
// instead of fataling when the search fails.
func TestSearchAnime_InvalidName(t *testing.T) {
	// SearchAnime may open an interactive fuzzy finder (tcell-based TUI) if
	// results are returned. On CI there is no TTY, so tcell panics (Windows)
	// or hangs waiting for terminal input.
	if os.Getenv("CI") != "" {
		t.Skip("Skipping interactive fuzzy-finder test in CI (no TTY available)")
	}

	anime, err := SearchAnime("zzzzz_nonexistent_anime_99999")

	// Either an error is returned or a nil anime — both are acceptable.
	// The key assertion is that we reach this line (no os.Exit).
	if err != nil {
		t.Logf("Got expected error: %v", err)
	} else {
		t.Logf("Search returned anime (may have fuzzy matched): %+v", anime)
	}
}

// TestSearchAnimeEnhanced_InvalidName verifies the enhanced search variant.
func TestSearchAnimeEnhanced_InvalidName(t *testing.T) {
	// SearchAnimeEnhanced may open an interactive fuzzy finder (tcell-based TUI)
	// if results are returned. On CI there is no TTY, so tcell panics (Windows)
	// or hangs waiting for terminal input.
	if os.Getenv("CI") != "" {
		t.Skip("Skipping interactive fuzzy-finder test in CI (no TTY available)")
	}

	anime, err := SearchAnimeEnhanced("zzzzz_nonexistent_anime_99999")

	if err != nil {
		t.Logf("Got expected error: %v", err)
	} else {
		t.Logf("Search returned anime (may have fuzzy matched): %+v", anime)
	}
}

// TestGetAnimeEpisodes_NilAnime verifies graceful handling of nil anime.
func TestGetAnimeEpisodes_NilAnime(t *testing.T) {
	defer func() {
		if r := recover(); r != nil {
			t.Fatalf("GetAnimeEpisodes panicked on nil anime: %v", r)
		}
	}()

	// This may panic or return error — we just want to verify no os.Exit
	episodes, err := GetAnimeEpisodes(&models.Anime{})
	if err != nil {
		t.Logf("Got expected error: %v", err)
	} else {
		t.Logf("Returned %d episodes", len(episodes))
	}
}
