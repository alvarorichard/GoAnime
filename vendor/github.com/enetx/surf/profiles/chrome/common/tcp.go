package common

import "github.com/enetx/surf/profiles"

// ConfigureTCP follows BoringSSL's full-sized application records and Chromium's
// socket-only Disconnect. These policies apply to HTTP/1 and HTTP/2 over TCP.
// https://github.com/google/boringssl/blob/ea9f3a34c3977144c45d272be81842fd891a0bdb/ssl/s3_pkt.cc
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/socket/ssl_client_socket_impl.cc
func ConfigureTCP(config *profiles.TCPConfig) {
	config.DynamicRecordSizing(false).CloseNotify(false)
}
