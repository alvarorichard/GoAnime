package quic

func (c *Conn) markApplicationLimited() {
	if handler, ok := c.sentPacketHandler.(interface{ OnApplicationLimited() }); ok {
		handler.OnApplicationLimited()
	}
}
