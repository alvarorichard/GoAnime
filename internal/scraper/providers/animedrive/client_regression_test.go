package animedrive

import (
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Live shapes captured 2026-10-16. Each test reproduces a real failure mode
// seen on animesdrive.cloud so a regression fails here instead of in prod.

// ccdnPlayer serves a Mobile/blogger + FullHD/HLS (ccdn.xyz) page exactly like
// One Piece 1088 live: the HLS <source> carries &#038; entities and the ccdn
// file is gone (404), while the blogger token is live (200 HTML).
func ccdnPlayerServer(t *testing.T, hlsStatus int) *httptest.Server {
	t.Helper()
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasPrefix(r.URL.Path, "/hls/"):
			w.WriteHeader(hlsStatus)
			return
		default:
			_, _ = fmt.Fprintf(w, `
			<div class="animeq-player__source" data-animeq-source="0"><div><iframe src="https://www.blogger.com/video.g?token=abc"></iframe></div></div>
			<div class="animeq-player__source" data-animeq-source="1"><video><source src="%s/hls/playlist.m3u8&#038;e=1&#038;t=2" type="video/mp4"></video></div>
			<button class="animeq-player__server" data-animeq-switch="0" data-source-name="Mobile / Celular">Mobile</button>
			<button class="animeq-player__server" data-animeq-switch="1" data-source-name="FullHD / HLS">HLS</button>
			`, server.URL)
		}
	}))
	t.Cleanup(server.Close)
	return server
}

// bloggerTransport answers blogger embeds like the live site (200 HTML) while
// delegating everything else to the default transport.
type bloggerLiveTransport struct{ fallback http.RoundTripper }

func (t bloggerLiveTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if strings.Contains(req.URL.Host, "blogger.com") {
		return &http.Response{
			StatusCode: http.StatusOK,
			Header:     make(http.Header),
			Body:       io.NopCloser(strings.NewReader("<html>blogger</html>")),
			Request:    req,
		}, nil
	}
	return t.fallback.RoundTrip(req)
}

func TestAnimeDriveRegression_DeadHLSFallsBackToBlogger(t *testing.T) {
	t.Parallel()

	server := ccdnPlayerServer(t, http.StatusNotFound)
	client := NewClientForTest(server.URL)
	client.client = &http.Client{Transport: bloggerLiveTransport{fallback: http.DefaultTransport}}

	stream, meta, err := client.GetStreamURL("/episodio/one-piece-1088/")
	require.NoError(t, err, "dead ccdn HLS must fall back to the live blogger embed")
	assert.Contains(t, stream, "blogger.com/video.g")
	assert.Equal(t, "Mobile / Celular", meta["audio_lang"])
}

func TestAnimeDriveRegression_EntityDecodedHLSSelectedWhenAlive(t *testing.T) {
	t.Parallel()

	server := ccdnPlayerServer(t, http.StatusPartialContent)
	client := NewClientForTest(server.URL)
	client.client = &http.Client{Transport: bloggerLiveTransport{fallback: http.DefaultTransport}}

	stream, meta, err := client.GetStreamURLWithQuality("/episodio/one-piece-1088/", "best")
	require.NoError(t, err)
	assert.Equal(t, server.URL+"/hls/playlist.m3u8&e=1&t=2", stream,
		"&#038; entities must be decoded before probing, or the live HLS 404s")
	assert.Equal(t, "FullHD / HLS", meta["audio_lang"])
	assert.NotContains(t, stream, "&#038;")
}

func TestAnimeDriveRegression_Blogger200CountsAsAvailable(t *testing.T) {
	t.Parallel()

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".mp4") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = fmt.Fprintf(w, `
		<div class="animeq-player__source" data-animeq-source="0"><video><source src="%s/dead.mp4"></video></div>
		<div class="animeq-player__source" data-animeq-source="1"><div><iframe src="https://www.blogger.com/video.g?token=live"></iframe></div></div>
		<button class="animeq-player__server" data-animeq-switch="0" data-source-name="FullHD / HLS">HLS</button>
		<button class="animeq-player__server" data-animeq-switch="1" data-source-name="Mobile / Celular">Mobile</button>
		`, server.URL)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	// Live blogger answers 200 HTML, not 206: the probe must accept it.
	client.client = &http.Client{Transport: bloggerLiveTransport{fallback: http.DefaultTransport}}

	stream, _, err := client.GetStreamURL("/episodio/x/")
	require.NoError(t, err)
	assert.Equal(t, "https://www.blogger.com/video.g?token=live", stream)
}

func TestAnimeDriveRegression_DeadLegendadoDubladoReportsLabels(t *testing.T) {
	t.Parallel()

	// Boruto/Naruto live shape: Legendado+Dublado direct mp4s, both 404, no
	// blogger fallback. The error must name the audio labels for triage.
	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.HasSuffix(r.URL.Path, ".mp4") {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = fmt.Fprintf(w, `
		<div class="animeq-player__source" data-animeq-source="0"><video><source src="%s/legendado.mp4"></video></div>
		<div class="animeq-player__source" data-animeq-source="1"><video><source src="%s/dublado.mp4"></video></div>
		<button class="animeq-player__server" data-animeq-switch="0" data-source-name="Legendado">Legendado</button>
		<button class="animeq-player__server" data-animeq-switch="1" data-source-name="Dublado">Dublado</button>
		`, server.URL, server.URL)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	_, _, err := client.GetStreamURL("/episodio/boruto-episodio-01/")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the 2 servers")
	assert.Contains(t, err.Error(), "Legendado")
	assert.Contains(t, err.Error(), "Dublado")
}

func TestAnimeDriveRegression_EmbedOnlyListsHosts(t *testing.T) {
	t.Parallel()

	// Kimetsu Mugen Jou-hen live shape: 3x FHD third-party iframes, nothing
	// playable. The error must list the embed hosts.
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = fmt.Fprint(w, `
		<div class="animeq-player__source" data-animeq-source="0"><div><iframe src="https://animeshd.cloud/#aaa"></iframe></div></div>
		<div class="animeq-player__source" data-animeq-source="1"><div><iframe src="https://animes.strp2p.com/#bbb"></iframe></div></div>
		<div class="animeq-player__source" data-animeq-source="2"><div><iframe src="https://animes.rpmplay.me/#ccc"></iframe></div></div>
		<button class="animeq-player__server" data-animeq-switch="0" data-source-name="FHD">FHD</button>
		<button class="animeq-player__server" data-animeq-switch="1" data-source-name="FHD">FHD</button>
		<button class="animeq-player__server" data-animeq-switch="2" data-source-name="FHD">FHD</button>
		`)
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	_, _, err := client.GetStreamURL("/episodio/kimetsu-movie/")
	require.Error(t, err)
	assert.Contains(t, err.Error(), "no video tracks found in animeq-player")
	assert.Contains(t, err.Error(), "animeshd.cloud")
	assert.Contains(t, err.Error(), "animes.strp2p.com")
	assert.Contains(t, err.Error(), "animes.rpmplay.me")
}

func TestAnimeDriveRegression_EmbedHostHelper(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "animeshd.cloud", animeDriveEmbedHost("https://animeshd.cloud/#aaa"))
	assert.Equal(t, "animes.strp2p.com", animeDriveEmbedHost("https://animes.strp2p.com/#bbb"))
	assert.Equal(t, "", animeDriveEmbedHost("://broken"))
	assert.Equal(t, "", animeDriveEmbedHost(""))
}

func TestAnimeDriveRegression_DedupStrings(t *testing.T) {
	t.Parallel()

	assert.Equal(t, []string{"a", "b"}, dedupStrings([]string{"a", "b", "a"}))
	assert.Empty(t, dedupStrings(nil))
}

// TestAnimeDriveRegression_ClientCascade drives search → details → episodes →
// stream through the REAL client against one stub, proving the stages stay
// wired together for both a recoverable (blogger fallback) and an unrecoverable
// (dead-only) title.
func TestAnimeDriveRegression_ClientCascade(t *testing.T) {
	t.Parallel()

	var server *httptest.Server
	server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch {
		case strings.HasSuffix(r.URL.Path, ".mp4"):
			if strings.Contains(r.URL.Path, "dead") {
				w.WriteHeader(http.StatusNotFound)
				return
			}
			w.WriteHeader(http.StatusPartialContent)
			return
		case r.URL.Path == "/" && r.URL.Query().Has("s"):
			_, _ = fmt.Fprint(w, `<html><body>
				<div class="result-item"><article><div class="title"><a href="/anime/recoverable/">Recoverable</a></div><img src="/r.jpg"/></article></div>
				<div class="result-item"><article><div class="title"><a href="/anime/deadonly/">Dead Only</a></div><img src="/d.jpg"/></article></div>
			</body></html>`)
			return
		case r.URL.Path == "/anime/recoverable/":
			_, _ = fmt.Fprint(w, `<html><body>
				<h1 class="entry-title">Recoverable</h1>
				<ul class="episodios"><li><a href="/episodio-recoverable-1/">Episode 1</a></li></ul>
			</body></html>`)
			return
		case r.URL.Path == "/anime/deadonly/":
			_, _ = fmt.Fprint(w, `<html><body>
				<h1 class="entry-title">Dead Only</h1>
				<ul class="episodios"><li><a href="/episodio-deadonly-1/">Episode 1</a></li></ul>
			</body></html>`)
			return
		case r.URL.Path == "/episodio-recoverable-1/":
			_, _ = fmt.Fprintf(w, `
			<div class="animeq-player__source" data-animeq-source="0"><video><source src="%s/dead.mp4"></video></div>
			<div class="animeq-player__source" data-animeq-source="1"><div><iframe src="https://www.blogger.com/video.g?token=live"></iframe></div></div>
			<button class="animeq-player__server" data-animeq-switch="0" data-source-name="FullHD / HLS">HLS</button>
			<button class="animeq-player__server" data-animeq-switch="1" data-source-name="Mobile / Celular">Mobile</button>
			`, server.URL)
			return
		case r.URL.Path == "/episodio-deadonly-1/":
			_, _ = fmt.Fprintf(w, `
			<div class="animeq-player__source" data-animeq-source="0"><video><source src="%s/dead-a.mp4"></video></div>
			<div class="animeq-player__source" data-animeq-source="1"><video><source src="%s/dead-b.mp4"></video></div>
			<button class="animeq-player__server" data-animeq-switch="0" data-source-name="Legendado">Legendado</button>
			<button class="animeq-player__server" data-animeq-switch="1" data-source-name="Dublado">Dublado</button>
			`, server.URL, server.URL)
			return
		default:
			w.WriteHeader(http.StatusNotFound)
		}
	}))
	defer server.Close()

	client := NewClientForTest(server.URL)
	client.client = &http.Client{Transport: bloggerLiveTransport{fallback: http.DefaultTransport}}

	results, err := client.SearchAnime("recoverable deadonly")
	require.NoError(t, err)
	require.Len(t, results, 2)

	byName := map[string]string{}
	for _, a := range results {
		byName[a.Name] = a.URL
	}
	require.Contains(t, byName, "Recoverable")
	require.Contains(t, byName, "Dead Only")

	// Recoverable title: dead direct + live blogger → blogger wins.
	eps, err := client.GetAnimeEpisodes(byName["Recoverable"])
	require.NoError(t, err)
	require.Len(t, eps, 1)
	stream, _, err := client.GetStreamURL(eps[0].URL)
	require.NoError(t, err)
	assert.Equal(t, "https://www.blogger.com/video.g?token=live", stream)

	// Dead-only title: both probes 404 → actionable error, not empty success.
	eps, err = client.GetAnimeEpisodes(byName["Dead Only"])
	require.NoError(t, err)
	require.Len(t, eps, 1)
	_, _, err = client.GetStreamURL(eps[0].URL)
	require.Error(t, err)
	assert.Contains(t, err.Error(), "none of the 2 servers")
}
