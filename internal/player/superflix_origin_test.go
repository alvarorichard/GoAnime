package player

import (
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestSuperFlixPlaybackSendsCORSOrigin is the regression guard for the
// 2026-08-26 "SuperFlix resolves but nothing plays" defect.
//
// SuperFlix's FirePlayer serves the HLS segments from rotating third-party
// hosts and validates the CORS origin on them. hls.js fetches segments as a
// cross-origin XHR, so the browser attaches Origin; mpv did not, and every
// segment came back 403 while the playlist itself loaded fine. Measured on the
// same live stream: 121 failed segments without the header, 0 with it.
func TestSuperFlixPlaybackSendsCORSOrigin(t *testing.T) {
	const referer = "https://xn--tckasiu6cvova0eb5fua2449g98vg.best/video/6bc24fc1ab650b25b4114e93a98f1eba"
	const origin = "https://xn--tckasiu6cvova0eb5fua2449g98vg.best"

	restore := snapshotGlobalReferer()
	t.Cleanup(restore)
	util.SetGlobalReferer(referer)

	args, got := appendPlaybackRefererArgs(nil, "https://cdn.example/master.txt", true, true)
	require.Equal(t, referer, got)
	require.Len(t, args, 1)

	assert.Contains(t, args[0], "Referer: "+referer)
	assert.Contains(t, args[0], "Origin: "+origin,
		"without Origin the CDN 403s every segment")
	// Origin is the bare scheme://host — never the /video/<hash> path.
	assert.NotContains(t, args[0], "Origin: "+referer)
}

// Other sources must not start receiving an Origin header they never had.
func TestNonSuperFlixPlaybackOmitsOrigin(t *testing.T) {
	restore := snapshotGlobalReferer()
	t.Cleanup(restore)
	util.SetGlobalReferer("https://animefire.io")

	args, _ := appendPlaybackRefererArgs(nil, "https://cdn.example/master.m3u8", true, false)
	require.Len(t, args, 1)
	assert.Contains(t, args[0], "Referer: https://animefire.io")
	assert.NotContains(t, args[0], "Origin:")
}

func TestCORSOriginOf(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "https://host.example", corsOriginOf("https://host.example/video/abc"))
	assert.Equal(t, "https://host.example", corsOriginOf("https://host.example/"))
	assert.Empty(t, corsOriginOf(""))
	assert.Empty(t, corsOriginOf("not-a-url"))
}

// TestLooksLikeHLSAcceptsMasterTxt pins the detection that gates every
// HLS-only mpv argument. SuperFlix serves its multivariant master as
// "master.txt"; while this returned false the forced lavf hls demuxer and
// allowed_extensions=ALL were both skipped and mpv got a text file with no
// hint it was a playlist.
func TestLooksLikeHLSAcceptsMasterTxt(t *testing.T) {
	t.Parallel()
	assert.True(t, LooksLikeHLS("https://host.best/tok/hash/1787770720/master.txt"))
	assert.True(t, LooksLikeHLS("https://cdn.example/x/master.m3u8"))
	assert.True(t, LooksLikeHLS("https://cdn.example/hls/index"))
	assert.False(t, LooksLikeHLS("https://cdn.example/video.mp4"))
	assert.False(t, LooksLikeHLS(""))
}

// A SuperFlix master.txt must end up with BOTH HLS demuxer flags, not just one.
func TestSuperFlixMasterTxtGetsFullHLSArgs(t *testing.T) {
	restore := snapshotGlobalReferer()
	t.Cleanup(restore)
	util.SetGlobalReferer("https://player.best/video/abc")

	const src = "https://player.best/tok/hash/1787770720/master.txt"
	args := buildPlaybackArgs(playbackArgsInput{
		VideoURL:    src,
		IsHLS:       LooksLikeHLS(src),
		IsSuperFlix: true,
	})
	joined := strings.Join(args, " ")
	assert.Contains(t, joined, hlsAllowAllExtensionsArg)
	assert.Contains(t, joined, hlsForceLavfFormatArg)
	assert.Contains(t, joined, "Origin: https://player.best")
}
