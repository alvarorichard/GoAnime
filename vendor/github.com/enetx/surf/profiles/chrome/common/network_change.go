package common

import "github.com/enetx/surf/profiles"

// ConfigureNetworkChange follows Chromium's default IP-change observers.
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/spdy/spdy_session_pool.cc
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/quic/quic_context.h
func ConfigureNetworkChange(config *profiles.NetworkChangeConfig) {
	config.Automatic(true).CoalescingDelay(0).Policy(NetworkChange)
}

// NetworkChange retires TCP sockets on stable address changes. Healthy QUIC,
// confirmed support, tickets and ordinary timed broken alternatives survive.
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/http/http_stream_pool.cc
// https://github.com/chromium/chromium/blob/158.0.8127.18/net/http/broken_alternative_services.cc
func NetworkChange(info profiles.NetworkChangeInfo) profiles.NetworkChangeAction {
	if !info.AddressChanged || info.TemporaryIPv6Only {
		return profiles.NetworkChangeAction{}
	}
	return profiles.NetworkChangeAction{CloseHTTP2: true, RetireHTTP1: true}
}
