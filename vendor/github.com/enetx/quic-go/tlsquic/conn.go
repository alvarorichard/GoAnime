// Package tlsquic defines the crypto/tls event interface used by QUIC.
// A client adapter can implement it without changing the QUIC packet engine.
package tlsquic

import (
	"context"
	"crypto/tls"
)

// Conn follows crypto/tls.QUICConn's ownership and event-draining rules.
// Methods are called serially by the QUIC connection. Session events must preserve
// edits to Extra and EarlyData, including edits made before the next NextEvent.
type Conn interface {
	Start(context.Context) error
	NextEvent() tls.QUICEvent
	HandleData(tls.QUICEncryptionLevel, []byte) error
	SetTransportParameters([]byte)
	SendSessionTicket(tls.QUICSessionTicketOptions) error
	StoreSession(*tls.SessionState) error
	ConnectionState() tls.ConnectionState
	Close() error
}

// NewClient constructs a new TLS client for each connection attempt. config is a
// private, normalized TLS config with the destination ServerName and ALPN set.
// Adapters must preserve certificate verification and return errors explicitly.
// EnableSessionEvents must be enabled. The factory must not perform network I/O.
type NewClient func(config *tls.Config) (Conn, error)

var _ Conn = (*tls.QUICConn)(nil)
