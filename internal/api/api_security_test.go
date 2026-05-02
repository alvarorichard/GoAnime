package api

import (
	"context"
	"crypto/tls"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDisallowedIP(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name string
		ip   string
		want bool
	}{
		{name: "empty", ip: "", want: true},
		{name: "invalid", ip: "not-an-ip", want: true},
		{name: "loopback ipv4", ip: "127.0.0.1", want: true},
		{name: "loopback ipv6", ip: "::1", want: true},
		{name: "private ipv4", ip: "192.168.1.10", want: true},
		{name: "multicast", ip: "224.0.0.1", want: true},
		{name: "unspecified", ip: "0.0.0.0", want: true},
		{name: "public", ip: "8.8.8.8", want: false},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.want, IsDisallowedIP(tc.ip))
		})
	}
}

func TestValidateExternalURL(t *testing.T) {
	t.Parallel()

	testCases := []struct {
		name        string
		rawURL      string
		wantErrLike string
	}{
		{name: "invalid url", rawURL: "://bad", wantErrLike: "invalid URL"},
		{name: "missing hostname", rawURL: "http://", wantErrLike: "URL has no hostname"},
		{name: "loopback ipv4", rawURL: "http://127.0.0.1:8080", wantErrLike: "disallowed IP 127.0.0.1"},
		{name: "loopback ipv6", rawURL: "http://[::1]:8080", wantErrLike: "disallowed IP ::1"},
		{name: "public ip literal", rawURL: "http://8.8.8.8", wantErrLike: ""},
	}

	for _, tc := range testCases {
		tc := tc
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()

			err := ValidateExternalURL(tc.rawURL)
			if tc.wantErrLike == "" {
				require.NoError(t, err)
				return
			}

			require.Error(t, err)
			assert.Contains(t, err.Error(), tc.wantErrLike)
		})
	}
}

func TestSafeTransportConfig(t *testing.T) {
	t.Parallel()

	timeout := 7 * time.Second
	transport := SafeTransport(timeout)
	require.NotNil(t, transport)
	require.NotNil(t, transport.DialContext)
	require.NotNil(t, transport.DialTLSContext)
	require.NotNil(t, transport.TLSClientConfig)

	assert.Equal(t, timeout, transport.TLSHandshakeTimeout)
	assert.Equal(t, uint16(0), transport.TLSClientConfig.MaxVersion)
	assert.Equal(t, uint16(tls.VersionTLS12), transport.TLSClientConfig.MinVersion)
	assert.Equal(t, 200, transport.MaxIdleConns)
	assert.Equal(t, 25, transport.MaxIdleConnsPerHost)
	assert.Equal(t, 120*time.Second, transport.IdleConnTimeout)
}

func TestSafeDialContextRejectsLoopbackConnection(t *testing.T) {
	listener, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	defer listener.Close()

	done := make(chan struct{})
	go func() {
		defer close(done)
		conn, acceptErr := listener.Accept()
		if acceptErr == nil {
			_ = conn.Close()
		}
	}()

	dialContext := SafeDialContext(time.Second)
	_, err = dialContext(context.Background(), "tcp", listener.Addr().String())
	require.Error(t, err)
	assert.Contains(t, err.Error(), "ip address is not allowed")
	<-done
}

func TestSafeGetRejectsLoopbackServer(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	resp, err := SafeGet(server.URL)
	require.Error(t, err)
	assert.Nil(t, resp)
	assert.Contains(t, err.Error(), "ip address is not allowed")
}
