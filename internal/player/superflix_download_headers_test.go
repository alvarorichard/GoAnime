package player

import (
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// pinSuperFlixStream points the playback globals at a SuperFlix stream the way
// enhanced.go does when one is resolved, and undoes it after the test.
func pinSuperFlixStream(t *testing.T, referer, userAgent string) {
	t.Helper()
	t.Cleanup(snapshotGlobalReferer())
	util.SetGlobalReferer(referer)
	util.SetGlobalUserAgent(userAgent)
	prev := util.GetGlobalAnimeSource()
	util.SetGlobalAnimeSource("SuperFlix")
	t.Cleanup(func() {
		util.ClearGlobalUserAgent()
		util.SetGlobalAnimeSource(prev)
	})
}

// TestFFmpegHLSDownloadSendsCDNContract_2026_08_31 guards the download half of
// the playback fix. ffmpeg's own User-Agent and a lone Referer get every
// segment 403'd by SuperFlix's player CDN, which matches the browser's exact
// UA, Accept-Language and Sec-CH-UA-* hints.
func TestFFmpegHLSDownloadSendsCDNContract_2026_08_31(t *testing.T) {
	const referer = "https://player.best/video/deadbeefdeadbeefdeadbeefdeadbeef"
	const ua = "Mozilla/5.0 (X11; Linux x86_64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/151.0.0.0 Safari/537.36"
	pinSuperFlixStream(t, referer, ua)

	args := ffmpegHLSDownloadArgs("https://cdn.example/master.txt", "/tmp/out.mp4", referer)

	headers := argValue(t, args, "-headers")
	for _, want := range []string{
		"Referer: " + referer,
		"User-Agent: " + ua,
		"Accept-Language: en-US,en;q=0.9",
		"Sec-CH-UA-Mobile: ?0",
		"Sec-CH-UA-Platform: ",
		"Origin: https://player.best",
	} {
		assert.Containsf(t, headers, want, "%s missing from ffmpeg -headers", want)
	}
	assert.True(t, strings.HasSuffix(headers, "\r\n"), "ffmpeg needs a trailing CRLF, got %q", headers)

	// The pinned UA has to REPLACE ffmpeg's default, which is passed earlier in
	// the same argv — ffmpeg keeps the last occurrence.
	assert.Equal(t, ua, lastArgValue(t, args, "-user_agent"))

	// The download still has to produce a real command.
	assert.Equal(t, "https://cdn.example/master.txt", argValue(t, args, "-i"))
	assert.Equal(t, "/tmp/out.mp4", args[len(args)-1])
}

func TestFFprobeHLSDurationSendsCDNContract(t *testing.T) {
	const referer = "https://player.best/video/abc"
	const ua = "Chrome/151-test"
	pinSuperFlixStream(t, referer, ua)

	args := ffprobeHLSDurationArgs("https://cdn.example/master.txt", referer)
	assert.Contains(t, argValue(t, args, "-headers"), "Sec-CH-UA-Mobile: ?0")
	assert.Equal(t, ua, lastArgValue(t, args, "-user_agent"))
}

// Sources that pin no User-Agent keep the plain Referer-only behaviour.
func TestFFmpegHLSDownloadKeepsPlainRefererForOtherSources(t *testing.T) {
	t.Cleanup(snapshotGlobalReferer())
	util.SetGlobalReferer("https://animefire.io")
	util.ClearGlobalUserAgent()

	args := ffmpegHLSDownloadArgs("https://cdn.example/master.m3u8", "/tmp/out.mp4", "https://animefire.io")
	headers := argValue(t, args, "-headers")
	assert.Equal(t, "Referer: https://animefire.io\r\n", headers)
	assert.Equal(t, downloadUserAgent, lastArgValue(t, args, "-user_agent"))
}

func argValue(t *testing.T, args []string, flag string) string {
	t.Helper()
	for i, a := range args {
		if a == flag {
			require.Less(t, i+1, len(args), "%s has no value", flag)
			return args[i+1]
		}
	}
	t.Fatalf("%s not found in %v", flag, args)
	return ""
}

func lastArgValue(t *testing.T, args []string, flag string) string {
	t.Helper()
	got := ""
	found := false
	for i, a := range args {
		if a == flag && i+1 < len(args) {
			got, found = args[i+1], true
		}
	}
	require.True(t, found, "%s not found in %v", flag, args)
	return got
}
