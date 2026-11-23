package quic

import "github.com/enetx/quic-go/internal/protocol"

// InitialPaddingMode selects where client datagrams containing Initial packets
// carry their size padding. The mandatory header-protection sample padding stays
// encrypted in the packet in either mode.
type InitialPaddingMode uint8

const (
	// InitialPaddingPacket retains native encrypted PADDING frames.
	InitialPaddingPacket InitialPaddingMode = iota
	// InitialPaddingDatagram appends zeros after coalesced long-header packets.
	// When a short-header packet is last, its padding is encrypted inside it.
	InitialPaddingDatagram
)

func (p *packetPacker) padInitialDatagram(packet *coalescedPacket, maxSize protocol.ByteCount) {
	if p.initialPaddingMode != InitialPaddingDatagram || p.perspective != protocol.PerspectiveClient || packet.shortHdrPacket != nil {
		return
	}
	for _, part := range packet.longHdrPackets {
		if part.EncryptionLevel() != protocol.EncryptionInitial {
			continue
		}
		padding := maxSize - packet.buffer.Len()
		if padding > 0 {
			packet.buffer.Data = append(packet.buffer.Data, make([]byte, int(padding))...)
			packet.initialDatagramPadding = padding
		}
		return
	}
}

// A short-header packet extends to the datagram boundary. Any padding following
// it must be encrypted inside that packet, otherwise AEAD authentication fails.
func (p *packetPacker) initialShortHeaderPadding(hasInitial bool, size, maxSize protocol.ByteCount) protocol.ByteCount {
	if hasInitial && p.perspective == protocol.PerspectiveClient && p.initialPaddingMode == InitialPaddingDatagram {
		return max(0, maxSize-size)
	}
	return 0
}
