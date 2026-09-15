package anidb

import (
	"context"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestHostIsPinned fails loudly if the base host rotates without a deliberate
// edit, per the provider contract (dated 2026-08-22).
func TestHostIsPinned(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "https://anidb.app", anidbBase,
		"host rotated? update the const and this assertion together")
}

// ---------------------------------------------------------------------------
// Identity
// ---------------------------------------------------------------------------

func TestAnimeID(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"full permalink", "https://anidb.app/anime/jojos-bizarre-adventure-golden-wind-2534", "2534", false},
		{"permalink with trailing slash", "https://anidb.app/anime/cowboy-bebop-42/", "42", false},
		{"bare numeric id", "2534", "2534", false},
		{"slug-id without prefix", "cowboy-bebop-42", "42", false},
		{"empty", "", "", true},
		{"unrelated URL", "https://animefire.io/animes/naruto", "", true},
		{"slug with no id", "https://anidb.app/anime/naruto", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := AnimeID(tt.input)
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
	tests := []struct {
		name    string
		input   string
		want    string
		wantErr bool
	}{
		{"canonical", "https://anidb.app/episode/20049", "20049", false},
		{"trailing slash", "https://anidb.app/episode/20049/", "20049", false},
		{"bare id", "20049", "20049", false},
		{"empty", "", "", true},
		{"anime URL is not an episode URL", "https://anidb.app/anime/x-1", "", true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := EpisodeID(tt.input)
			if tt.wantErr {
				require.Error(t, err)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestNormalizeQuality(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want int
	}{
		{"1080p", 1080}, {"1080", 1080}, {"720p", 720}, {"  480P ", 480},
		{"best", 0}, {"", 0}, {"auto", 0}, {"worst", 0}, {"garbage", 0},
	}
	for _, tt := range tests {
		t.Run(tt.in, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, normalizeQuality(tt.in))
		})
	}
}

func TestSelectLanguage(t *testing.T) {
	t.Parallel()
	langs := []apiLanguage{
		{Code: "eng", Name: "English", EmbedURL: "https://x/eng"},
		{Code: "jpn", Name: "Japanese", EmbedURL: "https://x/jpn"},
	}

	t.Run("prefers the requested code", func(t *testing.T) {
		t.Parallel()
		got, ok := selectLanguage(langs, "jpn")
		require.True(t, ok)
		assert.Equal(t, "jpn", got.Code)
	})

	t.Run("falls back rather than failing the episode", func(t *testing.T) {
		t.Parallel()
		got, ok := selectLanguage(langs, "por")
		require.True(t, ok)
		assert.NotEmpty(t, got.EmbedURL)
	})

	t.Run("skips entries with no embed URL", func(t *testing.T) {
		t.Parallel()
		got, ok := selectLanguage([]apiLanguage{
			{Code: "jpn", EmbedURL: ""},
			{Code: "eng", EmbedURL: "https://x/eng"},
		}, "jpn")
		require.True(t, ok)
		assert.Equal(t, "eng", got.Code, "an entry without an embed URL is not playable")
	})

	t.Run("no playable language", func(t *testing.T) {
		t.Parallel()
		_, ok := selectLanguage([]apiLanguage{{Code: "jpn"}}, "jpn")
		assert.False(t, ok)
	})
}

func TestResolveRef(t *testing.T) {
	t.Parallel()
	const base = "https://hls.anidb.app/stream/abc/master.m3u8"
	assert.Equal(t, "https://hls.anidb.app/stream/abc/index-f1.m3u8", resolveRef(base, "index-f1.m3u8"))
	assert.Equal(t, "https://hls.anidb.app/other.m3u8", resolveRef(base, "/other.m3u8"))
	assert.Equal(t, "https://cdn.example/x.m3u8", resolveRef(base, "https://cdn.example/x.m3u8"))
}

// ---------------------------------------------------------------------------
// Search
// ---------------------------------------------------------------------------

// searchPage mirrors the real browse markup: an anchor per card carrying the
// permalink, a title attribute and a poster.
const searchPage = `<!doctype html><html><body>
<div class="anime-grid">
  <a href="https://anidb.app/anime/steel-ball-run-jojos-bizarre-adventure-4979" class="anime-card" title="Steel Ball Run: JoJo&#039;s Bizarre Adventure">
    <img src="https://cdn.example/4979.jpg" alt="Steel Ball Run: JoJo&#039;s Bizarre Adventure">
  </a>
  <a href="https://anidb.app/anime/jojos-bizarre-adventure-golden-wind-2534" class="anime-card">
    <img src="https://cdn.example/2534.jpg" alt="JoJo&#039;s Bizarre Adventure: Golden Wind">
  </a>
  <a href="https://anidb.app/anime/jojos-bizarre-adventure-golden-wind-2534" class="anime-card" title="Duplicate card">
    <img src="https://cdn.example/2534.jpg" alt="dup">
  </a>
  <a href="/browse?sort=order_trending" class="btn-secondary">Clear</a>
  <a href="https://anidb.app/anime/no-title-here-77" class="anime-card"></a>
</div></body></html>`

func TestSearchAnime(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/browse", r.URL.Path)
		assert.Equal(t, "jojo", r.URL.Query().Get("q"))
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = w.Write([]byte(searchPage))
	}))
	defer srv.Close()

	got, err := NewClientForTest(srv.URL).SearchAnime(context.Background(), "jojo")
	require.NoError(t, err)
	require.Len(t, got, 2, "duplicates, nav links and title-less cards must be dropped")

	assert.Equal(t, "Steel Ball Run: JoJo's Bizarre Adventure", got[0].Name,
		"HTML entities in the title must be decoded")
	assert.Equal(t, srv.URL+"/anime/steel-ball-run-jojos-bizarre-adventure-4979", got[0].URL)
	assert.Equal(t, "https://cdn.example/4979.jpg", got[0].ImageURL)
	assert.Equal(t, sourceLabel, got[0].Source)

	assert.Equal(t, "JoJo's Bizarre Adventure: Golden Wind", got[1].Name,
		"a card without a title attribute must fall back to the poster alt text")
}

func TestSearchAnime_EmptyQuery(t *testing.T) {
	t.Parallel()
	_, err := NewClientForTest("http://127.0.0.1:1").SearchAnime(context.Background(), "   ")
	require.Error(t, err)
}

func TestSearchAnime_NoResults(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body><div class="anime-grid"></div></body></html>`))
	}))
	defer srv.Close()

	got, err := NewClientForTest(srv.URL).SearchAnime(context.Background(), "nothing")
	require.NoError(t, err, "an empty result set is not an error")
	assert.Empty(t, got)
}

func TestSearchAnime_ServerError(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	_, err := NewClientForTest(srv.URL).SearchAnime(context.Background(), "jojo")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "500")
}

func TestSearchAnime_ChallengePageIsDetected(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>Just a moment...</title></head>
<body><h1>Checking your browser before accessing</h1><p>Ray ID: 1234</p></body></html>`))
	}))
	defer srv.Close()

	_, err := NewClientForTest(srv.URL).SearchAnime(context.Background(), "jojo")
	require.Error(t, err, "an anti-bot interstitial must not be parsed as results")
}

// ---------------------------------------------------------------------------
// Episodes
// ---------------------------------------------------------------------------

// episodesFixture is the canonical /episodes payload. It is shared with
// meta_test.go, which mutates it to prove these tests actually have teeth.
const episodesFixture = `{"episodes":[
	{"id":20051,"number":3,"number2":null,"filler":true},
	{"id":20049,"number":1,"number2":null,"filler":false},
	{"id":20050,"number":2,"number2":null,"filler":false},
	{"id":0,"number":4,"filler":false}
]}`

// languagesFixture is the canonical /languages payload, with %s standing in for
// the test server's base URL.
const languagesFixture = `{"languages":[
	{"code":"eng","name":"English","embed_url":"%s/embed/eng"},
	{"code":"jpn","name":"Japanese","embed_url":"%s/embed/jpn"}
]}`

// embedFixture is the canonical embed page, with %s for the master playlist URL.
const embedFixture = `<html><script>jwplayer().setup({file: '%s'});</script></html>`

func TestGetAnimeEpisodes(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/api/frontend/anime/2534/episodes", r.URL.Path)
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(episodesFixture))
	}))
	defer srv.Close()

	got, err := NewClientForTest(srv.URL).GetAnimeEpisodes(context.Background(), srv.URL+"/anime/golden-wind-2534")
	require.NoError(t, err)
	require.Len(t, got, 3, "an entry without an id cannot be streamed and must be dropped")

	assert.Equal(t, []string{"1", "2", "3"}, []string{got[0].Number, got[1].Number, got[2].Number},
		"episodes must come back in ascending order regardless of API order")
	assert.Equal(t, srv.URL+"/episode/20049", got[0].URL)
	assert.Equal(t, "20049", got[0].DataID)
	assert.True(t, got[2].IsFiller, "the filler flag must survive")
}

func TestGetAnimeEpisodes_EmptyList(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"episodes":[]}`))
	}))
	defer srv.Close()

	_, err := NewClientForTest(srv.URL).GetAnimeEpisodes(context.Background(), "2534")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no episodes")
}

func TestGetAnimeEpisodes_BadURL(t *testing.T) {
	t.Parallel()
	_, err := NewClientForTest("http://127.0.0.1:1").GetAnimeEpisodes(context.Background(), "https://example.com/whatever")
	require.Error(t, err)
}

func TestGetAnimeEpisodes_MalformedJSON(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"episodes":[{"id":1,`))
	}))
	defer srv.Close()

	_, err := NewClientForTest(srv.URL).GetAnimeEpisodes(context.Background(), "2534")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Stream
// ---------------------------------------------------------------------------

const masterPlaylist = `#EXTM3U
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=1218557,RESOLUTION=1920x1080,CODECS="avc1.64001f"
index-f1-v1-a1.m3u8
#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=809659,RESOLUTION=1280x720,CODECS="avc1.64001f"
index-f2-v1-a1.m3u8
#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=192051,RESOLUTION=1920x1080,URI="iframes-f1.m3u8"
`

// streamServer serves the languages endpoint, the embed page and the master
// playlist — the whole stream chain.
func streamServer(t *testing.T) *httptest.Server {
	t.Helper()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, "/languages"):
			w.Header().Set("Content-Type", "application/json")
			fmt.Fprintf(w, languagesFixture, srv.URL, srv.URL)
		case strings.HasPrefix(r.URL.Path, "/embed/"):
			lang := strings.TrimPrefix(r.URL.Path, "/embed/")
			fmt.Fprintf(w, embedFixture, fmt.Sprintf("%s/stream/%s/master.m3u8", srv.URL, lang))
		case strings.HasSuffix(r.URL.Path, "master.m3u8"):
			_, _ = w.Write([]byte(masterPlaylist))
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	t.Cleanup(srv.Close)
	return srv
}

func TestGetEpisodeStreamURL_BestReturnsMaster(t *testing.T) {
	t.Parallel()
	srv := streamServer(t)

	got, meta, err := NewClientForTest(srv.URL).GetEpisodeStreamURL(context.Background(), srv.URL+"/episode/20049", "best")
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/stream/jpn/master.m3u8", got,
		"best must hand the master playlist to the player")
	assert.Equal(t, "jpn", meta["audio_lang"], "subbed is the default")
	assert.Equal(t, srv.URL+"/", meta["referer"])
	assert.Equal(t, "anidb", meta["source"])
}

func TestGetEpisodeStreamURL_PicksTheRequestedVariant(t *testing.T) {
	t.Parallel()
	srv := streamServer(t)

	got, _, err := NewClientForTest(srv.URL).GetEpisodeStreamURL(context.Background(), srv.URL+"/episode/20049", "720p")
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/stream/jpn/index-f2-v1-a1.m3u8", got,
		"a relative variant must be resolved against the master URL")
}

func TestGetEpisodeStreamURL_UnavailableQualityFallsBackToMaster(t *testing.T) {
	t.Parallel()
	srv := streamServer(t)

	got, _, err := NewClientForTest(srv.URL).GetEpisodeStreamURL(context.Background(), srv.URL+"/episode/20049", "2160p")
	require.NoError(t, err)
	assert.Equal(t, srv.URL+"/stream/jpn/master.m3u8", got,
		"an absent quality must not fail the episode")
}

func TestGetEpisodeStreamURL_DubHonoursTheEnvOverride(t *testing.T) {
	// Not parallel: mutates the process environment.
	t.Setenv("GOANIME_ANIDB_LANG", "dub")
	srv := streamServer(t)

	got, meta, err := NewClientForTest(srv.URL).GetEpisodeStreamURL(context.Background(), srv.URL+"/episode/20049", "best")
	require.NoError(t, err)
	assert.Equal(t, "eng", meta["audio_lang"])
	assert.Contains(t, got, "/stream/eng/")
}

func TestGetEpisodeStreamURL_NoLanguages(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`{"languages":[]}`))
	}))
	defer srv.Close()

	_, _, err := NewClientForTest(srv.URL).GetEpisodeStreamURL(context.Background(), "20049", "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no playable language")
}

func TestGetEpisodeStreamURL_EmbedWithoutPlaylist(t *testing.T) {
	t.Parallel()
	var srv *httptest.Server
	srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, "/languages") {
			fmt.Fprintf(w, `{"languages":[{"code":"jpn","embed_url":"%s/embed/jpn"}]}`, srv.URL)
			return
		}
		_, _ = w.Write([]byte(`<html><body>player layout changed</body></html>`))
	}))
	defer srv.Close()

	_, _, err := NewClientForTest(srv.URL).GetEpisodeStreamURL(context.Background(), "20049", "best")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no m3u8")
}

func TestGetEpisodeStreamURL_BadEpisodeURL(t *testing.T) {
	t.Parallel()
	_, _, err := NewClientForTest("http://127.0.0.1:1").GetEpisodeStreamURL(context.Background(), "https://example.com/x", "best")
	require.Error(t, err)
}

// ---------------------------------------------------------------------------
// Cancellation
// ---------------------------------------------------------------------------

// TestSearchAnime_CancellationAbortsInFlightRequest is the point of threading a
// context all the way down: when the dispatcher's per-source deadline fires, the
// HTTP request it was meant to abort must actually stop, not keep running until
// the client timeout. The three older providers cannot do this (dispatch.go
// calls it out); this one can.
func TestSearchAnime_CancellationAbortsInFlightRequest(t *testing.T) {
	t.Parallel()

	released := make(chan struct{})
	handlerReturned := make(chan struct{})
	srv := httptest.NewServer(http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) {
		defer close(handlerReturned)
		select {
		case <-r.Context().Done(): // the client hung up: cancellation reached the server
		case <-released:
		}
	}))
	defer srv.Close()
	defer close(released)

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := NewClientForTest(srv.URL).SearchAnime(ctx, "jojo")
		errCh <- err
	}()

	// Let the request reach the (blocking) handler, then cancel.
	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		require.Error(t, err)
		assert.ErrorIs(t, err, context.Canceled,
			"the cancellation must surface, not be masked as a generic failure")
	case <-time.After(5 * time.Second):
		t.Fatal("cancelling the context did not abort the request")
	}

	select {
	case <-handlerReturned:
	case <-time.After(5 * time.Second):
		t.Fatal("the server never saw the client hang up")
	}
}

// TestRetryBackoff_HonoursCancellation covers the other half: a cancelled call
// must not sit through its retry delay.
func TestRetryBackoff_HonoursCancellation(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError) // 5xx → retry path
	}))
	defer srv.Close()

	c := NewClientForTest(srv.URL)
	c.maxRetries = 3
	c.retryDelay = 30 * time.Second // absurd on purpose: only cancellation can cut it short

	ctx, cancel := context.WithCancel(context.Background())
	errCh := make(chan error, 1)
	go func() {
		_, err := c.SearchAnime(ctx, "jojo")
		errCh <- err
	}()

	time.Sleep(50 * time.Millisecond)
	cancel()

	select {
	case err := <-errCh:
		require.Error(t, err)
	case <-time.After(5 * time.Second):
		t.Fatal("the retry backoff ignored cancellation")
	}
}
