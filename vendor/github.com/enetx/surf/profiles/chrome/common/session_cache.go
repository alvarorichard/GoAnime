package common

import (
	"time"

	"github.com/enetx/surf/pkg/tlscache"
	"github.com/enetx/surf/profiles"
)

// ConfigureTCPSessionCache follows SSLClientSessionCache::Entry::Push/Pop:
// two newest TLS 1.3 tickets, consumed once; TLS 1.2 remains reusable.
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/ssl/ssl_client_session_cache.cc
func ConfigureTCPSessionCache(cache *profiles.SessionCacheConfig) {
	cache.Capacity(1024).TicketsPerKey(2).
		Order(tlscache.NewestFirst).ReuseTLS12(true).
		DiscardOnExpiry(false).SkipExpired(false).
		TLS12TicketPolicy(TLS12TicketPolicy)
}

// TLS12TicketPolicy follows Chromium's one-hour original-session limit and
// BoringSSL's second-resolution rebase. The server's lifetime hint is metadata,
// not the effective timeout. A renewal never refreshes the original deadline.
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/socket/ssl_client_socket_impl.cc#L208
// https://github.com/google/boringssl/blob/ea9f3a34c3977144c45d272be81842fd891a0bdb/ssl/ssl_session.cc#L169
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/ssl/ssl_client_session_cache.cc#L166
func TLS12TicketPolicy(info tlscache.TLS12TicketMetadata) tlscache.TLS12TicketValidity {
	if !info.OriginalSessionKnown {
		return tlscache.TLS12TicketValidity{}
	}

	previous := info.SessionCreatedAt.Truncate(time.Second)
	if info.PreviousTicketReceivedAtKnown {
		previous = info.PreviousTicketReceivedAt.Truncate(time.Second)
	}

	received := info.TicketReceivedAt.Truncate(time.Second)
	expires := info.SessionCreatedAt.Truncate(time.Second).Add(time.Hour)
	if received.Before(previous) || !received.Before(expires) {
		// BoringSSL stores timeout zero after a rollback or elapsed timeout;
		// insertion replaces the old reusable ticket even in this case.
		expires = received
	}

	return tlscache.TLS12TicketValidity{Accept: true, SessionExpiresAt: expires,
		CacheNotBefore: received.Add(-time.Second), OfferNotBefore: received,
		DeferSessionValidation: true}
}
