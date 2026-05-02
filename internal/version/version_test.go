package version

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDisplayVersion(t *testing.T) {
	originalVersion := Version
	t.Cleanup(func() {
		Version = originalVersion
	})

	tests := []struct {
		name     string
		version  string
		expected string
	}{
		{name: "plain version", version: "1.8.1", expected: "1.8.1"},
		{name: "lowercase prefix", version: "v1.8.1", expected: "1.8.1"},
		{name: "uppercase prefix with spaces", version: "  V1.8.1  ", expected: "1.8.1"},
		{name: "empty version", version: "   ", expected: "unknown"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			Version = tt.version
			assert.Equal(t, tt.expected, DisplayVersion())
		})
	}
}
