package http3

import (
	"fmt"

	"github.com/enetx/g"
	"github.com/enetx/quic-go/quicvarint"
)

// Per-connection reserved settings never change QPACK or protocol capabilities.
// Clone only after validating the factory output; caller settings stay immutable.
func connectionSettingsGrease(settings g.MapOrd[uint64, uint64], factory func() (uint64, uint64, error)) (g.MapOrd[uint64, uint64], error) {
	if factory == nil {
		return settings, nil
	}
	id, value, err := factory()
	if err != nil {
		return nil, err
	}
	if id < 0x21 || id > quicvarint.Max || (id-0x21)%0x1f != 0 || value > quicvarint.Max {
		return nil, fmt.Errorf("http3: invalid reserved SETTINGS identifier/value: %#x/%d", id, value)
	}
	if settings.Contains(id) {
		return nil, fmt.Errorf("http3: duplicate reserved setting %#x", id)
	}
	snapshot := settings.Clone()
	snapshot.Insert(id, value)
	return snapshot, nil
}
