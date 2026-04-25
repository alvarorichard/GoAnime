package player

import (
	"testing"

	"github.com/alvarorichard/Goanime/internal/util"
	"github.com/stretchr/testify/assert"
)

func TestAppendPlaybackRefererArgsAddsGlobalRefererForDirectHTTP(t *testing.T) {
	restore := snapshotGlobalReferer()
	defer restore()
	util.SetGlobalReferer("https://allmanga.to")

	args, referer := appendPlaybackRefererArgs(nil, "https://tools.fast4speed.rsvp//media9/videos/id/sub/4?v=22", false)

	assert.Equal(t, "https://allmanga.to", referer)
	assert.Contains(t, args, "--http-header-fields=Referer: https://allmanga.to")
}

func TestAppendPlaybackRefererArgsKeepsHLSFallbackReferer(t *testing.T) {
	restore := snapshotGlobalReferer()
	defer restore()
	util.ClearGlobalReferer()

	args, referer := appendPlaybackRefererArgs(nil, "https://cdn.example.com/master.m3u8", true)

	assert.Equal(t, defaultHLSReferer, referer)
	assert.Contains(t, args, "--http-header-fields=Referer: "+defaultHLSReferer)
}

func TestAppendPlaybackRefererArgsSkipsLocalFiles(t *testing.T) {
	restore := snapshotGlobalReferer()
	defer restore()
	util.SetGlobalReferer("https://allmanga.to")

	args, referer := appendPlaybackRefererArgs([]string{"--cache=yes"}, "/tmp/episode.mp4", false)

	assert.Empty(t, referer)
	assert.Equal(t, []string{"--cache=yes"}, args)
}
