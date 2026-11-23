package quic

import (
	"encoding/binary"
	"fmt"
	"slices"

	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/wire"
)

// CryptoRange selects Initial CRYPTO bytes by their absolute stream offset.
// Length must be positive. Offsets and lengths exclude CRYPTO frame headers.
type CryptoRange struct {
	// Offset is the zero-based position in the Initial CRYPTO stream.
	Offset int
	// Length is the number of ClientHello bytes in this range.
	Length int
}

// InitialCryptoLayoutInfo describes the first complete ClientHello and the
// packet available to carry it. InitialCryptoLayout runs before each packet
// while the first ClientHello has unsent bytes. Its result must cover Remaining
// exactly once; empty, overlapping, out-of-bounds or incomplete layouts fail
// before serialization. The engine sends only the first group, then invokes
// the callback again with the actual unsent ranges and next packet's budgets.
// A group may be split if it exceeds the frame budget. Nil initially retains
// native preparation; after a custom group it selects sequential unsent ranges.
// Subsequent TLS messages retain native first-send behavior. Optional
// InitialCryptoRecoveryLayout independently selects retransmitted byte ranges.
type InitialCryptoLayoutInfo struct {
	// Length includes the four-byte TLS handshake header. A recovery callback
	// receives the retained Initial stream length, including later messages.
	Length int
	// SNIOffset is the hostname position, or -1 when SNI is absent.
	SNIOffset int
	// SNILength is the hostname length, or zero when SNI is absent.
	SNILength int
	// ECHOffset is the extension-type position, or -1 when ECH is absent.
	ECHOffset int
	// ECHLength includes the extension type, length and value, or zero if absent.
	ECHLength int
	// PacketSizeLimit includes the QUIC header but excludes AEAD expansion.
	PacketSizeLimit int
	// FrameBudget is the space left after the QUIC header and any ACK frame.
	FrameBudget int
	// Remaining is a private snapshot of unsent ClientHello ranges in ascending
	// offset order. It can be changed by the callback without affecting recovery.
	Remaining []CryptoRange
}

type initialCryptoLayoutState struct {
	info      InitialCryptoLayoutInfo
	data      []byte
	remaining []CryptoRange
}

func completeInitialClientHello(data []byte) bool {
	if len(data) < 4 {
		return false
	}
	if data[0] != 1 {
		return true // let native preparation diagnose another handshake type
	}
	return len(data) >= 4+(int(data[1])<<16)+(int(data[2])<<8)+int(data[3])
}

func (s *initialCryptoStream) prepareInitialCryptoLayout(packetSizeLimit, frameBudget protocol.ByteCount) error {
	if s.layout == nil || s.layoutPrepared && s.layoutState == nil {
		return nil
	}
	state := s.layoutState
	if !s.layoutPrepared {
		s.layoutPrepared = true
		if s.writeOffset != 0 || !completeInitialClientHello(s.writeBuf) || s.writeBuf[0] != 1 {
			_, err := s.Write(nil)
			return err
		}
		total := 4 + (int(s.writeBuf[1]) << 16) + (int(s.writeBuf[2]) << 8) + int(s.writeBuf[3])
		hello := s.writeBuf[:total]
		sniOffset, sniLength, echOffset, err := findSNIAndECH(hello)
		if err != nil {
			return err
		}
		if sniOffset < 0 {
			sniLength = 0
		}
		var echLength int
		if echOffset >= 0 {
			echLength = 4 + int(binary.BigEndian.Uint16(hello[echOffset+2:]))
		}
		state = &initialCryptoLayoutState{
			info: InitialCryptoLayoutInfo{Length: total, SNIOffset: sniOffset, SNILength: sniLength, ECHOffset: echOffset, ECHLength: echLength},
			data: hello, remaining: []CryptoRange{{Length: total}},
		}
	}
	info := state.info
	info.PacketSizeLimit, info.FrameBudget = int(packetSizeLimit), int(frameBudget)
	info.Remaining = slices.Clone(state.remaining)
	layout := s.layout(info)
	if layout == nil {
		if s.layoutState == nil {
			_, err := s.Write(nil)
			return err
		}
		layout = [][]CryptoRange{slices.Clone(state.remaining)}
	}
	frames, err := initialCryptoLayoutFrames(state.data, layout, state.remaining)
	if err != nil {
		return err
	}
	if s.protection == nil {
		s.protection = &initialProtectionState{}
	}
	s.protection.prepared = true
	s.protection.ordered = true
	s.protection.active = frames[0]
	s.protection.pending = nil
	s.scramble = false
	if s.layoutState == nil {
		s.layoutState = state
		s.writeOffset = protocol.ByteCount(state.info.Length)
		s.writeBuf = s.writeBuf[state.info.Length:]
	}
	return nil
}

func initialCryptoLayoutFrames(hello []byte, layout [][]CryptoRange, remaining []CryptoRange) ([][]*wire.CryptoFrame, error) {
	if len(layout) == 0 {
		return nil, fmt.Errorf("Initial CRYPTO layout is empty")
	}
	frames := make([][]*wire.CryptoFrame, len(layout))
	var coverage []CryptoRange
	for i, group := range layout {
		if len(group) == 0 {
			return nil, fmt.Errorf("Initial CRYPTO layout packet %d is empty", i)
		}
		frames[i] = make([]*wire.CryptoFrame, len(group))
		for j, part := range group {
			if part.Offset < 0 || part.Length <= 0 || part.Offset > len(hello) || part.Length > len(hello)-part.Offset {
				return nil, fmt.Errorf("Initial CRYPTO layout range %d/%d is out of bounds", i, j)
			}
			coverage = append(coverage, part)
			frames[i][j] = &wire.CryptoFrame{Offset: protocol.ByteCount(part.Offset), Data: hello[part.Offset : part.Offset+part.Length]}
		}
	}
	slices.SortFunc(coverage, func(a, b CryptoRange) int { return a.Offset - b.Offset })
	end := remaining[0].Offset
	var next int
	for _, part := range coverage {
		if next == len(remaining) || part.Offset != end || part.Length > remaining[next].Offset+remaining[next].Length-end {
			return nil, fmt.Errorf("Initial CRYPTO layout overlaps or omits ClientHello bytes at %d", end)
		}
		end += part.Length
		if end == remaining[next].Offset+remaining[next].Length {
			next++
			if next < len(remaining) {
				end = remaining[next].Offset
			}
		}
	}
	if next != len(remaining) {
		return nil, fmt.Errorf("Initial CRYPTO layout omits ClientHello bytes at %d", end)
	}
	return frames, nil
}

func (s *initialCryptoStream) consumeCryptoLayoutRange(frame *wire.CryptoFrame) {
	state := s.layoutState
	if state == nil {
		return
	}
	start, end := int(frame.Offset), int(frame.Offset)+len(frame.Data)
	remaining := make([]CryptoRange, 0, len(state.remaining)+1)
	for _, part := range state.remaining {
		partEnd := part.Offset + part.Length
		if end <= part.Offset || start >= partEnd {
			remaining = append(remaining, part)
			continue
		}
		if part.Offset < start {
			remaining = append(remaining, CryptoRange{Offset: part.Offset, Length: start - part.Offset})
		}
		if end < partEnd {
			remaining = append(remaining, CryptoRange{Offset: end, Length: partEnd - end})
		}
	}
	state.remaining = remaining
	if len(remaining) == 0 {
		s.layoutState = nil
	}
}
