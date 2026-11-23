package quic

import (
	"slices"

	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/wire"
)

// cryptoRangeRecovery tracks bytes rather than the packetization of their
// previous transmission. The connection's send loop owns all access. CRYPTO
// key retirement releases the retained transcript and makes old handlers inert.
type cryptoRangeRecovery struct {
	data           []byte
	acked, pending []CryptoRange
	sent           []CryptoRange
	peerProbe      bool
	active         []*wire.CryptoFrame
	layout         func(InitialCryptoLayoutInfo) [][]CryptoRange
	closed         bool
	probePackets   uint8
}

func (r *cryptoRangeRecovery) OnAcked(frame wire.Frame) {
	if r.closed {
		return
	}
	if f, ok := frame.(*wire.CryptoFrame); ok {
		part := CryptoRange{Offset: int(f.Offset), Length: len(f.Data)}
		r.acked = unionCryptoRange(r.acked, part)
		r.pending = subtractCryptoRange(r.pending, part)
		r.active = nil // a late ACK may have satisfied part of a queued group
	}
}

func (r *cryptoRangeRecovery) OnLost(frame wire.Frame) {
	if r.closed {
		return
	}
	if f, ok := frame.(*wire.CryptoFrame); ok {
		part := CryptoRange{Offset: int(f.Offset), Length: len(f.Data)}
		parts := []CryptoRange{part}
		for _, ack := range r.acked {
			parts = subtractCryptoRange(parts, ack)
		}
		for _, part := range parts {
			r.pending = unionCryptoRange(r.pending, part)
		}
		r.active = nil
	}
}

func unionCryptoRange(ranges []CryptoRange, added CryptoRange) []CryptoRange {
	if added.Length <= 0 {
		return ranges
	}
	ranges = append(ranges, added)
	slices.SortFunc(ranges, func(a, b CryptoRange) int {
		if a.Offset < b.Offset {
			return -1
		}
		if a.Offset > b.Offset {
			return 1
		}
		return 0
	})
	merged := ranges[:0]
	for _, part := range ranges {
		if len(merged) == 0 || merged[len(merged)-1].Offset+merged[len(merged)-1].Length < part.Offset {
			merged = append(merged, part)
		} else {
			last := &merged[len(merged)-1]
			last.Length = max(last.Offset+last.Length, part.Offset+part.Length) - last.Offset
		}
	}
	return merged
}

func subtractCryptoRange(ranges []CryptoRange, removed CryptoRange) []CryptoRange {
	var result []CryptoRange
	end := removed.Offset + removed.Length
	for _, part := range ranges {
		partEnd := part.Offset + part.Length
		if end <= part.Offset || removed.Offset >= partEnd {
			result = append(result, part)
			continue
		}
		if part.Offset < removed.Offset {
			result = append(result, CryptoRange{part.Offset, removed.Offset - part.Offset})
		}
		if end < partEnd {
			result = append(result, CryptoRange{end, partEnd - end})
		}
	}
	return result
}

func (r *cryptoRangeRecovery) prepare(packetSize, budget protocol.ByteCount) error {
	if len(r.pending) == 0 {
		return nil
	}
	info := InitialCryptoLayoutInfo{Length: len(r.data), SNIOffset: -1, ECHOffset: -1, PacketSizeLimit: int(packetSize), FrameBudget: int(budget), Remaining: slices.Clone(r.pending)}
	// The SNI belongs to the first ClientHello. Later Initial messages keep
	// their absolute offsets, so a policy cannot mistake CH2 for offset zero.
	if completeInitialClientHello(r.data) && len(r.data) > 0 && r.data[0] == 1 {
		total := 4 + (int(r.data[1]) << 16) + (int(r.data[2]) << 8) + int(r.data[3])
		if sni, n, ech, err := findSNIAndECH(r.data[:total]); err == nil {
			info.SNIOffset, info.SNILength, info.ECHOffset = sni, max(n, 0), ech
			if ech >= 0 {
				info.ECHLength = 4 + (int(r.data[ech+2]) << 8) + int(r.data[ech+3])
			}
		}
	}
	var layout [][]CryptoRange
	if r.layout != nil {
		layout = r.layout(info)
	}
	if layout == nil {
		layout = [][]CryptoRange{slices.Clone(r.pending)}
	}
	frames, err := initialCryptoLayoutFrames(r.data, layout, r.pending)
	if err != nil {
		return err
	}
	r.active = frames[0]
	return nil
}

func (r *cryptoRangeRecovery) pop(budget protocol.ByteCount) *wire.CryptoFrame {
	if len(r.active) == 0 {
		return nil
	}
	f := r.active[0]
	if split, needed := f.MaybeSplitOffFrame(budget, protocol.Version1); needed {
		if split != nil {
			r.pending = subtractCryptoRange(r.pending, CryptoRange{int(split.Offset), len(split.Data)})
		}
		return split
	}
	r.active = r.active[1:]
	r.pending = subtractCryptoRange(r.pending, CryptoRange{int(f.Offset), len(f.Data)})
	return f
}

func (r *cryptoRangeRecovery) close() {
	r.closed = true
	r.data, r.acked, r.pending, r.active = nil, nil, nil, nil
	r.sent = nil
}

func (r *cryptoRangeRecovery) resendUnacked() {
	if r.closed || !r.peerProbe {
		return
	}
	r.rescheduleUnacked()
}

func (r *cryptoRangeRecovery) rescheduleUnacked() {
	if r.closed {
		return
	}
	parts := slices.Clone(r.sent)
	for _, ack := range r.acked {
		parts = subtractCryptoRange(parts, ack)
	}
	for _, part := range parts {
		r.pending = unionCryptoRange(r.pending, part)
	}
	r.active = nil
}
