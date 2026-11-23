package quic

import (
	"github.com/enetx/quic-go/internal/ackhandler"
	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
)

// releaseEarlyData gives setup writes their own send opportunity before callers
// can submit a request. It never waits for peer credit, pacing or the send queue.
func (c *Conn) releaseEarlyData(now monotime.Time) error {
	if !c.earlyDataReadyPending {
		return nil
	}
	if c.cryptoStreamManager.HasData(protocol.EncryptionInitial) && !c.sendQueue.WouldBlock() && c.sentPacketHandler.SendMode(now) == ackhandler.SendAny {
		return nil // Finish currently sendable CRYPTO before staging application data.
	}
	c.earlyDataReadyPending = false
	c.config.ClientHandshake.OnEarlyDataReady(c)
	for !c.sendQueue.WouldBlock() {
		if err := c.triggerSending(monotime.Now()); err != nil {
			return err
		}
		if !c.pacingDeadline.Equal(deadlineSendImmediately) {
			break
		}
	}
	close(c.earlyConnReadyChan)
	return nil
}
