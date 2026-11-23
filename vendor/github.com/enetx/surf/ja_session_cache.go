package surf

import (
	"crypto/sha256"
	"fmt"

	utls "github.com/enetx/utls"
)

// Browser caches distinguish the logical TLS destination, including its port.
// The native engine's SNI-only key remains an additional partition, so an
// application ServerName override does not merge different destinations.
// Custom caches retain their existing key contract and ownership.
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/socket/ssl_client_socket_impl.cc#L1628
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/security/manager/ssl/NSSSocketControl.cpp#L657
func (rt *roundtripper) sessionCacheForOrigin(address string) utls.ClientSessionCache {
	if !rt.partitionSessionCache {
		return rt.clientSessionCache
	}

	scope := sha256.Sum256([]byte(tlsOrigin(address)))
	return tcpOriginSessionCache{cache: rt.clientSessionCache, prefix: fmt.Sprintf("origin:%x:", scope)}
}

type tcpOriginSessionCache struct {
	cache  utls.ClientSessionCache
	prefix string
}

func (c tcpOriginSessionCache) Get(key string) (*utls.ClientSessionState, bool) {
	return c.cache.Get(c.prefix + key)
}

func (c tcpOriginSessionCache) Put(key string, state *utls.ClientSessionState) {
	c.cache.Put(c.prefix+key, state)
}
