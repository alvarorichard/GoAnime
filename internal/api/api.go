//package api
//
//import (
//	"crypto/tls"
//	"net"
//	"net/http"
//	"time"
//
//	"github.com/pkg/errors"
//	"golang.org/x/net/context"
//)
//
//// IsDisallowedIP checks if the given IP is disallowed.
//func IsDisallowedIP(hostIP string) bool {
//	ip := net.ParseIP(hostIP)
//	return ip.IsMulticast() || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate()
//}
//
//// customDial checks if the remote IP is allowed and returns a connection.
//func customDial(network, addr string, timeout time.Duration) (net.Conn, error) {
//	c, err := net.DialTimeout(network, addr, timeout)
//	if err != nil {
//		return nil, err
//	}
//	ip, _, _ := net.SplitHostPort(c.RemoteAddr().String())
//	if IsDisallowedIP(ip) {
//		return nil, errors.New("ip address is not allowed")
//	}
//	return c, nil
//}
//
//// customDialTLS checks if the remote IP is allowed and performs TLS handshake.
//func customDialTLS(network, addr string, timeout time.Duration) (net.Conn, error) {
//	dialer := &net.Dialer{Timeout: timeout}
//	c, err := tls.DialWithDialer(dialer, network, addr, &tls.Config{
//		MinVersion: tls.VersionTLS12,
//	})
//	if err != nil {
//		return nil, err
//	}
//	ip, _, _ := net.SplitHostPort(c.RemoteAddr().String())
//	if IsDisallowedIP(ip) {
//		return nil, errors.New("ip address is not allowed")
//	}
//	return c, c.Handshake()
//}
//
//// SafeTransport returns a http.Transport with custom dial functions.
//func SafeTransport(timeout time.Duration) *http.Transport {
//	return &http.Transport{
//		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
//			return customDial(network, addr, timeout)
//		},
//		DialTLS: func(network, addr string) (net.Conn, error) {
//			return customDialTLS(network, addr, timeout)
//		},
//		TLSHandshakeTimeout: timeout,
//	}
//}
//
//// SafeGet performs a GET request with a timeout and returns the response.
//func SafeGet(url string) (*http.Response, error) {
//	const clientConnectTimeout = time.Second * 10
//	httpClient := &http.Client{
//		Transport: SafeTransport(clientConnectTimeout),
//	}
//	return httpClient.Get(url)
//}

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

// customDial checks if the remote IP is allowed and returns a connection.
func customDial(network, addr string, timeout time.Duration) (net.Conn, error) {
	c, err := net.DialTimeout(network, addr, timeout)
	if err != nil {
		return nil, err
	}
	if ip, _, _ := net.SplitHostPort(c.RemoteAddr().String()); IsDisallowedIP(ip) {
		return nil, errors.New("ip address is not allowed")
	}
	return c, nil
}

// customDialTLS checks if the remote IP is allowed and performs TLS handshake.
func customDialTLS(network, addr string, timeout time.Duration) (net.Conn, error) {
	dialer := &net.Dialer{Timeout: timeout}
	c, err := tls.DialWithDialer(dialer, network, addr, &tls.Config{MinVersion: tls.VersionTLS12})
	if err != nil {
		return nil, err
	}
	if ip, _, _ := net.SplitHostPort(c.RemoteAddr().String()); IsDisallowedIP(ip) {
		return nil, errors.New("ip address is not allowed")
	}
	return c, c.Handshake()
}

// SafeTransport returns a http.Transport with custom dial functions.
func SafeTransport(timeout time.Duration) *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			return customDial(network, addr, timeout)
		},
		DialTLS: func(network, addr string) (net.Conn, error) {
			return customDialTLS(network, addr, timeout)
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