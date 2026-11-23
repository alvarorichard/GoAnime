//go:build !amd64 || purego

package encoder

// prefix2Mask64Available reports that no SIMD kernel is compiled in.
const prefix2Mask64Available = false

// prefix2Mask64 is never called on this target; it exists so the dispatch
// compiles.
func prefix2Mask64(p *byte, c0, c1 byte) uint64 { return 0 }
