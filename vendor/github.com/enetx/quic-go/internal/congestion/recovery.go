package congestion

import (
	"math/bits"

	"github.com/enetx/quic-go/internal/protocol"
)

// prr implements RFC 6937's proportional reduction and PRR-SSRB allowance.
// The accounting segment is supplied by the application, independently of MTU.
type prr struct {
	sent, delivered, priorInFlight protocol.ByteCount
	acks                           uint64
}

func (p *prr) lost(prior protocol.ByteCount) { *p = prr{priorInFlight: prior} }
func (p *prr) acked(n protocol.ByteCount)    { p.delivered += n; p.acks++ }

func (p *prr) canSend(cwnd, inFlight, threshold, segment protocol.ByteCount) bool {
	if p.sent == 0 || inFlight < segment {
		return true
	}
	if cwnd > inFlight {
		return uint64(p.delivered)+p.acks*uint64(segment) > uint64(p.sent)
	}
	// Compare products in 128 bits: large windows must not reverse the test.
	ah, al := bits.Mul64(uint64(p.delivered), uint64(threshold))
	bh, bl := bits.Mul64(uint64(p.sent), uint64(p.priorInFlight))
	return ah > bh || ah == bh && al > bl
}

// OnPersistentCongestion applies a verified persistent-loss collapse. A PTO alone
// never invokes this operation. Native Reno shares its existing state and helpers.
func OnPersistentCongestion(sender SendAlgorithm) {
	switch controller := sender.(type) {
	case *configuredSender:
		controller.OnPersistentCongestion()
	case *renoSender:
		controller.hybridSlowStart.Restart()
		controller.undoState.Reset()
		controller.numAckedPackets = 0
		controller.congestionWindow = controller.minCongestionWindow()
		controller.largestSentAtLastCutback = controller.largestSentPacketNumber
	}
}

func (c *configuredSender) OnPersistentCongestion() {
	c.hybridSlowStart.Restart()
	c.cubic.Reset()
	if c.rfcCubic != nil {
		*c.rfcCubic = rfcCubic{}
	}
	c.undoState.Reset()
	c.numAckedPackets = 0
	c.congestionWindow = c.minCongestionWindow()
	// Retain the boundary: late ACKs of pre-collapse sends cannot regrow cwnd.
	c.largestSentAtLastCutback = c.largestSentPacketNumber
	c.recoveryAllowance = 0
	c.connStats.CongestionWindow.Store(uint64(c.congestionWindow))
}
