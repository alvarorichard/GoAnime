package http3

import (
	"context"
	"errors"
	"fmt"
	"io"
	"sync"

	"github.com/enetx/g"
	"github.com/enetx/quic-go"
	dynamic "github.com/enetx/quic-go/internal/qpack"
	"github.com/enetx/quic-go/quicvarint"
	"github.com/quic-go/qpack"
)

type fieldDecoder interface{ Decode([]byte) qpack.DecodeFunc }

var errQPACKFeedbackBacklog = errors.New("http3: QPACK feedback backlog exceeded")

func qpackFeedbackQueueSize(capacity, blocked uint64) int {
	return int(min(uint64(65536), max(uint64(32), capacity/32+min(blocked, uint64(65536)))))
}

// Enable decoding exactly to the capacity and blocking allowance advertised by
// the application. A zero capacity retains the upstream static fast path.
func (c *rawConn) configureQPACK(settings g.MapOrd[uint64, uint64]) {
	capacity := settings.Get(1).UnwrapOrDefault()
	if capacity == 0 {
		return
	}
	blocked := settings.Get(7).UnwrapOrDefault()
	// Each instruction is at most ten bytes. Bound queued feedback independently
	// from application concurrency and from the peer's ability to read its stream.
	var emit func([]byte) error
	if c.clientQPACK != nil {
		emit = c.clientQPACK.decoder.enqueue
	} else {
		feedback := make(chan []byte, qpackFeedbackQueueSize(capacity, blocked))
		emit = func(data []byte) error {
			return c.enqueueQPACKFeedback(feedback, data)
		}
		go c.writeQPACKFeedback(feedback)
	}
	c.qpackDecoder = dynamic.NewDecoder(capacity, blocked, emit)
	context.AfterFunc(c.context(), func() { c.qpackDecoder.Close(context.Cause(c.context())) })
}

func (c *rawConn) enqueueQPACKFeedback(feedback chan<- []byte, data []byte) error {
	select {
	case feedback <- data:
		return nil
	case <-c.context().Done():
		return context.Cause(c.context())
	default:
		return errQPACKFeedbackBacklog
	}
}

func (c *rawConn) writeQPACKFeedback(feedback <-chan []byte) {
	var stream *quic.SendStream
	for {
		var data []byte
		select {
		case <-c.context().Done():
			return
		case data = <-feedback:
		}
		parts := [][]byte{data}
		if stream == nil {
			var err error
			stream, err = c.conn.OpenUniStreamSync(c.context())
			if err != nil {
				return
			}
			c.watchQPACKSendStream(stream)
			parts = [][]byte{quicvarint.Append(nil, streamTypeQPACKDecoderStream), data}
		}
		// Type and feedback retain their original separate writes. Both writes
		// have the same critical-stream and rejected-generation semantics.
		for _, part := range parts {
			if _, err := stream.Write(part); err != nil {
				if errors.Is(err, quic.Err0RTTRejected) {
					return
				}
				c.CloseWithError(quic.ApplicationErrorCode(ErrCodeClosedCriticalStream), "QPACK decoder stream write failed")
				return
			}
		}
	}
}

// Both reserved and lazy QPACK send streams are critical even while their
// writers are idle. Rejected 0-RTT generations and connection shutdown retain
// their own lifecycle; only a peer's stream cancellation closes the connection.
// https://www.rfc-editor.org/rfc/rfc9204.html#section-4.2
func (c *rawConn) watchQPACKSendStream(stream *quic.SendStream) {
	context.AfterFunc(stream.Context(), func() {
		_, stopped := errors.AsType[*quic.StreamError](context.Cause(stream.Context()))
		if stopped && c.context().Err() == nil {
			c.CloseWithError(quic.ApplicationErrorCode(ErrCodeClosedCriticalStream), "QPACK send stream closed")
		}
	})
}

func (c *rawConn) handleQPACKStream(stream *quic.ReceiveStream, encoder bool) {
	var err error
	code := ErrCodeQPACKDecoderStreamError
	if encoder {
		code = ErrCodeQPACKEncoderStreamError
		err = c.qpackDecoder.ReadEncoder(stream)
	} else if c.clientQPACK != nil {
		err = c.clientQPACK.fieldEncoder.ReadDecoder(stream)
	} else {
		err = dynamic.ReadDecoder(stream)
	}
	if c.context().Err() != nil || errors.Is(err, quic.Err0RTTRejected) {
		return
	}
	if errors.Is(err, io.EOF) || errors.Is(err, io.ErrUnexpectedEOF) {
		code = ErrCodeClosedCriticalStream
	}
	if _, ok := errors.AsType[*quic.StreamError](err); ok {
		code = ErrCodeClosedCriticalStream
	}
	if errors.Is(err, errQPACKFeedbackBacklog) {
		code = ErrCodeExcessiveLoad
	}
	c.CloseWithError(quic.ApplicationErrorCode(code), fmt.Sprint(err))
}

type streamFieldDecoder struct {
	conn    *rawConn
	context context.Context
	stream  *stateTrackingStream
	limit   int
}

func bindFieldDecoder(fallback fieldDecoder, conn *rawConn, ctx context.Context, stream *stateTrackingStream, limit int) fieldDecoder {
	if conn.qpackDecoder == nil {
		return fallback
	}
	return &streamFieldDecoder{conn: conn, context: ctx, stream: stream, limit: limit}
}
func (d *streamFieldDecoder) Decode(block []byte) qpack.DecodeFunc {
	if len(block) >= 2 && block[0] == 0 && block[1] == 0 {
		decode := qpack.NewDecoder().Decode(block)
		return func() (qpack.HeaderField, error) {
			field, err := decode()
			if err != nil && err != io.EOF {
				d.conn.CloseWithError(quic.ApplicationErrorCode(ErrCodeQPACKDecompressionFailed), err.Error())
			}
			return field, err
		}
	}
	var fields []qpack.HeaderField
	var err error
	initialized := false
	return func() (qpack.HeaderField, error) {
		if !initialized {
			initialized = true
			fields, err = d.conn.qpackDecoder.Decode(d.context, uint64(d.stream.StreamID()), block, uint64(d.limit), d.stream.WaitForReadReady)
			if errors.Is(err, dynamic.ErrDecompression) {
				d.conn.CloseWithError(quic.ApplicationErrorCode(ErrCodeQPACKDecompressionFailed), err.Error())
			}
			if errors.Is(err, errQPACKFeedbackBacklog) {
				d.conn.CloseWithError(quic.ApplicationErrorCode(ErrCodeExcessiveLoad), err.Error())
			}
			if err != nil && !errors.Is(err, dynamic.ErrDecompression) {
				d.stream.cancelQPACK()
			}
			if errors.Is(err, dynamic.ErrFieldSectionTooLarge) {
				err = errHeaderTooLarge
			}
		}
		if err != nil {
			return qpack.HeaderField{}, err
		}
		if len(fields) == 0 {
			return qpack.HeaderField{}, io.EOF
		}
		field := fields[0]
		fields = fields[1:]
		return field, nil
	}
}

func (s *stateTrackingStream) cancelQPACK() {
	if s.qpackCancel != nil {
		s.qpackCancel()
	}
}
func (c *rawConn) qpackCancellation(id quic.StreamID) func() {
	if c.qpackDecoder == nil {
		return nil
	}
	return sync.OnceFunc(func() {
		if c.context().Err() != nil {
			return
		}
		if err := c.qpackDecoder.Cancel(uint64(id)); err != nil && c.context().Err() == nil {
			c.CloseWithError(quic.ApplicationErrorCode(ErrCodeExcessiveLoad), err.Error())
		}
	})
}
