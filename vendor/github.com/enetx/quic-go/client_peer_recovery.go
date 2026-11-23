package quic

import (
	"github.com/enetx/quic-go/internal/ackhandler"
	"github.com/enetx/quic-go/internal/handshake"
	"github.com/enetx/quic-go/internal/monotime"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/wire"
	"github.com/enetx/quic-go/qlog"
)

// PeerPacketInfo describes a routed peer packet whose keys are unavailable.
// Authentication failures never reach this policy.
type PeerPacketInfo struct {
	// PacketType identifies the parsed QUIC encryption level.
	PacketType qlog.PacketType
	// KeysDropped distinguishes retired keys from keys not yet available.
	KeysDropped bool
	// ReceivedFirstPacket reports arrival of a successfully decrypted peer packet.
	ReceivedFirstPacket bool
	// HandshakeComplete reports completion of verified TLS.
	HandshakeComplete bool
	// BufferedHandshakePackets and BufferedApplicationPackets count packets
	// queued before the current packet, independently for each encryption level.
	BufferedHandshakePackets, BufferedApplicationPackets int
}

// PeerPacketRecovery selects transport responses to a missing-key hint.
// Neither response authenticates the packet or delivers its payload to TLS.
type PeerPacketRecovery struct {
	// ResendInitialCrypto schedules sent, unacknowledged Initial bytes.
	ResendInitialCrypto bool
	// ResendEarlyData discards outstanding 0-RTT packet accounting and schedules
	// their recovery tokens again, without a congestion loss or TLS restart.
	ResendEarlyData bool
	// ProbeHandshake schedules PING if Handshake keys are still available.
	ProbeHandshake bool
}

func (c *Conn) recoverFromPeerPacket(pt qlog.PacketType, dropped bool) {
	if c.perspective != protocol.PerspectiveClient || c.config.ClientHandshake == nil || c.config.ClientHandshake.PeerPacketRecoveryPolicy == nil {
		return
	}
	info := PeerPacketInfo{PacketType: pt, KeysDropped: dropped, ReceivedFirstPacket: c.receivedFirstPacket, HandshakeComplete: c.handshakeComplete}
	for _, p := range c.undecryptablePackets {
		if len(p.data) > 0 && wire.IsLongHeaderPacket(p.data[0]) {
			info.BufferedHandshakePackets++
		} else {
			info.BufferedApplicationPackets++
		}
	}
	action := c.config.ClientHandshake.PeerPacketRecoveryPolicy(info)
	if action.ResendInitialCrypto && !c.handshakeComplete && !dropped {
		if r := c.cryptoRecovery(protocol.EncryptionInitial); r != nil {
			r.rescheduleUnacked()
		}
	}
	if action.ResendEarlyData && !dropped {
		c.resendEarlyData(monotime.Now())
	}
	if action.ProbeHandshake && !c.handshakeComplete && !c.retransmissionQueue.HasData(protocol.EncryptionHandshake) {
		if packer, ok := c.packer.(*packetPacker); ok {
			if sealer, err := packer.cryptoSetup.GetHandshakeSealer(); err == nil && sealer != nil {
				c.retransmissionQueue.addHandshake(&wire.PingFrame{})
			}
		}
	}
}

func (c *Conn) resendEarlyData(now monotime.Time) {
	if c.perspective == protocol.PerspectiveClient && !c.handshakeComplete && !c.earlyDataRejected {
		ackhandler.ResendEarlyData(c.sentPacketHandler, now)
	}
}

func (c *Conn) recoverFromUnpackError(err error, pt qlog.PacketType) {
	if err == handshake.ErrKeysDropped || err == handshake.ErrKeysNotYetAvailable {
		c.recoverFromPeerPacket(pt, err == handshake.ErrKeysDropped)
	}
}
