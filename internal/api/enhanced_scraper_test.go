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

func (m *mockEpiScraper) GetStreamURL(_ string, _ ...any) (streamURL string, metadata map[string]string, err error) {
	return m.stURL, nil, m.stErr
}

func (m *mockEpiScraper) GetType() scraper.ScraperType {
	return m.tp
}

// wireEpisodesSeam installs the registry episode seam (episodesFetchFn) so
// DownloadEpisode*Enhanced — which fetch episodes through the registry since
// the per-source switch was deleted in 6.3 — see the given episodes/error
// without importing the providers package (which would cycle).
func wireEpisodesSeam(t *testing.T, eps []models.Episode, err error) {
	t.Helper()
	prev := episodesFetchFn
	episodesFetchFn = func(*models.Anime) ([]models.Episode, error) { return eps, err }
	t.Cleanup(func() { episodesFetchFn = prev })
}

// wireStreamSeam installs the registry stream seam (streamFetchFn) — the
// counterpart for the deleted GetEpisodeStreamURL switch — so
// DownloadEpisode*Enhanced resolve stream URLs without importing providers.
func wireStreamSeam(t *testing.T, url string, err error) {
	t.Helper()
	prev := streamFetchFn
	streamFetchFn = func(*models.Episode, *models.Anime, string) (string, error) { return url, err }
	t.Cleanup(func() { streamFetchFn = prev })
}

// injectScraper wires the registry episode seam from a *mockEpiScraper so
// episode-level DownloadEpisode*Enhanced tests see the mock's episodes/error.
// The scraperType arg is retained for call-site clarity; the ScraperManager it
// once fed no longer exists.
func injectScraper(t *testing.T, _ scraper.ScraperType, mock scraper.UnifiedScraper) {
	t.Helper()
	if m, ok := mock.(*mockEpiScraper); ok {
		wireEpisodesSeam(t, m.episodes, m.epsErr)
	}
}

// injectMultiScraper wires both the episode and stream registry seams from the
// AllAnime mock so stream-reaching DownloadEpisode*Enhanced tests see the mock's
// episodes and stream URL/error.
func injectMultiScraper(t *testing.T, mocks map[scraper.ScraperType]scraper.UnifiedScraper) {
	t.Helper()
	if m, ok := mocks[scraper.AniDBType].(*mockEpiScraper); ok {
		wireEpisodesSeam(t, m.episodes, m.epsErr)
		wireStreamSeam(t, m.stURL, m.stErr)
	}
}

// --- GetSuperFlixEpisodes ---

func TestGetSuperFlixEpisodes_EmptyURL(t *testing.T) {
	t.Parallel()
	media := &models.Anime{Source: "SuperFlix", URL: ""}
	_, err := GetSuperFlixEpisodes(media)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no TMDB ID")
}

func TestGetSuperFlixEpisodes_MovieType(t *testing.T) {
	t.Parallel()
	media := &models.Anime{
		Source:    "SuperFlix",
		URL:       "12345",
		Name:      "Avengers",
		MediaType: models.MediaTypeMovie,
	}
	eps, err := GetSuperFlixEpisodes(media)
	require.NoError(t, err)
	require.Len(t, eps, 1)
	assert.Equal(t, "1", eps[0].Number)
	assert.Equal(t, 1, eps[0].Num)
	assert.Equal(t, "12345", eps[0].URL)
	assert.Equal(t, "Avengers", eps[0].Title.English)
}

// NOTE: the per-source episode SWITCH (GetAnimeEpisodesEnhanced) was deleted in
// Etapa 6.3. Source detection is now source.Resolve (covered by
// providers.TestResolve_LiveRegistry) and dispatch by providers.FetchEpisodes
// (covered by providers.TestFetchEpisodes_*). SuperFlix-specific episode logic
// stays covered by TestGetSuperFlixEpisodes_* above.

// --- DownloadEpisodeEnhanced ---

func TestDownloadEpisodeEnhanced_EpisodesError(t *testing.T) {
	mock := &mockEpiScraper{epsErr: errors.New("upstream failed"), tp: scraper.AniDBType}
	injectScraper(t, scraper.AniDBType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	err := DownloadEpisodeEnhanced(anime, 1, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get episodes")
}

func TestDownloadEpisodeEnhanced_EpisodeNumTooLow(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AniDBType,
	}
	injectScraper(t, scraper.AniDBType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	err := DownloadEpisodeEnhanced(anime, 0, "best") // 0 < 1
	require.Error(t, err)
	assert.Contains(t, err.Error(), "not found")
}

func TestDownloadEpisodeEnhanced_EpisodeNumTooHigh(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AniDBType,
	}
	injectScraper(t, scraper.AniDBType, mock)

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
		tp:       scraper.AniDBType,
	}
	injectMultiScraper(t, map[scraper.ScraperType]scraper.UnifiedScraper{
		scraper.AniDBType: combined,
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
		tp:       scraper.AniDBType,
	}
	injectMultiScraper(t, map[scraper.ScraperType]scraper.UnifiedScraper{
		scraper.AniDBType: combined,
	})

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto", URL: "naruto-id"}
	err := DownloadEpisodeEnhanced(anime, 1, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get stream URL")
}

// --- DownloadEpisodeRangeEnhanced ---

func TestDownloadEpisodeRangeEnhanced_EpisodesError(t *testing.T) {
	mock := &mockEpiScraper{epsErr: errors.New("upstream failed"), tp: scraper.AniDBType}
	injectScraper(t, scraper.AniDBType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	err := DownloadEpisodeRangeEnhanced(anime, 1, 2, "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "failed to get episodes")
}

func TestDownloadEpisodeRangeEnhanced_InvalidRange(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}},
		tp:       scraper.AniDBType,
	}
	injectScraper(t, scraper.AniDBType, mock)

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto"}
	err := DownloadEpisodeRangeEnhanced(anime, 2, 3, "best") // only 1 ep; 2 > 1
	require.Error(t, err)
	assert.Contains(t, err.Error(), "invalid range")
}

func TestDownloadEpisodeRangeEnhanced_StartGreaterThanEnd(t *testing.T) {
	mock := &mockEpiScraper{
		episodes: []models.Episode{{Number: "1", Num: 1}, {Number: "2", Num: 2}},
		tp:       scraper.AniDBType,
	}
	injectScraper(t, scraper.AniDBType, mock)

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
		tp:    scraper.AniDBType,
	}
	injectMultiScraper(t, map[scraper.ScraperType]scraper.UnifiedScraper{
		scraper.AniDBType: combined,
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
		tp:       scraper.AniDBType,
	}
	injectMultiScraper(t, map[scraper.ScraperType]scraper.UnifiedScraper{
		scraper.AniDBType: combined,
	})

	anime := &models.Anime{Source: "AllAnime", Name: "Naruto", URL: "naruto-id"}
	err := DownloadEpisodeRangeEnhanced(anime, 1, 1, "best")
	assert.NoError(t, err) // errors are continue'd, not returned
}
