package appflow

import (
	"errors"
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/huh/spinner"
)

func SearchAnime(name string) *models.Anime {
	searchStart := time.Now()

	// Use enhanced API with source selection (spinner is inside api.SearchAnimeEnhanced)
	anime, err := api.SearchAnimeEnhanced(name, util.GlobalSource)
	if err != nil {
		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
	}

	util.Debugf("[PERF] SearchAnime completed in %v", time.Since(searchStart))
	return anime
}

// SearchAnimeEnhanced - busca em ambas as fontes (AllAnime e AnimeFire) simultaneamente
func SearchAnimeEnhanced(name string) *models.Anime {
	searchStart := time.Now()

	// Buscar em ambas as fontes (spinner is inside api.SearchAnimeEnhanced)
	anime, err := api.SearchAnimeEnhanced(name, "")
	if err != nil {
		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
	}

	util.Debugf("[PERF] SearchAnimeEnhanced completed in %v", time.Since(searchStart))
	return anime
}

// SearchAnimeWithRetry - searches for anime with retry logic on failure
func SearchAnimeWithRetry(name string) (*models.Anime, error) {
	currentName := name

	for {
		searchStart := time.Now()

		// Attempt to search for anime (spinner is inside api.SearchAnimeEnhanced)
		util.Debugf("Searching for: %s (searching all sources)", currentName)
		anime, searchErr := api.SearchAnimeEnhanced(currentName, "")

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
			// For FlixHQ movies/TV shows, use TMDB enrichment instead of AniList
			if anime.Source == "FlixHQ" || anime.MediaType == models.MediaTypeMovie || anime.MediaType == models.MediaTypeTV {
				util.Debugf("Skipping AniList enrichment for FlixHQ content: %s", anime.Name)
				// The anime already has ImageURL from FlixHQ scraping
				// Optionally enrich with TMDB for more metadata (IMDB ID, rating, etc.)
				if err := api.FetchAnimeDetails(anime); err != nil {
					util.Debugf("Failed to enrich FlixHQ content with TMDB: %v", err)
				}
				util.Debugf("[PERF] FetchAnimeDetails (FlixHQ) completed in %v", time.Since(detailsStart))
				return
			}

			// ALWAYS enrich anime with AniList data
			// This is essential for Discord integration, AniSkip, etc.
			// The original system ALWAYS uses AniList images for anime

			// Use the enrichment function from the original system
			aniListInfo, err := api.FetchAnimeFromAniList(anime.Name)
			if err != nil {
				util.Debugf("Failed to fetch from AniList: %v", err)
			} else {
				// Enrich the anime with AniList data
				anime.AnilistID = aniListInfo.Data.Media.ID
				anime.MalID = aniListInfo.Data.Media.IDMal
				anime.Details = aniListInfo.Data.Media

				// ALWAYS use AniList image (as in the original system)
				if cover := aniListInfo.Data.Media.CoverImage.Large; cover != "" {
					anime.ImageURL = cover
				} else {
					util.Debugf("Cover image not found for: %s", anime.Name)
				}

				util.Debugf("Anime enriched successfully with AniList data - ID: %d, MAL: %d, Image: %s",
					anime.AnilistID, anime.MalID, anime.ImageURL)
			}

			// Fallback: try to fetch source-specific details if needed
			if anime.Source == "AllAnime" && len(anime.URL) > 20 && strings.Contains(anime.URL, "allanime.to") {
				if err := api.FetchAnimeDetails(anime); err != nil {
					util.Debugf("Failed to fetch anime details from source: %v", err)
				}
			}
		}).
		Run()

	util.Debugf("[PERF] FetchAnimeDetails completed in %v", time.Since(detailsStart))
}

func GetAnimeEpisodes(anime *models.Anime) []models.Episode {
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

	if fetchErr != nil || len(episodes) == 0 {
		log.Fatalln("The selected anime does not have episodes on the server.")
	}

	util.Debugf("[PERF] GetAnimeEpisodes completed in %v", time.Since(episodesStart))
	return episodes
}

// GetAnimeEpisodesLegacy - compatibility function for old URL-based calls
func GetAnimeEpisodesLegacy(url string) []models.Episode {
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

	if fetchErr != nil || len(episodes) == 0 {
		log.Fatalln("The selected anime does not have episodes on the server.")
	}

	util.Debugf("[PERF] GetAnimeEpisodesLegacy completed in %v", time.Since(episodesStart))
	return episodes
}
