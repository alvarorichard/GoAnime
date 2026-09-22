package hianime

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
)

// TestLiveHiAnimeChain walks search → episodes → stream against the real site,
// and then fetches the playlist the way the player will, because the CDN's
// Referer check is the one failure a parser test cannot see.
//
// "best" returns the master, which the CDN serves either way; "720p" returns the
// variant, which is the level where it was measured to enforce the Referer. Both
// are checked so the stricter one is actually covered.
// Opt in with GOANIME_LIVE=1; never runs in CI.
func TestLiveHiAnimeChain(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	util.InitLogger()
	c := NewHiAnimeClient()

	results, err := c.SearchAnime(context.Background(), "jojo")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	if len(results) == 0 {
		t.Fatal("search returned no results")
	}
	fmt.Printf("search   : %d results\n", len(results))
	for i, r := range results[:min(3, len(results))] {
		fmt.Printf("           [%d] %-52s %s\n", i, truncate(r.Name, 52), r.URL)
	}

	target := results[0]
	for _, r := range results {
		if strings.Contains(strings.ToLower(r.Name), "golden wind") {
			target = r
			break
		}
	}

	eps, err := c.GetAnimeEpisodes(context.Background(), target.URL)
	if err != nil {
		t.Fatalf("episodes for %s: %v", target.URL, err)
	}
	fmt.Printf("episodes : %d for %q (first=%s last=%s)\n",
		len(eps), truncate(target.Name, 40), eps[0].Number, eps[len(eps)-1].Number)

	for _, q := range []string{"best", "720p"} {
		streamURL, meta, err := c.GetEpisodeStreamURL(context.Background(), eps[0].URL, q)
		if err != nil {
			t.Errorf("stream (%s): %v", q, err)
			continue
		}
		if !strings.Contains(streamURL, ".m3u8") {
			t.Errorf("stream (%s): not an m3u8: %s", q, streamURL)
			continue
		}
		fmt.Printf("stream %-5s: %s\n           meta=%v\n", q, truncate(streamURL, 88), meta)

		if status := fetchStatus(t, streamURL, meta["referer"]); status != http.StatusOK {
			t.Errorf("stream (%s): playlist answered %d with referer %q — playback would fail",
				q, status, meta["referer"])
		}
	}
}

// fetchStatus GETs a URL with the referer the scraper handed out.
func fetchStatus(t *testing.T, rawURL, referer string) int {
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
	return resp.StatusCode
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
