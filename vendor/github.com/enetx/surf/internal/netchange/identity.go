package netchange

import "net/netip"

// identityState belongs to one serial subscription, including its first value.
// An unavailable read must not erase a known identity or produce a change.
type identityState struct {
	read        func([]netip.Addr) (string, error)
	value       string
	initialized bool
}

func (s *identityState) refresh(neighbors []netip.Addr) bool {
	if s.read == nil {
		return false
	}

	value, err := s.read(neighbors)
	if err != nil {
		return false
	}

	changed := s.initialized && s.value != value
	s.value, s.initialized = value, true
	return changed
}

func (s *identityState) observe(event Event) Event {
	if event.Address || event.Link || event.Route || event.IdentityRoute {
		event.Identity = s.refresh(nil) || event.Identity
	} else if len(event.Neighbors) != 0 {
		event.Identity = s.refresh(event.Neighbors) || event.Identity
	}

	return event
}
