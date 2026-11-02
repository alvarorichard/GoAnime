//go:build live

package pobreflix_test

import (
	"context"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/cfsolver"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/pobreflix"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// TestLivePobreflixPlayerThroughCfsolve clears the real player gate for Todo
// Poderoso (TMDB 310) with cfsolve and resolves the dubbed proxy URL. It skips
// when cfsolve is not installed (GOANIME_CFSOLVE_BIN / PATH).
func TestLivePobreflixPlayerThroughCfsolve(t *testing.T) {
	if !cfsolver.Register() {
		t.Skip("cfsolve is not installed; set GOANIME_CFSOLVE_BIN")
	}
	t.Cleanup(func() { netx.RegisterChallengeSolver(nil) })
	util.InitLogger()

	ctx, cancel := context.WithTimeout(context.Background(), 90*time.Second)
	defer cancel()

	c := pobreflix.NewClient()
	stream, err := c.ResolvePlayer(ctx, "", pobreflix.PlayerTarget{
		TMDBID: 310,
		Kind:   pobreflix.KindMovie,
	}, pobreflix.AudioDubbed)
	if err != nil {
		t.Fatalf("resolve player: %v", err)
	}
	if stream.URL == "" || stream.Referer == "" {
		t.Fatalf("incomplete stream: %+v", stream)
	}
	if stream.Audio != pobreflix.AudioDubbed {
		t.Errorf("audio = %q, want the dubbed track", stream.Audio)
	}
	t.Logf("stream: %.100s (referer %.60s)", stream.URL, stream.Referer)

	// The proxy serves the file only with the proof-bearing page as Referer;
	// check the first bytes so the test fails on a URL that mpv cannot open.
	req, err := http.NewRequest(http.MethodGet, stream.URL, http.NoBody)
	if err != nil {
		t.Fatalf("build media request: %v", err)
	}
	req.Header.Set("Referer", stream.Referer)
	req.Header.Set("Range", "bytes=0-1023")
	resp, err := (&http.Client{Timeout: 30 * time.Second}).Do(req)
	if err != nil {
		t.Fatalf("media request: %v", err)
	}
	defer func() { _ = resp.Body.Close() }()
	first := make([]byte, 64)
	n, _ := resp.Body.Read(first)
	if resp.StatusCode/100 != 2 {
		t.Fatalf("media answered HTTP %d", resp.StatusCode)
	}
	if ct := resp.Header.Get("Content-Type"); !strings.Contains(ct, "video/mp4") {
		t.Errorf("media content type %q, want video/mp4", ct)
	}
	t.Logf("media: HTTP %d, content-type %q, head %.30q", resp.StatusCode, resp.Header.Get("Content-Type"), first[:n])
}
