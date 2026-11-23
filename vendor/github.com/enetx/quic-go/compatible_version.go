package quic

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/enetx/quic-go/internal/handshake"
	"github.com/enetx/quic-go/internal/protocol"
	"github.com/enetx/quic-go/internal/qerr"
	"github.com/enetx/quic-go/internal/wire"
	"github.com/enetx/quic-go/qlog"
)

// VersionNegotiationErrorCode is RFC 9368's VERSION_NEGOTIATION_ERROR.
const VersionNegotiationErrorCode = qerr.VersionNegotiationErrorCode

// Compatible switching is a client mechanism. Neither the default client nor
// server changes its negotiation policy. See RFC 9368 §§3–4 and RFC 9369 §4.1.
type compatibleVersionState struct {
	chosen    protocol.Version
	offered   *wire.VersionInformation
	confirmed bool
}

func (c *Conn) compatibleVersionEnabled() bool {
	return c.perspective == protocol.PerspectiveClient && c.config.ClientHandshake != nil && c.config.ClientHandshake.CompatibleVersionNegotiation
}

func (c *Conn) configureVersionInformation(params *wire.TransportParameters) error {
	info := &wire.VersionInformation{Chosen: c.version}
	if layout := params.ClientLayout; layout != nil {
		for i, parameter := range layout.Additional {
			if parameter.ID != 0x11 {
				continue
			}
			var err error
			info, err = wire.ParseVersionInformation(parameter.Value, protocol.PerspectiveClient)
			if err != nil {
				return err
			}
			// ClientLayout is private. Promote the raw extension to an engine-owned
			// parameter without changing its selected position or GREASE bytes.
			layout.Additional = slices.Delete(layout.Additional, i, i+1)
			break
		}
	}
	if len(info.Available) == 0 {
		for _, version := range c.config.Versions {
			if version == c.version || protocol.CanSwitchVersion(c.version, version) {
				info.Available = append(info.Available, version)
			}
		}
	}
	if info.Chosen != c.version || !slices.Contains(info.Available, c.version) {
		return fmt.Errorf("version_information does not describe the chosen version")
	}
	for _, version := range info.Available {
		if uint32(version)&0x0f0f0f0f == 0x0a0a0a0a {
			continue
		}
		if !slices.Contains(c.config.Versions, version) || version != c.version && !protocol.CanSwitchVersion(c.version, version) {
			return fmt.Errorf("version_information advertises an unavailable compatible version: %x", uint32(version))
		}
	}
	params.VersionInformation = info
	params.CompatibleVersionNegotiation = true
	c.compatibleVersion = &compatibleVersionState{chosen: c.version, offered: info}
	return nil
}

type compatibleCryptoSetup interface {
	InitialOpenerForVersion(protocol.Version) (handshake.LongHeaderOpener, error)
	SwitchVersion(protocol.Version) error
}

// acceptCompatibleVersion never commits to unauthenticated header bytes. A
// reordered Handshake can enter the normal bounded undecryptable-packet queue.
func (c *Conn) acceptCompatibleVersion(hdr *wire.Header, data []byte) bool {
	state := c.compatibleVersion
	if state == nil || state.confirmed || !protocol.CanSwitchVersion(c.version, hdr.Version) || !slices.Contains(state.offered.Available, hdr.Version) {
		return false
	}
	if hdr.Type == protocol.PacketTypeHandshake {
		return true
	}
	if hdr.Type != protocol.PacketTypeInitial || c.receivedFirstPacket && hdr.SrcConnectionID != c.handshakeDestConnID {
		return false
	}
	cs, ok := c.cryptoStreamHandler.(compatibleCryptoSetup)
	if !ok {
		return false
	}
	opener, err := cs.InitialOpenerForVersion(hdr.Version)
	if err != nil {
		return false
	}
	// Header and payload unprotection mutate their buffer. A failed candidate
	// must not alter either the actual packet or the active Initial PN decoder.
	_, _, err = (&packetUnpacker{}).unpackLongHeaderPacket(opener, hdr, bytes.Clone(data))
	if err != nil && err != wire.ErrInvalidReservedBits {
		return false
	}
	if err := cs.SwitchVersion(hdr.Version); err != nil {
		return false
	}
	c.version = hdr.Version
	c.connStateMutex.Lock()
	c.connState.Version = hdr.Version
	c.connStateMutex.Unlock()
	state.confirmed = true
	if c.baseTokenStoreKey != "" {
		c.tokenStoreKey = versionedTokenKey(c.baseTokenStoreKey, c.version)
	}
	if c.qlogger != nil {
		c.qlogger.RecordEvent(qlog.VersionInformation{ChosenVersion: c.version, ClientVersions: c.config.Versions})
	}
	return true
}

func (c *Conn) checkVersionInformation(params *wire.TransportParameters) error {
	state := c.compatibleVersion
	if state == nil {
		return nil
	}
	fail := func(message string) error {
		return &qerr.TransportError{ErrorCode: qerr.VersionNegotiationErrorCode, ErrorMessage: message}
	}
	info := params.VersionInformation
	if info == nil {
		if c.version != state.chosen {
			return fail("compatible version switch without authenticated version_information")
		}
		if !c.versionNegotiated {
			return nil
		}
		// RFC 9368 §8: old v1-only servers have implicit v1 availability.
		if c.version != protocol.Version1 {
			return fail("version negotiation without server version_information")
		}
		info = &wire.VersionInformation{Chosen: protocol.Version1, Available: []protocol.Version{protocol.Version1}}
	}
	if info.Chosen != c.version || !slices.Contains(state.offered.Available, info.Chosen) {
		return fail("server chosen version does not match the negotiated version")
	}
	if c.versionNegotiated {
		if len(info.Available) == 0 {
			return fail("empty server availability after version negotiation")
		}
		available := append(slices.Clone(info.Available), info.Chosen)
		selected, ok := protocol.ChooseSupportedVersion(c.config.Versions, available)
		if !ok || selected != c.version {
			return fail("version negotiation downgrade detected")
		}
	}
	return nil
}

func versionedTokenKey(key string, version protocol.Version) string {
	return fmt.Sprintf("quic:%08x:%s", uint32(version), key)
}
