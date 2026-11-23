//go:build !linux

package desktop

import "net/netip"

func networkIdentityReader() func([]netip.Addr) (string, error) { return nil }
