package pomfy

import (
	"context"
	"encoding/base64"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// The last edges: the seal's field decodes, the 307 re-send lane and the
// cancelled-retry lane of the HTTP layer, one refusal per shape.

// TestSealFieldFailures: the seal's iv and payload decode errors, driven
// directly against decryptBysePlayback with a valid key path.
func TestSealFieldFailures(t *testing.T) {
	good := base64KeyArena()

	badKeyParts := playbackSeal{Algorithm: "", IV: good, Payload: "AAAA", KeyParts: []string{"!!no base64!!"}, Version: ""}
	if _, err := decryptBysePlayback(&badKeyParts); err == nil || !strings.Contains(err.Error(), "key part") {
		t.Fatalf("err = %v, want the key-part error", err)
	}

	badPayload := sealFixture(t, []byte(`{"sources":[]}`))
	badPayload.Payload = "!!no base64!!"
	if _, err := decryptBysePlayback(&badPayload); err == nil || !strings.Contains(err.Error(), "payload:") {
		t.Fatalf("err = %v, want the payload decode error", err)
	}
}

// base64KeyArena hands a clean base64 key that joins to 32 bytes.
func base64KeyArena() string {
	return base64Raw(strings.Repeat("A", 42))
}

// TestPost307RunsGetBody: a 307 redirect re-sends the POST body through the
// request's GetBody — the seam the crowded counting tests never reach.
func TestPost307RunsGetBody(t *testing.T) {
	var sawBody bool
	redirects := 0
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/done" {
			body := make([]byte, 64)
			n, _ := r.Body.Read(body)
			if n > 0 {
				sawBody = true
			}
			_, _ = w.Write([]byte(`{"ok":true}`))
			return
		}
		redirects++
		if redirects > 2 {
			http.Error(w, `{"error":"loop"}`, http.StatusBadGateway)
			return
		}
		// 307 re-sends method + body
		w.Header().Set("Location", "/done")
		w.WriteHeader(http.StatusTemporaryRedirect)
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL)
	c.maxRetries = 0
	body, err := c.postJSON(context.Background(), srv.URL, "/start", map[string]any{"x": 1}, nil)
	if err != nil {
		t.Fatalf("postJSON over the 307: %v", err)
	}
	if !strings.Contains(string(body), "ok") {
		t.Errorf("body = %q", body)
	}
	if !sawBody {
		t.Error("the 307 re-send never carried the body (GetBody not run)")
	}
}

// TestGetRetryCancelMidLane: the cancelled context during the sleepy reset
// lane returns the ctx error from the select, not from the dial.
func TestGetRetryCancelMidLane(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"upstream"}`, http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL)
	c.maxRetries = 3
	c.retryDelay = 50 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	if _, err := c.get(ctx, srv.URL+"/x", nil); err == nil ||
		!strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("err = %v, want the cancelled-context answer", err)
	}
}

// TestResolveStreamErrLine: the all-sources dead lane already pins the
// return; here the light rider for the EditCheck sheet completes by handing
// the route's error forward.
func TestResolveStreamHostWhitelistSuicide(t *testing.T) {
	world := newFakeWorld(t)
	c := world.client()
	// the site names a source on a host with no route: dialing it dies
	world.site.mu.Lock()
	world.site.breakWatch = true
	world.site.mu.Unlock()

	_, err := c.ResolveStream(t.Context(), 550, KindMovie, 0, 0, DefaultBase+"/assistir/550")
	if err == nil {
		t.Fatal("a dead watch route must fail the resolve")
	}
}

// TestDecodeShowNodeEmptyPayload pins the empty-array refusal directly.
func TestDecodeShowNodeEmptyPayload(t *testing.T) {
	_, err := decodeShowNode(nil)
	if err == nil || !strings.Contains(err.Error(), "empty payload") {
		t.Fatalf("err = %v, want the empty-payload refusal", err)
	}
}

// TestSearchSkipsKindlessResults covers the continue on a result with no
// kind at all — the seam returning a shapeless entry.
func TestSearchSkipsKindlessResults(t *testing.T) {
	site := newFakeCatalogSite(t)
	c := site.newClient(t)
	c.SetSearchSeam(func(string) ([]Media, error) {
		return []Media{
			{}, // no kind, no type: skipped
			{Title: "Clube da Luta", TMDBID: 550, TMDBType: "movie"},
		}, nil
	})
	results, err := c.Search(t.Context(), "clube")
	if err != nil {
		t.Fatalf("Search: %v", err)
	}
	if len(results) != 1 || results[0].TMDBID != 550 {
		t.Fatalf("results = %+v, want only the typed one", results)
	}
}

// TestSearchMultiReadError: the truncated TMDB body keeps the search's read
// lane honest.
func TestSearchMultiReadError(t *testing.T) {
	hijack := hijackedShortServer(t)
	resetSearchCache(t)
	old := tmdbBase
	tmdbBase = hijack.URL
	t.Cleanup(func() { tmdbBase = old })
	if _, err := searchMultiWithKey(PublicKeyForTesting(0), "x"); err == nil {
		t.Error("a truncated TMDB body must fail the read")
	}
}

// --- small helpers, or the file reads naked ---

// base64Raw is base64.RawURLEncoding shortened to its import.
func base64Raw(s string) string {
	return base64.RawURLEncoding.EncodeToString([]byte(s))
}

// TestPostJSONCancelMidLane: the same cancelled-select for the POST lane —
// the counting server 502s until the ctx kills the retry sleep.
func TestPostJSONCancelMidLane(t *testing.T) {
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		http.Error(w, `{"error":"upstream"}`, http.StatusBadGateway)
	}))
	t.Cleanup(srv.Close)
	c := NewClientForTest(srv.Client(), srv.URL)
	c.maxRetries = 3
	c.retryDelay = 50 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(20 * time.Millisecond)
		cancel()
	}()
	if _, err := c.postJSON(ctx, srv.URL, "/x", map[string]any{}, nil); err == nil ||
		!strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("err = %v, want the cancelled POST", err)
	}
}

// TestParseMovieTitleSkipsNonArrayNode: the title payload whose first node's
// data is an object is skipped, and the second carries the record.
func TestParseMovieTitleSkipsNonArrayNode(t *testing.T) {
	shape := compact(`{"nodes":[
		{"data": {"a": 1}},
		{"data":[
			{"id":1,"tipo":2,"title":3,"year":4,"overview":5,"image":6},
			"550","filme","Clube da Luta","1999","desc","/x.jpg"
		]}
	]}`)
	title, err := parseMovieTitle([]byte(shape))
	if err != nil {
		t.Fatalf("parseMovieTitle: %v", err)
	}
	if title.Name != "Clube da Luta" || title.Year != "1999" {
		t.Errorf("title = %+v", title)
	}
}
