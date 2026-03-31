// Package providers defines the EpisodeProvider interface and its implementations.
package providers

import "github.com/alvarorichard/Goanime/internal/models"

// EpisodeProvider defines the unified interface for all anime/media source providers.
// Each provider encapsulates the logic for fetching episodes and resolving stream URLs
// for a specific source (AllAnime, AnimeFire, FlixHQ, etc.).
type EpisodeProvider interface {
	Name() string
	FetchEpisodes(anime *models.Anime) ([]models.Episode, error)
	GetStreamURL(episode *models.Episode, anime *models.Anime, quality string) (string, error)
	HasSeasons() bool
}
