package h3transport

import (
	"context"
	"crypto/tls"
	"errors"
	"sync/atomic"

	"github.com/enetx/http"
	"github.com/enetx/http/httptrace"
	"github.com/enetx/quic-go/http3"
	"github.com/enetx/surf/internal/connrace"
)

type Preconnector interface{ Preconnect(*http.Request) error }

func raceAddress(req *http.Request) (string, error) {
	return connrace.Address(req)
}

func (ut *Transport) roundTripRacing(req *http.Request) (*http.Response, error) {
	address, err := raceAddress(req)
	if err != nil {
		if req.Body != nil {
			req.Body.Close()
		}
		return nil, err
	}

	eligible := ut.eligibleEarlyData(req)

	// An already usable exact QUIC session removes the browser's main TCP job.
	// Check the actual partition and GOAWAY state before racing a ready H2 pool.
	// https://github.com/chromium/chromium/blob/158.0.8127.18/net/http/http_stream_factory_job_controller.cc
	ctx := req.Context()
	if ut.coalescer != nil {
		ctx = ut.coalescer.prepareContext(ctx, address)
	}

	if ut.http3tr.HasCachedConnection(ctx, address) {
		response, cachedErr := ut.roundTripHTTP3Attempt(req, true)
		if cachedErr != http3.ErrNoCachedConn {
			return response, cachedErr
		}
	}

	if eligible && ut.earlyDataBackup {
		return ut.roundTripEarlyBackup(req, address)
	}

	selection, err := ut.racer.SelectProtocol(req, address, eligible)
	if err != nil {
		if req.Body != nil {
			req.Body.Close()
		}
		return nil, err
	}

	// After selection, only the selected normal transport owns the body. Never
	// retry a failed application exchange against the other protocol here.
	if selection.Protocol == 3 {
		return ut.roundTripHTTP3(req)
	}

	return ut.Fallback.RoundTrip(req)
}

func (ut *Transport) preconnectQUIC(ctx context.Context, address string, early bool) error {
	if ut.coalescer != nil {
		ctx = ut.coalescer.prepareContext(ctx, address)
	}

	_, err := ut.http3tr.PreconnectWithOptions(ctx, address, http3.PreconnectOptions{
		Early:               early,
		OnConnectionStarted: func() { connrace.QUICStarted(ctx) },
	})

	return err
}

func (ut *Transport) preconnectTCP(req *http.Request) error {
	// JA wraps the fallback after the HTTP/3 factory runs. Resolve the current
	// transport at use time, so preparation and sending always use the same TLS.
	preconnect, ok := ut.Fallback.(Preconnector)
	if !ok {
		return errors.New("surf: TCP transport does not support Preconnect")
	}

	return preconnect.Preconnect(req)
}

// observeTCPPreparation uses the normal pool's trace instead of a mutable
// context flag. Native HTTP/1 Preconnect reports GotConn for fresh and reused
// sockets; native HTTP/2 preparation leaves its session in the shared H2 pool.
// The result is history metadata, not a request to switch transport pools.
func observeTCPPreparation(req *http.Request, prepare func(*http.Request) error) (bool, error) {
	var http1 atomic.Bool
	trace := &httptrace.ClientTrace{GotConn: func(info httptrace.GotConnInfo) {
		if conn, ok := info.Conn.(interface{ ConnectionState() tls.ConnectionState }); ok {
			http1.Store(conn.ConnectionState().NegotiatedProtocol != "h2")
		} else if req.URL.Scheme == "http" {
			http1.Store(true)
		}
	}}
	err := prepare(req.WithContext(httptrace.WithClientTrace(req.Context(), trace)))
	return http1.Load(), err
}
