package appflow

import (
	"fmt"
	"log"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/huh"
)

func SearchAnime(name string) *models.Anime {
	searchStart := time.Now()

	// Use enhanced API with source selection
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

	// Buscar em ambas as fontes (source = "" significa buscar em todas)
	anime, err := api.SearchAnimeEnhanced(name, "")
	if err != nil {
		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
	}

	util.Debugf("[PERF] SearchAnimeEnhanced completed in %v", time.Since(searchStart))
	return anime
}

// SearchAnimeWithRetry - searches for anime with retry logic on failure
func SearchAnimeWithRetry(name string) (*models.Anime, error) {
	const maxRetries = 3
	currentName := name

	for i := 0; i < maxRetries; i++ {
		searchStart := time.Now()

		// Attempt to search for anime (empty string means search all sources)
		util.Debugf("Search attempt %d/%d for: %s (searching all sources)", i+1, maxRetries, currentName)
		anime, err := api.SearchAnimeEnhanced(currentName, "")

		if err == nil && anime != nil {
			util.Debugf("[PERF] SearchAnimeWithRetry completed in %v", time.Since(searchStart))
			return anime, nil
		}

		// Display error message to user
		if i < maxRetries-1 {
			util.Errorf("No anime found with the name: %s", currentName)
			util.Infof("Please try again with a different search term.")

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
		} else {
			// Last attempt failed
			return nil, fmt.Errorf("failed to find anime after %d attempts", maxRetries)
		}
	}

	return nil, fmt.Errorf("failed to find anime after %d attempts", maxRetries)
}

func FetchAnimeDetails(anime *models.Anime) {
	detailsStart := time.Now()

	// SEMPRE enriquecer com dados do AniList para qualquer fonte
	// Isso é essencial para a integração com Discord, AniSkip, etc.
	// O sistema original SEMPRE usa imagens do AniList

	// Usar a função de enriquecimento que já existe no sistema original
	aniListInfo, err := api.FetchAnimeFromAniList(anime.Name)
	if err != nil {
		util.Debugf("Failed to fetch from AniList: %v", err)
	} else {
		// Enriquecer o anime com dados do AniList
		anime.AnilistID = aniListInfo.Data.Media.ID
		anime.MalID = aniListInfo.Data.Media.IDMal
		anime.Details = aniListInfo.Data.Media

		// SEMPRE usar imagem do AniList (como no sistema original)
		if cover := aniListInfo.Data.Media.CoverImage.Large; cover != "" {
			anime.ImageURL = cover
		} else {
			util.Debugf("Cover image not found for: %s", anime.Name)
		}

		util.Debugf("Anime enriched successfully with AniList data - ID: %d, MAL: %d, Image: %s",
			anime.AnilistID, anime.MalID, anime.ImageURL)
	}

	// Fallback: tentar buscar detalhes específicos da fonte se necessário
	if anime.Source == "AllAnime" && len(anime.URL) > 20 && strings.Contains(anime.URL, "allanime.to") {
		if err := api.FetchAnimeDetails(anime); err != nil {
			util.Debugf("Failed to fetch anime details from source: %v", err)
		}
	}

	util.Debugf("[PERF] FetchAnimeDetails completed in %v", time.Since(detailsStart))
}

func GetAnimeEpisodes(anime *models.Anime) []models.Episode {
	episodesStart := time.Now()

	// Use enhanced API for episode fetching
	episodes, err := api.GetAnimeEpisodesEnhanced(anime)
	if err != nil || len(episodes) == 0 {
		log.Fatalln("The selected anime does not have episodes on the server.")
	}

	util.Debugf("[PERF] GetAnimeEpisodes completed in %v", time.Since(episodesStart))
	return episodes
}

// GetAnimeEpisodesLegacy - compatibility function for old URL-based calls
func GetAnimeEpisodesLegacy(url string) []models.Episode {
	episodesStart := time.Now()
	episodes, err := api.GetAnimeEpisodes(url)
	if err != nil || len(episodes) == 0 {
		log.Fatalln("The selected anime does not have episodes on the server.")
	}
	util.Debugf("[PERF] GetAnimeEpisodesLegacy completed in %v", time.Since(episodesStart))
	return episodes
}
