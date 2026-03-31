//go:build !windows

package tui

import (
	"math"
	"os"

	"golang.org/x/sys/unix"
)

// drainStdin reads and discards any pending bytes on stdin using
// non-blocking I/O so it never blocks if there is nothing to read.
// This clears stale escape sequences left by tcell's application cursor
// key mode before the next interactive prompt runs.
func drainStdin() {
	rawFd := os.Stdin.Fd()
	if rawFd > math.MaxInt {
		return
	}
	fd := int(rawFd) //nolint:gosec // overflow guarded above

	// Set non-blocking mode on stdin
	if err := unix.SetNonblock(fd, true); err != nil {
		return
	}
	// Restore blocking mode when done — Go's runtime expects stdin to be blocking
	defer func() { _ = unix.SetNonblock(fd, false) }()

	// Read and discard any pending bytes
	buf := make([]byte, 256)
	for {
		n, err := unix.Read(fd, buf)
		if n <= 0 || err != nil {
			break
		}
	}
}
