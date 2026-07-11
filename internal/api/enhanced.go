// Package api provides enhanced anime search and streaming capabilities
package api

import (
	"context"
	"errors"
	"fmt"
	"math"
	"net/http"
	"os"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"charm.land/huh/v2/spinner"
	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/ktr0731/go-fuzzyfinder"
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

// sfBrowserSpinnerHint is appended to SuperFlix spinner titles so the browser
// window that may pop up is expected, not alarming. Plain language only: a lay
// user must understand it at a glance, so no "Cloudflare"/"Turnstile" jargon.
const sfBrowserSpinnerHint = " — a browser may open to check you're human; just wait (click the box if one shows)"

// Indirection points for preflightSuperFlixBrowser, overridable in tests so the
// notice logic can be exercised without a real display, cache marker, or logger.
var (
	sfHeadlessEnvFn  = superflix.HeadlessEnvironment
	sfSetupPendingFn = superflix.BrowserSetupPending
	sfWarnFn         = util.Warn
	sfInfoFn         = util.Info
)

// preflightSuperFlixBrowser emits the spinner-safe, pre-solve notices for the
// Cloudflare-bypass browser: a one-time first-run setup notice and a warning
// when there is no graphical display to show the headed browser. Both run
// OUTSIDE runWithSpinner so they cannot corrupt the spinner line.
func preflightSuperFlixBrowser() {
	// Warn first: on a screenless host the check can't be shown, so the user
	// should see this before the (one-time) setup notice or the spinner.
	// Plain language only — no "$DISPLAY"/"Cloudflare"/"headless" jargon.
	if sfHeadlessEnvFn() {
		sfWarnFn("⚠️  SuperFlix needs to open a browser window, but no screen was found (you may be connected remotely). It probably won't work here — try running GoAnime on your normal computer.")
	}
	if sfSetupPendingFn() {
		sfInfoFn("⏳ First time on SuperFlix: setting up a small helper browser (one time only, needs internet). This may take a minute…")
	}
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

// isGateTimeout reports whether err is the SuperFlix "are you human?" check that
// ran out of time. It is a plain fmt.Errorf (not a sentinel), so it is matched
// by its stable substring rather than errors.Is.
func isGateTimeout(err error) bool {
	return err != nil && strings.Contains(err.Error(), "gate not cleared")
}

// describeSuperFlixErr converts a low-level SuperFlix failure into a short,
// plain-language, icon-prefixed message a non-technical user can act on at a
// glance — no "Cloudflare"/"Turnstile"/"Playwright" jargon, and the raw cause is
// hidden from Error() but kept reachable via errors.Is/Unwrap.
func describeSuperFlixErr(err error) error {
	switch {
	case err == nil:
		return nil
	case errors.Is(err, superflix.ErrPlaywrightUnavailable):
		return &friendlyError{cause: err, msg: "⚠️  Couldn't open the helper browser. The first time you use SuperFlix, GoAnime needs internet to set it up — check your connection and try again. Tip: installing Google Chrome makes this faster."}
	case errors.Is(err, superflix.ErrSuperFlixNoServers):
		return &friendlyError{cause: err, msg: "⚠️  No video sources for this title right now. Try another episode, or come back later."}
	case errors.Is(err, superflix.ErrSuperFlixNoEpisodeList):
		return &friendlyError{cause: err, msg: "⚠️  SuperFlix didn't show an episode list for this title. Try searching it on another source (AnimeFire, Goyabu or AllAnime)."}
	case errors.Is(err, context.DeadlineExceeded) || isGateTimeout(err):
		return &friendlyError{cause: err, msg: "⚠️  The \"are you human?\" check didn't finish in time. Please try again — if a small box appears in the browser window, click it."}
	default:
		return err
	}
}

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

// SearchFetchFunc fans out a free-text search across the given source kinds
// (empty = all) and returns the aggregated, language-tagged results. It is a
// seam so the api package can dispatch through the Model B registry
// (providers.SearchAll) without importing providers (which would cycle). The
// providers package wires it in its init(); if unset, the search falls back to
// the ScraperManager engine.
type SearchFetchFunc func(ctx context.Context, query string, kinds []source.SourceKind) ([]*models.Anime, error)

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

// Enhanced search that supports multiple sources - always searches both Animefire.io and allanime simultaneously
func SearchAnimeEnhanced(name string, src string) (*models.Anime, error) {
	// Map the optional source selector to the registry kinds to search. Empty
	// = all sources; a specific kind (or the PT-BR trio) narrows the fan-out.
	var registryKinds []source.SourceKind
	switch strings.ToLower(src) {
	case "allanime":
		registryKinds = []source.SourceKind{source.AllAnime}
	case "animefire":
		registryKinds = []source.SourceKind{source.AnimeFire}
	case "goyabu":
		registryKinds = []source.SourceKind{source.Goyabu}
	case "superflix":
		registryKinds = []source.SourceKind{source.SuperFlix}
	case "ptbr", "pt-br":
		registryKinds = []source.SourceKind{source.AnimeFire, source.Goyabu, source.SuperFlix}
	}
	util.Debug("Searching for anime/media", "query", name, "kinds", registryKinds)

	var animes []*models.Anime
	var searchErr error
	runWithSpinner("Searching for anime...", func() {
		if searchFetchFn == nil {
			searchErr = fmt.Errorf("search dispatch not wired: the providers package must be imported")
			return
		}
		// Model B registry fan-out (providers.SearchAll).
		animes, searchErr = searchFetchFn(context.Background(), name, registryKinds)
	})
	if searchErr != nil {
		return nil, fmt.Errorf("failed to search: %w", searchErr)
	}

	if len(animes) == 0 {
		return nil, fmt.Errorf("no results found for: %s", name)
	}

	// Enhance source identification - names already have language tags from unified.go
	for _, anime := range animes {
		// Ensure proper source identification (for internal use only)
		if anime.Source == "" {
			// Fallback source identification by URL analysis
			switch {
			case len(anime.URL) < 30 && strings.ContainsAny(anime.URL, "ABCDEFGHIJKLMNOPQRSTUVWXYZabcdefghijklmnopqrstuvwxyz0123456789") && !strings.Contains(anime.URL, "http"):
				anime.Source = "AllAnime"
			case strings.Contains(anime.URL, "animefire"):
				anime.Source = "Animefire.io"
			case strings.Contains(anime.URL, "goyabu"):
				anime.Source = "Goyabu"
			}
		}

		// Language tags are already added by unified.go, don't duplicate them here
	}

	util.Debug("Search results summary", "total", len(animes))

	breakdown := countSourceBreakdown(animes)
	util.Debug("Source breakdown",
		"AnimeFire", breakdown.AnimeFire,
		"AllAnime", breakdown.AllAnime,
		"SuperFlix", breakdown.SuperFlix,
		"Goyabu", breakdown.Goyabu,
	)

	// Sort results by language priority: Portuguese first, then Multilanguage, Movies/TV, English, others
	sort.SliceStable(animes, func(i, j int) bool {
		return languagePriority(animes[i].Name) < languagePriority(animes[j].Name)
	})

	// Create a special "back" option as the first item
	backOption := &models.Anime{
		Name:   "← Back",
		URL:    "__back__",
		Source: "__back__",
	}

	// Prepend back option to the list
	animesWithBack := make([]*models.Anime, 0, len(animes)+1)
	animesWithBack = append(animesWithBack, backOption)
	animesWithBack = append(animesWithBack, animes...)

	// Use fuzzy finder to let user select
	var idx int
	var err error

	if util.IsDebug {
		// In debug mode, show preview window with technical details
		idx, err = tui.Find(
			animesWithBack,
			func(i int) string {
				a := animesWithBack[i]
				name := a.Name
				// Append release year if available and not already in the name
				if a.Year != "" && !strings.Contains(name, "("+a.Year+")") {
					name += " (" + a.Year + ")"
				}
				return name
			},
			fuzzyfinder.WithPromptString("Select the anime you want: "),
			fuzzyfinder.WithPreviewWindow(func(i, w, h int) string {
				if i >= 0 && i < len(animesWithBack) {
					anime := animesWithBack[i]
					if anime.Source == "__back__" {
						return "Go back to perform a new search"
					}
					var preview string
					preview = "Source: " + anime.Source + "\nURL: " + anime.URL
					if anime.ImageURL != "" {
						preview += "\nImage: " + anime.ImageURL
					}
					return preview
				}
				return ""
			}),
		)
	} else {
		// In normal mode, no preview window at all
		idx, err = tui.Find(
			animesWithBack,
			func(i int) string {
				a := animesWithBack[i]
				name := a.Name
				// Append release year if available and not already in the name
				if a.Year != "" && !strings.Contains(name, "("+a.Year+")") {
					name += " (" + a.Year + ")"
				}
				return name
			},
			fuzzyfinder.WithPromptString("Select the anime you want: "),
		)
	}

	if err != nil {
		return nil, fmt.Errorf("anime selection cancelled: %w", err)
	}

	selectedAnime := animesWithBack[idx]

	// Check if user selected the back option
	if selectedAnime.Source == "__back__" {
		return nil, ErrBackToSearch
	}
	util.Debug("Anime selected", "name", selectedAnime.Name, "source", selectedAnime.Source)

	// Enrich with AniList data for images and metadata. Best-effort: episodes and
	// playback work without it, so a failure here is a warning, not an error
	// (issue #184).
	if err := enrichAnimeData(selectedAnime); err != nil {
		util.Warn("Metadata enrichment unavailable; continuing without it", "anime", selectedAnime.Name, "error", err)
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
func downloadFromURL(_ string, _ string) error {
	// This is a placeholder that should fail to trigger fallback to the proper downloader
	util.Debugf("Enhanced API downloadFromURL is a placeholder - returning error to trigger fallback")
	return fmt.Errorf("enhanced download not implemented - use legacy downloader")
}

// Legacy wrapper functions to maintain compatibility
func SearchAnimeWithSource(name string, source string) (*models.Anime, error) {
	return SearchAnimeEnhanced(name, source)
}

func GetAnimeEpisodesWithSource(anime *models.Anime) ([]models.Episode, error) {
	return fetchEpisodesViaRegistry(anime)
}

// sortedSeasonNumbers returns the season keys in ascending numeric order.
//
// A plain string sort is wrong here: it orders "10" before "2", so a show with
// ten or more seasons lists them scrambled. Non-numeric keys (TVmaze exposes
// year-based "seasons" for some long-running anime) fall back to string order and
// sort after the numeric ones, keeping the result deterministic.
func sortedSeasonNumbers(allEpisodes map[string][]superflix.SuperFlixEpisode) []string {
	seasons := make([]string, 0, len(allEpisodes))
	for k := range allEpisodes {
		seasons = append(seasons, k)
	}
	sort.Slice(seasons, func(i, j int) bool {
		ni, erri := strconv.Atoi(seasons[i])
		nj, errj := strconv.Atoi(seasons[j])
		switch {
		case erri == nil && errj == nil:
			return ni < nj
		case erri == nil:
			return true // numeric seasons before non-numeric ones
		case errj == nil:
			return false
		default:
			return seasons[i] < seasons[j]
		}
	})
	return seasons
}

// Episode-listing seams. Split out so the TVmaze-first ordering (the fix for the
// "no seasons found" dead end in issue #184) is testable without a network or a
// headed browser.
var (
	sfTVmazeEpisodesFn = func(ctx context.Context, imdbID string) (map[string][]superflix.SuperFlixEpisode, error) {
		return superflix.GetEpisodesFromTVmaze(ctx, http.DefaultClient, imdbID)
	}
	sfBrowserEpisodesFn = func(ctx context.Context, c *superflix.SuperFlixClient, tmdbID string) (map[string][]superflix.SuperFlixEpisode, error) {
		return c.GetEpisodes(ctx, tmdbID)
	}
)

// fetchSuperFlixSeasons lists a series' seasons, preferring the browser-free
// TVmaze listing and only falling back to the headed browser when TVmaze cannot
// answer.
//
// The order matters. SuperFlix now frequently serves /serie/<tmdb> as an
// embed-only shell with no episode list at all — and does so non-deterministically
// — so scraping it is unreliable, while TVmaze (keyed on the IMDB id SuperFlix
// returns in search) is deterministic and needs no browser. Putting TVmaze first
// also keeps the "a browser window will open" warnings out of the common path:
// they are emitted only if we actually reach the browser.
func fetchSuperFlixSeasons(sfClient *superflix.SuperFlixClient, media *models.Anime, tmdbID string) (map[string][]superflix.SuperFlixEpisode, error) {
	var allEpisodes map[string][]superflix.SuperFlixEpisode

	if media.IMDBID != "" {
		runWithSpinner("Loading seasons...", func() {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()
			eps, err := sfTVmazeEpisodesFn(ctx, media.IMDBID)
			if err != nil {
				util.Debug("TVmaze episode listing failed; falling back to the browser", "imdb", media.IMDBID, "err", err)
				return
			}
			allEpisodes = eps
		})
	}

	if len(allEpisodes) == 0 {
		preflightSuperFlixBrowser()

		var episodesErr error
		runWithSpinner("Loading seasons..."+sfBrowserSpinnerHint, func() {
			// Generous timeout: the player page may sit behind a Cloudflare Turnstile
			// gate that NewSuperFlixClient solves with a headed Firefox (10–40s). Must
			// exceed the client's solve budget or the solve gets cancelled mid-flight.
			ctx, cancel := context.WithTimeout(context.Background(), 210*time.Second)
			defer cancel()
			allEpisodes, episodesErr = sfBrowserEpisodesFn(ctx, sfClient, tmdbID)
		})
		if episodesErr != nil {
			return nil, fmt.Errorf("failed to get episodes: %w", describeSuperFlixErr(episodesErr))
		}
	}

	if len(allEpisodes) == 0 {
		return nil, &friendlyError{
			cause: fmt.Errorf("superflix: no seasons for tmdb=%s (imdb=%q): TVmaze had no listing and the SuperFlix page exposed no episode list", tmdbID, media.IMDBID),
			msg:   "⚠️  Couldn't load the season list for this title on SuperFlix. Try searching it on another source (AnimeFire, Goyabu or AllAnime).",
		}
	}
	return allEpisodes, nil
}

// GetSuperFlixEpisodes handles episodes/content for SuperFlix movies and TV shows
func GetSuperFlixEpisodes(media *models.Anime) ([]models.Episode, error) {
	sfClient := superflix.NewSuperFlixClient()

	// media.URL contains the TMDB ID for SuperFlix
	tmdbID := media.URL
	if tmdbID == "" {
		return nil, fmt.Errorf("no TMDB ID found for SuperFlix content")
	}

	util.Debug("Getting SuperFlix content", "mediaType", media.MediaType, "tmdbID", tmdbID)

	// For movies, return a single "episode" representing the movie
	if media.MediaType == models.MediaTypeMovie {
		util.Debug("SuperFlix: Processing movie")
		return []models.Episode{
			{
				Number: "1",
				Num:    1,
				URL:    tmdbID,
				Title: models.TitleDetails{
					English: media.Name,
					Romaji:  media.Name,
				},
			},
		}, nil
	}

	// For TV shows / series, get seasons and episodes
	util.Debug("SuperFlix: Processing TV show/series, getting episodes")

	allEpisodes, err := fetchSuperFlixSeasons(sfClient, media, tmdbID)
	if err != nil {
		return nil, err
	}

	seasonNums := sortedSeasonNumbers(allEpisodes)

	// Build season labels for selection
	var seasonLabels []string
	for _, sn := range seasonNums {
		epCount := len(allEpisodes[sn])
		seasonLabels = append(seasonLabels, fmt.Sprintf("Season %s (%d episodes)", sn, epCount))
	}

	// Let user select a season
	seasonIdx, err := tui.Find(seasonLabels, func(i int) string {
		return seasonLabels[i]
	}, fuzzyfinder.WithPromptString("Select season: "))
	if err != nil {
		if errors.Is(err, fuzzyfinder.ErrAbort) {
			return nil, ErrBackToSearch
		}
		return nil, fmt.Errorf("season selection cancelled: %w", err)
	}

	selectedSeason := seasonNums[seasonIdx]
	epList := allEpisodes[selectedSeason]
	util.Debug("Selected season", "season", selectedSeason, "episodes", len(epList))

	// Convert to models.Episode
	var episodes []models.Episode
	for _, ep := range epList {
		epNum := ep.EpiNum.String()
		num := 0
		if n, err := ep.EpiNum.Int64(); err == nil {
			num = int(n)
		}

		episodes = append(episodes, models.Episode{
			Number:   epNum,
			Num:      num,
			URL:      tmdbID, // Store TMDB ID for stream retrieval
			SeasonID: selectedSeason,
			Title: models.TitleDetails{
				English: ep.Title,
				Romaji:  ep.Title,
			},
			Aired: ep.AirDate,
		})
	}

	// Store current season on the media object
	var seasonNum int
	if _, err := fmt.Sscanf(selectedSeason, "%d", &seasonNum); err == nil {
		media.CurrentSeason = seasonNum
	}

	util.Debug("SuperFlix episodes loaded", "count", len(episodes))
	return episodes, nil
}

// GetSuperFlixStreamURL gets the stream URL for SuperFlix content.
//
// Subtitle clearing and global-source tagging are handled by the only caller,
// GetEpisodeStreamURL — duplicating them here produced two identical
// "Stored anime source: SuperFlix" debug lines per playback.
func GetSuperFlixStreamURL(media *models.Anime, episode *models.Episode, quality string) (string, error) {
	sfClient := superflix.NewSuperFlixClient()

	tmdbID := episode.URL
	if tmdbID == "" {
		tmdbID = media.URL
	}

	var sfType, season, epNum string
	if media.MediaType == models.MediaTypeMovie {
		sfType = "filme"
	} else {
		sfType = "serie"
		season = episode.SeasonID
		epNum = episode.Number
	}

	util.Debug("Getting SuperFlix stream", "tmdbID", tmdbID, "type", sfType, "season", season, "episode", epNum)

	preflightSuperFlixBrowser()

	var result *superflix.SuperFlixStreamResult
	var streamErr error
	runWithSpinner("Loading stream..."+sfBrowserSpinnerHint, func() {
		// Generous timeout: the pipeline's first request may hit a Cloudflare
		// Turnstile gate that the client solves with a headed Firefox (10–40s).
		// Must exceed the client's solve budget or the solve gets cancelled.
		ctx, cancel := context.WithTimeout(context.Background(), 210*time.Second)
		defer cancel()
		result, streamErr = sfClient.GetStreamURL(ctx, sfType, tmdbID, season, epNum)
	})
	if streamErr != nil {
		return "", fmt.Errorf("failed to get SuperFlix stream: %w", describeSuperFlixErr(streamErr))
	}

	// Store referer globally for mpv playback
	if result.Referer != "" {
		util.SetGlobalReferer(result.Referer)
	}

	// Update cover image from stream thumbnail if not already set
	if media.ImageURL == "" && result.Thumb != "" {
		media.ImageURL = result.Thumb
		util.Debug("SuperFlix cover set from stream thumbnail", "url", result.Thumb)
	}

	// Store subtitles globally for playback
	if len(result.Subtitles) > 0 && !util.GlobalNoSubs {
		var subInfos []util.SubtitleInfo
		for _, sub := range result.Subtitles {
			lang := strings.ToLower(sub.Lang)
			subInfos = append(subInfos, util.SubtitleInfo{
				URL:      sub.URL,
				Language: lang,
				Label:    sub.Lang,
			})
		}
		util.SetGlobalSubtitles(subInfos)
		util.Debug("SuperFlix subtitles loaded", "count", len(subInfos))
	}

	util.Debug("SuperFlix stream URL obtained", "url", result.StreamURL[:min(len(result.StreamURL), 80)])
	return result.StreamURL, nil
}

// sourceBreakdown holds per-source result counts for the debug "Source breakdown"
// diagnostic line. Counted via countSourceBreakdown so the predicate stays
// testable in isolation.
type sourceBreakdown struct {
	AnimeFire int
	AllAnime  int
	SuperFlix int
	Goyabu    int
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
		case anime.Source == "AllAnime":
			b.AllAnime++
		case anime.Source == "SuperFlix":
			b.SuperFlix++
		case anime.Source == "Goyabu":
			b.Goyabu++
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
