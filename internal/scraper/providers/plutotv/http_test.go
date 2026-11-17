package plutotv

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGetRefusesAURLItCannotBuildARequestFor(t *testing.T) {
	_, status, err := NewClient().get(t.Context(), "http://[::1", "", "application/json", "layer", 10)
	require.ErrorContains(t, err, "bad request URL")
	assert.Zero(t, status)
}

func TestGetReportsABodyThatEndsEarly(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Length", strconv.Itoa(100))
		_, _ = w.Write([]byte("only ten b")) // then the connection closes short of the 100
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL)
	_, status, err := c.get(t.Context(), srv.URL, "", "application/json", "layer", 1000)
	require.ErrorContains(t, err, "failed to read response")
	assert.Equal(t, http.StatusOK, status)
}

func TestGetNeverCarriesATokenItWasNotGiven(t *testing.T) {
	f := newFakePluto(t)
	f.set("/x", "ok")
	c := f.client()
	_, _, err := c.get(t.Context(), f.url()+"/x", "", "text/plain", "layer", 10)
	require.NoError(t, err)
	assert.Empty(t, f.lastRequest("/x").Header.Get("Authorization"), "the media host is never sent the session's token")
	_, _, err = c.get(t.Context(), f.url()+"/x", "secret", "text/plain", "layer", 10)
	require.NoError(t, err)
	assert.Equal(t, "Bearer secret", f.lastRequest("/x").Header.Get("Authorization"))
}

func TestPlaylistsAreFetchedWithoutTheSessionToken(t *testing.T) {
	f := newFakePluto(t)
	f.serveEpisode(fakeEpisodeID)
	_, err := f.client().Stream(t.Context(), episodeURL, "best")
	require.NoError(t, err)
	for _, path := range []string{f.variantPath(fakeEpisodeID, 3321280), f.contentPath("hls_2400")} {
		r := f.lastRequest(path)
		require.NotNil(t, r, path)
		assert.Empty(t, r.Header.Get("Authorization"), "%s: a playlist host is not given the Bearer", path)
	}
}
