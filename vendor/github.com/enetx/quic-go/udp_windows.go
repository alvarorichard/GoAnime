//go:build windows

package quic

import "github.com/enetx/quic-go/internal/udp"

// IP_DONTFRAGMENT is the Windows IPv4 socket option retained for API compatibility.
const IP_DONTFRAGMENT = udp.IP_DONTFRAGMENT

// IPV6_DONTFRAG is the Windows IPv6 socket option retained for API compatibility.
const IPV6_DONTFRAG = udp.IPV6_DONTFRAG
