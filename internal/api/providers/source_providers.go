package providers

import (
	"context"
	"fmt"
	"strings"
	"sync"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
)

// Stream-fetch indirections. Production points at the proven api layer — the
// exact functions the legacy player dispatch called — so switching dispatch to
// the Source registry cannot change behavior. Tests swap these to avoid
// network. The api bodies migrate into per-source packages in Phase 3.
var (
	// startFlixStreamFn / startFlixEpisodesFn keep StartFlix delegating to the
	// api package's interactive paths (season picker, then Dublado/Legendado).
	// HiAnime/AnimeFire/Goyabu are self-contained (adapter-direct).
	startFlixStreamFn   = api.GetStartFlixStreamURL
	startFlixEpisodesFn = api.GetStartFlixEpisodes
)

// lazyGetAdapter returns a standalone adapter for a scraper type, built once and
// cached. This is how each Model B provider owns its scraper directly, without
// the ScraperManager.
type adapterSlot struct {
	scraper.UnifiedScraper
}

func lazyGetAdapter(once *sync.Once, cache *adapterSlot, st scraper.ScraperType) (scraper.UnifiedScraper, error) {
	once.Do(func() { cache.UnifiedScraper, _ = scraper.NewAdapter(st) })
	if cache.UnifiedScraper == nil {
		return nil, fmt.Errorf("no adapter for scraper type %v", st)
	}
	return cache.UnifiedScraper, nil
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

// --- AnimeFire Provider ---

type animeFireProvider struct {
	once    sync.Once
	adapter adapterSlot
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

// --- HiAnime Provider ---
//
// hianime.at is the English-language source, and the second host to hold that
// slot. AllAnime went first when mkissa.to removed the material its per-request
// key was derived from; anidb.app took over and then went to a site-wide 503
// "Under Maintenance" on every path, API included, which is where it still was
// on 2026-09-22. Priority 50 keeps it below the existing sources: it is the
// newest and least battle-tested, so it is picked only when nothing more
// specific matches.

type hianimeProvider struct {
	once    sync.Once
	adapter adapterSlot
}

func init() {
	source.Register(&hianimeProvider{})
}

func (p *hianimeProvider) scraper() (scraper.UnifiedScraper, error) {
	return lazyGetAdapter(&p.once, &p.adapter, scraper.HiAnimeType)
}

func (p *hianimeProvider) Describe() source.Descriptor {
	return source.Descriptor{
		Kind:     source.HiAnime,
		Priority: 50,
		// "anidb"/"anidb.app" stay accepted: that is what this slot was called
		// until 2026-09-22, and a --source flag in somebody's shell alias or a
		// title restored from history still spells it that way.
		Explicit: []string{"HiAnime", "hianime.at", "AniDB", "anidb.app"},
		// "[english]" moved here from the deleted AllAnime descriptor: tagging.go
		// stamps "[English]" on HiAnime results (it is the only English source
		// left), so the resolver must be able to route those names back. Without
		// it, a HiAnime result whose Source field was lost — a title restored from
		// history, say — resolved to Unknown.
		Tags:        []string{"[hianime]", "[anidb]", "[english]"},
		URLMatchers: []string{"hianime.at", "anidb.app"},
		ProbeURL:    "https://hianime.at",
	}
}

func (p *hianimeProvider) HasSeasons() bool { return false }

// The HiAnime adapter implements scraper.ContextualScraper, so every call below
// hands it the real context instead of letting a cancelled search keep an HTTP
// request alive until the client timeout. The type assertion is the Model C
// discovery pattern; the fallbacks keep this working if the capability is ever
// dropped.
func (p *hianimeProvider) Search(ctx context.Context, query string) ([]*models.Anime, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	adapter, err := p.scraper()
	if err != nil {
		return nil, err
	}
	var results []*models.Anime
	if ca, ok := adapter.(scraper.ContextualScraper); ok {
		results, err = ca.SearchAnimeContext(ctx, query)
	} else {
		results, err = adapter.SearchAnime(query)
	}
	if err != nil {
		return nil, err
	}
	tagResults(results, source.HiAnime)
	return results, nil
}

func (p *hianimeProvider) FetchEpisodes(ctx context.Context, anime *models.Anime) ([]models.Episode, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	adapter, err := p.scraper()
	if err != nil {
		return nil, err
	}
	if ca, ok := adapter.(scraper.ContextualScraper); ok {
		return ca.GetAnimeEpisodesContext(ctx, anime.URL)
	}
	return adapter.GetAnimeEpisodes(anime.URL)
}

// FetchStreamURL resolves an episode to its HLS URL. Unlike Goyabu, HiAnime
// honours the requested quality by picking the matching variant playlist.
//
// It is also the first provider here to READ the metadata its adapter returns.
// That map was being discarded, which was survivable while every source served
// its own media: HiAnime's does not. Its playlists live on a CDN that checks the
// Referer — on two of three titles sampled 2026-09-22 the variant playlist was
// 403 without one — so dropping the metadata means the URL resolves and then
// some titles play nothing.
func (p *hianimeProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
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
	var url string
	var metadata map[string]string
	if ca, ok := adapter.(scraper.ContextualScraper); ok {
		url, metadata, err = ca.GetStreamURLContext(ctx, episode.URL, quality)
	} else {
		url, metadata, err = adapter.GetStreamURL(episode.URL, quality)
	}
	if err != nil {
		return "", fmt.Errorf("hianime stream: %w", err)
	}
	if url == "" {
		return "", fmt.Errorf("empty stream URL returned from HiAnime")
	}
	applyPlaybackMetadata(metadata)
	return url, nil
}

// applyPlaybackMetadata moves a scraper's playback hints into the globals mpv
// and the downloader read.
//
// The two keys are a contract on the metadata map, not something specific to
// one source, so a future scraper can fill them and be played correctly without
// touching this function:
//
//	referer    the origin to send with every media request
//	subtitles  a JSON array of {"url","language","label"}
//
// A malformed subtitle payload is logged and dropped rather than failing the
// episode: subtitles are an enhancement, and refusing to play a video because
// its caption list did not parse would be the wrong trade.
func applyPlaybackMetadata(metadata map[string]string) {
	if referer := strings.TrimSpace(metadata["referer"]); referer != "" {
		util.SetGlobalReferer(referer)
	}
	raw := strings.TrimSpace(metadata["subtitles"])
	if raw == "" {
		return
	}
	var tracks []util.SubtitleInfo
	if err := jsonx.Unmarshal([]byte(raw), &tracks); err != nil {
		util.Debug("could not read the subtitle list from the scraper", "error", err)
		return
	}
	if len(tracks) > 0 {
		util.SetGlobalSubtitles(tracks)
	}
}

// --- Goyabu Provider ---

type goyabuProvider struct {
	once    sync.Once
	adapter adapterSlot
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
		// PARKED, not retired. Off until we find a way past the gate:
		// GOANIME_ENABLED_SOURCES=goyabu turns it back on at any time.
		//
		// Goyabu is the only source behind a bot gate. Everything else here is
		// plain HTTP; this one sits behind a Cloudflare managed challenge, and
		// GoAnime no longer ships the browser that used to try clearing it — on
		// a flagged network even that browser returned nothing. It stays disabled rather than deleted because the
		// scraper is complete and correct: it works today on a network that is
		// not flagged.
		//
		// What is known, so the next attempt does not restart from zero
		// (measured 2026-09-25):
		//
		//   - In a real browser the gate clears on its own in about ten seconds
		//     and leaves three cookies. It is not unsolvable.
		//   - Those cookies do NOT replay. Tried: the plain transport and the
		//     Chrome-impersonating one, each with the solving browser's
		//     User-Agent and with our own, plus curl with the same cookie and
		//     UA. All 403. Cloudflare is binding the clearance to something no
		//     HTTP client here presents.
		//   - Installing real Chrome (GOANIME_SF_CHROME_CHANNEL=chrome) instead
		//     of Playwright's bundled Chromium did not change that.
		//   - The site's own banner says ISPs are blocking it and recommends a
		//     VPN, and a datacenter IP reaches it unchallenged while a flagged
		//     residential one does not. The rule is reputation, not the path.
		//
		// The avenue NOT yet tried, and the most promising one: stop replaying
		// cookies and fetch through the browser itself — the solver already has
		// a context that passed the gate, so a transport that returns that
		// page's HTML would sidestep the binding entirely. It is real work
		// (every scraper call would route through it) and it was not started,
		// so it is written down rather than half-built.
		DefaultDisabled: true,
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

// --- StartFlix Provider ---
//
// StartFlix is the movie/TV source. It needs no browser: the catalog and its
// video panel are plain HTTP. What it cannot do is play every title —
// of the panel's player hosts only Byse is resolved, so a title offered solely
// on the others fails with a message naming them.

type startFlixProvider struct {
	once    sync.Once
	adapter adapterSlot
}

func init() {
	source.Register(&startFlixProvider{})
}

func (p *startFlixProvider) scraper() (scraper.UnifiedScraper, error) {
	return lazyGetAdapter(&p.once, &p.adapter, scraper.StartFlixType)
}

func (p *startFlixProvider) Describe() source.Descriptor {
	return source.Descriptor{
		Kind:     source.StartFlix,
		Priority: 40,
		Explicit: []string{"StartFlix"},
		Tags:     []string{"[startflix]"},
		// "painel-aso" is the video panel the title pages embed; an episode URL
		// points there rather than at startflix itself.
		URLMatchers: []string{"startflix", "painel-aso"},
		ProbeURL:    "https://www.startflix.biz",
	}
}

// HasSeasons: StartFlix is a movie/TV catalog organized into seasons.
func (p *startFlixProvider) HasSeasons() bool { return true }

func (p *startFlixProvider) Search(ctx context.Context, query string) ([]*models.Anime, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	adapter, err := p.scraper()
	if err != nil {
		return nil, err
	}
	var results []*models.Anime
	if ca, ok := adapter.(scraper.ContextualScraper); ok {
		results, err = ca.SearchAnimeContext(ctx, query)
	} else {
		results, err = adapter.SearchAnime(query)
	}
	if err != nil {
		return nil, err
	}
	tagResults(results, source.StartFlix)
	return results, nil
}

// FetchEpisodes runs the interactive listing (season, then audio) in the api
// package.
func (p *startFlixProvider) FetchEpisodes(ctx context.Context, anime *models.Anime) ([]models.Episode, error) {
	if err := ctx.Err(); err != nil {
		return nil, err
	}
	return startFlixEpisodesFn(anime)
}

func (p *startFlixProvider) FetchStreamURL(ctx context.Context, episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	if err := ctx.Err(); err != nil {
		return "", err
	}
	util.ClearGlobalSubtitles()
	if anime.Source != "" {
		util.SetGlobalAnimeSource(anime.Source)
	}
	return startFlixStreamFn(anime, episode, quality)
}
