//go:build unix

package diag

import "syscall"

// FreeSpace reports the bytes available to an unprivileged user on the
// filesystem holding path.
func FreeSpace(path string) (uint64, bool) {
	var st syscall.Statfs_t
	if err := syscall.Statfs(path, &st); err != nil {
		return 0, false
	}
	return uint64(st.Bavail) * uint64(st.Bsize), true // #nosec G115 -- block counts and sizes are non-negative
}
