package player

import (
	"errors"
	"fmt"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/tracking"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/charmbracelet/huh"
)

// ErrUserQuit is returned when the user chooses to quit the application
var ErrUserQuit = errors.New("user requested to quit application")

// ErrChangeAnime is returned when the user chooses to change anime
var ErrChangeAnime = errors.New("user requested to change anime")

// ErrBackToDownloadOptions is returned when user wants to go back to download options
var ErrBackToDownloadOptions = errors.New("back to download options")

// waitForVideoReady waits for the HLS video to be ready for playback
// Returns true if video is ready, false if timeout
func waitForVideoReady(socketPath string) bool {
	util.Debugf("Waiting for HLS video to be ready...")

	maxWait := 45 * time.Second // Increased for slow HLS streams
	pollInterval := 500 * time.Millisecond
	startTime := time.Now()

	for time.Since(startTime) < maxWait {
		// Try multiple properties to detect when video is ready
		// Method 1: Check duration (most reliable for HLS)
		durationResp, err := mpvSendCommand(socketPath, []interface{}{"get_property", "duration"})
		if err == nil {
			if duration, ok := durationResp.(float64); ok && duration > 0 {
				util.Debugf("HLS video ready (duration: %.0f seconds) after %.1fs", duration, time.Since(startTime).Seconds())
				return true
			}
		}

		// Method 2: Check if time-pos is available (video playing)
		posResp, posErr := mpvSendCommand(socketPath, []interface{}{"get_property", "time-pos"})
		if posErr == nil {
			if pos, ok := posResp.(float64); ok && pos >= 0 {
				util.Debugf("HLS video playing (position: %.1f seconds) after %.1fs", pos, time.Since(startTime).Seconds())
				return true
			}
		}

		// Method 3: Check playback-time (alternative property)
		playbackResp, playErr := mpvSendCommand(socketPath, []interface{}{"get_property", "playback-time"})
		if playErr == nil {
			if playback, ok := playbackResp.(float64); ok && playback >= 0 {
				util.Debugf("HLS video playback started (time: %.1f seconds) after %.1fs", playback, time.Since(startTime).Seconds())
				return true
			}
		}

		time.Sleep(pollInterval)
	}
	util.Debugf("Timeout waiting for HLS video (%.1fs)", time.Since(startTime).Seconds())
	return false
}

// seekToResumePosition seeks to the saved resume position for HLS streams
// This function is robust: it waits for video ready, seeks, and verifies the seek worked
func seekToResumePosition(socketPath string, resumeTime int) {
	if resumeTime <= 0 {
		return
	}

	util.Debugf("HLS resume: will seek to %d seconds", resumeTime)

	// Wait for video to be fully ready first
	if !waitForVideoReady(socketPath) {
		util.Debugf("HLS resume: video not ready after timeout, attempting seek anyway")
	}

	// Give mpv a moment to stabilize after video is ready
	time.Sleep(300 * time.Millisecond)

	// Try multiple seek methods with verification - more attempts, longer waits
	maxAttempts := 5
	for attempt := 1; attempt <= maxAttempts; attempt++ {
		util.Debugf("HLS resume: seek attempt %d/%d to %d seconds", attempt, maxAttempts, resumeTime)

		// Method 1: seek absolute (works best with HLS)
		_, err := mpvSendCommand(socketPath, []interface{}{"seek", float64(resumeTime), "absolute"})
		if err != nil {
			util.Debugf("HLS resume: seek absolute failed: %v, trying set_property", err)
			// Method 2: set_property time-pos
			_, _ = mpvSendCommand(socketPath, []interface{}{"set_property", "time-pos", float64(resumeTime)})
		}

		// Wait for seek to complete (HLS can be slow)
		time.Sleep(800 * time.Millisecond)

		// Verify position
		posResp, posErr := mpvSendCommand(socketPath, []interface{}{"get_property", "time-pos"})
		if posErr == nil {
			if pos, ok := posResp.(float64); ok {
				// Allow 10 second tolerance for HLS streams
				if pos >= float64(resumeTime-10) {
					util.Debugf("HLS resume: SUCCESS - position is %.0f seconds (target: %d)", pos, resumeTime)
					return
				}
				util.Debugf("HLS resume: position mismatch - got %.0f, want %d", pos, resumeTime)
			}
		} else {
			util.Debugf("HLS resume: could not get position: %v", posErr)
		}

		// Wait before retry, increasing delay
		if attempt < maxAttempts {
			waitTime := time.Duration(attempt) * 500 * time.Millisecond
			util.Debugf("HLS resume: waiting %v before retry", waitTime)
			time.Sleep(waitTime)
		}
	}

	util.Debugf("HLS resume: FAILED after %d attempts - video may start from beginning", maxAttempts)
}

// applySkipTimes applies skip times to an mpv instance
func applySkipTimes(socketPath string, episode *models.Episode) {
	var opts []string
	if episode.SkipTimes.Op.Start > 0 || episode.SkipTimes.Op.End > 0 {
		opts = append(opts, fmt.Sprintf("skip_op=%d-%d", episode.SkipTimes.Op.Start, episode.SkipTimes.Op.End))
	}
	if episode.SkipTimes.Ed.Start > 0 || episode.SkipTimes.Ed.End > 0 {
		opts = append(opts, fmt.Sprintf("skip_ed=%d-%d", episode.SkipTimes.Ed.Start, episode.SkipTimes.Ed.End))
	}

	if len(opts) > 0 {
		combinedOpts := strings.Join(opts, ",")
		_, cmdErr := mpvSendCommand(socketPath, []interface{}{"set_property", "script-opts", combinedOpts})
		if cmdErr != nil {
			util.Debugf("Failed to apply skip times: %v. Command: set_property script-opts %s", cmdErr, combinedOpts)
		} else {
			util.Debugf("Skip times applied successfully: %s", combinedOpts)
		}
	} else {
		util.Debugf("No skip times available for episode %s", episode.Number)
	}
}

// showResumeDialog displays a compact dialog asking if user wants to resume playback
func showResumeDialog(episodeNum int, timeSeconds int) (bool, error) {
	var resume bool

	// Convert seconds to minutes and seconds for better readability
	minutes := timeSeconds / 60
	seconds := timeSeconds % 60

	var timeStr string
	if minutes > 0 {
		timeStr = fmt.Sprintf("%dm %ds", minutes, seconds)
	} else {
		timeStr = fmt.Sprintf("%ds", seconds)
	}

	confirm := huh.NewConfirm().
		Title(fmt.Sprintf("Resume episode %d from %s?", episodeNum, timeStr)).
		Description("You can continue watching from where you left off.").
		Affirmative("Yes, resume").
		Negative("No, start from beginning").
		Value(&resume)

	if err := confirm.Run(); err != nil {
		return false, fmt.Errorf("error showing dialog: %w", err)
	}

	return resume, nil
}

// playVideo plays the video and manages interactions
// func playVideo(
// 	videoURL string,
// 	episodes []models.Episode,
// 	currentEpisodeNum int,
// 	anilistID int,
// 	updater *discord.RichPresenceUpdater,
// ) error {
// 	videoURL = strings.Replace(videoURL, "720pp.mp4", "720p.mp4", 1)
// 	util.Debugf("Video URL: %s", videoURL)

// 	currentEpisode, err := getCurrentEpisode(episodes, currentEpisodeNum)
// 	if err != nil {
// 		return fmt.Errorf("error getting current episode: %w", err)
// 	}

// 	mpvArgs := []string{
// 		"--hwdec=auto-safe",
// 		"--vo=gpu",
// 		"--profile=fast",
// 		"--cache=yes",
// 		"--demuxer-max-bytes=300M",
// 		"--demuxer-readahead-secs=20",
// 		"--no-config",
// 		"--video-latency-hacks=yes",
// 		"--audio-display=no",
// 	}

// 	tracker, resumeTime := initTracking(anilistID, currentEpisode, currentEpisodeNum)
// 	if resumeTime > 0 {
// 		mpvArgs = append(mpvArgs, fmt.Sprintf("--start=+%d", resumeTime))
// 	}

// 	skipDataChan := fetchAniSkipAsync(anilistID, currentEpisodeNum, currentEpisode)
// 	socketPath, err := StartVideo(videoURL, mpvArgs)
// 	if err != nil {
// 		return fmt.Errorf("failed to start video: %w", err)
// 	}

// 	applyAniSkipResults(skipDataChan, socketPath, currentEpisode, currentEpisodeNum)

// 	if updater != nil {
// 		initDiscordPresence(updater, socketPath, tracker, anilistID, currentEpisode, currentEpisodeNum)
// 		defer updater.Stop()
// 	}

// 	currentEpisodeIndex := findEpisodeIndex(episodes, currentEpisodeNum)
// 	if currentEpisodeIndex == -1 {
// 		return fmt.Errorf("episode %d not found in list", currentEpisodeNum)
// 	}

// 	preloadNextEpisode(episodes, currentEpisodeIndex)

// 	stopTracking := startTrackingRoutine(tracker, socketPath, anilistID, currentEpisode, currentEpisodeNum, updater)
// 	defer close(stopTracking)

// 	return handleUserInput(
// 		socketPath,
// 		episodes,
// 		currentEpisodeIndex,
// 		currentEpisodeNum,
// 		anilistID,
// 		updater,
// 		stopTracking,
// 		currentEpisode,
// 	)
// }

// playVideo plays the video and manages interactions
// playVideo plays the video and manages interactions
func playVideo(
	videoURL string,
	episodes []models.Episode,
	currentEpisodeNum int,
	anilistID int,
	updater *discord.RichPresenceUpdater,
) error {
	timer := util.StartTimer("playVideo:Total")
	defer timer.Stop()
	util.PerfCount("video_plays")

	// Log the episode number and URL for debugging
	util.Debugf("Playing video for episode %d, URL: %s", currentEpisodeNum, videoURL)

	// Normalize video URL if necessary
	videoURL = strings.Replace(videoURL, "720pp.mp4", "720p.mp4", 1)

	// Get the current episode
	currentEpisode, err := getCurrentEpisode(episodes, currentEpisodeNum)
	if err != nil {
		return fmt.Errorf("error getting current episode: %w", err)
	}

	// Set up mpv arguments for optimal playback
	mpvArgs := []string{
		"--hwdec=auto-safe",
		"--vo=gpu",
		"--profile=fast",
		"--cache=yes",
		"--demuxer-max-bytes=300M",
		"--demuxer-readahead-secs=20",
		"--no-config",
		"--video-latency-hacks=yes",
		"--audio-display=no",
	}

	// For HLS streams (.m3u8), we need to add HTTP headers for proper playback
	// Many streaming servers require specific User-Agent and Referer headers
	isHLSStream := strings.Contains(videoURL, ".m3u8") || strings.Contains(videoURL, "m3u8")
	if isHLSStream {
		// Add HTTP headers required by streaming servers
		// Note: Only Referer is typically needed; some servers also require User-Agent
		mpvArgs = append(mpvArgs,
			"--http-header-fields=Referer: https://streameeeeee.site/",
		)
		util.Debugf("HLS stream detected - adding HTTP Referer header")
	}

	// Only apply audio/subtitle language preferences for movies/TV (FlixHQ)
	// Check if this is a movie/TV content by examining the updater's anime source
	isMovieOrTV := false
	if updater != nil && updater.GetAnime() != nil {
		anime := updater.GetAnime()
		isMovieOrTV = anime.IsMovieOrTV() || strings.Contains(strings.ToLower(anime.Source), "flixhq")
	}

	if isMovieOrTV {
		// Audio and subtitle language preferences only for movies/TV
		audioLang := util.GlobalAudioLanguage
		if audioLang == "" {
			// Default: prefer Portuguese (Brazil), Portuguese, Spanish, English
			audioLang = "pt-BR,pt,por,pb,ptbr,portuguese,spa,es,spanish,eng,en,english"
		}
		subsLang := util.GlobalSubsLanguage
		if subsLang == "" {
			subsLang = "pt-BR,pt,por,pb,ptbr,portuguese,spa,es,spanish,eng,en,english"
		}
		mpvArgs = append(mpvArgs, fmt.Sprintf("--alang=%s", audioLang))
		mpvArgs = append(mpvArgs, fmt.Sprintf("--slang=%s", subsLang))
		util.Debugf("Movie/TV detected - applying language preferences: audio=%s, subs=%s", audioLang, subsLang)

		// Add external subtitle files if available (FlixHQ subtitles)
		// This follows the lobster.sh implementation for external subtitles
		subArgs := util.GetSubtitleArgs()
		if len(subArgs) > 0 {
			mpvArgs = append(mpvArgs, subArgs...)
			util.Debugf("Added external subtitles: %v", subArgs)
		}
	}

	// Initialize tracking and check for resume time
	tracker, resumeTime := initTracking(anilistID, currentEpisode, currentEpisodeNum)

	// For HLS streams, we'll seek after playback starts instead of using --start
	// because --start doesn't work reliably with HLS streams
	if resumeTime > 0 && !isHLSStream {
		mpvArgs = append(mpvArgs, fmt.Sprintf("--start=+%d", resumeTime))
	}

	// Fetch AniSkip data asynchronously
	skipDataChan := fetchAniSkipAsync(anilistID, currentEpisodeNum, currentEpisode)

	// Start the video with mpv
	mpvTimer := util.StartTimer("MPV:StartVideo")
	socketPath, err := StartVideo(videoURL, mpvArgs)
	mpvTimer.Stop()
	if err != nil {
		return fmt.Errorf("failed to start video: %w", err)
	}

	// For HLS streams, seek to resume position (includes waiting for video ready)
	if isHLSStream && resumeTime > 0 {
		util.Debugf("HLS stream detected with resume time %d seconds", resumeTime)
		hlsTimer := util.StartTimer("HLS:SeekToResume")
		seekToResumePosition(socketPath, resumeTime)
		hlsTimer.Stop()
	} else if isHLSStream {
		// Just wait for video ready without seeking
		hlsTimer := util.StartTimer("HLS:WaitForReady")
		waitForVideoReady(socketPath)
		hlsTimer.Stop()
	}

	// Apply AniSkip results to skip intros/outros
	applyAniSkipResults(skipDataChan, socketPath, currentEpisode, currentEpisodeNum)

	// Initialize Discord Rich Presence if updater is provided
	if updater != nil {
		initDiscordPresence(updater, socketPath, tracker, anilistID, currentEpisode, currentEpisodeNum)
		defer updater.Stop()
	}

	// Find the current episode index in the episode list
	currentEpisodeIndex := findEpisodeIndex(episodes, currentEpisodeNum)
	if currentEpisodeIndex == -1 {
		return fmt.Errorf("episode %d not found in list", currentEpisodeNum)
	}

	// Preload the next episode for seamless playback
	preloadNextEpisode(episodes, currentEpisodeIndex)

	// Start tracking routine if tracker is available
	stopTracking := startTrackingRoutine(tracker, socketPath, anilistID, currentEpisode, currentEpisodeNum, updater)

	// Handle user input for interactive controls
	err = handleUserInput(
		socketPath,
		episodes,
		currentEpisodeIndex,
		currentEpisodeNum,
		anilistID,
		updater,
		stopTracking,
		currentEpisode,
	)

	// Close the tracking channel if it's still open
	select {
	case <-stopTracking:
		// Channel already closed
	default:
		close(stopTracking)
	}

	return err
}

// getCurrentEpisode gets the current episode
// func getCurrentEpisode(episodes []models.Episode, num int) (*models.Episode, error) {
// 	if num < 1 || num > len(episodes) {
// 		return nil, fmt.Errorf("invalid episode number: %d", num)
// 	}
// 	return &episodes[num-1], nil
// }

// getCurrentEpisode retrieves the current episode based on the episode number
func getCurrentEpisode(episodes []models.Episode, num int) (*models.Episode, error) {
	numStr := fmt.Sprintf("%d", num)

	// First try: match by extracted episode number
	for i := range episodes {
		if ExtractEpisodeNumber(episodes[i].Number) == numStr {
			return &episodes[i], nil
		}
	}

	// Second try: match by Num field directly
	for i := range episodes {
		if episodes[i].Num == num {
			return &episodes[i], nil
		}
	}

	// Third try: match by Number field directly (for simple numeric cases)
	for i := range episodes {
		if episodes[i].Number == numStr {
			return &episodes[i], nil
		}
	}

	// Fourth try: check if we're within the bounds and use index-based access
	// This handles cases where episode numbering doesn't match exactly
	if num > 0 && num <= len(episodes) {
		return &episodes[num-1], nil
	}

	return nil, fmt.Errorf("episode %d not found", num)
}

// // initTracking initializes the tracking system
// func initTracking(anilistID int, episode *models.Episode, episodeNum int) (*tracking.LocalTracker, int) {
// 	if !tracking.IsCgoEnabled {
// 		if util.IsDebug {
// 			util.Debug("Tracking disabled: CGO not available")
// 		}
// 		return nil, 0
// 	}

// 	currentUser, err := user.Current()
// 	if err != nil {
// 		util.Errorf("Failed to get current user: %v", err)
// 		return nil, 0
// 	}

// 	var dbPath string
// 	if runtime.GOOS == "windows" {
// 		dbPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "GoAnime", "tracking", "progress.db")
// 	} else {
// 		dbPath = filepath.Join(currentUser.HomeDir, ".local", "goanime", "tracking", "progress.db")
// 	}

// 	tracker := tracking.NewLocalTracker(dbPath)
// 	if tracker == nil {
// 		return nil, 0
// 	}

// 	progress, err := tracker.GetAnime(anilistID, episode.URL)
// 	if err != nil || progress == nil || progress.PlaybackTime <= 0 {
// 		return tracker, 0
// 	}

// 	// Always use the selected episodeNum for the dialog, but use the tracked PlaybackTime if available
// 	if ok, _ := showResumeDialog(episodeNum, progress.PlaybackTime); ok {
// 		util.Debugf("Resuming from saved time: %d seconds for episode %d", progress.PlaybackTime, episodeNum)
// 		return tracker, progress.PlaybackTime
// 	}

// 	return tracker, 0
// }

// cachedDBPath stores the database path to avoid repeated user.Current() calls
var cachedDBPath string

// getTrackerDBPath returns the cached database path
func getTrackerDBPath() string {
	if cachedDBPath != "" {
		return cachedDBPath
	}

	currentUser, err := user.Current()
	if err != nil {
		util.Errorf("Failed to get current user: %v", err)
		return ""
	}

	if runtime.GOOS == "windows" {
		cachedDBPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "GoAnime", "tracking", "progress.db")
	} else {
		cachedDBPath = filepath.Join(currentUser.HomeDir, ".local", "goanime", "tracking", "progress.db")
	}

	return cachedDBPath
}

// initTracking inicializa o sistema de rastreamento
func initTracking(anilistID int, episode *models.Episode, episodeNum int) (*tracking.LocalTracker, int) {
	if !tracking.IsCgoEnabled {
		if util.IsDebug {
			util.Debug("Tracking disabled: CGO not available")
		}
		return nil, 0
	}

	dbPath := getTrackerDBPath()
	if dbPath == "" {
		return nil, 0
	}

	// First check if we have a cached tracker (fast path)
	tracker := tracking.GetGlobalTracker()
	if tracker == nil {
		// Need to initialize - this is only slow the first time
		tracker = tracking.NewLocalTracker(dbPath)
		if tracker == nil {
			return nil, 0
		}
	}

	// Debug: log what we're looking up
	util.Debugf("Tracking lookup: anilistID=%d, episode.URL=%s", anilistID, episode.URL)

	progress, err := tracker.GetAnime(anilistID, episode.URL)
	if err != nil {
		util.Debugf("Tracking lookup error: %v", err)
		return tracker, 0
	}
	if progress == nil {
		util.Debugf("Tracking lookup: no progress found")
		return tracker, 0
	}
	if progress.PlaybackTime <= 0 {
		util.Debugf("Tracking lookup: playback time is 0 or negative")
		return tracker, 0
	}

	util.Debugf("Tracking found: PlaybackTime=%d seconds", progress.PlaybackTime)

	// Usa o episodeNum selecionado para o diálogo, mas mantém o PlaybackTime do rastreamento
	if ok, _ := showResumeDialog(episodeNum, progress.PlaybackTime); ok {
		util.Debugf("Resuming from saved time: %d seconds for episode %d", progress.PlaybackTime, episodeNum)
		return tracker, progress.PlaybackTime
	}

	util.Debugf("User declined to resume")
	return tracker, 0
}

// InitTrackerAsync initializes the tracker in the background.
// Call this early in the application lifecycle to avoid delays later.
func InitTrackerAsync() {
	if !tracking.IsCgoEnabled {
		return
	}

	go func() {
		dbPath := getTrackerDBPath()
		if dbPath != "" {
			tracking.NewLocalTracker(dbPath)
		}
	}()
}

// fetchAniSkipAsync fetches AniSkip data in parallel
func fetchAniSkipAsync(anilistID, episodeNum int, episode *models.Episode) chan error {
	ch := make(chan error, 1)
	go func() {
		err := api.GetAndParseAniSkipData(anilistID, episodeNum, episode)
		ch <- err
	}()
	return ch
}

// applyAniSkipResults applies AniSkip results
// Reduced timeout from 3s to 2s for faster playback start
func applyAniSkipResults(ch chan error, socketPath string, episode *models.Episode, episodeNum int) {
	select {
	case err := <-ch:
		if err == nil {
			applySkipTimes(socketPath, episode)

			// For AllAnime episodes, also try to set chapter markers (like Curd does)
			if strings.Contains(episode.URL, "kibfyvtiFpKC") || len(episode.URL) < 30 {
				// This looks like an AllAnime episode ID, try to apply chapter markers
				allAnimeClient := scraper.NewAllAnimeClient()
				if chapterErr := allAnimeClient.SendSkipTimesToMPV(episode, socketPath, MpvSendCommand); chapterErr != nil {
					util.Debugf("Failed to set chapter markers: %v", chapterErr)
				}
			}
		} else {
			util.Debugf("AniSkip data unavailable for episode %d: %v", episodeNum, err)
		}
	case <-time.After(2 * time.Second):
		util.Debugf("Timeout fetching AniSkip data for episode %d", episodeNum)
	}
}

// initDiscordPresence initializes Discord presence
func initDiscordPresence(updater *discord.RichPresenceUpdater, socketPath string, tracker *tracking.LocalTracker, anilistID int, episode *models.Episode, episodeNum int) {
	updater.SetSocketPath(socketPath)
	updater.Start()

	go func() {
		waitForPlaybackStart(socketPath, updater)
		updateEpisodeDuration(socketPath, updater, tracker, anilistID, episode, episodeNum)
	}()
}

// waitForPlaybackStart waits for playback to start
func waitForPlaybackStart(socketPath string, updater *discord.RichPresenceUpdater) {
	maxAttempts := 30 // Max 30 seconds to wait
	for i := 0; i < maxAttempts; i++ {
		timePos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "time-pos"})
		if err == nil && timePos != nil && !updater.IsEpisodeStarted() {
			updater.SetEpisodeStarted(true)
			return
		}
		time.Sleep(1 * time.Second)
	}
	// Set as started anyway to avoid infinite loop
	updater.SetEpisodeStarted(true)
}

// updateEpisodeDuration updates the episode duration
func updateEpisodeDuration(socketPath string, updater *discord.RichPresenceUpdater, tracker *tracking.LocalTracker, anilistID int, episode *models.Episode, episodeNum int) {
	maxAttempts := 10 // Only try 10 times to get duration
	for i := 0; i < maxAttempts; i++ {
		if !updater.IsEpisodeStarted() {
			time.Sleep(2 * time.Second)
			continue
		}

		// If duration is already set, just update tracking and exit
		if updater.GetEpisodeDuration() > 0 {
			dur := updater.GetEpisodeDuration()
			updateTrackingWithDuration(tracker, anilistID, episode, episodeNum, dur)
			return
		}

		durationPos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "duration"})
		if err != nil || durationPos == nil {
			time.Sleep(2 * time.Second)
			continue
		}

		duration, ok := durationPos.(float64)
		if !ok || duration <= 0 {
			time.Sleep(2 * time.Second)
			continue
		}

		dur := time.Duration(duration * float64(time.Second))
		if dur < time.Second {
			dur = 24 * time.Minute
		}

		updater.SetEpisodeDuration(dur)
		updateTrackingWithDuration(tracker, anilistID, episode, episodeNum, dur)
		return
	}
}

// updateTrackingWithDuration updates the local tracker with episode info and duration
func updateTrackingWithDuration(tracker *tracking.LocalTracker, anilistID int, episode *models.Episode, episodeNum int, dur time.Duration) {
	if tracker != nil && dur > 0 {
		anime := tracking.Anime{
			AnilistID:     anilistID,
			AllanimeID:    episode.URL,
			EpisodeNumber: episodeNum,
			Duration:      int(dur.Seconds()),
			Title:         getEpisodeTitle(episode.Title),
			LastUpdated:   time.Now(),
		}
		if err := tracker.UpdateProgress(anime); err != nil {
			util.Errorf("Failed to update tracking: %v", err)
		}
	}
}

// getEpisodeTitle gets the episode title
func getEpisodeTitle(title models.TitleDetails) string {
	if title.English != "" {
		return title.English
	}
	if title.Romaji != "" {
		return title.Romaji
	}
	if title.Japanese != "" {
		return title.Japanese
	}
	return "No title"
}

// findEpisodeIndex finds the episode index
// func findEpisodeIndex(episodes []models.Episode, num int) int {
// 	episodeStr := strconv.Itoa(num)
// 	for i, ep := range episodes {
// 		if ExtractEpisodeNumber(ep.Number) == episodeStr {
// 			return i
// 		}
// 	}
// 	return -1
// }

// findEpisodeIndex finds the episode index based on the episode number
func findEpisodeIndex(episodes []models.Episode, num int) int {
	episodeStr := fmt.Sprintf("%d", num)

	// First try: match by extracted episode number
	for i, ep := range episodes {
		extractedNum := ExtractEpisodeNumber(ep.Number)
		if extractedNum == episodeStr {
			return i
		}
	}

	// Second try: match by Num field directly
	for i, ep := range episodes {
		if ep.Num == num {
			return i
		}
	}

	// Third try: match by Number field directly
	for i, ep := range episodes {
		if ep.Number == episodeStr {
			return i
		}
	}

	// Fourth try: use index-based access if within bounds
	if num > 0 && num <= len(episodes) {
		return num - 1
	}

	return -1 // Episode not found
}

// preloadNextEpisode preloads the next episode
func preloadNextEpisode(episodes []models.Episode, currentIndex int) {
	if currentIndex+1 >= len(episodes) {
		return
	}

	// Skip preloading for AllAnime episodes (they use IDs, not HTTP URLs)
	nextEpisodeURL := episodes[currentIndex+1].URL
	if len(nextEpisodeURL) < 30 && !strings.Contains(nextEpisodeURL, "http") {
		return
	}

	go func() {
		_, _ = GetVideoURLForEpisode(episodes[currentIndex+1].URL)
		// Preloading errors are ignored as this is not critical
	}()
}

// startTrackingRoutine starts the tracking routine
// Optimized with adaptive update interval to reduce overhead
func startTrackingRoutine(tracker *tracking.LocalTracker, socketPath string, anilistID int, episode *models.Episode, episodeNum int, updater *discord.RichPresenceUpdater) chan struct{} {
	stopChan := make(chan struct{})
	if tracker == nil {
		return stopChan
	}

	go func() {
		// Start with faster updates, then slow down for efficiency
		ticker := time.NewTicker(3 * time.Second) // Increased from 2s to reduce overhead
		defer ticker.Stop()

		updateCount := 0
		for {
			select {
			case <-ticker.C:
				updateTracking(tracker, socketPath, anilistID, episode, episodeNum, updater)
				updateCount++
				// After 30 updates (~90 seconds), slow down to every 5 seconds
				if updateCount == 30 {
					ticker.Reset(5 * time.Second)
				}
			case <-stopChan:
				// Final update before stopping
				updateTracking(tracker, socketPath, anilistID, episode, episodeNum, updater)
				return
			}
		}
	}()

	return stopChan
}

// updateTracking updates tracking
func updateTracking(tracker *tracking.LocalTracker, socketPath string, anilistID int, episode *models.Episode, episodeNum int, updater *discord.RichPresenceUpdater) {
	timePos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "time-pos"})
	if err != nil || timePos == nil {
		return
	}

	position, ok := timePos.(float64)
	if !ok {
		return
	}

	duration := 1440 // Default duration in seconds (24 minutes)
	if updater != nil {
		episodeDur := updater.GetEpisodeDuration()
		if episodeDur > 0 {
			duration = int(episodeDur.Seconds())
		}
	}

	// Ensure duration is valid before updating tracking
	if duration <= 0 {
		duration = 1440 // Fallback to default
	}

	anime := tracking.Anime{
		AnilistID:     anilistID,
		AllanimeID:    episode.URL,
		EpisodeNumber: episodeNum,
		PlaybackTime:  int(position),
		Duration:      duration,
		Title:         getEpisodeTitle(episode.Title),
		LastUpdated:   time.Now(),
	}

	if err := tracker.UpdateProgress(anime); err != nil {
		util.Errorf("Error updating tracking: %v", err)
	}
}

// showPlayerMenu displays an interactive menu using huh.Select
func showPlayerMenu(animeName string, currentEpisodeNum int, isMovieOrTV bool) (string, error) {
	var choice string

	title := "GoAnime Player Controls"
	if animeName != "" {
		title = fmt.Sprintf("Now playing: %s - Episode %d", animeName, currentEpisodeNum)
	}

	// Build menu options
	options := []huh.Option[string]{
		huh.NewOption("← Back ", "download_options"),
		huh.NewOption("Next episode", "next"),
		huh.NewOption("Previous episode", "previous"),
		huh.NewOption("Select episode", "select"),
		huh.NewOption("Change anime", "change"),
		huh.NewOption("Skip intro", "skip"),
		huh.NewOption("Exit", "quit"),
	}

	// isMovieOrTV parameter kept for future use but not currently adding extra options
	_ = isMovieOrTV

	menu := huh.NewSelect[string]().
		Title(title).
		Description("Choose an action:").
		Options(options...).
		Value(&choice)

	if err := menu.Run(); err != nil {
		return "", fmt.Errorf("error showing menu: %w", err)
	}

	return choice, nil
}

// handleUserInput manages user input
func handleUserInput(
	socketPath string,
	episodes []models.Episode,
	currentIndex int,
	currentEpisodeNum int,
	anilistID int,
	updater *discord.RichPresenceUpdater,
	stopTracking chan struct{},
	currentEpisode *models.Episode,
) error {
	// Get anime name for display and check if movie/TV
	var animeName string
	isMovieOrTV := false
	if updater != nil && updater.GetAnime() != nil {
		anime := updater.GetAnime()
		animeName = anime.Name
		isMovieOrTV = anime.IsMovieOrTV() || strings.Contains(strings.ToLower(anime.Source), "flixhq")
	}

	for {
		choice, err := showPlayerMenu(animeName, currentEpisodeNum, isMovieOrTV)
		if err != nil {
			return fmt.Errorf("error showing menu: %w", err)
		}

		switch choice {
		case "download_options":
			// Stop playback and go back to download options
			_, _ = mpvSendCommand(socketPath, []interface{}{"quit"})
			return ErrBackToDownloadOptions
		case "next":
			return playNextEpisode(currentIndex+1, episodes, anilistID, updater, stopTracking, socketPath)
		case "previous":
			return playPreviousEpisode(currentIndex-1, episodes, anilistID, updater, stopTracking, socketPath)
		case "quit":
			_, _ = mpvSendCommand(socketPath, []interface{}{"quit"})
			return ErrUserQuit
		case "change":
			_, _ = mpvSendCommand(socketPath, []interface{}{"quit"})
			return ErrChangeAnime
		case "select":
			return selectEpisode(episodes, anilistID, updater, stopTracking, socketPath)
		case "audio":
			selectAudioTrack(socketPath)
		case "subtitle":
			selectSubtitleTrack(socketPath)
		case "skip":
			skipIntro(socketPath, currentEpisode)
		}
	}
}

// playNextEpisode plays next episode
func playNextEpisode(newIndex int, episodes []models.Episode, anilistID int, updater *discord.RichPresenceUpdater, stopTracking chan struct{}, socketPath string) error {
	if newIndex >= len(episodes) {
		fmt.Println("You are on the last episode")
		return nil
	}
	return switchEpisode(newIndex, episodes, anilistID, updater, stopTracking, socketPath)
}

// playPreviousEpisode plays previous episode
func playPreviousEpisode(newIndex int, episodes []models.Episode, anilistID int, updater *discord.RichPresenceUpdater, stopTracking chan struct{}, socketPath string) error {
	if newIndex < 0 {
		fmt.Println("You are on the first episode")
		return nil
	}
	return switchEpisode(newIndex, episodes, anilistID, updater, stopTracking, socketPath)
}

// selectEpisode allows selecting an episode
func selectEpisode(episodes []models.Episode, anilistID int, updater *discord.RichPresenceUpdater, stopTracking chan struct{}, socketPath string) error {
	selectedURL, selectedNumStr, err := SelectEpisodeWithFuzzyFinder(episodes)
	if err != nil {
		// If user selected back, return nil to continue without action
		if errors.Is(err, ErrBackRequested) {
			return nil
		}
		return fmt.Errorf("failed to select episode: %w", err)
	}

	for i, ep := range episodes {
		if ep.URL == selectedURL {
			return switchEpisode(i, episodes, anilistID, updater, stopTracking, socketPath)
		}
	}

	return fmt.Errorf("episode %s not found", selectedNumStr)
}

// switchEpisode switches between episodes
func switchEpisode(newIndex int, episodes []models.Episode, anilistID int, updater *discord.RichPresenceUpdater, stopTracking chan struct{}, socketPath string) error {
	target := episodes[newIndex]
	targetNum, err := strconv.Atoi(ExtractEpisodeNumber(target.Number))
	if err != nil {
		return fmt.Errorf("invalid episode number: %w", err)
	}

	var anime *models.Anime
	if updater != nil {
		anime = updater.GetAnime()
	}

	// If no updater/anime context, try to synthesize from lastAnimeURL
	if anime == nil && lastAnimeURL != "" {
		guessedSource := ""
		if (len(lastAnimeURL) < 30 && !strings.Contains(lastAnimeURL, "http")) || strings.Contains(lastAnimeURL, "allanime") {
			guessedSource = "AllAnime"
		}
		anime = &models.Anime{URL: lastAnimeURL, Source: guessedSource}
	}

	targetURL, err := GetVideoURLForEpisodeEnhanced(&target, anime)
	if err != nil {
		return fmt.Errorf("failed to get video URL: %w", err)
	}

	if updater != nil {
		updater.Stop()
	}

	close(stopTracking)
	_, _ = mpvSendCommand(socketPath, []interface{}{"quit"})

	var newUpdater *discord.RichPresenceUpdater
	if updater != nil {
		duration := time.Duration(target.Duration) * time.Second
		newUpdater = discord.NewRichPresenceUpdater(
			updater.GetAnime(),
			updater.GetIsPaused(),
			updater.GetAnimeMutex(),
			updater.GetUpdateFreq(),
			duration,
			"",
			MpvSendCommand,
		)
		newUpdater.SetEpisodeStarted(false)
	}

	return playVideo(targetURL, episodes, targetNum, anilistID, newUpdater)
}

// skipIntro skips the intro
func skipIntro(socketPath string, episode *models.Episode) {
	if episode.SkipTimes.Op.End > 0 {
		_, _ = mpvSendCommand(socketPath, []interface{}{"seek", episode.SkipTimes.Op.End, "absolute"})
		fmt.Printf("Intro skipped to %ds\n", episode.SkipTimes.Op.End)
	} else {
		fmt.Println("Intro skip data not available")
	}
}

// selectAudioTrack shows a menu to select audio track
func selectAudioTrack(socketPath string) {
	tracks, err := GetAudioTracks(socketPath)
	if err != nil {
		fmt.Printf("Error getting audio tracks: %v\n", err)
		return
	}

	if len(tracks) == 0 {
		fmt.Println("No audio tracks available")
		return
	}

	currentID, _ := GetCurrentAudioTrack(socketPath)

	var options []huh.Option[int]
	for _, track := range tracks {
		id := int(track["id"].(float64))

		// Build label with all available info
		var parts []string

		// Language
		lang := ""
		if l, ok := track["lang"].(string); ok && l != "" {
			lang = l
		}

		// Title (often contains language info)
		title := ""
		if t, ok := track["title"].(string); ok && t != "" {
			title = t
		}

		// Codec
		codec := ""
		if c, ok := track["codec"].(string); ok && c != "" {
			codec = c
		}

		// Channels (stereo, 5.1, etc.)
		channels := ""
		if ch, ok := track["demux-channels"].(string); ok && ch != "" {
			channels = ch
		} else if ch, ok := track["audio-channels"].(float64); ok {
			channels = fmt.Sprintf("%.0fch", ch)
		}

		// Build the label
		if title != "" {
			parts = append(parts, title)
		} else if lang != "" {
			parts = append(parts, lang)
		} else {
			parts = append(parts, fmt.Sprintf("Audio %d", id))
		}

		if codec != "" {
			parts = append(parts, codec)
		}
		if channels != "" {
			parts = append(parts, channels)
		}

		label := fmt.Sprintf("Track %d: %s", id, strings.Join(parts, " | "))
		if id == currentID {
			label = "* " + label + " (current)"
		}
		options = append(options, huh.NewOption(label, id))
	}

	var selected int
	menu := huh.NewSelect[int]().
		Title("Select Audio Track").
		Description("Choose the audio language/track:").
		Options(options...).
		Value(&selected)

	if err := menu.Run(); err != nil {
		return
	}

	if err := SetAudioTrack(socketPath, selected); err != nil {
		fmt.Printf("Error setting audio track: %v\n", err)
	} else {
		fmt.Printf("Audio track changed to %d\n", selected)
	}
}

// selectSubtitleTrack shows a menu to select subtitle track
func selectSubtitleTrack(socketPath string) {
	tracks, err := GetSubtitleTracks(socketPath)
	if err != nil {
		fmt.Printf("Error getting subtitle tracks: %v\n", err)
		return
	}

	currentID, _ := GetCurrentSubtitleTrack(socketPath)

	var options []huh.Option[int]
	// Add option to disable subtitles
	disableLabel := "Disable subtitles"
	if currentID == 0 {
		disableLabel = "* " + disableLabel + " (current)"
	}
	options = append(options, huh.NewOption(disableLabel, 0))

	for _, track := range tracks {
		id := int(track["id"].(float64))

		// Build label with all available info
		var parts []string

		// Language
		lang := ""
		if l, ok := track["lang"].(string); ok && l != "" {
			lang = l
		}

		// Title (often contains language info)
		title := ""
		if t, ok := track["title"].(string); ok && t != "" {
			title = t
		}

		// Build the label
		if title != "" {
			parts = append(parts, title)
		} else if lang != "" {
			parts = append(parts, lang)
		} else {
			parts = append(parts, fmt.Sprintf("Subtitle %d", id))
		}

		label := fmt.Sprintf("Track %d: %s", id, strings.Join(parts, " | "))
		if id == currentID {
			label = "* " + label + " (current)"
		}
		options = append(options, huh.NewOption(label, id))
	}

	var selected int
	menu := huh.NewSelect[int]().
		Title("Select Subtitle Track").
		Description("Choose the subtitle language:").
		Options(options...).
		Value(&selected)

	if err := menu.Run(); err != nil {
		return
	}

	if selected == 0 {
		// Disable subtitles
		_, _ = mpvSendCommand(socketPath, []interface{}{"set_property", "sid", "no"})
		fmt.Println("Subtitles disabled")
	} else {
		if err := SetSubtitleTrack(socketPath, selected); err != nil {
			fmt.Printf("Error setting subtitle track: %v\n", err)
		} else {
			fmt.Printf("Subtitle track changed to %d\n", selected)
		}
	}
}
