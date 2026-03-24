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
func safeMovieTransport(timeout time.Duration) *http.Transport {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	return &http.Transport{
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
