package api

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"context"
	"github.com/pkg/errors"
)

// IsDisallowedIP checks if the given IP address falls under a disallowed category.
// It returns true if the IP address is multicast, unspecified, loopback, or private.
//
// Parameters:
// - hostIP: a string representing the IP address to check.
//
// Returns:
// - bool: true if the IP address is disallowed, false otherwise.
func IsDisallowedIP(hostIP string) bool {
	// Parse the provided IP address string into a net.IP object.
	ip := net.ParseIP(hostIP)

	// Check if the IP address is in one of the disallowed categories:
	// - IsMulticast: returns true if the IP is a multicast address.
	// - IsUnspecified: returns true if the IP is unspecified (e.g., 0.0.0.0).
	// - IsLoopback: returns true if the IP is a loopback address (e.g., 127.0.0.1).
	// - IsPrivate: returns true if the IP is in a private range (e.g., 192.168.x.x).
	return ip.IsMulticast() || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate()
}

// checkDisallowedIP validates the IP address of a connection to ensure it is allowed.
// If the IP address is disallowed, the connection is closed, and an error is returned.
//
// Parameters:
// - conn: a net.Conn representing the network connection to check.
//
// Returns:
// - error: an error if the IP address is disallowed or if there is an issue closing the connection.
func checkDisallowedIP(conn net.Conn) error {
	// Extract the IP address from the connection's remote address.
	ip, _, _ := net.SplitHostPort(conn.RemoteAddr().String())

	// Check if the IP address is disallowed using the IsDisallowedIP function.
	if IsDisallowedIP(ip) {
		// Close the connection if the IP address is disallowed.
		err := conn.Close()
		if err != nil {
			// Return an error if there was an issue closing the connection.
			return err
		}

		// Return an error indicating the IP address is not allowed.
		return errors.New("ip address is not allowed")
	}

	// Return nil if the IP address is allowed and the connection is valid.
	return nil
}

// dialFunc handles both regular and TLS connections.
// dialFunc handles both regular and TLS connections, establishing a network connection
// based on the provided network, address, timeout, and optional TLS configuration.
// It also checks if the IP address of the connection is allowed.
//
// Parameters:
// - network: the network type (e.g., "tcp", "udp") to use for the connection.
// - addr: the address to connect to, in the form "host:port".
// - timeout: the maximum amount of time allowed for the connection attempt.
// - tlsConfig: an optional *tls.Config for establishing a TLS connection.
//
// Returns:
// - net.Conn: the established network connection.
// - error: an error if the connection fails or if the IP address is disallowed.
func dialFunc(network, addr string, timeout time.Duration, tlsConfig *tls.Config) (net.Conn, error) {
	// Create a net.Dialer with the specified timeout.
	dialer := &net.Dialer{Timeout: timeout}

	var conn net.Conn
	var err error

	// If a TLS configuration is provided, use tls.DialWithDialer to establish a TLS connection.
	// Otherwise, establish a regular network connection using dialer.Dial.
	if tlsConfig != nil {
		conn, err = tls.DialWithDialer(dialer, network, addr, tlsConfig)
	} else {
		conn, err = dialer.Dial(network, addr)
	}

	// If there was an error during the connection attempt, return the error.
	if err != nil {
		return nil, err
	}

	// Check if the IP address of the connection is allowed. If not, return an error.
	if err := checkDisallowedIP(conn); err != nil {
		return nil, err
	}

	// Return the established connection.
	return conn, nil
}

// SafeTransport returns an http.Transport with custom dial functions for both regular and TLS connections.
// The transport is configured with a specified timeout and ensures that all TLS connections use a minimum version of TLS 1.2.
//
// Parameters:
// - timeout: the duration for both the connection timeout and the TLS handshake timeout.
//
// Returns:
// - *http.Transport: a pointer to an http.Transport configured with custom dial functions and security settings.
func SafeTransport(timeout time.Duration) *http.Transport {
	// Configure TLS settings, requiring at least TLS version 1.2.
	tlsConfig := &tls.Config{MinVersion: tls.VersionTLS12}

	return &http.Transport{
		// Custom dial function for regular (non-TLS) connections.
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialFunc(network, addr, timeout, nil)
		},
		// Custom dial function for TLS connections, using the specified TLS configuration.
		DialTLSContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return dialFunc(network, addr, timeout, tlsConfig)
		},
		// Set the timeout for the TLS handshake process.
		TLSHandshakeTimeout: timeout,
	}
}

// SafeGet performs an HTTP GET request to the specified URL using a custom HTTP client with a timeout.
// The function returns the response or an error if the request fails.
//
// Parameters:
// - url: the URL to send the GET request to.
//
// Returns:
// - *http.Response: a pointer to the HTTP response object containing the server's response.
// - error: an error if the request fails or if there is a problem during the request.
func SafeGet(url string) (*http.Response, error) {
	// Create an HTTP client with a custom transport that includes a 10-second timeout.
	httpClient := &http.Client{
		Transport: SafeTransport(10 * time.Second),
	}

	// Perform the GET request using the custom HTTP client and return the response.
	return httpClient.Get(url)
}
