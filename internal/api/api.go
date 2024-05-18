package api

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"github.com/pkg/errors"
	"golang.org/x/net/context"
)

// IsDisallowedIP checks if the given IP is disallowed.
func IsDisallowedIP(hostIP string) bool {
	ip := net.ParseIP(hostIP)
	return ip.IsMulticast() || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate()
}

// checkDisallowedIP validates the IP of a connection.
func checkDisallowedIP(conn net.Conn) error {
	ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())
	if IsDisallowedIP(ip) {
		conn.Close()
		return errors.New("ip address is not allowed")
	}
	return nil
}

// dialFunc handles both regular and TLS connections.
func dialFunc(network, addr string, timeout time.Duration, tlsConfig *tls.Config) (net.Conn, error) {
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
	if err := checkDisallowedIP(conn); err != nil {
		return nil, err
	}
	return conn, nil
}

// SafeTransport returns a http.Transport with custom dial functions.
func SafeTransport(timeout time.Duration) *http.Transport {
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialFunc(network, addr, timeout, nil)
		},
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialFunc(network, addr, timeout, tlsConfig)
		},
		TLSHandshakeTimeout: timeout,
	}
}

// SafeGet performs a GET request with a timeout and returns the response.
func SafeGet(url string) (*http.Response, error) {
	httpClient := &http.Client{
		Transport: SafeTransport(10 * time.Second),
	}
	return httpClient.Get(url)
}
