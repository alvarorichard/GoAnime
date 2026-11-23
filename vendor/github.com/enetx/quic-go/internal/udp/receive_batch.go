package udp

import "fmt"

// ValidateReceiveBatchSize bounds the message count used by the UDP reader.
// Zero preserves the platform default. The reader's position fits in a uint8.
func ValidateReceiveBatchSize(size int) error {
	if size < 0 || size > 255 {
		return fmt.Errorf("QUIC receive batch size must be between 0 and 255: %d", size)
	}

	return nil
}

// A PacketConn without OOB/batch support continues to read one datagram.
func (c *basicConn) setReadBatchSize(int) {}
