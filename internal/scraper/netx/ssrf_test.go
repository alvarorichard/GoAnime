package netx

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

func TestSafeDialFunc_RejectsLoopback(t *testing.T) {
	t.Parallel()
	srv := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
	t.Cleanup(srv.Close)

	_, err := SafeDialFunc("tcp", srv.Listener.Addr().String(), 2*time.Second, nil)
	assert.Error(t, err)
}

func TestSafeScraperTransport(t *testing.T) {
	t.Parallel()
	tr := SafeScraperTransport(5 * time.Second)
	require.NotNil(t, tr)
	assert.NotNil(t, tr.DialContext)
	assert.NotNil(t, tr.DialTLSContext)
	assert.Equal(t, 5*time.Second, tr.TLSHandshakeTimeout)
	assert.Equal(t, 100, tr.MaxIdleConns)
	assert.Equal(t, 15, tr.MaxIdleConnsPerHost)
}

// TestSafeDialFunc_RejectsBeforeConnecting pins that the check runs before
// connect(2): a disallowed listener never receives a connection — not even
// one closed right after, which is all a check on RemoteAddr can manage, and
// for TLS only after the ClientHello has already gone out.
func TestSafeDialFunc_RejectsBeforeConnecting(t *testing.T) {
	t.Parallel()
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	t.Cleanup(func() { _ = ln.Close() })

	for _, tlsConfig := range []*tls.Config{nil, SafeTLSConfig()} {
		start := time.Now()
		_, err := SafeDialFunc("tcp", ln.Addr().String(), 2*time.Second, tlsConfig)
		require.Error(t, err)
		assert.Less(t, time.Since(start), time.Second, "refusal should not wait on a handshake")
	}
	tcp, ok := ln.(*net.TCPListener)
	require.True(t, ok)
	require.NoError(t, tcp.SetDeadline(time.Now().Add(200*time.Millisecond)))
	if conn, err := ln.Accept(); err == nil {
		_ = conn.Close()
		t.Fatal("the disallowed listener accepted a connection: the check ran after connect")
	}
}

func TestSafeDialContext_HonorsCancellation(t *testing.T) {
	t.Parallel()
	ctx, cancel := context.WithCancel(context.Background())
	cancel()
	_, err := SafeDialContext(ctx, "tcp", "192.0.2.1:443", 5*time.Second, SafeTLSConfig())
	assert.ErrorIs(t, err, context.Canceled)
}
