package goyabu

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

// ---------------------------------------------------------------------------
// GoyabuClient.sleep
// ---------------------------------------------------------------------------

func TestGoyabuClient_Sleep_ZeroDelay(t *testing.T) {
	t.Parallel()
	client := NewGoyabuClient()
	client.retryDelay = 0
	// Must return without blocking
	client.sleep()
}

func TestGoyabuClient_Sleep_SmallDelay(t *testing.T) {
	t.Parallel()
	client := NewGoyabuClient()
	client.retryDelay = 1 * time.Millisecond
	start := time.Now()
	client.sleep()
	assert.GreaterOrEqual(t, time.Since(start), time.Millisecond)
}
