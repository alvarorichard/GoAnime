package api

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pobreflix"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Current Pobreflix pages embed their MP4 servers. These tests drive the api
// flow against a fake current site and prove playback never touches the
// StartFlix panel.

const directMoviePage = `<!doctype html><html><body>
<script id="initialMovieData" type="application/json">{
 "type":"movie",
 "item":{"id":1283515,"internal_id":74602,"tmdb_id":1283515,"slug":"verity-dublado",
         "type":"movie","name":"Verity","poster":"https://img.test/p.jpg","audio":"DUB"},
 "servers":{"dub":[{"id":"a","name":"Servidor 1 (1080p)","quality":"1080p",
                    "url":"https://cdn.test/verity.mp4","download":"https://cdn.test/verity.mp4",
                    "type":"mp4","audio":"dub"}],"leg":[]}
}</script></body></html>`

const directSeriesPage = `<!doctype html><html><body>
<script id="initialSeriesData" type="application/json">{
 "type":"series",
 "item":{"id":288458,"internal_id":74494,"tmdb_id":288458,"slug":"ben-hur-a-serie-dublado",
         "type":"series","name":"Ben-Hur: A Série","poster":"https://img.test/p.jpg","audio":"DUB"},
 "episodes":[
   {"id":1,"tmdb_id":288458,"season":1,"episode":1,"title":"O Judeu e o Romano","available":true,"audio":"dub"},
   {"id":2,"tmdb_id":288458,"season":1,"episode":2,"title":"Uma Esposa Para Judá","available":true,"audio":"dub"}
 ],
 "seasons":[{"season":1,"name":"Temporada 1"}]
}</script></body></html>`

const directEpisodePage = `<!doctype html><html><body>
<script id="initialEpisodeData" type="application/json">{
 "type":"series",
 "series":{"id":288458,"internal_id":74494,"tmdb_id":288458,"slug":"ben-hur-a-serie-dublado",
           "type":"series","name":"Ben-Hur: A Série","poster":"https://img.test/p.jpg","audio":"DUB"},
 "episode":{"id":1252009,"tmdb_id":288458,"season":1,"episode":1,
            "title":"O Judeu e o Romano","available":true,"audio":"dub"},
 "servers":{"dub":[{"id":"srv","name":"Servidor 1 (1080p)","quality":"1080p","type":"mp4","audio":"dub",
                    "download":"/api?action=episodeDownload&episode_id=1252009&audio=dub&server=srv"}],"leg":[]}
}</script></body></html>`

// pbDirectClient points the api's Pobreflix seam at a test server.
func pbDirectClient(t *testing.T, srv *httptest.Server) {
	t.Helper()
	prev := pbClientFn
	pbClientFn = func() *pobreflix.Client {
		return pobreflix.NewClientForTest(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "cache.json"))
	}
	t.Cleanup(func() { pbClientFn = prev })
}

// TestPobreflixDirectMovie: a current movie lists itself and plays its public
// MP4 without any panel request.
func TestPobreflixDirectMovie(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/filme/verity-dublado.html" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, directMoviePage)
	}))
	t.Cleanup(srv.Close)
	pbDirectClient(t, srv)

	anime := &models.Anime{
		Name:      "Verity",
		URL:       srv.URL + "/filme/verity-dublado.html",
		Source:    pobreflix.SourceName,
		MediaType: models.MediaTypeMovie,
	}
	eps, err := GetPobreflixEpisodes(anime)
	require.NoError(t, err)
	require.Len(t, eps, 1)
	assert.Equal(t, anime.URL, eps[0].URL, "the movie lists its own page")
	assert.Equal(t, models.MediaTypeMovie, anime.MediaType)
	assert.Equal(t, 1283515, anime.TMDBID)

	url, err := GetPobreflixStreamURL(anime, &eps[0], "best")
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.test/verity.mp4", url)
}

// TestPobreflixDirectSeries: a current series lists the embedded episodes and
// resolves an episode's download endpoint to the final MP4.
func TestPobreflixDirectSeries(t *testing.T) {
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/serie/ben-hur-a-serie-dublado.html":
			fmt.Fprint(w, directSeriesPage)
		case "/assistir/ben-hur-a-serie-dublado-t1-ep1.html":
			fmt.Fprint(w, directEpisodePage)
		case "/api":
			http.Redirect(w, r, srv.URL+"/media/ep1.mp4", http.StatusFound)
		case "/media/ep1.mp4":
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusOK)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	pbDirectClient(t, srv)

	anime := &models.Anime{
		Name:      "Ben-Hur: A Série",
		URL:       srv.URL + "/serie/ben-hur-a-serie-dublado.html",
		Source:    pobreflix.SourceName,
		MediaType: models.MediaTypeTV,
	}
	eps, err := GetPobreflixSeasonEpisodes(anime, 1)
	require.NoError(t, err)
	require.Len(t, eps, 2)
	assert.Equal(t, 1, anime.CurrentSeason)
	assert.Equal(t, srv.URL+"/assistir/ben-hur-a-serie-dublado-t1-ep1.html", eps[0].URL)
	assert.Equal(t, "O Judeu e o Romano", eps[0].Title.English)

	url, err := GetPobreflixStreamURL(anime, &eps[0], "best")
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/media/ep1.mp4", url)
}
