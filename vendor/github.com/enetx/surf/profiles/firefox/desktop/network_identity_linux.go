package desktop

import (
	"net/netip"

	"github.com/enetx/surf/internal/netchange"
)

// networkIdentityReader follows the Ethernet gateway/route-check part of
// Firefox's local network identification. The route lookups send no IP packets
// and do not resolve these addresses through the application's DNS provider.
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/system/NetworkLinkServiceDefines.h
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/system/netlink/NetlinkService.cpp
func networkIdentityReader() func([]netip.Addr) (string, error) {
	return netchange.GatewayIdentityReader(
		netip.MustParseAddr("23.219.91.27"),
		netip.MustParseAddr("2a02:26f0:40::17db:5b1b"),
	)
}
