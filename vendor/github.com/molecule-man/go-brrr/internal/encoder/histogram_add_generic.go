//go:build !amd64 || purego

package encoder

// histogramAdd adds src histogram into dst histogram element-wise.
func histogramAdd(dst, src []uint32, alphabetSize int) {
	for i := range alphabetSize {
		dst[i] += src[i]
	}
}
