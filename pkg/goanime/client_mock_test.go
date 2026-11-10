package goanime_test

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper"
	"github.com/alvarorichard/Goanime/v2/pkg/goanime"
	"github.com/alvarorichard/Goanime/v2/pkg/goanime/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockScraper is a UnifiedScraper that records its calls and serves canned
// answers, so Client is tested without any network.
type mockScraper struct {
	episodes   []models.Episode
	episodeErr error
	streamURL  string
	streamMeta map[string]string
	streamErr  error

	episodeCalls []string
	streamCalls  []string
	streamOpts   [][]any
}

func (m *mockScraper) SearchAnime(string, ...any) ([]*models.Anime, error) { return nil, nil }
func (m *mockScraper) GetAnimeEpisodes(u string) ([]models.Episode, error) {
	m.episodeCalls = append(m.episodeCalls, u)
	return m.episodes, m.episodeErr
}
func (m *mockScraper) GetStreamURL(u string, opts ...any) (streamURL string, meta map[string]string, err error) {
	m.streamCalls = append(m.streamCalls, u)
	m.streamOpts = append(m.streamOpts, opts)
	return m.streamURL, m.streamMeta, m.streamErr
}
func (m *mockScraper) GetType() scraper.ScraperType { return scraper.AnimefireType }

// mockManager is the Manager seam: it records which scraper type was asked for.
type mockManager struct {
	results []*models.Anime
	err     error

	searchQueries []string
	searchTypes   []*scraper.ScraperType

	scr         scraper.UnifiedScraper
	scraperErr  error
	scraperReqs []scraper.ScraperType
}

func (m *mockManager) SearchAnime(q string, st *scraper.ScraperType) ([]*models.Anime, error) {
	m.searchQueries = append(m.searchQueries, q)
	m.searchTypes = append(m.searchTypes, st)
	return m.results, m.err
}
func (m *mockManager) GetScraper(st scraper.ScraperType) (scraper.UnifiedScraper, error) {
	m.scraperReqs = append(m.scraperReqs, st)
	return m.scr, m.scraperErr
}

func TestClientSearchAnimeAllSources(t *testing.T) {
	m := &mockManager{results: []*models.Anime{{Name: "Naruto", URL: "u1", Source: "Animefire.io"}}}
	got, err := goanime.NewClientForTest(m).SearchAnime("naruto", nil)

	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "Naruto", got[0].Name)
	assert.Equal(t, []string{"naruto"}, m.searchQueries)
	assert.Equal(t, []*scraper.ScraperType{nil}, m.searchTypes, "nil source means every source")
}

func TestClientSearchAnimeSpecificSource(t *testing.T) {
	m := &mockManager{results: []*models.Anime{{Name: "Bleach"}}}
	src := types.SourceAnimeFire
	_, err := goanime.NewClientForTest(m).SearchAnime("bleach", &src)

	require.NoError(t, err)
	require.Len(t, m.searchTypes, 1)
	require.NotNil(t, m.searchTypes[0], "a chosen source is forwarded")
	assert.Equal(t, scraper.AnimefireType, *m.searchTypes[0])
}

func TestClientSearchAnimePropagatesErrorsAndReturnsNoResults(t *testing.T) {
	boom := errors.New("all sources down")
	m := &mockManager{err: boom, results: []*models.Anime{{Name: "ignored"}}}
	got, err := goanime.NewClientForTest(m).SearchAnime("x", nil)
	assert.ErrorIs(t, err, boom)
	assert.Nil(t, got, "no partial results next to an error")
}

func TestClientSearchAnimeEmptyResultIsEmptyNotNil(t *testing.T) {
	got, err := goanime.NewClientForTest(&mockManager{}).SearchAnime("zzz", nil)
	require.NoError(t, err)
	assert.NotNil(t, got)
	assert.Empty(t, got)
}

func TestClientGetAnimeEpisodes(t *testing.T) {
	scr := &mockScraper{episodes: []models.Episode{{Number: "1", Num: 1, URL: "e1"}, {Number: "2", Num: 2, URL: "e2"}}}
	m := &mockManager{scr: scr}
	got, err := goanime.NewClientForTest(m).GetAnimeEpisodes("https://x/anime", types.SourceAnimeFire)

	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, "e2", got[1].URL)
	assert.Equal(t, []string{"https://x/anime"}, scr.episodeCalls)
	assert.Equal(t, []scraper.ScraperType{scraper.AnimefireType}, m.scraperReqs)
}

func TestClientGetAnimeEpisodesErrors(t *testing.T) {
	t.Run("scraper unavailable", func(t *testing.T) {
		boom := errors.New("no adapter")
		m := &mockManager{scraperErr: boom}
		got, err := goanime.NewClientForTest(m).GetAnimeEpisodes("u", types.SourceAnimeFire)
		assert.ErrorIs(t, err, boom)
		assert.Nil(t, got)
	})
	t.Run("scraper fails", func(t *testing.T) {
		boom := errors.New("parse failed")
		scr := &mockScraper{episodeErr: boom, episodes: []models.Episode{{URL: "ignored"}}}
		got, err := goanime.NewClientForTest(&mockManager{scr: scr}).GetAnimeEpisodes("u", types.SourceAnimeFire)
		assert.ErrorIs(t, err, boom)
		assert.Nil(t, got)
	})
}

func TestClientGetStreamURL(t *testing.T) {
	scr := &mockScraper{streamURL: "https://cdn/v.m3u8", streamMeta: map[string]string{"Referer": "r"}}
	m := &mockManager{scr: scr}
	url, meta, err := goanime.NewClientForTest(m).GetStreamURL("ep", types.SourceAnimeFire, "opt1", 2)

	require.NoError(t, err)
	assert.Equal(t, "https://cdn/v.m3u8", url)
	assert.Equal(t, map[string]string{"Referer": "r"}, meta)
	assert.Equal(t, []string{"ep"}, scr.streamCalls)
	assert.Equal(t, [][]any{{"opt1", 2}}, scr.streamOpts, "options reach the scraper untouched")
}

func TestClientGetStreamURLErrors(t *testing.T) {
	boom := errors.New("nope")
	url, meta, err := goanime.NewClientForTest(&mockManager{scraperErr: boom}).GetStreamURL("e", types.SourceAnimeFire)
	assert.ErrorIs(t, err, boom)
	assert.Empty(t, url)
	assert.Nil(t, meta)

	scr := &mockScraper{streamErr: boom}
	_, _, err = goanime.NewClientForTest(&mockManager{scr: scr}).GetStreamURL("e", types.SourceAnimeFire)
	assert.ErrorIs(t, err, boom)
}

func TestClientGetEpisodeStreamURL(t *testing.T) {
	scr := &mockScraper{streamURL: "https://cdn/a.mp4", streamMeta: map[string]string{"q": "720p"}}
	m := &mockManager{scr: scr}
	anime := &types.Anime{Name: "Naruto", Source: types.SourceAnimeFire.String()}
	ep := &types.Episode{Number: "3", URL: "https://x/ep3"}

	url, meta, err := goanime.NewClientForTest(m).GetEpisodeStreamURL(anime, ep, nil)

	require.NoError(t, err)
	assert.Equal(t, "https://cdn/a.mp4", url)
	assert.Equal(t, "720p", meta["q"])
	assert.Equal(t, []string{"https://x/ep3"}, scr.streamCalls, "the episode URL, not the anime URL")
	assert.Equal(t, [][]any{nil}, scr.streamOpts, "no options are forwarded")
	assert.Equal(t, []scraper.ScraperType{scraper.AnimefireType}, m.scraperReqs)
}

func TestClientGetEpisodeStreamURLUnknownSource(t *testing.T) {
	scr := &mockScraper{}
	m := &mockManager{scr: scr}
	anime := &types.Anime{Name: "X", Source: "NoSuchSite"}
	url, meta, err := goanime.NewClientForTest(m).GetEpisodeStreamURL(anime, &types.Episode{URL: "e"}, &goanime.StreamOptions{})

	require.Error(t, err)
	assert.Contains(t, err.Error(), "NoSuchSite")
	assert.Empty(t, url)
	assert.Nil(t, meta)
	assert.Empty(t, m.scraperReqs, "an unknown source never reaches the manager")
	assert.Empty(t, scr.streamCalls)
}

func TestClientGetEpisodeStreamURLErrors(t *testing.T) {
	boom := errors.New("down")
	anime := &types.Anime{Source: types.SourceAnimeFire.String()}
	ep := &types.Episode{URL: "e"}

	_, _, err := goanime.NewClientForTest(&mockManager{scraperErr: boom}).GetEpisodeStreamURL(anime, ep, nil)
	assert.ErrorIs(t, err, boom, "scraper construction failure")

	_, _, err = goanime.NewClientForTest(&mockManager{scr: &mockScraper{streamErr: boom}}).GetEpisodeStreamURL(anime, ep, nil)
	assert.ErrorIs(t, err, boom, "stream failure")
}

func TestDefaultStreamOptions(t *testing.T) {
	assert.Equal(t, goanime.StreamOptions{Quality: "best", Mode: "sub"}, goanime.DefaultStreamOptions())
}

// TestClientContractMeta is the metatest: the client contract is expressed as a
// checker over a Manager, and deliberately broken Managers (a seam that lies)
// must trip at least one of its assertions. If a mutant passes, the mocks
// above are not sensitive enough.
func clientContractFailures(mgr func() (*mockManager, *mockScraper)) []string {
	var f []string
	m, scr := mgr()
	c := goanime.NewClientForTest(wrapManager{m})

	src := types.SourceAnimeFire
	res, err := c.SearchAnime("q", &src)
	if err != nil || len(res) != 1 || res[0].Name != "A" {
		f = append(f, "search result lost or altered")
	}
	if len(m.searchTypes) != 1 || m.searchTypes[0] == nil || *m.searchTypes[0] != scraper.AnimefireType {
		f = append(f, "source not forwarded")
	}
	eps, err := c.GetAnimeEpisodes("anime-url", src)
	if err != nil || len(eps) != 1 || eps[0].URL != "ep-url" {
		f = append(f, "episodes lost or altered")
	}
	if len(scr.episodeCalls) != 1 || scr.episodeCalls[0] != "anime-url" {
		f = append(f, "anime url not forwarded")
	}
	url, _, err := c.GetEpisodeStreamURL(&types.Anime{Source: src.String()}, &types.Episode{URL: "ep-url"}, nil)
	if err != nil || url != "stream" {
		f = append(f, "stream url lost")
	}
	if len(scr.streamCalls) != 1 || scr.streamCalls[0] != "ep-url" {
		f = append(f, "episode url not forwarded")
	}
	return f
}

// wrapManager adapts a *mockManager so mutants can wrap behaviour.
type wrapManager struct{ *mockManager }

func newContractMocks() (*mockManager, *mockScraper) {
	scr := &mockScraper{episodes: []models.Episode{{URL: "ep-url"}}, streamURL: "stream"}
	return &mockManager{results: []*models.Anime{{Name: "A"}}, scr: scr}, scr
}

func TestClientContractMeta(t *testing.T) {
	require.Empty(t, clientContractFailures(newContractMocks), "the real client satisfies the contract")

	mutants := map[string]func() (*mockManager, *mockScraper){
		"search results dropped": func() (*mockManager, *mockScraper) {
			m, s := newContractMocks()
			m.results = nil
			return m, s
		},
		"search result renamed": func() (*mockManager, *mockScraper) {
			m, s := newContractMocks()
			m.results = []*models.Anime{{Name: "B"}}
			return m, s
		},
		"episodes dropped": func() (*mockManager, *mockScraper) {
			m, s := newContractMocks()
			s.episodes = nil
			return m, s
		},
		"wrong episode url": func() (*mockManager, *mockScraper) {
			m, s := newContractMocks()
			s.episodes = []models.Episode{{URL: "other"}}
			return m, s
		},
		"empty stream": func() (*mockManager, *mockScraper) {
			m, s := newContractMocks()
			s.streamURL = ""
			return m, s
		},
		"search error": func() (*mockManager, *mockScraper) {
			m, s := newContractMocks()
			m.err = errors.New("x")
			return m, s
		},
	}
	for name, mk := range mutants {
		assert.NotEmpty(t, clientContractFailures(mk), "mutant %q survived", name)
	}
}

// The movie and series catalogues can be searched but not listed: the client
// says so, and never reaches the scraper (a mockManager with a scraper would
// answer for them).
func TestClientEpisodesUnsupportedSources(t *testing.T) {
	for _, src := range []types.Source{types.SourceTopCine, types.SourcePobreflix, types.SourcePomfy} {
		m := &mockManager{scr: &mockScraper{}}
		c := goanime.NewClientForTest(m)

		_, err := c.GetAnimeEpisodes("u", src)
		require.ErrorIs(t, err, goanime.ErrEpisodesUnsupported, src.String())
		assert.Contains(t, err.Error(), src.String())

		_, _, err = c.GetStreamURL("e", src)
		require.ErrorIs(t, err, goanime.ErrEpisodesUnsupported, src.String())

		_, _, err = c.GetEpisodeStreamURL(&types.Anime{Source: src.String()}, &types.Episode{URL: "e"}, nil)
		require.ErrorIs(t, err, goanime.ErrEpisodesUnsupported, src.String())

		assert.Empty(t, m.scraperReqs, "%s: no scraper was built", src)
	}
}

func TestClientEpisodesSupportedSourcesReachTheScraper(t *testing.T) {
	for _, src := range types.Sources() {
		if !src.SupportsEpisodes() || src.IsCatalog() {
			continue // catalogue sources are covered by the Season tests
		}
		m := &mockManager{scr: &mockScraper{episodes: []models.Episode{{Number: "1", URL: "e1"}}}}
		_, err := goanime.NewClientForTest(m).GetAnimeEpisodes("u", src)
		require.NoError(t, err, src.String())
		assert.Equal(t, []scraper.ScraperType{src.ToScraperType()}, m.scraperReqs, src.String())
	}
}

// mockCatalog is a UnifiedScraper that is also a CatalogScraper.
type mockCatalog struct {
	mockScraper
	seasons  []scraper.CatalogSeason
	episodes []models.Episode
	err      error

	gotURL    string
	gotSeason int
	gotAudio  string
}

func (m *mockCatalog) CatalogSeasons(_ context.Context, u string) ([]scraper.CatalogSeason, error) {
	m.gotURL = u
	return m.seasons, m.err
}
func (m *mockCatalog) CatalogEpisodes(_ context.Context, u string, season int, audio string) ([]models.Episode, error) {
	m.gotURL, m.gotSeason, m.gotAudio = u, season, audio
	return m.episodes, m.err
}
func (m *mockCatalog) CatalogStream(_ context.Context, u, audio string) (string, map[string]string, error) {
	m.gotURL, m.gotAudio = u, audio
	return "https://video/" + u, map[string]string{"audio": audio}, m.err
}

func startflixTitle() *types.Anime {
	return &types.Anime{Name: "Show", URL: "https://sf/series/x/", Source: types.SourceStartFlix.String()}
}

func TestClientGetSeasons(t *testing.T) {
	cat := &mockCatalog{seasons: []scraper.CatalogSeason{{Number: 1, Audios: []string{"Dublado"}}, {Number: 2, Audios: []string{"Dublado", "Legendado"}}}}
	m := &mockManager{scr: cat}
	got, err := goanime.NewClientForTest(m).GetSeasons(startflixTitle())
	require.NoError(t, err)
	require.Len(t, got, 2)
	assert.Equal(t, 2, got[1].Number)
	assert.Equal(t, []string{"Dublado", "Legendado"}, got[1].Audios)
	assert.Equal(t, "https://sf/series/x/", cat.gotURL)
	assert.Equal(t, []scraper.ScraperType{scraper.StartFlixType}, m.scraperReqs)
}

func TestClientGetSeasonsRefusals(t *testing.T) {
	c := goanime.NewClientForTest(&mockManager{scr: &mockCatalog{}})

	_, err := c.GetSeasons(nil)
	require.Error(t, err)

	_, err = c.GetSeasons(&types.Anime{Source: "NoSuchSite"})
	require.Error(t, err)

	_, err = c.GetSeasons(&types.Anime{Source: types.SourceAnimeFire.String()})
	require.ErrorIs(t, err, goanime.ErrNotCatalog, "an anime source has no seasons")

	// A catalogue source whose scraper is not a CatalogScraper.
	c = goanime.NewClientForTest(&mockManager{scr: &mockScraper{}})
	_, err = c.GetSeasons(startflixTitle())
	require.ErrorIs(t, err, goanime.ErrEpisodesUnsupported)

	c = goanime.NewClientForTest(&mockManager{scr: &mockCatalog{err: errors.New("boom")}})
	_, err = c.GetSeasons(startflixTitle())
	require.EqualError(t, err, "boom")
}

func TestClientGetSeasonEpisodes(t *testing.T) {
	cat := &mockCatalog{episodes: []models.Episode{{Number: "1", Num: 1, URL: "e1"}}}
	got, err := goanime.NewClientForTest(&mockManager{scr: cat}).GetSeasonEpisodes(startflixTitle(), 3, "Legendado")
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "e1", got[0].URL)
	assert.Equal(t, 3, cat.gotSeason)
	assert.Equal(t, "Legendado", cat.gotAudio)

	_, err = goanime.NewClientForTest(&mockManager{scr: &mockCatalog{err: scraper.ErrSeasonNotFound}}).GetSeasonEpisodes(startflixTitle(), 9, "")
	require.ErrorIs(t, err, goanime.ErrSeasonNotFound)
}

// For a catalogue source the old by-source calls go through the catalogue too:
// a movie is its one episode, a series asks for a season.
func TestClientGetAnimeEpisodesCatalogSource(t *testing.T) {
	cat := &mockCatalog{episodes: []models.Episode{{Number: "1", URL: "movie"}}}
	got, err := goanime.NewClientForTest(&mockManager{scr: cat}).GetAnimeEpisodes("https://sf/filmes/m/", types.SourceStartFlix)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, 0, cat.gotSeason)
	assert.Equal(t, "", cat.gotAudio)

	_, err = goanime.NewClientForTest(&mockManager{scr: &mockCatalog{err: scraper.ErrSeasonRequired}}).GetAnimeEpisodes("u", types.SourceStartFlix)
	require.ErrorIs(t, err, goanime.ErrSeasonRequired)
}

func TestClientGetEpisodeStreamURLCatalogAudio(t *testing.T) {
	cat := &mockCatalog{}
	c := goanime.NewClientForTest(&mockManager{scr: cat})

	u, meta, err := c.GetEpisodeStreamURL(startflixTitle(), &types.Episode{URL: "ep9"}, &goanime.StreamOptions{Audio: "Legendado"})
	require.NoError(t, err)
	assert.Equal(t, "https://video/ep9", u)
	assert.Equal(t, "Legendado", meta["audio"])
	assert.Equal(t, "Legendado", cat.gotAudio)

	_, _, err = c.GetEpisodeStreamURL(startflixTitle(), &types.Episode{URL: "ep9"}, nil)
	require.NoError(t, err)
	assert.Equal(t, "", cat.gotAudio, "nil options mean no preference")

	_, _, err = goanime.NewClientForTest(&mockManager{scr: &mockCatalog{err: scraper.ErrAudioUnavailable}}).
		GetEpisodeStreamURL(startflixTitle(), &types.Episode{URL: "e"}, &goanime.StreamOptions{Audio: "Legendado"})
	require.ErrorIs(t, err, goanime.ErrAudioUnavailable)
}

func TestClientGetStreamURLCatalogSource(t *testing.T) {
	cat := &mockCatalog{}
	u, _, err := goanime.NewClientForTest(&mockManager{scr: cat}).GetStreamURL("ep1", types.SourceStartFlix)
	require.NoError(t, err)
	assert.Equal(t, "https://video/ep1", u)
}
