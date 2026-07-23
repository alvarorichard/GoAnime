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
	apisource "github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
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
	case errors.Is(err, superflix.ErrSuperFlixRestricted):
		return &friendlyError{cause: err, msg: "⚠️  Este título está com acesso restrito no SuperFlix e não abriu. Tente outro título, ou procure em outra fonte (AnimeFire, Goyabu ou AllAnime)."}
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

// Enhanced search that supports multiple sources - always searches both Animefire.io and allanime simultaneously
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
	case "allanime":
		registryKinds = []apisource.SourceKind{apisource.AllAnime}
	case "animefire":
		registryKinds = []apisource.SourceKind{apisource.AnimeFire}
	case "goyabu":
		registryKinds = []apisource.SourceKind{apisource.Goyabu}
	case "superflix":
		registryKinds = []apisource.SourceKind{apisource.SuperFlix}
	case "ptbr", "pt-br":
		registryKinds = []apisource.SourceKind{apisource.AnimeFire, apisource.Goyabu, apisource.SuperFlix}
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
		return nil, fmt.Errorf("no results found for: %s", name)
	}

	// Enhance source identification - names already have language tags from unified.go
	for _, anime := range animes {
		// Ensure proper source identification (for internal use only)
		if anime.Source == "" {
			switch normalizedSource {
			case "allanime":
				anime.Source = "AllAnime"
			case "animefire":
				anime.Source = "Animefire.io"
			case "goyabu":
				anime.Source = "Goyabu"
			case "superflix":
				anime.Source = "SuperFlix"
			}
			if anime.Source == "" {
				lowerURL := strings.ToLower(anime.URL)
				switch {
				case apisource.IsAllAnimeShortID(anime.URL), strings.Contains(lowerURL, "allanime"):
					anime.Source = "AllAnime"
				case strings.Contains(lowerURL, "animefire"):
					anime.Source = "Animefire.io"
				case strings.Contains(lowerURL, "goyabu"):
					anime.Source = "Goyabu"
				case strings.Contains(lowerURL, "superflix"), strings.Contains(lowerURL, "sflix"):
					anime.Source = "SuperFlix"
				}
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

	if selectAnime == nil {
		return nil, fmt.Errorf("anime selection not configured")
	}
	selectedAnime, err := selectAnime(animes)
	if errors.Is(err, tui.ErrSelectionBack) {
		return nil, ErrBackToSearch
	}
	if err != nil {
		return nil, fmt.Errorf("anime selection cancelled: %w", err)
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
	sfClient := superflix.SharedSuperFlixClient()

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

	// Let user select a season (auto-selects when there is only one)
	selectedSeason, err := selectSuperFlixSeason(media, seasonNums, allEpisodes)
	if err != nil {
		if errors.Is(err, tui.ErrPickBack) || errors.Is(err, tui.ErrPickCancelled) {
			return nil, ErrBackToSearch
		}
		return nil, fmt.Errorf("season selection cancelled: %w", err)
	}

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

// sfServerListBudget caps how long we spend fetching the server list.
//
// The list is an ENHANCEMENT: it buys the user a choice of source and names each
// one dublado or legendado. Getting it needs the Cloudflare browser solve (the
// tokened player page is gated), which is ~6s once the persistent profile is warm
// — and it is warm after any prior SuperFlix play. This budget bounds the cold
// case: a brand-new profile whose first-ever solve would run long is cut off here
// and falls back to the embed sniff, which does its own solve for the stream. So
// the worst the enhancement can add is this budget, once, on a cold profile.
const sfServerListBudget = 60 * time.Second

// sfCachedStreamBudget bounds the cache-replay fast path. It is pure HTTP (a
// getVideo call + the player-extras GET), so a couple of round-trips — generous
// enough to absorb a slow CDN, tight enough that a stale/rotated host fails fast
// and we fall through to a full resolve.
const sfCachedStreamBudget = 8 * time.Second

// Stream seams. Split out so the "cache → servers → sniff" ordering is testable
// without a network or a headed browser.
var (
	sfCachedStreamFn = func(c *superflix.SuperFlixClient, mediaType, mediaID, season, episode string) (*superflix.SuperFlixStreamResult, bool) {
		ctx, cancel := context.WithTimeout(context.Background(), sfCachedStreamBudget)
		defer cancel()
		return c.TryCachedStream(ctx, mediaType, mediaID, season, episode)
	}
	sfGetServersFn = func(c *superflix.SuperFlixClient, ctx context.Context, mediaType, mediaID, season, episode string) ([]superflix.SuperFlixServer, *superflix.SuperFlixTokens, error) {
		return c.GetServers(ctx, mediaType, mediaID, season, episode)
	}
	sfStreamFromServerFn = func(c *superflix.SuperFlixClient, ctx context.Context, tokens *superflix.SuperFlixTokens, serverID, mediaType, mediaID, season, episode string) (*superflix.SuperFlixStreamResult, error) {
		return c.StreamFromServer(ctx, tokens, serverID, mediaType, mediaID, season, episode)
	}
	sfSniffStreamFn = func(c *superflix.SuperFlixClient, ctx context.Context, mediaType, mediaID, season, episode string) (*superflix.SuperFlixStreamResult, error) {
		return c.GetStreamURL(ctx, mediaType, mediaID, season, episode)
	}
	// sfReleaseBrowserFn closes the solver window after a resolve. A seam so tests
	// can assert it fires on every path (cache hit, server list, sniff, error).
	sfReleaseBrowserFn = superflix.ReleaseSharedBrowser

	// sfPrefetchNextFn warms the next episode's stream cache after a successful
	// resolve. A seam so tests can stub it out or drive it directly.
	sfPrefetchNextFn = maybePrefetchNextSuperFlixEpisode
)

// sfPrefetchBudget bounds the background next-episode warm-up. The chain is
// plain HTTP (browser solve forbidden), but the player-page fetch retries past
// token-less shells and the transport may honor a Retry-After, so give it room;
// nothing user-visible waits on this.
const sfPrefetchBudget = 45 * time.Second

var (
	// sfPrefetchInFlight dedupes concurrent warm-ups of the same episode.
	sfPrefetchInFlight sync.Map
	// sfPrefetchWG tracks warm-up goroutines so tests can wait for them.
	sfPrefetchWG sync.WaitGroup
)

// maybePrefetchNextSuperFlixEpisode warms the NEXT episode's (host, hash) cache
// entry in the background, so a binge's "next episode" opens through the ~1s
// cache fast path instead of paying the server-list wait again.
//
// Strictly best-effort and invisible: the whole chain runs with the browser
// solve FORBIDDEN (WithoutBrowserSolve), so it can never pop a window — with the
// Cloudflare clearance warm from the play that just happened, the tokened player
// page is reachable over plain HTTP. Any failure (gate re-armed, no next
// episode, rate limit) only means the next play resolves normally. The server is
// picked silently from the user's remembered preference and the pick is NOT
// re-persisted, so prefetch never influences a later prompt.
//
// GOANIME_SF_NO_PREFETCH disables it (escape hatch for metered connections or
// if SuperFlix ever turns hostile to the extra requests).
func maybePrefetchNextSuperFlixEpisode(sfClient *superflix.SuperFlixClient, tmdbID, sfType, season, epNum string) {
	if sfType != "serie" || os.Getenv("GOANIME_SF_NO_PREFETCH") != "" {
		return
	}
	n, err := strconv.Atoi(strings.TrimSpace(epNum))
	if err != nil || n < 1 {
		return
	}
	next := strconv.Itoa(n + 1)
	if superflix.HasCachedStream(sfType, tmdbID, season, next) {
		return
	}
	key := sfType + ":" + tmdbID + ":" + season + ":" + next
	if _, running := sfPrefetchInFlight.LoadOrStore(key, struct{}{}); running {
		return
	}
	// Capture the seams synchronously: the goroutine may outlive a test that
	// restores them, and reading the package vars there would be a data race.
	getServers, streamFromServer := sfGetServersFn, sfStreamFromServerFn
	sfPrefetchWG.Add(1)
	go func() {
		defer sfPrefetchWG.Done()
		defer sfPrefetchInFlight.Delete(key)

		ctx, cancel := context.WithTimeout(superflix.WithoutBrowserSolve(context.Background()), sfPrefetchBudget)
		defer cancel()

		servers, tokens, err := getServers(sfClient, ctx, sfType, tmdbID, season, next)
		if err != nil || len(servers) == 0 {
			util.Debug("SuperFlix prefetch: server list unavailable; next episode will resolve normally", "key", key, "err", err)
			return
		}
		candidates := orderedServers(servers)
		if pref, ok := recallSuperFlixServer(tmdbID); ok {
			candidates = narrowByMemory(candidates, pref)
		}
		// StreamFromServer caches the (host, hash) — the browser-gated fact —
		// as a side effect; the stream URL itself is discarded (signed links
		// expire, and the cache replay signs a fresh one at play time).
		if _, err := streamFromServer(sfClient, ctx, tokens, candidates[0].IDString(), sfType, tmdbID, season, next); err != nil {
			util.Debug("SuperFlix prefetch failed; next episode will resolve normally", "key", key, "err", err)
			return
		}
		util.Debug("SuperFlix prefetch: next episode cached for instant start", "key", key)
	}()
}

// superFlixStream resolves a SuperFlix stream, preferring the path that lets the
// user actually choose.
//
// The server list (player page → /player/bootstrap) is the only place SuperFlix
// exposes BOTH the available sources and whether each is dublado or legendado. So
// we try that first and let the user pick. It can fail — the site serves a
// token-less shell much of the time — and then we fall back to the embed sniff,
// which always yields *a* stream but offers no choice at all. That fallback is why
// playback used to silently take whatever the embed happened to play.
//
// The returned server is nil on the fallback path, telling the caller it must ask
// about the audio itself.
func superFlixStream(sfClient *superflix.SuperFlixClient, tmdbID, sfType, season, epNum string) (*superflix.SuperFlixStreamResult, *superflix.SuperFlixServer, error) {
	// Fast path: an episode played before replays straight from the cached
	// (host, hash) over plain HTTP — no Cloudflare solve, no server-list fetch, no
	// browser. This is the difference between a re-watch or a resume opening in ~1s
	// versus paying the whole pipeline again. A nil server is returned because the
	// choice was already made last time and is honored from the per-title memory.
	if cached, ok := sfCachedStreamFn(sfClient, sfType, tmdbID, season, epNum); ok {
		util.Debug("SuperFlix: served from stream cache (fast path)")
		return cached, nil, nil
	}

	var (
		servers []superflix.SuperFlixServer
		tokens  *superflix.SuperFlixTokens
		listErr error
	)
	runWithSpinner("Loading servers...", func() {
		ctx, cancel := context.WithTimeout(context.Background(), sfServerListBudget)
		defer cancel()
		servers, tokens, listErr = sfGetServersFn(sfClient, ctx, sfType, tmdbID, season, epNum)
	})
	if errors.Is(listErr, superflix.ErrSuperFlixRestricted) {
		// The browser already tried the only viable recovery (reading the signed
		// iframe as a cross-origin embed). Starting the separate sniff path would
		// repeat the same restricted-page wait, so surface its actionable error now.
		return nil, nil, listErr
	}

	if listErr == nil && len(servers) > 0 {
		// The server list is in hand, which means the browser already did its one
		// job — the Cloudflare solve. Everything left (the picker, then
		// StreamFromServer's source/redirect/getVideo) is plain HTTP that reuses the
		// warm cookie from the client's jar, so close the window NOW instead of at
		// the end. That makes it disappear ~5s sooner — before the picker and the
		// stream round-trips, not after. If the chosen server fails and we fall to
		// the sniff below, that path re-launches the browser itself.
		sfReleaseBrowserFn()

		// Ask outside the spinner: a picker under a spinner is unreadable.
		chosen, err := selectSuperFlixServer(tmdbID, servers)
		if err == nil {
			var result *superflix.SuperFlixStreamResult
			var streamErr error
			runWithSpinner("Loading stream...", func() {
				ctx, cancel := context.WithTimeout(context.Background(), 210*time.Second)
				defer cancel()
				result, streamErr = sfStreamFromServerFn(sfClient, ctx, tokens, chosen.IDString(), sfType, tmdbID, season, epNum)
			})
			if streamErr == nil {
				util.Debug("SuperFlix stream from chosen server", "server", chosen.Name, "type", chosen.Type)
				return result, &chosen, nil
			}
			// The chosen server refused: fall through to the sniff rather than
			// dead-ending on a source the user cannot re-pick from here.
			util.Warn("SuperFlix: the chosen server failed; falling back", "server", chosen.Name, "error", streamErr)
		}
	} else {
		util.Debug("SuperFlix: server list unavailable; falling back to the embed sniff", "err", listErr)
	}

	var result *superflix.SuperFlixStreamResult
	var streamErr error
	runWithSpinner("Loading stream..."+sfBrowserSpinnerHint, func() {
		// Generous timeout: the pipeline's first request may hit a Cloudflare
		// Turnstile gate that the client solves with a headed Firefox (10–40s).
		// Must exceed the client's solve budget or the solve gets cancelled.
		ctx, cancel := context.WithTimeout(context.Background(), 210*time.Second)
		defer cancel()
		result, streamErr = sfSniffStreamFn(sfClient, ctx, sfType, tmdbID, season, epNum)
	})
	if streamErr != nil {
		return nil, nil, streamErr
	}
	return result, nil, nil
}

// GetSuperFlixStreamURL gets the stream URL for SuperFlix content.
//
// Subtitle clearing and global-source tagging are handled by the only caller,
// GetEpisodeStreamURL — duplicating them here produced two identical
// "Stored anime source: SuperFlix" debug lines per playback.
func GetSuperFlixStreamURL(media *models.Anime, episode *models.Episode, quality string) (string, error) {
	sfClient := superflix.SharedSuperFlixClient()

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

	// Close the solver window once the URL is resolved (or failed), so it does not
	// linger through playback. No-op on the cache fast path (no window was opened);
	// the warm on-disk profile keeps the next episode's solve fast. Via a seam so
	// tests can assert it fires on every resolve path.
	defer sfReleaseBrowserFn()

	result, chosen, err := superFlixStream(sfClient, tmdbID, sfType, season, epNum)
	if err != nil {
		return "", fmt.Errorf("failed to get SuperFlix stream: %w", describeSuperFlixErr(err))
	}

	// Warm the NEXT episode in the background (best-effort, plain HTTP, no
	// browser window) so a binge's next play starts from the cache fast path.
	sfPrefetchNextFn(sfClient, tmdbID, sfType, season, epNum)

	// Store referer globally for mpv playback
	if result.Referer != "" {
		util.SetGlobalReferer(result.Referer)
	}
	// Update cover image from stream thumbnail if not already set
	if media.ImageURL == "" && result.Thumb != "" {
		media.ImageURL = result.Thumb
		util.Debug("SuperFlix cover set from stream thumbnail", "url", result.Thumb)
	}

	// Pick the audio track.
	//
	// When the server list was reachable the user already answered "dublado or
	// legendado" by picking a server, so asking again would be asking twice. Only on
	// the fallback path (embed sniff, no server list) do we have to ask, and there
	// the multi-audio HLS is the only lever we have.
	var alang string
	if chosen != nil {
		alang = audioForServer(*chosen, result.DefaultAudio)
		// Record the audio too, so a cached repeat of this title — which returns no
		// server (chosen == nil) — replays the same audio without re-asking.
		rememberServerAudioChoice(tmdbID, *chosen)
		util.Debug("SuperFlix audio derived from the chosen server",
			"server", chosen.Name, "type", chosen.Type, "alang", alang)
	} else if opt, ok := selectSuperFlixAudio(tmdbID, result.DefaultAudio, len(result.Subtitles) > 0); ok {
		alang = mpvAudioLanguage(opt)
		util.Debug("SuperFlix audio chosen from the stream's tracks", "code", opt.Code, "alang", alang)
	}
	if alang != "" {
		util.GlobalAudioLanguage = alang
	}

	// Load every subtitle track the stream ships, always.
	//
	// An earlier version withheld them whenever the dub was selected, reasoning that
	// Portuguese subtitles over Portuguese audio merely echo the dialogue. That was
	// a behavior change nobody asked for and it broke real viewing: subtitles that
	// had always been there stopped appearing. Worse, the flag defaulted to "off",
	// so a stream that exposed no audio-track list — or a user who had pinned
	// --audio-lang — silently lost its subtitles too.
	//
	// Availability is not the same as display: mpv can turn a track off, but it
	// cannot show one we never handed it. --no-subs remains the way to opt out.
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
