// Package api coordinates source resolution and playback-oriented orchestration.
package api

import (
	"context"
	"errors"
	"fmt"
	"strings"

	providerspkg "github.com/alvarorichard/Goanime/internal/api/providers"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
)

func getEpisodesByResolvedSource(anime *models.Anime, resolved ResolvedSource) ([]models.Episode, error) {
	resolved.Apply(anime)

	cleanName := strings.TrimSpace(strings.ReplaceAll(strings.ReplaceAll(anime.Name, "[English]", ""), "[PT-BR]", ""))
	util.Debug("Getting episodes", "source", resolved.Name, "anime", cleanName)

	provider, err := providerspkg.ForKind(resolved.Kind)
	if err != nil {
		return nil, err
	}

	episodes, err := provider.FetchEpisodes(context.Background(), anime)
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes from %s: %w", resolved.Name, err)
	}

	logEpisodeSourceDebug(resolved.Name, len(episodes))
	return episodes, nil
}

func getStreamURLByResolvedSource(anime *models.Anime, episode *models.Episode, quality string, resolved ResolvedSource) (string, error) {
	if anime == nil {
		return "", fmt.Errorf("cannot get stream URL for nil anime")
	}
	if episode == nil {
		return "", fmt.Errorf("cannot get stream URL for nil episode")
	}

	util.ResetPlaybackState()
	resolved.Apply(anime)
	util.SetGlobalAnimeSource(resolved.Name)

	util.Debug("Getting stream URL", "source", resolved.Name, "episode", episode.Number)
	util.Debug("Source details",
		"resolvedSource", resolved.Name,
		"animeURL", anime.URL,
		"episodeURL", episode.URL,
		"episodeNumber", episode.Number,
		"quality", quality)

	provider, err := providerspkg.ForKind(resolved.Kind)
	if err != nil {
		return "", err
	}

	streamURL, err := provider.FetchStreamURL(context.Background(), episode, anime, quality)
	if err != nil {
		if errors.Is(err, scraper.ErrBackRequested) {
			return "", err
		}
		return "", fmt.Errorf("failed to get stream URL from %s: %w", resolved.Name, err)
	}
	if streamURL == "" {
		return "", fmt.Errorf("empty stream URL returned from %s", resolved.Name)
	}

	util.Debug("Stream URL obtained", "source", resolved.Name)
	util.Debug("Stream URL details", "url", streamURL)
	return streamURL, nil
}

func logEpisodeSourceDebug(sourceName string, episodesCount int) {
	if episodesCount > 0 {
		util.Debug("Episodes found", "count", episodesCount, "source", sourceName)
		util.Debug("Source info", "type", sourceName, "features", "provider-backed")
		return
	}

	util.Warn("No episodes found", "source", sourceName)
}
