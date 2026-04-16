// Package api coordinates source resolution and playback-oriented orchestration.
package api

import (
	"fmt"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

// SourceProvider fetches episodes and stream URLs for a resolved source.
type SourceProvider interface {
	Kind() SourceKind
	FetchEpisodes(anime *models.Anime) ([]models.Episode, error)
	FetchStreamURL(anime *models.Anime, episode *models.Episode, quality string) (string, error)
}

type sourceProviderLookup func(SourceKind) (SourceProvider, bool)

var defaultSourceProviders = map[SourceKind]SourceProvider{
	SourceAllAnime:   allAnimeSourceProvider{},
	SourceAnimefire:  scraperBackedSourceProvider{kind: SourceAnimefire, streamQualityMode: streamQualityRequested},
	SourceAnimeDrive: scraperBackedSourceProvider{kind: SourceAnimeDrive, streamQualityMode: streamQualityAuto},
	SourceGoyabu:     scraperBackedSourceProvider{kind: SourceGoyabu, streamQualityMode: streamQualityNone},
}

type streamQualityMode int

const (
	streamQualityRequested streamQualityMode = iota
	streamQualityAuto
	streamQualityNone
)

type allAnimeSourceProvider struct{}

func (allAnimeSourceProvider) Kind() SourceKind {
	return SourceAllAnime
}

func (allAnimeSourceProvider) FetchEpisodes(anime *models.Anime) ([]models.Episode, error) {
	scraperInstance, err := getScraperForKind(SourceAllAnime)
	if err != nil {
		return nil, err
	}

	if allAnimeClient, ok := scraperInstance.(*scraper.AllAnimeClient); ok && anime.MalID > 0 {
		return allAnimeClient.GetAnimeEpisodesWithAniSkip(anime.URL, anime.MalID, GetAndParseAniSkipData)
	}

	return scraperInstance.GetAnimeEpisodes(anime.URL)
}

func (allAnimeSourceProvider) FetchStreamURL(anime *models.Anime, episode *models.Episode, quality string) (string, error) {
	streamURL, _, err := GetAllAnimeEpisodeURLDirect(anime, providerEpisodeNumber(episode), normalizeStreamQuality(quality))
	if err != nil {
		return "", err
	}
	return streamURL, nil
}

type scraperBackedSourceProvider struct {
	kind              SourceKind
	streamQualityMode streamQualityMode
}

func (p scraperBackedSourceProvider) Kind() SourceKind {
	return p.kind
}

func (p scraperBackedSourceProvider) FetchEpisodes(anime *models.Anime) ([]models.Episode, error) {
	scraperInstance, err := getScraperForKind(p.kind)
	if err != nil {
		return nil, err
	}
	return scraperInstance.GetAnimeEpisodes(anime.URL)
}

func (p scraperBackedSourceProvider) FetchStreamURL(_ *models.Anime, episode *models.Episode, quality string) (string, error) {
	scraperInstance, err := getScraperForKind(p.kind)
	if err != nil {
		return "", err
	}

	switch p.streamQualityMode {
	case streamQualityAuto:
		streamURL, _, streamErr := scraperInstance.GetStreamURL(episode.URL, "auto")
		return streamURL, streamErr
	case streamQualityNone:
		streamURL, _, streamErr := scraperInstance.GetStreamURL(episode.URL)
		return streamURL, streamErr
	default:
		streamURL, _, streamErr := scraperInstance.GetStreamURL(episode.URL, normalizeStreamQuality(quality))
		return streamURL, streamErr
	}
}

func getScraperForKind(kind SourceKind) (scraper.UnifiedScraper, error) {
	scraperType, ok := kind.ScraperType()
	if !ok {
		return nil, fmt.Errorf("source %s does not map to a scraper", kind)
	}
	return scraper.NewScraperManager().GetScraper(scraperType)
}

func sourceProviderFor(kind SourceKind) (SourceProvider, bool) {
	provider, ok := defaultSourceProviders[kind]
	return provider, ok
}

func fetchEpisodesWithResolvedSource(anime *models.Anime, resolved ResolvedSource, lookup sourceProviderLookup) ([]models.Episode, error) {
	provider, ok := lookup(resolved.Kind)
	if !ok {
		return nil, fmt.Errorf("no source provider registered for %s", resolved.Name)
	}

	episodes, err := provider.FetchEpisodes(anime)
	if err != nil {
		return nil, err
	}

	return episodes, nil
}

func fetchStreamURLWithResolvedSource(anime *models.Anime, episode *models.Episode, quality string, resolved ResolvedSource, lookup sourceProviderLookup) (string, error) {
	provider, ok := lookup(resolved.Kind)
	if !ok {
		return "", fmt.Errorf("no source provider registered for %s", resolved.Name)
	}

	streamURL, err := provider.FetchStreamURL(anime, episode, quality)
	if err != nil {
		return "", err
	}

	return streamURL, nil
}

func normalizeStreamQuality(quality string) string {
	if quality == "" {
		return "best"
	}
	return quality
}

func providerEpisodeNumber(episode *models.Episode) string {
	if episode == nil {
		return ""
	}
	if episode.Number != "" {
		return episode.Number
	}
	if episode.Num > 0 {
		return fmt.Sprintf("%d", episode.Num)
	}
	return "1"
}
