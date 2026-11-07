package api

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/startflix"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Not parallel: the fake StartFlix and the playback state are process-wide.

// startFlixShow is the fake series with a referer, a subtitle track and the
// Legendado audio, and its first episode.
func startFlixShow(t *testing.T) (*models.Anime, *models.Episode) {
	t.Helper()
	useFakeStartFlix(t, 0, 1)
	media := &models.Anime{Name: "Show", URL: "https://www.startflix.test/series/show/", Source: "StartFlix", MediaType: models.MediaTypeTV}
	eps, err := GetStartFlixEpisodes(media)
	require.NoError(t, err)
	return media, &eps[0]
}

func TestResolveStartFlixStreamReturnsTheContextAndPublishesNothing(t *testing.T) {
	media, ep := startFlixShow(t)
	util.SetGlobalReferer("https://sentinel.example/")
	util.SetGlobalAudioLanguage("sentinel")
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "sentinel"}})

	got, err := ResolveStartFlixStream(media, ep, "best")

	require.NoError(t, err)
	assert.Equal(t, source.Stream{
		URL:       "https://files.test/ep-1101.mp4",
		Referer:   "https://painel.test/",
		AudioLang: audioLangForStartFlix(startflix.AudioSubtitled),
		Subtitles: []source.Subtitle{{URL: "https://files.test/pt.vtt", Language: "pt-br", Label: "Português"}},
	}, got)
	assert.Equal(t, "https://sentinel.example/", util.GetGlobalReferer(), "resolving leaves the referer alone")
	assert.Equal(t, "sentinel", util.GetGlobalAudioLanguage())
	assert.Equal(t, []util.SubtitleInfo{{URL: "sentinel"}}, util.GetGlobalSubtitles())
}

func TestResolveStartFlixStreamThenPublishEqualsGetStartFlixStreamURL(t *testing.T) {
	media, ep := startFlixShow(t)
	snapshot := func() [3]any {
		return [3]any{util.GetGlobalReferer(), util.GetGlobalAudioLanguage(), util.GetGlobalSubtitles()}
	}

	util.ClearGlobalReferer()
	util.ClearGlobalSubtitles()
	url, err := GetStartFlixStreamURL(media, ep, "best")
	require.NoError(t, err)
	want := snapshot()

	util.ClearGlobalReferer()
	util.ClearGlobalSubtitles()
	stream, err := ResolveStartFlixStream(media, ep, "best")
	require.NoError(t, err)
	stream.Publish()

	assert.Equal(t, url, stream.URL)
	assert.Equal(t, want, snapshot(), "the two routes leave the player the same state")
}

func TestResolveStartFlixStreamKeepsTheAudioTheUserPinned(t *testing.T) {
	media, ep := startFlixShow(t)
	util.GlobalAudioLanguageExplicit = true
	util.SetGlobalAudioLanguage("eng")
	t.Cleanup(func() { util.GlobalAudioLanguageExplicit = false })

	got, err := ResolveStartFlixStream(media, ep, "best")

	require.NoError(t, err)
	assert.Empty(t, got.AudioLang, "a pinned --audio is not overridden by what the stream carries")
	got.Publish()
	assert.Equal(t, "eng", util.GetGlobalAudioLanguage())
}

func TestResolveStartFlixStreamAlwaysTellsMpvTheAudioWhenNothingIsPinned(t *testing.T) {
	media, ep := startFlixShow(t)
	util.GlobalAudioLanguageExplicit = false

	got, err := ResolveStartFlixStream(media, ep, "best")

	require.NoError(t, err)
	assert.Equal(t, audioLangForStartFlix(startflix.AudioSubtitled), got.AudioLang)
}

func TestResolveStartFlixStreamRejectsAMissingTitleOrEpisode(t *testing.T) {
	_, err := ResolveStartFlixStream(nil, &models.Episode{}, "")
	require.Error(t, err)
	_, err = ResolveStartFlixStream(&models.Anime{}, nil, "")
	require.Error(t, err)
}

func TestGetStartFlixStreamURLPublishesNothingOnFailure(t *testing.T) {
	useFakeStartFlix(t, 0, 0)
	util.ClearGlobalReferer()
	media := &models.Anime{Name: "Gone", URL: "https://www.startflix.test/series/gone/", Source: "StartFlix", MediaType: models.MediaTypeTV}

	url, err := GetStartFlixStreamURL(media, &models.Episode{Number: "1", Num: 1, URL: "https://painel.test/episodio/gone"}, "")

	require.Error(t, err)
	assert.Empty(t, url)
	assert.Empty(t, util.GetGlobalReferer(), "a failed resolve leaves no referer behind")
}

// ---- TopCine plays through the StartFlix panel ----------------------------------

func TestResolveTopCineStreamIsTheStartFlixStreamWithoutSideEffects(t *testing.T) {
	media, ep := startFlixShow(t)
	util.SetGlobalReferer("https://sentinel.example/")
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "sentinel"}})

	want, err := ResolveStartFlixStream(media, ep, "best")
	require.NoError(t, err)
	got, err := ResolveTopCineStream(media, ep, "best")

	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, "https://sentinel.example/", util.GetGlobalReferer(), "resolving leaves the referer alone")
	assert.Equal(t, []util.SubtitleInfo{{URL: "sentinel"}}, util.GetGlobalSubtitles())
}

func TestGetTopCineStreamURLStillPublishes(t *testing.T) {
	media, ep := startFlixShow(t)
	util.ClearGlobalReferer()
	util.ClearGlobalSubtitles()

	url, err := GetTopCineStreamURL(media, ep, "best")

	require.NoError(t, err)
	assert.Equal(t, "https://files.test/ep-1101.mp4", url)
	assert.Equal(t, "https://painel.test/", util.GetGlobalReferer())
	assert.Len(t, util.GetGlobalSubtitles(), 1)
	assert.Equal(t, audioLangForStartFlix(startflix.AudioSubtitled), util.GetGlobalAudioLanguage())
}

func TestResolveTopCineStreamRejectsAMissingTitleOrEpisode(t *testing.T) {
	_, err := ResolveTopCineStream(nil, &models.Episode{}, "")
	require.Error(t, err)
	_, err = ResolveTopCineStream(&models.Anime{}, nil, "")
	require.Error(t, err)
}
