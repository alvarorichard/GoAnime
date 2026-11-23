package quic

import (
	"fmt"
	"math/rand/v2"
	"slices"

	"github.com/enetx/quic-go/internal/ackhandler"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/wire"
	"github.com/enetx/quic-go/quicvarint"
)

// InitialPacketProtection controls optional client Initial CRYPTO fragmentation,
// PING insertion, padding distribution and frame shuffling. Counts are inclusive
// ranges of split attempts and PINGs, bounded by available padding. A zero first
// frame range leaves the native cross-packet CRYPTO layout. No browser values
// are selected by the engine. Nil preserves native behavior.
type InitialPacketProtection struct {
	MinCryptoSplits, MaxCryptoSplits     int
	MinPingFrames, MaxPingFrames         int
	MinFirstFrameSize, MaxFirstFrameSize int
}

func (p *InitialPacketProtection) validate() error {
	if p == nil {
		return nil
	}
	for _, pair := range [][3]int{{p.MinCryptoSplits, p.MaxCryptoSplits, 64}, {p.MinPingFrames, p.MaxPingFrames, 64}, {p.MinFirstFrameSize, p.MaxFirstFrameSize, 16384}} {
		if pair[0] < 0 || pair[1] < pair[0] || pair[1] > pair[2] {
			return fmt.Errorf("invalid Initial protection range")
		}
	}
	if p.MinFirstFrameSize == 0 && p.MaxFirstFrameSize != 0 {
		return fmt.Errorf("Initial first-frame range must be positive or both zero")
	}
	return nil
}

type initialProtectionState struct {
	policy   InitialPacketProtection
	prepared bool
	ordered  bool
	active   []*wire.CryptoFrame
	pending  [][]*wire.CryptoFrame
}

func (s *initialCryptoStream) hasProtectedData() bool {
	p := s.protection
	if s.layoutState != nil && len(s.layoutState.remaining) > 0 {
		return true
	}
	if len(p.active) > 0 || len(p.pending) > 0 {
		return true
	}
	if !p.prepared && s.writeOffset == 0 && len(s.writeBuf) > 0 && s.writeBuf[0] == 1 {
		// TLS may queue ClientHello in several writes. Wait for its complete
		// handshake message before selecting the first and trailing ranges.
		if len(s.writeBuf) < 4 {
			return false
		}
		length := int(s.writeBuf[1])<<16 | int(s.writeBuf[2])<<8 | int(s.writeBuf[3])
		return len(s.writeBuf) >= 4+length
	}
	return s.baseCryptoStream.HasData()
}

// Multi-packet layout follows the first-header/trailing-data strategy, with
// caller-selected first-frame sizes. Subsequent data retains its real offsets.
// https://github.com/google/quiche/blob/a1628192cf32ba40d21e0dd80f485d68c24ff87a/quiche/quic/core/quic_packet_creator.cc
func (s *initialCryptoStream) beginProtectedPacket(budget protocol.ByteCount, rng *rand.Rand) {
	p := s.protection
	if !p.prepared && s.writeOffset == 0 && len(s.writeBuf) >= 4 && s.writeBuf[0] == 1 && s.hasProtectedData() {
		p.prepared = true
		total := 4 + (int(s.writeBuf[1]) << 16) + (int(s.writeBuf[2]) << 8) + int(s.writeBuf[3])
		overhead := 1 + quicvarint.Len(uint64(total-1)) + quicvarint.Len(uint64(total))
		firstCapacity, otherCapacity := int(budget)-2*overhead, int(budget)-overhead
		if p.policy.MinFirstFrameSize > 0 && firstCapacity > 0 && total > firstCapacity && otherCapacity > 0 {
			numPackets := (total + overhead + otherCapacity - 1) / otherCapacity
			firstBytes := (total+overhead+numPackets-1)/numPackets - overhead
			prefix := randomRange(rng, p.policy.MinFirstFrameSize, p.policy.MaxFirstFrameSize)
			if firstBytes > prefix && firstBytes < total {
				tail := total - (firstBytes - prefix)
				p.pending = [][]*wire.CryptoFrame{
					{{Offset: 0, Data: s.writeBuf[:prefix]}, {Offset: protocol.ByteCount(tail), Data: s.writeBuf[tail:total]}},
					{{Offset: protocol.ByteCount(prefix), Data: s.writeBuf[prefix:tail]}},
				}
				s.writeOffset = protocol.ByteCount(total)
				s.writeBuf = s.writeBuf[total:]
			}
		}
	}
	if len(p.active) == 0 && len(p.pending) > 0 {
		p.active, p.pending = p.pending[0], p.pending[1:]
	}
}

func (s *initialCryptoStream) popProtectedFrame(budget protocol.ByteCount) *wire.CryptoFrame {
	p := s.protection
	if len(p.active) == 0 {
		if s.layoutState != nil {
			return nil // re-evaluate the next packet with its own live budget
		}
		if len(p.pending) > 0 {
			return nil // preserve the first packet's padding budget
		}
		return s.baseCryptoStream.PopCryptoFrame(budget)
	}
	f := p.active[0]
	if split, needed := f.MaybeSplitOffFrame(budget, protocol.Version1); needed {
		if split != nil {
			s.consumeCryptoLayoutRange(split)
		}
		return split
	}
	p.active = p.active[1:]
	s.consumeCryptoLayoutRange(f)
	return f
}

func randomRange(rng *rand.Rand, lo, hi int) int { return lo + rng.IntN(hi-lo+1) }

type protectedInitialFrame struct {
	frame   ackhandler.Frame
	padding protocol.ByteCount
}

// Packet-local protection spends padding rather than increasing packet length.
// Fragment handlers retain the native retransmission queue and exact offsets.
// https://github.com/google/quiche/blob/a1628192cf32ba40d21e0dd80f485d68c24ff87a/quiche/quic/core/quic_chaos_protector.cc
func (p *packetPacker) appendProtectedInitial(raw []byte, pl payload, padding protocol.ByteCount, v protocol.Version) ([]byte, payload, error) {
	policy := p.initialProtection
	frames := slices.Clone(pl.frames)
	hasCrypto := false
	for _, f := range frames {
		_, ok := f.Frame.(*wire.CryptoFrame)
		hasCrypto = hasCrypto || ok
	}
	if !hasCrypto || padding <= 0 {
		data, err := p.appendPacketPayload(raw, pl, padding, v)
		return data, pl, err
	}
	for range randomRange(&p.rand, policy.MinCryptoSplits, policy.MaxCryptoSplits) {
		index := p.rand.IntN(len(frames))
		original, ok := frames[index].Frame.(*wire.CryptoFrame)
		if !ok || len(original.Data) < 2 {
			continue
		}
		n := 1 + p.rand.IntN(len(original.Data)-1)
		a := &wire.CryptoFrame{Offset: original.Offset, Data: original.Data[:n]}
		b := &wire.CryptoFrame{Offset: original.Offset + protocol.ByteCount(n), Data: original.Data[n:]}
		extra := a.Length(v) + b.Length(v) - original.Length(v)
		if extra > padding {
			break
		}
		padding -= extra
		pl.length += extra
		handler := frames[index].Handler
		frames[index].Frame = a
		frames = append(frames, ackhandler.Frame{Frame: b, Handler: handler})
	}
	pings := min(protocol.ByteCount(randomRange(&p.rand, policy.MinPingFrames, policy.MaxPingFrames)), padding)
	for range int(pings) {
		frames = append(frames, ackhandler.Frame{Frame: &wire.PingFrame{}, Handler: emptyHandler{}})
	}
	padding -= pings
	pl.length += pings
	parts := make([]protectedInitialFrame, 0, 2*len(frames)+1)
	for _, frame := range frames {
		n := protocol.ByteCount(p.rand.Uint64N(uint64(padding) + 1))
		if n > 0 {
			parts = append(parts, protectedInitialFrame{padding: n})
			padding -= n
		}
		parts = append(parts, protectedInitialFrame{frame: frame})
	}
	if padding > 0 {
		parts = append(parts, protectedInitialFrame{padding: padding})
	}
	p.rand.Shuffle(len(parts), func(i, j int) { parts[i], parts[j] = parts[j], parts[i] })
	var err error
	if pl.ack != nil {
		raw, err = pl.ack.Append(raw, v)
		if err != nil {
			return nil, pl, err
		}
	}
	pl.frames = make([]ackhandler.Frame, 0, len(frames))
	for _, part := range parts {
		if part.padding > 0 {
			raw = append(raw, make([]byte, int(part.padding))...)
		} else {
			raw, err = part.frame.Frame.Append(raw, v)
			if err != nil {
				return nil, pl, err
			}
			pl.frames = append(pl.frames, part.frame)
		}
	}
	return raw, pl, nil
}
