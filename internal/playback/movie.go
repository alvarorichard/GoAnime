package playback

import (
	"errors"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"runtime"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
)

// HandleMovie gerencia a reprodução de filmes/OVAs
func HandleMovie(anime *models.Anime, episodes []models.Episode, discordEnabled bool) {
	for {
		animeMutex := sync.Mutex{}
		isPaused := false

		animeMutex.Lock()
		anime.Episodes = []models.Episode{episodes[0]}
		animeMutex.Unlock()

		if err := api.GetMovieData(anime.MalID, anime); err != nil {
			log.Printf("Error fetching movie/OVA data: %v", err)
		}

		videoURL, err := player.GetVideoURLForEpisodeEnhanced(&episodes[0], anime)
		if err != nil {
			log.Printf("Failed to extract video URL: %v", util.ErrorHandler(err))
			// Try to change anime immediately instead of exiting
			newAnime, newEpisodes, chErr := ChangeAnimeLocal()
			if chErr != nil {
				log.Printf("Error changing anime: %v", chErr)
				// If change fails, ask user on next loop iteration
				continue
			}
			anime = newAnime
			episodes = newEpisodes

			// If new anime is a series, delegate handling and exit movie loop
			series, totalEpisodes := CheckIfSeriesEnhanced(anime)
			if series {
				log.Printf("Switched to series: %s with %d episodes.\n", anime.Name, totalEpisodes)
				HandleSeries(anime, episodes, totalEpisodes, discordEnabled)
				break
			}
			// Otherwise continue loop to play the new movie
			fmt.Printf("Switched to movie: %s\n", anime.Name)
			continue
		}

		episodeDuration := time.Duration(episodes[0].Duration) * time.Second
		updater := createUpdater(anime, &isPaused, &animeMutex, episodeDuration, discordEnabled)

		err = player.HandleDownloadAndPlay(
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

		// Handle playback errors and user interaction
		if errors.Is(err, player.ErrUserQuit) {
			log.Println("Quitting application as per user request.")
			break
		}

		// Check if user requested to change anime during video playback
		if errors.Is(err, player.ErrChangeAnime) {
			newAnime, newEpisodes, err := ChangeAnimeLocal()
			if err != nil {
				log.Printf("Error changing anime: %v", err)
				continue // Stay with current anime if change fails
			}

			// Update anime and episodes
			anime = newAnime
			episodes = newEpisodes

			// Check if new anime is a series
			series, totalEpisodes := CheckIfSeriesEnhanced(anime)
			if series {
				// If new anime is a series, switch to series handler
				log.Printf("Switched to series: %s with %d episodes.\n", anime.Name, totalEpisodes)
				HandleSeries(anime, episodes, totalEpisodes, discordEnabled)
				break
			}

			fmt.Printf("Switched to movie: %s\n", anime.Name)
			continue // Continue with new movie
		}

		if err != nil {
			log.Printf("Error during movie playback: %v", err)
		}

		// Ask user what to do next after movie finishes
		userInput := GetUserInput()
		if userInput == "q" {
			log.Println("Quitting application as per user request.")
			break
		}

		// Handle anime change for movies
		if userInput == "c" {
			newAnime, newEpisodes, err := ChangeAnimeLocal()
			if err != nil {
				log.Printf("Error changing anime: %v", err)
				continue // Stay with current anime if change fails
			}

			// Update anime and episodes
			anime = newAnime
			episodes = newEpisodes

			// Check if new anime is a series
			series, totalEpisodes := CheckIfSeriesEnhanced(anime)
			if series {
				// If new anime is a series, switch to series handler
				log.Printf("Switched to series: %s with %d episodes.\n", anime.Name, totalEpisodes)
				HandleSeries(anime, episodes, totalEpisodes, discordEnabled)
				break
			}

			fmt.Printf("Switched to movie: %s\n", anime.Name)
			continue // Continue with new movie
		}

		// For movies, other navigation options don't make much sense, so just continue playing the same movie
		log.Println("Replaying the same movie...")
	}
}

// createUpdater cria um atualizador de Discord Rich Presence se estiver habilitado
func createUpdater(anime *models.Anime, isPaused *bool, animeMutex *sync.Mutex, episodeDuration time.Duration, discordEnabled bool) *discord.RichPresenceUpdater {
	if !discordEnabled {
		return nil
	}
	return discord.NewRichPresenceUpdater(
		anime,
		isPaused,
		animeMutex,
		1*time.Second,
		episodeDuration,
		getSocketPath(),
		player.MpvSendCommand,
	)
}

// getSocketPath retorna o caminho do socket MPV baseado no sistema operacional
func getSocketPath() string {
	if runtime.GOOS == "windows" {
		return `\\.\pipe\goanime_mpvsocket`
	}
	// Use os.TempDir() for macOS compatibility
	return filepath.Join(os.TempDir(), "mpvsocket")
}
