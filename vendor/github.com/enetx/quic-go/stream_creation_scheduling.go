package quic

import (
	"container/list"

	"github.com/enetx/quic-go/internal/ackhandler"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/wire"
)

// Protected by framer.mutex. Completion removes entries; an idle open stream
// keeps its place without retaining payload or a history of closed streams.
type streamCreationOrder struct {
	entries            map[protocol.StreamID]*list.Element
	ids                list.List
	prioritizeFinished bool
}

func (f *framer) enableStreamCreationOrder() {
	f.creationOrder = &streamCreationOrder{entries: make(map[protocol.StreamID]*list.Element)}
}

func (f *framer) recordStreamCreation(id protocol.StreamID) {
	if f == nil || f.creationOrder == nil {
		return
	}
	f.mutex.Lock()
	defer f.mutex.Unlock()
	f.recordStreamCreationLocked(id)
}

func (f *framer) recordStreamCreationLocked(id protocol.StreamID) {
	if f.creationOrder.entries[id] == nil {
		f.creationOrder.entries[id] = f.creationOrder.ids.PushBack(id)
	}
}

func (f *framer) appendStreamFramesByCreation(frames []ackhandler.Frame, streams []ackhandler.StreamFrame, maxLen, minimum protocol.ByteCount, v protocol.Version) ([]ackhandler.Frame, []ackhandler.StreamFrame, protocol.ByteCount) {
	f.controlFrameMutex.Lock()
	defer f.controlFrameMutex.Unlock()
	f.mutex.Lock()
	defer f.mutex.Unlock()
	var length protocol.ByteCount
	var last *wire.StreamFrame
	appendStream := func(frame ackhandler.StreamFrame) {
		streams = append(streams, frame)
		n := frame.Frame.Length(v)
		length += n
		maxLen -= n
		last = frame.Frame
	}
	appendBlocked := func(frame wire.Frame) {
		n := frame.Length(v)
		if n > maxLen {
			f.controlFrames = append(f.controlFrames, frame)
			return
		}
		frames = append(frames, ackhandler.Frame{Frame: frame})
		length += n
		maxLen -= n
	}
	appendNew := func(id protocol.StreamID) {
		str, ok := f.activeStreams[id]
		if !ok {
			return
		}
		for maxLen >= minimum {
			frame, blocked, more := f.popStreamFrame(id, str, maxLen, v)
			if frame.Frame != nil {
				appendStream(frame)
			}
			if blocked != nil {
				appendBlocked(blocked)
			}
			if !more || frame.Frame == nil {
				break
			}
		}
	}
	// Within the higher class, retain creation order across lost ranges and
	// optionally write-closed streams. A blocked stream never stalls a later
	// stream. The default retains a separate retransmissions-only pass.
	for entry := f.creationOrder.ids.Front(); entry != nil && maxLen >= minimum; entry = entry.Next() {
		id := entry.Value.(protocol.StreamID)
		if str := f.retransmissionStreams[id]; str != nil {
			for maxLen >= minimum {
				frame, more := popRetransmissionFrameForPacket(str, maxLen, v)
				if !more {
					delete(f.retransmissionStreams, id)
				}
				if frame.Frame != nil {
					appendStream(frame)
				}
				if !more || frame.Frame == nil {
					break
				}
			}
		}
		if f.creationOrder.prioritizeFinished && f.hasFinishedWrite(id) {
			appendNew(id)
		}
	}
	for entry := f.creationOrder.ids.Front(); entry != nil && maxLen >= minimum; entry = entry.Next() {
		id := entry.Value.(protocol.StreamID)
		if !f.creationOrder.prioritizeFinished || !f.hasFinishedWrite(id) {
			appendNew(id)
		}
	}
	if blocked, offset := f.connFlowController.IsNewlyBlocked(); blocked {
		appendBlocked(&wire.DataBlockedFrame{MaximumData: offset})
	}
	if last != nil {
		length -= last.Length(v)
		last.DataLenPresent = false
		length += last.Length(v)
	}
	return frames, streams, length
}

// Called with framer.mutex held. Concrete streams publish their write-side FIN
// under their own mutex; no application priority or stream ID is inferred.
func (f *framer) hasFinishedWrite(id protocol.StreamID) bool {
	stream, ok := f.activeStreams[id].streamFrameGetter.(*SendStream)
	if !ok {
		return false
	}
	stream.mutex.Lock()
	defer stream.mutex.Unlock()
	return stream.finishedWriting && stream.resetErr == nil && stream.shutdownErr == nil
}
