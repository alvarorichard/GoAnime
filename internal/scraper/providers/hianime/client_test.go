package hianime

import (
	"context"
	"encoding/base64"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHostIsPinned fails loudly if the base host or the API prefix rotates
// without a deliberate edit, per the provider contract.
func TestHostIsPinned(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "https://hianime.at", hianimeBase)
	assert.Equal(t, "/api/theme/", restPath)
	assert.Equal(t, "https://hianime.at", NewHiAnimeClient().baseURL)
}

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

func TestAnimeID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, in, want string
		wantErr        bool
	}{
		{name: "permalink", in: "https://hianime.at/naruto-1335", want: "1335"},
		{name: "hyphenated slug", in: "https://hianime.at/bleach-thousand-year-blood-war-the-calamity-5", want: "5"},
		{name: "watch URL", in: "https://hianime.at/watch/naruto-1335?ep=22676", want: "1335"},
		{name: "trailing slash", in: "https://hianime.at/naruto-1335/", want: "1335"},
		{name: "bare id", in: "1335", want: "1335"},
		{name: "empty", in: "", wantErr: true},
		{name: "no id", in: "https://hianime.at/az-list/all", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := AnimeID(tt.in)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestEpisodeID(t *testing.T) {
	t.Parallel()
	got, err := EpisodeID("https://hianime.at/watch/naruto-1335?ep=22676")
	require.NoError(t, err)
	assert.Equal(t, "22676", got)

	got, err = EpisodeID("22676")
	require.NoError(t, err)
	assert.Equal(t, "22676", got)

	_, err = EpisodeID("")
	require.Error(t, err)
}

// A watch URL carries the anime id AND the episode id. Falling back to the
// trailing number when ?ep= is missing would silently list anime 1335's
// servers as if they were episode 1335's — a wrong stream, not an error.
func TestEpisodeID_RefusesTheAnimeIDInAWatchURL(t *testing.T) {
	t.Parallel()
	_, err := EpisodeID("https://hianime.at/watch/naruto-1335")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "?ep=")
}

func TestNormalizeQuality(t *testing.T) {
	t.Parallel()
	for in, want := range map[string]int{
		"1080p": 1080, "720": 720, "480P": 480,
		"": 0, "best": 0, "auto": 0, "hd": 0,
	} {
		assert.Equal(t, want, normalizeQuality(in), "input %q", in)
	}
}

// Not parallel: the subtests set an environment variable, which t.Setenv
// refuses to do under a parallel parent.
func TestPreferredAudio(t *testing.T) {
	assert.Equal(t, audioSub, preferredAudio(), "subtitled is the default")

	t.Run("override", func(t *testing.T) {
		t.Setenv("GOANIME_HIANIME_AUDIO", "dub")
		assert.Equal(t, audioDub, preferredAudio())
	})
	t.Run("unknown value falls back to sub", func(t *testing.T) {
		t.Setenv("GOANIME_HIANIME_AUDIO", "klingon")
		assert.Equal(t, audioSub, preferredAudio())
	})
}

func TestResolveRef(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "https://cdn.example/v/abc/720/index.m3u8",
		resolveRef("https://cdn.example/v/abc/master.m3u8", "720/index.m3u8"))
	assert.Equal(t, "https://other.example/x.m3u8",
		resolveRef("https://cdn.example/v/abc/master.m3u8", "https://other.example/x.m3u8"))
}

func TestOriginOf(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "https://zokoanime.video", originOf("https://zokoanime.video/stream/mal/20/1/sub"))
	assert.Empty(t, originOf("not a url"))
}

func TestSelectServer(t *testing.T) {
	t.Parallel()
	servers := []server{
		{Name: "HD-1", Audio: "sub", Embed: "https://megaplay.example/a"},
		{Name: "ZokoAnime", Audio: "dub", Embed: "https://zoko.example/dub"},
		{Name: "ZokoAnime", Audio: "sub", Embed: "https://zoko.example/sub"},
	}

	got, ok := selectServer(servers, audioSub)
	require.True(t, ok)
	assert.Equal(t, "https://zoko.example/sub", got.Embed, "the preferred audio wins")

	got, ok = selectServer(servers, audioDub)
	require.True(t, ok)
	assert.Equal(t, "https://zoko.example/dub", got.Embed)

	// An episode that exists in one language only must still play.
	got, ok = selectServer(servers[:2], audioSub)
	require.True(t, ok)
	assert.Equal(t, "https://zoko.example/dub", got.Embed, "any readable server beats no stream")

	// Nothing readable is a real failure, not a silent pick of MegaPlay.
	_, ok = selectServer(servers[:1], audioSub)
	assert.False(t, ok)
}

// ---------------------------------------------------------------------------
// Embed payload
// ---------------------------------------------------------------------------

// obfuscate reproduces the player's encoding: JSON, XOR'd with the build tag,
// base64'd. Written out from the format rather than calling the decoder's own
// constant path, so an accidental change cannot cancel itself out.
func obfuscate(payload string) string {
	key := []byte("otaku-embed-v1")
	raw := []byte(payload)
	for i := range raw {
		raw[i] ^= key[i%len(key)]
	}
	return base64.StdEncoding.EncodeToString(raw)
}

func embedPage(config string) string {
	return `<!doctype html><html><body><div id="player"></div>` +
		`<script>window.__P="` + obfuscate(config) + `"</script></body></html>`
}

func TestDecodePayload(t *testing.T) {
	t.Parallel()
	page := embedPage(`{"src":"https://cdn.example/master.m3u8","subtitles":[` +
		`{"lang":"en","label":"English","src":"https://cdn.example/en.vtt","default":true}],` +
		`"skip":{"intro":{"start":0,"end":107},"outro":null}}`)

	got, err := decodePayload([]byte(page))
	require.NoError(t, err)
	assert.Equal(t, "https://cdn.example/master.m3u8", got.Src)
	require.Len(t, got.Subtitles, 1)
	assert.Equal(t, "English", got.Subtitles[0].Label)
	require.NotNil(t, got.Skip.Intro)
	assert.Equal(t, 107, got.Skip.Intro.End)
}

func TestDecodePayload_Failures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, page, wantMsg string
	}{
		{
			name:    "no payload at all",
			page:    `<html><script>var x = 1</script></html>`,
			wantMsg: "no player payload",
		},
		{
			name:    "payload is not valid base64",
			page:    `<html><script>window.__P="not/valid/base64=="</script></html>`,
			wantMsg: "base64",
		},
		{
			name: "the obfuscation changed",
			page: `<html><script>window.__P="` +
				base64.StdEncoding.EncodeToString([]byte(`{"src":"https://cdn.example/x.m3u8"}`)) +
				`"</script></html>`,
			wantMsg: "obfuscation changed",
		},
		{
			name:    "config carries no stream",
			page:    embedPage(`{"src":"","subtitles":[]}`),
			wantMsg: "no stream URL",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			_, err := decodePayload([]byte(tt.page))
			require.Error(t, err)
			assert.Contains(t, err.Error(), tt.wantMsg)
		})
	}
}

// ---------------------------------------------------------------------------
// Fixtures
// ---------------------------------------------------------------------------

// searchPage mirrors the real markup: a .flw-item per result, the permalink and
// title on .film-name a, the poster on .film-poster-img. The chrome around it
// is not decoration — the A-Z index below carries hrefs that a bare
// "<slug>-<digits>" pattern matches, and must not be read as results.
const searchPage = `<html><body>
	<div class="header"><a href="/az-list/0-9">0-9</a><a href="/az-list/a">A</a></div>
	<div class="film_list-wrap">
		<div class="flw-item">
			<div class="film-poster">
				<img src="https://cdn.example/26.jpg" class="film-poster-img" alt="Cowboy Bebop">
				<a href="https://hianime.at/watch/cowboy-bebop-26" class="film-poster-ahref" data-id="26"></a>
			</div>
			<div class="film-detail"><h3 class="film-name">
				<a href="https://hianime.at/cowboy-bebop-26" title="Cowboy Bebop" data-jname="Cowboy Bebop">
					Cowboy Bebop
				</a>
			</h3></div>
		</div>
		<div class="flw-item">
			<div class="film-poster">
				<img data-src="https://cdn.example/17.jpg" class="film-poster-img" alt="Cowboy Bebop: The Movie">
			</div>
			<div class="film-detail"><h3 class="film-name">
				<a href="https://hianime.at/cowboy-bebop-the-movie-17" title="Cowboy Bebop: The Movie">x</a>
			</h3></div>
		</div>
	</div>
	<div class="footer"><a href="/contact-2">Contact</a></div>
</body></html>`

// episodesFixture is the canonical episode-list envelope: rendered HTML inside
// JSON. Shared with meta_test.go, which mutates it to prove these tests have
// teeth.
const episodesFixture = `{"status":true,"totalItems":2,"html":"` +
	`<div class=\"ss-list ss-list-min active\">` +
	`<a title=\"Asteroid Blues\" class=\"ssl-item ep-item\" data-number=\"1\" data-id=\"9001\">` +
	`<div class=\"ep-name\" title=\"Asteroid Blues\">Asteroid Blues</div></a>` +
	`<a title=\"Stray Dog Strut\" class=\"ssl-item ep-item ssl-item-filler\" data-number=\"2\" data-id=\"9002\">` +
	`<div class=\"ep-name\" title=\"Stray Dog Strut\">Stray Dog Strut</div></a>` +
	`</div>"}`

// serversFixture is the canonical server list, with %s twice for the test
// server's base64'd embed URLs. MegaPlay is listed first on purpose.
const serversFixture = `{"status":true,"html":"` +
	`<div class=\"item server-item\" data-type=\"sub\" data-server-name=\"HD-1\" data-hash=\"bWVnYXBsYXk=\"></div>` +
	`<div class=\"item server-item\" data-type=\"sub\" data-server-name=\"ZokoAnime\" data-hash=\"%s\"></div>` +
	`<div class=\"item server-item\" data-type=\"dub\" data-server-name=\"ZokoAnime\" data-hash=\"%s\"></div>` +
	`"}`

// masterPlaylist is a two-variant multivariant playlist.
const masterPlaylist = "#EXTM3U\n" +
	"#EXT-X-STREAM-INF:BANDWIDTH=1218557,RESOLUTION=1920x1080\nindex-f1.m3u8\n" +
	"#EXT-X-STREAM-INF:BANDWIDTH=809659,RESOLUTION=1280x720\nindex-f2.m3u8\n"

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

func TestSearchAnime(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/search", r.URL.Path)
		assert.Equal(t, "cowboy bebop", r.URL.Query().Get("keyword"))
		w.Header().Set("Content-Type", "text/html")
		_, _ = w.Write([]byte(searchPage))
	}))
	defer srv.Close()

	got, err := NewClientForTest(srv.URL).SearchAnime(context.Background(), "cowboy bebop")
	require.NoError(t, err)
	require.Len(t, got, 2, "the A-Z index and footer links must not be read as results")

	assert.Equal(t, "Cowboy Bebop", got[0].Name)
	assert.Equal(t, srv.URL+"/cowboy-bebop-26", got[0].URL)
	assert.Equal(t, "https://cdn.example/26.jpg", got[0].ImageURL)
	assert.Equal(t, "HiAnime", got[0].Source)

	assert.Equal(t, "Cowboy Bebop: The Movie", got[1].Name)
	assert.Equal(t, "https://cdn.example/17.jpg", got[1].ImageURL, "lazy-loaded posters live in data-src")
}

func TestSearchAnime_EmptyQuery(t *testing.T) {
	t.Parallel()
	_, err := NewHiAnimeClient().SearchAnime(context.Background(), "   ")
	require.Error(t, err)
}

func TestSearchAnime_NoResults(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body><div class="film_list-wrap"></div></body></html>`))
	}))
	defer srv.Close()

	got, err := NewClientForTest(srv.URL).SearchAnime(context.Background(), "nothing")
	require.NoError(t, err, "an empty catalogue is not an error")
	assert.Empty(t, got)
}

func TestSearchAnime_ServerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	defer srv.Close()

	_, err := NewClientForTest(srv.URL).SearchAnime(context.Background(), "x")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "503")
}

func TestSearchAnime_ChallengePageIsDetected(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>Just a moment...</title></head>
			<body><div id="challenge-running"></div></body></html>`))
	}))
	defer srv.Close()

	_, err := NewClientForTest(srv.URL).SearchAnime(context.Background(), "x")
	require.Error(t, err, "a challenge page must not be reported as zero results")
}

// ---------------------------------------------------------------------------
// Episodes
// ---------------------------------------------------------------------------

func episodeServer(t *testing.T, body string) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/theme/episode/list/26", r.URL.Path)
		assert.Equal(t, "XMLHttpRequest", r.Header.Get("X-Requested-With"))
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestGetAnimeEpisodes(t *testing.T) {
	t.Parallel()
	srv := episodeServer(t, episodesFixture)

	got, err := NewClientForTest(srv.URL).GetAnimeEpisodes(context.Background(), srv.URL+"/cowboy-bebop-26")
	require.NoError(t, err)
	require.Len(t, got, 2)

	assert.Equal(t, "1", got[0].Number)
	assert.Equal(t, 1, got[0].Num)
	assert.Equal(t, "9001", got[0].DataID)
	assert.Equal(t, "Asteroid Blues", got[0].Title.English)
	assert.False(t, got[0].IsFiller)
	assert.Equal(t, srv.URL+"/watch/cowboy-bebop-26?ep=9001", got[0].URL,
		"the episode URL is the only channel back to the stream call")
	assert.True(t, got[1].IsFiller)
}

func TestGetAnimeEpisodes_BadURL(t *testing.T) {
	t.Parallel()
	_, err := NewHiAnimeClient().GetAnimeEpisodes(context.Background(), "https://hianime.at/az-list/all")
	require.Error(t, err)
}

func TestGetAnimeEpisodes_Failures(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name, body, wantMsg string
	}{
		{
			name:    "envelope says no",
			body:    `{"status":false,"html":""}`,
			wantMsg: "no content",
		},
		{
			name:    "malformed JSON",
			body:    `{"status":true,"html":`,
			wantMsg: "malformed JSON",
		},
		{
			name:    "fragment has no entries",
			body:    `{"status":true,"html":"<div class=\"ss-list\"></div>"}`,
			wantMsg: "no entries",
		},
		{
			name:    "entries carry no ids",
			body:    `{"status":true,"html":"<a class=\"ssl-item ep-item\" data-number=\"1\"></a>"}`,
			wantMsg: "none carried a usable id",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			srv := episodeServer(t, tt.body)
			_, err := NewClientForTest(srv.URL).GetAnimeEpisodes(context.Background(), srv.URL+"/cowboy-bebop-26")
			require.Error(t, err, "an unusable list must not read as an anime with no episodes")
			assert.Contains(t, err.Error(), tt.wantMsg)
		})
	}
}

// ---------------------------------------------------------------------------
// Stream
// ---------------------------------------------------------------------------

// streamServer serves the server list, the embed page and the master playlist —
// the whole stream chain.
func streamServer(t *testing.T) *httptest.Server {
	t.Helper()
	var base string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case r.URL.Path == "/api/theme/episode/servers":
			assert.Equal(t, "9001", r.URL.Query().Get("episodeId"))
			b64 := func(s string) string { return base64.StdEncoding.EncodeToString([]byte(s)) }
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, serversFixture, b64(base+"/embed/sub"), b64(base+"/embed/dub"))

		case strings.HasPrefix(r.URL.Path, "/embed/"):
			audio := strings.TrimPrefix(r.URL.Path, "/embed/")
			_, _ = w.Write([]byte(embedPage(fmt.Sprintf(
				`{"src":"%s/stream/%s/master.m3u8","subtitles":[`+
					`{"lang":"en","label":"English","src":"%s/en.vtt","default":true},`+
					`{"lang":"en","label":"Portuguese (- Brazilian)","src":"%s/pt.vtt"},`+
					`{"lang":"en","label":"Broken","src":""}]}`,
				base, audio, base, base))))

		case strings.HasSuffix(r.URL.Path, "master.m3u8"):
			_, _ = w.Write([]byte(masterPlaylist))

		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	base = srv.URL
	t.Cleanup(srv.Close)
	return srv
}

func TestGetEpisodeStreamURL_BestReturnsMaster(t *testing.T) {
	t.Parallel()
	srv := streamServer(t)

	got, meta, err := NewClientForTest(srv.URL).GetEpisodeStreamURL(
		context.Background(), srv.URL+"/watch/cowboy-bebop-26?ep=9001", "best")
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/stream/sub/master.m3u8", got)
	assert.Equal(t, "hianime", meta["source"])
	assert.Equal(t, audioSub, meta["audio_lang"])
	assert.Equal(t, srv.URL+"/", meta["referer"],
		"the CDN 403s part of its URLs without the embed host's referer")
}

func TestGetEpisodeStreamURL_PicksTheRequestedVariant(t *testing.T) {
	t.Parallel()
	srv := streamServer(t)

	got, _, err := NewClientForTest(srv.URL).GetEpisodeStreamURL(
		context.Background(), srv.URL+"/watch/cowboy-bebop-26?ep=9001", "720p")
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/stream/sub/index-f2.m3u8", got)
}

func TestGetEpisodeStreamURL_UnavailableQualityKeepsTheMaster(t *testing.T) {
	t.Parallel()
	srv := streamServer(t)

	got, _, err := NewClientForTest(srv.URL).GetEpisodeStreamURL(
		context.Background(), srv.URL+"/watch/cowboy-bebop-26?ep=9001", "2160p")
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/stream/sub/master.m3u8", got,
		"a missing quality must fall back to the master, not fail the episode")
}

func TestGetEpisodeStreamURL_HonoursTheAudioPreference(t *testing.T) {
	t.Setenv("GOANIME_HIANIME_AUDIO", "dub")
	srv := streamServer(t)

	got, meta, err := NewClientForTest(srv.URL).GetEpisodeStreamURL(
		context.Background(), srv.URL+"/watch/cowboy-bebop-26?ep=9001", "best")
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/stream/dub/master.m3u8", got)
	assert.Equal(t, audioDub, meta["audio_lang"])
}

// The subtitle list is what reaches mpv, and the payload's own "lang" field is
// unusable: every track on the live episodes sampled claimed "en", Portuguese
// and Spanish included. The label is what identifies a track.
func TestGetEpisodeStreamURL_CarriesSubtitlesByLabel(t *testing.T) {
	t.Parallel()
	srv := streamServer(t)

	_, meta, err := NewClientForTest(srv.URL).GetEpisodeStreamURL(
		context.Background(), srv.URL+"/watch/cowboy-bebop-26?ep=9001", "best")
	require.NoError(t, err)

	assert.Contains(t, meta["subtitles"], `"label":"Portuguese (- Brazilian)"`)
	assert.Contains(t, meta["subtitles"], `"language":"Portuguese (- Brazilian)"`)
	assert.NotContains(t, meta["subtitles"], `"label":"Broken"`, "a track with no URL is not a track")
}

func TestGetEpisodeStreamURL_NoReadableServer(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(`{"status":true,"html":"` +
			`<div class=\"item server-item\" data-type=\"sub\" data-server-name=\"HD-1\" data-hash=\"bWVnYQ==\"></div>"}`))
	}))
	defer srv.Close()

	_, _, err := NewClientForTest(srv.URL).GetEpisodeStreamURL(
		context.Background(), srv.URL+"/watch/cowboy-bebop-26?ep=9001", "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ZokoAnime")
}
