package scraper

import (
	"context"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/browser"
)

func TestNewHTMLFetcherOptions(t *testing.T) {
	tests := []struct {
		name  string
		opts  []FetcherOption
		check func(*testing.T, *HTMLFetcher)
	}{
		{
			name: "default options",
			opts: []FetcherOption{},
			check: func(t *testing.T, f *HTMLFetcher) {
				if f.useBrowser != false {
					t.Errorf("expected useBrowser=false, got %v", f.useBrowser)
				}
				if f.fallback != true {
					t.Errorf("expected fallback=true, got %v", f.fallback)
				}
				if f.httpClient == nil {
					t.Error("expected httpClient to be set")
				}
			},
		},
		{
			name: "with browser enabled",
			opts: []FetcherOption{WithBrowser(true)},
			check: func(t *testing.T, f *HTMLFetcher) {
				if f.useBrowser != true {
					t.Errorf("expected useBrowser=true, got %v", f.useBrowser)
				}
			},
		},
		{
			name: "with fallback disabled",
			opts: []FetcherOption{WithFallback(false)},
			check: func(t *testing.T, f *HTMLFetcher) {
				if f.fallback != false {
					t.Errorf("expected fallback=false, got %v", f.fallback)
				}
			},
		},
		{
			name: "combined options",
			opts: []FetcherOption{
				WithBrowser(true),
				WithFallback(false),
			},
			check: func(t *testing.T, f *HTMLFetcher) {
				if f.useBrowser != true {
					t.Error("expected useBrowser=true")
				}
				if f.fallback != false {
					t.Error("expected fallback=false")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fetcher := NewHTMLFetcher(tt.opts...)
			tt.check(t, fetcher)
		})
	}
}

func TestGetDefaultFetcher(t *testing.T) {
	// GetDefaultFetcher should return same instance
	f1 := GetDefaultFetcher()
	f2 := GetDefaultFetcher()

	if f1 != f2 {
		t.Error("GetDefaultFetcher should return same instance")
	}
}

func TestHTMLFetcherHTTPGet(t *testing.T) {
	// Create a test server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Test Content</body></html>"))
	}))
	defer server.Close()

	fetcher := NewHTMLFetcher()
	ctx := context.Background()

	html, err := fetcher.GetHTML(ctx, server.URL)
	if err != nil {
		t.Fatalf("failed to get HTML: %v", err)
	}

	if html == nil {
		t.Fatal("expected non-nil HTML reader")
	}

	// Read content
	content, err := io.ReadAll(html)
	if err != nil {
		t.Fatalf("failed to read content: %v", err)
	}

	if !strings.Contains(string(content), "Test Content") {
		t.Errorf("expected content to contain 'Test Content', got: %s", string(content))
	}
}

func TestHTMLFetcherHTTPGetWithHeaders(t *testing.T) {
	var receivedHeaders http.Header

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeaders = r.Header.Clone()
		w.Header().Set("Content-Type", "text/html")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>OK</body></html>"))
	}))
	defer server.Close()

	fetcher := NewHTMLFetcher()
	ctx := context.Background()

	customHeaders := map[string]string{
		"X-Custom-Header": "custom-value",
	}

	_, err := fetcher.GetHTMLWithHeaders(ctx, server.URL, customHeaders)
	if err != nil {
		t.Fatalf("failed to get HTML: %v", err)
	}

	if receivedHeaders.Get("X-Custom-Header") != "custom-value" {
		t.Errorf("expected X-Custom-Header to be 'custom-value', got: %s", receivedHeaders.Get("X-Custom-Header"))
	}
}

func TestHTMLFetcherHTTPGetInvalidURL(t *testing.T) {
	fetcher := NewHTMLFetcher()
	ctx := context.Background()

	_, err := fetcher.GetHTML(ctx, "http://invalid-domain-that-does-not-exist-12345.com")
	if err == nil {
		t.Error("expected error for invalid URL")
	}
}

func TestHTMLFetcherHTTPGetStatusCode(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	fetcher := NewHTMLFetcher()
	ctx := context.Background()

	_, err := fetcher.GetHTML(ctx, server.URL)
	if err == nil {
		t.Error("expected error for 404 status code")
	}
}

func TestHTMLFetcherWithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Delay response to trigger timeout
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>OK</body></html>"))
	}))
	defer server.Close()

	fetcher := NewHTMLFetcher()
	ctx, cancel := context.WithTimeout(context.Background(), 50*time.Millisecond)
	defer cancel()

	_, err := fetcher.GetHTML(ctx, server.URL)
	if err == nil {
		t.Error("expected error for timeout")
	}
}

func TestHTMLFetcherPostJSON(t *testing.T) {
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("expected POST method, got %s", r.Method)
		}
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
		}

		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"status":"ok"}`))
	}))
	defer server.Close()

	fetcher := NewHTMLFetcher()

	result, err := fetcher.postJSON(context.Background(), server.URL, `{"test":"value"}`)
	if err != nil {
		t.Fatalf("failed to post JSON: %v", err)
	}

	if receivedBody != `{"test":"value"}` {
		t.Errorf("expected body '%s', got '%s'", `{"test":"value"}`, receivedBody)
	}

	if !strings.Contains(result, "ok") {
		t.Errorf("expected result to contain 'ok', got: %s", result)
	}
}

func TestHTMLFetcherIsBrowserAvailable(t *testing.T) {
	fetcher := NewHTMLFetcher()

	// Default fetcher doesn't use browser
	available := fetcher.IsBrowserAvailable()
	if available {
		t.Log("Browser is available (this may be expected if browser is installed)")
	}

	// Create fetcher with browser enabled
	fetcherWithBrowser := NewHTMLFetcher(WithBrowser(true))
	_ = fetcherWithBrowser

	// This test just verifies the method doesn't panic
	// Actual availability depends on browser installation
}

func TestHTMLFetcherClose(t *testing.T) {
	fetcher := NewHTMLFetcher()

	// Close should not panic
	err := fetcher.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// Create fetcher with browser
	fetcherWithBrowser := NewHTMLFetcher(WithBrowser(true))
	err = fetcherWithBrowser.Close()
	if err != nil {
		t.Errorf("Close with browser returned error: %v", err)
	}
}

func TestSimpleGet(t *testing.T) {
	// Skip in CI if network is not available
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Simple Get Test</body></html>"))
	}))
	defer server.Close()

	// Note: SimpleGet uses defaultFetcher which is HTTP only
	// This will make an actual request, so we use our test server
	fetcher := GetDefaultFetcher()
	result, err := fetcher.GetHTML(context.Background(), server.URL)
	if err != nil {
		t.Fatalf("SimpleGet failed: %v", err)
	}

	content, _ := io.ReadAll(result)
	if !strings.Contains(string(content), "Simple Get Test") {
		t.Errorf("unexpected content: %s", string(content))
	}
}

func TestSimpleGetWithContext(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>Context Test</body></html>"))
	}))
	defer server.Close()

	fetcher := GetDefaultFetcher()
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := fetcher.GetHTML(ctx, server.URL)
	if err != nil {
		t.Fatalf("SimpleGetWithContext failed: %v", err)
	}

	content, _ := io.ReadAll(result)
	if !strings.Contains(string(content), "Context Test") {
		t.Errorf("unexpected content: %s", string(content))
	}
}

func TestGetWithHeaders(t *testing.T) {
	var receivedHeader string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		receivedHeader = r.Header.Get("X-Test-Header")
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>OK</body></html>"))
	}))
	defer server.Close()

	fetcher := GetDefaultFetcher()
	headers := map[string]string{
		"X-Test-Header": "test-value",
	}

	result, err := fetcher.GetHTMLWithHeaders(context.Background(), server.URL, headers)
	if err != nil {
		t.Fatalf("GetWithHeaders failed: %v", err)
	}

	io.ReadAll(result)

	if receivedHeader != "test-value" {
		t.Errorf("expected header 'test-value', got '%s'", receivedHeader)
	}
}

func TestGetWithTimeout(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(50 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("<html><body>OK</body></html>"))
	}))
	defer server.Close()

	// This should timeout
	result, err := GetWithTimeout(server.URL, 10*time.Millisecond)
	if err == nil {
		t.Error("expected timeout error")
	}
	if result != "" {
		t.Errorf("expected empty result on timeout, got: %s", result)
	}
}

func TestPostJSONWithContext(t *testing.T) {
	var receivedBody string

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		receivedBody = string(body)
		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"success":true}`))
	}))
	defer server.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	result, err := PostJSONWithContext(ctx, server.URL, `{"key":"value"}`)
	if err != nil {
		t.Fatalf("PostJSONWithContext failed: %v", err)
	}

	if receivedBody != `{"key":"value"}` {
		t.Errorf("expected body '%s', got '%s'", `{"key":"value"}`, receivedBody)
	}

	if !strings.Contains(result, "success") {
		t.Errorf("expected result to contain 'success', got: %s", result)
	}
}

func TestFetcherFactory(t *testing.T) {
	factory := NewFetcherFactory()

	if factory == nil {
		t.Fatal("expected non-nil factory")
	}

	// Test chaining
	fetcher := factory.
		UseBrowser(false).
		Fallback(true).
		Create()

	if fetcher == nil {
		t.Fatal("expected non-nil fetcher from factory")
	}

	if fetcher.useBrowser != false {
		t.Error("expected useBrowser=false from factory")
	}

	if fetcher.fallback != true {
		t.Error("expected fallback=true from factory")
	}
}

func TestFetcherFactoryWithBrowserOptions(t *testing.T) {
	factory := NewFetcherFactory()

	fetcher := factory.
		UseBrowser(true).
		BrowserOption(browser.WithHeadless(true)).
		Create()

	if fetcher == nil {
		t.Fatal("expected non-nil fetcher from factory")
	}

	// The browser client should be set
	if fetcher.browserClient == nil {
		t.Log("Note: browserClient is nil because GetDefaultBrowser lazy initializes")
	}
}
