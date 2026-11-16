package plutotv

import (
	"container/list"
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

type servedPlaylist struct{ id, text string }

func newPlaylistServer() *playlistServer {
	return &playlistServer{lists: make(map[string]*list.Element), recent: list.New()}
}

var playlists = newPlaylistServer()

// publishFn publishes a playlist and returns its local URL; a seam.
var publishFn = playlists.publish

// listenLoopback asks the system for a free port on 127.0.0.1.
func listenLoopback() (net.Listener, error) { return net.Listen("tcp", "127.0.0.1:0") }

// PlayURL is the address a player, or a downloader, should be given for s: the
// loopback address of its playlist when Stream read one, and s.URL otherwise
// (a title with no playlist of its own, or no port to serve one from, in which
// case the URL still plays wherever the CDN answers).
func PlayURL(s Stream) string {
	if s.Playlist == "" {
		return s.URL
	}
	local, err := publishFn(s.Playlist)
	if err != nil {
		util.Debug("Pluto TV playlist is played from its own URL", "error", err)
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
	s.lists[id] = s.recent.PushFront(&servedPlaylist{id: id, text: playlist})
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
	// An id is 32 hex digits, so a path that is not /plutotv/<id>.m3u8 leaves
	// something that is no id and is simply not found.
	id, isPlaylist := strings.CutSuffix(strings.TrimPrefix(r.URL.Path, "/plutotv/"), ".m3u8")
	s.mu.Lock()
	entry, found := s.lists[id]
	playlist := ""
	if found && isPlaylist {
		s.recent.MoveToFront(entry) // in use: not the one to forget
		playlist = entry.Value.(*servedPlaylist).text
	}
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
