package scraper

import (
	"errors"
	"testing"
	"time"

	"github.com/alvarorichard/Goanime/internal/scraper/netx"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewSourceCircuitBreaker(t *testing.T) {
	t.Parallel()
	cb := newSourceCircuitBreaker()
	require.NotNil(t, cb)
	assert.Equal(t, defaultSourceFailureThreshold, cb.threshold)
	assert.Equal(t, defaultSourceCooldown, cb.cooldown)
	assert.NotNil(t, cb.states)
}

func TestCircuitBreaker_IsOpen_EmptyState(t *testing.T) {
	t.Parallel()
	cb := newSourceCircuitBreaker()
	_, _, open := cb.isOpen(AllAnimeType)
	assert.False(t, open)
}

func TestCircuitBreaker_RecordFailure_OpensAfterThreshold(t *testing.T) {
	t.Parallel()
	cb := newSourceCircuitBreaker()
	diag := &netx.SourceDiagnostic{Kind: netx.DiagnosticSourceUnavailable}

	for i := 0; i < cb.threshold-1; i++ {
		assert.False(t, cb.recordFailure(AllAnimeType, diag), i)
	}
	assert.True(t, cb.recordFailure(AllAnimeType, diag), "threshold trip")

	_, last, open := cb.isOpen(AllAnimeType)
	assert.True(t, open)
	assert.Equal(t, diag, last)
}

func TestCircuitBreaker_RecordFailure_IgnoresNonRetryableKinds(t *testing.T) {
	t.Parallel()
	cb := newSourceCircuitBreaker()
	diag := &netx.SourceDiagnostic{Kind: netx.DiagnosticParserBroken}
	for i := 0; i < cb.threshold+1; i++ {
		assert.False(t, cb.recordFailure(AllAnimeType, diag))
	}
	_, _, open := cb.isOpen(AllAnimeType)
	assert.False(t, open)
}

func TestCircuitBreaker_RecordFailure_NilNoOp(t *testing.T) {
	t.Parallel()
	cb := newSourceCircuitBreaker()
	assert.False(t, cb.recordFailure(AllAnimeType, nil))
	var nilCB *sourceCircuitBreaker
	assert.False(t, nilCB.recordFailure(AllAnimeType, &netx.SourceDiagnostic{Kind: netx.DiagnosticSourceUnavailable}))
}

func TestCircuitBreaker_RecordSuccess_Resets(t *testing.T) {
	t.Parallel()
	cb := newSourceCircuitBreaker()
	diag := &netx.SourceDiagnostic{Kind: netx.DiagnosticSourceUnavailable}
	for i := 0; i < cb.threshold; i++ {
		cb.recordFailure(AllAnimeType, diag)
	}
	cb.recordSuccess(AllAnimeType)
	_, _, open := cb.isOpen(AllAnimeType)
	assert.False(t, open)
}

func TestCircuitBreaker_IsOpen_ClearsAfterCooldown(t *testing.T) {
	t.Parallel()
	cb := newSourceCircuitBreaker()
	now := time.Now()
	cb.now = func() time.Time { return now }
	cb.cooldown = time.Minute

	diag := &netx.SourceDiagnostic{Kind: netx.DiagnosticSourceUnavailable}
	for i := 0; i < cb.threshold; i++ {
		cb.recordFailure(AllAnimeType, diag)
	}
	_, _, open := cb.isOpen(AllAnimeType)
	assert.True(t, open)

	cb.now = func() time.Time { return now.Add(2 * time.Minute) }
	_, _, open = cb.isOpen(AllAnimeType)
	assert.False(t, open)
}

func TestEnsureCircuitBreaker_Idempotent(t *testing.T) {
	t.Parallel()
	sm := &ScraperManager{}
	first := sm.ensureCircuitBreaker()
	second := sm.ensureCircuitBreaker()
	assert.Same(t, first, second)
}

func TestScraperManager_RecordSourceSuccessAndFailure(t *testing.T) {
	t.Parallel()
	sm := &ScraperManager{}
	diag := &netx.SourceDiagnostic{Kind: netx.DiagnosticSourceUnavailable}

	for i := 0; i < defaultSourceFailureThreshold; i++ {
		sm.recordSourceFailure(AllAnimeType, diag)
	}
	d, _, open := sm.circuitOpenDiagnostic(AllAnimeType)
	assert.True(t, open)
	require.NotNil(t, d)
	assert.Equal(t, "AllAnime", d.Source)

	sm.recordSourceSuccess(AllAnimeType)
	_, _, open = sm.circuitOpenDiagnostic(AllAnimeType)
	assert.False(t, open)
}

func TestSentinelDistinct(t *testing.T) {
	t.Parallel()
	diag := &netx.SourceDiagnostic{Kind: netx.DiagnosticDecryptBroken, Err: errors.New("x")}
	assert.False(t, errors.Is(diag, netx.ErrSourceUnavailable))
}
