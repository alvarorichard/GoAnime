package desktop

import (
	"time"

	"github.com/enetx/surf/pkg/tlscache"
	"github.com/enetx/surf/profiles"
)

// ConfigureSessionCache follows SSLTokensCache's ten-token expiry-ordered queue.
// Tokens, including TLS 1.2 external tokens, are removed when retrieved.
// The 8 MiB budget charges each key and compressed record. Go and NSS serialize
// different state, so exact byte sizes differ. The key ceiling additionally
// bounds Surf's consumed-ticket history, which NSS does not need.
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/base/SSLTokensCache.cpp
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/modules/libpref/init/StaticPrefList.yaml
func ConfigureSessionCache(cache *profiles.SessionCacheConfig) {
	cache.Capacity(1024).TicketsPerKey(10).ByteCapacity(8 << 20).Brotli(1).
		EvictionPolicy(tlscache.EvictEarliestExpiry).
		Order(tlscache.EarliestExpiryFirst).TLS12TicketPolicy(TLS12TicketPolicy)
}

// TLS12TicketPolicy follows NSS's authenticated-completion session epoch and
// independent nonzero ticket hint. Resumed external TLS 1.2 tokens are not
// exported again; a server-declined resumption starts a new full session.
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/security/nss/lib/ssl/ssl3con.c#L12787
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/security/nss/lib/ssl/sslnonce.c#L796
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/base/SSLTokensCache.cpp#L953
func TLS12TicketPolicy(info tlscache.TLS12TicketMetadata) tlscache.TLS12TicketValidity {
	if !info.OriginalSessionKnown || info.Resumed {
		return tlscache.TLS12TicketValidity{}
	}

	validity := tlscache.TLS12TicketValidity{Accept: true,
		SessionExpiresAt:       info.SessionAuthenticatedAt.Truncate(time.Microsecond).Add(48 * time.Hour),
		DeferSessionValidation: true}
	if info.LifetimeHint != 0 {
		validity.OfferExpiresAt = info.TicketReceivedAt.Truncate(time.Microsecond).
			Add(time.Duration(info.LifetimeHint) * time.Second)
	}

	return validity
}
