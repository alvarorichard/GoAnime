//go:build !windows

package tui

import (
	"math"
	"os"
	"time"

	"golang.org/x/sys/unix"
	"golang.org/x/term"
)

// DrainTerminalResponses reads and discards pending bytes on stdin using
// non-blocking I/O. While waiting, stdin is temporarily put in raw/no-echo mode
// so delayed terminal responses do not get printed as literal escape text.
func DrainTerminalResponses(wait time.Duration) {
	rawFd := os.Stdin.Fd()
	if rawFd > math.MaxInt {
		return
	}
	fd := int(rawFd) //nolint:gosec // overflow guarded above

	if !term.IsTerminal(fd) {
		return
	}

	state, rawErr := term.MakeRaw(fd)
	if rawErr == nil {
		defer func() { _ = term.Restore(fd, state) }()
	}
	if wait > 0 {
		time.Sleep(wait)
	}

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
