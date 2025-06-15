package discord

import (
	"fmt"
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
				util.Debug("Rich Presence updater received stop signal.")
				return
			}
		}
	}()
	util.Debug("Rich Presence updater started.")
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
		util.Debug("Rich Presence updater stopped.")
	}
}

// updateDiscordPresence atualiza o status do Discord Rich Presence
func (rpu *RichPresenceUpdater) updateDiscordPresence() {
	rpu.animeMutex.Lock()
	defer rpu.animeMutex.Unlock()

	currentPosition, err := rpu.GetCurrentPlaybackPosition()
	if err != nil {
		util.Debugf("Error fetching playback position: %v", err)
		return
	}

	// Se a duração do episódio não estiver definida ou for 0, buscar do MPV
	if rpu.episodeDuration == 0 {
		durationResponse, err := rpu.mpvSendCommand(rpu.socketPath, []interface{}{"get_property", "duration"})
		if err == nil && durationResponse != nil {
			if durationSeconds, ok := durationResponse.(float64); ok && durationSeconds > 0 {
				rpu.episodeDuration = time.Duration(durationSeconds) * time.Second
			}
		}
	}

	// Debug log to check episode duration
	util.Debugf("Episode Duration in updateDiscordPresence: %v seconds (%v minutes)", rpu.episodeDuration.Seconds(), rpu.episodeDuration.Minutes())

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
		util.Debugf("Error updating Discord Rich Presence: %v", err)
	} else {
		util.Debugf("Discord Rich Presence updated with elapsed time: %s", timeInfo)
	}
}

// FetchDuration fetches the episode duration from MPV and calls the callback with duration in seconds
func (rpu *RichPresenceUpdater) FetchDuration(socketPath string, f func(durSec int)) {
	// Use the provided socketPath or fall back to the instance's socketPath
	path := socketPath
	if path == "" {
		path = rpu.socketPath
	}

	// Send command to MPV to get the duration property
	durationResponse, err := rpu.mpvSendCommand(path, []interface{}{"get_property", "duration"})
	if err != nil {
		util.Debugf("Error fetching duration from MPV: %v", err)
		return
	}

	// Check if we got a valid response
	if durationResponse == nil {
		util.Debug("Duration property not available from MPV")
		return
	}

	// Convert the response to float64 (MPV returns duration in seconds as a float)
	durationSeconds, ok := durationResponse.(float64)
	if !ok {
		util.Debugf("Failed to parse duration response: %v", durationResponse)
		return
	}

	// Convert to int seconds and call the callback
	durSec := int(durationSeconds)
	if durSec > 0 {
		f(durSec)
	} else {
		util.Debug("Duration is zero or negative, skipping callback")
	}
}

// WaitEpisodeStart waits for episode start (future feature)
func (rpu *RichPresenceUpdater) WaitEpisodeStart() {
	// TODO: Implement waiting for episode start
	panic("unimplemented")
}
