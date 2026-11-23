package http3

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/quic-go/qpack"
	"golang.org/x/net/http2/hpack"
)

// Configured clients collect fields for atomic connection-level encoding.
// The native path retains the official static encoder; its small adapter changes
// only the N bit or expands an exact reference into a literal when necessary.
func (w *requestWriter) writeQPACKField(name, value string) error {
	if w.encodeQPACK != nil || w.collectBodyFields {
		w.qpackFields = append(w.qpackFields, qpack.HeaderField{Name: name, Value: value})
		if w.encodeQPACK != nil {
			return nil
		}
	}
	start := w.headerBuf.Len()
	if err := w.encoder.WriteField(qpack.HeaderField{Name: name, Value: value}); err != nil {
		return err
	}
	if w.qpackConfig == nil || !slices.Contains(w.qpackConfig.NeverIndex, name) {
		return nil
	}
	if start == 0 {
		start = 2
	} // Required Insert Count and Delta Base are zero.
	return markQPACKFieldNeverIndexed(w.headerBuf, start, value)
}

// markQPACKFieldNeverIndexed rewrites one field representation, preserving all
// preceding fields. Keeping wire validation here also covers malformed encoder
// output independently of the official encoder's current static table.
func markQPACKFieldNeverIndexed(buffer *bytes.Buffer, start int, value string) error {
	data := buffer.Bytes()[start:]
	if len(data) == 0 {
		return fmt.Errorf("http3: empty QPACK field")
	}
	switch {
	case data[0]&0xc0 == 0xc0: // Indexed static entry -> literal static name reference.
		index := uint64(data[0] & 63)
		if index == 63 {
			for i, shift := 1, uint(0); ; i, shift = i+1, shift+7 {
				if i >= len(data) || shift >= 63 {
					return fmt.Errorf("http3: invalid QPACK static index")
				}
				index += uint64(data[i]&127) << shift
				if data[i]&128 == 0 {
					break
				}
			}
		}
		literal := appendQPACKInt(nil, 4, 0x70, index)
		literal = appendQPACKInt(literal, 7, 0x80, hpack.HuffmanEncodeLength(value))
		literal = hpack.AppendHuffmanString(literal, value)
		buffer.Truncate(start)
		_, err := buffer.Write(literal)
		return err
	case data[0]&0xf0 == 0x50: // Literal with static name reference.
		data[0] |= 0x20
	case data[0]&0xe0 == 0x20: // Literal with literal name.
		data[0] |= 0x10
	default:
		return fmt.Errorf("http3: unexpected QPACK static encoding")
	}
	return nil
}

func appendQPACKInt(dst []byte, prefix uint8, flags byte, value uint64) []byte {
	mask := uint64(1<<prefix) - 1
	if value < mask {
		return append(dst, flags|byte(value))
	}
	dst = append(dst, flags|byte(mask))
	value -= mask
	for value >= 128 {
		dst = append(dst, byte(value)|128)
		value >>= 7
	}
	return append(dst, byte(value))
}
