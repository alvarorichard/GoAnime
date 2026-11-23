package http3

import (
	"context"

	"github.com/enetx/quic-go"
)

// StreamPriority configures local QUIC scheduling of request stream data.
// Urgency is clipped to 0 through 7, with smaller values taking priority.
// Incremental selects round-robin scheduling within the urgency level.
// This calls quic.Stream.SetPriority; it does not send a PRIORITY_UPDATE frame
// or change the request's Priority header.
type StreamPriority struct {
	Urgency     int8
	Incremental bool
}

type streamPriorityKey struct{}

// WithStreamPriority overrides local QUIC scheduling for one HTTP/3 request.
// Pass the returned context to Request.WithContext. This takes precedence over
// Transport.StreamPriority and does not change HTTP headers or control frames.
func WithStreamPriority(ctx context.Context, priority StreamPriority) context.Context {
	return context.WithValue(ctx, streamPriorityKey{}, priority)
}

// SetPriority configures local scheduling of data sent on this request stream.
// It delegates to the official quic-go stream scheduler.
func (s *RequestStream) SetPriority(urgency int8, incremental bool) {
	s.str.SetPriority(urgency, incremental)
}

func (c *ClientConn) setStreamPriority(ctx context.Context, stream *quic.Stream) {
	if priority, ok := ctx.Value(streamPriorityKey{}).(StreamPriority); ok {
		stream.SetPriority(priority.Urgency, priority.Incremental)
	} else if priority := c.streamPriority; priority != nil {
		stream.SetPriority(priority.Urgency, priority.Incremental)
	}
}

func (t *Transport) newConfiguredClientConn(conn *quic.Conn) *ClientConn {
	c := newClientConnWithWire(conn, t.EnableDatagrams, t.AdditionalSettings,
		t.MaxResponseHeaderBytes, t.DisableCompression, t.Logger, t.WireConfig, t.QPACKConfig)
	c.prepareHeaders = t.PrepareHeaders
	if t.InitialPeerSettings != nil {
		c.rawConn.peerSettings = &clientPeerSettings{initial: t.InitialPeerSettings, ready: make(chan struct{})}
	}

	if t.StreamPriority != nil {
		priority := *t.StreamPriority
		c.streamPriority = &priority
	}
	return c
}
