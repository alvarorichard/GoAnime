package scraper

// Regression suite for the "search held hostage by a dead source" slowdown.
//
// Discovered: 2026-07-01 — user debug logs showed every search taking the full
//             12s perScraperTimeout because Goyabu hung, while SuperFlix and
//             AnimeFire had answered within ~1s.
// Fix:        searchAllScrapersConcurrent now arms a stragglerGrace window
//             once the FIRST source delivers results; remaining sources get
//             only that grace before the aggregate returns what it has.

import (
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// shrinkStragglerGrace overrides the package-level grace window for the test.
// Mutates a global, so callers must NOT use t.Parallel().
func shrinkStragglerGrace(t *testing.T, d time.Duration) {
	t.Helper()
	orig := stragglerGrace
	stragglerGrace = d
	t.Cleanup(func() { stragglerGrace = orig })
}

func TestSearchAllScrapersConcurrent_StragglerGraceCutsHangingSource(t *testing.T) {
	shrinkStragglerGrace(t, 300*time.Millisecond)

	fast := &MockScraper{
		searchFunc: func(string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "Naruto", URL: "fast-id"}}, nil
		},
	}
	// Hangs far beyond the grace window (but below perScraperTimeout, i.e. a
	// "slow but not yet failed" source).
	slow := &MockScraper{
		searchDelay: 5 * time.Second,
		searchFunc: func(string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "Naruto Slow", URL: "slow-id"}}, nil
		},
	}
	manager := createTestManager(fast, slow)

	start := time.Now()
	results, err := manager.searchAllScrapersConcurrent("naruto")
	elapsed := time.Since(start)

	require.NoError(t, err)
	require.Len(t, results, 1, "only the fast source's results should be in")
	assert.Equal(t, "fast-id", results[0].URL)
	assert.Less(t, elapsed, 3*time.Second,
		"search must return ~grace after the first results, not wait out the hanging source")
}

func TestSearchAllScrapersConcurrent_GraceOnlyArmsAfterFirstResults(t *testing.T) {
	shrinkStragglerGrace(t, 100*time.Millisecond)

	// The ONLY source is slower than the grace window. Since the grace must
	// arm only after the first results arrive, this source must still be
	// waited for — a search with zero results so far has nothing to return
	// early with.
	slow := &MockScraper{
		searchDelay: 600 * time.Millisecond,
		searchFunc: func(string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "Bleach", URL: "slow-only-id"}}, nil
		},
	}
	manager := createTestManager(slow, nil)

	results, err := manager.searchAllScrapersConcurrent("bleach")

	require.NoError(t, err)
	require.Len(t, results, 1)
	assert.Equal(t, "slow-only-id", results[0].URL)
}

func TestSearchAllScrapersConcurrent_GraceStillCollectsInFlightResults(t *testing.T) {
	shrinkStragglerGrace(t, 700*time.Millisecond)

	fast := &MockScraper{
		searchFunc: func(string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "One Piece", URL: "fast-id"}}, nil
		},
	}
	// Finishes INSIDE the grace window — its results must still be merged.
	slower := &MockScraper{
		searchDelay: 200 * time.Millisecond,
		searchFunc: func(string) ([]*models.Anime, error) {
			return []*models.Anime{{Name: "One Piece Film", URL: "slower-id"}}, nil
		},
	}
	manager := createTestManager(fast, slower)

	results, err := manager.searchAllScrapersConcurrent("one piece")

	require.NoError(t, err)
	assert.Len(t, results, 2, "sources answering within the grace window must not be dropped")
}
