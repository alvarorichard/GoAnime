package h3transport

import (
	"context"
	"sync"
	"sync/atomic"

	"github.com/enetx/g"
	"github.com/enetx/http"
	"github.com/enetx/quic-go"
	"github.com/enetx/quic-go/http3"
	"github.com/enetx/quic-go/tlsquic"
	"github.com/enetx/surf/internal/alps"
	"github.com/enetx/surf/internal/quictls"
)

// Each raw QUIC generation owns its authenticated ALPS. Coalesced requests
// consult that connection's canonical-origin table, never an authority cache.
type applicationSettings struct {
	mu          sync.RWMutex
	connections map[*quic.Conn]*atomic.Pointer[alps.Snapshot]
}

func (s *applicationSettings) bind(conn *quic.Conn) *atomic.Pointer[alps.Snapshot] {
	state := new(atomic.Pointer[alps.Snapshot])
	s.mu.Lock()
	defer s.mu.Unlock()

	if conn.Context().Err() != nil {
		return state
	}

	if s.connections == nil {
		s.connections = make(map[*quic.Conn]*atomic.Pointer[alps.Snapshot])
	}

	s.connections[conn] = state
	context.AfterFunc(conn.Context(), func() {
		s.mu.Lock()
		delete(s.connections, conn)
		s.mu.Unlock()
	})

	// The observer retains this object, not a map insertion callback. A late
	// handshake event therefore cannot resurrect a retired connection entry.
	return state
}

func (s *applicationSettings) snapshot(conn *quic.Conn) *alps.Snapshot {
	s.mu.RLock()
	state := s.connections[conn]
	s.mu.RUnlock()

	if state == nil {
		return nil
	}

	return state.Load()
}

func (s *applicationSettings) configure(config *quic.Config) {
	if config.ClientHandshake == nil {
		config.ClientHandshake = new(quic.ClientHandshake)
	}

	previous := config.ClientHandshake.WrapTLSClient
	config.ClientHandshake.WrapTLSClient = func(conn *quic.Conn, factory tlsquic.NewClient) tlsquic.NewClient {
		state := s.bind(conn)
		factory = quictls.ObserveApplicationSettings(factory, func(protocol string, raw []byte) error {
			return readApplicationSettings(state, protocol, raw)
		})

		if previous != nil {
			factory = previous(conn, factory)
		}

		return factory
	}
}

// Framing validation and publication are one step. Its caller supplies only
// authenticated completion events; a malformed snapshot never replaces state.
func readApplicationSettings(state *atomic.Pointer[alps.Snapshot], protocol string, raw []byte) error {
	if protocol != "h3" {
		return nil
	}

	snapshot, err := alps.DecodeHTTP3(raw)
	if err != nil {
		return err
	}

	state.Store(snapshot)
	return nil
}

func (ut *Transport) initialPeerSettings(conn *quic.Conn) (g.MapOrd[uint64, uint64], error) {
	return ut.applicationSettings.snapshot(conn).HTTP3Settings(), nil
}

func (ut *Transport) prepareHeaders(req *http.Request, conn *quic.Conn) error {
	// This decision is frozen by the engine before TLS can finish concurrently.
	// QUICHE populates ACCEPT_CH only in FinishHandshake; an early request must
	// not acquire a grant from a later fresh or inherited ALPS snapshot.
	// https://github.com/google/quiche/blob/132bfd90ea370faeacee616c70d10210bb12ba66/quiche/quic/core/tls_client_handshaker.cc#L648
	if !http3.IsEarlyRequest(req) {
		if err := alps.Prepare(req, ut.applicationSettings.snapshot(conn)); err != nil {
			return err
		}
	}

	if ut.protocolHeaders != nil {
		ut.protocolHeaders(req, 3)
	}

	return nil
}
