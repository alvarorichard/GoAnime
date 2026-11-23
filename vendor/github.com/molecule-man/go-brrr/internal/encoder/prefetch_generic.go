//go:build !amd64 || purego

package encoder

import "unsafe"

const hasPrefetch = false

func prefetch2(a, b unsafe.Pointer) {}
