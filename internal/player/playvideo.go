package player

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tracking"
	"github.com/alvarorichard/Goanime/internal/util"
)

// applySkipTimes attempts to set the skip times (OP and ED) on a running mpv instance
// by setting the "script-opts" property.
// Note: This relies on mpv's ability to update script-opts dynamically via set_property,
// or a script that interprets these options.
func applySkipTimes(socketPath string, episode *models.Episode) {
	var opts []string // <-- Estilo preferido para slices vazios
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
			if util.IsDebug {
				log.Printf("Failed to apply skip times to mpv via set_property script-opts: %v. Command: set_property script-opts %s", cmdErr, combinedOpts)
			}
		} else {
			if util.IsDebug {
				log.Printf("Successfully applied skip times to mpv: %s", combinedOpts)
			}
		}
	} else if util.IsDebug {
		log.Printf("No skip times to apply for episode %s", episode.Number)
	}
}

// playVideo handles the online playback of a video and user interaction.
func playVideo(
	videoURL string,
	episodes []models.Episode,
	currentEpisodeNum int,
	anilistID int,
	updater *RichPresenceUpdater,
) error {
	// The videoURL should already be processed by the caller (e.g., HandleDownloadAndPlay or GetVideoURLForEpisode)
	// to be a direct media link after any necessary quality selection.
	/*
		// Prompt for quality selection if multiple are available (streaming)
		if strings.HasPrefix(videoURL, "http") {
			if url, err := ExtractVideoSourcesWithPrompt(videoURL); err == nil {
				videoURL = url
			}
		}
	*/

	// Fix the URL if it has a double 'p' in the quality suffix
	videoURL = strings.Replace(videoURL, "720pp.mp4", "720p.mp4", 1)

	// Debug log video URL
	if util.IsDebug {
		log.Printf("Video URL: %s", videoURL)
	}

	// Get current episode and user
	currentEpisode := &episodes[currentEpisodeNum-1]
	currentUser, err := user.Current()
	if err != nil {
		return fmt.Errorf("failed to get current user: %w", err)
	}

	// Initialize mpv arguments
	// mpvArgs := []string{
	// 	"--hwdec=auto",             // Enable hardware decoding
	// 	"--vo=gpu",                 // Use GPU output
	// 	"--no-config",              // Skip loading config files
	// 	"--cache=yes",              // Enable caching
	// 	"--demuxer-max-bytes=500M", // Increase cache size
	// }
	mpvArgs := []string{
		"--hwdec=auto-safe",           // Hardware decoding com fallback seguro
		"--vo=gpu",                    // GPU rendering
		"--profile=fast",              // Perfil de performance
		"--cache=yes",                 // Habilita cache
		"--demuxer-max-bytes=300M",    // Tamanho do buffer de rede
		"--demuxer-readahead-secs=20", // Pré-carregamento em segundos
		"--no-config",                 // Ignora arquivos de configuração locais
		//"--input-ipc-server=auto", // Socket IPC automático (evita conflitos)
		"--video-latency-hacks=yes",
		"--audio-display=no",
	}
	var tracker *tracking.LocalTracker
	var resumeTime int
	// Try to initialize tracking system if CGO is enabled
	if tracking.IsCgoEnabled {
		var dbPath string
		if runtime.GOOS == "windows" {
			// Use %LOCALAPPDATA% on Windows
			dbPath = filepath.Join(os.Getenv("LOCALAPPDATA"), "GoAnime", "tracking", "progress.db")
		} else {
			// Use ~/.local/goanime on Unix systems
			dbPath = filepath.Join(
				currentUser.HomeDir, ".local", "goanime", "tracking", "progress.db",
			)
		}
		tracker = tracking.NewLocalTracker(dbPath)

		// Local tracking: ask to resume if progress exists and tracker initialized
		if tracker != nil {
			if progress, err := tracker.GetAnime(anilistID, currentEpisode.URL); err == nil &&
				progress != nil && progress.EpisodeNumber == currentEpisodeNum && progress.PlaybackTime > 0 {
				fmt.Printf("\nProgresso salvo encontrado: episódio %d, tempo %d segundos.\n",
					progress.EpisodeNumber, progress.PlaybackTime)
				if ok, _ := promptYesNo("Deseja retomar de onde parou?"); ok {
					resumeTime = progress.PlaybackTime
					if util.IsDebug {
						log.Printf("Retomando do tempo salvo: %d segundos", resumeTime)
					}
				}
			}
		}
	} else if util.IsDebug {
		log.Println("Tracking disabled: CGO not available")
	}

	// Add resume time to mpv arguments if available
	if resumeTime > 0 {
		mpvArgs = append(mpvArgs, fmt.Sprintf("--start=+%d", resumeTime))
	}

	// Channel to receive AniSkip results
	skipDataChan := make(chan error, 1)

	// Fetch AniSkip data concurrently.
	// currentEpisode is a pointer and will be populated by GetAndParseAniSkipData.
	go func() {
		//var aniskipErr error
		// GetAndParseAniSkipData should handle anilistID <= 0 gracefully (e.g., return nil error and no skip times).
		// If anilistID is 0, currentEpisode.SkipTimes will remain empty.
		aniskipErr := api.GetAndParseAniSkipData(anilistID, currentEpisodeNum, currentEpisode)
		skipDataChan <- aniskipErr
	}()

	// Start mpv with IPC support
	socketPath, err := StartVideo(videoURL, mpvArgs)
	if err != nil {
		return fmt.Errorf("failed to start video with IPC: %w", err)
	}

	// Wait for AniSkip data or timeout
	select {
	case errAniskip := <-skipDataChan:
		if errAniskip != nil {
			log.Printf("AniSkip data not available or failed to fetch for episode %d: %v", currentEpisodeNum, errAniskip)
		} else {
			// Successfully fetched (or anilistID was invalid and GetAndParseAniSkipData handled it gracefully by returning nil error)
			if util.IsDebug {
				if currentEpisode.SkipTimes.Op.Start > 0 || currentEpisode.SkipTimes.Ed.Start > 0 {
					log.Printf("AniSkip data fetched for episode %d: %+v", currentEpisodeNum, currentEpisode.SkipTimes)
				} else if anilistID > 0 { // anilistID was valid, but no skip times found in the data
					log.Printf("AniSkip data fetched for episode %d, but no skip intervals were found in the data.", currentEpisodeNum)
				} else { // anilistID was <= 0
					log.Println("AniSkip data fetch skipped or no data due to invalid/missing anilistID.")
				}
			}
			// applySkipTimes will check if SkipTimes has actual values.
			applySkipTimes(socketPath, currentEpisode)
		}
	case <-time.After(3 * time.Second):
		log.Printf("Timeout waiting for AniSkip data for episode %d. Continuing without applying skip times dynamically.", currentEpisodeNum)
	}

	// Helper to get a non-empty title
	getEpisodeTitle := func(title models.TitleDetails) string {
		if title.English != "" {
			return title.English
		}
		if title.Romaji != "" {
			return title.Romaji
		}
		if title.Japanese != "" {
			return title.Japanese
		}
		return "Unknown Title"
	}

	// Only proceed with Rich Presence updates if updater is not nil
	if updater != nil {
		// Wait for the episode to start before retrieving the duration
		go func() {
			for {
				timePos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "time-pos"})
				if err != nil {
					if util.IsDebug {
						log.Printf("Error getting playback time: %v", err)
					}
				}
				if timePos != nil {
					if !updater.episodeStarted {
						updater.episodeStarted = true
					}
					break
				}
				time.Sleep(1 * time.Second)
			}
		}()

		// Retrieve the video duration once the episode has started
		go func() {
			for {
				if updater.episodeStarted && updater.episodeDuration == 0 {
					durationPos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "duration"})
					if err != nil {
						log.Printf("Error getting video duration: %v", err)
					} else if durationPos != nil {
						if duration, ok := durationPos.(float64); ok {
							updater.episodeDuration = time.Duration(duration * float64(time.Second))
							if util.IsDebug {
								log.Printf("Retrieved Video duration: %v seconds", updater.episodeDuration.Seconds())
							}

							// Ensure we have a valid duration for database storage (minimum 1 second)
							if updater.episodeDuration < time.Second {
								log.Printf("Warning: Retrieved episode duration is very small (%v). Setting a default duration.", updater.episodeDuration)
								updater.episodeDuration = 24 * time.Minute
							}

							// Only update if we have a tracker and the duration is positive
							if tracker != nil && int(updater.episodeDuration.Seconds()) > 0 {
								anime := tracking.Anime{
									AnilistID:     anilistID,
									AllanimeID:    currentEpisode.URL,
									EpisodeNumber: currentEpisodeNum,
									Duration:      int(updater.episodeDuration.Seconds()),
									Title:         getEpisodeTitle(currentEpisode.Title),
									LastUpdated:   time.Now(),
								}
								if err := tracker.UpdateProgress(anime); err != nil {
									log.Printf("Failed to update local tracking: %v", err)
								}
							}
						} else {
							log.Printf("Error: duration is not a float64")
						}
					}
					break
				}
				time.Sleep(1 * time.Second)
			}
		}()

		updater.socketPath = socketPath
		updater.Start()
		defer updater.Stop()
	}

	// Optimize Episode Index Lookup
	// Assume episodes are ordered sequentially and try direct indexing first.
	currentEpisodeIndex := currentEpisodeNum - 1
	if currentEpisodeIndex < 0 || currentEpisodeIndex >= len(episodes) || ExtractEpisodeNumber(episodes[currentEpisodeIndex].Number) != strconv.Itoa(currentEpisodeNum) {
		// Fallback to search if assumption fails or direct index is out of bounds
		currentEpisodeIndex = -1 // Reset before searching
		for i, ep := range episodes {
			if ExtractEpisodeNumber(ep.Number) == strconv.Itoa(currentEpisodeNum) {
				currentEpisodeIndex = i
				break
			}
		}
	}

	if currentEpisodeIndex == -1 {
		return fmt.Errorf("current episode number %d not found in episodes list", currentEpisodeNum)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Press 'n' for next episode, 'p' for previous episode, 'e' to select episode, 'q' to quit, 's' to skip intro:")

	// Preload next episode in background
	if currentEpisodeIndex+1 < len(episodes) {
		go func() {
			nextEpisode := episodes[currentEpisodeIndex+1]
			// The result is intentionally ignored; this is for pre-caching/pre-resolving.
			_, err := GetVideoURLForEpisode(nextEpisode.URL)
			if err != nil && util.IsDebug {
				log.Printf("Error preloading next episode %s: %v", nextEpisode.Number, err)
			}
		}()
	}

	// Start a goroutine to periodically update local tracking
	stopTracking := make(chan struct{})
	if tracker != nil {
		go func() {
			ticker := time.NewTicker(2 * time.Second) // was 5s, now 2s for more responsive tracking
			defer ticker.Stop()
			for {
				select {
				case <-ticker.C:
					timePos, err := mpvSendCommand(socketPath, []interface{}{"get_property", "time-pos"})
					if err != nil {
						if util.IsDebug {
							log.Printf("Error getting playback time for tracking: %v", err)
						}
						continue
					}
					if timePos != nil {
						if position, ok := timePos.(float64); ok {
							// Ensure we have a valid duration (at least 1 second)
							var duration int
							if updater != nil {
								duration = int(updater.episodeDuration.Seconds())
							}
							if duration <= 0 {
								// Use a default duration if the real one isn't available yet
								duration = 1440 // 24 minutes in seconds as a safe default
							}

							anime := tracking.Anime{
								AnilistID:     anilistID,
								AllanimeID:    currentEpisode.URL,
								EpisodeNumber: currentEpisodeNum,
								PlaybackTime:  int(position),
								Duration:      duration,
								Title:         getEpisodeTitle(currentEpisode.Title),
								LastUpdated:   time.Now(),
							}
							if err := tracker.UpdateProgress(anime); err != nil {
								log.Printf("Failed to update local tracking: %v", err)
							}
						}
					}
				case <-stopTracking:
					return
				}
			}
		}()
	} else if util.IsDebug {
		log.Println("Skipping tracking: tracker not initialized")
	}

	for {
		char, _, err := reader.ReadRune()
		if err != nil {
			fmt.Printf("Failed to read command: %v\n", err)
			break
		}
		switch char {
		case 'n': // Next episode
			if currentEpisodeIndex+1 < len(episodes) {
				nextEpisode := episodes[currentEpisodeIndex+1]
				if updater != nil {
					updater.Stop()
				}
				nextVideoURL, err := GetVideoURLForEpisode(nextEpisode.URL)
				if err != nil {
					fmt.Printf("Failed to get video URL for next episode: %v\n", err)
					continue
				}
				nextEpisodeDuration := time.Duration(nextEpisode.Duration) * time.Second
				var newUpdater *RichPresenceUpdater
				if updater != nil {
					newUpdater = NewRichPresenceUpdater(
						updater.anime,
						updater.isPaused,
						updater.animeMutex,
						updater.updateFreq,
						nextEpisodeDuration,
						"",
					)
					updater.episodeStarted = false
				}
				close(stopTracking)
				return playVideo(nextVideoURL, episodes, currentEpisodeNum+1, anilistID, newUpdater)
			} else {
				fmt.Println("Already at the last episode.")
			}
		case 'p': // Previous episode
			if currentEpisodeIndex > 0 {
				prevEpisode := episodes[currentEpisodeIndex-1]
				if updater != nil {
					updater.Stop()
				}
				prevVideoURL, err := GetVideoURLForEpisode(prevEpisode.URL)
				if err != nil {
					fmt.Printf("Failed to get video URL for previous episode: %v\n", err)
					continue
				}
				prevEpisodeDuration := time.Duration(prevEpisode.Duration) * time.Second
				var newUpdater *RichPresenceUpdater
				if updater != nil {
					newUpdater = NewRichPresenceUpdater(
						updater.anime,
						updater.isPaused,
						updater.animeMutex,
						updater.updateFreq,
						prevEpisodeDuration,
						"",
					)
					updater.episodeStarted = false
				}
				close(stopTracking)
				return playVideo(prevVideoURL, episodes, currentEpisodeNum-1, anilistID, newUpdater)
			} else {
				fmt.Println("Already at the first episode.")
			}
		case 'q': // Quit
			fmt.Println("Quitting video playback.")
			close(stopTracking)
			_, _ = mpvSendCommand(socketPath, []interface{}{"quit"})
			return nil
		case 'e': // Select episode
			if updater != nil {
				updater.Stop()
			}
			selectedEpisodeURL, selectedEpisodeNumberStr, err := SelectEpisodeWithFuzzyFinder(episodes)
			if err != nil {
				fmt.Printf("Failed to select episode: %v\n", err)
				continue
			}
			selectedVideoURL, err := GetVideoURLForEpisode(selectedEpisodeURL)
			if err != nil {
				fmt.Printf("Failed to get video URL for selected episode: %v\n", err)
				continue
			}
			selectedEpisodeNum, _ := strconv.Atoi(ExtractEpisodeNumber(selectedEpisodeNumberStr))

			// Find the selected episode to get its duration
			var selectedEpisodeDuration time.Duration
			for _, ep := range episodes {
				if ExtractEpisodeNumber(ep.Number) == strconv.Itoa(selectedEpisodeNum) {
					selectedEpisodeDuration = time.Duration(ep.Duration) * time.Second
					break
				}
			}

			var newUpdater *RichPresenceUpdater
			if updater != nil {
				newUpdater = NewRichPresenceUpdater(
					updater.anime,
					updater.isPaused,
					updater.animeMutex,
					updater.updateFreq,
					selectedEpisodeDuration,
					"",
				)
				updater.episodeStarted = false
			}
			close(stopTracking)
			return playVideo(selectedVideoURL, episodes, selectedEpisodeNum, anilistID, newUpdater)
		case 's': // Skip intro (OP)
			if currentEpisode.SkipTimes.Op.End > 0 {
				fmt.Printf("Skipping intro to %d seconds.\n", currentEpisode.SkipTimes.Op.End)
				_, _ = mpvSendCommand(socketPath, []interface{}{"seek", currentEpisode.SkipTimes.Op.End, "absolute"})
			} else {
				fmt.Println("No intro skip data available for this episode.")
			}
		}
	}
	return nil
}
