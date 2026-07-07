package scraper

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSourceCircuitBreakerSkipsAfterRepeatedOriginFailures(t *testing.T) {
	unavailableErr := netx.NewHTTPStatusError("AllAnime", "search", 521)
	allAnimeMock := &MockScraper{
		searchFunc: func(_ string) ([]*models.Anime, error) {
			return nil, unavailableErr
		},
	}
	animefireMock := &MockScraper{
		searchFunc: func(_ string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "Naruto", URL: "ok"}}, nil
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)
	manager.breaker = newSourceCircuitBreaker()
	manager.breaker.threshold = 2
	manager.breaker.cooldown = time.Minute

	for range 2 {
		results, err := manager.SearchAnime("naruto", nil)
		require.NoError(t, err)
		require.Len(t, results, 1)
	}

	results, err := manager.SearchAnime("naruto", nil)
	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, int32(2), allAnimeMock.searchCallCount.Load(), "open circuit should skip the failing source")
	assert.Equal(t, int32(3), animefireMock.searchCallCount.Load())
}

func TestCheckSourceHealthFailsOnParserBreakButSkipsOffline(t *testing.T) {
	t.Parallel()

	manager := &ScraperManager{
		scrapers: map[ScraperType]UnifiedScraper{
			AllAnimeType: &MockScraper{
				scraperType: AllAnimeType,
				searchFunc: func(_ string) ([]*models.Anime, error) {
					return nil, netx.NewHTTPStatusError("AllAnime", "search", 521)
				},
			},
			AnimefireType: &MockScraper{
				scraperType: AnimefireType,
				searchFunc: func(_ string) ([]*models.Anime, error) {
					return nil, fmt.Errorf("no video URL found in AJAX response")
				},
			},
		},
		breaker: newSourceCircuitBreaker(),
	}

	offline := manager.CheckSourceHealth(context.Background(), AllAnimeType, "naruto")
	assert.Equal(t, SourceHealthSkipped, offline.Status)
	require.NotNil(t, offline.Diagnostic)
	assert.Equal(t, netx.DiagnosticSourceUnavailable, offline.Diagnostic.Kind)

	parserBroken := manager.CheckSourceHealth(context.Background(), AnimefireType, "naruto")
	assert.Equal(t, SourceHealthFailed, parserBroken.Status)
	require.NotNil(t, parserBroken.Diagnostic)
	assert.Equal(t, netx.DiagnosticParserBroken, parserBroken.Diagnostic.Kind)
}
