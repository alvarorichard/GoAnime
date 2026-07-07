package netx

import (
	"context"
	"errors"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewHTTPStatusErrorClassifiesCloudflareOriginDown(t *testing.T) {
	t.Parallel()

	err := NewHTTPStatusError("SFlix", "search", 521)
	diagnostic := DiagnoseError("SFlix", "search", err)

	require.NotNil(t, diagnostic)
	assert.Equal(t, DiagnosticSourceUnavailable, diagnostic.Kind)
	assert.Equal(t, 521, diagnostic.StatusCode)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))
	assert.True(t, diagnostic.ShouldSkipHealthCheck())
	assert.Contains(t, diagnostic.UserMessage(), "Cloudflare 521")
}

func TestNewHTTPStatusErrorClassifiesBlockedChallenge(t *testing.T) {
	t.Parallel()

	err := NewHTTPStatusError("SFlix", "search", http.StatusTooManyRequests)
	diagnostic := DiagnoseError("SFlix", "search", err)

	require.NotNil(t, diagnostic)
	assert.Equal(t, DiagnosticBlockedChallenge, diagnostic.Kind)
	assert.True(t, errors.Is(err, ErrSourceUnavailable))
	assert.True(t, diagnostic.ShouldOpenCircuit())
}

func TestDiagnoseErrorClassifiesParserAndDecryptFailures(t *testing.T) {
	t.Parallel()

	parserDiagnostic := DiagnoseError("Goyabu", "episode", errors.New("no video URL found in AJAX response"))
	require.NotNil(t, parserDiagnostic)
	assert.Equal(t, DiagnosticParserBroken, parserDiagnostic.Kind)
	assert.False(t, parserDiagnostic.ShouldSkipHealthCheck())

	decryptDiagnostic := DiagnoseError("AllAnime", "episode", errors.New("AES-GCM decryption failed: cipher: message authentication failed"))
	require.NotNil(t, decryptDiagnostic)
	assert.Equal(t, DiagnosticDecryptBroken, decryptDiagnostic.Kind)
	assert.False(t, decryptDiagnostic.ShouldSkipHealthCheck())
}

func TestDiagnoseErrorClassifiesTimeoutAsSourceUnavailable(t *testing.T) {
	t.Parallel()

	diagnostic := DiagnoseError("9Anime", "search", context.DeadlineExceeded)

	require.NotNil(t, diagnostic)
	assert.Equal(t, DiagnosticSourceUnavailable, diagnostic.Kind)
	assert.True(t, errors.Is(diagnostic, ErrSourceUnavailable))
	assert.True(t, diagnostic.ShouldSkipHealthCheck())
}
