//go:build !unix

package diag

// FreeSpace is not collected on this platform.
func FreeSpace(string) (uint64, bool) { return 0, false }
