package wire

import (
	"fmt"
	"io"
	"math"
	"time"

	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/quicvarint"
)

// LegacyAckFrequencyFrame is the ACK_FREQUENCY draft with a single-byte
// ignore_order flag, not the later varint reordering threshold. The same type
// code is used, so a connection must explicitly select the negotiated format.
type LegacyAckFrequencyFrame struct {
	SequenceNumber        uint64
	AckElicitingThreshold uint64
	RequestMaxAckDelay    time.Duration
	IgnoreOrder           bool
}

func parseLegacyAckFrequencyFrame(b []byte) (*LegacyAckFrequencyFrame, int, error) {
	start := len(b)
	values := [3]uint64{}
	for i := range values {
		value, n, err := quicvarint.Parse(b)
		if err != nil {
			return nil, 0, replaceUnexpectedEOF(err)
		}
		values[i], b = value, b[n:]
	}
	if len(b) == 0 {
		return nil, 0, io.EOF
	}
	if values[1] == 0 || b[0] > 1 {
		return nil, 0, fmt.Errorf("invalid legacy ACK_FREQUENCY tolerance or ignore_order")
	}
	delay := time.Duration(math.MaxInt64)
	if values[2] <= math.MaxInt64/uint64(time.Microsecond) {
		delay = time.Duration(values[2]) * time.Microsecond
	}
	return &LegacyAckFrequencyFrame{SequenceNumber: values[0], AckElicitingThreshold: values[1],
		RequestMaxAckDelay: delay, IgnoreOrder: b[0] == 1}, start - len(b) + 1, nil
}

func (f *LegacyAckFrequencyFrame) Append(b []byte, _ protocol.Version) ([]byte, error) {
	if f.SequenceNumber > quicvarint.Max || f.AckElicitingThreshold == 0 || f.AckElicitingThreshold > quicvarint.Max || f.RequestMaxAckDelay < 0 {
		return b, fmt.Errorf("invalid legacy ACK_FREQUENCY fields")
	}
	b = quicvarint.Append(b, uint64(FrameTypeAckFrequency))
	b = quicvarint.Append(b, f.SequenceNumber)
	b = quicvarint.Append(b, f.AckElicitingThreshold)
	b = quicvarint.Append(b, uint64(f.RequestMaxAckDelay/time.Microsecond))
	if f.IgnoreOrder {
		return append(b, 1), nil
	}
	return append(b, 0), nil
}

func (f *LegacyAckFrequencyFrame) Length(_ protocol.Version) protocol.ByteCount {
	return protocol.ByteCount(3 + quicvarint.Len(f.SequenceNumber) + quicvarint.Len(f.AckElicitingThreshold) + quicvarint.Len(uint64(f.RequestMaxAckDelay/time.Microsecond)))
}
