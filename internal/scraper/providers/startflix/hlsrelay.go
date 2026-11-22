package startflix

import (
	"container/list"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"io"
	"net"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/scraper/netx"
	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// Some hosts serve an HLS playlist at an address with no playlist extension
// and a MIME type that is not a playlist's (nightjarandnectar.space/pl/<id>,
// 2026-11-21, Oppenheimer and Jujutsu Kaisen). FFmpeg will not take that for
// HLS ("Not detecting m3u8/hls with non standard extension and non standard
// mime type"), so mpv read it as a plain list of files, could not open the
// root-relative paths in it, and quit in a third of a second; the downloader
// did not know it was HLS either.
//
// The resolvers that know a stream is HLS say so (Stream.HLS). For such a
// stream at an unnamed address, the player is handed a loopback address
// named .m3u8 instead: each request for it reads the host's playlist with the
// referer it wants and answers it with every URI absolute. Playlists listed
// in it (a master's renditions) are served the same way; segments and keys
// are fetched from the host directly, as before.

const (
	// maxRelayedPlaylists bounds the registry; least recently used go first.
	maxRelayedPlaylists = 512
	// maxRelayedBytes bounds a playlist read from a host.
	maxRelayedBytes = 8 << 20
	// relayTimeout bounds one read of a host's playlist.
	relayTimeout = 30 * time.Second
)

var relayURIAttrRe = regexp.MustCompile(`URI="([^"]*)"`)

type playlistRelay struct {
	mu         sync.Mutex
	base       string
	lists      map[string]*list.Element // id -> element; values are *relayedPlaylist
	lru        *list.List
	byUpstream map[string]string // upstream + "\x00" + referer -> id
	listen     func() (net.Listener, error)
}

type relayedPlaylist struct {
	id, upstream, referer string
	client                *http.Client
}

func newPlaylistRelay() *playlistRelay {
	return &playlistRelay{lists: map[string]*list.Element{}, lru: list.New(), byUpstream: map[string]string{}}
}

var (
	relayOnce   sync.Once
	relayShared *playlistRelay
)

func sharedPlaylistRelay() *playlistRelay {
	relayOnce.Do(func() { relayShared = newPlaylistRelay() })
	return relayShared
}

// hasPlaylistName reports whether the path of raw ends as a playlist's does,
// which is all FFmpeg looks at when the MIME type says nothing.
func hasPlaylistName(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	p := strings.ToLower(u.Path)
	return strings.HasSuffix(p, ".m3u8") || strings.HasSuffix(p, ".m3u")
}

// nameThePlaylist returns s with a URL the player takes for HLS, relaying it
// from loopback when the host's own address does not say so. The playlist is
// read with the client that read it first (Stream.client), or else with
// fallback. A stream that is not HLS, or already named as a playlist, is
// returned as it is, and so is one that cannot be relayed: it may still play
// where the host sends a playlist's MIME type.
func nameThePlaylist(s *Stream, fallback *http.Client) *Stream {
	if s == nil || !s.HLS || hasPlaylistName(s.URL) {
		return s
	}
	client := s.client
	if client == nil {
		client = fallback
	}
	local, err := sharedPlaylistRelay().publish(s.URL, s.Referer, client)
	if err != nil {
		util.Debug("StartFlix playlist is played from its own address", "host", s.Host, "err", err)
		return s
	}
	out := *s
	out.URL = local
	util.Debug("StartFlix playlist relayed under a playlist name", "host", s.Host)
	return &out
}

// publish registers upstream and returns its loopback address, starting the
// listener on first use. The same playlist with the same referer keeps one
// address, so a master read twice does not grow the registry.
func (r *playlistRelay) publish(upstream, referer string, client *http.Client) (string, error) {
	u, err := url.Parse(upstream)
	if err != nil || (u.Scheme != "http" && u.Scheme != "https") || u.Host == "" {
		return "", fmt.Errorf("not a playlist address: %q", upstream)
	}
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.base == "" {
		listen := r.listen
		if listen == nil {
			listen = func() (net.Listener, error) { return net.Listen("tcp", "127.0.0.1:0") }
		}
		ln, err := listen()
		if err != nil {
			return "", fmt.Errorf("local playlist relay: %w", err)
		}
		srv := &http.Server{Handler: r, ReadHeaderTimeout: 10 * time.Second}
		go func() { _ = srv.Serve(ln) }() // for the life of the process
		r.base = "http://" + ln.Addr().String()
		util.RegisterLocalProxyHost(ln.Addr().String())
		util.Debug("StartFlix playlist relay listening", "addr", r.base)
	}

	key := upstream + "\x00" + referer
	if id, ok := r.byUpstream[key]; ok {
		if el, ok := r.lists[id]; ok {
			r.lru.MoveToFront(el)
			return r.base + "/hls/" + id + ".m3u8", nil
		}
	}
	var raw [12]byte
	_, _ = rand.Read(raw[:]) // never fails since Go 1.24
	id := hex.EncodeToString(raw[:])
	r.lists[id] = r.lru.PushFront(&relayedPlaylist{id: id, upstream: upstream, referer: referer, client: client})
	r.byUpstream[key] = id
	for r.lru.Len() > maxRelayedPlaylists {
		oldest := r.lru.Back()
		r.lru.Remove(oldest)
		gone := oldest.Value.(*relayedPlaylist)
		delete(r.lists, gone.id)
		delete(r.byUpstream, gone.upstream+"\x00"+gone.referer)
	}
	return r.base + "/hls/" + id + ".m3u8", nil
}

func (r *playlistRelay) lookup(id string) (*relayedPlaylist, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()
	el, ok := r.lists[id]
	if !ok {
		return nil, false
	}
	r.lru.MoveToFront(el)
	return el.Value.(*relayedPlaylist), true
}

// ServeHTTP answers GET or HEAD /hls/<id>.m3u8 with the host's playlist, read
// now, every URI in it absolute.
func (r *playlistRelay) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet && req.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// An id is 24 hex digits: a path that is not /hls/<id>.m3u8 names none,
	// and is not found.
	id, ok := strings.CutSuffix(strings.TrimPrefix(req.URL.Path, "/hls/"), ".m3u8")
	if !ok {
		http.NotFound(w, req)
		return
	}
	p, found := r.lookup(id)
	if !found {
		http.NotFound(w, req)
		return
	}
	body, err := p.read(req)
	if err != nil {
		if req.Context().Err() == nil {
			util.Debug("StartFlix relayed playlist unavailable", "err", err)
		}
		http.Error(w, "playlist unavailable", http.StatusBadGateway)
		return
	}
	text := r.rewrite(body, p)
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Length", fmt.Sprint(len(text)))
	// net/http sends no body for a HEAD.
	_, _ = io.WriteString(w, text) // #nosec G705 -- an HLS playlist for the local player, served as one, never as HTML
}

// read fetches the host's playlist the way its player does.
func (p *relayedPlaylist) read(req *http.Request) (string, error) {
	ctx := req.Context()
	up, err := http.NewRequestWithContext(ctx, http.MethodGet, p.upstream, http.NoBody)
	if err != nil {
		return "", err
	}
	up.Header.Set("User-Agent", netx.UserAgent)
	up.Header.Set("Accept", "*/*")
	if p.referer != "" {
		up.Header.Set("Referer", p.referer)
		if o, err := url.Parse(p.referer); err == nil && o.Host != "" {
			up.Header.Set("Origin", o.Scheme+"://"+o.Host)
		}
	}
	client := p.client
	if client == nil {
		client = &http.Client{Transport: netx.SafeScraperTransport(relayTimeout)}
	}
	resp, err := client.Do(up) // #nosec G704 -- an address a resolver handed over, on the SSRF-guarded transport
	if err != nil {
		return "", err
	}
	defer func() { _ = resp.Body.Close() }()
	if resp.StatusCode != http.StatusOK {
		return "", fmt.Errorf("HTTP %d", resp.StatusCode)
	}
	raw, err := io.ReadAll(io.LimitReader(resp.Body, maxRelayedBytes+1))
	if err != nil {
		return "", err
	}
	if len(raw) > maxRelayedBytes {
		return "", errors.New("playlist exceeds size limit")
	}
	text := strings.TrimLeft(string(raw), "\uFEFF \t\r\n")
	if !strings.HasPrefix(text, "#EXTM3U") {
		return "", errors.New("the host did not answer with a playlist")
	}
	return text, nil
}

// rewrite resolves every URI of a playlist against the host's address.
// Playlists it lists (a master's renditions and alternative tracks) are
// relayed too, since their addresses are as unnamed as the master's; media
// segments, keys and init sections are fetched from the host.
//
// A playlist read with a query (the host's token) hands that query on to the
// URIs it lists on its own host that carry none: the host refuses the
// renditions and segments without it, and its own player adds it to every
// request (nightjarandnectar.space, 2026-11-21: master 200, rendition and
// segment 403 without the token, 200 with it). A URI with a query of its own,
// or on another host, is left as it is.
func (r *playlistRelay) rewrite(body string, p *relayedPlaylist) string {
	base, _ := url.Parse(p.upstream) // publish refused what does not parse
	resolve := func(ref string) string {
		u, err := url.Parse(strings.TrimSpace(ref))
		if err != nil {
			return ref
		}
		abs := base.ResolveReference(u)
		if abs.RawQuery == "" && base.RawQuery != "" && strings.EqualFold(abs.Host, base.Host) {
			abs.RawQuery = base.RawQuery
		}
		return abs.String()
	}
	relay := func(ref string) string {
		abs := resolve(ref)
		if hasPlaylistName(abs) {
			return abs
		}
		local, err := r.publish(abs, p.referer, p.client)
		if err != nil {
			return abs
		}
		return local
	}

	lines := strings.Split(strings.ReplaceAll(body, "\r\n", "\n"), "\n")
	nextIsPlaylist := false
	for i, line := range lines {
		trimmed := strings.TrimSpace(line)
		switch {
		case trimmed == "":
		case strings.HasPrefix(trimmed, "#EXT-X-STREAM-INF"):
			nextIsPlaylist = true
		case strings.HasPrefix(trimmed, "#EXT-X-MEDIA:"), strings.HasPrefix(trimmed, "#EXT-X-I-FRAME-STREAM-INF:"):
			lines[i] = relayURIAttrRe.ReplaceAllStringFunc(trimmed, func(attr string) string {
				return `URI="` + relay(attr[len(`URI="`):len(attr)-1]) + `"`
			})
		case strings.HasPrefix(trimmed, "#"):
			lines[i] = relayURIAttrRe.ReplaceAllStringFunc(trimmed, func(attr string) string {
				return `URI="` + resolve(attr[len(`URI="`):len(attr)-1]) + `"`
			})
		case nextIsPlaylist:
			lines[i] = relay(trimmed)
			nextIsPlaylist = false
		default:
			lines[i] = resolve(trimmed)
		}
	}
	return strings.Join(lines, "\n")
}
