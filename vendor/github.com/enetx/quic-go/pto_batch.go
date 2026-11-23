package quic

import (
	"errors"

	"github.com/enetx/quic-go/internal/ackhandler"
	"github.com/enetx/quic-go/internal/monotime"
)

func (c *Conn) primePTOHandshake(now monotime.Time) {
	if c.config.ClientHandshake == nil || !c.config.ClientHandshake.CoalescePTOProbes {
		return
	}
	if packer, ok := c.packer.(*packetPacker); ok {
		sealer, err := packer.cryptoSetup.GetHandshakeSealer()
		ackhandler.PrimePTOHandshake(c.sentPacketHandler, err == nil && sealer != nil, now)
	}
}

func (c *Conn) sendCoalescedPTOProbe(now monotime.Time) (handled bool, err error) {
	if c.config == nil || c.config.ClientHandshake == nil || !c.config.ClientHandshake.CoalescePTOProbes {
		return false, nil
	}
	packer, ok := c.packer.(*packetPacker)
	if !ok {
		return true, errors.New("quic: coalesced PTO requires the connection packet packer")
	}
	force, ok := ackhandler.BeginPTOProbeDatagram(c.sentPacketHandler, c.maxPacketSize())
	if !ok {
		return false, nil
	}
	packet, err := packer.packPTOProbeDatagram(force, c.maxPacketSize(), now, c.version)
	if err != nil || packet == nil {
		return true, err
	}
	return true, c.sendPackedCoalescedPacket(packet, c.sentPacketHandler.ECNMode(packet.IsOnlyShortHeaderPacket()), now)
}
