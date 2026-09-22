package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/util"
)

// TestLiveHiAnimeThroughRegistry drives the real registry the way the app does:
// search fan-out → Resolve → FetchEpisodes → FetchStreamURL. Opt in with
// GOANIME_LIVE=1; never runs in CI.
func TestLiveHiAnimeThroughRegistry(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	util.InitLogger()
	ctx := context.Background()

	results, err := SearchAll(ctx, "cowboy bebop", source.HiAnime)
	if err != nil {
		t.Fatalf("SearchAll(HiAnime): %v", err)
	}
	if len(results) == 0 {
		t.Fatal("the HiAnime fan-out returned nothing")
	}
	fmt.Printf("fan-out  : %d results, first=%q\n", len(results), results[0].Name)

	anime := results[0]
	if anime.Source != "HiAnime" {
		t.Errorf("Source = %q, want HiAnime", anime.Source)
	}

	src, resolved := source.Resolve(anime)
	if resolved.Kind != source.HiAnime {
		t.Fatalf("Resolve sent a tagged HiAnime result to %s (%s)", resolved.Kind, resolved.Reason)
	}
	fmt.Printf("resolve  : %s (%s)\n", resolved.Kind, resolved.Reason)

	eps, err := src.FetchEpisodes(ctx, anime)
	if err != nil {
		t.Fatalf("FetchEpisodes: %v", err)
	}
	fmt.Printf("episodes : %d\n", len(eps))

	streamURL, err := src.FetchStreamURL(ctx, &eps[0], anime, "best")
	if err != nil {
		t.Fatalf("FetchStreamURL: %v", err)
	}
	if !strings.Contains(streamURL, ".m3u8") {
		t.Fatalf("not a playlist: %s", streamURL)
	}
	fmt.Printf("stream   : %.86s…\n", streamURL)

	// Resolving a URL is not the same as being able to play it. The CDN checks
	// the Referer, so this is the step that would have caught the metadata the
	// provider used to discard.
	referer := util.GetGlobalReferer()
	fmt.Printf("referer  : %q\n", referer)
	if referer == "" {
		t.Fatal("FetchStreamURL set no referer; the CDN 403s part of its playlists without one")
	}
	body, status := fetchPlaylist(t, streamURL, referer)
	if status != http.StatusOK {
		t.Fatalf("the playlist answered %d with referer %q — playback would fail", status, referer)
	}
	if !strings.HasPrefix(strings.TrimSpace(body), "#EXTM3U") {
		t.Fatalf("not a playlist body: %.80s", body)
	}
	fmt.Printf("playlist : %d OK, %d bytes\n", status, len(body))

	if subs := util.GetGlobalSubtitles(); len(subs) > 0 {
		labels := make([]string, len(subs))
		for i, s := range subs {
			labels[i] = s.Label
		}
		fmt.Printf("subtitles: %s\n", strings.Join(labels, ", "))
	}
}

// fetchPlaylist GETs a playlist the way mpv will, with the referer the provider
// stored.
func fetchPlaylist(t *testing.T, rawURL, referer string) (string, int) {
	t.Helper()
	req, err := http.NewRequest(http.MethodGet, rawURL, http.NoBody)
	if err != nil {
		t.Fatalf("playlist request: %v", err)
	}
	req.Header.Set("Referer", referer)
	resp, err := util.NewFastClient().Do(req) // #nosec G704 -- URL came from the live scraper above
	if err != nil {
		t.Fatalf("playlist fetch: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	return string(body), resp.StatusCode
}
