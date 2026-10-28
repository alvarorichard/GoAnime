// Package scraper provides per-source adapters over a unified interface.
//
// The multi-source search/dispatch engine that used to live here (ScraperManager)
// has been replaced by the Model B registry in internal/api/providers: sources
// self-register and are fanned out by providers.SearchAll / source.Resolve. What
// remains here is the adapter layer — thin wrappers that expose each per-source
// client through the UnifiedScraper interface — plus NewAdapter, which the
// providers build lazily and own directly.
package scraper

import (
	"context"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/animedrive"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/animefire"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/animesdigital"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/anroll"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/goyabu"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/hianime"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pobreflix"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/topcine"
	"github.com/alvarorichard/Goanime/v2/internal/util/jsonx"
)

// ScraperType represents different scraper types
type ScraperType int

const (
	AnimefireType     ScraperType = iota
	GoyabuType                    // PT-BR anime source
	HiAnimeType                   // hianime.at — subbed/dubbed HLS
	StartFlixType                 // StartFlix PT-BR movies/series
	TopCineType                   // TopCine PT-BR movies/series, played through the StartFlix panel
	AnimesDigitalType             // AnimesDigital PT-BR anime source
	AnimeDriveType                // AnimeDrive — PT-BR anime source

	PobreflixType // Pobreflix PT-BR movies/series; direct MP4s, panel fallback

	// NOTE: appended at the end on purpose — the iota values stitch into
	// serialized/tracked positions; reordering the middle ones would migrate
	// user history.
	AnrollType // AnimesROLL PT-BR anime source
)

// SearchContextScraper is the optional search-only cancellation capability.
// UnifiedScraper predates context, so providers use this when available.
type SearchContextScraper interface {
	SearchAnimeContext(ctx context.Context, query string, options ...any) ([]*models.Anime, error)
}

// ContextualScraper is the optional capability (Model C: discovered by type
// assertion, never by a flag) for adapters whose client honours cancellation
// end to end.
//
// UnifiedScraper predates context and cannot carry one; dispatch.go documents
// the consequence — "the underlying adapter clients still issue non-cancelable
// http.NewRequest calls", so a per-source deadline can fire while the request it
// was meant to abort keeps running. Rather than widen the interface (which
// would touch every adapter at once), an adapter that CAN be cancelled declares
// it here, and a caller holding a context uses it when present.
type ContextualScraper interface {
	SearchAnimeContext(ctx context.Context, query string, options ...any) ([]*models.Anime, error)
	GetAnimeEpisodesContext(ctx context.Context, animeURL string) ([]models.Episode, error)
	GetStreamURLContext(ctx context.Context, episodeURL string, options ...any) (string, map[string]string, error)
}

// UnifiedScraper provides a common interface for all scrapers
type UnifiedScraper interface {
	SearchAnime(query string, options ...any) ([]*models.Anime, error)
	GetAnimeEpisodes(animeURL string) ([]models.Episode, error)
	GetStreamURL(episodeURL string, options ...any) (string, map[string]string, error)
	GetType() ScraperType
}

// NewAdapter constructs a standalone UnifiedScraper adapter for the given type,
// wrapping a freshly-built per-source client. It lets the Model B providers own
// their scraper directly; the clients are cheap, lazy structs so construction
// does no network I/O.
func NewAdapter(t ScraperType) (UnifiedScraper, error) {
	switch t {
	case AnimefireType:
		return &AnimefireAdapter{client: animefire.NewAnimefireClient()}, nil
	case GoyabuType:
		return &GoyabuAdapter{client: goyabu.NewGoyabuClient()}, nil
	case HiAnimeType:
		return &HiAnimeAdapter{client: hianime.NewHiAnimeClient()}, nil
	case StartFlixType:
		return &StartFlixAdapter{client: startflix.Shared()}, nil
	case TopCineType:
		return &TopCineAdapter{client: topcine.Shared(), panel: startflix.Shared()}, nil
	case AnimesDigitalType:
		return &AnimesDigitalAdapter{client: animesdigital.Shared()}, nil
	case AnimeDriveType:
		return &AnimeDriveAdapter{client: animedrive.NewAnimeDriveClient()}, nil
	case PobreflixType:
		return &PobreflixAdapter{client: pobreflix.Shared(), panel: startflix.Shared()}, nil
	case AnrollType:
		return &AnrollAdapter{client: anroll.Shared()}, nil
	default:
		return nil, fmt.Errorf("no adapter for scraper type %v", t)
	}
}

// scraperDisplayName returns a stable display name for the scraper type. It is
// the canonical Source spelling used by result tagging and diagnostics.
func scraperDisplayName(scraperType ScraperType) string {
	switch scraperType {
	case AnimefireType:
		return "Animefire.io"
	case GoyabuType:
		return "Goyabu"
	case HiAnimeType:
		return "HiAnime"
	case StartFlixType:
		return startflix.SourceName
	case TopCineType:
		return topcine.SourceName
	case AnimesDigitalType:
		return animesdigital.SourceName
	case AnimeDriveType:
		return "AnimeDrive"
	case PobreflixType:
		return pobreflix.SourceName
	case AnrollType:
		return anroll.SourceName
	default:
		return "Desconhecido"
	}
}

// scraperLanguageTag returns the language tag prefix for a source.
func scraperLanguageTag(scraperType ScraperType) string {
	switch scraperType {
	case AnimefireType:
		return "[PT-BR]"
	case GoyabuType:
		return "[PT-BR]"
	case HiAnimeType:
		return "[English]"
	case StartFlixType, TopCineType, AnimesDigitalType, AnimeDriveType, PobreflixType, AnrollType:
		return "[PT-BR]"
	default:
		return "[Unknown]"
	}
}

// sortPTBRFirst reorders results so that PT-BR entries appear before all others,
// preserving the relative order within each group.
func sortPTBRFirst(results []*models.Anime) {
	sort.SliceStable(results, func(i, j int) bool {
		iPTBR := strings.Contains(results[i].Name, "[PT-BR]")
		jPTBR := strings.Contains(results[j].Name, "[PT-BR]")
		// PT-BR entries come first; within the same group, keep original order.
		return iPTBR && !jPTBR
	})
}

// ptbr* are compiled regexes for cleaning PT-BR anime titles
var (
	ptbrSpaceRe      = regexp.MustCompile(`\s+`)
	ptbrAgeRatingRe  = regexp.MustCompile(`\bA\d{2}\b`)
	ptbrNumRatingRe  = regexp.MustCompile(`\b\d+[.,]\d+\b|\bN/A\b`)
	ptbrTypeSuffixRe = regexp.MustCompile(`(?i)\s*\((TV\s*Short|TV|Movie|OVA|ONA|Special|Filme|Especial|Longa-?Metragem)\)`)
	ptbrDubLegRe     = regexp.MustCompile(`(?i)\s*[\(\[]?(dublado|legendado)[\)\]]?`)
)

// cleanPTBRTitle removes noise from PT-BR anime titles such as ratings ("8.39"),
// age ratings ("A16"), type suffixes ("(TV)"), and extra whitespace.
func cleanPTBRTitle(title string) string {
	// Strip dublado/legendado labels — they will be re-added by tagging
	title = ptbrDubLegRe.ReplaceAllString(title, "")

	// Normalise whitespace (handles newlines / tabs from goquery.Text())
	title = ptbrSpaceRe.ReplaceAllString(strings.TrimSpace(title), " ")

	// Remove age ratings like A14, A16, A18
	title = ptbrAgeRatingRe.ReplaceAllString(title, "")

	// Remove numeric ratings like 8.39, N/A
	title = ptbrNumRatingRe.ReplaceAllString(title, "")

	// Remove media-type suffixes like (TV), (Movie), (OVA)
	title = ptbrTypeSuffixRe.ReplaceAllString(title, "")

	// Final whitespace cleanup
	title = strings.TrimSpace(ptbrSpaceRe.ReplaceAllString(title, " "))

	return title
}

// needsMediaTypeDisambig pre-scans results and returns a set of lowercased
// titles that appear with more than one MediaType in the batch. Only those
// entries need an explicit [Movie]/[TV] disambiguation tag.
func needsMediaTypeDisambig(results []*models.Anime) map[string]bool {
	titleTypes := make(map[string]models.MediaType, len(results))
	ambiguous := make(map[string]bool)
	for _, a := range results {
		key := strings.ToLower(strings.TrimSpace(a.Name))
		if prev, exists := titleTypes[key]; exists {
			if prev != a.MediaType {
				ambiguous[key] = true
			}
		} else {
			titleTypes[key] = a.MediaType
		}
	}
	return ambiguous
}

// AnimefireAdapter adapts animefire.AnimefireClient to UnifiedScraper interface
type AnimefireAdapter struct {
	client *animefire.AnimefireClient
}

func (a *AnimefireAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.client.SearchAnime(query)
}

func (a *AnimefireAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodes(animeURL)
}

func (a *AnimefireAdapter) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	url, err := a.client.GetEpisodeStreamURL(episodeURL)
	metadata = make(map[string]string)
	metadata["source"] = "animefire"
	return url, metadata, err
}

func (a *AnimefireAdapter) GetType() ScraperType {
	return AnimefireType
}

// AnimeDriveAdapter adapts animedrive.AnimeDriveClient to UnifiedScraper interface
type AnimeDriveAdapter struct {
	client *animedrive.AnimeDriveClient
}

func (a *AnimeDriveAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.client.SearchAnime(query)
}

// SearchAnimeContext searches AnimeDrive with the supplied context, allowing
// cancellation of its HTTP requests and retry delays.
func (a *AnimeDriveAdapter) SearchAnimeContext(ctx context.Context, query string, _ ...any) ([]*models.Anime, error) {
	return a.client.SearchAnimeContext(ctx, query)
}

func (a *AnimeDriveAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodes(animeURL)
}

func (a *AnimeDriveAdapter) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	quality := "best"
	for _, option := range options {
		if value, ok := option.(string); ok {
			quality = value
			break
		}
	}
	url, meta, err := a.client.GetStreamURLWithQuality(episodeURL, quality)
	metadata = meta
	if metadata == nil {
		metadata = make(map[string]string)
	}
	metadata["source"] = "animedrive"
	if err != nil {
		return url, metadata, err
	}
	return url, metadata, nil
}

func (a *AnimeDriveAdapter) GetType() ScraperType {
	return AnimeDriveType
}

// NewAnimeDriveAdapterForTest builds an AnimeDrive adapter whose client talks
// to a test server instead of animesdrive.cloud. It exists so registry-level
// cascade tests can drive the real adapter glue offline; without it a test can
// only fake the adapter, and an adapter bug would go unnoticed. Only for tests.
func NewAnimeDriveAdapterForTest(serverURL string) UnifiedScraper {
	return &AnimeDriveAdapter{client: animedrive.NewClientForTest(serverURL)}
}

// GoyabuAdapter adapts goyabu.GoyabuClient to UnifiedScraper interface
type GoyabuAdapter struct {
	client *goyabu.GoyabuClient
}

func (a *GoyabuAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.client.SearchAnime(query)
}

func (a *GoyabuAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodes(animeURL)
}

func (a *GoyabuAdapter) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	url, err := a.client.GetEpisodeStreamURL(episodeURL)
	metadata = make(map[string]string)
	metadata["source"] = "goyabu"
	return url, metadata, err
}

func (a *GoyabuAdapter) GetType() ScraperType {
	return GoyabuType
}

// AnimesDigitalAdapter adapts the plain-HTTP AnimesDigital client.
type AnimesDigitalAdapter struct {
	client *animesdigital.Client
}

func (a *AnimesDigitalAdapter) SearchAnimeContext(ctx context.Context, query string, _ ...any) ([]*models.Anime, error) {
	return a.client.SearchAnimeContext(ctx, query)
}

func (a *AnimesDigitalAdapter) GetAnimeEpisodesContext(ctx context.Context, animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodesContext(ctx, animeURL)
}

func (a *AnimesDigitalAdapter) GetStreamURLContext(ctx context.Context, episodeURL string, _ ...any) (streamURL string, metadata map[string]string, err error) {
	streamURL, err = a.client.GetEpisodeStreamURLContext(ctx, episodeURL)
	return streamURL, map[string]string{"source": "animesdigital"}, err
}

func (a *AnimesDigitalAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.SearchAnimeContext(context.Background(), query, options...)
}

func (a *AnimesDigitalAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.GetAnimeEpisodesContext(context.Background(), animeURL)
}

func (a *AnimesDigitalAdapter) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	return a.GetStreamURLContext(context.Background(), episodeURL, options...)
}

func (a *AnimesDigitalAdapter) GetType() ScraperType { return AnimesDigitalType }

// NewAnimesDigitalAdapterWithClient builds an adapter around a test client.
func NewAnimesDigitalAdapterWithClient(client *animesdigital.Client) *AnimesDigitalAdapter {
	return &AnimesDigitalAdapter{client: client}
}

// AnrollAdapter adapts the plain-HTTP AnimesROLL client.
type AnrollAdapter struct {
	client *anroll.Client
}

// SearchAnimeContext searches AnimesROLL for query.
func (a *AnrollAdapter) SearchAnimeContext(ctx context.Context, query string, _ ...any) ([]*models.Anime, error) {
	return a.client.SearchAnimeContext(ctx, query)
}

// GetAnimeEpisodesContext lists the episodes of an AnimesROLL title.
func (a *AnrollAdapter) GetAnimeEpisodesContext(ctx context.Context, animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodesContext(ctx, animeURL)
}

// GetStreamURLContext resolves an episode to its stream URL and the
// Referer/Origin the CDN expects.
func (a *AnrollAdapter) GetStreamURLContext(ctx context.Context, episodeURL string, _ ...any) (streamURL string, metadata map[string]string, err error) {
	stream, err := a.client.GetEpisodeStreamContext(ctx, episodeURL)
	if err != nil {
		return "", nil, err
	}
	metadata = map[string]string{"source": "anroll"}
	if stream.Referer != "" {
		metadata["referer"] = stream.Referer
	}
	if stream.Origin != "" {
		metadata["origin"] = stream.Origin
	}
	return stream.URL, metadata, nil
}

// SearchAnime is SearchAnimeContext with a background context.
func (a *AnrollAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.SearchAnimeContext(context.Background(), query, options...)
}

// GetAnimeEpisodes is GetAnimeEpisodesContext with a background context.
func (a *AnrollAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.GetAnimeEpisodesContext(context.Background(), animeURL)
}

// GetStreamURL is GetStreamURLContext with a background context.
func (a *AnrollAdapter) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	return a.GetStreamURLContext(context.Background(), episodeURL, options...)
}

// GetType reports AnrollType.
func (a *AnrollAdapter) GetType() ScraperType { return AnrollType }

// NewAnrollAdapterWithClient builds an adapter around a test client.
func NewAnrollAdapterWithClient(client *anroll.Client) *AnrollAdapter {
	return &AnrollAdapter{client: client}
}

// HiAnimeAdapter adapts hianime.HiAnimeClient to the UnifiedScraper interface.
type HiAnimeAdapter struct {
	client *hianime.HiAnimeClient
}

// The *Context methods are the real implementations; the UnifiedScraper ones
// below delegate with context.Background() because that interface has nowhere
// to put a context. Callers that have one should use ContextualScraper.
func (a *HiAnimeAdapter) SearchAnimeContext(ctx context.Context, query string, _ ...any) ([]*models.Anime, error) {
	return a.client.SearchAnime(ctx, query)
}

func (a *HiAnimeAdapter) GetAnimeEpisodesContext(ctx context.Context, animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodes(ctx, animeURL)
}

// GetStreamURLContext accepts an optional quality string ("best", "1080p", …)
// as the first variadic option, matching how the other adapters take theirs.
func (a *HiAnimeAdapter) GetStreamURLContext(ctx context.Context, episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	return a.client.GetEpisodeStreamURL(ctx, episodeURL, qualityOption(options))
}

func (a *HiAnimeAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.SearchAnimeContext(context.Background(), query, options...)
}

func (a *HiAnimeAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.GetAnimeEpisodesContext(context.Background(), animeURL)
}

func (a *HiAnimeAdapter) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	return a.GetStreamURLContext(context.Background(), episodeURL, options...)
}

// qualityOption reads the optional quality string from a variadic option list.
func qualityOption(options []any) string {
	if len(options) > 0 {
		if q, ok := options[0].(string); ok && q != "" {
			return q
		}
	}
	return "best"
}

func (a *HiAnimeAdapter) GetType() ScraperType {
	return HiAnimeType
}

// NewHiAnimeAdapterForTest builds an HiAnime adapter whose client talks to a test
// server instead of hianime.at. It exists so the registry-level cascade test can
// drive the real adapter glue offline; without it a test can only fake the
// adapter, and an adapter bug would go unnoticed. Only for tests.
func NewHiAnimeAdapterForTest(serverURL string) UnifiedScraper {
	return &HiAnimeAdapter{client: hianime.NewClientForTest(serverURL)}
}

// StartFlixAdapter adapts startflix.Client to the UnifiedScraper interface.
//
// The episode "URL" it streams from is the panel endpoint that lists the
// players — /episodio/<id> for an episode, /filme/<imdb> for a movie — so a
// stream resolves from the episode alone, which is what batch downloads need.
type StartFlixAdapter struct {
	client *startflix.Client
}

func (a *StartFlixAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.SearchAnimeContext(context.Background(), query, options...)
}

// SearchAnimeContext makes StartFlixAdapter a ContextualScraper: every request
// the client makes carries the caller's context.
func (a *StartFlixAdapter) SearchAnimeContext(ctx context.Context, query string, _ ...any) ([]*models.Anime, error) {
	media, err := a.client.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	animes := make([]*models.Anime, 0, len(media))
	for _, m := range media {
		animes = append(animes, m.ToAnimeModel())
	}
	return animes, nil
}

// GetAnimeEpisodesContext exists so the adapter satisfies ContextualScraper as
// a whole. StartFlix lists episodes through the season and audio pickers in
// the api package, not through this adapter.
func (a *StartFlixAdapter) GetAnimeEpisodesContext(_ context.Context, animeURL string) ([]models.Episode, error) {
	return a.GetAnimeEpisodes(animeURL)
}

func (a *StartFlixAdapter) GetAnimeEpisodes(string) ([]models.Episode, error) {
	return nil, fmt.Errorf("for StartFlix, use GetStartFlixEpisodes in the api package")
}

func (a *StartFlixAdapter) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	return a.GetStreamURLContext(context.Background(), episodeURL, options...)
}

// GetStreamURLContext resolves the first playable server listed at episodeURL.
// Metadata follows the providers.applyPlaybackMetadata contract: "referer",
// and "subtitles" as a JSON array of {"url","language","label"}.
func (a *StartFlixAdapter) GetStreamURLContext(ctx context.Context, episodeURL string, _ ...any) (streamURL string, metadata map[string]string, err error) {
	stream, err := a.client.Stream(ctx, episodeURL)
	if err != nil {
		return "", nil, err
	}
	metadata = map[string]string{
		"source":  "startflix",
		"referer": stream.Referer,
		"host":    stream.Host,
	}
	if len(stream.Subtitles) > 0 {
		tracks := make([]map[string]string, 0, len(stream.Subtitles))
		for _, sub := range stream.Subtitles {
			tracks = append(tracks, map[string]string{"url": sub.URL, "language": sub.Language, "label": sub.Label})
		}
		if raw, err := jsonx.Marshal(tracks); err == nil {
			metadata["subtitles"] = string(raw)
		}
	}
	return stream.URL, metadata, nil
}

func (a *StartFlixAdapter) GetType() ScraperType {
	return StartFlixType
}

// GetClient returns the underlying StartFlix client.
func (a *StartFlixAdapter) GetClient() *startflix.Client {
	return a.client
}

// NewStartFlixAdapterWithClient builds an adapter around a given client, for
// tests that point it at a mock server.
func NewStartFlixAdapterWithClient(client *startflix.Client) *StartFlixAdapter {
	return &StartFlixAdapter{client: client}
}

// TopCineAdapter adapts topcine.Client to the UnifiedScraper interface.
//
// TopCine's own player host is behind a Cloudflare Turnstile, so its titles
// play by TMDB id through the StartFlix panel: the episode "URL" it streams
// from is that panel's player list, exactly as for StartFlixAdapter.
type TopCineAdapter struct {
	client *topcine.Client
	panel  *startflix.Client
}

func (a *TopCineAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.SearchAnimeContext(context.Background(), query, options...)
}

// SearchAnimeContext makes TopCineAdapter a ContextualScraper.
func (a *TopCineAdapter) SearchAnimeContext(ctx context.Context, query string, _ ...any) ([]*models.Anime, error) {
	media, err := a.client.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	animes := make([]*models.Anime, 0, len(media))
	for _, m := range media {
		animes = append(animes, m.ToAnimeModel())
	}
	return animes, nil
}

// GetAnimeEpisodesContext exists so the adapter satisfies ContextualScraper as
// a whole. TopCine lists episodes through the season and audio pickers in the
// api package, not through this adapter.
func (a *TopCineAdapter) GetAnimeEpisodesContext(_ context.Context, animeURL string) ([]models.Episode, error) {
	return a.GetAnimeEpisodes(animeURL)
}

func (a *TopCineAdapter) GetAnimeEpisodes(string) ([]models.Episode, error) {
	return nil, fmt.Errorf("for TopCine, use GetTopCineEpisodes in the api package")
}

func (a *TopCineAdapter) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	return a.GetStreamURLContext(context.Background(), episodeURL, options...)
}

// GetStreamURLContext resolves the panel player list at episodeURL; the
// metadata contract is StartFlixAdapter's.
func (a *TopCineAdapter) GetStreamURLContext(ctx context.Context, episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	streamURL, metadata, err = (&StartFlixAdapter{client: a.panel}).GetStreamURLContext(ctx, episodeURL, options...)
	if metadata != nil {
		metadata["source"] = "topcine"
	}
	return streamURL, metadata, err
}

func (a *TopCineAdapter) GetType() ScraperType {
	return TopCineType
}

// GetClient returns the underlying TopCine client.
func (a *TopCineAdapter) GetClient() *topcine.Client {
	return a.client
}

// NewTopCineAdapterWithClients builds an adapter around given clients, for
// tests that point them at a mock server.
func NewTopCineAdapterWithClients(client *topcine.Client, panel *startflix.Client) *TopCineAdapter {
	return &TopCineAdapter{client: client, panel: panel}
}

// PobreflixAdapter adapts pobreflix.Client to the UnifiedScraper interface.
//
// Pobreflix's player is assembled client-side by an obfuscated bundle, so its
// titles play by TMDB id through the StartFlix panel, exactly as for
// TopCineAdapter. The catalog side — search over the sitemap corpus, title
// pages, season and episode names — is Pobreflix's own.
type PobreflixAdapter struct {
	client *pobreflix.Client
	panel  *startflix.Client
}

func (a *PobreflixAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.SearchAnimeContext(context.Background(), query, options...)
}

// SearchAnimeContext makes PobreflixAdapter a ContextualScraper.
func (a *PobreflixAdapter) SearchAnimeContext(ctx context.Context, query string, _ ...any) ([]*models.Anime, error) {
	media, err := a.client.Search(ctx, query)
	if err != nil {
		return nil, err
	}
	animes := make([]*models.Anime, 0, len(media))
	for _, m := range media {
		animes = append(animes, m.ToAnimeModel())
	}
	return animes, nil
}

// GetAnimeEpisodesContext exists so the adapter satisfies ContextualScraper as
// a whole. Pobreflix lists episodes through the season and audio pickers in
// the api package, not through this adapter.
func (a *PobreflixAdapter) GetAnimeEpisodesContext(_ context.Context, animeURL string) ([]models.Episode, error) {
	return a.GetAnimeEpisodes(animeURL)
}

func (a *PobreflixAdapter) GetAnimeEpisodes(string) ([]models.Episode, error) {
	return nil, fmt.Errorf("for Pobreflix, use GetPobreflixEpisodes in the api package")
}

func (a *PobreflixAdapter) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	return a.GetStreamURLContext(context.Background(), episodeURL, options...)
}

// GetStreamURLContext resolves the panel player list at episodeURL; the
// metadata contract is StartFlixAdapter's.
func (a *PobreflixAdapter) GetStreamURLContext(ctx context.Context, episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	streamURL, metadata, err = (&StartFlixAdapter{client: a.panel}).GetStreamURLContext(ctx, episodeURL, options...)
	if metadata != nil {
		metadata["source"] = "pobreflix"
	}
	return streamURL, metadata, err
}

func (a *PobreflixAdapter) GetType() ScraperType {
	return PobreflixType
}

// GetClient returns the underlying Pobreflix client.
func (a *PobreflixAdapter) GetClient() *pobreflix.Client {
	return a.client
}

// NewPobreflixAdapterWithClients builds an adapter around given clients, for
// tests that point them at a mock server.
func NewPobreflixAdapterWithClients(client *pobreflix.Client, panel *startflix.Client) *PobreflixAdapter {
	return &PobreflixAdapter{client: client, panel: panel}
}
