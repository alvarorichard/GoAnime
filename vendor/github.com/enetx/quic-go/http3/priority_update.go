package http3

import (
	"context"
	"errors"
	"fmt"
	"sync"
)

// ErrRequestCompleted means the request can no longer be reprioritized.
var ErrRequestCompleted = errors.New("http3: request has completed")

// PriorityUpdater changes the peer's response priority for one request. It is
// concurrency safe and independent of the local QUIC upload scheduler.
// Repeated values are suppressed. A successful Update queues a control frame;
// delivery still depends on the connection and peer.
type PriorityUpdater struct {
	mu                  sync.Mutex
	request, connection context.Context
	done                <-chan struct{}
	send                func(int8, bool) error
	urgency             int8
	incremental         bool
	ready               bool
	closed              bool
}

// Update queues RFC 9218 PRIORITY_UPDATE. Urgency must be between 0 and 7.
// Updating to (3, false) resets the peer to the default priority.
func (p *PriorityUpdater) Update(urgency int8, incremental bool) error {
	if urgency < 0 || urgency > 7 {
		return fmt.Errorf("http3: priority urgency must be between 0 and 7")
	}
	if p == nil {
		return errors.New("http3: priority updater is not ready")
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if !p.ready {
		return errors.New("http3: request headers have not been sent")
	}
	if p.closed {
		return ErrRequestCompleted
	}
	select {
	case <-p.done:
		return ErrRequestCompleted
	default:
	}
	if err := p.request.Err(); err != nil {
		return err
	}
	if err := p.connection.Err(); err != nil {
		return context.Cause(p.connection)
	}
	if p.urgency == urgency && p.incremental == incremental {
		return nil
	}
	if err := p.send(urgency, incremental); err != nil {
		return err
	}
	p.urgency, p.incremental = urgency, incremental
	return nil
}

type priorityUpdateHandlerKey struct{}

// WithPriorityUpdateHandler calls handler after successful request HEADERS on
// each HTTP/3 attempt. The handler must return promptly; it may retain the
// updater and use it until the response body is closed, read to EOF, or canceled.
// It does not change the Priority header or invent resource loading scenarios.
func WithPriorityUpdateHandler(ctx context.Context, handler func(*PriorityUpdater)) context.Context {
	return context.WithValue(ctx, priorityUpdateHandlerKey{}, handler)
}

// UpdatePriority changes the peer's response priority after SendRequestHeader.
// SetPriority instead controls local QUIC upload scheduling.
func (s *RequestStream) UpdatePriority(urgency int8, incremental bool) error {
	return s.priorityUpdater.Update(urgency, incremental)
}
