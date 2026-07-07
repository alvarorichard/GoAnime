package netx

import (
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
