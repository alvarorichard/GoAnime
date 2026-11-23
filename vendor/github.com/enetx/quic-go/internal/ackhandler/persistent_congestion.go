package ackhandler

import (
	"time"

	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
)

// persistentLossRange follows RFC 9002 section 7.6: loss of a long enough
// contiguous range, after an RTT sample, bounded by ack-eliciting packets.
// A PTO expiry, a path probe, or one isolated lost packet cannot trigger it.
type persistentLossRange struct {
	last   protocol.PacketNumber
	start  monotime.Time
	cutoff monotime.Time
	period time.Duration
	seen   bool
}

func (r *persistentLossRange) lost(pn protocol.PacketNumber, sent monotime.Time, eliciting bool) bool {
	if r.cutoff.IsZero() || sent < r.cutoff {
		return false
	}
	if !r.seen || pn != r.last+1 {
		r.start = 0
	}
	r.last, r.seen = pn, true
	if !eliciting {
		return false
	}
	if r.start.IsZero() {
		r.start = sent
		return false
	}
	return sent.Sub(r.start) > r.period
}
