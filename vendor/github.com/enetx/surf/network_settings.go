package surf

import (
	"net/netip"
	"time"

	"github.com/enetx/surf/profiles"
)

// NetworkChangeInfo describes an address, link, route or local identity change.
type NetworkChangeInfo = profiles.NetworkChangeInfo

// NetworkChangeAction selects maintenance on existing connection pools.
type NetworkChangeAction = profiles.NetworkChangeAction

// NetworkChangePolicy selects pool maintenance for a network notification.
type NetworkChangePolicy = profiles.NetworkChangePolicy

// NetworkSettings configures native network notifications through the client
// builder. Browser defaults apply first; explicit overrides apply last.
type NetworkSettings struct {
	builder   *Builder
	overrides []func(*profiles.NetworkChangeConfig)
}

// NetworkSettings starts the nested network-notification builder. Plain clients
// have no automatic subscription; browser profiles provide their own policy.
func (b *Builder) NetworkSettings() *NetworkSettings {
	if b.networkSettings == nil {
		b.networkSettings = &NetworkSettings{builder: b}
	}
	return b.networkSettings
}

// Automatic enables native notifications. False leaves explicit NetworkChanged
// calls available, for applications that already own network notifications.
func (n *NetworkSettings) Automatic(enabled bool) *NetworkSettings {
	n.overrides = append(n.overrides, func(c *profiles.NetworkChangeConfig) { c.Automatic(enabled) })
	return n
}

// CoalescingDelay merges native notifications from the first event without
// extending the window. Valid values are zero through one minute.
func (n *NetworkSettings) CoalescingDelay(delay time.Duration) *NetworkSettings {
	n.overrides = append(n.overrides, func(c *profiles.NetworkChangeConfig) { c.CoalescingDelay(delay) })
	return n
}

// Policy selects maintenance for each event. Nil disables maintenance. The
// callback must return promptly, be concurrency safe and not reenter this client.
func (n *NetworkSettings) Policy(policy NetworkChangePolicy) *NetworkSettings {
	n.overrides = append(n.overrides, func(c *profiles.NetworkChangeConfig) { c.Policy(policy) })
	return n
}

// Identity sets the native network identity reader. Nil disables it. A nil input
// requests a full refresh; other inputs identify neighbors acquiring a MAC.
// The first successful value is a baseline; later changes set IdentityChanged.
// Errors retain that baseline. The callback must return promptly, be safe across
// clients and not reenter the client. Identity values remain local.
func (n *NetworkSettings) Identity(reader func([]netip.Addr) (string, error)) *NetworkSettings {
	n.overrides = append(n.overrides, func(c *profiles.NetworkChangeConfig) { c.Identity(reader) })
	return n
}

// Set returns to the client builder.
func (n *NetworkSettings) Set() *Builder { return n.builder }
