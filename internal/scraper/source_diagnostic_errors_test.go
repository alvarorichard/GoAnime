package scraper

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSourceDiagnostic_Error(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		d    *SourceDiagnostic
		want string
	}{
		{"nil", nil, "<nil>"},
		{"full", &SourceDiagnostic{Source: "X", Layer: "search", Message: "msg"}, "X search: msg"},
		{"source only", &SourceDiagnostic{Source: "X", Message: "msg"}, "X: msg"},
		{"no source", &SourceDiagnostic{Message: "msg"}, "msg"},
		{"status appended", &SourceDiagnostic{Source: "X", Layer: "l", Kind: DiagnosticSourceUnavailable, StatusCode: 503, Message: "down"}, "X l: down (HTTP 503)"},
		{"kind fallback when no message", &SourceDiagnostic{Source: "X", Kind: DiagnosticParserBroken}, "X: ParserBroken"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.d.Error())
		})
	}
}

func TestSourceDiagnostic_Unwrap(t *testing.T) {
	t.Parallel()
	inner := errors.New("inner")
	d := &SourceDiagnostic{Err: inner}
	assert.Equal(t, inner, d.Unwrap())
	assert.Nil(t, (*SourceDiagnostic)(nil).Unwrap())
}

func TestSourceDiagnostic_Is(t *testing.T) {
	t.Parallel()
	t.Run("matches ErrSourceUnavailable when source unavailable", func(t *testing.T) {
		t.Parallel()
		d := &SourceDiagnostic{Kind: DiagnosticSourceUnavailable}
		assert.True(t, errors.Is(d, ErrSourceUnavailable))
	})
	t.Run("matches ErrSourceUnavailable when blocked", func(t *testing.T) {
		t.Parallel()
		d := &SourceDiagnostic{Kind: DiagnosticBlockedChallenge}
		assert.True(t, errors.Is(d, ErrSourceUnavailable))
	})
	t.Run("does not match for parser broken", func(t *testing.T) {
		t.Parallel()
		d := &SourceDiagnostic{Kind: DiagnosticParserBroken}
		assert.False(t, errors.Is(d, ErrSourceUnavailable))
	})
	t.Run("matches wrapped sentinel", func(t *testing.T) {
		t.Parallel()
		sentinel := errors.New("custom")
		d := &SourceDiagnostic{Err: sentinel, Kind: DiagnosticInternalBug}
		assert.True(t, errors.Is(d, sentinel))
	})
	t.Run("nil returns false", func(t *testing.T) {
		t.Parallel()
		assert.False(t, (*SourceDiagnostic)(nil).Is(ErrSourceUnavailable))
	})
}

func TestSourceDiagnostic_UserMessage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		d    *SourceDiagnostic
		want string
	}{
		{"nil", nil, "source diagnostic unavailable"},
		{"cloudflare origin", &SourceDiagnostic{Source: "X", Kind: DiagnosticSourceUnavailable, StatusCode: 522}, "X temporarily unavailable: Cloudflare 522/origin down"},
		{"unavailable status", &SourceDiagnostic{Source: "X", Kind: DiagnosticSourceUnavailable, StatusCode: 503}, "X temporarily unavailable: HTTP 503"},
		{"unavailable no status", &SourceDiagnostic{Source: "X", Kind: DiagnosticSourceUnavailable, Message: "dns"}, "X temporarily unavailable: dns"},
		{"blocked status", &SourceDiagnostic{Source: "X", Kind: DiagnosticBlockedChallenge, StatusCode: 403}, "X blocked the request: HTTP 403/challenge"},
		{"blocked no status", &SourceDiagnostic{Source: "X", Kind: DiagnosticBlockedChallenge}, "X blocked the request: captcha/challenge"},
		{"parser broken", &SourceDiagnostic{Source: "X", Kind: DiagnosticParserBroken, Message: "no selector"}, "X responded but the parser could not find the expected data: no selector"},
		{"decrypt", &SourceDiagnostic{Source: "X", Kind: DiagnosticDecryptBroken}, "X decrypt failed: format or key may have changed"},
		{"download expired with status", &SourceDiagnostic{Source: "X", Kind: DiagnosticDownloadExpired, StatusCode: 410}, "X download link expired or was rejected: HTTP 410"},
		{"download expired no status", &SourceDiagnostic{Source: "X", Kind: DiagnosticDownloadExpired}, "X download link expired or was rejected"},
		{"internal bug", &SourceDiagnostic{Source: "X", Kind: DiagnosticInternalBug, Message: "nil ptr"}, "X internal app error: nil ptr"},
		{"unknown kind", &SourceDiagnostic{Source: "X", Kind: DiagnosticUnknown, Message: "?"}, "X failed: ?"},
		{"missing source defaults", &SourceDiagnostic{Kind: DiagnosticDecryptBroken}, "Source decrypt failed: format or key may have changed"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, tt.d.UserMessage())
		})
	}
}

func TestSourceDiagnostic_ShouldSkipHealthCheckAndOpenCircuit(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		kind DiagnosticKind
		skip bool
	}{
		{"source unavailable", DiagnosticSourceUnavailable, true},
		{"blocked", DiagnosticBlockedChallenge, true},
		{"parser", DiagnosticParserBroken, false},
		{"decrypt", DiagnosticDecryptBroken, false},
		{"download expired", DiagnosticDownloadExpired, false},
		{"internal bug", DiagnosticInternalBug, false},
		{"unknown", DiagnosticUnknown, false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			d := &SourceDiagnostic{Kind: tt.kind}
			assert.Equal(t, tt.skip, d.ShouldSkipHealthCheck())
			assert.Equal(t, tt.skip, d.ShouldOpenCircuit())
		})
	}

	assert.False(t, (*SourceDiagnostic)(nil).ShouldSkipHealthCheck())
	assert.False(t, (*SourceDiagnostic)(nil).ShouldOpenCircuit())
}

func TestNewParserError(t *testing.T) {
	t.Parallel()
	err := NewParserError("X", "search", "no selector", errors.New("orig"))
	var diag *SourceDiagnostic
	assert.True(t, errors.As(err, &diag))
	assert.Equal(t, DiagnosticParserBroken, diag.Kind)
	assert.Equal(t, "no selector", diag.Message)
}

func TestNewDecryptError(t *testing.T) {
	t.Parallel()
	err := NewDecryptError("X", "decrypt", "aes failure", nil)
	var diag *SourceDiagnostic
	assert.True(t, errors.As(err, &diag))
	assert.Equal(t, DiagnosticDecryptBroken, diag.Kind)
}

func TestNewDownloadExpiredError(t *testing.T) {
	t.Parallel()
	err := NewDownloadExpiredError("X", "stream", 410, nil)
	var diag *SourceDiagnostic
	assert.True(t, errors.As(err, &diag))
	assert.Equal(t, DiagnosticDownloadExpired, diag.Kind)
	assert.Equal(t, 410, diag.StatusCode)
}

func TestNewInternalBugError(t *testing.T) {
	t.Parallel()
	err := NewInternalBugError("X", "logic", "nil ptr", nil)
	var diag *SourceDiagnostic
	assert.True(t, errors.As(err, &diag))
	assert.Equal(t, DiagnosticInternalBug, diag.Kind)
}

func TestIsBlockedStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code int
		want bool
	}{
		{403, true},
		{429, true},
		{1020, true},
		{200, false},
		{500, false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isBlockedStatus(tt.code))
	}
}

func TestIsOriginUnavailableStatus(t *testing.T) {
	t.Parallel()
	tests := []struct {
		code int
		want bool
	}{
		{500, true},
		{502, true},
		{503, true},
		{504, true},
		{521, true},
		{530, true},
		{200, false},
		{403, false},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, isOriginUnavailableStatus(tt.code))
	}
}

func TestIsCloudflareOriginStatus(t *testing.T) {
	t.Parallel()
	for _, code := range []int{521, 522, 523, 524, 530} {
		assert.True(t, isCloudflareOriginStatus(code), code)
	}
	assert.False(t, isCloudflareOriginStatus(503))
	assert.False(t, isCloudflareOriginStatus(200))
}

func TestStatusFromMessage(t *testing.T) {
	t.Parallel()
	tests := []struct {
		in   string
		want int
	}{
		{"http 503 service", 503},
		{"got 429 rate limit", 429},
		{"cf 1020", 1020},
		{"nothing here", 0},
	}
	for _, tt := range tests {
		assert.Equal(t, tt.want, statusFromMessage(tt.in))
	}
}

func TestContainsAny(t *testing.T) {
	t.Parallel()
	assert.True(t, containsAny("foo bar", "qux", "bar"))
	assert.False(t, containsAny("foo bar", "x", "y"))
	assert.False(t, containsAny("", "x"))
}

func TestIsNetworkUnavailable(t *testing.T) {
	t.Parallel()
	assert.True(t, isNetworkUnavailable(errors.New("connection refused")))
	assert.True(t, isNetworkUnavailable(errors.New("i/o timeout")))
	assert.True(t, isNetworkUnavailable(errors.New("no such host")))
	assert.False(t, isNetworkUnavailable(errors.New("malformed payload")))
}

func TestDiagnoseError_NilReturnsNil(t *testing.T) {
	t.Parallel()
	assert.Nil(t, DiagnoseError("X", "l", nil))
}

func TestDiagnoseError_BlockedKeywords(t *testing.T) {
	t.Parallel()
	d := DiagnoseError("X", "l", errors.New("got captcha"))
	assert.Equal(t, DiagnosticBlockedChallenge, d.Kind)
}

func TestDiagnoseError_DecryptKeyword(t *testing.T) {
	t.Parallel()
	d := DiagnoseError("X", "l", errors.New("aes-gcm failed"))
	assert.Equal(t, DiagnosticDecryptBroken, d.Kind)
}

func TestDiagnoseError_ParserKeyword(t *testing.T) {
	t.Parallel()
	d := DiagnoseError("X", "l", errors.New("no source urls"))
	assert.Equal(t, DiagnosticParserBroken, d.Kind)
}

func TestDiagnoseError_InternalKeyword(t *testing.T) {
	t.Parallel()
	d := DiagnoseError("X", "l", errors.New("nil pointer dereference"))
	assert.Equal(t, DiagnosticInternalBug, d.Kind)
}

func TestDiagnoseError_FillsMissingSourceLayer(t *testing.T) {
	t.Parallel()
	inner := &SourceDiagnostic{Kind: DiagnosticParserBroken, Message: "x"}
	d := DiagnoseError("Source", "layer", inner)
	assert.Equal(t, "Source", d.Source)
	assert.Equal(t, "layer", d.Layer)
}

func TestSentinelForDiagnosticKind(t *testing.T) {
	t.Parallel()
	assert.Equal(t, ErrSourceUnavailable, sentinelForDiagnosticKind(DiagnosticSourceUnavailable))
	assert.Equal(t, ErrSourceUnavailable, sentinelForDiagnosticKind(DiagnosticBlockedChallenge))
	assert.Nil(t, sentinelForDiagnosticKind(DiagnosticParserBroken))
}

func TestJoinDiagnosticErr(t *testing.T) {
	t.Parallel()
	a := errors.New("a")
	b := errors.New("b")
	assert.Equal(t, b, joinDiagnosticErr(nil, b))
	assert.Equal(t, a, joinDiagnosticErr(a, nil))
	joined := joinDiagnosticErr(a, b)
	assert.True(t, errors.Is(joined, a))
	assert.True(t, errors.Is(joined, b))
}
