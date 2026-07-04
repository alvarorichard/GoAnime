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

func TestAppendHLSDemuxerArgs(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		in          []string
		isHLS       bool
		wantAllow   bool // the allowed_extensions=ALL flag must be present
		wantPreface []string
	}{
		{
			name:        "hls appends allow-all flag",
			in:          nil,
			isHLS:       true,
			wantAllow:   true,
			wantPreface: nil,
		},
		{
			name:        "non-hls leaves args untouched",
			in:          []string{"--cache=yes"},
			isHLS:       false,
			wantAllow:   false,
			wantPreface: []string{"--cache=yes"},
		},
		{
			name:        "hls preserves existing args before the flag",
			in:          []string{"--cache=yes", "--http-header-fields=Referer: https://x/"},
			isHLS:       true,
			wantAllow:   true,
			wantPreface: []string{"--cache=yes", "--http-header-fields=Referer: https://x/"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			got := appendHLSDemuxerArgs(tt.in, tt.isHLS)

			if tt.wantAllow {
				assert.Contains(t, got, hlsAllowAllExtensionsArg)
				// The flag must be appended exactly once.
				count := 0
				for _, a := range got {
					if a == hlsAllowAllExtensionsArg {
						count++
					}
				}
				assert.Equal(t, 1, count, "flag must appear exactly once")
			} else {
				assert.NotContains(t, got, hlsAllowAllExtensionsArg)
			}

			// Existing args must be preserved, in order, at the front.
			if len(tt.wantPreface) > 0 {
				assert.Equal(t, tt.wantPreface, got[:len(tt.wantPreface)])
			}
		})
	}
}

// TestHLSAllowAllExtensionsArgValue pins the exact mpv option string so a typo
// (which mpv would silently ignore, re-breaking audio) fails the build.
func TestHLSAllowAllExtensionsArgValue(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "--demuxer-lavf-o=allowed_extensions=ALL", hlsAllowAllExtensionsArg)
}
