package pobreflix

import (
	"context"
	"os"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/util"
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
// production client asks the site's /api catalog first; the sitemap corpus is
// the browserless fallback. Opt in with GOANIME_LIVE=1.
func TestLivePobreflixCatalog(t *testing.T) {
	skipUnlessLive(t)
	util.InitLogger()
	c := NewClient()

	// Titles verified present, with episodes, in the live catalog (probe, 2026-10-28).
	for _, q := range []string{"stranger things", "todo mundo em panico", "interestelar"} {
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
	first.apiEnabled = false // exercise the sitemap corpus, not the catalog API
	if _, err := first.Search(timeoutCtx(t, 90*time.Second), "verdade"); err != nil {
		t.Fatalf("first search: %v", err)
	}
	second := NewClient() // same cache path from disk
	second.apiEnabled = false
	results, err := second.Search(timeoutCtx(t, 30*time.Second), "verdade")
	if err != nil || len(results) == 0 {
		t.Fatalf("warm search failed: %v (%d results)", err, len(results))
	}
}

// TestLivePobreflixDubIsGated is the canary for the 2026-10-27 change: every
// Pobreflix page lists one "Player Principal" that is a streambetter.lol page
// behind a Cloudflare Turnstile, not a file. It passes while that holds. It
// fails the day a page lists a file again — good news: the direct path takes it
// without a code change, so refresh the notes (CHANGELOG, source docs) and
// retire this canary.
func TestLivePobreflixDubIsGated(t *testing.T) {
	skipUnlessLive(t)
	util.InitLogger()
	c := NewClient()

	title, err := c.Title(timeoutCtx(t, 60*time.Second), "https://pobreflixtv.gay/filme/interestelar.html")
	if err != nil {
		t.Fatalf("title: %v", err)
	}
	if len(title.Servers) == 0 {
		t.Fatal("the page lists no servers at all; the embedded JSON may have changed shape")
	}
	if files := DirectServers(title.Servers); len(files) > 0 {
		t.Fatalf("Pobreflix lists a file again (%s): the dub can play directly; refresh the notes and retire this canary", files[0].URL)
	}
	t.Logf("gated as expected: %v", GatedHosts(title.Servers))
}
