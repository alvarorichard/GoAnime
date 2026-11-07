package providers

import (
	"context"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestStartFlixStreamPublishCascade is the Pomfy cascade for StartFlix: one
// episode from the picked title to the playback state mpv starts with, through
// the real registry and provider, with only the panel resolver stubbed.
func TestStartFlixStreamPublishCascade(t *testing.T) {
	cleanPlayback(t)
	util.SetGlobalReferer("https://last-episode.example/")
	util.SetGlobalOrigin("https://last-episode.example")
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://last-episode.example/old.vtt"}})

	var seen playbackSnapshot
	var resolved bool
	swapStartFlixStream(t, func(*models.Anime, *models.Episode, string) (source.Stream, error) {
		resolved = true
		seen = readPlayback()
		return source.Stream{
			URL:       "https://files.test/ep-1101.mp4",
			Referer:   "https://painel.test/",
			AudioLang: "eng",
			Subtitles: []source.Subtitle{{URL: "https://files.test/pt.vtt", Language: "pt-br", Label: "Português"}},
		}, nil
	})
	anime := &models.Anime{Name: "Show", URL: "https://www.startflix.biz/series/show/", Source: "StartFlix", MediaType: models.MediaTypeTV}
	var url string

	streamStage(t, "1 the registry routes the title to StartFlix", func(t *testing.T) {
		src, _ := source.Resolve(anime)
		require.NotNil(t, src)
		assert.Equal(t, source.StartFlix, src.Describe().Kind)
	})
	streamStage(t, "2 the StartFlix provider resolves through FetchStream", func(t *testing.T) {
		src, _ := source.Resolve(anime)
		_, ok := src.(source.StreamFetcher)
		assert.True(t, ok, "StartFlix no longer returns a Stream")
	})
	streamStage(t, "3 the dispatcher resolves the episode", func(t *testing.T) {
		var err error
		url, err = FetchStreamURL(context.Background(), &models.Episode{Number: "1"}, anime, "best")
		require.NoError(t, err)
		require.True(t, resolved, "the StartFlix resolver was never reached")
	})
	streamStage(t, "4 the last episode's playback state is gone before the resolve", func(t *testing.T) {
		assert.Empty(t, seen.referer)
		assert.Empty(t, seen.origin)
		assert.Empty(t, seen.subs)
	})
	streamStage(t, "5 the stream URL comes back", func(t *testing.T) {
		assert.Equal(t, "https://files.test/ep-1101.mp4", url)
	})
	streamStage(t, "6 the player gets this stream's referer, audio and subtitles", func(t *testing.T) {
		got := readPlayback()
		assert.Equal(t, "https://painel.test/", got.referer, "not the last episode's")
		assert.Equal(t, "eng", got.audio)
		assert.Equal(t, []util.SubtitleInfo{{URL: "https://files.test/pt.vtt", Language: "pt-br", Label: "Português"}}, got.subs)
		assert.Empty(t, got.origin, "the last episode's origin did not survive")
	})
	streamStage(t, "7 the current source is tagged for the player", func(t *testing.T) {
		assert.Equal(t, "StartFlix", util.GetGlobalAnimeSource())
	})
}
