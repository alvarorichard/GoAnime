package quic

import (
	"cmp"
	"slices"

	"github.com/enetx/quic-go/internal/ackhandler"
	"github.com/enetx/quic-go/internal/wire"
)

// The native control queue is LIFO. Reorder only NEW_CONNECTION_ID slots in
// descending order so popping emits ascending sequence numbers. PATH_RESPONSE,
// CRYPTO and flow-control frames retain their native relative positions.
func orderConnectionIDFrames(frames []wire.Frame) {
	var ids []*wire.NewConnectionIDFrame
	for _, f := range frames {
		if id, ok := f.(*wire.NewConnectionIDFrame); ok {
			ids = append(ids, id)
		}
	}
	if len(ids) < 2 {
		return
	}
	slices.SortFunc(ids, func(a, b *wire.NewConnectionIDFrame) int { return cmp.Compare(b.SequenceNumber, a.SequenceNumber) })
	i := 0
	for pos, frame := range frames {
		if _, ok := frame.(*wire.NewConnectionIDFrame); ok {
			frames[pos] = ids[i]
			i++
		}
	}
}

// Serialization shuffles control frames after queue admission. Restore only the
// NEW_CONNECTION_ID subsequence and move each frame with its loss/ACK handler.
func orderPayloadConnectionIDFrames(frames []ackhandler.Frame) {
	var ids []ackhandler.Frame
	for _, f := range frames {
		if _, ok := f.Frame.(*wire.NewConnectionIDFrame); ok {
			ids = append(ids, f)
		}
	}
	if len(ids) < 2 {
		return
	}
	slices.SortFunc(ids, func(a, b ackhandler.Frame) int {
		return cmp.Compare(a.Frame.(*wire.NewConnectionIDFrame).SequenceNumber, b.Frame.(*wire.NewConnectionIDFrame).SequenceNumber)
	})
	i := 0
	for pos, frame := range frames {
		if _, ok := frame.Frame.(*wire.NewConnectionIDFrame); ok {
			frames[pos] = ids[i]
			i++
		}
	}
}
