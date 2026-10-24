package pobreflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// The current site embeds the authoritative record as JSON. These tests pin
// the three page shapes and the URL grammar the embedded blocks use.

const embeddedMovieHTML = `<!doctype html><html><body>
<script id="initialMovieData" type="application/json">{
  "type":"movie",
  "item":{"id":1283515,"internal_id":74602,"tmdb_id":1283515,"slug":"verity-dublado",
          "type":"movie","name":"Verity","poster":"https://img.test/w185.jpg",
          "poster_large":"https://img.test/w342.jpg","audio":"DUB"},
  "servers":{"dub":[{"id":"a","name":"Servidor 1 (1080p)","quality":"1080p",
                     "url":"https://cdn.test/verity.mp4","download":"https://cdn.test/verity.mp4",
                     "type":"mp4","audio":"dub"}],"leg":[]}
}</script>
<div id="view" data-video-id="79524" data-tmdb-id="1477712"></div>
</body></html>`

const embeddedSeriesHTML = `<!doctype html><html><body>
<script id="initialSeriesData" type="application/json">{
  "type":"series",
  "item":{"id":288458,"internal_id":74494,"tmdb_id":288458,"slug":"ben-hur-a-serie-dublado",
          "type":"series","name":"Ben-Hur: A Série","poster":"https://img.test/p.jpg",
          "poster_large":"https://img.test/p342.jpg","audio":"DUB"},
  "episodes":[
    {"id":1,"tmdb_id":288458,"season":1,"episode":2,"title":"Uma Esposa Para Judá","available":true,"audio":"dub"},
    {"id":2,"tmdb_id":288458,"season":1,"episode":1,"title":"O Judeu e o Romano","available":true,"audio":"dub"},
    {"id":3,"tmdb_id":288458,"season":2,"episode":1,"title":"Outra Temporada","available":true,"audio":"leg"}
  ],
  "seasons":[{"season":1,"name":"Temporada 1"},{"season":2,"name":"Temporada 2"}]
}</script>
</body></html>`

const embeddedEpisodeHTML = `<!doctype html><html><body>
<script id="initialEpisodeData" type="application/json">{
  "type":"series",
  "series":{"id":288458,"internal_id":74494,"tmdb_id":288458,"slug":"ben-hur-a-serie-dublado",
            "type":"series","name":"Ben-Hur: A Série","poster":"https://img.test/p.jpg","audio":"DUB"},
  "episode":{"id":1252009,"tmdb_id":288458,"season":1,"episode":1,
             "title":"O Judeu e o Romano","available":true,"audio":"dub"},
  "servers":{"dub":[{"id":"srv","name":"Servidor 1 (1080p)","quality":"1080p","type":"mp4","audio":"dub",
                     "download":"/api?action=episodeDownload&episode_id=1252009&audio=dub&server=srv"}],"leg":[]}
}</script>
</body></html>`

// TestParseTitle_EmbeddedMovie: the embedded block wins over the fixed #view
// shell, whose data-tmdb-id belongs to the player, not the title.
func TestParseTitle_EmbeddedMovie(t *testing.T) {
	t.Parallel()
	doc := goqueryT(t, embeddedMovieHTML)
	title, ok := parseTitle(doc, "https://pobreflixtv.gay/filme/verity-dublado.html")
	require.True(t, ok)
	assert.Equal(t, KindMovie, title.Kind)
	assert.Equal(t, 1283515, title.TMDBID, "embedded tmdb_id, not the #view shell")
	assert.Equal(t, 74602, title.InternalID)
	assert.Equal(t, 74602, title.VideoID)
	assert.Equal(t, "Verity", title.Name)
	assert.Equal(t, "verity-dublado", title.Slug)
	assert.Equal(t, AudioDubbed, title.Audio)
	assert.True(t, title.HasDirect())
	require.Len(t, title.Servers, 1)
	assert.Equal(t, "https://cdn.test/verity.mp4", title.Servers[0].URL)
	assert.Equal(t, 1080, QualityRank(title.Servers[0].Quality))
}

// TestParseTitle_EmbeddedSeries: the embedded episode list is sorted and
// carries the site's current /assistir/ URLs and per-episode audio.
func TestParseTitle_EmbeddedSeries(t *testing.T) {
	t.Parallel()
	doc := goqueryT(t, embeddedSeriesHTML)
	title, ok := parseTitle(doc, "https://pobreflixtv.gay/serie/ben-hur-a-serie-dublado.html")
	require.True(t, ok)
	assert.Equal(t, KindSeries, title.Kind)
	assert.Equal(t, 288458, title.TMDBID)
	assert.Equal(t, "Ben-Hur: A Série", title.Name)
	require.Len(t, title.Seasons, 2)

	first := title.Seasons[0]
	assert.Equal(t, 1, first.Number)
	assert.Equal(t, AudioDubbed, first.Audio)
	require.Len(t, first.Episodes, 2)
	assert.Equal(t, 1, first.Episodes[0].Number)
	assert.Equal(t, "O Judeu e o Romano", first.Episodes[0].Title)
	assert.Equal(t, "https://pobreflixtv.gay/assistir/ben-hur-a-serie-dublado-t1-ep1.html", first.Episodes[0].URL)
	assert.Equal(t, "O Judeu e o Romano", title.EpisodeTitle(1, 1))

	second := title.Seasons[1]
	assert.Equal(t, 2, second.Number)
	assert.Equal(t, AudioSubtitled, second.Audio)
	assert.Equal(t, "https://pobreflixtv.gay/assistir/ben-hur-a-serie-dublado-t2-ep1.html", second.Episodes[0].URL)
}

// TestEpisode_EmbeddedPage: Episode reads the slot and resolves the relative
// download endpoint against the catalog host.
func TestEpisode_EmbeddedPage(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/assistir/ben-hur-a-serie-dublado-t1-ep1.html" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, embeddedEpisodeHTML)
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL, t.TempDir()+"/cache.json")

	data, err := c.Episode(context.Background(), srv.URL+"/assistir/ben-hur-a-serie-dublado-t1-ep1.html")
	require.NoError(t, err)
	assert.Equal(t, 288458, data.Series.TMDBID)
	assert.Equal(t, 1, data.Season)
	assert.Equal(t, 1, data.Number)
	assert.Equal(t, "O Judeu e o Romano", data.Name)
	require.Len(t, data.Servers, 1)
	assert.Equal(t, srv.URL+"/api?action=episodeDownload&episode_id=1252009&audio=dub&server=srv", data.Servers[0].Download)
}

// TestParseRecordPath_CurrentLayouts: the sitemap's current URLs map onto the
// same corpus entry shape as the old ones. The random suffix and appended
// TMDB id are display noise; a title's own year is not.
func TestParseRecordPath_CurrentLayouts(t *testing.T) {
	t.Parallel()
	tests := []struct {
		path   string
		kind   MediaKind
		slug   string
		audio  Audio
		wantOK bool
	}{
		{path: "/filme/verity-dublado.html", kind: KindMovie, slug: "verity", audio: AudioDubbed, wantOK: true},
		{path: "/filme/todo-poderoso-53qvb.html", kind: KindMovie, slug: "todo-poderoso", audio: AudioUnknown, wantOK: true},
		{path: "/filme/ben-hur-40360.html", kind: KindMovie, slug: "ben-hur", audio: AudioUnknown, wantOK: true},
		{path: "/filme/blade-runner-2049.html", kind: KindMovie, slug: "blade-runner-2049", audio: AudioUnknown, wantOK: true},
		{path: "/filme/007-cassino-royale-4hsp3.html", kind: KindMovie, slug: "007-cassino-royale", audio: AudioUnknown, wantOK: true},
		{path: "/serie/ben-hur-a-serie-dublado.html", kind: KindSeries, slug: "ben-hur-a-serie", audio: AudioDubbed, wantOK: true},
		{path: "/assistir/ben-hur-a-serie-dublado-t1-ep1.html", wantOK: false},
		{path: "/filme/", wantOK: false},
	}
	for _, tt := range tests {
		kind, slug, audio, _, ok := parseRecordPath(tt.path)
		if !tt.wantOK {
			assert.False(t, ok, "%q must not be a corpus record", tt.path)
			continue
		}
		require.True(t, ok, "%q must parse", tt.path)
		assert.Equal(t, tt.kind, kind, "%q kind", tt.path)
		assert.Equal(t, tt.slug, slug, "%q slug", tt.path)
		assert.Equal(t, tt.audio, audio, "%q audio", tt.path)
	}
}

// TestSitemapShards_CurrentIndex: when sitemap.php is gone, the current
// sitemap.xml index supplies the movie and series files in order.
func TestSitemapShards_CurrentIndex(t *testing.T) {
	t.Parallel()
	var index string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/sitemap.php":
			http.NotFound(w, r)
		case "/sitemap.xml":
			fmt.Fprint(w, index)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(srv.Close)
	index = `<?xml version="1.0"?><sitemapindex xmlns="http://www.sitemaps.org/schemas/sitemap/0.9">
	<sitemap><loc>` + srv.URL + `/sitemap-series.xml</loc></sitemap>
	<sitemap><loc>` + srv.URL + `/sitemap-movies.xml</loc></sitemap>
	<sitemap><loc>` + srv.URL + `/sitemap-episodes.xml</loc></sitemap>
	</sitemapindex>`

	c := NewClientForTest(srv.Client(), srv.URL, t.TempDir()+"/cache.json")
	shards, err := c.sitemapShards(context.Background())
	require.NoError(t, err)
	require.Len(t, shards, 2, "episodes and pages are not title records")
	assert.Equal(t, "movies", shards[0].name)
	assert.Equal(t, srv.URL+"/sitemap-movies.xml", shards[0].url)
	assert.Equal(t, "series", shards[1].name)
}

// TestSearch_APIUsesRealNames: production asks /api first; the answer's name
// and slug are authoritative and no sitemap fetch is needed.
func TestSearch_APIUsesRealNames(t *testing.T) {
	t.Parallel()
	var requests int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		requests++
		if r.URL.Path != "/api" {
			http.NotFound(w, r)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Query().Get("type") {
		case "movie":
			fmt.Fprint(w, `{"type":"movie","items":[{"id":310,"internal_id":7079,"tmdb_id":310,"slug":"todo-poderoso-53qvb","type":"movie","name":"Todo Poderoso","audio":"DUB"}]}`)
		default:
			fmt.Fprint(w, `{"type":"series","items":[]}`)
		}
	}))
	t.Cleanup(srv.Close)

	c := NewClientForTest(srv.Client(), srv.URL, t.TempDir()+"/cache.json")
	c.apiEnabled = true
	results, err := c.Search(context.Background(), "todo poderoso")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Todo Poderoso", results[0].Title)
	assert.Equal(t, srv.URL+"/filme/todo-poderoso-53qvb.html", results[0].URL)
	assert.Equal(t, KindMovie, results[0].Kind)
	assert.Equal(t, 2, requests, "one catalog request per type")
}
