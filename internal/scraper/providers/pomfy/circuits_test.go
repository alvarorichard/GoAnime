package pomfy

import (
	"context"
	"net/http"
	"net/http/httptest"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/v2/internal/api/movie"
)

// Circuit-level tests: the HTTP layer's retry paths, the transport's host
// errors, the seasons' parse shapes, and the search pool's rotation.

// countingServer answers the first n-1 tries with status and then the okBody.
func countingServer(t *testing.T, failTimes, status int, okBody string) (*httptest.Server, *atomic.Int64) {
	t.Helper()
	counter := &atomic.Int64{}
	mux := http.NewServeMux()
	mux.HandleFunc("POST /", func(w http.ResponseWriter, _ *http.Request) {
		n := counter.Add(1)
		if int(n) <= failTimes {
			http.Error(w, `{"error":"upstream dead"}`, status)
			return
		}
		_, _ = w.Write([]byte(okBody))
	})
	mux.HandleFunc("GET /", func(w http.ResponseWriter, _ *http.Request) {
		n := counter.Add(1)
		if int(n) <= failTimes {
			http.Error(w, `{"error":"upstream dead"}`, status)
			return
		}
		_, _ = w.Write([]byte(okBody))
	})
	srv := httptest.NewServer(mux)
	t.Cleanup(srv.Close)
	return srv, counter
}

// TestGetRetriesThenSucceeds: get() retries on 5xx/429 and the sleep lane of
// the retry loop is the one that waits for the second attempt.
func TestGetRetriesThenSucceeds(t *testing.T) {
	srv, counter := countingServer(t, 1, http.StatusInternalServerError, `{"filmes":[550],"series":[]}`)
	c := NewClientForTest(srv.Client(), srv.URL)
	c.maxRetries = 2
	c.retryDelay = 1 * time.Millisecond

	body, err := c.get(context.Background(), srv.URL+"/api/catalogo-ids", nil)
	if err != nil {
		t.Fatalf("get: %v", err)
	}
	if !strings.Contains(string(body), "filmes") {
		t.Errorf("body = %q", body)
	}
	if counter.Load() != 2 {
		t.Errorf("attempt count = %d, want 2 after one retry", counter.Load())
	}
}

// TestGetFailsFastOn4xx: non-retryable statuses stop the loop.
func TestGetFailsFastOn4xx(t *testing.T) {
	srv, counter := countingServer(t, 99, http.StatusNotFound, `{}`)
	c := NewClientForTest(srv.Client(), srv.URL)
	c.maxRetries = 2
	c.retryDelay = 1 * time.Millisecond

	if _, err := c.get(context.Background(), srv.URL+"/api/catalogo-ids", nil); err == nil {
		t.Fatal("expected the 404 to fail")
	}
	if counter.Load() != 1 {
		t.Errorf("attempt count = %d, want 1 (no retry on 4xx)", counter.Load())
	}
}

// TestGetRetryWaitsForCanceledContext: a canceled ctx during the sleep lane
// ends the call as a context error, not the transport's.
func TestGetRetryWaitsForCanceledContext(t *testing.T) {
	srv, counter := countingServer(t, 2, http.StatusInternalServerError, `{}`)
	c := NewClientForTest(srv.Client(), srv.URL)
	c.maxRetries = 2
	c.retryDelay = 30 * time.Millisecond

	ctx, cancel := context.WithCancel(context.Background())
	go func() {
		time.Sleep(10 * time.Millisecond)
		cancel()
	}()
	if _, err := c.get(ctx, srv.URL+"/api/catalogo-ids", nil); err == nil || !strings.Contains(err.Error(), "context canceled") {
		t.Fatalf("err = %v, want a canceled context", err)
	}
	counterWatcher := counter.Load()
	_ = counterWatcher
}

// TestAttemptRequestError covers the URL parse failure lane.
func TestAttemptRequestError(t *testing.T) {
	c := NewClientForTest(&http.Client{}, DefaultBase)
	if _, _, err := c.attempt(context.Background(), http.MethodGet, "http://[::1]:badport-no", nil, "", nil); err == nil {
		t.Fatal("an unparseable URL must fail NewRequest")
	}
}

// TestAttemptTransportError covers the dial-failure lane (a closed server).
func TestAttemptTransportError(t *testing.T) {
	dead := httptest.NewServer(http.HandlerFunc(func(http.ResponseWriter, *http.Request) {}))
	url := dead.URL
	dead.Close()
	c := NewClientForTest(&http.Client{}, DefaultBase)
	_, retry, err := c.attempt(context.Background(), http.MethodGet, url, nil, "", nil)
	if err == nil {
		t.Fatal("expected a transport error")
	}
	if !retry {
		t.Error("a connect error is retryable")
	}
}

// TestPostJSONRetryAndMarshalFail mirrors the get tests for the POST shape.
func TestPostJSONRetryAndMarshalFail(t *testing.T) {
	srv, counter := countingServer(t, 1, http.StatusBadGateway, `{"ok":true}`)
	c := NewClientForTest(srv.Client(), srv.URL)
	c.maxRetries = 2
	c.retryDelay = 1 * time.Millisecond

	body, err := c.postJSON(context.Background(), srv.URL, "/api/videos/x/embed/details", map[string]any{}, nil)
	if err != nil {
		t.Fatalf("postJSON: %v", err)
	}
	if !strings.Contains(string(body), "ok") {
		t.Errorf("body = %q", body)
	}
	if counter.Load() != 2 {
		t.Errorf("attempt count = %d, want 2", counter.Load())
	}

	// an unmarshalable payload fails fast
	if _, err := c.postJSON(context.Background(), srv.URL, "/x", make(chan int), nil); err == nil {
		t.Error("a channel payload must fail marshalling")
	}
}

// TestSeasonParseShapes keeps the season parser honest: garbage and a
// season with no episodes.
func TestSeasonParseShapes(t *testing.T) {
	garbage := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = w.Write([]byte(`{"epNums": "not-an-array"`))
	}))
	t.Cleanup(garbage.Close)
	g := NewClientForTest(garbage.Client(), garbage.URL)
	if _, err := g.Season(context.Background(), 1396, 1); err == nil {
		t.Error("an unparseable season must error")
	}
}

// --- the remaining ResolveByse edge: embeds that fail to parse name the --
// --- fallback, so no broken host URL is dialed blind.

func TestResolveByseUnparseableEmbed(t *testing.T) {
	c := NewClientForTest(&http.Client{Timeout: time.Second}, DefaultBase)
	_, err := c.ResolveByse(t.Context(), watchSource{EmbedURL: "not a url at all"}, DefaultBase+"/assistir/550")
	if err == nil || !strings.Contains(err.Error(), "carries no byse code") {
		t.Fatalf("err = %v, want the no-code refusal", err)
	}
	// An embed with a code but no host names no server: the contract refuses
	// before dialing anything.
	_, err = c.ResolveByse(t.Context(), watchSource{EmbedURL: "https:/e/abc12345"}, DefaultBase+"/assistir/550")
	if err == nil || !strings.Contains(err.Error(), "embed unreadable") {
		t.Fatalf("err = %v, want the unreadable-embed refusal", err)
	}
	_, err = c.ResolveByse(t.Context(), watchSource{EmbedURL: "https://cdn.test/e/"}, DefaultBase+"/assistir/550")
	if err == nil {
		t.Fatal("an embed with no code still needs a refusal")
	}
}

func TestSeriesPlayURLFormat(t *testing.T) {
	if got := SeriesPlayURL(1396, 2, 1); got != apiCatalog+"/serie/1396/2/1" {
		t.Errorf("SeriesPlayURL = %q", got)
	}
}

func TestMovieSearchMultiRunsOverTheRealClient(t *testing.T) {
	// The user-key client path runs over the movie client — a fake TMDB and
	// the ForTest constructor exercise its own parse layer, no network.
	fake := newFakeTMDB(t, "clube da luta", []map[string]any{
		{"id": 550, "media_type": "movie", "title": "Clube da Luta", "release_date": "1999-10-15", "vote_average": 8.4},
	})
	fake.userKey = "the-user-key00000"
	client := movie.NewTMDBClientForTest(fake.srv.Client(), fake.srv.URL+"/3", "the-user-key00000")
	out, err := movieSearchMulti(client, "clube da luta")
	if err != nil {
		t.Fatalf("movieSearchMulti: %v", err)
	}
	if len(out) != 1 || out[0].Title != "Clube da Luta" || out[0].Year != "1999" {
		t.Errorf("mapped = %+v", out)
	}

	// The garbage body surfaces as an error through the same client.
	junk := junkSearchServer(t)
	junkClient := movie.NewTMDBClientForTest(junk.Client(), junk.URL+"/3", "the-user-key00000")
	if _, err := movieSearchMulti(junkClient, "clube da luta"); err == nil {
		t.Error("an unreadable answer must error via the movie client")
	}
}

func junkSearchServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		_, _ = w.Write([]byte(`not json at all`))
	}))
	t.Cleanup(srv.Close)
	return srv
}
