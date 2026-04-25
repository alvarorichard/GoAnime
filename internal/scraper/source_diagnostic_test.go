package scraper

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPStatusErrorClassifiesCloudflareOriginDown(t *testing.T) {
	t.Parallel()

	err := NewHTTPStatusError("FlixHQ", "search", 521)
	diagnostic := DiagnoseError("FlixHQ", "search", err)

	require.NotNil(t, diagnostic)
	assert.Equal(t, DiagnosticSourceUnavailable, diagnostic.Kind)
	assert.Equal(t, 521, diagnostic.StatusCode)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))
	assert.True(t, diagnostic.ShouldSkipHealthCheck())
	assert.Contains(t, diagnostic.UserMessage(), "Cloudflare 521")
}

func TestNewHTTPStatusErrorClassifiesBlockedChallenge(t *testing.T) {
	t.Parallel()

	err := NewHTTPStatusError("SFlix", "search", http.StatusTooManyRequests)
	diagnostic := DiagnoseError("SFlix", "search", err)

	require.NotNil(t, diagnostic)
	assert.Equal(t, DiagnosticBlockedChallenge, diagnostic.Kind)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))
	assert.True(t, diagnostic.ShouldOpenCircuit())
}

func TestDiagnoseErrorClassifiesParserAndDecryptFailures(t *testing.T) {
	t.Parallel()

	parserDiagnostic := DiagnoseError("Goyabu", "episode", errors.New("no video URL found in AJAX response"))
	require.NotNil(t, parserDiagnostic)
	assert.Equal(t, DiagnosticParserBroken, parserDiagnostic.Kind)
	assert.False(t, parserDiagnostic.ShouldSkipHealthCheck())

	decryptDiagnostic := DiagnoseError("AllAnime", "episode", errors.New("AES-GCM decryption failed: cipher: message authentication failed"))
	require.NotNil(t, decryptDiagnostic)
	assert.Equal(t, DiagnosticDecryptBroken, decryptDiagnostic.Kind)
	assert.False(t, decryptDiagnostic.ShouldSkipHealthCheck())
}

func TestDiagnoseErrorClassifiesTimeoutAsSourceUnavailable(t *testing.T) {
	t.Parallel()

	diagnostic := DiagnoseError("9Anime", "search", context.DeadlineExceeded)

	require.NotNil(t, diagnostic)
	assert.Equal(t, DiagnosticSourceUnavailable, diagnostic.Kind)
	assert.True(t, errors.Is(diagnostic, ErrSourceUnavailable))
	assert.True(t, diagnostic.ShouldSkipHealthCheck())
}

func TestSourceCircuitBreakerSkipsAfterRepeatedOriginFailures(t *testing.T) {
	unavailableErr := NewHTTPStatusError("AllAnime", "search", 521)
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

	for i := 0; i < 2; i++ {
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
					return nil, NewHTTPStatusError("AllAnime", "search", 521)
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
	assert.Equal(t, DiagnosticSourceUnavailable, offline.Diagnostic.Kind)

	parserBroken := manager.CheckSourceHealth(context.Background(), AnimefireType, "naruto")
	assert.Equal(t, SourceHealthFailed, parserBroken.Status)
	require.NotNil(t, parserBroken.Diagnostic)
	assert.Equal(t, DiagnosticParserBroken, parserBroken.Diagnostic.Kind)
}
