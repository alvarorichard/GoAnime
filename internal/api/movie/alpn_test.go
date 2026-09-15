package movie

import (
	"crypto/tls"
	"testing"
	"time"
)

func TestSafeTLSConfig_AdvertisesHTTP2(t *testing.T) {
	cfg := safeTLSConfig()
	if cfg.MinVersion != tls.VersionTLS12 {
		t.Errorf("MinVersion = %x, want TLS 1.2", cfg.MinVersion)
	}
	if len(cfg.NextProtos) != 2 || cfg.NextProtos[0] != "h2" || cfg.NextProtos[1] != "http/1.1" {
		t.Fatalf("NextProtos = %v, want [h2 http/1.1]", cfg.NextProtos)
	}
}

func TestSafeMovieTransport_EnablesHTTP2(t *testing.T) {
	tr := safeMovieTransport(5 * time.Second)
	if !tr.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 must be set alongside the ALPN list")
	}
	if tr.DialContext == nil || tr.DialTLSContext == nil {
		t.Error("SSRF dial hooks lost")
	}
}
