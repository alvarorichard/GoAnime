package http3

import (
	"context"
	"errors"
	"sync"
	"sync/atomic"

	"github.com/enetx/http"
	"github.com/enetx/quic-go"
)

// earlyRestartClient owns successive HTTP/3 generations of one QUIC connection.
// Rejection retires the old workers before QUIC invalidates their streams. The
// verified application-ready hook creates the replacement exactly once.
type earlyRestartClient struct {
	conn     *quic.Conn
	current  atomic.Pointer[ClientConn]
	factory  func() *ClientConn
	rejected atomic.Bool
	staged   atomic.Bool
	ready    chan struct{}
	mu       sync.Mutex // serializes pool retirement with generation replacement
	retired  bool
}

func newEarlyRestartClient(client *ClientConn, factory func() *ClientConn) *earlyRestartClient {
	c := &earlyRestartClient{conn: client.conn, factory: factory, ready: make(chan struct{})}
	c.current.Store(client)
	return c
}

func (c *earlyRestartClient) reject() {
	// This hook is called synchronously before ResetFor0RTT. Canceling the
	// generation first prevents late workers from opening streams in new maps.
	c.rejected.Store(true)
	c.current.Load().rawConn.retireEarlyData()
}

func (c *earlyRestartClient) applicationReady() {
	if c.rejected.Load() && c.staged.CompareAndSwap(false, true) {
		c.mu.Lock()
		client := c.factory()
		if c.retired {
			client.retire()
		}
		c.current.Store(client)
		close(c.ready)
		c.mu.Unlock()
	}
}

func (c *earlyRestartClient) retire() {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.retired = true
	c.current.Load().retire()
}

func (c *earlyRestartClient) active(ctx context.Context) (*ClientConn, error) {
	if !c.rejected.Load() {
		return c.current.Load(), nil
	}
	select {
	case <-ctx.Done():
		return nil, context.Cause(ctx)
	case <-c.conn.Context().Done():
		return nil, context.Cause(c.conn.Context())
	case <-c.ready:
		return c.current.Load(), nil
	}
}

func (c *earlyRestartClient) RoundTrip(req *http.Request) (*http.Response, error) {
	var rejection error
	for {
		client, err := c.active(req.Context())
		if err != nil {
			return nil, err
		}

		request, reopened := req, false
		if rejection != nil {
			// A callback returning Err0RTTRejected cannot replace a generation.
			// Wait for the real replacement before reopening any request body.
			if !c.rejected.Load() {
				return nil, rejection
			}

			request, reopened, err = requestAfterEarlyRejection(req, rejection)
			if err != nil {
				return nil, err
			}
		}

		response, err := client.RoundTrip(request)
		if rejection != nil {
			if err != nil && reopened {
				request.Body.Close()
			}

			return response, err // Rejection is one-shot for this connection.
		}

		_, local := errors.AsType[*RequestPreparationError](err)
		if local || !errors.Is(err, quic.Err0RTTRejected) || req.Context().Err() != nil {
			return response, err
		}

		rejection = err
	}
}

// Request replay and body ownership are independent of generation readiness.
// An unopened attempt retains its original body; only an attempted body needs
// GetBody. The caller owns a reopened body until RoundTrip accepts or closes it.
func requestAfterEarlyRejection(req *http.Request, rejected error) (*http.Request, bool, error) {
	retry := *req
	switch retry.Method {
	case MethodGet0RTT:
		retry.Method = http.MethodGet
	case MethodHead0RTT:
		retry.Method = http.MethodHead
	}

	if retry.Body == nil || retry.Body == http.NoBody {
		return &retry, false, nil
	}

	if _, unopened := errors.AsType[*errConnUnusable](rejected); unopened {
		return &retry, false, nil
	}

	if retry.GetBody == nil {
		return nil, false, rejected
	}

	body, err := retry.GetBody()
	if err != nil {
		return nil, false, err
	}

	retry.Body = &requestBodyCloser{ReadCloser: body}
	return &retry, true, nil
}

func (c *earlyRestartClient) OpenRequestStream(ctx context.Context) (*RequestStream, error) {
	client, err := c.active(ctx)
	if err != nil {
		return nil, err
	}
	return client.OpenRequestStream(ctx)
}

func (c *earlyRestartClient) handleUnidirectionalStream(stream *quic.ReceiveStream) {
	client, err := c.active(c.conn.Context())
	if err == nil {
		client.handleUnidirectionalStream(stream)
	}
}
