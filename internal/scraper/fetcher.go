package scraper

import (
	"context"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/browser"
	"github.com/alvarorichard/Goanime/internal/util"
)

var defaultFetcher *HTMLFetcher

type HTMLFetcher struct {
	browserClient *browser.BrowserClient
	httpClient    *http.Client
	useBrowser    bool
	fallback      bool
}

type FetcherOption func(*HTMLFetcher)

func WithBrowser(enabled bool) FetcherOption {
	return func(f *HTMLFetcher) { f.useBrowser = enabled }
}

func WithFallback(enabled bool) FetcherOption {
	return func(f *HTMLFetcher) { f.fallback = enabled }
}

func NewHTMLFetcher(opts ...FetcherOption) *HTMLFetcher {
	fetcher := &HTMLFetcher{
		httpClient: util.GetFastClient(),
		useBrowser: false,
		fallback:   true,
	}

	for _, opt := range opts {
		opt(fetcher)
	}

	if fetcher.useBrowser {
		fetcher.browserClient = browser.GetDefaultBrowser()
	}

	return fetcher
}

func GetDefaultFetcher() *HTMLFetcher {
	if defaultFetcher == nil {
		defaultFetcher = NewHTMLFetcher()
	}
	return defaultFetcher
}

func (f *HTMLFetcher) GetHTML(ctx context.Context, url string) (io.Reader, error) {
	if f.useBrowser && f.browserClient != nil {
		html, err := f.browserClient.GetHTML(ctx, url)
		if err == nil {
			return html, nil
		}
		if !f.fallback {
			return nil, fmt.Errorf("browser fetch failed: %w", err)
		}
		util.Debug("browser fetch failed, falling back to HTTP", "error", err)
	}

	return f.httpGetHTML(ctx, url)
}

func (f *HTMLFetcher) httpGetHTML(ctx context.Context, url string) (io.Reader, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %s", resp.Status)
	}

	return resp.Body, nil
}

func (f *HTMLFetcher) GetHTMLWithHeaders(ctx context.Context, url string, headers map[string]string) (io.Reader, error) {
	if f.useBrowser && f.browserClient != nil {
		html, err := f.browserClient.GetHTML(ctx, url)
		if err == nil {
			return html, nil
		}
		if !f.fallback {
			return nil, fmt.Errorf("browser fetch failed: %w", err)
		}
		util.Debug("browser fetch failed, falling back to HTTP", "error", err)
	}

	return f.httpGetHTMLWithHeaders(ctx, url, headers)
}

func (f *HTMLFetcher) httpGetHTMLWithHeaders(ctx context.Context, url string, headers map[string]string) (io.Reader, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Accept", "text/html,application/xhtml+xml,application/xml;q=0.9,image/avif,image/webp,*/*;q=0.8")
	req.Header.Set("Accept-Language", "en-US,en;q=0.9")

	for k, v := range headers {
		req.Header.Set(k, v)
	}

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("server returned: %s", resp.Status)
	}

	return resp.Body, nil
}

func (f *HTMLFetcher) GetHTMLWithWait(ctx context.Context, url string, waitFor string) (io.Reader, error) {
	if f.useBrowser && f.browserClient != nil {
		html, err := f.browserClient.GetHTMLWithWait(ctx, url, waitFor)
		if err == nil {
			return html, nil
		}
		if !f.fallback {
			return nil, fmt.Errorf("browser fetch failed: %w", err)
		}
		util.Debug("browser fetch failed, falling back to HTTP", "error", err)
	}

	return f.httpGetHTML(ctx, url)
}

func (f *HTMLFetcher) IsBrowserAvailable() bool {
	if f.browserClient == nil {
		return false
	}
	return f.browserClient.IsAvailable()
}

func (f *HTMLFetcher) Close() error {
	if f.browserClient != nil {
		return f.browserClient.Close()
	}
	return nil
}

func InstallPlaywrightBrowsers() error {
	return browser.InstallBrowsers()
}

type FetcherFactory struct {
	useBrowser  bool
	fallback    bool
	browserOpts []browser.Option
}

func NewFetcherFactory() *FetcherFactory {
	return &FetcherFactory{
		useBrowser: false,
		fallback:   true,
	}
}

func (f *FetcherFactory) UseBrowser(enabled bool) *FetcherFactory {
	f.useBrowser = enabled
	return f
}

func (f *FetcherFactory) Fallback(enabled bool) *FetcherFactory {
	f.fallback = enabled
	return f
}

func (f *FetcherFactory) BrowserOption(opts ...browser.Option) *FetcherFactory {
	f.browserOpts = opts
	return f
}

func (f *FetcherFactory) Create() *HTMLFetcher {
	opts := []FetcherOption{
		WithBrowser(f.useBrowser),
		WithFallback(f.fallback),
	}

	if f.useBrowser && len(f.browserOpts) > 0 {
		browserClient := browser.NewBrowserClient(f.browserOpts...)
		return &HTMLFetcher{
			browserClient: browserClient,
			httpClient:    util.GetFastClient(),
			useBrowser:    f.useBrowser,
			fallback:      f.fallback,
		}
	}

	fetcher := &HTMLFetcher{
		httpClient: util.GetFastClient(),
		useBrowser: f.useBrowser,
		fallback:   f.fallback,
	}

	for _, opt := range opts {
		opt(fetcher)
	}

	return fetcher
}

func SimpleGet(url string) (string, error) {
	fetcher := GetDefaultFetcher()
	reader, err := fetcher.GetHTML(context.Background(), url)
	if err != nil {
		return "", err
	}
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func SimpleGetWithContext(ctx context.Context, url string) (string, error) {
	fetcher := GetDefaultFetcher()
	reader, err := fetcher.GetHTML(ctx, url)
	if err != nil {
		return "", err
	}
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func GetWithHeaders(url string, headers map[string]string) (string, error) {
	fetcher := GetDefaultFetcher()
	reader, err := fetcher.GetHTMLWithHeaders(context.Background(), url, headers)
	if err != nil {
		return "", err
	}
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func GetWithTimeout(url string, timeout time.Duration) (string, error) {
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()
	fetcher := GetDefaultFetcher()
	reader, err := fetcher.GetHTML(ctx, url)
	if err != nil {
		return "", err
	}
	content, err := io.ReadAll(reader)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

func PostJSON(url string, body string) (string, error) {
	fetcher := GetDefaultFetcher()
	return fetcher.postJSON(context.Background(), url, body)
}

func PostJSONWithContext(ctx context.Context, url string, body string) (string, error) {
	fetcher := GetDefaultFetcher()
	return fetcher.postJSON(ctx, url, body)
}

func (f *HTMLFetcher) postJSON(ctx context.Context, url string, body string) (string, error) {
	req, err := http.NewRequestWithContext(ctx, "POST", url, strings.NewReader(body))
	if err != nil {
		return "", fmt.Errorf("failed to create request: %w", err)
	}

	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Accept", "application/json")

	resp, err := f.httpClient.Do(req)
	if err != nil {
		return "", fmt.Errorf("failed to make request: %w", err)
	}
	defer resp.Body.Close()

	content, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	return string(content), nil
}
