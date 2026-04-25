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

	// Use winio to connect to Windows named pipe
	timeout := 5 * time.Second
	return winio.DialPipe(socketPath, &timeout)
}
