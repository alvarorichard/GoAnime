// Package goanime provides a public API for anime scraping and searching functionality.
// This package can be used as a library in other Go projects.
package goanime

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/api/providers"
	apisource "github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper"
	"github.com/alvarorichard/Goanime/v2/pkg/goanime/types"
)

// Manager abstracts the scraper backend used by Client. registryManager (the
// default) satisfies it on top of the Model B registry.
type Manager interface {
	SearchAnime(query string, scraperType *scraper.ScraperType) ([]*models.Anime, error)
	GetScraper(scraperType scraper.ScraperType) (scraper.UnifiedScraper, error)
}

// Client is the main client for interacting with anime sources
type Client struct {
	manager Manager
}

// NewClient creates a new GoAnime client backed by the Model B registry.
func NewClient() *Client {
	return &Client{
		manager: registryManager{},
	}
}

// registryManager implements Manager over the Model B registry: searches fan out
// through providers.SearchAll and per-source access builds a standalone adapter
// via scraper.NewAdapter — the same paths the CLI uses, so the SDK stays in sync
// with app behavior (kill-switch, tagging, circuit breaker) for free.
type registryManager struct{}

func (registryManager) SearchAnime(query string, scraperType *scraper.ScraperType) ([]*models.Anime, error) {
	if scraperType != nil {
		kind, ok := scraperKind(*scraperType)
		if !ok {
			return nil, fmt.Errorf("unknown scraper type %v", *scraperType)
		}
		return providers.SearchAll(context.Background(), query, kind)
	}
	return providers.SearchAll(context.Background(), query)
}

func (registryManager) GetScraper(scraperType scraper.ScraperType) (scraper.UnifiedScraper, error) {
	return scraper.NewAdapter(scraperType)
}

// scraperKind maps an internal ScraperType to its Model B SourceKind.
func scraperKind(st scraper.ScraperType) (apisource.SourceKind, bool) {
	return apisource.KindFor(st)
}

// SearchAnime searches for anime across all sources or a specific source.
// If source is nil, searches all available sources.
// Returns a list of anime results or an error.
func (c *Client) SearchAnime(query string, source *types.Source) ([]*types.Anime, error) {
	var scraperType *scraper.ScraperType
	if source != nil {
		st := source.ToScraperType()
		scraperType = &st
	}

	results, err := c.manager.SearchAnime(query, scraperType)
	if err != nil {
		return nil, err
	}

	// Convert internal models to public types
	return types.FromInternalAnimeList(results), nil
}

// GetAnimeEpisodes retrieves all episodes for a specific anime.
// The animeURL should be obtained from a SearchAnime result.
func (c *Client) GetAnimeEpisodes(animeURL string, source types.Source) ([]*types.Episode, error) {
	if err := requireEpisodes(source); err != nil {
		return nil, err
	}
	scr, err := c.manager.GetScraper(source.ToScraperType())
	if err != nil {
		return nil, err
	}

	var episodes []models.Episode
	if source.IsCatalog() {
		// A movie is its one episode; a series needs GetSeasonEpisodes.
		episodes, err = catalogEpisodes(scr, animeURL, 0, "")
	} else {
		episodes, err = scr.GetAnimeEpisodes(animeURL)
	}
	if err != nil {
		return nil, err
	}

	return types.FromInternalEpisodeList(episodes), nil
}

// GetStreamURL retrieves the streaming URL and headers for a specific episode.
// The episodeURL should be obtained from GetAnimeEpisodes.
//
// Deprecated: Use GetEpisodeStreamURL instead for better control over quality and mode.
func (c *Client) GetStreamURL(episodeURL string, source types.Source, options ...any) (streamURL string, metadata map[string]string, err error) {
	if err := requireEpisodes(source); err != nil {
		return "", nil, err
	}
	scr, err := c.manager.GetScraper(source.ToScraperType())
	if err != nil {
		return "", nil, err
	}

	if source.IsCatalog() {
		return catalogStream(scr, "", models.Episode{URL: episodeURL}, "")
	}
	return scr.GetStreamURL(episodeURL, options...)
}

// StreamOptions contains options for retrieving stream URLs
type StreamOptions struct {
	// Quality can be "best", "worst", "1080p", "720p", "480p", "360p"
	Quality string
	// Mode can be "sub" (subtitled) or "dub" (dubbed)
	Mode string
	// Audio is the audio of a movie or series from a catalogue source
	// (Source.IsCatalog): "Dublado" or "Legendado", or empty for no
	// preference. An audio that does not play is ErrAudioUnavailable; the
	// other audio is never played in its place. Anime sources ignore it.
	Audio string
}

// DefaultStreamOptions returns default stream options
func DefaultStreamOptions() StreamOptions {
	return StreamOptions{
		Quality: "best",
		Mode:    "sub",
	}
}

// GetEpisodeStreamURL retrieves the streaming URL for a specific episode with full control.
// This is the recommended method to get playback URLs.
//
// Parameters:
//   - anime: The anime object from SearchAnime
//   - episode: The episode object from GetAnimeEpisodes
//   - options: Audio is applied for catalogue sources (see StreamOptions); the
//     rest is accepted for forward compatibility and not applied yet: the
//     stream comes back in the source's default quality
//
// Returns:
//   - streamURL: Direct URL for video playback
//   - metadata: Additional info like quality, source, etc.
//   - error: Any error that occurred
func (c *Client) GetEpisodeStreamURL(anime *types.Anime, episode *types.Episode, opts *StreamOptions) (streamURL string, metadata map[string]string, err error) {
	if anime == nil || episode == nil {
		return "", nil, fmt.Errorf("no title or episode given")
	}
	source, err := types.ParseSource(anime.Source)
	if err != nil {
		return "", nil, err
	}
	if err := requireEpisodes(source); err != nil {
		return "", nil, err
	}

	scr, err := c.manager.GetScraper(source.ToScraperType())
	if err != nil {
		return "", nil, err
	}

	if source.IsCatalog() {
		audio := ""
		if opts != nil {
			audio = opts.Audio
		}
		return catalogStream(scr, anime.URL, models.Episode{
			Number: episode.Number, Num: episode.Num, URL: episode.URL,
			SeasonID: strconv.Itoa(episode.Season), Audio: episode.Audio,
		}, audio)
	}

	// The episode URL is the direct episode page. (The AllAnime branch that
	// passed an anime ID + episode number instead went with that source.)
	return scr.GetStreamURL(episode.URL)
}

// catalogTimeout bounds one catalogue call; the client has no context to carry.
const catalogTimeout = 45 * time.Second

// catalogStreamTimeout covers the panel's servers and the fallback, 45s each.
const catalogStreamTimeout = 100 * time.Second

func asCatalog(scr scraper.UnifiedScraper) (scraper.CatalogScraper, error) {
	cat, ok := scr.(scraper.CatalogScraper)
	if !ok {
		return nil, ErrEpisodesUnsupported
	}
	return cat, nil
}

func catalogEpisodes(scr scraper.UnifiedScraper, titleURL string, season int, audio string) ([]models.Episode, error) {
	cat, err := asCatalog(scr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), catalogTimeout)
	defer cancel()
	return cat.CatalogEpisodes(ctx, titleURL, season, audio)
}

// catalogStream resolves an episode. titleURL is the title page the fallback
// needs ("" when the caller only has the episode URL: no fallback then).
func catalogStream(scr scraper.UnifiedScraper, titleURL string, ep models.Episode, audio string) (string, map[string]string, error) {
	cat, err := asCatalog(scr)
	if err != nil {
		return "", nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), catalogStreamTimeout)
	defer cancel()
	return cat.CatalogStream(ctx, titleURL, ep, audio)
}

// GetSeasons lists the seasons of a series from a catalogue source
// (Source.IsCatalog), with the audios each is offered in. A movie has none:
// the list is empty, and GetAnimeEpisodes returns the movie itself.
// anime must be a SearchAnime result of that source.
func (c *Client) GetSeasons(anime *types.Anime) ([]*types.Season, error) {
	_, scr, err := c.catalogFor(anime)
	if err != nil {
		return nil, err
	}
	cat, err := asCatalog(scr)
	if err != nil {
		return nil, err
	}
	ctx, cancel := context.WithTimeout(context.Background(), catalogTimeout)
	defer cancel()
	seasons, err := cat.CatalogSeasons(ctx, anime.URL)
	if err != nil {
		return nil, err
	}
	out := make([]*types.Season, len(seasons))
	for i, s := range seasons {
		out[i] = &types.Season{Number: s.Number, Audios: s.Audios}
	}
	return out, nil
}

// GetSeasonEpisodes lists one season of a series in one audio ("Dublado" or
// "Legendado"; empty takes the dub when the season has one). A season the
// series lacks is ErrSeasonNotFound, an audio the season lacks is
// ErrAudioUnavailable, and both errors say what exists. For a movie, season
// and audio are ignored and the movie is the one episode.
func (c *Client) GetSeasonEpisodes(anime *types.Anime, season int, audio string) ([]*types.Episode, error) {
	_, scr, err := c.catalogFor(anime)
	if err != nil {
		return nil, err
	}
	episodes, err := catalogEpisodes(scr, anime.URL, season, audio)
	if err != nil {
		return nil, err
	}
	return types.FromInternalEpisodeList(episodes), nil
}

// catalogFor resolves the catalogue source of a search result.
func (c *Client) catalogFor(anime *types.Anime) (types.Source, scraper.UnifiedScraper, error) {
	if anime == nil {
		return 0, nil, fmt.Errorf("no title given")
	}
	source, err := types.ParseSource(anime.Source)
	if err != nil {
		return 0, nil, err
	}
	if !source.IsCatalog() {
		return 0, nil, fmt.Errorf("%s: %w", source, ErrNotCatalog)
	}
	scr, err := c.manager.GetScraper(source.ToScraperType())
	if err != nil {
		return 0, nil, err
	}
	return source, scr, nil
}

// requireEpisodes refuses a source that can only be searched.
func requireEpisodes(source types.Source) error {
	if !source.SupportsEpisodes() {
		return fmt.Errorf("%s: %w", source, ErrEpisodesUnsupported)
	}
	return nil
}

// NewClientForTest creates a Client backed by a custom Manager. Only for tests.
func NewClientForTest(manager Manager) *Client {
	return &Client{manager: manager}
}

// GetAvailableSources returns the sources whose episode lists and stream URLs
// the client can fetch. types.Sources lists every source that can be searched.
func (c *Client) GetAvailableSources() []types.Source {
	var out []types.Source
	for _, s := range types.Sources() {
		if s.SupportsEpisodes() {
			out = append(out, s)
		}
	}
	return out
}
