//go:build !windows

package player

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPlayerGoneIsFalseWhileMPVAnswers(t *testing.T) {
	t.Parallel()
	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK(1234.0) })
	assert.False(t, playerGone(sock), "mpv answers: it is there")
}
