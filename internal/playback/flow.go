package playback

import (
	"context"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/discord"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/player"
	"github.com/alvarorichard/Goanime/v2/internal/tui"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// flowEnv is everything the movie and series loops reach outside the package
// for: the player, the terminal, Discord and each other. realFlow wires the
// real ones; tests wire fakes, so the loops can be driven without mpv or a
// TTY.
type flowEnv struct {
	preWarmMPV         func()
	setEpisodeListCtx  func(name, source, audio string)
	resetTerminal      func()
	streamURL          func(ctx context.Context, ep *models.Episode, anime *models.Anime) (string, error)
	reportStreamFailed func(anime *models.Anime, err error)
	newUpdater         func(anime *models.Anime, isPaused *bool, mu *sync.Mutex, episodeDuration time.Duration, enabled bool) *discord.RichPresenceUpdater
	setExactMediaType  func(mediaType string)
	setMediaMeta       func(meta *util.MediaMeta)
	downloadAndPlay    func(videoURL string, episodes []models.Episode, selectedEpisodeNum int, animeURL, episodeNumberStr string, malID, anilistID int, updater *discord.RichPresenceUpdater, animeName string, season int, anime *models.Anime) error
	userInput          func(isMovie ...bool) string
	playEpisode        func(ctx context.Context, anime *models.Anime, episodes []models.Episode, episodeNum int, episodeURL, episodeNumberStr string, discordEnabled bool, isPaused *bool, animeMutex *sync.Mutex) error
	handleMovie        func(ctx context.Context, anime *models.Anime, episodes []models.Episode, discordEnabled bool) error
	handleSeries       func(ctx context.Context, anime *models.Anime, episodes []models.Episode, totalEpisodes int, discordEnabled bool) error
}

// realFlow is the wiring the program runs with.
func realFlow() flowEnv {
	return flowEnv{
		preWarmMPV:         player.PreWarmMPVPath,
		setEpisodeListCtx:  player.SetEpisodeListContext,
		resetTerminal:      tui.ResetTerminal,
		streamURL:          player.GetVideoURLForEpisodeEnhanced,
		reportStreamFailed: reportMovieStreamFailed,
		newUpdater:         createUpdater,
		setExactMediaType:  player.SetExactMediaType,
		setMediaMeta:       player.SetMediaMeta,
		downloadAndPlay:    player.HandleDownloadAndPlay,
		userInput:          GetUserInput,
		playEpisode:        PlayEpisode,
		handleMovie:        HandleMovie,
		handleSeries:       HandleSeries,
	}
}
