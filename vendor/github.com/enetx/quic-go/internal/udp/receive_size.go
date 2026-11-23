package udp

import (
	"fmt"
	"sync"

	"github.com/enetx/quic-go/internal/protocol"
)

// receivePacketBuffer preserves the default pool and opts into the existing
// large-buffer pool only when the application requests a larger receive limit.
func receivePacketBuffer(size int) *PacketBuffer {
	if size == 0 {
		size = protocol.MaxPacketBufferSize
	}
	var buffer *PacketBuffer
	if size > protocol.MaxLargePacketBufferSize {
		buffer = getJumboReceivePacketBuffer()
	} else if size > protocol.MaxPacketBufferSize {
		buffer = GetLargePacketBuffer()
	} else {
		buffer = GetPacketBuffer()
	}
	buffer.Data = buffer.Data[:size]
	return buffer
}
func ValidateReceivePacketSize(size int) error {
	if size != 0 && (size < protocol.MinInitialPacketSize || size > MaxUDPReceivePacketSize) {
		return fmt.Errorf("QUIC receive packet size must be between %d and %d", protocol.MinInitialPacketSize, MaxUDPReceivePacketSize)
	}
	return nil
}
func (c *basicConn) setReadPacketSize(size int) { c.receivePacketSize = size }

// MaxUDPReceivePacketSize is QUIC's default maximum UDP payload limit (RFC 9000).
// Applications omitting max_udp_payload_size must provision this receive size.
const MaxUDPReceivePacketSize = 65527

var jumboReceivePool = sync.Pool{New: func() any { return &PacketBuffer{Data: make([]byte, 0, MaxUDPReceivePacketSize)} }}

func getJumboReceivePacketBuffer() *PacketBuffer {
	buffer := jumboReceivePool.Get().(*PacketBuffer)
	buffer.refCount = 1
	buffer.Data = buffer.Data[:0]
	return buffer
}
