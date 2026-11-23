// Package transportparams controls the encoding of client transport parameters.
// Parameter values supported by the transport remain owned by the QUIC engine.
package transportparams

import (
	"bytes"
	"fmt"
	"slices"

	"github.com/enetx/quic-go/quicvarint"
)

// Grease selects the engine-generated reserved parameter in Order. It is a
// selector, not a wire parameter ID. The ID and value remain freshly randomized.
const Grease = ^uint64(0)

// Parameter is an extension parameter whose value is already wire-encoded.
// Additional parameters advertise extensions; they do not implement them.
type Parameter struct {
	ID    uint64
	Value []byte
}

// Layout specifies preferred parameter order and additional extension parameters.
// Unlisted parameters follow in their original order. An empty layout preserves
// upstream encoding. Known engine parameters cannot be overridden here.
type Layout struct {
	Order         []uint64
	Additional    []Parameter
	DisableGrease bool
}

// Clone returns an independent snapshot of all mutable slices.
func (l *Layout) Clone() *Layout {
	if l == nil {
		return nil
	}
	c := *l
	c.Order = slices.Clone(l.Order)
	c.Additional = slices.Clone(l.Additional)
	for i := range c.Additional {
		c.Additional[i].Value = bytes.Clone(c.Additional[i].Value)
	}
	return &c
}

// Validate rejects duplicates, reserved selectors and overrides of engine fields.
func (l *Layout) Validate() error {
	if l == nil {
		return nil
	}
	seen := make(map[uint64]bool, len(l.Order))
	for _, id := range l.Order {
		if (id > quicvarint.Max && id != Grease) || seen[id] {
			return fmt.Errorf("invalid or duplicate transport parameter order ID: %d", id)
		}
		seen[id] = true
	}
	clear(seen)
	for _, p := range l.Additional {
		if p.ID > quicvarint.Max || seen[p.ID] || engineParameter(p.ID) || (isGrease(p.ID) && !l.DisableGrease) {
			return fmt.Errorf("invalid, duplicate or engine-owned transport parameter: %d", p.ID)
		}
		seen[p.ID] = true
	}
	return nil
}

func engineParameter(id uint64) bool {
	return id <= 0x10 || id == 0x20 || id == 0x1d || id == 0x17f7586d2cb571 || id == 0xff04de1b || id == 0xff02de1a
}
func isGrease(id uint64) bool { return id >= 27 && (id-27)%31 == 0 }

// Apply rewrites a valid upstream encoding, preserving bytes of existing values.
// It is only used for clients, after Validate succeeds. It never mutates input.
func (l *Layout) Apply(encoded []byte) []byte {
	if l == nil {
		return encoded
	}
	type entry struct {
		id      uint64
		encoded []byte
	}
	entries := make([]entry, 0, 20+len(l.Additional))
	for rest := encoded; len(rest) > 0; {
		start := rest
		id, n, err := quicvarint.Parse(rest)
		if err != nil {
			panic("transportparams: invalid engine parameter ID")
		}
		rest = rest[n:]
		length, n, err := quicvarint.Parse(rest)
		if err != nil || length > uint64(len(rest)-n) {
			panic("transportparams: invalid engine parameter length")
		}
		rest = rest[n+int(length):]
		// Legacy min_ack_delay is engine-owned, despite sharing a reserved residue.
		if isGrease(id) && id != 0xff02de1a {
			if l.DisableGrease {
				continue
			}
			id = Grease
		}
		entries = append(entries, entry{id, start[:len(start)-len(rest)]})
	}
	for _, p := range l.Additional {
		b := quicvarint.Append(nil, p.ID)
		b = quicvarint.Append(b, uint64(len(p.Value)))
		b = append(b, p.Value...)
		entries = append(entries, entry{p.ID, b})
	}
	out := make([]byte, 0, len(encoded)+32)
	for _, id := range l.Order {
		for i := range entries {
			if entries[i].id == id && entries[i].encoded != nil {
				out = append(out, entries[i].encoded...)
				entries[i].encoded = nil
				break
			}
		}
	}
	for _, e := range entries {
		out = append(out, e.encoded...)
	}
	return out
}
