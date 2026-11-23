package congestion

import (
	"math"
	"math/bits"
	"time"

	"github.com/enetx/quic-go/internal/protocol"
)

// saturatedBandwidthFromDelta retains whole-byte rounding without large-window overflow.
func saturatedBandwidthFromDelta(bytes protocol.ByteCount, delta time.Duration) bandwidth {
	if bytes <= 0 || delta <= 0 {
		return 0
	}
	// Preserve byte-per-second rounding, without overflowing at large windows.
	hi, lo := bits.Mul64(uint64(bytes), uint64(time.Second))
	if hi >= uint64(delta) {
		return bandwidth(math.MaxUint64 / uint64(bytesPerSecond) * uint64(bytesPerSecond))
	}
	q, _ := bits.Div64(hi, lo, uint64(delta))
	if q > math.MaxUint64/uint64(bytesPerSecond) {
		q = math.MaxUint64 / uint64(bytesPerSecond)
	}
	return bandwidth(q) * bytesPerSecond
}
