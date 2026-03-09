package browser

import (
	"context"
	"fmt"
	"io"
	"strings"
	"sync"
	"time"

	"github.com/playwright-community/playwright-go"
)

const (
	DefaultTimeout = 30 * time.Second
	DefaultWidth   = 1920
	DefaultHeight  = 1080
)

type Config struct {
	Headless    bool
	Timeout     time.Duration
	UserAgent   string
	Width       int
	Height      int
	MaxPages    int
	InstallDeps bool
}

type Option func(*Config)

func WithHeadless(headless bool) Option {
	return func(c *Config) { c.Headless = headless }
}

func WithTimeout(timeout time.Duration) Option {
	return func(c *Config) { c.Timeout = timeout }
}

func WithUserAgent(ua string) Option {
	return func(c *Config) { c.UserAgent = ua }
}

func WithViewport(width, height int) Option {
	return func(c *Config) {
		c.Width = width
		c.Height = height
	}
}

func WithMaxPages(max int) Option {
	return func(c *Config) { c.MaxPages = max }
}

func WithInstallDeps(install bool) Option {
	return func(c *Config) { c.InstallDeps = install }
}

type BrowserClient struct {
	playwright  *playwright.Playwright
	browser     playwright.Browser
	pool        *pagePool
	config      Config
	mu          sync.RWMutex
	initialized bool
}

var (
	defaultUserAgent = "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/131.0.0.0 Safari/537.36"
	browserOnce      sync.Once
	browserInstance  *BrowserClient
	Playwright       *playwright.Playwright
)

func init() {
	Playwright = nil
}

func NewBrowserClient(opts ...Option) *BrowserClient {
	cfg := Config{
		Headless:    true,
		Timeout:     DefaultTimeout,
		UserAgent:   defaultUserAgent,
		Width:       DefaultWidth,
		Height:      DefaultHeight,
		MaxPages:    5,
		InstallDeps: true,
	}

	for _, opt := range opts {
		opt(&cfg)
	}

	return &BrowserClient{
		config: cfg,
	}
}

func GetDefaultBrowser() *BrowserClient {
	browserOnce.Do(func() {
		browserInstance = NewBrowserClient()
	})
	return browserInstance
}

func (c *BrowserClient) initialize() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.initialized {
		return nil
	}

	pw, err := playwright.Run()
	if err != nil {
		if c.config.InstallDeps {
			return c.tryInstallAndRetry()
		}
		return fmt.Errorf("failed to start playwright: %w", err)
	}
	c.playwright = pw

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(c.config.Headless),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--disable-dev-shm-usage",
			"--no-sandbox",
		},
	})
	if err != nil {
		pw.Stop()
		if c.config.InstallDeps {
			return c.tryInstallAndRetry()
		}
		return fmt.Errorf("failed to launch browser: %w", err)
	}
	c.browser = browser

	c.pool = newPagePool(browser, c.config)
	c.initialized = true

	return nil
}

func (c *BrowserClient) tryInstallAndRetry() error {
	installErr := InstallBrowsers()
	if installErr != nil {
		return fmt.Errorf("failed to install browsers: %w", installErr)
	}

	pw, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("failed to start playwright after install: %w", err)
	}
	c.playwright = pw

	browser, err := pw.Chromium.Launch(playwright.BrowserTypeLaunchOptions{
		Headless: playwright.Bool(c.config.Headless),
		Args: []string{
			"--disable-blink-features=AutomationControlled",
			"--disable-dev-shm-usage",
			"--no-sandbox",
		},
	})
	if err != nil {
		pw.Stop()
		return fmt.Errorf("failed to launch browser after install: %w", err)
	}
	c.browser = browser

	c.pool = newPagePool(browser, c.config)
	c.initialized = true

	return nil
}

func (c *BrowserClient) ensureInitialized() error {
	if !c.initialized {
		return c.initialize()
	}
	return nil
}

func (c *BrowserClient) GetHTML(ctx context.Context, url string) (io.Reader, error) {
	if err := c.ensureInitialized(); err != nil {
		return nil, err
	}

	page, release, err := c.pool.acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire page: %w", err)
	}
	defer release()

	timeout := c.config.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	timeoutMs := float64(timeout.Milliseconds())

	_, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   &timeoutMs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	html, err := page.Content()
	if err != nil {
		return nil, fmt.Errorf("failed to get page content: %w", err)
	}

	return strings.NewReader(html), nil
}

func (c *BrowserClient) GetHTMLWithWait(ctx context.Context, url string, waitFor string) (io.Reader, error) {
	if err := c.ensureInitialized(); err != nil {
		return nil, err
	}

	page, release, err := c.pool.acquire(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to acquire page: %w", err)
	}
	defer release()

	timeout := c.config.Timeout
	if timeout <= 0 {
		timeout = DefaultTimeout
	}

	timeoutMs := float64(timeout.Milliseconds())

	_, err = page.Goto(url, playwright.PageGotoOptions{
		WaitUntil: playwright.WaitUntilStateDomcontentloaded,
		Timeout:   &timeoutMs,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to navigate to %s: %w", url, err)
	}

	if waitFor != "" {
		_, err = page.WaitForSelector(waitFor, playwright.PageWaitForSelectorOptions{
			State:   playwright.WaitForSelectorStateVisible,
			Timeout: &timeoutMs,
		})
		if err != nil {
			return nil, fmt.Errorf("failed to wait for selector %s: %w", waitFor, err)
		}
	}

	html, err := page.Content()
	if err != nil {
		return nil, fmt.Errorf("failed to get page content: %w", err)
	}

	return strings.NewReader(html), nil
}

func (c *BrowserClient) IsAvailable() bool {
	if err := c.ensureInitialized(); err != nil {
		return false
	}
	return c.browser != nil && c.playwright != nil
}

func (c *BrowserClient) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if c.pool != nil {
		c.pool.close()
	}

	if c.browser != nil {
		c.browser.Close()
		c.browser = nil
	}

	if c.playwright != nil {
		c.playwright.Stop()
		c.playwright = nil
	}

	c.initialized = false
	return nil
}

func InstallBrowsers() error {
	_, err := playwright.Run()
	if err != nil {
		return fmt.Errorf("failed to run playwright: %w", err)
	}

	if err := playwright.Install(); err != nil {
		return fmt.Errorf("failed to install browsers: %w", err)
	}
	return nil
}

func EnsureBrowsers() error {
	client := NewBrowserClient(WithInstallDeps(true))
	return client.ensureInitialized()
}
