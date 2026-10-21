package topcine

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/alvarorichard/Goanime/internal/util/jsonx"
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

// TestLiveTopCineTodoPoderosoDubIsGated is the canary for the Todo Poderoso
// report. TopCine offers the movie Dublado, but only on 1take.top, which
// answers every request with a Cloudflare Turnstile — so GoAnime plays it
// through the panel, whose dub servers are dead, and asks before falling back
// to the original audio. The day this fails, the dub is reachable over plain
// HTTP and GoAnime should play it straight from TopCine's player.
func TestLiveTopCineTodoPoderosoDubIsGated(t *testing.T) {
	skipUnlessLive(t)
	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()
	c := NewClient()

	title, err := c.Title(ctx, c.BaseURL()+"/filme/todo-poderoso")
	if err != nil {
		t.Fatalf("title: %v", err)
	}
	if title.Kind != KindMovie || title.TMDBID != 310 {
		t.Fatalf("title = %s tmdb %d, want the movie with TMDB 310", title.Kind, title.TMDBID)
	}

	u, _ := playerURL(title.PlayerURL)
	q := u.Query()
	q.Set("acao", "carregar_idiomas")
	u.RawQuery = q.Encode()
	body, err := c.get(ctx, u.String(), getOptions{referer: title.PlayerURL, ajax: true, layer: "player"})
	if err != nil {
		t.Fatalf("languages: %v", err)
	}
	var rec languagesRecord
	if err := jsonx.Unmarshal(body, &rec); err != nil || !rec.HasDubbed || rec.DubbedURL == nil {
		t.Fatalf("the player no longer offers the dub: %s (%v)", body, err)
	}
	t.Logf("dub: %s", *rec.DubbedURL)

	// Gated either way: the TLS-impersonating client gets a 403 challenge, a
	// plain one the Turnstile page itself (both seen 2026-10-21).
	page, err := c.get(ctx, *rec.DubbedURL, getOptions{referer: title.PlayerURL, layer: "player"})
	var diag *netx.SourceDiagnostic
	if errors.As(err, &diag) && diag.Kind == netx.DiagnosticBlockedChallenge {
		t.Logf("dub embed gated: %v", err)
		return
	}
	if err != nil {
		t.Fatalf("dub embed: %v", err)
	}
	lower := strings.ToLower(string(page))
	if !strings.Contains(lower, "turnstile") && !strings.Contains(lower, "challenges.cloudflare.com") {
		t.Errorf("%s answers without a browser check now: GoAnime could play TopCine's Dublado directly (api/topcine.go)", *rec.DubbedURL)
	}
}
