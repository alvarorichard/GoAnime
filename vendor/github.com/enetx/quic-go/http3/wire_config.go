package http3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"sync"
	"sync/atomic"

	"github.com/enetx/http"
	httpcommon "github.com/enetx/http/headerorder"
	"github.com/enetx/quic-go"
	"github.com/enetx/quic-go/http3/qlog"
	"github.com/enetx/quic-go/qlogwriter"
	"github.com/enetx/quic-go/quicvarint"
	"github.com/quic-go/qpack"
)

// GreaseFrame is an RFC 9114 reserved frame (type 0x1f*N+0x21).
// Browser-specific types, payloads and random distributions belong to the caller.
type GreaseFrame struct {
	Type    uint64
	Payload []byte
}

// RequestPreparationError reports a local peer-settings, header preparation,
// frame factory or control queue failure before request bytes are written.
// It unwraps the original cause, but should
// not trigger protocol fallback or a retry on a different connection.
type RequestPreparationError struct{ err error }

func (e *RequestPreparationError) Error() string { return "http3: preparing request: " + e.err.Error() }
func (e *RequestPreparationError) Unwrap() error { return e.err }

// RequestBodyReadInfo describes the validated outgoing request before QPACK
// state changes. HeaderFields contains the actual wire fields, including
// implicit headers. It is a private copy and may be retained by the callback.
// Request must not be mutated. BufferSize is the effective body read buffer.
type RequestBodyReadInfo struct {
	Request      *http.Request
	HeaderFields []qpack.HeaderField
	BufferSize   int
}

// WireConfig adds optional client wire behavior. Nil preserves upstream behavior.
// Factories may run concurrently on different connections and must not mutate
// requests or reenter the transport. Returned payloads are copied before use.
type WireConfig struct {
	// OnControlFrame handles otherwise unknown frames after the peer's SETTINGS.
	// The reader is limited to this frame and valid only during the callback.
	// Unread bytes are drained; a truncated payload or callback error closes the
	// connection with H3_FRAME_ERROR. Calls are serialized per control stream.
	// It must not reenter this connection. Nil keeps native extension handling.
	OnControlFrame func(*quic.Conn, ControlFrame) error
	// OnResponseHeaders runs after each valid response header block, including
	// informational responses, before tracing or delivery to the application.
	// An error aborts the request. The callback must return promptly, must not
	// mutate the request or response or read its body, and must be safe for
	// concurrent requests. Nil leaves response handling unchanged.
	OnResponseHeaders func(*http.Request, *http.Response) error
	// CoalesceInitialSettings stages control-stream data before the first
	// application flight when peer credit allows. With early keys, setup runs
	// before Initial packetization unless FlushEarlySettings is enabled.
	// Backpressure keeps asynchronous writes on the same stream. Factories must not block.
	CoalesceInitialSettings bool
	// FlushEarlySettings stages initial control data and attempts to flush it
	// before releasing a resumed early connection to requests. The Initial flight
	// gets its own send opportunity first. Congestion and pacing can defer either
	// write; packet counts and delivery are not fixed. Does not enable early data.
	FlushEarlySettings bool
	// CoalesceRequestHeaders stages initial priority, ready QPACK encoder
	// instructions, request HEADERS and a bodyless request's FIN in one
	// connection-loop turn. Congestion and packet size limits still apply.
	// Busy or blocked critical streams retain their asynchronous owner.
	CoalesceRequestHeaders bool
	// RequestBodyBufferSize controls reads from request bodies. Zero uses 8 KiB.
	RequestBodyBufferSize int
	// RequestBodyBufferSizePolicy overrides the fixed size for a request body.
	// It receives ContentLength (-1 if unknown) and must be concurrency safe.
	// Returning zero selects the default. Valid sizes are at most 16 MiB.
	RequestBodyBufferSizePolicy func(contentLength int64) int
	// RequestBodyFirstReadSizePolicy caps the first nonempty body read. Zero
	// uses BufferSize. Results must be between zero and BufferSize. It runs
	// before request bytes or QPACK state changes and must be concurrency safe.
	RequestBodyFirstReadSizePolicy func(RequestBodyReadInfo) int
	// MaxDataFrameSize limits request DATA payloads. Zero preserves one frame
	// per Write. This does not control QUIC packet boundaries.
	MaxDataFrameSize int

	// SettingsGrease generates one reserved SETTINGS identifier/value per
	// connection, before any control-stream bytes. Nil adds nothing. The callback
	// must not block or reenter the transport. IDs must be 0x1f*N+0x21 and both
	// values must fit QUIC integers. It cannot replace existing settings.
	SettingsGrease func() (uint64, uint64, error)

	// ControlGreaseFrames runs once per connection, after SETTINGS.
	ControlGreaseFrames func() ([]GreaseFrame, error)
	// RequestGreaseFrames runs before each request's HEADERS, on its stream.
	RequestGreaseFrames func(*http.Request) ([]GreaseFrame, error)
	// SendPriorityUpdates sends the request's explicit Priority field on the
	// control stream when it differs from RFC 9218's default (u=3, i=?0).
	// It never invents a Priority header and is independent of GREASE.
	// Recognized urgency/incremental parameters are serialized canonically.
	SendPriorityUpdates bool
	// CompactPriorityUpdates omits the space after commas in RFC 9218 fields.
	CompactPriorityUpdates bool
}

const maxRequestBodyBufferSize = 16 << 20

func validateBodySize(name string, size int) error {
	if size < 0 || size > maxRequestBodyBufferSize {
		return fmt.Errorf("http3: %s must be between 0 and %d", name, maxRequestBodyBufferSize)
	}
	return nil
}

// Validate checks allocation and frame size limits without calling factories.
func (c *WireConfig) Validate() error {
	if c == nil {
		return nil
	}
	if err := validateBodySize("request body buffer size", c.RequestBodyBufferSize); err != nil {
		return err
	}
	return validateBodySize("maximum DATA frame size", c.MaxDataFrameSize)
}

func (c *WireConfig) Clone() *WireConfig {
	if c == nil {
		return nil
	}
	copy := *c
	return &copy
}

func appendGreaseFrames(frames []GreaseFrame) ([]byte, error) {
	var out []byte
	for _, frame := range frames {
		if frame.Type < 0x21 || frame.Type > quicvarint.Max || (frame.Type-0x21)%0x1f != 0 {
			return nil, fmt.Errorf("http3: invalid GREASE frame type %d", frame.Type)
		}
		out = quicvarint.Append(out, frame.Type)
		out = quicvarint.Append(out, uint64(len(frame.Payload)))
		out = append(out, frame.Payload...)
	}
	return out, nil
}

// A single bounded writer owns the send control stream. Backpressure on that
// stream never blocks request HEADERS indefinitely or creates per-request goroutines.
type clientControlWriter struct {
	owner   *rawConn
	queue   chan controlWrite
	mu      sync.Mutex
	stream  *quic.SendStream
	pending atomic.Int32
	initial []byte
}

type controlWrite struct {
	data     []byte
	recorded bool
}

func (w *clientControlWriter) enqueue(data []byte) error {
	return w.enqueueWrite(controlWrite{data: data})
}

func (w *clientControlWriter) enqueueWrite(data controlWrite) error {
	select {
	case <-w.owner.context().Done():
		return context.Cause(w.owner.context())
	default:
	}
	w.pending.Add(1)
	select {
	case w.queue <- data:
		return nil
	default:
		w.pending.Add(-1)
		return errors.New("http3: outgoing control frame queue is full")
	}
}

func (w *clientControlWriter) run(stream *quic.SendStream) {
	if w.owner.qlogger != nil {
		if !w.owner.AddQlogProducer() {
			return
		}
		defer w.owner.RemoveQlogProducer()
	}
	write := func(data []byte, recorded bool) bool {
		if len(data) == 0 {
			return true
		}
		w.mu.Lock()
		if !recorded {
			recordWireFrames(w.owner.qlogger, stream.StreamID(), data)
		}
		n, err := stream.Write(data)
		w.mu.Unlock()
		if err == nil && n != len(data) {
			err = io.ErrShortWrite
		}
		if err != nil {
			if !errors.Is(err, quic.Err0RTTRejected) && w.owner.context().Err() == nil {
				w.owner.CloseWithError(quic.ApplicationErrorCode(ErrCodeClosedCriticalStream), "control stream write failed")
			}
			return false
		}
		return true
	}
	if len(w.initial) > 0 {
		ok := write(w.initial, false)
		w.pending.Add(-1)
		if !ok {
			return
		}
	}
	for {
		select {
		case <-w.owner.context().Done():
			return
		case <-stream.Context().Done():
			// Rejection retires the early stream generation before the raw
			// connection's accept loop necessarily publishes the same cause.
			if !errors.Is(context.Cause(stream.Context()), quic.Err0RTTRejected) &&
				w.owner.context().Err() == nil {
				w.owner.CloseWithError(quic.ApplicationErrorCode(ErrCodeClosedCriticalStream), "control stream closed")
			}
			return
		case data := <-w.queue:
			ok := write(data.data, data.recorded)
			w.pending.Add(-1)
			if !ok {
				return
			}
		}
	}
}

func recordWireFrames(logger qlogwriter.Recorder, streamID quic.StreamID, data []byte) {
	if logger == nil {
		return
	}
	for len(data) > 0 {
		kind, n, err := quicvarint.Parse(data)
		if err != nil {
			return
		}
		length, m, err := quicvarint.Parse(data[n:])
		if err != nil || length > uint64(len(data)-n-m) {
			return
		}
		size := n + m + int(length)
		var frame any = qlog.ReservedFrame{Type: kind}
		if kind == 0xf0700 {
			id, k, err := quicvarint.Parse(data[n+m : size])
			if err != nil {
				return
			}
			frame = qlog.PriorityUpdateFrame{StreamID: quic.StreamID(id), PriorityFieldValue: string(data[n+m+k : size])}
		}
		logger.RecordEvent(qlog.FrameCreated{StreamID: streamID, Raw: qlog.RawInfo{Length: size, PayloadLength: int(length)}, Frame: qlog.Frame{Frame: frame}})
		data = data[size:]
	}
}

func (c *ClientConn) configureWire(config *WireConfig) {
	if config == nil {
		config = &WireConfig{}
	}
	c.wireConfig = config.Clone()
	// Install the control writer before the startup goroutine can send SETTINGS.
	c.controlWriter = &clientControlWriter{owner: c.rawConn, queue: make(chan controlWrite, 64)}
	c.requestWriter.greaseFrames = config.RequestGreaseFrames
	if config.CoalesceRequestHeaders {
		c.requestWriter.writeRequestHeaders = c.writeRequestHeaders
		if c.rawConn.clientQPACK != nil {
			c.requestWriter.prepareRequestHeaders = c.prepareQPACKRequestHeaders
		}
	}
}

func (c *ClientConn) queuePriorityUpdate(req *http.Request, id quic.StreamID) error {
	if c.wireConfig == nil || !c.wireConfig.SendPriorityUpdates || c.wireConfig.CoalesceRequestHeaders {
		return nil
	}
	u, incremental, explicit := requestPriority(req)
	if !explicit || u == defaultPriorityUrgency && !incremental {
		return nil
	}
	return c.enqueuePriorityUpdate(id, u, incremental)
}

func requestPriority(req *http.Request) (int8, bool, bool) {
	var values []string
	for _, field := range httpcommon.Sorted(req.Header, httpcommon.Indices(req.Header[httpcommon.HeaderOrderKey]), nil) {
		if strings.EqualFold(field.Key, "priority") {
			values = append(values, field.Values...)
		}
	}
	value := strings.Join(values, ", ")
	u, incremental := parsePriority(value)
	return u, incremental, value != ""
}

func (c *ClientConn) enqueuePriorityUpdate(id quic.StreamID, urgency int8, incremental bool) error {
	return c.controlWriter.enqueue(c.priorityUpdateFrame(id, urgency, incremental))
}

func (c *ClientConn) priorityUpdateFrame(id quic.StreamID, urgency int8, incremental bool) []byte {
	value := ""
	if urgency != defaultPriorityUrgency {
		value = "u=" + strconv.Itoa(int(urgency))
	}
	if incremental {
		if value != "" {
			value += ","
			if !c.wireConfig.CompactPriorityUpdates {
				value += " "
			}
		}
		value += "i"
	}
	payload := quicvarint.Append(nil, uint64(id))
	payload = append(payload, value...)
	data := quicvarint.Append(nil, 0xf0700)
	data = quicvarint.Append(data, uint64(len(payload)))
	return append(data, payload...)
}
