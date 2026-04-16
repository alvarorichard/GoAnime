// Package api coordinates source resolution and playback-oriented orchestration.
package api

import (
	"errors"
	"fmt"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
)

func getEpisodesByResolvedSource(anime *models.Anime, resolved ResolvedSource) ([]models.Episode, error) {
	resolved.Apply(anime)

	var episodes []models.Episode
	var err error
	switch resolved.Kind {
	case SourceSuperFlix:
		episodes, err = GetSuperFlixEpisodes(anime)
	case SourceFlixHQ:
		episodes, err = GetFlixHQEpisodes(anime)
	case SourceNineAnime:
		episodes, err = GetNineAnimeEpisodes(anime)
	default:
		episodes, err = fetchEpisodesWithResolvedSource(anime, resolved, sourceProviderFor)
		if err == nil && resolved.Kind == SourceAllAnime && anime.MalID > 0 {
			util.Debug("AniSkip integration enabled", "malID", anime.MalID)
		}
	}
	if err != nil {
		return nil, fmt.Errorf("failed to get episodes from %s: %w", resolved.Name, err)
	}
	return episodes, nil
}

func getStreamURLByResolvedSource(anime *models.Anime, episode *models.Episode, quality string) (string, error) {
	if anime == nil {
		return "", fmt.Errorf("cannot get stream URL for nil anime")
	}
	if episode == nil {
		return "", fmt.Errorf("cannot get stream URL for nil episode")
	}

	util.ClearGlobalSubtitles()

	resolved, err := ResolveSource(anime)
	if err != nil {
		return "", err
	}
	resolved.Apply(anime)
	util.SetGlobalAnimeSource(resolved.Name)

	switch resolved.Kind {
	case SourceSuperFlix:
		return GetSuperFlixStreamURL(anime, episode, quality)
	case SourceFlixHQ:
		streamURL, subtitles, streamErr := GetFlixHQStreamURL(anime, episode, quality)
		if streamErr != nil {
			return "", streamErr
		}

		if len(subtitles) > 0 && !util.GlobalNoSubs {
			var subInfos []util.SubtitleInfo
			for _, sub := range subtitles {
				subInfos = append(subInfos, util.SubtitleInfo{
					URL:      sub.URL,
					Language: sub.Language,
					Label:    sub.Label,
				})
			}
			util.SetGlobalSubtitles(subInfos)
		}
		return streamURL, nil
	case SourceNineAnime:
		return GetNineAnimeStreamURL(anime, episode, quality)
	default:
		util.Debug("Getting stream URL", "source", resolved.Name, "episode", episode.Number)
		util.Debug("Source details",
			"resolvedSource", resolved.Name,
			"animeURL", anime.URL,
			"episodeURL", episode.URL,
			"episodeNumber", episode.Number,
			"quality", quality)

		streamURL, streamErr := fetchStreamURLWithResolvedSource(anime, episode, quality, resolved, sourceProviderFor)
		if streamErr != nil {
			if errors.Is(streamErr, scraper.ErrBackRequested) {
				return "", streamErr
			}
			return "", fmt.Errorf("failed to get stream URL from %s: %w", resolved.Name, streamErr)
		}
		if streamURL == "" {
			return "", fmt.Errorf("empty stream URL returned from %s", resolved.Name)
		}

		util.Debug("Stream URL obtained", "source", resolved.Name)
		util.Debug("Stream URL details", "url", streamURL)
		return streamURL, nil
	}
}
