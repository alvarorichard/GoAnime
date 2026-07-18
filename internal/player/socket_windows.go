//go:build windows

// Windows-specific file that implements the MPV socket connection
// using named pipes instead of Unix domain sockets.

package player

import (
	"net"
	"path/filepath"
	"strings"
	"time"

	"github.com/Microsoft/go-winio"
)

// dialMPVSocket opens a connection to the MPV socket on Windows.
// On Windows, named pipes are used in the format \\\\.\\pipe\\PIPENAME.
// The go-winio package is used only on Windows for named pipe support.
func dialMPVSocket(socketPath string) (net.Conn, error) {
	// Windows uses named pipes format
	// Named pipes in Windows need to be in the format \\.\pipe\PIPENAME
	if !strings.HasPrefix(socketPath, `\\.\pipe\`) {
		socketPath = `\\.\pipe\` + filepath.Base(socketPath)
	}

	// Short probe timeout: ERROR_FILE_NOT_FOUND returns immediately; the
	// timeout only matters for ERROR_PIPE_BUSY. Keep it low so StartVideo can
	// poll process-exit between dial attempts instead of blocking 5s.
	timeout := 200 * time.Millisecond
	return winio.DialPipe(socketPath, &timeout)
}
