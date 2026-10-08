package util

import (
	"net/url"
	"sync"
)

// Local stream proxies are HTTP servers GoAnime itself runs on loopback to hand
// mpv and the downloader a plain media URL for a stream that has to be rewritten
// on the way — a CDN that wants headers the player cannot send, or a file whose
// first bytes arrive encrypted.
//
// The downloader's HTTP client refuses loopback on purpose (SSRF guard), which
// would also refuse these. This registry is the exact set of addresses that are
// GoAnime's own listeners, so the exemption is no wider than that.
var localProxyHosts sync.Map // "127.0.0.1:port" -> struct{}

// RegisterLocalProxyHost records host:port as one of GoAnime's own loopback
// stream proxies.
func RegisterLocalProxyHost(hostport string) {
	localProxyHosts.Store(hostport, struct{}{})
}

// IsLocalProxyURL reports whether raw points at a registered local proxy.
func IsLocalProxyURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "http" || u.Host == "" {
		return false
	}
	_, ok := localProxyHosts.Load(u.Host)
	return ok
}
