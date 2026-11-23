package ackhandler

import "time"

// SetMaxAckDelay configures the advertised bound before packets are received.
func (h *ReceivedPacketHandler) SetMaxAckDelay(delay time.Duration) {
	h.appDataPackets.ackDelayLimit = delay
}
func (h *appDataReceivedPacketTracker) maxAckDelay() time.Duration {
	if h.ackDelayLimit > 0 {
		return h.ackDelayLimit
	}
	return MaxAckDelay
}
