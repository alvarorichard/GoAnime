package http3

import (
	"context"
	"net"

	"github.com/enetx/quic-go"
)

// PreconnectOptions controls connection readiness without sending a request.
type PreconnectOptions struct {
	// Early returns at 0-RTT encryption readiness when available. False waits
	// for the authenticated TLS handshake, independently of early setup frames.
	Early bool
	// OnConnectionStarted runs once in this caller after the actual pooled
	// connection starts, including when joining a dial that has already started.
	// It must return promptly. The default dialer reports the event after DNS;
	// custom Dial implementations must call NotifyConnectionStarted themselves.
	OnConnectionStarted func()
}

// PreconnectWithOptions prepares the normal pooled connection with an optional
// start notification. It preserves Preconnect's shared-dial and cancellation
// ownership. It never reads or sends an application request.
func (t *Transport) PreconnectWithOptions(
	ctx context.Context,
	address string,
	options PreconnectOptions,
) (*quic.Conn, error) {
	return t.preconnect(ctx, address, options)
}

// Preconnect prepares a pooled HTTP/3 connection to address (host:port),
// waiting for the verified TLS handshake without sending an HTTP request.
// Connection SETTINGS and stream setup retain the transport's normal behavior.
// Cancellation releases this caller; a shared connection used by another caller
// is retained. Otherwise a failed or abandoned attempt is removed and closed.
func (t *Transport) Preconnect(ctx context.Context, address string) error {
	_, err := t.preconnect(ctx, address, PreconnectOptions{})
	return err
}

// PreconnectEarly prepares a pooled connection without sending an HTTP request.
// It returns when cached 0-RTT keys are available, or after the full handshake
// when early data is unavailable. The returned connection remains owned by the
// transport. Early readiness does not authenticate the current handshake.
// A custom Dial callback must itself support returning an early connection.
func (t *Transport) PreconnectEarly(ctx context.Context, address string) (*quic.Conn, error) {
	return t.preconnect(ctx, address, PreconnectOptions{Early: true})
}

func (t *Transport) preconnect(ctx context.Context, address string, options PreconnectOptions) (*quic.Conn, error) {
	if _, _, err := net.SplitHostPort(address); err != nil {
		return nil, err
	}
	if err := ctx.Err(); err != nil {
		return nil, context.Cause(ctx)
	}
	t.initOnce.Do(func() { t.initErr = t.init() })
	if t.initErr != nil {
		return nil, t.initErr
	}
	client, _, err := t.getClient(ctx, address, false)
	if err != nil {
		return nil, err
	}
	var failed bool
	defer func() {
		t.mutex.Lock()
		closeAttempt := client.useCount.Add(-1) == 0 && failed && client.authority == address && t.clients[client.poolKey] == client
		if closeAttempt {
			delete(t.clients, client.poolKey)
		}
		t.mutex.Unlock()
		if closeAttempt {
			client.Close()
		}
	}()
	started := client.connectStarted
	if options.OnConnectionStarted == nil {
		started = nil
	}
	for {
		select {
		case <-started:
			if ctx.Err() != nil {
				failed = true
				return nil, context.Cause(ctx)
			}
			options.OnConnectionStarted()
			started = nil
		case <-client.dialing:
			// DialEarly can publish readiness concurrently with its start event.
			// Deliver an already reported event before returning that readiness.
			if started != nil && ctx.Err() == nil {
				select {
				case <-started:
					options.OnConnectionStarted()
				default:
				}
			}
			goto ready
		case <-ctx.Done():
			failed = true
			return nil, context.Cause(ctx)
		}
	}

ready:
	if client.dialErr != nil {
		failed = true
		return nil, client.dialErr
	}
	if !options.Early {
		select {
		case <-client.conn.HandshakeComplete():
		case <-client.conn.Context().Done():
			failed = true
			return nil, context.Cause(client.conn.Context())
		case <-ctx.Done():
			failed = true
			return nil, context.Cause(ctx)
		}
	}
	if client.conn.Context().Err() != nil {
		failed = true
		return nil, context.Cause(client.conn.Context())
	}
	if ctx.Err() != nil {
		failed = true
		return nil, context.Cause(ctx)
	}
	return client.conn, nil
}
