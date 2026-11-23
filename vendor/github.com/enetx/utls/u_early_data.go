// SPDX-License-Identifier: BSD-3-Clause
package tls

import "io"

// EarlyDataExtension places a QUIC early_data offer in a custom template.
// The engine emits it only when a valid session permits 0-RTT, after checking
// cipher, ALPN and application session metadata. TCP never enables it.
type EarlyDataExtension struct{ enabled bool }

func (e *EarlyDataExtension) Len() int {
	if e.enabled {
		return 4
	}
	return 0
}

func (e *EarlyDataExtension) Read(b []byte) (int, error) {
	if !e.enabled {
		return 0, io.EOF
	}
	return (&GenericExtension{Id: extensionEarlyData}).Read(b)
}

func (*EarlyDataExtension) writeToUConn(*UConn) error { return nil }

func (u *UConn) setEarlyDataExtension(enabled bool) {
	for _, ext := range u.Extensions {
		if early, ok := ext.(*EarlyDataExtension); ok {
			early.enabled = enabled
			return
		}
	}
	if !enabled {
		return
	}
	// PSK must remain the final extension (RFC 8446 section 4.2.11).
	index := len(u.Extensions)
	for i, ext := range u.Extensions {
		if _, ok := ext.(PreSharedKeyExtension); ok {
			index = i
			break
		}
	}
	u.Extensions = append(u.Extensions, nil)
	copy(u.Extensions[index+1:], u.Extensions[index:])
	u.Extensions[index] = &EarlyDataExtension{enabled: true}
}
