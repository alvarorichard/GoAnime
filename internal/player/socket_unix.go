//go:build !windows

// Unix-specific file (Linux, macOS) that implements the MPV socket
// connection using standard Unix domain sockets.

package player

import (
	"net"
)

// dialMPVSocket opens a connection to the MPV socket on Unix systems.
// On Unix-like systems (Linux, macOS), standard Unix domain sockets are
// used via net.Dial with the "unix" network type.
func dialMPVSocket(socketPath string) (net.Conn, error) {
	// Unix-like system uses Unix sockets
	return net.Dial("unix", socketPath)
}
