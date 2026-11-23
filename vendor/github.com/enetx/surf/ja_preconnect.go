package surf

import (
	"github.com/enetx/http"
	h3transport "github.com/enetx/surf/internal/http3"
)

// HasConnectionGroup observes the same native TCP pool and proxy route as
// RoundTrip. It includes occupied H1 connections and pending connection dials;
// it does not create a proxy child, dial, reserve a connection or touch a body.
func (rt *roundtripper) HasConnectionGroup(req *http.Request) bool {
	if req == nil || req.URL == nil || req.URL.Scheme != "https" {
		return false
	}

	req, err := rt.prepareProxy(req)
	if err != nil {
		return false
	}

	proxy, _ := h3transport.SelectedProxy(req)
	if rt.requestProxy != nil && proxy != nil {
		rt.proxyMu.Lock()
		child := rt.proxyTransports[proxy.String()]
		rt.proxyMu.Unlock()

		return child != nil && child.HasConnectionGroup(req)
	}

	if rt.requiresHTTP1(req) {
		return rt.http1trFallback.HasConnectionGroup(req)
	}

	return rt.http1tr.HasConnectionGroup(req)
}

// HasHTTP2Connection observes the same native pool and proxy route as RoundTrip.
// It does not create a proxy child, dial or reserve a stream.
func (rt *roundtripper) HasHTTP2Connection(req *http.Request) bool {
	if req == nil || req.URL == nil || req.URL.Scheme != "https" {
		return false
	}

	req, err := rt.prepareProxy(req)
	if err != nil || rt.requiresHTTP1(req) {
		return false
	}

	proxy, _ := h3transport.SelectedProxy(req)
	if rt.requestProxy != nil && proxy != nil {
		rt.proxyMu.Lock()
		child := rt.proxyTransports[proxy.String()]
		rt.proxyMu.Unlock()

		return child != nil && child.HasHTTP2Connection(req)
	}

	return rt.http2tr != nil && rt.http2tr.HasHTTP2Connection(req)
}

// Preconnect warms the same JA transports used by RoundTrip. No alternate TLS
// implementation, request body reader or separate connection pool is introduced.
func (rt *roundtripper) Preconnect(req *http.Request) error {
	var err error
	req, err = rt.prepareProxy(req)
	if err != nil {
		return err
	}

	child, err := rt.proxyTransport(req)
	if err != nil {
		return err
	}

	if child != nil {
		return child.Preconnect(req)
	}

	if rt.requiresHTTP1(req) {
		return rt.http1trFallback.Preconnect(req)
	}

	if rt.http2tr != nil && req.URL.Scheme == "https" {
		return rt.http2tr.Preconnect(req)
	}

	return rt.http1tr.Preconnect(req)
}
