package scraper

import (
	"errors"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// MockScraper implements UnifiedScraper for testing
type MockScraper struct {
	searchFunc      func(query string) ([]*models.Anime, error)
	episodesFunc    func(url string) ([]models.Episode, error)
	streamURLFunc   func(url string) (string, map[string]string, error)
	scraperType     ScraperType
	searchCallCount atomic.Int32
	searchDelay     time.Duration
}

func (m *MockScraper) SearchAnime(query string, options ...interface{}) ([]*models.Anime, error) {
	m.searchCallCount.Add(1)
	if m.searchDelay > 0 {
		time.Sleep(m.searchDelay)
	}
	if m.searchFunc != nil {
		return m.searchFunc(query)
	}
	return nil, nil
}

func (m *MockScraper) GetAnimeEpisodes(animeURL string) ([]models.Episode, error) {
	if m.episodesFunc != nil {
		return m.episodesFunc(animeURL)
	}
	return nil, nil
}

func (m *MockScraper) GetStreamURL(episodeURL string, options ...interface{}) (string, map[string]string, error) {
	if m.streamURLFunc != nil {
		return m.streamURLFunc(episodeURL)
	}
	return "", nil, nil
}

func (m *MockScraper) GetType() ScraperType {
	return m.scraperType
}

// createTestManager creates a ScraperManager with mock scrapers
func createTestManager(allAnimeMock, animefireMock *MockScraper) *ScraperManager {
	manager := &ScraperManager{
		scrapers: make(map[ScraperType]UnifiedScraper),
	}
	if allAnimeMock != nil {
		allAnimeMock.scraperType = AllAnimeType
		manager.scrapers[AllAnimeType] = allAnimeMock
	}
	if animefireMock != nil {
		animefireMock.scraperType = AnimefireType
		manager.scrapers[AnimefireType] = animefireMock
	}
	return manager
}

// =============================================================================
// Test: Both sources return results successfully
// =============================================================================

func TestSearchAnime_BothSourcesSucceed(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{
				{Name: "Naruto", URL: "allanime-naruto-id"},
				{Name: "Naruto Shippuden", URL: "allanime-shippuden-id"},
			}, nil
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{
				{Name: "Naruto", URL: "https://animefire.io/anime/naruto"},
				{Name: "Naruto Classico", URL: "https://animefire.io/anime/naruto-classico"},
			}, nil
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)
	results, err := manager.SearchAnime("naruto", nil)

	require.NoError(t, err)
	assert.Len(t, results, 4, "Should have results from both sources")

	// Verify both scrapers were called
	assert.Equal(t, int32(1), allAnimeMock.searchCallCount.Load())
	assert.Equal(t, int32(1), animefireMock.searchCallCount.Load())

	// Verify source tags are added
	allAnimeCount := 0
	animefireCount := 0
	for _, anime := range results {
		switch anime.Source {
		case "AllAnime":
			allAnimeCount++
			assert.Contains(t, anime.Name, "[AllAnime]")
		case "AnimeFire.plus":
			animefireCount++
			assert.Contains(t, anime.Name, "[AnimeFire]")
		}
	}
	assert.Equal(t, 2, allAnimeCount)
	assert.Equal(t, 2, animefireCount)
}

// =============================================================================
// Test: AnimeFire fails, AllAnime succeeds (Portuguese results missing)
// =============================================================================

func TestSearchAnime_AnimefireFails_AllAnimeSucceeds(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{
				{Name: "Naruto", URL: "allanime-naruto-id"},
			}, nil
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return nil, errors.New("animefire returned a challenge page (try VPN or wait)")
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)
	results, err := manager.SearchAnime("naruto", nil)

	// Should still return results from AllAnime
	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "AllAnime", results[0].Source)

	// Both scrapers should have been called
	assert.Equal(t, int32(1), allAnimeMock.searchCallCount.Load())
	assert.Equal(t, int32(1), animefireMock.searchCallCount.Load())
}

// =============================================================================
// Test: AllAnime fails, AnimeFire succeeds
// =============================================================================

func TestSearchAnime_AllAnimeFails_AnimefireSucceeds(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return nil, errors.New("connection timeout")
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{
				{Name: "Naruto", URL: "https://animefire.io/anime/naruto"},
			}, nil
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)
	results, err := manager.SearchAnime("naruto", nil)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "AnimeFire.plus", results[0].Source)
}

// =============================================================================
// Test: Both sources fail
// =============================================================================

func TestSearchAnime_BothSourcesFail(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return nil, errors.New("API rate limited")
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return nil, errors.New("challenge page detected")
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)
	results, err := manager.SearchAnime("naruto", nil)

	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "no anime found")
	assert.Contains(t, err.Error(), "some sources failed")
}

// =============================================================================
// Test: Both sources return empty results
// =============================================================================

func TestSearchAnime_BothSourcesReturnEmpty(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{}, nil
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{}, nil
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)
	_, err := manager.SearchAnime("xyznonexistent", nil)

	require.Error(t, err)
	assert.Contains(t, err.Error(), "no anime found")
	// Should not mention failed sources since they didn't fail
	assert.NotContains(t, err.Error(), "some sources failed")
}

// =============================================================================
// Test: One source returns empty, other returns results
// =============================================================================

func TestSearchAnime_OneSourceEmpty_OtherHasResults(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{}, nil // Empty but no error
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{
				{Name: "Anime Brasileiro", URL: "https://animefire.io/anime/brasileiro"},
			}, nil
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)
	results, err := manager.SearchAnime("brasileiro", nil)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "AnimeFire.plus", results[0].Source)
}

// =============================================================================
// Test: Concurrent execution - both scrapers run in parallel
// =============================================================================

func TestSearchAnime_ConcurrentExecution(t *testing.T) {
	t.Parallel()

	var allAnimeStart, animefireStart time.Time
	var mu sync.Mutex

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			mu.Lock()
			allAnimeStart = time.Now()
			mu.Unlock()

			time.Sleep(100 * time.Millisecond)

			return []*models.Anime{{Name: "AllAnime Result", URL: "id1"}}, nil
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			mu.Lock()
			animefireStart = time.Now()
			mu.Unlock()

			time.Sleep(100 * time.Millisecond)

			return []*models.Anime{{Name: "AnimeFire Result", URL: "https://animefire.io/1"}}, nil
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)

	start := time.Now()
	results, err := manager.SearchAnime("test", nil)
	totalDuration := time.Since(start)

	require.NoError(t, err)
	assert.Len(t, results, 2)

	// If running concurrently, total time should be ~100ms, not ~200ms
	// Allow some buffer for test environment variations
	assert.Less(t, totalDuration, 180*time.Millisecond,
		"Searches should run concurrently, not sequentially")

	// Verify both started around the same time (within 50ms of each other)
	mu.Lock()
	startDiff := allAnimeStart.Sub(animefireStart)
	if startDiff < 0 {
		startDiff = -startDiff
	}
	mu.Unlock()

	assert.Less(t, startDiff, 50*time.Millisecond,
		"Both searches should start nearly simultaneously")
}

// =============================================================================
// Test: Slow source doesn't block fast source results
// =============================================================================

func TestSearchAnime_SlowSourceDoesNotBlockFastSource(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			time.Sleep(200 * time.Millisecond) // Slow
			return []*models.Anime{{Name: "Slow Result", URL: "id1"}}, nil
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			time.Sleep(10 * time.Millisecond) // Fast
			return []*models.Anime{{Name: "Fast Result", URL: "https://animefire.io/1"}}, nil
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)
	results, err := manager.SearchAnime("test", nil)

	require.NoError(t, err)
	// Both results should be present
	assert.Len(t, results, 2)
}

// =============================================================================
// Test: Specific scraper selection - AnimeFire only
// =============================================================================

func TestSearchAnime_SpecificScraper_AnimefireOnly(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "AllAnime Result", URL: "id1"}}, nil
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "AnimeFire Result", URL: "https://animefire.io/1"}}, nil
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)

	scraperType := AnimefireType
	results, err := manager.SearchAnime("test", &scraperType)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "AnimeFire.plus", results[0].Source)

	// Only AnimeFire should be called
	assert.Equal(t, int32(0), allAnimeMock.searchCallCount.Load())
	assert.Equal(t, int32(1), animefireMock.searchCallCount.Load())
}

// =============================================================================
// Test: Specific scraper selection - AllAnime only
// =============================================================================

func TestSearchAnime_SpecificScraper_AllAnimeOnly(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "AllAnime Result", URL: "id1"}}, nil
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "AnimeFire Result", URL: "https://animefire.io/1"}}, nil
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)

	scraperType := AllAnimeType
	results, err := manager.SearchAnime("test", &scraperType)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "AllAnime", results[0].Source)

	// Only AllAnime should be called
	assert.Equal(t, int32(1), allAnimeMock.searchCallCount.Load())
	assert.Equal(t, int32(0), animefireMock.searchCallCount.Load())
}

// =============================================================================
// Test: Specific scraper fails - returns error
// =============================================================================

func TestSearchAnime_SpecificScraper_Fails(t *testing.T) {
	t.Parallel()

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return nil, errors.New("Cloudflare challenge")
		},
	}

	manager := createTestManager(nil, animefireMock)

	scraperType := AnimefireType
	results, err := manager.SearchAnime("test", &scraperType)

	require.Error(t, err)
	assert.Nil(t, results)
	assert.Contains(t, err.Error(), "busca falhou")
	assert.Contains(t, err.Error(), "Cloudflare challenge")
}

// =============================================================================
// Test: Source tags are not duplicated
// =============================================================================

func TestSearchAnime_SourceTagsNotDuplicated(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{
				{Name: "[AllAnime] Naruto", URL: "id1"}, // Already has tag
			}, nil
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{
				{Name: "[AnimeFire] Naruto", URL: "https://animefire.io/1"}, // Already has tag
			}, nil
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)
	results, err := manager.SearchAnime("naruto", nil)

	require.NoError(t, err)

	for _, anime := range results {
		// Count occurrences of tags
		allAnimeTagCount := countOccurrences(anime.Name, "[AllAnime]")
		animefireTagCount := countOccurrences(anime.Name, "[AnimeFire]")

		// Should never have more than one of each tag
		assert.LessOrEqual(t, allAnimeTagCount, 1, "AllAnime tag duplicated")
		assert.LessOrEqual(t, animefireTagCount, 1, "AnimeFire tag duplicated")
	}
}

// =============================================================================
// Test: Race condition - multiple concurrent searches
// =============================================================================

func TestSearchAnime_NoConcurrentRaceConditions(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			time.Sleep(10 * time.Millisecond)
			return []*models.Anime{{Name: "Result " + query, URL: "id-" + query}}, nil
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			time.Sleep(10 * time.Millisecond)
			return []*models.Anime{{Name: "AF Result " + query, URL: "https://animefire.io/" + query}}, nil
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)

	// Run multiple concurrent searches
	var wg sync.WaitGroup
	errChan := make(chan error, 10)

	for i := 0; i < 10; i++ {
		wg.Add(1)
		go func(idx int) {
			defer wg.Done()
			results, err := manager.SearchAnime("query", nil)
			if err != nil {
				errChan <- err
				return
			}
			if len(results) != 2 {
				errChan <- errors.New("unexpected result count")
			}
		}(i)
	}

	wg.Wait()
	close(errChan)

	for err := range errChan {
		t.Errorf("Concurrent search error: %v", err)
	}
}

// =============================================================================
// Test: Network timeout simulation
// =============================================================================

func TestSearchAnime_NetworkTimeout(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "Quick Result", URL: "id1"}}, nil
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			// Simulate network timeout error (not actual timeout, just error)
			return nil, errors.New("connection timeout after 30s")
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)
	results, err := manager.SearchAnime("test", nil)

	require.NoError(t, err)
	assert.Len(t, results, 1)
	assert.Equal(t, "AllAnime", results[0].Source)
}

// =============================================================================
// Test: VPN required error from AnimeFire
// =============================================================================

func TestSearchAnime_VPNRequired(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "English Result", URL: "id1"}}, nil
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return nil, errors.New("access restricted: VPN may be required")
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)
	results, err := manager.SearchAnime("test", nil)

	require.NoError(t, err, "Should return results from working source")
	assert.Len(t, results, 1)
}

// =============================================================================
// Test: Cloudflare challenge detection
// =============================================================================

func TestSearchAnime_CloudflareChallenge(t *testing.T) {
	t.Parallel()

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "Result", URL: "id1"}}, nil
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			return nil, errors.New("animefire returned a challenge page (try VPN or wait)")
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)
	results, err := manager.SearchAnime("test", nil)

	require.NoError(t, err)
	assert.Len(t, results, 1)
}

// =============================================================================
// Test: Query is passed correctly to scrapers
// =============================================================================

func TestSearchAnime_QueryPassedCorrectly(t *testing.T) {
	t.Parallel()

	var capturedQueries []string
	var mu sync.Mutex

	allAnimeMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			mu.Lock()
			capturedQueries = append(capturedQueries, "allanime:"+query)
			mu.Unlock()
			return []*models.Anime{}, nil
		},
	}

	animefireMock := &MockScraper{
		searchFunc: func(query string) ([]*models.Anime, error) {
			mu.Lock()
			capturedQueries = append(capturedQueries, "animefire:"+query)
			mu.Unlock()
			return []*models.Anime{}, nil
		},
	}

	manager := createTestManager(allAnimeMock, animefireMock)
	_, _ = manager.SearchAnime("Shingeki no Kyojin", nil)

	mu.Lock()
	defer mu.Unlock()

	assert.Len(t, capturedQueries, 2)
	assert.Contains(t, capturedQueries, "allanime:Shingeki no Kyojin")
	assert.Contains(t, capturedQueries, "animefire:Shingeki no Kyojin")
}

// =============================================================================
// Helper functions
// =============================================================================

func countOccurrences(s, substr string) int {
	count := 0
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			count++
		}
	}
	return count
}
