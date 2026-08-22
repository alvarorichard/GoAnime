package hls

import (
	"net/http"
	"testing"
)

// httpTransport aliases *http.Transport so the assertion below reads clearly.
type httpTransport = http.Transport

// TestDownloaderStaysOnHTTP11 pins a deliberate exception to the HTTP/2 rollout.
// The SSRF-safe transports elsewhere now advertise h2 via ALPN, but the segment
// downloader must not: CDNs reset multiplexed HTTP/2 streams with INTERNAL_ERROR
// when dozens of segments are fetched concurrently over one connection, so this
// client keeps one TCP connection per request on purpose.
func TestDownloaderStaysOnHTTP11(t *testing.T) {
	d := NewDownloader()
	tr, ok := d.client.Transport.(*httpTransport)
	if !ok {
		t.Fatalf("unexpected transport type %T", d.client.Transport)
	}
	if tr.TLSNextProto == nil || len(tr.TLSNextProto) != 0 {
		t.Error("TLSNextProto must be a non-nil empty map to keep HTTP/2 disabled")
	}
	if tr.ForceAttemptHTTP2 {
		t.Error("ForceAttemptHTTP2 must stay off for segment downloads")
	}
	if tr.TLSClientConfig == nil || len(tr.TLSClientConfig.NextProtos) != 0 {
		t.Error("the segment downloader must not advertise h2 via ALPN")
	}
	if tr.DialContext == nil {
		t.Error("SSRF dial guard lost")
	}
}
