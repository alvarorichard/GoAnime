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

// streamStage runs one hop of a cascade; the cascade stops at the first broken
// hop, so the failure names the hop rather than everything downstream of it.
func streamStage(t *testing.T, name string, fn func(t *testing.T)) {
	t.Helper()
	if !t.Run(name, fn) {
		t.Fatalf("cascade stopped: stage %q failed", name)
	}
}

// TestPomfyStreamPublishCascade follows one Pomfy episode from the title the
// user picked to the playback state mpv is started with, through the real
// registry and the real Pomfy provider, with only the site's resolver stubbed.
// The previous episode left its referer, origin and subtitles behind; none of
// them may reach this one.
func TestPomfyStreamPublishCascade(t *testing.T) {
	cleanPlayback(t)
	util.SetGlobalReferer("https://last-episode.example/")
	util.SetGlobalOrigin("https://last-episode.example")
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://last-episode.example/old.vtt", Language: "en", Label: "old"}})

	var seen playbackSnapshot
	var resolved bool
	swapPomfyStream(t, func(a *models.Anime, _ *models.Episode, q string) (source.Stream, error) {
		resolved = true
		seen = readPlayback()
		assert.Equal(t, "1080p", q)
		return source.Stream{
			URL:       "https://edge.test/hls2/lang/por/master.m3u8",
			Referer:   "https://pomfy.test/assistir/550",
			AudioLang: "por",
			Subtitles: []source.Subtitle{{URL: "https://edge.test/pt-br.vtt", Language: "pt-br", Label: "Português"}},
		}, nil
	})
	anime := &models.Anime{Name: "Clube da Luta", URL: "https://pomfy.test/assistir/550", Source: "Pomfy", MediaType: models.MediaTypeMovie}
	var url string

	streamStage(t, "1 the registry routes the title to Pomfy", func(t *testing.T) {
		src, _ := source.Resolve(anime)
		require.NotNil(t, src)
		assert.Equal(t, source.Pomfy, src.Describe().Kind)
	})
	streamStage(t, "2 the Pomfy provider resolves through FetchStream", func(t *testing.T) {
		src, _ := source.Resolve(anime)
		_, ok := src.(source.StreamFetcher)
		assert.True(t, ok, "Pomfy no longer returns a Stream")
	})
	streamStage(t, "3 the dispatcher resolves the episode", func(t *testing.T) {
		var err error
		url, err = FetchStreamURL(context.Background(), &models.Episode{Number: "1"}, anime, "1080p")
		require.NoError(t, err)
		require.True(t, resolved, "the Pomfy resolver was never reached")
	})
	streamStage(t, "4 the last episode's playback state is gone before the resolve", func(t *testing.T) {
		assert.Empty(t, seen.referer)
		assert.Empty(t, seen.origin)
		assert.Empty(t, seen.subs)
	})
	streamStage(t, "5 the stream URL comes back", func(t *testing.T) {
		assert.Equal(t, "https://edge.test/hls2/lang/por/master.m3u8", url)
	})
	streamStage(t, "6 the player gets this stream's referer, audio and subtitles", func(t *testing.T) {
		got := readPlayback()
		assert.Equal(t, "https://pomfy.test/assistir/550", got.referer, "not the last episode's")
		assert.Equal(t, "por", got.audio)
		assert.Equal(t, []util.SubtitleInfo{{URL: "https://edge.test/pt-br.vtt", Language: "pt-br", Label: "Português"}}, got.subs)
		assert.Empty(t, got.origin, "the last episode's origin did not survive")
	})
	streamStage(t, "7 the current source is tagged for the player", func(t *testing.T) {
		assert.Equal(t, "Pomfy", util.GetGlobalAnimeSource())
	})
}
