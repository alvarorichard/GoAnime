package socks5

import (
	"bytes"
	"errors"
	"net"
	"time"
)

var (
	errUnsupportedMethod = errors.New("unsupported method")
)

type UDPConn struct {
	bufRead       [maxUdpPacket]byte
	bufWrite      [maxUdpPacket]byte
	proxyAddress  *net.UDPAddr
	defaultTarget *net.UDPAddr
	prefix        []byte
	control       net.Conn
	net.PacketConn
}

func convertToUDPAddr(addr net.Addr) (*net.UDPAddr, error) {
	udpAddr, ok := addr.(*net.UDPAddr)
	if ok {
		return udpAddr, nil
	}

	host, port, err := net.SplitHostPort(addr.String())
	if err != nil {
		return nil, err
	}
	ip := net.ParseIP(host)
	if ip == nil {
		ips, err := net.LookupIP(host)
		if err != nil {
			return nil, err
		}
		ip = ips[0]
	}

	portInt, err := net.LookupPort("udp", port)
	if err != nil {
		return nil, err
	}
	return &net.UDPAddr{
		IP:   ip,
		Port: portInt,
	}, nil
}

func newUDPConn(raw net.PacketConn, proxyAddress net.Addr, defaultTarget net.Addr, control net.Conn) (*UDPConn, error) {
	proxyAddr, err := convertToUDPAddr(proxyAddress)
	if err != nil {
		return nil, err
	}
	defaultTargetAddr, err := convertToUDPAddr(defaultTarget)
	if err != nil {
		return nil, err
	}

	conn := &UDPConn{
		PacketConn:    raw,
		proxyAddress:  proxyAddr,
		defaultTarget: defaultTargetAddr,
		prefix:        []byte{0, 0, 0},
		control:       control,
	}
	return conn, nil
}

// Close closes the connection and the control connection of its association.
func (c *UDPConn) Close() error {
	err := c.PacketConn.Close()
	if c.control != nil {
		c.control.Close()
	}
	return err
}

// ReadFrom implements the net.PacketConn ReadFrom method.
func (c *UDPConn) ReadFrom(p []byte) (n int, addr net.Addr, err error) {
	for {
		n, addr, err = c.PacketConn.ReadFrom(c.bufRead[:])
		if err != nil {
			return 0, nil, err
		}
		if n < len(c.prefix) || addr.String() != c.proxyAddress.String() {
			continue
		}
		buf := bytes.NewBuffer(c.bufRead[len(c.prefix):n])
		a, err := readAddr(buf)
		if err != nil {
			continue
		}
		n = copy(p, buf.Bytes())
		return n, a, nil
	}
}

// ReadFromUDP implements the net.UDPConn ReadFromUDP method.
func (c *UDPConn) ReadFromUDP(p []byte) (n int, addr *net.UDPAddr, err error) {
	udpConn, ok := c.PacketConn.(*net.UDPConn)
	if !ok {
		return 0, nil, errUnsupportedMethod
	}
	for {
		n, addr, err = udpConn.ReadFromUDP(c.bufRead[:])
		if err != nil {
			return 0, nil, err
		}
		if n < len(c.prefix) || addr.String() != c.proxyAddress.String() {
			continue
		}
		buf := bytes.NewBuffer(c.bufRead[len(c.prefix):n])
		a, err := readAddr(buf)
		if err != nil {
			continue
		}
		n = copy(p, buf.Bytes())
		return n, &net.UDPAddr{
			IP:   a.IP,
			Port: a.Port,
		}, nil
	}
}

// WriteTo implements the net.PacketConn WriteTo method.
func (c *UDPConn) WriteTo(p []byte, addr net.Addr) (n int, err error) {
	buf := bytes.NewBuffer(c.bufWrite[:0])
	buf.Write(c.prefix)
	err = writeAddrWithStr(buf, addr.String())
	if err != nil {
		return 0, err
	}
	n, err = buf.Write(p)
	if err != nil {
		return 0, err
	}

	data := buf.Bytes()
	_, err = c.PacketConn.WriteTo(data, c.proxyAddress)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// WriteToUDP implements the net.UDPConn WriteToUDP method.
func (c *UDPConn) WriteToUDP(p []byte, addr *net.UDPAddr) (n int, err error) {
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
	n, err = buf.Write(p)
	if err != nil {
		return 0, err
	}

	data := buf.Bytes()
	_, err = udpConn.WriteToUDP(data, c.proxyAddress)
	if err != nil {
		return 0, err
	}
	return n, nil
}

// Read implements the net.Conn Read method.
func (c *UDPConn) Read(b []byte) (int, error) {
	n, addr, err := c.ReadFrom(b)
	if err != nil {
		return 0, err
	}
	if addr.String() != c.defaultTarget.String() {
		return c.Read(b)
	}
	return n, nil
}

// Write implements the net.Conn Write method.
func (c *UDPConn) Write(b []byte) (int, error) {
	return c.WriteTo(b, c.defaultTarget)
}

// RemoteAddr implements the net.Conn RemoteAddr method.
func (c *UDPConn) RemoteAddr() net.Addr {
	return c.defaultTarget
}

// SetReadBuffer implements the net.UDPConn SetReadBuffer method.
func (c *UDPConn) SetReadBuffer(bytes int) error {
	udpConn, ok := c.PacketConn.(*net.UDPConn)
	if !ok {
		return errUnsupportedMethod
	}
	return udpConn.SetReadBuffer(bytes)
}

// SetWriteBuffer implements the net.UDPConn SetWriteBuffer method.
func (c *UDPConn) SetWriteBuffer(bytes int) error {
	udpConn, ok := c.PacketConn.(*net.UDPConn)
	if !ok {
		return errUnsupportedMethod
	}
	return udpConn.SetWriteBuffer(bytes)
}

// SetDeadline implements the Conn SetDeadline method.
func (c *UDPConn) SetDeadline(t time.Time) error {
	udpConn, ok := c.PacketConn.(*net.UDPConn)
	if !ok {
		return errUnsupportedMethod
	}
	return udpConn.SetDeadline(t)
}

// SetReadDeadline implements the Conn SetReadDeadline method.
func (c *UDPConn) SetReadDeadline(t time.Time) error {
	udpConn, ok := c.PacketConn.(*net.UDPConn)
	if !ok {
		return errUnsupportedMethod
	}
	return udpConn.SetReadDeadline(t)
}

// SetWriteDeadline implements the Conn SetWriteDeadline method.
func (c *UDPConn) SetWriteDeadline(t time.Time) error {
	udpConn, ok := c.PacketConn.(*net.UDPConn)
	if !ok {
		return errUnsupportedMethod
	}
	return udpConn.SetWriteDeadline(t)
}

// ReadMsgUDP implements the net.UDPConn ReadMsgUDP method.
func (c *UDPConn) ReadMsgUDP(b, oob []byte) (n, oobn, flags int, addr *net.UDPAddr, err error) {
	udpConn, ok := c.PacketConn.(*net.UDPConn)
	if !ok {
		return 0, 0, 0, nil, errUnsupportedMethod
	}

	for {
		n, oobn, flags, addr, err = udpConn.ReadMsgUDP(c.bufRead[:], oob)
		if err != nil {
			return 0, 0, 0, nil, err
		}
		if n < len(c.prefix) || addr.String() != c.proxyAddress.String() {
			continue
		}
		buf := bytes.NewBuffer(c.bufRead[len(c.prefix):n])
		a, err := readAddr(buf)
		if err != nil {
			continue
		}
		n = copy(b, buf.Bytes())
		return n, oobn, flags, &net.UDPAddr{
			IP:   a.IP,
			Port: a.Port,
		}, nil
	}
}

// WriteMsgUDP implements the net.UDPConn WriteMsgUDP method.
func (c *UDPConn) WriteMsgUDP(b, oob []byte, addr *net.UDPAddr) (n, oobn int, err error) {
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
	data := buf.Bytes()
	_, _, err = udpConn.WriteMsgUDP(data, oob, c.proxyAddress)
	if err != nil {
		return 0, 0, err
	}
	return len(b), len(oob), nil
}
