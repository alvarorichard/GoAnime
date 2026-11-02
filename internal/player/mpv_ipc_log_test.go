//go:build !windows

package player

import (
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// fakeMPV answers each connection with the given line, as mpv's IPC socket
// does for one command.
func fakeMPV(t *testing.T, reply string) string {
	t.Helper()
	dir, err := os.MkdirTemp("", "mpv")
	require.NoError(t, err)
	t.Cleanup(func() { _ = os.RemoveAll(dir) })
	sock := filepath.Join(dir, "s")
	l, err := net.Listen("unix", sock)
	require.NoError(t, err)
	t.Cleanup(func() { _ = l.Close() })
	go func() {
		for {
			c, err := l.Accept()
			if err != nil {
				return
			}
			go func() {
				defer func() { _ = c.Close() }()
				buf := make([]byte, 512)
				_, _ = c.Read(buf)
				_, _ = fmt.Fprint(c, reply)
			}()
		}
	}()
	return sock
}

func captureMPVLog(t *testing.T) func() []string {
	t.Helper()
	var mu sync.Mutex
	var lines []string
	prev := mpvDebugf
	mpvDebugf = func(format string, args ...any) {
		mu.Lock()
		defer mu.Unlock()
		lines = append(lines, fmt.Sprintf(format, args...))
	}
	t.Cleanup(func() { mpvDebugf = prev })
	return func() []string {
		mu.Lock()
		defer mu.Unlock()
		return append([]string(nil), lines...)
	}
}

// The player asks mpv for the position, the duration, the pause state and the
// speed several times a second. Logging every answer wrote 70 lines in 15
// seconds, each followed by a blank one, and buried everything else.
func TestAnsweringAnMPVQueryWritesNothingToTheLog(t *testing.T) {
	logged := captureMPVLog(t)
	sock := fakeMPV(t, `{"data":496.538782,"request_id":0,"error":"success"}`+"\n")

	for range 20 {
		data, err := mpvSendCommand(sock, []any{"get_property", "time-pos"})
		require.NoError(t, err)
		assert.InDelta(t, 496.538782, data, 1e-6)
	}

	assert.Empty(t, logged(), "a routine answer is not worth a line")
}

// "property unavailable" is what mpv says before a file is loaded: expected,
// and not worth a line either.
func TestPropertyUnavailableIsNotLogged(t *testing.T) {
	logged := captureMPVLog(t)
	sock := fakeMPV(t, `{"request_id":0,"error":"property unavailable"}`+"\n")

	_, err := mpvSendCommand(sock, []any{"get_property", "duration"})

	require.Error(t, err)
	assert.Empty(t, logged())
}

// A real mpv error is the rare thing a log should keep, with the command that
// caused it. Before, it surfaced only as the generic "no data field".
func TestARealMPVErrorIsLoggedWithItsCommand(t *testing.T) {
	logged := captureMPVLog(t)
	sock := fakeMPV(t, `{"request_id":0,"error":"invalid parameter"}`+"\n")

	_, err := mpvSendCommand(sock, []any{"set_property", "speed", "banana"})

	require.Error(t, err)
	got := logged()
	require.Len(t, got, 1)
	assert.Contains(t, got[0], "invalid parameter")
	assert.Contains(t, got[0], "set_property")
	assert.False(t, strings.HasSuffix(got[0], "\n"), "no trailing newline, which is what left a blank line after each entry")
}
