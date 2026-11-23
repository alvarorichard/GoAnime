package ackhandler

import (
	"math"
	"time"

	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/qerr"
	"github.com/enetx/quic-go/internal/wire"
)

type legacyAckFrequencyState struct {
	minimum, delay      time.Duration
	sequence, threshold uint64
	active, ignoreOrder bool
}

// SetLegacyAckFrequency enables only the negotiated legacy receive format.
// Initial/Handshake ACKs and default scheduling before an update stay unchanged.
func (h *ReceivedPacketHandler) SetLegacyAckFrequency(minimum time.Duration) {
	if minimum > 0 {
		h.appDataPackets.legacyAckFrequency = &legacyAckFrequencyState{minimum: minimum}
	}
}

func (h *ReceivedPacketHandler) HandleLegacyAckFrequency(f *wire.LegacyAckFrequencyFrame) error {
	s := h.appDataPackets.legacyAckFrequency
	if s == nil || f.RequestMaxAckDelay < s.minimum {
		return &qerr.TransportError{ErrorCode: qerr.ProtocolViolation, FrameType: uint64(wire.FrameTypeAckFrequency), ErrorMessage: "legacy ACK_FREQUENCY delay below advertised minimum or extension not enabled"}
	}
	if f.AckElicitingThreshold == 0 {
		return &qerr.TransportError{ErrorCode: qerr.FrameEncodingError, FrameType: uint64(wire.FrameTypeAckFrequency), ErrorMessage: "zero legacy ACK_FREQUENCY tolerance"}
	}
	// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/third_party/rust/neqo-transport/src/tracking.rs
	// Neqo accepts equal sequence numbers; the later draft's strictly newer
	// sequence rule must not be applied to this format.
	if f.SequenceNumber >= s.sequence {
		s.sequence, s.threshold, s.delay, s.ignoreOrder, s.active = f.SequenceNumber, f.AckElicitingThreshold, f.RequestMaxAckDelay, f.IgnoreOrder, true
	}
	return nil
}

func (h *appDataReceivedPacketTracker) receivedPacketWithLegacyAckFrequency(pn protocol.PacketNumber, now monotime.Time, eliciting, seen bool) error {
	s := h.legacyAckFrequency
	next := protocol.PacketNumber(0)
	if seen {
		next = h.largestObserved + 1
	}
	if !seen || pn >= h.largestObserved {
		h.largestObserved, h.largestObservedRcvdTime = pn, now
	}
	if !eliciting {
		return nil
	}
	h.ackElicitingPacketsReceivedSinceLastAck++
	if !s.ignoreOrder && pn != next || uint64(h.ackElicitingPacketsReceivedSinceLastAck) >= s.threshold {
		h.ackQueued, h.ackAlarm = true, 0
	}
	// A policy update never moves a pending timer, matching Neqo tracking.rs.
	if !h.ackQueued && h.ackAlarm.IsZero() {
		if now >= 0 && s.delay > time.Duration(math.MaxInt64-int64(now)) {
			h.ackAlarm = monotime.Time(math.MaxInt64)
		} else {
			h.ackAlarm = now.Add(s.delay)
		}
	}
	return nil
}
