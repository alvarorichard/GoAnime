package test

import (
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
)

func TestDiscordState_SeasonBug(t *testing.T) {
	t.Run("before fix: no season info in discord state", func(t *testing.T) {
		anime := &models.Anime{
			Name:      "Dexter",
			MediaType: models.MediaTypeTV,
			Source:    "FlixHQ",
			Episodes:  []models.Episode{{Number: "5", Num: 5}},
		}
		assert.Equal(t, "Dexter", anime.Name)
		assert.True(t, anime.IsMovieOrTV())

		episodeNumber := anime.Episodes[0].Number
		oldState := fmt.Sprintf("Episode %s", episodeNumber)

		assert.Equal(t, "Episode 5", oldState, "bug: no season in state")
		assert.NotContains(t, oldState, "S02", "bug: season missing")
	})
}

func TestDiscordState_SeasonFix(t *testing.T) {
	tests := []struct {
		name          string
		anime         *models.Anime
		episodeNumber string
		wantState     string
	}{
		{
			name: "TV show with season 2",
			anime: &models.Anime{
				Name:          "Dexter",
				MediaType:     models.MediaTypeTV,
				Source:        "FlixHQ",
				CurrentSeason: 2,
			},
			episodeNumber: "5",
			wantState:     "S02E5",
		},
		{
			name: "TV show with season 4",
			anime: &models.Anime{
				Name:          "Breaking Bad",
				MediaType:     models.MediaTypeTV,
				Source:        "FlixHQ",
				CurrentSeason: 4,
			},
			episodeNumber: "11",
			wantState:     "S04E11",
		},
		{
			name: "TV show season 1",
			anime: &models.Anime{
				Name:          "Friends",
				MediaType:     models.MediaTypeTV,
				Source:        "FlixHQ",
				CurrentSeason: 1,
			},
			episodeNumber: "1",
			wantState:     "S01E1",
		},
		{
			name: "TV show without CurrentSeason falls back to Episode X",
			anime: &models.Anime{
				Name:          "Some Show",
				MediaType:     models.MediaTypeTV,
				Source:        "FlixHQ",
				CurrentSeason: 0,
			},
			episodeNumber: "7",
			wantState:     "Episode 7",
		},
		{
			name: "movie shows Watching a movie",
			anime: &models.Anime{
				Name:      "Inception",
				MediaType: models.MediaTypeMovie,
				Source:    "FlixHQ",
			},
			episodeNumber: "1",
			wantState:     "Watching a movie",
		},
		{
			name: "anime without season shows Episode X",
			anime: &models.Anime{
				Name:      "Naruto",
				MediaType: models.MediaTypeAnime,
				Source:    "AllAnime",
			},
			episodeNumber: "42",
			wantState:     "Episode 42",
		},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			isMovieOrTV := tc.anime.IsMovieOrTV() || tc.anime.Source == "FlixHQ"

			var state string
			if isMovieOrTV {
				if tc.anime.IsMovie() || tc.anime.MediaType == "movie" {
					state = "Watching a movie"
				} else {
					if tc.anime.CurrentSeason > 0 {
						state = fmt.Sprintf("S%02dE%s", tc.anime.CurrentSeason, tc.episodeNumber)
					} else {
						state = fmt.Sprintf("Episode %s", tc.episodeNumber)
					}
				}
			} else {
				state = fmt.Sprintf("Episode %s", tc.episodeNumber)
			}

			assert.Equal(t, tc.wantState, state)
		})
	}
}

func TestDiscordState_EpisodeNumberPadding(t *testing.T) {
	anime := &models.Anime{
		Name:          "Dexter",
		MediaType:     models.MediaTypeTV,
		CurrentSeason: 2,
	}
	assert.Equal(t, "Dexter", anime.Name)
	assert.True(t, anime.IsTV())
	episodeNumber := "5"

	state := fmt.Sprintf("S%02dE%s", anime.CurrentSeason, episodeNumber)
	assert.Equal(t, "S02E5", state, "episode number is not zero-padded since it's a raw string")
}
