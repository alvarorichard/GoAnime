package plutotv

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// fetchLocal reads a URL the playlist server handed out.
func fetchLocal(t *testing.T, method, target string) (status int, contentType, body string) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), method, target, http.NoBody)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	raw, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, resp.Header.Get("Content-Type"), string(raw)
}

const samplePlaylist = "#EXTM3U\n#EXTINF:5,\nhttps://cdn/a.ts\n#EXT-X-ENDLIST\n"

func TestPlaylistServerServesOnlyWhatWasPublished(t *testing.T) {
	s := newPlaylistServer()
	target, err := s.publish(samplePlaylist)
	require.NoError(t, err)
	require.True(t, strings.HasPrefix(target, "http://127.0.0.1:"), target)
	require.True(t, strings.HasSuffix(target, ".m3u8"), "named as a playlist, which is how the player and the downloader tell")
	assert.True(t, util.IsLocalProxyURL(target), "registered, so the download paths let it through their loopback guard")

	status, ctype, body := fetchLocal(t, http.MethodGet, target)
	assert.Equal(t, http.StatusOK, status)
	assert.Equal(t, "application/vnd.apple.mpegurl", ctype)
	assert.Equal(t, samplePlaylist, body)

	other, err := s.publish(samplePlaylist)
	require.NoError(t, err)
	assert.NotEqual(t, target, other, "each publication has an address of its own, which cannot be guessed")
	assert.Equal(t, s.base, target[:len(s.base)], "one listener serves them all")

	id := strings.TrimSuffix(strings.TrimPrefix(target, s.base+"/plutotv/"), ".m3u8")
	require.Len(t, id, 32)
	for name, u := range map[string]string{
		"an id that was never published": s.base + "/plutotv/" + strings.Repeat("0", 32) + ".m3u8",
		"the id without the extension":   s.base + "/plutotv/" + id,
		"the id under another path":      s.base + "/other/" + id + ".m3u8",
		"the root":                       s.base + "/",
		"a path above":                   s.base + "/plutotv/../" + id + ".m3u8",
	} {
		status, _, body := fetchLocal(t, http.MethodGet, u)
		assert.Equal(t, http.StatusNotFound, status, name)
		assert.NotContains(t, body, "EXTM3U", name)
	}
	status, _, body = fetchLocal(t, http.MethodPost, target)
	assert.Equal(t, http.StatusMethodNotAllowed, status)
	assert.NotContains(t, body, "EXTM3U")
	status, _, body = fetchLocal(t, http.MethodHead, target)
	assert.Equal(t, http.StatusOK, status)
	assert.Empty(t, body)
}

func TestPlaylistServerRefusesWhatIsNotAPlaylist(t *testing.T) {
	s := newPlaylistServer()
	for _, text := range []string{"", "<html>", "https://cdn/x.m3u8", "\n\n  ", "x#EXTM3U"} {
		_, err := s.publish(text)
		require.Error(t, err, "%q", text)
	}
	assert.Empty(t, s.base, "nothing listens until there is something to serve")
	assert.Zero(t, s.recent.Len())
}

// A playlist that opens with a byte-order mark or blank lines is still one:
// the client accepts it, so the server must too, or the player would be handed
// the CDN's own URL, which is the one a CDN refuses.
func TestPlaylistServerServesAPlaylistThatOpensWithBlankLinesOrABOM(t *testing.T) {
	s := newPlaylistServer()
	for name, lead := range map[string]string{"blank lines": "\r\n\n", "spaces": "  \t", "a byte-order mark": "\ufeff", "both": "\ufeff\n "} {
		target, err := s.publish(lead + samplePlaylist)
		require.NoError(t, err, name)
		_, _, body := fetchLocal(t, http.MethodGet, target)
		assert.Equal(t, samplePlaylist, body, "%s: what is served opens with #EXTM3U, as a player expects", name)
	}
}

func TestPlaylistServerForgetsTheLeastRecentlyUsedPastItsBound(t *testing.T) {
	s := newPlaylistServer()
	var urls []string
	for i := range maxPlaylists {
		u, err := s.publish(fmt.Sprintf("#EXTM3U\n# %d\n", i))
		require.NoError(t, err)
		urls = append(urls, u)
	}
	// The oldest is asked for: it is in use, so it is not the one to forget.
	status, _, _ := fetchLocal(t, http.MethodGet, urls[0])
	require.Equal(t, http.StatusOK, status)

	for i := range 2 {
		u, err := s.publish(fmt.Sprintf("#EXTM3U\n# new %d\n", i))
		require.NoError(t, err)
		urls = append(urls, u)
	}
	assert.Len(t, s.lists, maxPlaylists)
	assert.Equal(t, maxPlaylists, s.recent.Len())

	want := func(i int) int {
		if i == 1 || i == 2 { // the two least recently used
			return http.StatusNotFound
		}
		return http.StatusOK
	}
	for _, i := range []int{0, 1, 2, 3, maxPlaylists - 1, maxPlaylists, maxPlaylists + 1} {
		status, _, _ := fetchLocal(t, http.MethodGet, urls[i])
		assert.Equal(t, want(i), status, "playlist %d", i)
	}
}

func TestPlaylistServerIsSafeToPublishToFromManyResolutions(t *testing.T) {
	s := newPlaylistServer()
	var wg sync.WaitGroup
	seen := make(chan string, 40)
	for range 40 {
		wg.Go(func() {
			u, err := s.publish("#EXTM3U\n")
			if assert.NoError(t, err) {
				seen <- u
			}
		})
	}
	wg.Wait()
	close(seen)
	unique := map[string]bool{}
	for u := range seen {
		unique[u] = true
		assert.True(t, strings.HasPrefix(u, s.base+"/plutotv/"), "one listener, started once")
	}
	assert.Len(t, unique, 40)
}

func TestPlaylistServerReportsThatItCannotListen(t *testing.T) {
	s := newPlaylistServer()
	s.listen = func() (net.Listener, error) { return nil, errors.New("no free port") }
	_, err := s.publish("#EXTM3U\n")
	require.ErrorContains(t, err, "local playlist server")
	require.ErrorContains(t, err, "no free port")
	assert.Empty(t, s.lists, "nothing is kept for an address that does not exist")

	s.listen = nil
	target, err := s.publish("#EXTM3U\n")
	require.NoError(t, err, "and it recovers once a port can be had")
	_, _, body := fetchLocal(t, http.MethodGet, target)
	assert.Equal(t, "#EXTM3U\n", body)
}

// ---- PlayURL ---------------------------------------------------------------

func TestPlayURLIsTheLoopbackAddressOfThePlaylist(t *testing.T) {
	got := PlayURL(Stream{URL: "https://cdn.example/clip/hls/0-end/hls_2400.m3u8", Playlist: samplePlaylist, AdFree: true})
	require.True(t, util.IsLocalProxyURL(got), "not the CDN's own URL, which one of the CDNs refuses: %s", got)
	_, _, body := fetchLocal(t, http.MethodGet, got)
	assert.Equal(t, samplePlaylist, body)
}

func TestPlayURLIsTheStreamsOwnURLWhenThereIsNoPlaylist(t *testing.T) {
	const master = "https://stitcher.example/master.m3u8?jwt=x"
	assert.Equal(t, master, PlayURL(Stream{URL: master}), "a title with no playlist of its own plays the stitched master")
	assert.Empty(t, PlayURL(Stream{}))
}

func TestPlayURLFallsBackToTheStreamsURLWhenNothingCanBeServed(t *testing.T) {
	prev := publishFn
	t.Cleanup(func() { publishFn = prev })
	var asked []string
	publishFn = func(text string) (string, error) {
		asked = append(asked, text)
		return "", errors.New("no port to listen on")
	}
	const own = "https://cdn.example/clip/hls/0-end/hls_2400.m3u8"
	assert.Equal(t, own, PlayURL(Stream{URL: own, Playlist: samplePlaylist, AdFree: true}), "it still plays wherever the CDN answers")
	assert.Equal(t, []string{samplePlaylist}, asked)

	// A playlist that is not one cannot be served either.
	publishFn = prev
	assert.Equal(t, own, PlayURL(Stream{URL: own, Playlist: "<html>"}))
}

func TestTheRealListenerIsOnLoopback(t *testing.T) {
	ln, err := listenLoopback()
	require.NoError(t, err)
	defer func() { _ = ln.Close() }()
	addr, ok := ln.Addr().(*net.TCPAddr)
	require.True(t, ok)
	assert.True(t, addr.IP.IsLoopback(), "never an address another machine can reach: %s", addr)
	assert.NotZero(t, addr.Port)
}
