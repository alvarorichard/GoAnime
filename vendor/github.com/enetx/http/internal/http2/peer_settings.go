package http2

import "slices"

// initPeerSettings runs before the connection is published, starts its read loop
// or writes its preface. It deliberately leaves wire SETTINGS/ACK state alone.
func (cc *ClientConn) initPeerSettings() error {
	if cc.fingerprint == nil || cc.fingerprint.InitialPeerSettings == nil {
		return nil
	}

	settings, err := cc.fingerprint.InitialPeerSettings(cc.tconn)
	if err != nil {
		return err
	}

	for _, setting := range slices.Clone(settings) {
		if err := cc.applyPeerSetting(setting); err != nil {
			return err
		}
	}

	return nil
}

// applyPeerSetting is shared by initialization and received wire SETTINGS.
// The read loop holds both connection locks; initialization is not yet shared.
func (cc *ClientConn) applyPeerSetting(setting Setting) error {
	if err := setting.Valid(); err != nil {
		return err
	}

	switch setting.ID {
	case SettingMaxFrameSize:
		cc.maxFrameSize = setting.Val
	case SettingMaxConcurrentStreams:
		cc.maxConcurrentStreams = setting.Val
		cc.peerMaxConcurrentStreamsSet = true
		cc.cond.Broadcast()
	case SettingMaxHeaderListSize:
		cc.peerMaxHeaderListSize = uint64(setting.Val)
	case SettingInitialWindowSize:
		// Valid has already enforced the same maximum as changeInitialWindowSize.
		cc.flow.changeInitialWindowSize(int64(setting.Val))
		cc.cond.Broadcast()
	case SettingHeaderTableSize:
		cc.setFingerprintEncoderTableSize(setting.Val)
		cc.peerMaxHeaderTableSize = setting.Val
	case SettingEnableConnectProtocol:
		// Extended CONNECT still waits for the first actual wire SETTINGS.
		// Later frames cannot change the capability, matching native behavior.
		if !cc.seenSettings {
			cc.extendedConnectAllowed = setting.Val == 1
		}
	default:
		cc.vlogf("Unhandled Setting: %v", setting)
	}

	return nil
}
