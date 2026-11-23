// Package alps decodes authenticated application-layer protocol settings.
// It owns framing and origin identity; browser values and request policy stay
// in Surf. Snapshots are immutable and never become response opt-in caches.
package alps

import (
	"net/url"
	"slices"

	"github.com/enetx/g"
	"github.com/enetx/http/http2"
	"github.com/enetx/surf/internal/origin"
)

// Snapshot retains one authenticated TLS connection's application settings.
// Getter results cannot mutate another request's connection state.
type Snapshot struct {
	http2    []http2.Setting
	http3    g.MapOrd[uint64, uint64]
	hints    map[string]string
	bypassed bool
}

// HTTP2Settings returns a private copy, with duplicate IDs resolved last-wins
// and IDs sorted as in Chromium's ALPS settings map.
func (s *Snapshot) HTTP2Settings() []http2.Setting {
	if s == nil {
		return nil
	}

	return slices.Clone(s.http2)
}

// HTTP3Settings returns private ordered raw settings, retaining explicit zero.
func (s *Snapshot) HTTP3Settings() g.MapOrd[uint64, uint64] {
	if s == nil {
		return g.NewMapOrd[uint64, uint64]()
	}

	return s.http3.Clone()
}

// AcceptCH looks up the actual URL's canonical origin, independently of SNI,
// Request.Host, physical peer address or proxy route. Empty values stay present.
func (s *Snapshot) AcceptCH(u *url.URL) (string, bool) {
	if s == nil {
		return "", false
	}

	key, valid := origin.Key(u)
	if !valid {
		return "", false
	}

	value, found := s.hints[key]
	return value, found
}

// AcceptCHBypassed reports Chromium's retained H2 malformed-entry-tail bypass.
// Complete preceding entries remain available; fatal frame errors still fail.
func (s *Snapshot) AcceptCHBypassed() bool { return s != nil && s.bypassed }

func (s *Snapshot) add(originText, value string) {
	u, err := url.Parse(originText)
	if err != nil {
		return
	}

	key, valid := origin.Key(u)
	if !valid || key != originText {
		return
	}

	if _, present := s.hints[key]; !present {
		s.hints[key] = value
	}
}
