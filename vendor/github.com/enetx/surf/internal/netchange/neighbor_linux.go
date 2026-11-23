package netchange

import (
	"encoding/binary"
	"net/netip"
	"syscall"

	"golang.org/x/sys/unix"
)

type linuxNeighborKey struct {
	index   uint32
	address netip.Addr
}

type linuxNeighbor struct {
	mac    [6]byte
	hasMAC bool
}

// The standard route-attribute parser must skip the same fixed header size.
var _ [unix.SizeofNdMsg]byte = [unix.SizeofRtMsg]byte{}

func (s *linuxState) identityLink(message syscall.NetlinkMessage, index, flags uint32) {
	if s.ethernet == nil {
		s.ethernet = make(map[uint32]bool)
	}

	if message.Header.Type == unix.RTM_DELLINK {
		delete(s.ethernet, index)
	} else {
		s.ethernet[index] = binary.NativeEndian.Uint16(message.Data[2:4]) == unix.ARPHRD_ETHER
	}

	if message.Header.Type == unix.RTM_DELLINK || s.links[index]&unix.IFF_UP != 0 && flags&unix.IFF_UP == 0 {
		for key := range s.neighbors {
			if key.index == index {
				delete(s.neighbors, key)
			}
		}

		for key, route := range s.gateways {
			if route.index == index {
				delete(s.gateways, key)
			}
		}
	}
}

// neighbor reports MAC acquisition only, rather than ARP state churn, known-MAC
// replacement or deletion. Identity readers decide which gateways matter.
func (s *linuxState) neighbor(message syscall.NetlinkMessage) (netip.Addr, bool) {
	if len(message.Data) < unix.SizeofNdMsg {
		return netip.Addr{}, false
	}

	if (len(message.Data)-unix.SizeofNdMsg)%unix.RTA_ALIGNTO != 0 {
		return netip.Addr{}, false
	}

	index := binary.NativeEndian.Uint32(message.Data[4:8])
	if !s.ethernet[index] {
		return netip.Addr{}, false
	}

	family := message.Data[0]
	if family != unix.AF_INET && family != unix.AF_INET6 {
		return netip.Addr{}, false
	}

	// NdMsg and RtMsg both have a twelve-byte fixed header on Linux. Let the
	// standard netlink attribute parser handle length/alignment validation.
	attributes := message
	attributes.Header.Type = unix.RTM_NEWROUTE
	attrs, err := syscall.ParseNetlinkRouteAttr(&attributes)
	if err != nil {
		return netip.Addr{}, false
	}

	var address netip.Addr
	var next linuxNeighbor
	for _, attr := range attrs {
		switch attr.Attr.Type {
		case unix.NDA_DST:
			if family == unix.AF_INET && len(attr.Value) == 4 || family == unix.AF_INET6 && len(attr.Value) == 16 {
				address, _ = netip.AddrFromSlice(attr.Value)
			}
		case unix.NDA_LLADDR:
			if len(attr.Value) == len(next.mac) {
				copy(next.mac[:], attr.Value)
				next.hasMAC = true
			}
		}
	}

	if !address.IsValid() {
		return netip.Addr{}, false
	}

	key := linuxNeighborKey{index: index, address: address}
	if message.Header.Type == unix.RTM_DELNEIGH {
		delete(s.neighbors, key)
		return netip.Addr{}, false
	}

	old := s.neighbors[key]
	if s.neighbors == nil {
		s.neighbors = make(map[linuxNeighborKey]linuxNeighbor)
	}

	s.neighbors[key] = next
	return address, next.hasMAC && !old.hasMAC
}

func linuxIdentityRoute(message syscall.NetlinkMessage) bool {
	return len(message.Data) >= unix.SizeofRtMsg &&
		(len(message.Data)-unix.SizeofRtMsg)%unix.RTA_ALIGNTO == 0 &&
		(message.Data[0] == unix.AF_INET || message.Data[0] == unix.AF_INET6) &&
		message.Data[6] == unix.RT_SCOPE_UNIVERSE && message.Data[7] == unix.RTN_UNICAST
}
