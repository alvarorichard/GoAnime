package playback

import (
	"log"
	"sync"
	"time"
	"strconv"
	

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
) {
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

	videoURL, err := player.GetVideoURLForEpisode(episodeURL)
	if err != nil {
		log.Fatalln("Failed to extract video URL:", util.ErrorHandler(err))
	}

	episodeDuration := time.Duration(episodes[0].Duration) * time.Second
	updater := createUpdater(anime, isPaused, animeMutex, episodeDuration, discordEnabled)

	player.HandleDownloadAndPlay(
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


