package main

import (
	"fmt"
	"log"
	"strconv"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/hugolgst/rich-go/client"
)

const discordClientID = "1302721937717334128" // Your Discord Client ID

func main() {
	var animeMutex sync.Mutex

	// Parse flags to get the anime name
	animeName, err := util.FlagParser()
	if err != nil {
		log.Fatalln(util.ErrorHandler(err))
	}

	// Initialize Discord Rich Presence
	discordEnabled := true
	if err := client.Login(discordClientID); err != nil {
		if util.IsDebug {
			log.Println("Failed to initialize Discord Rich Presence:", err)

		}
		discordEnabled = false
	} else {
		defer client.Logout() // Ensure logout on exit
	}

	// Search for the anime
	anime, err := api.SearchAnime(animeName)
	if err != nil {
		log.Fatalln("Failed to search for anime:", util.ErrorHandler(err))
	}

	// Fetch anime details, including cover image URL
	if err = api.FetchAnimeDetails(anime); err != nil {
		log.Println("Failed to fetch anime details:", err)
	}

	// Fetch episodes for the anime
	episodes, err := api.GetAnimeEpisodes(anime.URL)
	if err != nil || len(episodes) == 0 {
		log.Fatalln("The selected anime does not have episodes on the server.")
	}

	// Check if the anime is a series or a movie/OVA
	series, totalEpisodes, err := api.IsSeries(anime.URL)
	if err != nil {
		log.Fatalln("Error checking if the anime is a series:", util.ErrorHandler(err))
	}

	// Define a flag to track if the playback is paused
	isPaused := false
	socketPath := "/tmp/mpvsocket" // Adjust socket path as per your setup
	updateFreq := 1 * time.Second  // Update frequency for Rich Presence
	episodeDuration := time.Duration(episodes[0].Duration) * time.Second

	if series {
		fmt.Printf("The selected anime is a series with %d episodes.\n", totalEpisodes)

		for {
			// Select an episode using fuzzy finder
			selectedEpisodeURL, episodeNumberStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
			if err != nil {
				log.Fatalln(util.ErrorHandler(err))
			}

			selectedEpisodeNum, err := strconv.Atoi(player.ExtractEpisodeNumber(episodeNumberStr))
			if err != nil {
				log.Fatalln("Error converting episode number:", util.ErrorHandler(err))
			}

			// Lock anime struct and update with selected episode
			animeMutex.Lock()
			anime.Episodes = []api.Episode{
				{
					Number: episodeNumberStr,
					Num:    selectedEpisodeNum,
					URL:    selectedEpisodeURL,
				},
			}
			animeMutex.Unlock()

			// Fetch episode details and AniSkip data
			if err = api.GetEpisodeData(anime.MalID, selectedEpisodeNum, anime); err != nil {
				log.Printf("Error fetching episode data: %v", err)
			}

			// Retrieve video URL for the selected episode
			videoURL, err := player.GetVideoURLForEpisode(selectedEpisodeURL)
			if err != nil {
				log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
			}

			// Initialize a new RichPresenceUpdater for this episode if Discord is enabled
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
				defer updater.Stop() // Ensure updater is stopped when done
			} else {
				updater = nil
			}

			// Handle download and playback, updating paused state as necessary
			player.HandleDownloadAndPlay(
				videoURL,
				episodes,
				selectedEpisodeNum,
				anime.URL,
				episodeNumberStr,
				anime.MalID, // Pass the animeMalID here
				updater,
			)

			// Prompt user for next action
			var userInput string
			fmt.Print("Press 'n' for next episode, 'p' for previous episode, 'q' to quit: ")
			fmt.Scanln(&userInput)
			if userInput == "q" {
				log.Println("Quitting application as per user request.")
				break
			} else if userInput == "n" || userInput == "p" {
				continue // loop continues for next or previous episode
			} else {
				log.Println("Invalid input, continuing current episode.")
			}
		}

	} else {
		// Handle movie/OVA playback

		// Lock anime struct and update with the first episode
		animeMutex.Lock()
		anime.Episodes = []api.Episode{episodes[0]}
		animeMutex.Unlock()

		// Fetch details and AniSkip data for the movie/OVA
		if err = api.GetMovieData(anime.MalID, anime); err != nil {
			log.Printf("Error fetching movie/OVA data: %v", err)
		}

		// Get the video URL for the movie/OVA
		videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
		if err != nil {
			log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
		}

		// Initialize a new RichPresenceUpdater for the movie if Discord is enabled
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
			defer updater.Stop()
		} else {
			updater = nil
		}

		// Handle download and play, with Rich Presence updates
		player.HandleDownloadAndPlay(
			videoURL,
			episodes,
			1, // Episode number for movies/OVAs
			anime.URL,
			episodes[0].Number,
			anime.MalID, // Pass the animeMalID here
			updater,
		)
	}

	// No need to call updater.Stop() here as it's deferred after each initialization
}
