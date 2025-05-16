package player

import (
	"bufio"
	"fmt"
	"log"
	"os"
	"os/user"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/tracking"
	"github.com/alvarorichard/Goanime/internal/util"
)

// playVideo handles the online playback of a video and user interaction.
func playVideo(
	videoURL string,
	episodes []models.Episode,
	currentEpisodeNum int,
	animeMalID int,
	updater *RichPresenceUpdater,
) error {
	// Prompt for quality selection if multiple are available (streaming)
	if strings.HasPrefix(videoURL, "http") {
		if url, err := ExtractVideoSourcesWithPrompt(videoURL); err == nil {
			videoURL = url
		}
	}

	// Initialize mpv arguments
	mpvArgs := make([]string, 0)

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

	dbPath := filepath.Join(
		currentUser.HomeDir, ".local", "goanime", "tracking", "progress.db",
	)
	tracker := tracking.NewLocalTracker(dbPath)
	if tracker == nil {
		return fmt.Errorf("failed to initialize SQLite tracker")
	}

	// Local tracking: ask to resume if progress exists
	existingAnime, err := tracker.GetAnime(animeMalID, currentEpisode.URL)
	if err == nil && existingAnime != nil && existingAnime.PlaybackTime > 0 {
		fmt.Printf("\nProgresso salvo encontrado: episódio %d, tempo %d segundos.\n", existingAnime.EpisodeNumber, existingAnime.PlaybackTime)
		if ok, _ := promptYesNo("Deseja retomar de onde parou?"); ok {
			mpvArgs = append(mpvArgs, fmt.Sprintf("--start=+%d", existingAnime.PlaybackTime))
			if util.IsDebug {
				log.Printf("Retomando do tempo salvo: %d segundos", existingAnime.PlaybackTime)
			}
		}
	}

	err = api.GetAndParseAniSkipData(animeMalID, currentEpisodeNum, currentEpisode)
	if err != nil {
		log.Printf("AniSkip data not available for episode %d: %v\n", currentEpisodeNum, err)
	} else if util.IsDebug {
		log.Printf("AniSkip data for episode %d: %+v\n", currentEpisodeNum, currentEpisode.SkipTimes)
	}

	// Add skip times to mpv arguments if available
	if currentEpisode.SkipTimes.Op.Start > 0 || currentEpisode.SkipTimes.Op.End > 0 {
		opStart, opEnd := currentEpisode.SkipTimes.Op.Start, currentEpisode.SkipTimes.Op.End
		mpvArgs = append(mpvArgs, fmt.Sprintf("--script-opts=skip_op=%d-%d", opStart, opEnd))
	}
	if currentEpisode.SkipTimes.Ed.Start > 0 || currentEpisode.SkipTimes.Ed.End > 0 {
		edStart, edEnd := currentEpisode.SkipTimes.Ed.Start, currentEpisode.SkipTimes.Ed.End
		mpvArgs = append(mpvArgs, fmt.Sprintf("--script-opts=skip_ed=%d-%d", edStart, edEnd))
	}

	// Start mpv with IPC support
	socketPath, err := StartVideo(videoURL, mpvArgs)
	if err != nil {
		return fmt.Errorf("failed to start video with IPC: %w", err)
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
							if updater.episodeDuration < time.Second {
								log.Printf("Warning: Retrieved episode duration is very small (%v). Setting a default duration.", updater.episodeDuration)
								updater.episodeDuration = 24 * time.Minute
							}
							anime := tracking.Anime{
								AnilistID:     animeMalID,
								AllanimeID:    currentEpisode.URL,
								EpisodeNumber: currentEpisodeNum,
								Duration:      int(updater.episodeDuration.Seconds()),
								Title:         getEpisodeTitle(currentEpisode.Title),
								LastUpdated:   time.Now(),
							}
							if err := tracker.UpdateProgress(anime); err != nil {
								log.Printf("Failed to update local tracking: %v", err)
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

	// Locate the index of the current episode
	currentEpisodeIndex := -1
	for i, ep := range episodes {
		if ExtractEpisodeNumber(ep.Number) == strconv.Itoa(currentEpisodeNum) {
			currentEpisodeIndex = i
			break
		}
	}
	if currentEpisodeIndex == -1 {
		return fmt.Errorf("current episode number %d not found", currentEpisodeNum)
	}

	reader := bufio.NewReader(os.Stdin)
	fmt.Println("Press 'n' for next episode, 'p' for previous episode, 'q' to quit, 's' to skip intro:")

	// Start a goroutine to periodically update local tracking
	stopTracking := make(chan struct{})
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
						anime := tracking.Anime{
							AnilistID:     animeMalID,
							AllanimeID:    currentEpisode.URL,
							EpisodeNumber: currentEpisodeNum,
							PlaybackTime:  int(position),
							Duration:      int(updater.episodeDuration.Seconds()),
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
				return playVideo(nextVideoURL, episodes, currentEpisodeNum+1, animeMalID, newUpdater)
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
				return playVideo(prevVideoURL, episodes, currentEpisodeNum-1, animeMalID, newUpdater)
			} else {
				fmt.Println("Already at the first episode.")
			}
		case 'q': // Quit
			fmt.Println("Quitting video playback.")
			close(stopTracking)
			_, _ = mpvSendCommand(socketPath, []interface{}{"quit"})
			return nil
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
