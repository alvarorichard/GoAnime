package animesdigital

import (
	"context"
	"errors"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestSearchAnimeMergesAudioVariants(t *testing.T) {
	t.Parallel()
	var gotQuery, gotUA string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("s")
		gotUA = r.Header.Get("User-Agent")
		_, _ = w.Write(fixture(t, "search_2026_10_09.html"))
	}))
	defer srv.Close()

	c := NewClientForTest(srv.Client(), srv.URL)
	results, err := c.SearchAnime("  one-piece_ ")
	if err != nil {
		t.Fatal(err)
	}
	if gotQuery != "one piece" || gotUA == "" {
		t.Fatalf("request query=%q ua=%q", gotQuery, gotUA)
	}
	if len(results) != 3 {
		t.Fatalf("got %d results, want 3 merged titles", len(results))
	}

	var onePiece, dubbedOnly bool
	for _, anime := range results {
		variants := ParseVariants(anime.URL)
		switch anime.Name {
		case "One Piece":
			onePiece = variants.Subtitled != "" && variants.Dubbed != "" && !strings.Contains(anime.URL, "one-piece-dublado")
		case "One Piece: A Série Dublado":
			dubbedOnly = variants.Subtitled == "" && strings.Contains(variants.Dubbed, "one-piece-a-serie-2")
		}
	}
	if !onePiece || !dubbedOnly {
		t.Fatalf("merged variants not preserved: %+v", results)
	}
}

func TestGetAnimeEpisodesFollowsPaginationAndSorts(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/anime/a/one-piece":
			_, _ = w.Write(fixture(t, "episodes_page_1_2026_10_09.html"))
		case "/anime/b/one-piece-todos-episodios-06/page/2/":
			_, _ = w.Write(fixture(t, "episodes_page_2_2026_10_09.html"))
		case "/anime/b/one-piece-todos-episodios-06/page/3/":
			_, _ = w.Write(fixture(t, "episodes_page_3_2026_10_09.html"))
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()

	// Fixtures contain the dated production host; point their pagination links
	// back at the test server while retaining the real markup shape.
	c := NewClientForTest(&http.Client{Transport: rewriteHost{target: srv.URL}}, srv.URL)
	episodes, err := c.GetAnimeEpisodes(srv.URL + "/anime/a/one-piece")
	if err != nil {
		t.Fatal(err)
	}
	want := []string{"Episódio 0", "Episódio 1", "Episódio 1.5", "Episódio 2", "Episódio 3"}
	if len(episodes) != len(want) {
		t.Fatalf("got %d episodes, want %d: %+v", len(episodes), len(want), episodes)
	}
	for i := range want {
		if episodes[i].Number != want[i] {
			t.Errorf("episode %d = %q, want %q", i, episodes[i].Number, want[i])
		}
	}
}

type rewriteHost struct{ target string }

func (r rewriteHost) RoundTrip(req *http.Request) (*http.Response, error) {
	out := req.Clone(req.Context())
	target := strings.TrimPrefix(r.target, "http://")
	out.URL.Scheme = "http"
	out.URL.Host = target
	out.Host = target
	return http.DefaultTransport.RoundTrip(out)
}

func TestGetEpisodeStreamURLExtractsPlayerOne(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(fixture(t, "episode_2026_10_09.html"))
	}))
	defer srv.Close()
	c := NewClientForTest(srv.Client(), srv.URL)

	got, err := c.GetEpisodeStreamURL(srv.URL + "/video/a/139026/")
	if err != nil {
		t.Fatal(err)
	}
	want := "https://cdn-s01.mywallpaper-4k-image.net/stream/sv/o/one-piece/1180.mp4/index.m3u8"
	if got != want {
		t.Fatalf("stream = %q, want %q", got, want)
	}
}

func TestGetEpisodeStreamURLExplainsUnsupportedFallback(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><iframe class="metaframe rptss" src="https://animesdigital.org/token/61/bg.mp4?p=2"></iframe></html>`))
	}))
	defer srv.Close()
	c := NewClientForTest(srv.Client(), srv.URL)

	_, err := c.GetEpisodeStreamURL(srv.URL + "/video/a/1/")
	if err == nil || !strings.Contains(err.Error(), "player 2") {
		t.Fatalf("err = %v, want a clear player 2 diagnostic", err)
	}
}

func TestCloudflareChallengeIsNotBypassed(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><head><title>Just a moment...</title></head><body><form id="challenge-form"></form></body></html>`))
	}))
	defer srv.Close()
	c := NewClientForTest(srv.Client(), srv.URL)

	_, err := c.SearchAnimeContext(context.Background(), "naruto")
	if err == nil || !errors.Is(err, netx.ErrSourceUnavailable) || !strings.Contains(err.Error(), "Cloudflare challenge") {
		t.Fatalf("err = %v, want classified Cloudflare failure", err)
	}
}

func TestProductionHostPinned20261009(t *testing.T) {
	t.Parallel()
	if DefaultBase != "https://animesdigital.org" {
		t.Fatalf("AnimesDigital host changed: %s", DefaultBase)
	}
}
