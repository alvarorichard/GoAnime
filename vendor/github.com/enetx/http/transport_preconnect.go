package http

import (
	"context"
	"errors"

	"github.com/enetx/http/httptrace"
)

// Preconnect prepares a pooled connection for req without sending the request
// or reading or closing its body. Proxy handshakes and HTTP/2 connection setup
// are allowed. The subsequent RoundTrip uses the normal pool and retry rules;
// Preconnect does not reserve a connection for that request.
// Keep-alives must be enabled. Cancellation follows Transport's shared-dial
// behavior: a detached dial may finish and be cached for another request.
func (t *Transport) Preconnect(req *Request) error {
	if req == nil || req.URL == nil || req.URL.Host == "" {
		return errors.New("http: Preconnect requires a request URL and host")
	}
	if req.URL.Scheme != "http" && req.URL.Scheme != "https" {
		return errors.New("http: unsupported Preconnect scheme")
	}
	if t.DisableKeepAlives || t.MaxIdleConnsPerHost < 0 {
		return errors.ErrUnsupported
	}
	t.nextProtoOnce.Do(t.onceSetNextProtoDefaults)
	ctx, cancel := context.WithCancelCause(req.Context())
	defer cancel(nil)
	treq := &transportRequest{Request: req, ctx: ctx, cancel: cancel, trace: httptrace.ContextClientTrace(ctx)}
	cm, err := t.connectMethodForRequest(treq)
	if err != nil {
		return err
	}
	conn, err := t.getConn(treq, cm)
	if err != nil {
		return err
	}
	// HTTP/2 is already registered in its pool. Returning the persistConn also
	// makes it available to native connection waiters; HTTP/1 re-enters idle.
	t.putOrCloseIdleConn(conn)
	return nil
}
