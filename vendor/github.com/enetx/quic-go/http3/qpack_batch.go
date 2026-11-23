package http3

import (
	"errors"
	"io"

	"github.com/enetx/quic-go"
	"github.com/enetx/quic-go/quicvarint"
)

// A request may borrow an idle encoder writer through encoding and the
// nonblocking connection batch. A busy writer remains independent: requests
// must never wait for encoder credit while holding connection or QPACK locks.
type qpackHeaderBatch struct {
	stream *clientQPACKStream
	held   bool
}

func (b *qpackHeaderBatch) release() {
	if !b.held {
		return
	}
	b.held = false
	b.stream.writeMutex.Unlock()
	b.stream.notify()
}

func (c *ClientConn) prepareQPACKRequestHeaders(wr io.Writer) (io.Writer, func()) {
	s := c.rawConn.clientQPACK.encoder
	select {
	case <-s.ready:
	default:
		return wr, func() {}
	}
	if !s.writeMutex.TryLock() {
		return wr, func() {}
	}
	// A previous blocked batch retains its original bytes. Let the worker
	// drain it instead of repeatedly accumulating larger request batches.
	if s.stream == nil || s.pending != nil {
		s.writeMutex.Unlock()
		return wr, func() {}
	}

	batch := &qpackHeaderBatch{stream: s, held: true}
	writer := &requestHeaderWriter{Writer: wr, qpack: batch}
	if previous, ok := wr.(*requestHeaderWriter); ok {
		writer.fin = previous.fin
	}
	return writer, batch.release
}

// The caller owns writeMutex before encoding, so the worker cannot take one
// instruction and strand it behind this batch. SETTINGS arriving concurrently
// append to the same bounded queue and retain encoder-stream ordering.
func (b *qpackHeaderBatch) stage() []byte {
	s := b.stream
	// Bound the batch to the existing queue snapshot. New concurrent producers
	// cannot turn this into an unbounded drain or starve the connection loop.
	for range len(s.queue) {
		s.pending = append(s.pending, (<-s.queue)...)
	}
	if s.pending == nil && !(s.eager && !s.typeStaged) {
		return nil
	}
	if s.typeStaged {
		return s.pending
	}
	return append(quicvarint.Append(nil, uint64(s.typ)), s.pending...)
}

func (c *ClientConn) tryQPACKRequestBatch(
	batch *qpackHeaderBatch,
	priority []byte,
	request quic.StreamWrite,
) (bool, error) {
	control := c.controlWriter
	if len(priority) > 0 {
		if !control.mu.TryLock() {
			return false, enqueuePriorityFallback(control, priority, false)
		}
		defer control.mu.Unlock()
		if control.pending.Load() != 0 || control.stream == nil {
			return false, enqueuePriorityFallback(control, priority, false)
		}
	}

	writes := make([]quic.StreamWrite, 0, 3)
	if len(priority) > 0 {
		recordWireFrames(c.qlogger, control.stream.StreamID(), priority)
		writes = append(writes, quic.StreamWrite{StreamID: control.stream.StreamID(), Data: priority})
	}
	encoderAt := -1
	if data := batch.stage(); data != nil {
		encoderAt = len(writes)
		writes = append(writes, quic.StreamWrite{StreamID: batch.stream.stream.StreamID(), Data: data})
	}
	writes = append(writes, request)

	n, err := c.conn.TryWriteStreamBatch(writes)
	if encoderAt >= 0 && n > encoderAt {
		batch.stream.typeStaged = true
		batch.stream.pending = nil
	}
	if len(priority) > 0 && n == 0 && errors.Is(err, quic.ErrWouldBlock) {
		return false, enqueuePriorityFallback(control, priority, true)
	}
	return n == len(writes), err
}

func enqueuePriorityFallback(control *clientControlWriter, data []byte, recorded bool) error {
	if err := control.enqueueWrite(controlWrite{data: data, recorded: recorded}); err != nil {
		return &RequestPreparationError{err}
	}
	return quic.ErrWouldBlock
}
