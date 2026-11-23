package common

import (
	"net"
	"runtime"
	"time"

	"github.com/enetx/surf/profiles"
)

// KeepAlive follows TCPSocketPosix::SetDefaultOptionsForClient and
// SetTCPKeepAlive. Darwin sets only the idle period; other desktop targets
// also set the retry interval. Probe count remains the operating system's.
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/socket/tcp_socket_posix.cc
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/socket/tcp_socket_win.cc
func KeepAlive(profiles.TCPKeepAliveInfo) profiles.TCPKeepAliveOptions {
	interval := 45 * time.Second
	if runtime.GOOS == "darwin" {
		interval = -1
	}

	return profiles.TCPKeepAliveOptions{Initial: net.KeepAliveConfig{
		Enable: true, Idle: 45 * time.Second, Interval: interval, Count: -1,
	}}
}
