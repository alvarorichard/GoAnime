package api

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Not parallel: the Pomfy client and the playback state are process-wide.

func cleanPomfyPlayback(t *testing.T) {
	t.Helper()
	reset := func() {
		util.ClearGlobalReferer()
		util.ClearGlobalOrigin()
		util.ClearGlobalSubtitles()
		util.SetGlobalAudioLanguage("")
		util.GlobalNoSubs = false
	}
	reset()
	t.Cleanup(reset)
}

func pomfyMovie(world string) (*models.Anime, *models.Episode) {
	return &models.Anime{Name: "Clube da Luta", URL: world + "/assistir/550", Source: "Pomfy", MediaType: models.MediaTypeMovie, TMDBID: 550},
		&models.Episode{Number: "1", Num: 1, DataID: "550", SeasonID: "0", URL: world + "/assistir/550"}
}

func serveMasterAs(t *testing.T, url string) {
	t.Helper()
	old := apiByseMaster
	apiByseMaster = url
	t.Cleanup(func() { apiByseMaster = old })
}

// ---- characterization: what GetPomfyStreamURL has always left for the player ----

func TestGetPomfyStreamURLLeavesTheMetadataContract(t *testing.T) {
	cleanPomfyPlayback(t)
	serveMasterAs(t, "https://edge.test/hls2/lang/por/master.m3u8")
	world := newPomfyWorld(t)
	mountPomfyClient(t, world)
	media, ep := pomfyMovie(world.URL)

	url, err := GetPomfyStreamURL(media, ep, "")

	require.NoError(t, err)
	assert.Equal(t, "https://edge.test/hls2/lang/por/master.m3u8", url)
	assert.Equal(t, world.URL+"/assistir/550", util.GetGlobalReferer(), "mpv sends the watch page as referer")
	assert.Equal(t, "por", util.GetGlobalAudioLanguage(), "the URL spells the audio the file carries")
	assert.Equal(t, []util.SubtitleInfo{{URL: apiByseSubURL, Language: "pt-br", Label: "Português"}}, util.GetGlobalSubtitles())
}

func TestGetPomfyStreamURLHonoursNoSubs(t *testing.T) {
	cleanPomfyPlayback(t)
	util.GlobalNoSubs = true
	world := newPomfyWorld(t)
	mountPomfyClient(t, world)
	media, ep := pomfyMovie(world.URL)

	_, err := GetPomfyStreamURL(media, ep, "")

	require.NoError(t, err)
	assert.Empty(t, util.GetGlobalSubtitles())
	assert.NotEmpty(t, util.GetGlobalReferer(), "only the subtitles are held back")
}

func TestGetPomfyStreamURLWithoutALanguageInTheURLLeavesTheHintAlone(t *testing.T) {
	cleanPomfyPlayback(t)
	util.SetGlobalAudioLanguage("eng")
	world := newPomfyWorld(t)
	mountPomfyClient(t, world)
	media, ep := pomfyMovie(world.URL)

	_, err := GetPomfyStreamURL(media, ep, "")

	require.NoError(t, err)
	assert.Equal(t, "eng", util.GetGlobalAudioLanguage())
}

func TestGetPomfyStreamURLRejectsAMissingTitleOrEpisode(t *testing.T) {
	cleanPomfyPlayback(t)
	for name, c := range map[string]struct {
		m *models.Anime
		e *models.Episode
	}{
		"no media":   {nil, &models.Episode{}},
		"no episode": {&models.Anime{}, nil},
		"no TMDB id": {&models.Anime{Name: "x"}, &models.Episode{}},
	} {
		t.Run(name, func(t *testing.T) {
			url, err := GetPomfyStreamURL(c.m, c.e, "")
			require.Error(t, err)
			assert.Empty(t, url)
			assert.Empty(t, util.GetGlobalReferer(), "a failure publishes nothing")
		})
	}
}

// ---- ResolvePomfyStream: the same resolve, returned instead of published ----

func TestResolvePomfyStreamReturnsTheContextAndPublishesNothing(t *testing.T) {
	cleanPomfyPlayback(t)
	serveMasterAs(t, "https://edge.test/hls2/lang/por/master.m3u8")
	world := newPomfyWorld(t)
	mountPomfyClient(t, world)
	media, ep := pomfyMovie(world.URL)
	util.SetGlobalReferer("https://sentinel.example/")

	got, err := ResolvePomfyStream(media, ep, "")

	require.NoError(t, err)
	assert.Equal(t, source.Stream{
		URL:       "https://edge.test/hls2/lang/por/master.m3u8",
		Referer:   world.URL + "/assistir/550",
		AudioLang: "por",
		Subtitles: []source.Subtitle{{URL: apiByseSubURL, Language: "pt-br", Label: "Português"}},
	}, got)
	assert.Equal(t, "https://sentinel.example/", util.GetGlobalReferer(), "resolving leaves the referer alone")
	assert.Empty(t, util.GetGlobalAudioLanguage())
	assert.Empty(t, util.GetGlobalSubtitles())
}

func TestResolvePomfyStreamThenPublishEqualsGetPomfyStreamURL(t *testing.T) {
	serveMasterAs(t, "https://edge.test/hls2/lang/por/master.m3u8")
	world := newPomfyWorld(t)
	mountPomfyClient(t, world)
	media, ep := pomfyMovie(world.URL)

	cleanPomfyPlayback(t)
	url, err := GetPomfyStreamURL(media, ep, "")
	require.NoError(t, err)
	want := [3]any{util.GetGlobalReferer(), util.GetGlobalAudioLanguage(), util.GetGlobalSubtitles()}

	cleanPomfyPlayback(t)
	stream, err := ResolvePomfyStream(media, ep, "")
	require.NoError(t, err)
	stream.Publish()
	got := [3]any{util.GetGlobalReferer(), util.GetGlobalAudioLanguage(), util.GetGlobalSubtitles()}

	assert.Equal(t, url, stream.URL)
	assert.Equal(t, want, got, "the two routes leave the player the same state")
}

func TestResolvePomfyStreamRejectsAMissingTitleOrEpisode(t *testing.T) {
	_, err := ResolvePomfyStream(nil, &models.Episode{}, "")
	require.Error(t, err)
	_, err = ResolvePomfyStream(&models.Anime{}, nil, "")
	require.Error(t, err)
	_, err = ResolvePomfyStream(&models.Anime{Name: "x"}, &models.Episode{}, "")
	require.Error(t, err)
}
