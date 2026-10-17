package pobreflix

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/util"
)

func skipUnlessLive(t *testing.T) {
	t.Helper()
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
}

// timeoutCtx gives a live test a wall-clock context.
func timeoutCtx(t *testing.T, d time.Duration) context.Context {
	t.Helper()
	ctx, cancel := context.WithTimeout(context.Background(), d)
	t.Cleanup(cancel)
	return ctx
}

// TestLivePobreflixCatalog walks search → title page on the live site. The
// corpus build pulls ~74 sitemap shards (~13 MB) on the first run, then sits
// in the user-cache for the refresh window. Opt in with GOANIME_LIVE=1.
func TestLivePobreflixCatalog(t *testing.T) {
	skipUnlessLive(t)
	util.InitLogger()
	c := NewClient()

	// Titles verified present in the site's own sitemap corpus (probe,
	// 2026-10-16). Section records publish series/filmes only; searches over
	// the corpus prove rank, dedupe and the corpus disk cache.
	for _, q := range []string{"ben hur serie", "todo mundo em panico", "crise da meia idade"} {
		results, err := c.Search(timeoutCtx(t, 90*time.Second), q)
		if err != nil {
			t.Fatalf("search %q: %v", q, err)
		}
		if len(results) == 0 {
			t.Errorf("search %q returned nothing — the corpus may have failed to build", q)
			continue
		}
		t.Logf("%q → %d results; first: %s (%s, %s)", q, len(results), results[0].Title, results[0].Kind, results[0].URL)

		title, err := c.Title(timeoutCtx(t, 30*time.Second), results[0].URL)
		if err != nil {
			t.Errorf("title %s: %v", results[0].URL, err)
			continue
		}
		t.Logf("  title: video=%d tmdb=%d kind=%s name=%q seasons=%d",
			title.VideoID, title.TMDBID, title.Kind, title.Name, len(title.Seasons))
		if title.TMDBID <= 0 {
			t.Error("title page carries no TMDB id — the panel key is lost")
		}
		if title.Kind == KindSeries && len(title.Seasons) == 0 {
			t.Errorf("series page %s lists no seasons", results[0].URL)
		}
	}
}

// TestLivePobreflixSearchCorpusReload proves the persisted cache survives a
// client restart: the second client must answer from disk with no network.
func TestLivePobreflixSearchCorpusReload(t *testing.T) {
	skipUnlessLive(t)
	util.InitLogger()
	first := NewClient()
	if _, err := first.Search(timeoutCtx(t, 90*time.Second), "verdade"); err != nil {
		t.Fatalf("first search: %v", err)
	}
	second := NewClient() // same cache path from disk
	results, err := second.Search(timeoutCtx(t, 30*time.Second), "verdade")
	if err != nil || len(results) == 0 {
		t.Fatalf("warm search failed: %v (%d results)", err, len(results))
	}
}
