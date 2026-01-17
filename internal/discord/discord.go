package discord

import (
	"fmt"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/tr1xem/go-discordrpc/client"
)

// DiscordClientID is the Discord application client ID
const DiscordClientID = "1302721937717334128"

// Global state for smart updates
var (
	discordClient       *client.Client
	isLoggedIn          bool
	lastPausedState     bool
	lastEpisodeNumber   string
	lastTitle           string
	lastUpdateTime      time.Time
	lastForceUpdateTime time.Time
	clientMutex         sync.Mutex
)

// RichPresenceUpdater manages Discord Rich Presence updates
type RichPresenceUpdater struct {
	anime           *models.Anime
	isPaused        *bool
	animeMutex      *sync.Mutex
	updateFreq      time.Duration
	done            chan bool
	wg              sync.WaitGroup
	startTime       time.Time
	episodeDuration time.Duration
	episodeStarted  bool
	socketPath      string
	mpvSendCommand  func(string, []interface{}) (interface{}, error)
}

// NewRichPresenceUpdater creates a new Rich Presence updater instance
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

// LoginClient logs into Discord RPC
func LoginClient() error {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	if discordClient != nil && isLoggedIn {
		return nil // Already logged in
	}

	discordClient = client.NewClient(DiscordClientID)

	if err := discordClient.Login(); err != nil {
		return fmt.Errorf("discord login failed: %w", err)
	}

	isLoggedIn = true
	util.Debug("Discord RPC logged in successfully")
	return nil
}

// LogoutClient logs out from Discord RPC
func LogoutClient() error {
	clientMutex.Lock()
	defer clientMutex.Unlock()

	if discordClient != nil && isLoggedIn {
		if err := discordClient.Logout(); err != nil {
			return fmt.Errorf("discord logout failed: %w", err)
		}
		isLoggedIn = false
		discordClient = nil
		util.Debug("Discord RPC logged out")
	}
	return nil
}

// IsClientLoggedIn returns whether the Discord client is logged in
func IsClientLoggedIn() bool {
	clientMutex.Lock()
	defer clientMutex.Unlock()
	return isLoggedIn
}

// GetCurrentPlaybackPosition gets the current playback position from MPV
func (rpu *RichPresenceUpdater) GetCurrentPlaybackPosition() (time.Duration, error) {
	position, err := rpu.mpvSendCommand(rpu.socketPath, []interface{}{"get_property", "time-pos"})
	if err != nil {
		return 0, err
	}

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

// Start begins periodic Rich Presence updates
func (rpu *RichPresenceUpdater) Start() {
	// Ensure client is logged in
	if err := LoginClient(); err != nil {
		util.Debugf("Failed to login Discord client: %v", err)
		return
	}

	rpu.wg.Add(1)
	go func() {
		defer rpu.wg.Done()
		ticker := time.NewTicker(rpu.updateFreq)
		defer ticker.Stop()

		// Initial update
		rpu.updateDiscordPresence(false)

		for {
			select {
			case <-ticker.C:
				rpu.updateDiscordPresence(false)
			case <-rpu.done:
				util.Debug("Rich Presence updater received stop signal")
				return
			}
		}
	}()
	util.Debug("Rich Presence updater started")
}

// Stop signals the updater to stop and waits for the goroutine to finish
func (rpu *RichPresenceUpdater) Stop() {
	if rpu != nil {
		select {
		case <-rpu.done:
			// Channel already closed
		default:
			close(rpu.done)
		}
		rpu.wg.Wait()
		util.Debug("Rich Presence updater stopped")
	}
}

// updateDiscordPresence updates the Discord Rich Presence status
func (rpu *RichPresenceUpdater) updateDiscordPresence(forceUpdate bool) {
	// Use TryLock to avoid blocking playback
	if !rpu.animeMutex.TryLock() {
		return
	}
	defer rpu.animeMutex.Unlock()

	// Get current playback position
	currentPosition, err := rpu.GetCurrentPlaybackPosition()
	if err != nil {
		return
	}

	currentPositionSec := int(currentPosition.Seconds())

	// Fetch duration if not set
	if rpu.episodeDuration == 0 {
		durationResponse, err := rpu.mpvSendCommand(rpu.socketPath, []interface{}{"get_property", "duration"})
		if err == nil && durationResponse != nil {
			if durationSeconds, ok := durationResponse.(float64); ok && durationSeconds > 0 {
				rpu.episodeDuration = time.Duration(durationSeconds) * time.Second
			}
		}
	}

	totalDurationSec := int(rpu.episodeDuration.Seconds())

	// Check pause state
	pauseResponse, _ := rpu.mpvSendCommand(rpu.socketPath, []interface{}{"get_property", "pause"})
	isPaused := false
	if pause, ok := pauseResponse.(bool); ok {
		isPaused = pause
	}

	// Detect content type
	isMovieOrTV := rpu.anime.IsMovieOrTV() || rpu.anime.Source == "FlixHQ"

	// Get title
	title := rpu.getTitle(isMovieOrTV)

	// Get episode number
	episodeNumber := "1"
	if len(rpu.anime.Episodes) > 0 && rpu.anime.Episodes[0].Number != "" {
		episodeNumber = rpu.anime.Episodes[0].Number
	}

	// Smart update check - only update if something changed or forced
	now := time.Now()
	shouldUpdate := false

	// Force update every 2 minutes to keep presence alive
	if lastForceUpdateTime.IsZero() || time.Since(lastForceUpdateTime) >= 2*time.Minute {
		shouldUpdate = true
		lastForceUpdateTime = now
	}

	// Update if state changed
	if lastUpdateTime.IsZero() ||
		lastPausedState != isPaused ||
		lastEpisodeNumber != episodeNumber ||
		lastTitle != title ||
		forceUpdate {
		shouldUpdate = true
	}

	if !shouldUpdate {
		return
	}

	// Get image URL
	imageURL := rpu.anime.ImageURL
	if imageURL == "" {
		imageURL = "https://raw.githubusercontent.com/alvarorichard/Goanime/main/docs/assets/goanime-logo.png"
	}

	// Build timestamps
	var timestamps *client.Timestamps
	var smallImage, smallText string

	startTime := now.Add(-time.Duration(currentPositionSec) * time.Second)

	if isPaused {
		timestamps = &client.Timestamps{
			Start: &startTime,
			End:   nil,
		}
		smallImage = "pause-button"
		smallText = "Paused"
	} else {
		if totalDurationSec > 60 && totalDurationSec > currentPositionSec {
			remainingSeconds := totalDurationSec - currentPositionSec
			endTime := now.Add(time.Duration(remainingSeconds) * time.Second)
			timestamps = &client.Timestamps{
				Start: &startTime,
				End:   &endTime,
			}
		} else {
			timestamps = &client.Timestamps{
				Start: &startTime,
				End:   nil,
			}
		}
		smallImage = ""
		smallText = ""
	}

	// Build state text
	var state string
	if isMovieOrTV {
		if rpu.anime.IsMovie() || rpu.anime.MediaType == "movie" {
			state = "Watching a movie"
		} else {
			state = fmt.Sprintf("Episode %s", episodeNumber)
		}
	} else {
		state = fmt.Sprintf("Episode %s", episodeNumber)
	}

	// Build buttons
	buttons := rpu.buildButtons(isMovieOrTV)

	// Create and set activity
	activity := client.Activity{
		Type:       3, // Watching
		Name:       title,
		Details:    title,
		State:      state,
		LargeImage: imageURL,
		LargeText:  title,
		SmallImage: smallImage,
		SmallText:  smallText,
		Timestamps: timestamps,
		Buttons:    buttons,
	}

	clientMutex.Lock()
	if discordClient != nil && isLoggedIn {
		_ = discordClient.SetActivity(activity)
	}
	clientMutex.Unlock()

	// Update last state
	lastPausedState = isPaused
	lastEpisodeNumber = episodeNumber
	lastTitle = title
	lastUpdateTime = now
}

// getTitle extracts the appropriate title based on content type
func (rpu *RichPresenceUpdater) getTitle(isMovieOrTV bool) string {
	if isMovieOrTV {
		return rpu.anime.Name
	}

	title := rpu.anime.Details.Title.Romaji
	if title == "" {
		title = rpu.anime.Details.Title.English
	}
	if title == "" {
		title = rpu.anime.Name
		if idx := strings.Index(title, "]"); idx != -1 && idx < 20 {
			title = strings.TrimSpace(title[idx+1:])
		}
	}
	return title
}

// buildButtons creates the appropriate buttons based on content type
func (rpu *RichPresenceUpdater) buildButtons(isMovieOrTV bool) []*client.Button {
	var buttons []*client.Button

	if isMovieOrTV {
		if rpu.anime.IMDBID != "" {
			buttons = append(buttons, &client.Button{
				Label: "View on IMDB",
				Url:   fmt.Sprintf("https://www.imdb.com/title/%s", rpu.anime.IMDBID),
			})
		}
		if rpu.anime.TMDBID > 0 {
			mediaType := "movie"
			if rpu.anime.IsTV() || rpu.anime.MediaType == "tv" {
				mediaType = "tv"
			}
			buttons = append(buttons, &client.Button{
				Label: "View on TMDB",
				Url:   fmt.Sprintf("https://www.themoviedb.org/%s/%d", mediaType, rpu.anime.TMDBID),
			})
		}
	} else {
		if rpu.anime.AnilistID > 0 {
			buttons = append(buttons, &client.Button{
				Label: "View on AniList",
				Url:   fmt.Sprintf("https://anilist.co/anime/%d", rpu.anime.AnilistID),
			})
		}
		if rpu.anime.MalID > 0 {
			buttons = append(buttons, &client.Button{
				Label: "View on MAL",
				Url:   fmt.Sprintf("https://myanimelist.net/anime/%d", rpu.anime.MalID),
			})
		}
	}

	// Discord allows max 2 buttons
	if len(buttons) > 2 {
		buttons = buttons[:2]
	}

	return buttons
}

// FetchDuration fetches the episode duration from MPV
func (rpu *RichPresenceUpdater) FetchDuration(socketPath string, f func(durSec int)) {
	path := socketPath
	if path == "" {
		path = rpu.socketPath
	}

	durationResponse, err := rpu.mpvSendCommand(path, []interface{}{"get_property", "duration"})
	if err != nil {
		return
	}

	if durationResponse == nil {
		return
	}

	durationSeconds, ok := durationResponse.(float64)
	if !ok {
		return
	}

	durSec := int(durationSeconds)
	if durSec > 0 {
		f(durSec)
	}
}

// FormatTime formats seconds into human-readable time
func FormatTime(seconds int) string {
	hours := seconds / 3600
	minutes := (seconds % 3600) / 60
	remainingSeconds := seconds % 60

	if hours > 0 {
		return fmt.Sprintf("%d:%02d:%02d", hours, minutes, remainingSeconds)
	}
	return fmt.Sprintf("%d:%02d", minutes, remainingSeconds)
}
