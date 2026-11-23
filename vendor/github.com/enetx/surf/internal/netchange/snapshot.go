package netchange

import (
	"net"
	"sort"
	"strconv"
	"strings"
)

type snapshot struct{ addresses, links, ipv4 string }

func readSnapshot() (snapshot, error) {
	interfaces, err := net.Interfaces()
	if err != nil {
		return snapshot{}, err
	}

	var addresses, links, ipv4 []string
	for _, iface := range interfaces {
		if iface.Flags&net.FlagLoopback != 0 {
			continue
		}

		key := strconv.Itoa(iface.Index) + "/"
		links = append(links, key+strconv.Itoa(int(iface.Flags&(net.FlagUp|net.FlagRunning))))

		addrs, err := iface.Addrs()
		if err != nil {
			return snapshot{}, err
		}

		for _, addr := range addrs {
			addresses = append(addresses, key+addr.String())
			if ip, _, err := net.ParseCIDR(addr.String()); err == nil && ip.To4() != nil {
				ipv4 = append(ipv4, key+addr.String())
			}
		}
	}

	sort.Strings(addresses)
	sort.Strings(links)
	sort.Strings(ipv4)

	return snapshot{
		addresses: strings.Join(addresses, "\x00"),
		links:     strings.Join(links, "\x00"),
		ipv4:      strings.Join(ipv4, "\x00"),
	}, nil
}

func (s snapshot) difference(next snapshot) Event {
	changed := s.addresses != next.addresses
	return Event{Address: changed, Link: s.links != next.links, IPv6Only: changed && s.ipv4 == next.ipv4}
}
