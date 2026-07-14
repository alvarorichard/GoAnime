package netx

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func circuitWorthy() *SourceDiagnostic {
	// A 521 (Cloudflare origin down) is circuit-worthy.
	return DiagnoseError("Test", "search", NewHTTPStatusError("Test", "search", 521))
}

func TestCircuitBreaker_OpensAfterThreshold(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker()
	cb.threshold = 3

	assert.False(t, cb.RecordFailure("k", circuitWorthy()), "1st failure must not open")
	assert.False(t, cb.RecordFailure("k", circuitWorthy()), "2nd failure must not open")
	assert.True(t, cb.RecordFailure("k", circuitWorthy()), "3rd failure must open")

	_, _, open := cb.IsOpen("k")
	assert.True(t, open, "breaker must report open after threshold")
}

func TestCircuitBreaker_SuccessResets(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker()
	cb.threshold = 2

	cb.RecordFailure("k", circuitWorthy())
	cb.RecordSuccess("k")
	assert.False(t, cb.RecordFailure("k", circuitWorthy()), "success must reset the failure count")
	_, _, open := cb.IsOpen("k")
	assert.False(t, open)
}

func TestCircuitBreaker_ReopensAfterCooldown(t *testing.T) {
	t.Parallel()
	now := time.Now()
	cb := NewCircuitBreaker()
	cb.threshold = 1
	cb.cooldown = time.Minute
	cb.now = func() time.Time { return now }

	require.True(t, cb.RecordFailure("k", circuitWorthy()))
	_, _, open := cb.IsOpen("k")
	require.True(t, open)

	// Advance past the cooldown — breaker auto-closes.
	now = now.Add(2 * time.Minute)
	_, _, open = cb.IsOpen("k")
	assert.False(t, open, "breaker must reopen (close) after the cooldown window")
}

func TestCircuitBreaker_NonCircuitWorthyIgnored(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker()
	cb.threshold = 1
	// A parser error is not circuit-worthy (the site is up, our parser is stale).
	parser := DiagnoseError("Test", "episode", NewParserError("Test", "episode", "no match", nil))
	assert.False(t, cb.RecordFailure("k", parser))
	_, _, open := cb.IsOpen("k")
	assert.False(t, open)
}

func TestCircuitBreaker_KeysIndependent(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker()
	cb.threshold = 1
	require.True(t, cb.RecordFailure("a", circuitWorthy()))
	_, _, openA := cb.IsOpen("a")
	_, _, openB := cb.IsOpen("b")
	assert.True(t, openA)
	assert.False(t, openB, "one key opening must not affect another")
}

func TestCircuitBreaker_OpenDiagnostic(t *testing.T) {
	t.Parallel()
	cb := NewCircuitBreaker()
	cb.threshold = 1
	require.True(t, cb.RecordFailure("k", circuitWorthy()))

	diag, retry, ok := cb.OpenDiagnostic("k", "AllAnime")
	require.True(t, ok)
	require.NotNil(t, diag)
	assert.Equal(t, "AllAnime", diag.Source)
	assert.Equal(t, DiagnosticSourceUnavailable, diag.Kind)
	assert.Positive(t, retry)
}

func TestCircuitBreaker_NilSafe(t *testing.T) {
	t.Parallel()
	var cb *CircuitBreaker
	_, _, open := cb.IsOpen("k")
	assert.False(t, open)
	assert.False(t, cb.RecordFailure("k", circuitWorthy()))
	cb.RecordSuccess("k") // must not panic
}
