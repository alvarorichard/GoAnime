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

// TestLivePobreflixRedeflixPlaysWithoutASolver resolves Interestelar (TMDB
// 157336) through the player the catalog itself lists, with no solver registered,
// and downloads the first bytes of the file it points at.
func TestLivePobreflixRedeflixPlaysWithoutASolver(t *testing.T) {
	netx.RegisterChallengeSolver(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	c := pobreflix.NewClient()
	title, err := c.Title(ctx, c.BaseURL()+"/filmes/online/interestelar-dublado-157336")
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

// TestLivePobreflixSeriesEpisodePlaysWithoutASolver walks a series the way the
// app does (title page, its own episode list, the episode page, the RedeFlixApi
// player) and downloads the first bytes of the episode.
func TestLivePobreflixSeriesEpisodePlaysWithoutASolver(t *testing.T) {
	netx.RegisterChallengeSolver(nil)
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	c := pobreflix.NewClient()
	title, err := c.Title(ctx, c.BaseURL()+"/series/online/dexter-dublado-1405")
	if err != nil {
		t.Fatalf("title page: %v", err)
	}
	page := c.EpisodeURL(title, 1, 2)
	if page == "" {
		t.Fatalf("the title lists no season 1 episode 2 (%d seasons)", len(title.Seasons))
	}
	data, err := c.Episode(ctx, page)
	if err != nil {
		t.Fatalf("episode page %s: %v", page, err)
	}
	gated := pobreflix.GatedServers(data.Servers)
	if len(gated) == 0 || !c.HasSolverFreePlayer(data.Servers) {
		t.Skipf("the episode lists no RedeFlixApi player: %+v", data.Servers)
	}
	s, err := c.ResolveRedeflix(ctx, gated[0].URL, pobreflix.AudioDubbed)
	if err != nil {
		t.Fatalf("resolve %s: %v", gated[0].URL, err)
	}
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, http.NoBody)
	req.Header.Set("Range", "bytes=0-65535")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("media request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	n, _ := io.Copy(io.Discard, resp.Body)
	t.Logf("S%dE%d %q: HTTP %d, %d bytes, %s", data.Season, data.Number, data.Name, resp.StatusCode, n, resp.Header.Get("Content-Type"))
	if resp.StatusCode != http.StatusPartialContent || n != 65536 {
		t.Fatalf("expected 64 KiB of partial content, got HTTP %d and %d bytes", resp.StatusCode, n)
	}
}
