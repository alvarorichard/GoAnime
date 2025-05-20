package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/hugolgst/rich-go/client"
)

const discordClientID = "1302721937717334128"

func main() {
	startAll := time.Now()
	if util.IsDebug {
		log.Printf("[PERF] Início do programa")
	}
	var animeMutex sync.Mutex

	animeName, err := util.FlagParser()
	if err != nil {
		log.Fatalln(util.ErrorHandler(err))
	}

	discordStart := time.Now()
	discordEnabled := true
	if err := client.Login(discordClientID); err != nil {
		if util.IsDebug {
			log.Println("Failed to initialize Discord Rich Presence:", err)
		}
		discordEnabled = false
	} else {
		if util.IsDebug {
			log.Printf("[PERF] Discord pronto em %v", time.Since(discordStart))
		}
		defer client.Logout()
	}

	searchStart := time.Now()
	anime, err := api.SearchAnime(animeName)
	if err != nil {
		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
	}
	if util.IsDebug {
		log.Printf("[PERF] Busca de anime em %v", time.Since(searchStart))
	}

	detailsStart := time.Now()
	if err = api.FetchAnimeDetails(anime); err != nil {
		log.Println("Failed to fetch anime details:", err)
	}
	if util.IsDebug {
		log.Printf("[PERF] Busca de detalhes em %v", time.Since(detailsStart))
	}

	episodesStart := time.Now()
	episodes, err := api.GetAnimeEpisodes(anime.URL)
	if err != nil || len(episodes) == 0 {
		log.Fatalln("The selected anime does not have episodes on the server.")
	}
	if util.IsDebug {
		log.Printf("[PERF] Busca de episódios em %v", time.Since(episodesStart))
		log.Printf("[PERF] Inicialização total em %v", time.Since(startAll))
	}

	series, totalEpisodes, err := api.IsSeries(anime.URL)
	if err != nil {
		log.Fatalln("Error checking if the anime is a series:", util.ErrorHandler(err))
	}

	isPaused := false
	socketPath := "/tmp/mpvsocket"
	updateFreq := 1 * time.Second
	episodeDuration := time.Duration(episodes[0].Duration) * time.Second

	if series {
		fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)

		for {
			selectedEpisodeURL, episodeNumberStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
			if err != nil {
				log.Fatalln(util.ErrorHandler(err))
			}

			selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
			if err != nil {
				log.Fatalln("Error converting episode number:", util.ErrorHandler(err))
			}

			animeMutex.Lock()
			anime.Episodes = []models.Episode{
				{
					Number: episodeNumberStr,
					Num:    selectedEpisodeNum,
					URL:    selectedEpisodeURL,
				},
			}
			animeMutex.Unlock()

			if err = api.GetEpisodeData(anime.MalID, selectedEpisodeNum, anime); err != nil {
				log.Printf("Error fetching episode data: %v", err)
			}

			videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
			if err != nil {
				log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
			}

			var updater *player.RichPresenceUpdater
			if discordEnabled {
				updater = player.NewRichPresenceUpdater(
					anime,
					&isPaused,
					&animeMutex,
					updateFreq,
					episodeDuration,
					socketPath,
				)
			}

			player.HandleDownloadAndPlay(
				videoURL,
				episodes,
				selectedEpisodeNum,
				anime.URL,
				episodeNumberStr,
				anime.MalID,
				updater,
			)

			// Explicit cleanup after playback
			if updater != nil {
				updater.Stop()
			}

			var userInput string
			fmt.Print("Press 'n' for next episode, 'p' for previous episode, 'q' to quit: ")
			_, err = fmt.Scanln(&userInput)
			if err != nil {
				// Handle different error types
				if err.Error() == "unexpected newline" {
					log.Println("No input detected, continuing playback")
					userInput = "n" // Default to next episode
				} else {
					log.Printf("Error reading input: %v - defaulting to continue", util.ErrorHandler(err))
					userInput = "n"
				}
			}

			if userInput == "q" {
				log.Println("Quitting application as per user request.")
				break
			} else if userInput == "p" {
				// Handle previous episode logic
				selectedEpisodeNum = m(1, selectedEpisodeNum-1)
			} else {
				// Default to next episode
				selectedEpisodeNum = i(totalEpisodes, selectedEpisodeNum+1)
			}
		}
	} else {
		animeMutex.Lock()
		anime.Episodes = []models.Episode{episodes[0]}
		animeMutex.Unlock()

		if err = api.GetMovieData(anime.MalID, anime); err != nil {
			log.Printf("Error fetching movie/OVA data: %v", err)
		}

		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
		if err != nil {
			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
		}

		var updater *player.RichPresenceUpdater
		if discordEnabled {
			updater = player.NewRichPresenceUpdater(
				anime,
				&isPaused,
				&animeMutex,
				updateFreq,
				episodeDuration,
				socketPath,
			)
		}

		player.HandleDownloadAndPlay(
			videoURL,
			episodes,
			1,
			anime.URL,
			episodes[0].Number,
			anime.MalID,
			updater,
		)

		if updater != nil {
			updater.Stop()
		}
	}
}

func m(a, b int) int {
	if a > b {
		return a
	}
	return b
}

func i(a, b int) int {
	if a < b {
		return a
	}
	return b
}
