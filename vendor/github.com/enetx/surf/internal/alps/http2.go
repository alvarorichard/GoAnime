package alps

import (
	"encoding/binary"
	"errors"
	"sort"

	"github.com/enetx/http/http2"
)

// DecodeHTTP2 parses a complete ALPS frame sequence, without a connection
// preface. ALPS does not synthesize a received wire SETTINGS or its ACK.
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/spdy/alps_decoder.cc
func DecodeHTTP2(data []byte) (*Snapshot, error) {
	if len(data) > 65535 {
		return nil, errors.New("ALPS HTTP/2: TLS extension exceeds 65535 bytes")
	}

	snapshot := &Snapshot{hints: make(map[string]string)}
	settings := make(map[uint16]uint32)
	for len(data) != 0 {
		if len(data) < 9 {
			return nil, errors.New("ALPS HTTP/2: truncated frame header")
		}

		length := int(data[0])<<16 | int(data[1])<<8 | int(data[2])
		// Chromium's fresh decoder retains the default inbound frame limit;
		// a peer MAX_FRAME_SIZE setting controls the other direction only.
		if length > 16384 {
			return nil, errors.New("ALPS HTTP/2: frame exceeds 16384 bytes")
		}

		kind, flags := data[3], data[4]
		stream := binary.BigEndian.Uint32(data[5:9]) & 0x7fffffff
		data = data[9:]
		if length > len(data) {
			return nil, errors.New("ALPS HTTP/2: truncated frame payload")
		}

		payload := data[:length]
		data = data[length:]
		switch kind {
		case 0, 1, 2, 3, 5, 6, 7, 8, 9:
			return nil, errors.New("ALPS HTTP/2: forbidden core frame")
		case 4:
			if stream != 0 || flags&1 != 0 || length%6 != 0 {
				return nil, errors.New("ALPS HTTP/2: invalid SETTINGS frame")
			}

			for len(payload) != 0 {
				settings[binary.BigEndian.Uint16(payload)] = binary.BigEndian.Uint32(payload[2:])
				payload = payload[6:]
			}
		case 0x89:
			if stream != 0 || flags != 0 {
				return nil, errors.New("ALPS HTTP/2: invalid ACCEPT_CH stream or flags")
			}

			for len(payload) != 0 {
				text, rest, valid := uint16String(payload)
				if !valid {
					snapshot.bypassed = true
					break
				}

				value, remaining, valid := uint16String(rest)
				if !valid {
					snapshot.bypassed = true
					break
				}

				snapshot.add(text, value)
				payload = remaining
			}
		}
	}

	ids := make([]int, 0, len(settings))
	for id := range settings {
		ids = append(ids, int(id))
	}

	sort.Ints(ids)
	for _, id := range ids {
		snapshot.http2 = append(snapshot.http2, http2.Setting{ID: http2.SettingID(id), Val: settings[uint16(id)]})
	}

	return snapshot, nil
}

func uint16String(data []byte) (string, []byte, bool) {
	if len(data) < 2 {
		return "", nil, false
	}

	length := int(binary.BigEndian.Uint16(data))
	if length > len(data)-2 {
		return "", nil, false
	}

	return string(data[2 : 2+length]), data[2+length:], true
}
