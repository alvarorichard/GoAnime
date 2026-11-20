package animesdigital

import (
	"encoding/base64"
	"errors"
	"net/http"
	"strings"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
)

// Issue #237: on 2026-11-18 AnimesDigital changed its players. Player 1 is no
// longer api.anivideo.net/videohls.php?d=<playlist> but a page on that host,
// /cdn/v1.<token>, that sets its own player up with the playlist, and the host
// refuses (403) a request that does not name the episode as its Referer.
// Player 2 of older episodes became /<ROT13 of base64 of the post>/0/<n>/bg.mp4.
// Every episode on the site stopped playing in GoAnime.

const fbCDNIframe = "https://api.anivideo.net/cdn/v1.AbC-token_1?i=https%3A%2F%2Fanimesdigital.org%2Fx.jpg"

func TestStreamReadsPlayerOnesPageWhenItIsACDNFrame(t *testing.T) {
	s, c := newFBSite(t)
	s.set(func() { s.playerOne = fbCDNIframe })
	got, err := c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
	if err != nil {
		t.Fatal(err)
	}
	if want := "https://cdn-s01.mywallpaper-4k-image.net" + fbPlaylist; got != want {
		t.Fatalf("stream = %q, want the playlist the page names, %q", got, want)
	}
	// The fake host answers 403 unless the Referer is the episode, so getting
	// the playlist at all proves the page was asked for the way the site does.
	if n := s.hit("/cdn/v1.AbC-token_1"); n != 1 {
		t.Fatalf("player 1's page was asked for %d times, want 1", n)
	}
	if n := s.hit(fbPiece); n != 1 {
		t.Fatalf("the first piece of video was asked for %d times: the playlist must be probed like any other", n)
	}
}

func TestStreamFallsToPlayerTwoWhenPlayerOnesPageIsRefused(t *testing.T) {
	// Older episodes: the host answers 403 for their page and only player 2,
	// a Blogger video, has them.
	for _, status := range []int{http.StatusForbidden, http.StatusNotFound, 520} {
		s, c := newFBSite(t)
		s.set(func() {
			s.playerOne = fbCDNIframe
			s.cdnStatus = status
		})
		got, err := c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
		if err != nil || got != fbToken(4) {
			t.Fatalf("page %d: got %q, err %v, want player 2's video", status, got, err)
		}
	}
}

func TestStreamFallsToPlayerTwoWhenPlayerOnesPageNamesNoPlaylist(t *testing.T) {
	for name, body := range map[string]string{
		"no file":      `<html><script>jwplayer('x').setup({ playlist: [] })</script></html>`,
		"not a stream": `<html><script>setup({ file: "https:\/\/cdn.example\/video.mp4" })</script></html>`,
		"not https":    `<html><script>setup({ file: "http:\/\/cdn.example\/index.m3u8" })</script></html>`,
		"empty":        ``,
	} {
		s, c := newFBSite(t)
		s.set(func() {
			s.playerOne = fbCDNIframe
			s.cdnBody = body
		})
		got, err := c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
		if err != nil || got != fbToken(4) {
			t.Fatalf("%s: got %q, err %v, want player 2's video", name, got, err)
		}
		// What the page names is not asked for unless it is an https playlist.
		for _, path := range []string{"/video.mp4", "/index.m3u8"} {
			if n := s.hit(path); n != 0 {
				t.Fatalf("%s: %s was asked for %d times", name, path, n)
			}
		}
	}
}

func TestEpisodeWhosePlayerOnesPageIsRefusedAndHasNoPlayerTwoIsOffline(t *testing.T) {
	s, c := newFBSite(t)
	s.set(func() {
		s.playerOne = fbCDNIframe
		s.playerTwo = ""
		s.cdnStatus = http.StatusForbidden
	})
	_, err := c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
	if !errors.Is(err, netx.ErrMediaOffline) {
		t.Fatalf("err = %v, want the media-offline error, not a URL that plays nothing", err)
	}
}

func TestStreamFallsToPlayerTwoWhenPlayerOnesPiecesAreRefused(t *testing.T) {
	s, c := newFBSite(t)
	s.set(func() {
		s.playerOne = fbCDNIframe
		s.pieceStatus = http.StatusForbidden
	})
	got, err := c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
	if err != nil || got != fbToken(4) {
		t.Fatalf("got %q, err %v", got, err)
	}
}

func TestTheLegacyPlayerOneAddressStillWorks(t *testing.T) {
	s, c := newFBSite(t) // its playerOne is videohls.php?d=<playlist>
	got, err := c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
	if err != nil || !strings.HasSuffix(got, fbPlaylist) {
		t.Fatalf("got %q, err %v", got, err)
	}
	if n := s.hit("/cdn/v1.AbC-token_1"); n != 0 {
		t.Fatalf("the cdn page was asked for %d times for a legacy address", n)
	}
}

func TestPlayerTwoReadsTheOlderAddress(t *testing.T) {
	post := base64.StdEncoding.EncodeToString([]byte(fbPost))
	for name, tc := range map[string]struct {
		path string
		want int
	}{
		"rot13, 0/n":       {rot13(post) + "/0/7/bg.mp4", 7},
		"rot13, first":     {rot13(post) + "/0/0/bg.mp4", 0},
		"plain, 0/n":       {post + "/0/3/bg.mp4", 3},
		"plain, current":   {post + "/5/bg.mp4", 5},
		"rot13, current":   {rot13(post) + "/2/bg.mp4", 2},
		"query is ignored": {rot13(post) + "/0/9/bg.mp4?p=2&q=QmxlYWNo&nocache1", 9},
	} {
		s, c := newFBSite(t)
		got, err := c.playerTwoStream(t.Context(), s.srv.URL+"/"+tc.path)
		if err != nil || got != fbToken(tc.want) {
			t.Fatalf("%s: got %q, err %v, want video %d", name, got, err, tc.want)
		}
	}
}

func TestPlayerTwoRefusesAnOlderAddressThatIsWrong(t *testing.T) {
	post := base64.StdEncoding.EncodeToString([]byte(fbPost))
	for name, path := range map[string]string{
		"five segments":       rot13(post) + "/0/0/1/bg.mp4",
		"index is no number":  rot13(post) + "/0/x/bg.mp4",
		"negative index":      rot13(post) + "/0/-1/bg.mp4",
		"not a post":          "bm90IGEgdXJs/0/1/bg.mp4",
		"not a file":          rot13(post) + "/0/1/video.mp4",
		"outside blogger":     rot13(base64.StdEncoding.EncodeToString([]byte("https://evil.example/post"))) + "/0/1/bg.mp4",
		"past the last video": rot13(post) + "/0/10/bg.mp4",
	} {
		s, c := newFBSite(t)
		if got, err := c.playerTwoStream(t.Context(), s.srv.URL+"/"+path); err == nil {
			t.Fatalf("%s: got %q, want an error", name, got)
		}
	}
}

func TestRot13(t *testing.T) {
	for in, want := range map[string]string{
		"nUE0pUZ6": "aHR0cHM6",
		"aHR0cHM6": "nUE0pUZ6",
		"Az09+/=":  "Nm09+/=",
		"":         "",
	} {
		if got := rot13(in); got != want {
			t.Fatalf("rot13(%q) = %q, want %q", in, got, want)
		}
		if got := rot13(rot13(in)); got != in {
			t.Fatalf("rot13 twice of %q = %q", in, got)
		}
	}
}
