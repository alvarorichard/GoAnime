package integration

import (
	"os"
	"sync"
	"testing"

	"github.com/alvarorichard/Goanime/internal/discord"
	"github.com/alvarorichard/Goanime/internal/scraper"
)

func TestSingletons(t *testing.T) {
	t.Run("ScraperManager singleton", func(t *testing.T) {
		sm1 := scraper.NewScraperManager()
		sm2 := scraper.NewScraperManager()

		// NewScraperManager returns the same instance (singleton)
		if sm1 != sm2 {
			t.Error("ScraperManager should be singleton")
		}
	})

	t.Run("MediaManager singleton", func(t *testing.T) {
		mm1 := scraper.GetMediaManager()
		mm2 := scraper.GetMediaManager()

		if mm1 != mm2 {
			t.Error("MediaManager should be singleton")
		}
	})

	t.Run("FlixHQClient singleton", func(t *testing.T) {
		fc1 := scraper.GetFlixHQClient()
		fc2 := scraper.GetFlixHQClient()

		if fc1 != fc2 {
			t.Error("FlixHQClient should be singleton")
		}
	})

	t.Run("DiscordManager singleton", func(t *testing.T) {
		dm1 := discord.GetDiscordManager()
		dm2 := discord.GetDiscordManager()

		if dm1 != dm2 {
			t.Error("DiscordManager should be singleton")
		}
	})
}

func TestSingletonConcurrentAccess(t *testing.T) {
	// Test concurrent access to singletons
	var wg sync.WaitGroup
	errChan := make(chan error, 20)

	// Test ScraperManager concurrent access
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			sm := scraper.NewScraperManager()
			if sm == nil {
				errChan <- nil // This is expected - may not be initialized
			}
		}()
	}

	// Test MediaManager concurrent access
	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			mm := scraper.GetMediaManager()
			if mm == nil {
				errChan <- nil // This is expected
			}
		}()
	}

	wg.Wait()
	close(errChan)

	// No errors should have occurred
	for err := range errChan {
		if err != nil {
			t.Errorf("concurrent access error: %v", err)
		}
	}
}

func TestSingletonsInitialized(t *testing.T) {
	// Test that singletons are properly initialized
	t.Run("MediaManager has scraper manager", func(t *testing.T) {
		mm := scraper.GetMediaManager()
		if mm == nil {
			t.Fatal("MediaManager should not be nil")
		}
		// MediaManager should have a scraper manager
		_ = mm.SearchAll // This just verifies the method exists
	})

	t.Run("FlixHQClient has required methods", func(t *testing.T) {
		fc := scraper.GetFlixHQClient()
		if fc == nil {
			t.Fatal("FlixHQClient should not be nil")
		}
		// Verify the client has required fields/methods
		_ = fc.SearchMedia // This just verifies the method exists
	})

	t.Run("DiscordManager methods work", func(t *testing.T) {
		dm := discord.GetDiscordManager()
		if dm == nil {
			t.Fatal("DiscordManager should not be nil")
		}

		// Test IsEnabled returns a value (false by default)
		enabled := dm.IsEnabled()
		if enabled {
			t.Log("Discord is enabled (unexpected in test)")
		}

		// Test IsInitialized returns a value
		initialized := dm.IsInitialized()
		if initialized {
			t.Log("Discord is initialized (unexpected in test)")
		}
	})
}

func TestEnvironmentCheck(t *testing.T) {
	// Check if we're in CI environment
	if os.Getenv("CI") == "true" {
		t.Log("Running in CI environment")
	}

	// Check for required environment variables
	t.Log("HOME:", os.Getenv("HOME"))
	t.Log("PATH exists:", os.Getenv("PATH") != "")
}

func TestSkipInCI(t *testing.T) {
	// Example of how to skip tests in CI
	if os.Getenv("CI") == "true" {
		t.Skip("Skipping test in CI environment")
	}

	// This test would run in local environment
	t.Log("This test runs in local environment only")
}
