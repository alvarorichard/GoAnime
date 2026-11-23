package http3

import (
	"bytes"
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"slices"
	"sync"

	"github.com/enetx/quic-go"
	"github.com/enetx/quic-go/quicvarint"
	"github.com/enetx/quic-go/tlsquic"
)

const clientSettingsPrefix = "quic-go h3 peer settings v1\x00"
const maxPendingSettingsTickets = 64

// ClientSessionSettings returns a private snapshot of the server SETTINGS saved
// with this exact TLS ticket. It does not validate TLS lifetime or authenticity.
// Missing, malformed and duplicate metadata returns false; TLS owns validation.
func ClientSessionSettings(session *tls.ClientSessionState) (*Settings, bool) {
	if session == nil {
		return nil, false
	}
	_, state, err := session.ResumptionState()
	if err != nil || state == nil {
		return nil, false
	}
	saved, ok := clientSettingsFromExtra(state.Extra)
	if !ok {
		return nil, false
	}
	return &Settings{MaxFieldSectionSize: saved.MaxFieldSectionSize, EnableDatagrams: saved.Datagram,
		EnableExtendedConnect: saved.ExtendedConnect, Other: saved.Other.Clone()}, true
}

func clientSettingsFromExtra(extras [][]byte) (*settingsFrame, bool) {
	var data []byte
	for _, extra := range extras {
		if value, ok := bytes.CutPrefix(extra, []byte(clientSettingsPrefix)); ok {
			if data != nil || len(value) > 64<<10 {
				return nil, false
			}
			data = value
		}
	}
	if len(data) == 0 {
		return nil, false
	}
	r := bytes.NewReader(data)
	typ, err := quicvarint.Read(r)
	if err != nil || typ != 4 {
		return nil, false
	}
	n, err := quicvarint.Read(r)
	if err != nil || n != uint64(r.Len()) {
		return nil, false
	}
	settings, err := parseSettingsFrame(&countingByteReader{Reader: r}, n, 0, nil)
	return settings, err == nil
}

func encodeClientSettings(sf *settingsFrame) []byte {
	var payload []byte
	add := func(id, value uint64) { payload = quicvarint.Append(quicvarint.Append(payload, id), value) }
	if sf.MaxFieldSectionSize >= 0 {
		add(6, uint64(sf.MaxFieldSectionSize))
	}
	if sf.Datagram {
		add(settingDatagram, 1)
	}
	if sf.ExtendedConnect {
		add(settingExtendedConnect, 1)
	}
	sf.Other.Iter().ForEach(func(id, value uint64) { add(id, value) })
	data := quicvarint.Append([]byte(clientSettingsPrefix), 4)
	return append(quicvarint.Append(data, uint64(len(payload))), payload...)
}

func cloneClientTicket(session *tls.ClientSessionState, settings *settingsFrame, earlyData bool) (*tls.ClientSessionState, error) {
	ticket, original, err := session.ResumptionState()
	if err != nil || original == nil {
		return nil, errors.Join(errors.New("http3: missing TLS session state"), err)
	}
	data, err := original.Bytes()
	if err != nil {
		return nil, err
	}
	state, err := tls.ParseSessionState(data)
	if err != nil {
		return nil, err
	}
	state.EarlyData = original.EarlyData && earlyData
	state.Extra = slices.DeleteFunc(state.Extra, func(value []byte) bool { return bytes.HasPrefix(value, []byte(clientSettingsPrefix)) })
	if settings != nil {
		state.Extra = append(state.Extra, encodeClientSettings(settings))
	}
	return tls.NewResumptionState(bytes.Clone(ticket), state)
}

type pendingSettingsTicket struct {
	key    string
	ticket *tls.ClientSessionState
}

// Each TLS connection owns its selected ticket and pending cache insertions.
// There is no address-keyed SETTINGS history shared by racing connections.
type clientSettingsSession struct {
	mutex                   sync.Mutex
	storeMutex              sync.Mutex
	conn                    *quic.Conn
	cache                   tls.ClientSessionCache
	current, restored       *settingsFrame
	resume                  *tls.SessionState // borrowed only until the TLS provider's next event
	early, rejected, closed bool
	pending                 []pendingSettingsTicket
}

func (s *clientSettingsSession) Get(key string) (*tls.ClientSessionState, bool) {
	ticket, ok := s.cache.Get(key)
	if !ok || ticket == nil {
		return ticket, ok
	}
	_, state, err := ticket.ResumptionState()
	if err != nil || state == nil {
		return nil, false
	}
	settings, known := clientSettingsFromExtra(state.Extra)
	// Preserve ordinary TLS resumption when no usable application state exists.
	// Cloning leaves an application-owned cache and its ticket bytes untouched.
	copy, err := cloneClientTicket(ticket, settings, known)
	return copy, err == nil
}

func (s *clientSettingsSession) Put(key string, ticket *tls.ClientSessionState) {
	s.mutex.Lock()
	if ticket == nil {
		s.pending = slices.DeleteFunc(s.pending, func(t pendingSettingsTicket) bool { return t.key == key })
	}
	if s.closed {
		s.mutex.Unlock()
		return
	}
	if len(s.pending) == maxPendingSettingsTickets {
		copy(s.pending, s.pending[1:])
		s.pending = s.pending[:len(s.pending)-1]
	}
	s.pending = append(s.pending, pendingSettingsTicket{key, ticket})
	s.mutex.Unlock()
	s.flushTickets()
}

func (s *clientSettingsSession) flushTickets() {
	s.storeMutex.Lock()
	defer s.storeMutex.Unlock()
	for {
		s.mutex.Lock()
		if len(s.pending) == 0 || s.current == nil && s.pending[0].ticket != nil {
			s.mutex.Unlock()
			return
		}
		item, settings := s.pending[0], s.current
		s.pending[0] = pendingSettingsTicket{}
		s.pending = s.pending[1:]
		s.mutex.Unlock()
		if item.ticket == nil {
			s.cache.Put(item.key, nil)
		} else {
			s.store(item.key, item.ticket, settings)
		}
	}
}

func (s *clientSettingsSession) store(key string, ticket *tls.ClientSessionState, settings *settingsFrame) {
	copy, err := cloneClientTicket(ticket, settings, settings != nil)
	if err == nil {
		s.cache.Put(key, copy)
	}
}

func (s *clientSettingsSession) close() {
	s.mutex.Lock()
	s.closed = true
	// Without server SETTINGS there is no complete HTTP/3 resumption state.
	// Do not overwrite usable tickets from another connection with this attempt.
	s.pending = nil
	s.mutex.Unlock()
}

// https://www.rfc-editor.org/rfc/rfc9114#section-7.2.4.2
// https://www.rfc-editor.org/rfc/rfc9204#section-3.2.3
// https://www.rfc-editor.org/rfc/rfc9297#section-2.1.1
// Unknown settings are retained with the ticket but do not become capabilities.
func checkResumedSettings(old, current *settingsFrame) (ErrCode, error) {
	if old.MaxFieldSectionSize >= 0 && current.MaxFieldSectionSize < 0 ||
		old.MaxFieldSectionSize >= 0 && current.MaxFieldSectionSize >= 0 && current.MaxFieldSectionSize < old.MaxFieldSectionSize ||
		old.Datagram && !current.Datagram || old.ExtendedConnect && !current.ExtendedConnect ||
		current.Other.Get(7).UnwrapOrDefault() < old.Other.Get(7).UnwrapOrDefault() {
		return ErrCodeSettingsError, errors.New("server reduced remembered HTTP/3 SETTINGS after accepting early data")
	}
	if previous := old.Other.Get(1).UnwrapOrDefault(); previous > 0 && current.Other.Get(1).UnwrapOrDefault() != previous {
		return ErrCodeQPACKDecoderStreamError, errors.New("server changed remembered nonzero QPACK maximum capacity")
	}
	return 0, nil
}

func (s *clientSettingsSession) validateSettings(sf *settingsFrame, qpack *clientQPACK) (ErrCode, error) {
	s.mutex.Lock()
	restored, early, rejected := s.restored, s.early, s.rejected
	s.mutex.Unlock()
	if restored != nil && early && !rejected && s.conn.ConnectionState().Used0RTT {
		if code, err := checkResumedSettings(restored, sf); err != nil {
			if code == ErrCodeQPACKDecoderStreamError && qpack != nil {
				code = resumedCapacityErrorCode(qpack.config, restored.Other.Get(1).UnwrapOrDefault(), sf.Other.Get(1).UnwrapOrDefault())
			}
			return code, err
		}
	}
	return 0, nil
}

func resumedCapacityErrorCode(config *QPACKConfig, previous, current uint64) ErrCode {
	code := ErrCodeQPACKDecoderStreamError
	if config.ResumptionErrorCode != nil {
		if selected := config.ResumptionErrorCode(previous, current); selected != 0 {
			code = selected
		}
	}
	if uint64(code) > quicvarint.Max {
		return ErrCodeInternalError
	}
	return code
}

func (s *clientSettingsSession) commitSettings(sf *settingsFrame) {
	snapshot := *sf
	snapshot.Other = sf.Other.Clone()
	s.mutex.Lock()
	if s.closed {
		s.mutex.Unlock()
		return
	}
	s.current = &snapshot
	s.mutex.Unlock()
	s.flushTickets()
}

type clientSessionTLS struct {
	tlsquic.Conn
	session *clientSettingsSession
}

func (c *clientSessionTLS) syncResume() {
	c.session.mutex.Lock()
	if c.session.resume != nil {
		c.session.early = c.session.resume.EarlyData
		c.session.resume = nil
	}
	c.session.mutex.Unlock()
}

func (c *clientSessionTLS) NextEvent() tls.QUICEvent {
	c.syncResume()
	event := c.Conn.NextEvent()
	if event.Kind == tls.QUICResumeSession && event.SessionState != nil {
		saved, valid := clientSettingsFromExtra(event.SessionState.Extra)
		if !valid {
			event.SessionState.EarlyData = false
		}
		c.session.mutex.Lock()
		c.session.restored, c.session.resume = saved, event.SessionState
		c.session.mutex.Unlock()
	}
	if event.Kind == tls.QUICRejectedEarlyData {
		c.session.mutex.Lock()
		c.session.rejected = true
		c.session.mutex.Unlock()
	}
	return event
}

func (c *clientSessionTLS) StoreSession(state *tls.SessionState) error {
	c.syncResume()
	return c.Conn.StoreSession(state)
}

func (t *Transport) wrapSessionTLS(previous func(*quic.Conn, tlsquic.NewClient) tlsquic.NewClient) func(*quic.Conn, tlsquic.NewClient) tlsquic.NewClient {
	return func(conn *quic.Conn, factory tlsquic.NewClient) tlsquic.NewClient {
		if previous != nil {
			factory = previous(conn, factory)
		}
		return func(config *tls.Config) (tlsquic.Conn, error) {
			if factory == nil {
				return nil, fmt.Errorf("http3: TLS factory decorator returned nil")
			}
			if config.ClientSessionCache == nil {
				return factory(config)
			}
			session := &clientSettingsSession{conn: conn, cache: config.ClientSessionCache}
			config = config.Clone()
			config.ClientSessionCache = session
			provider, err := factory(config)
			if err != nil || provider == nil {
				return provider, err
			}
			t.clientSessions.Store(conn, session)
			context.AfterFunc(conn.Context(), func() {
				session.close()
				t.clientSessions.Delete(conn)
			})
			return &clientSessionTLS{Conn: provider, session: session}, nil
		}
	}
}

func (c *ClientConn) bindSession(session *clientSettingsSession) {
	c.rawConn.clientSession = session
	session.mutex.Lock()
	old, early, rejected := session.restored, session.early, session.rejected
	session.mutex.Unlock()
	if old != nil && early && !rejected && c.rawConn.clientQPACK != nil {
		if err := c.rawConn.clientQPACK.onSettings(old.Other.Get(1).UnwrapOrDefault(), old.Other.Get(7).UnwrapOrDefault(), old.Other.Get(1).IsSome()); err != nil {
			c.CloseWithError(quic.ApplicationErrorCode(ErrCodeExcessiveLoad), err.Error())
		}
	}
}

func (c *rawConn) peerFieldSectionLimit() int64 {
	if p := c.peerSettings; p != nil {
		p.mutex.Lock()
		defer p.mutex.Unlock()
		if p.current != nil {
			return p.current.MaxFieldSectionSize
		}
	}

	if s := c.clientSession; s != nil {
		s.mutex.Lock()
		defer s.mutex.Unlock()
		if s.current != nil {
			return s.current.MaxFieldSectionSize
		}
		if s.early && !s.rejected && s.restored != nil {
			return s.restored.MaxFieldSectionSize
		}
	}
	select {
	case <-c.receivedSettings:
		return c.settings.MaxFieldSectionSize
	default:
		return -1
	}
}

var _ tls.ClientSessionCache = (*clientSettingsSession)(nil)
var _ tlsquic.Conn = (*clientSessionTLS)(nil)
