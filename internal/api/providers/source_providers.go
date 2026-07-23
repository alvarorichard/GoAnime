package providers

import (
	"context"
	"fmt"
	"sync"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/alvarorichard/Goanime/internal/util"
)

// Stream-fetch indirections. Production points at the proven api layer — the
// exact functions the legacy player dispatch called — so switching dispatch to
// the Source registry cannot change behavior. Tests swap these to avoid
// network. The api bodies migrate into per-source packages in Phase 3.
var (
	// superFlixStreamFn / superFlixEpisodesFn keep SuperFlix delegating to the
	// api package's UX-heavy paths (spinner, browser preflight, season picker).
	// AllAnime/AnimeFire/Goyabu are self-contained (adapter-direct).
	superFlixStreamFn   = api.GetSuperFlixStreamURL
	superFlixEpisodesFn = api.GetSuperFlixEpisodes
)

// lazyGetAdapter returns a standalone adapter for a scraper type, built once and
// cached. This is how each Model B provider owns its scraper directly, without
// the ScraperManager.
func lazyGetAdapter(once *sync.Once, cache *scraper.UnifiedScraper, st scraper.ScraperType) (scraper.UnifiedScraper, error) {
	once.Do(func() { *cache, _ = scraper.NewAdapter(st) })
	if *cache == nil {
		return nil, fmt.Errorf("no adapter for scraper type %v", st)
	}
	return *cache, nil
}

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
	once    sync.Once
	adapter scraper.UnifiedScraper
}

func init() {
	source.Register(&allAnimeProvider{})
}

func (p *allAnimeProvider) scraper() (scraper.UnifiedScraper, error) {
	return lazyGetAdapter(&p.once, &p.adapter, scraper.AllAnimeType)
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

func (p *allAnimeProvider) HasSeasons() bool { return false }

func (p *allAnimeProvider) Search(ctx context.Context, query string) ([]*models.Anime, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	adapter, err := p.scraper()
	if err != nil {
		return nil, err
	}
	results, err := adapter.SearchAnime(query)
	if err != nil {
		return nil, err
	}
	tagResults(results, source.AllAnime)
	return results, nil
}

func (p *allAnimeProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.scraper()
	if err != nil {
		return nil, err
	}
	animeID := source.ExtractAllAnimeID(anime.URL)
	return adapter.GetAnimeEpisodes(animeID)
}

// FetchStreamURL resolves the AllAnime stream directly through the AllAnime
// adapter (self-contained — no delegation to the api package). The adapter's
// GetStreamURL drives the client's navigation-aware GetEpisodeURL and returns
// the referer in its metadata, which mpv needs; we publish it globally, exactly
// as the deleted api enhanced-navigation path did.
func (p *allAnimeProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	util.ClearGlobalSubtitles()
	if anime.Source != "" {
		util.SetGlobalAnimeSource(anime.Source)
	}
	adapter, err := p.scraper()
	if err != nil {
		return "", err
	}
	animeID := source.ExtractAllAnimeID(anime.URL)
	epNum := EpisodeNumber(episode)
	if quality == "" {
		quality = "best"
	}
	url, metadata, err := adapter.GetStreamURL(animeID, epNum, quality)
	if err != nil {
		return "", fmt.Errorf("allAnime stream: %w", err)
	}
	if url == "" {
		return "", fmt.Errorf("empty stream URL returned from AllAnime")
	}
	if ref := metadata["referer"]; ref != "" {
		util.SetGlobalReferer(ref)
	}
	return url, nil
}

// --- AnimeFire Provider ---

type animeFireProvider struct {
	once    sync.Once
	adapter scraper.UnifiedScraper
}

func init() {
	source.Register(&animeFireProvider{})
}

func (p *animeFireProvider) scraper() (scraper.UnifiedScraper, error) {
	return lazyGetAdapter(&p.once, &p.adapter, scraper.AnimefireType)
}

func (p *animeFireProvider) Describe() source.Descriptor {
	return source.Descriptor{
		Kind:        source.AnimeFire,
		Priority:    10,
		Explicit:    []string{"Animefire.io", "AnimeFire"},
		Tags:        []string{"[animefire]"},
		URLMatchers: []string{"animefire"},
		ProbeURL:    "https://animefire.io",
	}
}

func (p *animeFireProvider) HasSeasons() bool { return false }

func (p *animeFireProvider) Search(ctx context.Context, query string) ([]*models.Anime, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	adapter, err := p.scraper()
	if err != nil {
		return nil, err
	}
	results, err := adapter.SearchAnime(query)
	if err != nil {
		return nil, err
	}
	tagResults(results, source.AnimeFire)
	return results, nil
}

func (p *animeFireProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.scraper()
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

// FetchStreamURL mirrors api.GetEpisodeStreamURL's AnimeFire branch: same
// global side effects, same adapter call including the quality argument.
func (p *animeFireProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	util.ClearGlobalSubtitles()
	if anime.Source != "" {
		util.SetGlobalAnimeSource(anime.Source)
	}
	adapter, err := p.scraper()
	if err != nil {
		return "", err
	}
	if quality == "" {
		quality = "best"
	}
	url, _, err := adapter.GetStreamURL(episode.URL, quality)
	if err != nil {
		return "", fmt.Errorf("animeFire stream: %w", err)
	}
	if url == "" {
		return "", fmt.Errorf("empty stream URL returned from Animefire.io")
	}
	return url, nil
}

// --- Goyabu Provider ---

type goyabuProvider struct {
	once    sync.Once
	adapter scraper.UnifiedScraper
}

func init() {
	source.Register(&goyabuProvider{})
}

func (p *goyabuProvider) scraper() (scraper.UnifiedScraper, error) {
	return lazyGetAdapter(&p.once, &p.adapter, scraper.GoyabuType)
}

func (p *goyabuProvider) Describe() source.Descriptor {
	return source.Descriptor{
		Kind:        source.Goyabu,
		Priority:    20,
		Explicit:    []string{"Goyabu"},
		Tags:        []string{"[goyabu]"},
		URLMatchers: []string{"goyabu"},
		ProbeURL:    "https://goyabu.io",
	}
}

func (p *goyabuProvider) HasSeasons() bool { return false }

func (p *goyabuProvider) Search(ctx context.Context, query string) ([]*models.Anime, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	adapter, err := p.scraper()
	if err != nil {
		return nil, err
	}
	results, err := adapter.SearchAnime(query)
	if err != nil {
		return nil, err
	}
	tagResults(results, source.Goyabu)
	return results, nil
}

func (p *goyabuProvider) FetchEpisodes(_ context.Context, anime *models.Anime) ([]models.Episode, error) {
	adapter, err := p.scraper()
	if err != nil {
		return nil, err
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

// FetchStreamURL mirrors api.GetEpisodeStreamURL's Goyabu branch: same global
// side effects, same adapter call (Goyabu takes no quality argument).
func (p *goyabuProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, _ string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	util.ClearGlobalSubtitles()
	if anime.Source != "" {
		util.SetGlobalAnimeSource(anime.Source)
	}
	adapter, err := p.scraper()
	if err != nil {
		return "", err
	}
	url, _, err := adapter.GetStreamURL(episode.URL)
	if err != nil {
		return "", fmt.Errorf("goyabu stream: %w", err)
	}
	if url == "" {
		return "", fmt.Errorf("empty stream URL returned from Goyabu")
	}
	return url, nil
}

// --- SuperFlix Provider ---

type superFlixProvider struct {
	once    sync.Once
	adapter scraper.UnifiedScraper
}

func init() {
	source.Register(&superFlixProvider{})
}

func (p *superFlixProvider) scraper() (scraper.UnifiedScraper, error) {
	return lazyGetAdapter(&p.once, &p.adapter, scraper.SuperFlixType)
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

// HasSeasons satisfies both providers.Provider and the source.Seasoned
// capability: SuperFlix is a movie/TV catalog organized into seasons.
func (p *superFlixProvider) HasSeasons() bool { return true }

func (p *superFlixProvider) Search(ctx context.Context, query string) ([]*models.Anime, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	adapter, err := p.scraper()
	if err != nil {
		return nil, err
	}
	results, err := adapter.SearchAnime(query)
	if err != nil {
		return nil, err
	}
	tagResults(results, source.SuperFlix)
	return results, nil
}

// WarmUp satisfies the source.BrowserGated capability. SuperFlix clears a
// Cloudflare Turnstile gate with a headed browser; if there is no graphical
// display (and the user hasn't opted into headless), that solve is doomed, so
// fail fast here with a plain-language reason instead of letting the user wait
// out a browser that can never appear. The check is cheap and performs no
// eager solve — the happy path (display present) returns nil immediately.
func (p *superFlixProvider) WarmUp(ctx context.Context) error {
	if err := ctx.Err(); err != nil {
		return err
	}
	if superflix.HeadlessEnvironment() {
		return fmt.Errorf("SuperFlix needs a graphical browser to pass its \"are you human?\" check, but no screen was found — run GoAnime on your desktop session, or pass --sf-headless to try anyway")
	}
	return nil
}

// FetchEpisodes lists SuperFlix content. Unlike the anime sources, this is not
// a flat adapter call: it runs the season picker (TVmaze-first, browser
// fallback) and sets anime.CurrentSeason. Delegated to the proven api path so
// the interactive UX is byte-identical to the legacy episode switch.
func (p *superFlixProvider) FetchEpisodes(ctx context.Context, anime *models.Anime) ([]models.Episode, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return superFlixEpisodesFn(anime)
}

// FetchStreamURL mirrors api.GetEpisodeStreamURL's SuperFlix branch: the same
// entry side effects, then the full UX path (browser preflight notices,
// spinner, global referer/subtitles, friendly errors) via GetSuperFlixStreamURL.
func (p *superFlixProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	util.ClearGlobalSubtitles()
	if anime.Source != "" {
		util.SetGlobalAnimeSource(anime.Source)
	}
	return superFlixStreamFn(anime, episode, quality)
}
