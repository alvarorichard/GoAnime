package quic

import (
	"sync"

	"github.com/enetx/quic-go/internal/protocol"
)

// streamEpochSender prevents callbacks from rejected 0-RTT streams from reaching
// the replacement maps and framer, which reuse the same stream IDs. Retirement
// also waits for callbacks already executing before the maps are replaced.
type streamEpochSender struct {
	streamSender
	mutex   sync.RWMutex
	retired bool
}

func (s *streamEpochSender) retire() {
	s.mutex.Lock()
	s.retired = true
	s.mutex.Unlock()
}

func (s *streamEpochSender) call(f func()) {
	s.mutex.RLock()
	defer s.mutex.RUnlock()
	if !s.retired {
		f()
	}
}

func (s *streamEpochSender) onHasConnectionData() {
	s.call(s.streamSender.onHasConnectionData)
}

func (s *streamEpochSender) onHasStreamData(id protocol.StreamID, str *SendStream) {
	s.call(func() { s.streamSender.onHasStreamData(id, str) })
}

func (s *streamEpochSender) onHasStreamRetransmission(id protocol.StreamID, str *SendStream) {
	s.call(func() { s.streamSender.onHasStreamRetransmission(id, str) })
}

func (s *streamEpochSender) onHasStreamControlFrame(id protocol.StreamID, str streamControlFrameGetter) {
	s.call(func() { s.streamSender.onHasStreamControlFrame(id, str) })
}

func (s *streamEpochSender) updateStreamPriority(id protocol.StreamID) {
	s.call(func() { s.streamSender.updateStreamPriority(id) })
}

func (s *streamEpochSender) recordStreamPriorityUpdated(id protocol.StreamID, urgency int8, incremental bool) {
	s.call(func() { s.streamSender.recordStreamPriorityUpdated(id, urgency, incremental) })
}

func (s *streamEpochSender) onStreamCompleted(id protocol.StreamID) {
	s.call(func() { s.streamSender.onStreamCompleted(id) })
}

func (s *streamEpochSender) onStreamCreditReleased(id protocol.StreamID) {
	s.call(func() { s.streamSender.onStreamCreditReleased(id) })
}
