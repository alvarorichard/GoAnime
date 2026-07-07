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
	// sm is injected by the Provider factory (old path). When nil (Model B
	// path, registered in init), manager() falls back to the lazy global
	// ScraperManager singleton — nothing is built at init time.
	sm *scraper.ScraperManager
}

func init() {
	RegisterProvider(source.AllAnime, func(sm *scraper.ScraperManager) Provider {
		return &allAnimeProvider{sm: sm}
	})
	source.Register(&allAnimeProvider{})
}

func (p *allAnimeProvider) manager() *scraper.ScraperManager {
	if p.sm != nil {
		return p.sm
	}
	return scraper.NewScraperManager()
}

func (p *allAnimeProvider) Describe() source.Descriptor {
	return source.Descriptor{
		Kind:        source.AllAnime,
		Priority:    40,
		Explicit:    []string{"AllAnime"},
		Tags:        []string{"[english]"},
		URLMatchers: []string{"allanime"},
		ShortID:     true,
	}
}

func (p *allAnimeProvider) Kind() source.SourceKind { return source.AllAnime }
func (p *allAnimeProvider) HasSeasons() bool        { return false }

func (p *allAnimeProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.manager().GetScraper(scraper.AllAnimeType)
	if err != nil {
		return nil, err
	}
	animeID := source.ExtractAllAnimeID(anime.URL)
	return adapter.GetAnimeEpisodes(animeID)
}

func (p *allAnimeProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.manager().GetScraper(scraper.AllAnimeType)
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
	source.Register(&animeFireProvider{})
}

func (p *animeFireProvider) manager() *scraper.ScraperManager {
	if p.sm != nil {
		return p.sm
	}
	return scraper.NewScraperManager()
}

func (p *animeFireProvider) Describe() source.Descriptor {
	return source.Descriptor{
		Kind:        source.AnimeFire,
		Priority:    10,
		Explicit:    []string{"Animefire.io", "AnimeFire"},
		Tags:        []string{"[animefire]"},
		URLMatchers: []string{"animefire"},
	}
}

func (p *animeFireProvider) Kind() source.SourceKind { return source.AnimeFire }
func (p *animeFireProvider) HasSeasons() bool        { return false }

func (p *animeFireProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.manager().GetScraper(scraper.AnimefireType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *animeFireProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.manager().GetScraper(scraper.AnimefireType)
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
	source.Register(&goyabuProvider{})
}

func (p *goyabuProvider) manager() *scraper.ScraperManager {
	if p.sm != nil {
		return p.sm
	}
	return scraper.NewScraperManager()
}

func (p *goyabuProvider) Describe() source.Descriptor {
	return source.Descriptor{
		Kind:        source.Goyabu,
		Priority:    20,
		Explicit:    []string{"Goyabu"},
		Tags:        []string{"[goyabu]"},
		URLMatchers: []string{"goyabu"},
	}
}

func (p *goyabuProvider) Kind() source.SourceKind { return source.Goyabu }
func (p *goyabuProvider) HasSeasons() bool        { return false }

func (p *goyabuProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.manager().GetScraper(scraper.GoyabuType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *goyabuProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.manager().GetScraper(scraper.GoyabuType)
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
//
// TEMP-DISABLED: entire FlixHQ provider commented out until a fix lands.
// Restore the init() and the type+methods together.
/*
func init() {
	RegisterProvider(source.FlixHQ, func(sm *scraper.ScraperManager) Provider {
		return &flixHQProvider{sm: sm}
	})
}

type flixHQProvider struct {
	sm *scraper.ScraperManager
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
*/

// --- SFlix Provider ---
//
// TEMP-DISABLED: entire SFlix provider commented out until a fix lands.
// Restore the init() and the type+methods together.
/*
func init() {
	RegisterProvider(source.SFlix, func(sm *scraper.ScraperManager) Provider {
		return &sflixProvider{sm: sm}
	})
}

type sflixProvider struct {
	sm *scraper.ScraperManager
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
*/

// --- NineAnime Provider ---
//
// TEMP-DISABLED: entire NineAnime provider commented out until a fix lands.
// Restore the init() and the type+methods together.
/*
func init() {
	RegisterProvider(source.NineAnime, func(sm *scraper.ScraperManager) Provider {
		return &nineAnimeProvider{sm: sm}
	})
}

type nineAnimeProvider struct {
	sm *scraper.ScraperManager
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
*/

// --- SuperFlix Provider ---

type superFlixProvider struct {
	sm *scraper.ScraperManager
}

func init() {
	RegisterProvider(source.SuperFlix, func(sm *scraper.ScraperManager) Provider {
		return &superFlixProvider{sm: sm}
	})
	source.Register(&superFlixProvider{})
}

func (p *superFlixProvider) manager() *scraper.ScraperManager {
	if p.sm != nil {
		return p.sm
	}
	return scraper.NewScraperManager()
}

func (p *superFlixProvider) Describe() source.Descriptor {
	return source.Descriptor{
		Kind:        source.SuperFlix,
		Priority:    30,
		Explicit:    []string{"SuperFlix"},
		Tags:        []string{"[superflix]"},
		URLMatchers: []string{"superflix"},
	}
}

func (p *superFlixProvider) Kind() source.SourceKind { return source.SuperFlix }
func (p *superFlixProvider) HasSeasons() bool        { return true }

func (p *superFlixProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.manager().GetScraper(scraper.SuperFlixType)
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

func (p *superFlixProvider) FetchStreamURL(_ context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	adapter, err := p.manager().GetScraper(scraper.SuperFlixType)
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
