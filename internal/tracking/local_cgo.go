//go:build cgo

package tracking

// isCgoEnabled returns true when compiled with CGO
func isCgoEnabled() bool {
	return true
}
