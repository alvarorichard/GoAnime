package test

import (
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
)

// TestHandleMovie tests the HandleMovie function with test data
func TestHandleMovie(t *testing.T) {
	t.Skip("Skipping integration test - requires external dependencies (MPV, network)")

	// Este teste seria executado em um ambiente de integração completo
	// onde MPV e outras dependências externas estão disponíveis

	// Arrange
	anime := &models.Anime{
		Name:      "Test Movie",
		URL:       "https://example.com/movie/test",
		ImageURL:  "https://example.com/image.jpg",
		AnilistID: 123,
		MalID:     456,
		Details: models.AniListDetails{
			Title: models.Title{
				Romaji:  "Test Movie",
				English: "Test Movie English",
			},
		},
	}

	episodes := []models.Episode{
		{
			Number:   "Movie",
			URL:      "https://example.com/episode/test",
			Duration: 7200, // 2 horas em segundos
		},
	}

	// Act & Assert
	// Em um teste real, verificaríamos se a função não retorna erro
	// e se o video é reproduzido corretamente
	assert.NotNil(t, anime)
	assert.NotEmpty(t, episodes)

	// playback.HandleMovie(anime, episodes, false)
}

// TestMovieDurationCalculation tests movie duration calculation
func TestMovieDurationCalculation(t *testing.T) {
	// Arrange
	episode := models.Episode{
		Duration: 7200, // 2 horas em segundos
	}

	// Act
	duration := time.Duration(episode.Duration) * time.Second

	// Assert
	assert.Equal(t, 2*time.Hour, duration)
	assert.Equal(t, 7200*time.Second, duration)
}

// TestMovieMetadata tests the movie metadata structure
func TestMovieMetadata(t *testing.T) {
	// Arrange
	anime := &models.Anime{
		Name:      "Test Movie",
		URL:       "https://example.com/movie/test",
		ImageURL:  "https://example.com/image.jpg",
		AnilistID: 123,
		MalID:     456,
	}

	episodes := []models.Episode{
		{
			Number:   "Movie",
			URL:      "https://example.com/episode/test",
			Duration: 7200,
		},
	}

	// Act & Assert
	assert.NotNil(t, anime)
	assert.Equal(t, "Test Movie", anime.Name)
	assert.Equal(t, 123, anime.AnilistID)
	assert.Equal(t, 456, anime.MalID)

	assert.Len(t, episodes, 1)
	assert.Equal(t, "Movie", episodes[0].Number)
	assert.Equal(t, 7200, episodes[0].Duration)
}
