//go:build amd64 && !purego

package encoder

// prefix2Mask64Available reports that the SSE2 kernel is compiled in.
const prefix2Mask64Available = true

// prefix2Mask64 returns a bitmask whose bit j is set when p[j] == c0 and
// p[j+1] == c1. It reads 65 bytes starting at p.
//
//go:noescape
func prefix2Mask64(p *byte, c0, c1 byte) uint64
