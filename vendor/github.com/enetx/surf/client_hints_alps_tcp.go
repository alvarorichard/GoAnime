package surf

import (
	"context"
	"net"
	"sync"

	"github.com/enetx/http"
	"github.com/enetx/http/http2"
	"github.com/enetx/surf/internal/alps"
)

type selectedTCPSettingsKey struct{}

// A selection belongs to one TCP call, below the TCP/QUIC race. GotConn can
// replace it when the HTTP engine retries a closed pooled connection.
type selectedTCPSettings struct {
	mu       sync.Mutex
	snapshot *alps.Snapshot
}

func (s *selectedTCPSettings) selectConnection(conn net.Conn) {
	var snapshot *alps.Snapshot
	if secure, ok := conn.(*profileTLSConn); ok {
		snapshot = secure.applicationSettings
	}

	s.mu.Lock()
	s.snapshot = snapshot
	s.mu.Unlock()
}

func (s *selectedTCPSettings) load() *alps.Snapshot {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.snapshot
}

// Initial settings apply before the first stream and before ordinary SETTINGS.
// The engine owns setting validation and all resulting HPACK/flow state.
func (rt *roundtripper) initialTCPSettings(conn net.Conn) ([]http2.Setting, error) {
	var settings []http2.Setting
	if secure, ok := conn.(*profileTLSConn); ok {
		settings = secure.applicationSettings.HTTP2Settings()
		if rt.http2settings != nil && rt.http2settings.initialPeerSettings != nil {
			var err error
			settings, err = rt.http2settings.initialPeerSettings(settings)
			if err != nil {
				return nil, &alps.RequestPreparationError{Err: err}
			}
		}
	}

	return settings, nil
}

func (rt *roundtripper) prepareTCPHeaders(req *http.Request, major int) {
	if major == 2 {
		if selected, ok := req.Context().Value(selectedTCPSettingsKey{}).(*selectedTCPSettings); ok {
			if err := alps.Prepare(req, selected.load()); err != nil {
				// The HTTP hook owns this unsent clone. Cancel only its derived
				// context so the engine aborts before opening a stream or body.
				ctx, cancel := context.WithCancelCause(req.Context())
				cancel(&alps.RequestPreparationError{Err: err})
				*req = *req.WithContext(ctx)

				return
			}
		}
	}

	if rt.protocolHeaders != nil {
		rt.protocolHeaders(req, major)
	}
}
