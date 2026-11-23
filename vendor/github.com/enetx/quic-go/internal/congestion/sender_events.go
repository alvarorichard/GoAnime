package congestion

import (
	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
)

// TimeUntilSend supplies live flight accounting to configured pacing.
func TimeUntilSend(sender SendAlgorithm, inFlight protocol.ByteCount) monotime.Time {
	if configured, ok := sender.(*configuredSender); ok {
		return configured.timeUntilSend(inFlight)
	}

	return sender.TimeUntilSend()
}

// OnPacketSent preserves byte and flight context for configured controllers.
func OnPacketSent(sender SendAlgorithm, sentTime monotime.Time, inFlight protocol.ByteCount,
	number protocol.PacketNumber, size protocol.ByteCount, ackEliciting bool,
) {
	if configured, ok := sender.(*configuredSender); ok {
		configured.onPacketSent(sentTime, inFlight, number, size, ackEliciting)
		return
	}

	sender.OnPacketSent(sentTime, number, size, ackEliciting)
}

// OnPacketAcked supplies acknowledged bytes and event time to CUBIC and pacing.
func OnPacketAcked(sender SendAlgorithm, number protocol.PacketNumber,
	ackedBytes, priorInFlight protocol.ByteCount, eventTime monotime.Time,
) {
	if configured, ok := sender.(*configuredSender); ok {
		configured.onPacketAcked(number, ackedBytes, priorInFlight, eventTime)
		return
	}

	sender.OnPacketAcked(number, priorInFlight)
}

// OnCongestionEvent distinguishes byte loss from ECN and supplies PRR flight state.
func OnCongestionEvent(sender SendAlgorithm, number protocol.PacketNumber,
	lostBytes, priorInFlight protocol.ByteCount,
) {
	if configured, ok := sender.(*configuredSender); ok {
		configured.onCongestionEvent(number, lostBytes, priorInFlight)
		return
	}

	sender.OnCongestionEvent(number)
}
