package appflow

import (
	"log"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
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

func FetchAnimeDetails(anime *models.Anime) {
	detailsStart := time.Now()

	// For AllAnime animes with short IDs, we can skip detailed fetching
	// since we already have the essential information
	if anime.Source == "AllAnime" && len(anime.URL) > 20 && strings.Contains(anime.URL, "allanime.to") {
		if err := api.FetchAnimeDetails(anime); err != nil {
			log.Println("Failed to fetch anime details:", err)
		}
	} else {
		// For short IDs or other sources, we already have the basic info
		if util.IsDebug {
			util.Debugf("Skipping detailed fetch for %s anime with ID: %s", anime.Source, anime.URL)
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
