package quic

import (
	"github.com/enetx/quic-go/internal/ackhandler"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/wire"
)

type PacketNumberLengthInfo = ackhandler.PacketNumberLengthInfo
type AckProbeInfo = ackhandler.AckProbeInfo

// Keep the policy's temporary estimate separate from the path MTU. A remainder
// that cannot hold another protected header is not a usable packet-size limit.
func (p *packetPacker) recordCoalescedPacketSize(hdr *wire.ExtendedHeader, overhead, remaining protocol.ByteCount, v protocol.Version) {
	if p.coalescedPacketSize == nil {
		return
	}
	space := protocol.PacketNumberSpaceAppData
	if hdr != nil {
		switch hdr.Type {
		case protocol.PacketTypeInitial:
			space = protocol.PacketNumberSpaceInitial
		case protocol.PacketTypeHandshake:
			space = protocol.PacketNumberSpaceHandshake
		}
		minimum := hdr.GetLength(v) + overhead + max(0, 4-protocol.ByteCount(hdr.PacketNumberLen))
		if remaining < minimum {
			remaining = 0
		}
	} else {
		remaining = 0
	}
	p.coalescedPacketSize(space, remaining)
}
