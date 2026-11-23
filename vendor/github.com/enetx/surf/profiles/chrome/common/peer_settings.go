package common

import (
	"fmt"
	"math"

	"github.com/enetx/http/http2"
)

// InitialPeerSettings follows SpdySession::ParseAlps/HandleSetting, before the
// first wire SETTINGS. Chromium consumes the table, stream limit, send window
// and capability settings; frame/header limits and unknown IDs are ignored.
// The initial stream limit is capped at 256. An oversized initial send window
// is ignored rather than passed to the engine's ordinary wire validator.
// Returned settings are private; ordinary wire processing stays in the engine.
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/spdy/spdy_session.cc#L936
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/spdy/spdy_session.cc#L2352
func InitialPeerSettings(settings []http2.Setting) ([]http2.Setting, error) {
	var selected []http2.Setting
	for _, setting := range settings {
		switch setting.ID {
		case http2.SettingHeaderTableSize:
			selected = append(selected, setting)
		case http2.SettingEnableConnectProtocol:
			if setting.Val > 1 {
				return nil, fmt.Errorf("ALPS HTTP/2: invalid SETTINGS_ENABLE_CONNECT_PROTOCOL value %d", setting.Val)
			}

			selected = append(selected, setting)
		case http2.SettingMaxConcurrentStreams:
			setting.Val = min(setting.Val, 256)
			selected = append(selected, setting)
		case http2.SettingInitialWindowSize:
			if setting.Val <= math.MaxInt32 {
				selected = append(selected, setting)
			}
		case http2.SettingNoRFC7540Priorities:
			if setting.Val > 1 {
				return nil, fmt.Errorf("ALPS HTTP/2: invalid SETTINGS_DEPRECATE_HTTP2_PRIORITIES value %d", setting.Val)
			}

			selected = append(selected, setting)
		}
	}

	return selected, nil
}
