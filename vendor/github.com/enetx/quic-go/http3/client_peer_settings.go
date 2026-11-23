package http3

import (
	"context"
	"errors"
	"fmt"
	"sync"

	"github.com/enetx/g"
	"github.com/enetx/quic-go"
	"github.com/enetx/quic-go/quicvarint"
)

var errReservedPeerSetting = errors.New("reserved HTTP/2 setting")

// A protocol generation owns its authenticated prelude independently of the
// definitive control-stream SETTINGS and the session ticket's exact wire record.
type clientPeerSettings struct {
	initial func(*quic.Conn) (g.MapOrd[uint64, uint64], error)
	once    sync.Once
	ready   chan struct{}
	err     error
	mutex   sync.Mutex
	current *settingsFrame
}

func validatePeerSetting(id, value uint64) error {
	if id > quicvarint.Max || value > quicvarint.Max {
		return fmt.Errorf("HTTP/3 setting exceeds QUIC varint range: %d=%d", id, value)
	}

	switch id {
	case 2, 3, 4, 5:
		return fmt.Errorf("%w: %d", errReservedPeerSetting, id)
	case settingExtendedConnect:
		if value > 1 {
			return fmt.Errorf("invalid value for SETTINGS_ENABLE_CONNECT_PROTOCOL: %d", value)
		}
	case settingDatagram:
		if value > 1 {
			return fmt.Errorf("invalid value for SETTINGS_H3_DATAGRAM: %d", value)
		}
	}

	return nil
}

func peerSettingsFrame(values g.MapOrd[uint64, uint64]) (*settingsFrame, error) {
	sf := &settingsFrame{MaxFieldSectionSize: -1}
	seen := make(map[uint64]bool, len(values))
	for id, value := range values.Iter() {
		if seen[id] {
			return nil, fmt.Errorf("duplicate setting: %d", id)
		}

		if err := validatePeerSetting(id, value); err != nil {
			return nil, err
		}

		seen[id] = true
		switch id {
		case settingMaxFieldSectionSize:
			sf.MaxFieldSectionSize = int64(value)
		case settingDatagram:
			sf.Datagram, sf.datagramSet = value == 1, true
		case settingExtendedConnect:
			sf.ExtendedConnect, sf.extendedConnectSet = value == 1, true
		default:
			sf.Other.Insert(id, value)
		}
	}

	return sf, nil
}

func overlayPeerSettings(base, update *settingsFrame) *settingsFrame {
	if base == nil {
		base = &settingsFrame{MaxFieldSectionSize: -1}
	}

	merged := *base
	merged.Other = base.Other.Clone()
	if update.MaxFieldSectionSize >= 0 {
		merged.MaxFieldSectionSize = update.MaxFieldSectionSize
	}

	if update.Datagram || update.datagramSet {
		merged.Datagram, merged.datagramSet = update.Datagram, true
	}

	if update.ExtendedConnect || update.extendedConnectSet {
		merged.ExtendedConnect, merged.extendedConnectSet = update.ExtendedConnect, true
	}

	for id, value := range update.Other.Iter() {
		merged.Other.Insert(id, value)
	}

	return &merged
}

func (c *rawConn) initializePeerSettings(ctx context.Context) error {
	p := c.peerSettings
	if p == nil {
		return nil
	}

	if ctx.Err() != nil {
		return context.Cause(ctx)
	}

	if c.context().Err() != nil {
		return c.peerSettingsFailure(context.Cause(c.context()))
	}

	select {
	case <-ctx.Done():
		return context.Cause(ctx)
	case <-c.context().Done():
		return c.peerSettingsFailure(context.Cause(c.context()))
	case <-c.conn.HandshakeComplete():
	}

	owner := false
	p.once.Do(func() { owner = true })
	if !owner {
		select {
		case <-ctx.Done():
			return context.Cause(ctx)
		case <-c.context().Done():
			return c.peerSettingsFailure(context.Cause(c.context()))
		case <-p.ready:
			return c.peerSettingsFailure(context.Cause(c.context()))
		}
	}

	defer close(p.ready)
	func() {
		values, err := p.initial(c.conn)
		// Do not hold the generation lock across an application callback: early
		// rejection must retire the owner before QUIC replaces its stream maps.
		c.generationMx.RLock()
		defer c.generationMx.RUnlock()
		if c.context().Err() != nil {
			p.setError(context.Cause(c.context()))
			return
		}

		code := ErrCodeSettingsError
		if err == nil && len(values) > 0 {
			var sf *settingsFrame
			sf, err = peerSettingsFrame(values)
			if err == nil {
				code, err = c.applyClientPeerSettings(sf, true)
			}
		}

		if err != nil {
			p.setError(&RequestPreparationError{err})
			c.CloseWithError(quic.ApplicationErrorCode(code), err.Error())
		}
	}()

	return c.peerSettingsFailure(context.Cause(c.context()))
}

func (p *clientPeerSettings) setError(err error) {
	p.mutex.Lock()
	p.err = err
	p.mutex.Unlock()
}

func (c *rawConn) peerSettingsFailure(fallback error) error {
	if p := c.peerSettings; p != nil {
		p.mutex.Lock()
		defer p.mutex.Unlock()
		if p.err != nil {
			return p.err
		}
	}

	return fallback
}

// applyClientPeerSettings has the same validation and QPACK owner for initial
// and wire inputs. Only explicit prelude values overlay restored limits; later
// wire updates preserve prelude values whose identifiers they omit.
func (c *rawConn) applyClientPeerSettings(sf *settingsFrame, initial bool) (ErrCode, error) {
	effective := sf
	if p := c.peerSettings; p != nil {
		p.mutex.Lock()
		defer p.mutex.Unlock()
		base := p.current
		if initial && base == nil && c.clientSession != nil {
			s := c.clientSession
			s.mutex.Lock()
			if s.early && !s.rejected {
				base = s.restored
			}
			s.mutex.Unlock()
		}

		if initial || base != nil {
			effective = overlayPeerSettings(base, sf)
		}

		defer func() {
			// Publish only after every updater has accepted the same snapshot.
			if effective != nil {
				p.current = effective
			}
		}()
	}

	checked := sf
	if initial {
		checked = effective
	}

	if c.clientSession != nil {
		if code, err := c.clientSession.validateSettings(checked, c.clientQPACK); err != nil {
			effective = nil
			return code, err
		}
	}

	if effective.Datagram && c.enableDatagrams && !c.ConnectionState().SupportsDatagrams.Remote {
		effective = nil
		return ErrCodeSettingsError, fmt.Errorf("missing QUIC Datagram support")
	}

	if c.clientQPACK != nil {
		if err := c.clientQPACK.onSettings(effective.Other.Get(1).UnwrapOrDefault(), effective.Other.Get(7).UnwrapOrDefault(), effective.Other.Get(1).IsSome()); err != nil {
			effective = nil
			return ErrCodeExcessiveLoad, err
		}
	}

	return 0, nil
}

func (c *rawConn) applyWirePeerSettings(sf *settingsFrame) (ErrCode, error) {
	if err := c.initializePeerSettings(c.context()); err != nil {
		return ErrCodeSettingsError, err
	}

	if c.peerSettings != nil {
		c.generationMx.RLock()
		defer c.generationMx.RUnlock()
	}

	if c.context().Err() != nil {
		return ErrCodeInternalError, context.Cause(c.context())
	}

	if code, err := c.applyClientPeerSettings(sf, false); err != nil {
		return code, err
	}

	effective := sf
	if p := c.peerSettings; p != nil {
		p.mutex.Lock()
		effective = p.current
		p.mutex.Unlock()
	}

	c.settings = &Settings{
		MaxFieldSectionSize:   effective.MaxFieldSectionSize,
		EnableDatagrams:       effective.Datagram,
		EnableExtendedConnect: effective.ExtendedConnect,
		Other:                 effective.Other,
	}
	if c.clientSession != nil {
		c.clientSession.commitSettings(sf)
	}

	close(c.receivedSettings)
	return 0, nil
}
