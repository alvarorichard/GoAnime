// Package scraper — SSRF protection for scraper HTTP clients.
//
// Duplicates the core SSRF dial-check from internal/api to avoid an import
// cycle (api → scraper → api).
package netx

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// IsDisallowedIP returns true if the IP is loopback, private, multicast, or
// unspecified — mirrors api.IsDisallowedIP.
func IsDisallowedIP(hostIP string) bool {
	ip := net.ParseIP(hostIP)
	if ip == nil {
		return true
	}
	return ip.IsMulticast() || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate()
}

// SafeDialFunc establishes a connection and rejects disallowed IPs.
func SafeDialFunc(network, addr string, timeout time.Duration, tlsConfig *tls.Config) (net.Conn, error) {
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
	if IsDisallowedIP(ip) {
		_ = conn.Close()
		return nil, errors.New("ip address is not allowed")
	}
	return conn, nil
}

// SafeScraperTransport returns an *http.Transport with SSRF-safe dial hooks.
func SafeScraperTransport(timeout time.Duration) *http.Transport {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	return &http.Transport{
		DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			return SafeDialFunc(network, addr, timeout, nil)
		},
		DialTLSContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			return SafeDialFunc(network, addr, timeout, tlsConfig)
		},
		TLSHandshakeTimeout: timeout,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 15,
		IdleConnTimeout:     90 * time.Second,
	}
}
