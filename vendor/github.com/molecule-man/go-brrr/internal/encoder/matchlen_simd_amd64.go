// SIMD match length for amd64 (SSE2).

//go:build amd64 && !purego

package encoder

import "unsafe"

//go:noescape
func matchLenSIMD(dataPtr unsafe.Pointer, a, b uint, limit int) int
