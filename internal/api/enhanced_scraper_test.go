package api

import (
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// mockEpiScraper is a testable implementation of scraper.UnifiedScraper.
type mockEpiScraper struct {
	episodes []models.Episode
	epsErr   error
	stURL    string
	stErr    error
	tp       scraper.ScraperType
}

func (m *mockEpiScraper) SearchAnime(_ string, _ ...any) ([]*models.Anime, error) {
	return nil, nil
}

func (m *mockEpiScraper) GetAnimeEpisodes(_ string) ([]models.Episode, error) {
	return m.episodes, m.epsErr
}

func (m *mockEpiScraper) GetStreamURL(_ string, _ ...any) (string, map[string]string, error) {
	return m.stURL, nil, m.stErr
}

func (m *mockEpiScraper) GetType() scraper.ScraperType {
	return m.tp
}

// injectScraper wires a mock for scraperType into newScraperMgr and restores on cleanup.
func injectScraper(t *testing.T, scraperType scraper.ScraperType, mock scraper.UnifiedScraper) {
	t.Helper()
	mgr := scraper.NewScraperManagerForTest()
	mgr.RegisterScraperForTest(scraperType, mock)
	prev := newScraperMgr
	newScraperMgr = func() *scraper.ScraperManager { return mgr }
	t.Cleanup(func() { newScraperMgr = prev })
}

// injectMultiScraper wires multiple mock scrapers (e.g. for DownloadEpisodeEnhanced which
// calls both GetAnimeEpisodesEnhanced and GetEpisodeStreamURL, each calling newScraperMgr).
func injectMultiScraper(t *testing.T, mocks map[scraper.ScraperType]scraper.UnifiedScraper) {
	t.Helper()
	mgr := scraper.NewScraperManagerForTest()
	for tp, m := range mocks {
		mgr.RegisterScraperForTest(tp, m)
	}
	prev := newScraperMgr
	newScraperMgr = func() *scraper.ScraperManager { return mgr }
	t.Cleanup(func() { newScraperMgr = prev })
}

// --- GetAnimeEpisodesEnhanced ---

func TestGetAnimeEpisodesEnhanced_AllAnimeSource(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AllAnimeType,
	}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto", URL: "naruto-id"}
	eps, err := GetAnimeEpisodesEnhanced(anime)
	require.NoError(t, err)
	assert.Len(t, eps, 1)
}

func TestGetAnimeEpisodesEnhanced_AnimefireSource(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AnimefireType,
	}
	injectScraper(t, scraper.AnimefireType, mock)

	anime := &models.Anime{Source: "Animefire.io", Name: "Naruto"}
	eps, err := GetAnimeEpisodesEnhanced(anime)
	require.NoError(t, err)
	assert.Len(t, eps, 1)
}

func TestGetAnimeEpisodesEnhanced_GoyabuSource(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.GoyabuType,
	}
	injectScraper(t, scraper.GoyabuType, mock)

	anime := &models.Anime{Source: "Goyabu", Name: "Naruto"}
	eps, err := GetAnimeEpisodesEnhanced(anime)
	require.NoError(t, err)
	assert.Len(t, eps, 1)
}

func TestGetAnimeEpisodesEnhanced_EnglishTag(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AllAnimeType,
	}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "", Name: "Naruto [English]", URL: "naruto-id"}
	eps, err := GetAnimeEpisodesEnhanced(anime)
	require.NoError(t, err)
	assert.Len(t, eps, 1)
	assert.Equal(t, "AllAnime", anime.Source)
}

func TestGetAnimeEpisodesEnhanced_PTBRTagGoyabuURL(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.GoyabuType,
	}
	injectScraper(t, scraper.GoyabuType, mock)

	anime := &models.Anime{Source: "", Name: "Naruto [PT-BR]", URL: "https://goyabu.com/naruto"}
	eps, err := GetAnimeEpisodesEnhanced(anime)
	require.NoError(t, err)
	assert.Len(t, eps, 1)
	assert.Equal(t, "Goyabu", anime.Source)
}

func TestGetAnimeEpisodesEnhanced_PTBRTagAnimefireURL(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AnimefireType,
	}
	injectScraper(t, scraper.AnimefireType, mock)

	anime := &models.Anime{Source: "", Name: "Naruto [PT-BR]", URL: "https://animefire.net/naruto"}
	eps, err := GetAnimeEpisodesEnhanced(anime)
	require.NoError(t, err)
	assert.Len(t, eps, 1)
	assert.Equal(t, "Animefire.io", anime.Source)
}

func TestGetAnimeEpisodesEnhanced_PortuguesTag(t *testing.T) {
	// [Português] tag falls into the PT-BR branch; non-goyabu URL → AnimeFire.
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AnimefireType,
	}
	injectScraper(t, scraper.AnimefireType, mock)

	anime := &models.Anime{Source: "", Name: "Naruto [Português]", URL: "https://other.net/naruto"}
	eps, err := GetAnimeEpisodesEnhanced(anime)
	require.NoError(t, err)
	assert.Len(t, eps, 1)
}

func TestGetAnimeEpisodesEnhanced_URLShortIDAllAnime(t *testing.T) {
	// Short ID (< 30 chars, no http) is identified as AllAnime.
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AllAnimeType,
	}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "", Name: "Naruto", URL: "abc123XYZ"}
	eps, err := GetAnimeEpisodesEnhanced(anime)
	require.NoError(t, err)
	assert.Len(t, eps, 1)
	assert.Equal(t, "AllAnime", anime.Source)
}

func TestGetAnimeEpisodesEnhanced_URLAnimeFire(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AnimefireType,
	}
	injectScraper(t, scraper.AnimefireType, mock)

	anime := &models.Anime{Source: "", Name: "Naruto", URL: "https://animefire.net/animes/naruto"}
	eps, err := GetAnimeEpisodesEnhanced(anime)
	require.NoError(t, err)
	assert.Len(t, eps, 1)
	assert.Equal(t, "Animefire.io", anime.Source)
}

func TestGetAnimeEpisodesEnhanced_DefaultFallback_AllAnimeScraper(t *testing.T) {
	// Unknown source/URL hits the default switch case (sourceName="AllAnime (default)"),
	// which still routes through the AllAnime scraper branch.
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AllAnimeType,
	}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "Unknown", Name: "SomePlainAnime", URL: "https://other-site.net/anime"}
	eps, err := GetAnimeEpisodesEnhanced(anime)
	require.NoError(t, err)
	assert.Len(t, eps, 1)
	assert.Equal(t, "AllAnime", anime.Source)
}

func TestGetAnimeEpisodesEnhanced_ScraperReturnsError(t *testing.T) {
	mock := &mockEpiScraper{
		epsErr: errors.New("upstream scraper failed"),
		tp:     scraper.AllAnimeType,
	}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto", URL: "naruto-id"}
	_, err := GetAnimeEpisodesEnhanced(anime)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get episodes")
}

func TestGetAnimeEpisodesEnhanced_ScraperNotFound(t *testing.T) {
	// Empty test manager — no scraper registered → GetScraper returns error.
	prev := newScraperMgr
	newScraperMgr = func() *scraper.ScraperManager { return scraper.NewScraperManagerForTest() }
	t.Cleanup(func() { newScraperMgr = prev })

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto", URL: "naruto-id"}
	_, err := GetAnimeEpisodesEnhanced(anime)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "AllAnime scraper")
}

func TestGetAnimeEpisodesEnhanced_NoEpisodesFoundLogs(t *testing.T) {
	// Empty episode list exercises the "No episodes found" branch.
	mock := &mockEpiScraper{episodes: nil, tp: scraper.AnimefireType}
	injectScraper(t, scraper.AnimefireType, mock)

	anime := &models.Anime{Source: "Animefire.io", Name: "NoEpAnime"}
	eps, err := GetAnimeEpisodesEnhanced(anime)
	require.NoError(t, err)
	assert.Empty(t, eps)
}

// --- GetEpisodeStreamURL ---

func TestGetEpisodeStreamURL_AllAnimeSource(t *testing.T) {
	mock := &mockEpiScraper{stURL: "http://stream.test/ep1.m3u8", tp: scraper.AllAnimeType}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto", URL: "naruto-id"}
	ep := &models.Episode{Number: "1"}
	url, err := GetEpisodeStreamURL(ep, anime, "best")
	require.NoError(t, err)
	assert.Equal(t, "http://stream.test/ep1.m3u8", url)
}

func TestGetEpisodeStreamURL_AnimefireSource(t *testing.T) {
	mock := &mockEpiScraper{stURL: "http://stream.test/animefire.m3u8", tp: scraper.AnimefireType}
	injectScraper(t, scraper.AnimefireType, mock)

	anime := &models.Anime{Source: "Animefire.io", Name: "Naruto"}
	ep := &models.Episode{Number: "1", URL: "http://animefire.net/ep/1"}
	url, err := GetEpisodeStreamURL(ep, anime, "best")
	require.NoError(t, err)
	assert.Equal(t, "http://stream.test/animefire.m3u8", url)
}

func TestGetEpisodeStreamURL_GoyabuSource(t *testing.T) {
	mock := &mockEpiScraper{stURL: "http://stream.test/goyabu.m3u8", tp: scraper.GoyabuType}
	injectScraper(t, scraper.GoyabuType, mock)

	anime := &models.Anime{Source: "Goyabu", Name: "Naruto"}
	ep := &models.Episode{Number: "1", URL: "http://goyabu.com/ep/1"}
	url, err := GetEpisodeStreamURL(ep, anime, "best")
	require.NoError(t, err)
	assert.Equal(t, "http://stream.test/goyabu.m3u8", url)
}

func TestGetEpisodeStreamURL_EnglishTag(t *testing.T) {
	mock := &mockEpiScraper{stURL: "http://stream.test/en.m3u8", tp: scraper.AllAnimeType}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "", Name: "Naruto [English]", URL: "naruto-id"}
	ep := &models.Episode{Number: "1"}
	url, err := GetEpisodeStreamURL(ep, anime, "best")
	require.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestGetEpisodeStreamURL_PTBRTagGoyabuURL(t *testing.T) {
	mock := &mockEpiScraper{stURL: "http://stream.test/goyabu.m3u8", tp: scraper.GoyabuType}
	injectScraper(t, scraper.GoyabuType, mock)

	anime := &models.Anime{Source: "", Name: "Naruto [PT-BR]", URL: "https://goyabu.com/naruto"}
	ep := &models.Episode{Number: "1", URL: "http://goyabu.com/ep/1"}
	url, err := GetEpisodeStreamURL(ep, anime, "best")
	require.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestGetEpisodeStreamURL_PTBRTagAnimefireURL(t *testing.T) {
	mock := &mockEpiScraper{stURL: "http://stream.test/af.m3u8", tp: scraper.AnimefireType}
	injectScraper(t, scraper.AnimefireType, mock)

	anime := &models.Anime{Source: "", Name: "Naruto [PT-BR]", URL: "https://animefire.net/naruto"}
	ep := &models.Episode{Number: "1", URL: "http://animefire.net/ep/1"}
	url, err := GetEpisodeStreamURL(ep, anime, "best")
	require.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestGetEpisodeStreamURL_URLShortIDAllAnime(t *testing.T) {
	mock := &mockEpiScraper{stURL: "http://stream.test/short.m3u8", tp: scraper.AllAnimeType}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "", Name: "Naruto", URL: "abc123XYZ"}
	ep := &models.Episode{Number: "1"}
	url, err := GetEpisodeStreamURL(ep, anime, "best")
	require.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestGetEpisodeStreamURL_URLContainsAnimeFire(t *testing.T) {
	mock := &mockEpiScraper{stURL: "http://stream.test/af2.m3u8", tp: scraper.AnimefireType}
	injectScraper(t, scraper.AnimefireType, mock)

	anime := &models.Anime{Source: "", Name: "Naruto", URL: "https://animefire.net/animes/naruto"}
	ep := &models.Episode{Number: "1", URL: "http://animefire.net/ep/1"}
	url, err := GetEpisodeStreamURL(ep, anime, "best")
	require.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestGetEpisodeStreamURL_URLContainsGoyabu(t *testing.T) {
	mock := &mockEpiScraper{stURL: "http://stream.test/goyabu2.m3u8", tp: scraper.GoyabuType}
	injectScraper(t, scraper.GoyabuType, mock)

	anime := &models.Anime{Source: "", Name: "Naruto", URL: "https://goyabu.com/naruto"}
	ep := &models.Episode{Number: "1", URL: "http://goyabu.com/ep/1"}
	url, err := GetEpisodeStreamURL(ep, anime, "best")
	require.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestGetEpisodeStreamURL_URLContainsAllAnime(t *testing.T) {
	mock := &mockEpiScraper{stURL: "http://stream.test/aa2.m3u8", tp: scraper.AllAnimeType}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "", Name: "Naruto", URL: "https://allanime.to/naruto"}
	ep := &models.Episode{Number: "1"}
	url, err := GetEpisodeStreamURL(ep, anime, "best")
	require.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestGetEpisodeStreamURL_DefaultFallback(t *testing.T) {
	mock := &mockEpiScraper{stURL: "http://stream.test/default.m3u8", tp: scraper.AllAnimeType}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "UnknownSource", Name: "SomePlainAnime", URL: "https://other.net/anime"}
	ep := &models.Episode{Number: "1"}
	url, err := GetEpisodeStreamURL(ep, anime, "best")
	require.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestGetEpisodeStreamURL_EmptyStreamURL(t *testing.T) {
	// Mock returns empty stream URL → function returns error.
	mock := &mockEpiScraper{stURL: "", tp: scraper.AllAnimeType}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	ep := &models.Episode{Number: "1"}
	_, err := GetEpisodeStreamURL(ep, anime, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "empty stream URL")
}

func TestGetEpisodeStreamURL_ScraperError(t *testing.T) {
	mock := &mockEpiScraper{stErr: errors.New("scraper failed"), tp: scraper.AllAnimeType}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	ep := &models.Episode{Number: "1"}
	_, err := GetEpisodeStreamURL(ep, anime, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get stream URL")
}

func TestGetEpisodeStreamURL_DefaultQuality(t *testing.T) {
	// Empty quality string defaults to "best" internally.
	mock := &mockEpiScraper{stURL: "http://stream.test/default-q.m3u8", tp: scraper.AllAnimeType}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	ep := &models.Episode{Number: "1"}
	url, err := GetEpisodeStreamURL(ep, anime, "") // empty quality
	require.NoError(t, err)
	assert.NotEmpty(t, url)
}

func TestGetEpisodeStreamURL_ScraperNotFound(t *testing.T) {
	prev := newScraperMgr
	newScraperMgr = func() *scraper.ScraperManager { return scraper.NewScraperManagerForTest() }
	t.Cleanup(func() { newScraperMgr = prev })

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	ep := &models.Episode{Number: "1"}
	_, err := GetEpisodeStreamURL(ep, anime, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get scraper")
}

// --- DownloadEpisodeEnhanced ---

func TestDownloadEpisodeEnhanced_EpisodesError(t *testing.T) {
	mock := &mockEpiScraper{epsErr: errors.New("upstream failed"), tp: scraper.AllAnimeType}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	err := DownloadEpisodeEnhanced(anime, 1, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get episodes")
}

func TestDownloadEpisodeEnhanced_EpisodeNumTooLow(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AllAnimeType,
	}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	err := DownloadEpisodeEnhanced(anime, 0, "best") // 0 < 1
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDownloadEpisodeEnhanced_EpisodeNumTooHigh(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AllAnimeType,
	}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	err := DownloadEpisodeEnhanced(anime, 5, "best") // only 1 episode available
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDownloadEpisodeEnhanced_ReachesDownloadFromURL(t *testing.T) {
	// Both episode fetch and stream URL succeed; downloadFromURL always
	// returns its placeholder error — verifies we reach that code path.
	combined := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1, URL: "ep1-url"}},
		stURL:    "http://stream.test/ep1.m3u8",
		tp:       scraper.AllAnimeType,
	}
	injectMultiScraper(t, map[scraper.ScraperType]scraper.UnifiedScraper{
		scraper.AllAnimeType: combined,
	})

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto", URL: "naruto-id"}
	err := DownloadEpisodeEnhanced(anime, 1, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not implemented")
}

func TestDownloadEpisodeEnhanced_StreamURLError(t *testing.T) {
	// Episode fetch OK; stream URL fails → error propagates.
	combined := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1, URL: "ep1-url"}},
		stErr:    errors.New("no stream"),
		tp:       scraper.AllAnimeType,
	}
	injectMultiScraper(t, map[scraper.ScraperType]scraper.UnifiedScraper{
		scraper.AllAnimeType: combined,
	})

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto", URL: "naruto-id"}
	err := DownloadEpisodeEnhanced(anime, 1, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get stream URL")
}

// --- DownloadEpisodeRangeEnhanced ---

func TestDownloadEpisodeRangeEnhanced_EpisodesError(t *testing.T) {
	mock := &mockEpiScraper{epsErr: errors.New("upstream failed"), tp: scraper.AllAnimeType}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	err := DownloadEpisodeRangeEnhanced(anime, 1, 2, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get episodes")
}

func TestDownloadEpisodeRangeEnhanced_InvalidRange(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AllAnimeType,
	}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	err := DownloadEpisodeRangeEnhanced(anime, 2, 3, "best") // only 1 ep; 2 > 1
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid range")
}

func TestDownloadEpisodeRangeEnhanced_StartGreaterThanEnd(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}, {Number: "2", Num: 2}},
		tp:       scraper.AllAnimeType,
	}
	injectScraper(t, scraper.AllAnimeType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	err := DownloadEpisodeRangeEnhanced(anime, 3, 1, "best") // startEp > endEp
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid range")
}

func TestDownloadEpisodeRangeEnhanced_LoopIgnoresDownloadError(t *testing.T) {
	// downloadFromURL always errors; the loop swallows it (continue).
	// Function returns nil after the loop completes.
	combined := &mockEpiScraper{
		episodes: []models.Episode{
			{Number: "1", Num: 1, URL: "ep1-url"},
			{Number: "2", Num: 2, URL: "ep2-url"},
		},
		stURL: "http://stream.test/ep.m3u8",
		tp:    scraper.AllAnimeType,
	}
	injectMultiScraper(t, map[scraper.ScraperType]scraper.UnifiedScraper{
		scraper.AllAnimeType: combined,
	})

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto", URL: "naruto-id"}
	err := DownloadEpisodeRangeEnhanced(anime, 1, 2, "best")
	assert.NoError(t, err)
}

func TestDownloadEpisodeRangeEnhanced_StreamURLErrorContinues(t *testing.T) {
	// Stream URL errors for each episode; loop skips them, returns nil.
	combined := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1, URL: "ep1-url"}},
		stErr:    errors.New("no stream"),
		tp:       scraper.AllAnimeType,
	}
	injectMultiScraper(t, map[scraper.ScraperType]scraper.UnifiedScraper{
		scraper.AllAnimeType: combined,
	})

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto", URL: "naruto-id"}
	err := DownloadEpisodeRangeEnhanced(anime, 1, 1, "best")
	assert.NoError(t, err) // errors are continue'd, not returned
}
