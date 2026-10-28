package scraper

import (
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Live shapes captured 2026-10-16 against hianime.at: search cards,
// episode-list and server-list JSON envelopes, a ZokoAnime embed page whose
// window.__P__ blob is XOR'd with the documented build tag, and a master
// playlist. The adapter had no dedicated tests — every other source does —
// so this file drives search → episodes → stream through the real adapter
// glue against one stub.

// obfuscateHiAnime reproduces the player's encoding (JSON, XOR with the build
// tag, base64), mirroring the client package's own test helper.
func obfuscateHiAnime(payload string) string {
	key := []byte("otaku-embed-v1")
	raw := []byte(payload)
	for i := range raw {
		raw[i] ^= key[i%len(key)]
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func hiAnimeStub(t *testing.T) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/search":
			assert.Equal(t, "cowboy bebop", r.URL.Query().Get("keyword"))
			_, _ = fmt.Fprint(w, `<html><body>
				<div class="film_list-wrap"><div class="flw-item">
					<div class="film-poster"><img src="`+srv.URL+`/p.jpg" class="film-poster-img" alt="Cowboy Bebop"></div>
					<div class="film-detail"><h3 class="film-name">
						<a href="/cowboy-bebop-26" title="Cowboy Bebop">Cowboy Bebop</a>
					</h3></div>
				</div></div>
			</body></html>`)
		case "/api/theme/episode/list/26":
			_, _ = fmt.Fprint(w, `{"status":true,"totalItems":2,"html":"`+
				`<a title=\"Ballad\" class=\"ssl-item ep-item\" data-number=\"1\" data-id=\"9001\">`+
				`<div class=\"ep-name\" title=\"Ballad\">Ballad</div></a>`+
				`<a title=\"Stray\" class=\"ssl-item ep-item\" data-number=\"2\" data-id=\"9002\">`+
				`<div class=\"ep-name\" title=\"Stray\">Stray</div></a>`+
				`"}`)
		case "/api/theme/episode/servers":
			embed := base64.StdEncoding.EncodeToString([]byte(srv.URL + "/embed/1"))
			dub := base64.StdEncoding.EncodeToString([]byte(srv.URL + "/embed/2"))
			_, _ = fmt.Fprintf(w, `{"status":true,"html":"`+
				`<div class=\"item server-item\" data-type=\"sub\" data-server-name=\"HD-1\" data-hash=\"bWVnYXBsYXk=\"></div>`+
				`<div class=\"item server-item\" data-type=\"sub\" data-server-name=\"ZokoAnime\" data-hash=\"%s\"></div>`+
				`<div class=\"item server-item\" data-type=\"dub\" data-server-name=\"ZokoAnime\" data-hash=\"%s\"></div>`+
				`"}`, embed, dub)
		case "/embed/1", "/embed/2":
			config := `{"src":"` + srv.URL + `/master.m3u8","subtitles":[` +
				`{"lang":"en","label":"English","src":"` + srv.URL + `/en.vtt","default":true}],` +
				`"skip":{"intro":{"start":0,"end":107},"outro":null}}`
			_, _ = fmt.Fprint(w, `<html><body><div id="player"></div>`+
				`<script>window.__P="`+obfuscateHiAnime(config)+`"</script></body></html>`)
		case "/master.m3u8":
			_, _ = fmt.Fprint(w, "#EXTM3U\n"+
				"#EXT-X-STREAM-INF:BANDWIDTH=809659,RESOLUTION=1280x720\nindex-f2.m3u8\n")
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

// TestHiAnimeAdapter_SearchToStream drives the full chain through the real
// adapter: search, episodes, then the ZokoAnime stream with its metadata.
func TestHiAnimeAdapter_SearchToStream(t *testing.T) {
	t.Parallel()

	srv := hiAnimeStub(t)
	adapter := NewHiAnimeAdapterForTest(srv.URL)

	results, err := adapter.SearchAnime("cowboy bebop")
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "Cowboy Bebop", results[0].Name)
	assert.Equal(t, srv.URL+"/cowboy-bebop-26", results[0].URL)

	episodes, err := adapter.GetAnimeEpisodes(results[0].URL)
	require.NoError(t, err)
	require.Len(t, episodes, 2)
	assert.Equal(t, "1", episodes[0].Number)
	assert.Equal(t, "Ballad", episodes[0].Title.English)
	assert.True(t, strings.HasSuffix(episodes[0].URL, "?ep=9001"))

	stream, meta, err := adapter.GetStreamURL(episodes[0].URL, "best")
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/master.m3u8", stream)
	assert.Equal(t, "hianime", meta["source"])
	assert.Equal(t, srv.URL+"/", meta["referer"], "the CDN 403s the master without the embed origin")
	assert.Equal(t, "sub", meta["audio_lang"])
	assert.Contains(t, meta["subtitles"], srv.URL+"/en.vtt")
}

// TestHiAnimeAdapter_QualityReachesTheClient proves the quality option
// survives the adapter instead of being dropped in the glue.
func TestHiAnimeAdapter_QualityReachesTheClient(t *testing.T) {
	t.Parallel()

	srv := hiAnimeStub(t)
	adapter := NewHiAnimeAdapterForTest(srv.URL)

	episodes, err := adapter.GetAnimeEpisodes(srv.URL + "/cowboy-bebop-26")
	require.NoError(t, err)

	stream, _, err := adapter.GetStreamURL(episodes[0].URL, "720p")
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/index-f2.m3u8", stream)
}

// TestHiAnimeAdapter_GetType pins the registry identity.
func TestHiAnimeAdapter_GetType(t *testing.T) {
	t.Parallel()

	adapter := NewHiAnimeAdapterForTest("https://stub.test")
	assert.Equal(t, HiAnimeType, adapter.GetType())
	assert.Equal(t, "HiAnime", scraperDisplayName(HiAnimeType))
}
