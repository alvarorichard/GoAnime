package quic

import (
	"slices"

	"github.com/enetx/quic-go/internal/qerr"
	"github.com/enetx/quic-go/internal/wire"
)

func (c *Conn) validatePeerTransportParameters(params *wire.TransportParameters) error {
	if c.config.ValidatePeerTransportParameters == nil {
		return nil
	}
	private := slices.Clone(params.UnknownParameters)
	for i := range private {
		private[i].Value = slices.Clone(private[i].Value)
	}
	if err := c.config.ValidatePeerTransportParameters(private); err != nil {
		return &qerr.TransportError{ErrorCode: qerr.TransportParameterError, ErrorMessage: err.Error()}
	}
	return nil
}

// PeerTransportParameter returns a private copy of an unrecognized extension
// selected by Config.PeerTransportParameterIDs. It reports only the current
// verified handshake, never cached 0-RTT state. Known parameters are interpreted
// by the engine and are not returned here. Values are otherwise uninterpreted.
func (c *Conn) PeerTransportParameter(id uint64) ([]byte, bool) {
	select {
	case <-c.HandshakeComplete():
	default:
		return nil, false
	}
	c.connStateMutex.Lock()
	defer c.connStateMutex.Unlock()
	if c.peerParams != nil {
		for _, parameter := range c.peerParams.UnknownParameters {
			if parameter.ID == id {
				return slices.Clone(parameter.Value), true
			}
		}
	}
	return nil, false
}
