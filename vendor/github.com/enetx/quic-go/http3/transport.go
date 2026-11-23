package http3

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/url"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/enetx/http"
	"github.com/enetx/http/httptrace"

	"github.com/enetx/g"
	"github.com/enetx/quic-go"
	"golang.org/x/net/http/httpguts"
)

// Settings are HTTP/3 settings that apply to the underlying connection.
type Settings struct {
	// MaxFieldSectionSize bounds uncompressed field sections sent to this peer.
	// A negative value means the peer omitted the setting; zero is a real limit.
	MaxFieldSectionSize int64
	// Support for HTTP/3 datagrams (RFC 9297)
	EnableDatagrams bool
	// Extended CONNECT, RFC 9220
	EnableExtendedConnect bool
	// Other settings, defined by the application
	Other g.MapOrd[uint64, uint64]
}

// RoundTripOpt contains options for [Transport.RoundTripOpt].
type RoundTripOpt struct {
	allowRetryDial   bool // RoundTripCached preserves ownership only on an initial cache miss.
	requireConfirmed bool // The cache-only probe must not wait for a pending connection.
	// OnlyCachedConn controls whether the Transport may create a new QUIC connection.
	// If true and no cached connection is available, [Transport.RoundTripOpt] returns [ErrNoCachedConn].
	OnlyCachedConn bool
}

type clientConn interface {
	OpenRequestStream(context.Context) (*RequestStream, error)
	RoundTrip(*http.Request) (*http.Response, error)
	handleUnidirectionalStream(*quic.ReceiveStream)
}

type roundTripperWithCount struct {
	poolKey          string
	partition        string
	authority        string
	cancel           context.CancelFunc
	dialing          chan struct{} // closed as soon as quic.Dial(Early) returned
	connectStarted   chan struct{}
	connectStartOnce sync.Once
	dialErr          error
	conn             *quic.Conn
	clientConn       clientConn

	useCount atomic.Int64
}

func (r *roundTripperWithCount) Close() error {
	r.cancel()
	<-r.dialing
	if r.conn != nil {
		return r.conn.CloseWithError(quic.ApplicationErrorCode(ErrCodeNoError), "")
	}
	return nil
}

// Transport implements the [http.RoundTripper] interface.
type Transport struct {
	// InitialPeerSettings supplies authenticated settings before ordinary request
	// headers and control-stream SETTINGS are applied. It runs once per HTTP/3
	// protocol generation, after TLS completion. The result is copied. It must
	// be fast, concurrency safe, not perform I/O or reenter the transport.
	// Early requests do not wait for or initialize this callback.
	InitialPeerSettings func(*quic.Conn) (g.MapOrd[uint64, uint64], error)
	// PrepareHeaders runs on a private request/header clone for each RoundTrip
	// attempt after connection selection and normal TLS/settings initialization,
	// before a request stream or body read. It may change Header and derive the
	// request Context only. Errors are terminal RequestPreparationErrors.
	// IsEarlyRequest reports the frozen path that skipped authenticated
	// initialization, independently of packet encryption or handshake timing.
	// Raw OpenRequestStream callers retain explicit header ownership. The callback
	// must be fast, concurrency safe, not perform I/O or reenter the transport.
	PrepareHeaders func(*http.Request, *quic.Conn) error
	// ConnectionPoolPartition partitions both exact-authority lookup and
	// cross-authority reuse. The authority is always included in the actual key.
	// It must be pure, concurrency safe, stable for an attempt and not reenter
	// the transport. Nil retains the native single-partition pool.
	ConnectionPoolPartition func(context.Context, string) string
	// ConnectionReusePolicy may authorize an established connection for another
	// authority. The callback owns certificate, route and privacy checks. Only
	// live, handshake-complete, non-draining native clients are considered; nil
	// retains the native per-authority pool. It runs outside pool locks, may run
	// concurrently and must respect ctx. Preconnect invokes the same policy.
	ConnectionReusePolicy func(context.Context, ConnectionReuseInfo) bool
	// EarlyDataRestart keeps a rejected 0-RTT connection and recreates HTTP/3
	// protocol state after verified TLS. Replayable requests restart once;
	// the default preserves the caller-managed rejection behavior.
	// This option applies to pooled RoundTrip connections.
	EarlyDataRestart bool
	// TLSClientConfig specifies the TLS configuration to use with
	// [tls.Client]. If nil, the default configuration is used.
	TLSClientConfig *tls.Config

	// QUICConfig is the [quic.Config] used for dialing new connections.
	// If nil, reasonable default values will be used.
	QUICConfig *quic.Config

	// Dial specifies an optional dial function for creating QUIC
	// connections for requests.
	// If [Transport.Dial] is nil, a [net.UDPConn] will be created at the first request
	// and will be reused for subsequent connections to other servers.
	Dial func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (*quic.Conn, error)

	// Enable support for HTTP/3 datagrams (RFC 9297).
	// If [Transport.QUICConfig] is set, [quic.Config.EnableDatagrams] must also be enabled.
	EnableDatagrams bool

	// Additional HTTP/3 settings.
	// It is invalid to specify any settings defined by RFC 9114 (HTTP/3) and RFC 9297 (HTTP Datagrams).
	AdditionalSettings g.MapOrd[uint64, uint64]

	// MaxResponseHeaderBytes specifies a limit on how many response bytes are
	// allowed in the server's response header.
	// Zero means to use a default limit.
	MaxResponseHeaderBytes int

	// DisableCompression, if true, prevents the Transport from requesting compression with an
	// "Accept-Encoding: gzip" request header when the Request contains no existing Accept-Encoding value.
	// If the Transport requests gzip on its own and gets a gzipped response, it's transparently
	// decoded in the Response.Body.
	// However, if the user explicitly requested gzip it is not automatically uncompressed.
	DisableCompression bool

	// StreamPriority optionally overrides the local scheduling priority of new
	// request streams. Nil preserves quic-go's default scheduling.
	StreamPriority *StreamPriority

	// QPACKConfig optionally controls request compression, critical streams and
	// capacity announcements. DynamicTable enables dynamic request encoding.
	QPACKConfig *QPACKConfig

	// WireConfig optionally controls reserved frames and request priority updates.
	WireConfig *WireConfig

	Logger *slog.Logger

	mutex sync.Mutex

	initOnce sync.Once
	initErr  error

	newClientConn func(*quic.Conn) clientConn

	clients        map[string]*roundTripperWithCount
	retired        map[*roundTripperWithCount]struct{}
	transport      *quic.Transport
	closed         bool
	clientSessions sync.Map // *quic.Conn -> *clientSettingsSession
}

var (
	_ http.RoundTripper = &Transport{}
	_ io.Closer         = &Transport{}
)

var (
	// ErrNoCachedConn is returned by [Transport.RoundTripOpt] when [RoundTripOpt.OnlyCachedConn]
	// is true and no cached connection is available.
	ErrNoCachedConn = errors.New("http3: no cached connection was available")
	// ErrTransportClosed is returned by [Transport.RoundTrip] and [Transport.RoundTripOpt]
	// after the transport is closed.
	ErrTransportClosed = errors.New("http3: transport is closed")
)

func (t *Transport) init() error {
	if err := t.WireConfig.Validate(); err != nil {
		return err
	}
	if err := t.QPACKConfig.Validate(); err != nil {
		return err
	}
	if err := t.validateOrderedSettings(); err != nil {
		return err
	}
	if t.newClientConn == nil {
		t.newClientConn = func(conn *quic.Conn) clientConn {
			return t.newConfiguredClientConn(conn)
		}
	}
	if t.QUICConfig == nil {
		t.QUICConfig = defaultQuicConfig.Clone()
		t.QUICConfig.EnableDatagrams = t.EnableDatagrams
	}
	if t.EnableDatagrams && !t.QUICConfig.EnableDatagrams {
		return errors.New("HTTP Datagrams enabled, but QUIC Datagrams disabled")
	}
	if len(t.QUICConfig.Versions) == 0 {
		t.QUICConfig = t.QUICConfig.Clone()
		t.QUICConfig.Versions = []quic.Version{quic.SupportedVersions()[0]}
	}
	if len(t.QUICConfig.Versions) != 1 && (t.QUICConfig.ClientHandshake == nil || !t.QUICConfig.ClientHandshake.CompatibleVersionNegotiation) {
		return errors.New("can only use a single QUIC version for dialing a HTTP/3 connection")
	}
	if t.QUICConfig.MaxIncomingStreams == 0 {
		t.QUICConfig.MaxIncomingStreams = -1 // don't allow any bidirectional streams
	}
	if t.Dial == nil {
		udpConn, err := net.ListenUDP("udp", nil)
		if err != nil {
			return err
		}
		t.transport = &quic.Transport{Conn: udpConn}
	}
	return nil
}

// RoundTripOpt is like [Transport.RoundTrip], but takes options.
func (t *Transport) RoundTripOpt(req *http.Request, opt RoundTripOpt) (*http.Response, error) {
	req = t.requestWithPreparationBody(req)

	rsp, err := t.roundTripOpt(req, opt)
	if err != nil {
		if req.Body != nil {
			req.Body.Close()
		}
		return nil, err
	}
	return rsp, nil
}

func (t *Transport) roundTripOpt(req *http.Request, opt RoundTripOpt) (*http.Response, error) {
	t.initOnce.Do(func() { t.initErr = t.init() })
	if t.initErr != nil {
		return nil, t.initErr
	}

	if req.URL == nil {
		return nil, errors.New("http3: nil Request.URL")
	}
	if req.URL.Scheme != "https" {
		return nil, fmt.Errorf("http3: unsupported protocol scheme: %s", req.URL.Scheme)
	}
	if req.URL.Host == "" {
		return nil, errors.New("http3: no Host in request URL")
	}
	if req.Header == nil {
		return nil, errors.New("http3: nil Request.Header")
	}
	if req.Method != "" && !validMethod(req.Method) {
		return nil, fmt.Errorf("http3: invalid method %q", req.Method)
	}
	if err := validatePreparedHeaders(req.Header); err != nil {
		return nil, err
	}

	return t.doRoundTripOpt(req, opt, false)
}

func (t *Transport) doRoundTripOpt(req *http.Request, opt RoundTripOpt, isRetried bool) (*http.Response, error) {
	hostname := authorityAddr(hostnameFromURL(req.URL))
	trace := httptrace.ContextClientTrace(req.Context())
	traceGetConn(trace, hostname)
	getClient := t.getClient
	if isRetried || req.Host != "" && authorityAddr(req.Host) != hostname {
		getClient = t.getClientForHost
	}
	cl, isReused, err := getClient(req.Context(), hostname, opt.OnlyCachedConn)
	if err != nil {
		return nil, err
	}
	if opt.requireConfirmed && !reusableClient(cl) {
		cl.useCount.Add(-1)
		return nil, ErrNoCachedConn
	}
	defer cl.useCount.Add(-1)

	select {
	case <-cl.dialing:
	case <-req.Context().Done():
		return nil, context.Cause(req.Context())
	}

	if cl.dialErr != nil {
		t.removeClient(cl)
		return nil, cl.dialErr
	}
	traceGotConn(trace, cl.conn, isReused)
	rsp, err := cl.clientConn.RoundTrip(req)
	if err != nil {
		// request aborted due to context cancellation
		select {
		case <-req.Context().Done():
			return nil, err
		default:
		}
		if isRetried {
			return nil, err
		}

		if t.ConnectionReusePolicy == nil {
			t.removeClient(cl)
		}
		req, err = canRetryRequest(err, req)
		// A request-local failure must not retire another authority's healthy
		// shared connection. Connection failures and permitted retries still do.
		if t.ConnectionReusePolicy != nil && (err == nil || cl.conn.Context().Err() != nil) {
			t.removeClient(cl)
		}
		if err != nil {
			return nil, err
		}
		if opt.allowRetryDial {
			opt.OnlyCachedConn = false
			opt.requireConfirmed = false
		}
		return t.doRoundTripOpt(req, opt, true)
	}
	return rsp, nil
}

func canRetryRequest(err error, req *http.Request) (*http.Request, error) {
	if _, received := errors.AsType[*responseReceivedError](err); received {
		return nil, err
	}
	if _, local := errors.AsType[*RequestPreparationError](err); local {
		return nil, err
	}
	// error occurred while opening the stream, we can be sure that the request wasn't sent out
	if _, ok := errors.AsType[*errConnUnusable](err); ok {
		return req, nil
	}

	// If the request stream is reset, we can only be sure that the request wasn't processed
	// if the error code is H3_REQUEST_REJECTED.
	e, ok := errors.AsType[*Error](err)
	if !ok || e.ErrorCode != ErrCodeRequestRejected {
		return nil, err
	}
	// if the body is nil (or http.NoBody), it's safe to reuse this request and its body
	if req.Body == nil || req.Body == http.NoBody {
		return req, nil
	}
	// if the request body can be reset back to its original state via req.GetBody, do that
	if req.GetBody != nil {
		newBody, err := req.GetBody()
		if err != nil {
			return nil, err
		}
		reqCopy := *req
		reqCopy.Body = newBody
		req = &reqCopy
		return &reqCopy, nil
	}
	return nil, fmt.Errorf("http3: Transport: cannot retry err [%w] after Request.Body was written; define Request.GetBody to avoid this error", err)
}

// RoundTrip does a round trip.
func (t *Transport) RoundTrip(req *http.Request) (*http.Response, error) {
	return t.RoundTripOpt(req, RoundTripOpt{})
}

func (t *Transport) getClientForHost(ctx context.Context, hostname string, onlyCached bool) (rtc *roundTripperWithCount, isReused bool, err error) {
	key, partition := t.clientPoolKey(ctx, hostname)
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.closed {
		return nil, false, ErrTransportClosed
	}

	if t.clients == nil {
		t.clients = make(map[string]*roundTripperWithCount)
	}

	cl, ok := t.clients[key]
	if ok && !onlyCached && retiredClient(cl) {
		t.retireClientLocked(cl)
		cl, ok = nil, false
	}
	if !ok {
		if onlyCached {
			return nil, false, ErrNoCachedConn
		}
		// The dial is shared by all requests to this host,
		// so it must not be canceled when the first request is canceled.
		ctx, cancel := context.WithCancel(context.WithoutCancel(ctx))
		cl = &roundTripperWithCount{
			poolKey: key, partition: partition, authority: hostname,
			dialing:        make(chan struct{}),
			connectStarted: make(chan struct{}),
			cancel:         cancel,
		}
		ctx = context.WithValue(ctx, connectionStartedKey{}, func() {
			cl.connectStartOnce.Do(func() { close(cl.connectStarted) })
		})
		go func() {
			defer close(cl.dialing)
			defer cancel()
			conn, rt, err := t.dial(ctx, hostname)
			if err != nil {
				cl.dialErr = err
				return
			}
			cl.conn = conn
			cl.clientConn = rt
		}()
		t.clients[key] = cl
	}
	select {
	case <-cl.dialing:
		if cl.dialErr != nil {
			// A fresh dial's first owner still observes its reported start and
			// completion through the normal waiter, even if dial finished here.
			if ok {
				delete(t.clients, key)
				return nil, false, cl.dialErr
			}
		} else {
			select {
			case <-cl.conn.HandshakeComplete():
				isReused = true
			default:
			}
		}
	default:
	}
	cl.useCount.Add(1)
	return cl, isReused, nil
}

func (t *Transport) dial(ctx context.Context, hostname string) (*quic.Conn, clientConn, error) {
	var tlsConf *tls.Config
	if t.TLSClientConfig == nil {
		tlsConf = &tls.Config{}
	} else {
		tlsConf = t.TLSClientConfig.Clone()
	}
	if tlsConf.ServerName == "" {
		sni, _, err := net.SplitHostPort(hostname)
		if err != nil {
			// It's ok if net.SplitHostPort returns an error - it could be a hostname/IP address without a port.
			sni = hostname
		}
		tlsConf.ServerName = sni
	}
	// Replace existing ALPNs by H3
	tlsConf.NextProtos = []string{NextProtoH3}

	dial := t.Dial
	if dial == nil {
		dial = func(ctx context.Context, addr string, tlsCfg *tls.Config, cfg *quic.Config) (*quic.Conn, error) {
			network := "udp"
			udpAddr, err := t.resolveUDPAddr(ctx, network, addr)
			if err != nil {
				return nil, err
			}
			trace := httptrace.ContextClientTrace(ctx)
			NotifyConnectionStarted(ctx)
			traceConnectStart(trace, network, udpAddr.String())
			traceTLSHandshakeStart(trace)
			conn, err := t.transport.DialEarly(ctx, udpAddr, tlsCfg, cfg)
			var state tls.ConnectionState
			if conn != nil {
				state = conn.ConnectionState().TLS
			}
			traceTLSHandshakeDone(trace, state, err)
			traceConnectDone(trace, network, udpAddr.String(), err)
			return conn, err
		}
	}
	config := t.QUICConfig
	// A custom dialer can prepare several address candidates. Initial SETTINGS
	// belong to each candidate's own HTTP/3 connection, not whichever became
	// application-ready first. Retain only candidates still awaiting selection.
	var clientsMu sync.Mutex
	clients := make(map[*quic.Conn]clientConn)
	var prepared bool
	var activeRestart atomic.Pointer[earlyRestartClient]
	var initialInputReady chan struct{}
	if t.QPACKConfig != nil {
		initialInputReady = make(chan struct{})
	}
	defer func() {
		clientsMu.Lock()
		prepared = true
		clear(clients)
		clientsMu.Unlock()
	}()
	ensureClient := func(conn *quic.Conn) clientConn {
		clientsMu.Lock()
		defer clientsMu.Unlock()
		if prepared {
			if active := activeRestart.Load(); active != nil && active.conn == conn {
				return active
			}
			return nil // A late, unselected candidate no longer needs setup.
		}
		client := clients[conn]
		if client == nil {
			client = t.newClientConn(conn)
			if native, ok := client.(*ClientConn); ok {
				native.initialInputReady = initialInputReady
				if session, found := t.clientSessions.Load(conn); found {
					native.bindSession(session.(*clientSettingsSession))
				}
			}
			if native, ok := client.(*ClientConn); ok && t.EarlyDataRestart {
				client = newEarlyRestartClient(native, func() *ClientConn {
					fresh := t.newConfiguredClientConn(conn)
					fresh.initialInputReady = initialInputReady
					if session, found := t.clientSessions.Load(conn); found {
						fresh.bindSession(session.(*clientSettingsSession))
					}
					return fresh
				})
			}
			clients[conn] = client
		}
		return client
	}
	config = config.Clone()
	if config.ClientHandshake == nil {
		config.ClientHandshake = &quic.ClientHandshake{}
	}
	previousWrap := config.ClientHandshake.WrapTLSClient
	config.ClientHandshake.WrapTLSClient = t.wrapSessionTLS(previousWrap)
	if t.EarlyDataRestart {
		config.ClientHandshake.ResetStreamsOnEarlyDataRejection = true
		previous := config.ClientHandshake.OnEarlyDataRejected
		config.ClientHandshake.OnEarlyDataRejected = func(conn *quic.Conn) {
			if active := activeRestart.Load(); active != nil {
				if active.conn == conn {
					active.reject()
				}
			} else {
				clientsMu.Lock()
				if client, ok := clients[conn].(*earlyRestartClient); ok {
					client.reject()
				}
				clientsMu.Unlock()
			}
			if previous != nil {
				previous(conn)
			}
		}
	}
	if t.WireConfig != nil && (t.WireConfig.CoalesceInitialSettings || t.WireConfig.FlushEarlySettings) {
		if t.WireConfig.FlushEarlySettings {
			config.ClientHandshake.FlushEarlyData = true
		}
		previous := config.ClientHandshake.OnEarlyDataReady
		config.ClientHandshake.OnEarlyDataReady = func(conn *quic.Conn) {
			ensureClient(conn)
			if previous != nil {
				previous(conn)
			}
		}
	}
	if t.EarlyDataRestart || t.WireConfig != nil && (t.WireConfig.CoalesceInitialSettings || t.WireConfig.FlushEarlySettings) {
		previous := config.ClientHandshake.OnApplicationReady
		config.ClientHandshake.OnApplicationReady = func(conn *quic.Conn) {
			client := ensureClient(conn)
			if client, ok := client.(*earlyRestartClient); ok {
				client.applicationReady()
			}
			if previous != nil {
				previous(conn)
			}
		}
	}
	conn, err := dial(ctx, hostname, tlsConf, config)
	if err != nil {
		return nil, nil, err
	}
	clientConn := ensureClient(conn)
	if client, ok := clientConn.(*earlyRestartClient); ok {
		activeRestart.Store(client)
	}
	go func() {
		if initialInputReady != nil {
			select {
			case <-conn.HandshakeComplete():
				switch client := clientConn.(type) {
				case *ClientConn:
					client.dispatchBufferedInitialStreams()
				case *earlyRestartClient:
					client.current.Load().dispatchBufferedInitialStreams()
				}
			case <-conn.Context().Done():
			}
			close(initialInputReady)
		}

		for {
			str, err := conn.AcceptUniStream(context.Background())
			if err != nil {
				if client, ok := clientConn.(*earlyRestartClient); ok && errors.Is(err, quic.Err0RTTRejected) {
					if _, err := client.active(conn.Context()); err == nil {
						continue
					}
				}
				return
			}
			go clientConn.handleUnidirectionalStream(str)
		}
	}()
	return conn, clientConn, nil
}

func (t *Transport) resolveUDPAddr(ctx context.Context, network, addr string) (*net.UDPAddr, error) {
	host, portStr, err := net.SplitHostPort(addr)
	if err != nil {
		return nil, err
	}
	port, err := net.LookupPort(network, portStr)
	if err != nil {
		return nil, err
	}
	resolver := net.DefaultResolver
	ipAddrs, err := resolver.LookupIPAddr(ctx, host)
	if err != nil {
		return nil, err
	}
	addrs := addrList(ipAddrs)
	ip := addrs.forResolve(network, addr)
	return &net.UDPAddr{IP: ip.IP, Port: port, Zone: ip.Zone}, nil
}

func (t *Transport) removeClient(client *roundTripperWithCount) {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	if t.clients == nil {
		return
	}
	t.retireClientLocked(client)
}

// NewClientConn creates a new HTTP/3 client connection on top of a QUIC connection.
// Most users should use [Transport.RoundTrip] instead of creating a connection directly.
// Specifically, it is not needed to perform GET, POST, HEAD and CONNECT requests.
//
// Obtaining a [ClientConn] is only needed for more advanced use cases, such as
// using Extended CONNECT for WebTransport or the various MASQUE protocols.
func (t *Transport) NewClientConn(conn *quic.Conn) *ClientConn {
	c := t.newConfiguredClientConn(conn)
	go func() {
		for {
			str, err := conn.AcceptUniStream(context.Background())
			if err != nil {
				return
			}
			go c.handleUnidirectionalStream(str)
		}
	}()
	return c
}

// NewRawClientConn creates a new low-level HTTP/3 client connection on top of a QUIC connection.
// Unlike [Transport.NewClientConn], the returned [RawClientConn] allows the application to take control
// of the stream accept loops by calling [RawClientConn.HandleUnidirectionalStream] for incoming
// unidirectional streams and [ClientConn.HandleBidirectionalStream] for incoming bidirectional streams.
func (t *Transport) NewRawClientConn(conn *quic.Conn) *RawClientConn {
	return &RawClientConn{
		ClientConn: t.newConfiguredClientConn(conn),
	}
}

// Close closes the QUIC connections that this Transport has used.
// A Transport cannot be used after it has been closed.
func (t *Transport) Close() error {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for _, cl := range t.clients {
		if err := cl.Close(); err != nil {
			return err
		}
	}
	t.clients = nil
	for client := range t.retired {
		if err := client.Close(); err != nil {
			return err
		}
	}
	t.retired = nil
	if t.transport != nil {
		if err := t.transport.Close(); err != nil {
			return err
		}
		if err := t.transport.Conn.Close(); err != nil {
			return err
		}
		t.transport = nil
	}
	t.closed = true
	return nil
}

func hostnameFromURL(url *url.URL) string {
	if url != nil {
		return url.Host
	}
	return ""
}

func validMethod(method string) bool {
	// RFC 9110, section 9.1: a method is a nonempty token.
	return len(method) > 0 && strings.IndexFunc(method, isNotToken) == -1
}

// copied from net/http/http.go
func isNotToken(r rune) bool {
	return !httpguts.IsTokenRune(r)
}

// CloseIdleConnections closes any QUIC connections in the transport's pool that are currently idle.
// An idle connection is one that was previously used for requests but is now sitting unused.
// This method does not interrupt any connections currently in use.
// It also does not affect connections obtained via [Transport.NewClientConn].
func (t *Transport) CloseIdleConnections() {
	t.mutex.Lock()
	defer t.mutex.Unlock()
	for hostname, cl := range t.clients {
		if cl.useCount.Load() == 0 && !hasActiveClientStreams(cl) {
			cl.Close()
			delete(t.clients, hostname)
		}
	}
	for cl := range t.retired {
		if cl.useCount.Load() == 0 && !hasActiveClientStreams(cl) {
			cl.Close()
			delete(t.retired, cl)
		}
	}
}
