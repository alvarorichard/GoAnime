package topcine

import (
	"context"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
)

// retryingClient is a test client that retries like the real one, without
// waiting between attempts.
func retryingClient(t *testing.T, h http.Handler) *Client {
	t.Helper()
	c := newTestClient(t, h)
	c.maxRetries = 2
	return c
}

func TestSharedIsOneClient(t *testing.T) {
	t.Parallel()
	first, second := Shared(), Shared()
	if first != second {
		t.Error("Shared built two clients; the title cache would not be shared")
	}
}

func TestNewClientDefaults(t *testing.T) {
	t.Setenv("GOANIME_TOPCINE_URL", "  ")
	c := NewClient()
	if c.BaseURL() != DefaultBase {
		t.Errorf("BaseURL = %q, want %q for a blank override", c.BaseURL(), DefaultBase)
	}
	if c.maxRetries != 2 || c.retryDelay <= 0 || c.http == nil {
		t.Errorf("client = %+v, want retries with a delay on the guarded client", c)
	}
}

func TestGetRetries(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name     string
		statuses []int
		wantHits int32
		wantErr  bool
	}{
		{"server error then success", []int{http.StatusServiceUnavailable, http.StatusOK}, 2, false},
		{"rate limited then success", []int{http.StatusTooManyRequests, http.StatusOK}, 2, false},
		{"server error every time", []int{500, 502, 503}, 3, true},
		{"client error is not retried", []int{http.StatusNotFound}, 1, true},
		{"forbidden is not retried", []int{http.StatusForbidden}, 1, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var hits atomic.Int32
			c := retryingClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
				n := int(hits.Add(1)) - 1
				w.WriteHeader(tt.statuses[min(n, len(tt.statuses)-1)])
				_, _ = w.Write([]byte("ok"))
			}))
			body, err := c.get(context.Background(), "https://topcine.test/x", getOptions{layer: "search"})
			if got := hits.Load(); got != tt.wantHits {
				t.Errorf("hits = %d, want %d", got, tt.wantHits)
			}
			if tt.wantErr {
				if d, ok := errors.AsType[*netx.SourceDiagnostic](err); !ok || d.Source != SourceName || d.Layer != "search" {
					t.Errorf("err = %v, want a TopCine search-layer diagnostic", err)
				}
				return
			}
			if err != nil || string(body) != "ok" {
				t.Errorf("got (%q, %v)", body, err)
			}
		})
	}
}

func TestGetUnnamedLayerIsHTTP(t *testing.T) {
	t.Parallel()
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	}))
	_, err := c.get(context.Background(), "https://topcine.test/x", getOptions{})
	if d, ok := errors.AsType[*netx.SourceDiagnostic](err); !ok || d.Layer != "http" || d.StatusCode != http.StatusBadRequest {
		t.Errorf("err = %v, want an http-layer 400 diagnostic", err)
	}
}

func TestGetSendsBrowserHeaders(t *testing.T) {
	t.Parallel()
	var got http.Header
	c := newTestClient(t, http.HandlerFunc(func(_ http.ResponseWriter, r *http.Request) { got = r.Header.Clone() }))
	if _, err := c.get(context.Background(), "https://topcine.test/x", getOptions{}); err != nil {
		t.Fatal(err)
	}
	if got.Get("User-Agent") != userAgent || got.Get("Accept-Language") != netx.ChromeAcceptLanguage ||
		!strings.HasPrefix(got.Get("Accept"), "text/html") {
		t.Errorf("headers = %v", got)
	}
	if got.Get("Referer") != "" || got.Get("X-Requested-With") != "" {
		t.Errorf("unrequested referer/ajax headers sent: %v", got)
	}
}

func TestGetCancelledWhileWaitingToRetry(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	var hits atomic.Int32
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		hits.Add(1)
		cancel()
		w.WriteHeader(http.StatusServiceUnavailable)
	}))
	c.maxRetries, c.retryDelay = 3, time.Hour
	_, err := c.get(ctx, "https://topcine.test/x", getOptions{})
	if err == nil {
		t.Fatal("want an error")
	}
	if n := hits.Load(); n != 1 {
		t.Errorf("hits = %d, want 1: a cancelled call must not retry", n)
	}
}

func TestGetCancelledBeforeTheRetryWait(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	attempts := 0
	c := NewClientForTest(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("connection reset")
		}
		t.Error("retried after the context was cancelled")
		return nil, errors.New("unreachable")
	})}, "https://topcine.test")
	c.maxRetries, c.retryDelay = 1, time.Hour
	go func() { time.Sleep(20 * time.Millisecond); cancel() }()
	_, err := c.get(ctx, "https://topcine.test/x", getOptions{})
	if !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled from the retry wait", err)
	}
}

type roundTripFunc func(*http.Request) (*http.Response, error)

func (f roundTripFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestGetTransportErrorIsRetried(t *testing.T) {
	t.Parallel()
	attempts := 0
	c := NewClientForTest(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		attempts++
		if attempts == 1 {
			return nil, errors.New("connection reset")
		}
		return &http.Response{StatusCode: http.StatusOK, Body: io.NopCloser(strings.NewReader("ok")), Request: r}, nil
	})}, "https://topcine.test")
	c.maxRetries = 1
	body, err := c.get(context.Background(), "https://topcine.test/x", getOptions{})
	if err != nil || string(body) != "ok" || attempts != 2 {
		t.Errorf("got (%q, %v) after %d attempts, want ok after 2", body, err, attempts)
	}
}

type failingBody struct{}

func (failingBody) Read([]byte) (int, error) { return 0, errors.New("body cut") }
func (failingBody) Close() error             { return nil }

func TestGetBodyReadErrorIsRetried(t *testing.T) {
	t.Parallel()
	attempts := 0
	c := NewClientForTest(&http.Client{Transport: roundTripFunc(func(r *http.Request) (*http.Response, error) {
		attempts++
		return &http.Response{StatusCode: http.StatusOK, Body: failingBody{}, Request: r}, nil
	})}, "https://topcine.test")
	c.maxRetries = 1
	_, err := c.get(context.Background(), "https://topcine.test/x", getOptions{})
	if err == nil || !strings.Contains(err.Error(), "body cut") || attempts != 2 {
		t.Errorf("err = %v after %d attempts, want the read error after 2", err, attempts)
	}
}

func TestGetInvalidURL(t *testing.T) {
	t.Parallel()
	c := NewClientForTest(http.DefaultClient, "https://topcine.test")
	c.maxRetries = 2
	if _, err := c.get(context.Background(), "http://[::1", getOptions{}); err == nil {
		t.Error("an unparsable URL was requested")
	}
}

func TestSearchInvalidBaseURL(t *testing.T) {
	t.Parallel()
	c := NewClientForTest(http.DefaultClient, "http://[::1")
	if _, err := c.Search(context.Background(), "x"); err == nil || !strings.Contains(err.Error(), "invalid base URL") {
		t.Errorf("err = %v, want an invalid base URL error", err)
	}
}

func TestTitleHTTPErrorIsNotCached(t *testing.T) {
	t.Parallel()
	var hits atomic.Int32
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		if hits.Add(1) == 1 {
			w.WriteHeader(http.StatusBadGateway)
			return
		}
		_, _ = w.Write(fixture(t, "title_movie_2026_10_09.html"))
	}))
	page := "https://topcine.test/filme/zona-zero"
	if _, err := c.Title(context.Background(), page); err == nil {
		t.Fatal("want the 502")
	}
	title, err := c.Title(context.Background(), page)
	if err != nil || title.TMDBID != 1375646 {
		t.Errorf("second try = (%+v, %v), want the page once it answers", title, err)
	}
}

func TestLanguagesHTTPError(t *testing.T) {
	t.Parallel()
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	}))
	_, err := c.Languages(context.Background(), "https://azullog.top/filme/1")
	if d, ok := errors.AsType[*netx.SourceDiagnostic](err); !ok || d.Layer != "player" {
		t.Errorf("err = %v, want a player-layer diagnostic", err)
	}
}

func TestLanguagesKeepsThePlayerQuery(t *testing.T) {
	t.Parallel()
	var query, referer string
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		query, referer = r.URL.RawQuery, r.Header.Get("Referer")
		_, _ = w.Write([]byte(`{"sucesso":true,"tem_dublado":true,"url_dublado":"not a url","url_legendado":"https://b.test/x"}`))
	}))
	got, err := c.Languages(context.Background(), "https://azullog.top/filme/1?v=2")
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(query, "v=2") || !strings.Contains(query, "acao=carregar_idiomas") {
		t.Errorf("query = %q, want the player's own query kept", query)
	}
	if referer != "https://azullog.top/filme/1?v=2" {
		t.Errorf("referer = %q, want the player page", referer)
	}
	if len(got.Hosts) != 1 || got.Hosts[0] != "b.test" {
		t.Errorf("hosts = %v, want only the parsable track URL's host", got.Hosts)
	}
}

func TestSearchCancelledContext(t *testing.T) {
	t.Parallel()
	c := newTestClient(t, http.HandlerFunc(func(http.ResponseWriter, *http.Request) {
		t.Error("requested with a cancelled context")
	}))
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, err := c.Search(ctx, "x"); !errors.Is(err, context.Canceled) {
		t.Errorf("err = %v, want context.Canceled", err)
	}
}
