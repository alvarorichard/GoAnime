package plutotv

import (
	"bytes"
	"container/list"
	"context"
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// Pluto TV's media host is several CDNs behind one name, and one of them
// refuses every playlist under a clip while serving the segments the playlist
// lists. A player handed the playlist's URL therefore opens it or fails by the
// DNS answer it happens to get, and can fail again on a later request for it.
// So the player is not handed that URL: Stream reads the playlist (or
// assembles it), and this server hands the text to the player from loopback.
// The segments and keys in it are absolute and are fetched from the CDN
// directly; only the playlist's own text passes through here.

// maxPlaylists bounds what the server remembers. A playlist that is fetched
// stays: the bound drops the ones least recently asked for, so an episode that
// is playing, or queued in a long batch, is not the one forgotten.
const maxPlaylists = 256

// playlistServer serves resolved playlists on 127.0.0.1.
type playlistServer struct {
	mu    sync.Mutex
	base  string                   // "http://127.0.0.1:port", set once the listener is up
	lists map[string]*list.Element // id -> its place in recent
	// recent orders the playlists, most recently published or served first.
	recent *list.List
	// listen opens the loopback listener; a seam, nil meaning the real one.
	listen func() (net.Listener, error)
}

type servedPlaylist struct {
	id, text string
	// plan is set for a title whose segments are served from here (see
	// rebase.go); text is then its playlist.
	plan  *rebasePlan
	keys  keyCache
	cache segmentCache
}

func newPlaylistServer() *playlistServer {
	return &playlistServer{lists: make(map[string]*list.Element), recent: list.New()}
}

var playlists = newPlaylistServer()

// publishFn publishes a playlist and returns its local URL; a seam.
var publishFn = playlists.publish

// publishPlanFn publishes a rebased title and returns its local URL; a seam.
var publishPlanFn = playlists.publishPlan

// listenLoopback asks the system for a free port on 127.0.0.1.
func listenLoopback() (net.Listener, error) { return net.Listen("tcp", "127.0.0.1:0") }

// PlayURL is the address a player, or a downloader, should be given for s: the
// loopback address of its playlist when Stream read one, and s.URL otherwise
// (a title with no playlist of its own, or no port to serve one from, in which
// case the URL still plays wherever the CDN answers).
func PlayURL(s Stream) string {
	if s.rebase != nil {
		local, err := publishPlanFn(s.rebase)
		if err == nil {
			return local
		}
		util.Debug("Pluto TV title is played from the CDN's own timeline", "error", err)
	}
	// A stream with no playlist has nothing to publish: publishing refuses
	// what is not one, and the URL is what is left.
	local, err := publishFn(s.Playlist)
	if err != nil {
		if s.Playlist != "" {
			util.Debug("Pluto TV playlist is played from its own URL", "error", err)
		}
		return s.URL
	}
	return local
}

// publish registers playlist and returns the URL it is served at, starting
// the listener on first use.
func (s *playlistServer) publish(playlist string) (string, error) {
	// A playlist may open with a byte-order mark or blank lines; what follows
	// is what a player must see first.
	playlist = strings.TrimLeft(playlist, "\ufeff \t\r\n")
	if !strings.HasPrefix(playlist, "#EXTM3U") {
		return "", errors.New("not a playlist")
	}
	return s.register(func(string) string { return playlist }, nil)
}

// publishPlan registers a title whose segments are served from here.
func (s *playlistServer) publishPlan(plan *rebasePlan) (string, error) {
	if plan == nil || len(plan.segs) == 0 {
		return "", errors.New("no segments")
	}
	return s.register(plan.render, plan)
}

func (s *playlistServer) register(render func(id string) string, plan *rebasePlan) (string, error) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.base == "" {
		listen := s.listen
		if listen == nil {
			listen = listenLoopback
		}
		ln, err := listen()
		if err != nil {
			return "", fmt.Errorf("local playlist server: %w", err)
		}
		srv := &http.Server{Handler: s, ReadHeaderTimeout: 10 * time.Second}
		go func() { _ = srv.Serve(ln) }() // for the life of the process
		s.base = "http://" + ln.Addr().String()
		// The download paths refuse loopback unless it is one of GoAnime's own.
		util.RegisterLocalProxyHost(ln.Addr().String())
		util.Debug("Pluto TV playlist server listening", "addr", s.base)
	}

	var raw [16]byte
	_, _ = rand.Read(raw[:]) // never fails since Go 1.24
	id := hex.EncodeToString(raw[:])
	s.lists[id] = s.recent.PushFront(&servedPlaylist{id: id, text: render(id), plan: plan})
	for s.recent.Len() > maxPlaylists {
		oldest := s.recent.Back()
		s.recent.Remove(oldest)
		delete(s.lists, oldest.Value.(*servedPlaylist).id)
	}
	return s.base + "/plutotv/" + id + ".m3u8", nil
}

// ServeHTTP answers a GET or HEAD for a published playlist and nothing else.
func (s *playlistServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	rest := strings.TrimPrefix(r.URL.Path, "/plutotv/")
	if id, n, ok := strings.Cut(rest, "/"); ok {
		s.serveSegment(w, r, id, n)
		return
	}
	// An id is 32 hex digits, so a path that is not /plutotv/<id>.m3u8 leaves
	// something that is no id and is simply not found.
	id, isPlaylist := strings.CutSuffix(rest, ".m3u8")
	entry, found := s.lookup(id)
	if !isPlaylist || !found {
		http.NotFound(w, r)
		return
	}
	playlist := entry.text
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Length", fmt.Sprint(len(playlist)))
	_, _ = w.Write([]byte(playlist)) // net/http sends no body for a HEAD
}

// lookup finds a published playlist and marks it as in use.
func (s *playlistServer) lookup(id string) (*servedPlaylist, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()
	el, ok := s.lists[id]
	if !ok {
		return nil, false
	}
	s.recent.MoveToFront(el) // in use: not the one to forget
	return el.Value.(*servedPlaylist), true
}

// serveSegment answers /plutotv/<id>/<n>.ts for a title served from here.
func (s *playlistServer) serveSegment(w http.ResponseWriter, r *http.Request, id, name string) {
	num, isSegment := strings.CutSuffix(name, ".ts")
	n, err := strconv.Atoi(num)
	entry, found := s.lookup(id)
	if !isSegment || err != nil || n < 0 || !found || entry.plan == nil || n >= len(entry.plan.segs) {
		http.NotFound(w, r)
		return
	}
	data, err := entry.segment(r.Context(), n)
	if err != nil {
		if r.Context().Err() == nil {
			util.Debug("Pluto TV segment unavailable", "segment", n, "error", err)
		}
		http.Error(w, "segment unavailable", http.StatusBadGateway)
		return
	}
	w.Header().Set("Content-Type", "video/mp2t")
	w.Header().Set("Cache-Control", "no-store")
	http.ServeContent(w, r, "", time.Time{}, bytes.NewReader(data)) // ranges and HEAD
}

// segmentCache keeps the few segments around the one being played, and has a
// segment that two requests ask for at once fetched once.
type segmentCache struct {
	mu    sync.Mutex
	items map[int]*segmentResult
}

type segmentResult struct {
	done chan struct{}
	data []byte
	err  error
}

// segment returns segment n, and starts fetching the ones after it.
func (e *servedPlaylist) segment(ctx context.Context, n int) ([]byte, error) {
	res := e.cache.start(n, func() ([]byte, error) {
		// Not the request's context: a prefetch outlives it, and a request
		// that gives up must not poison what another is waiting for.
		bg, cancel := context.WithTimeout(context.WithoutCancel(ctx), 2*time.Minute)
		defer cancel()
		return e.plan.segment(bg, n, &e.keys)
	})
	for k := 1; k <= segmentPrefetch && n+k < len(e.plan.segs); k++ {
		k := k
		e.cache.start(n+k, func() ([]byte, error) {
			bg, cancel := context.WithTimeout(context.Background(), 2*time.Minute)
			defer cancel()
			return e.plan.segment(bg, n+k, &e.keys)
		})
	}
	e.cache.trim(n)
	select {
	case <-res.done:
		return res.data, res.err
	case <-ctx.Done():
		return nil, ctx.Err()
	}
}

// start begins fetching segment n unless that is under way or done, and
// returns its result.
func (c *segmentCache) start(n int, fetch func() ([]byte, error)) *segmentResult {
	c.mu.Lock()
	defer c.mu.Unlock()
	if res, ok := c.items[n]; ok {
		return res
	}
	if c.items == nil {
		c.items = make(map[int]*segmentResult)
	}
	res := &segmentResult{done: make(chan struct{})}
	c.items[n] = res
	go func() {
		res.data, res.err = fetch()
		close(res.done)
		if res.err != nil { // a failure is not remembered: the next ask retries
			c.mu.Lock()
			if c.items[n] == res {
				delete(c.items, n)
			}
			c.mu.Unlock()
		}
	}()
	return res
}

// trim forgets the segments far from n, behind it by more than a seek back
// would need and ahead of it by more than the prefetch.
func (c *segmentCache) trim(n int) {
	c.mu.Lock()
	defer c.mu.Unlock()
	for k := range c.items {
		if k < n-segmentKeepBehind || k > n+segmentPrefetch {
			delete(c.items, k)
		}
	}
}
