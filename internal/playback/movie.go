package playback

import (
	"log"
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
	animeMutex := sync.Mutex{}
	isPaused := false

	animeMutex.Lock()
	anime.Episodes = []models.Episode{episodes[0]}
	animeMutex.Unlock()

	if err := api.GetMovieData(anime.MalID, anime); err != nil {
		log.Printf("Error fetching movie/OVA data: %v", err)
	}

	videoURL, err := player.GetVideoURLForEpisode(episodes[0].URL)
	if err != nil {
		log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
	}

	episodeDuration := time.Duration(episodes[0].Duration) * time.Second
	updater := createUpdater(anime, &isPaused, &animeMutex, episodeDuration, discordEnabled)

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
	return "/tmp/mpvsocket"
}
