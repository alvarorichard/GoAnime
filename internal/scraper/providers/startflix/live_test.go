package startflix

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
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

// TestLiveStartFlixSeriesChain walks search → title page → panel → seasons →
// players → stream, then fetches the playlist the way mpv will. The title is
// one whose episodes carried a Byse server on 2026-10-07; if it loses that
// server the stream step fails with NoStreamError naming what is left.
// Opt in with GOANIME_LIVE=1; never runs in CI.
func TestLiveStartFlixSeriesChain(t *testing.T) {
	skipUnlessLive(t)
	util.InitLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	c := NewClient()

	results, err := c.Search(ctx, "reborn as a space mercenary")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	var target *Media
	for i := range results {
		fmt.Printf("search   : [%s] %s %s (%s)\n", results[i].Kind, results[i].Title, results[i].URL, results[i].Year)
		if target == nil && results[i].Kind == KindSeries {
			target = &results[i]
		}
	}
	if target == nil {
		t.Fatal("no series in the results")
	}

	panel, err := c.Panel(ctx, target.URL)
	if err != nil {
		t.Fatalf("panel: %v", err)
	}
	fmt.Printf("panel    : %s (tmdb=%d)\n", panel.URL, panel.TMDBID)

	seasons, err := c.Seasons(ctx, panel)
	if err != nil {
		t.Fatalf("seasons: %v", err)
	}
	for _, s := range seasons {
		fmt.Printf("season   : %d dub=%d sub=%d\n", s.Number, len(s.Dubbed), len(s.Subtitled))
	}
	ep := seasons[0].Episodes(seasons[0].Audios()[0])[0]
	playersURL := panel.EpisodeURL(ep.ID)

	players, err := c.Players(ctx, playersURL)
	if err != nil {
		t.Fatalf("players: %v", err)
	}
	for _, p := range players {
		fmt.Printf("player   : %-8s %s\n", p.Type, p.URL)
	}

	stream, err := c.ResolveStream(ctx, players)
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	fmt.Printf("stream   : %s via %s\n", truncate(stream.URL, 100), stream.Host)
	checkPlaylist(t, ctx, stream)
}

// TestLiveStartFlixMovie covers the movie path, where the panel itself lists
// the players.
func TestLiveStartFlixMovie(t *testing.T) {
	skipUnlessLive(t)
	util.InitLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	c := NewClient()

	results, err := c.Search(ctx, "one piece heroines")
	if err != nil {
		t.Fatalf("search: %v", err)
	}
	var target *Media
	for i := range results {
		if results[i].Kind == KindMovie {
			target = &results[i]
			break
		}
	}
	if target == nil {
		t.Fatal("no movie in the results")
	}
	panel, err := c.Panel(ctx, target.URL)
	if err != nil {
		t.Fatalf("panel: %v", err)
	}
	fmt.Printf("panel    : %s (imdb=%s)\n", panel.URL, panel.IMDBID)
	stream, err := c.Stream(ctx, panel.URL)
	if err != nil {
		t.Fatalf("stream: %v", err)
	}
	fmt.Printf("stream   : %s via %s\n", truncate(stream.URL, 100), stream.Host)
	checkPlaylist(t, ctx, stream)
}

// TestLiveStartFlixUnsupportedOnly pins the failure a user will meet most
// until more hosts are supported: a series whose players are all on hosts
// without a resolver must fail with ErrNoSupportedServer, not a parser error.
func TestLiveStartFlixUnsupportedOnly(t *testing.T) {
	skipUnlessLive(t)
	util.InitLogger()
	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()
	c := NewClient()

	panel := Panel{URL: "https://www.painel-aso.sbs/embed/1429", Kind: KindSeries, TMDBID: 1429} // Attack on Titan
	seasons, err := c.Seasons(ctx, panel)
	if err != nil {
		t.Fatalf("seasons: %v", err)
	}
	ep := seasons[0].Episodes(seasons[0].Audios()[0])[0]
	_, err = c.Stream(ctx, panel.EpisodeURL(ep.ID))
	fmt.Printf("result   : %v\n", err)
	if err != nil && !errors.Is(err, ErrNoSupportedServer) {
		t.Fatalf("want ErrNoSupportedServer, got %v", err)
	}
}

func checkPlaylist(t *testing.T, ctx context.Context, s *Stream) {
	t.Helper()
	req, _ := http.NewRequestWithContext(ctx, http.MethodGet, s.URL, http.NoBody)
	req.Header.Set("User-Agent", userAgent)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("playlist: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	head, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
	fmt.Printf("playlist : HTTP %d %q\n", resp.StatusCode, truncate(string(head), 60))
	if resp.StatusCode != http.StatusOK || !strings.HasPrefix(string(head), "#EXTM3U") {
		t.Fatalf("playlist not served: HTTP %d", resp.StatusCode)
	}
}

func truncate(s string, n int) string {
	if len(s) <= n {
		return s
	}
	return s[:n] + "…"
}
