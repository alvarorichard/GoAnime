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

	args, referer := appendPlaybackRefererArgs(nil, "https://tools.fast4speed.rsvp//media9/videos/id/sub/4?v=22", false, false)

	assert.Equal(t, "https://allmanga.to", referer)
	assert.Contains(t, args, "--http-header-fields=Referer: https://allmanga.to")
}

func TestAppendPlaybackRefererArgsKeepsHLSFallbackReferer(t *testing.T) {
	restore := snapshotGlobalReferer()
	defer restore()
	util.ClearGlobalReferer()

	args, referer := appendPlaybackRefererArgs(nil, "https://cdn.example.com/master.m3u8", true, false)

	assert.Equal(t, defaultHLSReferer, referer)
	assert.Contains(t, args, "--http-header-fields=Referer: "+defaultHLSReferer)
}

func TestAppendPlaybackRefererArgsSkipsLocalFiles(t *testing.T) {
	restore := snapshotGlobalReferer()
	defer restore()
	util.SetGlobalReferer("https://allmanga.to")

	args, referer := appendPlaybackRefererArgs([]string{"--cache=yes"}, "/tmp/episode.mp4", false, false)

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

// TestBuildPlaybackArgs pins the full mpv argument set the player assembles for
// each kind of stream. This is the wiring guard: it proves a SuperFlix HLS
// stream carries BOTH the Referer header AND allowed_extensions=ALL (the audio
// fix), that a local file carries neither, and that the 9Anime / upscaling /
// resume branches produce their expected flags. Not parallel: it sets the
// global referer that appendPlaybackRefererArgs reads.
func TestBuildPlaybackArgs(t *testing.T) {
	restore := snapshotGlobalReferer()
	defer restore()
	util.SetGlobalReferer("https://ref.test")

	t.Run("SuperFlix HLS movie carries referer + allowed_extensions + langs", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{
			VideoURL:    "https://cdn.test/master.m3u8",
			IsHLS:       true,
			IsMovieOrTV: true,
			AudioLang:   "por",
			SubsLang:    "por",
			Title:       "Psicose",
		})
		assert.Contains(t, args, hlsAllowAllExtensionsArg, "the audio fix flag must be present for HLS")
		assert.Contains(t, args, "--http-header-fields=Referer: https://ref.test")
		assert.Contains(t, args, "--alang=por")
		assert.Contains(t, args, "--slang=por")
		assert.Contains(t, args, "--force-media-title=Psicose")
		// Base args always present.
		assert.Contains(t, args, "--cache=yes")
		// Must NOT take the 9Anime yt-dlp path.
		assert.NotContains(t, args, "--script-opts=ytdl_hook-try_ytdl_first=yes")
	})

	t.Run("SuperFlix master txt forces the HLS demuxer", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{
			VideoURL: "https://cdn.test/cdn/hls/hash/master.txt",
			IsHLS:    true,
		})
		assert.Contains(t, args, hlsAllowAllExtensionsArg)
		assert.Contains(t, args, hlsForceLavfFormatArg)
	})

	t.Run("ordinary m3u8 does not need a forced demuxer", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{
			VideoURL: "https://cdn.test/video/master.m3u8",
			IsHLS:    true,
		})
		assert.Contains(t, args, hlsAllowAllExtensionsArg)
		assert.NotContains(t, args, hlsForceLavfFormatArg)
	})

	t.Run("local non-HLS file carries neither referer nor allowed_extensions", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{
			VideoURL: "/tmp/episode.mp4",
			IsHLS:    false,
		})
		assert.NotContains(t, args, hlsAllowAllExtensionsArg)
		assert.NotContains(t, args, "--http-header-fields=Referer: https://ref.test")
		assert.NotContains(t, args, "--alang=por")
		assert.Contains(t, args, "--no-config", "non-upscaled playback uses the standard profile")
	})

	t.Run("9Anime HLS adds yt-dlp impersonation plus the audio fix", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{
			VideoURL:       "https://9anime.cdn/master.m3u8",
			IsHLS:          true,
			Is9Anime:       true,
			CanImpersonate: true,
		})
		assert.Contains(t, args, hlsAllowAllExtensionsArg)
		assert.Contains(t, args, "--script-opts=ytdl_hook-try_ytdl_first=yes")
		assert.Contains(t, args, "--ytdl-raw-options-append=referer=https://ref.test")
		assert.Contains(t, args, "--ytdl-raw-options-append=impersonate=chrome")
	})

	t.Run("9Anime without impersonation support omits the impersonate flag", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{
			VideoURL:       "https://9anime.cdn/master.m3u8",
			IsHLS:          true,
			Is9Anime:       true,
			CanImpersonate: false,
		})
		assert.Contains(t, args, "--script-opts=ytdl_hook-try_ytdl_first=yes")
		assert.NotContains(t, args, "--ytdl-raw-options-append=impersonate=chrome")
	})

	t.Run("upscaling swaps the render profile and injects shader args", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{
			VideoURL:         "/tmp/x.mp4",
			UpscalingEnabled: true,
			ShaderArgs:       []string{"--glsl-shader=/a.glsl"},
		})
		assert.Contains(t, args, "--vo=gpu-next")
		assert.Contains(t, args, "--hwdec=no")
		assert.Contains(t, args, "--glsl-shader=/a.glsl")
		assert.NotContains(t, args, "--no-config", "upscaling must not use the standard profile")
	})

	t.Run("default VO uses Windows fallback chain", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{VideoURL: "/tmp/x.mp4"})
		assert.Contains(t, args, defaultVideoOutputArg())
		assert.Contains(t, args, "--hwdec=auto-safe")
	})

	t.Run("resume adds --start for non-HLS but not for HLS", func(t *testing.T) {
		nonHLS := buildPlaybackArgs(playbackArgsInput{VideoURL: "/tmp/x.mp4", ResumeTime: 30})
		assert.Contains(t, nonHLS, "--start=+30")

		hls := buildPlaybackArgs(playbackArgsInput{VideoURL: "https://cdn/master.m3u8", IsHLS: true, ResumeTime: 30})
		assert.NotContains(t, hls, "--start=+30", "HLS resume is handled by seeking, not --start")
	})

	t.Run("external subtitle args are appended verbatim", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{
			VideoURL:    "https://cdn/master.m3u8",
			IsHLS:       true,
			IsMovieOrTV: true,
			SubArgs:     []string{"--sub-file=/tmp/pt.srt"},
		})
		assert.Contains(t, args, "--sub-file=/tmp/pt.srt")
	})
}
