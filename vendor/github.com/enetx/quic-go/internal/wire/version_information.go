package wire

import (
	"encoding/binary"
	"errors"
	"slices"

	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/quicvarint"
)

// VersionInformation is RFC 9368's authenticated version negotiation data.
// Client availability describes this first flight, not every implemented version.
type VersionInformation struct {
	Chosen    protocol.Version
	Available []protocol.Version
}

func ParseVersionInformation(data []byte, sentBy protocol.Perspective) (*VersionInformation, error) {
	if len(data) < 4 || len(data)%4 != 0 {
		return nil, errors.New("invalid version_information length")
	}
	v := &VersionInformation{Chosen: protocol.Version(binary.BigEndian.Uint32(data))}
	if v.Chosen == 0 {
		return nil, errors.New("zero chosen version")
	}
	for data = data[4:]; len(data) > 0; data = data[4:] {
		version := protocol.Version(binary.BigEndian.Uint32(data))
		if version == 0 {
			return nil, errors.New("zero available version")
		}
		v.Available = append(v.Available, version)
	}
	if sentBy == protocol.PerspectiveClient && !slices.Contains(v.Available, v.Chosen) {
		return nil, errors.New("chosen version absent from client availability")
	}
	return v, nil
}

func (v *VersionInformation) appendParameter(b []byte) []byte {
	b = quicvarint.Append(b, 0x11)
	b = quicvarint.Append(b, uint64(4+4*len(v.Available)))
	b = binary.BigEndian.AppendUint32(b, uint32(v.Chosen))
	for _, version := range v.Available {
		b = binary.BigEndian.AppendUint32(b, uint32(version))
	}
	return b
}
