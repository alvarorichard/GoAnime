// Package providers defines the Provider interface and registry for media source dispatch.
// Each source (AllAnime, AnimeFire, etc.) implements Provider to handle episodes and streams.
package providers

import (
	"context"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
)

// Provider encapsulates the high-level logic for a media source.
// Implementations know how to extract parameters from models and call the scraper layer.
type Provider interface {
	// Kind returns the SourceKind for this provider.
	Kind() source.SourceKind

	// FetchEpisodes retrieves episodes for an anime from this source.
	FetchEpisodes(ctx context.Context, anime *models.Anime) ([]models.Episode, error)

	// FetchStreamURL resolves the streaming URL for a specific episode.
	FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error)

	// HasSeasons returns true if this source organizes content into seasons (e.g. FlixHQ, SFlix).
	HasSeasons() bool
}
