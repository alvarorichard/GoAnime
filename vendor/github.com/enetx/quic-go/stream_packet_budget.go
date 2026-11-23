package quic

import (
	"github.com/enetx/quic-go/internal/ackhandler"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/wire"
	"github.com/enetx/quic-go/quicvarint"
)

// The final STREAM omits its data-length field. Size the temporary explicit
// field from the data capacity after the actual ID and offset, not the packet
// budget: a 64-byte budget can hold less than 64 bytes of data and save only one
// length byte. The caller holds the stream lock while inspecting its offset.
func streamPacketSizeWithLength(
	budget protocol.ByteCount,
	frame *wire.StreamFrame,
	v protocol.Version,
) protocol.ByteCount {
	if !frame.DataLenPresent {
		return budget
	}
	header := wire.StreamFrame{StreamID: frame.StreamID, Offset: frame.Offset}
	return budget + protocol.ByteCount(quicvarint.Len(uint64(header.MaxDataLen(budget, v))))
}

func popStreamFrameForPacket(
	stream streamFrameGetter,
	budget protocol.ByteCount,
	v protocol.Version,
) (ackhandler.StreamFrame, *wire.StreamDataBlockedFrame, bool) {
	if native, ok := stream.(*SendStream); ok {
		return native.popStreamFrameWithLimit(budget, v, true)
	}
	// Preserve the existing getter contract for synthetic frame sources.
	return stream.popStreamFrame(budget+protocol.ByteCount(quicvarint.Len(uint64(budget))), v)
}

func popRetransmissionFrameForPacket(
	stream streamFrameGetter,
	budget protocol.ByteCount,
	v protocol.Version,
) (ackhandler.StreamFrame, bool) {
	if native, ok := stream.(*SendStream); ok {
		return native.popRetransmissionFrameWithLimit(budget, v, true)
	}
	return stream.popRetransmissionFrame(budget+protocol.ByteCount(quicvarint.Len(uint64(budget))), v)
}
