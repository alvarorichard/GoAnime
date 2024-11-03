package main

import (
	"fmt"
	"log"
	"strconv"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
)

func main() {
	animeName, err := util.FlagParser()
	if err != nil {
		log.Fatalln(util.ErrorHandler(err))
	}

	anime, err := api.SearchAnime(animeName)
	if err != nil {
		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
	}

	episodes, err := api.GetAnimeEpisodes(anime.URL)
	if err != nil || len(episodes) == 0 {
		if util.IsDebug {
			log.Fatalln("The selected anime has no episodes on the server:", util.ErrorHandler(err))
		}
		log.Fatalln("The selected anime has no episodes on the server.")
	}

	series, totalEpisodes, err := api.IsSeries(anime.URL)
	if err != nil {
		log.Fatalln("Error checking if the anime is a series:", util.ErrorHandler(err))
	}

	if series {
		fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)
		selectedEpisodeURL, episodeNumberStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
		if err != nil {
			log.Fatalln(util.ErrorHandler(err))
		}

		selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
		if err != nil {
			log.Fatalln("Error parsing episode number:", util.ErrorHandler(err))
		}

		videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
		if err != nil {
			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
		}

		player.HandleDownloadAndPlay(videoURL, episodes, selectedEpisodeNum, anime.URL, episodeNumberStr)
	} else {
		fmt.Println("The selected anime is a movie/OVA. Starting direct playback...")
		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
		if err != nil {
			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
		}

		player.HandleDownloadAndPlay(videoURL, episodes, 1, anime.URL, episodes[0].Number)
	}
}
