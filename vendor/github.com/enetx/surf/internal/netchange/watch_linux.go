package netchange

import (
	"encoding/binary"
	"encoding/hex"
	"errors"
	"maps"
	"net"
	"os"
	"sort"
	"strconv"
	"strings"
	"syscall"

	"golang.org/x/sys/unix"
)

func watch(notify func(Event)) (func(), error) {
	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW|unix.SOCK_CLOEXEC|unix.SOCK_NONBLOCK, unix.NETLINK_ROUTE)
	if err != nil {
		return nil, err
	}

	if err := unix.Bind(fd, &unix.SockaddrNetlink{
		Family: unix.AF_NETLINK,
		Groups: unix.RTMGRP_IPV4_IFADDR | unix.RTMGRP_IPV6_IFADDR | unix.RTMGRP_LINK |
			unix.RTMGRP_IPV4_ROUTE | unix.RTMGRP_IPV6_ROUTE | unix.RTMGRP_NEIGH,
	}); err != nil {
		unix.Close(fd)
		return nil, err
	}

	file := os.NewFile(uintptr(fd), "surf-network-notifications")
	state, err := newLinuxState()
	if err != nil {
		file.Close()
		return nil, err
	}

	done := make(chan struct{})
	go func() {
		defer close(done)

		buffer := make([]byte, 65536)
		for {
			n, err := file.Read(buffer)
			if err != nil {
				if errors.Is(err, syscall.ENOBUFS) {
					if next, syncErr := newLinuxState(); syncErr == nil {
						notify(Event{
							Address:       !maps.Equal(state.addresses, next.addresses),
							Link:          !maps.Equal(state.links, next.links),
							Route:         !maps.Equal(state.routes, next.routes),
							IdentityRoute: true,
						})
						state = next
					}
					continue
				}
				return
			}

			messages, err := syscall.ParseNetlinkMessage(buffer[:n])
			if err != nil {
				continue
			}

			event := state.apply(messages)
			if event.Address || event.Link || event.Route || event.IdentityRoute || len(event.Neighbors) != 0 {
				notify(event)
			}
		}
	}()

	return func() { file.Close(); <-done }, nil
}

// Maintain per-message state, rather than resampling all addresses after each
// batch. A resample can see a later queued change and accidentally suppress its
// notification (for example, stable IPv6 added after a temporary address).
type linuxState struct {
	routes    map[string]struct{}
	addresses map[string]string
	links     map[uint32]uint32
	loopbacks map[uint32]bool
	ethernet  map[uint32]bool
	neighbors map[linuxNeighborKey]linuxNeighbor
	gateways  map[string]linuxGateway
}

func newLinuxState() (*linuxState, error) {
	state := &linuxState{
		addresses: make(map[string]string),
		links:     make(map[uint32]uint32),
		loopbacks: make(map[uint32]bool),
		ethernet:  make(map[uint32]bool),
		neighbors: make(map[linuxNeighborKey]linuxNeighbor),
	}

	for _, kind := range []int{unix.RTM_GETLINK, unix.RTM_GETADDR, unix.RTM_GETROUTE, unix.RTM_GETNEIGH} {
		data, err := syscall.NetlinkRIB(kind, unix.AF_UNSPEC)
		if err != nil {
			return nil, err
		}

		messages, err := syscall.ParseNetlinkMessage(data)
		if err != nil {
			return nil, err
		}

		state.apply(messages)
	}

	return state, nil
}

func (s *linuxState) apply(messages []syscall.NetlinkMessage) Event {
	event := Event{TemporaryIPv6Only: true, IPv6Only: true}
	for _, message := range messages {
		switch message.Header.Type {
		case unix.RTM_NEWLINK, unix.RTM_DELLINK:
			if len(message.Data) < unix.SizeofIfInfomsg {
				continue
			}
			index := binary.NativeEndian.Uint32(message.Data[4:8])
			flags := binary.NativeEndian.Uint32(message.Data[8:12])
			s.identityLink(message, index, flags)
			old, exists := s.links[index]
			loopback := s.loopbacks[index] || flags&unix.IFF_LOOPBACK != 0

			if message.Header.Type == unix.RTM_DELLINK {
				delete(s.links, index)
				delete(s.loopbacks, index)
				event.Link = event.Link || exists && !loopback
			} else {
				value := flags & (unix.IFF_UP | unix.IFF_RUNNING)
				s.links[index] = value
				s.loopbacks[index] = loopback
				event.Link = event.Link || !loopback && (!exists || old != value)
			}
		case unix.RTM_NEWADDR, unix.RTM_DELADDR:
			if len(message.Data) < unix.SizeofIfAddrmsg {
				continue
			}

			index := binary.NativeEndian.Uint32(message.Data[4:8])
			if s.loopbacks[index] {
				continue
			}

			family := message.Data[0]
			if family != unix.AF_INET && family != unix.AF_INET6 {
				continue
			}

			attrs, err := syscall.ParseNetlinkRouteAttr(&message)
			if err != nil {
				continue
			}

			flags := uint32(message.Data[2])

			var address, local []byte
			for _, attr := range attrs {
				switch attr.Attr.Type {
				case unix.IFA_FLAGS:
					if len(attr.Value) >= 4 {
						flags = binary.NativeEndian.Uint32(attr.Value)
					}
				case unix.IFA_ADDRESS:
					address = attr.Value
				case unix.IFA_LOCAL:
					local = attr.Value
				}
			}

			if family == unix.AF_INET && len(local) > 0 {
				address = local
			}

			if family == unix.AF_INET && len(address) != net.IPv4len ||
				family == unix.AF_INET6 && len(address) != net.IPv6len {
				continue
			}

			key := strconv.FormatUint(
				uint64(index),
				10,
			) + "/" + string(
				[]byte{family, message.Data[1]},
			) + string(
				address,
			)

			value := strconv.FormatUint(uint64(flags), 10) + "/" + strconv.Itoa(int(message.Data[3]))
			old, exists := s.addresses[key]
			changed := false

			if message.Header.Type == unix.RTM_DELADDR {
				delete(s.addresses, key)
				changed = exists
			} else {
				s.addresses[key] = value
				changed = !exists || old != value
			}

			if !changed {
				continue
			}

			event.Address = true
			event.IPv6Only = event.IPv6Only && family == unix.AF_INET6
			event.TemporaryIPv6Only = event.TemporaryIPv6Only && family == unix.AF_INET6 &&
				flags&unix.IFA_F_TEMPORARY != 0
		case unix.RTM_NEWROUTE, unix.RTM_DELROUTE:
			event.IdentityRoute = event.IdentityRoute || linuxIdentityRoute(message)
			s.trackGateway(message)
			if key := linuxDefaultRoute(message); key != "" {
				_, exists := s.routes[key]
				if message.Header.Type == unix.RTM_DELROUTE {
					delete(s.routes, key)
					event.Route = event.Route || exists
				} else {
					if s.routes == nil {
						s.routes = make(map[string]struct{})
					}
					s.routes[key] = struct{}{}
					event.Route = event.Route || !exists
				}
			}
		case unix.RTM_NEWNEIGH, unix.RTM_DELNEIGH:
			if address, acquired := s.neighbor(message); acquired {
				event.Neighbors = append(event.Neighbors, address)
			}
		}
	}

	event.IPv6Only = event.IPv6Only && event.Address
	event.TemporaryIPv6Only = event.TemporaryIPv6Only && event.Address

	return event
}

func linuxDefaultRoute(message syscall.NetlinkMessage) string {
	if len(message.Data) < unix.SizeofRtMsg || message.Data[1] != 0 || message.Data[4] != unix.RT_TABLE_MAIN ||
		(message.Data[0] != unix.AF_INET && message.Data[0] != unix.AF_INET6) {
		return ""
	}
	attrs, err := syscall.ParseNetlinkRouteAttr(&message)
	if err != nil {
		return ""
	}
	parts := []string{string(message.Data[:1])}
	for _, attr := range attrs {
		switch attr.Attr.Type {
		case unix.RTA_GATEWAY, unix.RTA_OIF, unix.RTA_PRIORITY, unix.RTA_MULTIPATH:
			parts = append(parts, strconv.Itoa(int(attr.Attr.Type))+"/"+hex.EncodeToString(attr.Value))
		}
	}
	sort.Strings(parts)
	return strings.Join(parts, "\x00")
}
