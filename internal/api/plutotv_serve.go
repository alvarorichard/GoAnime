package api

import (
	"crypto/rand"
	"encoding/hex"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/util"
)

// Pluto TV's media host is several CDNs behind one name, and one of them
// refuses every playlist under a clip while serving the segments the playlist
// lists. A player handed the playlist's URL therefore opens it or fails by the
// DNS answer it happens to get, and can fail again on a later request for it.
// So the player is not handed that URL: the plutotv client reads the playlist
// (or assembles it), and this server hands the text to the player from
// loopback. The segments and keys in it are absolute and are fetched from the
// CDN directly; only the few kilobytes of playlist pass through here.

// maxPlutoPlaylists bounds what the server remembers: the playlists of the
// episodes resolved most recently (the one playing, a batch being downloaded).
const maxPlutoPlaylists = 64

// plutoPlaylistServer serves resolved playlists on 127.0.0.1.
type plutoPlaylistServer struct {
	mu    sync.Mutex
	base  string            // "http://127.0.0.1:port", set once the listener is up
	lists map[string]string // id -> playlist text
	order []string          // ids, oldest first
	// listen opens the loopback listener; a seam, nil meaning the real one.
	listen func() (net.Listener, error)
}

var plutoPlaylists = &plutoPlaylistServer{lists: make(map[string]string)}

// listenLoopback asks the system for a free port on 127.0.0.1.
func listenLoopback() (net.Listener, error) { return net.Listen("tcp", "127.0.0.1:0") }

// plutoPlaylistURLFn publishes a playlist and returns its local URL; a seam.
var plutoPlaylistURLFn = plutoPlaylists.publish

// publish registers playlist and returns the URL it is served at, starting
// the listener on first use.
func (s *plutoPlaylistServer) publish(playlist string) (string, error) {
	if !strings.HasPrefix(playlist, "#EXTM3U") {
		return "", errors.New("not a playlist")
	}
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
	s.lists[id] = playlist
	s.order = append(s.order, id)
	for len(s.order) > maxPlutoPlaylists {
		delete(s.lists, s.order[0])
		s.order = s.order[1:]
	}
	return s.base + "/plutotv/" + id + ".m3u8", nil
}

// ServeHTTP answers a GET or HEAD for a published playlist and nothing else.
func (s *plutoPlaylistServer) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet && r.Method != http.MethodHead {
		w.Header().Set("Allow", "GET, HEAD")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	// An id is 32 hex digits, so a path that is not /plutotv/<id>.m3u8 leaves
	// something that is no id and is simply not found.
	id, isPlaylist := strings.CutSuffix(strings.TrimPrefix(r.URL.Path, "/plutotv/"), ".m3u8")
	s.mu.Lock()
	playlist, found := s.lists[id]
	s.mu.Unlock()
	if !isPlaylist || !found {
		http.NotFound(w, r)
		return
	}
	w.Header().Set("Content-Type", "application/vnd.apple.mpegurl")
	w.Header().Set("Cache-Control", "no-store")
	w.Header().Set("Content-Length", fmt.Sprint(len(playlist)))
	_, _ = w.Write([]byte(playlist)) // net/http sends no body for a HEAD
}
