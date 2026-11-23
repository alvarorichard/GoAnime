package ackhandler

import (
	"fmt"
	"time"

	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/utils"
)

// AckReorderingMode selects when packet gaps prompt an immediate ACK.
type AckReorderingMode uint8

const (
	AckReorderingNative AckReorderingMode = iota
	// AckReorderingImmediate acknowledges every out-of-order ack-eliciting packet.
	AckReorderingImmediate
	// AckReorderingWindow acknowledges the first MissingPacketAckLimit packets
	// following a gap, and packets filling gaps already reported to the peer.
	AckReorderingWindow
)

type AckECNMode uint8

const (
	AckECNNative AckECNMode = iota
	AckECNTransitions
	AckECNCountOnly
)

// AckPolicy configures local Application Data acknowledgments. Nil preserves
// native behavior. It neither negotiates ACK_FREQUENCY nor changes peer ACKs.
// A configured policy keeps the earliest pending deadline rather than sliding it.
type AckPolicy struct {
	// FlushOnPacket gives a queued ACK a send opportunity before processing the
	// next received UDP datagram. False retains native receive-queue batching.
	// Socket backpressure can still postpone transmission.
	FlushOnPacket bool
	// PacketThreshold is the initial ack-eliciting packet threshold; zero uses native.
	PacketThreshold uint64
	// DecimationThreshold, when nonzero, replaces PacketThreshold after a packet
	// number reaches the smallest received number plus DecimationAfter. This is
	// a packet-number span, not a count of received or ack-eliciting packets.
	DecimationThreshold uint64
	DecimationAfter     uint64
	// RTTDivisor bounds delay by minRTT/divisor after the decimation span is
	// reached. Zero disables this bound. MinDelay supplies a timer floor.
	RTTDivisor            uint32
	MinDelay              time.Duration
	Reordering            AckReorderingMode
	MissingPacketAckLimit uint64
	ECN                   AckECNMode
}

// ShouldFlushACK reports a pending policy-selected ACK without consuming its
// ranges or resetting its timer. It never changes native receive batching.
func (h *ReceivedPacketHandler) ShouldFlushACK() bool {
	return h.appDataPackets.ackPolicy != nil &&
		h.appDataPackets.ackPolicy.FlushOnPacket &&
		h.appDataPackets.ackQueued && h.appDataPackets.hasNewAck
}

func (p *AckPolicy) Clone() *AckPolicy {
	if p == nil {
		return nil
	}
	copy := *p
	return &copy
}

func (p *AckPolicy) Validate(maxDelay time.Duration) error {
	if p == nil {
		return nil
	}
	if maxDelay == 0 {
		maxDelay = MaxAckDelay
	}
	const maxPacketNumber = 1<<62 - 1
	if p.PacketThreshold > maxPacketNumber || p.DecimationThreshold > maxPacketNumber || p.DecimationAfter > maxPacketNumber || p.MissingPacketAckLimit > maxPacketNumber {
		return fmt.Errorf("ACK policy exceeds packet number range")
	}
	if p.MinDelay < 0 || p.MinDelay > maxDelay {
		return fmt.Errorf("ACK minimum delay exceeds advertised maximum")
	}
	if p.Reordering > AckReorderingWindow || p.ECN > AckECNCountOnly {
		return fmt.Errorf("invalid ACK policy mode")
	}
	if p.Reordering == AckReorderingWindow && p.MissingPacketAckLimit == 0 {
		return fmt.Errorf("ACK missing packet window must be positive")
	}
	return nil
}

type ackPolicyState struct {
	first                       protocol.PacketNumber
	seen, decimated, previousCE bool
}

// SetAckPolicy snapshots receive policy before packets arrive. Initial and
// Handshake trackers are deliberately independent of this configuration.
func (h *ReceivedPacketHandler) SetAckPolicy(policy *AckPolicy, rtt *utils.RTTStats) {
	h.appDataPackets.ackPolicy = policy.Clone()
	h.appDataPackets.ackRTT = rtt
}

func (h *appDataReceivedPacketTracker) receivedPacketWithPolicy(pn protocol.PacketNumber, ecn protocol.ECN, now monotime.Time, eliciting bool) error {
	p, state := h.ackPolicy, &h.ackPolicyState
	next := protocol.PacketNumber(0)
	if state.seen {
		next = h.largestObserved + 1
	}
	if !state.seen || pn < state.first {
		state.first = pn
	}
	if !state.seen || pn >= h.largestObserved {
		h.largestObserved, h.largestObservedRcvdTime = pn, now
	}
	state.seen = true
	becameCE := ecn == protocol.ECNCE && !state.previousCE
	state.previousCE = ecn == protocol.ECNCE
	if p.Reordering == AckReorderingWindow && h.isMissing(pn) {
		h.hasNewAck = true
		h.ackQueued, h.ackAlarm = true, 0
		return nil
	}
	if p.ECN == AckECNTransitions && becameCE {
		h.hasNewAck = true
		h.ackQueued, h.ackAlarm = true, 0
		return nil
	}
	if !eliciting {
		return nil
	}
	h.ackElicitingPacketsReceivedSinceLastAck++
	spanReached := p.DecimationThreshold != 0 && uint64(pn-state.first) >= p.DecimationAfter
	state.decimated = state.decimated || spanReached
	threshold := p.PacketThreshold
	if threshold == 0 {
		threshold = packetsBeforeAck
	}
	if state.decimated {
		threshold = p.DecimationThreshold
	}
	immediate := uint64(h.ackElicitingPacketsReceivedSinceLastAck) >= threshold
	switch p.Reordering {
	case AckReorderingNative:
		immediate = immediate || h.isMissing(pn) || h.hasNewMissingPackets()
	case AckReorderingImmediate:
		immediate = immediate || pn != next
	case AckReorderingWindow:
		immediate = immediate || h.isMissing(pn) || h.withinMissingPacketWindow(p.MissingPacketAckLimit)
	}
	if p.ECN == AckECNNative && ecn == protocol.ECNCE {
		immediate = true
	}
	if immediate {
		h.ackQueued, h.ackAlarm = true, 0
	}
	if h.ackQueued {
		return nil
	}
	delay := h.maxAckDelay()
	if spanReached && p.RTTDivisor > 0 {
		var minRTT time.Duration
		if h.ackRTT != nil && h.ackRTT.HasMeasurement() {
			minRTT = h.ackRTT.MinRTT()
		}
		delay = max(p.MinDelay, min(delay, minRTT/time.Duration(p.RTTDivisor)))
	}
	deadline := now.Add(delay)
	if h.ackAlarm.IsZero() || deadline.Before(h.ackAlarm) {
		h.ackAlarm = deadline
	}
	return nil
}

func (h *appDataReceivedPacketTracker) withinMissingPacketWindow(limit uint64) bool {
	ranges := h.packetHistory.ranges
	if len(ranges) == 0 {
		return false
	}
	hasMissing := len(ranges) > 1 || h.ignoreBelow > 0 && ranges[0].Start > h.ignoreBelow
	last := ranges[len(ranges)-1]
	return hasMissing && uint64(last.End-last.Start+1) <= limit
}
