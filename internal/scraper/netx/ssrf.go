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
// SafeTLSConfig returns the TLS settings used for every SSRF-guarded dial.
//
// NextProtos is what enables HTTP/2. A transport that sets DialTLSContext takes
// over the TLS handshake, so net/http can no longer inject the ALPN protocol
// list for us: without this, every connection negotiates http/1.1 and the burst
// of requests these clients make is serialised over separate connections
// instead of multiplexed on one. It pairs with ForceAttemptHTTP2 on the
// transport — both are required, neither is enough alone.
func SafeTLSConfig() *tls.Config {
	return &tls.Config{
		MinVersion: tls.VersionTLS12,
		NextProtos: []string{"h2", "http/1.1"},
	}
}

func SafeScraperTransport(timeout time.Duration) *http.Transport {
	tlsConfig := SafeTLSConfig()
	return &http.Transport{
		// Required alongside NextProtos: net/http only upgrades a transport with
		// custom dial hooks to HTTP/2 when this is set.
		ForceAttemptHTTP2: true,
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
