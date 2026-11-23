package movie

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// IMDb's suggestion endpoint answers a name with titles and people; the name
// search must return the titles, keep their years and kinds, and ask the
// letter's shard (checked 2026-11-23 for "icarly").
const imdbICarlyJSON = `{"d":[
 {"id":"tt0972534","l":"iCarly","y":2007,"qid":"tvSeries","q":"TV series"},
 {"id":"tt13657062","l":"iCarly","y":2021,"qid":"tvSeries","q":"TV series"},
 {"id":"tt1286124","l":"iCarly: iGo to Japan","y":2008,"qid":"tvMovie"},
 {"id":"nm1234567","l":"Miranda Cosgrove"},
 {"id":"tt38627912","l":"Untitled iCarly movie","qid":"movie"},
 {"id":"tt0000009","l":""}
],"q":"icarly","v":1}`

func newIMDbTestServer(t *testing.T, status int, body string, gotPath *string) *IMDbClient {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		*gotPath = r.URL.EscapedPath()
		w.WriteHeader(status)
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return NewIMDbClientForTest(srv.Client(), srv.URL)
}

func TestIMDbSearchTitles(t *testing.T) {
	var path string
	c := newIMDbTestServer(t, http.StatusOK, imdbICarlyJSON, &path)
	titles, err := c.SearchTitles("  iCarly ")
	require.NoError(t, err)
	assert.Equal(t, "/i/icarly.json", path, "the first letter picks the shard and the query is lower case")
	require.Len(t, titles, 4, "a person and a nameless entry are not titles")
	assert.Equal(t, IMDbTitle{ID: "tt0972534", Title: "iCarly", Year: 2007, Kind: "tvSeries"}, titles[0])
	assert.Equal(t, 2021, titles[1].Year)
	assert.Equal(t, "tvMovie", titles[2].Kind)
	assert.Zero(t, titles[3].Year, "an unreleased title has no year")
}

func TestIMDbSearchTitlesShards(t *testing.T) {
	for q, want := range map[string]string{
		"24":         "/2/24.json",
		"Dexter":     "/d/dexter.json",
		"é um filme": "/x/%C3%A9%20um%20filme.json",
		"¿Qué?":      "/x/%C2%BFqu%C3%A9%3F.json",
	} {
		var path string
		c := newIMDbTestServer(t, http.StatusOK, `{"d":[]}`, &path)
		_, err := c.SearchTitles(q)
		require.NoError(t, err)
		assert.Equal(t, want, path, q)
	}
}

func TestIMDbSearchTitlesBlankAsksNothing(t *testing.T) {
	var path string
	c := newIMDbTestServer(t, http.StatusOK, imdbICarlyJSON, &path)
	titles, err := c.SearchTitles("   ")
	require.NoError(t, err)
	assert.Nil(t, titles)
	assert.Empty(t, path, "no request for a blank name")
}

func TestIMDbSearchTitlesErrors(t *testing.T) {
	var path string
	c := newIMDbTestServer(t, http.StatusTooManyRequests, `{}`, &path)
	_, err := c.SearchTitles("icarly")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "429")

	c = newIMDbTestServer(t, http.StatusOK, `<html>blocked</html>`, &path)
	_, err = c.SearchTitles("icarly")
	require.Error(t, err, "an answer that is not JSON is an error, not an empty list")
}

// TitleByID shares the fetch with SearchTitles; its contract must not move.
func TestIMDbTitleByIDStillWorks(t *testing.T) {
	var path string
	c := newIMDbTestServer(t, http.StatusOK, imdbICarlyJSON, &path)
	title, err := c.TitleByID("tt13657062")
	require.NoError(t, err)
	assert.Equal(t, "/t/tt13657062.json", path)
	assert.Equal(t, 2021, title.Year)
	_, err = c.TitleByID("tt0000404")
	require.ErrorIs(t, err, ErrIMDbTitleNotFound)
	_, err = c.TitleByID("nope")
	require.Error(t, err)
}
