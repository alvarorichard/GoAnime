package quic

import (
	"fmt"

	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/udp"
	"github.com/enetx/quic-go/internal/wire"
)

func (t *Transport) validateReceiveLimit(config *Config) error {
	if config == nil {
		return nil
	}
	if h := config.ClientHandshake; h != nil && (h.MaxUDPPayloadSize > 0 || h.OmitMaxUDPPayloadSize) {
		size := t.ReceivePacketSize
		if size == 0 {
			size = protocol.MaxPacketBufferSize
		}
		required := h.MaxUDPPayloadSize
		if h.OmitMaxUDPPayloadSize {
			required = MaxUDPReceivePacketSize
		}
		if required > uint64(size) {
			return fmt.Errorf("advertised max_udp_payload_size %d exceeds transport receive size %d", h.MaxUDPPayloadSize, size)
		}
	}
	return nil
}
func (c *Conn) maxReceiveDatagramFrameSize() protocol.ByteCount {
	if c.perspective == protocol.PerspectiveClient && c.config != nil && c.config.ClientHandshake != nil && c.config.ClientHandshake.MaxDatagramFrameSize > 0 {
		return protocol.ByteCount(c.config.ClientHandshake.MaxDatagramFrameSize)
	}
	return wire.MaxDatagramSize
}

// MaxUDPReceivePacketSize is QUIC's default maximum UDP payload limit (RFC 9000).
// Applications omitting max_udp_payload_size must provision this receive size.
const MaxUDPReceivePacketSize = udp.MaxUDPReceivePacketSize
