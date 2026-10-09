package topcine

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
)

// One test per client function not already named by another test.

func TestNewClientForTest(t *testing.T) {
	t.Parallel()
	hc := &http.Client{}
	c := NewClientForTest(hc, "https://topcine.test///")
	if c.http != hc {
		t.Error("the caller's HTTP client is not used")
	}
	if c.baseURL != "https://topcine.test" {
		t.Errorf("baseURL = %q, want trailing slashes trimmed", c.baseURL)
	}
	if c.maxRetries != 0 || c.retryDelay != 0 {
		t.Errorf("retries = %d/%v, want none so tests stay fast", c.maxRetries, c.retryDelay)
	}
	if c.userAgent != userAgent {
		t.Errorf("userAgent = %q", c.userAgent)
	}
}

func TestClientBaseURL(t *testing.T) {
	t.Parallel()
	tests := map[string]string{
		"https://topcine.test":   "https://topcine.test",
		"https://topcine.test/":  "https://topcine.test",
		"https://topcine4.site/": "https://topcine4.site",
	}
	for in, want := range tests {
		if got := NewClientForTest(http.DefaultClient, in).BaseURL(); got != want {
			t.Errorf("BaseURL(%q) = %q, want %q", in, got, want)
		}
	}
}

func TestGetOnce(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name      string
		status    int
		wantRetry bool
		wantErr   bool
	}{
		{"ok", http.StatusOK, false, false},
		{"no content is still a success", http.StatusNoContent, false, false},
		{"server error asks for a retry", http.StatusInternalServerError, true, true},
		{"bad gateway asks for a retry", http.StatusBadGateway, true, true},
		{"rate limit asks for a retry", http.StatusTooManyRequests, true, true},
		{"not found is final", http.StatusNotFound, false, true},
		{"forbidden is final", http.StatusForbidden, false, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			var gotReferer, gotAccept, gotAjax string
			c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				gotReferer, gotAccept, gotAjax = r.Header.Get("Referer"), r.Header.Get("Accept"), r.Header.Get("X-Requested-With")
				w.WriteHeader(tt.status)
				_, _ = w.Write([]byte("body"))
			}))
			body, retry, err := c.getOnce(context.Background(), "https://topcine.test/x", getOptions{
				referer: "https://topcine.test/", accept: "application/json", ajax: true, layer: "player",
			})
			if retry != tt.wantRetry {
				t.Errorf("retry = %v, want %v", retry, tt.wantRetry)
			}
			if tt.wantErr {
				if d, ok := errors.AsType[*netx.SourceDiagnostic](err); !ok || d.StatusCode != tt.status || d.Layer != "player" {
					t.Errorf("err = %v, want a player-layer %d diagnostic", err, tt.status)
				}
				if body != nil {
					t.Errorf("body = %q on a failure", body)
				}
			} else if err != nil || (tt.status == http.StatusOK && string(body) != "body") {
				t.Errorf("got (%q, %v)", body, err)
			}
			if gotReferer != "https://topcine.test/" || gotAccept != "application/json" || gotAjax != "XMLHttpRequest" {
				t.Errorf("headers referer=%q accept=%q ajax=%q, want the requested ones", gotReferer, gotAccept, gotAjax)
			}
		})
	}
}

func TestGetOnceTransportError(t *testing.T) {
	t.Parallel()
	c := NewClientForTest(&http.Client{Transport: roundTripFunc(func(*http.Request) (*http.Response, error) {
		return nil, errors.New("connection refused")
	})}, "https://topcine.test")

	_, retry, err := c.getOnce(context.Background(), "https://topcine.test/x", getOptions{})
	if err == nil || !retry {
		t.Errorf("(retry %v, err %v), want a retryable error", retry, err)
	}

	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	if _, retry, err := c.getOnce(ctx, "https://topcine.test/x", getOptions{}); err == nil || retry {
		t.Errorf("cancelled: (retry %v, err %v), want a final error", retry, err)
	}
}

func TestGetDocument(t *testing.T) {
	t.Parallel()
	c := newTestClient(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.URL.Path == "/missing" {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		_, _ = w.Write([]byte(`<html><body><h1>TopCine</h1></body></html>`))
	}))

	doc, err := c.getDocument(context.Background(), "https://topcine.test/", getOptions{layer: "title"})
	if err != nil {
		t.Fatal(err)
	}
	if got := doc.Find("h1").Text(); got != "TopCine" {
		t.Errorf("h1 = %q", got)
	}

	if _, err := c.getDocument(context.Background(), "https://topcine.test/missing", getOptions{layer: "title"}); err == nil {
		t.Error("a 404 page was parsed")
	}
}
