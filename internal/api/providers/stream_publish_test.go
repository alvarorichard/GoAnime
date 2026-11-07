package providers

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The pomfy provider is a StreamFetcher; the registry relies on it.
var _ source.StreamFetcher = (*pomfyProvider)(nil)

// playbackSnapshot is the player-facing state a stream resolve can leave.
type playbackSnapshot struct {
	referer, origin, audio, animeSource string
	subs                                []util.SubtitleInfo
}

func readPlayback() playbackSnapshot {
	return playbackSnapshot{
		referer:     util.GetGlobalReferer(),
		origin:      util.GetGlobalOrigin(),
		audio:       util.GetGlobalAudioLanguage(),
		animeSource: util.GetGlobalAnimeSource(),
		subs:        util.GetGlobalSubtitles(),
	}
}

// cleanPlayback starts from, and leaves, an empty playback state. The state is
// process-wide, so the tests using it are not parallel.
func cleanPlayback(t *testing.T) {
	t.Helper()
	reset := func() {
		util.ClearGlobalReferer()
		util.ClearGlobalOrigin()
		util.ClearGlobalSubtitles()
		util.SetGlobalAudioLanguage("")
		util.SetGlobalAnimeSource("")
		util.GlobalNoSubs = false
	}
	reset()
	t.Cleanup(reset)
}

// fetchStub is a Source that is also a StreamFetcher. It records how it was
// called and what the playback state looked like at that moment.
type fetchStub struct {
	desc   source.Descriptor
	stream source.Stream
	err    error

	calls           int
	gotEpisode      *models.Episode
	gotAnime        *models.Anime
	gotQuality      string
	seenDuringFetch playbackSnapshot
	legacyCalls     int
	legacyURL       string
}

func (s *fetchStub) Describe() source.Descriptor { return s.desc }
func (s *fetchStub) FetchEpisodes(context.Context, *models.Anime) ([]models.Episode, error) {
	return nil, nil
}

func (s *fetchStub) FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error) {
	s.legacyCalls++
	return s.legacyURL, nil
}

func (s *fetchStub) FetchStream(_ context.Context, ep *models.Episode, a *models.Anime, q string) (source.Stream, error) {
	s.calls++
	s.gotEpisode, s.gotAnime, s.gotQuality = ep, a, q
	s.seenDuringFetch = readPlayback()
	return s.stream, s.err
}

// legacyStub is a Source with no FetchStream: it still resolves through
// FetchStreamURL and writes whatever state it likes itself.
type legacyStub struct {
	desc  source.Descriptor
	url   string
	calls int
}

func (s *legacyStub) Describe() source.Descriptor { return s.desc }
func (s *legacyStub) FetchEpisodes(context.Context, *models.Anime) ([]models.Episode, error) {
	return nil, nil
}

func (s *legacyStub) FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error) {
	s.calls++
	util.SetGlobalReferer("https://legacy-wrote.example/")
	return s.url, nil
}

// quietLegacyStub is a legacy Source that publishes no playback context of its
// own, so any referer or origin left at the end is a previous stream's.
type quietLegacyStub struct {
	desc source.Descriptor
	seen playbackSnapshot
}

func (s *quietLegacyStub) Describe() source.Descriptor { return s.desc }
func (s *quietLegacyStub) FetchEpisodes(context.Context, *models.Anime) ([]models.Episode, error) {
	return nil, nil
}

func (s *quietLegacyStub) FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error) {
	s.seen = readPlayback()
	return "https://legacy.example/v.mp4", nil
}

func richStream() source.Stream {
	return source.Stream{
		URL:       "https://cdn.example/master.m3u8",
		Referer:   "https://ref.example/",
		Origin:    "https://origin.example",
		AudioLang: "por",
		Subtitles: []source.Subtitle{{URL: "https://s.example/a.vtt", Language: "pt", Label: "Portuguese"}},
	}
}

func TestFetchAndPublishResolvesThenPublishes(t *testing.T) {
	cleanPlayback(t)
	stub := &fetchStub{stream: richStream()}
	ep := &models.Episode{Number: "3"}
	anime := &models.Anime{Name: "X", Source: "Pomfy"}

	url, err := fetchAndPublish(context.Background(), stub, ep, anime, "1080p")

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/master.m3u8", url)
	assert.Equal(t, 1, stub.calls, "one resolve")
	assert.Same(t, ep, stub.gotEpisode)
	assert.Same(t, anime, stub.gotAnime)
	assert.Equal(t, "1080p", stub.gotQuality)
	got := readPlayback()
	assert.Equal(t, "https://ref.example/", got.referer)
	assert.Equal(t, "https://origin.example", got.origin)
	assert.Equal(t, "por", got.audio)
	assert.Equal(t, []util.SubtitleInfo{{URL: "https://s.example/a.vtt", Language: "pt", Label: "Portuguese"}}, got.subs)
}

func TestFetchAndPublishClearsThePreviousStreamsSubtitlesBeforeResolving(t *testing.T) {
	cleanPlayback(t)
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "from the last episode"}})
	stub := &fetchStub{stream: source.Stream{URL: "u"}} // this stream has no subtitles

	_, err := fetchAndPublish(context.Background(), stub, &models.Episode{}, &models.Anime{}, "")

	require.NoError(t, err)
	assert.Empty(t, stub.seenDuringFetch.subs, "the resolve starts from a clean slate")
	assert.Empty(t, util.GetGlobalSubtitles(), "and an episode without subtitles does not inherit the last one's")
}

func TestFetchAndPublishClearsThePreviousStreamsRefererAndOriginOnAnyEntrance(t *testing.T) {
	cleanPlayback(t)
	util.SetGlobalReferer("https://last-stream.example/")
	util.SetGlobalOrigin("https://last-stream.example")
	stub := &fetchStub{stream: source.Stream{URL: "u"}} // says neither

	_, err := fetchAndPublish(context.Background(), stub, &models.Episode{}, &models.Anime{}, "")

	require.NoError(t, err)
	assert.Empty(t, stub.seenDuringFetch.referer, "the resolve starts without the last stream's referer")
	assert.Empty(t, stub.seenDuringFetch.origin)
	assert.Empty(t, util.GetGlobalReferer(), "and a stream that names none leaves none behind, dispatcher or not")
	assert.Empty(t, util.GetGlobalOrigin())
}

func TestFetchAndPublishTagsTheCurrentSourceBeforeResolving(t *testing.T) {
	cleanPlayback(t)
	stub := &fetchStub{stream: source.Stream{URL: "u"}}

	_, err := fetchAndPublish(context.Background(), stub, &models.Episode{}, &models.Anime{Source: "Pomfy"}, "")

	require.NoError(t, err)
	assert.Equal(t, "Pomfy", stub.seenDuringFetch.animeSource, "the source is known while the stream resolves")
	assert.Equal(t, "Pomfy", util.GetGlobalAnimeSource())
}

func TestFetchAndPublishLeavesTheSourceAloneWhenTheAnimeHasNone(t *testing.T) {
	cleanPlayback(t)
	util.SetGlobalAnimeSource("Kept")
	stub := &fetchStub{stream: source.Stream{URL: "u"}}

	_, err := fetchAndPublish(context.Background(), stub, &models.Episode{}, &models.Anime{}, "")

	require.NoError(t, err)
	assert.Equal(t, "Kept", util.GetGlobalAnimeSource())
}

func TestFetchAndPublishPublishesNothingOnFailure(t *testing.T) {
	cleanPlayback(t)
	boom := errors.New("pomfy is down")
	stub := &fetchStub{stream: richStream(), err: boom} // even a half-filled stream must not leak
	util.SetGlobalReferer("https://previous.example/")

	url, err := fetchAndPublish(context.Background(), stub, &models.Episode{}, &models.Anime{}, "")

	require.ErrorIs(t, err, boom)
	assert.Empty(t, url, "no URL comes back with an error")
	got := readPlayback()
	assert.Empty(t, got.referer, "neither the failed stream's referer nor the previous one's is left: the slate was cleared before the resolve")
	assert.NotEqual(t, "https://ref.example/", got.referer, "the failed stream's referer was not published")
	assert.Empty(t, got.origin)
	assert.Empty(t, got.audio)
	assert.Empty(t, got.subs)
}

func TestFetchAndPublishDoesNothingOnACancelledContext(t *testing.T) {
	cleanPlayback(t)
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "kept"}})
	stub := &fetchStub{stream: richStream()}
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	url, err := fetchAndPublish(ctx, stub, &models.Episode{}, &models.Anime{Source: "Pomfy"}, "")

	require.ErrorIs(t, err, context.Canceled)
	assert.Empty(t, url)
	assert.Zero(t, stub.calls, "a cancelled resolve never starts")
	assert.Len(t, util.GetGlobalSubtitles(), 1, "and does not touch the playback state")
	assert.Empty(t, util.GetGlobalAnimeSource())
}

// ---- dispatch ---------------------------------------------------------------

func dispatchAnime() *models.Anime {
	return &models.Anime{Name: "Show", URL: "https://goyabu.io/show", Source: "Goyabu"}
}

func goyabuDesc() source.Descriptor {
	return source.Descriptor{Kind: source.Goyabu, Priority: 1, URLMatchers: []string{"goyabu"}, Explicit: []string{"Goyabu"}}
}

func TestFetchStreamURLPrefersFetchStreamWhenTheSourceHasIt(t *testing.T) {
	cleanPlayback(t)
	stub := &fetchStub{desc: goyabuDesc(), stream: richStream(), legacyURL: "https://legacy.example/should-not-be-used"}
	t.Cleanup(source.SwapRegistryForTesting(stub))

	url, err := FetchStreamURL(context.Background(), &models.Episode{}, dispatchAnime(), "720p")

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/master.m3u8", url)
	assert.Equal(t, 1, stub.calls)
	assert.Zero(t, stub.legacyCalls, "the legacy method is not called when FetchStream exists")
	assert.Equal(t, "720p", stub.gotQuality)
	assert.Equal(t, "https://ref.example/", util.GetGlobalReferer(), "published once, by the dispatcher")
}

func TestFetchStreamURLStillUsesTheLegacyPathForOtherSources(t *testing.T) {
	cleanPlayback(t)
	stub := &legacyStub{desc: goyabuDesc(), url: "https://legacy.example/v.mp4"}
	t.Cleanup(source.SwapRegistryForTesting(stub))

	url, err := FetchStreamURL(context.Background(), &models.Episode{}, dispatchAnime(), "")

	require.NoError(t, err)
	assert.Equal(t, "https://legacy.example/v.mp4", url)
	assert.Equal(t, 1, stub.calls)
	assert.Equal(t, "https://legacy-wrote.example/", util.GetGlobalReferer(), "a legacy source still publishes for itself")
}

func TestFetchStreamURLClearsTheLastEpisodesRefererAndOrigin(t *testing.T) {
	cleanPlayback(t)
	util.SetGlobalReferer("https://last-episode.example/")
	util.SetGlobalOrigin("https://last-episode.example")
	stub := &fetchStub{desc: goyabuDesc(), stream: source.Stream{URL: "https://cdn.example/x.mp4"}} // says neither
	t.Cleanup(source.SwapRegistryForTesting(stub))

	_, err := FetchStreamURL(context.Background(), &models.Episode{}, dispatchAnime(), "")

	require.NoError(t, err)
	assert.Empty(t, stub.seenDuringFetch.referer, "the resolve starts without the last episode's referer")
	assert.Empty(t, stub.seenDuringFetch.origin)
	assert.Empty(t, util.GetGlobalReferer(), "and a stream that names none leaves none behind")
	assert.Empty(t, util.GetGlobalOrigin())
}

func TestFetchStreamURLClearsTheLastEpisodesRefererAndOriginForLegacySources(t *testing.T) {
	cleanPlayback(t)
	util.SetGlobalReferer("https://last-episode.example/")
	util.SetGlobalOrigin("https://last-episode.example")
	stub := &quietLegacyStub{desc: goyabuDesc()}
	t.Cleanup(source.SwapRegistryForTesting(stub))

	_, err := FetchStreamURL(context.Background(), &models.Episode{}, dispatchAnime(), "")

	require.NoError(t, err)
	assert.Empty(t, stub.seen.referer, "a legacy source resolves without the last episode's referer")
	assert.Empty(t, stub.seen.origin, "or its origin")
}

func TestFetchStreamURLPropagatesAFailedResolve(t *testing.T) {
	cleanPlayback(t)
	boom := errors.New("no stream")
	stub := &fetchStub{desc: goyabuDesc(), stream: richStream(), err: boom}
	t.Cleanup(source.SwapRegistryForTesting(stub))

	url, err := FetchStreamURL(context.Background(), &models.Episode{}, dispatchAnime(), "")

	require.ErrorIs(t, err, boom)
	assert.Empty(t, url)
	assert.Empty(t, util.GetGlobalReferer(), "nothing of the failed stream was published")
}

// ---- the Pomfy provider --------------------------------------------------------

func swapPomfyStream(t *testing.T, fn func(*models.Anime, *models.Episode, string) (source.Stream, error)) {
	t.Helper()
	old := pomfyStreamFn
	pomfyStreamFn = fn
	t.Cleanup(func() { pomfyStreamFn = old })
}

func TestPomfyFetchStreamWritesNoPlaybackState(t *testing.T) {
	cleanPlayback(t)
	util.SetGlobalReferer("https://untouched.example/")
	want := richStream()
	swapPomfyStream(t, func(*models.Anime, *models.Episode, string) (source.Stream, error) { return want, nil })

	got, err := (&pomfyProvider{}).FetchStream(context.Background(), &models.Episode{}, &models.Anime{Source: "Pomfy"}, "")

	require.NoError(t, err)
	assert.Equal(t, want, got, "the whole Stream comes back")
	after := readPlayback()
	assert.Equal(t, "https://untouched.example/", after.referer, "resolving does not publish: that is the dispatcher's job")
	assert.Empty(t, after.origin)
	assert.Empty(t, after.audio)
	assert.Empty(t, after.subs)
	assert.Empty(t, after.animeSource)
}

func TestPomfyFetchStreamPassesTheArgumentsThrough(t *testing.T) {
	cleanPlayback(t)
	ep, anime := &models.Episode{Number: "2"}, &models.Anime{Name: "Pomfy show"}
	var gotEp *models.Episode
	var gotAnime *models.Anime
	var gotQuality string
	swapPomfyStream(t, func(a *models.Anime, e *models.Episode, q string) (source.Stream, error) {
		gotAnime, gotEp, gotQuality = a, e, q
		return source.Stream{URL: "u"}, nil
	})

	_, err := (&pomfyProvider{}).FetchStream(context.Background(), ep, anime, "480p")

	require.NoError(t, err)
	assert.Same(t, ep, gotEp)
	assert.Same(t, anime, gotAnime)
	assert.Equal(t, "480p", gotQuality)
}

func TestPomfyFetchStreamPropagatesErrorsAndCancellation(t *testing.T) {
	cleanPlayback(t)
	boom := errors.New("captcha failed")
	calls := 0
	swapPomfyStream(t, func(*models.Anime, *models.Episode, string) (source.Stream, error) {
		calls++
		return source.Stream{}, boom
	})

	_, err := (&pomfyProvider{}).FetchStream(context.Background(), &models.Episode{}, &models.Anime{}, "")
	require.ErrorIs(t, err, boom)
	assert.Equal(t, 1, calls)

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err = (&pomfyProvider{}).FetchStream(ctx, &models.Episode{}, &models.Anime{}, "")
	require.ErrorIs(t, err, context.Canceled)
	assert.Equal(t, 1, calls, "a cancelled context never reaches the resolver")
}

func TestPomfyFetchStreamURLPublishesLikeItAlwaysDid(t *testing.T) {
	cleanPlayback(t)
	swapPomfyStream(t, func(*models.Anime, *models.Episode, string) (source.Stream, error) { return richStream(), nil })

	url, err := (&pomfyProvider{}).FetchStreamURL(context.Background(), &models.Episode{}, &models.Anime{Source: "Pomfy"}, "")

	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/master.m3u8", url)
	got := readPlayback()
	assert.Equal(t, "https://ref.example/", got.referer)
	assert.Equal(t, "por", got.audio)
	assert.Equal(t, "Pomfy", got.animeSource)
	assert.Len(t, got.subs, 1)
}

// ---- every migrated provider, the same contract --------------------------------

var (
	_ source.StreamFetcher = (*startFlixProvider)(nil)
	_ source.StreamFetcher = (*topCineProvider)(nil)
	_ source.StreamFetcher = (*pobreflixProvider)(nil)
)

type streamFn = func(*models.Anime, *models.Episode, string) (source.Stream, error)

// migratedProviders are the providers that return a Stream. Each entry swaps
// the provider's resolver seam and builds the provider.
var migratedProviders = []struct {
	name string
	swap func(t *testing.T, fn streamFn)
	make func() interface {
		source.StreamFetcher
		FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error)
	}
}{
	{"pomfy", swapPomfyStream, func() interface {
		source.StreamFetcher
		FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error)
	} {
		return &pomfyProvider{}
	}},
	{"startflix", swapStartFlixStream, func() interface {
		source.StreamFetcher
		FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error)
	} {
		return &startFlixProvider{}
	}},
	{"pobreflix", swapPobreflixStream, func() interface {
		source.StreamFetcher
		FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error)
	} {
		return &pobreflixProvider{}
	}},
	{"topcine", swapTopCineStream, func() interface {
		source.StreamFetcher
		FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error)
	} {
		return &topCineProvider{}
	}},
}

func swapPobreflixStream(t *testing.T, fn streamFn) {
	t.Helper()
	old := pobreflixStreamFn
	pobreflixStreamFn = fn
	t.Cleanup(func() { pobreflixStreamFn = old })
}

func swapTopCineStream(t *testing.T, fn streamFn) {
	t.Helper()
	old := topCineStreamFn
	topCineStreamFn = fn
	t.Cleanup(func() { topCineStreamFn = old })
}

func swapStartFlixStream(t *testing.T, fn streamFn) {
	t.Helper()
	old := startFlixStreamFn
	startFlixStreamFn = fn
	t.Cleanup(func() { startFlixStreamFn = old })
}

func TestMigratedProvidersFetchStreamWritesNoPlaybackState(t *testing.T) {
	for _, pc := range migratedProviders {
		t.Run(pc.name, func(t *testing.T) {
			cleanPlayback(t)
			util.SetGlobalReferer("https://untouched.example/")
			want := richStream()
			pc.swap(t, func(*models.Anime, *models.Episode, string) (source.Stream, error) { return want, nil })

			got, err := pc.make().FetchStream(context.Background(), &models.Episode{}, &models.Anime{Source: "X"}, "")

			require.NoError(t, err)
			assert.Equal(t, want, got, "the whole Stream comes back")
			after := readPlayback()
			assert.Equal(t, "https://untouched.example/", after.referer, "resolving does not publish: that is the dispatcher's job")
			assert.Empty(t, after.origin)
			assert.Empty(t, after.audio)
			assert.Empty(t, after.subs)
			assert.Empty(t, after.animeSource)
		})
	}
}

func TestMigratedProvidersFetchStreamPassesTheArgumentsThrough(t *testing.T) {
	for _, pc := range migratedProviders {
		t.Run(pc.name, func(t *testing.T) {
			cleanPlayback(t)
			ep, anime := &models.Episode{Number: "2"}, &models.Anime{Name: "show"}
			var gotEp *models.Episode
			var gotAnime *models.Anime
			var gotQuality string
			pc.swap(t, func(a *models.Anime, e *models.Episode, q string) (source.Stream, error) {
				gotAnime, gotEp, gotQuality = a, e, q
				return source.Stream{URL: "u"}, nil
			})

			_, err := pc.make().FetchStream(context.Background(), ep, anime, "480p")

			require.NoError(t, err)
			assert.Same(t, ep, gotEp)
			assert.Same(t, anime, gotAnime)
			assert.Equal(t, "480p", gotQuality)
		})
	}
}

func TestMigratedProvidersFetchStreamPropagatesErrorsAndCancellation(t *testing.T) {
	for _, pc := range migratedProviders {
		t.Run(pc.name, func(t *testing.T) {
			cleanPlayback(t)
			boom := errors.New("resolver failed")
			calls := 0
			pc.swap(t, func(*models.Anime, *models.Episode, string) (source.Stream, error) {
				calls++
				return source.Stream{}, boom
			})

			_, err := pc.make().FetchStream(context.Background(), &models.Episode{}, &models.Anime{}, "")
			require.ErrorIs(t, err, boom)
			assert.Equal(t, 1, calls)

			ctx, cancel := context.WithCancel(context.Background())
			cancel()
			_, err = pc.make().FetchStream(ctx, &models.Episode{}, &models.Anime{}, "")
			require.ErrorIs(t, err, context.Canceled)
			assert.Equal(t, 1, calls, "a cancelled context never reaches the resolver")
		})
	}
}

func TestMigratedProvidersFetchStreamURLPublishes(t *testing.T) {
	for _, pc := range migratedProviders {
		t.Run(pc.name, func(t *testing.T) {
			cleanPlayback(t)
			pc.swap(t, func(*models.Anime, *models.Episode, string) (source.Stream, error) { return richStream(), nil })

			url, err := pc.make().FetchStreamURL(context.Background(), &models.Episode{}, &models.Anime{Source: "Tagged"}, "")

			require.NoError(t, err)
			assert.Equal(t, "https://cdn.example/master.m3u8", url)
			got := readPlayback()
			assert.Equal(t, "https://ref.example/", got.referer)
			assert.Equal(t, "por", got.audio)
			assert.Equal(t, "Tagged", got.animeSource)
			assert.Len(t, got.subs, 1)
		})
	}
}
