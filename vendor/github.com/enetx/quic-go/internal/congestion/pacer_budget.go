package congestion

import (
	"math"
	"time"

	"github.com/enetx/quic-go/internal/protocol"
)

// timeScaledBandwidth calculates the number of bytes that may be sent within
// a given duration, based on the current pacing rate.
// It caps the scaled value to the maximum allowed burst and handles overflows.
func (p *pacer) timeScaledBandwidth(d time.Duration, rate bandwidth) protocol.ByteCount {
	bw := uint64(rate / bytesPerSecond)
	if bw == 0 {
		return 0
	}
	const nsPerSecond = 1e9
	ns := uint64(d)
	maxBurst := maxBurstSizePackets * p.maxDatagramSize
	var scaled protocol.ByteCount
	if ns > math.MaxUint64/bw {
		scaled = maxBurst
	} else {
		scaled = protocol.ByteCount(bw * ns / nsPerSecond)
	}
	return scaled
}
