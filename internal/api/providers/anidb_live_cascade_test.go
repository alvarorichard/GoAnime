package providers

import (
	"context"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/util"
)

// TestLiveAniDBThroughRegistry drives the real registry the way the app does:
// search fan-out → Resolve → FetchEpisodes → FetchStreamURL. Opt in with
// GOANIME_LIVE=1; never runs in CI.
func TestLiveAniDBThroughRegistry(t *testing.T) {
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
	util.InitLogger()
	ctx := context.Background()

	results, err := SearchAll(ctx, "cowboy bebop", source.AniDB)
	if err != nil {
		t.Fatalf("SearchAll(AniDB): %v", err)
	}
	if len(results) == 0 {
		t.Fatal("the AniDB fan-out returned nothing")
	}
	fmt.Printf("fan-out  : %d results, first=%q\n", len(results), results[0].Name)

	anime := results[0]
	if anime.Source != "AniDB" {
		t.Errorf("Source = %q, want AniDB", anime.Source)
	}

	src, resolved := source.Resolve(anime)
	if resolved.Kind != source.AniDB {
		t.Fatalf("Resolve sent a tagged AniDB result to %s (%s)", resolved.Kind, resolved.Reason)
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

	_ = models.Anime{}
}
