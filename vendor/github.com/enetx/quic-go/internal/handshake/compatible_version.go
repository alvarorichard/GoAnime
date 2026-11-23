package handshake

import (
	"crypto/tls"
	"fmt"

	"github.com/enetx/quic-go/internal/protocol"
)

// InitialOpenerForVersion prepares independent candidate keys without changing
// the active connection. Authentication must succeed before SwitchVersion.
func (h *cryptoSetup) InitialOpenerForVersion(version protocol.Version) (LongHeaderOpener, error) {
	if h.perspective != protocol.PerspectiveClient || !h.ourParams.CompatibleVersionNegotiation ||
		!protocol.CanSwitchVersion(h.version, version) || h.handshakeOpener != nil || h.handshakeSealer != nil {
		return nil, fmt.Errorf("compatible version switch is unavailable")
	}
	if h.initialOpener == nil {
		return nil, ErrKeysDropped
	}
	_, opener := NewInitialAEAD(h.initialConnID, h.perspective, version)
	// Initial PN space survives the switch, including truncated PN decoding
	// after an ACK-only response using the original version.
	opener.(*longHeaderOpener).highestRcvdPN = h.initialOpener.(*longHeaderOpener).highestRcvdPN
	return opener, nil
}

func (h *cryptoSetup) SwitchVersion(version protocol.Version) error {
	opener, err := h.InitialOpenerForVersion(version)
	if err != nil {
		return err
	}
	sealer, _ := NewInitialAEAD(h.initialConnID, h.perspective, version)
	h.initialOpener, h.initialSealer = opener, sealer
	h.version = version
	h.aead.version = version
	// The original-version 0-RTT sealer is deliberately retained (RFC 9369 §4.1).
	return nil
}

// Session tickets belong to the negotiated version. Existing native cache keys
// remain untouched unless the client explicitly enables compatible negotiation.
type versionedSessionCache struct {
	cache   tls.ClientSessionCache
	version func() protocol.Version
}

func (c *versionedSessionCache) key(key string) string {
	return fmt.Sprintf("quic:%08x:%s", uint32(c.version()), key)
}
func (c *versionedSessionCache) Get(key string) (*tls.ClientSessionState, bool) {
	return c.cache.Get(c.key(key))
}
func (c *versionedSessionCache) Put(key string, state *tls.ClientSessionState) {
	c.cache.Put(c.key(key), state)
}
