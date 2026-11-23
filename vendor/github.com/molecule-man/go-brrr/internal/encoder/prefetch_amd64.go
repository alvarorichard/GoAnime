//go:build amd64 && !purego

package encoder

import "unsafe"

const hasPrefetch = true

//go:noescape
func prefetch2(a, b unsafe.Pointer)
