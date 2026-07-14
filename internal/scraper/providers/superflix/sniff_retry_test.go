package superflix

import (
	"context"
	"fmt"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// retrySolver satisfies BOTH embedStreamSolver (SniffEmbedStream) and cfSolver
// (Solve), so it can be used to unit-test sniffEmbedStreamWithRetry directly and
// to be assigned to SuperFlixClient.browserSolver for the integration path. Its
// SniffEmbedStream returns a scripted per-call sequence, so a test can drive
// fail→succeed and fail→fail paths.
type retrySolver struct {
	errs    []error
	results []*CFStreamResult
	calls   int
}

func (s *retrySolver) SniffEmbedStream(_ context.Context, _ string, _ time.Duration) (*CFStreamResult, error) {
	i := s.calls
	s.calls++
	if i < len(s.errs) && s.errs[i] != nil {
		return nil, s.errs[i]
	}
	if i < len(s.results) {
		return s.results[i], nil
	}
	return nil, fmt.Errorf("retrySolver: unexpected call %d", i)
}

// Solve is unused by the browser stream path (which drives SniffEmbedStream);
// it exists only so retrySolver satisfies cfSolver for browserSolver assignment.
func (s *retrySolver) Solve(context.Context, string, time.Duration) (*CFSolveResult, error) {
	return &CFSolveResult{}, nil
}

func TestSniffEmbedStreamWithRetry(t *testing.T) {
	t.Parallel()

	okRes := &CFStreamResult{StreamURL: "https://cdn.test/master.m3u8"}
	boom := fmt.Errorf("turnstile timeout")

	t.Run("succeeds on first attempt without retrying", func(t *testing.T) {
		t.Parallel()
		s := &retrySolver{results: []*CFStreamResult{okRes}}
		res, err := sniffEmbedStreamWithRetry(context.Background(), s, "https://embed/x")
		require.NoError(t, err)
		assert.Equal(t, okRes, res)
		assert.Equal(t, 1, s.calls, "a first-attempt success must not trigger a retry")
	})

	t.Run("retry rescues a transient first-attempt failure", func(t *testing.T) {
		t.Parallel()
		s := &retrySolver{
			errs:    []error{boom, nil},
			results: []*CFStreamResult{nil, okRes},
		}
		res, err := sniffEmbedStreamWithRetry(context.Background(), s, "https://embed/x")
		require.NoError(t, err)
		assert.Equal(t, okRes, res)
		assert.Equal(t, 2, s.calls, "a transient failure must be retried once")
	})

	t.Run("all attempts fail returns the last error", func(t *testing.T) {
		t.Parallel()
		last := fmt.Errorf("second boom")
		s := &retrySolver{errs: []error{boom, last}}
		_, err := sniffEmbedStreamWithRetry(context.Background(), s, "https://embed/x")
		require.Error(t, err)
		assert.Equal(t, last, err, "the last attempt's error must be returned")
		assert.Equal(t, sniffEmbedStreamAttempts, s.calls, "must exhaust all attempts")
	})

	t.Run("a cancelled context is not retried", func(t *testing.T) {
		t.Parallel()
		ctx, cancel := context.WithCancel(context.Background())
		cancel()
		s := &retrySolver{errs: []error{boom, boom}}
		_, err := sniffEmbedStreamWithRetry(ctx, s, "https://embed/x")
		require.Error(t, err)
		assert.Equal(t, 1, s.calls, "must not burn a second solve once the context is done")
	})

	// The "Acesso Restrito" shell is terminal — retrying it just burns another 90s
	// solve on a page that will never yield a stream. It must fail after ONE
	// attempt, and the sentinel must survive for the caller's friendly message.
	t.Run("the restricted-access error is terminal and not retried", func(t *testing.T) {
		t.Parallel()
		s := &retrySolver{errs: []error{
			fmt.Errorf("superflix embed stream sniff failed: %w", ErrSuperFlixRestricted),
			boom,
		}}
		_, err := sniffEmbedStreamWithRetry(context.Background(), s, "https://embed/x")
		require.Error(t, err)
		assert.ErrorIs(t, err, ErrSuperFlixRestricted)
		assert.Equal(t, 1, s.calls, "a restricted-access page must not be retried")
	})
}

// TestGetStreamURL_RetriesTransientSniffFailure is the integration counterpart:
// it proves the retry is actually wired into the public GetStreamURL browser
// path, not just the isolated helper. A solver that fails its first solve and
// succeeds on the second must yield a stream (no error) from GetStreamURL.
func TestGetStreamURL_RetriesTransientSniffFailure(t *testing.T) {
	// Not parallel: swaps the package-global stream cache.
	saved := defaultStreamCache
	t.Cleanup(func() { defaultStreamCache = saved })
	defaultStreamCache = &streamCache{path: filepath.Join(t.TempDir(), "c.json")}

	okRes := &CFStreamResult{
		StreamURL:  "https://cdn.test/master.m3u8",
		PlayerHost: "https://player.test",
		VideoHash:  "hash123",
	}
	solver := &retrySolver{
		errs:    []error{fmt.Errorf("turnstile timeout"), nil},
		results: []*CFStreamResult{nil, okRes},
	}

	c := NewSuperFlixClient()
	c.browserSolver = solver

	res, err := c.GetStreamURL(context.Background(), "filme", "1", "", "")
	require.NoError(t, err, "the retry must rescue the transient first-attempt failure")
	assert.Equal(t, okRes.StreamURL, res.StreamURL)
	assert.Equal(t, 2, solver.calls, "GetStreamURL must have retried the solve exactly once")

	// The freshly sniffed (host, hash) must be cached for browser-free replays.
	got, ok := defaultStreamCache.get(streamCacheKey("filme", "1", "", ""))
	require.True(t, ok, "a successful re-solve must populate the stream cache")
	assert.Equal(t, "https://player.test", got.Host)
	assert.Equal(t, "hash123", got.Hash)
}
