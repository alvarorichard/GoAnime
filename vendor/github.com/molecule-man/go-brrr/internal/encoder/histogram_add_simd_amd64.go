//go:build amd64 && !purego

package encoder

// histogramAdd adds src histogram into dst histogram element-wise.
func histogramAdd(dst, src []uint32, alphabetSize int) {
	histogramAddAsm(dst, src, alphabetSize)
}

// histogramAddAsm adds src into dst thirty-two uint32 per iteration
// across eight XMM register pairs. SSE2 is baseline on amd64, so this needs no
// CPU feature check.
//
//go:noescape
func histogramAddAsm(dst, src []uint32, n int)
