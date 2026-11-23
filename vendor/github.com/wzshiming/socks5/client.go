package socks5

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net"
	"net/url"
	"time"
)

// Dialer is a SOCKS5 dialer.
type Dialer struct {
	// ProxyNetwork network between a proxy server and a client
	ProxyNetwork string
	// ProxyAddress proxy server address
	ProxyAddress string
	// ProxyDial specifies the optional dial function for
	// establishing the transport connection.
	ProxyDial func(ctx context.Context, network string, address string) (net.Conn, error)
	// ProxyPacketDial specifies the optional proxyPacketDial function for
	// establishing the transport connection.
	ProxyPacketDial func(ctx context.Context, network string, address string) (net.PacketConn, error)
	// Username use username authentication if not empty
	Username string
	// Password use password authentication if not empty,
	// only valid if username is set
	Password string
	// IsResolve resolve domain name on locally
	IsResolve bool
	// Resolver optionally specifies an alternate resolver to use
	Resolver *net.Resolver
	// Timeout is the maximum amount of time a dial will wait for
	// a connect to complete. The default is no timeout
	Timeout time.Duration
}

// NewDialer returns a new Dialer that dials through the provided
// proxy server's network and address.
func NewDialer(addr string) (*Dialer, error) {
	d := &Dialer{
		ProxyNetwork: "tcp",
		Timeout:      time.Minute,
	}
	u, err := url.Parse(addr)
	if err != nil {
		return nil, err
	}
	switch u.Scheme {
	case "socks5":
		d.IsResolve = true
	case "socks5h":
	default:
		return nil, fmt.Errorf("unsupported protocol '%s'", u.Scheme)
	}
	host := u.Host
	port := u.Port()
	if port == "" {
		port = "1080"
		hostname := u.Hostname()
		host = net.JoinHostPort(hostname, port)
	}
	if u.User != nil {
		d.Username = u.User.Username()
		d.Password, _ = u.User.Password()
	}
	d.ProxyAddress = host
	return d, nil
}

// DialContext connects to the provided address on the provided network.
func (d *Dialer) DialContext(ctx context.Context, network, address string) (net.Conn, error) {
	switch network {
	default:
		return nil, fmt.Errorf("unsupported network %q", network)
	case "tcp", "tcp4", "tcp6":
		return d.do(ctx, ConnectCommand, address)
	case "udp", "udp4", "udp6":
		return d.do(ctx, AssociateCommand, address)
	}
}

// Dial connects to the provided address on the provided network.
func (d *Dialer) Dial(network, address string) (net.Conn, error) {
	return d.DialContext(context.Background(), network, address)
}

func (d *Dialer) Listen(ctx context.Context, network, address string) (net.Listener, error) {
	switch network {
	default:
		return nil, fmt.Errorf("unsupported network %q", network)
	case "tcp", "tcp4", "tcp6":
	}
	return &listener{ctx: ctx, d: d, address: address}, nil
}

func (d *Dialer) do(ctx context.Context, cmd Command, address string) (net.Conn, error) {
	if d.IsResolve {
		host, port, err := net.SplitHostPort(address)
		if err != nil {
			return nil, err
		}
		if host != "" {
			ip := net.ParseIP(host)
			if ip == nil {
				ipaddr, err := d.resolver().LookupIP(ctx, "ip4", host)
				if err != nil {
					ipaddr, err = d.resolver().LookupIP(ctx, "ip", host)
					if err != nil {
						return nil, err
					}
				}
				host := ipaddr[0].String()
				address = net.JoinHostPort(host, port)
			}
		}
	}

	conn, err := d.proxyDial(ctx, d.ProxyNetwork, d.ProxyAddress)
	if err != nil {
		return nil, err
	}

	c, err := d.connect(ctx, conn, cmd, address)
	if err != nil {
		_ = conn.Close()
		return nil, err
	}
	return c, nil
}

func (d *Dialer) connect(ctx context.Context, conn net.Conn, cmd Command, address string) (net.Conn, error) {
	if d.Timeout != 0 {
		deadline := time.Now().Add(d.Timeout)
		if d, ok := ctx.Deadline(); !ok || deadline.Before(d) {
			subCtx, cancel := context.WithDeadline(ctx, deadline)
			defer cancel()
			ctx = subCtx
		}
	}
	if deadline, ok := ctx.Deadline(); ok && !deadline.IsZero() {
		conn.SetDeadline(deadline)
		defer conn.SetDeadline(time.Time{})
	}

	err := d.connectAuth(conn)
	if err != nil {
		return nil, err
	}

	switch cmd {
	default:
		return nil, fmt.Errorf("unsupported Command %s", cmd)
	case ConnectCommand:
		_, err := d.connectCommand(conn, ConnectCommand, address)
		if err != nil {
			return nil, err
		}
		return conn, nil
	case BindCommand:
		_, err := d.connectCommand(conn, BindCommand, address)
		if err != nil {
			return nil, err
		}
		return conn, nil
	case AssociateCommand:
		targetIP, targetPort, err := splitHostPort(address)
		if err != nil {
			return nil, err
		}

		addr, err := d.connectCommand(conn, AssociateCommand, ":0")
		if err != nil {
			return nil, err
		}

		proxyIP, proxyPort, err := splitHostPort(addr.String())
		if err != nil {
			return nil, err
		}

		proxyAddr := &net.UDPAddr{
			IP:   net.ParseIP(proxyIP),
			Port: proxyPort,
		}
		network := "udp"
		if proxyAddr.IP.To4() != nil {
			network = "udp4"
		} else if proxyAddr.IP != nil {
			network = "udp6"
		}
		udpConn, err := d.proxyPacketDial(ctx, network, ":0")
		if err != nil {
			return nil, err
		}

		targetAddr := &net.UDPAddr{
			IP:   net.ParseIP(targetIP),
			Port: targetPort,
		}
		wrapConn, err := newUDPConn(udpConn, proxyAddr, targetAddr, conn)
		if err != nil {
			return nil, err
		}

		go func() {
			var buf [1]byte
			for {
				_, err := conn.Read(buf[:])
				if err != nil {
					wrapConn.Close()
					break
				}
			}
		}()
		return wrapConn, nil
	}

}

func (d *Dialer) connectAuth(conn net.Conn) error {
	_, err := conn.Write([]byte{socks5Version})
	if err != nil {
		return err
	}
	if d.Username == "" {
		err = writeBytes(conn, []byte{byte(noAuth)})
		if err != nil {
			return err
		}
	} else {
		err = writeBytes(conn, []byte{byte(noAuth), byte(userAuth)})
		if err != nil {
			return err
		}
	}

	var header [2]byte
	_, err = io.ReadFull(conn, header[:])
	if err != nil {
		return err
	}
	if header[0] != socks5Version {
		return fmt.Errorf("unexpected protocol version %d", header[0])
	}
	if authMethod(header[1]) == noAcceptable {
		return fmt.Errorf("no acceptable authentication methods %d", authMethod(header[1]))
	}
	switch authMethod(header[1]) {
	default:
		return fmt.Errorf("authentication method not supported %d", authMethod(header[1]))
	case noAuth:
	case userAuth:
		if d.Username == "" {
			return errors.New("need username/password")
		}

		if len(d.Username) == 0 || len(d.Username) > 255 || len(d.Password) == 0 || len(d.Password) > 255 {
			return errors.New("invalid username/password")
		}
		_, err = conn.Write([]byte{userAuthVersion})
		if err != nil {
			return err
		}
		err = writeBytes(conn, []byte(d.Username))
		if err != nil {
			return err
		}
		err = writeBytes(conn, []byte(d.Password))
		if err != nil {
			return err
		}

		_, err := io.ReadFull(conn, header[:])
		if err != nil {
			return err
		}
		if header[0] != userAuthVersion {
			return fmt.Errorf("invalid username/password version %d", header[0])
		}
		if header[1] != authSuccess {
			return fmt.Errorf("username/password authentication failed %d", header[1])
		}
	}
	return nil
}

func (d *Dialer) connectCommand(conn net.Conn, cmd Command, address string) (net.Addr, error) {
	_, err := conn.Write([]byte{socks5Version, byte(cmd), 0})
	if err != nil {
		return nil, err
	}
	err = writeAddrWithStr(conn, address)
	if err != nil {
		return nil, err
	}

	return d.readReply(conn)
}

func (d *Dialer) readReply(conn net.Conn) (net.Addr, error) {
	var header [3]byte
	_, err := io.ReadFull(conn, header[:])
	if err != nil {
		return nil, err
	}

	if header[0] != socks5Version {
		return nil, fmt.Errorf("unexpected protocol version %d", header[0])
	}

	if reply(header[1]) != successReply {
		return nil, fmt.Errorf("unknown error %s", reply(header[1]).String())
	}

	return readAddr(conn)
}

func (d *Dialer) resolver() *net.Resolver {
	if d.Resolver == nil {
		return net.DefaultResolver
	}
	return d.Resolver
}

func (d *Dialer) proxyDial(ctx context.Context, network, address string) (net.Conn, error) {
	proxyDial := d.ProxyDial
	if proxyDial == nil {
		var dialer net.Dialer
		proxyDial = dialer.DialContext
	}
	return proxyDial(ctx, network, address)
}

func (d *Dialer) proxyPacketDial(ctx context.Context, network, address string) (net.PacketConn, error) {
	proxyPacketDial := d.ProxyPacketDial
	if proxyPacketDial == nil {
		var listener net.ListenConfig
		proxyPacketDial = listener.ListenPacket
	}
	return proxyPacketDial(ctx, network, address)
}

type listener struct {
	ctx     context.Context
	d       *Dialer
	address string
}

// Accept waits for and returns the next connection to the listener.
func (l *listener) Accept() (net.Conn, error) {
	conn, err := l.d.do(l.ctx, BindCommand, l.address)
	if err != nil {
		return nil, err
	}
	addr, err := l.d.readReply(conn)
	if err != nil {
		return nil, err
	}
	return &connect{Conn: conn, remoteAddr: addr}, nil
}

// Close closes the listener.
func (l *listener) Close() error {
	return nil
}

// address returns the listener's network address.
func (l *listener) Addr() net.Addr {
	return nil
}

type connect struct {
	net.Conn
	remoteAddr net.Addr
}

func (c *connect) RemoteAddr() net.Addr {
	return c.remoteAddr
}
