package profiles

import (
	"net/netip"
	"time"
)

// NetworkChangeInfo describes a native notification or an application event.
// TemporaryIPv6Only is set only when every changed address is known temporary.
type NetworkChangeInfo struct {
	AddressChanged    bool // A non-loopback interface gained or lost an address.
	LinkChanged       bool // An interface's up/running state changed.
	RouteChanged      bool // A default route changed without necessarily changing an address.
	IdentityChanged   bool // The configured native network identity changed after coalescing.
	IPv6Only          bool // Only IPv6 addresses changed; used by platform notification policy.
	TemporaryIPv6Only bool // No stable IPv4/IPv6 address changed.
}

// NetworkChangeAction selects maintenance on existing transport-owned pools.
// Zero preserves connections and metadata. TLS tickets and QUIC RTT/version
// history are never cleared by these actions.
type NetworkChangeAction struct {
	HTTP2GoAwayOnClose      bool          // Send NO_ERROR GOAWAY only after the final H2 stream ends.
	CloseHTTP2              bool          // Interrupt existing H2 sessions; takes precedence over RetireHTTP2.
	RetireHTTP2             bool          // Prevent H2 reuse while active responses finish.
	RetireHTTP3             bool          // Prevent H3 reuse and discard coalescing authorizations.
	RetireHTTP1             bool          // Close idle H1 sockets and retire active sockets after their exchange.
	ClearFailedAlternatives bool          // Clear H3 failure exclusions, retaining advertisements and ECH.
	HTTP1LivenessTimeout    time.Duration // Close current active H1 sockets with no read/write progress; zero disables.
	HTTP2IdleProbeAfter     time.Duration // Probe H2 before retirement after this much socket inactivity; zero disables.
	HTTP2ProbePayload       *[8]byte      // Optional copied PING payload; nil selects the native random payload.
	HTTP2ProbeTimeout       time.Duration // Deadline for the H2 probe; an expired probe closes its connection.
}

// NetworkChangePolicy selects maintenance for one notification. It must return
// promptly, be concurrency safe and not reenter the client being maintained.
type NetworkChangePolicy func(NetworkChangeInfo) NetworkChangeAction

// NetworkChangeConfig is the concrete fluent profile configuration. Surf's
// NetworkSettings uses the same controls and applies application overrides last.
type NetworkChangeConfig struct {
	Enabled  bool                // Subscribe to native notifications when Policy is non-nil.
	Delay    time.Duration       // Coalescing window from the first native event.
	OnChange NetworkChangePolicy // Pure policy copied into the built client.
	// ReadIdentity returns an opaque local network identity. A nil neighbor list
	// requests a complete refresh; otherwise it lists neighbors that acquired a
	// previously unknown MAC. Readers can ignore unrelated neighbors. The first
	// successful result establishes a baseline; errors retain the previous one.
	// The reader runs serially before Policy, must return promptly and must not
	// reenter the client. Nil disables identity observation.
	ReadIdentity func([]netip.Addr) (string, error)
}

// Automatic enables native notifications. Manual NetworkChanged remains usable.
func (c *NetworkChangeConfig) Automatic(enabled bool) *NetworkChangeConfig {
	c.Enabled = enabled
	return c
}

// CoalescingDelay merges native events from the first event without extending
// the window. Zero delivers immediately; the supported range is zero to a minute.
func (c *NetworkChangeConfig) CoalescingDelay(delay time.Duration) *NetworkChangeConfig {
	c.Delay = delay
	return c
}

// Policy selects per-event maintenance. Nil disables automatic and manual work.
func (c *NetworkChangeConfig) Policy(policy NetworkChangePolicy) *NetworkChangeConfig {
	c.OnChange = policy
	return c
}

// Identity selects a local identity reader for native notifications. The result
// is compared only locally; Surf does not send it to peers or clear TLS tickets.
func (c *NetworkChangeConfig) Identity(reader func([]netip.Addr) (string, error)) *NetworkChangeConfig {
	c.ReadIdentity = reader
	return c
}
