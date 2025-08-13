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

	// Initialize tracking and check for resume time
	tracker, resumeTime := initTracking(anilistID, currentEpisode, currentEpisodeNum)
	if resumeTime > 0 {
		mpvArgs = append(mpvArgs, fmt.Sprintf("--start=+%d", resumeTime))
	}

	// Fetch AniSkip data asynchronously
	skipDataChan := fetchAniSkipAsync(anilistID, currentEpisodeNum, currentEpisode)

	// Start the video with mpv
	socketPath, err := StartVideo(videoURL, mpvArgs)
	if err != nil {
		return fmt.Errorf("failed to start video: %w", err)
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
	for _, ep := range episodes {
		if ExtractEpisodeNumber(ep.Number) == fmt.Sprintf("%d", num) {
			return &ep, nil
		}
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

// initTracking inicializa o sistema de rastreamento
func initTracking(anilistID int, episode *models.Episode, episodeNum int) (*tracking.LocalTracker, int) {
	if !tracking.IsCgoEnabled {
		if util.IsDebug {
			util.Debug("Tracking desabilitado: CGO não disponível")
		}
		return nil, 0
	}

	currentUser, err := user.Current()
	if err != nil {
		util.Errorf("Falha ao obter usuário atual: %v", err)
		return nil, 0
	}

	var dbPath string
	if runtime.GOOS == "windows" {
		dbPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "GoAnime", "tracking", "progress.db")
	} else {
		dbPath = filepath.Join(currentUser.HomeDir, ".local", "goanime", "tracking", "progress.db")
	}

	tracker := tracking.NewLocalTracker(dbPath)
	if tracker == nil {
		return nil, 0
	}

	progress, err := tracker.GetAnime(anilistID, episode.URL)
	if err != nil || progress == nil || progress.PlaybackTime <= 0 {
		return tracker, 0
	}

	// Usa o episodeNum selecionado para o diálogo, mas mantém o PlaybackTime do rastreamento
	if ok, _ := showResumeDialog(episodeNum, progress.PlaybackTime); ok {
		util.Debugf("Retomando do tempo salvo: %d segundos para o episódio %d", progress.PlaybackTime, episodeNum)
		return tracker, progress.PlaybackTime
	}

	return tracker, 0
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
	case <-time.After(3 * time.Second):
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
	for {
		timePos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "time-pos"})
		if err == nil && timePos != nil && !updater.IsEpisodeStarted() {
			updater.SetEpisodeStarted(true)
			return
		}
		time.Sleep(1 * time.Second)
	}
}

// updateEpisodeDuration updates the episode duration
func updateEpisodeDuration(socketPath string, updater *discord.RichPresenceUpdater, tracker *tracking.LocalTracker, anilistID int, episode *models.Episode, episodeNum int) {
	for {
		if !updater.IsEpisodeStarted() || updater.GetEpisodeDuration() == 0 {
			time.Sleep(1 * time.Second)
			continue
		}

		durationPos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "duration"})
		if err != nil || durationPos == nil {
			break
		}

		duration, ok := durationPos.(float64)
		if !ok {
			break
		}

		dur := time.Duration(duration * float64(time.Second))
		if dur < time.Second {
			dur = 24 * time.Minute
		}

		updater.SetEpisodeDuration(dur)

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
		break
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
	for i, ep := range episodes {
		extractedNum := ExtractEpisodeNumber(ep.Number)
		if extractedNum == episodeStr {
			return i
		}
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
func startTrackingRoutine(tracker *tracking.LocalTracker, socketPath string, anilistID int, episode *models.Episode, episodeNum int, updater *discord.RichPresenceUpdater) chan struct{} {
	stopChan := make(chan struct{})
	if tracker == nil {
		return stopChan
	}

	go func() {
		ticker := time.NewTicker(2 * time.Second)
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				updateTracking(tracker, socketPath, anilistID, episode, episodeNum, updater)
			case <-stopChan:
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
func showPlayerMenu(animeName string, currentEpisodeNum int) (string, error) {
	var choice string

	title := "GoAnime Player Controls"
	if animeName != "" {
		title = fmt.Sprintf("Now playing: %s - Episode %d", animeName, currentEpisodeNum)
	}

	menu := huh.NewSelect[string]().
		Title(title).
		Description("Choose an action:").
		Options(
			huh.NewOption("Next episode", "next"),
			huh.NewOption("Previous episode", "previous"),
			huh.NewOption("Select episode", "select"),
			huh.NewOption("Change anime", "change"),
			huh.NewOption("Skip intro", "skip"),
			huh.NewOption("Exit", "quit"),
		).
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
	// Get anime name for display
	var animeName string
	if updater != nil && updater.GetAnime() != nil {
		animeName = updater.GetAnime().Name
	}

	for {
		choice, err := showPlayerMenu(animeName, currentEpisodeNum)
		if err != nil {
			return fmt.Errorf("error showing menu: %w", err)
		}

		switch choice {
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
