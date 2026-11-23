package tlsstate

import (
	"crypto/sha256"
	"crypto/tls"
	"time"

	utls "github.com/enetx/utls"
)

// Info contains authenticated engine facts and certificate safety boundaries.
type Info struct {
	// Version is the decoded session's negotiated TLS version.
	Version uint16
	// From and Until bound certificate validity, additionally capped by TLS13 UseBy.
	From, Until time.Time
	// Identity hashes the separately serialized ticket identity.
	Identity [32]byte
	// TLS12Ticket contains engine-authenticated legacy ticket facts when present.
	TLS12Ticket utls.TLS12TicketMetadata
	// HasTLS12Ticket distinguishes missing or malformed legacy metadata.
	HasTLS12Ticket bool
}

// SessionInfo reads metadata through the public Go/uTLS serialization bridge.
// TLS retains validation of the ticket, certificates, ALPN and cryptography.
func SessionInfo(session *tls.ClientSessionState) (Info, bool) {
	if session == nil {
		return Info{}, false
	}
	ticket, state, err := session.ResumptionState()
	if err != nil || state == nil || len(ticket) == 0 {
		return Info{}, false
	}
	data, err := state.Bytes()
	if err != nil {
		return Info{}, false
	}
	decoded, err := utls.ParseSessionState(data)
	if err != nil {
		return Info{}, false
	}
	converted, err := utls.NewResumptionState(ticket, decoded)
	if err != nil || len(converted.ServerCertificates()) == 0 {
		return Info{}, false
	}
	certificate := converted.ServerCertificates()[0]
	until := certificate.NotAfter
	if converted.Vers() == tls.VersionTLS13 {
		if converted.UseBy() == 0 || converted.UseBy() > 1<<63-1 {
			return Info{}, false
		}
		if expiry := time.Unix(int64(converted.UseBy()), 0); expiry.Before(until) {
			until = expiry
		}
	}
	metadata, known := converted.TLS12TicketMetadata()
	return Info{Version: converted.Vers(), From: certificate.NotBefore, Until: until,
		Identity: sha256.Sum256(ticket), TLS12Ticket: metadata, HasTLS12Ticket: known}, true
}
