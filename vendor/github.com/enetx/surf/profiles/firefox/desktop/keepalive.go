package desktop

import (
	"net"
	"runtime"
	"time"

	"github.com/enetx/surf/profiles"
)

// KeepAlive follows StartShortLivedTCPKeepalives, its probe-adjusted transition
// timer and StartLongLivedTCPKeepalives. StartSpdy disables both the socket
// keepalive and transition timer. Values follow the host kernel, independently
// of the Windows User-Agent overlay; Darwin sets only the socket idle period.
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/protocol/http/nsHttpConnection.cpp
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/base/nsSocketTransport2.cpp
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/netwerk/base/nsSocketTransportService2.h
// https://github.com/mozilla-firefox/firefox/blob/FIREFOX_160_0_RELEASE/modules/libpref/init/all.js
func KeepAlive(info profiles.TCPKeepAliveInfo) profiles.TCPKeepAliveOptions {
	if info.Protocol == "h2" {
		return profiles.TCPKeepAliveOptions{}
	}

	interval := max(info.DialDuration/time.Second, 1) * time.Second
	probes, socketProbes := 4, 4
	socketInterval := interval
	switch runtime.GOOS {
	case "darwin":
		probes, socketProbes, socketInterval = 8, -1, -1
	case "windows":
		probes, socketProbes = 10, -1
	}

	short := net.KeepAliveConfig{Enable: true, Idle: 10 * time.Second, Interval: socketInterval, Count: socketProbes}
	long := short
	long.Idle = 600 * time.Second
	return profiles.TCPKeepAliveOptions{
		Initial: short, After: 60*time.Second + time.Duration(probes)*interval + 2*time.Second,
		LongLived: long,
	}
}
