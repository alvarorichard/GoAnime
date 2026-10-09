package topcine

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
)

func skipUnlessLive(t *testing.T) {
	t.Helper()
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
}

// TestLiveTopCineCatalog walks search → title page → player language check on
// the live site. Opt in with GOANIME_LIVE=1; never runs in CI.
func TestLiveTopCineCatalog(t *testing.T) {
	skipUnlessLive(t)
	util.InitLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	c := NewClient()

	results, err := c.Search(ctx, "breaking bad")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	var series *Media
	for i := range results {
		t.Logf("search: [%s] %s %s (%s)", results[i].Kind, results[i].Title, results[i].URL, results[i].Year)
		if series == nil && results[i].Kind == KindSeries && results[i].Title == "Breaking Bad" {
			series = &results[i]
		}
	}
	if series == nil {
		t.Fatal("Breaking Bad is not in the results")
	}

	title, err := c.Title(ctx, series.URL)
	if err != nil {
		t.Fatalf("title: %v", err)
	}
	if title.TMDBID != 1396 || len(title.Seasons) != 5 {
		t.Fatalf("title = tmdb %d with %d seasons, want 1396 with 5", title.TMDBID, len(title.Seasons))
	}

	langs, err := c.Languages(ctx, title.Seasons[0].Episodes[0].PlayerURL)
	if err != nil {
		t.Fatalf("languages: %v", err)
	}
	t.Logf("S01E01 player: dubbed=%v subtitled=%v hosts=%v", langs.Dubbed, langs.Subtitled, langs.Hosts)
	if !langs.Dubbed && !langs.Subtitled {
		t.Error("the player offers no audio track for S01E01")
	}
}
