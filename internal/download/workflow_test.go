package download

import (
	"context"
	"errors"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/providers/metadata"
	"github.com/alvarorichard/Goanime/internal/downloader"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/player"
	"github.com/alvarorichard/Goanime/internal/scraper"
	"github.com/alvarorichard/Goanime/internal/util"
)

func TestHandleDownloadRequestSingleAnimeUsesLegacyDownloader(t *testing.T) {
	restoreWorkflowState(t)
	installNoopMetadata(t)

	anime := &models.Anime{
		Name:      "Frieren",
		URL:       "https://anime.example/frieren",
		Source:    "AllAnime",
		MediaType: models.MediaTypeAnime,
	}
	episodes := []models.Episode{{Number: "7", Num: 7, URL: "https://anime.example/frieren/7"}}

	var searchedName, legacyURL string
	searchAnimeWithRetry = func(name string) (*models.Anime, error) {
		searchedName = name
		return anime, nil
	}
	getAnimeEpisodesLegacy = func(url string) ([]models.Episode, error) {
		legacyURL = url
		return episodes, nil
	}

	var gotDownloaderURL string
	var gotDownloaderAnime *models.Anime
	var gotSingleEpisode int
	newEpisodeDownloaderWithAnime = func(gotEpisodes []models.Episode, animeURL string, gotAnime *models.Anime) episodeDownloader {
		if len(gotEpisodes) != len(episodes) || gotEpisodes[0].URL != episodes[0].URL {
			t.Fatalf("episodes passed to downloader = %#v, want %#v", gotEpisodes, episodes)
		}
		gotDownloaderURL = animeURL
		gotDownloaderAnime = gotAnime
		return stubEpisodeDownloader{
			downloadSingle: func(episodeNum int) error {
				gotSingleEpisode = episodeNum
				return nil
			},
		}
	}

	err := HandleDownloadRequest(&util.DownloadRequest{
		AnimeName:  "Frieren",
		EpisodeNum: 7,
		Quality:    "720",
	})
	if err != nil {
		t.Fatalf("HandleDownloadRequest() error = %v", err)
	}
	if searchedName != "Frieren" {
		t.Fatalf("search name = %q, want Frieren", searchedName)
	}
	if legacyURL != anime.URL {
		t.Fatalf("legacy episode URL = %q, want %q", legacyURL, anime.URL)
	}
	if gotDownloaderURL != anime.URL || gotDownloaderAnime != anime {
		t.Fatalf("downloader got URL/anime = %q/%#v, want original anime", gotDownloaderURL, gotDownloaderAnime)
	}
	if gotSingleEpisode != 7 {
		t.Fatalf("DownloadSingleEpisode called with %d, want 7", gotSingleEpisode)
	}
}

func TestHandleDownloadRequestAllTreatsUserQuitAsSuccess(t *testing.T) {
	restoreWorkflowState(t)
	installNoopMetadata(t)

	anime := &models.Anime{Name: "Bocchi", URL: "anime-id", Source: "AllAnime", MediaType: models.MediaTypeAnime}
	episodes := []models.Episode{{Number: "1", Num: 1, URL: "episode-1"}}

	searchAnimeWithRetry = func(_ string) (*models.Anime, error) {
		return anime, nil
	}
	getAnimeEpisodesEnhanced = func(gotAnime *models.Anime) ([]models.Episode, error) {
		if gotAnime != anime {
			t.Fatalf("enhanced episode fetch got %#v, want anime", gotAnime)
		}
		return episodes, nil
	}

	var batchCalled bool
	handleBatchDownload = func(gotEpisodes []models.Episode, gotAnime *models.Anime) error {
		batchCalled = true
		if len(gotEpisodes) != 1 || gotEpisodes[0].URL != "episode-1" || gotAnime != anime {
			t.Fatalf("batch download got episodes/anime = %#v/%#v", gotEpisodes, gotAnime)
		}
		return player.ErrUserQuit
	}
	newEpisodeDownloaderWithAnime = func([]models.Episode, string, *models.Anime) episodeDownloader {
		t.Fatal("legacy downloader must not run when batch path returns ErrUserQuit")
		return nil
	}

	if err := HandleDownloadRequest(&util.DownloadRequest{AnimeName: "Bocchi", IsAll: true}); err != nil {
		t.Fatalf("HandleDownloadRequest() error = %v, want nil for ErrUserQuit", err)
	}
	if !batchCalled {
		t.Fatal("batch download path was not called")
	}
}

func TestHandleDownloadRequestRangeFallsBackToLegacyDownloader(t *testing.T) {
	restoreWorkflowState(t)
	installNoopMetadata(t)

	anime := &models.Anime{Name: "Dungeon Meshi", URL: "anime-url", Source: "AllAnime", MediaType: models.MediaTypeAnime}
	enhancedErr := errors.New("enhanced unavailable")

	searchAnimeWithRetry = func(_ string) (*models.Anime, error) {
		return anime, nil
	}
	getAnimeEpisodesEnhanced = func(gotAnime *models.Anime) ([]models.Episode, error) {
		return nil, enhancedErr
	}
	getAnimeEpisodesLegacy = func(url string) ([]models.Episode, error) {
		if url != anime.URL {
			t.Fatalf("legacy fetch URL = %q, want %q", url, anime.URL)
		}
		return []models.Episode{
			{Number: "3", Num: 3, URL: "episode-3"},
			{Number: "4", Num: 4, URL: "episode-4"},
		}, nil
	}

	var gotStart, gotEnd int
	newEpisodeDownloaderWithAnime = func([]models.Episode, string, *models.Anime) episodeDownloader {
		return stubEpisodeDownloader{
			downloadRange: func(startEp, endEp int) error {
				gotStart, gotEnd = startEp, endEp
				return nil
			},
		}
	}

	err := HandleDownloadRequest(&util.DownloadRequest{
		AnimeName:    "Dungeon Meshi",
		IsRange:      true,
		StartEpisode: 3,
		EndEpisode:   4,
	})
	if err != nil {
		t.Fatalf("HandleDownloadRequest() error = %v", err)
	}
	if gotStart != 3 || gotEnd != 4 {
		t.Fatalf("DownloadEpisodeRange called with %d-%d, want 3-4", gotStart, gotEnd)
	}
}

func TestHandleDownloadRequestRoutesNineAnimeRange(t *testing.T) {
	restoreWorkflowState(t)
	installNoopMetadata(t)

	anime := &models.Anime{Name: "Solo Leveling", URL: "solo-leveling-id", Source: "9Anime", MediaType: models.MediaTypeAnime}

	searchAnimeWithRetry = func(_ string) (*models.Anime, error) {
		return anime, nil
	}

	var gotConfig downloader.NineAnimeDownloadConfig
	var gotAnime *models.Anime
	var gotStart, gotEnd int
	newNineAnimeDownloader = func(config downloader.NineAnimeDownloadConfig) nineAnimeDownloader {
		gotConfig = config
		return stubNineAnimeDownloader{
			downloadRange: func(anime *models.Anime, startEp, endEp int) error {
				gotAnime, gotStart, gotEnd = anime, startEp, endEp
				return nil
			},
		}
	}

	err := HandleDownloadRequest(&util.DownloadRequest{
		AnimeName:     "Solo Leveling",
		IsRange:       true,
		StartEpisode:  2,
		EndEpisode:    5,
		Quality:       "1080p",
		SeasonNum:     2,
		SubsLanguage:  "portuguese",
		OutputDir:     "D:/anime",
		AllAnimeSmart: true,
	})
	if err != nil {
		t.Fatalf("HandleDownloadRequest() error = %v", err)
	}
	if gotAnime != anime || gotStart != 2 || gotEnd != 5 {
		t.Fatalf("9Anime range got anime/start/end = %#v/%d/%d", gotAnime, gotStart, gotEnd)
	}
	if gotConfig.AnimeName != anime.Name || gotConfig.Quality != "1080p" || gotConfig.Season != 2 ||
		gotConfig.SubsLanguage != "portuguese" || gotConfig.OutputDir != "D:/anime" {
		t.Fatalf("9Anime config = %+v, want request propagated", gotConfig)
	}
}

func TestHandleDownloadRequestRedirectsMovieContent(t *testing.T) {
	restoreWorkflowState(t)
	installNoopMetadata(t)

	searchAnimeWithRetry = func(_ string) (*models.Anime, error) {
		return &models.Anime{
			Name:      "Inception",
			URL:       "https://flixhq.to/movie/watch-inception-123",
			Source:    "FlixHQ",
			MediaType: models.MediaTypeMovie,
		}, nil
	}

	var gotReq *util.DownloadRequest
	handleMovieDownloadWorkflow = func(req *util.DownloadRequest) error {
		gotReq = req
		return nil
	}

	err := HandleDownloadRequest(&util.DownloadRequest{
		AnimeName:    "Inception",
		Quality:      "720",
		SubsLanguage: "english",
		OutputDir:    "D:/movies",
	})
	if err != nil {
		t.Fatalf("HandleDownloadRequest() error = %v", err)
	}
	if gotReq == nil {
		t.Fatal("movie workflow was not called")
	}
	if !gotReq.IsMovie || gotReq.AnimeName != "Inception" || gotReq.Quality != "720" ||
		gotReq.SubsLanguage != "english" || gotReq.OutputDir != "D:/movies" {
		t.Fatalf("movie redirect request = %+v, want movie request with original options", gotReq)
	}
}

func TestHandleMovieDownloadRequestDownloadsMovieWithoutSelectionUI(t *testing.T) {
	restoreWorkflowState(t)
	installNoopMetadata(t)

	newMediaManager = func() mediaCatalog {
		return stubMediaCatalog{
			results: []*scraper.FlixHQMedia{{
				Title:  "Inception",
				Type:   scraper.MediaTypeMovie,
				URL:    "https://flixhq.to/movie/watch-inception-123",
				Year:   "2010",
				Source: "FlixHQ",
			}},
		}
	}

	var gotConfig downloader.MovieDownloadConfig
	var gotMedia *models.Anime
	newMovieDownloaderWithConfig = func(config downloader.MovieDownloadConfig) movieDownloader {
		gotConfig = config
		return stubMovieDownloader{
			downloadMovie: func(media *models.Anime) error {
				gotMedia = media
				return nil
			},
		}
	}

	err := HandleMovieDownloadRequest(&util.DownloadRequest{
		AnimeName:    "Inception",
		IsMovie:      true,
		Quality:      "720",
		SubsLanguage: "portuguese",
	})
	if err != nil {
		t.Fatalf("HandleMovieDownloadRequest() error = %v", err)
	}
	if gotMedia == nil || gotMedia.Name != "Inception" || gotMedia.MediaType != models.MediaTypeMovie || gotMedia.Source != "FlixHQ" {
		t.Fatalf("movie downloader media = %#v, want selected movie model", gotMedia)
	}
	if gotConfig.Quality != scraper.Quality720 || gotConfig.SubsLanguage != "portuguese" || gotConfig.Provider != "Vidcloud" {
		t.Fatalf("movie downloader config = %+v, want request quality/subs propagated", gotConfig)
	}
}

func TestHandleMovieDownloadRequestDownloadsTVEpisodeWithoutSelectionUI(t *testing.T) {
	restoreWorkflowState(t)
	installNoopMetadata(t)

	newMediaManager = func() mediaCatalog {
		return stubMediaCatalog{
			results: []*scraper.FlixHQMedia{{
				Title:  "Breaking Bad",
				Type:   scraper.MediaTypeTV,
				URL:    "https://flixhq.to/tv/watch-breaking-bad-39506",
				Year:   "2008",
				Source: "FlixHQ",
			}},
		}
	}

	var gotMedia *models.Anime
	var gotSeason, gotEpisode int
	newMovieDownloaderWithConfig = func(_ downloader.MovieDownloadConfig) movieDownloader {
		return stubMovieDownloader{
			downloadTVEpisode: func(media *models.Anime, seasonNum, episodeNum int) error {
				gotMedia, gotSeason, gotEpisode = media, seasonNum, episodeNum
				return nil
			},
		}
	}

	err := HandleMovieDownloadRequest(&util.DownloadRequest{
		AnimeName:  "Breaking Bad",
		IsTV:       true,
		SeasonNum:  2,
		EpisodeNum: 4,
	})
	if err != nil {
		t.Fatalf("HandleMovieDownloadRequest() error = %v", err)
	}
	if gotMedia == nil || gotMedia.Name != "Breaking Bad" || gotMedia.MediaType != models.MediaTypeTV {
		t.Fatalf("TV downloader media = %#v, want selected TV model", gotMedia)
	}
	if gotSeason != 2 || gotEpisode != 4 {
		t.Fatalf("DownloadTVEpisode called with S%dE%d, want S2E4", gotSeason, gotEpisode)
	}
}

type stubEpisodeDownloader struct {
	downloadAll    func() error
	downloadRange  func(startEp, endEp int) error
	downloadSingle func(episodeNum int) error
}

func (s stubEpisodeDownloader) DownloadAllEpisodes() error {
	if s.downloadAll != nil {
		return s.downloadAll()
	}
	return nil
}

func (s stubEpisodeDownloader) DownloadEpisodeRange(startEp, endEp int) error {
	if s.downloadRange != nil {
		return s.downloadRange(startEp, endEp)
	}
	return nil
}

func (s stubEpisodeDownloader) DownloadSingleEpisode(episodeNum int) error {
	if s.downloadSingle != nil {
		return s.downloadSingle(episodeNum)
	}
	return nil
}

type stubNineAnimeDownloader struct {
	downloadAll    func(anime *models.Anime) error
	downloadRange  func(anime *models.Anime, startEp, endEp int) error
	downloadSingle func(anime *models.Anime, episodeNum int) error
}

func (s stubNineAnimeDownloader) DownloadAllEpisodes(anime *models.Anime) error {
	if s.downloadAll != nil {
		return s.downloadAll(anime)
	}
	return nil
}

func (s stubNineAnimeDownloader) DownloadEpisodeRange(anime *models.Anime, startEp, endEp int) error {
	if s.downloadRange != nil {
		return s.downloadRange(anime, startEp, endEp)
	}
	return nil
}

func (s stubNineAnimeDownloader) DownloadSingleEpisode(anime *models.Anime, episodeNum int) error {
	if s.downloadSingle != nil {
		return s.downloadSingle(anime, episodeNum)
	}
	return nil
}

type stubMovieDownloader struct {
	downloadMovie          func(media *models.Anime) error
	downloadTVEpisode      func(media *models.Anime, seasonNum, episodeNum int) error
	downloadTVEpisodeRange func(media *models.Anime, seasonNum, startEp, endEp int) error
	downloadAllSeasons     func(media *models.Anime) error
}

func (s stubMovieDownloader) DownloadMovie(media *models.Anime) error {
	if s.downloadMovie != nil {
		return s.downloadMovie(media)
	}
	return nil
}

func (s stubMovieDownloader) DownloadTVEpisode(media *models.Anime, seasonNum, episodeNum int) error {
	if s.downloadTVEpisode != nil {
		return s.downloadTVEpisode(media, seasonNum, episodeNum)
	}
	return nil
}

func (s stubMovieDownloader) DownloadTVEpisodeRange(media *models.Anime, seasonNum, startEp, endEp int) error {
	if s.downloadTVEpisodeRange != nil {
		return s.downloadTVEpisodeRange(media, seasonNum, startEp, endEp)
	}
	return nil
}

func (s stubMovieDownloader) DownloadAllSeasons(media *models.Anime) error {
	if s.downloadAllSeasons != nil {
		return s.downloadAllSeasons(media)
	}
	return nil
}

type stubMediaCatalog struct {
	results  []*scraper.FlixHQMedia
	seasons  []scraper.FlixHQSeason
	episodes []scraper.FlixHQEpisode
	err      error
}

func (s stubMediaCatalog) SearchMoviesAndTV(_ string) ([]*scraper.FlixHQMedia, error) {
	return s.results, s.err
}

func (s stubMediaCatalog) GetTVSeasons(_ string) ([]scraper.FlixHQSeason, error) {
	return s.seasons, s.err
}

func (s stubMediaCatalog) GetTVEpisodes(_ string) ([]scraper.FlixHQEpisode, error) {
	return s.episodes, s.err
}

func installNoopMetadata(t *testing.T) {
	t.Helper()

	enrichAnimeMetadata = func(context.Context, *models.Anime) ([]metadata.SeasonMapping, error) {
		return nil, nil
	}
	enrichMovieMedia = func(*models.Anime) error {
		return nil
	}
}

func restoreWorkflowState(t *testing.T) {
	t.Helper()

	originalGlobalRequest := util.GlobalDownloadRequest
	originalSearch := searchAnimeWithRetry
	originalEnhancedEpisodes := getAnimeEpisodesEnhanced
	originalLegacyEpisodes := getAnimeEpisodesLegacy
	originalSmartRange := downloadAllAnimeSmartRange
	originalEnhancedRange := downloadEpisodeRangeEnhanced
	originalBatchDownload := handleBatchDownload
	originalBatchDownloadRange := handleBatchDownloadRange
	originalMovieWorkflow := handleMovieDownloadWorkflow
	originalAnimeEnrichment := enrichAnimeMetadata
	originalMovieEnrichment := enrichMovieMedia
	originalEpisodeDownloader := newEpisodeDownloaderWithAnime
	originalNineAnimeDownloader := newNineAnimeDownloader
	originalMediaManager := newMediaManager
	originalMovieDownloader := newMovieDownloaderWithConfig

	t.Cleanup(func() {
		util.GlobalDownloadRequest = originalGlobalRequest
		searchAnimeWithRetry = originalSearch
		getAnimeEpisodesEnhanced = originalEnhancedEpisodes
		getAnimeEpisodesLegacy = originalLegacyEpisodes
		downloadAllAnimeSmartRange = originalSmartRange
		downloadEpisodeRangeEnhanced = originalEnhancedRange
		handleBatchDownload = originalBatchDownload
		handleBatchDownloadRange = originalBatchDownloadRange
		handleMovieDownloadWorkflow = originalMovieWorkflow
		enrichAnimeMetadata = originalAnimeEnrichment
		enrichMovieMedia = originalMovieEnrichment
		newEpisodeDownloaderWithAnime = originalEpisodeDownloader
		newNineAnimeDownloader = originalNineAnimeDownloader
		newMediaManager = originalMediaManager
		newMovieDownloaderWithConfig = originalMovieDownloader
	})
}
