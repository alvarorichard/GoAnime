// Package scraper — SSRF protection for scraper HTTP clients.
//
// Duplicates the core SSRF dial-check from internal/api to avoid an import
// cycle (api → scraper → api).
package scraper

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
)

// isDisallowedIP returns true if the IP is loopback, private, multicast, or
// unspecified — mirrors api.IsDisallowedIP.
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

// safeScraperTransport returns an *http.Transport with SSRF-safe dial hooks.
func safeScraperTransport(timeout time.Duration) *http.Transport {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	return &http.Transport{
		DialContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			return safeDialFunc(network, addr, timeout, nil)
		},
		DialTLSContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			return safeDialFunc(network, addr, timeout, tlsConfig)
		},
		TLSHandshakeTimeout: timeout,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 15,
		IdleConnTimeout:     90 * time.Second,
	}
}
