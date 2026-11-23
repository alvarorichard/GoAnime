package tlscache

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"io"

	"github.com/andybalholm/brotli"
)

const maxCompressedRecord = 256 << 10

// Storage uses the public Go TLS state, rather than pretending that NSS's opaque
// token and certificate metadata have the same representation. Compression uses
// the source browser's quality/window; the accounting model is documented.
func encodeRecord(state *tls.ClientSessionState, quality int) ([]byte, int, error) {
	label, session, err := state.ResumptionState()
	if err != nil || session == nil {
		return nil, 0, errors.New("invalid TLS cache session")
	}

	serialized, err := session.Bytes()
	if err != nil {
		return nil, 0, err
	}

	length := 4 + len(label) + len(serialized)
	if quality < 0 {
		return nil, length, nil
	}
	if length > maxCompressedRecord {
		return nil, 0, errors.New("TLS cache record exceeds compression limit")
	}

	payload := make([]byte, length)
	binary.LittleEndian.PutUint32(payload, uint32(len(label)))
	copy(payload[4:], label)
	copy(payload[4+len(label):], serialized)

	var compressed bytes.Buffer
	binary.Write(&compressed, binary.LittleEndian, uint32(length))
	writer := brotli.NewWriterOptions(&compressed, brotli.WriterOptions{Quality: quality, LGWin: 22})

	if _, err := writer.Write(payload); err != nil {
		return nil, 0, err
	}
	if err := writer.Close(); err != nil {
		return nil, 0, err
	}

	// Drop the buffer's spare capacity; retained storage follows the encoded
	// payload length rather than an intermediate allocation's growth factor.
	encoded := make([]byte, compressed.Len())
	copy(encoded, compressed.Bytes())

	return encoded, len(encoded), nil
}

func decodeRecord(record ticket) (*tls.ClientSessionState, error) {
	if record.state != nil {
		return record.state, nil
	}

	if len(record.encoded) < 4 {
		return nil, errors.New("truncated TLS cache record")
	}

	length := int(binary.LittleEndian.Uint32(record.encoded[:4]))
	if length < 4 || length > maxCompressedRecord {
		return nil, errors.New("invalid TLS cache record length")
	}

	reader := brotli.NewReader(bytes.NewReader(record.encoded[4:]))
	data, err := io.ReadAll(io.LimitReader(reader, int64(length)+1))
	if err != nil || len(data) != length {
		return nil, errors.New("invalid compressed TLS cache record")
	}

	labelSize := int(binary.LittleEndian.Uint32(data[:4]))
	if labelSize <= 0 || labelSize > len(data)-4 {
		return nil, errors.New("invalid TLS cache ticket length")
	}

	state, err := tls.ParseSessionState(data[4+labelSize:])
	if err != nil {
		return nil, err
	}

	return tls.NewResumptionState(data[4:4+labelSize], state)
}
