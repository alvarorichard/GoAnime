package startflix

import (
	"context"
	"crypto/aes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/url"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync/atomic"
	"testing"
)

// routeTo sends every request to srv whatever host it names, so tests can use
// the real host names the resolver classifies on.
type routeTo struct{ srv *httptest.Server }

func (r routeTo) RoundTrip(req *http.Request) (*http.Response, error) {
	target, _ := url.Parse(r.srv.URL)
	out := req.Clone(req.Context())
	out.Header.Set("X-Original-Host", req.URL.Host)
	out.URL.Scheme = target.Scheme
	out.URL.Host = target.Host
	out.Host = target.Host
	return http.DefaultTransport.RoundTrip(out)
}

func newTestClient(t *testing.T, h http.Handler) *Client {
	t.Helper()
	srv := httptest.NewServer(h)
	t.Cleanup(srv.Close)
	hc := &http.Client{Transport: routeTo{srv}}
	return NewClientForTest(hc, "https://www.startflix.test")
}

func fixture(t *testing.T, name string) []byte {
	t.Helper()
	raw, err := os.ReadFile(filepath.Join("testdata", name))
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestClientSearch(t *testing.T) {
	t.Parallel()
	var gotQuery, gotReferer string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		gotQuery = r.URL.Query().Get("s")
		gotReferer = r.Header.Get("Referer")
		_, _ = w.Write(fixture(t, "search_2026_10_07.html"))
	}))

	results, err := c.Search(context.Background(), "  naruto-shippuden_  ")
	if err != nil {
		t.Fatal(err)
	}
	if gotQuery != "naruto shippuden" {
		t.Errorf("query sent as %q, want slug separators turned into spaces", gotQuery)
	}
	if gotReferer != "https://www.startflix.test/" {
		t.Errorf("referer = %q", gotReferer)
	}
	if len(results) != 4 {
		t.Errorf("got %d results, want 4", len(results))
	}

	empty, err := c.Search(context.Background(), " - ")
	if err != nil || empty != nil {
		t.Errorf("blank query = (%v, %v), want no request and no results", empty, err)
	}
}

func TestClientSearchHTTPError(t *testing.T) {
	t.Parallel()
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusForbidden)
	}))
	if _, err := c.Search(context.Background(), "naruto"); err == nil || !strings.Contains(err.Error(), "403") {
		t.Fatalf("err = %v, want an HTTP 403 diagnostic", err)
	}
}

func TestClientPanelIsCached(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		_, _ = w.Write(fixture(t, "title_series_2026_10_07.html"))
	}))
	page := "https://www.startflix.test/series/the-walking-dead-dead-city/"
	for range 3 {
		p, err := c.Panel(context.Background(), page)
		if err != nil {
			t.Fatal(err)
		}
		if p.TMDBID != 194583 {
			t.Fatalf("panel = %+v", p)
		}
	}
	if n := hits.Load(); n != 1 {
		t.Errorf("title page fetched %d times, want 1", n)
	}
}

func TestClientPanelMissing(t *testing.T) {
	t.Parallel()
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`<html><body><iframe src="https://www.youtube.com/embed/x"></iframe></body></html>`))
	}))
	if _, err := c.Panel(context.Background(), "https://www.startflix.test/filmes/x/"); !errors.Is(err, ErrNoPanel) {
		t.Fatalf("err = %v, want ErrNoPanel", err)
	}
}

func TestClientSeasons(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name    string
		status  int
		body    string
		panel   Panel
		wantErr error
		seasons int
	}{
		{"panel", 200, string(fixture(t, "panel_series_2026_10_07.html")), Panel{URL: "https://p.test/embed/194583", Kind: KindSeries}, nil, 3},
		{"not on panel", 404, string(fixture(t, "panel_notfound_2026_10_07.html")), Panel{URL: "https://p.test/embed/46260", Kind: KindSeries}, ErrNotOnPanel, 0},
		{"movie panel", 200, "", Panel{URL: "https://p.test/filme/tt1", Kind: KindMovie}, nil, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte(tt.body))
			}))
			seasons, err := c.Seasons(context.Background(), tt.panel)
			if tt.panel.Kind == KindMovie {
				if err == nil {
					t.Fatal("a movie panel was accepted as a series")
				}
				return
			}
			if !errors.Is(err, tt.wantErr) {
				t.Fatalf("err = %v, want %v", err, tt.wantErr)
			}
			if len(seasons) != tt.seasons {
				t.Errorf("got %d seasons, want %d", len(seasons), tt.seasons)
			}
		})
	}
}

func TestClientPlayersRequest(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name        string
		url         string
		fixture     string
		wantReferer string
		wantAjax    bool
	}{
		{"episode", "https://p.test/episodio/299174", "episode_2026_10_07.html", "https://p.test/", true},
		{"movie", "https://p.test/filme/tt7526136", "panel_movie_2026_10_07.html", "https://www.startflix.test/", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var referer, ajax string
			c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				referer, ajax = r.Header.Get("Referer"), r.Header.Get("X-Requested-With")
				_, _ = w.Write(fixture(t, tt.fixture))
			}))
			players, err := c.Players(context.Background(), tt.url)
			if err != nil {
				t.Fatal(err)
			}
			if referer != tt.wantReferer || (ajax != "") != tt.wantAjax {
				t.Errorf("referer=%q ajax=%q, want referer %q ajax %v", referer, ajax, tt.wantReferer, tt.wantAjax)
			}
			if len(players) == 0 || players[0].referer != "https://p.test/" {
				t.Errorf("players carry no listing origin: %+v", players)
			}
		})
	}
}

// sealByse builds a playback record the way the Byse API does: 30 key parts
// with the real key split across parts v and 31-v.
func sealByse(t *testing.T, plaintext any, version int) bysePlayback {
	t.Helper()
	key := make([]byte, 32)
	iv := make([]byte, 12)
	_, _ = rand.Read(key)
	_, _ = rand.Read(iv)
	raw, _ := json.Marshal(plaintext)
	block, _ := aes.NewCipher(key)
	gcm, _ := cipher.NewGCM(block)
	sealed := gcm.Seal(nil, iv, raw, nil)

	enc := base64.RawURLEncoding.EncodeToString
	parts := make([]string, 30)
	for i := range parts {
		decoy := make([]byte, 16+i%9)
		_, _ = rand.Read(decoy)
		parts[i] = enc(decoy)
	}
	parts[version-1] = enc(key[:16])
	parts[30-version] = enc(key[16:])
	return bysePlayback{Algorithm: "AES-256-GCM", IV: enc(iv), Payload: enc(sealed), KeyParts: parts, Version: strconv.Itoa(version)}
}

func TestDecryptBysePlayback(t *testing.T) {
	t.Parallel()
	payload := map[string]any{
		"sources": []map[string]any{
			{"url": "https://cdn.test/low/master.m3u8", "height": 480},
			{"url": "https://cdn.test/hi/master.m3u8", "height": 1080},
		},
	}
	for _, v := range []int{1, 6, 15, 20} {
		pb := sealByse(t, payload, v)
		got, err := decryptBysePlayback(&pb)
		if err != nil {
			t.Fatalf("version %d: %v", v, err)
		}
		if src, ok := pickByseSource(got.Sources); !ok || src.URL != "https://cdn.test/hi/master.m3u8" {
			t.Errorf("version %d: picked %+v, want the 1080p source", v, src)
		}
	}

	pb := sealByse(t, payload, 6)
	pb.Version = "7" // wrong parts → wrong key
	if _, err := decryptBysePlayback(&pb); err == nil {
		t.Error("a wrong key opened the payload")
	}
	pb = sealByse(t, payload, 6)
	pb.Algorithm = "AES-CBC"
	if _, err := decryptBysePlayback(&pb); err == nil {
		t.Error("a non-GCM algorithm was accepted")
	}
	if _, err := decryptBysePlayback(nil); err == nil {
		t.Error("a missing playback record was accepted")
	}
}

func TestByseKeyFallsBackToAllParts(t *testing.T) {
	t.Parallel()
	enc := base64.RawURLEncoding.EncodeToString
	pb := &bysePlayback{Version: "99", KeyParts: []string{enc(make([]byte, 16)), enc(make([]byte, 16))}}
	key, err := byseKey(pb)
	if err != nil || len(key) != 32 {
		t.Fatalf("key = %d bytes, err %v; want every part joined into 32 bytes", len(key), err)
	}
	pb.KeyParts = []string{enc(make([]byte, 5))}
	if _, err := byseKey(pb); err == nil {
		t.Error("a 5-byte key was accepted")
	}
}

func TestByseCodeAndHost(t *testing.T) {
	t.Parallel()
	tests := []struct {
		raw  string
		code string
		byse bool
	}{
		{"https://embedplaybyse.top/e/3oip0k83cjd1/heart-of-the-beast", "3oip0k83cjd1", true},
		{"https://filemoon.sx/e/sti6qmfsrlda/Oppenheimer_2023", "sti6qmfsrlda", true},
		{"https://playembedapi.site/?v=H5QfVbK-c", "", false},
	}
	for _, tt := range tests {
		u, _ := url.Parse(tt.raw)
		code, _ := byseCode(u)
		if code != tt.code || isByseHost(u.Host) != tt.byse {
			t.Errorf("%s: code=%q byse=%v, want %q %v", tt.raw, code, isByseHost(u.Host), tt.code, tt.byse)
		}
	}
}

// fakeHosts serves a Byse host, a working direct file and a dead one, routed
// by the host the client asked for.
func fakeHosts(t *testing.T, bysePB *bysePlayback, byseErr string) http.Handler {
	t.Helper()
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch host := r.Header.Get("X-Original-Host"); {
		case host == "embedplaybyse.test":
			if !strings.HasPrefix(r.URL.Path, "/api/videos/") {
				http.NotFound(w, r)
				return
			}
			if byseErr != "" {
				w.WriteHeader(http.StatusNotFound)
				_, _ = fmt.Fprintf(w, `{"error":%q}`, byseErr)
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "x", "playback": bysePB})
		case host == "files.test" && r.URL.Path == "/ok.mp4":
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write([]byte{0, 0})
		case host == "files.test" && r.URL.Path == "/page.mp4":
			w.Header().Set("Content-Type", "text/html")
			_, _ = w.Write([]byte("Cai fora, irmão. >:("))
		default:
			w.WriteHeader(http.StatusForbidden)
		}
	})
}

func TestResolveStream(t *testing.T) {
	t.Parallel()
	pb := sealByse(t, map[string]any{
		"sources": []map[string]any{{"url": "https://cdn.test/master.m3u8", "height": 1080}},
		"tracks":  []map[string]any{{"file": "https://cdn.test/pt.vtt", "label": "Português", "kind": "captions"}},
	}, 6)

	abyss := Player{URL: "https://playembedapi.site/?v=a", Type: "iframe"}
	vidsrc := Player{URL: "https://vidsrcme.su/embed/tv?tmdb=1", Type: "iframe"}
	byse := Player{URL: "https://embedplaybyse.test/e/abc12345/x", Type: "iframe"}
	okFile := Player{URL: "https://files.test/ok.mp4", Type: "jwplayer", Subtitles: "https://subs.test/pt.srt", referer: "https://p.test/"}
	deadFile := Player{URL: "https://files.test/dead.mp4", Type: "jwplayer"}
	pageFile := Player{URL: "https://files.test/page.mp4", Type: "videojs"}

	tests := []struct {
		name        string
		players     []Player
		byseErr     string
		wantURL     string
		wantSubs    int
		wantErr     error
		unsupported int
		failures    int
	}{
		{name: "byse preferred over listing order", players: []Player{abyss, okFile, byse}, wantURL: "https://cdn.test/master.m3u8", wantSubs: 1},
		{name: "byse gone falls through to file", players: []Player{byse, okFile}, byseErr: "video record missing", wantURL: "https://files.test/ok.mp4", wantSubs: 1},
		// Abyss is supported; the fake answers its embed with a 403, so it
		// counts as a failure, not as an unsupported host.
		{name: "only unsupported or dead hosts", players: []Player{abyss, vidsrc, abyss}, wantErr: ErrNoSupportedServer, unsupported: 1, failures: 2},
		{name: "supported but dead", players: []Player{deadFile, pageFile, abyss}, wantErr: ErrNoSupportedServer, unsupported: 0, failures: 3},
		{name: "no players", players: nil, wantErr: ErrNoPlayers},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			c := newTestClient(t, fakeHosts(t, &pb, tt.byseErr))
			s, err := c.ResolveStream(context.Background(), tt.players)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("err = %v, want %v", err, tt.wantErr)
				}
				var ns *NoStreamError
				if errors.As(err, &ns) && (len(ns.Unsupported) != tt.unsupported || len(ns.Failures) != tt.failures) {
					t.Errorf("unsupported=%v failures=%v, want %d and %d", ns.Unsupported, ns.Failures, tt.unsupported, tt.failures)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if !strings.HasPrefix(s.URL, tt.wantURL) || len(s.Subtitles) != tt.wantSubs {
				t.Errorf("stream = %+v, want %s with %d subtitles", s, tt.wantURL, tt.wantSubs)
			}
		})
	}
}

// qualityHosts serves Byse videos by code, each with its own source height,
// and a working direct file of unknown height; it counts the Byse calls.
func qualityHosts(t *testing.T, heights map[string]int, calls *atomic.Int32) http.Handler {
	t.Helper()
	pbs := map[string]bysePlayback{}
	for code, h := range heights {
		pbs[code] = sealByse(t, map[string]any{
			"sources": []map[string]any{{"url": "https://cdn.test/" + code + "/master.m3u8", "height": h}},
		}, 6)
	}
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		switch host := r.Header.Get("X-Original-Host"); {
		case host == "embedplaybyse.test" && strings.HasPrefix(r.URL.Path, "/api/videos/"):
			calls.Add(1)
			pb, ok := pbs[strings.TrimPrefix(r.URL.Path, "/api/videos/")]
			if !ok {
				w.WriteHeader(http.StatusNotFound)
				_, _ = w.Write([]byte(`{"error":"video record missing"}`))
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]any{"title": "x", "playback": pb})
		case host == "files.test" && r.URL.Path == "/ok.mp4":
			w.Header().Set("Content-Type", "video/mp4")
			w.WriteHeader(http.StatusPartialContent)
			_, _ = w.Write([]byte{0, 0})
		default:
			w.WriteHeader(http.StatusForbidden)
		}
	})
}

// TestResolveStreamPicksTallest pins the quality rule: every supported player
// is resolved and the tallest stream wins, whatever order the players are
// listed in; equally tall streams go to the better-ranked, earlier player,
// and a failing player never hides a working one.
func TestResolveStreamPicksTallest(t *testing.T) {
	t.Parallel()
	byse := func(code string) Player {
		return Player{URL: "https://embedplaybyse.test/e/" + code + "/x", Type: "iframe"}
	}
	okFile := Player{URL: "https://files.test/ok.mp4", Type: "jwplayer", referer: "https://p.test/"}
	heights := map[string]int{"sd480000": 480, "hd720000": 720, "fhd10800": 1080, "fhd20800": 1080, "qhd14400": 1440}

	tests := []struct {
		name      string
		players   []Player
		wantCode  string
		wantH     int
		wantCalls int32
	}{
		{name: "taller listed later", players: []Player{byse("sd480000"), byse("fhd10800")}, wantCode: "fhd10800", wantH: 1080, wantCalls: 2},
		{name: "full HD is not the ceiling", players: []Player{byse("fhd10800"), byse("hd720000"), byse("qhd14400")}, wantCode: "qhd14400", wantH: 1440, wantCalls: 3},
		{name: "equal heights keep listing order", players: []Player{byse("fhd20800"), byse("fhd10800")}, wantCode: "fhd20800", wantH: 1080, wantCalls: 2},
		{name: "a failing player hides nothing", players: []Player{byse("missing1"), byse("hd720000"), byse("sd480000")}, wantCode: "hd720000", wantH: 720, wantCalls: 3},
		{name: "a known height beats an unknown one", players: []Player{okFile, byse("hd720000")}, wantCode: "hd720000", wantH: 720, wantCalls: 1},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var calls atomic.Int32
			c := newTestClient(t, qualityHosts(t, heights, &calls))
			s, err := c.ResolveStream(context.Background(), tt.players)
			if err != nil {
				t.Fatal(err)
			}
			if !strings.Contains(s.URL, "/"+tt.wantCode+"/") || s.Height != tt.wantH {
				t.Errorf("stream = %s (%dp), want %s (%dp)", s.URL, s.Height, tt.wantCode, tt.wantH)
			}
			if got := calls.Load(); got != tt.wantCalls {
				t.Errorf("Byse resolutions = %d, want %d", got, tt.wantCalls)
			}
		})
	}
}

func TestLabelHeight(t *testing.T) {
	t.Parallel()
	for label, want := range map[string]int{
		"1080p": 1080, "720p HD": 720, "FULL HD 1080P": 1080, "480p": 480, "4K": 2160,
		"Player #2": 0, "": 0, "HD": 0, "10800p": 0,
	} {
		if got := labelHeight(label); got != want {
			t.Errorf("labelHeight(%q) = %d, want %d", label, got, want)
		}
	}
}

func TestPickByseSourceFallsBackToLabel(t *testing.T) {
	t.Parallel()
	src, ok := pickByseSource([]byseSource{
		{URL: "https://cdn.test/a.m3u8", Label: "480p"},
		{URL: "https://cdn.test/b.m3u8", Label: "1080p"},
	})
	if !ok || src.URL != "https://cdn.test/b.m3u8" || src.height() != 1080 {
		t.Errorf("picked %+v, want the 1080p-labelled source", src)
	}
}

func TestResolveStreamHonoursCancellation(t *testing.T) {
	t.Parallel()
	c := newTestClient(t, http.NotFoundHandler())
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := c.ResolveStream(ctx, []Player{{URL: "https://embedplaybyse.test/e/abc12345", Type: "iframe"}})
	if !errors.Is(err, context.Canceled) {
		t.Fatalf("err = %v, want context.Canceled", err)
	}
}

func TestNoStreamErrorMessage(t *testing.T) {
	t.Parallel()
	err := &NoStreamError{Unsupported: []string{"playembedapi.site", "vidsrcme.su"}, Failures: []error{errors.New("files.test: HTTP 403")}}
	msg := err.Error()
	for _, want := range []string{"playembedapi.site", "vidsrcme.su", "files.test: HTTP 403"} {
		if !strings.Contains(msg, want) {
			t.Errorf("message %q lacks %q", msg, want)
		}
	}
}
