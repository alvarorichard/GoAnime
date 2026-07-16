package player

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Regression for the in-player menu boundary bug: playNextEpisode /
// playPreviousEpisode returned nil at the playlist edges, which unwound
// handleUserInput and surfaced the post-playback menu while mpv was still
// playing the current episode. They must return errStayInPlayerMenu so the
// menu loop keeps running.

func TestPlayNextEpisode_LastEpisodeStaysInPlayerMenu(t *testing.T) {
	t.Parallel()
	eps := []models.Episode{{Number: "1"}, {Number: "2"}}

	// currentIndex = 1 (last episode) → "next" asks for index 2.
	err := playNextEpisode(2, eps, 0, 0, nil, nil, "")

	require.Error(t, err, "nil at the boundary exits the player menu with mpv still running")
	assert.ErrorIs(t, err, errStayInPlayerMenu)
}

func TestPlayPreviousEpisode_FirstEpisodeStaysInPlayerMenu(t *testing.T) {
	t.Parallel()
	eps := []models.Episode{{Number: "1"}, {Number: "2"}}

	// currentIndex = 0 (first episode) → "previous" asks for index -1.
	err := playPreviousEpisode(-1, eps, 0, 0, nil, nil, "")

	require.Error(t, err, "nil at the boundary exits the player menu with mpv still running")
	assert.ErrorIs(t, err, errStayInPlayerMenu)
}

func TestFindSelectedEpisodeIndex(t *testing.T) {
	t.Parallel()

	sharedURL := []models.Episode{
		{URL: "animeID", Number: "1"},
		{URL: "animeID", Number: "2"},
		{URL: "animeID", Number: "3"},
	}
	uniqueURL := []models.Episode{
		{URL: "u1", Number: "1"},
		{URL: "u2", Number: "2"},
	}

	tests := []struct {
		name     string
		episodes []models.Episode
		url      string
		numStr   string
		want     int
	}{
		// Regression: AllAnime uses the anime ID as the URL of every episode,
		// so the old URL-only match always resolved to index 0 (episode 1)
		// regardless of which episode the user picked.
		{"shared url picks by number", sharedURL, "animeID", "3", 2},
		{"shared url middle episode", sharedURL, "animeID", "2", 1},
		{"unique url exact match", uniqueURL, "u2", "2", 1},
		{"number wins when url stale", uniqueURL, "unknown", "2", 1},
		{"url fallback when number unknown", uniqueURL, "u1", "weird", 0},
		{"not found", uniqueURL, "nope", "9", -1},
		{"empty list", nil, "u1", "1", -1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, findSelectedEpisodeIndex(tt.episodes, tt.url, tt.numStr))
		})
	}
}
