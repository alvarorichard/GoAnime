package anroll

import (
	"context"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name)) // nosemgrep: go_filesystem_rule-fileread -- fixed testdata directory
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestSearchAnimeMergesAudioVariants(t *testing.T) {
	t.Parallel()
	var gotQuery string
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("s")
		_, _ = w.Write(fixture(t, "search_2026_10_12.html")) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter -- test server serving a local fixture
	}))
	defer srv.Close()
	c := NewClientForTest(srv.Client(), srv.URL)
	results, err := c.SearchAnime(" one-piece_")
	if err != nil {
		t.Fatal(err)
	}
	if gotQuery != "one piece" || len(results) != 3 {
		t.Fatalf("query=%q results=%d", gotQuery, len(results))
	}
	variants := ParseVariants(results[0].URL)
	if results[0].Name != "One Piece" || variants.Subtitled == "" || variants.Dubbed == "" {
		t.Fatalf("paired result = %+v variants=%+v", results[0], variants)
	}
	if results[2].MediaType != models.MediaTypeMovie {
		t.Fatalf("movie media type = %v", results[2].MediaType)
	}
}

func TestGetAnimeEpisodesSortsAllItems(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write(fixture(t, "episodes_2026_10_12.html")) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter -- test server serving a local fixture
	}))
	defer srv.Close()
	c := NewClientForTest(srv.Client(), srv.URL)
	episodes, err := c.GetAnimeEpisodes(srv.URL + "/anime/one-piece/")
	if err != nil {
		t.Fatal(err)
	}
	if len(episodes) != 4 || episodes[0].Num != 1 || episodes[3].Num != 1180 {
		t.Fatalf("episodes = %+v", episodes)
	}
}

func TestGetEpisodeStreamDecodesHLSAndSendsIframeHeaders(t *testing.T) {
	t.Parallel()
	var iframeHeaders http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/68023/":
			_, _ = w.Write(fixture(t, "episode_hls_2026_10_12.html")) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter -- test server serving a local fixture
		case "/token/hv0688292sn4his":
			iframeHeaders = r.Header.Clone()
			_, _ = w.Write(fixture(t, "player_hls_2026_10_12.html")) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter -- test server serving a local fixture
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := NewClientForTest(srv.Client(), srv.URL)
	stream, err := c.GetEpisodeStream(srv.URL + "/68023/")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stream.URL, "85b6f01cfbee3c3ae01ea558888e8d7d-master.m3u8") ||
		stream.Referer != playerOrigin+"/" || stream.Origin != playerOrigin {
		t.Fatalf("stream = %+v", stream)
	}
	for key, want := range map[string]string{
		"Referer": "" + srv.URL + "/68023/", "Sec-Fetch-Dest": "iframe", "Sec-Fetch-Mode": "navigate",
		"Sec-Fetch-Site": "cross-site", "Upgrade-Insecure-Requests": "1", "Accept-Language": netxAcceptLanguageForTest(),
	} {
		if got := iframeHeaders.Get(key); got != want {
			t.Errorf("%s=%q, want %q", key, got, want)
		}
	}
}

func netxAcceptLanguageForTest() string { return "pt-BR,pt;q=0.9,en-US;q=0.8,en;q=0.7" }

func TestGetEpisodeStreamSelectsBestMP4FromJSONLDPlayer(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch r.URL.Path {
		case "/5831/":
			_, _ = w.Write(fixture(t, "episode_mp4_2026_10_12.html")) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter -- test server serving a local fixture
		case "/token/oHajlLyOvRJDcl8":
			_, _ = w.Write(fixture(t, "player_mp4_2026_10_12.html")) // nosemgrep: go.lang.security.audit.xss.no-direct-write-to-responsewriter.no-direct-write-to-responsewriter -- test server serving a local fixture
		default:
			http.NotFound(w, r)
		}
	}))
	defer srv.Close()
	c := NewClientForTest(srv.Client(), srv.URL)
	stream, err := c.GetEpisodeStream(srv.URL + "/5831/")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(stream.URL, "itag=22") || stream.Referer != "" || stream.Origin != "" {
		t.Fatalf("stream = %+v", stream)
	}
}

func TestDecodeScriptRejectsInvalidOrder(t *testing.T) {
	t.Parallel()
	_, err := decodeScriptLayers(`(function(){function x(){}x(["YQ=="],[2],"YQ==");})();`)
	if err == nil || !strings.Contains(err.Error(), "out of range") {
		t.Fatalf("err = %v", err)
	}
}

func TestLiveAnrollChain(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	c := NewClient()
	results, err := c.SearchAnimeContext(ctx, "one piece")
	if err != nil || len(results) == 0 {
		t.Fatalf("search: results=%d err=%v", len(results), err)
	}
	var variants Variants
	for _, result := range results {
		if result.Name == "One Piece" {
			variants = ParseVariants(result.URL)
		}
	}
	if variants.Subtitled == "" || variants.Dubbed == "" {
		t.Fatal("paired One Piece result not found")
	}
	episodes, err := c.GetAnimeEpisodesContext(ctx, variants.Subtitled)
	if err != nil || len(episodes) < 1000 {
		t.Fatalf("episodes: results=%d err=%v", len(episodes), err)
	}
	stream, err := c.GetEpisodeStreamContext(ctx, episodes[len(episodes)-1].URL)
	if err != nil || !strings.Contains(stream.URL, ".m3u8") {
		t.Fatalf("stream=%+v err=%v", stream, err)
	}
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, stream.URL, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	req.Header.Set("User-Agent", netx.UserAgent)
	req.Header.Set("Referer", stream.Referer)
	req.Header.Set("Origin", stream.Origin)
	resp, err := util.NewFastClient().Do(req) // #nosec G704 -- URL came from the live scraper.
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("playlist returned HTTP %d", resp.StatusCode)
	}

	dubbed, err := c.GetAnimeEpisodesContext(ctx, variants.Dubbed)
	if err != nil || len(dubbed) == 0 {
		t.Fatalf("dubbed episodes: results=%d err=%v", len(dubbed), err)
	}
	mp4, err := c.GetEpisodeStreamContext(ctx, dubbed[0].URL)
	if err != nil || !strings.Contains(mp4.URL, "googlevideo.com/") || !strings.Contains(mp4.URL, "itag=22") {
		t.Fatalf("dubbed MP4=%+v err=%v", mp4, err)
	}
}
