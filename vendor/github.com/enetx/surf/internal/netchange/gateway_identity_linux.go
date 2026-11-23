package netchange

import (
	"bytes"
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"errors"
	"net/netip"
	"slices"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/sys/unix"
)

type linuxGateway struct {
	index   uint32
	address netip.Addr
}

var errUnrelatedNeighbor = errors.New("neighbor is outside the selected gateway routes")

// GatewayIdentityReader constructs a local Ethernet gateway identity reader.
// Route targets are numeric kernel FIB queries, not outgoing IP/DNS requests.
// Unrelated neighbor acquisition returns an ignored hint without resampling;
// a nil input refreshes it. Each subscription retains its own result, so sharing
// the reader cannot advance another subscriber's identity baseline.
// This identifies gateway MAC history, not a browser's salted persistent ID or
// its separate cellular/VPN interface-prefix identity mechanisms.
func GatewayIdentityReader(targets ...netip.Addr) func([]netip.Addr) (string, error) {
	targets = slices.Clone(targets)
	var mu sync.Mutex
	var selected map[netip.Addr]struct{}

	return func(acquired []netip.Addr) (string, error) {
		mu.Lock()
		defer mu.Unlock()

		if selected != nil && len(acquired) != 0 {
			relevant := false
			for _, address := range acquired {
				_, found := selected[address]
				relevant = relevant || found
			}

			if !relevant {
				return "", errUnrelatedNeighbor
			}
		}

		state, err := newLinuxState()
		if err != nil {
			return "", err
		}

		defaults := make([]linuxGateway, 0, len(state.gateways))
		for _, route := range state.gateways {
			defaults = append(defaults, route)
		}
		checked := make([]linuxGateway, 0, len(targets))
		for _, target := range targets {
			route, err := lookupGateway(target)
			if err != nil {
				return "", err
			}

			checked = append(checked, route)
		}

		value, nextSelected := gatewayIdentity(state, defaults, checked)
		selected = nextSelected
		return value, nil
	}
}

func gatewayRoute(message syscall.NetlinkMessage) (linuxGateway, bool) {
	if !linuxIdentityRoute(message) {
		return linuxGateway{}, false
	}

	attrs, err := syscall.ParseNetlinkRouteAttr(&message)
	if err != nil {
		return linuxGateway{}, false
	}

	var route linuxGateway
	for _, attr := range attrs {
		switch attr.Attr.Type {
		case unix.RTA_OIF:
			if len(attr.Value) == 4 {
				route.index = binary.NativeEndian.Uint32(attr.Value)
			}
		case unix.RTA_GATEWAY:
			if message.Data[0] == unix.AF_INET && len(attr.Value) == 4 ||
				message.Data[0] == unix.AF_INET6 && len(attr.Value) == 16 {
				route.address, _ = netip.AddrFromSlice(attr.Value)
			}
		}
	}

	return route, route.index != 0
}

func (s *linuxState) trackGateway(message syscall.NetlinkMessage) {
	route, valid := gatewayRoute(message)
	if !valid || message.Data[1] != 0 || !route.address.IsValid() ||
		route.address.Is6() && !route.address.IsGlobalUnicast() {
		return
	}

	// Attribute order is not part of route identity. Keep distinct table,
	// priority and preferred-source routes so deleting one preserves others
	// using the same MAC, just as duplicate MACs are deduplicated when hashing.
	attrs, _ := syscall.ParseNetlinkRouteAttr(&message) // gatewayRoute validated it
	parts := []string{string(message.Data[:unix.SizeofRtMsg])}
	for _, attr := range attrs {
		switch attr.Attr.Type {
		case unix.RTA_GATEWAY, unix.RTA_OIF, unix.RTA_PRIORITY, unix.RTA_TABLE, unix.RTA_PREFSRC:
			parts = append(parts, strconv.Itoa(int(attr.Attr.Type))+"/"+hex.EncodeToString(attr.Value))
		}
	}

	slices.Sort(parts)
	key := strings.Join(parts, "\x00")
	if message.Header.Type == unix.RTM_DELROUTE {
		delete(s.gateways, key)
		return
	}

	if s.gateways == nil {
		s.gateways = make(map[string]linuxGateway)
	}

	s.gateways[key] = route
}

func gatewayIdentity(state *linuxState, defaults, checked []linuxGateway) (string, map[netip.Addr]struct{}) {
	selected := make(map[netip.Addr]struct{})
	var material []byte
	linkUp := slices.ContainsFunc(checked, func(route linuxGateway) bool { return route.index != 0 })

	for _, bits := range []int{32, 128} {
		unique := make(map[[6]byte]struct{})
		for _, route := range defaults {
			if route.address.BitLen() != bits || !state.ethernet[route.index] || !state.identityLinkUp(route.index) {
				continue
			}

			selected[route.address] = struct{}{}
			if neighbor := state.neighbors[linuxNeighborKey{route.index, route.address}]; neighbor.hasMAC && linkUp {
				unique[neighbor.mac] = struct{}{}
			}
		}

		macs := make([][6]byte, 0, len(unique))
		for mac := range unique {
			macs = append(macs, mac)
		}

		slices.SortFunc(macs, func(a, b [6]byte) int { return bytes.Compare(a[:], b[:]) })
		for _, mac := range macs {
			material = append(material, mac[:]...)
		}

		for _, route := range checked {
			if route.address.BitLen() != bits || !state.ethernet[route.index] {
				continue
			}

			selected[route.address] = struct{}{}
			if neighbor := state.neighbors[linuxNeighborKey{route.index, route.address}]; neighbor.hasMAC {
				material = append(material, neighbor.mac[:]...)
			}
		}
	}

	if len(material) == 0 {
		return "", selected
	}

	digest := sha256.Sum256(material)
	return hex.EncodeToString(digest[:]), selected
}

func (s *linuxState) identityLinkUp(index uint32) bool {
	if s.links[index]&unix.IFF_RUNNING == 0 || s.loopbacks[index] {
		return false
	}

	prefix := strconv.FormatUint(uint64(index), 10) + "/"
	for key, value := range s.addresses {
		if strings.HasPrefix(key, prefix) && strings.HasSuffix(value, "/0") { // RT_SCOPE_UNIVERSE
			return true
		}
	}

	return false
}

func lookupGateway(target netip.Addr) (linuxGateway, error) {
	if !target.IsValid() || target.Zone() != "" {
		return linuxGateway{}, syscall.EINVAL
	}

	fd, err := unix.Socket(unix.AF_NETLINK, unix.SOCK_RAW|unix.SOCK_CLOEXEC, unix.NETLINK_ROUTE)
	if err != nil {
		return linuxGateway{}, err
	}
	defer unix.Close(fd)

	if err := unix.SetsockoptTimeval(fd, unix.SOL_SOCKET, unix.SO_RCVTIMEO, &unix.Timeval{Sec: 2}); err != nil {
		return linuxGateway{}, err
	}

	kernel := &unix.SockaddrNetlink{Family: unix.AF_NETLINK}
	if err := unix.Bind(fd, kernel); err != nil {
		return linuxGateway{}, err
	}

	payload := target.AsSlice()
	request := make([]byte, unix.NLMSG_HDRLEN+unix.SizeofRtMsg+unix.SizeofRtAttr+len(payload))
	binary.NativeEndian.PutUint32(request[:4], uint32(len(request)))
	binary.NativeEndian.PutUint16(request[4:6], unix.RTM_GETROUTE)
	binary.NativeEndian.PutUint16(request[6:8], unix.NLM_F_REQUEST)
	binary.NativeEndian.PutUint32(request[8:12], 1)
	request[unix.NLMSG_HDRLEN], request[unix.NLMSG_HDRLEN+1] = unix.AF_INET6, byte(target.BitLen())
	if target.Is4() {
		request[unix.NLMSG_HDRLEN] = unix.AF_INET
	}

	attribute := request[unix.NLMSG_HDRLEN+unix.SizeofRtMsg:]
	binary.NativeEndian.PutUint16(attribute[:2], uint16(len(attribute)))
	binary.NativeEndian.PutUint16(attribute[2:4], unix.RTA_DST)
	copy(attribute[4:], payload)
	if err := unix.Sendto(fd, request, 0, kernel); err != nil {
		return linuxGateway{}, err
	}

	buffer := make([]byte, 65536)
	n, _, err := unix.Recvfrom(fd, buffer, 0)
	if err != nil {
		return linuxGateway{}, err
	}

	return gatewayReply(buffer[:n])
}

func gatewayReply(data []byte) (linuxGateway, error) {
	messages, err := syscall.ParseNetlinkMessage(data)
	if err != nil {
		return linuxGateway{}, err
	}

	for _, message := range messages {
		if message.Header.Seq != 1 {
			continue
		}

		if message.Header.Type == unix.NLMSG_ERROR {
			if len(message.Data) < 4 {
				return linuxGateway{}, syscall.EINVAL
			}

			err := syscall.Errno(-int32(binary.NativeEndian.Uint32(message.Data)))
			if err == 0 || errors.Is(err, syscall.ENETUNREACH) || errors.Is(err, syscall.EHOSTUNREACH) || errors.Is(err, syscall.ESRCH) {
				return linuxGateway{}, nil
			}

			return linuxGateway{}, err
		}

		if message.Header.Type == unix.RTM_NEWROUTE {
			route, _ := gatewayRoute(message)
			return route, nil
		}
	}

	return linuxGateway{}, syscall.EINVAL
}
