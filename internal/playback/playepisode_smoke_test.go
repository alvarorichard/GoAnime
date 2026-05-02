package playback

import (
	"errors"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func restorePlaybackSeams(t *testing.T) {
	t.Helper()

	origGetEpisodeData := playbackGetEpisodeData
	origGetVideo := playbackGetVideoURLForEpisodeEnhanced
	origHandle := playbackHandleDownloadAndPlay
	origSetExact := playbackSetExactMediaType
	origRunLoading := playbackRunLoadingAction

	t.Cleanup(func() {
		playbackGetEpisodeData = origGetEpisodeData
		playbackGetVideoURLForEpisodeEnhanced = origGetVideo
		playbackHandleDownloadAndPlay = origHandle
		playbackSetExactMediaType = origSetExact
		playbackRunLoadingAction = origRunLoading
	})

	playbackRunLoadingAction = func(action func()) {
		action()
	}
}

func TestPlayEpisodeSmoke(t *testing.T) {
	restorePlaybackSeams(t)

	anime := &models.Anime{
		Name:      "Example",
		URL:       "anime-id",
		MalID:     10,
		MediaType: models.MediaTypeAnime,
		Source:    "AllAnime",
	}
	episodes := []models.Episode{
		{Number: "1", Num: 1, URL: "episode-url", Duration: 1800},
	}

	var (
		gotEpisodeData int
		gotVideoURLFor string
		gotExactType   string
		gotPlayURL     string
		gotPlayEpisode int
	)

	playbackGetEpisodeData = func(malID, episodeNum int, anime *models.Anime) error {
		gotEpisodeData = episodeNum
		return nil
	}
	playbackGetVideoURLForEpisodeEnhanced = func(episode *models.Episode, anime *models.Anime) (string, error) {
		gotVideoURLFor = episode.URL
		return "https://cdn.example/stream.m3u8", nil
	}
	playbackSetExactMediaType = func(mediaType string) {
		gotExactType = mediaType
	}
	playbackHandleDownloadAndPlay = func(
		videoURL string,
		episodes []models.Episode,
		selectedEpisodeNum int,
		animeURL string,
		episodeNumberStr string,
		animeMalID int,
		updater *discord.RichPresenceUpdater,
		animeName string,
		animeSeason int,
	) error {
		gotPlayURL = videoURL
		gotPlayEpisode = selectedEpisodeNum
		assert.Nil(t, updater)
		assert.Equal(t, "Example", animeName)
		assert.Equal(t, "anime-id", animeURL)
		assert.Equal(t, "1", episodeNumberStr)
		assert.Equal(t, 10, animeMalID)
		return nil
	}

	isPaused := false
	var animeMutex sync.Mutex
	err := PlayEpisode(anime, episodes, 1, "episode-url", "1", false, &isPaused, &animeMutex)
	require.NoError(t, err)

	assert.Equal(t, 1, gotEpisodeData)
	assert.Equal(t, "episode-url", gotVideoURLFor)
	assert.Equal(t, string(models.MediaTypeAnime), gotExactType)
	assert.Equal(t, "https://cdn.example/stream.m3u8", gotPlayURL)
	assert.Equal(t, 1, gotPlayEpisode)
	require.Len(t, anime.Episodes, 1)
	assert.Equal(t, "1", anime.Episodes[0].Number)
}

func TestPlayEpisodeReturnsBackToSelectionOnStreamFailure(t *testing.T) {
	restorePlaybackSeams(t)

	anime := &models.Anime{
		Name:      "Example",
		URL:       "https://example.com/anime",
		MalID:     11,
		MediaType: models.MediaTypeAnime,
		Source:    "AnimeFire",
	}
	episodes := []models.Episode{
		{Number: "1", Num: 1, URL: "episode-url"},
	}

	playbackGetEpisodeData = func(malID, episodeNum int, anime *models.Anime) error {
		return nil
	}
	playbackGetVideoURLForEpisodeEnhanced = func(episode *models.Episode, anime *models.Anime) (string, error) {
		return "", errors.New("stream unavailable")
	}
	playbackHandleDownloadAndPlay = func(
		string,
		[]models.Episode,
		int,
		string,
		string,
		int,
		*discord.RichPresenceUpdater,
		string,
		int,
	) error {
		t.Fatal("download/play should not be called when stream resolution fails")
		return nil
	}

	isPaused := false
	var animeMutex sync.Mutex
	err := PlayEpisode(anime, episodes, 1, "episode-url", "1", false, &isPaused, &animeMutex)
	require.Error(t, err)
	assert.ErrorIs(t, err, player.ErrBackToEpisodeSelection)
}

func TestPlayEpisodeCreatesFallbackEpisodeForAllAnime(t *testing.T) {
	restorePlaybackSeams(t)

	anime := &models.Anime{
		Name:      "Example",
		URL:       "allanime-short-id",
		MalID:     12,
		MediaType: models.MediaTypeAnime,
		Source:    "AllAnime",
	}
	episodes := []models.Episode{
		{Number: "2", Num: 2, URL: "episode-two"},
	}

	var gotEpisodeURL string
	playbackGetEpisodeData = func(malID, episodeNum int, anime *models.Anime) error {
		return nil
	}
	playbackGetVideoURLForEpisodeEnhanced = func(episode *models.Episode, anime *models.Anime) (string, error) {
		gotEpisodeURL = episode.URL
		return "https://cdn.example/fallback.m3u8", nil
	}
	playbackSetExactMediaType = func(string) {}
	playbackHandleDownloadAndPlay = func(
		string,
		[]models.Episode,
		int,
		string,
		string,
		int,
		*discord.RichPresenceUpdater,
		string,
		int,
	) error {
		return nil
	}

	isPaused := false
	var animeMutex sync.Mutex
	err := PlayEpisode(anime, episodes, 3, "ignored-url", "3", false, &isPaused, &animeMutex)
	require.NoError(t, err)
	assert.Equal(t, "allanime-short-id", gotEpisodeURL)
}
