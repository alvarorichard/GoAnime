package discord

import (
	"fmt"
	"log"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/rich-go/client"
)

// RichPresenceUpdater gerencia as atualizações do Discord Rich Presence
type RichPresenceUpdater struct {
	anime           *models.Anime
	isPaused        *bool
	animeMutex      *sync.Mutex
	updateFreq      time.Duration
	done            chan bool
	wg              sync.WaitGroup
	startTime       time.Time                                        // Start time of playback
	episodeDuration time.Duration                                    // Total duration of the episode
	episodeStarted  bool                                             // Whether the episode has started
	socketPath      string                                           // Path to mpv IPC socket
	mpvSendCommand  func(string, []interface{}) (interface{}, error) // Função para enviar comandos ao MPV
}

// NewRichPresenceUpdater cria uma nova instância do atualizador de Rich Presence
func NewRichPresenceUpdater(
	anime *models.Anime,
	isPaused *bool,
	animeMutex *sync.Mutex,
	updateFreq time.Duration,
	episodeDuration time.Duration,
	socketPath string,
	mpvSendCommand func(string, []interface{}) (interface{}, error),
) *RichPresenceUpdater {
	return &RichPresenceUpdater{
		anime:           anime,
		isPaused:        isPaused,
		animeMutex:      animeMutex,
		updateFreq:      updateFreq,
		done:            make(chan bool),
		startTime:       time.Now(),
		episodeDuration: episodeDuration,
		episodeStarted:  false,
		socketPath:      socketPath,
		mpvSendCommand:  mpvSendCommand,
	}
}

// GetCurrentPlaybackPosition obtém a posição atual de reprodução do MPV
func (rpu *RichPresenceUpdater) GetCurrentPlaybackPosition() (time.Duration, error) {
	position, err := rpu.mpvSendCommand(rpu.socketPath, []interface{}{"get_property", "time-pos"})
	if err != nil {
		return 0, err
	}

	// Convert position to float64 and then to time.Duration
	posSeconds, ok := position.(float64)
	if !ok {
		return 0, fmt.Errorf("failed to parse playback position")
	}

	return time.Duration(posSeconds) * time.Second, nil
}

// SetSocketPath sets the MPV socket path
func (rpu *RichPresenceUpdater) SetSocketPath(socketPath string) {
	rpu.socketPath = socketPath
}

// GetSocketPath returns the MPV socket path
func (rpu *RichPresenceUpdater) GetSocketPath() string {
	return rpu.socketPath
}

// SetEpisodeStarted sets whether the episode has started
func (rpu *RichPresenceUpdater) SetEpisodeStarted(started bool) {
	rpu.episodeStarted = started
}

// IsEpisodeStarted returns whether the episode has started
func (rpu *RichPresenceUpdater) IsEpisodeStarted() bool {
	return rpu.episodeStarted
}

// SetEpisodeDuration sets the episode duration
func (rpu *RichPresenceUpdater) SetEpisodeDuration(duration time.Duration) {
	rpu.episodeDuration = duration
}

// GetEpisodeDuration returns the episode duration
func (rpu *RichPresenceUpdater) GetEpisodeDuration() time.Duration {
	return rpu.episodeDuration
}

// GetAnime returns the associated anime
func (rpu *RichPresenceUpdater) GetAnime() *models.Anime {
	return rpu.anime
}

// GetIsPaused returns the pointer to the pause state
func (rpu *RichPresenceUpdater) GetIsPaused() *bool {
	return rpu.isPaused
}

// GetAnimeMutex returns the anime mutex
func (rpu *RichPresenceUpdater) GetAnimeMutex() *sync.Mutex {
	return rpu.animeMutex
}

// GetUpdateFreq returns the update frequency
func (rpu *RichPresenceUpdater) GetUpdateFreq() time.Duration {
	return rpu.updateFreq
}

// Start inicia as atualizações periódicas do Rich Presence
func (rpu *RichPresenceUpdater) Start() {
	rpu.wg.Add(1)
	go func() {
		defer rpu.wg.Done()
		ticker := time.NewTicker(rpu.updateFreq)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				go rpu.updateDiscordPresence() // Run update asynchronously
			case <-rpu.done:
				if util.IsDebug {
					log.Println("Rich Presence updater received stop signal.")
				}
				return
			}
		}
	}()
	if util.IsDebug {
		log.Println("Rich Presence updater started.")
	}
}

// Stop signals the updater to stop and waits for the goroutine to finish
func (rpu *RichPresenceUpdater) Stop() {
	// Evita fechar o canal múltiplas vezes
	if rpu != nil {
		select {
		case <-rpu.done:
			// Canal já fechado
		default:
			close(rpu.done)
		}
		rpu.wg.Wait()
		if util.IsDebug {
			log.Println("Rich Presence updater stopped.")
		}
	}
}

// updateDiscordPresence atualiza o status do Discord Rich Presence
func (rpu *RichPresenceUpdater) updateDiscordPresence() {
	rpu.animeMutex.Lock()
	defer rpu.animeMutex.Unlock()

	currentPosition, err := rpu.GetCurrentPlaybackPosition()
	if err != nil {
		if util.IsDebug {
			log.Printf("Error fetching playback position: %v\n", err)
		}
		return
	}

	// Debug log to check episode duration
	if util.IsDebug {
		log.Printf("Episode Duration in updateDiscordPresence: %v seconds (%v minutes)\n", rpu.episodeDuration.Seconds(), rpu.episodeDuration.Minutes())
	}

	// Convert episode duration to minutes and seconds format
	totalMinutes := int(rpu.episodeDuration.Minutes())
	totalSeconds := int(rpu.episodeDuration.Seconds()) % 60 // Remaining seconds after full minutes

	// Format the current playback position as minutes and seconds
	timeInfo := fmt.Sprintf("%02d:%02d / %02d:%02d",
		int(currentPosition.Minutes()), int(currentPosition.Seconds())%60,
		totalMinutes, totalSeconds,
	)

	// Create the activity with updated Details
	activity := client.Activity{
		Details:    fmt.Sprintf("%s | Episode %s | %s / %d min", rpu.anime.Details.Title.Romaji, rpu.anime.Episodes[0].Number, timeInfo, totalMinutes),
		State:      "Watching",
		LargeImage: rpu.anime.ImageURL,
		LargeText:  rpu.anime.Details.Title.Romaji,
		Buttons: []*client.Button{
			{Label: "View on AniList", Url: fmt.Sprintf("https://anilist.co/anime/%d", rpu.anime.AnilistID)},
			{Label: "View on MAL", Url: fmt.Sprintf("https://myanimelist.net/anime/%d", rpu.anime.MalID)},
		},
	}

	// Set the activity in Discord Rich Presence
	if err := client.SetActivity(activity); err != nil {
		if util.IsDebug {
			log.Printf("Error updating Discord Rich Presence: %v\n", err)
		}
	} else {
		if util.IsDebug {
			log.Printf("Discord Rich Presence updated with elapsed time: %s\n", timeInfo)
		}
	}
}

// FetchDuration fetches the episode duration (future feature)
func (rpu *RichPresenceUpdater) FetchDuration(socketPath string, f func(durSec int)) {
	// TODO: Implement episode duration fetching
	panic("unimplemented")
}

// WaitEpisodeStart waits for episode start (future feature)
func (rpu *RichPresenceUpdater) WaitEpisodeStart() {
	// TODO: Implement waiting for episode start
	panic("unimplemented")
}
