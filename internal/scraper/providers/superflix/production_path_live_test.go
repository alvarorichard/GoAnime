package superflix

import (
	"context"
	"os"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
)

// TestProductionPathResolvesStream_Live drives the real entry point the player
// uses (GetStreamURL) and prints what it hands mpv, so the browser-gated half
// of the chain can be checked without launching the whole app.
//
// This is the harness that caught the 2026-08-26 Referer defect: GetStreamURL
// was failing with "sniffed a dead stream host" because streamURLDead probed
// the signed URL with "<playerHost>/" and the CDN answers 403 to anything but
// the player's own /video/<hash> page. See playerRefererFor.
//
//	GOANIME_RECON=1 go test ./internal/scraper/providers/superflix/ -run TestProductionPathResolvesStream_Live -v -count=1 -timeout 200s
func TestProductionPathResolvesStream_Live(t *testing.T) {
	if os.Getenv("GOANIME_RECON") == "" {
		t.Skip("set GOANIME_RECON=1 (launches a real browser + hits the live site)")
	}
	skipInCI(t)
	if testing.Short() {
		t.Skip("skipping live browser recon in -short")
	}
	util.InitLogger()

	res, err := NewSuperFlixClient().GetStreamURL(context.Background(), "serie", "76479", "5", "8")
	if err != nil {
		t.Fatalf("GetStreamURL: %v", err)
	}
	if res.StreamURL == "" {
		t.Fatal("resolved with no stream URL")
	}
	// The Referer is not decoration here: with the bare origin the CDN 403s the
	// signed playlist and the flow is discarded as a dead host before mpv runs.
	if res.Referer == "" {
		t.Error("no Referer — mpv will be handed a URL the CDN rejects")
	}
	t.Logf("stream : %s", res.StreamURL)
	t.Logf("referer: %s", res.Referer)
}
