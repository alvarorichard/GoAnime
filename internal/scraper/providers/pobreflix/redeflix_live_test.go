//go:build live

package pobreflix_test

import (
	"context"
	"io"
	"net/http"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pobreflix"
)

// TestLivePobreflixRedeflixPlaysWithoutASolver resolves Todo Poderoso (TMDB
// 310) through the player the catalog itself lists, with no solver registered,
// and downloads the first bytes of the file it points at.
func TestLivePobreflixRedeflixPlaysWithoutASolver(t *testing.T) {
	netx.RegisterChallengeSolver(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	c := pobreflix.NewClient()
	title, err := c.Title(ctx, c.BaseURL()+"/filme/todo-poderoso")
	if err != nil {
		t.Fatalf("title page: %v", err)
	}
	gated := pobreflix.GatedServers(title.Servers)
	if len(gated) == 0 {
		t.Skip("the title lists no player page right now")
	}
	t.Logf("catalog host now: %s; player page on %s", c.BaseURL(), gated[0].URL)
	if !c.HasSolverFreePlayer(title.Servers) {
		t.Skipf("the listed player is not a RedeFlixApi host: %s", gated[0].URL)
	}

	s, err := c.ResolveRedeflix(ctx, gated[0].URL, pobreflix.AudioDubbed)
	if err != nil {
		t.Fatalf("resolve: %v", err)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, http.NoBody)
	req.Header.Set("Range", "bytes=0-65535")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("media request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	n, _ := io.Copy(io.Discard, resp.Body)
	t.Logf("media: HTTP %d, %d bytes, %s, audio %q", resp.StatusCode, n, resp.Header.Get("Content-Type"), s.Audio)
	if resp.StatusCode != http.StatusPartialContent || n != 65536 {
		t.Fatalf("expected 64 KiB of partial content, got HTTP %d and %d bytes", resp.StatusCode, n)
	}
}
