package player

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/providers/superflix"
	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSuperFlixPlaysInMPV_Live is the end-to-end guard: it resolves a real
// stream, builds the mpv arguments with the SAME buildPlaybackArgs the app
// uses, runs mpv, and asserts frames actually came out.
//
// This is the one test that would have caught the 2026-08-26 outage on its own.
// Every layer beneath it was green while nothing played, because each failure
// was silent:
//
//   - LooksLikeHLS did not recognise "master.txt", so IsHLS was false and every
//     HLS-only argument was skipped;
//   - two of the four Referer producers handed mpv the bare origin;
//   - the CORS Origin header was missing, so the playlist loaded and every
//     segment 403'd.
//
// Any one of those returns mpv to producing nothing, and only running it shows
// that. Measured on the fix: 0 failed segments, ~12s of vp9/opus decoded.
//
// Live + opt-in, needs mpv and the network:
//
//	GOANIME_RECON=1 go test ./internal/player/ -run TestSuperFlixPlaysInMPV_Live -v -count=1 -timeout 400s
func TestSuperFlixPlaysInMPV_Live(t *testing.T) {
	if os.Getenv("GOANIME_RECON") == "" {
		t.Skip("set GOANIME_RECON=1 (hits the live site, may launch a browser, runs mpv)")
	}
	if testing.Short() {
		t.Skip("skipping live playback in -short")
	}
	if v := strings.ToLower(os.Getenv("CI")); v == "true" || v == "1" {
		t.Skip("skipped in CI: real upstream + mpv")
	}
	mpvBin, err := exec.LookPath("mpv")
	if err != nil {
		t.Skip("mpv not installed")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 300*time.Second)
	defer cancel()

	res, err := superflix.NewSuperFlixClient().GetStreamURL(ctx, "filme", "603", "", "")
	if err != nil {
		t.Skipf("SuperFlix unavailable (site may have rotated): %v", err)
	}
	require.NotEmpty(t, res.StreamURL)

	// Go through the real globals so the wiring in playVideo is exercised too,
	// not just the pure argument builder.
	restore := snapshotGlobalReferer()
	t.Cleanup(restore)
	util.SetGlobalReferer(res.Referer)
	// The CDN binds the signed URL to the UA that obtained it, so playback needs
	// the same handoff enhanced.go does — mpv's own libmpv UA is rejected.
	util.SetGlobalUserAgent(res.UserAgent)
	t.Cleanup(util.ClearGlobalUserAgent)
	prevSource := util.GetGlobalAnimeSource()
	util.SetGlobalAnimeSource("SuperFlix")
	t.Cleanup(func() { util.SetGlobalAnimeSource(prevSource) })

	args := buildPlaybackArgs(playbackArgsInput{
		VideoURL:    res.StreamURL,
		IsHLS:       LooksLikeHLS(res.StreamURL),
		IsSuperFlix: util.IsSuperFlixSource(),
	})
	joined := strings.Join(args, " ")
	// Fail loudly here rather than after a two-minute mpv run.
	require.Contains(t, joined, hlsForceLavfFormatArg, "master.txt must be forced through the lavf hls demuxer")
	require.Contains(t, joined, hlsAllowAllExtensionsArg, "disguised segment extensions must be allowed")
	require.Contains(t, joined, "Origin: ", "segments need the CORS Origin header")
	require.Contains(t, joined, "--user-agent="+res.UserAgent,
		"the CDN only serves the signed URL to the UA that obtained it")
	require.Contains(t, joined, "Accept-Language: en-US,en;q=0.9",
		"the CDN matches Chromium's Accept-Language byte for byte")

	out := filepath.Join(t.TempDir(), "out.mkv")
	// buildPlaybackArgs returns the flags only; StartVideo appends the URL last.
	full := append([]string{"--no-config", "--length=8", "--o=" + out}, args...)
	// --o= puts mpv in encoding mode, where the output drivers must be lavc:
	// the app's own --vo=gpu (and a plain --vo=null) fail to initialise here and
	// mpv writes nothing even though it decoded fine.
	full = append(full, "--vo=lavc", "--ao=lavc", res.StreamURL)
	cmd := exec.CommandContext(ctx, mpvBin, full...)
	combined, _ := cmd.CombinedOutput()
	log := string(combined)

	failed := strings.Count(log, "Failed to open segment")
	assert.Zero(t, failed, "mpv could not fetch %d segments — the playback headers no longer satisfy the CDN:\n%s",
		failed, tailLines(log, 12))

	fi, statErr := os.Stat(out)
	require.NoError(t, statErr, "mpv produced no output file at all:\n%s", tailLines(log, 12))
	assert.Greater(t, fi.Size(), int64(100_000),
		"mpv produced only %d bytes — it started but decoded nothing", fi.Size())
}

func tailLines(s string, n int) string {
	lines := strings.Split(strings.TrimRight(s, "\n"), "\n")
	if len(lines) > n {
		lines = lines[len(lines)-n:]
	}
	return strings.Join(lines, "\n")
}
