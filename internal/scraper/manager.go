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
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/anidb"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/animefire"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/goyabu"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
)

// ScraperType represents different scraper types
type ScraperType int

const (
	AnimefireType ScraperType = iota
	GoyabuType                // PT-BR anime source
	SuperFlixType             // SuperFlix PT-BR movies/series/animes/doramas
	AniDBType                 // anidb.app — subbed/dubbed HLS
)

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
	case SuperFlixType:
		return &SuperFlixAdapter{client: superflix.NewSuperFlixClient()}, nil
	case AniDBType:
		return &AniDBAdapter{client: anidb.NewAniDBClient()}, nil
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
	case SuperFlixType:
		return "SuperFlix"
	case AniDBType:
		return "AniDB"
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
	case SuperFlixType:
		return "[PT-BR]"
	case AniDBType:
		return "[English]"
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

// AniDBAdapter adapts anidb.AniDBClient to the UnifiedScraper interface.
type AniDBAdapter struct {
	client *anidb.AniDBClient
}

// The *Context methods are the real implementations; the UnifiedScraper ones
// below delegate with context.Background() because that interface has nowhere
// to put a context. Callers that have one should use ContextualScraper.
func (a *AniDBAdapter) SearchAnimeContext(ctx context.Context, query string, _ ...any) ([]*models.Anime, error) {
	return a.client.SearchAnime(ctx, query)
}

func (a *AniDBAdapter) GetAnimeEpisodesContext(ctx context.Context, animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodes(ctx, animeURL)
}

// GetStreamURLContext accepts an optional quality string ("best", "1080p", …)
// as the first variadic option, matching how the other adapters take theirs.
func (a *AniDBAdapter) GetStreamURLContext(ctx context.Context, episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	return a.client.GetEpisodeStreamURL(ctx, episodeURL, qualityOption(options))
}

func (a *AniDBAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.SearchAnimeContext(context.Background(), query, options...)
}

func (a *AniDBAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.GetAnimeEpisodesContext(context.Background(), animeURL)
}

func (a *AniDBAdapter) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
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

func (a *AniDBAdapter) GetType() ScraperType {
	return AniDBType
}

// NewAniDBAdapterForTest builds an AniDB adapter whose client talks to a test
// server instead of anidb.app. It exists so the registry-level cascade test can
// drive the real adapter glue offline; without it a test can only fake the
// adapter, and an adapter bug would go unnoticed. Only for tests.
func NewAniDBAdapterForTest(serverURL string) UnifiedScraper {
	return &AniDBAdapter{client: anidb.NewClientForTest(serverURL)}
}

// SuperFlixAdapter adapts superflix.SuperFlixClient to UnifiedScraper interface
type SuperFlixAdapter struct {
	client *superflix.SuperFlixClient
}

func (a *SuperFlixAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	media, err := a.client.SearchMedia(query)
	if err != nil {
		return nil, err
	}

	var animes []*models.Anime
	for _, m := range media {
		animes = append(animes, m.ToAnimeModel())
	}
	return animes, nil
}

func (a *SuperFlixAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	// For SuperFlix, animeURL contains the TMDB ID
	return nil, fmt.Errorf("for SuperFlix, use GetSuperFlixEpisodes in enhanced.go")
}

func (a *SuperFlixAdapter) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	// episodeURL = TMDB ID
	// options[0] = media type ("filme" or "serie")
	// options[1] = season (optional)
	// options[2] = episode number (optional)
	mediaType := "filme"
	season := ""
	episode := ""

	if len(options) > 0 {
		if s, ok := options[0].(string); ok {
			mediaType = s
		}
	}
	if len(options) > 1 {
		if s, ok := options[1].(string); ok {
			season = s
		}
	}
	if len(options) > 2 {
		if s, ok := options[2].(string); ok {
			episode = s
		}
	}

	// Generous timeout: the first request in the pipeline may hit a Cloudflare
	// Turnstile gate the client solves with a headed Firefox (10–40s); a
	// shorter deadline cancels the solve mid-flight.
	ctx, cancel := context.WithTimeout(context.Background(), 210*time.Second)
	defer cancel()

	result, err := a.client.GetStreamURL(ctx, mediaType, episodeURL, season, episode)
	if err != nil {
		return "", nil, err
	}

	metadata = make(map[string]string)
	metadata["source"] = "superflix"
	metadata["referer"] = result.Referer
	metadata["title"] = result.Title

	if len(result.Subtitles) > 0 {
		var subURLs, subLabels []string
		for _, sub := range result.Subtitles {
			subURLs = append(subURLs, sub.URL)
			subLabels = append(subLabels, sub.Lang)
		}
		metadata["subtitles"] = strings.Join(subURLs, ",")
		metadata["subtitle_labels"] = strings.Join(subLabels, ",")
	}

	if len(result.DefaultAudio) > 0 {
		metadata["audio_lang"] = result.DefaultAudio[0]
	}

	return result.StreamURL, metadata, nil
}

func (a *SuperFlixAdapter) GetType() ScraperType {
	return SuperFlixType
}

// GetClient returns the underlying SuperFlix client for direct access
func (a *SuperFlixAdapter) GetClient() *superflix.SuperFlixClient {
	return a.client
}

// NewSuperFlixAdapterWithClient creates a SuperFlixAdapter with a pre-configured client.
// Useful for testing with mock servers.
func NewSuperFlixAdapterWithClient(client *superflix.SuperFlixClient) *SuperFlixAdapter {
	return &SuperFlixAdapter{client: client}
}
