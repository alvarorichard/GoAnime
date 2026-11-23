package http3

import (
	"fmt"

	"github.com/enetx/g"
)

// Preserve the explicit wire position of H3_DATAGRAM when it is already in the
// ordered settings. The runtime capability is still controlled by EnableDatagrams.
func initialClientSettings(datagrams bool, settings g.MapOrd[uint64, uint64]) *settingsFrame {
	if settings.Contains(settingDatagram) {
		datagrams = false
	}
	return &settingsFrame{Datagram: datagrams, Other: settings, MaxFieldSectionSize: -1}
}

func (t *Transport) validateOrderedSettings() error {
	for id, value := range t.AdditionalSettings.Iter() {
		if id > 1<<62-1 || value > 1<<62-1 {
			return fmt.Errorf("http3: setting %#x exceeds the QUIC variable integer range", id)
		}
		if id >= 2 && id <= 5 {
			return fmt.Errorf("http3: setting %#x is reserved by HTTP/3", id)
		}
		if id == settingDatagram || id == settingExtendedConnect {
			if value > 1 {
				return fmt.Errorf("http3: setting %#x must be 0 or 1", id)
			}
		}
		if id == settingDatagram && t.EnableDatagrams && value != 1 {
			return fmt.Errorf("http3: ordered H3_DATAGRAM setting disables enabled datagrams")
		}
	}
	return nil
}
