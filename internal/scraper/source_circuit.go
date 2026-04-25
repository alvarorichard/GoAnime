// Package scraper implements provider search, stream extraction, and source diagnostics.
package scraper

import (
	"fmt"
	"sync"
	"time"
)

const (
	defaultSourceFailureThreshold = 3
	defaultSourceCooldown         = 10 * time.Minute
)

type sourceCircuitState struct {
	failures       int
	openUntil      time.Time
	lastDiagnostic *SourceDiagnostic
}

type sourceCircuitBreaker struct {
	mu        sync.Mutex
	threshold int
	cooldown  time.Duration
	now       func() time.Time
	states    map[ScraperType]*sourceCircuitState
}

func newSourceCircuitBreaker() *sourceCircuitBreaker {
	return &sourceCircuitBreaker{
		threshold: defaultSourceFailureThreshold,
		cooldown:  defaultSourceCooldown,
		now:       time.Now,
		states:    make(map[ScraperType]*sourceCircuitState),
	}
}

func (cb *sourceCircuitBreaker) isOpen(source ScraperType) (time.Time, *SourceDiagnostic, bool) {
	if cb == nil {
		return time.Time{}, nil, false
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	state := cb.states[source]
	if state == nil || state.openUntil.IsZero() {
		return time.Time{}, nil, false
	}

	now := cb.now()
	if !now.Before(state.openUntil) {
		state.openUntil = time.Time{}
		state.failures = 0
		state.lastDiagnostic = nil
		return time.Time{}, nil, false
	}

	return state.openUntil, state.lastDiagnostic, true
}

func (cb *sourceCircuitBreaker) recordSuccess(source ScraperType) {
	if cb == nil {
		return
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()
	delete(cb.states, source)
}

func (cb *sourceCircuitBreaker) recordFailure(source ScraperType, diagnostic *SourceDiagnostic) bool {
	if cb == nil || diagnostic == nil || !diagnostic.ShouldOpenCircuit() {
		return false
	}

	cb.mu.Lock()
	defer cb.mu.Unlock()

	state := cb.states[source]
	if state == nil {
		state = &sourceCircuitState{}
		cb.states[source] = state
	}

	state.failures++
	state.lastDiagnostic = diagnostic
	if state.failures < cb.threshold {
		return false
	}

	state.openUntil = cb.now().Add(cb.cooldown)
	return true
}

func (sm *ScraperManager) ensureCircuitBreaker() *sourceCircuitBreaker {
	sm.breakerMu.Lock()
	defer sm.breakerMu.Unlock()

	if sm.breaker == nil {
		sm.breaker = newSourceCircuitBreaker()
	}
	return sm.breaker
}

func (sm *ScraperManager) circuitOpenDiagnostic(source ScraperType) (*SourceDiagnostic, time.Duration, bool) {
	breaker := sm.ensureCircuitBreaker()
	openUntil, lastDiagnostic, ok := breaker.isOpen(source)
	if !ok {
		return nil, 0, false
	}

	sourceName := sm.getScraperDisplayName(source)
	message := fmt.Sprintf("circuit breaker open until %s", openUntil.Format(time.RFC3339))
	if lastDiagnostic != nil {
		message = fmt.Sprintf("%s; last failure: %s", message, lastDiagnostic.UserMessage())
	}

	diagnostic := &SourceDiagnostic{
		Source:  sourceName,
		Layer:   "circuit-breaker",
		Kind:    DiagnosticSourceUnavailable,
		Message: message,
		Err:     ErrSourceUnavailable,
	}

	return diagnostic, time.Until(openUntil), true
}

func (sm *ScraperManager) recordSourceSuccess(source ScraperType) {
	sm.ensureCircuitBreaker().recordSuccess(source)
}

func (sm *ScraperManager) recordSourceFailure(source ScraperType, diagnostic *SourceDiagnostic) bool {
	return sm.ensureCircuitBreaker().recordFailure(source, diagnostic)
}
