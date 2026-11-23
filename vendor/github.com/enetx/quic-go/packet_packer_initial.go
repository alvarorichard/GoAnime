package quic

import "github.com/enetx/quic-go/internal/protocol"

// prepareInitialCryptoPacket keeps optional Initial layout/protection outside
// the native packer. The manager remains the sole owner of the crypto streams.
func (p *packetPacker) prepareInitialCryptoPacket(
	level protocol.EncryptionLevel,
	packetSizeLimit, available protocol.ByteCount,
) (bool, error) {
	if level != protocol.EncryptionInitial {
		return false, nil
	}

	stream := p.cryptoStreamManager.initialStream
	if stream == nil {
		return false, nil
	}

	if err := stream.prepareInitialCryptoLayout(packetSizeLimit, available); err != nil {
		return false, err
	}

	if stream.protection == nil {
		return false, nil
	}

	stream.beginProtectedPacket(available, &p.rand)
	state := stream.protection
	return state.ordered && len(state.active) > 0, nil
}
