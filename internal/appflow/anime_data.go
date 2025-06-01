package appflow

import (
	"log"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

func SearchAnime(name string) *models.Anime {
	searchStart := time.Now()
	anime, err := api.SearchAnime(name)
	if err != nil {
		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
	}
	if util.IsDebug {
		log.Printf("[PERF] Busca de anime em %v", time.Since(searchStart))
	}
	return anime
}

func FetchAnimeDetails(anime *models.Anime) {
	detailsStart := time.Now()
	if err := api.FetchAnimeDetails(anime); err != nil {
		log.Println("Failed to fetch anime details:", err)
	}
	if util.IsDebug {
		log.Printf("[PERF] Search in details %v", time.Since(detailsStart))
	}
}

func GetAnimeEpisodes(url string) []models.Episode {
	episodesStart := time.Now()
	episodes, err := api.GetAnimeEpisodes(url)
	if err != nil || len(episodes) == 0 {
		log.Fatalln("The selected anime does not have episodes on the server.")
	}
	if util.IsDebug {
		log.Printf("[PERF] Search Episode in %v", time.Since(episodesStart))
	}
	return episodes
}
