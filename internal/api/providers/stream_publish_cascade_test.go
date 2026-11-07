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

// cascadeCase is one provider's end-to-end path: the title, the registry kind
// it must route to, and the swap of its resolver seam.
type cascadeCase struct {
	anime *models.Anime
	kind  source.SourceKind
	tag   string // the source name the player is told
	swap  func(t *testing.T, fn streamFn)
	url   string
}

// runStreamCascade follows one episode from the title the user picked to the
// playback state mpv starts with, through the real registry and the real
// provider, with only the site's resolver stubbed. The previous episode left
// its referer, origin and subtitles behind; none of them may reach this one.
func runStreamCascade(t *testing.T, c cascadeCase) {
	t.Helper()
	cleanPlayback(t)
	util.SetGlobalReferer("https://last-episode.example/")
	util.SetGlobalOrigin("https://last-episode.example")
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://last-episode.example/old.vtt", Language: "en", Label: "old"}})

	var seen playbackSnapshot
	var resolved bool
	c.swap(t, func(_ *models.Anime, _ *models.Episode, q string) (source.Stream, error) {
		resolved = true
		seen = readPlayback()
		assert.Equal(t, "1080p", q)
		return source.Stream{
			URL:       c.url,
			Referer:   "https://site.test/watch/1",
			AudioLang: "por",
			Subtitles: []source.Subtitle{{URL: "https://edge.test/pt-br.vtt", Language: "pt-br", Label: "Português"}},
		}, nil
	})
	var url string

	streamStage(t, "1 the registry routes the title to its provider", func(t *testing.T) {
		src, _ := source.Resolve(c.anime)
		require.NotNil(t, src)
		assert.Equal(t, c.kind, src.Describe().Kind)
	})
	streamStage(t, "2 the provider resolves through FetchStream", func(t *testing.T) {
		src, _ := source.Resolve(c.anime)
		_, ok := src.(source.StreamFetcher)
		assert.True(t, ok, "%s no longer returns a Stream", c.kind)
	})
	streamStage(t, "3 the dispatcher resolves the episode", func(t *testing.T) {
		var err error
		url, err = FetchStreamURL(context.Background(), &models.Episode{Number: "1"}, c.anime, "1080p")
		require.NoError(t, err)
		require.True(t, resolved, "the %s resolver was never reached", c.kind)
	})
	streamStage(t, "4 the last episode's playback state is gone before the resolve", func(t *testing.T) {
		assert.Empty(t, seen.referer)
		assert.Empty(t, seen.origin)
		assert.Empty(t, seen.subs)
	})
	streamStage(t, "5 the stream URL comes back", func(t *testing.T) {
		assert.Equal(t, c.url, url)
	})
	streamStage(t, "6 the player gets this stream's referer, audio and subtitles", func(t *testing.T) {
		got := readPlayback()
		assert.Equal(t, "https://site.test/watch/1", got.referer, "not the last episode's")
		assert.Equal(t, "por", got.audio)
		assert.Equal(t, []util.SubtitleInfo{{URL: "https://edge.test/pt-br.vtt", Language: "pt-br", Label: "Português"}}, got.subs)
		assert.Empty(t, got.origin, "the last episode's origin did not survive")
	})
	streamStage(t, "7 the current source is tagged for the player", func(t *testing.T) {
		assert.Equal(t, c.tag, util.GetGlobalAnimeSource())
	})
}

func TestPomfyStreamPublishCascade(t *testing.T) {
	runStreamCascade(t, cascadeCase{
		anime: &models.Anime{Name: "Clube da Luta", URL: "https://pomfy.test/assistir/550", Source: "Pomfy", MediaType: models.MediaTypeMovie},
		kind:  source.Pomfy, tag: "Pomfy", swap: swapPomfyStream, url: "https://edge.test/hls2/lang/por/master.m3u8",
	})
}

func TestStartFlixStreamPublishCascade(t *testing.T) {
	runStreamCascade(t, cascadeCase{
		anime: &models.Anime{Name: "Show", URL: "https://www.startflix.biz/series/show/", Source: "StartFlix", MediaType: models.MediaTypeTV},
		kind:  source.StartFlix, tag: "StartFlix", swap: swapStartFlixStream, url: "https://files.test/ep-1101.mp4",
	})
}

func TestTopCineStreamPublishCascade(t *testing.T) {
	runStreamCascade(t, cascadeCase{
		anime: &models.Anime{Name: "Todo Poderoso", URL: "https://topcine.test/filme/todo-poderoso/", Source: "TopCine", MediaType: models.MediaTypeMovie},
		kind:  source.TopCine, tag: "TopCine", swap: swapTopCineStream, url: "https://cdn.example/tc.m3u8",
	})
}
