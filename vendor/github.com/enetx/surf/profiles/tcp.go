package profiles

import (
	"net"
	"time"
)

// TCPKeepAliveInfo describes a Surf-owned TCP connection. Protocol is empty
// before TLS negotiation or without ALPN, "http/1.1" for plain HTTP, or the negotiated ALPN.
// DialDuration measures the TCP connect operation. Reused marks pool reuse.
type TCPKeepAliveInfo struct {
	Protocol     string        // Empty before TLS/without ALPN; otherwise the selected HTTP protocol.
	DialDuration time.Duration // Winning TCP connect, excluding DNS and address fallback wait.
	Reused       bool          // True when an existing pooled connection is activated again.
}

// TCPKeepAliveOptions selects initial socket settings and an optional timed
// transition. After zero disables the transition. Negative net.KeepAliveConfig
// values retain the kernel value; zero values use Go's defaults.
type TCPKeepAliveOptions struct {
	Initial   net.KeepAliveConfig // Settings on connection activation.
	After     time.Duration       // Delay before LongLived; restarted on reuse.
	LongLived net.KeepAliveConfig // Settings after After while the connection lives.
}

// TCPConfig describes TLS record and shutdown behavior for JA-owned TCP
// connections. QUIC carries TLS handshake messages without TLS records.
type TCPConfig struct {
	// KeepAlive configures owned TCP sockets at dial, ALPN negotiation and pool
	// reuse. Nil retains the application's dialer settings. A custom DialContext
	// or customized net.Dialer keepalive settings retain application ownership.
	KeepAlive func(TCPKeepAliveInfo) TCPKeepAliveOptions
	// DynamicRecordSizingDisabled sends up to 16 KiB from the first application
	// write instead of gradually increasing the record size.
	DynamicRecordSizingDisabled bool
	// OmitCloseNotify closes the underlying socket without a TLS close alert.
	// Explicit half-close operations retain the TLS engine's semantics.
	OmitCloseNotify bool
}

// KeepAlivePolicy configures socket keepalive and its optional timed transition.
func (c *TCPConfig) KeepAlivePolicy(policy func(TCPKeepAliveInfo) TCPKeepAliveOptions) *TCPConfig {
	c.KeepAlive = policy
	return c
}

// DynamicRecordSizing selects the TLS engine's adaptive application record sizes.
func (c *TCPConfig) DynamicRecordSizing(enabled bool) *TCPConfig {
	c.DynamicRecordSizingDisabled = !enabled
	return c
}

// CloseNotify selects whether closing a completed TLS connection sends an alert.
func (c *TCPConfig) CloseNotify(enabled bool) *TCPConfig {
	c.OmitCloseNotify = !enabled
	return c
}
