package profiles

import "time"

// ConnectionRacingInfo describes the current origin and its client's transport
// session. It does not include persisted browser history or OS network changes.
type ConnectionRacingInfo struct {
	// Origin is the HTTPS origin whose connections are being prepared.
	Origin string
	// QUICWorked reports a confirmed QUIC handshake for this origin.
	QUICWorked bool
	// QUICWorkedInClient reports a confirmed QUIC handshake to any origin in
	// this client's HTTP/3 transport. Independent clients do not share it;
	// closing idle connections retains it. It is not an OS network probe.
	QUICWorkedInClient bool
	// HTTP2Available reports a reusable session in this request's TCP route.
	// Stream admission can still queue, and a later GOAWAY can invalidate it.
	HTTP2Available bool
	// TCPConnectionGroup reports pool-owned TCP connections or pending dials
	// for this route. An occupied group does not promise an idle socket.
	TCPConnectionGroup bool
	// DNSOnly reports HTTP/3 discovery without an authenticated Alt-Svc route.
	// An Alt-Svc alternative retains its separate main-job hold policy.
	DNSOnly bool
	// SmoothedRTT is the measured QUIC RTT for this origin, when available.
	SmoothedRTT time.Duration
	// LastQUICError is the most recent non-cancellation preparation failure
	// for this origin. A later confirmed preparation clears it.
	LastQUICError error
}

// ConnectionRacingPolicy selects the TCP backup delay. It must return a
// nonnegative duration promptly and be safe for concurrent calls.
type ConnectionRacingPolicy func(ConnectionRacingInfo) time.Duration

// EarlyRequestPolicy admits an otherwise eligible bodyless GET/HEAD before
// authenticated TLS completes. It observes the built client's transport history;
// it does not disable TLS early-data offers or admit unsafe request bodies.
// Callbacks must return promptly and be safe for concurrent calls.
type EarlyRequestPolicy func(ConnectionRacingInfo) bool
