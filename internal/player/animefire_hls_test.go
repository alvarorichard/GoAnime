package player

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// AnimeFire's CDN serves its HLS playlists under an image name:
//
//	https://akumast.net/i/<token>/h.jpg      multivariant master
//	.../<variant>/p.jpg                      variant playlists
//
// Both answer application/vnd.apple.mpegurl with a body starting #EXTM3U.
//
// LooksLikeHLS did not know that, so a download took the plain-MP4 branch and
// saved the ~200-byte playlist as the episode; the 10 MB floor then deleted it
// and reported a failed download, with nothing to say the URL had been misread.
// Playback shares this helper, so it was routing the same URL wrong too.
//
// This is the third shape to catch this code out — SuperFlix's master.txt and
// its later path change are documented beside it for the same reason.

func TestLooksLikeHLS_KnowsAnimeFiresDisguisedPlaylist(t *testing.T) {
	t.Parallel()
	for name, tt := range map[string]struct {
		url  string
		want bool
	}{
		"animefire master":            {"https://akumast.net/i/mqkkYJoeiFt51sJd/h.jpg", true},
		"animefire variant":           {"https://akumast.net/Vmhagys/p.jpg", true},
		"animefire master with query": {"https://akumast.net/i/tok/h.jpg?x=1", true},
		"still a plain m3u8":          {"https://cdn.example.com/master.m3u8", true},
		"still an hls path":           {"https://cdn.example.com/hls/stream", true},
		"still superflix master.txt":  {"https://host.best/tok/id/exp/master.txt", true},

		// Must NOT match: these would be routed to the HLS downloader for nothing.
		"an ordinary mp4":          {"https://cdn.example.com/video.mp4", false},
		"a poster image":           {"https://image.tmdb.org/t/p/w500/poster.jpg", false},
		"a path merely ending jpg": {"https://cdn.example.com/thumbnails/frame01.jpg", false},
		"empty":                    {"", false},
	} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, LooksLikeHLS(tt.url))
		})
	}
}

// The match is on the final path segment, not the host: the CDN host rotates
// and the naming has not.
func TestIsAnimeFireHLS_MatchesTheSegmentNotTheHost(t *testing.T) {
	t.Parallel()

	assert.True(t, isAnimeFireHLS("https://some-other-cdn.example/i/tok/h.jpg"),
		"a rotated host must still be recognised")
	assert.False(t, isAnimeFireHLS("https://akumast.net/i/tok/video.mp4"),
		"the host alone means nothing; the playlist naming is the signal")
}
