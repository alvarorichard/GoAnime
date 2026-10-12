package animesdigital

import (
	"context"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// TestLiveAnimesDigitalChain walks search -> episodes -> stream and verifies
// the extracted playlist. Opt in with GOANIME_LIVE=1; never runs in CI.
func TestLiveAnimesDigitalChain(t *testing.T) {
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
	var target *models.Anime
	for _, result := range results {
		if result.Name == "One Piece" {
			target = result
			break
		}
	}
	if target == nil {
		t.Fatal("exact One Piece result not found")
	}
	variants := ParseVariants(target.URL)
	if variants.Subtitled == "" || variants.Dubbed == "" {
		t.Fatalf("One Piece variants were not paired: %+v", variants)
	}
	titleURL := variants.Subtitled
	episodes, err := c.GetAnimeEpisodesContext(ctx, titleURL)
	if err != nil || len(episodes) < 1000 {
		t.Fatalf("episodes: results=%d err=%v", len(episodes), err)
	}
	if episodes[0].Num > episodes[len(episodes)-1].Num {
		t.Fatalf("episodes are not ascending: first=%s last=%s", episodes[0].Number, episodes[len(episodes)-1].Number)
	}
	streamURL, err := c.GetEpisodeStreamURLContext(ctx, episodes[len(episodes)-1].URL)
	if err != nil || !strings.Contains(streamURL, ".m3u8") {
		t.Fatalf("stream=%q err=%v", streamURL, err)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, streamURL, http.NoBody)
	resp, err := util.NewFastClient().Do(req) // #nosec G704 -- URL came from the live scraper.
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		t.Fatalf("playlist returned HTTP %d", resp.StatusCode)
	}
}
