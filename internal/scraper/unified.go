// Package scraper provides a unified interface for different anime sources
package scraper

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// ScraperType represents different scraper types
type ScraperType int

// Timeout configurations – we wait for ALL sources to finish (or the hard
// timeout) so that slower scrapers like SuperFlix are never silently dropped.
const (
	// searchTimeout is the maximum time to wait for all scrapers.
	searchTimeout = 15 * time.Second
	// perScraperTimeout is the timeout for individual scrapers.
	perScraperTimeout = 12 * time.Second
)

const (
	AllAnimeType ScraperType = iota
	AnimefireType
	AnimeDriveType
	FlixHQType    // Movies and TV Shows source
	SFlixType     // Alternative Movies and TV Shows source
	NineAnimeType // 9animetv.to anime source
	GoyabuType    // PT-BR anime source
	SuperFlixType // SuperFlix PT-BR movies/series/animes/doramas
)

// UnifiedScraper provides a common interface for all scrapers
type UnifiedScraper interface {
	SearchAnime(query string, options ...any) ([]*models.Anime, error)
	GetAnimeEpisodes(animeURL string) ([]models.Episode, error)
	GetStreamURL(episodeURL string, options ...any) (string, map[string]string, error)
	GetType() ScraperType
}

// ScraperManager manages multiple scrapers
type ScraperManager struct {
	scrapers  map[ScraperType]UnifiedScraper
	breaker   *sourceCircuitBreaker
	breakerMu sync.Mutex
}

// Singleton ScraperManager — scrapers are stateless HTTP clients, no need to recreate
var (
	globalScraperManager     *ScraperManager
	globalScraperManagerOnce sync.Once
)

// PreWarmScraperManager triggers background initialization of the scraper
// manager singleton so it's ready when the first search happens.
func PreWarmScraperManager() {
	go func() { NewScraperManager() }()
}

// NewScraperManager returns a cached scraper manager singleton.
// Scrapers are stateless HTTP clients so a single instance is reused.
func NewScraperManager() *ScraperManager {
	globalScraperManagerOnce.Do(func() {
		manager := &ScraperManager{
			scrapers: make(map[ScraperType]UnifiedScraper),
			breaker:  newSourceCircuitBreaker(),
		}

		// Initialize scrapers
		manager.scrapers[AllAnimeType] = &AllAnimeAdapter{client: NewAllAnimeClient()}
		manager.scrapers[AnimefireType] = &AnimefireAdapter{client: NewAnimefireClient()}
		manager.scrapers[FlixHQType] = &FlixHQAdapter{client: NewFlixHQClient()}
		manager.scrapers[SFlixType] = &SFlixAdapter{client: NewSFlixClient()}
		manager.scrapers[NineAnimeType] = &NineAnimeAdapter{client: NewNineAnimeClient()}
		manager.scrapers[GoyabuType] = &GoyabuAdapter{client: NewGoyabuClient()}
		manager.scrapers[SuperFlixType] = &SuperFlixAdapter{client: NewSuperFlixClient()}

		// AnimeDrive disabled — Cloudflare protection blocks all requests.
		// Kept on standby until a bypass/solution is found.
		// manager.scrapers[AnimeDriveType] = &AnimeDriveAdapter{client: NewAnimeDriveClient()}

		globalScraperManager = manager
	})
	return globalScraperManager
}

// SearchAnime searches across all available scrapers with enhanced Portuguese messaging.
// All sources are queried concurrently and results are collected until every
// scraper finishes or the hard timeout expires.
func (sm *ScraperManager) SearchAnime(query string, scraperType *ScraperType) ([]*models.Anime, error) {
	timer := util.StartTimer("SearchAnime:Total")
	defer timer.Stop()

	util.PerfCount("search_requests")

	if scraperType != nil {
		return sm.searchSpecificScraper(query, *scraperType)
	}

	return sm.searchAllScrapersConcurrent(query)
}

// searchSpecificScraper searches using a single specific scraper
func (sm *ScraperManager) searchSpecificScraper(query string, scraperType ScraperType) ([]*models.Anime, error) {
	scraper, exists := sm.scrapers[scraperType]
	if !exists {
		return nil, fmt.Errorf("scraper type %v not found", scraperType)
	}

	sourceName := sm.getScraperDisplayName(scraperType)
	if diagnostic, retryAfter, open := sm.circuitOpenDiagnostic(scraperType); open {
		util.Warn("Search source skipped", "source", sourceName, "diagnostic", diagnostic.UserMessage(), "retry_after", retryAfter.Round(time.Second))
		return nil, fmt.Errorf("busca pulada em %s: %w", sourceName, diagnostic)
	}

	util.Debug("Searching specific scraper", "scraper", sourceName)

	results, err := scraper.SearchAnime(query)
	if err != nil {
		diagnostic := DiagnoseError(sourceName, "search", err)
		if sm.recordSourceFailure(scraperType, diagnostic) {
			util.Warn("Source circuit breaker opened", "source", sourceName, "diagnostic", diagnostic.UserMessage())
		}
		return nil, fmt.Errorf("busca falhou em %s: %w", sourceName, diagnostic)
	}
	sm.recordSourceSuccess(scraperType)

	// Add language tags
	sm.tagResults(results, scraperType)

	if len(results) > 0 {
		util.Debug("Search completed", "scraper", sourceName, "results", len(results))
	}

	return results, nil
}

// searchResult holds the result from a single scraper goroutine
type searchResult struct {
	scraperType ScraperType
	results     []*models.Anime
	err         error
}

// searchAllScrapersConcurrent searches all scrapers in parallel and waits for
// every scraper to finish (or the hard searchTimeout to expire).  This ensures
// slower sources like SuperFlix are never silently dropped.
func (sm *ScraperManager) searchAllScrapersConcurrent(query string) ([]*models.Anime, error) {
	util.Debug("Starting concurrent search across all sources", "query", query)

	ctx, cancel := context.WithTimeout(context.Background(), searchTimeout)
	defer cancel()

	// Thread-safe result collection
	var (
		allResults         []*models.Anime
		resultsMutex       sync.Mutex
		searchErrors       []string
		searchSourceErrors []error
		errorsMutex        sync.Mutex
	)

	resultChan := make(chan searchResult, len(sm.scrapers))
	var wg sync.WaitGroup

	// Launch all scrapers concurrently
	for sType, scraper := range sm.scrapers {
		sourceName := sm.getScraperDisplayName(sType)
		if diagnostic, retryAfter, open := sm.circuitOpenDiagnostic(sType); open {
			util.Warn("Search source skipped", "source", sourceName, "diagnostic", diagnostic.UserMessage(), "retry_after", retryAfter.Round(time.Second))
			searchErrors = append(searchErrors, diagnostic.UserMessage())
			searchSourceErrors = append(searchSourceErrors, fmt.Errorf("%s: %w", sourceName, diagnostic))
			continue
		}

		wg.Add(1)
		go func(st ScraperType, s UnifiedScraper) {
			defer wg.Done()
			result := sm.searchWithTimeout(ctx, st, s, query)
			resultChan <- result
		}(sType, scraper)
	}

	// Close channel when all goroutines complete
	go func() {
		wg.Wait()
		close(resultChan)
	}()

	// Collect results – wait for ALL scrapers to finish or the context to expire.
	for {
		select {
		case res, ok := <-resultChan:
			if !ok {
				// Channel closed – all scrapers finished.
				goto done
			}

			if res.err != nil {
				errorsMutex.Lock()
				sourceName := sm.getScraperDisplayName(res.scraperType)
				diagnostic := DiagnoseError(sourceName, "search", res.err)
				util.Debug("Search error", "source", sourceName, "kind", diagnostic.Kind, "error", diagnostic)
				if sm.recordSourceFailure(res.scraperType, diagnostic) {
					util.Warn("Source circuit breaker opened", "source", sourceName, "diagnostic", diagnostic.UserMessage())
				}
				searchErrors = append(searchErrors, diagnostic.UserMessage())
				searchSourceErrors = append(searchSourceErrors, fmt.Errorf("%s: %w", sourceName, diagnostic))
				errorsMutex.Unlock()
				continue
			}

			sm.recordSourceSuccess(res.scraperType)
			if len(res.results) > 0 {
				sm.tagResults(res.results, res.scraperType)
				resultsMutex.Lock()
				allResults = append(allResults, res.results...)
				resultsMutex.Unlock()

				util.Debug("Search results received",
					"source", sm.getScraperDisplayName(res.scraperType),
					"count", len(res.results))
			}

		case <-ctx.Done():
			util.Debug("Search timeout reached, returning collected results")
			goto done
		}
	}

done:
	// Log warnings for failed sources
	errorsMutex.Lock()
	if len(searchErrors) > 0 {
		for _, errMsg := range searchErrors {
			util.Warn("Search source diagnostic", "details", errMsg)
		}
	}
	errorsMutex.Unlock()

	resultsMutex.Lock()
	finalResults := allResults
	resultsMutex.Unlock()

	if len(finalResults) == 0 {
		util.Debug("No anime found", "query", query)
		errorsMutex.Lock()
		defer errorsMutex.Unlock()
		if len(searchErrors) > 0 {
			return nil, fmt.Errorf("no anime found with name: %s (some sources failed: %s): %w", query, strings.Join(searchErrors, "; "), errors.Join(searchSourceErrors...))
		}
		return nil, fmt.Errorf("no anime found with name: %s", query)
	}

	// Sort results: PT-BR first, then everything else
	sortPTBRFirst(finalResults)

	sm.logSearchSummary(finalResults)
	return finalResults, nil
}

// searchWithTimeout executes a single scraper search with timeout
func (sm *ScraperManager) searchWithTimeout(ctx context.Context, st ScraperType, s UnifiedScraper, query string) searchResult {
	sourceName := sm.getScraperDisplayName(st)
	timer := util.StartTimer("Search:" + sourceName)
	util.Debug("Searching in source", "source", sourceName)

	// Create individual timeout context
	scraperCtx, scraperCancel := context.WithTimeout(ctx, perScraperTimeout)
	defer scraperCancel()

	// Channel for search result
	done := make(chan searchResult, 1)

	go func() {
		results, err := s.SearchAnime(query)
		done <- searchResult{
			scraperType: st,
			results:     results,
			err:         err,
		}
	}()

	// Wait for result or timeout
	select {
	case result := <-done:
		timer.Stop()
		if result.err == nil {
			util.PerfCount("search_success:" + sourceName)
		} else {
			util.PerfCount("search_error:" + sourceName)
		}
		return result
	case <-scraperCtx.Done():
		timer.Stop()
		util.PerfCount("search_timeout:" + sourceName)
		util.Debug("Search timeout", "source", sourceName)
		return searchResult{
			scraperType: st,
			results:     nil,
			err:         fmt.Errorf("search timed out after %v", perScraperTimeout),
		}
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

// ptbrTitleCleanRe are compiled regexes for cleaning PT-BR anime titles
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
	// Strip dublado/legendado labels — they will be re-added by tagResults
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
// titles that appear with more than one MediaType in the batch.  Only those
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

// tagResults adds language tags and source metadata to results
func (sm *ScraperManager) tagResults(results []*models.Anime, scraperType ScraperType) {
	sourceName := sm.getScraperDisplayName(scraperType)
	isPTBR := scraperType == AnimefireType || scraperType == AnimeDriveType || scraperType == GoyabuType

	// For FlixHQ/SFlix, pre-scan to find titles that need a [Movie]/[TV]
	// disambiguation tag (same title appears as both movie and TV show).
	// SuperFlix always shows media type since it mixes movies/series/animes/doramas.
	var disambig map[string]bool
	if scraperType == FlixHQType || scraperType == SFlixType {
		disambig = needsMediaTypeDisambig(results)
	}

	for _, anime := range results {
		// Clean PT-BR titles before tagging
		if isPTBR {
			anime.Name = cleanPTBRTitle(anime.Name)
		}

		// Check if the anime name already has any language tag
		hasLanguageTag := strings.Contains(anime.Name, "[English]") ||
			strings.Contains(anime.Name, "[PT-BR]") ||
			strings.Contains(anime.Name, "[Portuguese]") ||
			strings.Contains(anime.Name, "[Português]") ||
			strings.Contains(anime.Name, "[Multilanguage]") ||
			strings.Contains(anime.Name, "[Movie]") ||
			strings.Contains(anime.Name, "[TV]")

		if !hasLanguageTag {
			switch scraperType {
			case SuperFlixType:
				// SuperFlix is PT-BR. Movies get [Movie], TV series get [TV],
				// but anime/dorama only need [PT-BR].
				switch anime.MediaType {
				case models.MediaTypeMovie:
					anime.Name = fmt.Sprintf("[Movie] [PT-BR] %s", anime.Name)
				case models.MediaTypeTV:
					anime.Name = fmt.Sprintf("[TV] [PT-BR] %s", anime.Name)
				default:
					anime.Name = fmt.Sprintf("[PT-BR] %s", anime.Name)
				}
			case FlixHQType, SFlixType:
				// FlixHQ/SFlix: add [Movie]/[TV] only when disambiguation is needed.
				key := strings.ToLower(strings.TrimSpace(anime.Name))
				if disambig[key] {
					switch anime.MediaType {
					case models.MediaTypeMovie:
						anime.Name = fmt.Sprintf("[Movie] %s", anime.Name)
					case models.MediaTypeTV:
						anime.Name = fmt.Sprintf("[TV] %s", anime.Name)
					default:
						anime.Name = fmt.Sprintf("[English] %s", anime.Name)
					}
				} else {
					anime.Name = fmt.Sprintf("[English] %s", anime.Name)
				}
			default:
				languageTag := sm.getLanguageTag(scraperType)
				anime.Name = fmt.Sprintf("%s %s", languageTag, anime.Name)
			}
		}

		// Add audio type for PT-BR sources only when detectable
		if isPTBR {
			lowerURL := strings.ToLower(anime.URL)
			lowerName := strings.ToLower(anime.Name)
			if strings.Contains(lowerName, "dublado") || strings.Contains(lowerURL, "dublado") {
				if !strings.Contains(anime.Name, "(Dublado)") {
					anime.Name = anime.Name + " (Dublado)"
				}
			} else if strings.Contains(lowerName, "legendado") || strings.Contains(lowerURL, "legendado") {
				if !strings.Contains(anime.Name, "(Legendado)") {
					anime.Name = anime.Name + " (Legendado)"
				}
			}
		}

		anime.Source = sourceName
	}
}

// logSearchSummary logs a summary of search results by source
func (sm *ScraperManager) logSearchSummary(results []*models.Anime) {
	if !util.IsDebug {
		return
	}

	counts := make(map[string]int)
	for _, anime := range results {
		counts[anime.Source]++
	}

	util.Debug("Search summary",
		"animeFire", counts["Animefire.io"],
		"allAnime", counts["AllAnime"],
		"animeDrive", counts["AnimeDrive"],
		"flixHQ", counts["FlixHQ"],
		"9anime", counts["9Anime"],
		"goyabu", counts["Goyabu"],
		"superflix", counts["SuperFlix"],
		"total", len(results))
}

// SearchAnimePTBR searches PT-BR sources (AnimeFire and Goyabu) concurrently
func (sm *ScraperManager) SearchAnimePTBR(query string) ([]*models.Anime, error) {
	ptbrTypes := []ScraperType{AnimefireType, GoyabuType, SuperFlixType}

	var (
		allResults []*models.Anime
		mu         sync.Mutex
		wg         sync.WaitGroup
	)

	for _, st := range ptbrTypes {
		wg.Add(1)
		go func(scraperType ScraperType) {
			defer wg.Done()
			results, err := sm.searchSpecificScraper(query, scraperType)
			if err != nil {
				util.Debug("PT-BR search error", "source", sm.getScraperDisplayName(scraperType), "error", err)
				return
			}
			mu.Lock()
			allResults = append(allResults, results...)
			mu.Unlock()
		}(st)
	}

	wg.Wait()

	if len(allResults) == 0 {
		return nil, fmt.Errorf("no PT-BR results found for: %s", query)
	}
	return allResults, nil
}

// GetScraper returns a specific scraper by type
func (sm *ScraperManager) GetScraper(scraperType ScraperType) (UnifiedScraper, error) {
	if scraper, exists := sm.scrapers[scraperType]; exists {
		return scraper, nil
	}
	return nil, fmt.Errorf("scraper type %v not found", scraperType)
}

// getScraperDisplayName returns a Portuguese display name for the scraper type
func (sm *ScraperManager) getScraperDisplayName(scraperType ScraperType) string {
	switch scraperType {
	case AllAnimeType:
		return "AllAnime"
	case AnimefireType:
		return "Animefire.io"
	case AnimeDriveType:
		return "AnimeDrive"
	case FlixHQType:
		return "FlixHQ"
	case SFlixType:
		return "SFlix"
	case NineAnimeType:
		return "9Anime"
	case GoyabuType:
		return "Goyabu"
	case SuperFlixType:
		return "SuperFlix"
	default:
		return "Desconhecido"
	}
}

// getLanguageTag returns a language tag for the source
func (sm *ScraperManager) getLanguageTag(scraperType ScraperType) string {
	switch scraperType {
	case AllAnimeType:
		return "[English]"
	case AnimefireType:
		return "[PT-BR]"
	case AnimeDriveType:
		return "[PT-BR]"
	case FlixHQType:
		return "[English]"
	case SFlixType:
		return "[English]"
	case NineAnimeType:
		return "[Multilanguage]"
	case GoyabuType:
		return "[PT-BR]"
	case SuperFlixType:
		return "[PT-BR]"
	default:
		return "[Unknown]"
	}
}

// AllAnimeAdapter adapts AllAnimeClient to UnifiedScraper interface
type AllAnimeAdapter struct {
	client *AllAnimeClient
}

// Client returns the underlying AllAnimeClient for direct access to enhanced features.
func (a *AllAnimeAdapter) Client() *AllAnimeClient {
	return a.client
}

func (a *AllAnimeAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	// mode is now hardcoded in the new implementation
	return a.client.SearchAnime(query)
}

func (a *AllAnimeAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	// For AllAnime, animeURL is actually the anime ID
	episodes, err := a.client.GetEpisodesList(animeURL, "sub")
	if err != nil {
		return nil, err
	}

	var episodeModels []models.Episode
	for i, ep := range episodes {
		episodeModels = append(episodeModels, models.Episode{
			Number: ep,
			Num:    i + 1,
			URL:    animeURL, // Store the anime ID in URL field
			Title: models.TitleDetails{
				Romaji: fmt.Sprintf("Episode %s", ep),
			},
		})
	}

	return episodeModels, nil
}

func (a *AllAnimeAdapter) GetStreamURL(episodeURL string, options ...any) (string, map[string]string, error) {
	// For AllAnime, episodeURL contains the anime ID
	animeID := episodeURL

	// Parse options to get episode number
	episodeNo := "1"
	if len(options) > 0 {
		if ep, ok := options[0].(string); ok {
			episodeNo = ep
		}
	}

	quality := "best"
	if len(options) > 1 {
		if q, ok := options[1].(string); ok {
			quality = q
		}
	}

	mode := "sub"
	if len(options) > 2 {
		if m, ok := options[2].(string); ok {
			mode = m
		}
	}

	return a.client.GetEpisodeURL(animeID, episodeNo, mode, quality)
}

func (a *AllAnimeAdapter) GetType() ScraperType {
	return AllAnimeType
}

// AnimefireAdapter adapts AnimefireClient to UnifiedScraper interface
type AnimefireAdapter struct {
	client *AnimefireClient
}

func (a *AnimefireAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.client.SearchAnime(query)
}

func (a *AnimefireAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodes(animeURL)
}

func (a *AnimefireAdapter) GetStreamURL(episodeURL string, options ...any) (string, map[string]string, error) {
	url, err := a.client.GetEpisodeStreamURL(episodeURL)
	metadata := make(map[string]string)
	metadata["source"] = "animefire"
	return url, metadata, err
}

func (a *AnimefireAdapter) GetType() ScraperType {
	return AnimefireType
}

// AnimeDriveAdapter adapts AnimeDriveClient to UnifiedScraper interface
type AnimeDriveAdapter struct {
	client *AnimeDriveClient
}

func (a *AnimeDriveAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.client.SearchAnime(query)
}

func (a *AnimeDriveAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodes(animeURL)
}

func (a *AnimeDriveAdapter) GetStreamURL(episodeURL string, options ...any) (string, map[string]string, error) {
	// Check if server selection is requested via options
	selectServer := true // Default to showing server selection
	for _, opt := range options {
		if s, ok := opt.(string); ok && s == "auto" {
			selectServer = false
			break
		}
		if b, ok := opt.(bool); ok {
			selectServer = b
		}
	}

	if selectServer {
		return a.client.GetStreamURLWithSelection(episodeURL)
	}
	return a.client.GetStreamURL(episodeURL)
}

func (a *AnimeDriveAdapter) GetType() ScraperType {
	return AnimeDriveType
}

// FlixHQAdapter adapts FlixHQClient to UnifiedScraper interface for movies and TV shows
type FlixHQAdapter struct {
	client *FlixHQClient
}

func (a *FlixHQAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	media, err := a.client.SearchMedia(query)
	if err != nil {
		return nil, err
	}

	var animes []*models.Anime
	for _, m := range media {
		anime := m.ToAnimeModel()
		// Set the media type
		if m.Type == MediaTypeMovie {
			anime.MediaType = models.MediaTypeMovie
		} else {
			anime.MediaType = models.MediaTypeTV
		}
		anime.Year = m.Year
		animes = append(animes, anime)
	}

	return animes, nil
}

func (a *FlixHQAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	// For FlixHQ, animeURL contains the media ID
	// This needs to be called differently for movies vs TV shows
	// For movies, return a single "episode"
	// For TV shows, we need to get seasons first

	// This is a simplified implementation - in practice, you'd need to know if it's a movie or TV show
	return nil, fmt.Errorf("for FlixHQ, use GetSeasons and GetEpisodes directly on the client")
}

func (a *FlixHQAdapter) GetStreamURL(episodeURL string, options ...any) (string, map[string]string, error) {
	// Parse options
	provider := "Vidcloud"
	quality := "1080"
	subsLanguage := "english"

	for i, opt := range options {
		if s, ok := opt.(string); ok {
			switch i {
			case 0:
				provider = s
			case 1:
				quality = s
			case 2:
				subsLanguage = s
			}
		}
	}

	// Get embed link directly from episode ID
	embedLink, err := a.client.GetEmbedLink(episodeURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get embed link: %w", err)
	}

	streamInfo, err := a.client.ExtractStreamInfo(embedLink, quality, subsLanguage)
	if err != nil {
		return "", nil, fmt.Errorf("failed to extract stream info: %w", err)
	}

	metadata := make(map[string]string)
	metadata["source"] = "flixhq"
	metadata["provider"] = provider
	metadata["quality"] = quality

	// Include subtitle URLs in metadata
	if len(streamInfo.Subtitles) > 0 {
		var subURLs []string
		for _, sub := range streamInfo.Subtitles {
			subURLs = append(subURLs, sub.URL)
		}
		metadata["subtitles"] = strings.Join(subURLs, ",")
	}

	return streamInfo.VideoURL, metadata, nil
}

func (a *FlixHQAdapter) GetType() ScraperType {
	return FlixHQType
}

// GetClient returns the underlying FlixHQ client for direct access
func (a *FlixHQAdapter) GetClient() *FlixHQClient {
	return a.client
}

// SFlixAdapter adapts SFlixClient to UnifiedScraper interface for movies and TV shows
type SFlixAdapter struct {
	client *SFlixClient
}

func (a *SFlixAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	media, err := a.client.SearchMedia(query)
	if err != nil {
		return nil, err
	}

	var animes []*models.Anime
	for _, m := range media {
		anime := m.ToAnimeModel()
		// Set the media type
		if m.Type == MediaTypeMovie {
			anime.MediaType = models.MediaTypeMovie
		} else {
			anime.MediaType = models.MediaTypeTV
		}
		anime.Year = m.Year
		animes = append(animes, anime)
	}

	return animes, nil
}

func (a *SFlixAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	// For SFlix, animeURL contains the media ID
	// This needs to be called differently for movies vs TV shows
	return nil, fmt.Errorf("for SFlix, use GetSeasons and GetEpisodes directly on the client")
}

func (a *SFlixAdapter) GetStreamURL(episodeURL string, options ...any) (string, map[string]string, error) {
	// Parse options
	provider := "Vidcloud"
	quality := "1080"
	subsLanguage := "english"

	for i, opt := range options {
		if s, ok := opt.(string); ok {
			switch i {
			case 0:
				provider = s
			case 1:
				quality = s
			case 2:
				subsLanguage = s
			}
		}
	}

	// Get embed link directly from episode ID
	embedLink, err := a.client.GetEmbedLink(episodeURL)
	if err != nil {
		return "", nil, fmt.Errorf("failed to get embed link: %w", err)
	}

	streamInfo, err := a.client.ExtractStreamInfo(embedLink, quality, subsLanguage)
	if err != nil {
		return "", nil, fmt.Errorf("failed to extract stream info: %w", err)
	}

	metadata := make(map[string]string)
	metadata["source"] = "sflix"
	metadata["provider"] = provider
	metadata["quality"] = quality

	// Include subtitle URLs in metadata
	if len(streamInfo.Subtitles) > 0 {
		var subURLs []string
		for _, sub := range streamInfo.Subtitles {
			subURLs = append(subURLs, sub.URL)
		}
		metadata["subtitles"] = strings.Join(subURLs, ",")
	}

	return streamInfo.VideoURL, metadata, nil
}

func (a *SFlixAdapter) GetType() ScraperType {
	return SFlixType
}

// GetClient returns the underlying SFlix client for direct access
func (a *SFlixAdapter) GetClient() *SFlixClient {
	return a.client
}

// NineAnimeAdapter adapts NineAnimeClient to UnifiedScraper interface
type NineAnimeAdapter struct {
	client *NineAnimeClient
}

func (a *NineAnimeAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.client.SearchAnime(query)
}

func (a *NineAnimeAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodes(animeURL)
}

func (a *NineAnimeAdapter) GetStreamURL(episodeURL string, options ...any) (string, map[string]string, error) {
	// episodeURL contains the episode data-id for 9anime
	// options[0] = audio preference ("sub" or "dub")
	return a.client.GetStreamURL(episodeURL, options...)
}

func (a *NineAnimeAdapter) GetType() ScraperType {
	return NineAnimeType
}

// GetClient returns the underlying NineAnime client for direct access
func (a *NineAnimeAdapter) GetClient() *NineAnimeClient {
	return a.client
}

// GoyabuAdapter adapts GoyabuClient to UnifiedScraper interface
type GoyabuAdapter struct {
	client *GoyabuClient
}

func (a *GoyabuAdapter) SearchAnime(query string, options ...any) ([]*models.Anime, error) {
	return a.client.SearchAnime(query)
}

func (a *GoyabuAdapter) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	return a.client.GetAnimeEpisodes(animeURL)
}

func (a *GoyabuAdapter) GetStreamURL(episodeURL string, options ...any) (string, map[string]string, error) {
	url, err := a.client.GetEpisodeStreamURL(episodeURL)
	metadata := make(map[string]string)
	metadata["source"] = "goyabu"
	return url, metadata, err
}

func (a *GoyabuAdapter) GetType() ScraperType {
	return GoyabuType
}

// SuperFlixAdapter adapts SuperFlixClient to UnifiedScraper interface
type SuperFlixAdapter struct {
	client *SuperFlixClient
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

func (a *SuperFlixAdapter) GetStreamURL(episodeURL string, options ...any) (string, map[string]string, error) {
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

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	result, err := a.client.GetStreamURL(ctx, mediaType, episodeURL, season, episode)
	if err != nil {
		return "", nil, err
	}

	metadata := make(map[string]string)
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
func (a *SuperFlixAdapter) GetClient() *SuperFlixClient {
	return a.client
}

// NewSuperFlixAdapterWithClient creates a SuperFlixAdapter with a pre-configured client.
// Useful for testing with mock servers.
func NewSuperFlixAdapterWithClient(client *SuperFlixClient) *SuperFlixAdapter {
	return &SuperFlixAdapter{client: client}
}
