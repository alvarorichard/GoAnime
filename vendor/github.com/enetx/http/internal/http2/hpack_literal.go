package http2

import (
	"fmt"

	"golang.org/x/net/http2/hpack"
)

// Only literal framing is adapted. Table storage, eviction and Huffman codes
// continue to come from the official x/net codec.

// RFC 7541 4.4 empties the table on an oversized incremental insertion. The
// upstream encoder only attempts inserts that fit, so clear it through its
// public capacity API after emitting that literal. Real pending size updates
// have already been emitted by WriteField. Drain the reset's artificial updates
// into the private buffer with a non-inserting field, then discard those bytes;
// neither a reset notification nor the drain field is sent to the peer.
func (cc *ClientConn) clearFingerprintEncoderTable() {
	capacity := cc.henc.MaxDynamicTableSize()
	if capacity == 0 {
		return
	}

	cc.henc.SetMaxDynamicTableSize(0)
	cc.henc.SetMaxDynamicTableSize(capacity)
	end := cc.hbuf.Len()
	// Like writeFingerprintHeader, this encoder writes only to cc.hbuf.
	cc.henc.WriteField(hpack.HeaderField{Name: ":method", Value: "GET", Sensitive: true})
	cc.hbuf.Truncate(end)
	if cc.fingerprintTable != nil {
		cc.fingerprintTable.resize(0)
		cc.fingerprintTable.resize(capacity)
	}
}

// finishFingerprintHeader validates and adapts one encoded field after policy
// selection. Keep wire framing here so codec changes can be checked against
// independent byte vectors without replacing the official dynamic encoder.
func (cc *ClientConn) finishFingerprintHeader(
	start int,
	field hpack.HeaderField,
	rep HPACKRepresentation,
	config *HPACKConfig,
	nameReference uint64,
) error {
	name, value := field.Name, field.Value
	data := skipHPACKSizeUpdates(cc.hbuf.Bytes()[start:])
	oversizedInsert := rep == HPACKIncrementalIndexing && len(data) > 0 && data[0]&0x80 == 0 && uint64(len(name))+uint64(len(value))+32 > uint64(cc.henc.MaxDynamicTableSize())

	if config != nil && len(data) > 0 && data[0]&0x80 == 0 && (config.Huffman != HPACKHuffmanDefault || config.NameReferencePolicy != HPACKNameReferenceDefault || oversizedInsert) {
		prefix := uint8(4)
		if data[0]&0xc0 == 0x40 {
			prefix = 6
		}
		index, err := readHPACKIndex(data, prefix)
		if err != nil {
			return err
		}

		if config.NameReferencePolicy == HPACKNameReferenceLast {
			index = nameReference
		} else if config.NameReferencePolicy == HPACKNameReferenceFirst {
			if first := staticHPACKFirstNames()[name]; first != 0 {
				index = first
			}
		}

		flags := data[0] & byte(0xff<<prefix)
		if rep == HPACKWithoutIndexing {
			flags &^= 0x10
		} else if oversizedInsert {
			prefix, flags = 6, 0x40
		}

		literal := appendHPACKIndex(nil, prefix, flags, index)
		if index == 0 {
			literal = appendHPACKString(literal, name, config.Huffman)
		}
		literal = appendHPACKString(literal, value, config.Huffman)
		cc.hbuf.Truncate(cc.hbuf.Len() - len(data))
		cc.hbuf.Write(literal)
		data = skipHPACKSizeUpdates(cc.hbuf.Bytes()[start:])
	}

	if oversizedInsert {
		cc.clearFingerprintEncoderTable()
	} else if cc.fingerprintTable != nil && len(data) > 0 && data[0]&0xc0 == 0x40 {
		cc.fingerprintTable.add(field)
	}

	if rep == HPACKWithoutIndexing && (config == nil || config.Huffman == HPACKHuffmanDefault && config.NameReferencePolicy == HPACKNameReferenceDefault) {
		// Sensitive suppresses exact matches and insertion in the official encoder.
		// Both literal forms share the four-bit name index and string encoding;
		// only the never-index bit differs. Skip any pending table size updates.
		if len(data) == 0 || data[0]&0xf0 != 0x10 {
			return fmt.Errorf("http2: unexpected HPACK literal encoding")
		}

		data[0] &^= 0x10
	}

	return nil
}

func readHPACKIndex(data []byte, prefix uint8) (uint64, error) {
	mask := uint64(1<<prefix) - 1
	index := uint64(data[0]) & mask
	if index < mask {
		return index, nil
	}
	for i, shift := 1, uint(0); i < len(data) && shift < 63; i, shift = i+1, shift+7 {
		index += uint64(data[i]&127) << shift
		if data[i]&128 == 0 {
			return index, nil
		}
	}
	return 0, fmt.Errorf("http2: invalid HPACK name index")
}

func appendHPACKIndex(dst []byte, prefix uint8, flags byte, index uint64) []byte {
	mask := uint64(1<<prefix) - 1
	if index < mask {
		return append(dst, flags|byte(index))
	}
	dst = append(dst, flags|byte(mask))
	index -= mask
	for index >= 128 {
		dst = append(dst, byte(index)|128)
		index >>= 7
	}
	return append(dst, byte(index))
}

func appendHPACKString(dst []byte, value string, mode HPACKHuffmanMode) []byte {
	n := hpack.HuffmanEncodeLength(value)
	if mode == HPACKHuffmanAlways || mode == HPACKHuffmanDefault && n < uint64(len(value)) {
		dst = appendHPACKIndex(dst, 7, 128, n)
		return hpack.AppendHuffmanString(dst, value)
	}
	dst = appendHPACKIndex(dst, 7, 0, uint64(len(value)))
	return append(dst, value...)
}
