package pobreflix

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// 2026-10-27: every Pobreflix page began listing one "Player Principal" per
// audio, whose URL is a streambetter.lol page. That page answers with HTML
// titled "Verificando..." and a Cloudflare Turnstile; no media is behind it
// until a browser passes the check. GoAnime counted any listed URL as a file,
// returned the challenge page as the stream, and never tried its fallbacks, so
// the dub that played days earlier (a public MP4) stopped playing for every
// title. The fixtures below are the live pages' JSON, copied verbatim.

const gatedMovieHTML = `<!doctype html><html><body>
<script id="initialMovieData" type="application/json">{"type":"movie","item":{"id":310,"tmdb_id":310,"name":"Todo Poderoso","original_name":"Todo Poderoso","slug":"todo-poderoso","poster":"https://image.tmdb.org/t/p/w500/wcA8udbEqqU5tPzsyqyZCDX6OwI.jpg","year":2003,"genre":"Fantasia, Comédia","rating":6.727},"servers":{"dub":[{"name":"Player Principal HD","quality":"HD","url":"https://streambetter.lol/filme/310"}],"leg":[{"name":"Player Principal Legendado","quality":"HD","url":"https://streambetter.lol/filme/310"}]}}</script>
</body></html>`

// The episode id is a string here ("<tmdb>-<season>-<episode>"); it was a
// number, and the typed field made json.Unmarshal fail, so the page read as
// "no embedded episode data".
const gatedEpisodeHTML = `<!doctype html><html><body>
<script id="initialEpisodeData" type="application/json">{"series":{"id":66732,"tmdb_id":66732,"name":"Stranger Things","slug":"stranger-things","poster":"https://image.tmdb.org/t/p/w500/twfKp60THrcOIep9sjHODOOfO8d.jpg"},"episode":{"id":"66732-1-1","tmdb_id":66732,"season":1,"episode":1,"title":"Episódio 1"},"prevEpisode":null,"nextEpisode":{"season":1,"episode":2},"servers":{"dub":[{"name":"Player Principal HD","quality":"HD","url":"https://streambetter.lol/serie/66732/1/1"}],"leg":[{"name":"Player Principal Legendado","quality":"HD","url":"https://streambetter.lol/serie/66732/1/1"}]}}</script>
</body></html>`

func TestServerDirectVersusGated(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name       string
		server     Server
		wantDirect bool
		wantGated  bool
	}{
		{"public mp4", Server{URL: "https://cdn.test/verity.mp4"}, true, false},
		{"mp4 with a signed query", Server{URL: "https://r2.test/a/b.mp4?X-Amz-Signature=abc"}, true, false},
		{"hls master", Server{URL: "https://cdn.test/pl/master.m3u8"}, true, false},
		{"upper-case extension", Server{URL: "https://cdn.test/Verity.MP4"}, true, false},
		{"the site's download endpoint", Server{Download: "/api?action=episodeDownload&episode_id=1&audio=dub&server=s"}, true, false},
		{"file in download, page in url", Server{URL: "https://player.test/e/1", Download: "https://cdn.test/v.mkv"}, true, false},
		{"the live streambetter movie page", Server{URL: "https://streambetter.lol/filme/310"}, false, true},
		{"the live streambetter episode page", Server{URL: "https://streambetter.lol/serie/66732/1/1"}, false, true},
		{"a path that merely contains mp4", Server{URL: "https://player.test/mp4/embed"}, false, true},
		{"nothing listed", Server{}, false, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.wantDirect, tt.server.Direct(), "Direct")
			assert.Equal(t, tt.wantGated, tt.server.Gated(), "Gated")
		})
	}
}

func TestDirectServersAndGatedHosts(t *testing.T) {
	t.Parallel()
	servers := []Server{
		{URL: "https://streambetter.lol/filme/310", Audio: AudioDubbed},
		{URL: "https://streambetter.lol/filme/310", Audio: AudioSubtitled},
		{URL: "https://cdn.test/a.mp4", Audio: AudioDubbed},
		{URL: "https://other.test/embed/9", Audio: AudioDubbed},
	}
	assert.Equal(t, []Server{servers[2]}, DirectServers(servers))
	assert.Equal(t, []string{"streambetter.lol", "other.test"}, GatedHosts(servers), "distinct hosts, in listing order")
	assert.Empty(t, GatedHosts(DirectServers(servers)))
}

// TestParseTitle_GatedMovieIsNotDirect: the live movie page parses, keeps the
// servers it lists, and is not "direct": there is no file to play.
func TestParseTitle_GatedMovieIsNotDirect(t *testing.T) {
	t.Parallel()
	title, ok := parseTitle(goqueryT(t, gatedMovieHTML), "https://pobreflixtv.gay/filme/todo-poderoso.html")
	require.True(t, ok)
	assert.Equal(t, KindMovie, title.Kind)
	assert.Equal(t, 310, title.TMDBID, "the TMDB id still keys the panel and the vidsrc fallback")
	require.Len(t, title.Servers, 1, "dub and leg name one player page, kept once")
	assert.False(t, title.HasDirect(), "a player page is not a file")
	assert.Equal(t, []string{"streambetter.lol"}, GatedHosts(title.Servers))
}

// TestEpisode_StringIDDecodes: the live episode page decodes, so the series
// can tell its servers are gated instead of reporting a missing page.
func TestEpisode_StringIDDecodes(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path != "/assistir/stranger-things-t1-ep1.html" {
			http.NotFound(w, r)
			return
		}
		fmt.Fprint(w, gatedEpisodeHTML)
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL, filepath.Join(t.TempDir(), "cache.json"))

	ep, err := c.Episode(context.Background(), srv.URL+"/assistir/stranger-things-t1-ep1.html")

	require.NoError(t, err, "a string episode id must not break the page")
	assert.Equal(t, 66732, ep.Series.TMDBID)
	assert.Equal(t, 1, ep.Season)
	assert.Equal(t, 1, ep.Number)
	require.Len(t, ep.Servers, 1, "dub and leg name one player page, kept once")
	assert.Empty(t, DirectServers(ep.Servers))
	assert.Equal(t, []string{"streambetter.lol"}, GatedHosts(ep.Servers))
}
