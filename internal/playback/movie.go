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
func HandleMovie(anime *models.Anime, episodes []models.Episode, discordEnabled bool) error {
	for {
		animeMutex := sync.Mutex{}
		isPaused := false

		animeMutex.Lock()
		anime.Episodes = []models.Episode{episodes[0]}
		animeMutex.Unlock()

		// Only fetch movie data from Jikan API for anime content (not FlixHQ movies/TV)
		// FlixHQ content already has metadata from TMDB/OMDb
		if !anime.IsMovieOrTV() && anime.MalID > 0 {
			if err := api.GetMovieData(anime.MalID, anime); err != nil {
				log.Printf("Error fetching movie/OVA data: %v", err)
			}
		}

		videoURL, err := player.GetVideoURLForEpisodeEnhanced(&episodes[0], anime)
		if err != nil {
			log.Printf("Failed to extract video URL: %v", util.ErrorHandler(err))
			// Return to anime selection
			return player.ErrBackToAnimeSelection
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
				if err := HandleSeries(anime, episodes, totalEpisodes, discordEnabled); err != nil {
					if errors.Is(err, player.ErrBackToAnimeSelection) {
						return err
					}
				}
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

		// Handle back/change anime for movies - both options allow searching for a new anime
		if userInput == "c" || userInput == "back" {
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
				if err := HandleSeries(anime, episodes, totalEpisodes, discordEnabled); err != nil {
					if errors.Is(err, player.ErrBackToAnimeSelection) {
						return err
					}
				}
				break
			}

			fmt.Printf("Switched to movie: %s\n", anime.Name)
			continue // Continue with new movie
		}

		// For movies, other navigation options don't make much sense, so just continue playing the same movie
		log.Println("Replaying the same movie...")
	}
	return nil
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
