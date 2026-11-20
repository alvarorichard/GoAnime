package player

import (
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/providers/anroll"
)

// AnimesROLL's Blogger links are signed for the User-Agent that asked for the
// player; the downloader and mpv must send that exact one or Google answers 403.
func TestAnrollPlayerUserAgentMatchesDownloader(t *testing.T) {
	if anroll.PlayerUserAgent != downloadUserAgent {
		t.Fatalf("anroll.PlayerUserAgent = %q, downloadUserAgent = %q: they must be equal", anroll.PlayerUserAgent, downloadUserAgent)
	}
}

func TestIsDirectBloggerGoogleVideo(t *testing.T) {
	cases := map[string]bool{
		"https://redirector.googlevideo.com/videoplayback?expire=1&source=blogger&itag=22": true,
		"https://rr2---sn-x.googlevideo.com/videoplayback?source=blogger":                  true,
		"https://rr2---sn-x.googlevideo.com/videoplayback?source=youtube":                  false,
		"http://127.0.0.1:5000/blogger_proxy":                                              false,
		"https://static.anidrive.click/master.m3u8":                                        false,
		"not a url": false,
	}
	for u, want := range cases {
		if got := isDirectBloggerGoogleVideo(u); got != want {
			t.Errorf("isDirectBloggerGoogleVideo(%q) = %v, want %v", u, got, want)
		}
	}
}
