package http3

import (
	"bytes"

	"github.com/enetx/quic-go/quicvarint"
)

const serverSettingsSessionTicketPrefix = "quic-go h3 settings v1"

type settings = settingsFrame

func settingsForSessionTicket(current *settings) []byte {
	return current.Append([]byte(serverSettingsSessionTicketPrefix))
}

func settingsDataFromSessionTicket(extras [][]byte) ([]byte, bool) {
	prefix := []byte(serverSettingsSessionTicketPrefix)
	for _, extra := range extras {
		if data, ok := bytes.CutPrefix(extra, prefix); ok {
			return data, true
		}
	}
	return nil, false
}

func settingsCompatibleFor0RTT(data []byte, current *settings) bool {
	r := bytes.NewReader(data)
	frameType, err := quicvarint.Read(r)
	if err != nil || frameType != 0x4 {
		return false
	}
	length, err := quicvarint.Read(r)
	if err != nil || uint64(r.Len()) != length {
		return false
	}
	old, err := parseSettingsFrame(&countingByteReader{Reader: r}, length, 0, nil)
	if err != nil {
		return false
	}
	// An unknown setting might belong to an extension that this version no longer supports.
	for _, field := range old.Other {
		id := field.Key
		if id != 1 && id != 7 && !(id >= 33 && (id-33)%31 == 0) {
			return false
		}
	}
	if previous := old.Other.Get(1).UnwrapOrDefault(); previous > 0 && previous != current.Other.Get(1).UnwrapOrDefault() {
		return false
	}
	if old.Other.Get(7).UnwrapOrDefault() > current.Other.Get(7).UnwrapOrDefault() {
		return false
	}

	if current.MaxFieldSectionSize > 0 && (old.MaxFieldSectionSize < 0 || old.MaxFieldSectionSize > current.MaxFieldSectionSize) {
		return false
	}
	if old.Datagram && !current.Datagram {
		return false
	}
	if old.ExtendedConnect && !current.ExtendedConnect {
		return false
	}
	return true
}
