package anidb

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
)

// TestLiveAniDBChain walks search → episodes → stream against the real site.
// Opt in with GOANIME_LIVE=1; never runs in CI.
func TestLiveAniDBChain(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	util.InitLogger()
	c := NewAniDBClient()

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
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n-1] + "…"
}
