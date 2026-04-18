package providers

import (
	"context"
	"fmt"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

// EpisodeNumber extracts the episode number string from an Episode model.
// Returns "" if indeterminate — caller must decide how to handle.
func EpisodeNumber(ep *models.Episode) string {
	if ep == nil {
		return ""
	}
	if ep.Number != "" {
		return ep.Number
	}
	if ep.Num > 0 {
		return fmt.Sprintf("%d", ep.Num)
	}
	return ""
}

// --- AllAnime Provider ---

type allAnimeProvider struct {
	sm *scraper.ScraperManager
}

func init() {
	RegisterProvider(source.AllAnime, func(sm *scraper.ScraperManager) Provider {
		return &allAnimeProvider{sm: sm}
	})
}

func (p *allAnimeProvider) Kind() source.SourceKind { return source.AllAnime }
func (p *allAnimeProvider) HasSeasons() bool        { return false }

func (p *allAnimeProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.sm.GetScraper(scraper.AllAnimeType)
	if err != nil {
		return nil, err
	}
	animeID := source.ExtractAllAnimeID(anime.URL)
	return adapter.GetAnimeEpisodes(animeID)
}

func (p *allAnimeProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.sm.GetScraper(scraper.AllAnimeType)
	if err != nil {
		return "", err
	}
	animeID := source.ExtractAllAnimeID(anime.URL)
	epNum := EpisodeNumber(episode)
	if quality == "" {
		quality = "best"
	}
	url, _, err := adapter.GetStreamURL(animeID, epNum, quality)
	if err != nil {
		return "", fmt.Errorf("allAnime stream: %w", err)
	}
	return url, nil
}

// --- AnimeFire Provider ---

type animeFireProvider struct {
	sm *scraper.ScraperManager
}

func init() {
	RegisterProvider(source.AnimeFire, func(sm *scraper.ScraperManager) Provider {
		return &animeFireProvider{sm: sm}
	})
}

func (p *animeFireProvider) Kind() source.SourceKind { return source.AnimeFire }
func (p *animeFireProvider) HasSeasons() bool        { return false }

func (p *animeFireProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.sm.GetScraper(scraper.AnimefireType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *animeFireProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.sm.GetScraper(scraper.AnimefireType)
	if err != nil {
		return "", err
	}
	url, _, err := adapter.GetStreamURL(episode.URL)
	if err != nil {
		return "", fmt.Errorf("animeFire stream: %w", err)
	}
	return url, nil
}

// --- Goyabu Provider ---

type goyabuProvider struct {
	sm *scraper.ScraperManager
}

func init() {
	RegisterProvider(source.Goyabu, func(sm *scraper.ScraperManager) Provider {
		return &goyabuProvider{sm: sm}
	})
}

func (p *goyabuProvider) Kind() source.SourceKind { return source.Goyabu }
func (p *goyabuProvider) HasSeasons() bool        { return false }

func (p *goyabuProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.sm.GetScraper(scraper.GoyabuType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *goyabuProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.sm.GetScraper(scraper.GoyabuType)
	if err != nil {
		return "", err
	}
	url, _, err := adapter.GetStreamURL(episode.URL)
	if err != nil {
		return "", fmt.Errorf("goyabu stream: %w", err)
	}
	return url, nil
}

// --- FlixHQ Provider ---

type flixHQProvider struct {
	sm *scraper.ScraperManager
}

func init() {
	RegisterProvider(source.FlixHQ, func(sm *scraper.ScraperManager) Provider {
		return &flixHQProvider{sm: sm}
	})
}

func (p *flixHQProvider) Kind() source.SourceKind { return source.FlixHQ }
func (p *flixHQProvider) HasSeasons() bool        { return true }

func (p *flixHQProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.sm.GetScraper(scraper.FlixHQType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *flixHQProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.sm.GetScraper(scraper.FlixHQType)
	if err != nil {
		return "", err
	}
	if quality == "" {
		quality = "auto"
	}
	url, _, err := adapter.GetStreamURL(episode.URL, "upcloud", quality, "english")
	if err != nil {
		return "", fmt.Errorf("flixHQ stream: %w", err)
	}
	return url, nil
}

// --- SFlix Provider ---

type sflixProvider struct {
	sm *scraper.ScraperManager
}

func init() {
	RegisterProvider(source.SFlix, func(sm *scraper.ScraperManager) Provider {
		return &sflixProvider{sm: sm}
	})
}

func (p *sflixProvider) Kind() source.SourceKind { return source.SFlix }
func (p *sflixProvider) HasSeasons() bool        { return true }

func (p *sflixProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.sm.GetScraper(scraper.SFlixType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *sflixProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.sm.GetScraper(scraper.SFlixType)
	if err != nil {
		return "", err
	}
	if quality == "" {
		quality = "auto"
	}
	url, _, err := adapter.GetStreamURL(episode.URL, "upcloud", quality, "english")
	if err != nil {
		return "", fmt.Errorf("sflix stream: %w", err)
	}
	return url, nil
}

// --- NineAnime Provider ---

type nineAnimeProvider struct {
	sm *scraper.ScraperManager
}

func init() {
	RegisterProvider(source.NineAnime, func(sm *scraper.ScraperManager) Provider {
		return &nineAnimeProvider{sm: sm}
	})
}

func (p *nineAnimeProvider) Kind() source.SourceKind { return source.NineAnime }
func (p *nineAnimeProvider) HasSeasons() bool        { return false }

func (p *nineAnimeProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.sm.GetScraper(scraper.NineAnimeType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *nineAnimeProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.sm.GetScraper(scraper.NineAnimeType)
	if err != nil {
		return "", err
	}
	url, _, err := adapter.GetStreamURL(episode.URL)
	if err != nil {
		return "", fmt.Errorf("9anime stream: %w", err)
	}
	return url, nil
}

// --- SuperFlix Provider ---

type superFlixProvider struct {
	sm *scraper.ScraperManager
}

func init() {
	RegisterProvider(source.SuperFlix, func(sm *scraper.ScraperManager) Provider {
		return &superFlixProvider{sm: sm}
	})
}

func (p *superFlixProvider) Kind() source.SourceKind { return source.SuperFlix }
func (p *superFlixProvider) HasSeasons() bool        { return true }

func (p *superFlixProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.sm.GetScraper(scraper.SuperFlixType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *superFlixProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.sm.GetScraper(scraper.SuperFlixType)
	if err != nil {
		return "", err
	}
	epNum := EpisodeNumber(episode)
	mediaType := "serie"
	if anime.MediaType == models.MediaTypeMovie {
		mediaType = "filme"
	}
	season := "1"
	if anime.CurrentSeason > 0 {
		season = fmt.Sprintf("%d", anime.CurrentSeason)
	}
	url, _, err := adapter.GetStreamURL(episode.URL, mediaType, season, epNum)
	if err != nil {
		return "", fmt.Errorf("superFlix stream: %w", err)
	}
	return url, nil
}

// --- AnimeDrive Provider ---

type animeDriveProvider struct {
	sm *scraper.ScraperManager
}

func init() {
	RegisterProvider(source.AnimeDrive, func(sm *scraper.ScraperManager) Provider {
		return &animeDriveProvider{sm: sm}
	})
}

func (p *animeDriveProvider) Kind() source.SourceKind { return source.AnimeDrive }
func (p *animeDriveProvider) HasSeasons() bool        { return false }

func (p *animeDriveProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.sm.GetScraper(scraper.AnimeDriveType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *animeDriveProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.sm.GetScraper(scraper.AnimeDriveType)
	if err != nil {
		return "", err
	}
	url, _, err := adapter.GetStreamURL(episode.URL, "auto")
	if err != nil {
		return "", fmt.Errorf("animeDrive stream: %w", err)
	}
	return url, nil
}
