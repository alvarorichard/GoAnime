package plutotv

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
)

// A fake Pluto TV: the session, the catalog and the stitcher, with the rules
// the real service enforced when this was written (2026-11-16) and nothing
// more:
//
//   - the catalog wants the session token as a Bearer, the stitcher wants it as
//     the jwt query parameter and a session id beside it;
//   - the media host answers 403 to a request that names no User-Agent;
//   - a token the service has renewed is the only one it accepts.
//
// TestTheFakeEnforcesWhatTheRealServiceDoes holds it to that, because every
// test below leans on it: a fake that lets a wrong request through would make
// them pass for the wrong reason.

const (
	fakeSeriesID  = "6206bba315c8610013683974"
	fakeEpisodeID = "6206bbaa15c8610013683a28"
	fakeMovieID   = "62b9ea9ec4eb8f00140cb633"
	fakeSID       = "11111111-2222-4333-8444-555555555555"

	// fakeClip is where the title's own media lives on the fake CDN.
	fakeClip = "/media/clip/abc_Title/720p/20220214/hls/"
)

// fakeLadder is Pluto TV's ladder: the stitched bandwidth of each rung and the
// name of its rendition on the CDN.
var fakeLadder = []struct {
	bandwidth int
	rendition string
}{
	{640930, "hls_300"},
	{1042180, "hls_600"},
	{1577180, "hls_1000"},
	{2294080, "hls_1600"},
	{3321280, "hls_2400"},
}

type fakeRoute struct {
	status int
	body   string
	ctype  string
}

type fakePluto struct {
	t   *testing.T
	srv *httptest.Server

	mu       sync.Mutex
	boots    int
	token    string // the one token the service accepts now
	tokenFmt string // how a token is spelled; %d is the boot count
	refresh  int    // refreshInSec of the boot answer
	reject   int    // the next N authenticated calls answer 401
	bootBody func(f *fakePluto) string
	bootCode int
	routes   map[string]fakeRoute
	hits     map[string]int
	requests []*http.Request
	onHit    func(path string)
}

func newFakePluto(t *testing.T) *fakePluto {
	t.Helper()
	f := &fakePluto{t: t, tokenFmt: "tok-%d", routes: map[string]fakeRoute{}, hits: map[string]int{}}
	f.srv = httptest.NewServer(http.HandlerFunc(f.serve))
	t.Cleanup(f.srv.Close)
	return f
}

// client is a client pointed at the fake, with the fake's host trusted.
func (f *fakePluto) client() *Client { return NewClientForTest(f.srv.Client(), f.srv.URL) }

func (f *fakePluto) url() string { return f.srv.URL }

func (f *fakePluto) set(path, body string) { f.setRoute(path, http.StatusOK, body) }

func (f *fakePluto) setRoute(path string, status int, body string) {
	f.mu.Lock()
	defer f.mu.Unlock()
	f.routes[path] = fakeRoute{status: status, body: body}
}

func (f *fakePluto) hitCount(path string) int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.hits[path]
}

func (f *fakePluto) bootCount() int {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.boots
}

// lastRequest is the latest request for path, or nil.
func (f *fakePluto) lastRequest(path string) *http.Request {
	f.mu.Lock()
	defer f.mu.Unlock()
	for i := len(f.requests) - 1; i >= 0; i-- {
		if f.requests[i].URL.Path == path {
			return f.requests[i]
		}
	}
	return nil
}

func (f *fakePluto) bootJSON() string {
	if f.bootBody != nil {
		return f.bootBody(f)
	}
	return fmt.Sprintf(`{
 "sessionToken": %q,
 "stitcherParams": "appName=web&sid=%s&deviceId=DEV&serverSideAds=false",
 "refreshInSec": %d,
 "servers": {"search": %q, "vod": %q, "stitcher": %q}
}`, f.token, fakeSID, f.refresh, f.srv.URL, f.srv.URL, f.srv.URL)
}

func (f *fakePluto) serve(w http.ResponseWriter, r *http.Request) {
	f.mu.Lock()
	f.hits[r.URL.Path]++
	f.requests = append(f.requests, r.Clone(r.Context()))
	hook := f.onHit
	f.mu.Unlock()
	if hook != nil {
		hook(r.URL.Path)
	}

	path := r.URL.Path
	switch {
	case path == "/v4/start":
		f.mu.Lock()
		f.boots++
		f.token = fmt.Sprintf(f.tokenFmt, f.boots)
		code := f.bootCode
		body := f.bootJSON()
		f.mu.Unlock()
		if code != 0 {
			http.Error(w, "boot failed", code)
			return
		}
		w.Header().Set("Content-Type", "application/json")
		_, _ = w.Write([]byte(body))
		return
	case strings.HasPrefix(path, "/v1/search"), strings.HasPrefix(path, "/v4/vod/"):
		if !f.authorized(r.Header.Get("Authorization") == "Bearer "+f.currentToken()) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	case strings.HasPrefix(path, "/v2/stitch/") && strings.HasSuffix(path, "/master.m3u8"):
		q := r.URL.Query()
		if q.Get("sid") == "" || q.Get("masterJWTPassthrough") != "true" {
			http.Error(w, "bad stitcher query", http.StatusBadRequest)
			return
		}
		if !f.authorized(q.Get("jwt") == f.currentToken()) {
			http.Error(w, "unauthorized", http.StatusUnauthorized)
			return
		}
	case strings.HasPrefix(path, "/media/"):
		if strings.TrimSpace(r.Header.Get("User-Agent")) == "" {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
	}

	f.mu.Lock()
	route, ok := f.routes[path]
	f.mu.Unlock()
	if !ok {
		http.NotFound(w, r)
		return
	}
	if route.ctype != "" {
		w.Header().Set("Content-Type", route.ctype)
	}
	w.WriteHeader(route.status)
	_, _ = w.Write([]byte(route.body))
}

func (f *fakePluto) currentToken() string {
	f.mu.Lock()
	defer f.mu.Unlock()
	return f.token
}

// authorized is whether the caller's credentials are right, unless the fake
// was told to answer 401 to the next few calls whatever they carry.
func (f *fakePluto) authorized(credentialsOK bool) bool {
	f.mu.Lock()
	defer f.mu.Unlock()
	if f.reject > 0 {
		f.reject--
		return false
	}
	return credentialsOK
}

// ---- builders --------------------------------------------------------------

func (f *fakePluto) masterPath(id string) string {
	return "/v2/stitch/hls/episode/" + id + "/master.m3u8"
}

func (f *fakePluto) variantPath(id string, bandwidth int) string {
	return fmt.Sprintf("/v2/stitch/hls/episode/%s/%d/playlist.m3u8", id, bandwidth)
}

func (f *fakePluto) contentPath(rendition string) string {
	return fakeClip + "0-end/" + rendition + ".m3u8"
}

// masterOf lists the rungs at the given indexes of fakeLadder, in that order,
// with relative URIs like the real master.
func masterOf(rungs ...int) string {
	var b strings.Builder
	b.WriteString("#EXTM3U\n#EXT-X-VERSION:5\n")
	b.WriteString("#EXT-X-MEDIA:TYPE=AUDIO,GROUP-ID=\"audio\",LANGUAGE=\"pt\",NAME=\"Portuguese [Original]\",AUTOSELECT=YES,DEFAULT=YES,CHANNELS=\"2\",URI=\"audio/audio/Portuguese/audio.m3u8?sid=" + fakeSID + "\"\n")
	for _, i := range rungs {
		fmt.Fprintf(&b, "#EXT-X-STREAM-INF:PROGRAM-ID=1,BANDWIDTH=%d,AUDIO=\"audio\"\n%d/playlist.m3u8?sid=%s\n", fakeLadder[i].bandwidth, fakeLadder[i].bandwidth, fakeSID)
	}
	return b.String()
}

// stitched is the title cut into ranges with an ad between each, the shape the
// stitcher answers with. Each range is `segments` long.
func (f *fakePluto) stitched(rendition string, segments int, ranges ...string) string {
	var b strings.Builder
	b.WriteString("#EXTM3U\n#EXT-X-VERSION:5\n#EXT-X-PLAYLIST-TYPE:VOD\n#EXT-X-TARGETDURATION:5\n#EXT-X-MEDIA-SEQUENCE:0\n")
	for i, r := range ranges {
		if i > 0 {
			b.WriteString("#EXT-X-DISCONTINUITY\n#EXT-X-KEY:METHOD=NONE\n")
			fmt.Fprintf(&b, "#EXTINF:5,\n%s/v1/mp4/c(ts)/h(default)/p(1_ad%%2Fcreative%%2F%d_ad%%2F720p%%2Fvideo_600.mp4)\n", f.url(), i)
			b.WriteString("#EXT-X-DISCONTINUITY\n")
		}
		for s := 0; s < segments; s++ {
			fmt.Fprintf(&b, "#EXT-X-KEY:METHOD=AES-128,URI=\"%s%s%s/%s_keyfile_0.key\",IV=0x00000000000000000000000000000001\n", f.url(), fakeClip, r, rendition)
			fmt.Fprintf(&b, "#EXTINF:5,\n%s%s%s/%s-%05d.ts\n", f.url(), fakeClip, r, rendition, s)
		}
	}
	// The service's own closing bumper, a clip of its own.
	fmt.Fprintf(&b, "#EXT-X-DISCONTINUITY\n#EXT-X-KEY:METHOD=NONE\n#EXTINF:5,\n%s/media/clip/bumper/1080pDRM/hls/0-end/hls_bumper-00000.ts\n#EXT-X-ENDLIST\n", f.url())
	return b.String()
}

// content is the title's own playlist: one clip, start to end.
func content(rendition string, segments int) string {
	var b strings.Builder
	b.WriteString("#EXTM3U\n#EXT-X-VERSION:3\n#EXT-X-PLAYLIST-TYPE:VOD\n#EXT-X-TARGETDURATION:5\n#EXT-X-MEDIA-SEQUENCE:0\n")
	for s := 0; s < segments; s++ {
		fmt.Fprintf(&b, "#EXT-X-KEY:METHOD=AES-128,URI=\"%s_keyfile_0.key\",IV=0x%032x\n#EXTINF:5.000000,\n%s-%05d.ts\n", rendition, s+1, rendition, s)
	}
	b.WriteString("#EXT-X-ENDLIST\n")
	return b.String()
}

// serveEpisode registers an episode (or movie) every rung of which resolves to
// the title's own playlist, split into three ranges by default.
func (f *fakePluto) serveEpisode(id string, rungs ...int) {
	f.t.Helper()
	if len(rungs) == 0 {
		rungs = []int{0, 1, 2, 3, 4}
	}
	f.set(f.masterPath(id), masterOf(rungs...))
	for _, i := range rungs {
		rung := fakeLadder[i]
		f.set(f.variantPath(id, rung.bandwidth), f.stitched(rung.rendition, 3, "0-100", "101-200", "201-end"))
		f.set(f.contentPath(rung.rendition), content(rung.rendition, 9))
	}
}

// ---- catalog bodies --------------------------------------------------------

func searchBody(hits ...string) string {
	return `{"suggestions": [], "data": [` + strings.Join(hits, ",") + `]}`
}

func hit(kind, id, name, slug, lang string) string {
	return fmt.Sprintf(`{"id": %q, "slug": %q, "name": %q, "type": %q, "language": %q, "rating": "A14", "distributeAs": {"AVOD": true}}`, id, slug, name, kind, lang)
}

func seasonsBody(id, name string, seasons ...string) string {
	return fmt.Sprintf(`{"_id": %q, "name": %q, "type": "series", "seasons": [%s]}`, id, name, strings.Join(seasons, ","))
}

func seasonOf(number int, episodes ...string) string {
	return fmt.Sprintf(`{"number": %d, "episodes": [%s]}`, number, strings.Join(episodes, ","))
}

func episodeOf(id, name string, number, season int, durationMs int64) string {
	return fmt.Sprintf(`{"_id": %q, "name": %q, "description": "about %s", "number": %d, "season": %d, "originalContentDuration": %d, "type": "episode"}`, id, name, name, number, season, durationMs)
}

// oid is a catalog id made of one repeated hex digit.
func oid(d byte) string { return strings.Repeat(string(d), 24) }

// spyHost stands in for a host that is not Pluto's: it answers every request
// with `body`, like a healthy server would, and counts who came. A client that
// follows a host it should not have is caught by the count, not by a network
// failure that would hide it.
type spyHost struct {
	*httptest.Server
	hits  atomic.Int32
	auths atomic.Value // the Authorization of the last request
}

func newSpyHost(t *testing.T, body string) *spyHost {
	t.Helper()
	s := &spyHost{}
	s.auths.Store("")
	s.Server = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		s.hits.Add(1)
		s.auths.Store(r.Header.Get("Authorization"))
		_, _ = w.Write([]byte(body))
	}))
	t.Cleanup(s.Close)
	return s
}
