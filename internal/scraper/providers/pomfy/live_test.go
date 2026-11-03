package pomfy

import (
	"context"
	"os"
	"strings"
	"testing"
	"time"
)

// Live proofs for the whole umbrella: catalog, seasons and one full
// captcha→PoW→playback→decrypt chain against pomfy.my and the byse CDN.
// Opt in with GOANIME_LIVE=1; it never runs in CI.
func requireLive(t *testing.T) {
	t.Helper()
	if os.Getenv("GOANIME_LIVE") == "" || testing.Short() || os.Getenv("CI") != "" {
		t.Skip("set GOANIME_LIVE=1 to run against the live network")
	}
}

func TestLiveByseChainEndToEnd(t *testing.T) {
	requireLive(t)
	c := Shared()
	ctx, cancel := context.WithTimeout(context.Background(), 3*time.Minute)
	defer cancel()

	// Movie 550 (Clube da Luta): the chain every piece of this package serves.
	stream, err := c.ResolveStream(ctx, 550, KindMovie, 0, 0, DefaultBase+"/assistir/550")
	if err != nil {
		t.Fatalf("movie 550: %v", err)
	}
	if !strings.HasPrefix(stream.URL, "https://") || !strings.Contains(stream.URL, "master.m3u8") {
		t.Errorf("stream = %q, want an HLS master", stream.URL)
	}
	if stream.Height <= 0 {
		t.Errorf("height = %d", stream.Height)
	}

	// Series: S2E1 of Breaking Bad (1396), whose master spells the audio.
	stream, err = c.ResolveStream(ctx, 1396, KindSeries, 2, 1, DefaultBase+"/serie/1396")
	if err != nil {
		t.Fatalf("series S2E1: %v", err)
	}
	if !strings.Contains(stream.URL, "master.m3u8") {
		t.Errorf("series stream = %q, want an HLS master", stream.URL)
	}
}
