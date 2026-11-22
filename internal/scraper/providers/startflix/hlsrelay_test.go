package startflix

import (
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// The host that broke Oppenheimer and Jujutsu Kaisen (2026-11-21): a
// playlist at /pl/<id>, no extension, served as text/plain, every URI in it
// root-relative, refused without the embed's referer. These tests build it
// and hold the relay to handing the player something it takes for HLS that
// leads to the same segments.

const relayReferer = "https://stellarconductornexus.com/"

type unnamedHost struct {
	*httptest.Server
	hits   sync.Map // path -> *atomic.Int32
	master string
	media  string
	status int // answered to every playlist when non-zero
	// statusBody is what comes with status: by default an error text.
	statusBody string
	noReply    bool
}

func (h *unnamedHost) hit(path string) int {
	v, _ := h.hits.LoadOrStore(path, new(atomic.Int32))
	return int(v.(*atomic.Int32).Load())
}

func newUnnamedHost(t *testing.T) *unnamedHost {
	t.Helper()
	h := &unnamedHost{}
	h.media = "#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-TARGETDURATION:6\n" +
		"#EXT-X-KEY:METHOD=AES-128,URI=\"/pl/abc/key\"\n" +
		"#EXT-X-MAP:URI=\"init\"\n" +
		"#EXTINF:6,\n/pl/abc/seg0\n#EXTINF:6,\n/pl/abc/seg1\n#EXTINF:6,\nseg2\n#EXT-X-ENDLIST\n"
	h.master = "#EXTM3U\n" +
		"#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"a\",NAME=\"pt\",URI=\"/pl/audio\"\n" +
		"#EXT-X-STREAM-INF:BANDWIDTH=1000,RESOLUTION=1280x720,AUDIO=\"a\"\n/pl/abc\n" +
		"#EXT-X-STREAM-INF:BANDWIDTH=500\nhttps://other.example/named.m3u8\n" +
		"#EXT-X-I-FRAME-STREAM-INF:BANDWIDTH=90,URI=\"/pl/iframes\"\n"
	h.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		v, _ := h.hits.LoadOrStore(r.URL.Path, new(atomic.Int32))
		v.(*atomic.Int32).Add(1)
		if r.Header.Get("Referer") != relayReferer || r.Header.Get("User-Agent") == "" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}
		if h.status != 0 {
			body := h.statusBody
			if body == "" {
				body = "down"
			}
			w.WriteHeader(h.status)
			_, _ = io.WriteString(w, body)
			return
		}
		w.Header().Set("Content-Type", "text/plain") // what ffmpeg will not take for HLS
		switch r.URL.Path {
		case "/pl/master":
			_, _ = io.WriteString(w, h.master)
		case "/pl/abc", "/pl/audio", "/pl/iframes":
			if h.noReply {
				_, _ = io.WriteString(w, "<html>not a playlist</html>")
				return
			}
			_, _ = io.WriteString(w, h.media)
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(h.Close)
	return h
}

func getText(t *testing.T, target string) (int, http.Header, string) {
	t.Helper()
	req, err := http.NewRequestWithContext(t.Context(), http.MethodGet, target, http.NoBody)
	require.NoError(t, err)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	defer func() { _ = resp.Body.Close() }()
	body, err := io.ReadAll(resp.Body)
	require.NoError(t, err)
	return resp.StatusCode, resp.Header, string(body)
}

func TestHasPlaylistName(t *testing.T) {
	for raw, want := range map[string]bool{
		"https://h.example/a/master.m3u8":         true,
		"https://h.example/a/MASTER.M3U8?t=1&s=2": true,
		"https://h.example/list.m3u":              true,
		"https://h.example/pl/noDuNSK1xn":         false,
		"https://h.example/pl/x?file=a.m3u8":      false, // the query is not the name
		"https://h.example/v.mp4":                 false,
		"::not a url":                             false,
	} {
		assert.Equal(t, want, hasPlaylistName(raw), raw)
	}
}

func TestTheRelayServesAnUnnamedMediaPlaylistAsOneWithAbsoluteURIs(t *testing.T) {
	h := newUnnamedHost(t)
	r := newPlaylistRelay()
	local, err := r.publish(h.URL+"/pl/abc", relayReferer, h.Client())
	require.NoError(t, err)
	assert.True(t, strings.HasPrefix(local, "http://127.0.0.1:"), local)
	assert.True(t, hasPlaylistName(local), "named as a playlist: %s", local)
	assert.True(t, util.IsLocalProxyURL(local), "the downloader's loopback guard knows it")

	status, hdr, body := getText(t, local)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "application/vnd.apple.mpegurl", hdr.Get("Content-Type"), "a playlist's MIME type")
	assert.Contains(t, body, "\n"+h.URL+"/pl/abc/seg0\n", "root-relative segments resolved against the host")
	assert.Contains(t, body, "\n"+h.URL+"/pl/seg2\n", "relative ones too")
	assert.Contains(t, body, `URI="`+h.URL+`/pl/abc/key"`, "the key from the host")
	assert.Contains(t, body, `URI="`+h.URL+`/pl/init"`, "and the init section")
	assert.Contains(t, body, "#EXT-X-ENDLIST")
	assert.NotContains(t, body, "127.0.0.1:"+strings.Split(local, ":")[2][:0]+"/hls/", "a media playlist lists no playlist of ours")
	assert.Equal(t, 1, h.hit("/pl/abc"), "read once, with the host's referer")
}

func TestTheRelayServesTheRenditionsOfAnUnnamedMasterToo(t *testing.T) {
	h := newUnnamedHost(t)
	r := newPlaylistRelay()
	local, err := r.publish(h.URL+"/pl/master", relayReferer, h.Client())
	require.NoError(t, err)
	_, _, body := getText(t, local)

	lines := strings.Split(body, "\n")
	var rendition string
	for i, l := range lines {
		if strings.HasPrefix(l, "#EXT-X-STREAM-INF:BANDWIDTH=1000") {
			rendition = lines[i+1]
		}
	}
	require.NotEmpty(t, rendition)
	assert.True(t, strings.HasPrefix(rendition, r.base+"/hls/"), "an unnamed rendition is relayed: %s", rendition)
	assert.Contains(t, body, "\nhttps://other.example/named.m3u8\n", "a named one is left to the host")
	assert.Regexp(t, `#EXT-X-MEDIA:.*URI="`+r.base+`/hls/[0-9a-f]+\.m3u8"`, body, "an alternative audio playlist is relayed")
	assert.Regexp(t, `#EXT-X-I-FRAME-STREAM-INF:.*URI="`+r.base+`/hls/[0-9a-f]+\.m3u8"`, body)

	status, _, media := getText(t, rendition)
	require.Equal(t, http.StatusOK, status)
	assert.Contains(t, media, h.URL+"/pl/abc/seg0", "and the rendition leads to the host's segments")

	// Reading the master again does not grow the registry.
	before := r.lru.Len()
	_, _, again := getText(t, local)
	assert.Equal(t, body, again)
	assert.Equal(t, before, r.lru.Len(), "a playlist keeps its address")
}

func TestTheRelayTellsThePlayerWhenTheHostFails(t *testing.T) {
	for name, set := range map[string]func(h *unnamedHost){
		"the host answers 500":         func(h *unnamedHost) { h.status = http.StatusInternalServerError },
		"the host answers no playlist": func(h *unnamedHost) { h.noReply = true },
		"the host answers 404 with a playlist's body": func(h *unnamedHost) {
			h.status = http.StatusNotFound
			h.statusBody = h.media
		},
	} {
		t.Run(name, func(t *testing.T) {
			h := newUnnamedHost(t)
			set(h)
			r := newPlaylistRelay()
			local, err := r.publish(h.URL+"/pl/abc", relayReferer, h.Client())
			require.NoError(t, err)
			status, _, _ := getText(t, local)
			assert.Equal(t, http.StatusBadGateway, status)
		})
	}
	t.Run("the host refuses a request without its referer", func(t *testing.T) {
		h := newUnnamedHost(t)
		r := newPlaylistRelay()
		local, err := r.publish(h.URL+"/pl/abc", "", h.Client())
		require.NoError(t, err)
		status, _, _ := getText(t, local)
		assert.Equal(t, http.StatusBadGateway, status, "the referer is what the host checks; without it there is nothing to relay")
	})
	t.Run("a playlist over the size limit", func(t *testing.T) {
		big := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
			_, _ = io.WriteString(w, "#EXTM3U\n"+strings.Repeat("#EXTINF:1,\nx\n", maxRelayedBytes/12+10))
		}))
		t.Cleanup(big.Close)
		r := newPlaylistRelay()
		local, err := r.publish(big.URL+"/pl/big", "", big.Client())
		require.NoError(t, err)
		status, _, _ := getText(t, local)
		assert.Equal(t, http.StatusBadGateway, status)
	})
	t.Run("the host cannot be reached", func(t *testing.T) {
		gone := httptest.NewServer(http.NotFoundHandler())
		addr := gone.URL
		gone.Close()
		r := newPlaylistRelay()
		local, err := r.publish(addr+"/pl/x", "", http.DefaultClient)
		require.NoError(t, err)
		status, _, _ := getText(t, local)
		assert.Equal(t, http.StatusBadGateway, status)
	})
}

func TestTheRelayServesNothingElse(t *testing.T) {
	h := newUnnamedHost(t)
	r := newPlaylistRelay()
	local, err := r.publish(h.URL+"/pl/abc", relayReferer, h.Client())
	require.NoError(t, err)
	id := strings.TrimSuffix(local[strings.LastIndex(local, "/")+1:], ".m3u8")
	for name, target := range map[string]string{
		"an id never published": r.base + "/hls/" + strings.Repeat("0", 24) + ".m3u8",
		"the id with no name":   r.base + "/hls/" + id,
		"the id under a path":   r.base + "/hls/x/" + id + ".m3u8",
		"another prefix":        r.base + "/abyss/" + id + ".m3u8",
		"the root":              r.base + "/",
	} {
		status, _, _ := getText(t, target)
		assert.Equal(t, http.StatusNotFound, status, name)
	}
	req, _ := http.NewRequestWithContext(t.Context(), http.MethodPost, local, http.NoBody)
	resp, err := http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusMethodNotAllowed, resp.StatusCode)

	req, _ = http.NewRequestWithContext(t.Context(), http.MethodHead, local, http.NoBody)
	resp, err = http.DefaultClient.Do(req)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode, "a player may ask HEAD first")
	assert.Equal(t, "application/vnd.apple.mpegurl", resp.Header.Get("Content-Type"))
}

func TestTheRelayRefusesWhatIsNoAddress(t *testing.T) {
	r := newPlaylistRelay()
	for _, bad := range []string{"", "/pl/abc", "ftp://h.example/pl/x", "::"} {
		_, err := r.publish(bad, "", nil)
		assert.Error(t, err, bad)
	}
	assert.Empty(t, r.base, "nothing was started for them")
}

func TestTheRelayReportsThatItCannotListen(t *testing.T) {
	r := newPlaylistRelay()
	r.listen = func() (net.Listener, error) { return nil, errors.New("no port") }
	_, err := r.publish("https://h.example/pl/x", "", nil)
	assert.ErrorContains(t, err, "no port")
}

func TestTheRelayForgetsTheLeastRecentlyUsedPastItsBound(t *testing.T) {
	r := newPlaylistRelay()
	first, err := r.publish("https://h.example/pl/0", "", nil)
	require.NoError(t, err)
	second, err := r.publish("https://h.example/pl/1", "", nil)
	require.NoError(t, err)
	_, _ = r.lookup(strings.TrimSuffix(first[strings.LastIndex(first, "/")+1:], ".m3u8")) // in use
	for i := 2; i <= maxRelayedPlaylists; i++ {
		_, err := r.publish(fmt.Sprintf("https://h.example/pl/%d", i), "", nil)
		require.NoError(t, err)
	}
	assert.Equal(t, maxRelayedPlaylists, r.lru.Len())
	assert.Len(t, r.byUpstream, maxRelayedPlaylists, "the index forgets with the list")
	_, kept := r.lookup(strings.TrimSuffix(first[strings.LastIndex(first, "/")+1:], ".m3u8"))
	assert.True(t, kept, "the playlist in use stays")
	_, gone := r.lookup(strings.TrimSuffix(second[strings.LastIndex(second, "/")+1:], ".m3u8"))
	assert.False(t, gone, "the least recently used goes")
	again, err := r.publish("https://h.example/pl/1", "", nil)
	require.NoError(t, err)
	assert.NotEqual(t, second, again, "a forgotten playlist gets a new address")
}

func TestThePlayerIsHandedARelayedNameOnlyForAnUnnamedHLSStream(t *testing.T) {
	h := newUnnamedHost(t)
	relayed := nameThePlaylist(&Stream{URL: h.URL + "/pl/abc", Referer: relayReferer, Host: "nightjar", HLS: true, Height: 872}, h.Client())
	require.NotNil(t, relayed)
	assert.True(t, hasPlaylistName(relayed.URL), relayed.URL)
	assert.True(t, util.IsLocalProxyURL(relayed.URL))
	assert.Equal(t, relayReferer, relayed.Referer, "the segments still come from the host, with its referer")
	assert.Equal(t, 872, relayed.Height, "the rest of the stream is kept")
	status, _, body := getText(t, relayed.URL)
	assert.Equal(t, http.StatusOK, status)
	assert.Contains(t, body, h.URL+"/pl/abc/seg0")

	for name, s := range map[string]*Stream{
		"an HLS stream already named":  {URL: "https://h.example/master.m3u8", HLS: true},
		"a file that is not HLS":       {URL: "https://h.example/pl/video"},
		"an abyss file from the proxy": {URL: "http://127.0.0.1:9/abyss/x.mp4"},
		"an HLS stream at no address":  {URL: "::", HLS: true},
	} {
		got := nameThePlaylist(s, h.Client())
		assert.Same(t, s, got, name)
	}
	assert.Nil(t, nameThePlaylist(nil, nil))

	// The client that read the playlist first is the one the relay uses.
	own := &Stream{URL: h.URL + "/pl/audio", Referer: relayReferer, HLS: true, client: h.Client()} // not yet published
	viaOwn := nameThePlaylist(own, &http.Client{Transport: refuseAll{}})
	status, _, _ = getText(t, viaOwn.URL)
	assert.Equal(t, http.StatusOK, status, "read with the stream's own client, not the fallback that refuses everything")
}

// The whole path, as the panel serves it: a Legendado vidsrc player whose
// master sits at an address with no playlist name. ResolveStreamFor hands over
// the relay of it, the master read through the relay carries the host token,
// and so does every rendition the host listed without one.
func TestResolveStreamForRelaysAnUnnamedHLSStream(t *testing.T) {
	srvUpn := deadUpnFake(t)
	playerPage := fakeMovieConfig("SRV/data/api.php?type=movie&imdb=tt0315327", "tok1")
	f := &fakeVIDSRC{
		kind:         KindMovie,
		id:           "tt0315327",
		playerPages:  []string{playerPage},
		plainEnvJSON: `{"status_code":"200","data":{"stream_urls":["SRV/pl/cascade/master"]}}`,
	}
	srvVIDSRC := f.start(t)
	routeBothFamilies(t, hostRouter{byHost: map[string]*httptest.Server{
		"embedplayapiupn.upns.xyz": srvUpn,
		"vidsrc.sh":                srvVIDSRC,
	}})
	psrv := panelServer(t,
		`<button data-show-player="true" data-source="https://embedplayapiupn.upns.xyz/#8txjhj" data-type="iframe">#1 Dublado</button>`+
			`<button data-show-player="true" data-source="https://vidsrc.sh/embed/movie/tt0315327?ds_lang=pt" data-type="iframe">Legendado</button>`)
	c := NewClientForTest(psrv.Client(), "https://www.startflix.test")
	players, err := c.Players(t.Context(), psrv.URL+"/episodio/x")
	require.NoError(t, err)

	stream, err := c.ResolveStreamFor(t.Context(), players, AudioDubbed)
	require.NoError(t, err)
	assert.True(t, stream.HLS)
	require.True(t, hasPlaylistName(stream.URL), "handed over under a playlist name: %s", stream.URL)
	require.True(t, util.IsLocalProxyURL(stream.URL), stream.URL)
	assert.Equal(t, 1080, stream.Height, "the rest of the stream is what the resolver found")

	hitsBefore := f.masterHits
	status, hdr, body := getText(t, stream.URL)
	require.Equal(t, http.StatusOK, status)
	assert.Equal(t, "application/vnd.apple.mpegurl", hdr.Get("Content-Type"))
	assert.Equal(t, hitsBefore+1, f.masterHits, "the relay read the master, with the host token (the fake asserts it)")
	assert.True(t, strings.HasPrefix(body, "#EXTM3U"))
}

type refuseAll struct{}

func (refuseAll) RoundTrip(*http.Request) (*http.Response, error) {
	return nil, errors.New("this client is not to be used")
}

// The host itself (nightjarandnectar.space, 2026-11-21): the master is read
// with ?token=…, and its renditions and segments, listed root-relative with no
// query, are refused (403, a Cloudflare page) unless the same token comes with
// them. Its own player adds it to every request; the relay hands it on.
func TestTheRelayHandsTheTokenOnToWhatTheHostListsWithoutOne(t *testing.T) {
	const token = "token=abc%2B123"
	var refused atomic.Int32
	host := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.RawQuery != token {
			refused.Add(1)
			http.Error(w, "<!DOCTYPE html>cloudflare", http.StatusForbidden)
			return
		}
		w.Header().Set("Content-Type", "text/plain")
		switch r.URL.Path {
		case "/pl/master":
			_, _ = io.WriteString(w, "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1\n/pl/master/rendition\n")
		case "/pl/master/rendition":
			_, _ = io.WriteString(w, "#EXTM3U\n#EXT-X-KEY:METHOD=AES-128,URI=\"/content/k\"\n#EXTINF:5,\n/content/x/page-0.html\n"+
				"#EXTINF:5,\nhttps://other-cdn.example/seg1.ts\n#EXTINF:5,\n/content/x/page-2.html?sig=own\n#EXT-X-ENDLIST\n")
		default:
			_, _ = io.WriteString(w, "segment bytes")
		}
	}))
	t.Cleanup(host.Close)

	r := newPlaylistRelay()
	local, err := r.publish(host.URL+"/pl/master?"+token, "", host.Client())
	require.NoError(t, err)
	status, _, master := getText(t, local)
	require.Equal(t, http.StatusOK, status)
	var rendition string
	for _, l := range strings.Split(master, "\n") {
		if strings.HasPrefix(l, "http") {
			rendition = l
		}
	}
	require.True(t, strings.HasPrefix(rendition, r.base+"/hls/"), "the rendition is relayed: %s", rendition)
	p, ok := r.lookup(strings.TrimSuffix(rendition[strings.LastIndex(rendition, "/")+1:], ".m3u8"))
	require.True(t, ok)
	assert.Equal(t, host.URL+"/pl/master/rendition?"+token, p.upstream, "and read with the master's token")

	status, _, media := getText(t, rendition)
	require.Equal(t, http.StatusOK, status, "the host answered the relay, token and all")
	assert.Contains(t, media, "\n"+host.URL+"/content/x/page-0.html?"+token+"\n", "a segment of the host gets the token")
	assert.Contains(t, media, `URI="`+host.URL+`/content/k?`+token+`"`, "and so does its key")
	assert.Contains(t, media, "\nhttps://other-cdn.example/seg1.ts\n", "another host's segment is not handed the token")
	assert.Contains(t, media, "\n"+host.URL+"/content/x/page-2.html?sig=own\n", "nor is one that carries a query of its own")
	assert.Zero(t, refused.Load(), "nothing was ever asked of the host without its token")

	// And the segment, asked for the way the player will, is served.
	seg := host.URL + "/content/x/page-0.html?" + token
	resp, err := host.Client().Get(seg)
	require.NoError(t, err)
	_ = resp.Body.Close()
	assert.Equal(t, http.StatusOK, resp.StatusCode)
}

func TestTheRelayHandsNoQueryOnWhenThePlaylistHadNone(t *testing.T) {
	h := newUnnamedHost(t)
	r := newPlaylistRelay()
	local, err := r.publish(h.URL+"/pl/abc", relayReferer, h.Client())
	require.NoError(t, err)
	_, _, body := getText(t, local)
	assert.Contains(t, body, "\n"+h.URL+"/pl/abc/seg0\n", "no stray ? on a segment")
}

func TestTheRelayReportsAPlaylistItCannotRead(t *testing.T) {
	r := newPlaylistRelay()
	req := httptest.NewRequest(http.MethodGet, "/hls/x.m3u8", http.NoBody)

	_, err := (&relayedPlaylist{upstream: "http://bad host/x", client: http.DefaultClient}).read(req)
	assert.Error(t, err, "an address no request can be made for")

	// A host that cuts the playlist short.
	cut := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		conn, buf, err := w.(http.Hijacker).Hijack()
		if err != nil {
			return
		}
		_, _ = fmt.Fprint(buf, "HTTP/1.1 200 OK\r\nContent-Length: 1000\r\n\r\n#EXTM3U\n#EXTINF")
		_ = buf.Flush()
		_ = conn.Close()
	}))
	t.Cleanup(cut.Close)
	_, err = (&relayedPlaylist{upstream: cut.URL + "/pl/x", client: cut.Client()}).read(req)
	assert.Error(t, err, "half a playlist is not one")
	_ = r
}

func TestTheRelayWithNoClientOfItsOwnKeepsTheSSRFGuard(t *testing.T) {
	h := newUnnamedHost(t) // on loopback, which the scraper transport refuses
	r := newPlaylistRelay()
	local, err := r.publish(h.URL+"/pl/abc", relayReferer, nil)
	require.NoError(t, err)
	status, _, _ := getText(t, local)
	assert.Equal(t, http.StatusBadGateway, status, "the default client is the SSRF-guarded one")
	assert.Zero(t, h.hit("/pl/abc"), "so the loopback host was never asked")
}

func TestTheRelayLeavesWhatItCannotResolveOrRelay(t *testing.T) {
	host := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1\nftp://h.example/rendition\n#EXTINF:5,\n%zz\n")
	}))
	t.Cleanup(host.Close)
	r := newPlaylistRelay()
	local, err := r.publish(host.URL+"/pl/x", "", host.Client())
	require.NoError(t, err)
	_, _, body := getText(t, local)
	assert.Contains(t, body, "\nftp://h.example/rendition\n", "a rendition that cannot be relayed is left as the host wrote it")
	assert.Contains(t, body, "\n%zz\n", "and so is a line that is no URL")
}

func TestTheRelayReadsACRLFPlaylist(t *testing.T) {
	host := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "#EXTM3U\r\n\r\n#EXTINF:5,\r\nseg0\r\n#EXT-X-ENDLIST\r\n")
	}))
	t.Cleanup(host.Close)
	r := newPlaylistRelay()
	local, err := r.publish(host.URL+"/pl/x", "", host.Client())
	require.NoError(t, err)
	_, _, body := getText(t, local)
	assert.Contains(t, body, "\n"+host.URL+"/pl/seg0\n", "a segment line ending in CR is still a URL")
	assert.NotContains(t, body, "\r")
}

// A rendition's address is the one line after its #EXT-X-STREAM-INF: a URI
// further down with no tag of its own is not taken for another playlist.
func TestOnlyTheLineAfterAStreamInfIsARendition(t *testing.T) {
	host := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = io.WriteString(w, "#EXTM3U\n#EXT-X-STREAM-INF:BANDWIDTH=1\n/pl/rendition\n/pl/stray\n")
	}))
	t.Cleanup(host.Close)
	r := newPlaylistRelay()
	local, err := r.publish(host.URL+"/pl/x", "", host.Client())
	require.NoError(t, err)
	_, _, body := getText(t, local)
	assert.Contains(t, body, "\n"+host.URL+"/pl/stray", "the stray line is resolved, not relayed")
	assert.Equal(t, 2, r.lru.Len(), "the master and its one rendition")
}
