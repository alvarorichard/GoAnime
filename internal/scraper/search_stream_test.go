package scraper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------------
// AnimefireClient.GetAnimeEpisodes
// ---------------------------------------------------------------------------

func TestAnimefireClient_GetAnimeEpisodes_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><body>`+
			`<a class="lEp epT divNumEp smallbox px-2 mx-1 text-left d-flex" href="/ep/1">Episódio 1</a>`+
			`<a class="lEp epT divNumEp smallbox px-2 mx-1 text-left d-flex" href="/ep/2">Episódio 2</a>`+
			`</body></html>`)
	}))
	t.Cleanup(srv.Close)

	client := NewAnimefireClient()
	client.baseURL = srv.URL
	client.maxRetries = 0
	client.retryDelay = 0

	episodes, err := client.GetAnimeEpisodes(srv.URL + "/anime/test")
	require.NoError(t, err)
	assert.Len(t, episodes, 2)
}

func TestAnimefireClient_GetAnimeEpisodes_HTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	client := NewAnimefireClient()
	client.baseURL = srv.URL
	client.maxRetries = 0
	client.retryDelay = 0

	_, err := client.GetAnimeEpisodes(srv.URL + "/anime/test")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// AnimefireClient.parseEpisodes (covered via GetAnimeEpisodes above, but
// also tested directly via a goquery document)
// ---------------------------------------------------------------------------

func TestAnimefireClient_ParseEpisodes_ReturnsEpisodes(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><body>`+
			`<a class="lEp epT divNumEp smallbox px-2 mx-1 text-left d-flex" href="/ep/5">Episódio 5</a>`+
			`<a class="lEp epT divNumEp smallbox px-2 mx-1 text-left d-flex" href="/ep/6">Episódio 6</a>`+
			`</body></html>`)
	}))
	t.Cleanup(srv.Close)

	client := NewAnimefireClient()
	client.baseURL = srv.URL
	client.maxRetries = 0
	client.retryDelay = 0

	episodes, err := client.GetAnimeEpisodes(srv.URL + "/anime/x")
	require.NoError(t, err)
	// parseEpisodes is called internally — verify its output
	assert.Len(t, episodes, 2)
	assert.Equal(t, 5, episodes[0].Num)
	assert.Equal(t, 6, episodes[1].Num)
}

// ---------------------------------------------------------------------------
// AnimefireClient.GetAnimeDetails
// ---------------------------------------------------------------------------

func TestAnimefireClient_GetAnimeDetails_ReturnsError(t *testing.T) {
	t.Parallel()
	client := NewAnimefireClient()
	anime, err := client.GetAnimeDetails("https://animefire.io/anime/naruto")
	assert.Nil(t, anime)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API layer")
}

// ---------------------------------------------------------------------------
// GoyabuClient.sleep
// ---------------------------------------------------------------------------

func TestGoyabuClient_Sleep_ZeroDelay(t *testing.T) {
	t.Parallel()
	client := NewGoyabuClient()
	client.retryDelay = 0
	// Must return without blocking
	client.sleep()
}

func TestGoyabuClient_Sleep_SmallDelay(t *testing.T) {
	t.Parallel()
	client := NewGoyabuClient()
	client.retryDelay = 1 * time.Millisecond
	start := time.Now()
	client.sleep()
	assert.GreaterOrEqual(t, time.Since(start), time.Millisecond)
}

// ---------------------------------------------------------------------------
// MediaManager.GetAnimeStreamURL
// ---------------------------------------------------------------------------

func TestMediaManager_GetAnimeStreamURL_AllAnime(t *testing.T) {
	t.Parallel()
	mock := &MockScraper{
		streamURLFunc: func(url string) (string, map[string]string, error) {
			return "https://cdn.test/video.m3u8", map[string]string{"source": "allanime"}, nil
		},
	}
	mock.scraperType = AllAnimeType

	mm := &MediaManager{
		scraperManager: createTestManager(mock, nil),
	}

	anime := &models.Anime{Name: "Naruto", Source: "AllAnime", URL: "allanime-id"}
	url, _, err := mm.GetAnimeStreamURL(anime, "1", "best", "sub")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.test/video.m3u8", url)
}

func TestMediaManager_GetAnimeStreamURL_AnimeFire(t *testing.T) {
	t.Parallel()
	mock := &MockScraper{
		streamURLFunc: func(url string) (string, map[string]string, error) {
			return "https://cdn.test/af.mp4", nil, nil
		},
	}
	mock.scraperType = AnimefireType

	mm := &MediaManager{
		scraperManager: createTestManager(nil, mock),
	}

	anime := &models.Anime{Name: "Test", Source: "AnimeFire", URL: "https://af.test/anime/1"}
	url, _, err := mm.GetAnimeStreamURL(anime, "1", "best", "sub")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.test/af.mp4", url)
}

func TestMediaManager_GetAnimeStreamURL_UnknownSource(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	anime := &models.Anime{Name: "X", Source: "UnknownSource99"}
	_, _, err := mm.GetAnimeStreamURL(anime, "1", "best", "sub")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unknown source")
}

// ---------------------------------------------------------------------------
// MediaManager.GetScraperManager
// ---------------------------------------------------------------------------

func TestMediaManager_GetScraperManager_ReturnsSame(t *testing.T) {
	t.Parallel()
	mm := NewMediaManager()
	sm := mm.GetScraperManager()
	require.NotNil(t, sm)
	// Calling twice must return the same singleton pointer
	assert.Same(t, sm, mm.GetScraperManager())
}

// ---------------------------------------------------------------------------
// ScraperManager.SearchAnimePTBR
// ---------------------------------------------------------------------------

func TestScraperManager_SearchAnimePTBR_BothReturn(t *testing.T) {
	t.Parallel()
	afMock := &MockScraper{
		searchFunc: func(q string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "Naruto AF", Source: "AnimeFire"}}, nil
		},
	}
	afMock.scraperType = AnimefireType

	manager := NewScraperManagerForTest()
	manager.RegisterScraperForTest(AnimefireType, afMock)
	// GoyabuType and SuperFlixType not registered → errors silently ignored

	results, err := manager.SearchAnimePTBR("Naruto")
	require.NoError(t, err)
	require.Len(t, results, 1)
	// SearchAnimePTBR tags AnimeFire results with [PT-BR] prefix
	assert.Contains(t, results[0].Name, "Naruto AF")
}

func TestScraperManager_SearchAnimePTBR_AllFail(t *testing.T) {
	t.Parallel()
	manager := NewScraperManagerForTest()
	// No PT-BR scrapers registered
	_, err := manager.SearchAnimePTBR("Naruto")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// AllAnimeAdapter.GetAnimeEpisodes
// ---------------------------------------------------------------------------

func TestAllAnimeAdapter_GetAnimeEpisodes_ServerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	t.Cleanup(srv.Close)

	client := newTestClient(srv.URL)
	adapter := &AllAnimeAdapter{client: client}

	_, err := adapter.GetAnimeEpisodes("test-anime-id")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// AllAnimeAdapter.GetStreamURL
// ---------------------------------------------------------------------------

func TestAllAnimeAdapter_GetStreamURL_ServerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)

	client := newTestClient(srv.URL)
	adapter := &AllAnimeAdapter{client: client}

	_, _, err := adapter.GetStreamURL("test-anime-id", "1", "best", "sub")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// AnimefireAdapter.GetAnimeEpisodes
// ---------------------------------------------------------------------------

func TestAnimefireAdapter_GetAnimeEpisodes_Success(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		fmt.Fprint(w, `<html><body>`+
			`<a class="lEp epT divNumEp smallbox px-2 mx-1 text-left d-flex" href="/ep/1">Episódio 1</a>`+
			`</body></html>`)
	}))
	t.Cleanup(srv.Close)

	afClient := NewAnimefireClient()
	afClient.baseURL = srv.URL
	afClient.maxRetries = 0
	afClient.retryDelay = 0

	adapter := &AnimefireAdapter{client: afClient}
	episodes, err := adapter.GetAnimeEpisodes(srv.URL + "/anime/test")
	require.NoError(t, err)
	assert.Len(t, episodes, 1)
}

// ---------------------------------------------------------------------------
// AnimefireAdapter.GetStreamURL
// ---------------------------------------------------------------------------

func TestAnimefireAdapter_GetStreamURL_HTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	t.Cleanup(srv.Close)

	afClient := NewAnimefireClient()
	afClient.baseURL = srv.URL
	afClient.maxRetries = 0
	afClient.retryDelay = 0

	adapter := &AnimefireAdapter{client: afClient}
	_, metadata, err := adapter.GetStreamURL(srv.URL + "/ep/1")
	require.Error(t, err)
	// metadata is always populated by AnimefireAdapter
	assert.Equal(t, "animefire", metadata["source"])
}

// ---------------------------------------------------------------------------
// GoyabuAdapter.GetAnimeEpisodes
// ---------------------------------------------------------------------------

func TestGoyabuAdapter_GetAnimeEpisodes_HTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	t.Cleanup(srv.Close)

	gClient := NewGoyabuClient()
	gClient.baseURL = srv.URL
	gClient.maxRetries = 0
	gClient.retryDelay = 0

	adapter := &GoyabuAdapter{client: gClient}
	_, err := adapter.GetAnimeEpisodes(srv.URL + "/anime/test")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// GoyabuAdapter.GetStreamURL
// ---------------------------------------------------------------------------

func TestGoyabuAdapter_GetStreamURL_HTTPError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	t.Cleanup(srv.Close)

	gClient := NewGoyabuClient()
	gClient.baseURL = srv.URL
	gClient.maxRetries = 0
	gClient.retryDelay = 0

	adapter := &GoyabuAdapter{client: gClient}
	_, metadata, err := adapter.GetStreamURL(srv.URL + "/ep/1")
	require.Error(t, err)
	assert.Equal(t, "goyabu", metadata["source"])
}
