package providers

import (
	"context"
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/api/providers/naming"
	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// The Pluto TV provider is a StreamFetcher and a seasoned source; the registry
// and the downloads rely on both.
var (
	_ source.StreamFetcher = (*plutoTVProvider)(nil)
	_ source.Seasoned      = (*plutoTVProvider)(nil)
	_ source.Searchable    = (*plutoTVProvider)(nil)
)

// stubPluto swaps the two api functions the provider calls and records them.
type stubPluto struct {
	episodesCalls int
	streamCalls   int
	gotAnime      *models.Anime
	gotEpisode    *models.Episode
	gotQuality    string
	episodes      []models.Episode
	episodesErr   error
	stream        source.Stream
	streamErr     error
}

func installPlutoStub(t *testing.T) *stubPluto {
	t.Helper()
	s := &stubPluto{}
	prevEpisodes, prevStream := plutoTVEpisodesFn, plutoTVStreamFn
	plutoTVEpisodesFn = func(a *models.Anime) ([]models.Episode, error) {
		s.episodesCalls++
		s.gotAnime = a
		return s.episodes, s.episodesErr
	}
	plutoTVStreamFn = func(a *models.Anime, e *models.Episode, quality string) (source.Stream, error) {
		s.streamCalls++
		s.gotAnime, s.gotEpisode, s.gotQuality = a, e, quality
		return s.stream, s.streamErr
	}
	t.Cleanup(func() { plutoTVEpisodesFn, plutoTVStreamFn = prevEpisodes, prevStream })
	return s
}

func TestPlutoTVDescribesItself(t *testing.T) {
	d := (&plutoTVProvider{}).Describe()
	assert.Equal(t, source.PlutoTV, d.Kind)
	assert.Contains(t, d.Explicit, "PlutoTV")
	assert.Contains(t, d.Explicit, "Pluto TV")
	assert.Equal(t, []string{"[plutotv]"}, d.Tags)
	assert.Equal(t, []string{"pluto.tv"}, d.URLMatchers)
	assert.Equal(t, "https://pluto.tv", d.ProbeURL)
	assert.False(t, d.DefaultDisabled, "it ships on")
	assert.True(t, (&plutoTVProvider{}).HasSeasons())
}

func TestPlutoTVIsRegisteredAndSeasoned(t *testing.T) {
	src, ok := source.Registered(source.PlutoTV)
	require.True(t, ok, "the provider registers itself from init")
	assert.True(t, source.IsSeasoned(src))
	assert.False(t, source.IsBrowserGated(src))
	assert.Contains(t, source.Kinds(), source.PlutoTV)
}

func TestPlutoTVPriorityLeavesTheOtherSourcesWhereTheyWere(t *testing.T) {
	// Resolution is by priority; a tie would fall to the kind's spelling.
	seen := map[int]source.SourceKind{}
	for _, kind := range source.Kinds() {
		src, _ := source.Registered(kind)
		p := src.Describe().Priority
		if other, dup := seen[p]; dup {
			t.Errorf("%s and %s share priority %d", kind, other, p)
		}
		seen[p] = kind
	}
}

func TestPlutoTVResolves(t *testing.T) {
	cases := []struct {
		name  string
		anime *models.Anime
	}{
		{"by its source", &models.Anime{Source: "PlutoTV"}},
		{"by its spaced source", &models.Anime{Source: "Pluto TV"}},
		{"by a movie URL", &models.Anime{URL: "https://pluto.tv/on-demand/movies/62b9ea9ec4eb8f00140cb633"}},
		{"by a series URL", &models.Anime{URL: "https://pluto.tv/on-demand/series/6206bba315c8610013683974"}},
		{"by its tag", &models.Anime{Name: "[plutotv] x"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			_, resolved := source.Resolve(tc.anime)
			assert.Equal(t, source.PlutoTV, resolved.Kind, resolved.Reason)
		})
	}
	_, byURL := source.ResolveURL("https://pluto.tv/on-demand/series/6206bba315c8610013683974/episode/6206bbaa15c8610013683a28")
	assert.Equal(t, source.PlutoTV, byURL.Kind)
}

func TestPlutoTVDoesNotClaimAnotherSourcesTitles(t *testing.T) {
	for _, anime := range []*models.Anime{
		{Source: "Pomfy", URL: "https://pomfy.my/assistir/1"},
		{Source: "StartFlix", URL: "https://startflix.test/x"},
		{Source: "AnimesFHD", URL: "https://animesfhd.net/serie/1"},
	} {
		_, resolved := source.Resolve(anime)
		assert.NotEqual(t, source.PlutoTV, resolved.Kind, anime.Source)
	}
}

func TestPlutoTVFetchEpisodesDelegates(t *testing.T) {
	stub := installPlutoStub(t)
	stub.episodes = []models.Episode{{Number: "1", Num: 1}, {Number: "2", Num: 2}}
	anime := &models.Anime{Name: "x"}

	eps, err := (&plutoTVProvider{}).FetchEpisodes(t.Context(), anime)
	require.NoError(t, err)
	assert.Equal(t, stub.episodes, eps)
	assert.Same(t, anime, stub.gotAnime, "the title itself, so the chosen season is recorded on it")

	stub.episodesErr = errors.New("catalog down")
	_, err = (&plutoTVProvider{}).FetchEpisodes(t.Context(), anime)
	assert.EqualError(t, err, "catalog down", "an error is passed on as it came")
}

func TestPlutoTVFetchStreamDelegates(t *testing.T) {
	stub := installPlutoStub(t)
	stub.stream = source.Stream{URL: "https://cdn/x.m3u8"}
	anime, episode := &models.Anime{Name: "x"}, &models.Episode{URL: "e"}

	got, err := (&plutoTVProvider{}).FetchStream(t.Context(), episode, anime, "720p")
	require.NoError(t, err)
	assert.Equal(t, stub.stream, got)
	assert.Same(t, anime, stub.gotAnime)
	assert.Same(t, episode, stub.gotEpisode)
	assert.Equal(t, "720p", stub.gotQuality, "the quality reaches the resolver")

	stub.streamErr = errors.New("blocked")
	_, err = (&plutoTVProvider{}).FetchStream(t.Context(), episode, anime, "")
	assert.EqualError(t, err, "blocked")
}

func TestPlutoTVNeverReachesTheServiceOnACancelledContext(t *testing.T) {
	stub := installPlutoStub(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	p := &plutoTVProvider{}

	_, err := p.FetchEpisodes(ctx, &models.Anime{})
	assert.ErrorIs(t, err, context.Canceled)
	_, err = p.FetchStream(ctx, &models.Episode{}, &models.Anime{}, "best")
	assert.ErrorIs(t, err, context.Canceled)
	_, err = p.FetchStreamURL(ctx, &models.Episode{}, &models.Anime{}, "best")
	assert.ErrorIs(t, err, context.Canceled)
	_, err = p.Search(ctx, "x")
	assert.ErrorIs(t, err, context.Canceled)
	assert.Zero(t, stub.episodesCalls+stub.streamCalls)
}

func TestPlutoTVFetchStreamURLPublishesAndClearsWhatTheLastStreamLeft(t *testing.T) {
	cleanPlayback(t)
	stub := installPlutoStub(t)
	stub.stream = source.Stream{URL: "https://cdn/x.m3u8"}
	util.SetGlobalReferer("https://stale.example/")
	util.SetGlobalOrigin("https://stale.example")
	util.SetGlobalSubtitles([]util.SubtitleInfo{{URL: "stale.vtt"}})

	url, err := (&plutoTVProvider{}).FetchStreamURL(t.Context(), &models.Episode{URL: "e"}, &models.Anime{Source: "PlutoTV"}, "best")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn/x.m3u8", url)
	snap := readPlayback()
	assert.Empty(t, snap.referer, "Pluto TV's media takes none; the last source's must not ride along")
	assert.Empty(t, snap.origin)
	assert.Empty(t, snap.subs)
	assert.Equal(t, "PlutoTV", snap.animeSource)
}

func TestPlutoTVAFailedFetchStreamURLPublishesNothing(t *testing.T) {
	cleanPlayback(t)
	stub := installPlutoStub(t)
	stub.streamErr = errors.New("blocked")
	_, err := (&plutoTVProvider{}).FetchStreamURL(t.Context(), &models.Episode{}, &models.Anime{Source: "PlutoTV"}, "best")
	require.Error(t, err)
	assert.Empty(t, readPlayback().referer)
}

func TestTagResultsForPlutoTV(t *testing.T) {
	results := []*models.Anime{
		{Name: "[Movie] [PT-BR] Halloween II", MediaType: models.MediaTypeMovie, Source: "x"},
		{Name: "[TV] [English] Death Note A18 (Dublado)", MediaType: models.MediaTypeTV},
		{Name: "Untagged", MediaType: models.MediaTypeAnime},
	}
	tagResults(results, source.PlutoTV)
	for _, r := range results {
		assert.Equal(t, "PlutoTV", r.Source)
	}
	assert.Equal(t, "[Movie] [PT-BR] Halloween II", results[0].Name, "already tagged: left as the scraper wrote it")
	assert.Equal(t, "[TV] [English] Death Note A18 (Dublado)", results[1].Name, "not a Portuguese-only source: no title cleaning, no dub label")
	assert.Equal(t, "[PT-BR] Untagged", results[2].Name, "a title that lost its tags gets the catalog's default")
}

func TestPlutoTVNamesAreCleanedLikeTheOthers(t *testing.T) {
	assert.Equal(t, "PlutoTV", sourceDisplayName(source.PlutoTV))
	assert.Equal(t, "Death Note", naming.CleanTitle("[TV] [PT-BR] Death Note"), "the tags a result carries")
	assert.Equal(t, "Death Note", naming.CleanTitle("Death Note [PlutoTV]"), "and the source's own, should it ever be stamped")
	assert.Equal(t, "Halloween II", naming.CleanTitle("[Movie] [English] Halloween II"))
}

// ---- search, through an adapter of the test's own --------------------------

// providerWith is a provider whose adapter is already built: the sync.Once is
// spent, so the registry's real adapter (and its network) is never made.
func providerWith(adapter scraper.UnifiedScraper) *plutoTVProvider {
	p := &plutoTVProvider{}
	p.once.Do(func() { p.adapter.UnifiedScraper = adapter })
	return p
}

type plutoPlainScraper struct {
	results []*models.Anime
	err     error
	query   string
}

func (s *plutoPlainScraper) SearchAnime(q string, _ ...any) ([]*models.Anime, error) {
	s.query = q
	return s.results, s.err
}
func (s *plutoPlainScraper) GetAnimeEpisodes(string) ([]models.Episode, error) { return nil, nil }
func (s *plutoPlainScraper) GetStreamURL(string, ...any) (string, map[string]string, error) {
	return "", nil, nil
}
func (s *plutoPlainScraper) GetType() scraper.ScraperType { return scraper.PlutoTVType }

// plutoCtxScraper is the same, and a ContextualScraper: it keeps the context
// it was searched with.
type plutoCtxScraper struct {
	plutoPlainScraper
	ctx   context.Context
	plain int
}

func (s *plutoCtxScraper) SearchAnime(q string, o ...any) ([]*models.Anime, error) {
	s.plain++
	return s.plutoPlainScraper.SearchAnime(q, o...)
}
func (s *plutoCtxScraper) SearchAnimeContext(ctx context.Context, q string, _ ...any) ([]*models.Anime, error) {
	s.ctx = ctx
	s.query = q
	return s.results, s.err
}
func (s *plutoCtxScraper) GetAnimeEpisodesContext(context.Context, string) ([]models.Episode, error) {
	return nil, nil
}
func (s *plutoCtxScraper) GetStreamURLContext(context.Context, string, ...any) (string, map[string]string, error) {
	return "", nil, nil
}

type plutoCtxKey struct{}

func TestPlutoTVSearchTagsWhatTheAdapterFoundAndSearchesWithTheCallersContext(t *testing.T) {
	adapter := &plutoCtxScraper{plutoPlainScraper: plutoPlainScraper{results: []*models.Anime{
		{Name: "[Movie] [PT-BR] Em Nome do Pai", MediaType: models.MediaTypeMovie, Source: "x"},
	}}}
	ctx := context.WithValue(t.Context(), plutoCtxKey{}, "mine")

	results, err := providerWith(adapter).Search(ctx, "pai")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "PlutoTV", results[0].Source, "tagged as Pluto TV's, not another source's")
	assert.Equal(t, "pai", adapter.query)
	assert.Equal(t, "mine", adapter.ctx.Value(plutoCtxKey{}), "the fan-out's deadline reaches the request")
	assert.Zero(t, adapter.plain, "the context-aware call is the one used")
}

func TestPlutoTVSearchWorksWithAnAdapterThatTakesNoContext(t *testing.T) {
	adapter := &plutoPlainScraper{results: []*models.Anime{{Name: "Untagged"}}}
	results, err := providerWith(adapter).Search(t.Context(), "x")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "PlutoTV", results[0].Source)
	assert.Equal(t, "x", adapter.query)
}

func TestPlutoTVSearchPassesAFailureOnWithoutResults(t *testing.T) {
	for name, adapter := range map[string]scraper.UnifiedScraper{
		"plain":      &plutoPlainScraper{results: []*models.Anime{{Name: "x"}}, err: errors.New("down")},
		"contextual": &plutoCtxScraper{plutoPlainScraper: plutoPlainScraper{results: []*models.Anime{{Name: "x"}}, err: errors.New("down")}},
	} {
		t.Run(name, func(t *testing.T) {
			results, err := providerWith(adapter).Search(t.Context(), "x")
			assert.EqualError(t, err, "down")
			assert.Nil(t, results, "a failed search tags and returns nothing")
		})
	}
}

func TestPlutoTVSearchWithoutAnAdapterSaysSo(t *testing.T) {
	_, err := providerWith(nil).Search(t.Context(), "x")
	require.ErrorContains(t, err, "no adapter")
}

func TestPlutoTVBuildsItsOwnAdapterOnce(t *testing.T) {
	p := &plutoTVProvider{}
	first, err := p.scraper()
	require.NoError(t, err)
	assert.Equal(t, scraper.PlutoTVType, first.GetType(), "Pluto TV's adapter, not another source's")
	second, err := p.scraper()
	require.NoError(t, err)
	assert.Same(t, first, second, "built once and kept")
}
