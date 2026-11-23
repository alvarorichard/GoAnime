package alps

import (
	"errors"

	"github.com/enetx/g"
	"github.com/enetx/quic-go/quicvarint"
)

// DecodeHTTP3 parses ALPS without a control-stream prefix. Initial settings do
// not stand in for the definitive wire SETTINGS event or its ticket snapshot.
// https://github.com/google/quiche/blob/132bfd90ea370faeacee616c70d10210bb12ba66/quiche/quic/core/http/quic_spdy_session.cc#L76
// https://github.com/google/quiche/blob/132bfd90ea370faeacee616c70d10210bb12ba66/quiche/quic/core/http/http_decoder.cc#L150
func DecodeHTTP3(data []byte) (*Snapshot, error) {
	if len(data) > 65535 {
		return nil, errors.New("ALPS HTTP/3: TLS extension exceeds 65535 bytes")
	}

	snapshot := &Snapshot{hints: make(map[string]string), http3: g.NewMapOrd[uint64, uint64]()}
	seenSettings := false
	for len(data) != 0 {
		kind, n, err := quicvarint.Parse(data)
		if err != nil {
			return nil, errors.New("ALPS HTTP/3: truncated frame type")
		}

		data = data[n:]
		length, n, err := quicvarint.Parse(data)
		if err != nil {
			return nil, errors.New("ALPS HTTP/3: truncated frame length")
		}

		data = data[n:]
		if length > uint64(len(data)) {
			return nil, errors.New("ALPS HTTP/3: truncated frame payload")
		}

		payload := data[:int(length)]
		data = data[int(length):]
		switch kind {
		case 0, 1, 2, 3, 5, 6, 7, 8, 9, 0xd, 0x4d, 0xf0700:
			return nil, errors.New("ALPS HTTP/3: forbidden frame")
		case 4:
			if seenSettings {
				return nil, errors.New("ALPS HTTP/3: multiple SETTINGS frames")
			}

			seenSettings = true
			for len(payload) != 0 {
				id, n, err := quicvarint.Parse(payload)
				if err != nil {
					return nil, errors.New("ALPS HTTP/3: truncated setting ID")
				}

				payload = payload[n:]
				value, n, err := quicvarint.Parse(payload)
				if err != nil {
					return nil, errors.New("ALPS HTTP/3: truncated setting value")
				}

				if snapshot.http3.Contains(id) {
					return nil, errors.New("ALPS HTTP/3: duplicate setting ID")
				}

				snapshot.http3.Insert(id, value)
				payload = payload[n:]
			}
		case 0x89:
			for len(payload) != 0 {
				text, rest, valid := varintString(payload)
				if !valid {
					return nil, errors.New("ALPS HTTP/3: truncated ACCEPT_CH origin")
				}

				value, remaining, valid := varintString(rest)
				if !valid {
					return nil, errors.New("ALPS HTTP/3: truncated ACCEPT_CH value")
				}

				snapshot.add(text, value)
				payload = remaining
			}
		case 0xc:
			// Chromium's external default enables ORIGIN frame parsing. Its
			// ALPS visitor ignores entries, but malformed entry lengths still fail.
			for len(payload) != 0 {
				_, remaining, valid := uint16String(payload)
				if !valid {
					return nil, errors.New("ALPS HTTP/3: truncated ORIGIN entry")
				}

				payload = remaining
			}
		}
	}

	return snapshot, nil
}

func varintString(data []byte) (string, []byte, bool) {
	length, n, err := quicvarint.Parse(data)
	if err != nil || length > uint64(len(data)-n) {
		return "", nil, false
	}

	return string(data[n : n+int(length)]), data[n+int(length):], true
}
