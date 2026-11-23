package congestion

import (
	"time"

	"github.com/enetx/quic-go/internal/protocol"
)

// bandwidth of a connection
type bandwidth uint64

const (
	// bitsPerSecond is 1 bit per second
	bitsPerSecond bandwidth = 1
	// bytesPerSecond is 1 byte per second
	bytesPerSecond = 8 * bitsPerSecond
)

// bandwidthFromDelta calculates the bandwidth from a number of bytes and a time delta
func bandwidthFromDelta(bytes protocol.ByteCount, delta time.Duration) bandwidth {
	return saturatedBandwidthFromDelta(bytes, delta)
}
