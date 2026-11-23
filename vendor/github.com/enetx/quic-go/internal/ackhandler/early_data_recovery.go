package ackhandler

import (
	"slices"

	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
)

// ResendEarlyData releases early packet accounting without congestion loss and
// returns its recovery tokens to their owners. An Application ACK permanently
// ends this hint mechanism. It neither rejects TLS early data nor resets PTO.
// Keeping this optional concrete helper avoids extending SentPacketHandler.
func ResendEarlyData(handler SentPacketHandler, now monotime.Time) {
	h, ok := handler.(*sentPacketHandler)
	if !ok || h.perspective != protocol.PerspectiveClient || h.earlyDataAcked || h.appDataPackets.largestAcked != protocol.InvalidPacketNumber {
		return
	}
	history := h.history(protocol.PacketNumberSpaceAppData)
	for pn, p := range history.Packets() {
		if p.EncryptionLevel != protocol.Encryption0RTT {
			continue
		}
		h.removeFromBytesInFlight(p)
		history.Remove(pn)
		if len(p.Frames) > 0 || len(p.StreamFrames) > 0 {
			h.queueFramesForRetransmission(p)
		}
	}
	// Peer-hint removal is distinct from ordinary loss: these originals no
	// longer supply late-ACK recovery tokens, matching packet-history removal.
	h.retainedStreamPackets = slices.DeleteFunc(h.retainedStreamPackets, func(p retainedStreamPacket) bool {
		return p.level == protocol.Encryption0RTT
	})
	if h.appDataPackets.numOutstanding == 0 {
		h.appDataPackets.lossTime = 0
	}
	h.setLossDetectionTimer(now)
}
