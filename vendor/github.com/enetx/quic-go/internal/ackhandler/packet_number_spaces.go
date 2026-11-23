package ackhandler

import (
	"iter"

	"github.com/enetx/quic-go/internal/protocol"
)

// Each wire space owns its generator and history. Congestion control uses a
// separate transmission sequence, so an ACK of Initial 10 can never put an
// already-sent Handshake 0 on the wrong side of a recovery boundary.
type separatePacketNumberSpaces struct {
	generators       [3]*packetNumberGenerator
	histories        [3]*sentPacketHistory
	nextTransmission protocol.PacketNumber
}

func newSeparatePacketNumberSpaces(initial protocol.PacketNumber) *separatePacketNumberSpaces {
	s := new(separatePacketNumberSpaces)
	for i := range s.generators {
		first := protocol.PacketNumber(0)
		if protocol.PacketNumberSpace(i) == protocol.PacketNumberSpaceInitial {
			first = initial
		}
		s.generators[i] = newPacketNumberGenerator(first, skipPacketInitialPeriod, skipPacketMaxPeriod)
		s.histories[i] = newSentPacketHistory()
	}
	return s
}

func (h *sentPacketHandler) history(space protocol.PacketNumberSpace) *sentPacketHistory {
	if h.separateSpaces != nil {
		return h.separateSpaces.histories[space]
	}
	return &h.hist
}

func (h *sentPacketHandler) histories() iter.Seq[*sentPacketHistory] {
	return func(yield func(*sentPacketHistory) bool) {
		if h.separateSpaces == nil {
			yield(&h.hist)
			return
		}
		for _, history := range h.separateSpaces.histories {
			if !yield(history) {
				return
			}
		}
	}
}

func (h *sentPacketHandler) packetNumbers(space protocol.PacketNumberSpace) *packetNumberGenerator {
	if h.separateSpaces != nil {
		return h.separateSpaces.generators[space]
	}
	return h.pns
}

func (h *sentPacketHandler) trackedPackets() (tracked, outstanding int) {
	for history := range h.histories() {
		tracked += history.Len()
		outstanding += history.NumOutstanding()
	}
	return
}

func (h *sentPacketHandler) hasOutstandingPackets() bool {
	_, outstanding := h.trackedPackets()
	return outstanding > 0
}
