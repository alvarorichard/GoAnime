package api

import (
	"crypto/tls"
	"net"
	"net/http"
	"time"

	"golang.org/x/net/context"

	"github.com/pkg/errors"
)

func IsDisallowedIP(hostIP string) bool {
	ip := net.ParseIP(hostIP)
	return ip.IsMulticast() || ip.IsUnspecified() || ip.IsLoopback() || ip.IsPrivate()
}

// SafeTransport returns a http.Transport with a custom DialContext and DialTLS functions that check if the remote IP is allowed.
func SafeTransport(timeout time.Duration) *http.Transport {
	return &http.Transport{
		DialContext: func(ctx context.Context, network, addr string) (net.Conn, error) {
			c, err := net.DialTimeout(network, addr, timeout)
			if err != nil {
				return nil, err
			}
			ip, _, _ := net.SplitHostPort(c.RemoteAddr().String())
			if IsDisallowedIP(ip) {
				return nil, errors.New("ip address is not allowed")
			}
			return c, err
		},
		DialTLS: func(network, addr string) (net.Conn, error) {
			dialer := &net.Dialer{Timeout: timeout}
			c, err := tls.DialWithDialer(dialer, network, addr, &tls.Config{
				MinVersion: tls.VersionTLS12, // Set minimum TLS version to 1.3 or 1.2 in case break download
			})
			if err != nil {
				return nil, err
			}

			ip, _, _ := net.SplitHostPort(c.RemoteAddr().String())
			if IsDisallowedIP(ip) {
				return nil, errors.New("ip address is not allowed")
			}

			err = c.Handshake()
			if err != nil {
				return c, err
			}

			return c, c.Handshake()
		},
		TLSHandshakeTimeout: timeout,
	}
}

// GetAnimeEpisodes returns a slice of Episode structs containing the episode number and URL.

// SafeGet performs a GET request with a timeout and returns the response.
func SafeGet(url string) (*http.Response, error) {
	const clientConnectTimeout = time.Second * 10
	httpClient := &http.Client{
		Transport: SafeTransport(clientConnectTimeout),
	}
	return httpClient.Get(url)
}
