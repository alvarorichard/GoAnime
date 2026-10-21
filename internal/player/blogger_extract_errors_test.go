package player

import (
	"errors"
	"fmt"
	"strings"
	"testing"
)

func isTransientBloggerExtractError(err error) bool {
	if err == nil {
		return false
	}
	var rpcErr bloggerRPCError
	if errors.As(err, &rpcErr) {
		_, terminal := bloggerRPCStatusName(rpcErr.Code)
		return !terminal
	}
	errMsg := strings.ToLower(err.Error())
	transient := []string{
		"no such host",
		"timeout",
		"temporary failure",
		"connection refused",
		"connection reset",
		"network is unreachable",
		"tls handshake timeout",
		"server returned",
		"status 403",
		"status 429",
		"status 500",
		"status 502",
		"status 503",
		"status 521",
		"status 522",
		"status 523",
		"status 524",
		"status 530",
	}
	for _, marker := range transient {
		if strings.Contains(errMsg, marker) {
			return true
		}
	}
	return false
}

func TestIsTransientBloggerExtractError_RPCStatus(t *testing.T) {
	for _, code := range []int{2, 4, 8, 13, 14} {
		t.Run(fmt.Sprintf("retryable_%d", code), func(t *testing.T) {
			err := fmt.Errorf("Blogger extraction failed: %w", bloggerRPCError{Code: code})
			if !isTransientBloggerExtractError(err) {
				t.Fatalf("RPC status %d should remain retryable in live diagnostics", code)
			}
		})
	}
	for _, code := range []int{3, 5, 7, 9, 16} {
		t.Run(fmt.Sprintf("terminal_%d", code), func(t *testing.T) {
			err := fmt.Errorf("request timeout followed by: %w", bloggerRPCError{Code: code})
			if isTransientBloggerExtractError(err) {
				t.Fatalf("terminal RPC status %d must fail live diagnostics", code)
			}
		})
	}
}

func TestIsTransientBloggerExtractError(t *testing.T) {
	for _, tc := range []struct {
		name string
		err  error
		want bool
	}{
		{name: "nil"},
		{name: "network", err: errors.New("TLS handshake timeout"), want: true},
		{name: "http", err: errors.New("server returned status 503"), want: true},
		{name: "parser", err: errors.New("no video URL found in batchexecute response")},
	} {
		t.Run(tc.name, func(t *testing.T) {
			if got := isTransientBloggerExtractError(tc.err); got != tc.want {
				t.Errorf("isTransientBloggerExtractError(%v) = %v, want %v", tc.err, got, tc.want)
			}
		})
	}
}
