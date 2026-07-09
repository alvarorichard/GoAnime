package appflow

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/api/providers"

	"charm.land/huh/v2"
	"charm.land/huh/v2/spinner"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tui"
	"github.com/alvarorichard/Goanime/internal/util"
)

// Injectable package-level dependencies. Tests swap these via the
// helpers in *_test.go (which serialise on appflowOverrideMu). Production
// callers never touch them.
var (
	// searchEnhancedFn is the underlying search implementation.
	searchEnhancedFn = api.SearchAnimeEnhanced

	// searchWithRetryFn is the per-attempt search used inside SearchAnimeWithRetry.
	searchWithRetryFn = api.SearchAnimeEnhanced

	// aniListFetchFn fetches AniList metadata for an anime title.
	aniListFetchFn = api.FetchAnimeFromAniList

	// sourceDetailsFetchFn enriches anime with provider-specific details.
	sourceDetailsFetchFn = api.FetchAnimeDetails

	// getAnimeEpisodesEnhancedFn returns the episode list for an anime. It now
	// dispatches through the Model B registry (providers.FetchEpisodes) instead
	// of api's legacy per-source switch.
	getAnimeEpisodesEnhancedFn = func(anime *models.Anime) ([]models.Episode, error) {
		return providers.FetchEpisodes(context.Background(), anime)
	}

	// getAnimeEpisodesLegacyFn returns episodes by URL (legacy API).
	getAnimeEpisodesLegacyFn = api.GetAnimeEpisodes

	// runSpinnerFn wraps a long action in a TUI spinner. Default uses
	// huh.spinner + tui.RunClean. Tests inject a synchronous passthrough.
	runSpinnerFn = defaultRunSpinner

	// promptForNameFn asks the user for a new search name. Default uses
	// huh.NewInput inside tui.RunClean. Tests inject a scripted sequence.
	promptForNameFn = defaultPromptForName
)

// defaultRunSpinner is the production spinner wrapper. It is replaced by
// tests with a synchronous passthrough so the action runs inline.
func defaultRunSpinner(title string, action func()) {
	_ = tui.RunClean(func() error {
		return spinner.New().
			Title(title).
			Type(spinner.Dots).
			Action(action).
			Run()
	})
}

// defaultPromptForName is the production prompt. Returns the user's input
// trimmed, or an error if cancelled / empty / TTY unavailable.
func defaultPromptForName(_ string) (string, error) {
	var newName string
	prompt := huh.NewInput().
		Title("Search Again").
		Description("Enter a new anime name to search for:").
		Value(&newName).
		Validate(func(v string) error {
			if len(strings.TrimSpace(v)) < 2 {
				return fmt.Errorf("anime name must be at least 2 characters")
			}
			return nil
		})
	if err := tui.RunClean(prompt.Run); err != nil {
		return "", fmt.Errorf("search cancelled by user")
	}
	name := strings.TrimSpace(newName)
	if name == "" {
		return "", fmt.Errorf("search cancelled: empty name provided")
	}
	return name, nil
}

// SearchAnime searches for an anime by name using the globally configured source.
func SearchAnime(name string) (*models.Anime, error) {
	searchStart := time.Now()

	// Use enhanced API with source selection (spinner is inside api.SearchAnimeEnhanced)
	anime, err := searchEnhancedFn(name, util.GlobalSource)
	if err != nil {
		return nil, fmt.Errorf("failed to search for anime: %w", err)
	}

	util.Debugf("[PERF] SearchAnime completed in %v", time.Since(searchStart))
	return anime, nil
}

// SearchAnimeEnhanced - busca em ambas as fontes (AllAnime e AnimeFire) simultaneamente
func SearchAnimeEnhanced(name string) (*models.Anime, error) {
	searchStart := time.Now()

	// Buscar em ambas as fontes (spinner is inside api.SearchAnimeEnhanced)
	anime, err := searchEnhancedFn(name, "")
	if err != nil {
		return nil, fmt.Errorf("failed to search for anime: %w", err)
	}

	util.Debugf("[PERF] SearchAnimeEnhanced completed in %v", time.Since(searchStart))
	return anime, nil
}

// SearchAnimeWithRetry - searches for anime with retry logic on failure
func SearchAnimeWithRetry(name string) (*models.Anime, error) {
	currentName := name

	for {
		searchStart := time.Now()

		// Attempt to search for anime (spinner is inside api.SearchAnimeEnhanced)
		// Respect user's --source flag (e.g. --source allanime) via GlobalSource
		source := util.GlobalSource
		if source != "" {
			util.Debugf("Searching for: %s (source: %s)", currentName, source)
		} else {
			util.Debugf("Searching for: %s (searching all sources)", currentName)
		}
		anime, searchErr := searchWithRetryFn(currentName, source)

		if searchErr == nil && anime != nil {
			util.Debugf("[PERF] SearchAnimeWithRetry completed in %v", time.Since(searchStart))
			return anime, nil
		}

		// Check if user requested to go back to search
		if errors.Is(searchErr, api.ErrBackToSearch) {
			util.Infof("Going back to new search...")
		} else {
			// Display error message to user for other errors
			util.Errorf("No anime found with the name: %s", currentName)
		}

		util.Infof("Please enter a new search term.")

		nextName, promptErr := promptForNameFn(currentName)
		if promptErr != nil {
			return nil, promptErr
		}
		currentName = nextName
	}
}

// FetchAnimeDetails enriches anime with metadata from AniList and/or the
// source provider. Spinner runs around the enrichment via runSpinnerFn.
func FetchAnimeDetails(anime *models.Anime) {
	detailsStart := time.Now()
	runSpinnerFn("Fetching anime details...", func() {
		fetchAnimeDetailsCore(anime)
	})
	util.Debugf("[PERF] FetchAnimeDetails completed in %v", time.Since(detailsStart))
}

// fetchAnimeDetailsCore is the pure orchestration: branch on source / metadata
// state, dispatch to aniListFetchFn / sourceDetailsFetchFn. Fully testable
// with mocks — no TUI, no time-sensitive code paths.
func fetchAnimeDetailsCore(anime *models.Anime) {
	if anime == nil {
		return
	}
	// For FlixHQ/SuperFlix movies/TV shows: skip AniList, optionally enrich.
	if anime.Source == "SFlix" || anime.Source == "SuperFlix" ||
		anime.MediaType == models.MediaTypeMovie || anime.MediaType == models.MediaTypeTV {
		util.Debugf("Skipping AniList enrichment for movie/TV content: %s (source: %s)", anime.Name, anime.Source)
		if anime.Source != "SuperFlix" {
			if err := sourceDetailsFetchFn(anime); err != nil {
				util.Debugf("Failed to enrich content with TMDB: %v", err)
			}
		}
		return
	}

	needsAniList := anime.AnilistID <= 0 || anime.MalID <= 0 || anime.ImageURL == ""
	needsSourceDetails := anime.Source == "AllAnime" && len(anime.URL) > 20 && strings.Contains(anime.URL, "allanime.to")

	switch {
	case needsAniList && needsSourceDetails:
		var wg sync.WaitGroup
		wg.Add(2)
		go func() {
			defer wg.Done()
			enrichFromAniList(anime)
		}()
		go func() {
			defer wg.Done()
			if err := sourceDetailsFetchFn(anime); err != nil {
				util.Debugf("Failed to fetch anime details from source: %v", err)
			}
		}()
		wg.Wait()
	case needsAniList:
		enrichFromAniList(anime)
	default:
		util.Debugf("AniList data already present (ID: %d, MAL: %d), skipping redundant fetch", anime.AnilistID, anime.MalID)
		if needsSourceDetails {
			if err := sourceDetailsFetchFn(anime); err != nil {
				util.Debugf("Failed to fetch anime details from source: %v", err)
			}
		}
	}
}

// enrichFromAniList fetches AniList metadata via aniListFetchFn and applies
// it to the anime. Errors are logged at debug level and ignored — the call
// site treats AniList as best-effort.
func enrichFromAniList(anime *models.Anime) {
	aniListInfo, err := aniListFetchFn(anime.Name)
	if err != nil {
		util.Debugf("Failed to fetch from AniList: %v", err)
		return
	}
	anime.AnilistID = aniListInfo.Data.Media.ID
	anime.MalID = aniListInfo.Data.Media.IDMal
	anime.Details = aniListInfo.Data.Media
	if cover := aniListInfo.Data.Media.CoverImage.Large; cover != "" {
		anime.ImageURL = cover
	}
	util.Debugf("Anime enriched with AniList data - ID: %d, MAL: %d", anime.AnilistID, anime.MalID)
}

// GetAnimeEpisodes fetches the episode list for the given anime from its source.
// FlixHQ content bypasses the spinner since it has its own UI interactions.
func GetAnimeEpisodes(anime *models.Anime) ([]models.Episode, error) {
	episodesStart := time.Now()

	var episodes []models.Episode
	var fetchErr error

	if anime.Source == "SFlix" || anime.MediaType == models.MediaTypeMovie || anime.MediaType == models.MediaTypeTV {
		episodes, fetchErr = getAnimeEpisodesEnhancedFn(anime)
	} else {
		runSpinnerFn("Loading episodes...", func() {
			episodes, fetchErr = getAnimeEpisodesEnhancedFn(anime)
		})
	}

	if fetchErr != nil {
		return nil, fmt.Errorf("failed to fetch episodes: %w", fetchErr)
	}
	if len(episodes) == 0 {
		return nil, fmt.Errorf("the selected anime does not have episodes on the server")
	}

	util.Debugf("[PERF] GetAnimeEpisodes completed in %v", time.Since(episodesStart))
	return episodes, nil
}

// GetAnimeEpisodesLegacy is the URL-based compatibility shim.
func GetAnimeEpisodesLegacy(url string) ([]models.Episode, error) {
	episodesStart := time.Now()

	var episodes []models.Episode
	var fetchErr error

	runSpinnerFn("Loading episodes...", func() {
		episodes, fetchErr = getAnimeEpisodesLegacyFn(url)
	})

	if fetchErr != nil {
		return nil, fmt.Errorf("failed to fetch episodes: %w", fetchErr)
	}
	if len(episodes) == 0 {
		return nil, fmt.Errorf("the selected anime does not have episodes on the server")
	}

	util.Debugf("[PERF] GetAnimeEpisodesLegacy completed in %v", time.Since(episodesStart))
	return episodes, nil
}
