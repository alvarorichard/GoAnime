// Package scraper implements provider search, stream extraction, and source diagnostics.
package scraper

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"strconv"
	"strings"
)

// DiagnosticKind identifies the layer where a provider failure happened.
type DiagnosticKind string

const (
	// DiagnosticUnknown is used when no more specific classification exists.
	DiagnosticUnknown DiagnosticKind = "Unknown"
	// DiagnosticSourceUnavailable means the upstream origin/network is down.
	DiagnosticSourceUnavailable DiagnosticKind = "SourceUnavailable"
	// DiagnosticBlockedChallenge means the upstream blocked or challenged us.
	DiagnosticBlockedChallenge DiagnosticKind = "BlockedOrChallenge"
	// DiagnosticParserBroken means the response shape changed after a successful request.
	DiagnosticParserBroken DiagnosticKind = "ParserBroken"
	// DiagnosticDecryptBroken means a decrypt/decode layer changed or returned invalid data.
	DiagnosticDecryptBroken DiagnosticKind = "DecryptBroken"
	// DiagnosticDownloadExpired means an extracted CDN URL expired or was rejected.
	DiagnosticDownloadExpired DiagnosticKind = "DownloadExpired"
	// DiagnosticInternalBug means the failure appears to be local application logic.
	DiagnosticInternalBug DiagnosticKind = "InternalBug"
)

// SourceDiagnostic carries machine-readable context while still behaving like a
// normal wrapped error.
type SourceDiagnostic struct {
	Source     string
	Layer      string
	Kind       DiagnosticKind
	StatusCode int
	Message    string
	Err        error
}

func (d *SourceDiagnostic) Error() string {
	if d == nil {
		return "<nil>"
	}

	message := d.Message
	if message == "" {
		message = string(d.Kind)
	}
	if d.StatusCode > 0 && !strings.Contains(message, strconv.Itoa(d.StatusCode)) {
		message = fmt.Sprintf("%s (HTTP %d)", message, d.StatusCode)
	}

	switch {
	case d.Source != "" && d.Layer != "":
		return fmt.Sprintf("%s %s: %s", d.Source, d.Layer, message)
	case d.Source != "":
		return fmt.Sprintf("%s: %s", d.Source, message)
	default:
		return message
	}
}

func (d *SourceDiagnostic) Unwrap() error {
	if d == nil {
		return nil
	}
	return d.Err
}

// Is lets errors.Is match ErrSourceUnavailable for source-side diagnostics.
func (d *SourceDiagnostic) Is(target error) bool {
	if d == nil {
		return false
	}
	if target == ErrSourceUnavailable {
		return d.Kind == DiagnosticSourceUnavailable || d.Kind == DiagnosticBlockedChallenge
	}
	return errors.Is(d.Err, target)
}

// UserMessage is safe to show in app logs.
func (d *SourceDiagnostic) UserMessage() string {
	if d == nil {
		return "source diagnostic unavailable"
	}

	source := d.Source
	if source == "" {
		source = "Source"
	}

	switch d.Kind {
	case DiagnosticSourceUnavailable:
		if isCloudflareOriginStatus(d.StatusCode) {
			return fmt.Sprintf("%s temporarily unavailable: Cloudflare %d/origin down", source, d.StatusCode)
		}
		if d.StatusCode > 0 {
			return fmt.Sprintf("%s temporarily unavailable: HTTP %d", source, d.StatusCode)
		}
		return fmt.Sprintf("%s temporarily unavailable: %s", source, d.reason())
	case DiagnosticBlockedChallenge:
		if d.StatusCode > 0 {
			return fmt.Sprintf("%s blocked the request: HTTP %d/challenge", source, d.StatusCode)
		}
		return fmt.Sprintf("%s blocked the request: captcha/challenge", source)
	case DiagnosticParserBroken:
		return fmt.Sprintf("%s responded but the parser could not find the expected data: %s", source, d.reason())
	case DiagnosticDecryptBroken:
		return fmt.Sprintf("%s decrypt failed: format or key may have changed", source)
	case DiagnosticDownloadExpired:
		if d.StatusCode > 0 {
			return fmt.Sprintf("%s download link expired or was rejected: HTTP %d", source, d.StatusCode)
		}
		return fmt.Sprintf("%s download link expired or was rejected", source)
	case DiagnosticInternalBug:
		return fmt.Sprintf("%s internal app error: %s", source, d.reason())
	default:
		return fmt.Sprintf("%s failed: %s", source, d.reason())
	}
}

func (d *SourceDiagnostic) reason() string {
	if d == nil {
		return "unknown"
	}
	if d.Message != "" {
		return d.Message
	}
	if d.Err != nil {
		return d.Err.Error()
	}
	return string(d.Kind)
}

// ShouldSkipHealthCheck returns true for provider-side failures. CI health checks
// should skip these instead of failing a PR.
func (d *SourceDiagnostic) ShouldSkipHealthCheck() bool {
	if d == nil {
		return false
	}
	return d.Kind == DiagnosticSourceUnavailable || d.Kind == DiagnosticBlockedChallenge
}

// ShouldOpenCircuit returns true when retrying immediately is likely to hammer a
// dead or blocked upstream provider.
func (d *SourceDiagnostic) ShouldOpenCircuit() bool {
	if d == nil {
		return false
	}
	return d.Kind == DiagnosticSourceUnavailable || d.Kind == DiagnosticBlockedChallenge
}

// NewHTTPStatusError classifies an upstream HTTP status into a diagnostic error.
func NewHTTPStatusError(source, layer string, statusCode int) error {
	kind := DiagnosticParserBroken
	message := fmt.Sprintf("returned HTTP %d", statusCode)

	switch {
	case isBlockedStatus(statusCode):
		kind = DiagnosticBlockedChallenge
		message = fmt.Sprintf("blocked or challenged with HTTP %d", statusCode)
	case isOriginUnavailableStatus(statusCode):
		kind = DiagnosticSourceUnavailable
		message = fmt.Sprintf("upstream unavailable with HTTP %d", statusCode)
	}

	return &SourceDiagnostic{
		Source:     source,
		Layer:      layer,
		Kind:       kind,
		StatusCode: statusCode,
		Message:    message,
		Err:        sentinelForDiagnosticKind(kind),
	}
}

// NewBlockedChallengeError creates a diagnostic for captcha/challenge blocks.
func NewBlockedChallengeError(source, layer, message string, err error) error {
	return &SourceDiagnostic{
		Source:  source,
		Layer:   layer,
		Kind:    DiagnosticBlockedChallenge,
		Message: message,
		Err:     joinDiagnosticErr(err, ErrSourceUnavailable),
	}
}

// NewParserError creates a diagnostic for source markup/API shape changes.
func NewParserError(source, layer, message string, err error) error {
	return &SourceDiagnostic{
		Source:  source,
		Layer:   layer,
		Kind:    DiagnosticParserBroken,
		Message: message,
		Err:     err,
	}
}

// NewDecryptError creates a diagnostic for broken decrypt/decode layers.
func NewDecryptError(source, layer, message string, err error) error {
	return &SourceDiagnostic{
		Source:  source,
		Layer:   layer,
		Kind:    DiagnosticDecryptBroken,
		Message: message,
		Err:     err,
	}
}

// NewDownloadExpiredError creates a diagnostic for expired or rejected CDN links.
func NewDownloadExpiredError(source, layer string, statusCode int, err error) error {
	return &SourceDiagnostic{
		Source:     source,
		Layer:      layer,
		Kind:       DiagnosticDownloadExpired,
		StatusCode: statusCode,
		Message:    fmt.Sprintf("download URL expired or rejected with HTTP %d", statusCode),
		Err:        err,
	}
}

// NewInternalBugError creates a diagnostic for local application errors.
func NewInternalBugError(source, layer, message string, err error) error {
	return &SourceDiagnostic{
		Source:  source,
		Layer:   layer,
		Kind:    DiagnosticInternalBug,
		Message: message,
		Err:     err,
	}
}

// DiagnoseError classifies legacy/plain errors so callers can log and decide
// whether a CI health check should fail or skip.
func DiagnoseError(source, layer string, err error) *SourceDiagnostic {
	if err == nil {
		return nil
	}

	var diag *SourceDiagnostic
	if errors.As(err, &diag) {
		copyDiag := *diag
		if copyDiag.Source == "" {
			copyDiag.Source = source
		}
		if copyDiag.Layer == "" {
			copyDiag.Layer = layer
		}
		return &copyDiag
	}

	if isNetworkUnavailable(err) {
		return &SourceDiagnostic{
			Source:  source,
			Layer:   layer,
			Kind:    DiagnosticSourceUnavailable,
			Message: err.Error(),
			Err:     joinDiagnosticErr(err, ErrSourceUnavailable),
		}
	}

	lower := strings.ToLower(err.Error())
	if status := statusFromMessage(lower); status > 0 {
		typedErr := NewHTTPStatusError(source, layer, status)
		var typedDiag *SourceDiagnostic
		if errors.As(typedErr, &typedDiag) {
			typedDiag.Err = joinDiagnosticErr(err, typedDiag.Err)
			return typedDiag
		}
	}

	switch {
	case containsAny(lower, "captcha", "challenge", "cloudflare", "access denied", "error 1020", "forbidden", "rate limit"):
		return &SourceDiagnostic{Source: source, Layer: layer, Kind: DiagnosticBlockedChallenge, Message: err.Error(), Err: joinDiagnosticErr(err, ErrSourceUnavailable)}
	case containsAny(lower, "decrypt", "decryption", "aes-gcm", "base64 decode", "tobeparsed"):
		return &SourceDiagnostic{Source: source, Layer: layer, Kind: DiagnosticDecryptBroken, Message: err.Error(), Err: err}
	case containsAny(lower, "no source url", "no source urls", "no suitable quality", "no server found", "no embed", "no video url", "no video sources", "failed to parse", "selector", "expected payload"):
		return &SourceDiagnostic{Source: source, Layer: layer, Kind: DiagnosticParserBroken, Message: err.Error(), Err: err}
	case containsAny(lower, "panic", "nil pointer", "deadlock", "infinite loop"):
		return &SourceDiagnostic{Source: source, Layer: layer, Kind: DiagnosticInternalBug, Message: err.Error(), Err: err}
	default:
		return &SourceDiagnostic{Source: source, Layer: layer, Kind: DiagnosticInternalBug, Message: err.Error(), Err: err}
	}
}

func isNetworkUnavailable(err error) bool {
	if errors.Is(err, context.DeadlineExceeded) || errors.Is(err, context.Canceled) {
		return true
	}

	var dnsErr *net.DNSError
	if errors.As(err, &dnsErr) {
		return true
	}

	var netErr net.Error
	if errors.As(err, &netErr) && netErr.Timeout() {
		return true
	}

	lower := strings.ToLower(err.Error())
	return containsAny(lower,
		"no such host",
		"connection refused",
		"connection reset",
		"i/o timeout",
		"tls handshake timeout",
		"context deadline exceeded",
		"timed out",
		"timeout",
	)
}

func isBlockedStatus(statusCode int) bool {
	return statusCode == http.StatusForbidden ||
		statusCode == http.StatusTooManyRequests ||
		statusCode == 1020
}

func isOriginUnavailableStatus(statusCode int) bool {
	switch statusCode {
	case http.StatusInternalServerError,
		http.StatusBadGateway,
		http.StatusServiceUnavailable,
		http.StatusGatewayTimeout,
		521, 522, 523, 524, 530:
		return true
	default:
		return false
	}
}

func isCloudflareOriginStatus(statusCode int) bool {
	switch statusCode {
	case 521, 522, 523, 524, 530:
		return true
	default:
		return false
	}
}

func statusFromMessage(lower string) int {
	statuses := []int{1020, 530, 524, 523, 522, 521, 504, 503, 502, 500, 429, 404, 403, 405}
	for _, status := range statuses {
		if strings.Contains(lower, strconv.Itoa(status)) {
			return status
		}
	}
	return 0
}

func sentinelForDiagnosticKind(kind DiagnosticKind) error {
	if kind == DiagnosticSourceUnavailable || kind == DiagnosticBlockedChallenge {
		return ErrSourceUnavailable
	}
	return nil
}

func joinDiagnosticErr(err, sentinel error) error {
	switch {
	case err == nil:
		return sentinel
	case sentinel == nil:
		return err
	default:
		return errors.Join(err, sentinel)
	}
}

func containsAny(s string, needles ...string) bool {
	for _, needle := range needles {
		if strings.Contains(s, needle) {
			return true
		}
	}
	return false
}
