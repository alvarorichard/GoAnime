package quictls

import (
	"bytes"
	"crypto/tls"
	"encoding/binary"
	"errors"
	"slices"

	"github.com/enetx/quic-go/tlsquic"
	utls "github.com/enetx/utls"
)

const applicationSettingsPrefix = "surf quic alps\x00"

type savedApplicationSettings struct {
	alpn string
	peer []byte
}

type applicationSettingsState struct {
	observers       []func(string, []byte) error
	restored        *savedApplicationSettings
	current         *savedApplicationSettings
	early, rejected bool
	err             error
}

// ObserveApplicationSettings decorates Surf's TLS provider without changing
// custom providers. The observer runs on authenticated HandshakeDone, before
// that event is forwarded. It receives final ALPN and private peer ALPS bytes;
// nil means absent, while nonnil empty preserves negotiated empty settings.
// Accepted early data may inherit settings from its exact ticket at completion;
// ticket selection and early-key events never invoke the observer.
// An observer error becomes a terminal QUICErrorEvent and prevents ticket stores.
// Observers compose in factory order and must return promptly without reentry.
func ObserveApplicationSettings(
	factory tlsquic.NewClient,
	observe func(alpn string, raw []byte) error,
) tlsquic.NewClient {
	if factory == nil || observe == nil {
		return factory
	}

	return func(config *tls.Config) (tlsquic.Conn, error) {
		provider, err := factory(config)
		if err != nil {
			return provider, err
		}

		if own, ok := provider.(*client); ok {
			if own.alps == nil {
				own.alps = &applicationSettingsState{}
			}

			own.alps.observers = append(own.alps.observers, observe)
		}

		return provider, nil
	}
}

func applicationSettingsProtocols(extensions []utls.TLSExtension) []string {
	var protocols []string
	for _, extension := range extensions {
		switch settings := extension.(type) {
		case *utls.ApplicationSettingsExtension:
			protocols = append(protocols, settings.SupportedProtocols...)
		case *utls.ApplicationSettingsExtensionNew:
			protocols = append(protocols, settings.SupportedProtocols...)
		}
	}

	return protocols
}

func (c *client) observeApplicationEvent(event tls.QUICEvent) error {
	state := c.alps
	switch event.Kind {
	case tls.QUICResumeSession:
		state.restored = applicationSettingsFromExtra(event.SessionState.Extra)
		if state.restored != nil && !slices.Contains(c.alpsProtocols, state.restored.alpn) {
			// This factory's local ALPS payload is empty. A ticket which
			// negotiated ALPS cannot use early data if support was removed.
			event.SessionState.EarlyData = false
		}
	case tls.QUICSetWriteSecret:
		if event.Level == tls.QUICEncryptionLevelEarly {
			state.early = true
		}
	case tls.QUICRejectedEarlyData:
		state.rejected = true
	case tls.QUICHandshakeDone:
		connection := c.conn.ConnectionState()
		peer := bytes.Clone(connection.PeerApplicationSettings)
		if state.early && !state.rejected && connection.DidResume {
			// BoringSSL implicitly retains ticket ALPS only after accepted
			// early data; separately negotiating ALPS in that EE is forbidden.
			if peer != nil {
				return errors.New("QUIC TLS: ALPS extension after accepted early data")
			}

			if state.restored != nil && state.restored.alpn == connection.NegotiatedProtocol {
				peer = bytes.Clone(state.restored.peer)
			}
		}

		for _, observe := range state.observers {
			if err := observe(connection.NegotiatedProtocol, bytes.Clone(peer)); err != nil {
				return err
			}
		}

		if peer != nil {
			state.current = &savedApplicationSettings{alpn: connection.NegotiatedProtocol, peer: peer}
		}
	}

	return nil
}

func applicationSettingsFromExtra(extras [][]byte) *savedApplicationSettings {
	var saved *savedApplicationSettings
	for _, extra := range extras {
		data, own := bytes.CutPrefix(extra, []byte(applicationSettingsPrefix))
		if !own {
			continue
		}

		// Version1 records the factory's known empty local ALPS payload.
		if saved != nil || len(data) < 4 || data[0] != 1 {
			return nil
		}

		n := int(data[1])
		if n == 0 || n > len(data)-4 {
			return nil
		}

		size := int(binary.BigEndian.Uint16(data[2+n:]))
		if size != len(data)-4-n {
			return nil
		}

		saved = &savedApplicationSettings{alpn: string(data[2 : 2+n]), peer: bytes.Clone(data[4+n:])}
	}

	return saved
}

func (state *applicationSettingsState) ticketExtra(extras [][]byte) [][]byte {
	copy := slices.DeleteFunc(slices.Clone(extras), func(extra []byte) bool {
		return bytes.HasPrefix(extra, []byte(applicationSettingsPrefix))
	})
	if state.current == nil {
		return copy
	}

	// TLS bounds ALPN to 255 bytes and extension data to 65535 bytes. These are final
	// authenticated fields, never caller-sized or callback-mutated input.
	saved := state.current
	record := append([]byte(applicationSettingsPrefix), 1, byte(len(saved.alpn)))
	record = append(record, saved.alpn...)
	record = binary.BigEndian.AppendUint16(record, uint16(len(saved.peer)))
	record = append(record, saved.peer...)

	return append(copy, record)
}
