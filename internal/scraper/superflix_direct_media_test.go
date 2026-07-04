package scraper

// Tests for sfDirectMediaRe, the pattern feeding SniffEmbedStream's
// last-resort capture (added 2026-07-01 after the rotating player host broke
// the do=getVideo interception while the video played on in the solver
// window). It must recognize real media traffic and must NOT match the
// player's API URLs — matching those would "adopt" a JSON endpoint as the
// stream and hand mpv an unplayable URL.

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSFDirectMediaRe(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		url  string
		want bool
	}{
		{"signed HLS master", "https://xn--kcksk7a2bl5le7b6doc1h3f.com/cdn/hls/c8a4e367/master.m3u8?md5=x&expires=1", true},
		{"bare m3u8", "https://host.test/stream.m3u8", true},
		{"m3u8 with fragment", "https://host.test/stream.m3u8#frag", true},
		{"mp4 file", "https://host.test/video/movie.mp4", true},
		{"mp4 with query", "https://host.test/movie.mp4?token=abc", true},
		{"hls segment path", "https://host.test/cdn/hls/c8a4e367/seg-1.ts", true},
		{"unsigned master.txt", "https://host.test/video/master.txt", true},
		{"getVideo API must not match", "https://host.test/player/index.php?data=abc123&do=getVideo", false},
		{"securedLink API must not match", "https://host.test/api?do=securedLink", false},
		{"embed page must not match", "https://superflixapi.pro/filme/136797", false},
		{"m3u8 in the middle of a path must not match", "https://host.test/stream.m3u8.html", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, sfDirectMediaRe.MatchString(tt.url), "url: %s", tt.url)
		})
	}
}
