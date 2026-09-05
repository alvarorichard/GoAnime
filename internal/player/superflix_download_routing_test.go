package player

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// superFlixMasterURLs are the real shapes SuperFlix's FirePlayer has served,
// newest first. Every one of them must reach the ffmpeg-HLS downloader.
var superFlixMasterURLs = []struct {
	name string
	url  string
}{
	{
		// Live 2026-09-01: /<token>/<contentid>/<expires>/master.txt.
		// This is the shape that broke downloads — it has no "/cdn/hls/".
		name: "current: token/contentid/expires",
		url:  "https://xn--tckasiu6cvova0eb5fua2449g98vg.best/E5QvcAOtvokp4_b1yDauAw/f6e229b53d19b2257a19223e49a1acac/1788245971/master.txt",
	},
	{
		name: "legacy: /cdn/hls/<hash>/master.txt",
		url:  "https://xn--kcksk7a2bl5le7b6doc1h3f.com/cdn/hls/c8a4e367/master.txt",
	},
	{
		name: "legacy with query",
		url:  "https://host.best/cdn/hls/abc/master.txt?md5=x&expires=1",
	},
}

// TestIsSuperFlixTextHLS_AcceptsEveryServedShape_2026_09_02 is the direct
// regression for "downloads stopped working".
//
// The detector required "/cdn/hls/" AND "master.txt". When the player host
// changed its path to /<token>/<contentid>/<expires>/master.txt the first half
// stopped matching, so every SuperFlix URL missed the ffmpeg-HLS branch and
// fell through the routing switch to the plain-MP4 Range downloader — which
// dutifully saved the playlist TEXT as the episode file. Playback was
// unaffected, because LooksLikeHLS had already been taught about master.txt;
// only the downloader still had the old rule.
func TestIsSuperFlixTextHLS_AcceptsEveryServedShape_2026_09_02(t *testing.T) {
	t.Parallel()
	for _, tc := range superFlixMasterURLs {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Truef(t, isSuperFlixTextHLS(tc.url),
				"a SuperFlix master playlist must route to ffmpeg, not the MP4 downloader:\n  %s", tc.url)
		})
	}
}

func TestIsSuperFlixTextHLS_RejectsNonPlaylists(t *testing.T) {
	t.Parallel()
	for _, u := range []string{
		"",
		"https://cdn.example/video.mp4",
		"https://www.blogger.com/video.g?token=abc",
		"https://lightspeedst.net/s6/mp4/show/hd/20.mp4",
		"https://cdn.example/master.txt.html", // not a playlist, just a page
	} {
		assert.Falsef(t, isSuperFlixTextHLS(u), "%q must not be treated as a SuperFlix playlist", u)
	}
}

// TestDownloadAndPlaybackAgreeOnSuperFlixHLS pins the invariant whose violation
// caused the outage: playback and download must classify a SuperFlix master
// playlist the same way.
//
// They drifted because the knowledge lived in two independent predicates and
// only one was updated. This test fails the moment they disagree again.
func TestDownloadAndPlaybackAgreeOnSuperFlixHLS(t *testing.T) {
	t.Parallel()
	for _, tc := range superFlixMasterURLs {
		require.Truef(t, LooksLikeHLS(tc.url), "playback must see HLS: %s", tc.url)
		assert.Equalf(t, LooksLikeHLS(tc.url), isSuperFlixTextHLS(tc.url),
			"playback and download must agree on %s (%s)", tc.name, tc.url)
	}
}

// routeOf mirrors the classification order of the download switches. It exists
// so the routing decision can be asserted at all: the real switches are three
// inline copies inside long functions, which is exactly why a stale predicate
// went unnoticed.
//
// It is deliberately kept in the test, next to assertions that pin the real
// predicates, so a drift between this mirror and production shows up as a
// failing predicate test rather than silently passing.
func routeOf(videoURL string) string {
	switch {
	case isSuperFlixTextHLS(videoURL):
		return "ffmpeg-hls"
	case LooksLikeHLS(videoURL) || hasUnsafeExtension(videoURL):
		return "native-hls"
	default:
		return "direct-mp4"
	}
}

// TestSuperFlixNeverRoutesToTheMP4Downloader is the guard with teeth: a
// playlist handed to the Range downloader produces a file containing the
// playlist text, which then fails to play with no obvious cause.
func TestSuperFlixNeverRoutesToTheMP4Downloader(t *testing.T) {
	t.Parallel()
	for _, tc := range superFlixMasterURLs {
		assert.Equalf(t, "ffmpeg-hls", routeOf(tc.url),
			"%s must not reach the MP4 Range downloader", tc.name)
	}
}

// A plain HLS playlist that is not SuperFlix still has to reach an HLS path,
// never the MP4 downloader.
func TestPlainHLSRoutesToAnHLSDownloader(t *testing.T) {
	t.Parallel()
	for _, u := range []string{
		"https://cdn.example/stream.m3u8",
		"https://cdn.example/hls/index", // no extension at all
		"https://cdn.example/play.aspx", // yt-dlp rejects the extension
	} {
		assert.NotEqualf(t, "direct-mp4", routeOf(u), "%s must not be Range-downloaded", u)
	}
}

// Ordinary media must keep taking the fast Range path — the fix must not drag
// plain MP4s through ffmpeg.
func TestPlainMediaStillUsesTheRangeDownloader(t *testing.T) {
	t.Parallel()
	for _, u := range []string{
		"https://lightspeedst.net/s6/mp4/show/hd/20.mp4",
		"https://cdn.example/episode.mp4?token=x",
		"http://127.0.0.1:37229/blogger_proxy",
	} {
		assert.Equalf(t, "direct-mp4", routeOf(u), "%s must stay on the Range downloader", u)
	}
}

// TestFFmpegHLSArgsCarryTheCDNContractForSuperFlix ties the routing fix to the
// header fix: reaching ffmpeg is only useful if ffmpeg is given the headers the
// CDN demands. Without them every segment 403s and the download dies mid-file.
func TestFFmpegHLSArgsCarryTheCDNContractForSuperFlix(t *testing.T) {
	const referer = "https://player.best/video/deadbeefdeadbeefdeadbeefdeadbeef"
	const ua = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 Chrome/151.0.0.0 Safari/537.36"
	pinSuperFlixStream(t, referer, ua)

	url := superFlixMasterURLs[0].url
	require.Equal(t, "ffmpeg-hls", routeOf(url), "precondition: the URL must reach ffmpeg")

	args := ffmpegHLSDownloadArgs(url, "/tmp/out.mp4", referer)
	headers := argValue(t, args, "-headers")
	for _, want := range []string{
		"Referer: " + referer,
		"Accept-Language: en-US,en;q=0.9",
		"Sec-CH-UA-Mobile: ?0",
	} {
		assert.Containsf(t, headers, want, "%s missing from ffmpeg -headers", want)
	}
	assert.Equal(t, ua, lastArgValue(t, args, "-user_agent"),
		"the CDN binds the signed URL to the UA that obtained it")
	assert.Equal(t, url, argValue(t, args, "-i"))
}

// Every shape must survive the whole argument build, not just the first.
func TestFFmpegHLSArgsForEveryServedShape(t *testing.T) {
	pinSuperFlixStream(t, "https://player.best/video/abc", "UA/1.0")
	for _, tc := range superFlixMasterURLs {
		args := ffmpegHLSDownloadArgs(tc.url, "/tmp/out.mp4", "https://player.best/video/abc")
		assert.Equalf(t, tc.url, argValue(t, args, "-i"), "%s: input URL must survive", tc.name)
		assert.Contains(t, fmt.Sprint(args), "-f", "the HLS demuxer must stay forced")
	}
}
