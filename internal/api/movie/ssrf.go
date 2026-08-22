// Package movie — SSRF protection for movie API HTTP clients.
//
// This file duplicates the core SSRF dial-check logic from internal/api so
// that the movie package can use it without importing api (which would create
// an import cycle: api → api/movie → api).
package movie

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"context"

	"github.com/pkg/errors"
)

// isDisallowedIP returns true if ip is loopback, private, multicast, or
// unspecified — the same check as api.IsDisallowedIP.
func isDisallowedIP(hostIP string) bool {
	ip := net.ParseIP(hostIP)
	if ip == nil {
		return true
	}
	return ip.IsMulticast() || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate()
}

// safeDialFunc establishes a connection and rejects disallowed IPs.
func safeDialFunc(network, addr string, timeout time.Duration, tlsConfig *tls.Config) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	var conn net.Conn
	var err error
	if tlsConfig != nil {
		conn, err = tls.DialWithDialer(dialer, network, addr, tlsConfig)
	} else {
		conn, err = dialer.Dial(network, addr)
	}
	if err != nil {
		return nil, err
	}
	ip, _, err := net.SplitHostPort(conn.RemoteAddr().String())
	if err != nil {
		_ = conn.Close()
		return nil, errors.New("failed to parse remote address")
	}
	if isDisallowedIP(ip) {
		_ = conn.Close()
		return nil, errors.New("ip address is not allowed")
	}
	return conn, nil
}

// safeMovieTransport returns an *http.Transport with SSRF-safe dial hooks.
// safeTLSConfig returns the TLS settings used for every SSRF-guarded dial.
//
// NextProtos is what enables HTTP/2. A transport that sets DialTLSContext takes
// over the TLS handshake, so net/http can no longer inject the ALPN protocol
// list for us: without this, every connection negotiates http/1.1 and the burst
// of requests these clients make is serialised over separate connections
// instead of multiplexed on one. It pairs with ForceAttemptHTTP2 on the
// transport — both are required, neither is enough alone.
func safeTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		NextProtos: []string{"h2", "http/1.1"},
	}
}

func safeMovieTransport(timeout time.Duration) *http.Transport {
	tlsConfig := safeTLSConfig()
	return &http.Transport{
		// Required alongside NextProtos: net/http only upgrades a transport with
		// custom dial hooks to HTTP/2 when this is set.
		ForceAttemptHTTP2: true,
		DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			return safeDialFunc(network, addr, timeout, nil)
		},
		DialTLSContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			return safeDialFunc(network, addr, timeout, tlsConfig)
		},
		TLSHandshakeTimeout: timeout,
		MaxIdleConns:        50,
		MaxIdleConnsPerHost: 10,
		IdleConnTimeout:     90 * time.Second,
	}
}
