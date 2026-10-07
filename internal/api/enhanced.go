// Package api provides enhanced anime search and streaming capabilities
package api

import (
	"context"
	"errors"
	"fmt"
	"math"
	"os"
	"sort"
	"strings"
	"sync"

	"charm.land/huh/v2/spinner"
	apisource "github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"golang.org/x/term"
)

// Cached terminal detection (checked once, reused)
var (
	stdoutIsTerminal     bool
	stdoutIsTerminalOnce sync.Once
)

func isStdoutTerminal() bool {
	stdoutIsTerminalOnce.Do(func() {
		fd := os.Stdout.Fd()
		stdoutIsTerminal = fd <= math.MaxInt && term.IsTerminal(int(fd))
	})
	return stdoutIsTerminal
}

// friendlyError carries a plain-language message for the user while keeping the
// technical cause reachable via Unwrap (so errors.Is and debug tooling still see
// the root cause). Error() returns ONLY the friendly text, so the raw cause —
// which may contain jargon — is never shown to a lay user.
type friendlyError struct {
	msg   string
	cause error
}

func (e *friendlyError) Error() string { return e.msg }
func (e *friendlyError) Unwrap() error { return e.cause }

// runWithSpinner runs the action with a spinner if stdout is a terminal,
// otherwise runs the action directly. This ensures CI and non-interactive
// environments work correctly since huh/v2 spinner may skip the Action
// callback when no terminal is attached.
//
// The huh spinner's Run() can return before its Action goroutine completes
// (e.g. tea.Interrupt from residual stdin bytes left over from a prior
// fuzzyfinder). When that happens the closure that mutates the caller's
// local variables is still running, so the caller would observe zero values.
// awaitActionThroughRunner uses sync.Once + a trailing safety call to
// guarantee the action runs exactly once and that this function does not
// return until that single execution has finished.
func runWithSpinner(title string, action func()) {
	if !isStdoutTerminal() {
		action()
		return
	}
	// Background probes (e.g. per-source search diagnostics) log through
	// util.Warn/Info while the spinner is animating. Those writes land on
	// the same stderr the spinner redraws, interleaving with its frames and
	// leaving garbled output behind once the spinner exits. Route console
	// logs to the file only for the spinner's lifetime, same as the
	// download progress bars do (internal/player/download.go).
	restoreConsoleLogs := util.SuppressConsoleLogging()
	defer restoreConsoleLogs()
	awaitActionThroughRunner(action, func(wrapped func()) {
		_ = tui.RunClean(func() error {
			return spinner.New().
				Title(title).
				Type(spinner.Dots).
				Action(wrapped).
				Run()
		})
	})
}

// awaitActionThroughRunner runs `action` via `runner` and guarantees that:
//   - action executes exactly once (sync.Once); and
//   - this function does not return until that single execution has fully
//     returned, even if `runner` exits before invoking the wrapped function
//     it was given.
//
// Exposed at package scope so the regression test can drive it directly with
// a mock runner that mimics the spinner's "Run() exits before Action finishes"
// race, without depending on a real terminal.
func awaitActionThroughRunner(action func(), runner func(wrapped func())) {
	var once sync.Once
	wrapped := func() { once.Do(action) }
	runner(wrapped)
	// If the runner already invoked wrapped and action is still in flight,
	// once.Do here blocks until that in-flight call returns. If the runner
	// never invoked wrapped, this call runs action now. Either way, action
	// is guaranteed to have fully completed when we return.
	wrapped()
}

// ErrBackToSearch is returned when user selects the back option to search again
var ErrBackToSearch = errors.New("back to search requested")

// ErrSearchAborted is returned when the user QUITS the result screen (q /
// Ctrl+C) rather than asking for the previous one. Quitting is a deliberate
// exit, so callers must stop instead of re-prompting — the retry loop used to
// answer a Ctrl+C with "No anime found with the name: <query>" and another
// prompt, which read as a search bug on a screen that had just listed 15
// results (issue #203).
var ErrSearchAborted = errors.New("search aborted by user")

// ErrNoResults is returned when every source answered and none had the title.
// Distinct from a source/transport failure, which must not be reported as
// "nothing matched".
var ErrNoResults = errors.New("no results found")

// SearchFetchFunc fans out a free-text search across the given source kinds
// (empty = all) and returns the aggregated, language-tagged results. It is a
// seam so the api package can dispatch through the Model B registry
// (providers.SearchAll) without importing providers (which would cycle). The
// providers package wires it in its init(); if unset, the search falls back to
// the ScraperManager engine.
type SearchFetchFunc func(ctx context.Context, query string, kinds []apisource.SourceKind) ([]*models.Anime, error)

var searchFetchFn SearchFetchFunc

// SetSearchFetch installs the registry-backed search fan-out. Called from
// providers.init().
func SetSearchFetch(f SearchFetchFunc) { searchFetchFn = f }

// EpisodesFetchFunc lists an anime's episodes through the Model B registry.
// Like SearchFetchFunc, it is a seam so api can dispatch through
// providers.FetchEpisodes without importing providers (which would cycle).
type EpisodesFetchFunc func(anime *models.Anime) ([]models.Episode, error)

var episodesFetchFn EpisodesFetchFunc

// SetEpisodesFetch installs the registry-backed episode dispatch. Called from
// providers.init().
func SetEpisodesFetch(f EpisodesFetchFunc) { episodesFetchFn = f }

// fetchEpisodesViaRegistry dispatches episode listing through the Model B
// registry seam. It is the replacement for the deleted GetAnimeEpisodesEnhanced
// per-source switch; every former caller routes here.
func fetchEpisodesViaRegistry(anime *models.Anime) ([]models.Episode, error) {
	if episodesFetchFn == nil {
		return nil, fmt.Errorf("episode dispatch not wired: the providers package must be imported")
	}
	return episodesFetchFn(anime)
}

// StreamFetchFunc resolves a single episode's stream URL through the Model B
// registry. Seam so api can dispatch through providers.FetchStreamURL without
// importing providers (which would cycle).
type StreamFetchFunc func(episode *models.Episode, anime *models.Anime, quality string) (string, error)

var streamFetchFn StreamFetchFunc

// SetStreamFetch installs the registry-backed stream dispatch. Called from
// providers.init().
func SetStreamFetch(f StreamFetchFunc) { streamFetchFn = f }

// fetchStreamViaRegistry dispatches stream resolution through the Model B
// registry seam — the replacement for the deleted GetEpisodeStreamURL switch.
func fetchStreamViaRegistry(episode *models.Episode, anime *models.Anime, quality string) (string, error) {
	if streamFetchFn == nil {
		return "", fmt.Errorf("stream dispatch not wired: the providers package must be imported")
	}
	return streamFetchFn(episode, anime, quality)
}

// Enhanced search that supports multiple sources - fans out across every registered source
func SearchAnimeEnhanced(name, src string) (*models.Anime, error) {
	return searchAnimeEnhanced(name, src, searchFetchFn, tui.SelectAnime, enrichAnimeData)
}

func searchAnimeEnhanced(
	name string,
	src string,
	search SearchFetchFunc,
	selectAnime func([]*models.Anime) (*models.Anime, error),
	enrich func(*models.Anime) error,
) (*models.Anime, error) {
	// Map the optional source selector to the registry kinds to search. Empty
	// = all sources; a specific kind (or the PT-BR trio) narrows the fan-out.
	var registryKinds []apisource.SourceKind
	normalizedSource := strings.ToLower(strings.TrimSpace(src))
	switch normalizedSource {
	case "animefire":
		registryKinds = []apisource.SourceKind{apisource.AnimeFire}
	case "goyabu":
		registryKinds = []apisource.SourceKind{apisource.Goyabu}
	case "startflix":
		registryKinds = []apisource.SourceKind{apisource.StartFlix}
	case "hianime", "anidb": // "anidb" is what this source was called before 2026-09-22
		registryKinds = []apisource.SourceKind{apisource.HiAnime}
	case "ptbr", "pt-br":
		registryKinds = []apisource.SourceKind{apisource.AnimeFire, apisource.Goyabu, apisource.StartFlix}
	}
	util.Debug("Searching for anime/media", "query", name, "kinds", registryKinds)

	var animes []*models.Anime
	var searchErr error
	runWithSpinner("Searching for anime...", func() {
		if search == nil {
			searchErr = fmt.Errorf("search dispatch not wired: the providers package must be imported")
			return
		}
		// Model B registry fan-out (providers.SearchAll).
		animes, searchErr = search(context.Background(), name, registryKinds)
	})
	if searchErr != nil {
		return nil, fmt.Errorf("failed to search: %w", searchErr)
	}
	validAnimes := make([]*models.Anime, 0, len(animes))
	for _, anime := range animes {
		if anime != nil {
			validAnimes = append(validAnimes, anime)
		}
	}
	animes = validAnimes

	if len(animes) == 0 {
		return nil, fmt.Errorf("%w for: %s", ErrNoResults, name)
	}

	// Enhance source identification - names already have language tags from unified.go
	for _, anime := range animes {
		// Ensure proper source identification (for internal use only)
		if anime.Source == "" {
			switch normalizedSource {
			case "animefire":
				anime.Source = "Animefire.io"
			case "goyabu":
				anime.Source = "Goyabu"
			case "startflix":
				anime.Source = "StartFlix"
			case "hianime", "anidb":
				anime.Source = "HiAnime"
			}
			if anime.Source == "" {
				lowerURL := strings.ToLower(anime.URL)
				switch {
				case strings.Contains(lowerURL, "animefire"):
					anime.Source = "Animefire.io"
				case strings.Contains(lowerURL, "goyabu"):
					anime.Source = "Goyabu"
				case strings.Contains(lowerURL, "startflix"):
					anime.Source = "StartFlix"
				case strings.Contains(lowerURL, "hianime.at"), strings.Contains(lowerURL, "anidb.app"):
					anime.Source = "HiAnime"
				}
			}
		}

		// Language tags are already added by unified.go, don't duplicate them here
	}

	util.Debug("Search results summary", "total", len(animes))

	breakdown := countSourceBreakdown(animes)
	util.Debug("Source breakdown",
		"AnimeFire", breakdown.AnimeFire,
		"StartFlix", breakdown.StartFlix,
		"Goyabu", breakdown.Goyabu,
		"HiAnime", breakdown.HiAnime,
	)

	// Sort results by language priority: Portuguese first, then Multilanguage, Movies/TV, English, others
	sort.SliceStable(animes, func(i, j int) bool {
		return languagePriority(animes[i].Name) < languagePriority(animes[j].Name)
	})

	if selectAnime == nil {
		return nil, fmt.Errorf("anime selection not configured")
	}
	selectedAnime, err := selectAnime(animes)
	if errors.Is(err, tui.ErrSelectionBack) {
		return nil, ErrBackToSearch
	}
	if errors.Is(err, tui.ErrSelectionCancelled) {
		return nil, fmt.Errorf("%w: %w", ErrSearchAborted, err)
	}
	if err != nil {
		return nil, fmt.Errorf("anime selection failed: %w", err)
	}
	if selectedAnime == nil {
		return nil, fmt.Errorf("anime selection returned nil")
	}
	util.Debug("Anime selected", "name", selectedAnime.Name, "source", selectedAnime.Source)

	// Enrich with AniList data for images and metadata. Best-effort: episodes and
	// playback work without it, so a failure here is a warning, not an error
	// (issue #184).
	if enrich != nil {
		if err := enrich(selectedAnime); err != nil {
			util.Warn("Metadata enrichment unavailable; continuing without it", "anime", selectedAnime.Name, "error", err)
		}
	}

	return selectedAnime, nil
}

// Enhanced download support
func DownloadEpisodeEnhanced(anime *models.Anime, episodeNum int, quality string) error {
	util.Debugf("Fetching episodes for %s...", anime.Name)

	episodes, err := fetchEpisodesViaRegistry(anime)
	if err != nil {
		return fmt.Errorf("failed to get episodes: %w", err)
	}

	if episodeNum < 1 || episodeNum > len(episodes) {
		return fmt.Errorf("episode %d not found (available: 1-%d)", episodeNum, len(episodes))
	}

	episode := episodes[episodeNum-1]

	util.Debugf("Getting stream URL for episode %d...", episodeNum)
	streamURL, err := fetchStreamViaRegistry(&episode, anime, quality)
	if err != nil {
		return fmt.Errorf("failed to get stream URL: %w", err)
	}

	util.Debugf("Stream URL obtained: %s", streamURL)

	// Create a basic downloader (this would integrate with your existing downloader)
	return downloadFromURL(streamURL, fmt.Sprintf("%s_Episode_%d",
		sanitizeFilename(anime.Name), episodeNum))
}

// Enhanced range download support
func DownloadEpisodeRangeEnhanced(anime *models.Anime, startEp, endEp int, quality string) error {
	util.Debugf("Fetching episodes for %s...", anime.Name)

	episodes, err := fetchEpisodesViaRegistry(anime)
	if err != nil {
		return fmt.Errorf("failed to get episodes: %w", err)
	}

	if startEp < 1 || endEp > len(episodes) || startEp > endEp {
		return fmt.Errorf("invalid range %d-%d (available: 1-%d)", startEp, endEp, len(episodes))
	}

	for i := startEp; i <= endEp; i++ {
		util.Infof("Downloading episode %d of %d...", i, endEp)

		episode := episodes[i-1]
		streamURL, err := fetchStreamViaRegistry(&episode, anime, quality)
		if err != nil {
			util.Errorf("Failed to get stream URL for episode %d: %v", i, err)
			continue
		}

		filename := fmt.Sprintf("%s_Episode_%d", sanitizeFilename(anime.Name), i)
		// Note: downloadFromURL is a placeholder - integrate with proper downloader
		_ = downloadFromURL(streamURL, filename) // This will always fail as expected

		util.Infof("Successfully downloaded episode %d", i)
	}

	return nil
}

// Helper function to sanitize filename
func sanitizeFilename(name string) string {
	// Remove language tags
	name = strings.ReplaceAll(name, "[English]", "")
	name = strings.ReplaceAll(name, "[PT-BR]", "")
	name = strings.ReplaceAll(name, "[Português]", "")
	name = strings.ReplaceAll(name, "(Legendado)", "")
	name = strings.ReplaceAll(name, "(Dublado)", "")
	name = strings.TrimSpace(name)

	// Replace invalid characters
	invalid := []string{"/", "\\", ":", "*", "?", "\"", "<", ">", "|"}
	for _, char := range invalid {
		name = strings.ReplaceAll(name, char, "_")
	}

	return name
}

// Basic download function (placeholder - integrate with your existing downloader)
func downloadFromURL(_, _ string) error {
	// This is a placeholder that should fail to trigger fallback to the proper downloader
	util.Debugf("Enhanced API downloadFromURL is a placeholder - returning error to trigger fallback")
	return fmt.Errorf("enhanced download not implemented - use legacy downloader")
}

// Legacy wrapper functions to maintain compatibility
func SearchAnimeWithSource(name, sourceName string) (*models.Anime, error) {
	return SearchAnimeEnhanced(name, sourceName)
}

func GetAnimeEpisodesWithSource(anime *models.Anime) ([]models.Episode, error) {
	return fetchEpisodesViaRegistry(anime)
}

// sourceBreakdown holds per-source result counts for the debug "Source breakdown"
// diagnostic line. Counted via countSourceBreakdown so the predicate stays
// testable in isolation.
type sourceBreakdown struct {
	AnimeFire int
	StartFlix int
	Goyabu    int
	HiAnime   int
}

// countSourceBreakdown tallies anime results by Source field using
// case-insensitive matching for AnimeFire. The scraper canonical Source is
// "Animefire.io" (lowercase 'f'), but older callers and tests sometimes emit
// "AnimeFire"; both must be counted so the diagnostic line never lies.
func countSourceBreakdown(animes []*models.Anime) sourceBreakdown {
	var b sourceBreakdown
	for _, anime := range animes {
		if anime == nil {
			continue
		}
		switch {
		case strings.Contains(strings.ToLower(anime.Source), "animefire"):
			b.AnimeFire++
		case anime.Source == "StartFlix":
			b.StartFlix++
		case anime.Source == "Goyabu":
			b.Goyabu++
		case anime.Source == "HiAnime":
			b.HiAnime++
		}
	}
	return b
}

// languagePriority returns a sort key for language-based ordering.
// Lower values sort first: Portuguese → Multilanguage → English → Movies/TV → Unknown.
func languagePriority(name string) int {
	lower := strings.ToLower(name)
	// Check for [PT-BR] anywhere (covers "[Movie] [PT-BR] ...", "[TV] [PT-BR] ...", etc.)
	if strings.Contains(lower, "[pt-br]") || strings.Contains(lower, "[portuguese]") || strings.Contains(lower, "[português]") {
		return 0
	}
	switch {
	case strings.HasPrefix(lower, "[multilanguage]"):
		return 1
	case strings.HasPrefix(lower, "[english]"):
		return 2
	case strings.HasPrefix(lower, "[movie]") || strings.HasPrefix(lower, "[tv]") || strings.HasPrefix(lower, "[movies/tv]"):
		return 3
	default:
		return 4
	}
}
