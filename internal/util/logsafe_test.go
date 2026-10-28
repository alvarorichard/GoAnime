package util

import (
	"bytes"
	"errors"
	"log"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestLogSafe(t *testing.T) {
	t.Parallel()
	tests := []struct {
		name string
		in   any
		want string
	}{
		{"plain title", "Todo Poderoso (2003)", "Todo Poderoso (2003)"},
		{"accents and CJK stay", "Pânico 続き", "Pânico 続き"},
		{"newline cannot start a forged line", "x\n2026/01/01 ERROR forged", "x 2026/01/01 ERROR forged"},
		{"carriage return", "a\r\nb", "a  b"},
		{"escape sequence", "a\x1b[31mred\x1b[0m", "a[31mred[0m"},
		{"tab and NUL", "a\tb\x00c", "abc"},
		{"error value", errors.New("boom\nnext"), "boom next"},
		{"number", 42, "42"},
		{"nil", nil, "<nil>"},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tt.want, LogSafe(tt.in))
		})
	}
}

// TestLogSafe_OneLogLinePerCall is the property the helper exists for: the log
// package writes one line per call only if nothing in the arguments breaks it.
func TestLogSafe_OneLogLinePerCall(t *testing.T) {
	var buf bytes.Buffer
	logger := log.New(&buf, "", 0)

	logger.Printf("Switched to series: %s", LogSafe("Naruto\nERROR forged entry\r"))

	assert.Equal(t, 1, strings.Count(buf.String(), "\n"), "got %q", buf.String())
}
