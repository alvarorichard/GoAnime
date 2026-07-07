package util

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStrictSourceResolution(t *testing.T) {
	// Uses t.Setenv — not parallel.
	tests := []struct {
		name  string
		value string
		want  bool
	}{
		{"unset", "", false},
		{"one", "1", true},
		{"true lower", "true", true},
		{"true upper", "TRUE", true},
		{"zero", "0", false},
		{"false", "false", false},
		{"garbage", "yes", false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			t.Setenv("GOANIME_STRICT_SOURCE", tt.value)
			assert.Equal(t, tt.want, StrictSourceResolution())
		})
	}
}
