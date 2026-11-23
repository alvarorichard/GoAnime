package desktop

import (
	"github.com/enetx/surf/profiles"
	"github.com/enetx/surf/profiles/chrome/common"
)

// ConfigureTCP applies Chromium desktop TLS records, shutdown and TCP keepalive.
// Mobile socket defaults have a separate lifetime and are not selected here.
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/socket/tcp_socket_posix.cc
func ConfigureTCP(config *profiles.TCPConfig) {
	common.ConfigureTCP(config)
	config.KeepAlivePolicy(common.KeepAlive)
}
