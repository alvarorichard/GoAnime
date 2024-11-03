//package main
//
//import (
//	"fmt"
//	"log"
//	"strconv"
//
//	"github.com/alvarorichard/Goanime/internal/api"
//	"github.com/alvarorichard/Goanime/internal/player"
//	"github.com/alvarorichard/Goanime/internal/util"
//)
//
//func main() {
//	animeName, err := util.FlagParser()
//	if err != nil {
//		log.Fatalln(util.ErrorHandler(err))
//	}
//
//	anime, err := api.SearchAnime(animeName)
//	if err != nil {
//		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
//	}
//
//	episodes, err := api.GetAnimeEpisodes(anime.URL)
//	if err != nil || len(episodes) == 0 {
//		if util.IsDebug {
//			log.Fatalln("The selected anime has no episodes on the server:", util.ErrorHandler(err))
//		}
//		log.Fatalln("The selected anime has no episodes on the server.")
//	}
//
//	series, totalEpisodes, err := api.IsSeries(anime.URL)
//	if err != nil {
//		log.Fatalln("Error checking if the anime is a series:", util.ErrorHandler(err))
//	}
//
//	if series {
//		fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)
//		selectedEpisodeURL, episodeNumberStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
//		if err != nil {
//			log.Fatalln(util.ErrorHandler(err))
//		}
//
//		selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
//		if err != nil {
//			log.Fatalln("Error parsing episode number:", util.ErrorHandler(err))
//		}
//
//		videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
//		if err != nil {
//			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
//		}
//
//		player.HandleDownloadAndPlay(videoURL, episodes, selectedEpisodeNum, anime.URL, episodeNumberStr)
//	} else {
//		fmt.Println("The selected anime is a movie/OVA. Starting direct playback...")
//		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
//		if err != nil {
//			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
//		}
//
//		player.HandleDownloadAndPlay(videoURL, episodes, 1, anime.URL, episodes[0].Number)
//	}
//}

package main

import (
	"fmt"
	"log"
	"strconv"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/hugolgst/rich-go/client"
)

const discordClientID = "1302721937717334128" // Seu Client ID do Discord

// Goroutine to continuously update Discord Rich Presence
func startDiscordPresenceUpdater(anime *api.Anime, isPaused *bool) {
	ticker := time.NewTicker(5 * time.Second) // Update every 5 seconds
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			// Update the Discord presence
			err := api.DiscordPresence(discordClientID, *anime, *isPaused)
			if err != nil {
				log.Println("Erro ao atualizar o Discord Rich Presence:", err)
			}
		}
	}
}

func main() {
	// Parse flags to get the anime name
	animeName, err := util.FlagParser()
	if err != nil {
		log.Fatalln(util.ErrorHandler(err))
	}

	// Initialize Discord Rich Presence
	err = client.Login(discordClientID)
	if err != nil {
		log.Fatalln("Falha ao inicializar Discord Rich Presence:", err)
	}
	defer client.Logout() // Ensure logout on exit

	// Search for the anime
	anime, err := api.SearchAnime(animeName)
	if err != nil {
		log.Fatalln("Falha ao buscar anime:", util.ErrorHandler(err))
	}

	// Fetch anime details, including cover image URL
	err = api.FetchAnimeDetails(anime)
	if err != nil {
		log.Println("Falha ao buscar detalhes do anime:", err)
	}

	// Fetch episodes for the anime
	episodes, err := api.GetAnimeEpisodes(anime.URL)
	if err != nil || len(episodes) == 0 {
		log.Fatalln("O anime selecionado não possui episódios no servidor.")
	}

	// Check if the anime is a series or a movie/OVA
	series, totalEpisodes, err := api.IsSeries(anime.URL)
	if err != nil {
		log.Fatalln("Erro ao verificar se o anime é uma série:", util.ErrorHandler(err))
	}

	// Define a flag to track whether the playback is paused
	isPaused := false

	if series {
		fmt.Printf("O anime selecionado é uma série com %d episódios.\n", totalEpisodes)

		// Select an episode with fuzzy finder
		selectedEpisodeURL, episodeNumberStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
		if err != nil {
			log.Fatalln(util.ErrorHandler(err))
		}

		selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
		if err != nil {
			log.Fatalln("Erro ao converter número do episódio:", util.ErrorHandler(err))
		}

		// Update the anime struct with the selected episode
		anime.Episodes = []api.Episode{
			{
				Number: episodeNumberStr,
				Num:    selectedEpisodeNum,
				URL:    selectedEpisodeURL,
			},
		}

		// Start the goroutine to continuously update Discord Rich Presence
		go startDiscordPresenceUpdater(anime, &isPaused)

		// Get the video URL for the selected episode
		videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
		if err != nil {
			log.Fatalln("Falha ao extrair URL do vídeo:", util.ErrorHandler(err))
		}

		// Handle download and play, updating the paused state as necessary
		player.HandleDownloadAndPlay(videoURL, episodes, selectedEpisodeNum, anime.URL, episodeNumberStr)

	} else {
		fmt.Println("O anime selecionado é um filme/OVA. Iniciando reprodução direta...")

		// Update the anime struct with the first episode (movie/OVA)
		anime.Episodes = []api.Episode{
			episodes[0],
		}

		// Start the goroutine to continuously update Discord Rich Presence
		go startDiscordPresenceUpdater(anime, &isPaused)

		// Get the video URL for the movie/OVA
		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
		if err != nil {
			log.Fatalln("Falha ao extrair URL do vídeo:", util.ErrorHandler(err))
		}

		// Handle download and play, updating the paused state as necessary
		player.HandleDownloadAndPlay(videoURL, episodes, 1, anime.URL, episodes[0].Number)
	}
}
