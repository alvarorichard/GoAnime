//go:build !amd64 || purego

package encoder

// histogramTotalCount sums all entries in a histogram.
func histogramTotalCount(h []uint32, alphabetSize int) uint32 {
	var total uint32
	for i := range alphabetSize {
		total += h[i]
	}
	return total
}
