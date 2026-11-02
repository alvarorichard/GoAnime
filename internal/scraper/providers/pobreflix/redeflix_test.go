package pobreflix

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync"
	"testing"
)

type seenCall struct {
	path, origin, referer, requestedWith string
	body                                 map[string]string
}

// fakeRedeflix serves the RedeFlixApi player: the page with its ticket and the
// two POST endpoints, recording every call.
type fakeRedeflix struct {
	srv          *httptest.Server
	mu           sync.Mutex
	calls        []seenCall
	resolveCode  int
	resolveError string
	media        string
	resolveHost  string // overrides the host the page names for playback-resolve
	direct       bool   // answer playback-resolve with the media URL itself
}

func newFakeRedeflix(t *testing.T) *fakeRedeflix {
	t.Helper()
	f := &fakeRedeflix{resolveCode: http.StatusOK, media: "https://cdn.example/m/film.mp4?X-Amz-Signature=abc"}
	f.srv = httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		f.mu.Lock()
		defer f.mu.Unlock()
		if r.Method == http.MethodGet {
			resolve := f.srv.URL + "/playback-resolve.php"
			if f.resolveHost != "" {
				resolve = f.resolveHost + "/playback-resolve.php"
			}
			_, _ = w.Write([]byte(`<html><title>x</title><script>var playbackResolveTicket = "v1.TICKET_ONE-abc";
var playbackResolveUrl = "` + resolve + `";</script></html>`))
			return
		}
		raw, _ := io.ReadAll(r.Body)
		c := seenCall{path: r.URL.Path, origin: r.Header.Get("Origin"), referer: r.Header.Get("Referer"),
			requestedWith: r.Header.Get("X-Requested-With"), body: map[string]string{}}
		_ = json.Unmarshal(raw, &c.body)
		f.calls = append(f.calls, c)
		w.Header().Set("Content-Type", "application/json")
		switch r.URL.Path {
		case "/playback-resolve.php":
			if f.resolveCode != http.StatusOK {
				w.WriteHeader(f.resolveCode)
				_ = json.NewEncoder(w).Encode(map[string]string{"error": f.resolveError})
				return
			}
			if f.direct {
				_ = json.NewEncoder(w).Encode(map[string]string{"url": f.media})
				return
			}
			_ = json.NewEncoder(w).Encode(map[string]string{"launchTicket": "v1.LAUNCH", "nextTicket": "v1.NEXT"})
		case "/playerkys/source.php":
			_ = json.NewEncoder(w).Encode(map[string]string{"url": f.media})
		default:
			http.NotFound(w, r)
		}
	}))
	t.Cleanup(f.srv.Close)
	return f
}

func (f *fakeRedeflix) client(t *testing.T) *Client {
	return NewClientForTest(http.DefaultClient, "https://catalog.example", t.TempDir()+"/c.json")
}

func TestResolveRedeflixDubbedUsesThePlayersOwnTwoStepFlow(t *testing.T) {
	f := newFakeRedeflix(t)
	got, err := f.client(t).ResolveRedeflix(t.Context(), f.srv.URL+"/filme/310", AudioDubbed)
	if err != nil {
		t.Fatal(err)
	}
	if got.URL != f.media || got.Audio != AudioDubbed {
		t.Fatalf("stream = %+v; want %s, dubbed", got, f.media)
	}
	if len(f.calls) != 2 || f.calls[0].path != "/playback-resolve.php" || f.calls[1].path != "/playerkys/source.php" {
		t.Fatalf("calls = %+v; want playback-resolve then source", f.calls)
	}
	if f.calls[0].body["ticket"] != "v1.TICKET_ONE-abc" || f.calls[0].body["server"] != "default" {
		t.Errorf("resolve body = %v; want the page's ticket and server default", f.calls[0].body)
	}
	if f.calls[1].body["ticket"] != "v1.LAUNCH" {
		t.Errorf("source body = %v; want the launch ticket", f.calls[1].body)
	}
	for i, c := range f.calls {
		if c.origin != f.srv.URL {
			t.Errorf("call %d Origin = %q; the player's own origin %q is the only one to send", i, c.origin, f.srv.URL)
		}
		if !strings.HasPrefix(c.referer, f.srv.URL) {
			t.Errorf("call %d Referer = %q; must stay on the player's host", i, c.referer)
		}
		if c.requestedWith != "RedeFlixPlayer" {
			t.Errorf("call %d X-Requested-With = %q", i, c.requestedWith)
		}
	}
}

func TestResolveRedeflixSubtitledAsksForLegendado(t *testing.T) {
	f := newFakeRedeflix(t)
	got, err := f.client(t).ResolveRedeflix(t.Context(), f.srv.URL+"/filme/310", AudioSubtitled)
	if err != nil {
		t.Fatal(err)
	}
	if f.calls[0].body["server"] != "legendado" || got.Audio != AudioSubtitled {
		t.Fatalf("server = %q, audio = %q; want legendado, subtitled", f.calls[0].body["server"], got.Audio)
	}
}

func TestResolveRedeflixAcceptsAMediaURLInTheFirstAnswer(t *testing.T) {
	f := newFakeRedeflix(t)
	f.direct = true
	got, err := f.client(t).ResolveRedeflix(t.Context(), f.srv.URL+"/filme/310", AudioDubbed)
	if err != nil || got.URL != f.media || len(f.calls) != 1 {
		t.Fatalf("got %+v, err %v, %d calls; want the URL after a single call", got, err, len(f.calls))
	}
}

// A page in the Turnstile format is not this player: no POST goes out and the
// caller is told to try the solver.
func TestResolveRedeflixLeavesOtherPagesAlone(t *testing.T) {
	var posts int
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method == http.MethodPost {
			posts++
		}
		_, _ = w.Write([]byte(`<html><title>Verificando...</title><div class="cf-turnstile" data-sitekey="k"></div></html>`))
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(http.DefaultClient, "https://catalog.example", t.TempDir()+"/c.json")
	_, err := c.ResolveRedeflix(t.Context(), srv.URL+"/filme/1", AudioDubbed)
	if !errors.Is(err, ErrNotRedeflixPlayer) || posts != 0 {
		t.Fatalf("err = %v, posts = %d; want ErrNotRedeflixPlayer and no POST", err, posts)
	}
}

// The ticket is only ever sent to the page's own host, even if the page names
// another one for its resolve endpoint.
func TestResolveRedeflixNeverSendsTheTicketToAnotherHost(t *testing.T) {
	var leaked int
	other := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) { leaked++ }))
	t.Cleanup(other.Close)
	f := newFakeRedeflix(t)
	f.resolveHost = other.URL
	_, err := f.client(t).ResolveRedeflix(t.Context(), f.srv.URL+"/filme/310", AudioDubbed)
	if err == nil || leaked != 0 || len(f.calls) != 0 {
		t.Fatalf("err = %v, leaked = %d, calls = %d; the ticket must stay on the page's host", err, leaked, len(f.calls))
	}
}

func TestResolveRedeflixReportsTheSitesOwnError(t *testing.T) {
	f := newFakeRedeflix(t)
	f.resolveCode, f.resolveError = http.StatusForbidden, "Origem de reprodução não autorizada."
	_, err := f.client(t).ResolveRedeflix(t.Context(), f.srv.URL+"/filme/310", AudioDubbed)
	if err == nil || !strings.Contains(err.Error(), "403") || !strings.Contains(err.Error(), "não autorizada") {
		t.Fatalf("err = %v; want the status and the site's message", err)
	}
	if strings.Contains(err.Error(), "TICKET_ONE") {
		t.Fatalf("the error leaks the ticket: %v", err)
	}
}

func TestResolveRedeflixRejectsAMediaURLThatIsNotHTTPS(t *testing.T) {
	for _, bad := range []string{"http://cdn.example/f.mp4", "file:///etc/passwd", "javascript:alert(1)", ""} {
		f := newFakeRedeflix(t)
		f.media = bad
		if got, err := f.client(t).ResolveRedeflix(t.Context(), f.srv.URL+"/filme/310", AudioDubbed); err == nil {
			t.Errorf("media %q accepted: %+v", bad, got)
		}
	}
}

func TestSolverFreePlayerIsJudgedByHostAlone(t *testing.T) {
	c := NewClientForTest(http.DefaultClient, "https://catalog.example", t.TempDir()+"/c.json")
	for rawURL, want := range map[string]bool{
		"https://redeflixapi.bar/filme/310":   true,
		"https://RedeFlixApi.example/filme/1": true,
		"https://streambetter.lat/filme/310":  false,
		"https://streambetter.test/filme/1":   false,
		"https://cdn.example/f.mp4":           false,
		"/api/player?id=310":                  false,
		"":                                    false,
	} {
		if got := c.IsSolverFreePlayer(rawURL); got != want {
			t.Errorf("IsSolverFreePlayer(%q) = %v; want %v", rawURL, got, want)
		}
	}
	t.Setenv(envOpenPlayerHosts, " player.moved.example , other.example")
	if !c.IsSolverFreePlayer("https://player.moved.example/filme/1") || c.IsSolverFreePlayer("https://moved.example/filme/1") {
		t.Error("the env list must match whole hosts only")
	}

	turnstileOnly := []Server{{URL: "https://streambetter.lat/filme/310"}}
	both := []Server{{URL: "https://streambetter.lat/filme/310"}, {URL: "https://redeflixapi.bar/filme/310"}}
	files := []Server{{URL: "https://cdn.example/redeflix.mp4"}}
	if c.HasSolverFreePlayer(turnstileOnly) || !c.HasSolverFreePlayer(both) || c.HasSolverFreePlayer(files) {
		t.Error("HasSolverFreePlayer must be true only for a player page on such a host")
	}
}
