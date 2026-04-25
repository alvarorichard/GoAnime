//go:build sourcehealth

package scraper

import (
	"context"
	"testing"
	"time"
)

func TestSourceHealthLive(t *testing.T) {
	manager := NewScraperManager()

	for _, source := range manager.AvailableSources() {
		source := source
		t.Run(manager.getScraperDisplayName(source), func(t *testing.T) {
			ctx, cancel := context.WithTimeout(context.Background(), 20*time.Second)
			defer cancel()

			result := manager.CheckSourceHealth(ctx, source, DefaultHealthCheckQuery(source))
			switch result.Status {
			case SourceHealthHealthy:
				t.Logf("%s", result.Description)
			case SourceHealthSkipped:
				t.Skipf("%s", result.Description)
			case SourceHealthFailed:
				t.Fatalf("%s", result.Description)
			default:
				t.Fatalf("unknown source health status: %s", result.Status)
			}
		})
	}
}
