package ackhandler

import "github.com/enetx/quic-go/internal/congestion"

// ClientSendConfig separates wire/loss bookkeeping from sender arithmetic.
// Server constructors ignore these application-selected client mechanisms.
type ClientSendConfig struct {
	// RetainInitialCryptoProbes keeps original packet ACK callbacks after a PTO.
	RetainInitialCryptoProbes bool
	// StreamDataRecovery retains original Application STREAM probe callbacks.
	StreamDataRecovery bool
	// CoalescePTOProbes selects a shared datagram allowance across timed spaces.
	CoalescePTOProbes bool
	// InitialACKResetsPTO resets backoff before peer address validation.
	InitialACKResetsPTO bool
	// StreamLossRetention selects the lifetime of lost STREAM ACK metadata.
	StreamLossRetention func(StreamLossInfo) bool
	// InitialProbePackets overrides the Initial PTO packet allowance; zero keeps native behavior.
	InitialProbePackets uint8
	// InitialProbeSkipPacketNumber enables a deliberate Initial PTO number gap.
	InitialProbeSkipPacketNumber bool
	// RetainHandshakeCryptoProbes keeps original Handshake ACK callbacks on PTO.
	RetainHandshakeCryptoProbes bool
	// HandshakeProbePackets and ApplicationProbePackets override PTO allowances.
	// Zero retains the native allowance; client config validates the maximum two.
	HandshakeProbePackets, ApplicationProbePackets uint8
	// HandshakeProbeSkipPacketNumber enables a deliberate Handshake PTO gap.
	HandshakeProbeSkipPacketNumber bool
	// ApplicationProbeSkipPacketNumber separates PTO gaps from random gaps.
	// Nil preserves DisablePacketNumberSkipping's native combined behavior.
	ApplicationProbeSkipPacketNumber *bool
	DisablePacketNumberSkipping      bool
	Congestion                       congestion.SenderConfig
	SeparatePacketNumberSpaces       bool
	PersistentCongestionThreshold    uint32
}

func (h *sentPacketHandler) OnApplicationLimited() {
	if sender, ok := h.congestion.(congestion.ApplicationLimitedHandler); ok {
		sender.OnApplicationLimited()
	}
}

// SkipEmptyHandshakeProbe cancels only the unused follow-up allowance after a
// Handshake PTO packet was sent. The first probe is always allowed to elicit ACK.
func SkipEmptyHandshakeProbe(handler SentPacketHandler) bool {
	h, ok := handler.(*sentPacketHandler)
	if !ok || h.ptoMode != SendPTOHandshake {
		return false
	}
	allowance := int(h.handshakeProbePackets)
	if allowance == 0 {
		allowance = 2
	}
	if h.numProbesToSend <= 0 || h.numProbesToSend >= allowance {
		return false
	}
	h.numProbesToSend = 0
	return true
}
