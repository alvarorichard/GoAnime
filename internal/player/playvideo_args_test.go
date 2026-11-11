package player

import (
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/util"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAppendPlaybackRefererArgsAddsGlobalRefererForDirectHTTP(t *testing.T) {
	restore := snapshotGlobalReferer()
	defer restore()
	util.SetGlobalReferer("https://allmanga.to")

	args, referer := appendPlaybackRefererArgs(nil, "https://tools.fast4speed.rsvp//media9/videos/id/sub/4?v=22", false)

	assert.Equal(t, "https://allmanga.to", referer)
	assert.Contains(t, args, "--http-header-fields=Referer: https://allmanga.to")
}

func TestAppendPlaybackRefererArgsAddsExplicitOrigin(t *testing.T) {
	restore := snapshotGlobalReferer()
	defer restore()
	previousOrigin := util.GetGlobalOrigin()
	defer util.SetGlobalOrigin(previousOrigin)
	util.SetGlobalReferer("https://anidrive.click/")
	util.SetGlobalOrigin("https://anidrive.click")

	args, referer := appendPlaybackRefererArgs(nil, "https://static.anidrive.click/hls/x/master.m3u8", true)

	assert.Equal(t, "https://anidrive.click/", referer)
	assert.Contains(t, args, "--http-header-fields=Referer: https://anidrive.click/,Origin: https://anidrive.click")
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

func TestAnimesFHDPlaybackUsesBrowserHeaders(t *testing.T) {
	// These arguments read process-wide playback context, so do not parallelize.
	restore := snapshotGlobalReferer()
	defer restore()
	util.SetGlobalReferer("https://animesfhd.net/")
	for _, tc := range []struct {
		url string
		hls bool
	}{
		{"https://animesfhd.net/api/v/s6?source=1&playlist=.m3u8", true},
		{"https://cdn.example/video.mp4", false},
	} {
		args := buildPlaybackArgs(playbackArgsInput{VideoURL: tc.url, IsHLS: tc.hls})
		assert.Contains(t, args, "--user-agent="+downloadUserAgent)
		assert.Contains(t, strings.Join(args, " "), "Referer: https://animesfhd.net/")
	}
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
// each kind of stream. This is the wiring guard: it proves a movie/TV HLS
// stream carries BOTH the Referer header AND allowed_extensions=ALL (the audio
// fix), that a local file carries neither, and that the 9Anime / upscaling /
// resume branches produce their expected flags. Not parallel: it sets the
// global referer that appendPlaybackRefererArgs reads.
func TestBuildPlaybackArgs(t *testing.T) {
	restore := snapshotGlobalReferer()
	defer restore()
	util.SetGlobalReferer("https://ref.test")

	t.Run("StartFlix HLS movie carries referer + allowed_extensions + langs", func(t *testing.T) {
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

	t.Run("HLS never forces a lavf demuxer format", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{
			VideoURL: "https://cdn.test/video/master.m3u8",
			IsHLS:    true,
		})
		assert.Contains(t, args, hlsAllowAllExtensionsArg)
		for _, a := range args {
			assert.Falsef(t, strings.HasPrefix(a, "--demuxer-lavf-format="), "unexpected forced format: %q", a)
		}
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
		assert.Contains(t, args, defaultVideoOutputArg(), "upscaling keeps the fallbacks too")
		assert.Contains(t, args, "--hwdec=no")
		assert.Contains(t, args, "--glsl-shader=/a.glsl")
		assert.NotContains(t, args, "--no-config", "upscaling must not use the standard profile")
	})

	t.Run("default VO is this OS's gpu-next chain", func(t *testing.T) {
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

	// HLS streams and plain files both receive their external subtitles.
	t.Run("external subtitles reach mpv for HLS and files alike", func(t *testing.T) {
		for _, url := range []string{"https://cdn/master.m3u8", "https://cdn/movie.mp4"} {
			args := buildPlaybackArgs(playbackArgsInput{
				VideoURL:    url,
				IsHLS:       LooksLikeHLS(url),
				IsMovieOrTV: true,
				SubArgs:     []string{"--sub-file=https://subs.example/pt.vtt"},
			})
			assert.Containsf(t, args, "--sub-file=https://subs.example/pt.vtt", "%s must keep its subtitles", url)
		}
	})
}

// userConfigViolations checks the contract of the user-config argument set and
// returns every rule broken. Shared by the test and its metatest.
func userConfigViolations(args []string) []string {
	var v []string
	has := func(a string) bool {
		for _, x := range args {
			if x == a {
				return true
			}
		}
		return false
	}
	if has("--no-config") {
		v = append(v, "passes --no-config")
	}
	if has("--profile=fast") {
		v = append(v, "forces --profile=fast over the user's quality settings")
	}
	if has("--video-latency-hacks=yes") {
		v = append(v, "forces latency hacks")
	}
	if !has("--hwdec=auto-safe") {
		v = append(v, "lost hwdec")
	}
	if !has(defaultVideoOutputArg()) {
		v = append(v, "lost the VO fallback chain")
	}
	for _, g := range userConfigGuardArgs() {
		if !has(g) {
			v = append(v, "missing guard "+g)
		}
	}
	return v
}

func TestBuildPlaybackArgsUserMPVConfig(t *testing.T) {
	t.Run("user config drops --no-config and keeps the guards", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{VideoURL: "/tmp/x.mp4", UserMPVConfig: true})
		assert.Empty(t, userConfigViolations(args))
	})

	t.Run("isolated profile is unchanged", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{VideoURL: "/tmp/x.mp4"})
		assert.Contains(t, args, "--no-config")
		assert.Contains(t, args, "--profile=fast")
		for _, g := range userConfigGuardArgs() {
			assert.NotContains(t, args, g, "guards are only for loaded configs")
		}
	})

	t.Run("upscaling loads the config so it gets the guards", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{
			VideoURL: "/tmp/x.mp4", UpscalingEnabled: true, ShaderArgs: []string{"--glsl-shader=/a.glsl"},
		})
		assert.NotContains(t, args, "--no-config")
		for _, g := range userConfigGuardArgs() {
			assert.Contains(t, args, g)
		}
	})

	t.Run("referer, HLS and resume still apply with the user config", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{
			VideoURL: "/tmp/x.mp4", UserMPVConfig: true, ResumeTime: 30, Title: "T",
		})
		assert.Contains(t, args, "--start=+30")
		assert.Contains(t, args, "--force-media-title=T")
	})

	t.Run("every emitted flag passes the allowlist", func(t *testing.T) {
		args := buildPlaybackArgs(playbackArgsInput{VideoURL: "/tmp/x.mp4", UserMPVConfig: true})
		assert.Equal(t, args, filterMPVArgs(args), "StartVideo must not silently drop any of them")
	})
}

// TestUserConfigViolationsMeta proves the contract check has teeth: each
// mutation of a correct argument list must be reported.
func TestUserConfigViolationsMeta(t *testing.T) {
	good := buildPlaybackArgs(playbackArgsInput{VideoURL: "/tmp/x.mp4", UserMPVConfig: true})
	require.Empty(t, userConfigViolations(good))

	without := func(drop string) []string {
		var out []string
		for _, a := range good {
			if a != drop {
				out = append(out, a)
			}
		}
		return out
	}
	mutants := map[string][]string{
		"no-config added":   append(append([]string{}, good...), "--no-config"),
		"profile fast":      append(append([]string{}, good...), "--profile=fast"),
		"latency hacks":     append(append([]string{}, good...), "--video-latency-hacks=yes"),
		"hwdec dropped":     without("--hwdec=auto-safe"),
		"vo chain dropped":  without(defaultVideoOutputArg()),
		"keep-open dropped": without("--keep-open=no"),
		"idle dropped":      without("--idle=no"),
		"loop-file dropped": without("--loop-file=no"),
		"resume dropped":    without("--resume-playback=no"),
		"isolated profile":  buildPlaybackArgs(playbackArgsInput{VideoURL: "/tmp/x.mp4"}),
	}
	for name, m := range mutants {
		assert.NotEmpty(t, userConfigViolations(m), "mutant %q survived", name)
	}
}
