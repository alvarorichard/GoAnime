//go:build amd64 && !purego

package encoder

// findBlocksDP runs the forward pass of findBlocks: per byte it adds the
// symbol's insertCost row into cost, stores the first index holding the minimum
// as the byte's block type, rebases cost against the minimum and clamps it at
// the switch cost.
//
// First-occurrence tie-breaking is load-bearing: resolving a tie differently
// changes the compressed output.
//
// Implemented in findblocks_kernel_simd_amd64.s. SSE2 is baseline on amd64, so
// this needs no CPU feature check.
//
//go:noescape
func findBlocksDP(data []uint16, insertCost, cost []float64, switchSignal, blockID []byte, blockSwitchBitcost float64)
