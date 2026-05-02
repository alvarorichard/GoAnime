package download

import (
	"errors"
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/internal/downloader"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests in this file mutate package-level vars used as DI seams.
// Do not add t.Parallel() — these tests must run sequentially.

type mockEpisodeDownloader struct {
	singleEpisode int
	rangeCalls    [][2]int
	err           error
}

func (m *mockEpisodeDownloader) DownloadEpisodeRange(startEp, endEp int) error {
	m.rangeCalls = append(m.rangeCalls, [2]int{startEp, endEp})
	return m.err
}

func (m *mockEpisodeDownloader) DownloadSingleEpisode(episodeNum int) error {
	m.singleEpisode = episodeNum
	return m.err
}

type mockNineAnimeDownloader struct {
	singleEpisode int
	rangeCalls    [][2]int
	allCalled     bool
	gotAnime      *models.Anime
	err           error
}

func (m *mockNineAnimeDownloader) DownloadEpisodeRange(anime *models.Anime, startEp, endEp int) error {
	m.gotAnime = anime
	m.rangeCalls = append(m.rangeCalls, [2]int{startEp, endEp})
	return m.err
}

func (m *mockNineAnimeDownloader) DownloadAllEpisodes(anime *models.Anime) error {
	m.gotAnime = anime
	m.allCalled = true
	return m.err
}

func (m *mockNineAnimeDownloader) DownloadSingleEpisode(anime *models.Anime, episodeNum int) error {
	m.gotAnime = anime
	m.singleEpisode = episodeNum
	return m.err
}

type mockMovieDownloader struct {
	gotMovie      *models.Anime
	gotTVSeason   int
	gotTVEpisode  int
	gotRangeStart int
	gotRangeEnd   int
	err           error
}

func (m *mockMovieDownloader) DownloadMovie(media *models.Anime) error {
	m.gotMovie = media
	return m.err
}

func (m *mockMovieDownloader) DownloadTVEpisode(media *models.Anime, seasonNum, episodeNum int) error {
	m.gotMovie = media
	m.gotTVSeason = seasonNum
	m.gotTVEpisode = episodeNum
	return m.err
}

func (m *mockMovieDownloader) DownloadTVEpisodeRange(media *models.Anime, seasonNum, startEp, endEp int) error {
	m.gotMovie = media
	m.gotTVSeason = seasonNum
	m.gotRangeStart = startEp
	m.gotRangeEnd = endEp
	return m.err
}

type mockMediaCatalog struct {
	searchResults []*scraper.FlixHQMedia
	searchErr     error
	seasons       []scraper.FlixHQSeason
	seasonsErr    error
	episodes      []scraper.FlixHQEpisode
	episodesErr   error
}

func (m *mockMediaCatalog) SearchMoviesAndTV(query string) ([]*scraper.FlixHQMedia, error) {
	return m.searchResults, m.searchErr
}

func (m *mockMediaCatalog) GetTVSeasons(mediaID string) ([]scraper.FlixHQSeason, error) {
	return m.seasons, m.seasonsErr
}

func (m *mockMediaCatalog) GetTVEpisodes(seasonID string) ([]scraper.FlixHQEpisode, error) {
	return m.episodes, m.episodesErr
}

func restoreWorkflowState() {
	searchAnimeWithRetry = appSearchAnimeWithRetry
	getAnimeEpisodesEnhanced = apiGetAnimeEpisodesEnhanced
	getAnimeEpisodesLegacy = appGetAnimeEpisodesLegacy
	downloadAllAnimeSmartRange = apiDownloadAllAnimeSmartRange
	downloadEpisodeRange = apiDownloadEpisodeRange
	handleBatchDownloadRange = playerHandleBatchDownloadRange
	setAnimeName = playerSetAnimeName
	setExactMediaType = playerSetExactMediaType
	handleMovieDownloadFlow = HandleMovieDownloadRequest
	selectMovieFromResultsFn = selectMovieFromResults
	selectSeasonFn = selectSeason
	selectEpisodeFn = selectEpisode
	newEpisodeDownloader = newDefaultEpisodeDownloader
	newNineAnimeDownloader = newDefaultNineAnimeDownloader
	newMediaCatalog = newDefaultMediaCatalog
	newMovieDownloader = newDefaultMovieDownloader
}

var (
	appSearchAnimeWithRetry        = searchAnimeWithRetry
	apiGetAnimeEpisodesEnhanced    = getAnimeEpisodesEnhanced
	appGetAnimeEpisodesLegacy      = getAnimeEpisodesLegacy
	apiDownloadAllAnimeSmartRange  = downloadAllAnimeSmartRange
	apiDownloadEpisodeRange        = downloadEpisodeRange
	playerHandleBatchDownloadRange = handleBatchDownloadRange
	playerSetAnimeName             = setAnimeName
	playerSetExactMediaType        = setExactMediaType
	newDefaultEpisodeDownloader    = newEpisodeDownloader
	newDefaultNineAnimeDownloader  = newNineAnimeDownloader
	newDefaultMediaCatalog         = newMediaCatalog
	newDefaultMovieDownloader      = newMovieDownloader
)

func TestHandleDownloadRequestRoutesMovieToMovieFlow(t *testing.T) {
	restoreWorkflowState()
	t.Cleanup(restoreWorkflowState)

	sentinel := errors.New("movie workflow called")
	searchAnimeWithRetry = func(name string) (*models.Anime, error) {
		assert.Equal(t, "Inception", name)
		return &models.Anime{Name: "Inception", MediaType: models.MediaTypeMovie}, nil
	}
	setAnimeName = func(string, int) {}
	setExactMediaType = func(string) {}

	var gotRequest *util.DownloadRequest
	handleMovieDownloadFlow = func(request *util.DownloadRequest) error {
		gotRequest = request
		return sentinel
	}

	err := HandleDownloadRequest(&util.DownloadRequest{AnimeName: "Inception"})
	require.ErrorIs(t, err, sentinel)
	require.NotNil(t, gotRequest)
	assert.True(t, gotRequest.IsMovie)
	assert.Equal(t, "best", gotRequest.Quality)
	assert.Equal(t, "Inception", gotRequest.AnimeName)
}

func TestHandleDownloadRequestRoutesNineAnimeDownloader(t *testing.T) {
	restoreWorkflowState()
	t.Cleanup(restoreWorkflowState)

	searchAnimeWithRetry = func(string) (*models.Anime, error) {
		return &models.Anime{Name: "Naruto", Source: "9Anime", MediaType: models.MediaTypeAnime}, nil
	}
	setAnimeName = func(string, int) {}
	setExactMediaType = func(string) {}

	mockDownloader := &mockNineAnimeDownloader{}
	var gotConfig downloader.NineAnimeDownloadConfig
	newNineAnimeDownloader = func(config downloader.NineAnimeDownloadConfig) nineAnimeDownloader {
		gotConfig = config
		return mockDownloader
	}

	err := HandleDownloadRequest(&util.DownloadRequest{
		AnimeName:    "Naruto",
		EpisodeNum:   3,
		SeasonNum:    2,
		SubsLanguage: "portuguese",
		OutputDir:    "D:\\Downloads",
	})
	require.NoError(t, err)
	assert.Equal(t, 3, mockDownloader.singleEpisode)
	assert.Equal(t, "best", gotConfig.Quality)
	assert.Equal(t, 2, gotConfig.Season)
	assert.Equal(t, "portuguese", gotConfig.SubsLanguage)
	assert.Equal(t, "D:\\Downloads", gotConfig.OutputDir)
	require.NotNil(t, mockDownloader.gotAnime)
	assert.Equal(t, "Naruto", mockDownloader.gotAnime.Name)
}

func TestHandleDownloadRequestAllAnimeSmartRangeUsesBatchPathFirst(t *testing.T) {
	restoreWorkflowState()
	t.Cleanup(restoreWorkflowState)

	searchAnimeWithRetry = func(string) (*models.Anime, error) {
		return &models.Anime{Name: "Naruto", Source: "AllAnime", URL: "naruto123abc"}, nil
	}
	setAnimeName = func(string, int) {}
	setExactMediaType = func(string) {}

	getAnimeEpisodesEnhanced = func(anime *models.Anime) ([]models.Episode, error) {
		require.Equal(t, "naruto123abc", anime.URL)
		return []models.Episode{{Number: "1", Num: 1}, {Number: "2", Num: 2}}, nil
	}

	var gotRange [2]int
	handleBatchDownloadRange = func(episodes []models.Episode, anime *models.Anime, startEp, endEp int) error {
		assert.Len(t, episodes, 2)
		require.NotNil(t, anime)
		assert.Equal(t, "naruto123abc", anime.URL)
		gotRange = [2]int{startEp, endEp}
		return nil
	}

	smartCalled := false
	downloadAllAnimeSmartRange = func(*models.Anime, int, int, string) error {
		smartCalled = true
		return nil
	}

	err := HandleDownloadRequest(&util.DownloadRequest{
		AnimeName:     "Naruto",
		IsRange:       true,
		StartEpisode:  1,
		EndEpisode:    2,
		AllAnimeSmart: true,
	})
	require.NoError(t, err)
	assert.Equal(t, [2]int{1, 2}, gotRange)
	assert.False(t, smartCalled)
}

func TestHandleDownloadRequestRangeFallsBackToLegacyDownloader(t *testing.T) {
	restoreWorkflowState()
	t.Cleanup(restoreWorkflowState)

	searchAnimeWithRetry = func(string) (*models.Anime, error) {
		return &models.Anime{Name: "Bleach", Source: "Animefire.io", URL: "https://animefire.plus/bleach"}, nil
	}
	setAnimeName = func(string, int) {}
	setExactMediaType = func(string) {}

	getAnimeEpisodesEnhanced = func(*models.Anime) ([]models.Episode, error) {
		return nil, fmt.Errorf("enhanced unavailable")
	}
	getAnimeEpisodesLegacy = func(url string) ([]models.Episode, error) {
		assert.Equal(t, "https://animefire.plus/bleach", url)
		return []models.Episode{{Number: "1", Num: 1}}, nil
	}

	mockDownloader := &mockEpisodeDownloader{}
	newEpisodeDownloader = func(episodes []models.Episode, animeURL string, anime *models.Anime) episodeDownloader {
		assert.Len(t, episodes, 1)
		assert.Equal(t, "https://animefire.plus/bleach", animeURL)
		assert.Equal(t, "Bleach", anime.Name)
		return mockDownloader
	}

	err := HandleDownloadRequest(&util.DownloadRequest{
		AnimeName:    "Bleach",
		IsRange:      true,
		StartEpisode: 2,
		EndEpisode:   5,
	})
	require.NoError(t, err)
	require.Len(t, mockDownloader.rangeCalls, 1)
	assert.Equal(t, [2]int{2, 5}, mockDownloader.rangeCalls[0])
}

func TestHandleDownloadRequestSingleEpisodeUsesLegacyDownloader(t *testing.T) {
	restoreWorkflowState()
	t.Cleanup(restoreWorkflowState)

	searchAnimeWithRetry = func(string) (*models.Anime, error) {
		return &models.Anime{Name: "One Piece", Source: "Animefire.io", URL: "https://animefire.plus/one-piece"}, nil
	}
	setAnimeName = func(string, int) {}
	setExactMediaType = func(string) {}
	getAnimeEpisodesLegacy = func(string) ([]models.Episode, error) {
		return []models.Episode{{Number: "7", Num: 7}}, nil
	}

	mockDownloader := &mockEpisodeDownloader{}
	newEpisodeDownloader = func([]models.Episode, string, *models.Anime) episodeDownloader {
		return mockDownloader
	}

	err := HandleDownloadRequest(&util.DownloadRequest{
		AnimeName:  "One Piece",
		EpisodeNum: 7,
	})
	require.NoError(t, err)
	assert.Equal(t, 7, mockDownloader.singleEpisode)
}

func TestHandleMovieDownloadRequestRoutesMovieDownload(t *testing.T) {
	restoreWorkflowState()
	t.Cleanup(restoreWorkflowState)

	selected := &scraper.FlixHQMedia{
		Title:  "Inception",
		Type:   scraper.MediaTypeMovie,
		URL:    "https://flixhq.to/movie/inception-12345",
		Source: "FlixHQ",
	}

	newMediaCatalog = func() mediaCatalog {
		return &mockMediaCatalog{searchResults: []*scraper.FlixHQMedia{selected}}
	}
	selectMovieFromResultsFn = func(results []*scraper.FlixHQMedia, preferMovie, preferTV bool) (*scraper.FlixHQMedia, error) {
		assert.Len(t, results, 1)
		assert.True(t, preferMovie)
		assert.False(t, preferTV)
		return selected, nil
	}

	var gotExactType, gotAnimeName string
	var gotSeason int
	setExactMediaType = func(mediaType string) { gotExactType = mediaType }
	setAnimeName = func(name string, season int) {
		gotAnimeName = name
		gotSeason = season
	}

	mockDownloader := &mockMovieDownloader{}
	var gotConfig downloader.MovieDownloadConfig
	newMovieDownloader = func(config downloader.MovieDownloadConfig) movieDownloader {
		gotConfig = config
		return mockDownloader
	}

	err := HandleMovieDownloadRequest(&util.DownloadRequest{
		AnimeName: "Inception",
		IsMovie:   true,
	})
	require.NoError(t, err)
	assert.Equal(t, "movie", gotExactType)
	assert.Equal(t, "Inception", gotAnimeName)
	assert.Equal(t, 0, gotSeason)
	assert.Equal(t, scraper.Quality("1080"), gotConfig.Quality)
	assert.Equal(t, "english", gotConfig.SubsLanguage)
	require.NotNil(t, mockDownloader.gotMovie)
	assert.Equal(t, "Inception", mockDownloader.gotMovie.Name)
}

func TestHandleMovieDownloadRequestRoutesTVRangeDownload(t *testing.T) {
	restoreWorkflowState()
	t.Cleanup(restoreWorkflowState)

	selected := &scraper.FlixHQMedia{
		Title:  "Dexter",
		Type:   scraper.MediaTypeTV,
		URL:    "https://flixhq.to/tv/dexter-39448",
		Source: "FlixHQ",
	}

	newMediaCatalog = func() mediaCatalog {
		return &mockMediaCatalog{searchResults: []*scraper.FlixHQMedia{selected}}
	}
	selectMovieFromResultsFn = func(results []*scraper.FlixHQMedia, preferMovie, preferTV bool) (*scraper.FlixHQMedia, error) {
		assert.Len(t, results, 1)
		assert.False(t, preferMovie)
		assert.True(t, preferTV)
		return selected, nil
	}
	setExactMediaType = func(string) {}
	setAnimeName = func(string, int) {}

	mockDownloader := &mockMovieDownloader{}
	newMovieDownloader = func(config downloader.MovieDownloadConfig) movieDownloader {
		assert.Equal(t, scraper.Quality("720"), config.Quality)
		assert.Equal(t, "spanish", config.SubsLanguage)
		return mockDownloader
	}

	err := HandleMovieDownloadRequest(&util.DownloadRequest{
		AnimeName:    "Dexter",
		IsTV:         true,
		IsRange:      true,
		SeasonNum:    2,
		StartEpisode: 4,
		EndEpisode:   6,
		Quality:      "720",
		SubsLanguage: "spanish",
	})
	require.NoError(t, err)
	require.NotNil(t, mockDownloader.gotMovie)
	assert.Equal(t, "Dexter", mockDownloader.gotMovie.Name)
	assert.Equal(t, 2, mockDownloader.gotTVSeason)
	assert.Equal(t, 4, mockDownloader.gotRangeStart)
	assert.Equal(t, 6, mockDownloader.gotRangeEnd)
}

func TestSelectMovieFromResultsPrefersMovieWhenFilterLeavesSingleResult(t *testing.T) {
	results := []*scraper.FlixHQMedia{
		{Title: "Dexter", Type: scraper.MediaTypeTV},
		{Title: "Inception", Type: scraper.MediaTypeMovie},
	}

	selected, err := selectMovieFromResults(results, true, false)
	require.NoError(t, err)
	require.NotNil(t, selected)
	assert.Equal(t, "Inception", selected.Title)
}

func TestSelectSeasonSingleSeasonSkipsPrompt(t *testing.T) {
	mm := &mockMediaCatalog{
		seasons: []scraper.FlixHQSeason{{ID: "season-1", Title: "Season 1"}},
	}

	season, err := selectSeason(mm, "39448")
	require.NoError(t, err)
	assert.Equal(t, 1, season)
}

func TestSelectEpisodeValidatesSeasonIndexBeforePrompt(t *testing.T) {
	mm := &mockMediaCatalog{
		seasons: []scraper.FlixHQSeason{{ID: "season-1", Title: "Season 1"}},
	}

	_, err := selectEpisode(mm, "39448", 2)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "season 2 not found")
}

func TestExtractIDFromURL(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "12345", extractIDFromURL("https://flixhq.to/movie/inception-12345"))
	assert.Equal(t, "98765", extractIDFromURL("watch-dexter-98765"))
	assert.Equal(t, "", extractIDFromURL(""))
}

func TestHandleDownloadRequestTreatsUserQuitAsSuccessOnBatchPath(t *testing.T) {
	restoreWorkflowState()
	t.Cleanup(restoreWorkflowState)

	searchAnimeWithRetry = func(string) (*models.Anime, error) {
		return &models.Anime{Name: "Naruto", Source: "Animefire.io", URL: "https://animefire.plus/naruto"}, nil
	}
	setAnimeName = func(string, int) {}
	setExactMediaType = func(string) {}
	getAnimeEpisodesEnhanced = func(*models.Anime) ([]models.Episode, error) {
		return []models.Episode{{Number: "1", Num: 1}}, nil
	}
	handleBatchDownloadRange = func([]models.Episode, *models.Anime, int, int) error {
		return player.ErrUserQuit
	}

	err := HandleDownloadRequest(&util.DownloadRequest{
		AnimeName:    "Naruto",
		IsRange:      true,
		StartEpisode: 1,
		EndEpisode:   1,
	})
	require.NoError(t, err)
}
