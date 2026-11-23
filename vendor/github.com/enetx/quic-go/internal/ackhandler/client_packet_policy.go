package ackhandler

import (
	"time"

	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
)

// PacketNumberLengthInfo is a send-loop snapshot for choosing 1..4 encoded
// bytes. LargestAcknowledged is -1 before the first ACK in this number space.
// LeastUnacknowledged includes tracked ACK-only packets. MaxPacketsInFlight
// estimates the congestion window in full-size datagrams.
type PacketNumberLengthInfo struct {
	PacketNumber        uint64
	LargestAcknowledged int64
	LeastUnacknowledged uint64
	MaxPacketsInFlight  uint64
	// CoalescedMaxPacketsInFlight estimates the same window using the remaining
	// unpadded datagram budget after the previous long-header packet in this
	// number space. Zero means that no usable coalescing budget is available.
	CoalescedMaxPacketsInFlight uint64
	NativeLength                uint8
}

// SetCoalescedPacketSize records the remaining long-header datagram budget for
// an optional client packet-number policy. It does not change congestion state,
// the real path MTU, recovery, or the native packet-number length selection.
func SetCoalescedPacketSize(handler SentPacketHandler, space protocol.PacketNumberSpace, size protocol.ByteCount) {
	if h, ok := handler.(*sentPacketHandler); ok && h.perspective == protocol.PerspectiveClient {
		h.coalescedPacketSpace = space
		h.coalescedPacketSize = max(0, size)
	}
}

// AckProbeInfo describes tracked application-data packets, including ACK-only
// packets, and time since the last ACK-eliciting send. PTO excludes backoff.
type AckProbeInfo struct {
	OutstandingPackets    uint64
	HasLastAckEliciting   bool
	SinceLastAckEliciting time.Duration
	PTO                   time.Duration
}

// SetPacketNumberLengthPolicy installs a client-only policy without extending
// the native handler interface. Callbacks run in the send loop and must not block
// or reenter the connection. Invalid results fall back to the native length.
func SetPacketNumberLengthPolicy(handler SentPacketHandler, policy func(PacketNumberLengthInfo) uint8) {
	if h, ok := handler.(*sentPacketHandler); ok && h.perspective == protocol.PerspectiveClient {
		h.packetNumberLengthPolicy = policy
	}
}

func (h *sentPacketHandler) packetNumberLength(pn, largestAcked protocol.PacketNumber, space protocol.PacketNumberSpace) protocol.PacketNumberLen {
	native := protocol.PacketNumberLengthForHeader(pn, largestAcked)
	if h.packetNumberLengthPolicy == nil {
		return native
	}
	least := pn
	for n, p := range h.history(space).Packets() {
		if p.EncryptionLevel.PacketNumberSpace() == space {
			least = n
			break
		}
	}
	var inFlight uint64
	if h.maxDatagramSize > 0 {
		inFlight = uint64(h.congestion.GetCongestionWindow() / h.maxDatagramSize)
	}
	var coalescedInFlight uint64
	if h.coalescedPacketSpace != space {
		// A change of encryption/number space discards the preceding packet's
		// temporary limit, including when returning to an earlier space later.
		h.coalescedPacketSize = 0
	}
	if h.coalescedPacketSize > 0 {
		coalescedInFlight = uint64(h.congestion.GetCongestionWindow() / h.coalescedPacketSize)
	}
	length := h.packetNumberLengthPolicy(PacketNumberLengthInfo{
		PacketNumber: uint64(pn), LargestAcknowledged: int64(largestAcked),
		LeastUnacknowledged: uint64(least), MaxPacketsInFlight: inFlight, NativeLength: uint8(native),
		CoalescedMaxPacketsInFlight: coalescedInFlight,
	})
	if length < 1 || length > 4 {
		return native
	}
	// RFC 9000 section 17.1: before an ACK encode the full number, then
	// retain enough bits for more than twice the unacknowledged range.
	required := uint64(pn) + 1
	if largestAcked != protocol.InvalidPacketNumber {
		required = uint64(pn-largestAcked) * 2
	}
	for length < 4 && required >= uint64(1)<<(8*length) {
		length++
	}
	return protocol.PacketNumberLen(length)
}

// AckProbeNeeded evaluates a client policy only when an application-data ACK
// is being built. It never creates an independent timer or changes PTO recovery.
func AckProbeNeeded(handler SentPacketHandler, now monotime.Time, policy func(AckProbeInfo) bool) bool {
	h, ok := handler.(*sentPacketHandler)
	if !ok || policy == nil || h.perspective != protocol.PerspectiveClient {
		return false
	}
	info := AckProbeInfo{PTO: h.rttStats.PTO(h.handshakeConfirmed)}
	for _, p := range h.history(protocol.PacketNumberSpaceAppData).Packets() {
		if p.EncryptionLevel.PacketNumberSpace() == protocol.PacketNumberSpaceAppData && !p.isPathProbePacket {
			info.OutstandingPackets++
		}
	}
	last := h.appDataPackets.lastAckElicitingPacketTime
	info.HasLastAckEliciting = !last.IsZero()
	if info.HasLastAckEliciting {
		info.SinceLastAckEliciting = max(0, now.Sub(last))
	}
	return policy(info)
}
