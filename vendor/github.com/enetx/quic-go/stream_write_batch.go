package quic

import (
	"context"
	"fmt"

	"github.com/enetx/quic-go/internal/protocol"
)

// StreamWrite stages bytes and optionally FIN on an already opened local stream.
type StreamWrite struct {
	StreamID StreamID
	Data     []byte
	Fin      bool
}

type streamWriteResult struct {
	count int
	err   error
}

type streamWriteBatch struct {
	writes []StreamWrite
	done   chan streamWriteResult
}

// TryWriteStreamBatch queues writes in order in one connection-loop turn, before
// packetization can consume them. It waits for that turn, never for stream credit
// or another writer. The result counts fully accepted elements; retry only the
// remaining suffix after ErrWouldBlock. FIN is queued with its element's bytes.
// Packet size, pacing and flow control still determine actual packet boundaries.
// At most 64 elements and 1 MiB of data are accepted per call. Only already opened
// locally initiated streams are supported. Caller buffers may be reused on return.
func (c *Conn) TryWriteStreamBatch(writes []StreamWrite) (int, error) {
	if len(writes) == 0 {
		return 0, nil
	}
	if len(writes) > 64 {
		return 0, ErrWouldBlock
	}
	total := 0
	for _, w := range writes {
		if w.StreamID < 0 || uint64(w.StreamID) >= 1<<62 || protocol.StreamInitiator(w.StreamID) != c.perspective {
			return 0, fmt.Errorf("quic: stream batch requires a locally initiated stream: %d", w.StreamID)
		}
		if len(w.Data) > 1<<20-total {
			return 0, ErrWouldBlock
		}
		total += len(w.Data)
	}

	batch := streamWriteBatch{writes: make([]StreamWrite, len(writes)), done: make(chan streamWriteResult, 1)}
	for i, w := range writes {
		batch.writes[i] = StreamWrite{StreamID: w.StreamID, Data: append([]byte(nil), w.Data...), Fin: w.Fin}
	}

	select {
	case c.streamWriteBatches <- batch:
	case <-c.ctx.Done():
		return 0, context.Cause(c.ctx)
	}

	return batch.wait(c.ctx)
}

// A completed admission owns its accepted count even if shutdown becomes ready
// at the same time. Only an unpublished result can fall back to cancellation.
func (b *streamWriteBatch) wait(ctx context.Context) (int, error) {
	select {
	case result := <-b.done:
		return result.count, result.err
	case <-ctx.Done():
		select {
		case result := <-b.done:
			return result.count, result.err
		default:
		}

		return 0, context.Cause(ctx)
	}
}

func (c *Conn) writeStreamBatch(batch streamWriteBatch) {
	result := streamWriteResult{}
	defer func() { batch.done <- result }()
	streams := make([]*SendStream, len(batch.writes))
	for i, w := range batch.writes {
		stream, err := c.streamsMap.getSendStream(w.StreamID)
		if err != nil {
			result.err = err
			return
		}
		switch stream := stream.(type) {
		case *SendStream:
			streams[i] = stream
		case *Stream:
			streams[i] = stream.sendStr
		default:
			result.err = ErrStreamClosed
			return
		}
	}
	for i, w := range batch.writes {
		if result.err = streams[i].TryWriteAll(w.Data); result.err != nil {
			return
		}
		result.count++
		if w.Fin {
			if result.err = streams[i].Close(); result.err != nil {
				return
			}
		}
	}
}
