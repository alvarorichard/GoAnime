package congestion

import (
	"time"

	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
)

const maxBurstSizePackets = 10

// minPacingDelay is the minimum duration that is used for packet pacing
// If the packet packing frequency is higher, multiple packets might be sent at once.
// Example: For a packet pacing delay of 200μs, we would send 5 packets at once, wait for 1ms, and so forth.
const minPacingDelay = time.Millisecond

// The pacer implements a token bucket pacing algorithm.
type pacer struct {
	budgetAtLastSent protocol.ByteCount
	maxDatagramSize  protocol.ByteCount
	lastSentTime     monotime.Time
}

func newPacer(maxDatagramSize protocol.ByteCount) *pacer {
	return &pacer{
		maxDatagramSize: maxDatagramSize,
	}
}

func (p *pacer) SentPacket(sendTime monotime.Time, size protocol.ByteCount, rate bandwidth) {
	budget := p.Budget(sendTime, rate)
	if size >= budget {
		p.budgetAtLastSent = 0
	} else {
		p.budgetAtLastSent = budget - size
	}
	p.lastSentTime = sendTime
}

func (p *pacer) Budget(now monotime.Time, rate bandwidth) protocol.ByteCount {
	if p.lastSentTime.IsZero() {
		return p.maxBurstSize(rate)
	}
	delta := now.Sub(p.lastSentTime)
	var added protocol.ByteCount
	if delta > 0 {
		added = p.timeScaledBandwidth(delta, rate)
	}
	budget := p.budgetAtLastSent + added
	if added > 0 && budget < p.budgetAtLastSent {
		budget = protocol.MaxByteCount
	}
	return min(p.maxBurstSize(rate), budget)
}

func (p *pacer) maxBurstSize(rate bandwidth) protocol.ByteCount {
	return max(
		p.timeScaledBandwidth(minPacingDelay+protocol.TimerGranularity, rate),
		maxBurstSizePackets*p.maxDatagramSize,
	)
}

// TimeUntilSend returns when the next packet should be sent.
// It returns zero if a packet can be sent immediately.
func (p *pacer) TimeUntilSend(rate bandwidth) monotime.Time {
	if p.lastSentTime.IsZero() || p.budgetAtLastSent >= p.maxDatagramSize {
		return 0
	}
	diff := 1e9 * uint64(p.maxDatagramSize-p.budgetAtLastSent)
	bw := uint64(rate / bytesPerSecond)
	// We might need to round up this value.
	// Otherwise, we might have a budget (slightly) smaller than the datagram size when the timer expires.
	d := diff / bw
	// this is effectively a math.Ceil, but using only integer math
	if diff%bw > 0 {
		d++
	}
	return p.lastSentTime.Add(max(minPacingDelay, time.Duration(d)*time.Nanosecond))
}

func (p *pacer) SetMaxDatagramSize(s protocol.ByteCount) {
	p.maxDatagramSize = s
}
