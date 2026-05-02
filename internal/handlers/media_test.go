package handlers

import (
	"context"
	"fmt"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type mockMediaCatalog struct {
	searchAnimeOnlyResults []*models.Anime
	searchAnimeOnlyErr     error
	searchMovieResults     []*scraper.FlixHQMedia
	searchMovieErr         error
	searchAllResults       []*models.Anime
	searchAllErr           error
	seasons                []scraper.FlixHQSeason
	seasonsErr             error
	episodes               []scraper.FlixHQEpisode
	episodesErr            error
	movieStreamInfo        *scraper.FlixHQStreamInfo
	movieStreamErr         error
	tvStreamInfo           *scraper.FlixHQStreamInfo
	tvStreamErr            error
	streamWithQualityInfo  *scraper.FlixHQStreamInfo
	streamWithQualityErr   error
	qualities              []scraper.Quality
	qualitiesErr           error
	animeStreamURL         string
	animeStreamMetadata    map[string]string
	animeStreamErr         error
	movieQualities         []scraper.QualityOption
	movieQualitiesErr      error
	movieStreamWithQuality *scraper.FlixHQStreamInfo
	movieStreamWithErr     error
	episodeQualities       []scraper.QualityOption
	episodeQualitiesErr    error

	gotSearchAnimeQuery string
	gotSearchMovieQuery string
	gotSearchAllQuery   string
	gotMovieStreamID    string
	gotTVEpisodeID      string
	gotStreamQualityID  string
	gotAnimeEpisodeNum  string
	gotAnimeMode        string
}

func (m *mockMediaCatalog) SearchAnimeOnly(query string) ([]*models.Anime, error) {
	m.gotSearchAnimeQuery = query
	return m.searchAnimeOnlyResults, m.searchAnimeOnlyErr
}

func (m *mockMediaCatalog) SearchMoviesAndTV(query string) ([]*scraper.FlixHQMedia, error) {
	m.gotSearchMovieQuery = query
	return m.searchMovieResults, m.searchMovieErr
}

func (m *mockMediaCatalog) SearchAll(query string) ([]*models.Anime, error) {
	m.gotSearchAllQuery = query
	return m.searchAllResults, m.searchAllErr
}

func (m *mockMediaCatalog) GetTVSeasons(mediaID string) ([]scraper.FlixHQSeason, error) {
	return m.seasons, m.seasonsErr
}

func (m *mockMediaCatalog) GetTVEpisodes(seasonID string) ([]scraper.FlixHQEpisode, error) {
	return m.episodes, m.episodesErr
}

func (m *mockMediaCatalog) GetMovieStreamInfo(mediaID, provider, quality, subsLanguage string) (*scraper.FlixHQStreamInfo, error) {
	m.gotMovieStreamID = mediaID
	return m.movieStreamInfo, m.movieStreamErr
}

func (m *mockMediaCatalog) GetTVEpisodeStreamInfo(dataID, provider, quality, subsLanguage string) (*scraper.FlixHQStreamInfo, error) {
	m.gotTVEpisodeID = dataID
	return m.tvStreamInfo, m.tvStreamErr
}

func (m *mockMediaCatalog) GetStreamWithQuality(episodeID string, isMovie bool, quality scraper.Quality, subsLanguage string) (*scraper.FlixHQStreamInfo, error) {
	m.gotStreamQualityID = episodeID
	return m.streamWithQualityInfo, m.streamWithQualityErr
}

func (m *mockMediaCatalog) GetStreamWithQualityWithContext(ctx context.Context, episodeID string, isMovie bool, quality scraper.Quality, subsLanguage string) (*scraper.FlixHQStreamInfo, error) {
	m.gotStreamQualityID = episodeID
	return m.streamWithQualityInfo, m.streamWithQualityErr
}

func (m *mockMediaCatalog) GetAvailableQualities(episodeID string, isMovie bool) ([]scraper.Quality, error) {
	return m.qualities, m.qualitiesErr
}

func (m *mockMediaCatalog) GetAnimeStreamURL(anime *models.Anime, episodeNum string, quality, mode string) (string, map[string]string, error) {
	m.gotAnimeEpisodeNum = episodeNum
	m.gotAnimeMode = mode
	return m.animeStreamURL, m.animeStreamMetadata, m.animeStreamErr
}

func (m *mockMediaCatalog) GetMovieQualities(mediaID string) ([]scraper.QualityOption, error) {
	return m.movieQualities, m.movieQualitiesErr
}

func (m *mockMediaCatalog) GetMovieStreamWithQuality(mediaID string, quality scraper.Quality, subsLanguage string) (*scraper.FlixHQStreamInfo, error) {
	return m.movieStreamWithQuality, m.movieStreamWithErr
}

func (m *mockMediaCatalog) GetEpisodeQualities(dataID string) ([]scraper.QualityOption, error) {
	return m.episodeQualities, m.episodeQualitiesErr
}

func TestMediaHandlerSetOptions(t *testing.T) {
	t.Parallel()

	handler := &MediaHandler{
		provider:     "Vidcloud",
		quality:      scraper.Quality720,
		subsLanguage: "english",
	}

	handler.SetOptions("UpCloud", "1080", "spanish")
	assert.Equal(t, "UpCloud", handler.provider)
	assert.Equal(t, scraper.Quality1080, handler.quality)
	assert.Equal(t, "spanish", handler.subsLanguage)

	handler.SetOptions("", "", "")
	assert.Equal(t, "UpCloud", handler.provider)
	assert.Equal(t, scraper.Quality1080, handler.quality)
	assert.Equal(t, "spanish", handler.subsLanguage)
}

func TestSearchMediaRoutesByContentType(t *testing.T) {
	t.Parallel()

	manager := &mockMediaCatalog{
		searchAnimeOnlyResults: []*models.Anime{{Name: "Naruto"}},
		searchMovieResults: []*scraper.FlixHQMedia{{
			Title:  "Dexter",
			Type:   scraper.MediaTypeTV,
			URL:    "https://flixhq.to/tv/dexter-39448",
			Source: "FlixHQ",
		}},
		searchAllResults: []*models.Anime{{Name: "Bleach"}},
	}
	handler := &MediaHandler{mediaManager: manager}

	animeResults, err := handler.SearchMedia("naruto", models.MediaTypeAnime)
	require.NoError(t, err)
	require.Len(t, animeResults, 1)
	assert.Equal(t, "naruto", manager.gotSearchAnimeQuery)

	movieResults, err := handler.SearchMedia("dexter", models.MediaTypeTV)
	require.NoError(t, err)
	require.Len(t, movieResults, 1)
	assert.Equal(t, "Dexter", movieResults[0].Name)
	assert.Equal(t, models.MediaTypeTV, movieResults[0].MediaType)
	assert.Equal(t, "dexter", manager.gotSearchMovieQuery)

	allResults, err := handler.SearchMedia("bleach", "")
	require.NoError(t, err)
	require.Len(t, allResults, 1)
	assert.Equal(t, "bleach", manager.gotSearchAllQuery)
}

func TestGetStreamInfoWithContextRejectsNonFlixHQSources(t *testing.T) {
	t.Parallel()

	handler := &MediaHandler{mediaManager: &mockMediaCatalog{}}
	_, err := handler.GetStreamInfoWithContext(context.Background(), &models.Anime{
		Name:      "Naruto",
		Source:    "Animefire.io",
		MediaType: models.MediaTypeAnime,
	}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "does not support FlixHQ streaming")
}

func TestGetStreamInfoWithContextMovieUsesMovieStreamInfo(t *testing.T) {
	t.Parallel()

	manager := &mockMediaCatalog{
		movieStreamInfo: &scraper.FlixHQStreamInfo{VideoURL: "https://example.com/movie.m3u8"},
	}
	handler := &MediaHandler{
		mediaManager: manager,
		provider:     "Vidcloud",
		quality:      scraper.Quality1080,
		subsLanguage: "english",
	}

	info, err := handler.GetStreamInfoWithContext(context.Background(), &models.Anime{
		Name:      "Inception",
		Source:    "FlixHQ",
		MediaType: models.MediaTypeMovie,
		URL:       "https://flixhq.to/movie/inception-12345",
	}, nil)
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "https://example.com/movie.m3u8", info.VideoURL)
	assert.Equal(t, "12345", manager.gotMovieStreamID)
}

func TestGetStreamInfoWithContextTVRequiresEpisode(t *testing.T) {
	t.Parallel()

	handler := &MediaHandler{mediaManager: &mockMediaCatalog{}}
	_, err := handler.GetStreamInfoWithContext(context.Background(), &models.Anime{
		Name:      "Dexter",
		Source:    "FlixHQ",
		MediaType: models.MediaTypeTV,
		URL:       "https://flixhq.to/tv/dexter-39448",
	}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "episode is required")
}

func TestGetStreamInfoWithContextTVUsesEpisodeDataID(t *testing.T) {
	t.Parallel()

	manager := &mockMediaCatalog{
		tvStreamInfo: &scraper.FlixHQStreamInfo{VideoURL: "https://example.com/tv.m3u8"},
	}
	handler := &MediaHandler{
		mediaManager: manager,
		provider:     "Vidcloud",
		quality:      scraper.Quality720,
		subsLanguage: "portuguese",
	}

	info, err := handler.GetStreamInfoWithContext(context.Background(), &models.Anime{
		Name:      "Dexter",
		Source:    "FlixHQ",
		MediaType: models.MediaTypeTV,
		URL:       "https://flixhq.to/tv/dexter-39448",
	}, &scraper.FlixHQEpisode{DataID: "ep-data-id"})
	require.NoError(t, err)
	require.NotNil(t, info)
	assert.Equal(t, "https://example.com/tv.m3u8", info.VideoURL)
	assert.Equal(t, "ep-data-id", manager.gotTVEpisodeID)
}

func TestMediaHandlerDelegatesStreamHelpers(t *testing.T) {
	t.Parallel()

	manager := &mockMediaCatalog{
		streamWithQualityInfo: &scraper.FlixHQStreamInfo{VideoURL: "https://example.com/quality.m3u8"},
		qualities:             []scraper.Quality{scraper.Quality720, scraper.Quality1080},
		animeStreamURL:        "https://example.com/anime.m3u8",
		animeStreamMetadata:   map[string]string{"referer": "https://example.com"},
	}
	handler := &MediaHandler{
		mediaManager: manager,
		quality:      scraper.Quality1080,
		subsLanguage: "english",
	}

	info, err := handler.GetStreamWithQuality("episode-id", false)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/quality.m3u8", info.VideoURL)
	assert.Equal(t, "episode-id", manager.gotStreamQualityID)

	info, err = handler.GetStreamWithQualityContext(context.Background(), "episode-id-ctx", true)
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/quality.m3u8", info.VideoURL)
	assert.Equal(t, "episode-id-ctx", manager.gotStreamQualityID)

	qualities, err := handler.GetAvailableQualities("episode-id", false)
	require.NoError(t, err)
	assert.Equal(t, []scraper.Quality{scraper.Quality720, scraper.Quality1080}, qualities)

	streamURL, metadata, err := handler.GetAnimeStreamURL(&models.Anime{Name: "Naruto", Source: "AllAnime"}, "12", "dub")
	require.NoError(t, err)
	assert.Equal(t, "https://example.com/anime.m3u8", streamURL)
	assert.Equal(t, map[string]string{"referer": "https://example.com"}, metadata)
	assert.Equal(t, "12", manager.gotAnimeEpisodeNum)
	assert.Equal(t, "dub", manager.gotAnimeMode)
}

func TestHelperConversions(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "12345", extractIDFromURL("https://flixhq.to/movie/inception-12345"))
	assert.Equal(t, "", extractIDFromURL(""))

	subs := convertSubtitles([]scraper.FlixHQSubtitle{{
		URL:      "https://example.com/sub.vtt",
		Language: "eng",
		Label:    "English",
	}})
	require.Len(t, subs, 1)
	assert.Equal(t, models.Subtitle{
		URL:      "https://example.com/sub.vtt",
		Language: "eng",
		Label:    "English",
	}, subs[0])
}

func TestGetStreamInfoWithContextFailsWhenMediaIDCannotBeExtracted(t *testing.T) {
	t.Parallel()

	handler := &MediaHandler{mediaManager: &mockMediaCatalog{}}
	_, err := handler.GetStreamInfoWithContext(context.Background(), &models.Anime{
		Name:      "Inception",
		Source:    "FlixHQ",
		MediaType: models.MediaTypeMovie,
		URL:       "",
	}, nil)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "could not extract media ID")
}

func TestSearchMediaPropagatesMovieSearchError(t *testing.T) {
	t.Parallel()

	handler := &MediaHandler{mediaManager: &mockMediaCatalog{searchMovieErr: fmt.Errorf("movie search failed")}}
	_, err := handler.SearchMedia("dexter", models.MediaTypeMovie)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "movie search failed")
}
