package http3

import (
	"errors"
	"io"

	"github.com/enetx/http"
	"github.com/enetx/quic-go"
)

type requestHeaderWriter struct {
	io.Writer
	fin   bool
	qpack *qpackHeaderBatch
}

func (w *clientControlWriter) attachStream(stream *quic.SendStream) {
	w.mu.Lock()
	defer w.mu.Unlock()
	w.stream = stream
	if len(w.initial) > 0 {
		w.pending.Add(1)
	}
}

func (w *clientControlWriter) tryRequestBatch(priority []byte, request quic.StreamWrite) (count int, recorded bool, err error) {
	if !w.mu.TryLock() {
		return 0, false, quic.ErrWouldBlock
	}
	defer w.mu.Unlock()
	if w.pending.Load() != 0 || w.stream == nil {
		return 0, false, quic.ErrWouldBlock
	}
	// Older updates and blocked control writes retain ownership. Requests never
	// wait for critical-stream credit or for the asynchronous writer.
	recordWireFrames(w.owner.qlogger, w.stream.StreamID(), priority)
	n, err := w.owner.conn.TryWriteStreamBatch([]quic.StreamWrite{{StreamID: w.stream.StreamID(), Data: priority}, request})
	return n, true, err
}

func (c *ClientConn) writeRequestHeaders(wr io.Writer, req *http.Request, id quic.StreamID, data []byte) error {
	request := quic.StreamWrite{StreamID: id, Data: data}
	if writer, ok := wr.(*requestHeaderWriter); ok {
		request.Fin = writer.fin
	}
	var priority []byte
	u, incremental, explicit := requestPriority(req)
	if c.wireConfig.SendPriorityUpdates && explicit && (u != defaultPriorityUrgency || incremental) {
		priority = c.priorityUpdateFrame(id, u, incremental)
	}
	if writer, ok := wr.(*requestHeaderWriter); ok && writer.qpack != nil {
		done, err := c.tryQPACKRequestBatch(writer.qpack, priority, request)
		// The fallback can wait for request-stream credit. Return the critical
		// stream to its asynchronous owner before any blocking request write.
		writer.qpack.release()
		if done || !errors.Is(err, quic.ErrWouldBlock) {
			return err
		}
		return c.writeRequestOnly(wr, request)
	}
	if len(priority) == 0 {
		n, err := c.conn.TryWriteStreamBatch([]quic.StreamWrite{request})
		if n == 1 || !errors.Is(err, quic.ErrWouldBlock) {
			return err
		}
	} else {
		n, recorded, err := c.controlWriter.tryRequestBatch(priority, request)
		if n == 2 || !errors.Is(err, quic.ErrWouldBlock) {
			return err
		}
		if n == 0 {
			if err := c.controlWriter.enqueueWrite(controlWrite{data: priority, recorded: recorded}); err != nil {
				return &RequestPreparationError{err}
			}
		}
		return c.writeRequestOnly(wr, request)
	}
	// Backpressure retains the native blocking request write; no body is read
	// or replayed. The bounded asynchronous writer owns the remaining update.
	return writeRequestHeaderFallback(wr, data)
}

// An earlier critical-stream element can block a batch before the request is
// attempted. Queue the request independently, retaining its FIN when its own
// credit permits; only real request backpressure needs the blocking fallback.
func (c *ClientConn) writeRequestOnly(wr io.Writer, request quic.StreamWrite) error {
	n, err := c.conn.TryWriteStreamBatch([]quic.StreamWrite{request})
	if n == 1 || !errors.Is(err, quic.ErrWouldBlock) {
		return err
	}
	return writeRequestHeaderFallback(wr, request.Data)
}

func writeRequestHeaderFallback(wr io.Writer, data []byte) error {
	n, err := wr.Write(data)
	if err == nil && n != len(data) {
		err = io.ErrShortWrite
	}
	return err
}
