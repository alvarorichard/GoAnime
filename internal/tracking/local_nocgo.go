//go:build !cgo

package tracking

// isCgoEnabled returns false when compiled without CGO
func isCgoEnabled() bool {
	return false
}
