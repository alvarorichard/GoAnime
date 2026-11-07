package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pobreflix"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Not parallel: the Pobreflix client seam and the playback state are
// process-wide. Characterization tests: they describe what a direct Pobreflix
// movie has always left for the player, and must pass before and after the
// Stream hand-off.

func cleanPobreflixPlayback(t *testing.T) {
	t.Helper()
	reset := func() {
		util.ClearGlobalReferer()
		util.ClearGlobalOrigin()
		util.ClearGlobalSubtitles()
		util.SetGlobalAudioLanguage("")
		util.GlobalAudioLanguageExplicit = false
		util.GlobalNoSubs = false
	}
	reset()
	t.Cleanup(reset)
}

// directPobreflixMovie serves the dubbed Verity page and returns it with its
// one episode.
func directPobreflixMovie(t *testing.T) (*models.Anime, *models.Episode) {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/filme/verity-dublado.html" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, directMoviePage)
	}))
	t.Cleanup(srv.Close)
	pbDirectClient(t, srv)
	anime := &models.Anime{Name: "Verity", URL: srv.URL + "/filme/verity-dublado.html", Source: pobreflix.SourceName, MediaType: models.MediaTypeMovie}
	eps, err := GetPobreflixEpisodes(anime)
	require.NoError(t, err)
	require.Len(t, eps, 1)
	return anime, &eps[0]
}

func TestGetPobreflixStreamURLDirectMovieClearsTheLastStreamsState(t *testing.T) {
	cleanPobreflixPlayback(t)
	anime, ep := directPobreflixMovie(t)
	util.SetGlobalReferer("https://last-stream.example/")
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://last-stream.example/old.vtt"}})

	url, err := GetPobreflixStreamURL(anime, ep, "best")

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.test/verity.mp4", url)
	assert.Empty(t, util.GetGlobalReferer(), "a public file needs no referer, and the last stream's must not follow it")
	assert.Empty(t, util.GetGlobalSubtitles(), "nor its subtitles")
	assert.Equal(t, portugueseALang, util.GetGlobalAudioLanguage(), "the file is the dub, and mpv is told so")
}

func TestGetPobreflixStreamURLDirectMovieKeepsTheAudioTheUserPinned(t *testing.T) {
	cleanPobreflixPlayback(t)
	anime, ep := directPobreflixMovie(t)
	util.GlobalAudioLanguageExplicit = true
	util.SetGlobalAudioLanguage("eng")

	_, err := GetPobreflixStreamURL(anime, ep, "best")

	require.NoError(t, err)
	assert.Equal(t, "eng", util.GetGlobalAudioLanguage(), "a pinned --audio is not overridden by the file's audio")
}

// ---- ResolvePobreflixStream: the same resolve, returned instead of published ----

func TestResolvePobreflixStreamReturnsTheContextAndPublishesNothing(t *testing.T) {
	cleanPobreflixPlayback(t)
	anime, ep := directPobreflixMovie(t)
	util.SetGlobalReferer("https://sentinel.example/")
	util.SetGlobalAudioLanguage("sentinel")
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "sentinel"}})

	got, err := ResolvePobreflixStream(anime, ep, "best")

	require.NoError(t, err)
	assert.Equal(t, source.Stream{URL: "https://cdn.test/verity.mp4", AudioLang: portugueseALang}, got)
	assert.Equal(t, "https://sentinel.example/", util.GetGlobalReferer(), "resolving leaves the referer alone")
	assert.Equal(t, "sentinel", util.GetGlobalAudioLanguage())
	assert.Equal(t, []util.SubtitleInfo{{URL: "sentinel"}}, util.GetGlobalSubtitles())
}

func TestResolvePobreflixStreamThenPublishEqualsGetPobreflixStreamURL(t *testing.T) {
	cleanPobreflixPlayback(t)
	anime, ep := directPobreflixMovie(t)
	snapshot := func() [3]any {
		return [3]any{util.GetGlobalReferer(), util.GetGlobalAudioLanguage(), util.GetGlobalSubtitles()}
	}

	util.SetGlobalReferer("stale")
	url, err := GetPobreflixStreamURL(anime, ep, "best")
	require.NoError(t, err)
	want := snapshot()

	util.ClearGlobalReferer()
	util.ClearGlobalSubtitles()
	util.SetGlobalAudioLanguage("")
	stream, err := ResolvePobreflixStream(anime, ep, "best")
	require.NoError(t, err)
	stream.Publish()

	assert.Equal(t, url, stream.URL)
	assert.Equal(t, want, snapshot(), "the two routes leave the player the same state")
}

func TestResolvePobreflixStreamKeepsTheAudioTheUserPinned(t *testing.T) {
	cleanPobreflixPlayback(t)
	anime, ep := directPobreflixMovie(t)
	util.GlobalAudioLanguageExplicit = true
	util.SetGlobalAudioLanguage("eng")

	got, err := ResolvePobreflixStream(anime, ep, "best")

	require.NoError(t, err)
	assert.Empty(t, got.AudioLang)
}

func TestResolvePobreflixStreamRejectsAMissingTitleOrEpisode(t *testing.T) {
	_, err := ResolvePobreflixStream(nil, &models.Episode{}, "")
	require.Error(t, err)
	_, err = ResolvePobreflixStream(&models.Anime{}, nil, "")
	require.Error(t, err)
}

func TestPobreflixStreamBuildsTheContext(t *testing.T) {
	cleanPobreflixPlayback(t)
	assert.Equal(t, source.Stream{URL: "u", Referer: "r", AudioLang: portugueseALang}, pobreflixStream("u", pobreflix.AudioDubbed, "r"))
	assert.Equal(t, source.Stream{URL: "u", AudioLang: originalALang}, pobreflixStream("u", pobreflix.AudioSubtitled, ""))

	util.GlobalAudioLanguageExplicit = true
	util.SetGlobalAudioLanguage("eng")
	assert.Equal(t, source.Stream{URL: "u", Referer: "r"}, pobreflixStream("u", pobreflix.AudioDubbed, "r"), "a pinned --audio is not overridden")
}
