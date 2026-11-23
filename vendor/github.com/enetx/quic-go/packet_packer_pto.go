package quic

import (
	"github.com/enetx/quic-go/internal/ackhandler"
	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/wire"
)

func (p *packetPacker) shouldForcePTOProbe(space protocol.PacketNumberSpace) bool {
	return p.ptoProbeForce != nil && p.ptoProbeForce[space]
}

// forcePTOProbe adds an eliciting frame even if the only existing payload is an
// ACK. It reserves a byte by truncating the ACK when necessary; ordinary packet
// composition and explicit ACK-only paths do not call this with active policy.
func (p *packetPacker) forcePTOProbe(pl payload, space protocol.PacketNumberSpace, budget protocol.ByteCount, v protocol.Version) payload {
	if !p.shouldForcePTOProbe(space) || ackhandler.HasAckElicitingFrames(pl.frames) || len(pl.streamFrames) > 0 || budget < 1 {
		return pl
	}
	if pl.ack != nil && pl.length >= budget {
		old := pl.ack.Length(v)
		pl.ack.Truncate(budget-1, v)
		pl.length += pl.ack.Length(v) - old
	}
	if pl.length < budget {
		pl.frames = append(pl.frames, ackhandler.Frame{Frame: &wire.PingFrame{}, Handler: emptyHandler{}})
		pl.length++
		p.numNonAckElicitingAcks = 0
	}
	return pl
}

func (p *packetPacker) packPTOProbeDatagram(force [3]bool, maxSize protocol.ByteCount, now monotime.Time, v protocol.Version) (*coalescedPacket, error) {
	p.ptoProbeForce = &force
	defer func() { p.ptoProbeForce = nil }()
	return p.packCoalescedPacket(false, maxSize, now, v, false)
}
