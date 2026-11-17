package animesdigital

import (
	"context"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
)

// Issue #236: one episode of a title made the player close at once, while the
// others, and the same episode on another source, played. The site's video
// host has no file for some episodes it still lists and answers 520 for their
// playlist; that URL was handed to the player unasked. These tests hold the
// fix: player 1's playlist is asked for first, and when its host does not
// answer, the episode's copy on player 2 (a video uploaded to a Blogger post)
// is used; when neither has it, the episode is reported as offline.

const (
	fbPlaylist = "/stream/sv/c/cyberpunk-edgerunners-2/05.mp4/index.m3u8"
	fbPost     = "https://bulborn.blogspot.com/2026/10/cyberp-ed-2.html"
)

// fbSite is the site, its video host and Blogger behind one test server: the
// client's transport sends every host there (rewriteHost), and the handler
// tells them apart by path.
type fbSite struct {
	srv *httptest.Server
	mu  sync.Mutex
	// what the video host answers for player 1's playlist
	playlistStatus int
	playlistBody   string
	// the episode page's iframes
	playerOne, playerTwo string
	// the Blogger post
	postStatus int
	postBody   string
	hits       map[string]int
	// onHit, when set, is called with each request's path before it is answered.
	onHit func(path string)
}

func newFBSite(t *testing.T) (*fbSite, *Client) {
	t.Helper()
	s := &fbSite{
		playlistStatus: http.StatusOK,
		playlistBody:   "#EXTM3U\n#EXT-X-VERSION:3\n#EXTINF:10.000,\nseg-1-v1-a1.webp\n#EXT-X-ENDLIST\n",
		postStatus:     http.StatusOK,
		postBody:       fbPostWith(10),
		hits:           map[string]int{},
	}
	s.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.mu.Lock()
		s.hits[r.URL.Path]++
		playlistStatus, playlistBody := s.playlistStatus, s.playlistBody
		postStatus, postBody := s.postStatus, s.postBody
		page := fbEpisodePage(s.playerOne, s.playerTwo)
		hook := s.onHit
		s.mu.Unlock()
		if hook != nil {
			hook(r.URL.Path)
		}
		switch {
		case strings.HasPrefix(r.URL.Path, "/video/"):
			_, _ = w.Write([]byte(page))
		case r.URL.Path == fbPlaylist:
			w.WriteHeader(playlistStatus)
			_, _ = w.Write([]byte(playlistBody))
		case r.URL.Path == "/2026/10/cyberp-ed-2.html":
			w.WriteHeader(postStatus)
			_, _ = w.Write([]byte(postBody))
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(s.srv.Close)
	s.playerOne = "https://api.anivideo.net/videohls.php?d=" + strings.ReplaceAll("https://cdn-s01.mywallpaper-4k-image.net"+fbPlaylist, "/", "%2F") + "&nocache1"
	s.playerTwo = s.playerTwoURL(fbPost, 4)
	c := NewClientForTest(&http.Client{Transport: rewriteHost{target: s.srv.URL}}, s.srv.URL)
	return s, c
}

// playerTwoURL is player 2's address for the n-th video of post, as the site
// writes it.
func (s *fbSite) playerTwoURL(post string, n int) string {
	return fmt.Sprintf("%s/%s/%d/bg.mp4?p=2&q=Q3liZXI=&nocache1", s.srv.URL, base64.StdEncoding.EncodeToString([]byte(post)), n)
}

func (s *fbSite) hit(path string) int {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.hits[path]
}

func (s *fbSite) set(fn func()) {
	s.mu.Lock()
	defer s.mu.Unlock()
	fn()
}

func fbEpisodePage(playerOne, playerTwo string) string {
	var b strings.Builder
	b.WriteString("<html><body>")
	for _, src := range []string{playerOne, playerTwo} {
		if src != "" {
			fmt.Fprintf(&b, `<iframe class="metaframe rptss no-lazy" src="%s" frameborder="0"></iframe>`, strings.ReplaceAll(src, "&", "&#038;"))
		}
	}
	b.WriteString("</body></html>")
	return b.String()
}

func fbToken(n int) string { return fmt.Sprintf("%stoken-%02d", bloggerVideoPrefix, n) }

// fbPostWith is a Blogger post with n uploaded videos, and around them what a
// real post also holds: the comment editor's frame and an embed from elsewhere.
func fbPostWith(n int) string {
	var b strings.Builder
	b.WriteString(`<html><body><iframe src="https://www.youtube.com/embed/x"></iframe>`)
	for i := range n {
		fmt.Fprintf(&b, `<iframe allowfullscreen="allowfullscreen" class="b-hbp-video b-uploaded" id="BLOGGER-video-%d" src="%s"></iframe>`, i, fbToken(i))
	}
	b.WriteString(`<iframe id='comment-editor' src=''></iframe><iframe src="` + bloggerVideoPrefix + `"></iframe></body></html>`)
	return b.String()
}

const fbEpisode = "/video/a/139346/"

func TestStreamIsPlayerOneWhenItsPlaylistAnswers(t *testing.T) {
	s, c := newFBSite(t)
	got, err := c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
	if err != nil {
		t.Fatal(err)
	}
	if want := "https://cdn-s01.mywallpaper-4k-image.net" + fbPlaylist; got != want {
		t.Fatalf("stream = %q, want %q", got, want)
	}
	if n := s.hit(fbPlaylist); n != 1 {
		t.Fatalf("the playlist was asked for %d times, want once", n)
	}
	if n := s.hit("/2026/10/cyberp-ed-2.html"); n != 0 {
		t.Fatalf("player 2's post was read %d times although player 1 answers", n)
	}
}

func TestStreamFallsBackToPlayerTwoWhenPlayerOneDoesNotAnswer(t *testing.T) {
	refusals := map[string]func(s *fbSite){
		"HTTP 520, as in the issue":    func(s *fbSite) { s.playlistStatus = 520 },
		"HTTP 404":                     func(s *fbSite) { s.playlistStatus = http.StatusNotFound },
		"HTTP 403":                     func(s *fbSite) { s.playlistStatus = http.StatusForbidden },
		"a page instead of a playlist": func(s *fbSite) { s.playlistBody = "<html>error code: 520</html>" },
		"an empty answer":              func(s *fbSite) { s.playlistBody = "" },
	}
	for name, refuse := range refusals {
		t.Run(name, func(t *testing.T) {
			s, c := newFBSite(t)
			s.set(func() { refuse(s) })
			got, err := c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
			if err != nil {
				t.Fatal(err)
			}
			if got != fbToken(4) {
				t.Fatalf("stream = %q, want the post's fifth video %q", got, fbToken(4))
			}
		})
	}
}

func TestPlayerTwoPicksTheVideoItsAddressNames(t *testing.T) {
	for _, n := range []int{0, 4, 9} {
		s, c := newFBSite(t)
		s.set(func() {
			s.playlistStatus = 520
			s.playerTwo = s.playerTwoURL(fbPost, n)
		})
		got, err := c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
		if err != nil {
			t.Fatalf("video %d: %v", n, err)
		}
		if got != fbToken(n) {
			t.Fatalf("video %d: stream = %q, want %q (the embeds that are not uploaded videos do not count)", n, got, fbToken(n))
		}
	}
}

func TestAnEpisodeNeitherPlayerHasIsOffline(t *testing.T) {
	cases := map[string]func(s *fbSite){
		"no player 2 at all":                func(s *fbSite) { s.playerTwo = "" },
		"the post has fewer videos":         func(s *fbSite) { s.postBody = fbPostWith(3) },
		"the post is gone":                  func(s *fbSite) { s.postStatus = http.StatusNotFound },
		"the post holds no uploaded video":  func(s *fbSite) { s.postBody = "<html><body>removed</body></html>" },
		"player 2 has an address not known": func(s *fbSite) { s.playerTwo = s.srv.URL + "/token/61/other.mp4?p=2" },
		"player 2 points outside Blogger":   func(s *fbSite) { s.playerTwo = s.playerTwoURL("https://evil.example/2026/10/cyberp-ed-2.html", 4) },
		"player 2 points at Blogger over http": func(s *fbSite) {
			s.playerTwo = s.playerTwoURL("http://bulborn.blogspot.com/2026/10/cyberp-ed-2.html", 4)
		},
	}
	for name, breakIt := range cases {
		t.Run(name, func(t *testing.T) {
			s, c := newFBSite(t)
			s.set(func() {
				s.playlistStatus = 520
				breakIt(s)
			})
			got, err := c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
			if got != "" {
				t.Fatalf("stream = %q, want none: a dead URL is what closed the player", got)
			}
			if !errors.Is(err, netx.ErrMediaOffline) {
				t.Fatalf("err = %v, want ErrMediaOffline, which the player explains instead of crashing", err)
			}
			if !strings.Contains(err.Error(), "HTTP 520") {
				t.Fatalf("err = %v, want it to say how the video host answered", err)
			}
		})
	}
}

func TestPlayerTwoNeverReadsAPostOutsideBlogger(t *testing.T) {
	s, c := newFBSite(t)
	s.set(func() {
		s.playlistStatus = 520
		// The same path on another host: were it fetched, the rewriting
		// transport would land it on the post and the test would see the hit.
		s.playerTwo = s.playerTwoURL("https://evil.example/2026/10/cyberp-ed-2.html", 4)
	})
	_, _ = c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
	if n := s.hit("/2026/10/cyberp-ed-2.html"); n != 0 {
		t.Fatalf("a post outside Blogger was fetched %d times", n)
	}
}

func TestStreamWithOnlyPlayerTwo(t *testing.T) {
	s, c := newFBSite(t)
	s.set(func() { s.playerOne = "" })
	got, err := c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
	if err != nil || got != fbToken(4) {
		t.Fatalf("stream = %q, err = %v; want player 2's video", got, err)
	}
	if n := s.hit(fbPlaylist); n != 0 {
		t.Fatalf("no player 1, yet a playlist was asked for %d times", n)
	}

	s.set(func() { s.postStatus = http.StatusNotFound })
	_, err = c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
	if err == nil || !strings.Contains(err.Error(), "player 2") || errors.Is(err, netx.ErrMediaOffline) {
		t.Fatalf("err = %v, want a diagnostic that names player 2: a page with no player 1 is a layout to look at, not an offline episode", err)
	}
}

func TestStreamWithNoPlayerIsAParserError(t *testing.T) {
	s, c := newFBSite(t)
	s.set(func() { s.playerOne, s.playerTwo = "", "" })
	_, err := c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
	if err == nil || !strings.Contains(err.Error(), "no supported player iframe") {
		t.Fatalf("err = %v", err)
	}
}

func TestStreamStopsWhenTheCallerLeaves(t *testing.T) {
	s, c := newFBSite(t)
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	_, err := c.GetEpisodeStreamURLContext(ctx, s.srv.URL+fbEpisode)
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
	if n := s.hit("/2026/10/cyberp-ed-2.html"); n != 0 {
		t.Fatal("player 2 was tried for a caller who had left")
	}
}

func TestPlayerOneWithAnAddressThatIsNoURLIsRefused(t *testing.T) {
	s, c := newFBSite(t)
	s.set(func() { s.playerOne = "https://api.anivideo.net/videohls.php?d=ftp%3A%2F%2Fcdn%2Fx.m3u8" })
	_, err := c.GetEpisodeStreamURL(s.srv.URL + fbEpisode)
	if !errors.Is(err, netx.ErrInvalidStreamURL) {
		t.Fatalf("err = %v, want ErrInvalidStreamURL", err)
	}
}

// ---- playlistAnswers -------------------------------------------------------

func TestPlaylistAnswers(t *testing.T) {
	s, c := newFBSite(t)
	target := "https://cdn-s01.mywallpaper-4k-image.net" + fbPlaylist
	if err := c.playlistAnswers(t.Context(), target); err != nil {
		t.Fatalf("a playlist: %v", err)
	}
	s.set(func() { s.playlistBody = "\ufeff\r\n #EXTM3U\n" })
	if err := c.playlistAnswers(t.Context(), target); err != nil {
		t.Fatalf("a playlist behind a byte-order mark and blank lines: %v", err)
	}
	s.set(func() { s.playlistStatus = 520 })
	if err := c.playlistAnswers(t.Context(), target); err == nil || err.Error() != "HTTP 520" {
		t.Fatalf("err = %v, want HTTP 520", err)
	}
	s.set(func() { s.playlistStatus, s.playlistBody = http.StatusOK, "#EXTINF without the header" })
	if err := c.playlistAnswers(t.Context(), target); err == nil || err.Error() != "not a playlist" {
		t.Fatalf("err = %v, want not a playlist", err)
	}
	if err := c.playlistAnswers(t.Context(), "http://[::1"); err == nil {
		t.Fatal("an address no request can be made for must be an error")
	}
	s.srv.Close()
	if err := c.playlistAnswers(t.Context(), target); err == nil {
		t.Fatal("a host that is down must be an error")
	}
}

func TestPlaylistAnswersAsksTheWayThePlayerWill(t *testing.T) {
	var got http.Header
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		got = r.Header.Clone()
		_, _ = w.Write([]byte("#EXTM3U\n"))
	}))
	defer srv.Close()
	c := NewClientForTest(srv.Client(), srv.URL)
	if err := c.playlistAnswers(t.Context(), srv.URL+"/x.m3u8"); err != nil {
		t.Fatal(err)
	}
	if got.Get("Referer") != srv.URL+"/" {
		t.Fatalf("Referer = %q, want the site's, which the video host expects", got.Get("Referer"))
	}
	if got.Get("User-Agent") == "" || !strings.Contains(got.Get("Accept"), "mpegurl") {
		t.Fatalf("headers = %v", got)
	}
}

// ---- playerTwoStream -------------------------------------------------------

func TestPlayerTwoStreamRefusesAddressesItDoesNotKnow(t *testing.T) {
	s, c := newFBSite(t)
	b64 := base64.StdEncoding.EncodeToString([]byte(fbPost))
	cases := map[string]string{
		"too few parts":      s.srv.URL + "/" + b64 + "/bg.mp4",
		"too many parts":     s.srv.URL + "/x/" + b64 + "/4/bg.mp4",
		"another file name":  s.srv.URL + "/" + b64 + "/4/video.mp4",
		"no number":          s.srv.URL + "/" + b64 + "/four/bg.mp4",
		"a negative number":  s.srv.URL + "/" + b64 + "/-1/bg.mp4",
		"not base64":         s.srv.URL + "/!!!!/4/bg.mp4",
		"not a URL inside":   s.srv.URL + "/" + base64.StdEncoding.EncodeToString([]byte("://nope")) + "/4/bg.mp4",
		"credentials inside": s.srv.URL + "/" + base64.StdEncoding.EncodeToString([]byte("https://u:p@bulborn.blogspot.com/p.html")) + "/4/bg.mp4",
		"a lookalike host":   s.srv.URL + "/" + base64.StdEncoding.EncodeToString([]byte("https://blogspot.com.evil.example/p.html")) + "/4/bg.mp4",
		"not a URL at all":   "http://[::1",
	}
	for name, addr := range cases {
		if got, err := c.playerTwoStream(t.Context(), addr); err == nil || got != "" {
			t.Errorf("%s: got %q, err %v; want an error and nothing", name, got, err)
		}
	}
	if n := s.hit("/2026/10/cyberp-ed-2.html"); n != 0 {
		t.Fatalf("a refused address still fetched the post %d times", n)
	}
}

func TestPlayerTwoStreamReadsBothSpellingsOfBase64(t *testing.T) {
	s, c := newFBSite(t)
	// A post whose URL encodes to '+' or '/' in standard base64 is written
	// with '-' and '_' when it is URL-safe, and may lose its padding.
	post := "https://bulborn.blogspot.com/2026/10/cyberp-ed-2.html?>>>"
	urlSafe := strings.TrimRight(base64.URLEncoding.EncodeToString([]byte(post)), "=")
	if !strings.ContainsAny(urlSafe, "-_") {
		t.Fatalf("the test's post does not exercise the URL-safe alphabet: %s", urlSafe)
	}
	got, err := c.playerTwoStream(t.Context(), s.srv.URL+"/"+urlSafe+"/2/bg.mp4")
	if err != nil || got != fbToken(2) {
		t.Fatalf("got %q, err %v", got, err)
	}
}

func TestBlogHosts(t *testing.T) {
	c := NewClient()
	for host, want := range map[string]bool{
		"bulborn.blogspot.com": true, "X.BLOGSPOT.COM": true,
		"blogspot.com": false, "blogspot.com.evil.example": false, "evilblogspot.com": false, "animesdigital.org": false, "": false,
	} {
		if got := c.isBlogHost(host); got != want {
			t.Errorf("isBlogHost(%q) = %v, want %v", host, got, want)
		}
	}
	test := NewClientForTest(http.DefaultClient, "http://127.0.0.1:9999")
	if !test.isBlogHost("127.0.0.1") || test.isBlogHost("127.0.0.2") {
		t.Fatal("a test client trusts its own server's host, and no other, beyond Blogger")
	}
	if odd := NewClientForTest(http.DefaultClient, "://nope"); odd.isBlogHost("anything") {
		t.Fatal("a base that is no URL adds no trusted host")
	}
}

// The caller may leave while a player is being asked: the answer is then that
// they left, not a fallback and not an episode declared offline.
func TestStreamStopsWhenTheCallerLeavesWhileAPlayerIsAsked(t *testing.T) {
	for name, path := range map[string]string{"player 1's playlist": fbPlaylist, "player 2's post": "/2026/10/cyberp-ed-2.html"} {
		t.Run(name, func(t *testing.T) {
			s, c := newFBSite(t)
			ctx, cancel := context.WithCancel(t.Context())
			defer cancel()
			s.set(func() {
				s.playlistStatus = 520
				s.onHit = func(hit string) {
					if hit == path {
						cancel()
					}
				}
			})
			got, err := c.GetEpisodeStreamURLContext(ctx, s.srv.URL+fbEpisode)
			if got != "" || !errors.Is(err, context.Canceled) {
				t.Fatalf("stream = %q, err = %v; want context.Canceled and nothing", got, err)
			}
			if errors.Is(err, netx.ErrMediaOffline) {
				t.Fatal("a caller who left is not an episode that is offline")
			}
			if path == fbPlaylist && s.hit("/2026/10/cyberp-ed-2.html") != 0 {
				t.Fatal("player 2 was tried after the caller left")
			}
		})
	}
}

func TestStreamReportsAnEpisodePageThatCannotBeRead(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, "gone", http.StatusNotFound)
	}))
	defer srv.Close()
	c := NewClientForTest(srv.Client(), srv.URL)
	got, err := c.GetEpisodeStreamURL(srv.URL + fbEpisode)
	if got != "" || err == nil || errors.Is(err, netx.ErrMediaOffline) {
		t.Fatalf("stream = %q, err = %v; want the page's own failure", got, err)
	}
}

// A player frame with no address, or one that is no URL, is skipped: the
// players around it are still read.
func TestStreamSkipsPlayerFramesItCannotRead(t *testing.T) {
	const playlist = "/stream/x/index.m3u8"
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == playlist {
			_, _ = w.Write([]byte("#EXTM3U\n"))
			return
		}
		_, _ = w.Write([]byte(`<html><body>
<iframe class="metaframe rptss"></iframe>
<iframe class="metaframe rptss" src="   "></iframe>
<iframe class="metaframe rptss" src="http://[::1"></iframe>
<iframe class="metaframe rptss" src="https://api.anivideo.net/videohls.php?d="></iframe>
<iframe class="metaframe rptss" src="https://api.anivideo.net/videohls.php?d=https%3A%2F%2Fcdn.example%2Fstream%2Fx%2Findex.m3u8"></iframe>
</body></html>`))
	}))
	defer srv.Close()
	c := NewClientForTest(&http.Client{Transport: rewriteHost{target: srv.URL}}, srv.URL)
	got, err := c.GetEpisodeStreamURL(srv.URL + fbEpisode)
	if err != nil || got != "https://cdn.example"+playlist {
		t.Fatalf("stream = %q, err = %v; want the one frame that names a playlist", got, err)
	}
}
