package appflow

import (
	"errors"
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"

	"charm.land/huh/v2"
	"charm.land/huh/v2/spinner"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// SearchAnime searches for an anime by name using the globally configured source.
func SearchAnime(name string) (*models.Anime, error) {
	searchStart := time.Now()

	// Use enhanced API with source selection (spinner is inside api.SearchAnimeEnhanced)
	anime, err := api.SearchAnimeEnhanced(name, util.GlobalSource)
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
	anime, err := api.SearchAnimeEnhanced(name, "")
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
		anime, searchErr := api.SearchAnimeEnhanced(currentName, source)

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

		// Prompt user for new input
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

		if promptErr := prompt.Run(); promptErr != nil {
			return nil, fmt.Errorf("search cancelled by user")
		}

		currentName = strings.TrimSpace(newName)
		if currentName == "" {
			return nil, fmt.Errorf("search cancelled: empty name provided")
		}
	}
}

func FetchAnimeDetails(anime *models.Anime) {
	detailsStart := time.Now()

	// Use spinner while fetching details
	_ = spinner.New().
		Title("Fetching anime details...").
		Type(spinner.Dots).
		Action(func() {
			// For FlixHQ/SuperFlix movies/TV shows, use TMDB enrichment instead of AniList
			if anime.Source == "FlixHQ" || anime.Source == "SuperFlix" || anime.MediaType == models.MediaTypeMovie || anime.MediaType == models.MediaTypeTV {
				util.Debugf("Skipping AniList enrichment for movie/TV content: %s (source: %s)", anime.Name, anime.Source)
				// SuperFlix stores TMDB ID in URL, not a web page URL, so skip the old page scraping
				if anime.Source != "SuperFlix" {
					if err := api.FetchAnimeDetails(anime); err != nil {
						util.Debugf("Failed to enrich content with TMDB: %v", err)
					}
				}
				util.Debugf("[PERF] FetchAnimeDetails (movie/TV) completed in %v", time.Since(detailsStart))
				return
			}

			// Skip AniList enrichment if already done during search (enrichAnimeData)
			needsAniList := anime.AnilistID <= 0 || anime.MalID <= 0 || anime.ImageURL == ""
			needsSourceDetails := anime.Source == "AllAnime" && len(anime.URL) > 20 && strings.Contains(anime.URL, "allanime.to")

			if needsAniList && needsSourceDetails {
				// Both needed — run in parallel
				var wg sync.WaitGroup
				wg.Add(2)

				go func() {
					defer wg.Done()
					aniListInfo, err := api.FetchAnimeFromAniList(anime.Name)
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
				}()

				go func() {
					defer wg.Done()
					if err := api.FetchAnimeDetails(anime); err != nil {
						util.Debugf("Failed to fetch anime details from source: %v", err)
					}
				}()

				wg.Wait()
			} else if needsAniList {
				aniListInfo, err := api.FetchAnimeFromAniList(anime.Name)
				if err != nil {
					util.Debugf("Failed to fetch from AniList: %v", err)
				} else {
					anime.AnilistID = aniListInfo.Data.Media.ID
					anime.MalID = aniListInfo.Data.Media.IDMal
					anime.Details = aniListInfo.Data.Media
					if cover := aniListInfo.Data.Media.CoverImage.Large; cover != "" {
						anime.ImageURL = cover
					}
					util.Debugf("Anime enriched with AniList data - ID: %d, MAL: %d", anime.AnilistID, anime.MalID)
				}
			} else {
				util.Debugf("AniList data already present (ID: %d, MAL: %d), skipping redundant fetch", anime.AnilistID, anime.MalID)
				if needsSourceDetails {
					if err := api.FetchAnimeDetails(anime); err != nil {
						util.Debugf("Failed to fetch anime details from source: %v", err)
					}
				}
			}
		}).
		Run()

	util.Debugf("[PERF] FetchAnimeDetails completed in %v", time.Since(detailsStart))
}

// GetAnimeEpisodes fetches the episode list for the given anime from its source.
func GetAnimeEpisodes(anime *models.Anime) ([]models.Episode, error) {
	episodesStart := time.Now()

	var episodes []models.Episode
	var fetchErr error

	// For FlixHQ content, don't wrap in spinner here because GetFlixHQEpisodes
	// has UI interactions (season selection) and handles its own spinners for network calls
	if anime.Source == "FlixHQ" || anime.MediaType == models.MediaTypeMovie || anime.MediaType == models.MediaTypeTV {
		episodes, fetchErr = api.GetAnimeEpisodesEnhanced(anime)
	} else {
		// Use spinner while fetching episodes for non-FlixHQ content
		_ = spinner.New().
			Title("Loading episodes...").
			Type(spinner.Dots).
			Action(func() {
				// Use enhanced API for episode fetching
				episodes, fetchErr = api.GetAnimeEpisodesEnhanced(anime)
			}).
			Run()
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

// GetAnimeEpisodesLegacy - compatibility function for old URL-based calls
func GetAnimeEpisodesLegacy(url string) ([]models.Episode, error) {
	episodesStart := time.Now()

	var episodes []models.Episode
	var fetchErr error

	// Use spinner while fetching episodes
	_ = spinner.New().
		Title("Loading episodes...").
		Type(spinner.Dots).
		Action(func() {
			episodes, fetchErr = api.GetAnimeEpisodes(url)
		}).
		Run()

	if fetchErr != nil {
		return nil, fmt.Errorf("failed to fetch episodes: %w", fetchErr)
	}
	if len(episodes) == 0 {
		return nil, fmt.Errorf("the selected anime does not have episodes on the server")
	}

	util.Debugf("[PERF] GetAnimeEpisodesLegacy completed in %v", time.Since(episodesStart))
	return episodes, nil
}
