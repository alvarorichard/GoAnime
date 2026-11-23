package socks5

import (
	"bytes"
	"net"
	"net/netip"
)

// ReadFromUDPAddrPort implements the net.UDPConn ReadFromUDPAddrPort method.
func (c *UDPConn) ReadFromUDPAddrPort(b []byte) (n int, addr netip.AddrPort, err error) {
	udpConn, ok := c.PacketConn.(*net.UDPConn)
	if !ok {
		return 0, addr, errUnsupportedMethod
	}
	var a *address
	for {
		n, addr, err = udpConn.ReadFromUDPAddrPort(c.bufRead[:])
		if err != nil {
			return 0, addr, err
		}
		if n < len(c.prefix) || addr.String() != c.proxyAddress.String() {
			continue
		}
		buf := bytes.NewBuffer(c.bufRead[len(c.prefix):n])
		a, err = readAddr(buf)
		if err != nil {
			continue
		}
		n = copy(b, buf.Bytes())
		break
	}
	netipaddr, err := netip.ParseAddr(a.IP.String())
	if err != nil {
		return 0, addr, err
	}
	return n, netip.AddrPortFrom(netipaddr, uint16(a.Port)), nil
}

// ReadMsgUDPAddrPort implements the net.UDPConn ReadMsgUDPAddrPort method.
func (c *UDPConn) ReadMsgUDPAddrPort(b, oob []byte) (n, oobn, flags int, addr netip.AddrPort, err error) {
	udpConn, ok := c.PacketConn.(*net.UDPConn)
	if !ok {
		return 0, 0, 0, addr, errUnsupportedMethod
	}
	var a *address
	for {
		n, oobn, flags, addr, err = udpConn.ReadMsgUDPAddrPort(c.bufRead[:], oob)
		if err != nil {
			return 0, 0, 0, addr, err
		}
		if n < len(c.prefix) || addr.String() != c.proxyAddress.String() {
			continue
		}
		buf := bytes.NewBuffer(c.bufRead[len(c.prefix):n])
		a, err = readAddr(buf)
		if err != nil {
			continue
		}
		n = copy(b, buf.Bytes())
		break
	}
	netipaddr, err := netip.ParseAddr(a.IP.String())
	if err != nil {
		return 0, 0, 0, addr, err
	}
	return n, oobn, flags, netip.AddrPortFrom(netipaddr, uint16(a.Port)), nil
}

// WriteToUDPAddrPort implements the net.UDPConn WriteToUDPAddrPort method.
func (c *UDPConn) WriteToUDPAddrPort(b []byte, addr netip.AddrPort) (n int, err error) {
	udpConn, ok := c.PacketConn.(*net.UDPConn)
	if !ok {
		return 0, errUnsupportedMethod
	}
	buf := bytes.NewBuffer(c.bufWrite[:0])
	buf.Write(c.prefix)
	err = writeAddrWithStr(buf, addr.String())
	if err != nil {
		return 0, err
	}
	_, err = buf.Write(b)
	if err != nil {
		return 0, err
	}

	netipaddrport, err := netip.ParseAddrPort(c.proxyAddress.String())
	if err != nil {
		return 0, err
	}

	data := buf.Bytes()
	_, err = udpConn.WriteToUDPAddrPort(data, netipaddrport)
	if err != nil {
		return 0, err
	}
	return len(b), nil
}

// WriteMsgUDPAddrPort implements the net.UDPConn WriteMsgUDPAddrPort method.
func (c *UDPConn) WriteMsgUDPAddrPort(b, oob []byte, addr netip.AddrPort) (n, oobn int, err error) {
	udpConn, ok := c.PacketConn.(*net.UDPConn)
	if !ok {
		return 0, 0, errUnsupportedMethod
	}

	buf := bytes.NewBuffer(c.bufWrite[:0])
	buf.Write(c.prefix)
	err = writeAddrWithStr(buf, addr.String())
	if err != nil {
		return 0, 0, err
	}

	_, err = buf.Write(b)
	if err != nil {
		return 0, 0, err
	}

	netipaddrport, err := netip.ParseAddrPort(c.proxyAddress.String())
	if err != nil {
		return 0, 0, err
	}

	data := buf.Bytes()
	_, _, err = udpConn.WriteMsgUDPAddrPort(data, oob, netipaddrport)
	if err != nil {
		return 0, 0, err
	}
	return len(b), len(oob), nil
}
