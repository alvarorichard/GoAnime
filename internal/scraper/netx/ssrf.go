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
	"syscall"
	"time"

	"github.com/pkg/errors"
)

// IsDisallowedIP returns true if the IP is loopback, private, link-local,
// multicast, or unspecified — api.IsDisallowedIP's check plus link-local,
// which covers the 169.254.169.254 cloud metadata endpoint and fe80::/10.
func IsDisallowedIP(hostIP string) bool {
	ip := net.ParseIP(hostIP)
	if ip == nil {
		return true
	}
	return ip.IsMulticast() || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate() || ip.IsLinkLocalUnicast()
}

// SafeDialFunc establishes a connection and rejects disallowed IPs.
func SafeDialFunc(network, addr string, timeout time.Duration, tlsConfig *tls.Config) (net.Conn, error) {
	return SafeDialContext(context.Background(), network, addr, timeout, tlsConfig)
}

// SafeDialContext establishes a connection unless addr resolves to a
// disallowed IP. The check runs in the dialer's Control hook: after DNS, on
// the exact address about to be connected, before connect(2). A forbidden
// host therefore never sees a SYN, let alone a TLS ClientHello, and a name
// that re-resolves to an internal address (DNS rebinding) is judged by the
// address actually used. ctx cancels both the dial and the TLS handshake.
func SafeDialContext(ctx context.Context, network, addr string, timeout time.Duration, tlsConfig *tls.Config) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout, ControlContext: rejectDisallowedIP}
	if tlsConfig != nil {
		return (&tls.Dialer{NetDialer: dialer, Config: tlsConfig}).DialContext(ctx, network, addr)
	}
	return dialer.DialContext(ctx, network, addr)
}

func rejectDisallowedIP(_ context.Context, _, address string, _ syscall.RawConn) error {
	ip, _, err := net.SplitHostPort(address)
	if err != nil {
		return errors.New("failed to parse remote address")
	}
	if IsDisallowedIP(ip) {
		return errors.New("ip address is not allowed")
	}
	return nil
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
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return SafeDialContext(ctx, network, addr, timeout, nil)
		},
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return SafeDialContext(ctx, network, addr, timeout, tlsConfig)
		},
		TLSHandshakeTimeout: timeout,
		MaxIdleConns:        100,
		MaxIdleConnsPerHost: 15,
		IdleConnTimeout:     90 * time.Second,
	}
}
