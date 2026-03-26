package providers

import (
	"fmt"
	"strings"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
)

// AniSkipFunc is the function signature for AniSkip data fetching.
// Injected to avoid circular dependency with the api package.
type AniSkipFunc func(animeMalId int, episodeNum int, episode *models.Episode) error

// AllAnimeProvider handles episode fetching and stream resolution for AllAnime sources.
type AllAnimeProvider struct {
	manager     *scraper.ScraperManager
	aniSkipFunc AniSkipFunc
}

func NewAllAnimeProvider() *AllAnimeProvider {
	return &AllAnimeProvider{manager: scraper.NewScraperManager()}
}

// SetAniSkipFunc injects the AniSkip integration function.
// Called by the api package during initialization to avoid circular imports.
func (p *AllAnimeProvider) SetAniSkipFunc(fn AniSkipFunc) {
	p.aniSkipFunc = fn
}

func (p *AllAnimeProvider) Name() string { return "AllAnime" }

func (p *AllAnimeProvider) HasSeasons() bool { return false }

func (p *AllAnimeProvider) FetchEpisodes(anime *models.Anime) ([]models.Episode, error) {
	scraperInstance, err := p.manager.GetScraper(scraper.AllAnimeType)
	if err != nil {
		return nil, fmt.Errorf("failed to get AllAnime scraper: %w", err)
	}

	if p.aniSkipFunc != nil && anime.MalID > 0 {
		if adapter, ok := scraperInstance.(interface {
			Client() *scraper.AllAnimeClient
		}); ok {
			client := adapter.Client()
			episodes, aniErr := client.GetAnimeEpisodesWithAniSkip(anime.URL, anime.MalID, p.aniSkipFunc)
			if aniErr == nil {
				util.Debug("AniSkip integration enabled", "malID", anime.MalID)
				return episodes, nil
			}
			util.Debug("AniSkip fallback to regular episodes", "error", aniErr)
		}
	}

	episodes, err := scraperInstance.GetAnimeEpisodes(anime.URL)
	if err != nil {
		return nil, fmt.Errorf("failed to get AllAnime episodes: %w", err)
	}
	return episodes, nil
}

func (p *AllAnimeProvider) GetStreamURL(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	scraperInstance, err := p.manager.GetScraper(scraper.AllAnimeType)
	if err != nil {
		return "", fmt.Errorf("failed to get AllAnime scraper: %w", err)
	}

	if quality == "" {
		quality = "best"
	}

	streamURL, _, streamErr := scraperInstance.GetStreamURL(anime.URL, episode.Number, quality)
	if streamErr != nil {
		return "", fmt.Errorf("failed to get AllAnime stream URL: %w", streamErr)
	}
	if streamURL == "" {
		return "", fmt.Errorf("empty stream URL returned from AllAnime")
	}
	return streamURL, nil
}

// ExtractAllAnimeID extracts the anime ID from a URL or returns the raw ID.
func ExtractAllAnimeID(urlStr string) string {
	if !strings.Contains(urlStr, "http") && len(urlStr) < 30 {
		return urlStr
	}
	if strings.Contains(urlStr, "allanime") {
		for part := range strings.SplitSeq(urlStr, "/") {
			if len(part) > 5 && len(part) < 30 &&
				strings.ContainsAny(part, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") {
				return part
			}
		}
	}
	return urlStr
}
