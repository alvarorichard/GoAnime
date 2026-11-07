package providers

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper"
	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// adapterStub is a scraper adapter that returns what it is told and records
// how it was called. It has no context-aware method, so providers take their
// plain GetStreamURL path.
type adapterStub struct {
	url      string
	metadata map[string]string
	err      error

	calls      int
	gotEpisode string
	gotOptions []any
}

func (a *adapterStub) SearchAnime(string, ...any) ([]*models.Anime, error) { return nil, nil }
func (a *adapterStub) GetAnimeEpisodes(string) ([]models.Episode, error)   { return nil, nil }
func (a *adapterStub) GetType() scraper.ScraperType                        { return scraper.AnimeDriveType }

func (a *adapterStub) GetStreamURL(episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	a.calls++
	a.gotEpisode, a.gotOptions = episodeURL, options
	return a.url, a.metadata, a.err
}

// ctxAdapterStub is an adapterStub that also speaks the context-aware
// interface; providers that can must prefer it and pass the caller's context.
type ctxAdapterStub struct {
	adapterStub
	ctxCalls int
	gotCtx   context.Context
}

func (a *ctxAdapterStub) SearchAnimeContext(context.Context, string, ...any) ([]*models.Anime, error) {
	return nil, nil
}

func (a *ctxAdapterStub) GetAnimeEpisodesContext(context.Context, string) ([]models.Episode, error) {
	return nil, nil
}

func (a *ctxAdapterStub) GetStreamURLContext(ctx context.Context, episodeURL string, options ...any) (streamURL string, metadata map[string]string, err error) {
	a.ctxCalls++
	a.gotCtx = ctx
	a.gotEpisode, a.gotOptions = episodeURL, options
	return a.url, a.metadata, a.err
}

type streamProvider interface {
	source.StreamFetcher
	FetchStreamURL(context.Context, *models.Episode, *models.Anime, string) (string, error)
}

// adapterProviders are the providers that resolve through a scraper adapter.
// build wires the stub in as the provider's adapter.
var adapterProviders = []struct {
	name        string
	build       func(a scraper.UnifiedScraper) streamProvider
	metadata    bool   // whether the provider reads the scraper's playback metadata
	errPrefix   string // what a scraper failure is wrapped in
	emptyReason string // the error for an empty URL
	quality     bool   // whether the quality reaches the adapter
}{
	{"hianime", func(a scraper.UnifiedScraper) streamProvider {
		p := &hianimeProvider{}
		p.adapter.UnifiedScraper = a
		p.once.Do(func() {})
		return p
	}, true, "hianime stream: ", "empty stream URL returned from HiAnime", true},
	{"anroll", func(a scraper.UnifiedScraper) streamProvider {
		p := &anrollProvider{}
		p.adapter.UnifiedScraper = a
		p.once.Do(func() {})
		return p
	}, true, "anroll stream: ", "empty stream URL returned from AnimesROLL", false},
	{"animedrive", func(a scraper.UnifiedScraper) streamProvider {
		p := &animeDriveProvider{}
		p.adapter.UnifiedScraper = a
		p.once.Do(func() {})
		return p
	}, true, "AnimeDrive stream: ", "empty stream URL returned from AnimeDrive", true},
	{"animefire", func(a scraper.UnifiedScraper) streamProvider {
		p := &animeFireProvider{}
		p.adapter.UnifiedScraper = a
		p.once.Do(func() {})
		return p
	}, false, "animeFire stream: ", "empty stream URL returned from Animefire.io", true},
	{"goyabu", func(a scraper.UnifiedScraper) streamProvider {
		p := &goyabuProvider{}
		p.adapter.UnifiedScraper = a
		p.once.Do(func() {})
		return p
	}, false, "goyabu stream: ", "empty stream URL returned from Goyabu", false},
	{"animesdigital", func(a scraper.UnifiedScraper) streamProvider {
		p := &animesDigitalProvider{}
		p.adapter.UnifiedScraper = a
		p.once.Do(func() {})
		return p
	}, false, "animesdigital stream: ", "empty stream URL returned from AnimesDigital", false},
}

const metadataSubs = `[{"url":"https://s.example/pt.vtt","language":"pt","label":"Portuguese"},{"url":"https://s.example/en.vtt","language":"en","label":"English"}]`

func fullMetadata() map[string]string {
	return map[string]string{"referer": "https://ref.example/", "origin": "https://origin.example", "subtitles": metadataSubs}
}

func TestAdapterProvidersFetchStreamReturnsTheContextAndWritesNone(t *testing.T) {
	for _, pc := range adapterProviders {
		t.Run(pc.name, func(t *testing.T) {
			cleanPlayback(t)
			util.SetGlobalReferer("https://untouched.example/")
			stub := &adapterStub{url: "https://cdn.example/v.m3u8", metadata: fullMetadata()}

			got, err := pc.build(stub).FetchStream(context.Background(), &models.Episode{URL: "https://site.example/ep/1"}, &models.Anime{Source: "X"}, "720p")

			require.NoError(t, err)
			want := source.Stream{URL: "https://cdn.example/v.m3u8"}
			if pc.metadata {
				want.Referer, want.Origin = "https://ref.example/", "https://origin.example"
				want.Subtitles = []source.Subtitle{
					{URL: "https://s.example/pt.vtt", Language: "pt", Label: "Portuguese"},
					{URL: "https://s.example/en.vtt", Language: "en", Label: "English"},
				}
			}
			assert.Equal(t, want, got, "a provider that never read the scraper's metadata still does not")
			assert.Equal(t, "https://site.example/ep/1", stub.gotEpisode)
			after := readPlayback()
			assert.Equal(t, "https://untouched.example/", after.referer, "resolving does not publish")
			assert.Empty(t, after.origin)
			assert.Empty(t, after.subs)
			assert.Empty(t, after.animeSource)
		})
	}
}

func TestAdapterProvidersPassTheQualityOnlyWhereTheyAlwaysDid(t *testing.T) {
	for _, pc := range adapterProviders {
		t.Run(pc.name, func(t *testing.T) {
			stub := &adapterStub{url: "u"}
			_, err := pc.build(stub).FetchStream(context.Background(), &models.Episode{URL: "e"}, &models.Anime{}, "480p")
			require.NoError(t, err)
			if pc.quality {
				assert.Equal(t, []any{"480p"}, stub.gotOptions)
			} else {
				assert.Empty(t, stub.gotOptions)
			}
		})
	}
}

func TestAdapterProvidersWrapAScraperFailureAndKeepItInspectable(t *testing.T) {
	for _, pc := range adapterProviders {
		t.Run(pc.name, func(t *testing.T) {
			cleanPlayback(t)
			boom := errors.New("scraper down")
			stub := &adapterStub{url: "ignored", metadata: fullMetadata(), err: boom}

			got, err := pc.build(stub).FetchStream(context.Background(), &models.Episode{}, &models.Anime{}, "")

			require.ErrorIs(t, err, boom)
			assert.ErrorContains(t, err, pc.errPrefix)
			assert.Equal(t, source.Stream{}, got, "a failed resolve returns no half-filled stream")
		})
	}
}

func TestAdapterProvidersRejectAnEmptyURLWithoutPublishingItsContext(t *testing.T) {
	for _, pc := range adapterProviders {
		t.Run(pc.name, func(t *testing.T) {
			cleanPlayback(t)
			stub := &adapterStub{url: "", metadata: fullMetadata()}
			p := pc.build(stub)

			got, err := p.FetchStream(context.Background(), &models.Episode{}, &models.Anime{}, "")
			require.EqualError(t, err, pc.emptyReason)
			assert.Equal(t, source.Stream{}, got)

			url, err := p.FetchStreamURL(context.Background(), &models.Episode{}, &models.Anime{}, "")
			require.EqualError(t, err, pc.emptyReason)
			assert.Empty(t, url)
			after := readPlayback()
			assert.Empty(t, after.referer, "the metadata of a stream that did not resolve is not left behind")
			assert.Empty(t, after.origin)
			assert.Empty(t, after.subs)
		})
	}
}

func TestAdapterProvidersNeverReachTheAdapterOnACancelledContext(t *testing.T) {
	for _, pc := range adapterProviders {
		t.Run(pc.name, func(t *testing.T) {
			stub := &ctxAdapterStub{}
			ctx, cancel := context.WithCancel(context.Background())
			cancel()

			_, err := pc.build(stub).FetchStream(ctx, &models.Episode{}, &models.Anime{}, "")

			require.ErrorIs(t, err, context.Canceled)
			assert.Zero(t, stub.calls+stub.ctxCalls)
		})
	}
}

func TestAdapterProvidersFetchStreamURLPublishesTheContext(t *testing.T) {
	for _, pc := range adapterProviders {
		t.Run(pc.name, func(t *testing.T) {
			cleanPlayback(t)
			util.SetGlobalReferer("https://last.example/")
			util.SetGlobalOrigin("https://last.example")
			stub := &adapterStub{url: "https://cdn.example/v.m3u8", metadata: fullMetadata()}

			url, err := pc.build(stub).FetchStreamURL(context.Background(), &models.Episode{URL: "e"}, &models.Anime{Source: "Tagged"}, "")

			require.NoError(t, err)
			assert.Equal(t, "https://cdn.example/v.m3u8", url)
			got := readPlayback()
			assert.Equal(t, "Tagged", got.animeSource)
			if pc.metadata {
				assert.Equal(t, "https://ref.example/", got.referer)
				assert.Equal(t, "https://origin.example", got.origin)
				assert.Equal(t, []util.SubtitleInfo{
					{URL: "https://s.example/pt.vtt", Language: "pt", Label: "Portuguese"},
					{URL: "https://s.example/en.vtt", Language: "en", Label: "English"},
				}, got.subs)
			} else {
				assert.Empty(t, got.referer, "this provider has never published the scraper's metadata")
				assert.Empty(t, got.origin)
				assert.Empty(t, got.subs)
			}
		})
	}
}

func TestAdapterProvidersFetchStreamURLReplacesTheLastEpisodesContext(t *testing.T) {
	for _, pc := range adapterProviders {
		t.Run(pc.name, func(t *testing.T) {
			cleanPlayback(t)
			util.SetGlobalReferer("https://last.example/")
			util.SetGlobalOrigin("https://last.example")
			util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "https://last.example/old.vtt"}})
			stub := &adapterStub{url: "u"} // no metadata at all

			_, err := pc.build(stub).FetchStreamURL(context.Background(), &models.Episode{}, &models.Anime{}, "")

			require.NoError(t, err)
			got := readPlayback()
			assert.Empty(t, got.referer, "an episode that names no referer does not inherit the last one's")
			assert.Empty(t, got.origin)
			assert.Empty(t, got.subs)
		})
	}
}

func TestContextAwareAdaptersAreUsedWithTheCallersContext(t *testing.T) {
	type ctxKey struct{}
	for _, name := range []string{"hianime", "anroll", "animesdigital"} {
		for _, pc := range adapterProviders {
			if pc.name != name {
				continue
			}
			t.Run(pc.name, func(t *testing.T) {
				stub := &ctxAdapterStub{adapterStub: adapterStub{url: "u", metadata: map[string]string{"referer": "r"}}}
				ctx := context.WithValue(context.Background(), ctxKey{}, "marker")

				got, err := pc.build(stub).FetchStream(ctx, &models.Episode{URL: "e"}, &models.Anime{}, "best")

				require.NoError(t, err)
				assert.Equal(t, 1, stub.ctxCalls, "the context-aware method is preferred")
				assert.Zero(t, stub.calls, "the plain one is not also called")
				assert.Equal(t, "marker", stub.gotCtx.Value(ctxKey{}), "the caller's context reaches the scraper")
				assert.Equal(t, "u", got.URL)
				if pc.metadata {
					assert.Equal(t, "r", got.Referer)
				} else {
					assert.Empty(t, got.Referer)
				}
			})
		}
	}
}

func TestAnimeFireDefaultsTheQualityToBest(t *testing.T) {
	stub := &adapterStub{url: "u"}
	p := &animeFireProvider{}
	p.adapter.UnifiedScraper = stub
	p.once.Do(func() {})

	_, err := p.FetchStream(context.Background(), &models.Episode{URL: "e"}, &models.Anime{}, "")

	require.NoError(t, err)
	assert.Equal(t, []any{"best"}, stub.gotOptions)
}
