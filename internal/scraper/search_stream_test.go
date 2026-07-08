package scraper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/allanime"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/animefire"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/goyabu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	client := allanime.NewClientForTest(srv.URL)
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

	client := allanime.NewClientForTest(srv.URL)
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

	afClient := animefire.NewClientForTest(srv.URL)

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

	afClient := animefire.NewClientForTest(srv.URL)

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

	gClient := goyabu.NewClientForTest(srv.URL)

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

	gClient := goyabu.NewClientForTest(srv.URL)

	adapter := &GoyabuAdapter{client: gClient}
	_, metadata, err := adapter.GetStreamURL(srv.URL + "/ep/1")
	require.Error(t, err)
	assert.Equal(t, "goyabu", metadata["source"])
}
