//go:build darwin || linux || freebsd

package udp

import "golang.org/x/net/ipv4"

// setReadBatchSize runs before the first read, while all buffers are unowned.
// Single-read platforms retain their native capacity instead of allocating an
// unused batch. Configuration has already been validated by Wrap.
func (c *oobConn) setReadBatchSize(size int) {
	if size == 0 || size == batchSize || batchSize == 1 {
		return
	}

	c.messages = make([]ipv4.Message, size)
	c.buffers = make([]*PacketBuffer, size)
	c.readPos = uint8(size)

	for i := range c.messages {
		c.messages[i].Buffers = make([][]byte, 1)
		c.messages[i].OOB = make([]byte, oobBufferSize)
	}
}
