package discord

import (
	"fmt"
	"strings"
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
	// Use TryLock to avoid blocking playback if mutex is held
	if !rpu.animeMutex.TryLock() {
		util.Debug("Discord update skipped - mutex busy")
		return
	}
	defer rpu.animeMutex.Unlock()

	currentPosition, err := rpu.GetCurrentPlaybackPosition()
	if err != nil {
		// Don't spam logs on temporary connection issues
		return
	}

	// Only fetch duration once if not set
	if rpu.episodeDuration == 0 {
		durationResponse, err := rpu.mpvSendCommand(rpu.socketPath, []interface{}{"get_property", "duration"})
		if err == nil && durationResponse != nil {
			if durationSeconds, ok := durationResponse.(float64); ok && durationSeconds > 0 {
				rpu.episodeDuration = time.Duration(durationSeconds) * time.Second
			}
		}
	}

	// Format time as HH:MM:SS for movies/long content (>= 60 min) or MM:SS for episodes
	timeInfo := formatPlaybackTime(currentPosition, rpu.episodeDuration)

	// Detect if this is a movie/TV show (FlixHQ content) or anime
	isMovieOrTV := rpu.anime.IsMovieOrTV() || rpu.anime.Source == "FlixHQ"

	// Get title - prefer romaji for anime, use Name for movies/TV
	var title string
	if isMovieOrTV {
		// For movies/TV, use the Name directly (it comes from FlixHQ/TMDB)
		title = rpu.anime.Name
	} else {
		// For anime, prefer romaji, fall back to english, then name
		title = rpu.anime.Details.Title.Romaji
		if title == "" {
			title = rpu.anime.Details.Title.English
		}
		if title == "" {
			// Clean the anime name from source tags for display
			title = rpu.anime.Name
			// Remove language/source tags like [English], [Portuguese], [AnimeFire] etc.
			if idx := strings.Index(title, "]"); idx != -1 && idx < 20 {
				title = strings.TrimSpace(title[idx+1:])
			}
		}
	}

	// Get image URL with fallback - Discord requires an externally accessible URL
	imageURL := rpu.anime.ImageURL
	if imageURL == "" {
		// Use a default placeholder image
		imageURL = "https://raw.githubusercontent.com/alvarorichard/Goanime/main/docs/assets/goanime-logo.png"
	}

	// Build activity details based on content type
	totalDurationDisplay := formatDurationDisplay(rpu.episodeDuration)
	var details string
	var state string

	if isMovieOrTV {
		// For movies/TV, don't show "Episode X"
		if rpu.anime.IsMovie() || rpu.anime.MediaType == "movie" {
			details = fmt.Sprintf("%s | %s / %s", title, timeInfo, totalDurationDisplay)
			state = "Watching a movie"
		} else {
			// For TV shows, show season/episode info if available
			episodeNumber := "1"
			if len(rpu.anime.Episodes) > 0 && rpu.anime.Episodes[0].Number != "" {
				episodeNumber = rpu.anime.Episodes[0].Number
			}
			details = fmt.Sprintf("%s | Ep %s | %s / %s", title, episodeNumber, timeInfo, totalDurationDisplay)
			state = "Watching a TV show"
		}
	} else {
		// For anime, show episode number
		episodeNumber := "1"
		if len(rpu.anime.Episodes) > 0 && rpu.anime.Episodes[0].Number != "" {
			episodeNumber = rpu.anime.Episodes[0].Number
		}
		details = fmt.Sprintf("%s | Episode %s | %s / %s", title, episodeNumber, timeInfo, totalDurationDisplay)
		state = "Watching anime"
	}

	// Create the activity with updated Details
	activity := client.Activity{
		Details:    details,
		State:      state,
		LargeImage: imageURL,
		LargeText:  title,
	}

	// Add buttons based on content type
	var buttons []*client.Button
	if isMovieOrTV {
		// For movies/TV from FlixHQ, add IMDB button if available
		if rpu.anime.IMDBID != "" {
			buttons = append(buttons, &client.Button{Label: "View on IMDB", Url: fmt.Sprintf("https://www.imdb.com/title/%s", rpu.anime.IMDBID)})
		}
		if rpu.anime.TMDBID > 0 {
			mediaType := "movie"
			if rpu.anime.IsTV() || rpu.anime.MediaType == "tv" {
				mediaType = "tv"
			}
			buttons = append(buttons, &client.Button{Label: "View on TMDB", Url: fmt.Sprintf("https://www.themoviedb.org/%s/%d", mediaType, rpu.anime.TMDBID)})
		}
	} else {
		// For anime, add AniList and MAL buttons
		if rpu.anime.AnilistID > 0 {
			buttons = append(buttons, &client.Button{Label: "View on AniList", Url: fmt.Sprintf("https://anilist.co/anime/%d", rpu.anime.AnilistID)})
		}
		if rpu.anime.MalID > 0 {
			buttons = append(buttons, &client.Button{Label: "View on MAL", Url: fmt.Sprintf("https://myanimelist.net/anime/%d", rpu.anime.MalID)})
		}
	}
	if len(buttons) > 0 {
		activity.Buttons = buttons
	}

	// Set the activity in Discord Rich Presence (silently ignore errors to reduce log spam)
	_ = client.SetActivity(activity)
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

// formatPlaybackTime formats the current position and total duration
// Uses HH:MM:SS format for content >= 60 minutes (movies), MM:SS for shorter content (episodes)
func formatPlaybackTime(currentPosition, totalDuration time.Duration) string {
	// Use hours format if total duration is >= 60 minutes (movies/long content)
	useHoursFormat := totalDuration.Minutes() >= 60

	if useHoursFormat {
		// Format as HH:MM:SS / HH:MM:SS for movies
		currentHours := int(currentPosition.Hours())
		currentMinutes := int(currentPosition.Minutes()) % 60
		currentSeconds := int(currentPosition.Seconds()) % 60

		totalHours := int(totalDuration.Hours())
		totalMinutes := int(totalDuration.Minutes()) % 60
		totalSeconds := int(totalDuration.Seconds()) % 60

		return fmt.Sprintf("%d:%02d:%02d / %d:%02d:%02d",
			currentHours, currentMinutes, currentSeconds,
			totalHours, totalMinutes, totalSeconds,
		)
	}

	// Format as MM:SS / MM:SS for episodes
	return fmt.Sprintf("%02d:%02d / %02d:%02d",
		int(currentPosition.Minutes()), int(currentPosition.Seconds())%60,
		int(totalDuration.Minutes()), int(totalDuration.Seconds())%60,
	)
}

// formatDurationDisplay returns a human-readable duration string
// Shows hours and minutes for long content, just minutes for short content
func formatDurationDisplay(duration time.Duration) string {
	totalMinutes := int(duration.Minutes())
	if totalMinutes >= 60 {
		hours := totalMinutes / 60
		minutes := totalMinutes % 60
		if minutes > 0 {
			return fmt.Sprintf("%dh%02dm", hours, minutes)
		}
		return fmt.Sprintf("%dh", hours)
	}
	return fmt.Sprintf("%d min", totalMinutes)
}
