package surf

import (
	"context"
	"crypto/tls"
	"fmt"
	"net"
	"sync"
	"syscall"
	"time"

	"github.com/enetx/http"
	"github.com/enetx/http/httptrace"
	"github.com/enetx/surf/pkg/connectproxy"
	"github.com/enetx/surf/profiles"
)

func (rt *roundtripper) dialKeepAlive(
	ctx context.Context, network, address string, source *net.Dialer,
) (net.Conn, error) {
	return dialTCPKeepAlive(ctx, network, address, source, rt.tcpConfig.KeepAlive)
}

func (c *Client) dialKeepAlive(
	ctx context.Context, network, address string, source *net.Dialer,
) (net.Conn, error) {
	return dialTCPKeepAlive(ctx, network, address, source, c.tcpKeepAlive)
}

func dialTCPKeepAlive(
	ctx context.Context, network, address string, source *net.Dialer,
	policy func(profiles.TCPKeepAliveInfo) profiles.TCPKeepAliveOptions,
) (net.Conn, error) {
	if policy == nil || source.KeepAlive != _TCPKeepAlive ||
		source.KeepAliveConfig != (net.KeepAliveConfig{}) {
		return source.DialContext(ctx, network, address)
	}

	// Keep the caller's dialer unchanged. Disabling Go's initial setup preserves
	// kernel interval/count defaults where a policy deliberately leaves them alone.
	dialer := *source
	dialer.KeepAlive = -1
	controlContext, control := dialer.ControlContext, dialer.Control
	var connectMu sync.Mutex
	connectStarted := make(map[string]time.Time)
	// Control runs after name resolution for each concrete connection attempt.
	// Keep the winning TCP connect measurement independent of DNS and address
	// fallback delays, while preserving either application control callback.
	dialer.ControlContext = func(ctx context.Context, network, address string, raw syscall.RawConn) error {
		var err error
		if controlContext != nil {
			err = controlContext(ctx, network, address, raw)
		} else if control != nil {
			err = control(network, address, raw)
		}

		if err != nil {
			return err
		}

		connectMu.Lock()
		connectStarted[address] = time.Now()
		connectMu.Unlock()
		return nil
	}

	conn, err := dialer.DialContext(ctx, network, address)
	if err != nil {
		return nil, err
	}

	tcp, ok := conn.(*net.TCPConn)
	if !ok {
		return conn, nil
	}

	connectMu.Lock()
	var elapsed time.Duration
	if started, ok := connectStarted[tcp.RemoteAddr().String()]; ok {
		elapsed = time.Since(started)
	}

	connectMu.Unlock()
	managed := &tcpKeepAliveConn{
		TCPConn: tcp,
		policy:  policy,
		elapsed: elapsed,
	}

	if err := managed.activate("", false); err != nil {
		managed.Close()
		return nil, err
	}

	return managed, nil
}

func (rt *roundtripper) withKeepAliveTrace(req *http.Request) *http.Request {
	if rt.tcpConfig.KeepAlive == nil {
		return req
	}

	trace := &httptrace.ClientTrace{GotConn: func(info httptrace.GotConnInfo) {
		protocol := "http/1.1"
		if secure, ok := info.Conn.(*profileTLSConn); ok {
			protocol = secure.ConnectionState().NegotiatedProtocol
		}

		activateTCPKeepAlive(info.Conn, protocol, info.Reused)
	}}
	return req.WithContext(httptrace.WithClientTrace(req.Context(), trace))
}

func activateTCPKeepAlive(conn net.Conn, protocol string, reused bool) {
	for range 8 {
		switch c := conn.(type) {
		case *tcpKeepAliveConn:
			// Browser socket-option failures do not turn an otherwise valid
			// reused HTTP transaction into a new dial or replay.
			_ = c.activate(protocol, reused)
			return
		case *profileTLSConn:
			conn = c.UConn.NetConn()
		case *tls.Conn:
			conn = c.NetConn()
		default:
			underlying, ok := connectproxy.UnderlyingConn(conn)
			if !ok {
				return // Custom and multiplexed stream connections retain ownership.
			}

			conn = underlying
		}
	}
}

type tcpKeepAliveConn struct {
	*net.TCPConn
	mu         sync.Mutex
	policy     func(profiles.TCPKeepAliveInfo) profiles.TCPKeepAliveOptions
	elapsed    time.Duration
	timer      *time.Timer
	generation uint64
	closed     bool
}

func (c *tcpKeepAliveConn) activate(protocol string, reused bool) error {
	options := c.policy(profiles.TCPKeepAliveInfo{
		Protocol: protocol, DialDuration: c.elapsed, Reused: reused,
	})
	if options.After < 0 {
		return fmt.Errorf("TCP keepalive transition delay must not be negative")
	}

	c.mu.Lock()
	defer c.mu.Unlock()
	if c.closed {
		return net.ErrClosed
	}

	c.generation++
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}

	if err := c.TCPConn.SetKeepAliveConfig(options.Initial); err != nil {
		return err
	}

	if options.After > 0 {
		generation := c.generation
		c.timer = time.AfterFunc(options.After, func() {
			c.mu.Lock()
			defer c.mu.Unlock()
			if !c.closed && c.generation == generation {
				_ = c.TCPConn.SetKeepAliveConfig(options.LongLived)
				c.timer = nil
			}
		})
	}

	return nil
}

func (c *tcpKeepAliveConn) Close() error {
	c.mu.Lock()
	c.closed = true
	c.generation++
	if c.timer != nil {
		c.timer.Stop()
		c.timer = nil
	}

	c.mu.Unlock()
	return c.TCPConn.Close()
}
