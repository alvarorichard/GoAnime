package handshake

import (
	"bytes"
	"encoding/binary"
	"time"

	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/quicvarint"
)

// ClientSessionInfo is a snapshot of QUIC's locally remembered transport state.
type ClientSessionInfo struct {
	Version             protocol.Version
	RTT                 time.Duration
	ConnectionID        []byte
	TransportParameters []byte
}

// SessionInfo leaves TLS encoding and ticket validation to the TLS engine.
func SessionInfo(extras [][]byte) (ClientSessionInfo, bool) {
	var result ClientSessionInfo
	found := false
	for _, extra := range extras {
		data, ok := bytes.CutPrefix(extra, []byte(extraPrefix))
		if !ok {
			continue
		}
		revision, width, err := quicvarint.Parse(data)
		if err != nil || (revision != 6 && revision != clientSessionStateRevision) || len(data)-width < 4 || found {
			return ClientSessionInfo{}, false
		}
		data = data[width:]
		result.Version = protocol.Version(binary.BigEndian.Uint32(data))
		data = data[4:]
		if result.Version != protocol.Version1 && result.Version != protocol.Version2 {
			return ClientSessionInfo{}, false
		}
		if revision == clientSessionStateRevision {
			rtt, width, err := quicvarint.Parse(data)
			if err != nil {
				return ClientSessionInfo{}, false
			}
			result.RTT = time.Duration(rtt)
			data = data[width:]
			cidLen, width, err := quicvarint.Parse(data)
			if err != nil {
				return ClientSessionInfo{}, false
			}
			data = data[width:]
			if cidLen > protocol.MaxConnIDLen || cidLen > uint64(len(data)) {
				return ClientSessionInfo{}, false
			}
			result.ConnectionID = bytes.Clone(data[:cidLen])
			data = data[cidLen:]
		}
		result.TransportParameters = bytes.Clone(data)
		found = true
	}
	return result, found
}

// SessionVersion reads the negotiated version from local session metadata.
func SessionVersion(extras [][]byte) (protocol.Version, bool) {
	info, ok := SessionInfo(extras)
	return info.Version, ok
}
