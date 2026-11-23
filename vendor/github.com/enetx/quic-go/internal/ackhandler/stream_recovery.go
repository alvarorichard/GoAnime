package ackhandler

import (
	"slices"
	"time"

	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/wire"
)

// StreamLossInfo describes retained ACK metadata for an Application STREAM packet.
// It contains no payload or application state. Packet numbers use -1 when unknown.
type StreamLossInfo struct {
	// TimeSinceLoss is elapsed time since actual loss declaration, excluding PTO copies.
	TimeSinceLoss time.Duration
	// ProbeTimeout includes current exponential backoff and handshake ACK-delay rules.
	ProbeTimeout time.Duration
	// FirstPacketAfterLoss is the next packet-number boundary at loss declaration.
	// Shared numbering includes subsequent packets from other encryption levels.
	FirstPacketAfterLoss int64
	// LargestAcknowledged is the largest ACK in the applicable numbering domain.
	LargestAcknowledged int64
}

type retainedStreamPacket struct {
	number         protocol.PacketNumber
	level          protocol.EncryptionLevel
	lossTime       monotime.Time
	firstAfterLoss protocol.PacketNumber
	handlers       []FrameHandler
}

func (h *sentPacketHandler) retainStreamPacket(pn protocol.PacketNumber, p *packet, now monotime.Time) {
	if !h.streamDataRecovery || p.EncryptionLevel.PacketNumberSpace() != protocol.PacketNumberSpaceAppData {
		return
	}
	var handlers []FrameHandler
	for _, f := range p.StreamFrames {
		if handler, ok := f.Handler.(interface{ HasOutstandingData() bool }); ok && handler.HasOutstandingData() {
			handlers = append(handlers, f.Handler)
		}
	}
	if len(handlers) == 0 {
		return
	}
	// The generic limit is the existing packet-history limit, not browser policy.
	// Profiles normally expire metadata much earlier through StreamLossRetention.
	if len(h.retainedStreamPackets) >= protocol.MaxTrackedSentPackets {
		h.retainedStreamPackets = slices.Delete(h.retainedStreamPackets, 0, 1)
	}
	h.retainedStreamPackets = append(h.retainedStreamPackets, retainedStreamPacket{number: pn, level: p.EncryptionLevel, lossTime: now, firstAfterLoss: h.packetNumbers(protocol.PacketNumberSpaceAppData).Peek(false), handlers: handlers})
}

func (h *sentPacketHandler) acknowledgeRetainedStreams(ack *wire.AckFrame, level protocol.EncryptionLevel) (matched, oneRTT bool) {
	if level.PacketNumberSpace() != protocol.PacketNumberSpaceAppData {
		return false, false
	}
	h.retainedStreamPackets = slices.DeleteFunc(h.retainedStreamPackets, func(p retainedStreamPacket) bool {
		if !ack.AcksPacket(p.number) {
			return false
		}
		matched = true
		h.earlyDataAcked = true
		oneRTT = oneRTT || p.level == protocol.Encryption1RTT
		for _, handler := range p.handlers {
			handler.OnAcked(nil)
		}
		return true
	})
	return matched, oneRTT
}

func (h *sentPacketHandler) expireRetainedStreams(now monotime.Time) {
	if len(h.retainedStreamPackets) == 0 {
		return
	}
	h.retainedStreamPackets = slices.DeleteFunc(h.retainedStreamPackets, func(p retainedStreamPacket) bool {
		outstanding := false
		for _, handler := range p.handlers {
			if handler.(interface{ HasOutstandingData() bool }).HasOutstandingData() {
				outstanding = true
				break
			}
		}
		if !outstanding {
			return true
		}
		return h.streamLossRetention != nil && !h.streamLossRetention(StreamLossInfo{TimeSinceLoss: now.Sub(p.lossTime), ProbeTimeout: h.getScaledPTO(h.handshakeConfirmed), FirstPacketAfterLoss: int64(p.firstAfterLoss), LargestAcknowledged: int64(h.streamLargestAcked)})
	})
}

func (h *sentPacketHandler) queueStreamProbe(space protocol.PacketNumberSpace) bool {
	for _, p := range h.history(space).Packets() {
		if !p.Outstanding() || p.EncryptionLevel.PacketNumberSpace() != space || p.probeQueued {
			continue
		}
		p.probeQueued = true
		copied := false
		for i, f := range p.Frames {
			if f.Handler != nil {
				f.Handler.OnLost(f.Frame)
				p.Frames[i].Handler = nil
				copied = true
			}
		}
		copied = h.queueStreamProbeFrames(p) || copied
		if copied {
			return true
		}
	}
	return false
}
