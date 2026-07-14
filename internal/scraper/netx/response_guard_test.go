package netx

import (
	"fmt"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestIsDisallowedIP(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name       string
		ip         string
		disallowed bool
	}{
		{"loopback IPv4", "127.0.0.1", true},
		{"loopback IPv6", "::1", true},
		{"private 10.x", "10.0.0.1", true},
		{"private 192.168.x", "192.168.1.1", true},
		{"private 172.16.x", "172.16.0.1", true},
		{"multicast", "224.0.0.1", true},
		{"unspecified", "0.0.0.0", true},
		{"public IP", "8.8.8.8", false},
		{"public IP 2", "93.184.216.34", false},
		{"invalid IP", "notanip", true},
		{"empty", "", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.disallowed, IsDisallowedIP(tc.ip))
		})
	}
}

// =============================================================================
// Unit Tests: Error helpers
// =============================================================================

func TestCheckHTTPStatus_Blocked(t *testing.T) {
	t.Parallel()

	blockedCodes := []int{http.StatusForbidden, http.StatusTooManyRequests, http.StatusServiceUnavailable}
	for _, code := range blockedCodes {
		t.Run(fmt.Sprintf("status_%d", code), func(t *testing.T) {
			t.Parallel()
			resp := &http.Response{StatusCode: code}
			err := CheckHTTPStatus(resp, "test")
			require.Error(t, err)
			assert.ErrorIs(t, err, ErrSourceUnavailable)
		})
	}
}

func TestCheckHTTPStatus_Success(t *testing.T) {
	t.Parallel()

	resp := &http.Response{StatusCode: http.StatusOK}
	err := CheckHTTPStatus(resp, "test")
	assert.NoError(t, err)
}

func TestCheckHTTPStatus_OtherError(t *testing.T) {
	t.Parallel()

	resp := &http.Response{StatusCode: http.StatusNotFound}
	err := CheckHTTPStatus(resp, "test")
	require.Error(t, err)
	assert.NotErrorIs(t, err, ErrSourceUnavailable, "404 is not a source-unavailable error")
}

func TestCheckHTMLResponse_JSONContentType(t *testing.T) {
	t.Parallel()

	resp := &http.Response{Header: http.Header{"Content-Type": []string{"application/json"}}}
	err := CheckHTMLResponse(resp, []byte(`{"ok":true}`), "test")
	assert.NoError(t, err)
}

func TestCheckHTMLResponse_HTMLContentType(t *testing.T) {
	t.Parallel()

	resp := &http.Response{Header: http.Header{"Content-Type": []string{"text/html; charset=utf-8"}}}
	err := CheckHTMLResponse(resp, []byte(`<html>`), "test")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSourceUnavailable)
}

func TestCheckHTMLResponse_HTMLBody(t *testing.T) {
	t.Parallel()

	resp := &http.Response{Header: http.Header{"Content-Type": []string{"application/octet-stream"}}}
	err := CheckHTMLResponse(resp, []byte(`  <html>`), "test")
	require.Error(t, err)
	assert.ErrorIs(t, err, ErrSourceUnavailable)
}

func TestValidateStreamURL(t *testing.T) {
	t.Parallel()

	tests := []struct {
		name      string
		url       string
		expectErr bool
	}{
		{"valid HTTPS", "https://cdn.example.com/stream.m3u8", false},
		{"valid HTTP", "http://cdn.example.com/stream.m3u8", false},
		{"FTP scheme", "ftp://example.com/file", true},
		{"no scheme", "cdn.example.com/stream", true},
		{"empty", "", true},
		{"relative path", "/video/stream.m3u8", true},
		{"javascript scheme", "javascript:alert(1)", true},
		{"data URI", "data:text/html,<h1>hi</h1>", true},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			result, err := ValidateStreamURL(tc.url, "test")
			if tc.expectErr {
				require.Error(t, err)
				assert.ErrorIs(t, err, ErrInvalidStreamURL)
			} else {
				require.NoError(t, err)
				assert.NotEmpty(t, result)
			}
		})
	}
}

// =============================================================================
// Unit Tests: Scraper manager integration with SuperFlix
// =============================================================================
