package providers

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/api/source"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// TestLiveMiruroThroughRegistry drives the real registry the way the app does:
// search fan-out → Resolve → FetchEpisodes → FetchStreamURL, then fetches the
// playlist with the Referer and Origin the provider published, which is what
// mpv will send. Opt in with GOANIME_LIVE=1; never runs in CI.
func TestLiveMiruroThroughRegistry(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	util.InitLogger()
	ctx := context.Background()

	results, err := SearchAll(ctx, "cowboy bebop", source.Miruro)
	if err != nil {
		t.Fatalf("SearchAll(Miruro): %v", err)
	}
	if len(results) == 0 {
		t.Fatal("the Miruro fan-out returned nothing")
	}
	anime := results[0]
	fmt.Printf("fan-out  : %d results, first=%q anilist=%d\n", len(results), anime.Name, anime.AnilistID)
	if anime.Source != "Miruro" {
		t.Errorf("Source = %q, want Miruro", anime.Source)
	}

	src, resolved := source.Resolve(anime)
	if resolved.Kind != source.Miruro {
		t.Fatalf("Resolve sent a Miruro result to %s (%s)", resolved.Kind, resolved.Reason)
	}

	eps, err := src.FetchEpisodes(ctx, anime)
	if err != nil {
		t.Fatalf("FetchEpisodes: %v", err)
	}
	fmt.Printf("episodes : %d\n", len(eps))

	streamURL, err := src.FetchStreamURL(ctx, &eps[0], anime, "best")
	if err != nil {
		t.Fatalf("FetchStreamURL: %v", err)
	}
	referer, origin := util.GetGlobalReferer(), util.GetGlobalOrigin()
	fmt.Printf("stream   : %.86s…\nheaders  : referer=%q origin=%q\n", streamURL, referer, origin)

	req, err := http.NewRequest(http.MethodGet, streamURL, http.NoBody)
	if err != nil {
		t.Fatal(err)
	}
	if referer != "" {
		req.Header.Set("Referer", referer)
	}
	if origin != "" {
		req.Header.Set("Origin", origin)
	}
	resp, err := util.NewFastClient().Do(req) // #nosec G704 -- URL came from the live scraper above
	if err != nil {
		t.Fatalf("playlist fetch: %v", err)
	}
	body, _ := io.ReadAll(io.LimitReader(resp.Body, 1<<20))
	_ = resp.Body.Close()
	if resp.StatusCode != http.StatusOK || !strings.HasPrefix(strings.TrimSpace(string(body)), "#EXTM3U") {
		t.Fatalf("the playlist answered %d (%.60q); playback would fail", resp.StatusCode, body)
	}
	fmt.Printf("playlist : %d OK, %d bytes\n", resp.StatusCode, len(body))

	if subs := util.GetGlobalSubtitles(); len(subs) > 0 {
		labels := make([]string, len(subs))
		for i, s := range subs {
			labels[i] = s.Language + ": " + s.Label
		}
		fmt.Printf("subtitles: %s\n", strings.Join(labels, " | "))
	}
}
