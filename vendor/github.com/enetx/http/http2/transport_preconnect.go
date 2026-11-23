package http2

import (
	"context"
	"errors"

	"github.com/enetx/http"
	"github.com/enetx/http/httptrace"
)

// Preconnect prepares an HTTP/2 connection in the transport's normal pool
// without sending req or touching its body. Custom ConnPool implementations
// are unsupported. Dial coalescing, TLS options and connection settings are
// identical to RoundTrip.
func (t *Transport) Preconnect(req *http.Request) error {
	if t.ConnPool != nil {
		return errors.ErrUnsupported
	}
	if req == nil || req.URL == nil || req.URL.Scheme == "http" && !t.AllowHTTP {
		return errors.New("http2: invalid Preconnect request")
	}
	dial := &transportRoundTripState{}
	defer dial.roundTripDone()
	ctx := context.WithValue(req.Context(), http2TransportContextKey{}, dial)
	ctx = httptrace.WithClientTrace(ctx, &httptrace.ClientTrace{
		GotConn: func(httptrace.GotConnInfo) { dial.gotConn() },
	})
	return t.init().Preconnect(req.WithContext(ctx))
}
