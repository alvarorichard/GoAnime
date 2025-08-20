package playback

import (
	"fmt"
	"log"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/util"
)

func PlayEpisode(
	anime *models.Anime,
	episodes []models.Episode,
	episodeNum int,
	episodeURL string,
	episodeNumberStr string,
	discordEnabled bool,
	isPaused *bool,
	animeMutex *sync.Mutex,
) error {
	animeMutex.Lock()
	anime.Episodes = []models.Episode{{
		Number: episodeNumberStr,
		Num:    episodeNum,
		URL:    episodeURL,
	}}
	animeMutex.Unlock()

	if err := api.GetEpisodeData(anime.MalID, episodeNum, anime); err != nil {
		log.Printf("Error fetching episode data: %v", err)
	}

	// Find the specific episode to pass to enhanced API
	var currentEpisode *models.Episode
	for i := range episodes {
		if episodes[i].Number == episodeNumberStr {
			currentEpisode = &episodes[i]
			break
		}
	}

	if currentEpisode == nil {
		// Create episode if not found
		// For AllAnime, use the anime ID as URL instead of episode-specific URL
		episodeURLForCreation := episodeURL
		if anime.Source == "AllAnime" || (len(anime.URL) < 30 && !strings.Contains(anime.URL, "http")) {
			episodeURLForCreation = anime.URL // Use anime ID for AllAnime
		}

		currentEpisode = &models.Episode{
			Number: episodeNumberStr,
			Num:    episodeNum,
			URL:    episodeURLForCreation,
		}
	}

	// Try enhanced API first, fallback to legacy if needed
	videoURL, err := player.GetVideoURLForEpisodeEnhanced(currentEpisode, anime)
	if err != nil {
		// Bubble up so callers can handle (e.g., prompt to change anime) instead of exiting the app
		return fmt.Errorf("failed to extract video URL: %w", err)
	}

	// Guard against empty or missing durations
	var episodeDuration time.Duration
	if len(episodes) > 0 && episodes[0].Duration > 0 {
		episodeDuration = time.Duration(episodes[0].Duration) * time.Second
	} else {
		episodeDuration = 0
	}
	updater := createUpdater(anime, isPaused, animeMutex, episodeDuration, discordEnabled)

	err = player.HandleDownloadAndPlay(
		videoURL,
		episodes,
		episodeNum,
		anime.URL,
		episodeNumberStr,
		anime.MalID,
		updater,
	)

	if updater != nil {
		updater.Stop()
	}
	return err
}

func SelectEpisodeWithFuzzy(episodes []models.Episode) (string, string, int) {
	url, numStr, err := player.SelectEpisodeWithFuzzyFinder(episodes)
	if err != nil {
		log.Fatalln(util.ErrorHandler(err))
	}
	epNum, err := strconv.Atoi(player.ExtractEpisodeNumber(numStr))
	if err != nil {
		log.Fatalln("Error converting episode number:", util.ErrorHandler(err))
	}
	return url, numStr, epNum
}

func FindEpisodeByNumber(episodes []models.Episode, num int) (string, string, int) {
	for _, ep := range episodes {
		if epNum, err := strconv.Atoi(player.ExtractEpisodeNumber(ep.Number)); err == nil && epNum == num {
			return ep.URL, ep.Number, num
		}
	}
	log.Printf("Warning: Episode number %d not found. Re-selecting.", num)
	return SelectEpisodeWithFuzzy(episodes)
}
