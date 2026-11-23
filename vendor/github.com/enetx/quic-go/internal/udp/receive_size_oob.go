//go:build darwin || linux || freebsd

package udp

func (c *oobConn) setReadPacketSize(size int) { c.receivePacketSize = size }
