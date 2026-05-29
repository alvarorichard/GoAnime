package scraper

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/internal/models"
)

const animeWorldSearchFixture = `<html><body>
<div class="film-list">
    <div class="item"><div class="inner">
        <a href="/play/naruto.abc123" class="poster">
            <img src="https://img.example.com/naruto.jpg" alt="Naruto">
        </a>
        <a href="/play/naruto.abc123" class="name">Naruto</a>
    </div></div>
    <div class="item"><div class="inner">
        <a href="/play/naruto-shippuden.def456" class="poster">
            <img src="https://img.example.com/shippuden.jpg" alt="Naruto Shippuden">
        </a>
        <a href="/play/naruto-shippuden.def456" class="name">Naruto Shippuden</a>
    </div></div>
</div>
</body></html>`

const animeWorldEpisodesFixture = `
<html>
	<body>
      <ul class="episodes">
        <li class="episode"> <a href="/play/naruto-ita1" data-episode-num="1"> 1 </a></li>
        <li class="episode"> <a href="/play/naruto-ita2" data-episode-num="2"> 2 </a></li>
        <li class="episode"> <a href="/play/naruto-ita3" data-episode-num="3"> 3 </a></li>
      </ul>
      <ul class="episodes">
        <li class="episode"> <a href="/play/naruto-ita4" data-episode-num="4"> 4 </a></li>
        <li class="episode"> <a href="/play/naruto-ita5" data-episode-num="5"> 5 </a></li>
        <li class="episode"> <a href="/play/naruto-ita6" data-episode-num="6"> 6 </a></li>
      </ul>
	</body>
</html>
`

// --- SearchAnime ---

func TestAnimeWorldSearchAnime(t *testing.T) {
	t.Parallel()

	var receivedKeyword string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedKeyword = r.URL.Query().Get("keyword")
		_, _ = fmt.Fprint(w, animeWorldSearchFixture)
	}))
	defer srv.Close()

	client := NewAnimeWorldClient()
	client.baseURL = srv.URL
	client.maxRetries = 0
	client.retryDelay = 0

	results, err := client.SearchAnime("naruto-shippuden")
	require.NoError(t, err)
	require.Len(t, results, 2)

	assert.Equal(t, "naruto shippuden", receivedKeyword, "hyphens should normalize to spaces")
	assert.Equal(t, "Naruto", results[0].Name)
	assert.Equal(t, animeWorldSource, results[0].Source)
	assert.Equal(t, models.MediaTypeAnime, results[0].MediaType)
	assert.Contains(t, results[0].URL, "/play/naruto.abc123")
	assert.Equal(t, "https://img.example.com/naruto.jpg", results[0].ImageURL)
}

func TestAnimeWorldSearchAnime_Empty(t *testing.T) {
	t.Parallel()

	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `<html><body><div class="film-list"></div></body></html>`)
	}))
	defer srv.Close()

	client := NewAnimeWorldClient()
	client.baseURL = srv.URL
	client.maxRetries = 0
	client.retryDelay = 0

	results, err := client.SearchAnime("nothing-matches")
	require.NoError(t, err)
	assert.Empty(t, results)
}

func TestAnimeWorldSearchAnime_RetriesGiveUp(t *testing.T) {
	t.Parallel()

	var calls int32
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		atomic.AddInt32(&calls, 1)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer srv.Close()

	client := NewAnimeWorldClient()
	client.baseURL = srv.URL
	client.maxRetries = 2
	client.retryDelay = 0

	_, err := client.SearchAnime("naruto")
	require.Error(t, err)
	assert.Equal(t, int32(3), atomic.LoadInt32(&calls), "should attempt maxRetries+1 times")
}

// --- normalizeEpisodes ---

func TestAnimeWorld_normalizeEpisodes(t *testing.T) {
	client := NewAnimeWorldClient()

	raws := []rawEpisode{
		{numberStr: "1", number: 1, url: "/ep1"},
		{numberStr: "2", number: 2, url: "/ep2"},
		{numberStr: "3-4", number: 3, url: "/ep34"},
		{numberStr: "5-6", number: 4, url: "/ep56"},
		{numberStr: "7", number: 5, url: "/ep7"},
		{numberStr: "8-9", number: 6, url: "/ep89"},
	}
	expected := []models.Episode{
		{Number: "Episodio 1", Num: 1, URL: client.baseURL + "/ep1"},
		{Number: "Episodio 2", Num: 2, URL: client.baseURL + "/ep2"},
		{Number: "Episodio 3-4", Num: 3, URL: client.baseURL + "/ep34"},
		{Number: "Episodio 5-6", Num: 5, URL: client.baseURL + "/ep56"},
		{Number: "Episodio 7", Num: 7, URL: client.baseURL + "/ep7"},
		{Number: "Episodio 8-9", Num: 8, URL: client.baseURL + "/ep89"},
	}

	assert.Equal(t, expected, client.normalizeEpisodes(raws))
}

// --- GetAnimeEpisodes ---

func TestAnimeWorldGetAnimeEpisodes(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/play/naruto-ita1", func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, err := fmt.Fprint(w, animeWorldEpisodesFixture)
		require.NoError(t, err)
	})

	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewAnimeWorldClient()
	client.baseURL = srv.URL
	client.maxRetries = 0
	client.retryDelay = 0

	animeURL := fmt.Sprintf("%s/play/naruto-ita1", client.baseURL)

	eps, err := client.GetAnimeEpisodes(animeURL)
	assert.NoError(t, err)
	assert.Len(t, eps, 6)
	for i, ep := range eps {
		assert.Equal(t, i+1, ep.Num)
		assert.Equal(t, fmt.Sprintf("Episodio %d", i+1), ep.Number)
		assert.Equal(t, fmt.Sprintf("%s/play/naruto-ita%d", client.baseURL, ep.Num), ep.URL)
	}
}

func TestAnimeWorldGetAnimeEpisodes_NoEpisode(t *testing.T) {
	t.Parallel()

	mux := http.NewServeMux()
	mux.HandleFunc("/play/naruto-ita1", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, err := fmt.Fprint(w, "<html><body></body></html>")
		require.NoError(t, err)
	})
	srv := httptest.NewServer(mux)
	defer srv.Close()

	client := NewAnimeWorldClient()
	client.baseURL = srv.URL
	client.maxRetries = 0
	client.retryDelay = 0

	eps, err := client.GetAnimeEpisodes(fmt.Sprintf("%s/play/naruto-ita1", srv.URL))
	assert.NoError(t, err)
	assert.Empty(t, eps)
}

// --- GetStreamURL ---

const (
	animeWorldHellsingEpPath    = "/play/hellsing-ep1"
	animeWorldHellsingDirectMP4 = "https://srv28-greeneyes.sweetpixel.org/DDL/ANIME/HellsingITA/Hellsing_Ep_01_ITA.mp4"
	animeWorldHellsingWrapURL   = "https://srv28-greeneyes.sweetpixel.org/download-file.php?id=DDL/ANIME/HellsingITA/Hellsing_Ep_01_ITA.mp4"
)

// Snippet taken from a real animeworld.ac episode page (Hellsing ep. 1).
const hellsingEpisodePage = `<html><body>
<div class="widget-body" style="padding-top: 17px;padding-bottom: 17px;">
	<center>
		<a href="https://srv28-greeneyes.sweetpixel.org/download-file.php?id=DDL/ANIME/HellsingITA/Hellsing_Ep_01_ITA.mp4" id="downloadLink" class="m-1 btn btn-sm btn-primary" target="_blank">Download Diretto - Ep. 1</a>
		<a href="https://srv28-greeneyes.sweetpixel.org/DDL/ANIME/HellsingITA/Hellsing_Ep_01_ITA.mp4" id="alternativeDownloadLink" class="m-1 btn btn-sm btn-primary" target="_blank" download>Download Alternativo - Ep. 1</a>
		<a href="" id="customDownloadButton" class="m-1 btn btn-sm btn-primary hidden" target="_blank">Download - Ep. 1</a>
	</center>
</div>
</body></html>`

func newAnimeWorldTestClient(t *testing.T, body string) (*AnimeWorldClient, string) {
	t.Helper()

	mux := http.NewServeMux()
	mux.HandleFunc(animeWorldHellsingEpPath, func(w http.ResponseWriter, _ *http.Request) {
		w.Header().Set("Content-Type", "text/html; charset=utf-8")
		_, _ = fmt.Fprint(w, body)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	c := NewAnimeWorldClient()
	c.baseURL = srv.URL
	c.maxRetries = 0
	c.retryDelay = 0

	return c, srv.URL + animeWorldHellsingEpPath
}

// Both links present → alternative (direct .mp4) wins.
func TestAnimeWorldGetStreamURL_PrefersAlternativeLink(t *testing.T) {
	t.Parallel()
	client, epURL := newAnimeWorldTestClient(t, hellsingEpisodePage)

	streamURL, err := client.GetStreamURL(epURL)
	require.NoError(t, err)
	assert.Equal(t, animeWorldHellsingDirectMP4, streamURL)
}

// Only #downloadLink present → wrapper URL is normalized to the direct mp4.
func TestAnimeWorldGetStreamURL_FallsBackToDownloadLink(t *testing.T) {
	t.Parallel()
	body := fmt.Sprintf(`<html><body><center>
		<a href="%s" id="downloadLink">Diretto</a>
	</center></body></html>`, animeWorldHellsingWrapURL)
	client, epURL := newAnimeWorldTestClient(t, body)

	streamURL, err := client.GetStreamURL(epURL)
	require.NoError(t, err)
	assert.Equal(t, animeWorldHellsingDirectMP4, streamURL)
}

// Page has no recognized links.
func TestAnimeWorldGetStreamURL_NoLinks(t *testing.T) {
	t.Parallel()
	client, epURL := newAnimeWorldTestClient(t, `<html><body><center></center></body></html>`)

	_, err := client.GetStreamURL(epURL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no video source found")
}

// download-file.php wrapper without the id param → normalize fails, falls through to error.
func TestAnimeWorldGetStreamURL_DownloadLinkMissingID(t *testing.T) {
	t.Parallel()
	body := `<html><body><center>
		<a href="https://srv28-greeneyes.sweetpixel.org/download-file.php" id="downloadLink">Diretto</a>
	</center></body></html>`
	client, epURL := newAnimeWorldTestClient(t, body)

	_, err := client.GetStreamURL(epURL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no video source found")
}

// Alternative link present but with a junk href → must fall through to downloadLink.
func TestAnimeWorldGetStreamURL_FallsThroughBadAlternative(t *testing.T) {
	t.Parallel()
	body := fmt.Sprintf(`<html><body><center>
		<a href="not-a-url" id="alternativeDownloadLink">Alt</a>
		<a href="%s" id="downloadLink">Diretto</a>
	</center></body></html>`, animeWorldHellsingWrapURL)
	client, epURL := newAnimeWorldTestClient(t, body)

	streamURL, err := client.GetStreamURL(epURL)
	require.NoError(t, err)
	assert.Equal(t, animeWorldHellsingDirectMP4, streamURL)
}

// URL doesn't match /play/<slug> → validation rejects it before any HTTP call.
func TestAnimeWorldGetStreamURL_InvalidEpisodeURL(t *testing.T) {
	t.Parallel()
	client := NewAnimeWorldClient()

	_, err := client.GetStreamURL("https://www.animeworld.ac/not-a-play-path")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "unexpected URL path")
}

// Server replies non-2xx → bubbles up as an error.
func TestAnimeWorldGetStreamURL_ServerError(t *testing.T) {
	t.Parallel()
	mux := http.NewServeMux()
	mux.HandleFunc(animeWorldHellsingEpPath, func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "boom", http.StatusInternalServerError)
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)

	client := NewAnimeWorldClient()
	client.baseURL = srv.URL
	client.maxRetries = 0
	client.retryDelay = 0

	_, err := client.GetStreamURL(srv.URL + animeWorldHellsingEpPath)
	require.Error(t, err)
}

// --- normalizeVideoURL ---

func TestAnimeWorldNormalizationStreamURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name    string
		input   string
		want    string
		wantErr string // substring match; empty = no error
	}{
		{
			name:  "wrapper url is normalized to direct mp4",
			input: "https://srv28-greeneyes.sweetpixel.org/download-file.php?id=DDL/ANIME/HellsingITA/Hellsing_Ep_01_ITA.mp4",
			want:  "https://srv28-greeneyes.sweetpixel.org/DDL/ANIME/HellsingITA/Hellsing_Ep_01_ITA.mp4",
		},
		{
			name:  "wrapper id with leading slash is not doubled",
			input: "https://srv28-greeneyes.sweetpixel.org/download-file.php?id=/DDL/foo.mp4",
			want:  "https://srv28-greeneyes.sweetpixel.org/DDL/foo.mp4",
		},
		{
			name:  "wrapper id is a flat filename",
			input: "https://example.com/download-file.php?id=video.mp4",
			want:  "https://example.com/video.mp4",
		},
		{
			name:  "direct mp4 is returned as-is",
			input: "https://srv28-greeneyes.sweetpixel.org/DDL/ANIME/HellsingITA/Hellsing_Ep_01_ITA.mp4",
			want:  "https://srv28-greeneyes.sweetpixel.org/DDL/ANIME/HellsingITA/Hellsing_Ep_01_ITA.mp4",
		},
		{
			name:    "wrapper without id param",
			input:   "https://srv28-greeneyes.sweetpixel.org/download-file.php",
			wantErr: "missing id",
		},
		{
			name:    "wrapper with empty id",
			input:   "https://srv28-greeneyes.sweetpixel.org/download-file.php?id=",
			wantErr: "missing id",
		},
		{
			name:    "non-mp4 non-wrapper url is rejected",
			input:   "https://example.com/stream/playlist.m3u8",
			wantErr: "unsupported video URL",
		},
		{
			name:    "malformed url is rejected",
			input:   "http://[::1",
			wantErr: "invalid video URL",
		},
	}

	c := NewAnimeWorldClient()
	for _, tt := range tests {
		tt := tt
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got, err := c.normalizeVideoURL(tt.input)
			if tt.wantErr != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), tt.wantErr)
				return
			}
			require.NoError(t, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

// --- API path ---

func TestAnimeWorldStreamURL_FromAPI(t *testing.T) {
	t.Parallel()

	streamURL := "https://www.animeworld.ac/jojo.mp4"
	epID := "TMWIn"
	episodeURL := fmt.Sprintf("https://www.animeworld.ac/play/%s", epID)

	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, epID, r.URL.Query().Get("id"))
		w.Header().Set("Content-Type", "application/json")
		b, err := json.Marshal(animeWorldAPIResponse{
			Grabber: streamURL,
			Name:    "Jojo",
			Target:  "/api/jojo",
		})
		require.NoError(t, err)
		_, err = w.Write(b)
		require.NoError(t, err)
	})
	s := httptest.NewServer(mux)
	defer s.Close()

	client := NewAnimeWorldClient()
	client.episodeAPIURL = s.URL

	respURL, err := client.GetStreamURL(episodeURL)
	assert.NoError(t, err)
	assert.Equal(t, streamURL, respURL)
}
