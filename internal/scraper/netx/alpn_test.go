package netx

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"io"
	"net"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"
)

func TestSafeTLSConfig_AdvertisesHTTP2(t *testing.T) {
	cfg := SafeTLSConfig()
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %x, want TLS 1.2", cfg.MinVersion)
	}
	want := []string{"h2", "http/1.1"}
	if len(cfg.NextProtos) != len(want) {
		t.Fatalf("NextProtos = %v, want %v", cfg.NextProtos, want)
	}
	for i, p := range want {
		if cfg.NextProtos[i] != p {
			t.Fatalf("NextProtos = %v, want %v", cfg.NextProtos, want)
		}
	}
}

func TestSafeTLSConfig_ReturnsAFreshConfigEachCall(t *testing.T) {
	a, b := SafeTLSConfig(), SafeTLSConfig()
	if a == b {
		t.Fatal("callers must not share one *tls.Config: net/http mutates it")
	}
	a.NextProtos = append(a.NextProtos, "mutated")
	if len(b.NextProtos) != 2 {
		t.Fatal("configs share their NextProtos backing array")
	}
}

func TestSafeScraperTransport_ForceAttemptHTTP2(t *testing.T) {
	tr := SafeScraperTransport(5 * time.Second)
	if !tr.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 must be set: with a custom DialTLSContext, " +
			"net/http will not enable HTTP/2 without it")
	}
	if tr.DialTLSContext == nil || tr.DialContext == nil {
		t.Error("SSRF dial hooks lost")
	}
	if tr.MaxIdleConnsPerHost == 0 {
		t.Error("connection pooling settings lost")
	}
}

// h2Probe dials a TLS server with the given config and reports the negotiated
// protocol as seen by net/http. It mirrors the production transport exactly,
// minus the loopback IP guard (which exists precisely to stop us from reaching
// a local test server).
func h2Probe(t *testing.T, srv *httptest.Server, cfg *tls.Config) string {
	t.Helper()
	pool := x509.NewCertPool()
	pool.AddCert(srv.Certificate())
	cfg.RootCAs = pool

	tr := &http.Transport{
		ForceAttemptHTTP2: true,
		DialTLSContext: func(_ context.Context, network, addr string) (net.Conn, error) {
			return tls.DialWithDialer(&net.Dialer{Timeout: 5 * time.Second}, network, addr, cfg)
		},
	}
	t.Cleanup(tr.CloseIdleConnections)

	resp, err := (&http.Client{Transport: tr, Timeout: 10 * time.Second}).Get(srv.URL)
	if err != nil {
		t.Fatalf("GET %s: %v", srv.URL, err)
	}
	defer func() { _ = resp.Body.Close() }()
	if _, err := io.Copy(io.Discard, resp.Body); err != nil {
		t.Fatalf("read body: %v", err)
	}
	return resp.Proto
}

func h2TestServer(t *testing.T) *httptest.Server {
	t.Helper()
	srv := httptest.NewUnstartedServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		_, _ = io.WriteString(w, r.Proto)
	}))
	srv.EnableHTTP2 = true
	srv.StartTLS()
	t.Cleanup(srv.Close)
	return srv
}

// TestSafeTLSConfigNegotiatesHTTP2 is the end-to-end proof that the ALPN list
// does what it claims against a real HTTP/2 server.
func TestSafeTLSConfigNegotiatesHTTP2(t *testing.T) {
	srv := h2TestServer(t)
	if got := h2Probe(t, srv, SafeTLSConfig()); got != "HTTP/2.0" {
		t.Fatalf("negotiated %s, want HTTP/2.0", got)
	}
}

// TestWithoutALPNFallsBackToHTTP11 is the negative control: the same transport
// with the ALPN list removed — the configuration this repository shipped before
// — silently drops to HTTP/1.1. Without this case the test above could pass for
// the wrong reason.
func TestWithoutALPNFallsBackToHTTP11(t *testing.T) {
	srv := h2TestServer(t)
	cfg := SafeTLSConfig()
	cfg.NextProtos = nil
	if got := h2Probe(t, srv, cfg); got != "HTTP/1.1" {
		t.Fatalf("negotiated %s, want HTTP/1.1 without ALPN", got)
	}
}

// TestSafeDialFunc_StillRejectsLoopback guards the property the ALPN change must
// not weaken.
func TestSafeDialFunc_StillRejectsLoopback(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		t.Fatalf("listen: %v", err)
	}
	defer func() { _ = ln.Close() }()
	go func() {
		for {
			c, err := ln.Accept()
			if err != nil {
				return
			}
			_ = c.Close()
		}
	}()

	if _, err := SafeDialFunc("tcp", ln.Addr().String(), 2*time.Second, nil); err == nil {
		t.Fatal("expected the loopback address to be rejected")
	}
}
