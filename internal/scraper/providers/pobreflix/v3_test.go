package pobreflix

import (
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The v3 fixtures are the real pages of pobreflixtvfree.cv, captured
// 2026-11-14, the day the site moved off pobreflixtv.* and dropped the
// embedded JSON.
const (
	v3MovieURL   = "https://pobreflixtvfree.cv/filmes/online/interestelar-dublado-157336"
	v3SeriesURL  = "https://pobreflixtvfree.cv/series/online/dexter-dublado-1405"
	v3EpisodeURL = "https://pobreflixtvfree.cv/episodios/online/dexter-1x2-dublado-1405"
)

func TestParseV3Title_Movie(t *testing.T) {
	t.Parallel()
	title, ok := parseTitle(fixtureDoc(t, "v3_movie.html"), v3MovieURL)
	require.True(t, ok, "the page mounts a player, so it is a title")
	assert.Equal(t, KindMovie, title.Kind)
	assert.Equal(t, 157336, title.TMDBID, "the id that ends the slug is the TMDB id")
	assert.Equal(t, "Interestelar", title.Name)
	assert.Equal(t, AudioDubbed, title.Audio)
	require.Len(t, title.Servers, 1)
	assert.Equal(t, "https://redeflixapi.click/filme/157336", title.Servers[0].URL)
	assert.True(t, title.Servers[0].Gated(), "a player page, not a file")
	assert.False(t, title.HasDirect())
	assert.True(t, NewClientForTest(http.DefaultClient, "https://pobreflixtvfree.cv", filepath.Join(t.TempDir(), "c.json")).
		HasSolverFreePlayer(title.Servers), "RedeFlixApi needs no solver")
}

func TestParseV3Title_Series(t *testing.T) {
	t.Parallel()
	title, ok := parseTitle(fixtureDoc(t, "v3_series.html"), v3SeriesURL)
	require.True(t, ok)
	assert.Equal(t, KindSeries, title.Kind)
	assert.Equal(t, 1405, title.TMDBID)
	assert.Equal(t, "Dexter", title.Name)
	assert.Empty(t, title.Servers, "a series lists its players per episode")
	assert.True(t, title.HasDirect())
	require.Len(t, title.Seasons, 8)
	s1, ok := title.Season(1)
	require.True(t, ok)
	require.Len(t, s1.Episodes, 12)
	assert.Equal(t, 1, s1.Episodes[0].Number)
	assert.Equal(t, "Piloto", s1.Episodes[0].Title)
	assert.Equal(t, "https://pobreflixtvfree.cv/episodios/online/dexter-1x1-dublado-1405", s1.Episodes[0].URL)
}

func TestParseV3Episode(t *testing.T) {
	t.Parallel()
	data, ok := parseV3Episode(fixtureDoc(t, "v3_episode.html"), v3EpisodeURL)
	require.True(t, ok)
	assert.Equal(t, 1, data.Season)
	assert.Equal(t, 2, data.Number)
	assert.Equal(t, 1405, data.Series.TMDBID)
	require.Len(t, data.Servers, 1)
	assert.Equal(t, "https://redeflixapi.click/serie/1405/1/2", data.Servers[0].URL)
}

// A page that is the home page (what the old host answers to every path) is
// no title.
func TestParseV3Title_HomePageIsNoTitle(t *testing.T) {
	t.Parallel()
	_, ok := parseV3Title(goqueryT(t, `<html><body><div class="hero-slide"></div></body></html>`), v3MovieURL)
	assert.False(t, ok)
	_, ok = parseV3Episode(goqueryT(t, `<html><body></body></html>`), v3EpisodeURL)
	assert.False(t, ok)
}

func TestIsEpisodePage(t *testing.T) {
	t.Parallel()
	assert.True(t, IsEpisodePage(v3EpisodeURL))
	assert.True(t, IsEpisodePage("https://x.test/assistir/dexter-t1-ep2.html"))
	assert.False(t, IsEpisodePage(v3SeriesURL))
	assert.False(t, IsEpisodePage(v3MovieURL))
}

// TestClientV3Flow walks search → title → episode against a server that
// answers like the new site.
func TestClientV3Flow(t *testing.T) {
	t.Parallel()
	serve := func(w http.ResponseWriter, name string) {
		raw, err := os.ReadFile(filepath.Join("testdata", name))
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}
		_, _ = w.Write(raw)
	}
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/api/busca":
			assert.Equal(t, "dexter", r.URL.Query().Get("q"))
			_, _ = w.Write([]byte(`[{"title":"Dexter","year":2006,"type":"series","url":"/series/online/dexter-dublado-1405","poster":"x"},{"title":"","type":"movie","url":"/filmes/online/vazio-dublado-1"},{"title":"Lixo","type":"movie","url":"/sobre"}]`))
		case "/series/online/dexter-dublado-1405":
			serve(w, "v3_series.html")
		case "/episodios/online/dexter-1x2-dublado-1405":
			serve(w, "v3_episode.html")
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	c := NewClientForTest(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "corpus.json"))
	c.apiEnabled = true

	results, err := c.Search(t.Context(), "dexter")
	require.NoError(t, err)
	require.Len(t, results, 1, "an unnamed record and a path that is no record are dropped")
	assert.Equal(t, srv.URL+"/series/online/dexter-dublado-1405", results[0].URL)
	assert.Equal(t, KindSeries, results[0].Kind)
	assert.Equal(t, AudioDubbed, results[0].Audio)

	title, err := c.Title(t.Context(), results[0].URL)
	require.NoError(t, err)
	assert.Equal(t, 1405, title.TMDBID)

	pageURL := c.EpisodeURL(title, 1, 2)
	assert.Equal(t, srv.URL+"/episodios/online/dexter-1x2-dublado-1405", pageURL,
		"the episode page comes from the title's own list, not from the old /assistir/ path")
	data, err := c.Episode(t.Context(), pageURL)
	require.NoError(t, err)
	require.Len(t, data.Servers, 1)
	assert.Equal(t, "https://redeflixapi.click/serie/1405/1/2", data.Servers[0].URL)
}
