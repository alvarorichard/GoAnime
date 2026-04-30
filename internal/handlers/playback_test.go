package handlers

import (
	"errors"
	"reflect"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api"
	"github.com/alvarorichard/Goanime/internal/models"
)

type stubPlaybackModeDiscordManager struct {
	enabled         bool
	initializeCalls int
	shutdownCalls   int
}

func (m *stubPlaybackModeDiscordManager) Initialize() error {
	m.initializeCalls++
	return nil
}

func (m *stubPlaybackModeDiscordManager) Shutdown() {
	m.shutdownCalls++
}

func (m *stubPlaybackModeDiscordManager) IsEnabled() bool {
	return m.enabled
}

func installPlaybackModeTestHooks(t *testing.T) *stubPlaybackModeDiscordManager {
	t.Helper()

	oldInitLogger := playbackModeInitLogger
	oldPreWarmConnections := playbackModePreWarmConnections
	oldHandleTrackingNotice := playbackModeHandleTrackingNotice
	oldNewDiscordManager := playbackModeNewDiscordManager
	oldSearchAnimeWithRetry := playbackModeSearchAnimeWithRetry
	oldFetchAnimeDetails := playbackModeFetchAnimeDetails
	oldGetAnimeEpisodes := playbackModeGetAnimeEpisodes
	oldHandleSeries := playbackModeHandleSeries
	oldHandleMovie := playbackModeHandleMovie

	manager := &stubPlaybackModeDiscordManager{enabled: true}
	unexpectedCall := errors.New("unexpected playback mode test hook call")

	playbackModeInitLogger = func() {}
	playbackModePreWarmConnections = func() {}
	playbackModeHandleTrackingNotice = func() {}
	playbackModeNewDiscordManager = func() playbackModeDiscordManager {
		return manager
	}
	playbackModeSearchAnimeWithRetry = func(name string) (*models.Anime, error) {
		t.Errorf("unexpected search for %q", name)
		return nil, unexpectedCall
	}
	playbackModeFetchAnimeDetails = func(anime *models.Anime) {}
	playbackModeGetAnimeEpisodes = func(anime *models.Anime) ([]models.Episode, error) {
		t.Errorf("unexpected episode fetch for %#v", anime)
		return nil, unexpectedCall
	}
	playbackModeHandleSeries = func(anime *models.Anime, episodes []models.Episode, totalEpisodes int, discordEnabled bool) error {
		t.Errorf("unexpected series playback for %#v", anime)
		return nil
	}
	playbackModeHandleMovie = func(anime *models.Anime, episodes []models.Episode, discordEnabled bool) error {
		t.Errorf("unexpected movie playback for %#v", anime)
		return nil
	}

	t.Cleanup(func() {
		playbackModeInitLogger = oldInitLogger
		playbackModePreWarmConnections = oldPreWarmConnections
		playbackModeHandleTrackingNotice = oldHandleTrackingNotice
		playbackModeNewDiscordManager = oldNewDiscordManager
		playbackModeSearchAnimeWithRetry = oldSearchAnimeWithRetry
		playbackModeFetchAnimeDetails = oldFetchAnimeDetails
		playbackModeGetAnimeEpisodes = oldGetAnimeEpisodes
		playbackModeHandleSeries = oldHandleSeries
		playbackModeHandleMovie = oldHandleMovie
	})

	return manager
}

func TestHandlePlaybackModeSeriesAnimeUsesSeriesHandler(t *testing.T) {
	manager := installPlaybackModeTestHooks(t)
	anime := &models.Anime{Name: "Series Anime", Source: "AllAnime", MediaType: models.MediaTypeAnime}
	episodes := []models.Episode{
		{Number: "1", URL: "https://example.test/1"},
		{Number: "2", URL: "https://example.test/2"},
		{Number: "3", URL: "https://example.test/3"},
	}

	searchCalls := 0
	playbackModeSearchAnimeWithRetry = func(name string) (*models.Anime, error) {
		searchCalls++
		if name != "series query" {
			t.Errorf("search name = %q, want %q", name, "series query")
		}
		return anime, nil
	}
	playbackModeGetAnimeEpisodes = func(got *models.Anime) ([]models.Episode, error) {
		if got != anime {
			t.Errorf("episode fetch anime = %#v, want %#v", got, anime)
		}
		return episodes, nil
	}

	seriesCalls := 0
	var gotEpisodes []models.Episode
	var gotTotal int
	var gotDiscordEnabled bool
	playbackModeHandleSeries = func(got *models.Anime, eps []models.Episode, totalEpisodes int, discordEnabled bool) error {
		seriesCalls++
		if got != anime {
			t.Errorf("series anime = %#v, want %#v", got, anime)
		}
		gotEpisodes = append([]models.Episode(nil), eps...)
		gotTotal = totalEpisodes
		gotDiscordEnabled = discordEnabled
		return nil
	}
	playbackModeHandleMovie = func(anime *models.Anime, episodes []models.Episode, discordEnabled bool) error {
		t.Errorf("movie playback called for series anime")
		return nil
	}

	HandlePlaybackMode("series query")

	if searchCalls != 1 {
		t.Fatalf("search calls = %d, want 1", searchCalls)
	}
	if seriesCalls != 1 {
		t.Fatalf("series playback calls = %d, want 1", seriesCalls)
	}
	if gotTotal != len(episodes) {
		t.Errorf("series total episodes = %d, want %d", gotTotal, len(episodes))
	}
	if !reflect.DeepEqual(gotEpisodes, episodes) {
		t.Errorf("series episodes = %#v, want %#v", gotEpisodes, episodes)
	}
	if !gotDiscordEnabled {
		t.Errorf("discord enabled = false, want true")
	}
	if manager.initializeCalls != 1 || manager.shutdownCalls != 1 {
		t.Errorf("discord manager calls = init:%d shutdown:%d, want 1 each", manager.initializeCalls, manager.shutdownCalls)
	}
}

func TestHandlePlaybackModeMovieOrSingleEpisodeUsesMovieHandler(t *testing.T) {
	tests := []struct {
		name     string
		anime    *models.Anime
		episodes []models.Episode
	}{
		{
			name:  "movie media type",
			anime: &models.Anime{Name: "Standalone Movie", Source: "FlixHQ", MediaType: models.MediaTypeMovie},
			episodes: []models.Episode{
				{Number: "1", URL: "https://example.test/movie"},
				{Number: "2", URL: "https://example.test/extra"},
			},
		},
		{
			name:  "single anime episode",
			anime: &models.Anime{Name: "OVA", Source: "AllAnime", MediaType: models.MediaTypeAnime},
			episodes: []models.Episode{
				{Number: "1", URL: "https://example.test/ova"},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			installPlaybackModeTestHooks(t)

			playbackModeSearchAnimeWithRetry = func(name string) (*models.Anime, error) {
				if name != "movie query" {
					t.Errorf("search name = %q, want %q", name, "movie query")
				}
				return tt.anime, nil
			}
			playbackModeGetAnimeEpisodes = func(got *models.Anime) ([]models.Episode, error) {
				if got != tt.anime {
					t.Errorf("episode fetch anime = %#v, want %#v", got, tt.anime)
				}
				return tt.episodes, nil
			}

			movieCalls := 0
			var gotEpisodes []models.Episode
			playbackModeHandleMovie = func(got *models.Anime, eps []models.Episode, discordEnabled bool) error {
				movieCalls++
				if got != tt.anime {
					t.Errorf("movie anime = %#v, want %#v", got, tt.anime)
				}
				gotEpisodes = append([]models.Episode(nil), eps...)
				return nil
			}
			playbackModeHandleSeries = func(anime *models.Anime, episodes []models.Episode, totalEpisodes int, discordEnabled bool) error {
				t.Errorf("series playback called for %s", tt.name)
				return nil
			}

			HandlePlaybackMode("movie query")

			if movieCalls != 1 {
				t.Fatalf("movie playback calls = %d, want 1", movieCalls)
			}
			if !reflect.DeepEqual(gotEpisodes, tt.episodes) {
				t.Errorf("movie episodes = %#v, want %#v", gotEpisodes, tt.episodes)
			}
		})
	}
}

func TestHandlePlaybackModeBackToSearchContinuesWithFreshSearch(t *testing.T) {
	installPlaybackModeTestHooks(t)

	firstAnime := &models.Anime{Name: "Needs Season Selection", Source: "FlixHQ", MediaType: models.MediaTypeTV}
	secondAnime := &models.Anime{Name: "Fresh Result", Source: "AllAnime", MediaType: models.MediaTypeAnime}
	secondEpisodes := []models.Episode{
		{Number: "1", URL: "https://example.test/fresh-1"},
		{Number: "2", URL: "https://example.test/fresh-2"},
	}

	searchNames := []string{}
	playbackModeSearchAnimeWithRetry = func(name string) (*models.Anime, error) {
		searchNames = append(searchNames, name)
		switch len(searchNames) {
		case 1:
			return firstAnime, nil
		case 2:
			return secondAnime, nil
		default:
			t.Errorf("unexpected extra search for %q", name)
			return nil, errors.New("unexpected extra search")
		}
	}

	episodeFetchCalls := 0
	playbackModeGetAnimeEpisodes = func(anime *models.Anime) ([]models.Episode, error) {
		episodeFetchCalls++
		switch episodeFetchCalls {
		case 1:
			if anime != firstAnime {
				t.Errorf("first episode fetch anime = %#v, want %#v", anime, firstAnime)
			}
			return nil, api.ErrBackToSearch
		case 2:
			if anime != secondAnime {
				t.Errorf("second episode fetch anime = %#v, want %#v", anime, secondAnime)
			}
			return secondEpisodes, nil
		default:
			t.Errorf("unexpected extra episode fetch for %#v", anime)
			return nil, errors.New("unexpected extra episode fetch")
		}
	}

	seriesCalls := 0
	playbackModeHandleSeries = func(anime *models.Anime, episodes []models.Episode, totalEpisodes int, discordEnabled bool) error {
		seriesCalls++
		if anime != secondAnime {
			t.Errorf("series anime = %#v, want %#v", anime, secondAnime)
		}
		if totalEpisodes != len(secondEpisodes) {
			t.Errorf("series total episodes = %d, want %d", totalEpisodes, len(secondEpisodes))
		}
		return nil
	}
	playbackModeHandleMovie = func(anime *models.Anime, episodes []models.Episode, discordEnabled bool) error {
		t.Errorf("movie playback called after back-to-search")
		return nil
	}

	HandlePlaybackMode("initial query")

	wantSearchNames := []string{"initial query", ""}
	if !reflect.DeepEqual(searchNames, wantSearchNames) {
		t.Fatalf("search names = %#v, want %#v", searchNames, wantSearchNames)
	}
	if episodeFetchCalls != 2 {
		t.Fatalf("episode fetch calls = %d, want 2", episodeFetchCalls)
	}
	if seriesCalls != 1 {
		t.Fatalf("series playback calls = %d, want 1", seriesCalls)
	}
}
