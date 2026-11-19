package animesdigital

import (
	"context"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/models"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
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

// TestLiveAnimesDigitalReportedTitlesPlay walks every episode of the titles
// users reported (issues #229, #236 and #237) and requires that what the client
// returns is something the player can start: for player 1, its playlist and
// first piece of video with the Referer the player sends; otherwise player 2's
// Blogger video. An episode neither player has is allowed only as the
// media-offline error, never as a URL that plays nothing. Opt in with
// GOANIME_LIVE=1; never runs in CI.
func TestLiveAnimesDigitalReportedTitlesPlay(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	titles := map[string]string{
		"Cyberpunk: Edgerunners 2": "/anime/a/cyberpunk-edgerunners-2-2",
		"Cyberpunk: Edgerunners":   "/anime/a/cyberpunk-edgerunners",
		"Wind Breaker":             "/anime/a/wind-breaker",
		"Takagi-san 2":             "/anime/a/karakai-jouzu-no-takagi-san-2",
		// Issue #237. An older title: its episodes are on player 2 only.
		"Bleach": "/anime/a/bleach",
	}
	// Bleach lists 366 episodes; the first few say whether the title plays.
	limits := map[string]int{"Bleach": 8}
	c := NewClient()
	for name, path := range titles {
		t.Run(name, func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
			defer cancel()
			episodes, err := c.GetAnimeEpisodesContext(ctx, DefaultBase+path)
			if err != nil || len(episodes) == 0 {
				t.Fatalf("episodes: %d, err=%v", len(episodes), err)
			}
			if limit := limits[name]; limit > 0 && len(episodes) > limit {
				episodes = episodes[:limit]
			}
			for _, episode := range episodes {
				stream, err := c.GetEpisodeStreamURLContext(ctx, episode.URL)
				if errors.Is(err, netx.ErrMediaOffline) {
					t.Logf("%s: offline on both players (reported, not a failure)", episode.Number)
					continue
				}
				if err != nil {
					t.Errorf("%s: %v", episode.Number, err)
					continue
				}
				if StreamReferer(stream) == "" {
					if !strings.HasPrefix(stream, bloggerVideoPrefix) {
						t.Errorf("%s: %q is neither player 1 nor player 2", episode.Number, stream)
					}
					continue
				}
				if err := c.playlistAnswers(ctx, stream); err != nil {
					t.Errorf("%s: %s would not play: %v", episode.Number, stream, err)
				}
			}
		})
	}
}
