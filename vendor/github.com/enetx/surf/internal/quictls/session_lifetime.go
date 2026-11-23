package quictls

import (
	"crypto/tls"
	"time"

	"github.com/enetx/surf/internal/tlsstate"
)

// SessionLifetime uses the public serialization bridge and requires TLS 1.3:
// QUIC session selection must not consume a TCP TLS 1.2 ticket.
func SessionLifetime(session *tls.ClientSessionState) (time.Time, time.Time, bool) {
	info, ok := tlsstate.SessionInfo(session)
	return info.From, info.Until, ok && info.Version == tls.VersionTLS13
}
