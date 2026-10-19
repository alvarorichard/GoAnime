package providers

import (
	"context"
	"math/rand"
	"sync"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/api/source"
	"github.com/alvarorichard/Goanime/internal/models"
	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// contextualStubSource honours context cancellation during search.
type contextualStubSource struct {
	epStubSource
	results []*models.Anime
	delay   time.Duration
}
func (s *contextualStubSource) Search(ctx context.Context, _ string) ([]*models.Anime, error) {
	if s.delay <= 0 {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		default:
			return s.results, nil
		}
	}
	select {
	case <-ctx.Done():
		return nil, ctx.Err()
	case <-time.After(s.delay):
		return s.results, nil
	}
}

// TestSearch_StressCancellationAndTimeout_Race stresses SearchAll and SearchLive
// under concurrent execution with mixed timeouts, cancellations, and completions
// to verify race-freedom, error preservation, and circuit breaker isolation.
func TestSearch_StressCancellationAndTimeout_Race(t *testing.T) {
	prevBreaker := searchBreaker
	searchBreaker = netx.NewCircuitBreaker()
	t.Cleanup(func() { searchBreaker = prevBreaker })

	prevTimeout := perSourceSearchTimeout
	perSourceSearchTimeout = 50 * time.Millisecond
	t.Cleanup(func() { perSourceSearchTimeout = prevTimeout })

	fastSrc := &contextualStubSource{
		epStubSource: epStubSource{desc: source.Descriptor{Kind: source.AnimeFire, Priority: 1}},
		results:      []*models.Anime{{Name: "Naruto"}},
		delay:        5 * time.Millisecond,
	}
	slowSrc := &contextualStubSource{
		epStubSource: epStubSource{desc: source.Descriptor{Kind: source.AnimeDrive, Priority: 2}},
		results:      []*models.Anime{{Name: "Naruto Shippuden"}},
		delay:        30 * time.Millisecond,
	}
	hangingSrc := &hangingSearchSource{
		epStubSource: epStubSource{desc: source.Descriptor{Kind: source.Goyabu, Priority: 3}},
		release:      make(chan struct{}),
	}
	t.Cleanup(func() { close(hangingSrc.release) })

	restore := source.SwapRegistryForTesting(fastSrc, slowSrc, hangingSrc)
	t.Cleanup(restore)

	const concurrency = 20
	const iterations = 5

	var wg sync.WaitGroup
	for worker := 0; worker < concurrency; worker++ {
		wg.Add(1)
		go func(w int) {
			defer wg.Done()
			rng := rand.New(rand.NewSource(int64(w*1000 + 42)))

			for i := 0; i < iterations; i++ {
				mode := rng.Intn(4)
				var ctx context.Context
				var cancel context.CancelFunc

				switch mode {
				case 0:
					// Pre-cancelled context
					ctx, cancel = context.WithCancel(context.Background())
					cancel()
				case 1:
					// Quick cancellation while running
					ctx, cancel = context.WithCancel(context.Background())
					delay := time.Duration(rng.Intn(10)+1) * time.Millisecond
					time.AfterFunc(delay, cancel)
				case 2:
					// Short timeout that may beat or lose to fastSrc
					ctx, cancel = context.WithTimeout(context.Background(), time.Duration(rng.Intn(40)+10)*time.Millisecond)
				case 3:
					// Normal context allowing fastSrc to complete
					ctx, cancel = context.WithTimeout(context.Background(), 200*time.Millisecond)
				}

				if w%2 == 0 {
					results, err := SearchAll(ctx, "naruto")
					if err != nil {
						if mode == 0 {
							require.ErrorIs(t, err, context.Canceled)
						}
					} else {
						assert.NotEmpty(t, results)
					}
				} else {
					var events []models.SearchEvent
					var mu sync.Mutex
					results, err := SearchLive(ctx, "naruto", nil, func(ev models.SearchEvent) {
						mu.Lock()
						events = append(events, ev)
						mu.Unlock()
					})
					if err != nil {
						if mode == 0 {
							require.ErrorIs(t, err, context.Canceled)
						}
					} else {
						assert.NotEmpty(t, results)
					}
					mu.Lock()
					assert.NotEmpty(t, events)
					mu.Unlock()
				}
				cancel()
			}
		}(worker)
	}

	wg.Wait()
}
