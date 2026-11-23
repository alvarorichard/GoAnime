package http

import (
	"context"
	"time"
)

// HTTP2Connection is a maintenance handle for a connection owned by a Transport.
// Requests continue to use the transport's normal pool. Handles remain valid
// after removal from the pool; operations on a closed connection are harmless.
type HTTP2Connection struct {
	close  func() error
	retire func(HTTP2RetirementOptions)
	ping   func(context.Context) error
	idle   func() time.Duration
}

// HTTP2RetirementOptions controls optional wire traffic when draining a session.
// The zero value retires quietly, preserving the native pool behavior.
type HTTP2RetirementOptions struct {
	// GoAway sends a NO_ERROR GOAWAY with last server stream ID zero after the
	// final active/reserved stream ends. It never precedes an active response.
	GoAway bool
	// ProbeContext sends a PING before closing an idle session. Its deadline
	// closes an unresponsive session; cancellation only stops waiting.
	ProbeContext context.Context
	// ProbePayload is copied. Nil selects native randomness. Equal outstanding
	// payloads share a probe. It is unused when ProbeContext is nil.
	ProbePayload *[8]byte
}

// Close interrupts active requests and closes the underlying connection.
func (c *HTTP2Connection) Close() error { return c.close() }

// Retire prevents new requests and closes the connection after active requests
// finish. Optional settings select a probe and/or GOAWAY on drain. With no
// settings it sends neither. Active response bodies remain readable.
func (c *HTTP2Connection) Retire(options ...HTTP2RetirementOptions) {
	var selected HTTP2RetirementOptions
	if len(options) > 0 {
		selected = options[0]
	}
	if selected.ProbePayload != nil {
		payload := *selected.ProbePayload
		selected.ProbePayload = &payload
	}
	c.retire(selected)
}

// Ping sends a PING and waits for its acknowledgement or context cancellation.
func (c *HTTP2Connection) Ping(ctx context.Context) error { return c.ping(ctx) }

// IdleTime is the duration since the last successful socket read or write.
// Buffered application reads do not count as new network activity.
func (c *HTTP2Connection) IdleTime() time.Duration { return c.idle() }

// HTTP2Connections snapshots the native pool across all routes without dialing,
// reserving streams or sending frames. Each underlying connection appears once.
// Connections established after the snapshot are not included. Custom protocol
// implementations and application-supplied connection pools are not inspected.
func (t *Transport) HTTP2Connections() []*HTTP2Connection {
	t.nextProtoOnce.Do(t.onceSetNextProtoDefaults)
	return t.http2Connections()
}
