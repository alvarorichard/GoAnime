// Package qpack adds RFC 9204 dynamic tables around the official static table
// and Huffman codec. It contains no browser presets. Limits come from HTTP/3
// SETTINGS and explicit local encoder/decoder budgets.
package qpack

import (
	"bufio"
	"errors"
	"io"
	"strings"

	base "github.com/quic-go/qpack"
	"golang.org/x/net/http2/hpack"
)

const maxInt = uint64(1<<62 - 1)

// AppendEncoderCapacity appends an RFC 9204 Set Dynamic Table Capacity instruction.
func AppendEncoderCapacity(dst []byte, capacity uint64) []byte {
	return appendInt(dst, 5, 0x20, capacity)
}

var ErrDecompression = errors.New("QPACK decompression failed")
var ErrEncoderStream = errors.New("QPACK encoder stream error")
var ErrFieldSectionTooLarge = errors.New("QPACK decoded field section exceeds limit")

func readInt(r io.ByteReader, first byte, prefix uint8) (uint64, error) {
	mask := uint64(1<<prefix) - 1
	value := uint64(first) & mask
	if value < mask {
		return value, nil
	}
	for shift := uint(0); shift < 63; shift += 7 {
		b, err := r.ReadByte()
		if err != nil {
			return 0, io.ErrUnexpectedEOF
		}
		part := uint64(b & 127)
		if part > (maxInt-value)>>shift {
			return 0, ErrDecompression
		}
		value += part << shift
		if b&128 == 0 {
			return value, nil
		}
	}
	return 0, ErrDecompression
}
func appendInt(dst []byte, prefix uint8, flags byte, value uint64) []byte {
	mask := uint64(1<<prefix) - 1
	if value < mask {
		return append(dst, flags|byte(value))
	}
	dst = append(dst, flags|byte(mask))
	value -= mask
	for value >= 128 {
		dst = append(dst, byte(value&127)|128)
		value >>= 7
	}
	return append(dst, byte(value))
}
func readString(r interface {
	io.Reader
	io.ByteReader
}, first byte, prefix uint8, huffman bool, limit uint64) (string, error) {
	length, err := readInt(r, first, prefix)
	if err != nil {
		return "", err
	}
	// Every Huffman symbol needs at least five bits. Bound encoded allocation too.
	encodedLimit := limit
	if huffman {
		if limit > (maxInt-4)/4 {
			encodedLimit = maxInt
		} else {
			encodedLimit = limit*4 + 4
		}
	}
	if length > encodedLimit || length > uint64(int(^uint(0)>>1)) {
		return "", ErrFieldSectionTooLarge
	}
	data := make([]byte, int(length))
	if _, err := io.ReadFull(r, data); err != nil {
		return "", io.ErrUnexpectedEOF
	}
	if !huffman {
		return string(data), nil
	}
	var output strings.Builder
	writer := &boundedWriter{writer: &output, left: limit}
	if _, err := hpack.HuffmanDecode(writer, data); err != nil {
		return "", err
	}
	return output.String(), nil
}

type boundedWriter struct {
	writer io.Writer
	left   uint64
}

func (w *boundedWriter) Write(p []byte) (int, error) {
	if uint64(len(p)) > w.left {
		return 0, ErrFieldSectionTooLarge
	}
	w.left -= uint64(len(p))
	return w.writer.Write(p)
}
func staticField(index uint64) (base.HeaderField, error) {
	encoded := appendInt([]byte{0, 0}, 6, 0xc0, index)
	return base.NewDecoder().Decode(encoded)()
}

// ReadDecoder validates feedback for the unchanged static-only request encoder.
// There are no dynamic references or insertions to acknowledge. Cancellation is
// valid even when there were no references on that stream (RFC 9204 section 4.4).
func ReadDecoder(input io.Reader) error {
	reader := bufio.NewReader(input)
	for {
		first, err := reader.ReadByte()
		if err != nil {
			return err
		}
		prefix := uint8(6)
		if first&0x80 != 0 {
			prefix = 7
		}
		_, err = readInt(reader, first, prefix)
		if err != nil {
			return errors.Join(ErrDecoderStream, err)
		}
		if first&0xc0 != 0x40 {
			return ErrDecoderStream
		}
	}
}

var ErrDecoderStream = errors.New("QPACK decoder stream error")
