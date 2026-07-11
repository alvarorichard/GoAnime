package scraper

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/providers/allanime"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/animefire"
	"github.com/alvarorichard/Goanime/internal/scraper/providers/goyabu"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
