package tlscache

import (
	"time"

	"github.com/enetx/surf/internal/tlsstate"
	utls "github.com/enetx/utls"
)

// TLS12TicketMetadata contains authenticated ticket issuance and session lineage.
// OriginalSessionKnown is false for legacy imports without proven original age.
type TLS12TicketMetadata = utls.TLS12TicketMetadata

// TLS12TicketValidity separates session ordering, cache retention and wire offers.
// Certificate validity always bounds storage and use independently of this policy.
type TLS12TicketValidity struct {
	// Accept allows insertion. A zero SessionExpiresAt still rejects the ticket.
	Accept bool
	// SessionExpiresAt is the session deadline and the queue/pressure ordering key.
	SessionExpiresAt time.Time
	// CacheNotBefore rejects cache lookup before this time; zero adds no bound.
	CacheNotBefore time.Time
	// OfferNotBefore prevents offering a retained ticket before this time.
	OfferNotBefore time.Time
	// OfferExpiresAt adds a wire-offer deadline, such as a nonzero lifetime hint.
	OfferExpiresAt time.Time
	// DeferSessionValidation retains future or expired sessions until lookup or
	// capacity eviction. Insertion also leaves other session-expired records in
	// place. Certificate safety and duplicate/reuse rules remain enforced.
	DeferSessionValidation bool
}

// TLS12TicketPolicy controls legacy tickets using authenticated engine facts.
// The callback runs once per insertion, outside locks, and must be concurrency
// safe. Its returned timestamps are frozen for the lifetime of that record.
// Missing/malformed metadata is a miss without calling the policy. Nil restores
// generic legacy caching. TLS 1.3 and application-supplied caches are unaffected.
func (b *Builder) TLS12TicketPolicy(policy func(TLS12TicketMetadata) TLS12TicketValidity) *Builder {
	b.tls12TicketPolicy = policy

	return b
}

// validity computes separate certificate, session and offer bounds before locks.
func (c *Cache) validity(info tlsstate.Info, now time.Time) (ticket, bool) {
	record := ticket{version: info.Version, from: info.From, until: info.Until,
		certificateUntil: info.Until, rank: info.Until, id: info.Identity}
	if now.Before(info.From) || !now.Before(info.Until) {
		return ticket{}, false
	}

	if info.Version >= utls.VersionTLS13 || c.policy.tls12TicketPolicy == nil {
		return record, true
	}

	if !info.HasTLS12Ticket {
		return ticket{}, false
	}

	validity := c.policy.tls12TicketPolicy(info.TLS12Ticket)
	if !validity.Accept || validity.SessionExpiresAt.IsZero() {
		return ticket{}, false
	}

	record.rank = validity.SessionExpiresAt
	if validity.SessionExpiresAt.Before(record.until) {
		record.until = validity.SessionExpiresAt
	}

	if validity.CacheNotBefore.After(record.from) {
		record.from = validity.CacheNotBefore
	}

	record.offerFrom, record.offerUntil = validity.OfferNotBefore, validity.OfferExpiresAt
	record.deferSessionValidation = validity.DeferSessionValidation
	if !record.deferSessionValidation && (now.Before(record.from) || !now.Before(record.until)) {
		return ticket{}, false
	}

	return record, true
}
