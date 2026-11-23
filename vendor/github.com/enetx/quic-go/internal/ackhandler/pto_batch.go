package ackhandler

import (
	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/wire"
	"github.com/enetx/quic-go/qlog"
)

type ptoProbeBatch struct {
	force, retransmit [3]bool
	cause             protocol.PacketNumberSpace
	trigger           protocol.PacketNumberSpace
	firedAt           monotime.Time
	total             uint8
	queued            bool
}

func probeLevel(space protocol.PacketNumberSpace) protocol.EncryptionLevel {
	switch space {
	case protocol.PacketNumberSpaceInitial:
		return protocol.EncryptionInitial
	case protocol.PacketNumberSpaceHandshake:
		return protocol.EncryptionHandshake
	default:
		return protocol.Encryption1RTT
	}
}

func (h *sentPacketHandler) cryptoPTOTime(now monotime.Time) (deadline monotime.Time, level protocol.EncryptionLevel) {
	spaces := []protocol.PacketNumberSpace{protocol.PacketNumberSpaceInitial, protocol.PacketNumberSpaceHandshake}
	if h.handshakeConfirmed {
		spaces = []protocol.PacketNumberSpace{protocol.PacketNumberSpaceAppData}
	}
	for _, space := range spaces {
		ps := h.getPacketNumberSpace(space)
		if ps == nil || ps.lastAckElicitingPacketTime.IsZero() || space == protocol.PacketNumberSpaceAppData && ps.numOutstanding == 0 {
			continue
		}
		next := ps.lastAckElicitingPacketTime.Add(h.getScaledPTO(h.handshakeConfirmed))
		if deadline.IsZero() || next.Before(deadline) {
			deadline, level = next, probeLevel(space)
		}
	}
	if deadline.IsZero() && !h.peerCompletedAddressValidation {
		level = protocol.EncryptionInitial
		if h.initialPackets == nil {
			level = protocol.EncryptionHandshake
		}
		deadline = now.Add(h.getScaledPTO(false))
	}
	return deadline, level
}

func (h *sentPacketHandler) onCoalescedPTOTimeout(now monotime.Time) error {
	if h.handshakeConfirmed {
		h.detectLostPathProbes(now)
	}
	for _, space := range []protocol.PacketNumberSpace{protocol.PacketNumberSpaceInitial, protocol.PacketNumberSpaceHandshake, protocol.PacketNumberSpaceAppData} {
		ps := h.getPacketNumberSpace(space)
		if ps != nil && !ps.lossTime.IsZero() && !ps.lossTime.After(now) {
			h.detectLostPackets(now, probeLevel(space))
		}
	}
	var allowed, due [3]bool
	cause := protocol.PacketNumberSpaceAppData
	fired := false
	period := h.getScaledPTO(h.handshakeConfirmed)
	for _, space := range []protocol.PacketNumberSpace{protocol.PacketNumberSpaceInitial, protocol.PacketNumberSpaceHandshake, protocol.PacketNumberSpaceAppData} {
		ps := h.getPacketNumberSpace(space)
		if ps == nil || ps.lastAckElicitingPacketTime.IsZero() || space == protocol.PacketNumberSpaceAppData && ps.numOutstanding == 0 {
			continue
		}
		allowed[space] = true
		if !ps.lastAckElicitingPacketTime.Add(period).After(now) {
			due[space] = true
			if !fired {
				cause = space
				fired = true
			}
		}
	}
	if !fired && h.bytesInFlight == 0 && !h.peerCompletedAddressValidation {
		cause = protocol.PacketNumberSpaceInitial
		if h.initialPackets == nil {
			cause = protocol.PacketNumberSpaceHandshake
		}
		allowed[cause], due[cause], fired = true, true, true
	}
	if !fired {
		return nil
	}
	// Retransmission selection uses the old backoff for every space before this
	// single increment. A Handshake timeout also releases Initial CRYPTO tokens.
	if due[protocol.PacketNumberSpaceHandshake] && h.initialPackets != nil {
		due[protocol.PacketNumberSpaceInitial] = true
	}
	count := h.applicationProbePackets
	switch cause {
	case protocol.PacketNumberSpaceInitial:
		count = h.initialProbePackets
	case protocol.PacketNumberSpaceHandshake:
		count = h.handshakeProbePackets
	}
	if count == 0 {
		count = 2
	}
	trigger := cause
	if old := h.probeBatch; old != nil {
		cause = min(cause, old.cause)
		for i := range allowed {
			allowed[i] = allowed[i] || old.force[i]
		}
	}
	h.ptoCount++
	h.probeBatch = &ptoProbeBatch{force: allowed, retransmit: due, cause: cause, trigger: trigger, firedAt: now, total: count}
	h.numProbesToSend = int(count)
	switch cause {
	case protocol.PacketNumberSpaceInitial:
		h.ptoMode = SendPTOInitial
	case protocol.PacketNumberSpaceHandshake:
		h.ptoMode = SendPTOHandshake
	default:
		h.ptoMode = SendPTOAppData
	}
	for _, space := range []protocol.PacketNumberSpace{protocol.PacketNumberSpaceInitial, protocol.PacketNumberSpaceHandshake, protocol.PacketNumberSpaceAppData} {
		if !allowed[space] {
			continue
		}
		skip := h.applicationProbeSkipPacketNumber
		switch space {
		case protocol.PacketNumberSpaceInitial:
			skip = h.initialProbeSkipPacketNumber
		case protocol.PacketNumberSpaceHandshake:
			skip = h.handshakeProbeSkipPacketNumber
		}
		if skip {
			pn := h.PopPacketNumber(space)
			h.history(space).SkippedPacket(pn)
		}
	}
	if h.qlogger != nil {
		h.qlogger.RecordEvent(qlog.TimerUpdated{Type: qlog.TimerEventTypeExpired, TimerType: qlog.TimerTypePTO, EncLevel: probeLevel(trigger)})
		h.qlogger.RecordEvent(qlog.PTOCountUpdated{PTOCount: h.ptoCount})
	}
	return nil
}

// BeginPTOProbeDatagram consumes one shared datagram allowance, independently of
// how many protected packets it carries. Only eligible spaces force PING; all
// spaces may still contribute data. This does not extend SentPacketHandler.
func BeginPTOProbeDatagram(handler SentPacketHandler, maxSize protocol.ByteCount) (force [3]bool, ok bool) {
	h, ok := handler.(*sentPacketHandler)
	if !ok || !h.coalescePTOProbes || h.probeBatch == nil || h.numProbesToSend <= 0 {
		return force, false
	}
	batch := h.probeBatch
	if !batch.queued {
		batch.queued = true
		budget := maxSize * protocol.ByteCount(batch.total)
		var used protocol.ByteCount
		for _, space := range []protocol.PacketNumberSpace{protocol.PacketNumberSpaceInitial, protocol.PacketNumberSpaceHandshake, protocol.PacketNumberSpaceAppData} {
			if !batch.retransmit[space] {
				continue
			}
			for _, p := range h.history(space).Packets() {
				if p.EncryptionLevel.PacketNumberSpace() != space || p.probeQueued {
					continue
				}
				// The source iterator releases its token before checking the shared
				// byte limit, including the first packet beyond each space's limit.
				p.probeQueued = true
				used += p.Length
				if used > budget {
					break
				}
				retainCrypto := space == protocol.PacketNumberSpaceInitial && h.retainInitialCryptoProbes || space == protocol.PacketNumberSpaceHandshake && h.retainHandshakeCryptoProbes
				for i, f := range p.Frames {
					if f.Handler != nil {
						f.Handler.OnLost(f.Frame)
						if !retainCrypto {
							p.Frames[i].Handler = nil
						}
					}
				}
				h.queueStreamProbeFrames(p)
			}
		}
	}
	h.numProbesToSend--
	return batch.force, true
}

// PrimePTOHandshake establishes a missing Handshake timer baseline only after an
// Initial PTO, with real Handshake TX keys and a received Initial ACK.
func PrimePTOHandshake(handler SentPacketHandler, keysAvailable bool, now monotime.Time) {
	h, ok := handler.(*sentPacketHandler)
	if !ok || !h.coalescePTOProbes || !keysAvailable || h.probeBatch == nil || h.probeBatch.trigger != protocol.PacketNumberSpaceInitial || h.probeBatch.firedAt != now || h.initialPackets == nil || h.initialPackets.largestAcked == protocol.InvalidPacketNumber {
		return
	}
	ps := h.handshakePackets
	if ps != nil && ps.lastAckElicitingPacketTime.IsZero() && ps.largestAcked == protocol.InvalidPacketNumber {
		ps.lastAckElicitingPacketTime = now
		h.setLossDetectionTimer(now)
	}
}

// queueStreamProbeFrames preserves immutable range callbacks where supported.
func (h *sentPacketHandler) queueStreamProbeFrames(p *packet) bool {
	copied := false
	for i, f := range p.StreamFrames {
		if probe, ok := f.Handler.(interface{ OnProbe(wire.Frame) bool }); ok {
			copied = probe.OnProbe(f.Frame) || copied
		} else if f.Handler != nil {
			f.Handler.OnLost(f.Frame)
			p.StreamFrames[i].Handler = nil
			copied = true
		}
	}
	return copied
}
