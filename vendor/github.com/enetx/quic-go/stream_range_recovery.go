package quic

import (
	"slices"

	"github.com/enetx/quic-go/internal/ackhandler"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/wire"
)

// streamRangeRecovery is protected by its SendStream's mutex. ACK state belongs
// to byte ranges and FIN, independently of the packets carrying their copies.
// Unlike the native frame handler, each transmission has immutable metadata.
type streamRangeRecovery struct {
	acked    []streamByteRange
	finAcked bool
	live     map[*streamRangeFrame]struct{}
	closed   bool
}

type streamByteRange struct{ start, end protocol.ByteCount }

func subtractStreamRange(parts []streamByteRange, removed streamByteRange) []streamByteRange {
	result := make([]streamByteRange, 0, len(parts)+1)
	for _, p := range parts {
		if removed.end <= p.start || removed.start >= p.end {
			result = append(result, p)
			continue
		}
		if p.start < removed.start {
			result = append(result, streamByteRange{p.start, removed.start})
		}
		if removed.end < p.end {
			result = append(result, streamByteRange{removed.end, p.end})
		}
	}
	return result
}

func (r *streamRangeRecovery) acknowledge(part streamByteRange) {
	if part.start == part.end {
		return
	}
	r.acked = append(r.acked, part)
	slices.SortFunc(r.acked, func(a, b streamByteRange) int {
		if a.start < b.start {
			return -1
		}
		if a.start > b.start {
			return 1
		}
		return 0
	})
	merged := r.acked[:0]
	for _, p := range r.acked {
		if len(merged) == 0 || merged[len(merged)-1].end < p.start {
			merged = append(merged, p)
		} else {
			merged[len(merged)-1].end = max(merged[len(merged)-1].end, p.end)
		}
	}
	r.acked = merged
}

func (r *streamRangeRecovery) missing(s *SendStream, start, end protocol.ByteCount) []streamByteRange {
	if r.closed || s.shutdownErr != nil {
		return nil
	}
	if s.resetErr != nil {
		end = min(end, s.reliableOffset())
	}
	if start >= end {
		return nil
	}
	parts := []streamByteRange{{start, end}}
	for _, ack := range r.acked {
		parts = subtractStreamRange(parts, ack)
	}
	return parts
}

func (r *streamRangeRecovery) needsFIN(s *SendStream, fin bool) bool {
	return fin && !r.closed && !r.finAcked && s.resetErr == nil && s.shutdownErr == nil
}

func (r *streamRangeRecovery) satisfied(s *SendStream, f *streamRangeFrame) bool {
	return len(r.missing(s, f.part.start, f.part.end)) == 0 && !r.needsFIN(s, f.fin)
}

func (r *streamRangeRecovery) retireSatisfied(s *SendStream) {
	for f := range r.live {
		if r.satisfied(s, f) {
			s.numOutstandingFrames--
			f.counted = false
			delete(r.live, f)
		}
	}
}

// fragments copies only missing ranges. Queued copies are excluded when a new
// loss/probe is scheduled, so overlapping originals cannot queue duplicate data.
func (r *streamRangeRecovery) fragments(s *SendStream, f *wire.StreamFrame, excludeQueued bool) []*wire.StreamFrame {
	end := f.Offset + f.DataLen()
	parts := r.missing(s, f.Offset, end)
	fin := r.needsFIN(s, f.Fin)
	if excludeQueued {
		for _, queued := range s.retransmissionQueue {
			parts = subtractStreamRange(parts, streamByteRange{queued.Offset, queued.Offset + queued.DataLen()})
			if queued.Fin {
				fin = false
			}
		}
	}
	result := make([]*wire.StreamFrame, 0, len(parts)+1)
	for _, p := range parts {
		copy := s.newPooledStreamFrame(p.start, f.Data[p.start-f.Offset:p.end-f.Offset])
		if p.end == end && fin {
			copy.Fin = true
			fin = false
		}
		result = append(result, copy)
	}
	if fin {
		result = append(result, &wire.StreamFrame{StreamID: s.streamID, Offset: end, DataLenPresent: true, Fin: true})
	}
	return result
}

func (r *streamRangeRecovery) trimQueue(s *SendStream) {
	old := s.retransmissionQueue
	s.retransmissionQueue = nil
	for _, f := range old {
		parts := r.missing(s, f.Offset, f.Offset+f.DataLen())
		fin := r.needsFIN(s, f.Fin)
		if len(parts) == 1 && parts[0].start == f.Offset && parts[0].end == f.Offset+f.DataLen() && fin == f.Fin || len(f.Data) == 0 && fin {
			s.retransmissionQueue = append(s.retransmissionQueue, f)
		} else {
			s.retransmissionQueue = append(s.retransmissionQueue, r.fragments(s, f, false)...)
			f.PutBack()
		}
	}
	clear(old)
}

// Coalesce adjacent queued bytes without tying their next packet boundary to a
// previous transmission. Larger ranges stay in bounded pooled chunks; a split
// remainder is coalesced again against the following chunk on the next pop.
func (r *streamRangeRecovery) coalesceQueue(s *SendStream) {
	queue := s.retransmissionQueue
	slices.SortFunc(queue, func(a, b *wire.StreamFrame) int {
		if a.Offset < b.Offset {
			return -1
		}
		if a.Offset > b.Offset {
			return 1
		}
		return 0
	})
	merged := queue[:0]
	for _, f := range queue {
		if len(merged) > 0 {
			last := merged[len(merged)-1]
			if !last.Fin && last.Offset+last.DataLen() == f.Offset && last.DataLen()+f.DataLen() <= protocol.MaxPacketBufferSize {
				last.Data = append(last.Data, f.Data...)
				last.Fin = f.Fin
				f.PutBack()
				continue
			}
		}
		merged = append(merged, f)
	}
	clear(queue[len(merged):])
	s.retransmissionQueue = merged
}

func (s *SendStream) streamFrameHandler(f *wire.StreamFrame) ackhandler.FrameHandler {
	if s.recovery == nil {
		return (*sendStreamAckHandler)(s)
	}
	handler := &streamRangeFrame{stream: s, frame: f, part: streamByteRange{f.Offset, f.Offset + f.DataLen()}, fin: f.Fin, counted: true}
	if s.recovery.live == nil {
		s.recovery.live = make(map[*streamRangeFrame]struct{})
	}
	s.recovery.live[handler] = struct{}{}
	return handler
}

type streamRangeFrame struct {
	stream       *SendStream
	frame        *wire.StreamFrame // released once; late ACKs use only part and fin
	part         streamByteRange
	fin, counted bool
}

func (f *streamRangeFrame) release() {
	if f.frame != nil {
		f.frame.PutBack()
		f.frame = nil
	}
}

func (f *streamRangeFrame) HasOutstandingData() bool {
	s := f.stream
	s.mutex.Lock()
	defer s.mutex.Unlock()
	return !s.recovery.satisfied(s, f)
}

func (f *streamRangeFrame) OnAcked(wire.Frame) {
	s := f.stream
	s.mutex.Lock()
	r := s.recovery
	if !r.closed {
		r.acknowledge(f.part)
		r.finAcked = r.finAcked || f.fin
		r.trimQueue(s)
		r.retireSatisfied(s)
	}
	f.release()
	completed := !r.closed && s.isNewlyCompleted()
	s.mutex.Unlock()
	if completed {
		s.sender.onStreamCompleted(s.streamID)
	}
}

func (f *streamRangeFrame) schedule(probe bool) bool {
	s := f.stream
	s.mutex.Lock()
	r := s.recovery
	wasEmpty := len(s.retransmissionQueue) == 0
	added := false
	if f.frame != nil && !r.closed {
		copies := r.fragments(s, f.frame, true)
		added = len(copies) > 0
		s.retransmissionQueue = append(s.retransmissionQueue, copies...)
		if added {
			r.coalesceQueue(s)
		}
	}
	if !probe {
		if f.counted && !r.closed {
			s.numOutstandingFrames--
			f.counted = false
			delete(r.live, f)
		}
		f.release()
	}
	completed := !r.closed && s.isNewlyCompleted()
	s.mutex.Unlock()
	if added && wasEmpty {
		s.sender.onHasStreamRetransmission(s.streamID, s)
	}
	if completed {
		s.sender.onStreamCompleted(s.streamID)
	}
	return added
}

func (f *streamRangeFrame) OnLost(wire.Frame)       { f.schedule(false) }
func (f *streamRangeFrame) OnProbe(wire.Frame) bool { return f.schedule(true) }
