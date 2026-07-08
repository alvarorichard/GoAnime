package animefire

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

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
	anime, err := client.GetAnimeDetails("https://io/anime/naruto")
	assert.Nil(t, anime)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "API layer")
}
