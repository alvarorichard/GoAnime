package player

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// These tests swap package seams and process globals; none run in parallel.

// cleanStreamGlobals starts from, and leaves, an empty playback state.
func cleanStreamGlobals(t *testing.T) {
	t.Helper()
	reset := func() {
		util.ClearGlobalReferer()
		util.ClearGlobalOrigin()
		util.ClearGlobalSubtitles()
		util.SetGlobalAudioLanguage("")
		util.SetGlobalAnimeSource("")
	}
	reset()
	t.Cleanup(reset)
}

// withResolver swaps the batch's per-episode resolver for fn.
func withResolver(t *testing.T, fn func(models.Episode, *models.Anime) (source.Stream, error)) {
	t.Helper()
	prev := resolveEpisodeStreamFn
	resolveEpisodeStreamFn = fn
	t.Cleanup(func() { resolveEpisodeStreamFn = prev })
}

func fixedStream(s source.Stream) func(models.Episode, *models.Anime) (source.Stream, error) {
	return func(models.Episode, *models.Anime) (source.Stream, error) { return s, nil }
}

// ---- resolveBatchEpisode ---------------------------------------------------------

func TestResolveBatchEpisodeKeepsTheStreamsContext(t *testing.T) {
	cleanStreamGlobals(t)
	withResolver(t, fixedStream(source.Stream{
		URL: "https://cdn.test/ep1.mp4", Referer: "https://site.test/", Origin: "https://origin.test",
		Subtitles: []source.Subtitle{{URL: "https://s.test/pt.vtt", Language: "pt", Label: "Português"}},
	}))

	be, err := resolveBatchEpisode(models.Episode{Num: 4}, &models.Anime{Source: "StartFlix"})

	require.NoError(t, err)
	assert.Equal(t, 4, be.num)
	assert.Equal(t, "https://cdn.test/ep1.mp4", be.sourceURL)
	assert.Equal(t, "https://cdn.test/ep1.mp4", be.url)
	assert.Equal(t, "https://site.test/", be.referer)
	assert.Equal(t, "https://origin.test", be.origin)
	assert.Equal(t, []util.SubtitleInfo{{URL: "https://s.test/pt.vtt", Language: "pt", Label: "Português"}}, be.subtitles)
}

func TestResolveBatchEpisodeNeverReadsThePlayersGlobals(t *testing.T) {
	cleanStreamGlobals(t)
	// Whatever the player holds from something else must not become this
	// episode's referer, origin or tracks.
	util.SetGlobalReferer("https://something-else.example/")
	util.SetGlobalOrigin("https://something-else.example")
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://something-else.example/x.vtt"}})
	withResolver(t, fixedStream(source.Stream{URL: "https://cdn.test/ep1.mp4"})) // names nothing

	be, err := resolveBatchEpisode(models.Episode{Num: 1}, &models.Anime{Source: "StartFlix"})

	require.NoError(t, err)
	assert.Empty(t, be.referer, "an episode whose resolver named no referer has none")
	assert.Empty(t, be.origin)
	assert.Empty(t, be.subtitles)
}

func TestResolveBatchEpisodeLeavesThePlayingRefererOriginAndAudioAlone(t *testing.T) {
	cleanStreamGlobals(t)
	util.SetGlobalReferer("https://playing-now.example/")
	util.SetGlobalOrigin("https://playing-now.example")
	util.SetGlobalAudioLanguage("eng")
	withResolver(t, fixedStream(source.Stream{
		URL: "https://cdn.test/ep1.mp4", Referer: "https://other.test/", Origin: "https://other.test", AudioLang: "por",
	}))

	_, err := resolveBatchEpisode(models.Episode{Num: 1}, &models.Anime{Source: "StartFlix"})

	require.NoError(t, err)
	assert.Equal(t, "https://playing-now.example/", util.GetGlobalReferer(), "resolving a download does not replace the playing episode's referer")
	assert.Equal(t, "https://playing-now.example", util.GetGlobalOrigin())
	assert.Equal(t, "eng", util.GetGlobalAudioLanguage())
	// The subtitles are not on this list: a batch still clears the global
	// tracks (see TestResolveBatchEpisodeClearsTheTracksAnEarlierEpisodeLeft)
	// until the player stops reading them from there.
}

func TestResolveBatchEpisodeTagsTheSource(t *testing.T) {
	cleanStreamGlobals(t)
	withResolver(t, fixedStream(source.Stream{URL: "https://cdn.test/ep1.mp4"}))

	_, err := resolveBatchEpisode(models.Episode{Num: 1}, &models.Anime{Source: "StartFlix"})
	require.NoError(t, err)
	assert.Equal(t, "StartFlix", util.GetGlobalAnimeSource(), "what is watched afterwards is tracked under this source")

	_, err = resolveBatchEpisode(models.Episode{Num: 2}, &models.Anime{})
	require.NoError(t, err)
	assert.Equal(t, "StartFlix", util.GetGlobalAnimeSource(), "an anime with no source does not clear the tag")

	_, err = resolveBatchEpisode(models.Episode{Num: 3}, nil)
	require.NoError(t, err)
	assert.Equal(t, "StartFlix", util.GetGlobalAnimeSource(), "nor does a missing one")
}

func TestResolveBatchEpisodeClearsTheTracksAnEarlierEpisodeLeft(t *testing.T) {
	cleanStreamGlobals(t)
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://earlier.example/old.vtt"}})
	withResolver(t, fixedStream(source.Stream{URL: "https://cdn.test/ep1.mp4"}))

	_, err := resolveBatchEpisode(models.Episode{Num: 1}, &models.Anime{Source: "StartFlix"})

	require.NoError(t, err)
	assert.Empty(t, util.GetGlobalSubtitles(), "a batch does not leave an earlier episode's online tracks for a downloaded file")
}

func TestResolveBatchEpisodeKeepsTheUsersLanguageChoiceFor9Anime(t *testing.T) {
	cleanStreamGlobals(t)
	util.SetGlobalAnimeSource("9Anime")
	chosen := []util.SubtitleInfo{{URL: "https://9anime.test/en.vtt", Language: "en"}}
	util.SetGlobalSubtitles(chosen)
	withResolver(t, fixedStream(source.Stream{
		URL: "https://cdn.test/ep1.mp4", Subtitles: []source.Subtitle{{URL: "https://s.test/other.vtt"}},
	}))

	be, err := resolveBatchEpisode(models.Episode{Num: 1}, &models.Anime{Source: "9Anime"})

	require.NoError(t, err)
	assert.Nil(t, be.subtitles, "9Anime's tracks are the user's choice, held in the global")
	assert.Equal(t, chosen, util.GetGlobalSubtitles(), "and that choice is not cleared")
}

func TestResolveBatchEpisodeFailures(t *testing.T) {
	cleanStreamGlobals(t)
	boom := errors.New("offline")

	t.Run("the resolver fails", func(t *testing.T) {
		withResolver(t, func(models.Episode, *models.Anime) (source.Stream, error) { return source.Stream{}, boom })
		_, err := resolveBatchEpisode(models.Episode{Num: 1}, &models.Anime{})
		require.ErrorIs(t, err, boom)
		assert.ErrorContains(t, err, "failed to resolve stream")
	})
	t.Run("the resolver returns no URL", func(t *testing.T) {
		withResolver(t, fixedStream(source.Stream{Referer: "https://site.test/"}))
		_, err := resolveBatchEpisode(models.Episode{Num: 1}, &models.Anime{})
		require.EqualError(t, err, "failed to resolve stream: empty stream URL")
	})
}

// ---- the per-download request context ---------------------------------------------

func TestDownloadContextPinsAreIndependentOfTheGlobals(t *testing.T) {
	cleanStreamGlobals(t)
	util.SetGlobalReferer("https://global.example/")
	util.SetGlobalOrigin("https://global.example")
	const u = "https://cdn.test/ep1.m3u8"

	t.Run("unpinned follows the globals", func(t *testing.T) {
		assert.Equal(t, "https://global.example/", downloadReferer(u))
		assert.Equal(t, "https://global.example", downloadOrigin(u))
	})
	t.Run("a pinned context wins, empty origin included", func(t *testing.T) {
		forget := rememberDownloadContext(u, "https://pinned.test/", "")
		assert.Equal(t, "https://pinned.test/", downloadReferer(u))
		assert.Empty(t, downloadOrigin(u), "the resolver gave no origin, so the global one is not stood in for it")
		forget()
		assert.Equal(t, "https://global.example/", downloadReferer(u), "forgetting hands the URL back to the globals")
		assert.Equal(t, "https://global.example", downloadOrigin(u))
	})
	t.Run("a pinned origin is used", func(t *testing.T) {
		defer rememberDownloadContext(u, "https://pinned.test/", "https://pinned-origin.test")()
		assert.Equal(t, "https://pinned-origin.test", downloadOrigin(u))
	})
	t.Run("pinning only the referer leaves the origin on the global", func(t *testing.T) {
		defer rememberDownloadReferer(u, "https://pinned.test/")()
		assert.Equal(t, "https://pinned.test/", downloadReferer(u))
		assert.Equal(t, "https://global.example", downloadOrigin(u))
	})
	t.Run("pins are per URL", func(t *testing.T) {
		defer rememberDownloadContext(u, "https://pinned.test/", "https://pinned-origin.test")()
		assert.Equal(t, "https://global.example/", downloadReferer("https://cdn.test/other.m3u8"))
		assert.Equal(t, "https://global.example", downloadOrigin("https://cdn.test/other.m3u8"))
	})
}

func TestBatchEpisodesKeepTheirOwnOriginInParallel(t *testing.T) {
	restore := installDownloadRangeTestState(t.TempDir())
	defer restore()
	cleanStreamGlobals(t)
	home := t.TempDir()
	streams := map[int]source.Stream{
		1: {URL: "https://a.cdn.test/ep1/master.m3u8", Referer: "https://a.test/", Origin: "https://origin-a.test"},
		2: {URL: "https://b.cdn.test/ep2/master.m3u8", Referer: "https://b.test/", Origin: "https://origin-b.test"},
		3: {URL: "https://c.cdn.test/ep3/master.m3u8", Referer: "https://c.test/"}, // no origin of its own
	}
	withResolver(t, func(ep models.Episode, _ *models.Anime) (source.Stream, error) { return streams[ep.Num], nil })
	util.SetGlobalOrigin("https://stale-global.example") // must not stand in for episode 3's missing origin

	var mu sync.Mutex
	gotOrigin := map[string]string{}
	gotReferer := map[string]string{}
	useHLSChainMocks(t, func(u, p string, _ *model) error {
		mu.Lock()
		gotOrigin[u], gotReferer[u] = downloadOrigin(u), downloadReferer(u)
		mu.Unlock()
		return writeVideo(p)
	}, ok, ok)

	var eps []*batchEpisode
	for n := 1; n <= 3; n++ {
		be, err := resolveBatchEpisode(models.Episode{Num: n}, &models.Anime{Source: "HiAnime"})
		require.NoError(t, err)
		be.num, be.path = n, filepath.Join(home, "Show", "ep"+string(rune('0'+n))+".mp4")
		eps = append(eps, &be)
	}
	var wg sync.WaitGroup
	errs := make([]error, len(eps))
	for i, be := range eps {
		wg.Go(func() { errs[i] = downloadBatchEpisode(be, &models.Anime{Source: "HiAnime"}, nil, nil) })
	}
	wg.Wait()
	for i, err := range errs {
		require.NoError(t, err, "episode %d", i+1)
	}

	for n, s := range streams {
		assert.Equal(t, s.Origin, gotOrigin[s.URL], "episode %d sent its own origin", n)
		assert.Equal(t, s.Referer, gotReferer[s.URL], "episode %d sent its own referer", n)
	}
}

// ---- getBestQualityStream: the production resolver through the real registry -----

type streamSourceStub struct {
	source.Descriptor
	stream source.Stream
	err    error

	gotEpisode *models.Episode
	gotAnime   *models.Anime
	gotQuality string
}

func (s *streamSourceStub) Describe() source.Descriptor { return s.Descriptor }
func (s *streamSourceStub) FetchEpisodes(context.Context, *models.Anime) ([]models.Episode, error) {
	return nil, nil
}

func (s *streamSourceStub) FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error) {
	return "", errors.New("the legacy method must not be used")
}

func (s *streamSourceStub) FetchStream(_ context.Context, ep *models.Episode, a *models.Anime, q string) (source.Stream, error) {
	s.gotEpisode, s.gotAnime, s.gotQuality = ep, a, q
	return s.stream, s.err
}

func useStreamSource(t *testing.T, stub *streamSourceStub) *models.Anime {
	t.Helper()
	stub.Descriptor = source.Descriptor{Kind: source.Goyabu, Priority: 1, URLMatchers: []string{"goyabu"}, Explicit: []string{"Goyabu"}}
	t.Cleanup(source.SwapRegistryForTesting(stub))
	return &models.Anime{Name: "Show", URL: "https://goyabu.io/show", Source: "Goyabu"}
}

func TestGetBestQualityStreamReturnsTheSourcesStreamWithoutPublishingIt(t *testing.T) {
	cleanStreamGlobals(t)
	util.SetGlobalReferer("https://playing-now.example/")
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://playing-now.example/en.vtt"}})
	want := source.Stream{
		URL: "https://cdn.test/v.m3u8", Referer: "https://r.test/", Origin: "https://o.test", AudioLang: "por",
		Subtitles: []source.Subtitle{{URL: "https://s.test/pt.vtt", Language: "pt", Label: "P"}},
	}
	stub := &streamSourceStub{stream: want}
	anime := useStreamSource(t, stub)
	prev := util.GlobalQuality
	util.GlobalQuality = "720p"
	t.Cleanup(func() { util.GlobalQuality = prev })

	got, err := getBestQualityStream(models.Episode{Number: "3", Num: 3, URL: "e3", DataID: "d", SeasonID: "2"}, anime)

	require.NoError(t, err)
	assert.Equal(t, want, got)
	assert.Equal(t, "720p", stub.gotQuality, "the configured quality reaches the source")
	assert.Same(t, anime, stub.gotAnime)
	assert.Equal(t, &models.Episode{Number: "3", Num: 3, URL: "e3", DataID: "d", SeasonID: "2"}, stub.gotEpisode)
	assert.Equal(t, "https://playing-now.example/", util.GetGlobalReferer(), "the playing episode's referer is untouched")
	assert.Len(t, util.GetGlobalSubtitles(), 1, "and so are its subtitles")
	assert.Empty(t, util.GetGlobalOrigin())
}

func TestGetBestQualityStreamNumbersAnEpisodeThatHasOnlyNum(t *testing.T) {
	cleanStreamGlobals(t)
	stub := &streamSourceStub{stream: source.Stream{URL: "u"}}
	anime := useStreamSource(t, stub)

	_, err := getBestQualityStream(models.Episode{Num: 7}, anime)

	require.NoError(t, err)
	assert.Equal(t, "7", stub.gotEpisode.Number)
}

func TestGetBestQualityStreamNamesTheSourceAndEpisodeOnFailure(t *testing.T) {
	cleanStreamGlobals(t)
	boom := errors.New("down")
	stub := &streamSourceStub{err: boom}
	anime := useStreamSource(t, stub)

	_, err := getBestQualityStream(models.Episode{Number: "5"}, anime)
	require.ErrorIs(t, err, boom)
	assert.ErrorContains(t, err, "failed to get stream URL from Goyabu for episode 5")

	stub.err, stub.stream = nil, source.Stream{Referer: "https://r.test/"}
	_, err = getBestQualityStream(models.Episode{Number: "5"}, anime)
	require.EqualError(t, err, "empty stream URL from Goyabu for episode 5")
}

// ---- the headers native HLS really sends -------------------------------------------

// headerRecorder is a one-segment HLS CDN that records the Referer and Origin
// of every request it gets.
func headerRecorder(t *testing.T) (*httptest.Server, func() (referers, origins []string)) {
	t.Helper()
	var mu sync.Mutex
	var refs, origs []string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		mu.Lock()
		refs, origs = append(refs, r.Header.Get("Referer")), append(origs, r.Header.Get("Origin"))
		mu.Unlock()
		if r.URL.Path == "/normal.m3u8" {
			fmt.Fprint(w, "#EXTM3U\n#EXT-X-TARGETDURATION:5\n#EXTINF:5.0,\nseg.ts\n#EXT-X-ENDLIST\n")
			return
		}
		_, _ = w.Write(make([]byte, 2048))
	}))
	t.Cleanup(srv.Close)
	return srv, func() ([]string, []string) {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), refs...), append([]string(nil), origs...)
	}
}

func uniq(values []string) []string {
	seen := map[string]bool{}
	var out []string
	for _, v := range values {
		if !seen[v] {
			seen[v] = true
			out = append(out, v)
		}
	}
	return out
}

func TestNativeHLSSendsThePinnedOrigin(t *testing.T) {
	cleanStreamGlobals(t)
	util.SetGlobalReferer("https://global.example/")
	util.SetGlobalOrigin("https://global.example")

	cases := []struct {
		name                  string
		pin                   func(url string) func()
		wantReferer, wantOrig string
	}{
		{"a pinned origin is sent as is", func(u string) func() {
			return rememberDownloadContext(u, "https://site.test/watch/", "https://stated-origin.test")
		}, "https://site.test/watch/", "https://stated-origin.test"},
		{"a pinned referer with no origin derives it from the referer, not the global", func(u string) func() {
			return rememberDownloadContext(u, "https://site.test/watch/", "")
		}, "https://site.test/watch/", "https://site.test/watch"},
		{"a referer-only pin keeps following the global origin", func(u string) func() {
			return rememberDownloadReferer(u, "https://site.test/watch/")
		}, "https://site.test/watch/", "https://global.example"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			srv, seen := headerRecorder(t)
			streamURL := srv.URL + "/normal.m3u8"
			defer tc.pin(streamURL)()

			require.NoError(t, downloadWithNativeHLS(streamURL, homePath(t, "ep.ts"), &model{}))

			refs, origs := seen()
			assert.Equal(t, []string{tc.wantReferer}, uniq(refs), "every request carries the pinned referer")
			assert.Equal(t, []string{tc.wantOrig}, uniq(origs), "and the origin that goes with it")
		})
	}
}

func TestStreamSubtitlesKeepsEveryTrack(t *testing.T) {
	assert.Nil(t, streamSubtitles(source.Stream{}), "a stream with no tracks has none")
	got := streamSubtitles(source.Stream{Subtitles: []source.Subtitle{
		{URL: "1", Language: "pt", Label: "P"}, {URL: "2", Language: "en", Label: "E"}, {URL: "3", Language: "es", Label: "S"},
	}})
	assert.Equal(t, []util.SubtitleInfo{
		{URL: "1", Language: "pt", Label: "P"}, {URL: "2", Language: "en", Label: "E"}, {URL: "3", Language: "es", Label: "S"},
	}, got, "every track, in order, each field in its own place")
}
