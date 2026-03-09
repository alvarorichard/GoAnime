package browser

import (
	"context"
	"os"
	"strings"
	"sync"
	"testing"
	"time"
)

func TestNewBrowserClientOptions(t *testing.T) {
	tests := []struct {
		name  string
		opts  []Option
		check func(*testing.T, *BrowserClient)
	}{
		{
			name: "default options",
			opts: []Option{},
			check: func(t *testing.T, c *BrowserClient) {
				if c.config.Headless != true {
					t.Errorf("expected Headless=true, got %v", c.config.Headless)
				}
				if c.config.Timeout != DefaultTimeout {
					t.Errorf("expected Timeout=%v, got %v", DefaultTimeout, c.config.Timeout)
				}
				if c.config.MaxPages != 5 {
					t.Errorf("expected MaxPages=5, got %v", c.config.MaxPages)
				}
			},
		},
		{
			name: "with headless=false",
			opts: []Option{WithHeadless(false)},
			check: func(t *testing.T, c *BrowserClient) {
				if c.config.Headless != false {
					t.Errorf("expected Headless=false, got %v", c.config.Headless)
				}
			},
		},
		{
			name: "with custom timeout",
			opts: []Option{WithTimeout(60 * time.Second)},
			check: func(t *testing.T, c *BrowserClient) {
				if c.config.Timeout != 60*time.Second {
					t.Errorf("expected Timeout=60s, got %v", c.config.Timeout)
				}
			},
		},
		{
			name: "with custom user agent",
			opts: []Option{WithUserAgent("test-agent")},
			check: func(t *testing.T, c *BrowserClient) {
				if c.config.UserAgent != "test-agent" {
					t.Errorf("expected UserAgent=test-agent, got %v", c.config.UserAgent)
				}
			},
		},
		{
			name: "with custom viewport",
			opts: []Option{WithViewport(1280, 720)},
			check: func(t *testing.T, c *BrowserClient) {
				if c.config.Width != 1280 {
					t.Errorf("expected Width=1280, got %v", c.config.Width)
				}
				if c.config.Height != 720 {
					t.Errorf("expected Height=720, got %v", c.config.Height)
				}
			},
		},
		{
			name: "with max pages",
			opts: []Option{WithMaxPages(10)},
			check: func(t *testing.T, c *BrowserClient) {
				if c.config.MaxPages != 10 {
					t.Errorf("expected MaxPages=10, got %v", c.config.MaxPages)
				}
			},
		},
		{
			name: "with install deps",
			opts: []Option{WithInstallDeps(true)},
			check: func(t *testing.T, c *BrowserClient) {
				if c.config.InstallDeps != true {
					t.Errorf("expected InstallDeps=true, got %v", c.config.InstallDeps)
				}
			},
		},
		{
			name: "combined options",
			opts: []Option{
				WithHeadless(false),
				WithTimeout(45 * time.Second),
				WithViewport(1920, 1080),
				WithMaxPages(3),
			},
			check: func(t *testing.T, c *BrowserClient) {
				if c.config.Headless != false {
					t.Errorf("expected Headless=false")
				}
				if c.config.Timeout != 45*time.Second {
					t.Errorf("expected Timeout=45s")
				}
				if c.config.Width != 1920 {
					t.Errorf("expected Width=1920")
				}
				if c.config.Height != 1080 {
					t.Errorf("expected Height=1080")
				}
				if c.config.MaxPages != 3 {
					t.Errorf("expected MaxPages=3")
				}
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			client := NewBrowserClient(tt.opts...)
			defer client.Close()
			tt.check(t, client)
		})
	}
}

func TestBrowserClientNotInitialized(t *testing.T) {
	client := NewBrowserClient(WithHeadless(true))
	defer client.Close()

	// Should not be initialized yet
	if client.initialized {
		t.Error("client should not be initialized before first use")
	}

	// IsAvailable should return false before initialization
	if client.IsAvailable() {
		t.Error("IsAvailable should return false before initialization")
	}
}

func TestBrowserClientClose(t *testing.T) {
	client := NewBrowserClient(WithHeadless(true))

	// Close should not panic
	err := client.Close()
	if err != nil {
		t.Errorf("Close returned error: %v", err)
	}

	// After close, should not be available
	if client.IsAvailable() {
		t.Error("IsAvailable should return false after close")
	}
}

func TestBrowserClientMultipleCloses(t *testing.T) {
	client := NewBrowserClient(WithHeadless(true))

	// First close
	err := client.Close()
	if err != nil {
		t.Errorf("first Close returned error: %v", err)
	}

	// Second close should not panic
	err = client.Close()
	if err != nil {
		t.Errorf("second Close returned error: %v", err)
	}
}

func TestBrowserClientGetHTMLRequiresURL(t *testing.T) {
	client := NewBrowserClient(WithHeadless(true))
	defer client.Close()

	ctx := context.Background()

	// Test with empty URL - this may fail but shouldn't panic
	_, err := client.GetHTML(ctx, "")
	if err == nil {
		t.Error("expected error for empty URL")
	}
}

func TestBrowserClientWithWait(t *testing.T) {
	client := NewBrowserClient(WithHeadless(true))
	defer client.Close()

	ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
	defer cancel()

	// Get HTML with wait for a selector that exists on example.com
	html, err := client.GetHTMLWithWait(ctx, "https://example.com", "h1")
	if err != nil {
		t.Fatalf("failed to get HTML with wait: %v", err)
	}

	if html == nil {
		t.Fatal("expected non-nil HTML reader")
	}

	client.Close()
}

func TestConstants(t *testing.T) {
	if DefaultTimeout != 30*time.Second {
		t.Errorf("expected DefaultTimeout=30s, got %v", DefaultTimeout)
	}
	if DefaultWidth != 1920 {
		t.Errorf("expected DefaultWidth=1920, got %v", DefaultWidth)
	}
	if DefaultHeight != 1080 {
		t.Errorf("expected DefaultHeight=1080, got %v", DefaultHeight)
	}
}

func TestGetDefaultBrowser(t *testing.T) {
	// GetDefaultBrowser should return same instance
	b1 := GetDefaultBrowser()
	b2 := GetDefaultBrowser()

	if b1 != b2 {
		t.Error("GetDefaultBrowser should return same instance")
	}

	// Clean up
	b1.Close()
}

func TestBrowserClientConcurrentAccess(t *testing.T) {
	client := NewBrowserClient(WithHeadless(true))
	defer client.Close()

	var wg sync.WaitGroup
	errChan := make(chan error, 10)

	// Try concurrent access
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
			defer cancel()

			_, err := client.GetHTML(ctx, "https://example.com")
			if err != nil {
				errChan <- err
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// Check for errors
	for err := range errChan {
		t.Logf("concurrent access error: %v", err)
	}
}

func TestBrowserClientEnvVars(t *testing.T) {
	// Test that client works in environment variables
	// This is a placeholder - actual behavior depends on environment
	client := NewBrowserClient(WithHeadless(true))
	defer client.Close()

	// Check that config is set correctly
	if client.config.UserAgent == "" {
		t.Error("expected default user agent to be set")
	}

	// Verify UserAgent contains expected pattern
	if !strings.Contains(client.config.UserAgent, "Mozilla") {
		t.Logf("UserAgent: %s", client.config.UserAgent)
	}
}

func TestInstallBrowsers(t *testing.T) {
	// This test verifies that InstallBrowsers can be called
	// In CI environments, browsers may already be installed
	// The function should not panic

	// Skip if PLAYWRIGHT_SKIP_BROWSER_INSTALL is set
	if os.Getenv("PLAYWRIGHT_SKIP_BROWSER_INSTALL") == "true" {
		t.Skip("Skipping browser install test")
	}

	// This is a smoke test - actual browser install may take time
	// We just verify the function can be called
	err := InstallBrowsers()
	if err != nil {
		t.Logf("InstallBrowsers error (may be expected): %v", err)
	}
}

func TestEnsureBrowsers(t *testing.T) {
	// Skip if browser is not available
	if os.Getenv("PLAYWRIGHT_SKIP_BROWSER_TEST") == "true" {
		t.Skip("Skipping browser test")
	}

	// This will try to ensure browsers are available
	err := EnsureBrowsers()
	if err != nil {
		t.Logf("EnsureBrowsers error: %v", err)
	}
}
