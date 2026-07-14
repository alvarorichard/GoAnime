package netx

import (
	"fmt"
	"sync"
	"time"
)

// Per-source circuit breaker (R5). A source that fails repeatedly with a
// circuit-worthy diagnostic is skipped for a cooldown window instead of being
// retried into the ground. Keyed by an opaque string so both the legacy
// ScraperManager (by display name) and the Model B registry (by SourceKind)
// can share the mechanism without an import cycle.

const (
	defaultFailureThreshold = 3
	defaultCooldown         = 10 * time.Minute
)

type circuitState struct {
	failures       int
	openUntil      time.Time
	lastDiagnostic *SourceDiagnostic
}

// CircuitBreaker tracks per-key failure state and opens after a threshold of
// circuit-worthy failures. Safe for concurrent use.
type CircuitBreaker struct {
	mu        sync.Mutex
	threshold int
	cooldown  time.Duration
	now       func() time.Time
	states    map[string]*circuitState
}

// NewCircuitBreaker returns a breaker with the default threshold and cooldown.
func NewCircuitBreaker() *CircuitBreaker {
	return &CircuitBreaker{
		threshold: defaultFailureThreshold,
		cooldown:  defaultCooldown,
		now:       time.Now,
		states:    make(map[string]*circuitState),
	}
}

// IsOpen reports whether key's breaker is currently open (and the last
// diagnostic + when it reopens). A lapsed window auto-resets.
func (cb *CircuitBreaker) IsOpen(key string) (time.Time, *SourceDiagnostic, bool) {
	if cb == nil {
		return time.Time{}, nil, false
	}
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state := cb.states[key]
	if state == nil || state.openUntil.IsZero() {
		return time.Time{}, nil, false
	}
	if !cb.now().Before(state.openUntil) {
		state.openUntil = time.Time{}
		state.failures = 0
		state.lastDiagnostic = nil
		return time.Time{}, nil, false
	}
	return state.openUntil, state.lastDiagnostic, true
}

// RecordSuccess clears any accumulated failure state for key.
func (cb *CircuitBreaker) RecordSuccess(key string) {
	if cb == nil {
		return
	}
	cb.mu.Lock()
	defer cb.mu.Unlock()
	delete(cb.states, key)
}

// RecordFailure records a failure for key. It returns true when this failure
// opened the breaker. Only circuit-worthy diagnostics count.
func (cb *CircuitBreaker) RecordFailure(key string, diagnostic *SourceDiagnostic) bool {
	if cb == nil || diagnostic == nil || !diagnostic.ShouldOpenCircuit() {
		return false
	}
	cb.mu.Lock()
	defer cb.mu.Unlock()

	state := cb.states[key]
	if state == nil {
		state = &circuitState{}
		cb.states[key] = state
	}
	state.failures++
	state.lastDiagnostic = diagnostic
	if state.failures < cb.threshold {
		return false
	}
	state.openUntil = cb.now().Add(cb.cooldown)
	return true
}

// OpenDiagnostic returns a typed SourceDiagnostic describing an open breaker for
// key (labeled with sourceName) plus the remaining cooldown, or ok=false when
// the breaker is closed.
func (cb *CircuitBreaker) OpenDiagnostic(key, sourceName string) (*SourceDiagnostic, time.Duration, bool) {
	openUntil, last, ok := cb.IsOpen(key)
	if !ok {
		return nil, 0, false
	}
	message := fmt.Sprintf("circuit breaker open until %s", openUntil.Format(time.RFC3339))
	if last != nil {
		message = fmt.Sprintf("%s; last failure: %s", message, last.UserMessage())
	}
	return &SourceDiagnostic{
		Source:  sourceName,
		Layer:   "circuit-breaker",
		Kind:    DiagnosticSourceUnavailable,
		Message: message,
		Err:     ErrSourceUnavailable,
	}, time.Until(openUntil), true
}
