package http3

import (
	"context"
	"errors"
	"fmt"

	"github.com/enetx/http"
	httpcommon "github.com/enetx/http/headerorder"
	"golang.org/x/net/http/httpguts"
)

type earlyRequestContextKey struct{}

// IsEarlyRequest reports whether this request's header preparation skipped
// authenticated connection initialization through the early request path.
// It is intended for Transport.PrepareHeaders and the resulting Response.Request.
// The decision is frozen before preparation, regardless of concurrent handshake
// completion; it does not prove that any packet used 0-RTT encryption.
// Magic early methods with an already-complete handshake initialize normally
// and report false, including warm requests after an accepted 0-RTT connection.
// Nil and requests outside this preparation path return false.
func IsEarlyRequest(req *http.Request) bool {
	if req == nil {
		return false
	}

	early, _ := req.Context().Value(earlyRequestContextKey{}).(bool)
	return early
}

func (c *ClientConn) pendingEarlyRequest(req *http.Request) bool {
	if req.Method != MethodGet0RTT && req.Method != MethodHead0RTT {
		return false
	}

	if c.prepareHeaders == nil && c.rawConn.peerSettings == nil {
		return true // Preserve the existing fast path when neither hook is configured.
	}

	select {
	case <-c.conn.HandshakeComplete():
		return false
	default:
		return true
	}
}

// Native failures before request admission belong to the existing unopened
// retry owner. Local prelude failures are terminal even if their cause resembles
// early rejection or a connection failure.
func requestInitializationError(err error) error {
	_, local := errors.AsType[*RequestPreparationError](err)
	if local {
		return err
	}

	return &errConnUnusable{e: err}
}

// A completed handshake retains native admission/retry behavior when its cached
// connection has already closed. Only cancellation before that gate aborts here.
func (c *ClientConn) handshakeCancellationError() error {
	select {
	case <-c.conn.HandshakeComplete():
		return nil
	default:
		return requestInitializationError(c.preparationFailure(context.Cause(c.rawConn.context())))
	}
}

// Startup publishes its error before closing the connection. Read that existing
// publication without waiting, preserving a local factory failure over its
// secondary connection-close error. Peer initialization similarly owns its cause.
func (c *ClientConn) preparationFailure(fallback error) error {
	select {
	case <-c.initialSettingsWritten:
		if err := c.initialSettingsErr; err != nil {
			if _, local := errors.AsType[*RequestPreparationError](err); local {
				return err
			}

			fallback = err
		}
	default:
	}

	return c.rawConn.peerSettingsFailure(fallback)
}

// The transport, direct client and an early-data retry share this close owner.
// It protects unsent failures without changing the caller's request or body.
func requestWithPreparationBody(req *http.Request) *http.Request {
	if req.Body == nil || req.Body == http.NoBody {
		return req
	}

	if _, ok := req.Body.(*requestBodyCloser); ok {
		return req
	}

	copy := *req
	copy.Body = &requestBodyCloser{ReadCloser: req.Body}
	return &copy
}

func (t *Transport) requestWithPreparationBody(req *http.Request) *http.Request {
	if t.PrepareHeaders == nil && t.InitialPeerSettings == nil {
		return req
	}

	return requestWithPreparationBody(req)
}

func (c *ClientConn) prepareRequestHeaders(req *http.Request, early bool) (*http.Request, error) {
	if c.prepareHeaders == nil {
		return req, nil
	}

	if err := c.preparationFailure(nil); err != nil {
		return nil, requestInitializationError(err)
	}

	copy := req.WithContext(context.WithValue(req.Context(), earlyRequestContextKey{}, early))
	copy.Header = req.Header.Clone()
	err := context.Cause(req.Context())
	if err == nil {
		if cause := context.Cause(c.rawConn.context()); cause != nil {
			return nil, requestInitializationError(c.preparationFailure(cause))
		}
	}

	if err == nil {
		err = c.prepareHeaders(copy, c.conn)
	}

	if err == nil {
		err = validatePreparedHeaders(copy.Header)
	}

	if err == nil {
		err = context.Cause(req.Context())
	}

	if err == nil {
		err = context.Cause(copy.Context())
	}

	if err != nil {
		if req.Body != nil {
			req.Body.Close()
		}

		return nil, &RequestPreparationError{err}
	}

	if cause := context.Cause(c.rawConn.context()); cause != nil {
		return nil, requestInitializationError(c.preparationFailure(cause))
	}

	return copy, nil
}

func validatePreparedHeaders(header http.Header) error {
	for name, values := range header {
		if httpcommon.IsMetadataKey(name) {
			continue
		}

		if !httpguts.ValidHeaderFieldName(name) {
			return fmt.Errorf("http3: invalid http header field name %q", name)
		}

		for _, value := range values {
			if !httpguts.ValidHeaderFieldValue(value) {
				return fmt.Errorf("http3: invalid http header field value %q for key %v", value, name)
			}
		}
	}

	return nil
}
