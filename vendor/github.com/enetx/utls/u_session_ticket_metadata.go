package tls

import (
	"bytes"
	"encoding/binary"
	"time"
)

// TLS12TicketMetadata contains facts recorded for a TLS 1.0–1.2 session ticket.
// The engine records these only after verifying the handshake. It does not use
// them to change ticket validity, cache policy or the TLS 1.3 UseBy contract.
// Metadata survives SessionState.Bytes and the standard library's Extra bridge.
type TLS12TicketMetadata struct {
	// SessionCreatedAt is the original full handshake's ServerHello time.
	SessionCreatedAt time.Time
	// SessionAuthenticatedAt is the original full handshake's completion time.
	SessionAuthenticatedAt time.Time
	// TicketReceivedAt is when this handshake received NewSessionTicket.
	TicketReceivedAt time.Time
	// PreviousTicketReceivedAt is the preceding ticket's receipt time, when known.
	PreviousTicketReceivedAt time.Time
	// LifetimeHint is the server's unmodified lifetime hint, in seconds. Zero
	// remains a present zero hint; it does not itself mean immediate expiration.
	LifetimeHint uint32
	// Resumed reports whether the server actually resumed this handshake.
	Resumed bool
	// OriginalSessionKnown reports proven original creation/completion times.
	// Imported legacy sessions without this metadata do not gain a fresh age.
	OriginalSessionKnown bool
	// PreviousTicketReceivedAtKnown distinguishes an unavailable predecessor
	// timestamp from a recorded time. It can be true on legacy renewal lineage
	// whose original full-handshake times remain unknown.
	PreviousTicketReceivedAtKnown bool
}

const (
	tls12TicketMetadataPrefix  = "enetx/tls12-ticket\x00"
	tls12TicketMetadataVersion = 1
	tls12TicketMetadataSize    = len(tls12TicketMetadataPrefix) + 2 + 4 + 4*12
	tls12MetadataOriginal      = 1
	tls12MetadataResumed       = 2
	tls12MetadataPrevious      = 4
)

// TLS12TicketMetadata returns recorded legacy-ticket facts. False means absent,
// malformed, duplicate or unsupported metadata, or a non-legacy session. The
// returned value is independent of the session's mutable Extra byte slices.
// A caller implementing age limits must not infer original age when
// OriginalSessionKnown is false. Application Extra records remain opaque.
func (css *ClientSessionState) TLS12TicketMetadata() (TLS12TicketMetadata, bool) {
	if css == nil || css.session == nil || css.session.version < VersionTLS10 || css.session.version > VersionTLS12 {
		return TLS12TicketMetadata{}, false
	}

	var result TLS12TicketMetadata
	found := false
	for _, extra := range css.session.Extra {
		if !bytes.HasPrefix(extra, []byte(tls12TicketMetadataPrefix)) {
			continue
		}

		if found {
			return TLS12TicketMetadata{}, false
		}

		var ok bool
		result, ok = parseTLS12TicketMetadata(extra)
		if !ok {
			return TLS12TicketMetadata{}, false
		}

		found = true
	}

	return result, found
}

func validTLS12TicketMetadata(info TLS12TicketMetadata) bool {
	if info.TicketReceivedAt.IsZero() ||
		!info.Resumed && (!info.OriginalSessionKnown || info.PreviousTicketReceivedAtKnown) ||
		info.Resumed && info.OriginalSessionKnown && !info.PreviousTicketReceivedAtKnown {
		return false
	}

	if info.OriginalSessionKnown != (!info.SessionCreatedAt.IsZero() && !info.SessionAuthenticatedAt.IsZero()) ||
		!info.OriginalSessionKnown && (!info.SessionCreatedAt.IsZero() || !info.SessionAuthenticatedAt.IsZero()) ||
		info.PreviousTicketReceivedAtKnown != !info.PreviousTicketReceivedAt.IsZero() {
		return false
	}

	for _, stamp := range []time.Time{
		info.SessionCreatedAt, info.SessionAuthenticatedAt,
		info.TicketReceivedAt, info.PreviousTicketReceivedAt,
	} {
		if year := stamp.UTC().Year(); year < 1 || year > 9999 {
			return false
		}
	}

	return true
}

func marshalTLS12TicketMetadata(info TLS12TicketMetadata) ([]byte, bool) {
	if !validTLS12TicketMetadata(info) {
		return nil, false
	}

	data := make([]byte, tls12TicketMetadataSize)
	copy(data, tls12TicketMetadataPrefix)
	position := len(tls12TicketMetadataPrefix)
	data[position] = tls12TicketMetadataVersion
	if info.OriginalSessionKnown {
		data[position+1] |= tls12MetadataOriginal
	}

	if info.Resumed {
		data[position+1] |= tls12MetadataResumed
	}

	if info.PreviousTicketReceivedAtKnown {
		data[position+1] |= tls12MetadataPrevious
	}

	position += 2
	binary.BigEndian.PutUint32(data[position:], info.LifetimeHint)
	position += 4
	for _, stamp := range []time.Time{
		info.SessionCreatedAt, info.SessionAuthenticatedAt,
		info.TicketReceivedAt, info.PreviousTicketReceivedAt,
	} {
		binary.BigEndian.PutUint64(data[position:], uint64(stamp.Unix()))
		binary.BigEndian.PutUint32(data[position+8:], uint32(stamp.Nanosecond()))
		position += 12
	}

	return data, true
}

func parseTLS12TicketMetadata(data []byte) (TLS12TicketMetadata, bool) {
	position := len(tls12TicketMetadataPrefix)
	if len(data) != tls12TicketMetadataSize || !bytes.HasPrefix(data, []byte(tls12TicketMetadataPrefix)) ||
		data[position] != tls12TicketMetadataVersion || data[position+1]&^byte(7) != 0 {
		return TLS12TicketMetadata{}, false
	}

	info := TLS12TicketMetadata{
		OriginalSessionKnown:          data[position+1]&tls12MetadataOriginal != 0,
		Resumed:                       data[position+1]&tls12MetadataResumed != 0,
		PreviousTicketReceivedAtKnown: data[position+1]&tls12MetadataPrevious != 0,
		LifetimeHint:                  binary.BigEndian.Uint32(data[position+2:]),
	}
	position += 6
	for _, stamp := range []*time.Time{
		&info.SessionCreatedAt, &info.SessionAuthenticatedAt,
		&info.TicketReceivedAt, &info.PreviousTicketReceivedAt,
	} {
		seconds := int64(binary.BigEndian.Uint64(data[position:]))
		nanos := binary.BigEndian.Uint32(data[position+8:])
		if seconds < -62135596800 || seconds > 253402300799 || nanos >= 1_000_000_000 {
			return TLS12TicketMetadata{}, false
		}

		*stamp = time.Unix(seconds, int64(nanos)).UTC()
		position += 12
	}

	if !validTLS12TicketMetadata(info) {
		return TLS12TicketMetadata{}, false
	}

	return info, true
}

func (hs *clientHandshakeState) recordTLS12TicketMetadata(session *SessionState) {
	info := hs.ticketMetadata
	info.Resumed = hs.c.didResume
	if info.Resumed {
		info.SessionCreatedAt, info.SessionAuthenticatedAt = time.Time{}, time.Time{}
		if previous, ok := (&ClientSessionState{session: hs.session}).TLS12TicketMetadata(); ok {
			info.OriginalSessionKnown = previous.OriginalSessionKnown
			info.SessionCreatedAt, info.SessionAuthenticatedAt = previous.SessionCreatedAt, previous.SessionAuthenticatedAt
			info.PreviousTicketReceivedAt = previous.TicketReceivedAt
			info.PreviousTicketReceivedAtKnown = true
		}
	} else {
		info.OriginalSessionKnown = true
		info.SessionAuthenticatedAt = hs.c.config.time()
	}

	if encoded, ok := marshalTLS12TicketMetadata(info); ok {
		// The destination is the engine's fresh session. Preserve its Extra
		// and carry only proven owned lineage, not the predecessor's app data.
		session.Extra = append(session.Extra, encoded)
	}
}
