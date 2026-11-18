package player

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

// Issue #229: on Windows mpv's pipe, once mpv has exited, answers "The system
// cannot find the file specified." isMPVConnectionDead knew only the Unix
// wording, so waitForVideoReady never saw the player was gone and waited its
// full 45 s, twice (the second time after "Watch now").
func TestIsMPVConnectionDeadKnowsWindowsPipeErrors(t *testing.T) {
	t.Parallel()
	for _, msg := range []string{
		`open \\.\pipe\goanime_mpv_18e90d225a2f5c34: The system cannot find the file specified.`,
		`open \\.\pipe\x: The system cannot find the path specified.`,
		"read \\\\.\\pipe\\x: The pipe has been ended.",
		"write \\\\.\\pipe\\x: The pipe is being closed.",
	} {
		assert.True(t, isMPVConnectionDead(errors.New(msg)), msg)
	}
}

func TestPlayerGone(t *testing.T) {
	t.Parallel()
	assert.True(t, playerGone("/tmp/goanime_mpv_does_not_exist.sock"), "no socket: mpv has exited")

	sock := startMockMPVSocket(t, func(map[string]any) []byte { return mpvOK(1234.0) })
	assert.False(t, playerGone(sock), "mpv answers: it is there")
}

func TestSeekToResumePositionReportsAClosedPlayer(t *testing.T) {
	t.Parallel()
	assert.False(t, seekToResumePosition("/tmp/goanime_mpv_does_not_exist.sock", 30),
		"the player closed before the video was ready: say so instead of seeking a dead socket")
	assert.True(t, seekToResumePosition("/tmp/goanime_mpv_does_not_exist.sock", 0), "nothing to seek")
}
