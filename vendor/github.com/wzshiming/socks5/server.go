package socks5

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"
)

// Server is accepting connections and handling the details of the SOCKS5 protocol
type Server struct {
	// Authentication is proxy authentication
	Authentication Authentication
	// ProxyDial specifies the optional proxyDial function for
	// establishing the transport connection.
	ProxyDial func(ctx context.Context, network string, address string) (net.Conn, error)
	// ProxyListen specifies the optional proxyListen function for
	// establishing the transport connection.
	ProxyListen func(ctx context.Context, network string, address string) (net.Listener, error)
	// ProxyListenPacket specifies the optional proxyListenPacket function for
	// establishing the transport connection.
	ProxyListenPacket func(ctx context.Context, network string, address string) (net.PacketConn, error)
	// ProxyOutgoingListenPacket specifies the optional proxyOutgoingListenPacket function for
	// establishing outgoing UDP connections. If not set, ProxyListenPacket is used for both
	// relay and outgoing connections.
	ProxyOutgoingListenPacket func(ctx context.Context, network string, address string) (net.PacketConn, error)
	// PacketForwardAddress specifies the packet forwarding address
	PacketForwardAddress func(ctx context.Context, destinationAddr string, packet net.PacketConn, conn net.Conn) (net.IP, int, error)
	// ProxyListenBind specifies the optional proxyListenBind function for
	// establishing the transport connection.
	ProxyListenBind func(ctx context.Context, network string, address string) (net.Listener, error)
	// ListenBindReuseTimeout is the timeout for reusing bind listener
	ListenBindReuseTimeout time.Duration
	// ListenBindAcceptTimeout is the timeout for accepting connections on bind listener
	ListenBindAcceptTimeout time.Duration
	// HandshakeTimeout is the maximum amount of time to complete the SOCKS5
	// negotiation before the request is processed, protecting the server
	// from clients that connect and then stall. Zero means no timeout.
	HandshakeTimeout time.Duration
	// reserveListenBind is a pool for reusing bind listeners across requests.
	reserveListenBind reserveListen
	// Logger error log
	Logger Logger
	// Context is default context
	Context context.Context
	// BytesPool getting and returning temporary bytes for use by io.CopyBuffer
	BytesPool BytesPool
}

type Logger interface {
	Println(v ...interface{})
}

// NewServer creates a new Server
func NewServer() *Server {
	return &Server{
		ListenBindReuseTimeout: time.Second / 2,
		HandshakeTimeout:       10 * time.Second,
	}
}

// ListenAndServe is used to create a listener and serve on it
func (s *Server) ListenAndServe(network, addr string) error {
	l, err := s.proxyListen(s.context(), network, addr)
	if err != nil {
		return err
	}
	return s.Serve(l)
}

func (s *Server) proxyListen(ctx context.Context, network, address string) (net.Listener, error) {
	proxyListen := s.ProxyListen
	if proxyListen == nil {
		var listenConfig net.ListenConfig
		proxyListen = listenConfig.Listen
	}
	return proxyListen(ctx, network, address)
}

func (s *Server) proxyListenBind(ctx context.Context, network, address string) (net.Listener, error) {
	proxyListenBind := s.ProxyListenBind
	if proxyListenBind == nil {
		var listenConfig net.ListenConfig
		proxyListenBind = listenConfig.Listen
	}
	return proxyListenBind(ctx, network, address)
}

// Serve is used to serve connections from a listener
func (s *Server) Serve(l net.Listener) error {
	for {
		conn, err := l.Accept()
		if err != nil {
			return err
		}
		go s.ServeConn(conn)
	}
}

// ServeConn is used to serve a single connection.
func (s *Server) ServeConn(conn net.Conn) {
	defer conn.Close()
	err := s.serveConn(conn)
	if err != nil && s.Logger != nil && !isClosedConnError(err) {
		s.Logger.Println(err)
	}
}

func (s *Server) serveConn(conn net.Conn) error {
	if s.HandshakeTimeout > 0 {
		conn.SetDeadline(time.Now().Add(s.HandshakeTimeout))
	}

	version, err := readByte(conn)
	if err != nil {
		return err
	}
	if version != socks5Version {
		return fmt.Errorf("unsupported SOCKS version: %d", version)
	}

	req := &request{
		Version: socks5Version,
		Conn:    conn,
	}

	methods, err := readBytes(conn)
	if err != nil {
		return err
	}

	if s.Authentication != nil && bytes.IndexByte(methods, byte(userAuth)) != -1 {
		_, err := conn.Write([]byte{socks5Version, byte(userAuth)})
		if err != nil {
			return err
		}

		header, err := readByte(conn)
		if err != nil {
			return err
		}
		if header != userAuthVersion {
			return fmt.Errorf("unsupported auth version: %d", header)
		}

		username, err := readBytes(conn)
		if err != nil {
			return err
		}
		req.Username = string(username)

		password, err := readBytes(conn)
		if err != nil {
			return err
		}
		req.Password = string(password)

		if !s.Authentication.Auth(req.Command, req.Username, req.Password) {
			_, err := conn.Write([]byte{userAuthVersion, authFailure})
			if err != nil {
				return err
			}
			return errUserAuthFailed
		}
		_, err = conn.Write([]byte{userAuthVersion, authSuccess})
		if err != nil {
			return err
		}
	} else if s.Authentication == nil && bytes.IndexByte(methods, byte(noAuth)) != -1 {
		_, err := conn.Write([]byte{socks5Version, byte(noAuth)})
		if err != nil {
			return err
		}
	} else {
		_, err := conn.Write([]byte{socks5Version, byte(noAcceptable)})
		if err != nil {
			return err
		}
		return errNoSupportedAuth
	}

	var header [3]byte
	_, err = io.ReadFull(conn, header[:])
	if err != nil {
		return err
	}

	if header[0] != socks5Version {
		return fmt.Errorf("unsupported Command version: %d", header[0])
	}

	req.Command = Command(header[1])

	dest, err := readAddr(conn)
	if err != nil {
		if err == errUnrecognizedAddrType {
			err := sendReply(conn, addrTypeNotSupported, nil)
			if err != nil {
				return err
			}
		}
		return err
	}
	req.DestinationAddr = dest

	if s.HandshakeTimeout > 0 {
		conn.SetDeadline(time.Time{})
	}

	err = s.handle(req)
	if err != nil {
		return err
	}

	return nil
}

func (s *Server) handle(req *request) error {
	switch req.Command {
	case ConnectCommand:
		return s.handleConnect(req)
	case BindCommand:
		return s.handleBind(req)
	case AssociateCommand:
		return s.handleAssociate(req)
	default:
		if err := sendReply(req.Conn, commandNotSupported, nil); err != nil {
			return err
		}
		return fmt.Errorf("unsupported Command: %v", req.Command)
	}
}

func (s *Server) handleConnect(req *request) error {
	ctx := s.context()
	target, err := s.proxyDial(ctx, "tcp", req.DestinationAddr.Address())
	if err != nil {
		if err := sendReply(req.Conn, errToReply(err), nil); err != nil {
			return fmt.Errorf("failed to send reply: %v", err)
		}
		return fmt.Errorf("connect to %v failed: %w", req.DestinationAddr, err)
	}
	defer target.Close()

	localAddr := target.LocalAddr()
	local, ok := localAddr.(*net.TCPAddr)
	if !ok {
		return fmt.Errorf("connect to %v failed: local address is %s://%s", req.DestinationAddr, localAddr.Network(), localAddr.String())
	}
	bind := address{IP: local.IP, Port: local.Port}
	if err := sendReply(req.Conn, successReply, &bind); err != nil {
		return fmt.Errorf("failed to send reply: %v", err)
	}

	var buf1, buf2 []byte
	if s.BytesPool != nil {
		buf1 = s.BytesPool.Get()
		buf2 = s.BytesPool.Get()
		defer func() {
			s.BytesPool.Put(buf1)
			s.BytesPool.Put(buf2)
		}()
	} else {
		buf1 = make([]byte, 32*1024)
		buf2 = make([]byte, 32*1024)
	}
	return tunnel(ctx, target, req.Conn, buf1, buf2)
}

func (s *Server) handleBind(req *request) error {
	ctx := s.context()
	addr := req.DestinationAddr.String()

	var listener net.Listener
	var err error
	if s.ListenBindReuseTimeout > 0 {
		listener, err = s.reserveListenBind.getOrNew(addr, func() (net.Listener, error) {
			return s.proxyListenBind(ctx, "tcp", addr)
		}, s.ListenBindReuseTimeout, s.ListenBindAcceptTimeout, s.Logger)
	} else {
		listener, err = s.proxyListenBind(ctx, "tcp", addr)
	}
	if err != nil {
		if err := sendReply(req.Conn, errToReply(err), nil); err != nil {
			return fmt.Errorf("failed to send reply: %v", err)
		}
		return fmt.Errorf("connect to %v failed: %w", req.DestinationAddr, err)
	}

	localAddr := listener.Addr()
	local, ok := localAddr.(*net.TCPAddr)
	if !ok {
		listener.Close()
		return fmt.Errorf("connect to %v failed: local address is %s://%s", req.DestinationAddr, localAddr.Network(), localAddr.String())
	}
	bind := address{IP: local.IP, Port: local.Port}
	if err := sendReply(req.Conn, successReply, &bind); err != nil {
		listener.Close()
		return fmt.Errorf("failed to send reply: %v", err)
	}

	conn, err := listener.Accept()
	if err != nil {
		listener.Close()
		if err := sendReply(req.Conn, errToReply(err), nil); err != nil {
			return fmt.Errorf("failed to send reply: %v", err)
		}
		return fmt.Errorf("connect to %v failed: %w", req.DestinationAddr, err)
	}
	listener.Close()

	remoteAddr := conn.RemoteAddr()
	local, ok = remoteAddr.(*net.TCPAddr)
	if !ok {
		return fmt.Errorf("connect to %v failed: remote address is %s://%s", req.DestinationAddr, localAddr.Network(), localAddr.String())
	}
	bind = address{IP: local.IP, Port: local.Port}
	if err := sendReply(req.Conn, successReply, &bind); err != nil {
		return fmt.Errorf("failed to send reply: %v", err)
	}

	var buf1, buf2 []byte
	if s.BytesPool != nil {
		buf1 = s.BytesPool.Get()
		buf2 = s.BytesPool.Get()
		defer func() {
			s.BytesPool.Put(buf1)
			s.BytesPool.Put(buf2)
		}()
	} else {
		buf1 = make([]byte, 32*1024)
		buf2 = make([]byte, 32*1024)
	}
	return tunnel(ctx, conn, req.Conn, buf1, buf2)
}

func (s *Server) handleAssociate(req *request) error {
	ctx := s.context()
	destinationAddr := req.DestinationAddr.String()
	udpConn, err := s.proxyListenPacket(ctx, "udp", destinationAddr)
	if err != nil {
		if err := sendReply(req.Conn, errToReply(err), nil); err != nil {
			return fmt.Errorf("failed to send reply: %v", err)
		}
		return fmt.Errorf("connect to %v failed: %w", req.DestinationAddr, err)
	}
	defer udpConn.Close()

	replyPacketForwardAddress := defaultReplyPacketForwardAddress
	if s.PacketForwardAddress != nil {
		replyPacketForwardAddress = s.PacketForwardAddress
	}
	ip, port, err := replyPacketForwardAddress(ctx, destinationAddr, udpConn, req.Conn)
	if err != nil {
		return err
	}
	bind := address{IP: ip, Port: port}
	if err := sendReply(req.Conn, successReply, &bind); err != nil {
		return fmt.Errorf("failed to send reply: %v", err)
	}

	if s.ProxyOutgoingListenPacket == nil {
		return s.handleAssociateLegacy(ctx, req, udpConn)
	}

	outgoingConn, err := s.ProxyOutgoingListenPacket(ctx, "udp", ":0")
	if err != nil {
		if err := sendReply(req.Conn, errToReply(err), nil); err != nil {
			return fmt.Errorf("failed to send reply: %v", err)
		}
		return fmt.Errorf("outgoing connect failed: %w", err)
	}
	defer outgoingConn.Close()
	return s.handleAssociateWithSeparateConns(ctx, req, udpConn, outgoingConn)
}

func (s *Server) handleAssociateLegacy(ctx context.Context, req *request, udpConn net.PacketConn) error {
	go func() {
		var buf [1]byte
		for {
			_, err := req.Conn.Read(buf[:])
			if err != nil {
				udpConn.Close()
				break
			}
		}
	}()

	var (
		sourceAddr net.Addr
		wantSource string
		buf        [maxUdpPacket]byte
		replyBuf   [maxHeaderSize]byte
	)

	for {
		n, addr, err := udpConn.ReadFrom(buf[:])
		if err != nil {
			return err
		}

		if sourceAddr == nil {
			sourceAddr = addr
			wantSource = sourceAddr.String()
		}

		gotAddr := addr.String()
		if wantSource == gotAddr {
			// Packet from client to target
			if n < 3 {
				continue
			}
			reader := bytes.NewBuffer(buf[3:n])
			targetAddr, err := readAddr(reader)
			if err != nil {
				if s.Logger != nil {
					s.Logger.Println(err)
				}
				continue
			}
			var targetIP net.IP
			if targetAddr.IP != nil {
				targetIP = targetAddr.IP
			} else if targetAddr.Name != "" {
				ips, err := net.LookupIP(targetAddr.Name)
				if err != nil {
					if s.Logger != nil {
						s.Logger.Println(fmt.Errorf("failed to resolve %s: %w", targetAddr.Name, err))
					}
					continue
				}
				if len(ips) == 0 {
					if s.Logger != nil {
						s.Logger.Println(fmt.Errorf("no IP addresses found for %s", targetAddr.Name))
					}
					continue
				}
				targetIP = ips[0]
				if s.Logger != nil {
					s.Logger.Println(fmt.Sprintf("Resolved %s to %v", targetAddr.Name, targetIP))
				}
			} else {
				if s.Logger != nil {
					s.Logger.Println(fmt.Errorf("no valid address in UDP packet"))
				}
				continue
			}
			target := &net.UDPAddr{
				IP:   targetIP,
				Port: targetAddr.Port,
			}
			_, err = udpConn.WriteTo(reader.Bytes(), target)
			if err != nil {
				return err
			}
		} else {
			headWriter := bytes.NewBuffer(replyBuf[:0])
			headWriter.Write([]byte{0, 0, 0})
			err = writeAddrWithStr(headWriter, gotAddr)
			if err != nil {
				if s.Logger != nil {
					s.Logger.Println(err)
				}
				continue
			}
			prefixLen := headWriter.Len()

			// Check if data length plus header exceeds maximum UDP packet limit
			if prefixLen+n > maxUdpPacket {
				if s.Logger != nil {
					s.Logger.Println(fmt.Errorf("dropping packet: data length (%d) + header length (%d) = %d exceeds max UDP packet size %d", n, prefixLen, prefixLen+n, maxUdpPacket))
				}
				continue
			}

			copy(buf[prefixLen:prefixLen+n], buf[:n])
			copy(buf[:prefixLen], headWriter.Bytes())

			_, err = udpConn.WriteTo(buf[:prefixLen+n], sourceAddr)
			if err != nil {
				return err
			}
		}
	}
}

func (s *Server) handleAssociateWithSeparateConns(ctx context.Context, req *request, udpConn, outgoingConn net.PacketConn) error {
	errChan := make(chan error, 3)

	go func() {
		var buf [1]byte
		for {
			_, err := req.Conn.Read(buf[:])
			if err != nil {
				udpConn.Close()
				outgoingConn.Close()
				errChan <- nil
				break
			}
		}
	}()

	var (
		sourceAddr net.Addr
		wantSource string
	)

	go func() {
		var buf [maxUdpPacket]byte
		for {
			n, addr, err := udpConn.ReadFrom(buf[:])
			if err != nil {
				errChan <- err
				return
			}

			if sourceAddr == nil {
				sourceAddr = addr
				wantSource = sourceAddr.String()
				go func() {
					var buf [maxUdpPacket]byte
					var replyBuf [maxHeaderSize]byte

					for {
						n, addr, err := outgoingConn.ReadFrom(buf[:])
						if err != nil {
							errChan <- err
							return
						}

						gotAddr := addr.String()
						headWriter := bytes.NewBuffer(replyBuf[:0])
						headWriter.Write([]byte{0, 0, 0})
						err = writeAddrWithStr(headWriter, gotAddr)
						if err != nil {
							if s.Logger != nil {
								s.Logger.Println(err)
							}
							continue
						}
						prefixLen := headWriter.Len()

						// Check if data length plus header exceeds maximum UDP packet limit
						if prefixLen+n > maxUdpPacket {
							if s.Logger != nil {
								s.Logger.Println(fmt.Errorf("dropping packet: data length (%d) + header length (%d) = %d exceeds max UDP packet size %d", n, prefixLen, prefixLen+n, maxUdpPacket))
							}
							continue
						}

						copy(buf[prefixLen:prefixLen+n], buf[:n])
						copy(buf[:prefixLen], headWriter.Bytes())

						_, err = udpConn.WriteTo(buf[:prefixLen+n], sourceAddr)
						if err != nil {
							errChan <- err
							return
						}
					}
				}()
			}

			// Packet from client to target
			if n < 3 {
				continue
			}

			if addr.String() != wantSource {
				continue
			}

			reader := bytes.NewBuffer(buf[3:n])
			targetAddr, err := readAddr(reader)
			if err != nil {
				if s.Logger != nil {
					s.Logger.Println(err)
				}
				continue
			}
			var targetIP net.IP
			if targetAddr.IP != nil {
				targetIP = targetAddr.IP
			} else if targetAddr.Name != "" {
				ips, err := net.LookupIP(targetAddr.Name)
				if err != nil {
					if s.Logger != nil {
						s.Logger.Println(fmt.Errorf("failed to resolve %s: %w", targetAddr.Name, err))
					}
					continue
				}
				if len(ips) == 0 {
					if s.Logger != nil {
						s.Logger.Println(fmt.Errorf("no IP addresses found for %s", targetAddr.Name))
					}
					continue
				}
				targetIP = ips[0]
				if s.Logger != nil {
					s.Logger.Println(fmt.Sprintf("Resolved %s to %v", targetAddr.Name, targetIP))
				}
			} else {
				if s.Logger != nil {
					s.Logger.Println(fmt.Errorf("no valid address in UDP packet"))
				}
				continue
			}
			target := &net.UDPAddr{
				IP:   targetIP,
				Port: targetAddr.Port,
			}
			_, err = outgoingConn.WriteTo(reader.Bytes(), target)
			if err != nil {
				errChan <- err
				return
			}
		}
	}()

	return <-errChan
}

func (s *Server) proxyDial(ctx context.Context, network, address string) (net.Conn, error) {
	proxyDial := s.ProxyDial
	if proxyDial == nil {
		var dialer net.Dialer
		proxyDial = dialer.DialContext
	}
	return proxyDial(ctx, network, address)
}

func (s *Server) proxyListenPacket(ctx context.Context, network, address string) (net.PacketConn, error) {
	proxyListenPacket := s.ProxyListenPacket
	if proxyListenPacket == nil {
		var listener net.ListenConfig
		proxyListenPacket = listener.ListenPacket
	}
	return proxyListenPacket(ctx, network, address)
}

func (s *Server) context() context.Context {
	if s.Context == nil {
		return context.Background()
	}
	return s.Context
}

func sendReply(w io.Writer, resp reply, addr *address) error {
	_, err := w.Write([]byte{socks5Version, byte(resp), 0})
	if err != nil {
		return err
	}
	err = writeAddr(w, addr)
	return err
}

type request struct {
	Version         uint8
	Command         Command
	DestinationAddr *address
	Username        string
	Password        string
	Conn            net.Conn
}

func defaultReplyPacketForwardAddress(ctx context.Context, destinationAddr string, packet net.PacketConn, conn net.Conn) (net.IP, int, error) {
	udpLocal := packet.LocalAddr()
	udpLocalAddr, ok := udpLocal.(*net.UDPAddr)
	if !ok {
		return nil, 0, fmt.Errorf("connect to %v failed: local address is %s://%s", destinationAddr, udpLocal.Network(), udpLocal.String())
	}

	tcpLocal := conn.LocalAddr()
	tcpLocalAddr, ok := tcpLocal.(*net.TCPAddr)
	if !ok {
		return nil, 0, fmt.Errorf("connect to %v failed: local address is %s://%s", destinationAddr, tcpLocal.Network(), tcpLocal.String())
	}
	return tcpLocalAddr.IP, udpLocalAddr.Port, nil
}

type reserveListen struct {
	mut               sync.Mutex
	reservedListeners map[string]*reserved
}

type reserved struct {
	key   string
	base  net.Listener
	conns chan net.Conn
}

type holdListener struct {
	r      *reserved
	closed atomic.Bool
}

func (r *reserveListen) getOrNew(key string, newFunc func() (net.Listener, error), reuse, accept time.Duration, logger Logger) (net.Listener, error) {
	r.mut.Lock()
	defer r.mut.Unlock()

	reserve := r.reservedListeners[key]
	if reserve != nil {
		return &holdListener{r: reserve}, nil
	}

	listener, err := newFunc()
	if err != nil {
		return nil, err
	}
	reserve = &reserved{
		key:   key,
		base:  listener,
		conns: make(chan net.Conn),
	}
	if r.reservedListeners == nil {
		r.reservedListeners = map[string]*reserved{}
	}
	r.reservedListeners[key] = reserve

	if accept > 0 {
		_, ok := listener.(setDeadline)
		if !ok {
			accept = 0
			if logger != nil {
				logger.Println("reserve bind listener does not support SetDeadline, disabling accept timeout")
			}
		}
	}
	go reserve.run(reuse, accept, logger)
	return &holdListener{r: reserve}, nil
}

type setDeadline interface {
	SetDeadline(t time.Time) error
}

func (r *reserved) run(reuse, accept time.Duration, logger Logger) {
	defer func() {
		r.base.Close()
		close(r.conns)
	}()

	for {
		if accept > 0 {
			r.base.(setDeadline).SetDeadline(time.Now().Add(accept))
		}
		conn, err := r.base.Accept()
		if err != nil {
			if logger != nil {
				logger.Println("reserve bind listen accept error:", err)
			}
			return
		}

		select {
		case r.conns <- conn:
		case <-time.After(reuse):
			conn.Close()
			if logger != nil {
				logger.Println("reserve bind listen reuse timeout")
			}
			return
		}
	}
}

func (h *holdListener) Accept() (net.Conn, error) {
	if h.closed.Load() {
		return nil, net.ErrClosed
	}
	conn, ok := <-h.r.conns
	if !ok {
		h.closed.Store(true)
		return nil, net.ErrClosed
	}
	return conn, nil
}

func (h *holdListener) Close() error {
	if h.closed.Swap(true) {
		return net.ErrClosed
	}
	return nil
}

func (h *holdListener) Addr() net.Addr {
	return h.r.base.Addr()
}
