package quic

import (
	"crypto/tls"
	"time"

	"github.com/enetx/quic-go/internal/handshake"
)

// SessionVersion returns the negotiated QUIC version recorded locally alongside
// a client session ticket. It does not validate ticket expiry or authenticity;
// the TLS engine retains all session validation. Unmarked/invalid states return
// false. The metadata uses SessionState.Extra and does not change ticket bytes.
func SessionVersion(session *tls.ClientSessionState) (Version, bool) {
	info, ok := ClientSessionMetadata(session)
	return info.Version, ok
}

// SessionMetadata describes locally remembered client transport state. RTT is
// zero without an actual measurement; it is never an advertised parameter.
// TransportParameters contains the serialized remembered 0-RTT limits and is
// copied, so a caller cannot mutate cached TLS metadata.
type SessionMetadata struct {
	Version Version
	RTT     time.Duration
	// ConnectionID identifies the Initial key context of the issuing connection.
	// It is local metadata and does not change the opaque server ticket.
	ConnectionID        []byte
	TransportParameters []byte
}

// ClientSessionMetadata reads QUIC's local SessionState.Extra entry. It does not
// validate TLS ticket lifetime, certificates or authenticity. TLS retains those
// checks. Earlier version-bearing metadata remains readable with a zero RTT.
func ClientSessionMetadata(session *tls.ClientSessionState) (SessionMetadata, bool) {
	if session == nil {
		return SessionMetadata{}, false
	}
	_, state, err := session.ResumptionState()
	if err != nil || state == nil {
		return SessionMetadata{}, false
	}
	info, ok := handshake.SessionInfo(state.Extra)
	return SessionMetadata{Version: info.Version, RTT: info.RTT, ConnectionID: info.ConnectionID, TransportParameters: info.TransportParameters}, ok
}
