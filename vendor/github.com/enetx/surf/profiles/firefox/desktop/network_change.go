package desktop

import (
	"runtime"
	"time"

	"github.com/enetx/surf/profiles"
)

// ConfigureNetworkChange follows Firefox's notification coalescing and pool
// maintenance. TLS tokens are independent of network-link notifications.
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/system/netlink/NetlinkService.cpp
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/protocol/http/nsHttpHandler.cpp
func ConfigureNetworkChange(config *profiles.NetworkChangeConfig) {
	config.Automatic(true).CoalescingDelay(time.Second).Policy(NetworkChange).Identity(networkIdentityReader())
}

// NetworkChange retires multiplexed sessions, clears H3 exclusions and verifies
// active HTTP/1 traffic. Healthy idle HTTP/1 sockets remain reusable.
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/protocol/http/ConnectionEntry.cpp
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/protocol/http/nsHttpConnection.cpp
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/modules/libpref/init/StaticPrefList.yaml
func NetworkChange(info profiles.NetworkChangeInfo) profiles.NetworkChangeAction {
	// Windows defaults network.notify.IPv6 to false; link/route notifications
	// remain independently relevant. The host OS controls native notifications.
	if runtime.GOOS == "windows" && info.IPv6Only && !info.LinkChanged && !info.RouteChanged && !info.IdentityChanged {
		return profiles.NetworkChangeAction{}
	}

	if !info.AddressChanged && !info.LinkChanged && !info.RouteChanged && !info.IdentityChanged {
		return profiles.NetworkChangeAction{}
	}

	return profiles.NetworkChangeAction{
		HTTP2GoAwayOnClose: true,
		RetireHTTP2:        true, RetireHTTP3: true, ClearFailedAlternatives: true,
		HTTP1LivenessTimeout: 5 * time.Second,
		HTTP2IdleProbeAfter:  500 * time.Millisecond,
		HTTP2ProbePayload:    &[8]byte{},
		HTTP2ProbeTimeout:    8 * time.Second,
	}
}
